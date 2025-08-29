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

package pod

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"k8s.io/api/core/v1"
)

var missingMemRegex = regexp.MustCompile(`missing pod memory usage|missing container .+ memory usage`)

// targetContainerState is the default implementation of podcommon.TargetContainerState.
type targetContainerState struct {
	podHelper       kubecommon.PodHelper
	containerHelper kubecommon.ContainerHelper
}

func newTargetContainerState(
	podHelper kubecommon.PodHelper,
	containerHelper kubecommon.ContainerHelper,
) targetContainerState {
	return targetContainerState{
		podHelper:       podHelper,
		containerHelper: containerHelper,
	}
}

// States calculates and returns states from the supplied pod and config.
func (s targetContainerState) States(
	ctx context.Context,
	pod *v1.Pod,
	targetContainer *v1.Container,
	scaleConfigs scalecommon.Configurations,
) (podcommon.States, error) {
	ret := podcommon.NewStatesAllUnknown()
	ret.StartupProbe = s.stateStartupProbe(targetContainer)
	ret.ReadinessProbe = s.stateReadinessProbe(targetContainer)

	var err error
	ret.Container, err = s.stateContainer(pod, targetContainer)
	if err != nil {
		if !s.shouldReturnError(ctx, err) {
			return ret, nil
		}
		return ret, common.WrapErrorf(err, "unable to determine container state")
	}

	ret.Started, err = s.stateStarted(pod, targetContainer)
	if err != nil {
		if !s.shouldReturnError(ctx, err) {
			return ret, nil
		}
		return ret, common.WrapErrorf(err, "unable to determine started state")
	}

	ret.Ready, err = s.stateReady(pod, targetContainer)
	if err != nil {
		if !s.shouldReturnError(ctx, err) {
			return ret, nil
		}
		return ret, common.WrapErrorf(err, "unable to determine ready state")
	}

	scaleStates := scale.NewStates(scaleConfigs, s.containerHelper)
	ret.Resources = s.stateResources(
		scaleStates.IsStartupConfigurationAppliedAll(targetContainer),
		scaleStates.IsPostStartupConfigurationAppliedAll(targetContainer),
	)

	ret.StatusResources, err = s.stateStatusResources(pod, targetContainer, scaleStates)
	if err != nil {
		if !s.shouldReturnError(ctx, err) {
			return ret, nil
		}
		return ret, common.WrapErrorf(err, "unable to determine status resources states")
	}

	ret.Resize, err = s.stateResize(pod)
	if err != nil {
		if !s.shouldReturnError(ctx, err) {
			return ret, nil
		}
		return ret, common.WrapErrorf(err, "unable to determine resize state")
	}

	return ret, nil
}

// stateStartupProbe returns the startup probe state for the target container.
func (s targetContainerState) stateStartupProbe(container *v1.Container) podcommon.StateBool {
	if s.containerHelper.HasStartupProbe(container) {
		return podcommon.StateBoolTrue
	}

	return podcommon.StateBoolFalse
}

// stateReadinessProbe returns the readiness probe state for the target container.
func (s targetContainerState) stateReadinessProbe(container *v1.Container) podcommon.StateBool {
	if s.containerHelper.HasReadinessProbe(container) {
		return podcommon.StateBoolTrue
	}

	return podcommon.StateBoolFalse
}

// stateContainer returns the container state for the target container.
func (s targetContainerState) stateContainer(pod *v1.Pod, targetContainer *v1.Container) (podcommon.StateContainer, error) {
	containerState, err := s.containerHelper.State(pod, targetContainer)
	if err != nil {
		return podcommon.StateContainerUnknown, common.WrapErrorf(err, "unable to get container state")
	}

	if containerState.Running != nil {
		return podcommon.StateContainerRunning, nil
	}

	if containerState.Waiting != nil {
		return podcommon.StateContainerWaiting, nil
	}

	if containerState.Terminated != nil {
		return podcommon.StateContainerTerminated, nil
	}

	return podcommon.StateContainerUnknown, nil
}

// stateStarted returns the ready state for the target container.
func (s targetContainerState) stateStarted(pod *v1.Pod, targetContainer *v1.Container) (podcommon.StateBool, error) {
	started, err := s.containerHelper.IsStarted(pod, targetContainer)
	if err != nil {
		return podcommon.StateBoolUnknown, common.WrapErrorf(err, "unable to get container ready status")
	}

	if started {
		return podcommon.StateBoolTrue, nil
	}

	return podcommon.StateBoolFalse, nil
}

// stateReady returns the ready state for the target container.
func (s targetContainerState) stateReady(pod *v1.Pod, targetContainer *v1.Container) (podcommon.StateBool, error) {
	ready, err := s.containerHelper.IsReady(pod, targetContainer)
	if err != nil {
		return podcommon.StateBoolUnknown, common.WrapErrorf(err, "unable to get container ready status")
	}

	if ready {
		return podcommon.StateBoolTrue, nil
	}

	return podcommon.StateBoolFalse, nil
}

