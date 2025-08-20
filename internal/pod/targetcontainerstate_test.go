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
	"errors"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contexttest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scaletest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewTargetContainerState(t *testing.T) {
	pHelper := kube.NewPodHelper(nil)
	cHelper := kube.NewContainerHelper()
	state := newTargetContainerState(pHelper, cHelper)
	expected := targetContainerState{
		podHelper:       pHelper,
		containerHelper: cHelper,
	}
	assert.Equal(t, expected, state)
}

func TestTargetContainerStateStates(t *testing.T) {
	tests := []struct {
		name                  string
		targetContainer       *v1.Container
		configPHelperMockFunc func(*kubetest.MockPodHelper)
		configCHelperMockFunc func(*kubetest.MockContainerHelper)
		wantErrMsg            string
		wantStates            podcommon.States
		wantStateResources    podcommon.StateResources
	}{
		{
			"UnableToDetermineContainerState",
			&v1.Container{},
			nil,
			func(m *kubetest.MockContainerHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{}, errors.New(""))
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
			},
			"unable to determine container state",
			podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerUnknown,
				podcommon.StateBoolUnknown,
				podcommon.StateBoolUnknown,
				podcommon.StateResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
			podcommon.StateResources(""),
		},
		{
			"UnableToDetermineStartedState",
			&v1.Container{},
			nil,
			func(m *kubetest.MockContainerHelper) {
				m.On("IsStarted", mock.Anything, mock.Anything).Return(false, errors.New(""))
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
			},
			"unable to determine started state",
			podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolUnknown,
				podcommon.StateBoolUnknown,
				podcommon.StateResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
			podcommon.StateResources(""),
		},
		{
			"UnableToDetermineReadyState",
			&v1.Container{},
			nil,
			func(m *kubetest.MockContainerHelper) {
				m.On("IsReady", mock.Anything, mock.Anything).Return(false, errors.New(""))
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
			},
			"unable to determine ready state",
			podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolTrue,
				podcommon.StateBoolUnknown,
				podcommon.StateResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
			podcommon.StateResources(""),
		},
		{
			"UnableToDetermineStatusResourcesStates",
			kubetest.NewContainerBuilder().Build(),
			nil,
			func(m *kubetest.MockContainerHelper) {
				m.On("CurrentRequests", mock.Anything, mock.Anything, mock.Anything).
					Return(resource.Quantity{}, errors.New(""))
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
				m.IsReadyDefault()
				applyMockRequestsLimitsStartup(m)
			},
			"unable to determine status resources states",
			podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateResourcesStartup,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
			podcommon.StateResources(""),
		},
		{
			"UnableToDetermineResizeState",
			kubetest.NewContainerBuilder().Build(),
			func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsUnknownConditions)
			},
			nil,
			"unable to determine resize state",
			podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateResourcesStartup,
				podcommon.StateStatusResourcesContainerResourcesMatch,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
			podcommon.StateResources(""),
		},
		{
			"StateNotShouldReturnError",
			&v1.Container{},
			nil,
			func(m *kubetest.MockContainerHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{}, kube.NewContainerStatusNotPresentError())
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
			},
			"",
			podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerUnknown,
				podcommon.StateBoolUnknown,
				podcommon.StateBoolUnknown,
				podcommon.StateResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
			podcommon.StateResources(""),
		},
		{
			"IsStartedNotShouldReturnError",
			&v1.Container{},
			nil,
			func(m *kubetest.MockContainerHelper) {
				m.On("IsStarted", mock.Anything, mock.Anything).Return(false, kube.NewContainerStatusNotPresentError())
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
			},
			"",
			podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolUnknown,
				podcommon.StateBoolUnknown,
				podcommon.StateResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
			podcommon.StateResources(""),
		},
		{
			"IsReadyNotShouldReturnError",
			&v1.Container{},
			nil,
			func(m *kubetest.MockContainerHelper) {
				m.On("IsReady", mock.Anything, mock.Anything).Return(false, kube.NewContainerStatusNotPresentError())
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
			},
			"",
			podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolTrue,
				podcommon.StateBoolUnknown,
				podcommon.StateResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
			podcommon.StateResources(""),
		},
		{
			"StateStatusResourcesNotShouldReturnError",
			kubetest.NewContainerBuilder().Build(),
			nil,
			func(m *kubetest.MockContainerHelper) {
				m.On("CurrentRequests", mock.Anything, mock.Anything, mock.Anything).
					Return(resource.Quantity{}, kube.NewContainerStatusNotPresentError())
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
				m.IsReadyDefault()
				applyMockRequestsLimitsStartup(m)
			},
			"",
			podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateResourcesStartup,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
			podcommon.StateResources(""),
		},
		{
			string(podcommon.StateResourcesStartup),
			kubetest.NewContainerBuilder().Build(),
			nil,
			func(m *kubetest.MockContainerHelper) {
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
				m.IsReadyDefault()
				applyMockRequestsLimitsStartup(m)
				m.CurrentRequestsDefault()
				m.CurrentLimitsDefault()
			},
			"",
			podcommon.States{},
			podcommon.StateResourcesStartup,
		},
		{
			string(podcommon.StateResourcesPostStartup),
			kubetest.NewContainerBuilder().Build(),
			nil,
			func(m *kubetest.MockContainerHelper) {
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
				m.IsReadyDefault()
				applyMockRequestsLimitsPostStartup(m)
				m.CurrentRequestsDefault()
				m.CurrentLimitsDefault()
			},
			"",
			podcommon.States{},
			podcommon.StateResourcesPostStartup,
		},
		{
			string(podcommon.StateResourcesUnknown),
			kubetest.NewContainerBuilder().Build(),
			nil,
			func(m *kubetest.MockContainerHelper) {
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
				m.IsReadyDefault()
				applyMockRequestsLimitsUnknown(m)
				m.CurrentRequestsDefault()
				m.CurrentLimitsDefault()
			},
			"",
			podcommon.States{},
			podcommon.StateResourcesUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(
				kubetest.NewMockPodHelper(tt.configPHelperMockFunc),
				kubetest.NewMockContainerHelper(tt.configCHelperMockFunc),
			)

			cpuConfig := scaletest.NewMockConfiguration(func(m *scaletest.MockConfiguration) {
				m.On("Resources").Return(scalecommon.Resources{
					Startup:             kubetest.PodCpuStartupEnabled,
					PostStartupRequests: kubetest.PodCpuPostStartupRequestsEnabled,
					PostStartupLimits:   kubetest.PodCpuPostStartupLimitsEnabled,
				})
				m.IsEnabledDefault()
			})

			memoryConfig := scaletest.NewMockConfiguration(func(m *scaletest.MockConfiguration) {
				m.On("Resources").Return(scalecommon.Resources{
					Startup:             kubetest.PodMemoryStartupEnabled,
					PostStartupRequests: kubetest.PodMemoryPostStartupRequestsEnabled,
					PostStartupLimits:   kubetest.PodMemoryPostStartupLimitsEnabled,
				})
				m.IsEnabledDefault()
			})

			configs := scaletest.NewMockConfigurations(func(m *scaletest.MockConfigurations) {
				m.On("ConfigurationFor", v1.ResourceCPU).Return(cpuConfig)
				m.On("ConfigurationFor", v1.ResourceMemory).Return(memoryConfig)
			})

			got, err := s.States(
				contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				&v1.Pod{},
				tt.targetContainer,
				configs,
			)
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantStates != (podcommon.States{}) {
				assert.Equal(t, tt.wantStates, got)
			}
			if tt.wantStateResources != "" {
				assert.Equal(t, tt.wantStateResources, got.Resources)
			}
		})
	}
}

