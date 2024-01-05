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

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
)

// TargetContainerState performs operations relating to determining target container state.
type TargetContainerState interface {
	States(context.Context, *v1.Pod, podcommon.ScaleConfig) (podcommon.States, error)
}

// targetContainerState is the default implementation of TargetContainerState.
type targetContainerState struct {
	containerKubeHelper ContainerKubeHelper
}

func newTargetContainerState(containerKubeHelper ContainerKubeHelper) targetContainerState {
	return targetContainerState{containerKubeHelper: containerKubeHelper}
}

// States calculates and returns states from the supplied pod and config.
func (s targetContainerState) States(ctx context.Context, pod *v1.Pod, config podcommon.ScaleConfig) (podcommon.States, error) {
	container, err := s.containerKubeHelper.Get(pod, config.GetTargetContainerName())
	if err != nil {
		return podcommon.NewStatesAllUnknown(),
			errors.Wrap(err, "unable to get container")
	}

	ret := podcommon.NewStatesAllUnknown()
	ret.StartupProbe = s.stateStartupProbe(container)
	ret.ReadinessProbe = s.stateReadinessProbe(container)

	ret.Container, err = s.stateContainer(pod, config)
	if err != nil {
		if !s.shouldReturnError(ctx, err) {
			return ret, nil
		}
		return ret, errors.Wrap(err, "unable to determine container state")
	}

	ret.Started, err = s.stateStarted(pod, config)
	if err != nil {
		if !s.shouldReturnError(ctx, err) {
			return ret, nil
		}
		return ret, errors.Wrap(err, "unable to determine started state")
	}

	ret.Ready, err = s.stateReady(pod, config)
	if err != nil {
		if !s.shouldReturnError(ctx, err) {
			return ret, nil
		}
		return ret, errors.Wrap(err, "unable to determine ready state")
	}

	ret.Resources = s.stateResources(
		s.isStartupConfigApplied(container, config),
		s.isPostStartupConfigApplied(container, config),
	)

	ret.AllocatedResources, err = s.stateAllocatedResources(pod, container, config)
	if err != nil {
		if !s.shouldReturnError(ctx, err) {
			return ret, nil
		}
		return ret, errors.Wrap(err, "unable to determine allocated resources states")
	}

	ret.StatusResources, err = s.stateStatusResources(pod, container, config)
	if err != nil {
		if !s.shouldReturnError(ctx, err) {
			return ret, nil
		}
		return ret, errors.Wrap(err, "unable to determine status resources states")
	}

	return ret, nil
}

// stateStartupProbe returns the startup probe state for the target container.
func (s targetContainerState) stateStartupProbe(container *v1.Container) podcommon.StateBool {
	if s.containerKubeHelper.HasStartupProbe(container) {
		return podcommon.StateBoolTrue
	}

	return podcommon.StateBoolFalse
}

// stateReadinessProbe returns the readiness probe state for the target container.
func (s targetContainerState) stateReadinessProbe(container *v1.Container) podcommon.StateBool {
	if s.containerKubeHelper.HasReadinessProbe(container) {
		return podcommon.StateBoolTrue
	}

	return podcommon.StateBoolFalse
}

// stateContainer returns the container state for the target container, using the supplied config.
func (s targetContainerState) stateContainer(pod *v1.Pod, config podcommon.ScaleConfig) (podcommon.StateContainer, error) {
	state, err := s.containerKubeHelper.State(pod, config.GetTargetContainerName())
	if err != nil {
		return podcommon.StateContainerUnknown, errors.Wrap(err, "unable to get container state")
	}

	if state.Running != nil {
		return podcommon.StateContainerRunning, nil
	}

	if state.Waiting != nil {
		return podcommon.StateContainerWaiting, nil
	}

	if state.Terminated != nil {
		return podcommon.StateContainerTerminated, nil
	}

	return podcommon.StateContainerUnknown, nil
}

