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
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scaletest"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
)

// MockConfiguration is a generic mock for pod.Configuration.
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

func (m *MockConfiguration) Configure(pod *v1.Pod) (scale.Configs, error) {
	args := m.Called(pod)
	return args.Get(0).(scale.Configs), args.Error(1)
}

func (m *MockConfiguration) ConfigureDefault() {
	m.On("Configure", mock.Anything).Return(scaletest.NewMockConfigs(nil), nil)
}

func (m *MockConfiguration) AllDefaults() {
	m.ConfigureDefault()
}
