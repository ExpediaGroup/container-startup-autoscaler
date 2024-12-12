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
	"encoding/json"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contexttest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/metricscommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/scale"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podtest"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/component-base/metrics/testutil"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestStatusUpdateCore(t *testing.T) {
	t.Run("UnableToPatchPod", func(t *testing.T) {
		s := newStatus(newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset() },
			func() interceptor.Funcs { return interceptor.Funcs{Patch: podtest.InterceptorFuncPatchFail()} },
		)))

		got, err := s.Update(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
			"test",
			podcommon.States{},
			podcommon.StatusScaleStateNotApplicable,
		)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "unable to patch pod")
	})

	t.Run("UnableToGetStatusAnnotationFromString", func(t *testing.T) {
		s := newStatus(newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset {
				return kubefake.NewClientset(
					podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				)
			},
			func() interceptor.Funcs { return interceptor.Funcs{} },
		)))

		_, err := s.Update(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
				AdditionalAnnotations(map[string]string{podcommon.AnnotationStatus: "test"}).
				Build(),
			"test",
			podcommon.States{},
			podcommon.StatusScaleStateNotApplicable,
		)
		assert.Contains(t, err.Error(), "unable to get status annotation from string")
	})

	t.Run("OkNoPreviousStatus", func(t *testing.T) {
		pod := podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		s := newStatus(newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		)))

		got, err := s.Update(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
			"test",
			podcommon.States{},
			podcommon.StatusScaleStateNotApplicable,
		)
		assert.Nil(t, err)
		ann, gotAnn := got.Annotations[podcommon.AnnotationStatus]
		assert.True(t, gotAnn)
		stat := &podcommon.StatusAnnotation{}
		_ = json.Unmarshal([]byte(ann), stat)
		assert.Equal(t, "Test", stat.Status)
		assert.NotEmpty(t, stat.LastUpdated)

		// Ensure pod isn't mutated
		_, gotAnn = pod.Annotations[podcommon.AnnotationStatus]
		assert.False(t, gotAnn)
	})

	t.Run("OkPreviousStatusSame", func(t *testing.T) {
		s := newStatus(newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset {
				return kubefake.NewClientset(
					podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				)
			},
			func() interceptor.Funcs { return interceptor.Funcs{} },
		)))

		previousStat := podcommon.NewStatusAnnotation(
			"Test",
			podcommon.States{},
			podcommon.NewEmptyStatusAnnotationScale(),
			"",
		).Json()
		got, err := s.Update(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
				AdditionalAnnotations(map[string]string{podcommon.AnnotationStatus: previousStat}).
				Build(),
			"test",
			podcommon.States{},
			podcommon.StatusScaleStateNotApplicable,
		)
		assert.Nil(t, err)

		stat := &podcommon.StatusAnnotation{}
		_ = json.Unmarshal([]byte(got.Annotations[podcommon.AnnotationStatus]), stat)
		assert.Empty(t, stat.LastUpdated)
	})
}

