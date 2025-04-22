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
	assert.Equal(t, recorder, val.recorder)
	assert.Equal(t, stat, val.status)
	assert.Equal(t, podHelper, val.podHelper)
	assert.Equal(t, containerHelper, val.containerHelper)
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
			name: "UnableToGetEnabledLabelValue",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ExpectedLabelValueAs", mock.Anything, kubecommon.LabelEnabled, kubecommon.DataTypeBool).
					Return(nil, errors.New(""))
			},
			wantErrMsg:       "unable to get pod enabled label value",
			wantNilContainer: true,
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: unable to get pod enabled label value",
		},
		{
			name: "EnabledLabelFalse",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ExpectedLabelValueAs", mock.Anything, kubecommon.LabelEnabled, kubecommon.DataTypeBool).
					Return(false, nil)
			},
			wantErrMsg:       "pod enabled label value is unexpectedly 'false'",
			wantNilContainer: true,
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: pod enabled label value is unexpectedly 'false'",
		},
		{
			name: "VpaNotSupported",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, KnownVpaAnnotations[0]).Return(true, "")
				m.ExpectedLabelValueAsDefault()
			},
			wantErrMsg:       "vpa not supported",
			wantNilContainer: true,
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: vpa not supported",
		},
		{
			name: "TargetContainerNotInPodSpec",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.On("IsContainerInSpec", mock.Anything, mock.Anything).Return(false)
				m.ExpectedLabelValueAsDefault()
			},
			wantErrMsg:       "target container not in pod spec",
			wantNilContainer: true,
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: target container not in pod spec",
		},
		{
			name: "TargetContainerNoProbes",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
			},
			configContHelperMockFunc: func(m *kubetest.MockContainerHelper) {
				m.On("HasStartupProbe", mock.Anything).Return(false)
				m.On("HasReadinessProbe", mock.Anything).Return(false)
				m.GetDefault()
			},
			wantErrMsg:       "target container does not specify startup probe or readiness probe",
			wantNilContainer: true,
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: target container does not specify startup probe or readiness probe",
		},
		{
			name: "UnableToDeterminePodQosClass",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.On("QOSClass", mock.Anything).Return(v1.PodQOSClass(""), errors.New(""))
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()

			},
			wantErrMsg:       "unable to determine pod qos class",
			wantNilContainer: true,
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: unable to determine pod qos class",
		},
		{
			name: "PodQosClassIsNotGuaranteed",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.On("QOSClass", mock.Anything).Return(v1.PodQOSBurstable, nil)
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
			},
			wantErrMsg:       "pod qos class is not guaranteed",
			wantNilContainer: true,
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: pod qos class is not guaranteed",
		},
		{
			name: "UnableToValidateScaleConfiguration",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
				m.QOSClassDefault()
			},
			configScaleConfigsMockFunc: func(m *scaletest.MockConfigurations) {
				m.On("ValidateAll", mock.Anything).Return(errors.New(""))
			},
			wantErrMsg:       "unable to validate scale configuration",
			wantNilContainer: true,
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: unable to validate scale configuration",
		},
		{
			name: "UnableToValidateScaleConfigurationCollection",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
				m.QOSClassDefault()
			},
			configScaleConfigsMockFunc: func(m *scaletest.MockConfigurations) {
				m.On("ValidateAll", mock.Anything).Return(nil)
				m.On("ValidateCollection", mock.Anything).Return(errors.New(""))
			},
			wantErrMsg:       "unable to validate scale configuration collection",
			wantNilContainer: true,
			wantStatusUpdate: true,
			wantEventMsg:     "Validation error: unable to validate scale configuration collection",
		},
		{
			name: "Ok",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				m.ExpectedLabelValueAsDefault()
				m.IsContainerInSpecDefault()
				m.QOSClassDefault()
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
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
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
