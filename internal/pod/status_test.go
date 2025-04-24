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
	"encoding/json"
	"testing"
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contexttest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/metricscommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/scale"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scaletest"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"
	"k8s.io/component-base/metrics/testutil"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestNewStatus(t *testing.T) {
	recorder := &record.FakeRecorder{}
	podHelper := kube.NewPodHelper(nil)
	expected := &status{
		recorder:  recorder,
		podHelper: podHelper,
	}
	assert.Equal(t, expected, newStatus(recorder, podHelper))
}

func TestStatusUpdateCore(t *testing.T) {
	t.Run("NoFailReasonPanics", func(t *testing.T) {
		s := newStatus(nil, nil)

		fun := func() {
			_, _ = s.Update(
				nil,
				nil,
				"",
				podcommon.States{},
				podcommon.StatusScaleStateUpFailed,
				nil,
				" ",
			)
		}
		assert.PanicsWithError(t, "failReason not provided for failed scale state", fun)
	})

	t.Run("UnableToPatchPod", func(t *testing.T) {
		s := newStatus(
			&record.FakeRecorder{},
			kube.NewPodHelper(
				kubetest.ControllerRuntimeFakeClientWithKubeFake(
					func() *kubefake.Clientset { return kubefake.NewClientset() },
					func() interceptor.Funcs { return interceptor.Funcs{Patch: kubetest.InterceptorFuncPatchFail()} },
				),
			),
		)

		got, err := s.Update(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			kubetest.NewPodBuilder().Build(),
			"test",
			podcommon.States{},
			podcommon.StatusScaleStateNotApplicable,
			scaletest.NewMockConfigurations(nil),
			"",
		)
		assert.Nil(t, got)
		assert.ErrorContains(t, err, "unable to patch pod")
	})

	t.Run("UnableToGetStatusAnnotationFromString", func(t *testing.T) {
		s := newStatus(
			&record.FakeRecorder{},
			kube.NewPodHelper(
				kubetest.ControllerRuntimeFakeClientWithKubeFake(
					func() *kubefake.Clientset {
						return kubefake.NewClientset(kubetest.NewPodBuilder().Build())
					},
					func() interceptor.Funcs { return interceptor.Funcs{} },
				),
			),
		)

		_, err := s.Update(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			kubetest.NewPodBuilder().AdditionalAnnotations(map[string]string{kubecommon.AnnotationStatus: "test"}).Build(),
			"test",
			podcommon.States{},
			podcommon.StatusScaleStateNotApplicable,
			scaletest.NewMockConfigurations(nil),
			"",
		)
		assert.ErrorContains(t, err, "unable to get status annotation from string")
	})

	t.Run("OkNoPreviousStatus", func(t *testing.T) {
		pod := kubetest.NewPodBuilder().Build()
		s := newStatus(
			&record.FakeRecorder{},
			kube.NewPodHelper(
				kubetest.ControllerRuntimeFakeClientWithKubeFake(
					func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
					func() interceptor.Funcs { return interceptor.Funcs{} },
				),
			),
		)

		got, err := s.Update(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			kubetest.NewPodBuilder().Build(),
			"test",
			podcommon.States{},
			podcommon.StatusScaleStateNotApplicable,
			scaletest.NewMockConfigurations(nil),
			"",
		)
		assert.NoError(t, err)
		ann, gotAnn := got.Annotations[kubecommon.AnnotationStatus]
		assert.True(t, gotAnn)
		stat := &StatusAnnotation{}
		_ = json.Unmarshal([]byte(ann), stat)
		assert.Equal(t, "Test", stat.Status)
		assert.NotEmpty(t, stat.LastUpdated)

		// Ensure pod isn't mutated
		_, gotAnn = pod.Annotations[kubecommon.AnnotationStatus]
		assert.False(t, gotAnn)
	})

	t.Run("OkPreviousStatusSame", func(t *testing.T) {
		s := newStatus(
			&record.FakeRecorder{},
			kube.NewPodHelper(
				kubetest.ControllerRuntimeFakeClientWithKubeFake(
					func() *kubefake.Clientset {
						return kubefake.NewClientset(kubetest.NewPodBuilder().Build())
					},
					func() interceptor.Funcs { return interceptor.Funcs{} },
				),
			),
		)

		previousStat := NewStatusAnnotation(
			"Test",
			podcommon.States{},
			NewEmptyStatusAnnotationScale([]v1.ResourceName{v1.ResourceCPU}),
			"",
		).Json()
		got, err := s.Update(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			kubetest.NewPodBuilder().AdditionalAnnotations(map[string]string{kubecommon.AnnotationStatus: previousStat}).Build(),
			"test",
			podcommon.States{},
			podcommon.StatusScaleStateNotApplicable,
			scaletest.NewMockConfigurations(nil),
			"",
		)
		assert.NoError(t, err)

		stat := &StatusAnnotation{}
		_ = json.Unmarshal([]byte(got.Annotations[kubecommon.AnnotationStatus]), stat)
		assert.Empty(t, stat.LastUpdated)
	})
}

