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
	"errors"
	"testing"
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contexttest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/controller/controllercommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/metricscommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/scale"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podtest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scaletest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/component-base/metrics/testutil"
)

func TestNewTargetContainerAction(t *testing.T) {
	config := controllercommon.ControllerConfig{}
	recorder := &record.FakeRecorder{}
	podHelper := kube.NewPodHelper(nil)
	stat := newStatus(podHelper)
	s := newTargetContainerAction(config, recorder, stat, podHelper)
	assert.Equal(t, config, s.controllerConfig)
	assert.Equal(t, recorder, s.recorder)
	assert.Equal(t, stat, s.status)
	assert.Equal(t, podHelper, s.podHelper)
}

func TestTargetContainerActionExecute(t *testing.T) {
	tests := []struct {
		name                 string
		scaleWhenUnknownRes  bool
		configStatusMockFunc func(*podtest.MockStatus, func())
		states               podcommon.States
		wantPanicErrMsg      string
		wantErrMsg           string
		wantLogMsg           string
		wantStatusUpdate     bool
	}{
		{
			name:                 "StartupProbeUnknownPanics",
			configStatusMockFunc: nil,
			states:               podcommon.States{StartupProbe: podcommon.StateBoolUnknown},
			wantPanicErrMsg:      "unsupported startup probe state 'unknown'",
		},
		{
			name:                 "ReadinessProbeUnknownPanics",
			configStatusMockFunc: nil,
			states: podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolUnknown,
			},
			wantPanicErrMsg: "unsupported readiness probe state 'unknown'",
		},
		{
			name: "ContainerNotRunningAction",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			states: podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerWaiting,
			},
			wantLogMsg:       "target container currently not running",
			wantStatusUpdate: true,
		},
		{
			name: "StartedUnknownAction",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			states: podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolUnknown,
			},
			wantLogMsg:       "target container started status currently unknown",
			wantStatusUpdate: true,
		},
		{
			name: "ReadyUnknownAction",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			states: podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolTrue,
				Ready:          podcommon.StateBoolUnknown,
			},
			wantLogMsg:       "target container ready status currently unknown",
			wantStatusUpdate: true,
		},
		{
			name:                "ResUnknownAction",
			scaleWhenUnknownRes: false,
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			states: podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolTrue,
				Ready:          podcommon.StateBoolTrue,
				Resources:      podcommon.StateResourcesUnknown,
			},
			wantErrMsg:       "unknown resources applied",
			wantStatusUpdate: true,
		},
		{
			name: "NeitherProbePresentPanics",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			states: podcommon.States{
				StartupProbe:    podcommon.StateBoolFalse,
				ReadinessProbe:  podcommon.StateBoolFalse,
				Container:       podcommon.StateContainerRunning,
				Started:         podcommon.StateBoolFalse,
				Ready:           podcommon.StateBoolFalse,
				Resources:       podcommon.StateResourcesStartup,
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
			},
			wantPanicErrMsg: "neither startup probe or readiness probe present",
		},
		{
			name: "NotStartedWithStartupResActionStartupProbe",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			states: podcommon.States{
				StartupProbe:    podcommon.StateBoolTrue,
				ReadinessProbe:  podcommon.StateBoolFalse,
				Container:       podcommon.StateContainerRunning,
				Started:         podcommon.StateBoolFalse,
				Ready:           podcommon.StateBoolFalse,
				Resources:       podcommon.StateResourcesStartup,
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
			},
			wantLogMsg:       "startup resources enacted",
			wantStatusUpdate: true,
		},
		{
			name: "NotStartedWithStartupResActionReadinessProbe",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			states: podcommon.States{
				StartupProbe:    podcommon.StateBoolFalse,
				ReadinessProbe:  podcommon.StateBoolTrue,
				Container:       podcommon.StateContainerRunning,
				Started:         podcommon.StateBoolFalse,
				Ready:           podcommon.StateBoolFalse,
				Resources:       podcommon.StateResourcesStartup,
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
			},
			wantLogMsg:       "startup resources enacted",
			wantStatusUpdate: true,
		},
		{
			name: "StartedWithStartupResAction",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.PodMutationFuncDefaultAndRun(run)
			},
			states: podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolTrue,
				Ready:          podcommon.StateBoolTrue,
				Resources:      podcommon.StateResourcesStartup,
			},
			wantLogMsg:       "post-startup resources commanded",
			wantStatusUpdate: true,
		},
		{
			name: "NotStartedWithPostStartupResAction",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.PodMutationFuncDefaultAndRun(run)
			},
			states: podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolFalse,
				Ready:          podcommon.StateBoolFalse,
				Resources:      podcommon.StateResourcesPostStartup,
			},
			wantLogMsg:       "startup resources commanded",
			wantStatusUpdate: true,
		},
		{
			name: "StartedWithPostStartupResAction",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			states: podcommon.States{
				StartupProbe:    podcommon.StateBoolTrue,
				ReadinessProbe:  podcommon.StateBoolTrue,
				Container:       podcommon.StateContainerRunning,
				Started:         podcommon.StateBoolTrue,
				Ready:           podcommon.StateBoolTrue,
				Resources:       podcommon.StateResourcesPostStartup,
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
			},
			wantLogMsg:       "post-startup resources enacted",
			wantStatusUpdate: true,
		},
		{
			name:                "NotStartedWithUnknownResAction",
			scaleWhenUnknownRes: true,
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.PodMutationFuncDefaultAndRun(run)
			},
			states: podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolFalse,
				Ready:          podcommon.StateBoolFalse,
				Resources:      podcommon.StateResourcesUnknown,
			},
			wantLogMsg:       "startup resources commanded (unknown resources applied)",
			wantStatusUpdate: true,
		},
		{
			name:                "StartedWithUnknownResAction",
			scaleWhenUnknownRes: true,
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) {
				m.PodMutationFuncDefaultAndRun(run)
			},
			states: podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolTrue,
				Ready:          podcommon.StateBoolTrue,
				Resources:      podcommon.StateResourcesUnknown,
			},
			wantLogMsg:       "post-startup resources commanded (unknown resources applied)",
			wantStatusUpdate: true,
		},
		{
			name:                 "NoActionPanics",
			configStatusMockFunc: nil,
			states: podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolTrue,
				Ready:          podcommon.StateBoolTrue,
				Resources:      podcommon.StateResources("test"),
			},
			wantPanicErrMsg: "no action to invoke",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			run := func() { statusUpdated = true }
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{ScaleWhenUnknownResources: tt.scaleWhenUnknownRes},
				&record.FakeRecorder{},
				podtest.NewMockStatusWithRun(tt.configStatusMockFunc, run),
				kubetest.NewMockPodHelper(nil),
			)

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() {
					_ = a.Execute(nil, tt.states, &v1.Pod{}, &v1.Container{}, scaletest.NewMockConfigurations(nil))
				})
				return
			}

			buffer := bytes.Buffer{}
			err := a.Execute(
				contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(&buffer)).Build(),
				tt.states,
				&v1.Pod{},
				&v1.Container{},
				scaletest.NewMockConfigurations(nil),
			)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			if tt.wantLogMsg != "" {
				assert.Contains(t, buffer.String(), tt.wantLogMsg)
			}
			if tt.wantStatusUpdate {
				assert.True(t, statusUpdated)
			} else {
				assert.False(t, statusUpdated)
			}
		})
	}
}

