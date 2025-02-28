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

// MockConfig is a generic mock for scale.Config.
type MockConfig struct {
	mock.Mock
}

func NewMockConfig(configFunc func(*MockConfig)) *MockConfig {
	m := &MockConfig{}
	if configFunc == nil {
		configFunc(m)
	} else {
		m.AllDefaults()
	}

	return m
}

func (m *MockConfig) ResourceName() v1.ResourceName {
	args := m.Called()
	return args.Get(0).(v1.ResourceName)
}

func (m *MockConfig) IsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockConfig) Resources() scale.Resources {
	args := m.Called()
	return args.Get(0).(scale.Resources)
}

func (m *MockConfig) StoreFromAnnotations(pod *v1.Pod) error {
	args := m.Called(pod)
	return args.Error(0)
}

func (m *MockConfig) Validate(container *v1.Container) error {
	args := m.Called(container)
	return args.Error(0)
}

func (m *MockConfig) String() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfig) ResourceNameDefault() {
	m.On("ResourceName").Return(v1.ResourceCPU)
}

func (m *MockConfig) IsEnabledDefault() {
	m.On("IsEnabled").Return(true)
}

func (m *MockConfig) ResourcesDefault() {
	m.On("Resources").Return(scale.Resources{})
}

func (m *MockConfig) StoreFromAnnotationsDefault() {
	m.On("StoreFromAnnotations", mock.Anything).Return(nil)
}

func (m *MockConfig) ValidateDefault() {
	m.On("Validate", mock.Anything).Return(nil)
}

func (m *MockConfig) StringDefault() {
	m.On("String").Return("")
}

func (m *MockConfig) AllDefaults() {
	m.ResourceNameDefault()
	m.IsEnabledDefault()
	m.ResourcesDefault()
	m.StoreFromAnnotationsDefault()
	m.ValidateDefault()
	m.StringDefault()
}
