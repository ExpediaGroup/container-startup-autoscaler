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

package pod

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contexttest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podtest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scaletest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

func TestNewValidation(t *testing.T) {
	recorder := &record.FakeRecorder{}
	podHelper := kube.NewPodHelper(nil)
	containerHelper := kube.NewContainerHelper()
	stat := newStatus(podHelper)
	val := newValidation(recorder, stat, podHelper, containerHelper)
	expected := &validation{
		recorder:        recorder,
		status:          stat,
		podHelper:       podHelper,
		containerHelper: containerHelper,
	}
	assert.Equal(t, expected, val)
}

func TestValidationValidate(t *testing.T) {
	tests := []struct {
		name                       string
		configStatusMockFunc       func(*podtest.MockStatus, func())
		configPodHelperMockFunc    func(*kubetest.MockPodHelper)
		configContHelperMockFunc   func(*kubetest.MockContainerHelper)
		configScaleConfigsMockFunc func(*scaletest.MockConfigurations)
		wantErrMsg                 string
		wantNilContainer           bool
		wantStatusUpdate           bool
		wantEventMsg               string
	}{
		{
			"UnableToGetEnabledLabelValue",
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			func(m *kubetest.MockPodHelper) {
				m.On("ExpectedLabelValueAs", mock.Anything, kubecommon.LabelEnabled, kubecommon.DataTypeBool).
					Return(nil, errors.New(""))
			},
			nil,
			nil,
			"unable to get pod enabled label value",
			true,
			true,
			"Validation error: unable to get pod enabled label value",
		},
		{
			"EnabledLabelFalse",
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			func(m *kubetest.MockPodHelper) {
				m.On("ExpectedLabelValueAs", mock.Anything, kubecommon.LabelEnabled, kubecommon.DataTypeBool).
					Return(false, nil)
			},
			nil,
			nil,
			"pod enabled label value is unexpectedly 'false'",
			true,
			true,
			"Validation error: pod enabled label value is unexpectedly 'false'",
		},
		{
			"VpaNotSupported",
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, KnownVpaAnnotations[0]).Return(true, "")
				m.ExpectedLabelValueAsDefault()
			},
			nil,
			nil,
			"vpa not supported",
			true,
			true,
			"Validation error: vpa not supported",
		},
		{
			"TargetContainerNotInPodSpec",
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.On("IsContainerInSpec", mock.Anything, mock.Anything).Return(false)
				m.ExpectedLabelValueAsDefault()
			},
			nil,
			nil,
			"target container not in pod spec",
			true,
			true,
			"Validation error: target container not in pod spec",
		},
		{
			"TargetContainerNoProbes",
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
			},
			func(m *kubetest.MockContainerHelper) {
				m.On("HasStartupProbe", mock.Anything).Return(false)
				m.On("HasReadinessProbe", mock.Anything).Return(false)
				m.GetDefault()
			},
			nil,
			"target container does not specify startup probe or readiness probe",
			true,
			true,
			"Validation error: target container does not specify startup probe or readiness probe",
		},
		{
			"UnableToDeterminePodQosClass",
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.On("QOSClass", mock.Anything).Return(v1.PodQOSClass(""), errors.New(""))
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()

			},
			nil,
			nil,
			"unable to determine pod qos class",
			true,
			true,
			"Validation error: unable to determine pod qos class",
		},
		{
			"PodQosClassIsNotGuaranteed",
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.On("QOSClass", mock.Anything).Return(v1.PodQOSBurstable, nil)
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
			},
			nil,
			nil,
			"pod qos class is not guaranteed",
			true,
			true,
			"Validation error: pod qos class is not guaranteed",
		},
		{
			"UnableToValidateScaleConfiguration",
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
				m.QOSClassDefault()
			},
			nil,
			func(m *scaletest.MockConfigurations) {
				m.On("ValidateAll", mock.Anything).Return(errors.New("text"))
			},
			"text",
			true,
			true,
			"Validation error: text",
		},
		{
			"UnableToValidateScaleConfigurationCollection",
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
				m.QOSClassDefault()
			},
			nil,
			func(m *scaletest.MockConfigurations) {
				m.On("ValidateAll", mock.Anything).Return(nil)
				m.On("ValidateCollection", mock.Anything).Return(errors.New("text"))
			},
			"text",
			true,
			true,
			"Validation error: text",
		},
		{
			"Ok",
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
				m.QOSClassDefault()
			},
			nil,
			nil,
			"",
			false,
			false,
			"",
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
				kubetest.NewMockPodHelper(tt.configPodHelperMockFunc),
				kubetest.NewMockContainerHelper(tt.configContHelperMockFunc),
			)
			configs := scaletest.NewMockConfigurations(tt.configScaleConfigsMockFunc)

			container, err := v.Validate(
				contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				&v1.Pod{},
				"",
				configs,
			)
			if tt.wantErrMsg != "" {
				assert.True(t, errors.As(err, &ValidationError{}))
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantNilContainer {
				assert.Nil(t, container)
			} else {
				assert.NotNil(t, container)
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
			m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
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
			nil,
		)
		assert.Contains(t, buffer.String(), "unable to update status (will continue)")
	})
}