func TestTargetContainerActionContainerNotRunningAction(t *testing.T) {
	statusUpdated := false
	configStatusMock := podtest.NewMockStatusWithRun(
		func(status *podtest.MockStatus, run func()) { status.UpdateDefaultAndRun(run) },
		func() { statusUpdated = true },
	)
	a := newTargetContainerAction(
		controllercommon.ControllerConfig{},
		&record.FakeRecorder{},
		configStatusMock,
		kubetest.NewMockPodHelper(nil),
	)

	buffer := bytes.Buffer{}
	_ = a.containerNotRunningAction(
		contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(&buffer)).Build(),
		podcommon.States{},
		&v1.Pod{},
		scaletest.NewMockConfigurations(nil),
	)
	assert.Contains(t, buffer.String(), "target container currently not running")
	assert.True(t, statusUpdated)
}

func TestTargetContainerActionStartedUnknownAction(t *testing.T) {
	statusUpdated := false
	configStatusMock := podtest.NewMockStatusWithRun(
		func(status *podtest.MockStatus, run func()) { status.UpdateDefaultAndRun(run) },
		func() { statusUpdated = true },
	)
	a := newTargetContainerAction(
		controllercommon.ControllerConfig{},
		&record.FakeRecorder{},
		configStatusMock,
		kubetest.NewMockPodHelper(nil),
	)

	buffer := bytes.Buffer{}
	_ = a.startedUnknownAction(
		contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(&buffer)).Build(),
		podcommon.States{},
		&v1.Pod{},
		scaletest.NewMockConfigurations(nil),
	)
	assert.Contains(t, buffer.String(), "target container started status currently unknown")
	assert.True(t, statusUpdated)
}

