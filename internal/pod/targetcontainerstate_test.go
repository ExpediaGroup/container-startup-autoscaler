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
	"errors"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contexttest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewTargetContainerState(t *testing.T) {
	cHelper := newContainerKubeHelper()
	s := newTargetContainerState(cHelper)
	assert.Equal(t, cHelper, s.containerKubeHelper)
}

func TestTargetContainerStateStates(t *testing.T) {
	tests := []struct {
		name               string
		configMockFunc     func(*podtest.MockContainerKubeHelper)
		wantStates         podcommon.States
		wantStateResources podcommon.StateResources
		wantErrMsg         string
	}{
		{
			name: "UnableToGetContainer",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("Get", mock.Anything, mock.Anything).Return(&v1.Container{}, errors.New(""))
			},
			wantStates: podcommon.NewStatesAllUnknown(),
			wantErrMsg: "unable to get container",
		},
		{
			name: "UnableToDetermineContainerState",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{}, errors.New(""))
				m.GetDefault()
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
				podcommon.StateAllocatedResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
			),
			wantErrMsg: "unable to determine container state",
		},
		{
			name: "UnableToDetermineStartedState",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("IsStarted", mock.Anything, mock.Anything).Return(false, errors.New(""))
				m.GetDefault()
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
				podcommon.StateAllocatedResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
			),
			wantErrMsg: "unable to determine started state",
		},

		{
			name: "UnableToDetermineReadyState",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("IsReady", mock.Anything, mock.Anything).Return(false, errors.New(""))
				m.GetDefault()
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
				podcommon.StateAllocatedResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
			),
			wantErrMsg: "unable to determine ready state",
		},
		{
			name: "UnableToGetAllocatedResourcesState",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("AllocatedResources", mock.Anything, mock.Anything, mock.Anything).
					Return(resource.Quantity{}, errors.New(""))
				m.GetDefault()
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
				podcommon.StateAllocatedResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
			),
			wantErrMsg: "unable to determine allocated resources states",
		},
		{
			name: "UnableToGetStatusResourcesState",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("CurrentRequests", mock.Anything, mock.Anything, mock.Anything).
					Return(resource.Quantity{}, errors.New(""))
				m.GetDefault()
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
				m.IsReadyDefault()
				applyMockRequestsLimitsStartup(m)
				m.AllocatedResourcesDefault()
			},
			wantStates: podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateResourcesStartup,
				podcommon.StateAllocatedResourcesContainerRequestsMismatch,
				podcommon.StateStatusResourcesUnknown,
			),
			wantErrMsg: "unable to determine status resources states",
		},
		{
			name: "StateNotShouldReturnError",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{}, NewContainerStatusNotPresentError())
				m.GetDefault()
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
				podcommon.StateAllocatedResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
			),
		},
		{
			name: "IsStartedNotShouldReturnError",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("IsStarted", mock.Anything, mock.Anything).Return(false, NewContainerStatusNotPresentError())
				m.GetDefault()
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
				podcommon.StateAllocatedResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
			),
		},
		{
			name: "IsReadyNotShouldReturnError",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("IsReady", mock.Anything, mock.Anything).Return(false, NewContainerStatusNotPresentError())
				m.GetDefault()
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
				podcommon.StateAllocatedResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
			),
		},
		{
			name: "StateAllocatedResourcesNotShouldReturnError",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("AllocatedResources", mock.Anything, mock.Anything, mock.Anything).
					Return(resource.Quantity{}, NewContainerStatusNotPresentError())
				m.GetDefault()
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
				podcommon.StateAllocatedResourcesUnknown,
				podcommon.StateStatusResourcesUnknown,
			),
		},
		{
			name: "StateStatusResourcesNotShouldReturnError",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("CurrentRequests", mock.Anything, mock.Anything, mock.Anything).
					Return(resource.Quantity{}, NewContainerStatusNotPresentError())
				m.GetDefault()
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
				m.IsReadyDefault()
				applyMockRequestsLimitsStartup(m)
				m.AllocatedResourcesDefault()
			},
			wantStates: podcommon.NewStates(
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateContainerRunning,
				podcommon.StateBoolTrue,
				podcommon.StateBoolTrue,
				podcommon.StateResourcesStartup,
				podcommon.StateAllocatedResourcesContainerRequestsMismatch,
				podcommon.StateStatusResourcesUnknown,
			),
		},
		{
			name: string(podcommon.StateResourcesStartup),
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.GetDefault()
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
				m.IsReadyDefault()
				applyMockRequestsLimitsStartup(m)
				m.AllocatedResourcesDefault()
				m.CurrentRequestsDefault()
				m.CurrentLimitsDefault()
			},
			wantStateResources: podcommon.StateResourcesStartup,
		},
		{
			name: string(podcommon.StateResourcesPostStartup),
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.GetDefault()
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
				m.IsReadyDefault()
				applyMockRequestsLimitsPostStartup(m)
				m.AllocatedResourcesDefault()
				m.CurrentRequestsDefault()
				m.CurrentLimitsDefault()
			},
			wantStateResources: podcommon.StateResourcesPostStartup,
		},
		{
			name: string(podcommon.StateResourcesUnknown),
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.GetDefault()
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.StateDefault()
				m.IsStartedDefault()
				m.IsReadyDefault()
				applyMockRequestsLimitsUnknown(m)
				m.AllocatedResourcesDefault()
				m.CurrentRequestsDefault()
				m.CurrentLimitsDefault()
			},
			wantStateResources: podcommon.StateResourcesUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(podtest.NewMockContainerKubeHelper(tt.configMockFunc))
			config := &scaleConfig{
				targetContainerName: podtest.DefaultContainerName,
				cpuConfig: podcommon.NewCpuConfig(
					podtest.PodAnnotationCpuStartupQuantity,
					podtest.PodAnnotationCpuPostStartupRequestsQuantity,
					podtest.PodAnnotationCpuPostStartupLimitsQuantity,
				),
				memoryConfig: podcommon.NewMemoryConfig(
					podtest.PodAnnotationMemoryStartupQuantity,
					podtest.PodAnnotationMemoryPostStartupRequestsQuantity,
					podtest.PodAnnotationMemoryPostStartupLimitsQuantity,
				),
			}

			got, err := s.States(contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(), &v1.Pod{}, config)
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
		configMockFunc func(*podtest.MockContainerKubeHelper)
		want           podcommon.StateBool
	}{
		{
			name: "StateBoolTrue",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("HasStartupProbe", mock.Anything, mock.Anything).Return(true, nil)
			},
			want: podcommon.StateBoolTrue,
		},
		{
			name: "StateBoolFalse",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("HasStartupProbe", mock.Anything, mock.Anything).Return(false, nil)
			},
			want: podcommon.StateBoolFalse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(podtest.NewMockContainerKubeHelper(tt.configMockFunc))
			assert.Equal(t, tt.want, s.stateStartupProbe(&v1.Container{}))
		})
	}
}