func TestTargetContainerStateStateStartupProbe(t *testing.T) {
	tests := []struct {
		name           string
		configMockFunc func(*kubetest.MockContainerHelper)
		want           podcommon.StateBool
	}{
		{
			"StateBoolTrue",
			func(m *kubetest.MockContainerHelper) {
				m.On("HasStartupProbe", mock.Anything, mock.Anything).Return(true, nil)
			},
			podcommon.StateBoolTrue,
		},
		{
			"StateBoolFalse",
			func(m *kubetest.MockContainerHelper) {
				m.On("HasStartupProbe", mock.Anything, mock.Anything).Return(false, nil)
			},
			podcommon.StateBoolFalse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(nil, kubetest.NewMockContainerHelper(tt.configMockFunc))
			assert.Equal(t, tt.want, s.stateStartupProbe(&v1.Container{}))
		})
	}
}

func TestTargetContainerStateStateReadinessProbe(t *testing.T) {
	tests := []struct {
		name           string
		configMockFunc func(*kubetest.MockContainerHelper)
		want           podcommon.StateBool
	}{
		{
			"StateBoolTrue",
			func(m *kubetest.MockContainerHelper) {
				m.On("HasReadinessProbe", mock.Anything, mock.Anything).Return(true, nil)
			},
			podcommon.StateBoolTrue,
		},
		{
			"StateBoolFalse",
			func(m *kubetest.MockContainerHelper) {
				m.On("HasReadinessProbe", mock.Anything, mock.Anything).Return(false, nil)
			},
			podcommon.StateBoolFalse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(nil, kubetest.NewMockContainerHelper(tt.configMockFunc))
			assert.Equal(t, tt.want, s.stateReadinessProbe(&v1.Container{}))
		})
	}
}