func TestTargetContainerActionReadyUnknownAction(t *testing.T) {
	statusUpdated := false
	configStatusMock := podtest.NewMockStatusWithRun(
		func(status *podtest.MockStatus, run func()) { status.UpdateDefaultAndRun(run) },
		func() { statusUpdated = true },
	)
	a := newTargetContainerAction(
		controllercommon.ControllerConfig{},
		&record.FakeRecorder{},
		configStatusMock,
		kubetest.NewMockPodHelper(nil),
	)

	buffer := bytes.Buffer{}
	_ = a.readyUnknownAction(
		contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(&buffer)).Build(),
		podcommon.States{},
		&v1.Pod{},
		scaletest.NewMockConfigurations(nil),
	)
	assert.Contains(t, buffer.String(), "target container ready status currently unknown")
	assert.True(t, statusUpdated)
}

func TestTargetContainerActionResUnknownAction(t *testing.T) {
	statusUpdated := false
	configStatusMock := podtest.NewMockStatusWithRun(
		func(status *podtest.MockStatus, run func()) { status.UpdateDefaultAndRun(run) },
		func() { statusUpdated = true },
	)
	a := newTargetContainerAction(
		controllercommon.ControllerConfig{},
		&record.FakeRecorder{},
		configStatusMock,
		kubetest.NewMockPodHelper(nil),
	)

	err := a.resUnknownAction(
		contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
		podcommon.States{},
		&v1.Pod{},
		&v1.Container{},
		scaletest.NewMockConfigurations(nil),
	)
	assert.Contains(t, err.Error(), "unknown resources applied")
	assert.True(t, statusUpdated)
}

func TestTargetContainerActionNotStartedWithStartupResAction(t *testing.T) {
	tests := []struct {
		name                    string
		states                  podcommon.States
		configPodHelperMockFunc func(*kubetest.MockPodHelper)
		wantErr                 bool
		wantStatusUpdate        bool
	}{
		{
			name: "Error",
			states: podcommon.States{
				Resources:       podcommon.StateResourcesStartup,
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMismatch,
			},
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeStatus", mock.Anything).Return(v1.PodResizeStatusInfeasible)
			},
			wantErr:          true,
			wantStatusUpdate: true,
		},
		{
			name: "Ok",
			states: podcommon.States{
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
			},
			configPodHelperMockFunc: nil,
			wantErr:                 false,
			wantStatusUpdate:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{},
				&record.FakeRecorder{},
				podtest.NewMockStatusWithRun(
					func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
					func() { statusUpdated = true },
				),
				kubetest.NewMockPodHelper(tt.configPodHelperMockFunc),
			)

			err := a.notStartedWithStartupResAction(
				contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				tt.states,
				&v1.Pod{},
				&v1.Container{},
				scaletest.NewMockConfigurations(nil),
			)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			if tt.wantStatusUpdate {
				assert.True(t, statusUpdated)
			} else {
				assert.False(t, statusUpdated)
			}
		})
	}
}

