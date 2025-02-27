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

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource/config"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
)

// MockValidation is a generic mock for pod.Validation.
type MockValidation struct {
	mock.Mock
}

func NewMockValidation(configFunc func(*MockValidation)) *MockValidation {
	mockValidation := &MockValidation{}
	configFunc(mockValidation)
	return mockValidation
}

func (m *MockValidation) Validate(
	ctx context.Context,
	pod *v1.Pod,
	targetContainerName string,
	scaleConfigs config.ScaleConfigs,
) (*v1.Container, error) {
	args := m.Called(ctx, pod, targetContainerName, scaleConfigs)
	return args.Get(0).(*v1.Container), args.Error(0)
}

func (m *MockValidation) ValidateDefault() {
	m.On("Validate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&v1.Container{}, nil)
}

func (m *MockValidation) AllDefaults() {
	m.ValidateDefault()
}
