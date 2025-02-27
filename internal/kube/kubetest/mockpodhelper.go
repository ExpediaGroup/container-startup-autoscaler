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
	"context"
	"strings"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// MockPodHelper is a generic mock for pod.PodHelper.
type MockPodHelper struct {
	mock.Mock
}

func NewMockPodHelper(configFunc func(*MockPodHelper)) *MockPodHelper {
	mockHelper := &MockPodHelper{}
	configFunc(mockHelper)
	return mockHelper
}

func (m *MockPodHelper) Get(ctx context.Context, name types.NamespacedName) (bool, *v1.Pod, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Get(1).(*v1.Pod), args.Error(2)
}

func (m *MockPodHelper) Patch(
	ctx context.Context,
	originalPod *v1.Pod,
	mutatePodFunc func(*v1.Pod) (bool, *v1.Pod, error),
	patchResize bool,
	mustSyncCache bool,
) (*v1.Pod, error) {
	args := m.Called(ctx, originalPod, mutatePodFunc, patchResize, mustSyncCache)
	return args.Get(0).(*v1.Pod), args.Error(1)
}

func (m *MockPodHelper) UpdateContainerResources(
	ctx context.Context,
	pod *v1.Pod,
	addPodMutationFunc func(pod *v1.Pod) (bool, *v1.Pod, error),
	addPodMutationMustSyncCache bool,
) (*v1.Pod, error) {
	args := m.Called(ctx, pod, addPodMutationFunc, addPodMutationMustSyncCache)
	return args.Get(0).(*v1.Pod), args.Error(1)
}

func (m *MockPodHelper) HasAnnotation(pod *v1.Pod, name string) (bool, string) {
	args := m.Called(pod, name)
	return args.Bool(0), args.String(1)
}

func (m *MockPodHelper) ExpectedLabelValueAs(pod *v1.Pod, name string, as kubecommon.DataType) (any, error) {
	args := m.Called(pod, name, as)
	return args.Get(0), args.Error(1)
}

func (m *MockPodHelper) ExpectedAnnotationValueAs(pod *v1.Pod, name string, as kubecommon.DataType) (any, error) {
	args := m.Called(pod, name, as)
	return args.Get(0), args.Error(1)
}

func (m *MockPodHelper) IsContainerInSpec(pod *v1.Pod, containerName string) bool {
	args := m.Called(pod, containerName)
	return args.Bool(0)
}

func (m *MockPodHelper) ResizeStatus(pod *v1.Pod) v1.PodResizeStatus {
	args := m.Called(pod)
	return args.Get(0).(v1.PodResizeStatus)
}

func (m *MockPodHelper) GetDefault() {
	m.On("Get", mock.Anything, mock.Anything).Return(true, &v1.Pod{}, nil)
}

func (m *MockPodHelper) PatchDefault() {
	m.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&v1.Pod{}, nil)
}

func (m *MockPodHelper) UpdateContainerResourcesDefault() {
	m.On("UpdateContainerResources", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&v1.Pod{}, nil)
}

func (m *MockPodHelper) HasAnnotationDefault() {
	m.On("HasAnnotation", mock.Anything, mock.Anything).Return(true, "")
}

func (m *MockPodHelper) ExpectedLabelValueAsDefault() {
	enabledMatchFunc := func(label string) bool {
		return strings.Contains(label, podcommon.LabelEnabled)
	}

	m.On("ExpectedLabelValueAs", mock.Anything, mock.MatchedBy(enabledMatchFunc), kubecommon.DataTypeBool).
		Return(true, nil)
}

func (m *MockPodHelper) ExpectedAnnotationValueAsDefault() {
	targetContainerNameMatchFunc := func(ann string) bool {
		return strings.Contains(ann, scalecommon.AnnotationTargetContainerName)
	}

	cpuMatchFunc := func(ann string) bool {
		return strings.Contains(ann, "cpu")
	}

	memoryMatchFunc := func(ann string) bool {
		return strings.Contains(ann, "memory")
	}

	m.On("ExpectedAnnotationValueAs", mock.Anything, mock.MatchedBy(targetContainerNameMatchFunc), kubecommon.DataTypeString).
		Return(DefaultContainerName, nil)
	m.On("ExpectedAnnotationValueAs", mock.Anything, mock.MatchedBy(cpuMatchFunc), kubecommon.DataTypeString).
		Return(MockDefaultCpu, nil)
	m.On("ExpectedAnnotationValueAs", mock.Anything, mock.MatchedBy(memoryMatchFunc), kubecommon.DataTypeString).
		Return(MockDefaultMemory, nil)
}

func (m *MockPodHelper) IsContainerInSpecDefault() {
	m.On("IsContainerInSpec", mock.Anything, mock.Anything).Return(true)
}

func (m *MockPodHelper) ResizeStatusDefault() {
	m.On("ResizeStatus", mock.Anything).Return(v1.PodResizeStatus(""), nil)
}

func (m *MockPodHelper) AllDefaults() {
	m.GetDefault()
	m.PatchDefault()
	m.UpdateContainerResourcesDefault()
	m.HasAnnotationDefault()
	m.ExpectedLabelValueAsDefault()
	m.ExpectedAnnotationValueAsDefault()
	m.IsContainerInSpecDefault()
	m.ResizeStatusDefault()
}
