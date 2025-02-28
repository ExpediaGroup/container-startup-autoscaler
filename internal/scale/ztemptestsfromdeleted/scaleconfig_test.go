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

package ztemptestsfromdeleted

//import (
//	"errors"
//	"fmt"
//	"testing"
//
//	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
//	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podtest"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/mock"
//	"k8s.io/api/core/v1"
//	"k8s.io/apimachinery/pkg/api/resource"
//)
//
//func TestNewScaleConfig(t *testing.T) {
//	config := NewScaleConfig(nil)
//	assert.Empty(t, config.GetTargetContainerName())
//	assert.Empty(t, config.GetCpuConfig())
//	assert.Empty(t, config.GetMemoryConfig())
//}
//
//func TestScaleConfigGetTargetContainerName(t *testing.T) {
//	config := &scaleConfig{targetContainerName: podtest.DefaultContainerName}
//	assert.Equal(t, podtest.DefaultContainerName, config.GetTargetContainerName())
//}
//
//func TestScaleConfigGetCpuConfig(t *testing.T) {
//	cConfig := podcommon.NewCpuConfig(podtest.MockDefaultCpuQuantity, podtest.MockDefaultCpuQuantity, podtest.MockDefaultCpuQuantity)
//	config := &scaleConfig{cpuConfig: cConfig}
//	assert.Equal(t, cConfig, config.GetCpuConfig())
//}
//
//func TestScaleConfigGetMemoryConfig(t *testing.T) {
//	mConfig := podcommon.NewMemoryConfig(podtest.MockDefaultMemoryQuantity, podtest.MockDefaultMemoryQuantity, podtest.MockDefaultMemoryQuantity)
//	config := &scaleConfig{memoryConfig: mConfig}
//	assert.Equal(t, mConfig, config.GetMemoryConfig())
//}
//
//func TestScaleConfigStoreFromAnnotations(t *testing.T) {
//	type want struct {
//		targetContainerName        string
//		cpuConfig                  podcommon.CpuConfig
//		memoryConfig               podcommon.MemoryConfig
//		hasAssignedFromAnnotations bool
//	}
//	tests := []struct {
//		name           string
//		configMockFunc func(*podtest.MockPodHelper)
//		want           *want
//		wantErrMsg     string
//	}{
//		{
//			name: "UnableToGet" + podcommon.AnnotationTargetContainerName,
//			configMockFunc: func(m *podtest.MockPodHelper) {
//				m.On("ExpectedAnnotationValueAs", mock.Anything, podcommon.AnnotationTargetContainerName, podcommon.TypeString).
//					Return(nil, errors.New(""))
//				m.ExpectedAnnotationValueAsDefault()
//			},
//			wantErrMsg: fmt.Sprintf("unable to get '%s' annotation value", podcommon.AnnotationTargetContainerName),
//		},
//		{
//			name: "UnableToGet" + podcommon.AnnotationCpuStartup,
//			configMockFunc: func(m *podtest.MockPodHelper) {
//				m.On("ExpectedAnnotationValueAs", mock.Anything, podcommon.AnnotationCpuStartup, podcommon.TypeString).
//					Return("", errors.New(""))
//				m.ExpectedAnnotationValueAsDefault()
//			},
//			wantErrMsg: fmt.Sprintf("unable to get '%s' annotation value", podcommon.AnnotationCpuStartup),
//		},
//		{
//			name: "UnableToParse" + podcommon.AnnotationCpuStartup,
//			configMockFunc: func(m *podtest.MockPodHelper) {
//				m.On("ExpectedAnnotationValueAs", mock.Anything, podcommon.AnnotationCpuStartup, podcommon.TypeString).
//					Return("test", nil)
//				m.ExpectedAnnotationValueAsDefault()
//			},
//			wantErrMsg: fmt.Sprintf("unable to parse '%s' annotation value ('test')", podcommon.AnnotationCpuStartup),
//		},
//		{
//			name: "UnableToGet" + podcommon.AnnotationCpuPostStartupRequests,
//			configMockFunc: func(m *podtest.MockPodHelper) {
//				m.On("ExpectedAnnotationValueAs", mock.Anything, podcommon.AnnotationCpuPostStartupRequests, podcommon.TypeString).
//					Return("", errors.New(""))
//				m.ExpectedAnnotationValueAsDefault()
//			},
//			wantErrMsg: fmt.Sprintf("unable to get '%s' annotation value", podcommon.AnnotationCpuPostStartupRequests),
//		},
//		{
//			name: "UnableToParse" + podcommon.AnnotationCpuPostStartupRequests,
//			configMockFunc: func(m *podtest.MockPodHelper) {
//				m.On("ExpectedAnnotationValueAs", mock.Anything, podcommon.AnnotationCpuPostStartupRequests, podcommon.TypeString).
//					Return("test", nil)
//				m.ExpectedAnnotationValueAsDefault()
//			},
//			wantErrMsg: fmt.Sprintf("unable to parse '%s' annotation value ('test')", podcommon.AnnotationCpuPostStartupRequests),
//		},
//		{
//			name: "UnableToGet" + podcommon.AnnotationCpuPostStartupLimits,
//			configMockFunc: func(m *podtest.MockPodHelper) {
//				m.On("ExpectedAnnotationValueAs", mock.Anything, podcommon.AnnotationCpuPostStartupLimits, podcommon.TypeString).
//					Return("", errors.New(""))
//				m.ExpectedAnnotationValueAsDefault()
//			},
//			wantErrMsg: fmt.Sprintf("unable to get '%s' annotation value", podcommon.AnnotationCpuPostStartupLimits),
//		},
//		{
//			name: "UnableToParse" + podcommon.AnnotationCpuPostStartupLimits,
//			configMockFunc: func(m *podtest.MockPodHelper) {
//				m.On("ExpectedAnnotationValueAs", mock.Anything, podcommon.AnnotationCpuPostStartupLimits, podcommon.TypeString).
//					Return("test", nil)
//				m.ExpectedAnnotationValueAsDefault()
//			},
//			wantErrMsg: fmt.Sprintf("unable to parse '%s' annotation value ('test')", podcommon.AnnotationCpuPostStartupLimits),
//		},
//		{
//			name: "UnableToGet" + podcommon.AnnotationMemoryStartup,
//			configMockFunc: func(m *podtest.MockPodHelper) {
//				m.On("ExpectedAnnotationValueAs", mock.Anything, podcommon.AnnotationMemoryStartup, podcommon.TypeString).
//					Return("", errors.New(""))
//				m.ExpectedAnnotationValueAsDefault()
//			},
//			wantErrMsg: fmt.Sprintf("unable to get '%s' annotation value", podcommon.AnnotationMemoryStartup),
//		},
//		{
//			name: "UnableToParse" + podcommon.AnnotationMemoryStartup,
//			configMockFunc: func(m *podtest.MockPodHelper) {
//				m.On("ExpectedAnnotationValueAs", mock.Anything, podcommon.AnnotationMemoryStartup, podcommon.TypeString).
//					Return("test", nil)
//				m.ExpectedAnnotationValueAsDefault()
//			},
//			wantErrMsg: fmt.Sprintf("unable to parse '%s' annotation value ('test')", podcommon.AnnotationMemoryStartup),
//		},
//		{
//			name: "UnableToGet" + podcommon.AnnotationMemoryPostStartupRequests,
//			configMockFunc: func(m *podtest.MockPodHelper) {
//				m.On("ExpectedAnnotationValueAs", mock.Anything, podcommon.AnnotationMemoryPostStartupRequests, podcommon.TypeString).
//					Return("", errors.New(""))
//				m.ExpectedAnnotationValueAsDefault()
//			},
//			wantErrMsg: fmt.Sprintf("unable to get '%s' annotation value", podcommon.AnnotationMemoryPostStartupRequests),
//		},
//		{
//			name: "UnableToParse" + podcommon.AnnotationMemoryPostStartupRequests,
//			configMockFunc: func(m *podtest.MockPodHelper) {
//				m.On("ExpectedAnnotationValueAs", mock.Anything, podcommon.AnnotationMemoryPostStartupRequests, podcommon.TypeString).
//					Return("test", nil)
//				m.ExpectedAnnotationValueAsDefault()
//			},
//			wantErrMsg: fmt.Sprintf("unable to parse '%s' annotation value ('test')", podcommon.AnnotationMemoryPostStartupRequests),
//		},
//		{
//			name: "UnableToGet" + podcommon.AnnotationMemoryPostStartupLimits,
//			configMockFunc: func(m *podtest.MockPodHelper) {
//				m.On("ExpectedAnnotationValueAs", mock.Anything, podcommon.AnnotationMemoryPostStartupLimits, podcommon.TypeString).
//					Return("", errors.New(""))
//				m.ExpectedAnnotationValueAsDefault()
//			},
//			wantErrMsg: fmt.Sprintf("unable to get '%s' annotation value", podcommon.AnnotationMemoryPostStartupLimits),
//		},
//		{
//			name: "UnableToParse" + podcommon.AnnotationMemoryPostStartupLimits,
//			configMockFunc: func(m *podtest.MockPodHelper) {
//				m.On("ExpectedAnnotationValueAs", mock.Anything, podcommon.AnnotationMemoryPostStartupLimits, podcommon.TypeString).
//					Return("test", nil)
//				m.ExpectedAnnotationValueAsDefault()
//			},
//			wantErrMsg: fmt.Sprintf("unable to parse '%s' annotation value ('test')", podcommon.AnnotationMemoryPostStartupLimits),
//		},
//		{
//			name: "Ok",
//			configMockFunc: func(m *podtest.MockPodHelper) {
//				m.ExpectedAnnotationValueAsDefault()
//			},
//			want: &want{
//				targetContainerName:        podtest.DefaultContainerName,
//				cpuConfig:                  podcommon.NewCpuConfig(podtest.MockDefaultCpuQuantity, podtest.MockDefaultCpuQuantity, podtest.MockDefaultCpuQuantity),
//				memoryConfig:               podcommon.NewMemoryConfig(podtest.MockDefaultMemoryQuantity, podtest.MockDefaultMemoryQuantity, podtest.MockDefaultMemoryQuantity),
//				hasAssignedFromAnnotations: true,
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			config := &scaleConfig{podHelper: podtest.NewMockPodHelper(tt.configMockFunc)}
//			err := config.StoreFromAnnotations(&v1.Pod{})
//
//			if tt.wantErrMsg != "" {
//				assert.Contains(t, err.Error(), tt.wantErrMsg)
//			} else {
//				assert.Nil(t, err)
//			}
//
//			if tt.want != nil {
//				assert.Equal(t, tt.want.targetContainerName, config.targetContainerName)
//				assert.Equal(t, tt.want.cpuConfig, config.cpuConfig)
//				assert.Equal(t, tt.want.memoryConfig, config.memoryConfig)
//				assert.Equal(t, tt.want.hasAssignedFromAnnotations, config.hasAssignedFromAnnotations)
//			}
//		})
//	}
//}
//
//func TestScaleConfigValidate(t *testing.T) {
//	tests := []struct {
//		name            string
//		ScaleConfig     *scaleConfig
//		wantPanicErrMsg string
//		wantErrMsg      string
//	}{
//		{
//			name:            "AssignFromAnnotationsNotInvokedFirst",
//			ScaleConfig:     &scaleConfig{},
//			wantPanicErrMsg: "StoreFromAnnotations() hasn't been invoked first",
//		},
//		{
//			name:        "NameEmpty",
//			ScaleConfig: &scaleConfig{hasAssignedFromAnnotations: true},
//			wantErrMsg:  "target container name is empty",
//		},
//		{
//			name: "CpuPostStartupRequestsMustEqualPostStartupRequests",
//			ScaleConfig: &scaleConfig{
//				targetContainerName: "test",
//				cpuConfig: podcommon.CpuConfig{
//					PostStartupRequests: resource.MustParse("1m"),
//					PostStartupLimits:   resource.MustParse("2m"),
//				},
//				hasAssignedFromAnnotations: true,
//			},
//			wantErrMsg: "cpu post-startup requests (1m) must equal post-startup limits (2m) - change in qos class is not yet permitted by kube",
//		},
//		{
//			name: "MemoryPostStartupRequestsMustEqualPostStartupRequests",
//			ScaleConfig: &scaleConfig{
//				targetContainerName: "test",
//				memoryConfig: podcommon.MemoryConfig{
//					PostStartupRequests: resource.MustParse("1M"),
//					PostStartupLimits:   resource.MustParse("2M"),
//				},
//				hasAssignedFromAnnotations: true,
//			},
//			wantErrMsg: "memory post-startup requests (1M) must equal post-startup limits (2M) - change in qos class is not yet permitted by kube",
//		},
//		{
//			name: "CpuPostStartupRequestsGreaterThanStartup",
//			ScaleConfig: &scaleConfig{
//				targetContainerName: "test",
//				cpuConfig: podcommon.CpuConfig{
//					Startup:             resource.MustParse("1m"),
//					PostStartupRequests: resource.MustParse("2m"),
//					PostStartupLimits:   resource.MustParse("2m"),
//				},
//				hasAssignedFromAnnotations: true,
//			},
//			wantErrMsg: "cpu post-startup requests (2m) is greater than startup value (1m)",
//		},
//		{
//			name: "MemoryPostStartupRequestsGreaterThanStartup",
//			ScaleConfig: &scaleConfig{
//				targetContainerName: "test",
//				memoryConfig: podcommon.MemoryConfig{
//					Startup:             resource.MustParse("1M"),
//					PostStartupRequests: resource.MustParse("2M"),
//					PostStartupLimits:   resource.MustParse("2M"),
//				},
//				hasAssignedFromAnnotations: true,
//			},
//			wantErrMsg: "memory post-startup requests (2M) is greater than startup value (1M)",
//		},
//		// TODO(wt) reinstate once change in qos class is permitted by Kube (get rid of this)
//		//{
//		//	name: "CpuPostStartupLimitsLessThanRequests",
//		//	ScaleConfig: &scaleConfig{
//		//		targetContainerName: "test",
//		//		cpuConfig: podcommon.CpuConfig{
//		//			Startup:             resource.MustParse("2m"),
//		//			PostStartupRequests: resource.MustParse("2m"),
//		//			PostStartupLimits:   resource.MustParse("1m"),
//		//		},
//		//		hasAssignedFromAnnotations: true,
//		//	},
//		//	wantErrMsg: "cpu post-startup limits (1m) is less than post-startup requests (2m)",
//		//},
//		//{
//		//	name: "MemoryPostStartupLimitsLessThanRequests",
//		//	ScaleConfig: &scaleConfig{
//		//		targetContainerName: "test",
//		//		cpuConfig: podcommon.CpuConfig{
//		//			Startup:             resource.MustParse("1m"),
//		//			PostStartupRequests: resource.MustParse("1m"),
//		//			PostStartupLimits:   resource.MustParse("1m"),
//		//		},
//		//		memoryConfig: podcommon.MemoryConfig{
//		//			Startup:             resource.MustParse("2M"),
//		//			PostStartupRequests: resource.MustParse("2M"),
//		//			PostStartupLimits:   resource.MustParse("1M"),
//		//		},
//		//		hasAssignedFromAnnotations: true,
//		//	},
//		//	wantErrMsg: "memory post-startup limits (1M) is less than post-startup requests (2M)",
//		//},
//		{
//			name: "Ok",
//			ScaleConfig: &scaleConfig{
//				targetContainerName: "test",
//				cpuConfig: podcommon.CpuConfig{
//					Startup:             resource.MustParse("1m"),
//					PostStartupRequests: resource.MustParse("1m"),
//					PostStartupLimits:   resource.MustParse("1m"),
//				},
//				memoryConfig: podcommon.MemoryConfig{
//					Startup:             resource.MustParse("1M"),
//					PostStartupRequests: resource.MustParse("1M"),
//					PostStartupLimits:   resource.MustParse("1M"),
//				},
//				hasAssignedFromAnnotations: true,
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if tt.wantPanicErrMsg != "" {
//				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() { _ = tt.ScaleConfig.Validate() })
//				return
//			}
//
//			err := tt.ScaleConfig.Validate()
//			if tt.wantErrMsg != "" {
//				assert.Contains(t, err.Error(), tt.wantErrMsg)
//			} else {
//				assert.Nil(t, err)
//			}
//		})
//	}
//}
//
//func TestScaleConfigString(t *testing.T) {
//	s := &scaleConfig{
//		targetContainerName: podtest.DefaultContainerName,
//		cpuConfig: podcommon.NewCpuConfig(
//			podtest.PodAnnotationCpuStartupQuantity,
//			podtest.PodAnnotationCpuPostStartupRequestsQuantity,
//			podtest.PodAnnotationCpuPostStartupLimitsQuantity,
//		),
//		memoryConfig: podcommon.NewMemoryConfig(
//			podtest.PodAnnotationMemoryStartupQuantity,
//			podtest.PodAnnotationMemoryPostStartupRequestsQuantity,
//			podtest.PodAnnotationMemoryPostStartupLimitsQuantity,
//		),
//	}
//
//	expected := fmt.Sprintf(
//		"cpu: %s/%s/%s, memory: %s/%s/%s",
//		podtest.PodAnnotationCpuStartupQuantity.String(),
//		podtest.PodAnnotationCpuPostStartupRequestsQuantity.String(),
//		podtest.PodAnnotationCpuPostStartupLimitsQuantity.String(),
//		podtest.PodAnnotationMemoryStartupQuantity.String(),
//		podtest.PodAnnotationMemoryPostStartupRequestsQuantity.String(),
//		podtest.PodAnnotationMemoryPostStartupLimitsQuantity.String(),
//	)
//	assert.Equal(t, expected, s.String())
//}