// stateReady returns the resources state using the supplied startupConfigApplied and postStartupConfigApplied.
func (s targetContainerState) stateResources(
	startupConfigApplied bool,
	postStartupConfigApplied bool,
) podcommon.StateResources {
	if startupConfigApplied {
		return podcommon.StateResourcesStartup
	} else if postStartupConfigApplied {
		return podcommon.StateResourcesPostStartup
	} else {
		return podcommon.StateResourcesUnknown
	}
}

// stateStatusResources returns the status resources state for the target container.
func (s targetContainerState) stateStatusResources(
	pod *v1.Pod,
	targetContainer *v1.Container,
	scaleStates scalecommon.States,
) (podcommon.StateStatusResources, error) {
	zero, err := scaleStates.IsAnyCurrentZeroAll(pod, targetContainer)
	if err != nil {
		return podcommon.StateStatusResourcesUnknown, common.WrapErrorf(err, "unable to determine if any current resources are zero")
	}
	if zero {
		return podcommon.StateStatusResourcesIncomplete, nil
	}

	requestsMatch, err := scaleStates.DoesRequestsCurrentMatchSpecAll(pod, targetContainer)
	if err != nil {
		return podcommon.StateStatusResourcesUnknown, common.WrapErrorf(err, "unable to determine if current requests matches spec")
	}

	limitsMatch, err := scaleStates.DoesLimitsCurrentMatchSpecAll(pod, targetContainer)
	if err != nil {
		return podcommon.StateStatusResourcesUnknown, common.WrapErrorf(err, "unable to determine if current limits matches spec")
	}

	if requestsMatch && limitsMatch {
		return podcommon.StateStatusResourcesContainerResourcesMatch, nil
	}

	return podcommon.StateStatusResourcesContainerResourcesMismatch, nil
}

// stateResize returns the resize state for the pod.
func (s targetContainerState) stateResize(pod *v1.Pod) (podcommon.ResizeState, error) {
	// Reference: callers of SetPodResize*Condition in
	// https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/status/status_manager.go

	// Both resize conditions are potentially transient.
	resizeConditions := s.podHelper.ResizeConditions(pod)

	// Neither PodResizePending nor PodResizeInProgress conditions will be present if 1) a resize has not yet been
	// started or 2) has completed successfully.
	if len(resizeConditions) == 0 {
		return podcommon.NewResizeState(podcommon.StateResizeNotStartedOrCompleted, ""), nil
	}

	var condition v1.PodCondition

	// Both PodResizePending and PodResizeInProgress conditions will be present if a new resize was requested in the
	// middle of a previous pod resize that is still in progress. Inspect the PodResizePending condition in this case.
	if len(resizeConditions) == 2 {
		for _, cond := range resizeConditions {
			if cond.Type == v1.PodResizePending {
				condition = cond
				break
			}
		}
	} else {
		condition = resizeConditions[0]
	}

	// condition.Status is always ConditionTrue for both PodResizePending and PodResizeInProgress.
	if condition.Type == v1.PodResizePending {
		if condition.Reason == v1.PodReasonDeferred {
			return podcommon.NewResizeState(podcommon.StateResizeDeferred, condition.Message), nil
		}

		if condition.Reason == v1.PodReasonInfeasible {
			return podcommon.NewResizeState(podcommon.StateResizeInfeasible, condition.Message), nil
		}

		return podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			fmt.Errorf("unknown pod resize pending condition reason '%s'", condition.Reason)
	}

	if condition.Type == v1.PodResizeInProgress {
		if condition.Reason == v1.PodReasonError {
			// TODO(wt-later) matching messages is brittle but no other way of detecting this condition. Ref:
			// 	Ref: https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/kuberuntime/kuberuntime_manager.go
			if missingMemRegex.MatchString(condition.Message) {
				// kubelet is awaiting memory utilization for downsizing - treat as in progress with message.
				return podcommon.NewResizeState(
					podcommon.StateResizeInProgress,
					"kubelet is awaiting memory utilization for downsizing",
				), nil
			}

			// An error has occurred when resizing.
			return podcommon.NewResizeState(podcommon.StateResizeError, condition.Message), nil
		}

		// The resize is in progress.
		return podcommon.NewResizeState(podcommon.StateResizeInProgress, ""), nil
	}

	return podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
		fmt.Errorf("unexpected pod resize conditions (%#v)", resizeConditions)
}

// shouldReturnError returns whether to return an error after examining the type of the supplied err. Certain errors
// should not propagate since they are likely transient in nature i.e. 'resolved' in future reconciles.
func (s targetContainerState) shouldReturnError(ctx context.Context, err error) bool {
	if errors.As(err, &kube.ContainerStatusNotPresentError{}) {
		logging.Infof(ctx, logging.VDebug, "container status not yet present")
		return false
	}

	if errors.As(err, &kube.ContainerStatusResourcesNotPresentError{}) {
		logging.Infof(ctx, logging.VDebug, "container status resources not yet present")
		return false
	}

	return true
}