func TestTargetContainerStateStateContainer(t *testing.T) {
	tests := []struct {
		name           string
		configMockFunc func(*kubetest.MockContainerHelper)
		wantErrMsg     string
		want           podcommon.StateContainer
	}{
		{
			"UnableToGetContainerState",
			func(m *kubetest.MockContainerHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{}, errors.New(""))
			},
			"unable to get container state",
			podcommon.StateContainerUnknown,
		},
		{
			"StateContainerRunning",
			func(m *kubetest.MockContainerHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{
					Running: &v1.ContainerStateRunning{},
				}, nil)
			},
			"",
			podcommon.StateContainerRunning,
		},
		{
			"StateContainerWaiting",
			func(m *kubetest.MockContainerHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{},
				}, nil)
			},
			"",
			podcommon.StateContainerWaiting,
		},
		{
			"StateContainerTerminated",
			func(m *kubetest.MockContainerHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{},
				}, nil)
			},
			"",
			podcommon.StateContainerTerminated,
		},
		{
			"StateContainerUnknown",
			func(m *kubetest.MockContainerHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{}, nil)
			},
			"",
			podcommon.StateContainerUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(nil, kubetest.NewMockContainerHelper(tt.configMockFunc))

			got, err := s.stateContainer(&v1.Pod{}, &v1.Container{})
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTargetContainerStateStateStarted(t *testing.T) {
	tests := []struct {
		name           string
		configMockFunc func(*kubetest.MockContainerHelper)
		wantErrMsg     string
		want           podcommon.StateBool
	}{
		{
			"UnableToGetContainerReadyStatus",
			func(m *kubetest.MockContainerHelper) {
				m.On("IsStarted", mock.Anything, mock.Anything).Return(false, errors.New(""))
			},
			"unable to get container ready status",
			podcommon.StateBoolUnknown,
		},
		{
			"StateBoolTrue",
			func(m *kubetest.MockContainerHelper) {
				m.On("IsStarted", mock.Anything, mock.Anything).Return(true, nil)
			},
			"",
			podcommon.StateBoolTrue,
		},
		{
			"StateBoolFalse",
			func(m *kubetest.MockContainerHelper) {
				m.On("IsStarted", mock.Anything, mock.Anything).Return(false, nil)
			},
			"",
			podcommon.StateBoolFalse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(nil, kubetest.NewMockContainerHelper(tt.configMockFunc))

			got, err := s.stateStarted(&v1.Pod{}, &v1.Container{})
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTargetContainerStateStateReady(t *testing.T) {
	tests := []struct {
		name           string
		configMockFunc func(*kubetest.MockContainerHelper)
		wantErrMsg     string
		want           podcommon.StateBool
	}{
		{
			"UnableToGetContainerReadyStatus",
			func(m *kubetest.MockContainerHelper) {
				m.On("IsReady", mock.Anything, mock.Anything).Return(false, errors.New(""))
			},
			"unable to get container ready status",
			podcommon.StateBoolUnknown,
		},
		{
			"StateBoolTrue",
			func(m *kubetest.MockContainerHelper) {
				m.On("IsReady", mock.Anything, mock.Anything).Return(true, nil)
			},
			"",
			podcommon.StateBoolTrue,
		},
		{
			"StateBoolFalse",
			func(m *kubetest.MockContainerHelper) {
				m.On("IsReady", mock.Anything, mock.Anything).Return(false, nil)
			},
			"",
			podcommon.StateBoolFalse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(nil, kubetest.NewMockContainerHelper(tt.configMockFunc))

			got, err := s.stateReady(&v1.Pod{}, &v1.Container{})
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTargetContainerStateStateStatusResources(t *testing.T) {
	tests := []struct {
		name           string
		configMockFunc func(states *scaletest.MockStates)
		wantErrMsg     string
		want           podcommon.StateStatusResources
	}{
		{
			"UnableToDetermineIfAnyCurrentResourcesAreZero",
			func(m *scaletest.MockStates) {
				m.On("IsAnyCurrentZeroAll", mock.Anything, mock.Anything).Return(false, errors.New(""))
			},
			"unable to determine if any current resources are zero",
			podcommon.StateStatusResourcesUnknown,
		},
		{
			"UnableToDetermineIfCurrentRequestsMatchesSpec",
			func(m *scaletest.MockStates) {
				m.On("DoesRequestsCurrentMatchSpecAll", mock.Anything, mock.Anything).Return(false, errors.New(""))
				m.IsAnyCurrentZeroAllDefault()
			},
			"unable to determine if current requests matches spec",
			podcommon.StateStatusResourcesUnknown,
		},
		{
			"UnableToDetermineIfCurrentLimitsMatchesSpec",
			func(m *scaletest.MockStates) {
				m.On("DoesLimitsCurrentMatchSpecAll", mock.Anything, mock.Anything).Return(false, errors.New(""))
				m.IsAnyCurrentZeroAllDefault()
				m.DoesRequestsCurrentMatchSpecAllDefault()
			},
			"unable to determine if current limits matches spec",
			podcommon.StateStatusResourcesUnknown,
		},
		{
			string(podcommon.StateStatusResourcesIncomplete),
			func(m *scaletest.MockStates) {
				m.On("IsAnyCurrentZeroAll", mock.Anything, mock.Anything).Return(true, nil)
			},
			"",
			podcommon.StateStatusResourcesIncomplete,
		},
		{
			string(podcommon.StateStatusResourcesContainerResourcesMatch),
			func(m *scaletest.MockStates) {
				m.On("DoesRequestsCurrentMatchSpecAll", mock.Anything, mock.Anything).Return(true, nil)
				m.On("DoesLimitsCurrentMatchSpecAll", mock.Anything, mock.Anything).Return(true, nil)
				m.IsAnyCurrentZeroAllDefault()
			},
			"",
			podcommon.StateStatusResourcesContainerResourcesMatch,
		},
		{
			string(podcommon.StateStatusResourcesContainerResourcesMismatch),
			func(m *scaletest.MockStates) {
				m.On("DoesRequestsCurrentMatchSpecAll", mock.Anything, mock.Anything).Return(true, nil)
				m.On("DoesLimitsCurrentMatchSpecAll", mock.Anything, mock.Anything).Return(false, nil)
				m.IsAnyCurrentZeroAllDefault()
			},
			"",
			podcommon.StateStatusResourcesContainerResourcesMismatch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(nil, kubetest.NewMockContainerHelper(nil))

			got, err := s.stateStatusResources(&v1.Pod{}, &v1.Container{}, scaletest.NewMockStates(tt.configMockFunc))
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}

}

func TestTargetContainerStateStateResize(t *testing.T) {
	tests := []struct {
		name           string
		configMockFunc func(*kubetest.MockPodHelper)
		wantErrMsg     string
		want           podcommon.ResizeState
	}{
		{
			"UnknownPodResizePendingConditionReason",
			func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsUnknownPending)

			},
			"unknown pod resize pending condition reason",
			podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
		},
		{
			"UnexpectedPodResizeConditions",
			func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsUnknownConditions)

			},
			"unexpected pod resize conditions",
			podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
		},
		{
			"StateResizeNotStartedOrCompletedNoConditions",
			func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsNotStartedOrCompletedNoConditions)

			},
			"",
			podcommon.NewResizeState(podcommon.StateResizeNotStartedOrCompleted, ""),
		},
		{
			"StateResizeDeferred",
			func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsDeferred)

			},
			"",
			podcommon.NewResizeState(podcommon.StateResizeDeferred, "message"),
		},
		{
			"StateResizeInfeasible",
			func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsInfeasible)

			},
			"",
			podcommon.NewResizeState(podcommon.StateResizeInfeasible, "message"),
		},
		{
			"StateResizeErrorMissingMem1",
			func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsErrorInProgress1)

			},
			"",
			podcommon.NewResizeState(podcommon.StateResizeInProgress, "kubelet is awaiting memory utilization for downsizing"),
		},
		{
			"StateResizeErrorMissingMem2",
			func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsErrorInProgress2)

			},
			"",
			podcommon.NewResizeState(podcommon.StateResizeInProgress, "kubelet is awaiting memory utilization for downsizing"),
		},
		{
			"StateResizeError",
			func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsError)

			},
			"",
			podcommon.NewResizeState(podcommon.StateResizeError, "message"),
		},
		{
			"StateResizeInProgress",
			func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsInProgress)

			},
			"",
			podcommon.NewResizeState(podcommon.StateResizeInProgress, ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(kubetest.NewMockPodHelper(tt.configMockFunc), nil)

			got, err := s.stateResize(&v1.Pod{})
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}

}

