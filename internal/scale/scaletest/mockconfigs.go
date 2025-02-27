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

package scaletest

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

// MockConfigs is a generic mock for scale.Configs.
type MockConfigs struct {
	mock.Mock
}

func NewMockConfigs(configFunc func(*MockConfigs)) *MockConfigs {
	mockConfigs := &MockConfigs{}
	configFunc(mockConfigs)
	return mockConfigs
}

func (m *MockConfigs) TargetContainerName(pod *v1.Pod) (string, error) {
	args := m.Called(pod)
	return args.String(0), args.Error(0)
}

func (m *MockConfigs) StoreFromAnnotationsAll(pod *v1.Pod) error {
	args := m.Called(pod)
	return args.Error(0)
}

func (m *MockConfigs) ValidateAll(container *v1.Container) error {
	args := m.Called(container)
	return args.Error(0)
}

func (m *MockConfigs) ValidateCollection() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConfigs) ConfigFor(resourceName v1.ResourceName) scale.Config {
	args := m.Called(resourceName)
	return args.Get(0).(scale.Config)
}

func (m *MockConfigs) AllConfigs() []scale.Config {
	args := m.Called()
	return args.Get(0).([]scale.Config)
}

func (m *MockConfigs) AllEnabledConfigs() []scale.Config {
	args := m.Called()
	return args.Get(0).([]scale.Config)
}

func (m *MockConfigs) AllEnabledConfigsResourceNames() []v1.ResourceName {
	args := m.Called()
	return args.Get(0).([]v1.ResourceName)
}

func (m *MockConfigs) String() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfigs) TargetContainerNameDefault() {
	m.On("TargetContainerName", mock.Anything).Return("", nil)
}

func (m *MockConfigs) StoreFromAnnotationsAllDefault() {
	m.On("StoreFromAnnotationsAll", mock.Anything).Return(nil)
}

func (m *MockConfigs) ValidateAllDefault() {
	m.On("ValidateAll", mock.Anything).Return(nil)
}

func (m *MockConfigs) ValidateCollectionDefault() {
	m.On("ValidateCollection").Return(nil)
}

func (m *MockConfigs) ConfigForDefault() {
	m.On("ConfigFor", mock.Anything).
		Return(scale.NewConfig(v1.ResourceCPU, "", "", "", true, nil, nil))
}

func (m *MockConfigs) AllConfigsDefault() {
	m.On("AllConfigs").Return([]scale.Config{})
}

func (m *MockConfigs) AllEnabledConfigsDefault() {
	m.On("AllEnabledConfigs").Return([]scale.Config{})
}

func (m *MockConfigs) AllEnabledConfigsResourceNamesDefault() {
	m.On("AllEnabledConfigsResourceNames").Return([]v1.ResourceName{})
}

func (m *MockConfigs) StringDefault() {
	m.On("String").Return("")
}

func (m *MockConfigs) AllDefaults() {
	m.TargetContainerNameDefault()
	m.StoreFromAnnotationsAllDefault()
	m.ValidateAllDefault()
	m.ValidateCollectionDefault()
	m.ConfigForDefault()
	m.AllConfigsDefault()
	m.AllEnabledConfigsDefault()
	m.AllEnabledConfigsResourceNamesDefault()
	m.StringDefault()
}
