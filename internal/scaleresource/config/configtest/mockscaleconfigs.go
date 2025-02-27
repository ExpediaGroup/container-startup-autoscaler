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

// MockScaleConfigs is a generic mock for config.ScaleConfigs.
type MockScaleConfigs struct {
	mock.Mock
}

func NewMockScaleConfigs(configFunc func(*MockScaleConfigs)) *MockScaleConfigs {
	mockScaleConfigs := &MockScaleConfigs{}
	configFunc(mockScaleConfigs)
	return mockScaleConfigs
}

func (m *MockScaleConfigs) TargetContainerName(pod *v1.Pod) (string, error) {
	args := m.Called(pod)
	return args.String(0), args.Error(0)
}

func (m *MockScaleConfigs) StoreFromAnnotationsAll(pod *v1.Pod) error {
	args := m.Called(pod)
	return args.Error(0)
}

func (m *MockScaleConfigs) ValidateAll(container *v1.Container) error {
	args := m.Called(container)
	return args.Error(0)
}

func (m *MockScaleConfigs) ValidateCollection() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockScaleConfigs) ScaleConfigFor(resourceType scaleresource.ResourceType) config.ScaleConfig {
	args := m.Called(resourceType)
	return args.Get(0).(config.ScaleConfig)
}

func (m *MockScaleConfigs) AllScaleConfigs() []config.ScaleConfig {
	args := m.Called()
	return args.Get(0).([]config.ScaleConfig)
}

func (m *MockScaleConfigs) AllEnabledScaleConfigs() []config.ScaleConfig {
	args := m.Called()
	return args.Get(0).([]config.ScaleConfig)
}

func (m *MockScaleConfigs) AllEnabledScaleConfigsTypes() []scaleresource.ResourceType {
	args := m.Called()
	return args.Get(0).([]scaleresource.ResourceType)
}

func (m *MockScaleConfigs) String() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockScaleConfigs) TargetContainerNameDefault() {
	m.On("TargetContainerName", mock.Anything).Return("", nil)
}

func (m *MockScaleConfigs) StoreFromAnnotationsAllDefault() {
	m.On("StoreFromAnnotationsAll", mock.Anything).Return(nil)
}

func (m *MockScaleConfigs) ValidateAllDefault() {
	m.On("ValidateAll", mock.Anything).Return(nil)
}

func (m *MockScaleConfigs) ValidateCollectionDefault() {
	m.On("ValidateCollection").Return(nil)
}

func (m *MockScaleConfigs) ScaleConfigForDefault() {
	m.On("ScaleConfigFor", mock.Anything).Return(config.NewCpuScaleConfig(false, nil, nil))
}

func (m *MockScaleConfigs) AllScaleConfigsDefault() {
	m.On("AllScaleConfigs").Return([]config.ScaleConfig{config.NewCpuScaleConfig(false, nil, nil)})
}

func (m *MockScaleConfigs) AllEnabledScaleConfigsDefault() {
	m.On("AllEnabledScaleConfigs").Return([]config.ScaleConfig{config.NewCpuScaleConfig(false, nil, nil)})
}

func (m *MockScaleConfigs) AllEnabledScaleConfigsTypesDefault() {
	m.On("AllEnabledScaleConfigsTypes").Return([]scaleresource.ResourceType{scaleresource.ResourceTypeCpu})
}

func (m *MockScaleConfigs) StringDefault() {
	m.On("String").Return("")
}
