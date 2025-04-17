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
	s := newTargetContainerState(pHelper, cHelper)
	assert.Equal(t, pHelper, s.podHelper)
	assert.Equal(t, cHelper, s.containerHelper)
}

func TestTargetContainerStateStates(t *testing.T) {
	tests := []struct {
		name                  string
		targetContainer       *v1.Container
		configPHelperMockFunc func(*kubetest.MockPodHelper)
		configCHelperMockFunc func(*kubetest.MockContainerHelper)
		wantStates            podcommon.States
		wantStateResources    podcommon.StateResources
		wantErrMsg            string
	}{
		{
			name:            "UnableToDetermineContainerState",
			targetContainer: &v1.Container{},
			configCHelperMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{}, errors.New(""))
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
			},
			wantStates: podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerUnknown,
				podcommon.StateBoolUnknown,
				podcommon.StateBoolUnknown,
				podcommon.StateResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
			wantErrMsg: "unable to determine container state",
		},
		{
			name:            "UnableToDetermineStartedState",
			targetContainer: &v1.Container{},
			configCHelperMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("IsStarted", mock.Anything, mock.Anything).Return(false, errors.New(""))
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
			},
			wantStates: podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolUnknown,
				podcommon.StateBoolUnknown,
				podcommon.StateResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
			wantErrMsg: "unable to determine started state",
		},
		{
			name:            "UnableToDetermineReadyState",
			targetContainer: &v1.Container{},
			configCHelperMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("IsReady", mock.Anything, mock.Anything).Return(false, errors.New(""))
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
			},
			wantStates: podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolTrue,
				podcommon.StateBoolUnknown,
				podcommon.StateResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
			wantErrMsg: "unable to determine ready state",
		},
		{
			name:            "UnableToDetermineStatusResourcesStates",
			targetContainer: kubetest.NewContainerBuilder().Build(),
			configCHelperMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("CurrentRequests", mock.Anything, mock.Anything, mock.Anything).
					Return(resource.Quantity{}, errors.New(""))
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
				m.IsReadyDefault()
				applyMockRequestsLimitsStartup(m)
			},
			wantStates: podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateResourcesStartup,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
			wantErrMsg: "unable to determine status resources states",
		},
		{
			name:            "UnableToDetermineResizeState",
			targetContainer: kubetest.NewContainerBuilder().Build(),
			configPHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsUnknownConditions)
			},
			wantStates: podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateResourcesStartup,
				podcommon.StateStatusResourcesContainerResourcesMatch,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
			wantErrMsg: "unable to determine resize state",
		},
		{
			name:            "StateNotShouldReturnError",
			targetContainer: &v1.Container{},
			configCHelperMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{}, kube.NewContainerStatusNotPresentError())
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
			},
			wantStates: podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerUnknown,
				podcommon.StateBoolUnknown,
				podcommon.StateBoolUnknown,
				podcommon.StateResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
		},
		{
			name:            "IsStartedNotShouldReturnError",
			targetContainer: &v1.Container{},
			configCHelperMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("IsStarted", mock.Anything, mock.Anything).Return(false, kube.NewContainerStatusNotPresentError())
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
			},
			wantStates: podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolUnknown,
				podcommon.StateBoolUnknown,
				podcommon.StateResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
		},
		{
			name:            "IsReadyNotShouldReturnError",
			targetContainer: &v1.Container{},
			configCHelperMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("IsReady", mock.Anything, mock.Anything).Return(false, kube.NewContainerStatusNotPresentError())
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
			},
			wantStates: podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolTrue,
				podcommon.StateBoolUnknown,
				podcommon.StateResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
		},
		{
			name:            "StateStatusResourcesNotShouldReturnError",
			targetContainer: kubetest.NewContainerBuilder().Build(),
			configCHelperMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("CurrentRequests", mock.Anything, mock.Anything, mock.Anything).
					Return(resource.Quantity{}, kube.NewContainerStatusNotPresentError())
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
				m.IsReadyDefault()
				applyMockRequestsLimitsStartup(m)
			},
			wantStates: podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateResourcesStartup,
				podcommon.StateStatusResourcesUnknown,
				podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			),
		},
		{
			name:            string(podcommon.StateResourcesStartup),
			targetContainer: kubetest.NewContainerBuilder().Build(),
			configCHelperMockFunc: func(m *kubetest.MockContainerHelper) {
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
				m.IsReadyDefault()
				applyMockRequestsLimitsStartup(m)
				m.CurrentRequestsDefault()
				m.CurrentLimitsDefault()
			},
			wantStateResources: podcommon.StateResourcesStartup,
		},
		{
			name:            string(podcommon.StateResourcesPostStartup),
			targetContainer: kubetest.NewContainerBuilder().Build(),
			configCHelperMockFunc: func(m *kubetest.MockContainerHelper) {
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
				m.IsReadyDefault()
				applyMockRequestsLimitsPostStartup(m)
				m.CurrentRequestsDefault()
				m.CurrentLimitsDefault()
			},
			wantStateResources: podcommon.StateResourcesPostStartup,
		},
		{
			name:            string(podcommon.StateResourcesUnknown),
			targetContainer: kubetest.NewContainerBuilder().Build(),
			configCHelperMockFunc: func(m *kubetest.MockContainerHelper) {
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
				m.IsReadyDefault()
				applyMockRequestsLimitsUnknown(m)
				m.CurrentRequestsDefault()
				m.CurrentLimitsDefault()
			},
			wantStateResources: podcommon.StateResourcesUnknown,
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
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
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
			name: "StateBoolTrue",
			configMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("HasStartupProbe", mock.Anything, mock.Anything).Return(true, nil)
			},
			want: podcommon.StateBoolTrue,
		},
		{
			name: "StateBoolFalse",
			configMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("HasStartupProbe", mock.Anything, mock.Anything).Return(false, nil)
			},
			want: podcommon.StateBoolFalse,
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
			name: "StateBoolTrue",
			configMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("HasReadinessProbe", mock.Anything, mock.Anything).Return(true, nil)
			},
			want: podcommon.StateBoolTrue,
		},
		{
			name: "StateBoolFalse",
			configMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("HasReadinessProbe", mock.Anything, mock.Anything).Return(false, nil)
			},
			want: podcommon.StateBoolFalse,
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
		want           podcommon.StateContainer
		wantErrMsg     string
	}{
		{
			name: "UnableToGetContainerState",
			configMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{}, errors.New(""))
			},
			want:       podcommon.StateContainerUnknown,
			wantErrMsg: "unable to get container state",
		},
		{
			name: "StateContainerRunning",
			configMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{
					Running: &v1.ContainerStateRunning{},
				}, nil)
			},
			want: podcommon.StateContainerRunning,
		},
		{
			name: "StateContainerWaiting",
			configMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{},
				}, nil)
			},
			want: podcommon.StateContainerWaiting,
		},
		{
			name: "StateContainerTerminated",
			configMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{},
				}, nil)
			},
			want: podcommon.StateContainerTerminated,
		},
		{
			name: "StateContainerUnknown",
			configMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{}, nil)
			},
			want: podcommon.StateContainerUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(nil, kubetest.NewMockContainerHelper(tt.configMockFunc))

			got, err := s.stateContainer(&v1.Pod{}, &v1.Container{})
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTargetContainerStateStateStarted(t *testing.T) {
	tests := []struct {
		name           string
		configMockFunc func(*kubetest.MockContainerHelper)
		want           podcommon.StateBool
		wantErrMsg     string
	}{
		{
			name: "UnableToGetContainerReadyStatus",
			configMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("IsStarted", mock.Anything, mock.Anything).Return(false, errors.New(""))
			},
			want:       podcommon.StateBoolUnknown,
			wantErrMsg: "unable to get container ready status",
		},
		{
			name: "StateBoolTrue",
			configMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("IsStarted", mock.Anything, mock.Anything).Return(true, nil)
			},
			want: podcommon.StateBoolTrue,
		},
		{
			name: "StateBoolFalse",
			configMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("IsStarted", mock.Anything, mock.Anything).Return(false, nil)
			},
			want: podcommon.StateBoolFalse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(nil, kubetest.NewMockContainerHelper(tt.configMockFunc))

			got, err := s.stateStarted(&v1.Pod{}, &v1.Container{})
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTargetContainerStateStateReady(t *testing.T) {
	tests := []struct {
		name           string
		configMockFunc func(*kubetest.MockContainerHelper)
		want           podcommon.StateBool
		wantErrMsg     string
	}{
		{
			name: "UnableToGetContainerReadyStatus",
			configMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("IsReady", mock.Anything, mock.Anything).Return(false, errors.New(""))
			},
			want:       podcommon.StateBoolUnknown,
			wantErrMsg: "unable to get container ready status",
		},
		{
			name: "StateBoolTrue",
			configMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("IsReady", mock.Anything, mock.Anything).Return(true, nil)
			},
			want: podcommon.StateBoolTrue,
		},
		{
			name: "StateBoolFalse",
			configMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("IsReady", mock.Anything, mock.Anything).Return(false, nil)
			},
			want: podcommon.StateBoolFalse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(nil, kubetest.NewMockContainerHelper(tt.configMockFunc))

			got, err := s.stateReady(&v1.Pod{}, &v1.Container{})
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTargetContainerStateStateStatusResources(t *testing.T) {
	tests := []struct {
		name           string
		configMockFunc func(states *scaletest.MockStates)
		want           podcommon.StateStatusResources
		wantErrMsg     string
	}{
		{
			name: "UnableToDetermineIfAnyCurrentResourcesAreZero",
			configMockFunc: func(m *scaletest.MockStates) {
				m.On("IsAnyCurrentZeroAll", mock.Anything, mock.Anything).Return(false, errors.New(""))
			},
			want:       podcommon.StateStatusResourcesUnknown,
			wantErrMsg: "unable to determine if any current resources are zero",
		},
		{
			name: "UnableToDetermineIfCurrentRequestsMatchesSpec",
			configMockFunc: func(m *scaletest.MockStates) {
				m.On("DoesRequestsCurrentMatchSpecAll", mock.Anything, mock.Anything).Return(false, errors.New(""))
				m.IsAnyCurrentZeroAllDefault()
			},
			want:       podcommon.StateStatusResourcesUnknown,
			wantErrMsg: "unable to determine if current requests matches spec",
		},
		{
			name: "UnableToDetermineIfCurrentLimitsMatchesSpec",
			configMockFunc: func(m *scaletest.MockStates) {
				m.On("DoesLimitsCurrentMatchSpecAll", mock.Anything, mock.Anything).Return(false, errors.New(""))
				m.IsAnyCurrentZeroAllDefault()
				m.DoesRequestsCurrentMatchSpecAllDefault()
			},
			want:       podcommon.StateStatusResourcesUnknown,
			wantErrMsg: "unable to determine if current limits matches spec",
		},
		{
			name: string(podcommon.StateStatusResourcesIncomplete),
			configMockFunc: func(m *scaletest.MockStates) {
				m.On("IsAnyCurrentZeroAll", mock.Anything, mock.Anything).Return(true, nil)
			},
			want: podcommon.StateStatusResourcesIncomplete,
		},
		{
			name: string(podcommon.StateStatusResourcesContainerResourcesMatch),
			configMockFunc: func(m *scaletest.MockStates) {
				m.On("DoesRequestsCurrentMatchSpecAll", mock.Anything, mock.Anything).Return(true, nil)
				m.On("DoesLimitsCurrentMatchSpecAll", mock.Anything, mock.Anything).Return(true, nil)
				m.IsAnyCurrentZeroAllDefault()
			},
			want: podcommon.StateStatusResourcesContainerResourcesMatch,
		},
		{
			name: string(podcommon.StateStatusResourcesContainerResourcesMismatch),
			configMockFunc: func(m *scaletest.MockStates) {
				m.On("DoesRequestsCurrentMatchSpecAll", mock.Anything, mock.Anything).Return(true, nil)
				m.On("DoesLimitsCurrentMatchSpecAll", mock.Anything, mock.Anything).Return(false, nil)
				m.IsAnyCurrentZeroAllDefault()
			},
			want: podcommon.StateStatusResourcesContainerResourcesMismatch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(nil, kubetest.NewMockContainerHelper(nil))

			got, err := s.stateStatusResources(&v1.Pod{}, &v1.Container{}, scaletest.NewMockStates(tt.configMockFunc))
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}

}

func TestTargetContainerStateStateResize(t *testing.T) {
	tests := []struct {
		name           string
		configMockFunc func(*kubetest.MockPodHelper)
		want           podcommon.ResizeState
		wantErrMsg     string
	}{
		{
			name: "UnknownPodResizePendingConditionReason",
			configMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsUnknownPending)

			},
			want:       podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			wantErrMsg: "unknown pod resize pending condition reason",
		},
		{
			name: "UnknownPodResizeInProgressConditionState",
			configMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsUnknownInProgress)

			},
			want:       podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			wantErrMsg: "unknown pod resize in progress condition state",
		},
		{
			name: "UnexpectedPodResizeConditions",
			configMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsUnknownConditions)

			},
			want:       podcommon.NewResizeState(podcommon.StateResizeUnknown, ""),
			wantErrMsg: "unexpected pod resize conditions",
		},
		{
			name: "StateResizeNotStartedOrCompletedNoConditions",
			configMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsNotStartedOrCompletedNoConditions)

			},
			want: podcommon.NewResizeState(podcommon.StateResizeNotStartedOrCompleted, ""),
		},
		{
			name: "StateResizeDeferred",
			configMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsDeferred)

			},
			want: podcommon.NewResizeState(podcommon.StateResizeDeferred, "message"),
		},
		{
			name: "StateResizeInfeasible",
			configMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsInfeasible)

			},
			want: podcommon.NewResizeState(podcommon.StateResizeInfeasible, "message"),
		},
		{
			name: "StateResizeError",
			configMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsError)

			},
			want: podcommon.NewResizeState(podcommon.StateResizeError, "message"),
		},
		{
			name: "StateResizeInProgress",
			configMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsInProgress)

			},
			want: podcommon.NewResizeState(podcommon.StateResizeInProgress, ""),
		},
		{
			name: "StateResizeNotStartedOrCompletedResizeInProgressTrue",
			configMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeConditions", mock.Anything).Return(kubetest.PodResizeConditionsNotStartedOrCompletedResizeInProgressTrue)

			},
			want: podcommon.NewResizeState(podcommon.StateResizeNotStartedOrCompleted, ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(kubetest.NewMockPodHelper(tt.configMockFunc), nil)

			got, err := s.stateResize(&v1.Pod{})
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
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
