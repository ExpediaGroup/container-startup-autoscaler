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
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/metricscommon"
	metricsscale "github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/scale"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

const (
	postPatchPauseSecs = 3
	timeFormatSecs     = "2006-01-02T15:04:05-0700"
	timeFormatMilli    = "2006-01-02T15:04:05.000-0700"
)

// status is the default implementation of podcommon.Status.
type status struct {
	recorder  record.EventRecorder
	podHelper kubecommon.PodHelper
}

func newStatus(
	recorder record.EventRecorder,
	podHelper kubecommon.PodHelper,
) *status {
	return &status{
		recorder:  recorder,
		podHelper: podHelper,
	}
}

// Update updates controller status by applying mutations to the supplied pod. The supplied pod is never mutated.
// Under specific circumstances, a pause is observed after the patch is applied to allow Kubelet time to react. Returns
// the new server representation of the pod.
func (s *status) Update(
	ctx context.Context,
	pod *v1.Pod,
	status string,
	states podcommon.States,
	statusScaleState podcommon.StatusScaleState,
	scaleConfigs scalecommon.Configurations,
	failReason string,
) (*v1.Pod, error) {
	if (statusScaleState == podcommon.StatusScaleStateUpFailed ||
		statusScaleState == podcommon.StatusScaleStateDownFailed) &&
		strings.TrimSpace(failReason) == "" {

		panic(errors.New("failReason not provided for failed scale state"))
	}

	shouldPause := false
	postPatchPauseCallback := func(pause bool) { shouldPause = pause }
	mutatePodFunc := s.podMutationFunc(
		ctx,
		status,
		states,
		statusScaleState,
		scaleConfigs,
		failReason,
		postPatchPauseCallback,
	)

	newPod, err := s.podHelper.Patch(ctx, pod, []func(*v1.Pod) (bool, error){mutatePodFunc}, false, true)
	if err != nil {
		return nil, common.WrapErrorf(err, "unable to patch pod")
	}

	if shouldPause {
		logging.Infof(ctx, logging.VDebug, "pausing for %d seconds after status patch", postPatchPauseSecs)
		time.Sleep(postPatchPauseSecs * time.Second)
	}

	return newPod, nil
}

