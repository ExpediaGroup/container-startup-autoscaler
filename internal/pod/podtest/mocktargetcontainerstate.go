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

package podtest

import (
	"context"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
)

// MockTargetContainerState is a generic mock for pod.TargetContainerState.
type MockTargetContainerState struct {
	mock.Mock
}

func NewMockTargetContainerState(configFunc func(*MockTargetContainerState)) *MockTargetContainerState {
	mockTargetContainerState := &MockTargetContainerState{}
	configFunc(mockTargetContainerState)
	return mockTargetContainerState
}

func (m *MockTargetContainerState) States(
	ctx context.Context,
	pod *v1.Pod,
	config podcommon.ScaleConfig,
) (podcommon.States, error) {
	args := m.Called(ctx, pod, config)
	return args.Get(0).(podcommon.States), args.Error(1)
}

func (m *MockTargetContainerState) StatesDefault() {
	m.On("States", mock.Anything, mock.Anything, mock.Anything).Return(
		podcommon.NewStates(
			podcommon.StateBoolTrue,
			podcommon.StateBoolTrue,
			podcommon.StateContainerRunning,
			podcommon.StateBoolTrue,
			podcommon.StateBoolTrue,
			podcommon.StateResourcesStartup,
			podcommon.StateAllocatedResourcesContainerRequestsMatch,
			podcommon.StateStatusResourcesContainerResourcesMatch,
		),
		nil,
	)
}
