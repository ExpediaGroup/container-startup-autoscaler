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
	ccontext "github.com/ExpediaGroup/container-startup-autoscaler/internal/context"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/event/eventcommon"
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
	waitForCacheUpdateMaxWaitSecs = 5
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

// Patch applies the mutations dictated by podMutationFuncs to either the 'resize' subresource of the supplied pod, or
// the pod itself. If any podMutationFunc specifies a non-nil function return value, it waits for the patched pod to be
// updated in the informer cache using the conditions specified by these functions.  It returns the new server
// representation of the pod. The patch is retried and specially handled if there's a conflict: the latest version is
// retrieved and the mutations are reapplied before attempting again. The supplied pod is never mutated.
func (h *podHelper) Patch(
	ctx context.Context,
	podEventPublisher eventcommon.PodEventPublisher,
	pod *v1.Pod,
	podMutationFuncs []func(podToMutate *v1.Pod) (bool, func(currentPod *v1.Pod) bool, error),
	patchResize bool,
) (*v1.Pod, error) {
	mutatedPod := pod.DeepCopy() // Apply mutations to a copy.
	var waitCacheConditionsMetFuncs []func(*v1.Pod) bool
	shouldPatch := false

	for _, podMutationFunc := range podMutationFuncs {
		shouldPatchFunc, waitCacheConditionsMetFunc, err := podMutationFunc(mutatedPod)
		if err != nil {
			return nil, common.WrapErrorf(err, "unable to mutate pod")
		}

		waitCacheConditionsMetFuncs = append(waitCacheConditionsMetFuncs, waitCacheConditionsMetFunc)
		shouldPatch = shouldPatch || shouldPatchFunc
	}

	// Only patch if at least one podMutationFunc indicated to do so.
	if !shouldPatch {
		return pod, nil
	}

	var podEventCh <-chan eventcommon.PodEvent
	shouldWaitForCacheUpdate := h.shouldWaitForCacheUpdate(waitCacheConditionsMetFuncs)
	if shouldWaitForCacheUpdate {
		defer func() { podEventPublisher.Unsubscribe(podEventCh) }()
	}

	var err error
	retryableFunc := func() error {
		if shouldWaitForCacheUpdate {
			// Subscribe to pod update events for later informer cache waiting.
			podEventCh = podEventPublisher.Subscribe(
				mutatedPod.Namespace,
				mutatedPod.Name,
				[]eventcommon.PodEventType{eventcommon.PodEventTypeUpdate},
			)
		}

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
					_, _, _ = podMutationFunc(mutatedPod)
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

	// If indicated, wait for the patched pod to be updated in the informer cache. For example, this is necessary
	// when updating the status annotation since the cache may not be updated immediately upon the next
	// reconciliation, leading to inaccurate status updates that rely on accurate current status. The reconciler
	// doesn't allow concurrent reconciles for same pod so subsequent reconciles will not start until this wait has
	// completed.
	if shouldWaitForCacheUpdate {
		mutatedPod = h.waitForCacheUpdate(ctx, mutatedPod, waitCacheConditionsMetFuncs, podEventCh)
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

// shouldWaitForCacheUpdate determines whether the local informer cache should be waited upon based on the supplied
// conditions functions.
func (h *podHelper) shouldWaitForCacheUpdate(conditionsMetFuncs []func(*v1.Pod) bool) bool {
	if len(conditionsMetFuncs) == 0 {
		return false
	}

	allNil := true
	for _, conditionsMetFunc := range conditionsMetFuncs {
		if conditionsMetFunc != nil {
			allNil = false
			break
		}
	}
	if allNil {
		return false
	}

	return true
}

// waitForCacheUpdate waits for the local informer cache to reflect the pod with the conditions specified within
// conditionsMetFuncs. Returns the new representation of the pod if found within a timeout period, otherwise the
// original pod.
func (h *podHelper) waitForCacheUpdate(
	ctx context.Context,
	pod *v1.Pod,
	conditionsMetFuncs []func(*v1.Pod) bool,
	podEventCh <-chan eventcommon.PodEvent,
) *v1.Pod {
	var timeoutDuration = waitForCacheUpdateMaxWaitSecs * time.Second
	timeoutOverride := ccontext.TimeoutOverride(ctx)
	if timeoutOverride != 0 {
		timeoutDuration = timeoutOverride
		logging.Infof(ctx, logging.VInfo, "default cache update timeout overridden")
	}

	for {
		select {
		case event := <-podEventCh:
			if event.EventType != eventcommon.PodEventTypeUpdate {
				panic(fmt.Errorf("unexpected event type '%s'", event.EventType))
			}

			allConditionsMet := true
			for _, conditionsMetFunc := range conditionsMetFuncs {
				if !conditionsMetFunc(event.Pod) {
					allConditionsMet = false
				}
			}
			if allConditionsMet {
				return event.Pod
			}

		case <-time.After(timeoutDuration):
			logging.Infof(ctx, logging.VDebug, "cache wasn't updated in time")
			informercache.SyncTimeout().Inc()
			return pod
		}
	}
}