func TestTargetContainerStateStateReadinessProbe(t *testing.T) {
	tests := []struct {
		name           string
		configMockFunc func(*podtest.MockContainerKubeHelper)
		want           podcommon.StateBool
	}{
		{
			name: "StateBoolTrue",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("HasReadinessProbe", mock.Anything, mock.Anything).Return(true, nil)
			},
			want: podcommon.StateBoolTrue,
		},
		{
			name: "StateBoolFalse",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("HasReadinessProbe", mock.Anything, mock.Anything).Return(false, nil)
			},
			want: podcommon.StateBoolFalse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(podtest.NewMockContainerKubeHelper(tt.configMockFunc))
			assert.Equal(t, tt.want, s.stateReadinessProbe(&v1.Container{}))
		})
	}
}

func TestTargetContainerStateStateContainer(t *testing.T) {
	tests := []struct {
		name           string
		configMockFunc func(*podtest.MockContainerKubeHelper)
		want           podcommon.StateContainer
		wantErrMsg     string
	}{
		{
			name: "UnableToGetContainerState",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{}, errors.New(""))
			},
			want:       podcommon.StateContainerUnknown,
			wantErrMsg: "unable to get container state",
		},
		{
			name: "StateContainerRunning",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{
					Running: &v1.ContainerStateRunning{},
				}, nil)
			},
			want: podcommon.StateContainerRunning,
		},
		{
			name: "StateContainerWaiting",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{},
				}, nil)
			},
			want: podcommon.StateContainerWaiting,
		},
		{
			name: "StateContainerTerminated",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{},
				}, nil)
			},
			want: podcommon.StateContainerTerminated,
		},
		{
			name: "StateContainerUnknown",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{}, nil)
			},
			want: podcommon.StateContainerUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(podtest.NewMockContainerKubeHelper(tt.configMockFunc))

			got, err := s.stateContainer(&v1.Pod{}, NewScaleConfig(nil))
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
		configMockFunc func(*podtest.MockContainerKubeHelper)
		want           podcommon.StateBool
		wantErrMsg     string
	}{
		{
			name: "UnableToGetContainerReadyStatus",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("IsStarted", mock.Anything, mock.Anything).Return(false, errors.New(""))
			},
			want:       podcommon.StateBoolUnknown,
			wantErrMsg: "unable to get container ready status",
		},
		{
			name: "StateBoolTrue",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("IsStarted", mock.Anything, mock.Anything).Return(true, nil)
			},
			want: podcommon.StateBoolTrue,
		},
		{
			name: "StateBoolFalse",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("IsStarted", mock.Anything, mock.Anything).Return(false, nil)
			},
			want: podcommon.StateBoolFalse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(podtest.NewMockContainerKubeHelper(tt.configMockFunc))

			got, err := s.stateStarted(&v1.Pod{}, NewScaleConfig(nil))
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
		configMockFunc func(*podtest.MockContainerKubeHelper)
		want           podcommon.StateBool
		wantErrMsg     string
	}{
		{
			name: "UnableToGetContainerReadyStatus",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("IsReady", mock.Anything, mock.Anything).Return(false, errors.New(""))
			},
			want:       podcommon.StateBoolUnknown,
			wantErrMsg: "unable to get container ready status",
		},
		{
			name: "StateBoolTrue",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("IsReady", mock.Anything, mock.Anything).Return(true, nil)
			},
			want: podcommon.StateBoolTrue,
		},
		{
			name: "StateBoolFalse",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("IsReady", mock.Anything, mock.Anything).Return(false, nil)
			},
			want: podcommon.StateBoolFalse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(podtest.NewMockContainerKubeHelper(tt.configMockFunc))

			got, err := s.stateReady(&v1.Pod{}, NewScaleConfig(nil))
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTargetContainerStateStateAllocatedResources(t *testing.T) {
	tests := []struct {
		name           string
		configMockFunc func(*podtest.MockContainerKubeHelper)
		want           podcommon.StateAllocatedResources
		wantErrMsg     string
	}{
		{
			name: "UnableToGetAllocatedCpuResources",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("AllocatedResources", mock.Anything, mock.Anything, v1.ResourceCPU).
					Return(resource.Quantity{}, errors.New(""))
			},
			want:       podcommon.StateAllocatedResourcesUnknown,
			wantErrMsg: "unable to get allocated cpu resources",
		},
		{
			name: "UnableToGetAllocatedMemoryResources",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("AllocatedResources", mock.Anything, mock.Anything, v1.ResourceMemory).
					Return(resource.Quantity{}, errors.New(""))
				m.AllocatedResourcesDefault()
			},
			want:       podcommon.StateAllocatedResourcesUnknown,
			wantErrMsg: "unable to get allocated memory resources",
		},
		{
			name: string(podcommon.StateAllocatedResourcesIncomplete),
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("AllocatedResources", mock.Anything, mock.Anything, v1.ResourceCPU).
					Return(resource.Quantity{}, nil)
				m.AllocatedResourcesDefault()
			},
			want: podcommon.StateAllocatedResourcesIncomplete,
		},
		{
			name: string(podcommon.StateAllocatedResourcesContainerRequestsMatch),
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("AllocatedResources", mock.Anything, mock.Anything, v1.ResourceCPU).
					Return(podtest.PodAnnotationCpuStartupQuantity, nil)
				m.On("AllocatedResources", mock.Anything, mock.Anything, v1.ResourceMemory).
					Return(podtest.PodAnnotationMemoryStartupQuantity, nil)
				m.On("Requests", mock.Anything, v1.ResourceCPU).
					Return(podtest.PodAnnotationCpuStartupQuantity, nil)
				m.On("Requests", mock.Anything, v1.ResourceMemory).
					Return(podtest.PodAnnotationMemoryStartupQuantity, nil)
			},
			want: podcommon.StateAllocatedResourcesContainerRequestsMatch,
		},
		{
			name: string(podcommon.StateAllocatedResourcesContainerRequestsMismatch),
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("AllocatedResources", mock.Anything, mock.Anything, v1.ResourceCPU).
					Return(podtest.PodAnnotationCpuStartupQuantity, nil)
				m.On("AllocatedResources", mock.Anything, mock.Anything, v1.ResourceMemory).
					Return(podtest.PodAnnotationMemoryStartupQuantity, nil)
				m.On("Requests", mock.Anything, v1.ResourceCPU).
					Return(podtest.PodAnnotationCpuPostStartupRequestsQuantity, nil)
				m.On("Requests", mock.Anything, v1.ResourceMemory).
					Return(podtest.PodAnnotationMemoryStartupQuantity, nil)
			},
			want: podcommon.StateAllocatedResourcesContainerRequestsMismatch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(podtest.NewMockContainerKubeHelper(tt.configMockFunc))

			got, err := s.stateAllocatedResources(&v1.Pod{}, &v1.Container{}, NewScaleConfig(nil))
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
		configMockFunc func(*podtest.MockContainerKubeHelper)
		want           podcommon.StateStatusResources
		wantErrMsg     string
	}{
		{
			name: "UnableToGetStatusResourcesCpuRequests",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("CurrentRequests", mock.Anything, mock.Anything, v1.ResourceCPU).
					Return(resource.Quantity{}, errors.New(""))
			},
			want:       podcommon.StateStatusResourcesUnknown,
			wantErrMsg: "unable to get status resources cpu requests",
		},
		{
			name: "UnableToGetStatusResourcesCpuLimits",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("CurrentLimits", mock.Anything, mock.Anything, v1.ResourceCPU).
					Return(resource.Quantity{}, errors.New(""))
				m.CurrentRequestsDefault()
			},
			want:       podcommon.StateStatusResourcesUnknown,
			wantErrMsg: "unable to get status resources cpu limits",
		},
		{
			name: "UnableToGetStatusResourcesMemoryRequests",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("CurrentRequests", mock.Anything, mock.Anything, v1.ResourceMemory).
					Return(resource.Quantity{}, errors.New(""))
				m.CurrentRequestsDefault()
				m.CurrentLimitsDefault()
			},
			want:       podcommon.StateStatusResourcesUnknown,
			wantErrMsg: "unable to get status resources memory requests",
		},
		{
			name: "UnableToGetStatusResourcesMemoryLimits",
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("CurrentLimits", mock.Anything, mock.Anything, v1.ResourceMemory).
					Return(resource.Quantity{}, errors.New(""))
				m.CurrentRequestsDefault()
				m.CurrentLimitsDefault()
			},
			want:       podcommon.StateStatusResourcesUnknown,
			wantErrMsg: "unable to get status resources memory limits",
		},
		{
			name: string(podcommon.StateStatusResourcesIncomplete),
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("CurrentRequests", mock.Anything, mock.Anything, mock.Anything).
					Return(resource.Quantity{}, nil)
				m.CurrentLimitsDefault()
			},
			want: podcommon.StateStatusResourcesIncomplete,
		},
		{
			name: string(podcommon.StateStatusResourcesContainerResourcesMatch),
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("CurrentRequests", mock.Anything, mock.Anything, v1.ResourceCPU).
					Return(podtest.PodAnnotationCpuStartupQuantity, nil)
				m.On("CurrentRequests", mock.Anything, mock.Anything, v1.ResourceMemory).
					Return(podtest.PodAnnotationMemoryStartupQuantity, nil)
				m.On("CurrentLimits", mock.Anything, mock.Anything, v1.ResourceCPU).
					Return(podtest.PodAnnotationCpuStartupQuantity, nil)
				m.On("CurrentLimits", mock.Anything, mock.Anything, v1.ResourceMemory).
					Return(podtest.PodAnnotationMemoryStartupQuantity, nil)
				m.On("Requests", mock.Anything, v1.ResourceCPU).
					Return(podtest.PodAnnotationCpuStartupQuantity, nil)
				m.On("Requests", mock.Anything, v1.ResourceMemory).
					Return(podtest.PodAnnotationMemoryStartupQuantity, nil)
				m.On("Limits", mock.Anything, v1.ResourceCPU).
					Return(podtest.PodAnnotationCpuStartupQuantity, nil)
				m.On("Limits", mock.Anything, v1.ResourceMemory).
					Return(podtest.PodAnnotationMemoryStartupQuantity, nil)
			},
			want: podcommon.StateStatusResourcesContainerResourcesMatch,
		},
		{
			name: string(podcommon.StateStatusResourcesContainerResourcesMismatch),
			configMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("CurrentRequests", mock.Anything, mock.Anything, v1.ResourceCPU).
					Return(podtest.PodAnnotationCpuStartupQuantity, nil)
				m.On("CurrentRequests", mock.Anything, mock.Anything, v1.ResourceMemory).
					Return(podtest.PodAnnotationMemoryStartupQuantity, nil)
				m.On("CurrentLimits", mock.Anything, mock.Anything, v1.ResourceCPU).
					Return(podtest.PodAnnotationCpuStartupQuantity, nil)
				m.On("CurrentLimits", mock.Anything, mock.Anything, v1.ResourceMemory).
					Return(podtest.PodAnnotationMemoryStartupQuantity, nil)
				m.On("Requests", mock.Anything, v1.ResourceCPU).
					Return(podtest.PodAnnotationCpuPostStartupRequestsQuantity, nil)
				m.On("Requests", mock.Anything, v1.ResourceMemory).
					Return(podtest.PodAnnotationMemoryPostStartupRequestsQuantity, nil)
				m.On("Limits", mock.Anything, v1.ResourceCPU).
					Return(podtest.PodAnnotationCpuPostStartupRequestsQuantity, nil)
				m.On("Limits", mock.Anything, v1.ResourceMemory).
					Return(podtest.PodAnnotationMemoryPostStartupRequestsQuantity, nil)
			},
			want: podcommon.StateStatusResourcesContainerResourcesMismatch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(podtest.NewMockContainerKubeHelper(tt.configMockFunc))

			got, err := s.stateStatusResources(&v1.Pod{}, &v1.Container{}, NewScaleConfig(nil))
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
		{"ContainerStatusResourcesNotPresentErrorFalse", NewContainerStatusResourcesNotPresentError(), false},
		{"ContainerStatusAllocatedResourcesNotPresentErrorFalse", NewContainerStatusAllocatedResourcesNotPresentError(), false},
		{"NewContainerStatusResourcesNotPresentErrorFalse", NewContainerStatusResourcesNotPresentError(), false},
		{"True", errors.New(""), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTargetContainerState(nil)
			assert.Equal(
				t,
				tt.want,
				s.shouldReturnError(contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(), tt.err),
			)
		})
	}
}

