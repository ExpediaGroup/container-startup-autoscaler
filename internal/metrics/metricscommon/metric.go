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

package metricscommon

import (
	"github.com/prometheus/client_golang/prometheus"
)

// ResetMetrics resets all supplied metrics.
func ResetMetrics(metrics []prometheus.Collector) {
	for _, metric := range metrics {
		switch metric.(type) {
		case *prometheus.CounterVec:
			metric.(*prometheus.CounterVec).Reset()
		case *prometheus.GaugeVec:
			metric.(*prometheus.GaugeVec).Reset()
		case *prometheus.HistogramVec:
			metric.(*prometheus.HistogramVec).Reset()
		case *prometheus.SummaryVec:
			metric.(*prometheus.SummaryVec).Reset()
		}
	}
}
