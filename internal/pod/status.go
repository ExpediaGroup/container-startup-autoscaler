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
	"fmt"
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/metricscommon"
	metricsscale "github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/scale"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"k8s.io/api/core/v1"
)

const (
	timeFormatSecs  = "2006-01-02T15:04:05-0700"
	timeFormatMilli = "2006-01-02T15:04:05.000-0700"
)

// status is the default implementation of podcommon.Status.
type status struct {
	podHelper kubecommon.PodHelper
}

func newStatus(podHelper kubecommon.PodHelper) *status {
	return &status{podHelper: podHelper}
}

// Update updates controller status by applying mutations to the supplied pod. The supplied pod is never mutated.
// Returns the new server representation of the pod.
func (s *status) Update(
	ctx context.Context,
	pod *v1.Pod,
	status string,
	states podcommon.States,
	statusScaleState podcommon.StatusScaleState,
	scaleConfigs scalecommon.Configurations,
) (*v1.Pod, error) {
	mutatePodFunc := s.PodMutationFunc(ctx, status, states, statusScaleState, scaleConfigs)

	newPod, err := s.podHelper.Patch(ctx, pod, []func(*v1.Pod) error{mutatePodFunc}, false, true)
	if err != nil {
		return nil, common.WrapErrorf(err, "unable to patch pod")
	}

	return newPod, nil
}

// PodMutationFunc returns a function that performs the actual work of updating controller status.
func (s *status) PodMutationFunc(
	ctx context.Context,
	status string,
	states podcommon.States,
	statusScaleState podcommon.StatusScaleState,
	scaleConfigs scalecommon.Configurations,
) func(pod *v1.Pod) error {
	return func(pod *v1.Pod) error {
		var currentStat StatusAnnotation
		currentStatAnn, gotStatAnn := pod.Annotations[kubecommon.AnnotationStatus]
		if gotStatAnn {
			var err error
			currentStat, err = StatusAnnotationFromString(currentStatAnn)
			if err != nil {
				return common.WrapErrorf(err, "unable to get status annotation from string")
			}
		}

		statScale := NewEmptyStatusAnnotationScale(scaleConfigs.AllEnabledConfigurationsResourceNames())

		switch statusScaleState {
		case podcommon.StatusScaleStateNotApplicable: // Preserve current status.
			if gotStatAnn {
				statScale.LastCommanded = currentStat.Scale.LastCommanded
				statScale.LastEnacted = currentStat.Scale.LastEnacted
				statScale.LastFailed = currentStat.Scale.LastFailed
			}
		case podcommon.StatusScaleStateDownCommanded, podcommon.StatusScaleStateUpCommanded:
			statScale.LastCommanded = s.formattedNow(timeFormatMilli) // i.e. others ""
		case podcommon.StatusScaleStateUnknownCommanded:
			statScale.LastCommanded = s.formattedNow(timeFormatMilli) // i.e. others ""
			metricsscale.CommandedUnknownRes().Inc()
		case podcommon.StatusScaleStateDownEnacted, podcommon.StatusScaleStateUpEnacted:
			statScale.LastCommanded = currentStat.Scale.LastCommanded
			statScale.LastEnacted = currentStat.Scale.LastEnacted

			// Only update if not already set.
			if !gotStatAnn || (gotStatAnn && currentStat.Scale.LastCommanded != "" && currentStat.Scale.LastEnacted == "") {
				now := s.formattedNow(timeFormatMilli)
				statScale.LastEnacted = now
				s.updateDurationMetric(
					ctx,
					statusScaleState.Direction(), metricscommon.OutcomeSuccess,
					statScale.LastCommanded, now,
				)
			}
		case podcommon.StatusScaleStateDownFailed, podcommon.StatusScaleStateUpFailed:
			statScale.LastCommanded = currentStat.Scale.LastCommanded
			statScale.LastFailed = currentStat.Scale.LastEnacted

			// Only update if not already set.
			if !gotStatAnn || (gotStatAnn && currentStat.Scale.LastFailed == "") {
				now := s.formattedNow(timeFormatMilli)
				statScale.LastFailed = now
				s.updateDurationMetric(
					ctx,
					statusScaleState.Direction(), metricscommon.OutcomeFailure,
					statScale.LastCommanded, now,
				)
			}
		default:
			panic(fmt.Errorf("statusScaleState '%s' not supported", statusScaleState))
		}

		newStat := NewStatusAnnotation(common.CapitalizeFirstChar(status), states, statScale, s.formattedNow(timeFormatSecs))
		if gotStatAnn && newStat.Equal(currentStat) {
			return nil
		}

		pod.Annotations[kubecommon.AnnotationStatus] = newStat.Json()
		return nil
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
		logging.Errorf(ctx, nil, "negative commanded/now seconds difference ('%s'/'%s') (won't update metric)", commanded, now)
		return
	}

	metricsscale.Duration(direction, outcome).Observe(diffSecs)
}