func applyMockRequestsLimitsStartup(m *podtest.MockContainerKubeHelper) {
	m.On("Requests", mock.Anything, v1.ResourceCPU).Return(podtest.PodAnnotationCpuStartupQuantity)
	m.On("Requests", mock.Anything, v1.ResourceMemory).Return(podtest.PodAnnotationMemoryStartupQuantity)
	m.On("Limits", mock.Anything, v1.ResourceCPU).Return(podtest.PodAnnotationCpuStartupQuantity)
	m.On("Limits", mock.Anything, v1.ResourceMemory).Return(podtest.PodAnnotationMemoryStartupQuantity)
}

func applyMockRequestsLimitsPostStartup(m *podtest.MockContainerKubeHelper) {
	m.On("Requests", mock.Anything, v1.ResourceCPU).Return(podtest.PodAnnotationCpuPostStartupRequestsQuantity)
	m.On("Requests", mock.Anything, v1.ResourceMemory).Return(podtest.PodAnnotationMemoryPostStartupRequestsQuantity)
	m.On("Limits", mock.Anything, v1.ResourceCPU).Return(podtest.PodAnnotationCpuPostStartupLimitsQuantity)
	m.On("Limits", mock.Anything, v1.ResourceMemory).Return(podtest.PodAnnotationMemoryPostStartupLimitsQuantity)
}

func applyMockRequestsLimitsUnknown(m *podtest.MockContainerKubeHelper) {
	m.On("Requests", mock.Anything, v1.ResourceCPU).Return(podtest.PodAnnotationCpuUnknownQuantity)
	m.On("Requests", mock.Anything, v1.ResourceMemory).Return(podtest.PodAnnotationMemoryUnknownQuantity)
	m.On("Limits", mock.Anything, v1.ResourceCPU).Return(podtest.PodAnnotationCpuUnknownQuantity)
	m.On("Limits", mock.Anything, v1.ResourceMemory).Return(podtest.PodAnnotationMemoryUnknownQuantity)
}
