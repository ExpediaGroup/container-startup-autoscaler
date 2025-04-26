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

package kubetest

import (
	"context"
	"strings"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type MockPodHelper struct {
	mock.Mock
}

func NewMockPodHelper(configFunc func(*MockPodHelper)) *MockPodHelper {
	m := &MockPodHelper{}
	if configFunc != nil {
		configFunc(m)
	} else {
		m.AllDefaults()
	}

	return m
}

func (m *MockPodHelper) Get(ctx context.Context, name types.NamespacedName) (bool, *v1.Pod, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Get(1).(*v1.Pod), args.Error(2)
}

func (m *MockPodHelper) Patch(
	ctx context.Context,
	pod *v1.Pod,
	podMutationFuncs []func(*v1.Pod) (bool, error),
	patchResize bool,
	mustSyncCache bool,
) (*v1.Pod, error) {
	args := m.Called(ctx, pod, podMutationFuncs, patchResize, mustSyncCache)
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

func (m *MockPodHelper) ResizeConditions(pod *v1.Pod) []v1.PodCondition {
	args := m.Called(pod)
	return args.Get(0).([]v1.PodCondition)
}

func (m *MockPodHelper) QOSClass(pod *v1.Pod) (v1.PodQOSClass, error) {
	args := m.Called(pod)
	return args.Get(0).(v1.PodQOSClass), args.Error(1)
}

func (m *MockPodHelper) GetDefault() {
	m.On("Get", mock.Anything, mock.Anything).Return(true, &v1.Pod{}, nil)
}

func (m *MockPodHelper) PatchDefault() {
	m.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		&v1.Pod{},
		nil,
	)
}

func (m *MockPodHelper) HasAnnotationDefault() {
	m.On("HasAnnotation", mock.Anything, mock.Anything).Return(true, "")
}

func (m *MockPodHelper) ExpectedLabelValueAsDefault() {
	enabledMatchFunc := func(label string) bool {
		return strings.Contains(label, kubecommon.LabelEnabled)
	}

	m.On("ExpectedLabelValueAs", mock.Anything, mock.MatchedBy(enabledMatchFunc), kubecommon.DataTypeBool).
		Return(true, nil)
}

func (m *MockPodHelper) ExpectedAnnotationValueAsDefault() {
	targetContainerNameMatchFunc := func(ann string) bool {
		return strings.Contains(ann, scalecommon.AnnotationTargetContainerName)
	}

	cpuStartupMatchFunc := func(ann string) bool {
		return strings.Contains(ann, scalecommon.AnnotationCpuStartup)
	}

	cpuPostStartupRequestsMatchFunc := func(ann string) bool {
		return strings.Contains(ann, scalecommon.AnnotationCpuPostStartupRequests)
	}

	cpuPostStartupLimitsMatchFunc := func(ann string) bool {
		return strings.Contains(ann, scalecommon.AnnotationCpuPostStartupLimits)
	}

	memoryStartupMatchFunc := func(ann string) bool {
		return strings.Contains(ann, scalecommon.AnnotationMemoryStartup)
	}

	memoryPostStartupRequestsMatchFunc := func(ann string) bool {
		return strings.Contains(ann, scalecommon.AnnotationMemoryPostStartupRequests)
	}

	memoryPostStartupLimitsMatchFunc := func(ann string) bool {
		return strings.Contains(ann, scalecommon.AnnotationMemoryPostStartupLimits)
	}

	m.On("ExpectedAnnotationValueAs", mock.Anything, mock.MatchedBy(targetContainerNameMatchFunc), kubecommon.DataTypeString).
		Return(DefaultContainerName, nil)

	m.On("ExpectedAnnotationValueAs", mock.Anything, mock.MatchedBy(cpuStartupMatchFunc), kubecommon.DataTypeString).
		Return(PodAnnotationCpuStartup, nil)

	m.On("ExpectedAnnotationValueAs", mock.Anything, mock.MatchedBy(cpuPostStartupRequestsMatchFunc), kubecommon.DataTypeString).
		Return(PodAnnotationCpuPostStartupRequests, nil)

	m.On("ExpectedAnnotationValueAs", mock.Anything, mock.MatchedBy(cpuPostStartupLimitsMatchFunc), kubecommon.DataTypeString).
		Return(PodAnnotationCpuPostStartupLimits, nil)

	m.On("ExpectedAnnotationValueAs", mock.Anything, mock.MatchedBy(memoryStartupMatchFunc), kubecommon.DataTypeString).
		Return(PodAnnotationMemoryStartup, nil)

	m.On("ExpectedAnnotationValueAs", mock.Anything, mock.MatchedBy(memoryPostStartupRequestsMatchFunc), kubecommon.DataTypeString).
		Return(PodAnnotationMemoryPostStartupRequests, nil)

	m.On("ExpectedAnnotationValueAs", mock.Anything, mock.MatchedBy(memoryPostStartupLimitsMatchFunc), kubecommon.DataTypeString).
		Return(PodAnnotationMemoryPostStartupLimits, nil)
}

func (m *MockPodHelper) IsContainerInSpecDefault() {
	m.On("IsContainerInSpec", mock.Anything, mock.Anything).Return(true)
}

func (m *MockPodHelper) ResizeConditionsDefault() {
	m.On("ResizeConditions", mock.Anything).Return([]v1.PodCondition{}, nil)
}

func (m *MockPodHelper) QOSClassDefault() {
	m.On("QOSClass", mock.Anything).Return(v1.PodQOSGuaranteed, nil)
}

func (m *MockPodHelper) AllDefaults() {
	m.GetDefault()
	m.PatchDefault()
	m.HasAnnotationDefault()
	m.ExpectedLabelValueAsDefault()
	m.ExpectedAnnotationValueAsDefault()
	m.IsContainerInSpecDefault()
	m.ResizeConditionsDefault()
	m.QOSClassDefault()
}
