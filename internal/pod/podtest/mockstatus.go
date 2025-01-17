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
	v1 "k8s.io/api/core/v1"
)

// MockStatus is a generic mock for pod.Status.
type MockStatus struct {
	mock.Mock
}

func NewMockStatus(configFunc func(*MockStatus)) *MockStatus {
	mockStatus := &MockStatus{}
	configFunc(mockStatus)
	return mockStatus
}

func NewMockStatusWithRun(configFunc func(*MockStatus, func()), run func()) *MockStatus {
	mockStatus := &MockStatus{}
	configFunc(mockStatus, run)
	return mockStatus
}

func (m *MockStatus) Update(
	ctx context.Context,
	pod *v1.Pod,
	status string,
	states podcommon.States,
	scaleState podcommon.StatusScaleState,
) (*v1.Pod, error) {
	args := m.Called(ctx, pod, status, states, scaleState)
	return args.Get(0).(*v1.Pod), args.Error(1)
}

func (m *MockStatus) PodMutationFunc(
	ctx context.Context,
	status string,
	states podcommon.States,
	scaleState podcommon.StatusScaleState,
) func(pod *v1.Pod) (bool, *v1.Pod, error) {
	args := m.Called(ctx, status, states, scaleState)
	return args.Get(0).(func(pod *v1.Pod) (bool, *v1.Pod, error))
}

func (m *MockStatus) UpdateDefault() {
	m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&v1.Pod{}, nil)
}

func (m *MockStatus) UpdateDefaultAndRun(run func()) {
	m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&v1.Pod{}, nil).
		Run(func(args mock.Arguments) { run() })
}

func (m *MockStatus) PodMutationFuncDefault() {
	m.On("PodMutationFunc", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(pod *v1.Pod) (bool, *v1.Pod, error) {
			return true, &v1.Pod{}, nil
		},
	)
}

func (m *MockStatus) PodMutationFuncDefaultAndRun(run func()) {
	m.On("PodMutationFunc", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(pod *v1.Pod) (bool, *v1.Pod, error) {
			return true, &v1.Pod{}, nil
		},
	).Run(func(args mock.Arguments) { run() })
}