func TestStatusUpdateScaleStatus(t *testing.T) {
	type args struct {
		pod        *v1.Pod
		scaleState podcommon.StatusScaleState
	}
	tests := []struct {
		name                   string
		args                   args
		wantPanicErrMsg        string
		wantLastScaleCommanded bool
		wantLastScaleEnacted   bool
		wantLastScaleFailed    bool
	}{
		{
			name: "StatusScaleStateNotApplicablePrevious",
			args: args{
				podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
					AdditionalAnnotations(map[string]string{podcommon.AnnotationStatus: fullStatusAnnotationString()}).
					Build(),
				podcommon.StatusScaleStateNotApplicable,
			},
			wantLastScaleCommanded: true,
			wantLastScaleEnacted:   true,
			wantLastScaleFailed:    true,
		},
		{
			name: "StatusScaleStateCommanded",
			args: args{
				podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				podcommon.StatusScaleStateUpCommanded,
			},
			wantLastScaleCommanded: true,
			wantLastScaleEnacted:   false,
			wantLastScaleFailed:    false,
		},
		{
			name: "StatusScaleStateUnknownCommanded",
			args: args{
				podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				podcommon.StatusScaleStateUnknownCommanded,
			},
			wantLastScaleCommanded: true,
			wantLastScaleEnacted:   false,
			wantLastScaleFailed:    false,
		},
		{
			name: "StatusScaleStateEnactedNoPrevious",
			args: args{
				podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				podcommon.StatusScaleStateUpEnacted,
			},
			wantLastScaleCommanded: false,
			wantLastScaleEnacted:   true,
			wantLastScaleFailed:    false,
		},
		{
			name: "StatusScaleStateEnactedPrevious",
			args: args{
				podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
					AdditionalAnnotations(map[string]string{podcommon.AnnotationStatus: fullStatusAnnotationString()}).
					Build(),
				podcommon.StatusScaleStateUpEnacted,
			},
			wantLastScaleCommanded: true,
			wantLastScaleEnacted:   true,
			wantLastScaleFailed:    false,
		},
		{
			name: "StatusScaleStateFailedNoPrevious",
			args: args{
				podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				podcommon.StatusScaleStateUpFailed,
			},
			wantLastScaleCommanded: false,
			wantLastScaleEnacted:   false,
			wantLastScaleFailed:    true,
		},
		{
			name: "StatusScaleStateFailedPrevious",
			args: args{
				podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
					AdditionalAnnotations(map[string]string{podcommon.AnnotationStatus: fullStatusAnnotationString()}).
					Build(),
				podcommon.StatusScaleStateUpFailed,
			},
			wantLastScaleCommanded: true,
			wantLastScaleEnacted:   false,
			wantLastScaleFailed:    true,
		},
		{
			name: "ScaleStateNotSupported",
			args: args{
				&v1.Pod{},
				podcommon.StatusScaleState("test"),
			},
			wantPanicErrMsg: "scaleState 'test' not supported",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newStatus(newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
				func() *kubefake.Clientset {
					return kubefake.NewClientset(
						podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
					)
				},
				func() interceptor.Funcs { return interceptor.Funcs{} },
			)))
			ctx := contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build()

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() {
					_, _ = s.Update(
						ctx,
						tt.args.pod,
						"test",
						podcommon.States{},
						tt.args.scaleState,
					)
				})
				return
			}

			got, err := s.Update(
				ctx,
				tt.args.pod,
				"test",
				podcommon.States{},
				tt.args.scaleState,
			)
			assert.Nil(t, err)

			stat := &podcommon.StatusAnnotation{}
			_ = json.Unmarshal([]byte(got.Annotations[podcommon.AnnotationStatus]), stat)
			if tt.wantLastScaleCommanded {
				assert.NotEmpty(t, stat.Scale.LastCommanded)
			} else {
				assert.Empty(t, stat.Scale.LastCommanded)
			}
			if tt.wantLastScaleEnacted {
				assert.NotEmpty(t, stat.Scale.LastEnacted)
			} else {
				assert.Empty(t, stat.Scale.LastEnacted)
			}
			if tt.wantLastScaleFailed {
				assert.NotEmpty(t, stat.Scale.LastFailed)
			} else {
				assert.Empty(t, stat.Scale.LastFailed)
			}
		})
	}
}

func TestStatusUpdateDurationMetric(t *testing.T) {
	type args struct {
		commanded string
		now       string
	}
	tests := []struct {
		name                    string
		configMetricAssertsFunc func(t *testing.T)
		args                    args
		wantLogMsg              string
	}{
		{
			name:       "Empty",
			args:       args{},
			wantLogMsg: "",
		},
		{
			name: "UnableToParseCommandedTime",
			args: args{
				commanded: "test",
				now:       "test",
			},
			wantLogMsg: "unable to parse commanded time",
		},
		{
			name: "UnableToParseNowTime",
			args: args{
				commanded: "2023-01-01T00:00:00.000-0000",
				now:       "test",
			},
			wantLogMsg: "unable to parse now time",
		},
		{
			name: "NegativeDiff",
			args: args{
				commanded: "2023-01-01T00:01:00.000-0000",
				now:       "2023-01-01T00:00:00.000-0000",
			},
			wantLogMsg: "negative commanded/now seconds difference",
		},
		{
			name: "Ok",
			configMetricAssertsFunc: func(t *testing.T) {
				metricVal, _ := testutil.GetHistogramMetricCount(scale.Duration(metricscommon.DirectionUp, metricscommon.OutcomeSuccess))
				assert.Equal(t, uint64(1), metricVal)
			},
			args: args{
				commanded: "2023-01-01T00:00:00.000-0000",
				now:       "2023-01-01T00:01:00.000-0000",
			},
			wantLogMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newStatus(nil)
			buffer := &bytes.Buffer{}
			ctx := contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(buffer)).Build()
			s.updateDurationMetric(
				ctx,
				metricscommon.DirectionUp, metricscommon.OutcomeSuccess,
				tt.args.commanded,
				tt.args.now,
			)
			assert.Contains(t, buffer.String(), tt.wantLogMsg)
			if tt.configMetricAssertsFunc != nil {
				tt.configMetricAssertsFunc(t)
			}
		})
	}
}

func fullStatusAnnotationString() string {
	now := newStatus(nil).formattedNow(timeFormatMilli)

	return podcommon.NewStatusAnnotation(
		"test",
		podcommon.States{},
		podcommon.NewStatusAnnotationScale(now, now, now),
		now,
	).Json()
}