func TestStatusUpdateScaleStatus(t *testing.T) {
	type args struct {
		pod        *v1.Pod
		scaleState podcommon.StatusScaleState
		failReason string
	}
	tests := []struct {
		name                    string
		args                    args
		wantPanicErrMsg         string
		wantLastScaleCommanded  bool
		wantLastScaleEnacted    bool
		wantLastScaleFailed     bool
		wantEventMsg            string
		wantPause               bool
		configMetricAssertsFunc func(t *testing.T)
	}{
		{
			"StatusScaleStateNotApplicablePrevious",
			args{
				kubetest.NewPodBuilder().
					AdditionalAnnotations(map[string]string{kubecommon.AnnotationStatus: fullStatusAnnotationString()}).
					Build(),
				podcommon.StatusScaleStateNotApplicable,
				"",
			},
			"",
			true,
			true,
			true,
			"",
			false,
			nil,
		},
		{
			"StatusScaleStateCommandedNoPreviousFail",
			args{
				kubetest.NewPodBuilder().Build(),
				podcommon.StatusScaleStateUpCommanded,
				"",
			},
			"",
			true,
			false,
			false,
			"Normal Scaling Test",
			false,
			nil,
		},
		{
			"StatusScaleStateCommandedPreviousFail",
			args{
				kubetest.NewPodBuilder().
					AdditionalAnnotations(map[string]string{kubecommon.AnnotationStatus: fullStatusAnnotationString()}).
					Build(),
				podcommon.StatusScaleStateUpCommanded,
				"",
			},
			"",
			true,
			false,
			false,
			"Normal Scaling Test",
			true,
			nil,
		},
		{
			"StatusScaleStateUnknownCommandedNoPreviousFail",
			args{
				kubetest.NewPodBuilder().Build(),
				podcommon.StatusScaleStateUnknownCommanded,
				"",
			},
			"",
			true,
			false,
			false,
			"Normal Scaling Test",
			false,
			func(t *testing.T) {
				metricVal, _ := testutil.GetCounterMetricValue(scale.CommandedUnknownRes())
				assert.Equal(t, float64(1), metricVal)
			},
		},
		{
			"StatusScaleStateUnknownCommandedPreviousFail",
			args{
				kubetest.NewPodBuilder().
					AdditionalAnnotations(map[string]string{kubecommon.AnnotationStatus: fullStatusAnnotationString()}).
					Build(),
				podcommon.StatusScaleStateUnknownCommanded,
				"",
			},
			"",
			true,
			false,
			false,
			"Normal Scaling Test",
			true,
			func(t *testing.T) {
				metricVal, _ := testutil.GetCounterMetricValue(scale.CommandedUnknownRes())
				assert.Equal(t, float64(1), metricVal)
			},
		},
		{
			"StatusScaleStateEnactedNoPrevious",
			args{
				kubetest.NewPodBuilder().
					AdditionalAnnotations(map[string]string{kubecommon.AnnotationStatus: statusAnnotationString(true, false, false)}).
					Build(),
				podcommon.StatusScaleStateUpEnacted,
				"",
			},
			"",
			true,
			true,
			false,
			"Normal Scaling Test",
			false,
			func(t *testing.T) {
				metricVal, _ := testutil.GetHistogramMetricCount(scale.Duration(metricscommon.DirectionUp, metricscommon.OutcomeSuccess))
				assert.Equal(t, uint64(1), metricVal)
			},
		},
		{
			"StatusScaleStateEnactedPrevious",
			args{
				kubetest.NewPodBuilder().
					AdditionalAnnotations(map[string]string{kubecommon.AnnotationStatus: fullStatusAnnotationString()}).
					Build(),
				podcommon.StatusScaleStateUpEnacted,
				"",
			},
			"",
			true,
			true,
			false,
			"",
			false,
			func(t *testing.T) {
				metricVal, _ := testutil.GetHistogramMetricCount(scale.Duration(metricscommon.DirectionUp, metricscommon.OutcomeSuccess))
				assert.Equal(t, uint64(0), metricVal)
			},
		},
		{
			"StatusScaleStateFailedNoPrevious",
			args{
				kubetest.NewPodBuilder().
					AdditionalAnnotations(map[string]string{kubecommon.AnnotationStatus: statusAnnotationString(true, false, false)}).
					Build(),
				podcommon.StatusScaleStateUpFailed,
				"failReason",
			},
			"",
			true,
			false,
			true,
			"Warning Scaling Test",
			false,
			func(t *testing.T) {
				durationMetricVal, _ := testutil.GetHistogramMetricCount(scale.Duration(metricscommon.DirectionUp, metricscommon.OutcomeFailure))
				assert.Equal(t, uint64(1), durationMetricVal)
				failureMetricVal, _ := testutil.GetCounterMetricValue(scale.Failure(metricscommon.DirectionUp, "failReason"))
				assert.Equal(t, float64(1), failureMetricVal)
			},
		},
		{
			"StatusScaleStateFailedPrevious",
			args{
				kubetest.NewPodBuilder().
					AdditionalAnnotations(map[string]string{kubecommon.AnnotationStatus: fullStatusAnnotationString()}).
					Build(),
				podcommon.StatusScaleStateUpFailed,
				"failReason",
			},
			"",
			true,
			false,
			true,
			"",
			false,
			func(t *testing.T) {
				durationMetricVal, _ := testutil.GetHistogramMetricCount(scale.Duration(metricscommon.DirectionUp, metricscommon.OutcomeFailure))
				assert.Equal(t, uint64(0), durationMetricVal)
				failureMetricVal, _ := testutil.GetCounterMetricValue(scale.Failure(metricscommon.DirectionUp, "failReason"))
				assert.Equal(t, float64(0), failureMetricVal)
			},
		},
		{
			"StatusScaleStateNotSupported",
			args{
				&v1.Pod{},
				podcommon.StatusScaleState("test"),
				"",
			},
			"statusScaleState 'test' not supported",
			false,
			false,
			false,
			"",
			false,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scale.ResetMetrics()

			eventRecorder := record.NewFakeRecorder(1)
			s := newStatus(
				eventRecorder,
				kube.NewPodHelper(
					kubetest.ControllerRuntimeFakeClientWithKubeFake(
						func() *kubefake.Clientset { return kubefake.NewClientset(kubetest.NewPodBuilder().Build()) },
						func() interceptor.Funcs { return interceptor.Funcs{} },
					),
				),
			)
			ctx := contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build()

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() {
					_, _ = s.Update(
						ctx,
						tt.args.pod,
						"test",
						podcommon.States{Resources: podcommon.StateResourcesStartup},
						tt.args.scaleState,
						scaletest.NewMockConfigurations(nil),
						tt.args.failReason,
					)
				})
				return
			}

			currentTimeMillis := time.Now().UnixMilli()
			got, err := s.Update(
				ctx,
				tt.args.pod,
				"test",
				podcommon.States{Resources: podcommon.StateResourcesStartup},
				tt.args.scaleState,
				scaletest.NewMockConfigurations(nil),
				tt.args.failReason,
			)
			durationMillis := time.Now().UnixMilli() - currentTimeMillis
			assert.NoError(t, err)

			stat := &StatusAnnotation{}
			_ = json.Unmarshal([]byte(got.Annotations[kubecommon.AnnotationStatus]), stat)
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
			if tt.wantEventMsg != "" {
				select {
				case res := <-eventRecorder.Events:
					assert.Contains(t, res, tt.wantEventMsg)
				case <-time.After(5 * time.Second):
					t.Fatalf("event not generated")
				}
			} else {
				select {
				case <-eventRecorder.Events:
					t.Fatalf("event unexpectedly generated")
				case <-time.After(1 * time.Second):
				}
			}
			if tt.wantPause {
				assert.GreaterOrEqual(t, durationMillis, int64(postPatchPauseSecs*1000))
			} else {
				assert.Less(t, durationMillis, int64(postPatchPauseSecs*1000))
			}
			if tt.configMetricAssertsFunc != nil {
				tt.configMetricAssertsFunc(t)
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
			"Empty",
			func(t *testing.T) {},
			args{
				"",
				"",
			},
			"",
		},
		{
			"UnableToParseCommandedTime",
			func(t *testing.T) {},
			args{
				"test",
				"test",
			},
			"unable to parse commanded time",
		},
		{
			"UnableToParseNowTime",
			func(t *testing.T) {},
			args{
				"2023-01-01T00:00:00.000-0000",
				"test",
			},
			"unable to parse now time",
		},
		{
			"NegativeDiff",
			func(t *testing.T) {},
			args{
				"2023-01-01T00:01:00.000-0000",
				"2023-01-01T00:00:00.000-0000",
			},
			"negative commanded/now seconds difference",
		},
		{
			"Ok",
			func(t *testing.T) {
				metricVal, _ := testutil.GetHistogramMetricCount(scale.Duration(metricscommon.DirectionUp, metricscommon.OutcomeSuccess))
				assert.Equal(t, uint64(1), metricVal)
			},
			args{
				"2023-01-01T00:00:00.000-0000",
				"2023-01-01T00:01:00.000-0000",
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newStatus(&record.FakeRecorder{}, nil)
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
	return statusAnnotationString(true, true, true)
}

func statusAnnotationString(lastCommanded bool, lastEnacted bool, lastFailed bool) string {
	now := newStatus(&record.FakeRecorder{}, nil).formattedNow(timeFormatMilli)

	lastCommandedString, lastEnactedString, lastFailedString := "", "", ""

	if lastCommanded {
		lastCommandedString = now
	}

	if lastEnacted {
		lastEnactedString = now
	}

	if lastFailed {
		lastFailedString = now
	}

	return NewStatusAnnotation(
		"test",
		podcommon.States{},
		NewStatusAnnotationScale([]v1.ResourceName{v1.ResourceCPU}, lastCommandedString, lastEnactedString, lastFailedString),
		now,
	).Json()
}