func TestTargetContainerActionNotStartedWithPostStartupResAction(t *testing.T) {
	tests := []struct {
		name                    string
		configPodHelperMockFunc func(*kubetest.MockPodHelper)
		wantErrMsg              string
		wantStatusUpdate        bool
		wantEventMsg            string
	}{
		{
			name: "UnableToPatchContainerResources",
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&v1.Pod{}, errors.New(""))
			},
			wantErrMsg:       "unable to patch container resources",
			wantStatusUpdate: false,
		},
		{
			name:                    "Ok",
			configPodHelperMockFunc: nil,
			wantStatusUpdate:        true,
			wantEventMsg:            "Startup resources commanded",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			eventRecorder := record.NewFakeRecorder(1)
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{},
				eventRecorder,
				podtest.NewMockStatusWithRun(
					func(m *podtest.MockStatus, run func()) { m.PodMutationFuncDefaultAndRun(run) },
					func() { statusUpdated = true },
				),
				kubetest.NewMockPodHelper(tt.configPodHelperMockFunc),
			)

			err := a.notStartedWithPostStartupResAction(
				contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				podcommon.States{},
				&v1.Pod{},
				&v1.Container{},
				scaletest.NewMockConfigurations(nil),
			)
			if tt.wantErrMsg != "" {
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

func TestTargetContainerActionStartedWithStartupResAction(t *testing.T) {
	tests := []struct {
		name                    string
		configPodHelperMockFunc func(*kubetest.MockPodHelper)
		wantErrMsg              string
		wantStatusUpdate        bool
		wantEventMsg            string
	}{
		{
			name: "UnableToPatchContainerResources",
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&v1.Pod{}, errors.New(""))
			},
			wantErrMsg:       "unable to patch container resources",
			wantStatusUpdate: false,
		},
		{
			name:                    "Ok",
			configPodHelperMockFunc: nil,
			wantStatusUpdate:        true,
			wantEventMsg:            "Post-startup resources commanded",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			eventRecorder := record.NewFakeRecorder(1)
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{},
				eventRecorder,
				podtest.NewMockStatusWithRun(
					func(m *podtest.MockStatus, run func()) { m.PodMutationFuncDefaultAndRun(run) },
					func() { statusUpdated = true },
				),
				kubetest.NewMockPodHelper(tt.configPodHelperMockFunc),
			)

			err := a.startedWithStartupResAction(
				contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				podcommon.States{},
				&v1.Pod{},
				&v1.Container{},
				scaletest.NewMockConfigurations(nil),
			)
			if tt.wantErrMsg != "" {
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

func TestTargetContainerActionStartedWithPostStartupResAction(t *testing.T) {
	tests := []struct {
		name                    string
		states                  podcommon.States
		configPodHelperMockFunc func(*kubetest.MockPodHelper)
		wantErr                 bool
		wantStatusUpdate        bool
	}{
		{
			name: "Error",
			states: podcommon.States{
				Resources:       podcommon.StateResourcesPostStartup,
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMismatch,
			},
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeStatus", mock.Anything).Return(v1.PodResizeStatusInfeasible)
			},
			wantErr:          true,
			wantStatusUpdate: true,
		},
		{
			name: "Ok",
			states: podcommon.States{
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
			},
			configPodHelperMockFunc: nil,
			wantErr:                 false,
			wantStatusUpdate:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{},
				&record.FakeRecorder{},
				podtest.NewMockStatusWithRun(
					func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
					func() { statusUpdated = true },
				),
				kubetest.NewMockPodHelper(tt.configPodHelperMockFunc),
			)

			err := a.startedWithPostStartupResAction(
				contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				tt.states,
				&v1.Pod{},
				&v1.Container{},
				scaletest.NewMockConfigurations(nil),
			)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			if tt.wantStatusUpdate {
				assert.True(t, statusUpdated)
			} else {
				assert.False(t, statusUpdated)
			}
		})
	}
}

