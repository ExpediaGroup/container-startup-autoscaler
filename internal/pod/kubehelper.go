/*
Copyright 2024 Expedia Group, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pod

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/retry"
	retrygo "github.com/avast/retry-go/v4"
	"k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KubeHelper performs operations relating to Kube pods.
type KubeHelper interface {
	Get(context.Context, types.NamespacedName) (bool, *v1.Pod, error)
	Patch(context.Context, *v1.Pod, func(*v1.Pod) (bool, *v1.Pod, error)) (*v1.Pod, error)
	UpdateContainerResources(
		context.Context,
		*v1.Pod,
		string,
		resource.Quantity, resource.Quantity,
		resource.Quantity, resource.Quantity,
		func(pod *v1.Pod) (bool, *v1.Pod, error),
	) (*v1.Pod, error)
	HasAnnotation(pod *v1.Pod, name string) (bool, string)
	ExpectedLabelValueAs(*v1.Pod, string, podcommon.Type) (any, error)
	ExpectedAnnotationValueAs(*v1.Pod, string, podcommon.Type) (any, error)
	IsContainerInSpec(*v1.Pod, string) bool
	ResizeStatus(*v1.Pod) v1.PodResizeStatus
}

type mapFor string

const (
	mapForLabel      mapFor = "label"
	mapForAnnotation mapFor = "annotation"
)

// kubeHelper is the default implementation of KubeHelper.
type kubeHelper struct {
	client client.Client
}

func newKubeHelper(client client.Client) *kubeHelper {
	return &kubeHelper{client: client}
}

// Get returns the pod with the supplied name, along with whether the pod exists.
func (h *kubeHelper) Get(ctx context.Context, name types.NamespacedName) (bool, *v1.Pod, error) {
	pod := &v1.Pod{}
	retryableFunc := func() error {
		return h.client.Get(ctx, name, pod)
	}

	err := retry.DoStandardRetryWithMoreOpts(ctx, retryableFunc, kubeApiRetryOptions(ctx))
	if err != nil {
		if kerrors.IsNotFound(err) {
			return false, &v1.Pod{}, nil
		}

		return false, &v1.Pod{}, common.WrapErrorf(err, "unable to get pod")
	}

	return true, pod, nil
}

// Patch patches the supplied pod with mutations dictated by the supplied mutatePodFunc and returns the new server
// representation of the pod. Patches are retried and specially handled if there's a conflict: the latest version is
// retrieved and the patch is reapplied before attempting again. The supplied pod is never mutated.
func (h *kubeHelper) Patch(
	ctx context.Context,
	pod *v1.Pod,
	mutatePodFunc func(*v1.Pod) (bool, *v1.Pod, error),
) (*v1.Pod, error) {
	shouldPatch, mutatedPod, err := mutatePodFunc(pod)
	if err != nil {
		return nil, common.WrapErrorf(err, "unable to mutate pod")
	}
	if !shouldPatch {
		return pod, nil
	}

	retryableFunc := func() error {
		if err = h.client.Patch(ctx, mutatedPod, client.MergeFrom(pod)); err != nil {
			if kerrors.IsConflict(err) {
				// Get latest pod and re-apply patch for next attempt.
				exists, latestPod, getErr := h.Get(ctx, types.NamespacedName{
					Namespace: pod.Namespace,
					Name:      pod.Name,
				})
				if getErr != nil {
					return common.WrapErrorf(err, "unable to get pod when resolving conflict")
				}
				if !exists {
					// Mark as unrecoverable so not to retry further.
					return retrygo.Unrecoverable(errors.New("pod doesn't exist when resolving conflict"))
				}

				_, mutatedPod, _ = mutatePodFunc(latestPod)
			}

			return err
		}

		return nil
	}

	err = retry.DoStandardRetryWithMoreOpts(ctx, retryableFunc, kubeApiRetryOptions(ctx))
	if err != nil {
		return nil, common.WrapErrorf(err, "unable to patch pod")
	}

	return mutatedPod, nil
}

// UpdateContainerResources updates the resources (requests and limits) of the supplied containerName within the
// supplied pod. Optional additional mutations may be supplied via addMutations. The update is serviced via a patch,
// which behaves per Patch. The supplied pod is never mutated. Returns the new server representation of the pod.
func (h *kubeHelper) UpdateContainerResources(
	ctx context.Context,
	pod *v1.Pod,
	containerName string,
	cpuRequests resource.Quantity, cpuLimits resource.Quantity,
	memoryRequests resource.Quantity, memoryLimits resource.Quantity,
	addMutations func(pod *v1.Pod) (bool, *v1.Pod, error),
) (*v1.Pod, error) {
	mutatePodFunc := func(pod *v1.Pod) (bool, *v1.Pod, error) {
		mutatedPod := pod.DeepCopy()
		var container *v1.Container

		for _, c := range mutatedPod.Spec.Containers {
			if c.Name == containerName {
				container = &c
				break
			}
		}
		if container == nil {
			return false, nil, errors.New("container not present")
		}

		container.Resources.Requests[v1.ResourceCPU] = cpuRequests
		container.Resources.Limits[v1.ResourceCPU] = cpuLimits
		container.Resources.Requests[v1.ResourceMemory] = memoryRequests
		container.Resources.Limits[v1.ResourceMemory] = memoryLimits

		if addMutations != nil {
			var err error
			// 'Should patch' ignored here as supplementary to patching resources.
			_, mutatedPod, err = addMutations(mutatedPod)
			if err != nil {
				return false, nil, common.WrapErrorf(err, "unable to apply additional pod mutations")
			}
		}

		return true, mutatedPod, nil
	}

	newPod, err := h.Patch(ctx, pod, mutatePodFunc)
	if err != nil {
		return nil, common.WrapErrorf(err, "unable to patch pod")
	}

	return newPod, nil
}

// HasAnnotation reports whether the supplied pod has the supplied name annotation.
func (h *kubeHelper) HasAnnotation(pod *v1.Pod, name string) (bool, string) {
	if value, has := pod.Annotations[name]; has {
		return true, value
	}

	return false, ""
}

// ExpectedLabelValueAs returns the value of the supplied pod's supplied name label, as a specific type. The label is
// expected to exist.
func (h *kubeHelper) ExpectedLabelValueAs(pod *v1.Pod, name string, as podcommon.Type) (any, error) {
	return h.expectedLabelOrAnnotationAs(mapForLabel, pod.Labels, name, as)
}

// ExpectedAnnotationValueAs returns the value of the supplied pod's supplied name annotation, as a specific type. The
// annotation is expected to exist.
func (h *kubeHelper) ExpectedAnnotationValueAs(pod *v1.Pod, name string, as podcommon.Type) (any, error) {
	return h.expectedLabelOrAnnotationAs(mapForAnnotation, pod.Annotations, name, as)
}

// IsContainerInSpec reports whether the container with the supplied containerName is present in the supplied pod's
// spec.
func (h *kubeHelper) IsContainerInSpec(pod *v1.Pod, containerName string) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			return true
		}
	}

	return false
}

// ResizeStatus returns the resize status for the supplied pod.
func (h *kubeHelper) ResizeStatus(pod *v1.Pod) v1.PodResizeStatus {
	return pod.Status.Resize
}

// expectedLabelOrAnnotationAs retrieves an expected label or annotation and returns the indicated type.
func (h *kubeHelper) expectedLabelOrAnnotationAs(
	mapFor mapFor,
	m map[string]string,
	name string,
	as podcommon.Type,
) (any, error) {
	var value string
	var present bool
	if value, present = m[name]; !present {
		return nil, fmt.Errorf("%s '%s' not present", mapFor, name)
	}

	switch as {
	case podcommon.TypeString:
		return value, nil
	case podcommon.TypeBool:
		valueBool, err := strconv.ParseBool(value)
		if err != nil {
			return nil, common.WrapErrorf(err, "unable to parse '%s' %s value '%s' as %s", name, mapFor, value, as)
		}

		return valueBool, nil
	}

	panic(fmt.Errorf("as '%s' not supported", as))
}
