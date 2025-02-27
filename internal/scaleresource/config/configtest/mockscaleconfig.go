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

package configtest

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource/config"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

// MockScaleConfig is a generic mock for config.ScaleConfig.
type MockScaleConfig struct {
	mock.Mock
}

func NewMockScaleConfig(configFunc func(*MockScaleConfig)) *MockScaleConfig {
	mockScaleConfig := &MockScaleConfig{}
	configFunc(mockScaleConfig)
	return mockScaleConfig
}

func (m *MockScaleConfig) ResourceType() scaleresource.ResourceType {
	args := m.Called()
	return args.Get(0).(scaleresource.ResourceType)
}

func (m *MockScaleConfig) IsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockScaleConfig) IsCsaEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockScaleConfig) IsUserEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockScaleConfig) Resources() config.Resources {
	args := m.Called()
	return args.Get(0).(config.Resources)
}

func (m *MockScaleConfig) StoreFromAnnotations(pod *v1.Pod) error {
	args := m.Called(pod)
	return args.Error(0)
}

func (m *MockScaleConfig) Validate(container *v1.Container) error {
	args := m.Called(container)
	return args.Error(0)
}

func (m *MockScaleConfig) ResourceTypeDefault() {
	m.On("ResourceType").Return(scaleresource.ResourceTypeCpu)
}

func (m *MockScaleConfig) IsEnabledDefault() {
	m.On("IsEnabled").Return(true)
}

func (m *MockScaleConfig) IsCsaEnabledDefault() {
	m.On("IsCsaEnabled").Return(true)
}

func (m *MockScaleConfig) IsUserEnabledDefault() {
	m.On("IsUserEnabled").Return(true)
}

func (m *MockScaleConfig) ResourcesDefault() {
	m.On("Resources").Return(config.Resources{})
}

func (m *MockScaleConfig) StoreFromAnnotationsDefault() {
	m.On("StoreFromAnnotations", mock.Anything).Return(nil)
}

func (m *MockScaleConfig) ValidateDefault() {
	m.On("Validate", mock.Anything).Return(nil)
}

func (m *MockScaleConfig) StringDefault() {
	m.On("String").Return("")
}
