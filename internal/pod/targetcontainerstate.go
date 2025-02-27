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

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource/config"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource/state"
	"k8s.io/api/core/v1"
)

// TargetContainerState performs operations relating to determining target container state.
type TargetContainerState interface {
	States(context.Context, *v1.Pod, *v1.Container, config.ScaleConfigs) (podcommon.States, error)
}

// targetContainerState is the default implementation of TargetContainerState.
type targetContainerState struct {
	containerHelper kube.ContainerHelper
}

func newTargetContainerState(containerHelper kube.ContainerHelper) targetContainerState {
	return targetContainerState{containerHelper: containerHelper}
}

// States calculates and returns states from the supplied pod and config.
func (s targetContainerState) States(
	ctx context.Context,
	pod *v1.Pod,
	targetContainer *v1.Container,
	scaleConfigs config.ScaleConfigs,
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

	scaleStates := state.NewScaleStates(scaleConfigs, s.containerHelper)
	ret.Resources = s.stateResources(
		scaleStates.IsStartupConfigAppliedAll(targetContainer),
		scaleStates.IsPostStartupConfigAppliedAll(targetContainer),
	)

	ret.StatusResources, err = s.stateStatusResources(pod, targetContainer, scaleStates)
	if err != nil {
		if !s.shouldReturnError(ctx, err) {
			return ret, nil
		}
		return ret, common.WrapErrorf(err, "unable to determine status resources states")
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

// stateContainer returns the container state for the target container, using the supplied config.
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

// stateStarted returns the ready state for the target container, using the supplied config.
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

// stateReady returns the ready state for the target container, using the supplied config.
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

// stateStatusResources returns the status resources state for the target container, using the supplied config.
func (s targetContainerState) stateStatusResources(
	pod *v1.Pod,
	targetContainer *v1.Container,
	scaleStates state.ScaleStates,
) (podcommon.StateStatusResources, error) {
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

// shouldReturnError reports whether to return an error after examining the type of the supplied err. Certain errors
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
