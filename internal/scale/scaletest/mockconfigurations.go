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
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

type MockConfigurations struct {
	mock.Mock
}

func NewMockConfigurations(configFunc func(*MockConfigurations)) *MockConfigurations {
	m := &MockConfigurations{}
	if configFunc != nil {
		configFunc(m)
	} else {
		m.AllDefaults()
	}

	return m
}

func (m *MockConfigurations) TargetContainerName(pod *v1.Pod) (string, error) {
	args := m.Called(pod)
	return args.String(0), args.Error(1)
}

func (m *MockConfigurations) StoreFromAnnotationsAll(pod *v1.Pod) error {
	args := m.Called(pod)
	return args.Error(0)
}

func (m *MockConfigurations) ValidateAll(container *v1.Container) error {
	args := m.Called(container)
	return args.Error(0)
}

func (m *MockConfigurations) ValidateCollection(container *v1.Container) error {
	args := m.Called(container)
	return args.Error(0)
}

func (m *MockConfigurations) ConfigurationFor(resourceName v1.ResourceName) scalecommon.Configuration {
	args := m.Called(resourceName)
	return args.Get(0).(scalecommon.Configuration)
}

func (m *MockConfigurations) AllConfigurations() []scalecommon.Configuration {
	args := m.Called()
	return args.Get(0).([]scalecommon.Configuration)
}

func (m *MockConfigurations) AllEnabledConfigurations() []scalecommon.Configuration {
	args := m.Called()
	return args.Get(0).([]scalecommon.Configuration)
}

func (m *MockConfigurations) AllEnabledConfigurationsResourceNames() []v1.ResourceName {
	args := m.Called()
	return args.Get(0).([]v1.ResourceName)
}

func (m *MockConfigurations) String() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfigurations) TargetContainerNameDefault() {
	m.On("TargetContainerName", mock.Anything).Return(kubetest.DefaultContainerName, nil)
}

func (m *MockConfigurations) StoreFromAnnotationsAllDefault() {
	m.On("StoreFromAnnotationsAll", mock.Anything).Return(nil)
}

func (m *MockConfigurations) ValidateAllDefault() {
	m.On("ValidateAll", mock.Anything).Return(nil)
}

func (m *MockConfigurations) ValidateCollectionDefault() {
	m.On("ValidateCollection", mock.Anything).Return(nil)
}

func (m *MockConfigurations) ConfigForDefault() {
	m.On("ConfigurationFor", mock.Anything).Return(NewMockConfiguration(nil))
}

func (m *MockConfigurations) AllConfigsDefault() {
	m.On("AllConfigurations").Return([]scalecommon.Configuration{NewMockConfiguration(nil)})
}

func (m *MockConfigurations) AllEnabledConfigsDefault() {
	m.On("AllEnabledConfigurations").Return([]scalecommon.Configuration{NewMockConfiguration(nil)})
}

func (m *MockConfigurations) AllEnabledConfigsResourceNamesDefault() {
	m.On("AllEnabledConfigurationsResourceNames").Return([]v1.ResourceName{v1.ResourceCPU})
}

func (m *MockConfigurations) StringDefault() {
	m.On("String").Return("")
}

func (m *MockConfigurations) AllDefaults() {
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
