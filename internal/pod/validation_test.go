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

package pod

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contexttest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/tools/record"
)

func TestNewValidation(t *testing.T) {
	recorder := &record.FakeRecorder{}
	helper := newKubeHelper(nil)
	cHelper := newContainerKubeHelper()
	stat := newStatus(helper)
	val := newValidation(recorder, stat, helper, cHelper)
	assert.Equal(t, recorder, val.recorder)
	assert.Equal(t, stat, val.status)
	assert.Equal(t, helper, val.kubeHelper)
	assert.Equal(t, cHelper, val.containerKubeHelper)
}

func TestValidationValidate(t *testing.T) {
	tests := []struct {
		name                      string
		configStatusMockFunc      func(*podtest.MockStatus, func())
		configHelperMockFunc      func(*podtest.MockKubeHelper)
		configContHelperMockFunc  func(*podtest.MockContainerKubeHelper)
		configScaleConfigMockFunc func(*podtest.MockScaleConfig)
		wantErrMsg                string
		wantStatusUpdate          bool
		wantEventMsg              string
	}{
		{
			name: "UnableToGetEnabledLabelValue",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("ExpectedLabelValueAs", mock.Anything, podcommon.LabelEnabled, podcommon.TypeBool).
					Return(nil, errors.New(""))
			},
			configContHelperMockFunc:  func(m *podtest.MockContainerKubeHelper) {},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {},
			wantErrMsg:                "unable to get pod enabled label value",
			wantStatusUpdate:          true,
			wantEventMsg:              "Validation error: unable to get pod enabled label value",
		},
		{
			name: "EnabledLabelFalse",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("ExpectedLabelValueAs", mock.Anything, podcommon.LabelEnabled, podcommon.TypeBool).
					Return(false, nil)
			},
			configContHelperMockFunc:  func(m *podtest.MockContainerKubeHelper) {},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {},
			wantErrMsg:                "pod enabled label value is unexpectedly 'false'",
			wantStatusUpdate:          true,
			wantEventMsg:              "Validation error: pod enabled label value is unexpectedly 'false'",
		},
		{
			name: "VpaNotSupported",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("HasAnnotation", mock.Anything, podcommon.KnownVpaAnnotations[0]).Return(true, "")
				m.ExpectedLabelValueAsDefault()
			},
			configContHelperMockFunc:  func(m *podtest.MockContainerKubeHelper) {},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {},
			wantErrMsg:                "vpa not supported",
			wantStatusUpdate:          true,
			wantEventMsg:              "Validation error: vpa not supported",
		},
		{
			name: "UnableToGetAnnotationConfigValues",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
			},
			configContHelperMockFunc: func(m *podtest.MockContainerKubeHelper) {},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {
				m.On("StoreFromAnnotations", mock.Anything, mock.Anything).Return(errors.New(""))
			},
			wantErrMsg:       "unable to get annotation configuration values",
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: unable to get annotation configuration values",
		},
		{
			name: "UnableToValidateConfigValues",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
			},
			configContHelperMockFunc: func(m *podtest.MockContainerKubeHelper) {},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {
				m.On("Validate").Return(errors.New(""))
				m.StoreFromAnnotationsDefault()
				m.GetTargetContainerNameDefault()
			},
			wantErrMsg:       "unable to validate configuration values",
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: unable to validate configuration values",
		},
		{
			name: "TargetContainerNotInPodSpec",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.On("IsContainerInSpec", mock.Anything, podtest.DefaultContainerName).Return(false)
				m.ExpectedLabelValueAsDefault()
			},
			configContHelperMockFunc: func(m *podtest.MockContainerKubeHelper) {},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {
				m.StoreFromAnnotationsDefault()
				m.GetTargetContainerNameDefault()
				m.ValidateDefault()
			},
			wantErrMsg:       "target container not in pod spec",
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: target container not in pod spec",
		},
		{
			name: "TargetContainerNoProbes",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
			},
			configContHelperMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("HasStartupProbe", mock.Anything).Return(false)
				m.On("HasReadinessProbe", mock.Anything).Return(false)
				m.GetDefault()
			},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {
				m.StoreFromAnnotationsDefault()
				m.GetTargetContainerNameDefault()
				m.ValidateDefault()
			},
			wantErrMsg:       "target container does not specify startup probe or readiness probe",
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: target container does not specify startup probe or readiness probe",
		},
		{
			name: "TargetContainerNoCpuRequests",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
			},
			configContHelperMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("Requests", mock.Anything, v1.ResourceCPU).Return(resource.Quantity{})
				m.GetDefault()
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
			},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {
				m.StoreFromAnnotationsDefault()
				m.GetTargetContainerNameDefault()
				m.ValidateDefault()
			},
			wantErrMsg:       "target container does not specify cpu requests",
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: target container does not specify cpu requests",
		},
		{
			name: "TargetContainerNoMemoryRequests",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
			},
			configContHelperMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("Requests", mock.Anything, v1.ResourceMemory).Return(resource.Quantity{})
				m.GetDefault()
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.RequestsDefault()
			},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {
				m.StoreFromAnnotationsDefault()
				m.GetTargetContainerNameDefault()
				m.ValidateDefault()
			},
			wantErrMsg:       "target container does not specify memory requests",
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: target container does not specify memory requests",
		},
		{
			name: "TargetContainerCpuRequestsMustEqualLimits",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
			},
			configContHelperMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("Requests", mock.Anything, v1.ResourceCPU).Return(resource.MustParse("1m"))
				m.On("Limits", mock.Anything, v1.ResourceCPU).Return(resource.MustParse("2m"))
				m.GetDefault()
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.RequestsDefault()
			},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {
				m.StoreFromAnnotationsDefault()
				m.GetTargetContainerNameDefault()
				m.ValidateDefault()
			},
			wantErrMsg:       "target container cpu requests (1m) must equal limits (2m)",
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: target container cpu requests (1m) must equal limits (2m)",
		},
		{
			name: "TargetContainerMemoryRequestsMustEqualLimits",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
			},
			configContHelperMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("Requests", mock.Anything, v1.ResourceMemory).Return(resource.MustParse("1M"))
				m.On("Limits", mock.Anything, v1.ResourceMemory).Return(resource.MustParse("2M"))
				m.GetDefault()
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.RequestsDefault()
				m.LimitsDefault()
			},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {
				m.StoreFromAnnotationsDefault()
				m.GetTargetContainerNameDefault()
				m.ValidateDefault()
			},
			wantErrMsg:       "target container memory requests (1M) must equal limits (2M)",
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: target container memory requests (1M) must equal limits (2M)",
		},
		{
			name: "UnableToGetCpuResizePolicy",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
			},
			configContHelperMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("ResizePolicy", mock.Anything, v1.ResourceCPU).Return(v1.ResourceResizeRestartPolicy(""), errors.New(""))
				m.GetDefault()
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.RequestsDefault()
				m.LimitsDefault()
			},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {
				m.StoreFromAnnotationsDefault()
				m.GetTargetContainerNameDefault()
				m.ValidateDefault()
			},
			wantErrMsg:       "unable to get target container cpu resize policy",
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: unable to get target container cpu resize policy",
		},
		{
			name: "TargetContainerCpuResizePolicyIncorrect",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
			},
			configContHelperMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("ResizePolicy", mock.Anything, v1.ResourceCPU).Return(v1.RestartContainer, nil)
				m.GetDefault()
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.RequestsDefault()
				m.LimitsDefault()
			},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {
				m.StoreFromAnnotationsDefault()
				m.GetTargetContainerNameDefault()
				m.GetTargetContainerNameDefault()
				m.ValidateDefault()
			},
			wantErrMsg:       fmt.Sprintf("target container cpu resize policy is not '%s' ('%s')", v1.NotRequired, v1.RestartContainer),
			wantStatusUpdate: true,
			wantEventMsg:     fmt.Sprintf("Validation error: target container cpu resize policy is not '%s' ('%s')", v1.NotRequired, v1.RestartContainer),
		},
		{
			name: "UnableToGetMemoryResizePolicy",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
			},
			configContHelperMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("ResizePolicy", mock.Anything, v1.ResourceCPU).Return(v1.NotRequired, nil)
				m.On("ResizePolicy", mock.Anything, v1.ResourceMemory).Return(v1.ResourceResizeRestartPolicy(""), errors.New(""))
				m.GetDefault()
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.RequestsDefault()
				m.LimitsDefault()
			},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {
				m.StoreFromAnnotationsDefault()
				m.GetTargetContainerNameDefault()
				m.GetTargetContainerNameDefault()
				m.ValidateDefault()
			},
			wantErrMsg:       "unable to get target container memory resize policy",
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: unable to get target container memory resize policy",
		},
		{
			name: "TargetContainerMemoryResizePolicyIncorrect",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
			},
			configContHelperMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.On("ResizePolicy", mock.Anything, v1.ResourceCPU).Return(v1.NotRequired, nil)
				m.On("ResizePolicy", mock.Anything, v1.ResourceMemory).Return(v1.RestartContainer, nil)
				m.GetDefault()
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.RequestsDefault()
				m.LimitsDefault()
			},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {
				m.StoreFromAnnotationsDefault()
				m.GetTargetContainerNameDefault()
				m.GetTargetContainerNameDefault()
				m.ValidateDefault()
			},
			wantErrMsg:       fmt.Sprintf("target container memory resize policy is not '%s' ('%s')", v1.NotRequired, v1.RestartContainer),
			wantStatusUpdate: true,
			wantEventMsg:     fmt.Sprintf("Validation error: target container memory resize policy is not '%s' ('%s')", v1.NotRequired, v1.RestartContainer),
		},
		{
			name: "Ok",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configHelperMockFunc: func(m *podtest.MockKubeHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
			},
			configContHelperMockFunc: func(m *podtest.MockContainerKubeHelper) {
				m.GetDefault()
				m.HasStartupProbeDefault()
				m.HasReadinessProbeDefault()
				m.RequestsDefault()
				m.LimitsDefault()
				m.ResizePolicyDefault()
			},
			configScaleConfigMockFunc: func(m *podtest.MockScaleConfig) {
				m.StoreFromAnnotationsDefault()
				m.GetTargetContainerNameDefault()
				m.GetTargetContainerNameDefault()
				m.ValidateDefault()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			run := func() { statusUpdated = true }
			eventRecorder := record.NewFakeRecorder(1)
			v := newValidation(
				eventRecorder,
				podtest.NewMockStatusWithRun(tt.configStatusMockFunc, run),
				podtest.NewMockKubeHelper(tt.configHelperMockFunc),
				podtest.NewMockContainerKubeHelper(tt.configContHelperMockFunc),
			)
			config := podtest.NewMockScaleConfig(tt.configScaleConfigMockFunc)

			err := v.Validate(
				contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				&v1.Pod{},
				config,
				func(podcommon.ScaleConfig) {},
			)
			if tt.wantErrMsg != "" {
				assert.True(t, errors.Is(err, ValidationError{}))
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			if tt.wantStatusUpdate {
				assert.True(t, statusUpdated)
			} else {
				assert.False(t, statusUpdated)
			}
			if tt.wantEventMsg != "" {
				select {
				case res := <-eventRecorder.Events:
					assert.Contains(t, res, tt.wantEventMsg)
				case <-time.After(5 * time.Second):
					t.Fatalf("event not generated")
				}
			}
		})
	}
}

func TestValidationUpdateStatusAndGetError(t *testing.T) {
	t.Run("UnableToUpdateStatus", func(t *testing.T) {
		configStatusMockFunc := func(m *podtest.MockStatus) {
			m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(&v1.Pod{}, errors.New(""))
		}
		v := newValidation(
			&record.FakeRecorder{},
			podtest.NewMockStatus(configStatusMockFunc),
			nil,
			nil,
		)

		buffer := &bytes.Buffer{}
		_ = v.updateStatusAndGetError(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(buffer)).Build(),
			&v1.Pod{},
			"",
			nil,
		)
		assert.Contains(t, buffer.String(), "unable to update status (will continue)")
	})
}
