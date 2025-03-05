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
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

type MockState struct {
	mock.Mock
}

func NewMockState(configFunc func(*MockState)) *MockState {
	m := &MockState{}
	if configFunc != nil {
		configFunc(m)
	} else {
		m.AllDefaults()
	}

	return m
}

func (m *MockState) ResourceName() v1.ResourceName {
	args := m.Called()
	return args.Get(0).(v1.ResourceName)
}

func (m *MockState) IsStartupConfigurationApplied(container *v1.Container) *bool {
	args := m.Called(container)
	return args.Get(0).(*bool)
}

func (m *MockState) IsPostStartupConfigurationApplied(container *v1.Container) *bool {
	args := m.Called(container)
	return args.Get(0).(*bool)
}

func (m *MockState) IsAnyCurrentZero(pod *v1.Pod, container *v1.Container) (*bool, error) {
	args := m.Called(pod, container)
	return args.Get(0).(*bool), args.Error(1)
}

func (m *MockState) DoesRequestsCurrentMatchSpec(pod *v1.Pod, container *v1.Container) (*bool, error) {
	args := m.Called(pod, container)
	return args.Get(0).(*bool), args.Error(1)
}

func (m *MockState) DoesLimitsCurrentMatchSpec(pod *v1.Pod, container *v1.Container) (*bool, error) {
	args := m.Called(pod, container)
	return args.Get(0).(*bool), args.Error(1)
}

func (m *MockState) ResourceNameDefault() {
	m.On("ResourceName").Return(v1.ResourceCPU)
}

func (m *MockState) IsStartupConfigAppliedDefault() {
	ret := true
	m.On("IsStartupConfigurationApplied", mock.Anything).Return(&ret)
}

func (m *MockState) IsPostStartupConfigAppliedDefault() {
	ret := false
	m.On("IsPostStartupConfigurationApplied", mock.Anything).Return(&ret)
}

func (m *MockState) IsAnyCurrentZeroDefault() {
	ret := false
	m.On("IsAnyCurrentZero", mock.Anything, mock.Anything).Return(&ret, nil)
}

func (m *MockState) DoesRequestsCurrentMatchSpecDefault() {
	ret := true
	m.On("DoesRequestsCurrentMatchSpec", mock.Anything, mock.Anything).Return(&ret, nil)
}

func (m *MockState) DoesLimitsCurrentMatchSpecDefault() {
	ret := true
	m.On("DoesLimitsCurrentMatchSpec", mock.Anything, mock.Anything).Return(&ret, nil)
}

func (m *MockState) AllDefaults() {
	m.ResourceNameDefault()
	m.IsStartupConfigAppliedDefault()
	m.IsPostStartupConfigAppliedDefault()
	m.IsAnyCurrentZeroDefault()
	m.DoesRequestsCurrentMatchSpecDefault()
	m.DoesLimitsCurrentMatchSpecDefault()
}
