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

package controller

// TODO(wt) tests fixed here
import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contexttest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/controller/controllercommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/reconciler"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podtest"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/component-base/metrics/testutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestNewContainerStartupAutoscalerReconciler(t *testing.T) {
	p := &pod.Pod{}
	c := controllercommon.NewControllerConfig()
	r := newContainerStartupAutoscalerReconciler(p, c)
	assert.Equal(t, p, r.pod)
	assert.Equal(t, c, r.controllerConfig)
	assert.NotNil(t, r.reconcilingPods)
}

func TestContainerStartupAutoscalerReconcilerReconcile(t *testing.T) {
	type fields struct {
		controllerConfig controllercommon.ControllerConfig
	}
	type mocks struct {
		configuration         pod.Configuration
		validation            pod.Validation
		targetContainerState  pod.TargetContainerState
		targetContainerAction pod.TargetContainerAction
		podHelper             kube.PodHelper
	}
	tests := []struct {
		name                    string
		configMapFunc           func(cmap.ConcurrentMap[string, any], string)
		configMetricAssertsFunc func(t *testing.T)
		fields                  fields
		mocks                   mocks
		podNamespace            string
		podName                 string
		want                    reconcile.Result
		wantErrMsg              string
		wantLogMsg              string
		wantEmptyMap            bool
	}{
		{
			name: "ExistingReconcileInProgress",
			configMapFunc: func(cmap cmap.ConcurrentMap[string, any], podNamespacedName string) {
				cmap.Set(podNamespacedName, nil)
			},
			configMetricAssertsFunc: func(t *testing.T) {
				metricVal, _ := testutil.GetCounterMetricValue(reconciler.ExistingInProgress())
				assert.Equal(t, float64(1), metricVal)
			},
			fields:       fields{controllercommon.ControllerConfig{RequeueDurationSecs: 10}},
			podNamespace: "namespace",
			podName:      "name",
			want:         reconcile.Result{RequeueAfter: 10 * time.Second},
			wantLogMsg:   "existing reconcile in progress (will requeue)",
			wantEmptyMap: false,
		},
		{
			name:   "UnableToGetPod",
			fields: fields{controllercommon.ControllerConfig{RequeueDurationSecs: 10}},
			mocks: mocks{
				podHelper: kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) {
					m.On("Get", mock.Anything, mock.Anything).Return(false, &v1.Pod{}, errors.New(""))
				}),
			},
			podNamespace: "namespace",
			podName:      "name",
			want:         reconcile.Result{RequeueAfter: 10 * time.Second},
			wantLogMsg:   "unable to get pod (will requeue)",
			wantEmptyMap: true,
		},
		{
			name:   "PodDoesntExist",
			fields: fields{},
			mocks: mocks{
				podHelper: kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) {
					m.On("Get", mock.Anything, mock.Anything).Return(false, &v1.Pod{}, nil)
				}),
			},
			podNamespace: "namespace",
			podName:      "name",
			want:         reconcile.Result{},
			wantErrMsg:   "pod doesn't exist (won't requeue)",
			wantLogMsg:   "pod doesn't exist (won't requeue)",
			wantEmptyMap: true,
		},
		{
			name:   "UnableToConfigurePod",
			fields: fields{},
			mocks: mocks{
				configuration: podtest.NewMockConfiguration(func(m *podtest.MockConfiguration) {
					m.On("Configure", mock.Anything).Return(errors.New(""))
				}),
				podHelper: kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) { m.GetDefault() }),
			},
			podNamespace: "namespace",
			podName:      "name",
			want:         reconcile.Result{},
			wantErrMsg:   "unable to configure pod (won't requeue)",
			wantLogMsg:   "unable to configure pod (won't requeue)",
			wantEmptyMap: true,
		},
		{
			name:   "UnableToValidatePod",
			fields: fields{},
			mocks: mocks{ // TODO(wt) continue here - need to return mock scaleconfigs
				configuration: podtest.NewMockConfiguration(func(m *podtest.MockConfiguration) { m.ConfigureDefault() }),
				validation: podtest.NewMockValidation(func(m *podtest.MockValidation) {
					m.On("Validate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(errors.New(""))
				}),
				podHelper: kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) { m.GetDefault() }),
			},
			podNamespace: "namespace",
			podName:      "name",
			want:         reconcile.Result{},
			wantErrMsg:   "unable to validate pod (won't requeue)",
			wantLogMsg:   "unable to validate pod (won't requeue)",
			wantEmptyMap: true,
		},
		{
			name:   "UnableToDetermineTargetContainerStates",
			fields: fields{},
			mocks: mocks{
				configuration: podtest.NewMockConfiguration(func(m *podtest.MockConfiguration) { m.ConfigureDefault() }),
				validation:    podtest.NewMockValidation(func(m *podtest.MockValidation) { m.ValidateDefault() }),
				targetContainerState: podtest.NewMockTargetContainerState(func(m *podtest.MockTargetContainerState) {
					m.On("States", mock.Anything, mock.Anything, mock.Anything).
						Return(podcommon.States{}, errors.New(""))
				}),
				podHelper: kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) { m.GetDefault() }),
			},
			podNamespace: "namespace",
			podName:      "name",
			want:         reconcile.Result{},
			wantErrMsg:   "unable to determine target container states (won't requeue)",
			wantLogMsg:   "unable to determine target container states (won't requeue)",
			wantEmptyMap: true,
		},
		{
			name:   "UnableToActionTargetContainerStates",
			fields: fields{},
			mocks: mocks{
				configuration: podtest.NewMockConfiguration(func(m *podtest.MockConfiguration) { m.ConfigureDefault() }),
				validation:    podtest.NewMockValidation(func(m *podtest.MockValidation) { m.ValidateDefault() }),
				targetContainerState: podtest.NewMockTargetContainerState(func(m *podtest.MockTargetContainerState) {
					m.StatesDefault()
				}),
				targetContainerAction: podtest.NewMockTargetContainerAction(func(m *podtest.MockTargetContainerAction) {
					m.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(errors.New(""))
				}),
				podHelper: kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) { m.GetDefault() }),
			},
			podNamespace: "namespace",
			podName:      "name",
			want:         reconcile.Result{},
			wantErrMsg:   "unable to action target container states (won't requeue)",
			wantLogMsg:   "unable to action target container states (won't requeue)",
			wantEmptyMap: true,
		},
		{
			name:   "Ok",
			fields: fields{},
			mocks: mocks{
				configuration: podtest.NewMockConfiguration(func(m *podtest.MockConfiguration) { m.ConfigureDefault() }),
				validation:    podtest.NewMockValidation(func(m *podtest.MockValidation) { m.ValidateDefault() }),
				targetContainerState: podtest.NewMockTargetContainerState(func(m *podtest.MockTargetContainerState) {
					m.StatesDefault()
				}),
				targetContainerAction: podtest.NewMockTargetContainerAction(func(m *podtest.MockTargetContainerAction) {
					m.ExecuteDefault()
				}),
				podHelper: kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) { m.GetDefault() }),
			},
			podNamespace: "namespace",
			podName:      "name",
			want:         reconcile.Result{},
			wantEmptyMap: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			namespacedName := types.NamespacedName{
				Namespace: tt.podNamespace,
				Name:      tt.podName,
			}
			c := cmap.New[any]()
			if tt.configMapFunc != nil {
				tt.configMapFunc(c, namespacedName.String())
			}

			p := &pod.Pod{
				Validation:            tt.mocks.validation,
				TargetContainerState:  tt.mocks.targetContainerState,
				TargetContainerAction: tt.mocks.targetContainerAction,
				PodHelper:             tt.mocks.podHelper,
			}
			r := &containerStartupAutoscalerReconciler{
				pod:              p,
				controllerConfig: tt.fields.controllerConfig,
				reconcilingPods:  c,
			}

			buffer := &bytes.Buffer{}
			ctx := contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(buffer)).Build()
			got, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})

			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantEmptyMap, c.IsEmpty())
			if tt.configMetricAssertsFunc != nil {
				tt.configMetricAssertsFunc(t)
			}
		})
	}
}