func TestTargetContainerActionNotStartedWithUnknownResAction(t *testing.T) {
	tests := []struct {
		name                    string
		configPodHelperMockFunc func(*kubetest.MockPodHelper)
		wantErrMsg              string
		wantStatusUpdate        bool
		wantEventMsg            string
	}{
		{
			name: "UnableToPatchContainerResources",
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&v1.Pod{}, errors.New(""))
			},
			wantErrMsg:       "unable to patch container resources",
			wantStatusUpdate: false,
		},
		{
			name:                    "Ok",
			configPodHelperMockFunc: nil,
			wantStatusUpdate:        true,
			wantEventMsg:            "Startup resources commanded (unknown resources applied)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			eventRecorder := record.NewFakeRecorder(1)
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{},
				eventRecorder,
				podtest.NewMockStatusWithRun(
					func(m *podtest.MockStatus, run func()) { m.PodMutationFuncDefaultAndRun(run) },
					func() { statusUpdated = true },
				),
				kubetest.NewMockPodHelper(tt.configPodHelperMockFunc),
			)

			err := a.notStartedWithUnknownResAction(
				contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				podcommon.States{},
				&v1.Pod{},
				&v1.Container{},
				scaletest.NewMockConfigurations(nil),
			)
			if tt.wantErrMsg != "" {
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

func TestTargetContainerActionStartedWithUnknownResAction(t *testing.T) {
	tests := []struct {
		name                    string
		configPodHelperMockFunc func(*kubetest.MockPodHelper)
		wantErrMsg              string
		wantStatusUpdate        bool
		wantEventMsg            string
	}{
		{
			name: "UnableToPatchContainerResources",
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&v1.Pod{}, errors.New(""))
			},
			wantErrMsg:       "unable to patch container resources",
			wantStatusUpdate: false,
		},
		{
			name:                    "Ok",
			configPodHelperMockFunc: nil,
			wantStatusUpdate:        true,
			wantEventMsg:            "Post-startup resources commanded (unknown resources applied)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			eventRecorder := record.NewFakeRecorder(1)
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{},
				eventRecorder,
				podtest.NewMockStatusWithRun(
					func(m *podtest.MockStatus, run func()) { m.PodMutationFuncDefaultAndRun(run) },
					func() { statusUpdated = true },
				),
				kubetest.NewMockPodHelper(tt.configPodHelperMockFunc),
			)

			err := a.startedWithUnknownResAction(
				contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				podcommon.States{},
				&v1.Pod{},
				&v1.Container{},
				scaletest.NewMockConfigurations(nil),
			)
			if tt.wantErrMsg != "" {
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

func TestTargetContainerActionProcessConfigEnacted(t *testing.T) {
	tests := []struct {
		name                    string
		configStatusMockFunc    func(*podtest.MockStatus, func())
		configPodHelperMockFunc func(*kubetest.MockPodHelper)
		configMetricAssertsFunc func(t *testing.T)
		beforeTestFunc          func()
		states                  podcommon.States
		wantPanicErrMsg         string
		wantErrMsg              string
		wantStatusUpdate        bool
		wantLogMsg              string
		wantEventMsg            string
	}{
		{
			name:                 "ScaleNotYetCompletedProposed",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeStatus", mock.Anything).Return(v1.PodResizeStatusProposed)
			},
			wantStatusUpdate: true,
			wantLogMsg:       "scale not yet completed - has been proposed",
		},
		{
			name:                 "ScaleNotYetCompletedInProgress",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeStatus", mock.Anything).Return(v1.PodResizeStatusInProgress)
			},
			wantStatusUpdate: true,
			wantLogMsg:       "scale not yet completed - in progress",
		},
		{
			name:                 "ScaleNotYetCompletedDeferred",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeStatus", mock.Anything).Return(v1.PodResizeStatusDeferred)
			},
			wantStatusUpdate: true,
			wantLogMsg:       "scale not yet completed - deferred",
		},
		{
			name:                 "ScaleFailedInfeasibleStateResourcesPostStartup",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeStatus", mock.Anything).Return(v1.PodResizeStatusInfeasible)
			},
			configMetricAssertsFunc: func(t *testing.T) {
				metricVal, _ := testutil.GetCounterMetricValue(scale.Failure(metricscommon.DirectionDown, "infeasible"))
				assert.Equal(t, float64(1), metricVal)
			},
			beforeTestFunc: func() {
				scale.ResetMetrics()
			},
			states: podcommon.States{
				Resources: podcommon.StateResourcesPostStartup,
			},
			wantErrMsg:       "post-startup scale failed - infeasible",
			wantStatusUpdate: true,
			wantEventMsg:     "Post-startup scale failed - infeasible",
		},
		{
			name:                 "ScaleFailedInfeasibleStateResourcesStartup",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeStatus", mock.Anything).Return(v1.PodResizeStatusInfeasible)
			},
			configMetricAssertsFunc: func(t *testing.T) {
				metricVal, _ := testutil.GetCounterMetricValue(scale.Failure(metricscommon.DirectionUp, "infeasible"))
				assert.Equal(t, float64(1), metricVal)
			},
			beforeTestFunc: func() {
				scale.ResetMetrics()
			},
			states: podcommon.States{
				Resources: podcommon.StateResourcesStartup,
			},
			wantErrMsg:       "startup scale failed - infeasible",
			wantStatusUpdate: true,
			wantEventMsg:     "Startup scale failed - infeasible",
		},
		{
			name:                 "UnknownStatusStateResourcesPostStartup",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeStatus", mock.Anything).Return(v1.PodResizeStatus("test"))
			},
			configMetricAssertsFunc: func(t *testing.T) {
				metricVal, _ := testutil.GetCounterMetricValue(scale.Failure(metricscommon.DirectionDown, "unknownstatus"))
				assert.Equal(t, float64(1), metricVal)
			},
			states: podcommon.States{
				Resources: podcommon.StateResourcesPostStartup,
			},
			wantErrMsg:       "post-startup scale: unknown status 'test'",
			wantStatusUpdate: true,
			wantEventMsg:     "Post-startup scale: unknown status",
		},
		{
			name:                 "UnknownStatusStateResourcesStartup",
			configStatusMockFunc: func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) {
				m.On("ResizeStatus", mock.Anything).Return(v1.PodResizeStatus("test"))
			},
			configMetricAssertsFunc: func(t *testing.T) {
				metricVal, _ := testutil.GetCounterMetricValue(scale.Failure(metricscommon.DirectionUp, "unknownstatus"))
				assert.Equal(t, float64(1), metricVal)
			},
			states: podcommon.States{
				Resources: podcommon.StateResourcesStartup,
			},
			wantErrMsg:       "startup scale: unknown status 'test'",
			wantStatusUpdate: true,
			wantEventMsg:     "Startup scale: unknown status",
		},
		{
			name:                    string(podcommon.StateStatusResourcesIncomplete),
			configStatusMockFunc:    func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) { m.ResizeStatusDefault() },
			states: podcommon.States{
				StatusResources: podcommon.StateStatusResourcesIncomplete,
			},
			wantStatusUpdate: true,
			wantLogMsg:       "target container current cpu and/or memory resources currently missing",
		},
		{
			name:                    string(podcommon.StateStatusResourcesContainerResourcesMismatch),
			configStatusMockFunc:    func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) { m.ResizeStatusDefault() },
			states: podcommon.States{
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMismatch,
			},
			wantStatusUpdate: true,
			wantLogMsg:       "target container current cpu and/or memory resources currently don't match target container's 'requests'",
		},
		{
			name:                    string(podcommon.StateStatusResourcesUnknown),
			configStatusMockFunc:    func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) { m.ResizeStatusDefault() },
			states: podcommon.States{
				StatusResources: podcommon.StateStatusResourcesUnknown,
			},
			wantStatusUpdate: true,
			wantLogMsg:       "target container current cpu and/or memory resources currently unknown",
		},
		{
			name:                    "UnknownStatusResourcesStatePanics",
			configStatusMockFunc:    func(m *podtest.MockStatus, run func()) {},
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) { m.ResizeStatusDefault() },
			states: podcommon.States{
				StatusResources: podcommon.StateStatusResources("test"),
			},
			wantPanicErrMsg:  "unknown state 'test'",
			wantStatusUpdate: true,
		},
		{
			name:                    string(podcommon.StateStatusResourcesContainerResourcesMatch) + string(podcommon.StateResourcesPostStartup),
			configStatusMockFunc:    func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) { m.ResizeStatusDefault() },
			states: podcommon.States{
				Resources:       podcommon.StateResourcesPostStartup,
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
			},
			wantStatusUpdate: true,
			wantLogMsg:       "post-startup resources enacted",
			wantEventMsg:     "Post-startup resources enacted",
		},
		{
			name:                    string(podcommon.StateStatusResourcesContainerResourcesMatch) + string(podcommon.StateResourcesStartup),
			configStatusMockFunc:    func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			configPodHelperMockFunc: func(m *kubetest.MockPodHelper) { m.ResizeStatusDefault() },
			states: podcommon.States{
				Resources:       podcommon.StateResourcesStartup,
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
			},
			wantStatusUpdate: true,
			wantLogMsg:       "startup resources enacted",
			wantEventMsg:     "Startup resources enacted",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			eventRecorder := record.NewFakeRecorder(1)
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{},
				eventRecorder,
				podtest.NewMockStatusWithRun(tt.configStatusMockFunc, func() { statusUpdated = true }),
				kubetest.NewMockPodHelper(tt.configPodHelperMockFunc),
			)

			if tt.beforeTestFunc != nil {
				tt.beforeTestFunc()
			}

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() {
					_ = a.processConfigEnacted(nil, tt.states, &v1.Pod{}, &v1.Container{}, scaletest.NewMockConfigurations(nil))
				})
				return
			}

			buffer := bytes.Buffer{}
			err := a.processConfigEnacted(
				contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(&buffer)).Build(),
				tt.states,
				&v1.Pod{},
				&v1.Container{},
				scaletest.NewMockConfigurations(nil),
			)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			if tt.wantStatusUpdate {
				assert.True(t, statusUpdated)
			} else {
				assert.False(t, statusUpdated)
			}
			if tt.wantLogMsg != "" {
				assert.Contains(t, buffer.String(), tt.wantLogMsg)
			}
			if tt.configMetricAssertsFunc != nil {
				tt.configMetricAssertsFunc(t)
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

func TestTargetContainerActionContainerResourceConfig(t *testing.T) {
	a := newTargetContainerAction(
		controllercommon.ControllerConfig{},
		&record.FakeRecorder{},
		nil,
		nil,
	)

	mockContainer := kubetest.NewContainerBuilder().Build()
	mockScaleConfigs := scaletest.NewMockConfigurations(func(m *scaletest.MockConfigurations) {
		m.On("String").Return("test")
	})
	got := a.containerResourceConfig(mockContainer, mockScaleConfigs)
	assert.Contains(t, got, "target container resources: [")
	assert.Contains(t, got, "], configurations: [test]")
}

func TestTargetContainerActionUpdateStatus(t *testing.T) {
	t.Run("UnableToUpdateStatus", func(t *testing.T) {
		mockStatus := podtest.NewMockStatus(func(m *podtest.MockStatus) {
			m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(&v1.Pod{}, errors.New(""))
		})
		a := newTargetContainerAction(
			controllercommon.ControllerConfig{},
			&record.FakeRecorder{},
			mockStatus,
			nil,
		)

		buffer := bytes.Buffer{}
		a.updateStatus(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(&buffer)).Build(),
			podcommon.States{},
			podcommon.StatusScaleStateNotApplicable,
			&v1.Pod{},
			"",
			scaletest.NewMockConfigurations(nil),
		)
		assert.Contains(t, buffer.String(), "unable to update status")
	})
}
