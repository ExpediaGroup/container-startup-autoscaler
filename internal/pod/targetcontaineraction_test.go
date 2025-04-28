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

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contexttest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/controller/controllercommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podtest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scaletest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

func TestNewTargetContainerAction(t *testing.T) {
	recorder := &record.FakeRecorder{}
	config := controllercommon.ControllerConfig{}
	podHelper := kube.NewPodHelper(nil)
	stat := newStatus(recorder, podHelper)
	action := newTargetContainerAction(config, stat, podHelper)
	expected := &targetContainerAction{
		controllerConfig: config,
		status:           stat,
		podHelper:        podHelper,
	}
	assert.Equal(t, expected, action)
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
			"StartupProbeUnknownPanics",
			false,
			nil,
			podcommon.States{StartupProbe: podcommon.StateBoolUnknown},
			"unsupported startup probe state 'unknown'",
			"",
			"",
			false,
		},
		{
			"ReadinessProbeUnknownPanics",
			false,
			nil,
			podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolUnknown,
			},
			"unsupported readiness probe state 'unknown'",
			"",
			"",
			false,
		},
		{
			"ContainerNotRunningAction",
			false,
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerWaiting,
			},
			"",
			"",
			"target container currently not running",
			true,
		},
		{
			"StartedUnknownAction",
			false,
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolUnknown,
			},
			"",
			"",
			"target container started status currently unknown",
			false,
		},
		{
			"ReadyUnknownAction",
			false,
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolTrue,
				Ready:          podcommon.StateBoolUnknown,
			},
			"",
			"",
			"target container ready status currently unknown",
			false,
		},
		{
			"ResUnknownAction",
			false,
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolTrue,
				Ready:          podcommon.StateBoolTrue,
				Resources:      podcommon.StateResourcesUnknown,
			},
			"",
			"unknown resources applied",
			"",
			true,
		},
		{
			"NeitherProbePresentPanics",
			false,
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			podcommon.States{
				StartupProbe:    podcommon.StateBoolFalse,
				ReadinessProbe:  podcommon.StateBoolFalse,
				Container:       podcommon.StateContainerRunning,
				Started:         podcommon.StateBoolFalse,
				Ready:           podcommon.StateBoolFalse,
				Resources:       podcommon.StateResourcesStartup,
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
			},
			"neither startup probe or readiness probe present",
			"",
			"",
			false,
		},
		{
			"NotStartedWithStartupResActionStartupProbe",
			false,
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			podcommon.States{
				StartupProbe:    podcommon.StateBoolTrue,
				ReadinessProbe:  podcommon.StateBoolFalse,
				Container:       podcommon.StateContainerRunning,
				Started:         podcommon.StateBoolFalse,
				Ready:           podcommon.StateBoolFalse,
				Resources:       podcommon.StateResourcesStartup,
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
				Resize:          podcommon.NewResizeState(podcommon.StateResizeNotStartedOrCompleted, ""),
			},
			"",
			"",
			"startup resources enacted",
			true,
		},
		{
			"NotStartedWithStartupResActionReadinessProbe",
			false,
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			podcommon.States{
				StartupProbe:    podcommon.StateBoolFalse,
				ReadinessProbe:  podcommon.StateBoolTrue,
				Container:       podcommon.StateContainerRunning,
				Started:         podcommon.StateBoolFalse,
				Ready:           podcommon.StateBoolFalse,
				Resources:       podcommon.StateResourcesStartup,
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
				Resize:          podcommon.NewResizeState(podcommon.StateResizeNotStartedOrCompleted, ""),
			},
			"",
			"",
			"startup resources enacted",
			true,
		},
		{
			"StartedWithStartupResAction",
			false,
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolTrue,
				Ready:          podcommon.StateBoolTrue,
				Resources:      podcommon.StateResourcesStartup,
			},
			"",
			"",
			"post-startup resources commanded",
			true,
		},
		{
			"NotStartedWithPostStartupResAction",
			false,
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolFalse,
				Ready:          podcommon.StateBoolFalse,
				Resources:      podcommon.StateResourcesPostStartup,
			},
			"",
			"",
			"startup resources commanded",
			true,
		},
		{
			"StartedWithPostStartupResAction",
			false,
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			podcommon.States{
				StartupProbe:    podcommon.StateBoolTrue,
				ReadinessProbe:  podcommon.StateBoolTrue,
				Container:       podcommon.StateContainerRunning,
				Started:         podcommon.StateBoolTrue,
				Ready:           podcommon.StateBoolTrue,
				Resources:       podcommon.StateResourcesPostStartup,
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
				Resize:          podcommon.NewResizeState(podcommon.StateResizeNotStartedOrCompleted, ""),
			},
			"",
			"",
			"post-startup resources enacted",
			true,
		},
		{
			"NotStartedWithUnknownResAction",
			true,
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolFalse,
				Ready:          podcommon.StateBoolFalse,
				Resources:      podcommon.StateResourcesUnknown,
			},
			"",
			"",
			"startup resources commanded (unknown resources applied)",
			true,
		},
		{
			"StartedWithUnknownResAction",
			true,
			func(m *podtest.MockStatus, run func()) {
				m.UpdateDefaultAndRun(run)
			},
			podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolTrue,
				Ready:          podcommon.StateBoolTrue,
				Resources:      podcommon.StateResourcesUnknown,
			},
			"",
			"",
			"post-startup resources commanded (unknown resources applied)",
			true,
		},
		{
			"NoActionPanics",
			false,
			nil,
			podcommon.States{
				StartupProbe:   podcommon.StateBoolTrue,
				ReadinessProbe: podcommon.StateBoolTrue,
				Container:      podcommon.StateContainerRunning,
				Started:        podcommon.StateBoolTrue,
				Ready:          podcommon.StateBoolTrue,
				Resources:      podcommon.StateResources("test"),
			},
			"no action to invoke",
			"",
			"",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			run := func() { statusUpdated = true }
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{ScaleWhenUnknownResources: tt.scaleWhenUnknownRes},
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
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
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
		configStatusMock,
		kubetest.NewMockPodHelper(nil),
	)

	buffer := bytes.Buffer{}
	_ = a.startedUnknownAction(contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(&buffer)).Build())
	assert.Contains(t, buffer.String(), "target container started status currently unknown")
	assert.False(t, statusUpdated)
}

