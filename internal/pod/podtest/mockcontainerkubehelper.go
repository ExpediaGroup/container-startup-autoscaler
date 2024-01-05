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
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// MockContainerKubeHelper is a generic mock for pod.ContainerKubeHelper.
type MockContainerKubeHelper struct {
	mock.Mock
}

func NewMockContainerKubeHelper(configFunc func(*MockContainerKubeHelper)) *MockContainerKubeHelper {
	mockHelper := &MockContainerKubeHelper{}
	configFunc(mockHelper)
	return mockHelper
}

func (m *MockContainerKubeHelper) Get(pod *v1.Pod, containerName string) (*v1.Container, error) {
	args := m.Called(pod, containerName)
	return args.Get(0).(*v1.Container), args.Error(1)
}

func (m *MockContainerKubeHelper) HasStartupProbe(container *v1.Container) bool {
	args := m.Called(container)
	return args.Bool(0)
}

func (m *MockContainerKubeHelper) HasReadinessProbe(container *v1.Container) bool {
	args := m.Called(container)
	return args.Bool(0)
}

func (m *MockContainerKubeHelper) State(pod *v1.Pod, containerName string) (v1.ContainerState, error) {
	args := m.Called(pod, containerName)
	return args.Get(0).(v1.ContainerState), args.Error(1)
}

func (m *MockContainerKubeHelper) IsStarted(pod *v1.Pod, containerName string) (bool, error) {
	args := m.Called(pod, containerName)
	return args.Bool(0), args.Error(1)
}

func (m *MockContainerKubeHelper) IsReady(pod *v1.Pod, containerName string) (bool, error) {
	args := m.Called(pod, containerName)
	return args.Bool(0), args.Error(1)
}

func (m *MockContainerKubeHelper) Requests(container *v1.Container, resourceName v1.ResourceName) resource.Quantity {
	args := m.Called(container, resourceName)
	return args.Get(0).(resource.Quantity)
}

func (m *MockContainerKubeHelper) Limits(container *v1.Container, resourceName v1.ResourceName) resource.Quantity {
	args := m.Called(container, resourceName)
	return args.Get(0).(resource.Quantity)
}

func (m *MockContainerKubeHelper) ResizePolicy(
	container *v1.Container,
	resourceName v1.ResourceName,
) (v1.ResourceResizeRestartPolicy, error) {
	args := m.Called(container, resourceName)
	return args.Get(0).(v1.ResourceResizeRestartPolicy), args.Error(1)
}

func (m *MockContainerKubeHelper) AllocatedResources(
	pod *v1.Pod,
	containerName string,
	resourceName v1.ResourceName,
) (resource.Quantity, error) {
	args := m.Called(pod, containerName, resourceName)
	return args.Get(0).(resource.Quantity), args.Error(1)
}

func (m *MockContainerKubeHelper) CurrentRequests(
	pod *v1.Pod,
	containerName string,
	resourceName v1.ResourceName,
) (resource.Quantity, error) {
	args := m.Called(pod, containerName, resourceName)
	return args.Get(0).(resource.Quantity), args.Error(1)
}

func (m *MockContainerKubeHelper) CurrentLimits(
	pod *v1.Pod,
	containerName string,
	resourceName v1.ResourceName,
) (resource.Quantity, error) {
	args := m.Called(pod, containerName, resourceName)
	return args.Get(0).(resource.Quantity), args.Error(1)
}

func (m *MockContainerKubeHelper) GetDefault() {
	m.On("Get", mock.Anything, mock.Anything).Return(&v1.Container{}, nil)
}

func (m *MockContainerKubeHelper) HasStartupProbeDefault() {
	m.On("HasStartupProbe", mock.Anything).Return(true)
}

func (m *MockContainerKubeHelper) HasReadinessProbeDefault() {
	m.On("HasReadinessProbe", mock.Anything).Return(true)
}

func (m *MockContainerKubeHelper) StateDefault() {
	m.On("State", mock.Anything, mock.Anything).Return(v1.ContainerState{Running: &v1.ContainerStateRunning{}}, nil)
}

func (m *MockContainerKubeHelper) IsStartedDefault() {
	m.On("IsStarted", mock.Anything, mock.Anything).Return(true, nil)
}

func (m *MockContainerKubeHelper) IsReadyDefault() {
	m.On("IsReady", mock.Anything, mock.Anything).Return(true, nil)
}

func (m *MockContainerKubeHelper) RequestsDefault() {
	m.On("Requests", mock.Anything, v1.ResourceCPU).Return(MockDefaultCpuQuantity)
	m.On("Requests", mock.Anything, v1.ResourceMemory).Return(MockDefaultMemoryQuantity)
}

func (m *MockContainerKubeHelper) LimitsDefault() {
	m.On("Limits", mock.Anything, v1.ResourceCPU).Return(MockDefaultCpuQuantity)
	m.On("Limits", mock.Anything, v1.ResourceMemory).Return(MockDefaultMemoryQuantity)
}

func (m *MockContainerKubeHelper) ResizePolicyDefault() {
	m.On("ResizePolicy", mock.Anything, mock.Anything).Return(v1.NotRequired, nil)
}

func (m *MockContainerKubeHelper) AllocatedResourcesDefault() {
	m.On("AllocatedResources", mock.Anything, mock.Anything, v1.ResourceCPU).Return(MockDefaultCpuQuantity, nil)
	m.On("AllocatedResources", mock.Anything, mock.Anything, v1.ResourceMemory).Return(MockDefaultMemoryQuantity, nil)
}

func (m *MockContainerKubeHelper) CurrentRequestsDefault() {
	m.On("CurrentRequests", mock.Anything, mock.Anything, v1.ResourceCPU).Return(MockDefaultCpuQuantity, nil)
	m.On("CurrentRequests", mock.Anything, mock.Anything, v1.ResourceMemory).Return(MockDefaultMemoryQuantity, nil)
}

func (m *MockContainerKubeHelper) CurrentLimitsDefault() {
	m.On("CurrentLimits", mock.Anything, mock.Anything, v1.ResourceCPU).Return(MockDefaultCpuQuantity, nil)
	m.On("CurrentLimits", mock.Anything, mock.Anything, v1.ResourceMemory).Return(MockDefaultMemoryQuantity, nil)
}
