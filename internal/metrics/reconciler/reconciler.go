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

package reconciler

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/metricscommon"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	Subsystem = "reconciler"
)

const (
	skippedOnlyStatusChangeName = "skipped_only_status_change"
	existingInProgressName      = "existing_in_progress"
	failureName                 = "failure"
)

type FailureReason string

const (
	FailureReasonUnableToGetPod      = FailureReason("unable_to_get_pod")
	FailureReasonPodDoesNotExist     = FailureReason("pod_does_not_exist")
	FailureReasonConfiguration       = FailureReason("configuration")
	FailureReasonValidation          = FailureReason("validation")
	FailureReasonStatesDetermination = FailureReason("states_determination")
	FailureReasonStatesAction        = FailureReason("states_action")
)

var (
	skippedOnlyStatusChange = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      skippedOnlyStatusChangeName,
		Help:      "Number of reconciles that were skipped because only the status changed",
	}, []string{})

	existingInProgress = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      existingInProgressName,
		Help:      "Number of attempted reconciles where one was already in progress for the same namespace/name (results in a requeue)",
	}, []string{})

	failure = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      failureName,
		Help:      "Number of reconciles where there was a failure",
	}, []string{metricscommon.ReasonLabelName})
)

// allMetrics must include all metrics defined above.
var allMetrics = []prometheus.Collector{
	skippedOnlyStatusChange, existingInProgress, failure,
}

func RegisterMetrics(registry metrics.RegistererGatherer) {
	registry.MustRegister(allMetrics...)
}

func ResetMetrics() {
	metricscommon.ResetMetrics(allMetrics)
}

func SkippedOnlyStatusChange() prometheus.Counter {
	return skippedOnlyStatusChange.WithLabelValues()
}

func ExistingInProgress() prometheus.Counter {
	return existingInProgress.WithLabelValues()
}

func Failure(reason FailureReason) prometheus.Counter {
	return failure.WithLabelValues(string(reason))
}
