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

package scale

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/metricscommon"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	Subsystem = "scale"
)

const (
	failureName             = "failure"
	commandedUnknownResName = "commanded_unknown_resources"
	durationName            = "duration_seconds"
)

var cName string

var (
	failure = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      failureName,
		Help:      "Number of scale failures (by scale direction, reason)",
	}, []string{metricscommon.ControllerNameLabelName, metricscommon.DirectionLabelName, metricscommon.ReasonLabelName})

	commandedUnknownRes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      commandedUnknownResName,
		Help:      "Number of scales commanded upon encountering unknown resources",
	}, []string{metricscommon.ControllerNameLabelName})

	duration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      durationName,
		Help:      "Scale duration (from commanded to enacted) in seconds (by scale direction, outcome)",
		Buckets:   []float64{1, 2, 4, 8, 16, 32, 64, 128},
	}, []string{metricscommon.ControllerNameLabelName, metricscommon.DirectionLabelName, metricscommon.OutcomeLabelName})
)

// allMetrics must include all metrics defined above.
var allMetrics = []prometheus.Collector{
	failure, commandedUnknownRes, duration,
}

func RegisterMetrics(registry metrics.RegistererGatherer, controllerName string) {
	cName = controllerName
	registry.MustRegister(allMetrics...)
}

func ResetMetrics() {
	metricscommon.ResetMetrics(allMetrics)
}

func Failure(direction metricscommon.Direction, reason string) prometheus.Counter {
	return failure.WithLabelValues(cName, string(direction), reason)
}

func CommandedUnknownRes() prometheus.Counter {
	return commandedUnknownRes.WithLabelValues(cName)
}

func Duration(direction metricscommon.Direction, outcome metricscommon.Outcome) prometheus.Observer {
	return duration.WithLabelValues(cName, string(direction), string(outcome))
}
