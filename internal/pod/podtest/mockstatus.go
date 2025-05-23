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

package podtest

import (
	"context"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/event/eventcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

type MockStatus struct {
	mock.Mock
}

func NewMockStatus(configFunc func(*MockStatus)) *MockStatus {
	m := &MockStatus{}
	if configFunc != nil {
		configFunc(m)
	} else {
		m.AllDefaults()
	}

	return m
}

func NewMockStatusWithRun(configFunc func(*MockStatus, func()), run func()) *MockStatus {
	m := &MockStatus{}
	if configFunc != nil {
		configFunc(m, run)
	} else {
		m.AllDefaults()
	}

	return m
}

func (m *MockStatus) Update(
	ctx context.Context,
	podEventPublisher eventcommon.PodEventPublisher,
	pod *v1.Pod,
	status string,
	states podcommon.States,
	statusScaleState podcommon.StatusScaleState,
	scaleConfigs scalecommon.Configurations,
	failReason string,
) (*v1.Pod, error) {
	args := m.Called(ctx, podEventPublisher, pod, status, states, statusScaleState, scaleConfigs, failReason)
	return args.Get(0).(*v1.Pod), args.Error(1)
}

func (m *MockStatus) UpdateDefault() {
	m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&v1.Pod{}, nil)
}

func (m *MockStatus) UpdateDefaultAndRun(run func()) {
	m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&v1.Pod{}, nil).
		Run(func(args mock.Arguments) { run() })
}

func (m *MockStatus) AllDefaults() {
	m.UpdateDefault()
}
