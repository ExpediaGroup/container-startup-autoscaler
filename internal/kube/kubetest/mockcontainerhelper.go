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

package kubetest

import (
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// MockContainerHelper is a generic mock for pod.ContainerHelper.
type MockContainerHelper struct {
	mock.Mock
}

func NewMockContainerHelper(configFunc func(*MockContainerHelper)) *MockContainerHelper {
	m := &MockContainerHelper{}
	if configFunc == nil {
		configFunc(m)
	} else {
		m.AllDefaults()
	}

	return m
}

func (m *MockContainerHelper) Get(pod *v1.Pod, containerName string) (*v1.Container, error) {
	args := m.Called(pod, containerName)
	return args.Get(0).(*v1.Container), args.Error(1)
}

func (m *MockContainerHelper) HasStartupProbe(container *v1.Container) bool {
	args := m.Called(container)
	return args.Bool(0)
}

func (m *MockContainerHelper) HasReadinessProbe(container *v1.Container) bool {
	args := m.Called(container)
	return args.Bool(0)
}

func (m *MockContainerHelper) State(pod *v1.Pod, container *v1.Container) (v1.ContainerState, error) {
	args := m.Called(pod, container)
	return args.Get(0).(v1.ContainerState), args.Error(1)
}

func (m *MockContainerHelper) IsStarted(pod *v1.Pod, container *v1.Container) (bool, error) {
	args := m.Called(pod, container)
	return args.Bool(0), args.Error(1)
}

func (m *MockContainerHelper) IsReady(pod *v1.Pod, container *v1.Container) (bool, error) {
	args := m.Called(pod, container)
	return args.Bool(0), args.Error(1)
}

func (m *MockContainerHelper) Requests(container *v1.Container, resourceName v1.ResourceName) resource.Quantity {
	args := m.Called(container, resourceName)
	return args.Get(0).(resource.Quantity)
}

func (m *MockContainerHelper) Limits(container *v1.Container, resourceName v1.ResourceName) resource.Quantity {
	args := m.Called(container, resourceName)
	return args.Get(0).(resource.Quantity)
}

func (m *MockContainerHelper) ResizePolicy(
	container *v1.Container,
	resourceName v1.ResourceName,
) (v1.ResourceResizeRestartPolicy, error) {
	args := m.Called(container, resourceName)
	return args.Get(0).(v1.ResourceResizeRestartPolicy), args.Error(1)
}

func (m *MockContainerHelper) CurrentRequests(
	pod *v1.Pod,
	container *v1.Container,
	resourceName v1.ResourceName,
) (resource.Quantity, error) {
	args := m.Called(pod, container, resourceName)
	return args.Get(0).(resource.Quantity), args.Error(1)
}

func (m *MockContainerHelper) CurrentLimits(
	pod *v1.Pod,
	container *v1.Container,
	resourceName v1.ResourceName,
) (resource.Quantity, error) {
	args := m.Called(pod, container, resourceName)
	return args.Get(0).(resource.Quantity), args.Error(1)
}

func (m *MockContainerHelper) GetDefault() {
	m.On("Get", mock.Anything, mock.Anything).Return(
		NewContainerBuilder(NewStartupContainerConfig()).Build(),
		nil,
	)
}

func (m *MockContainerHelper) HasStartupProbeDefault() {
	m.On("HasStartupProbe", mock.Anything).Return(true)
}

func (m *MockContainerHelper) HasReadinessProbeDefault() {
	m.On("HasReadinessProbe", mock.Anything).Return(true)
}

func (m *MockContainerHelper) StateDefault() {
	m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{Running: &v1.ContainerStateRunning{}}, nil)
}

func (m *MockContainerHelper) IsStartedDefault() {
	m.On("IsStarted", mock.Anything, mock.Anything).Return(true, nil)
}

func (m *MockContainerHelper) IsReadyDefault() {
	m.On("IsReady", mock.Anything, mock.Anything).Return(true, nil)
}

func (m *MockContainerHelper) RequestsDefault() {
	m.On("Requests", mock.Anything, v1.ResourceCPU).Return(MockDefaultCpuQuantity)
	m.On("Requests", mock.Anything, v1.ResourceMemory).Return(MockDefaultMemoryQuantity)
}

func (m *MockContainerHelper) LimitsDefault() {
	m.On("Limits", mock.Anything, v1.ResourceCPU).Return(MockDefaultCpuQuantity)
	m.On("Limits", mock.Anything, v1.ResourceMemory).Return(MockDefaultMemoryQuantity)
}

func (m *MockContainerHelper) ResizePolicyDefault() {
	m.On("ResizePolicy", mock.Anything, mock.Anything).Return(v1.NotRequired, nil)
}

func (m *MockContainerHelper) CurrentRequestsDefault() {
	m.On("CurrentRequests", mock.Anything, mock.Anything, v1.ResourceCPU).Return(MockDefaultCpuQuantity, nil)
	m.On("CurrentRequests", mock.Anything, mock.Anything, v1.ResourceMemory).Return(MockDefaultMemoryQuantity, nil)
}

func (m *MockContainerHelper) CurrentLimitsDefault() {
	m.On("CurrentLimits", mock.Anything, mock.Anything, v1.ResourceCPU).Return(MockDefaultCpuQuantity, nil)
	m.On("CurrentLimits", mock.Anything, mock.Anything, v1.ResourceMemory).Return(MockDefaultMemoryQuantity, nil)
}

func (m *MockContainerHelper) AllDefaults() {
	m.GetDefault()
	m.HasStartupProbeDefault()
	m.HasReadinessProbeDefault()
	m.StateDefault()
	m.IsStartedDefault()
	m.IsReadyDefault()
	m.RequestsDefault()
	m.LimitsDefault()
	m.ResizePolicyDefault()
	m.CurrentRequestsDefault()
	m.CurrentLimitsDefault()
}