func TestTargetContainerStateShouldReturnError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"ContainerStatusResourcesNotPresentErrorFalse", kube.NewContainerStatusResourcesNotPresentError(), false},
		{"NewContainerStatusResourcesNotPresentErrorFalse", kube.NewContainerStatusResourcesNotPresentError(), false},
		{"True", errors.New(""), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(nil, nil)
			assert.Equal(
				t,
				tt.want,
				s.shouldReturnError(contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(), tt.err),
			)
		})
	}
}

func applyMockRequestsLimitsStartup(m *kubetest.MockContainerHelper) {
	m.On("Requests", mock.Anything, v1.ResourceCPU).Return(kubetest.PodCpuStartupEnabled)
	m.On("Requests", mock.Anything, v1.ResourceMemory).Return(kubetest.PodMemoryStartupEnabled)
	m.On("Limits", mock.Anything, v1.ResourceCPU).Return(kubetest.PodCpuStartupEnabled)
	m.On("Limits", mock.Anything, v1.ResourceMemory).Return(kubetest.PodMemoryStartupEnabled)
}

func applyMockRequestsLimitsPostStartup(m *kubetest.MockContainerHelper) {
	m.On("Requests", mock.Anything, v1.ResourceCPU).Return(kubetest.PodCpuPostStartupRequestsEnabled)
	m.On("Requests", mock.Anything, v1.ResourceMemory).Return(kubetest.PodMemoryPostStartupRequestsEnabled)
	m.On("Limits", mock.Anything, v1.ResourceCPU).Return(kubetest.PodCpuPostStartupLimitsEnabled)
	m.On("Limits", mock.Anything, v1.ResourceMemory).Return(kubetest.PodMemoryPostStartupLimitsEnabled)
}

func applyMockRequestsLimitsUnknown(m *kubetest.MockContainerHelper) {
	m.On("Requests", mock.Anything, v1.ResourceCPU).Return(kubetest.PodCpuUnknown)
	m.On("Requests", mock.Anything, v1.ResourceMemory).Return(kubetest.PodMemoryUnknown)
	m.On("Limits", mock.Anything, v1.ResourceCPU).Return(kubetest.PodCpuUnknown)
	m.On("Limits", mock.Anything, v1.ResourceMemory).Return(kubetest.PodMemoryUnknown)
}