func TestTargetContainerActionReadyUnknownAction(t *testing.T) {
	statusUpdated := false
	configStatusMock := podtest.NewMockStatusWithRun(
		func(status *podtest.MockStatus, run func()) { status.UpdateDefaultAndRun(run) },
		func() { statusUpdated = true },
	)
	a := newTargetContainerAction(
		controllercommon.ControllerConfig{},
		configStatusMock,
		kubetest.NewMockPodHelper(nil),
	)

	buffer := bytes.Buffer{}
	_ = a.readyUnknownAction(contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(&buffer)).Build())
	assert.Contains(t, buffer.String(), "target container ready status currently unknown")
	assert.False(t, statusUpdated)
}

func TestTargetContainerActionResUnknownAction(t *testing.T) {
	statusUpdated := false
	configStatusMock := podtest.NewMockStatusWithRun(
		func(status *podtest.MockStatus, run func()) { status.UpdateDefaultAndRun(run) },
		func() { statusUpdated = true },
	)
	a := newTargetContainerAction(
		controllercommon.ControllerConfig{},
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
	assert.ErrorContains(t, err, "unknown resources applied")
	assert.True(t, statusUpdated)
}

func TestTargetContainerActionNotStartedWithStartupResAction(t *testing.T) {
	tests := []struct {
		name             string
		states           podcommon.States
		wantErr          bool
		wantStatusUpdate bool
	}{
		{
			"Ok",
			podcommon.States{
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
				Resize:          podcommon.NewResizeState(podcommon.StateResizeNotStartedOrCompleted, ""),
			},
			false,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{},
				podtest.NewMockStatusWithRun(
					func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
					func() { statusUpdated = true },
				),
				nil,
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
				assert.NoError(t, err)
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
	}{
		{
			"UnableToPatchContainerResources",
			func(m *kubetest.MockPodHelper) {
				m.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&v1.Pod{}, errors.New(""))
			},
			"unable to patch container resources",
			false,
		},
		{
			"Ok",
			nil,
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{},
				podtest.NewMockStatusWithRun(
					func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
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
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantStatusUpdate {
				assert.True(t, statusUpdated)
			} else {
				assert.False(t, statusUpdated)
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
	}{
		{
			"UnableToPatchContainerResources",
			func(m *kubetest.MockPodHelper) {
				m.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&v1.Pod{}, errors.New(""))
			},
			"unable to patch container resources",
			false,
		},
		{
			"Ok",
			nil,
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{},
				podtest.NewMockStatusWithRun(
					func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
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
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantStatusUpdate {
				assert.True(t, statusUpdated)
			} else {
				assert.False(t, statusUpdated)
			}
		})
	}
}

func TestTargetContainerActionStartedWithPostStartupResAction(t *testing.T) {
	tests := []struct {
		name             string
		states           podcommon.States
		wantErr          bool
		wantStatusUpdate bool
	}{
		{
			"Ok",
			podcommon.States{
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
				Resize:          podcommon.NewResizeState(podcommon.StateResizeNotStartedOrCompleted, ""),
			},
			false,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{},
				podtest.NewMockStatusWithRun(
					func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
					func() { statusUpdated = true },
				),
				nil,
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
				assert.NoError(t, err)
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
	}{
		{
			"UnableToPatchContainerResources",
			func(m *kubetest.MockPodHelper) {
				m.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&v1.Pod{}, errors.New(""))
			},
			"unable to patch container resources",
			false,
		},
		{
			"Ok",
			nil,
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{},
				podtest.NewMockStatusWithRun(
					func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
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
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantStatusUpdate {
				assert.True(t, statusUpdated)
			} else {
				assert.False(t, statusUpdated)
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
	}{
		{
			"UnableToPatchContainerResources",
			func(m *kubetest.MockPodHelper) {
				m.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&v1.Pod{}, errors.New(""))
			},
			"unable to patch container resources",
			false,
		},
		{
			"Ok",
			nil,
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{},
				podtest.NewMockStatusWithRun(
					func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
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
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantStatusUpdate {
				assert.True(t, statusUpdated)
			} else {
				assert.False(t, statusUpdated)
			}
		})
	}
}

func TestTargetContainerActionProcessConfigEnacted(t *testing.T) {
	tests := []struct {
		name                 string
		configStatusMockFunc func(*podtest.MockStatus, func())
		states               podcommon.States
		wantPanicErrMsg      string
		wantErrMsg           string
		wantStatusUpdate     bool
		wantLogMsg           string
	}{
		{
			"ScaleNotYetCompletedInProgress",
			func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			podcommon.States{
				Resize: podcommon.NewResizeState(podcommon.StateResizeInProgress, ""),
			},
			"",
			"",
			true,
			"scale not yet completed - in progress",
		},
		{
			"ScaleNotYetCompletedDeferred",
			func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			podcommon.States{
				Resize: podcommon.NewResizeState(podcommon.StateResizeDeferred, "message"),
			},
			"",
			"",
			true,
			"scale not yet completed - deferred (message)",
		},
		{
			"ScaleFailedInfeasibleStateResourcesPostStartup",
			func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			podcommon.States{
				Resources: podcommon.StateResourcesPostStartup,
				Resize:    podcommon.NewResizeState(podcommon.StateResizeInfeasible, "message"),
			},
			"",
			"post-startup scale failed - infeasible (message)",
			true,
			"",
		},
		{
			"ScaleFailedInfeasibleStateResourcesStartup",
			func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			podcommon.States{
				Resources: podcommon.StateResourcesStartup,
				Resize:    podcommon.NewResizeState(podcommon.StateResizeInfeasible, "message"),
			},
			"",
			"startup scale failed - infeasible (message)",
			true,
			"",
		},
		{
			"ScaleFailedErrorStateResourcesPostStartup",
			func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			podcommon.States{
				Resources: podcommon.StateResourcesPostStartup,
				Resize:    podcommon.NewResizeState(podcommon.StateResizeError, "message"),
			},
			"",
			"post-startup scale failed - error (message)",
			true,
			"",
		},
		{
			"ScaleFailedErrorStateResourcesStartup",
			func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			podcommon.States{
				Resources: podcommon.StateResourcesStartup,
				Resize:    podcommon.NewResizeState(podcommon.StateResizeError, "message"),
			},
			"",
			"startup scale failed - error (message)",
			true,
			"",
		},
		{
			"UnknownResizeStatePanics",
			func(m *podtest.MockStatus, run func()) {},
			podcommon.States{
				Resize: podcommon.NewResizeState("unknown", ""),
			},
			"unknown resize state 'unknown'",
			"",
			false,
			"",
		},
		{
			string(podcommon.StateStatusResourcesIncomplete),
			func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			podcommon.States{
				StatusResources: podcommon.StateStatusResourcesIncomplete,
				Resize:          podcommon.NewResizeState(podcommon.StateResizeNotStartedOrCompleted, ""),
			},
			"",
			"",
			true,
			"target container current cpu and/or memory resources currently missing",
		},
		{
			string(podcommon.StateStatusResourcesContainerResourcesMismatch),
			func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			podcommon.States{
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMismatch,
				Resize:          podcommon.NewResizeState(podcommon.StateResizeNotStartedOrCompleted, ""),
			},
			"",
			"",
			true,
			"target container current cpu and/or memory resources currently don't match target container's 'requests'",
		},
		{
			string(podcommon.StateStatusResourcesUnknown),
			func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			podcommon.States{
				StatusResources: podcommon.StateStatusResourcesUnknown,
				Resize:          podcommon.NewResizeState(podcommon.StateResizeNotStartedOrCompleted, ""),
			},
			"",
			"",
			true,
			"target container current cpu and/or memory resources currently unknown",
		},
		{
			"UnknownStatusResourcesStatePanics",
			func(m *podtest.MockStatus, run func()) {},
			podcommon.States{
				StatusResources: podcommon.StateStatusResources("test"),
				Resize:          podcommon.NewResizeState(podcommon.StateResizeNotStartedOrCompleted, ""),
			},
			"unknown state 'test'",
			"",
			true,
			"",
		},
		{
			string(podcommon.StateStatusResourcesContainerResourcesMatch) + string(podcommon.StateResourcesPostStartup),
			func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			podcommon.States{
				Resources:       podcommon.StateResourcesPostStartup,
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
				Resize:          podcommon.NewResizeState(podcommon.StateResizeNotStartedOrCompleted, ""),
			},
			"",
			"",
			true,
			"post-startup resources enacted",
		},
		{
			string(podcommon.StateStatusResourcesContainerResourcesMatch) + string(podcommon.StateResourcesStartup),
			func(m *podtest.MockStatus, run func()) { m.UpdateDefaultAndRun(run) },
			podcommon.States{
				Resources:       podcommon.StateResourcesStartup,
				StatusResources: podcommon.StateStatusResourcesContainerResourcesMatch,
				Resize:          podcommon.NewResizeState(podcommon.StateResizeNotStartedOrCompleted, ""),
			},
			"",
			"",
			true,
			"startup resources enacted",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusUpdated := false
			a := newTargetContainerAction(
				controllercommon.ControllerConfig{},
				podtest.NewMockStatusWithRun(tt.configStatusMockFunc, func() { statusUpdated = true }),
				nil,
			)

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
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantStatusUpdate {
				assert.True(t, statusUpdated)
			} else {
				assert.False(t, statusUpdated)
			}
			if tt.wantLogMsg != "" {
				assert.Contains(t, buffer.String(), tt.wantLogMsg)
			}
		})
	}
}

func TestTargetContainerActionContainerResourceConfig(t *testing.T) {
	a := newTargetContainerAction(
		controllercommon.ControllerConfig{},
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
			m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(&v1.Pod{}, errors.New(""))
		})
		a := newTargetContainerAction(
			controllercommon.ControllerConfig{},
			mockStatus,
			nil,
		)

		buffer := bytes.Buffer{}
		a.updateStatus(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(&buffer)).Build(),
			&v1.Pod{},
			"",
			podcommon.States{},
			podcommon.StatusScaleStateNotApplicable,
			scaletest.NewMockConfigurations(nil),
			"",
		)
		assert.Contains(t, buffer.String(), "unable to update status")
	})
}