// podMutationFunc returns a function that performs the actual work of updating controller status.
func (s *status) podMutationFunc(
	ctx context.Context,
	status string,
	states podcommon.States,
	statusScaleState podcommon.StatusScaleState,
	scaleConfigs scalecommon.Configurations,
	failReason string,
	postPatchPauseCallback func(pause bool),
) func(pod *v1.Pod) (bool, error) {
	return func(pod *v1.Pod) (bool, error) {
		shouldPause := false
		defer func() { postPatchPauseCallback(shouldPause) }()

		currentStat := NewEmptyStatusAnnotation()
		currentStatAnn, gotStatAnn := pod.Annotations[kubecommon.AnnotationStatus]
		if gotStatAnn {
			var err error
			currentStat, err = StatusAnnotationFromString(currentStatAnn)
			if err != nil {
				logging.Errorf(ctx, err, "unable to get status annotation from string (will ignore)")
				currentStat = NewEmptyStatusAnnotation()
				gotStatAnn = false
			}
		}

		statScale := NewEmptyStatusAnnotationScale(scaleConfigs.AllEnabledConfigurationsResourceNames())

		switch statusScaleState {
		case podcommon.StatusScaleStateNotApplicable:
			if gotStatAnn {
				// Preserve current status.
				statScale.LastCommanded = currentStat.Scale.LastCommanded
				statScale.LastEnacted = currentStat.Scale.LastEnacted
				statScale.LastFailed = currentStat.Scale.LastFailed
			}

		case podcommon.StatusScaleStateDownCommanded, podcommon.StatusScaleStateUpCommanded:
			statScale.LastCommanded = s.formattedNow(timeFormatMilli)
			statScale.LastEnacted = ""
			statScale.LastFailed = ""

			// Indicate a pause should be observed if commanding a scale after a previously failed scale. This is to
			// ensure Kubelet has time to update pod conditions which could otherwise lead to spurious Kube events
			// being emitted.
			if currentStat.Scale.LastFailed != "" {
				shouldPause = true
			}

			s.normalEvent(pod, eventReasonScaling, status)

		case podcommon.StatusScaleStateUnknownCommanded:
			statScale.LastCommanded = s.formattedNow(timeFormatMilli)
			statScale.LastEnacted = ""
			statScale.LastFailed = ""

			if currentStat.Scale.LastFailed != "" {
				shouldPause = true
			}

			metricsscale.CommandedUnknownRes().Inc()
			s.normalEvent(pod, eventReasonScaling, status)

		case podcommon.StatusScaleStateDownEnacted, podcommon.StatusScaleStateUpEnacted:
			if currentStat.Scale.LastCommanded == "" {
				// Detected enacted but wasn't previously commanded. This happens if container resources are already
				// correctly applied for the desired state e.g. admitting a pod with startup resources already
				// applied. Only set empty timestamps in this case.
				statScale.LastCommanded = ""
				statScale.LastEnacted = ""
				statScale.LastFailed = ""
			} else {
				statScale.LastCommanded = currentStat.Scale.LastCommanded
				statScale.LastEnacted = currentStat.Scale.LastEnacted
				statScale.LastFailed = ""

				// Only update if not already set.
				if !gotStatAnn || (gotStatAnn && currentStat.Scale.LastEnacted == "") {
					now := s.formattedNow(timeFormatMilli)
					statScale.LastEnacted = now
					s.updateDurationMetric(
						ctx,
						statusScaleState.Direction(), metricscommon.OutcomeSuccess,
						statScale.LastCommanded, now,
					)
					s.normalEvent(pod, eventReasonScaling, status)
				}
			}

		case podcommon.StatusScaleStateDownFailed, podcommon.StatusScaleStateUpFailed:
			statScale.LastCommanded = currentStat.Scale.LastCommanded
			statScale.LastEnacted = "" // Assumes can't fail after enacted.
			statScale.LastFailed = currentStat.Scale.LastFailed

			// Only update if not already set.
			if !gotStatAnn || (gotStatAnn && currentStat.Scale.LastFailed == "") {
				now := s.formattedNow(timeFormatMilli)
				statScale.LastFailed = now
				s.updateDurationMetric(
					ctx,
					statusScaleState.Direction(), metricscommon.OutcomeFailure,
					statScale.LastCommanded, now,
				)
				metricsscale.Failure(states.Resources.Direction(), failReason).Inc()
				s.warningEvent(pod, eventReasonScaling, status)
			}

		default:
			panic(fmt.Errorf("statusScaleState '%s' not supported", statusScaleState))
		}

		newStat := NewStatusAnnotation(common.CapitalizeFirstChar(status), statScale, s.formattedNow(timeFormatMilli))
		if gotStatAnn && newStat.Equal(currentStat) {
			logging.Infof(ctx, logging.VDebug, "status annotation not changed so will not patch")
			return false, nil
		}

		pod.Annotations[kubecommon.AnnotationStatus] = newStat.Json()
		return true, nil
	}
}

// formattedNow returns a time per format in UTC.
func (s *status) formattedNow(format string) string {
	return time.Now().UTC().Format(format)
}

// updateDurationMetric attempts to update the scale duration metric according to the supplied arguments.
func (s *status) updateDurationMetric(
	ctx context.Context,
	direction metricscommon.Direction,
	outcome metricscommon.Outcome,
	commanded string,
	now string,
) {
	if commanded == "" || now == "" {
		return
	}

	cTime, err := time.Parse(timeFormatMilli, commanded)
	if err != nil {
		logging.Errorf(ctx, err, "unable to parse commanded time '%s' as string (won't update metric)", commanded)
		return
	}

	nTime, err := time.Parse(timeFormatMilli, now)
	if err != nil {
		logging.Errorf(ctx, err, "unable to parse now time '%s' as string (won't update metric)", now)
		return
	}

	diffSecs := nTime.Sub(cTime).Seconds()
	if diffSecs < 0 {
		logging.Errorf(ctx, nil, "negative commanded/now seconds difference '%s'/'%s' (won't update metric)", commanded, now)
		return
	}

	metricsscale.Duration(direction, outcome).Observe(diffSecs)
}

// normalEvent yields a 'normal' Kube event for the supplied pod with the supplied reason and message.
func (s *status) normalEvent(pod *v1.Pod, reason string, message string) {
	s.recorder.Event(pod, v1.EventTypeNormal, reason, common.CapitalizeFirstChar(message))
}

// warningEvent yields a 'warning' Kube event for the supplied pod with the supplied reason and message.
func (s *status) warningEvent(pod *v1.Pod, reason string, message string) {
	s.recorder.Event(pod, v1.EventTypeWarning, reason, common.CapitalizeFirstChar(message))
}
