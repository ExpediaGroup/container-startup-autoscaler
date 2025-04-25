/*
Copyright 2025 Expedia Group, Inc.

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

package kube

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/informercache"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/retry"
	retrygo "github.com/avast/retry-go/v4"
	"k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	waitForCacheUpdatePollMillis  = 100
	waitForCacheUpdateMaxWaitSecs = 3
)

type mapFor string

const (
	mapForLabel      mapFor = "label"
	mapForAnnotation mapFor = "annotation"
)

// podHelper is the default implementation of kubecommon.PodHelper.
type podHelper struct {
	client client.Client
}

func NewPodHelper(client client.Client) kubecommon.PodHelper {
	return &podHelper{client: client}
}

// Get returns the pod with the supplied name, along with whether the pod exists.
func (h *podHelper) Get(ctx context.Context, name types.NamespacedName) (bool, *v1.Pod, error) {
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

// Patch applies the mutations dictated by mutatePodFuncs to either the 'resize' subresource of the supplied pod, or the
// pod itself. If mustSyncCache is true, it waits for the patched pod to be updated in the informer cache. It returns
// the new server representation of the pod. The patch is retried and specially handled if there's a conflict: the
// latest version is retrieved and the mutations are reapplied before attempting again. The supplied pod is never
// mutated.
func (h *podHelper) Patch(
	ctx context.Context,
	pod *v1.Pod,
	podMutationFuncs []func(*v1.Pod) error,
	patchResize bool,
	mustSyncCache bool,
) (*v1.Pod, error) {
	// Copy and apply mutations.
	mutatedPod := pod.DeepCopy()

	for _, podMutationFunc := range podMutationFuncs {
		err := podMutationFunc(mutatedPod)
		if err != nil {
			return nil, common.WrapErrorf(err, "unable to mutate pod")
		}
	}

	var err error
	retryableFunc := func() error {
		if patchResize {
			err = h.client.SubResource("resize").Patch(ctx, mutatedPod, client.MergeFrom(pod))
		} else {
			err = h.client.Patch(ctx, mutatedPod, client.MergeFrom(pod))
		}

		if err != nil {
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

				// Reapply mutations to latest pod.
				mutatedPod = latestPod
				for _, podMutationFunc := range podMutationFuncs {
					_ = podMutationFunc(mutatedPod)
				}
			}

			return err
		}

		return nil
	}

	err = retry.DoStandardRetryWithMoreOpts(ctx, retryableFunc, kubeApiRetryOptions(ctx))
	if err != nil {
		return nil, common.WrapErrorf(err, "unable to patch pod")
	}

	if mustSyncCache {
		// Wait for the patched pod to be updated in the informer cache. For example, this is necessary when updating
		// the status annotation since the cache may not be updated immediately upon the next reconciliation, leading
		// to inaccurate status updates that rely on accurate current status. The reconciler doesn't allow concurrent
		// reconciles for same pod so subsequent reconciles will not start until this wait has completed.
		_ = h.waitForCacheUpdate(ctx, mutatedPod)
	}

	return mutatedPod, nil
}

// HasAnnotation returns whether the supplied pod has the supplied name annotation.
func (h *podHelper) HasAnnotation(pod *v1.Pod, name string) (bool, string) {
	if value, has := pod.Annotations[name]; has {
		return true, value
	}

	return false, ""
}

// ExpectedLabelValueAs returns the value of the supplied pod's supplied name label, as a specific type. The label is
// expected to exist.
func (h *podHelper) ExpectedLabelValueAs(pod *v1.Pod, name string, as kubecommon.DataType) (any, error) {
	return h.expectedLabelOrAnnotationAs(mapForLabel, pod.Labels, name, as)
}

// ExpectedAnnotationValueAs returns the value of the supplied pod's supplied name annotation, as a specific type. The
// annotation is expected to exist.
func (h *podHelper) ExpectedAnnotationValueAs(pod *v1.Pod, name string, as kubecommon.DataType) (any, error) {
	return h.expectedLabelOrAnnotationAs(mapForAnnotation, pod.Annotations, name, as)
}

// IsContainerInSpec returns whether the container with the supplied containerName is present in the supplied pod's
// spec.
func (h *podHelper) IsContainerInSpec(pod *v1.Pod, containerName string) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			return true
		}
	}

	return false
}

// ResizeConditions returns the resize-related conditions for the supplied pod.
func (h *podHelper) ResizeConditions(pod *v1.Pod) []v1.PodCondition {
	var resizeConditions []v1.PodCondition

	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodResizePending || condition.Type == v1.PodResizeInProgress {
			resizeConditions = append(resizeConditions, condition)
		}
	}

	return resizeConditions
}

// QOSClass returns the QOS class of the supplied pod. It returns an error if the QOS class is not (yet) present in the
// pod's status.
func (h *podHelper) QOSClass(pod *v1.Pod) (v1.PodQOSClass, error) {
	if pod.Status.QOSClass == "" {
		return pod.Status.QOSClass, errors.New("pod status qos class not present")
	}

	return pod.Status.QOSClass, nil
}

// expectedLabelOrAnnotationAs retrieves an expected label or annotation and returns the indicated type.
func (h *podHelper) expectedLabelOrAnnotationAs(
	mapFor mapFor,
	m map[string]string,
	name string,
	as kubecommon.DataType,
) (any, error) {
	var value string
	var present bool
	if value, present = m[name]; !present {
		return nil, fmt.Errorf("%s '%s' not present", mapFor, name)
	}

	switch as {
	case kubecommon.DataTypeString:
		return value, nil
	case kubecommon.DataTypeBool:
		valueBool, err := strconv.ParseBool(value)
		if err != nil {
			return nil, common.WrapErrorf(err, "unable to parse '%s' %s value '%s' as %s", name, mapFor, value, as)
		}

		return valueBool, nil
	}

	panic(fmt.Errorf("as '%s' not supported", as))
}

// waitForCacheUpdate waits for the informer cache to update a pod with at least the resource version indicated by the
// supplied pod. Returns the new representation of the pod if found within a timeout period, otherwise nil.
func (h *podHelper) waitForCacheUpdate(ctx context.Context, pod *v1.Pod) *v1.Pod {
	ticker := time.NewTicker(waitForCacheUpdatePollMillis * time.Millisecond)
	defer ticker.Stop()
	timeout := time.After(waitForCacheUpdateMaxWaitSecs * time.Second)

	pollCount := 0
	for {
		select {
		case <-ticker.C:
			pollCount++
			exists, podFromCache, err := h.Get(ctx, types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
			if err == nil && exists && podFromCache.ResourceVersion >= pod.ResourceVersion {
				logging.Infof(ctx, logging.VDebug, "pod polled from cache %d time(s) in total", pollCount)
				informercache.SyncPoll().Observe(float64(pollCount))
				return podFromCache
			}

		case <-timeout:
			logging.Infof(ctx, logging.VDebug, "cache wasn't updated in time")
			informercache.SyncTimeout().Inc()
			return nil
		}
	}
}
