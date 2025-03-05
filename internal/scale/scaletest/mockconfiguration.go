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

package scaletest

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

type MockConfiguration struct {
	mock.Mock
}

func NewMockConfiguration(configFunc func(*MockConfiguration)) *MockConfiguration {
	m := &MockConfiguration{}
	if configFunc != nil {
		configFunc(m)
	} else {
		m.AllDefaults()
	}

	return m
}

func (m *MockConfiguration) ResourceName() v1.ResourceName {
	args := m.Called()
	return args.Get(0).(v1.ResourceName)
}

func (m *MockConfiguration) IsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockConfiguration) Resources() scalecommon.Resources {
	args := m.Called()
	return args.Get(0).(scalecommon.Resources)
}

func (m *MockConfiguration) StoreFromAnnotations(pod *v1.Pod) error {
	args := m.Called(pod)
	return args.Error(0)
}

func (m *MockConfiguration) Validate(container *v1.Container) error {
	args := m.Called(container)
	return args.Error(0)
}

func (m *MockConfiguration) String() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfiguration) ResourceNameDefault() {
	m.On("ResourceName").Return(v1.ResourceCPU)
}

func (m *MockConfiguration) IsEnabledDefault() {
	m.On("IsEnabled").Return(true)
}

func (m *MockConfiguration) ResourcesDefault() {
	m.On("Resources").Return(ResourcesCpuEnabled)
}

func (m *MockConfiguration) StoreFromAnnotationsDefault() {
	m.On("StoreFromAnnotations", mock.Anything).Return(nil)
}

func (m *MockConfiguration) ValidateDefault() {
	m.On("Validate", mock.Anything).Return(nil)
}

func (m *MockConfiguration) StringDefault() {
	m.On("String").Return("")
}

func (m *MockConfiguration) AllDefaults() {
	m.ResourceNameDefault()
	m.IsEnabledDefault()
	m.ResourcesDefault()
	m.StoreFromAnnotationsDefault()
	m.ValidateDefault()
	m.StringDefault()
}
