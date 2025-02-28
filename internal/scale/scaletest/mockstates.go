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

// MockStates is a generic mock for scale.States.
type MockStates struct {
	mock.Mock
}

func NewMockStates(configFunc func(*MockStates)) *MockStates {
	m := &MockStates{}
	if configFunc == nil {
		configFunc(m)
	} else {
		m.AllDefaults()
	}

	return m
}

func (m *MockStates) IsStartupConfigAppliedAll(container *v1.Container) bool {
	args := m.Called(container)
	return args.Bool(0)
}

func (m *MockStates) IsPostStartupConfigAppliedAll(container *v1.Container) bool {
	args := m.Called(container)
	return args.Bool(0)
}

func (m *MockStates) IsAnyCurrentZeroAll(pod *v1.Pod, container *v1.Container) (bool, error) {
	args := m.Called(pod, container)
	return args.Bool(0), args.Error(1)
}

func (m *MockStates) DoesRequestsCurrentMatchSpecAll(pod *v1.Pod, container *v1.Container) (bool, error) {
	args := m.Called(pod, container)
	return args.Bool(0), args.Error(1)
}

func (m *MockStates) DoesLimitsCurrentMatchSpecAll(pod *v1.Pod, container *v1.Container) (bool, error) {
	args := m.Called(pod, container)
	return args.Bool(0), args.Error(1)
}

func (m *MockStates) StateFor(name v1.ResourceName) scale.State {
	args := m.Called(name)
	return args.Get(0).(scale.State)
}

func (m *MockStates) AllStates() []scale.State {
	args := m.Called()
	return args.Get(0).([]scale.State)
}

func (m *MockStates) IsStartupConfigAppliedAllDefault() {
	m.On("IsStartupConfigAppliedAll", mock.Anything).Return(true)
}

func (m *MockStates) IsPostStartupConfigAppliedAllDefault() {
	m.On("IsPostStartupConfigAppliedAll", mock.Anything).Return(true)
}

func (m *MockStates) IsAnyCurrentZeroAllDefault() {
	m.On("IsAnyCurrentZeroAll", mock.Anything, mock.Anything).Return(false, nil)
}

func (m *MockStates) DoesRequestsCurrentMatchSpecAllDefault() {
	m.On("DoesRequestsCurrentMatchSpecAll", mock.Anything, mock.Anything).Return(true, nil)
}

func (m *MockStates) DoesLimitsCurrentMatchSpecAllDefault() {
	m.On("DoesLimitsCurrentMatchSpecAll", mock.Anything, mock.Anything).Return(true, nil)
}

func (m *MockStates) StateForDefault() {
	m.On("StateFor", mock.Anything).Return(NewMockState(nil))
}

func (m *MockStates) AllStatesDefault() {
	m.On("AllStates").Return([]scale.State{NewMockState(nil)})
}

func (m *MockStates) AllDefaults() {
	m.IsStartupConfigAppliedAllDefault()
	m.IsPostStartupConfigAppliedAllDefault()
	m.IsAnyCurrentZeroAllDefault()
	m.DoesRequestsCurrentMatchSpecAllDefault()
	m.DoesLimitsCurrentMatchSpecAllDefault()
	m.StateForDefault()
	m.AllStatesDefault()
}
