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
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
)

// MockTargetContainerAction is a generic mock for pod.TargetContainerAction.
type MockTargetContainerAction struct {
	mock.Mock
}

func NewMockTargetContainerAction(configFunc func(*MockTargetContainerAction)) *MockTargetContainerAction {
	m := &MockTargetContainerAction{}
	if configFunc != nil {
		configFunc(m)
	} else {
		m.AllDefaults()
	}

	return m
}

func (m *MockTargetContainerAction) Execute(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	targetContainer *v1.Container,
	scaleConfigs scalecommon.Configurations,
) error {
	args := m.Called(ctx, states, pod, targetContainer, scaleConfigs)
	return args.Error(0)
}

func (m *MockTargetContainerAction) ExecuteDefault() {
	m.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
}

func (m *MockTargetContainerAction) AllDefaults() {
	m.ExecuteDefault()
}
