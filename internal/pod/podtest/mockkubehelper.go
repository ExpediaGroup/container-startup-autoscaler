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
	"context"
	"strings"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
)

// MockKubeHelper is a generic mock for pod.KubeHelper.
type MockKubeHelper struct {
	mock.Mock
}

func NewMockKubeHelper(configFunc func(*MockKubeHelper)) *MockKubeHelper {
	mockHelper := &MockKubeHelper{}
	configFunc(mockHelper)
	return mockHelper
}

func (m *MockKubeHelper) Get(ctx context.Context, name types.NamespacedName) (bool, *v1.Pod, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Get(1).(*v1.Pod), args.Error(2)
}

func (m *MockKubeHelper) Patch(
	ctx context.Context,
	originalPod *v1.Pod,
	mutatePodFunc func(*v1.Pod) (bool, *v1.Pod, error),
	patchResize bool,
	mustSyncCache bool,
) (*v1.Pod, error) {
	args := m.Called(ctx, originalPod, mutatePodFunc, patchResize, mustSyncCache)
	return args.Get(0).(*v1.Pod), args.Error(1)
}

func (m *MockKubeHelper) UpdateContainerResources(
	ctx context.Context,
	pod *v1.Pod,
	containerName string,
	cpuRequests resource.Quantity, cpuLimits resource.Quantity,
	memoryRequests resource.Quantity, memoryLimits resource.Quantity,
	addPodMutationFunc func(pod *v1.Pod) (bool, *v1.Pod, error),
	addPodMutationMustSyncCache bool,
) (*v1.Pod, error) {
	args := m.Called(ctx, pod, containerName, cpuRequests, cpuLimits, memoryRequests, memoryLimits, addPodMutationFunc, addPodMutationMustSyncCache)
	return args.Get(0).(*v1.Pod), args.Error(1)
}

func (m *MockKubeHelper) HasAnnotation(pod *v1.Pod, name string) (bool, string) {
	args := m.Called(pod, name)
	return args.Bool(0), args.String(1)
}

func (m *MockKubeHelper) ExpectedLabelValueAs(pod *v1.Pod, name string, as podcommon.Type) (any, error) {
	args := m.Called(pod, name, as)
	return args.Get(0), args.Error(1)
}

func (m *MockKubeHelper) ExpectedAnnotationValueAs(pod *v1.Pod, name string, as podcommon.Type) (any, error) {
	args := m.Called(pod, name, as)
	return args.Get(0), args.Error(1)
}

func (m *MockKubeHelper) IsContainerInSpec(pod *v1.Pod, containerName string) bool {
	args := m.Called(pod, containerName)
	return args.Bool(0)
}

func (m *MockKubeHelper) ResizeStatus(pod *v1.Pod) v1.PodResizeStatus {
	args := m.Called(pod)
	return args.Get(0).(v1.PodResizeStatus)
}

func (m *MockKubeHelper) GetDefault() {
	m.On("Get", mock.Anything, mock.Anything).Return(true, &v1.Pod{}, nil)
}

func (m *MockKubeHelper) PatchDefault() {
	m.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&v1.Pod{}, nil)
}

func (m *MockKubeHelper) UpdateContainerResourcesDefault() {
	m.On("UpdateContainerResources", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&v1.Pod{}, nil)
}

func (m *MockKubeHelper) HasAnnotationDefault() {
	m.On("HasAnnotation", mock.Anything, mock.Anything).Return(true, "")
}

func (m *MockKubeHelper) ExpectedLabelValueAsDefault() {
	enabledMatchFunc := func(label string) bool {
		return strings.Contains(label, podcommon.LabelEnabled)
	}

	m.On("ExpectedLabelValueAs", mock.Anything, mock.MatchedBy(enabledMatchFunc), podcommon.TypeBool).
		Return(true, nil)
}

func (m *MockKubeHelper) ExpectedAnnotationValueAsDefault() {
	targetContainerNameMatchFunc := func(ann string) bool {
		return strings.Contains(ann, podcommon.AnnotationTargetContainerName)
	}

	cpuMatchFunc := func(ann string) bool {
		return strings.Contains(ann, "cpu")
	}

	memoryMatchFunc := func(ann string) bool {
		return strings.Contains(ann, "memory")
	}

	m.On("ExpectedAnnotationValueAs", mock.Anything, mock.MatchedBy(targetContainerNameMatchFunc), podcommon.TypeString).
		Return(DefaultContainerName, nil)
	m.On("ExpectedAnnotationValueAs", mock.Anything, mock.MatchedBy(cpuMatchFunc), podcommon.TypeString).
		Return(MockDefaultCpu, nil)
	m.On("ExpectedAnnotationValueAs", mock.Anything, mock.MatchedBy(memoryMatchFunc), podcommon.TypeString).
		Return(MockDefaultMemory, nil)
}

func (m *MockKubeHelper) IsContainerInSpecDefault() {
	m.On("IsContainerInSpec", mock.Anything, mock.Anything).Return(true)
}

func (m *MockKubeHelper) ResizeStatusDefault() {
	m.On("ResizeStatus", mock.Anything).Return(v1.PodResizeStatus(""), nil)
}
