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
	skippedOnlyStatusChangeName    = "skipped_only_status_change"
	existingInProgressName         = "existing_in_progress"
	failureUnableToGetPodName      = "failure_unable_to_get_pod"
	failurePodDoesntExistName      = "failure_pod_doesnt_exist"
	failureValidationName          = "failure_validation"
	failureStatesDeterminationName = "failure_states_determination"
	failureStatesActionName        = "failure_states_action"
)

var (
	skippedOnlyStatusChange = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      skippedOnlyStatusChangeName,
		Help:      "Number of reconciles that were skipped because only the scaler controller status changed",
	}, []string{})

	existingInProgress = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      existingInProgressName,
		Help:      "Number of attempted reconciles where one was already in progress for the same namespace/name (results in a requeue)",
	}, []string{})

	failureUnableToGetPod = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      failureUnableToGetPodName,
		Help:      "Number of reconciles where there was a failure to get the pod (results in a requeue)",
	}, []string{})

	failurePodDoesntExist = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      failurePodDoesntExistName,
		Help:      "Number of reconciles where the pod was found not to exist (results in failure)",
	}, []string{})

	failureValidation = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      failureValidationName,
		Help:      "Number of reconciles where there was a failure to validate (results in failure)",
	}, []string{})

	failureStatesDetermination = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      failureStatesDeterminationName,
		Help:      "Number of reconciles where there was a failure to determine states (results in failure)",
	}, []string{})

	failureStatesAction = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      failureStatesActionName,
		Help:      "Number of reconciles where there was a failure to action the determined states (results in failure)",
	}, []string{})
)

// allMetrics must include all metrics defined above.
var allMetrics = []prometheus.Collector{
	skippedOnlyStatusChange, existingInProgress, failureUnableToGetPod, failurePodDoesntExist, failureValidation,
	failureStatesDetermination, failureStatesAction,
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

func FailureUnableToGetPod() prometheus.Counter {
	return failureUnableToGetPod.WithLabelValues()
}

func FailurePodDoesntExist() prometheus.Counter {
	return failurePodDoesntExist.WithLabelValues()
}

func FailureValidation() prometheus.Counter {
	return failureValidation.WithLabelValues()
}

func FailureStatesDetermination() prometheus.Counter {
	return failureStatesDetermination.WithLabelValues()
}

func FailureStatesAction() prometheus.Counter {
	return failureStatesAction.WithLabelValues()
}