// stateStarted returns the ready state for the target container, using the supplied config.
func (s targetContainerState) stateStarted(pod *v1.Pod, config podcommon.ScaleConfig) (podcommon.StateBool, error) {
	started, err := s.containerKubeHelper.IsStarted(pod, config.GetTargetContainerName())
	if err != nil {
		return podcommon.StateBoolUnknown, errors.Wrap(err, "unable to get container ready status")
	}

	if started {
		return podcommon.StateBoolTrue, nil
	}

	return podcommon.StateBoolFalse, nil
}

// stateReady returns the ready state for the target container, using the supplied config.
func (s targetContainerState) stateReady(pod *v1.Pod, config podcommon.ScaleConfig) (podcommon.StateBool, error) {
	ready, err := s.containerKubeHelper.IsReady(pod, config.GetTargetContainerName())
	if err != nil {
		return podcommon.StateBoolUnknown, errors.Wrap(err, "unable to get container ready status")
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

// stateAllocatedResources returns the allocated resources state for the target container, using the supplied config.
func (s targetContainerState) stateAllocatedResources(
	pod *v1.Pod,
	container *v1.Container,
	config podcommon.ScaleConfig,
) (podcommon.StateAllocatedResources, error) {
	allocatedCpu, err := s.containerKubeHelper.AllocatedResources(pod, config.GetTargetContainerName(), v1.ResourceCPU)
	if err != nil {
		return podcommon.StateAllocatedResourcesUnknown, errors.Wrap(err, "unable to get allocated cpu resources")
	}

	allocatedMemory, err := s.containerKubeHelper.AllocatedResources(pod, config.GetTargetContainerName(), v1.ResourceMemory)
	if err != nil {
		return podcommon.StateAllocatedResourcesUnknown, errors.Wrap(err, "unable to get allocated memory resources")
	}

	if allocatedCpu.IsZero() || allocatedMemory.IsZero() {
		return podcommon.StateAllocatedResourcesIncomplete, nil
	}

	requestsCpu := s.containerKubeHelper.Requests(container, v1.ResourceCPU)
	requestsMemory := s.containerKubeHelper.Requests(container, v1.ResourceMemory)

	if allocatedCpu.Equal(requestsCpu) && allocatedMemory.Equal(requestsMemory) {
		return podcommon.StateAllocatedResourcesContainerRequestsMatch, nil
	}

	return podcommon.StateAllocatedResourcesContainerRequestsMismatch, nil
}

// stateAllocatedResources returns the status resources state for the target container, using the supplied config.
func (s targetContainerState) stateStatusResources(
	pod *v1.Pod,
	container *v1.Container,
	config podcommon.ScaleConfig,
) (podcommon.StateStatusResources, error) {
	currentRequestsCpu, err := s.containerKubeHelper.CurrentRequests(pod, config.GetTargetContainerName(), v1.ResourceCPU)
	if err != nil {
		return podcommon.StateStatusResourcesUnknown, errors.Wrap(err, "unable to get status resources cpu requests")
	}

	currentLimitsCpu, err := s.containerKubeHelper.CurrentLimits(pod, config.GetTargetContainerName(), v1.ResourceCPU)
	if err != nil {
		return podcommon.StateStatusResourcesUnknown, errors.Wrap(err, "unable to get status resources cpu limits")
	}

	currentRequestsMemory, err := s.containerKubeHelper.CurrentRequests(pod, config.GetTargetContainerName(), v1.ResourceMemory)
	if err != nil {
		return podcommon.StateStatusResourcesUnknown, errors.Wrap(err, "unable to get status resources memory requests")
	}

	currentLimitsMemory, err := s.containerKubeHelper.CurrentLimits(pod, config.GetTargetContainerName(), v1.ResourceMemory)
	if err != nil {
		return podcommon.StateStatusResourcesUnknown, errors.Wrap(err, "unable to get status resources memory limits")
	}

	if currentRequestsCpu.IsZero() || currentLimitsCpu.IsZero() || currentRequestsMemory.IsZero() || currentLimitsMemory.IsZero() {
		return podcommon.StateStatusResourcesIncomplete, nil
	}

	requestsCpu := s.containerKubeHelper.Requests(container, v1.ResourceCPU)
	limitsCpu := s.containerKubeHelper.Limits(container, v1.ResourceCPU)
	requestsMemory := s.containerKubeHelper.Requests(container, v1.ResourceMemory)
	limitsMemory := s.containerKubeHelper.Limits(container, v1.ResourceMemory)

	if currentRequestsCpu.Equal(requestsCpu) &&
		currentLimitsCpu.Equal(limitsCpu) &&
		currentRequestsMemory.Equal(requestsMemory) &&
		currentLimitsMemory.Equal(limitsMemory) {
		return podcommon.StateStatusResourcesContainerResourcesMatch, nil
	}

	return podcommon.StateStatusResourcesContainerResourcesMismatch, nil
}

// isStartupConfigApplied reports whether the supplied container has its startup configuration applied, using the
// supplied config.
func (s targetContainerState) isStartupConfigApplied(container *v1.Container, config podcommon.ScaleConfig) bool {
	cpuStartupRequestsApplied := s.containerKubeHelper.Requests(container, v1.ResourceCPU).Equal(config.GetCpuConfig().Startup)
	cpuStartupLimitsApplied := s.containerKubeHelper.Limits(container, v1.ResourceCPU).Equal(config.GetCpuConfig().Startup)
	cpuStartupApplied := cpuStartupRequestsApplied && cpuStartupLimitsApplied

	memoryRequestsStartupApplied := s.containerKubeHelper.Requests(container, v1.ResourceMemory).Equal(config.GetMemoryConfig().Startup)
	memoryLimitsStartupApplied := s.containerKubeHelper.Limits(container, v1.ResourceMemory).Equal(config.GetMemoryConfig().Startup)
	memoryStartupApplied := memoryRequestsStartupApplied && memoryLimitsStartupApplied

	return cpuStartupApplied && memoryStartupApplied
}

// isPostStartupConfigApplied reports whether the supplied container has its post-startup configuration applied, using
// the supplied config.
func (s targetContainerState) isPostStartupConfigApplied(container *v1.Container, config podcommon.ScaleConfig) bool {
	cpuPostStartupRequestsApplied := s.containerKubeHelper.Requests(container, v1.ResourceCPU).Equal(config.GetCpuConfig().PostStartupRequests)
	cpuPostStartupLimitsApplied := s.containerKubeHelper.Limits(container, v1.ResourceCPU).Equal(config.GetCpuConfig().PostStartupLimits)
	cpuPostStartupApplied := cpuPostStartupRequestsApplied && cpuPostStartupLimitsApplied

	memoryPostStartupRequestsApplied := s.containerKubeHelper.Requests(container, v1.ResourceMemory).Equal(config.GetMemoryConfig().PostStartupRequests)
	memoryPostStartupLimitsApplied := s.containerKubeHelper.Limits(container, v1.ResourceMemory).Equal(config.GetMemoryConfig().PostStartupLimits)
	memoryPostStartupApplied := memoryPostStartupRequestsApplied && memoryPostStartupLimitsApplied

	return cpuPostStartupApplied && memoryPostStartupApplied
}

// shouldReturnError reports whether to return an error after examining the type of the supplied err. Certain errors
// should not propagate since they are likely transient in nature i.e. 'resolved' in future reconciles.
func (s targetContainerState) shouldReturnError(ctx context.Context, err error) bool {
	if errors.Is(err, ContainerStatusNotPresentError{}) {
		logging.Infof(ctx, logging.VDebug, "container status not yet present")
		return false
	}

	if errors.Is(err, ContainerStatusAllocatedResourcesNotPresentError{}) {
		logging.Infof(ctx, logging.VDebug, "container status allocated resources not yet present")
		return false
	}

	if errors.Is(err, ContainerStatusResourcesNotPresentError{}) {
		logging.Infof(ctx, logging.VDebug, "container status resources not yet present")
		return false
	}

	return true
}
