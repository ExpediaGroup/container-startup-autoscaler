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

package informercache

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/metricscommon"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	Subsystem = "informercache"
)

const (
	patchSyncPollName    = "patch_sync_poll"
	patchSyncTimeoutName = "patch_sync_timeout"
)

var (
	patchSyncPoll = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      patchSyncPollName,
		Help:      "Number of informer cache sync polls after a Kube API patch was performed",
		Buckets:   []float64{1, 2, 4, 8, 16, 32, 64},
	}, []string{})

	patchSyncTimeout = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      patchSyncTimeoutName,
		Help:      "Number of informer cache sync timeouts after a Kube API patch was performed (may result in inconsistent CSA status updates)",
	}, []string{})
)

// allMetrics must include all metrics defined above.
var allMetrics = []prometheus.Collector{
	patchSyncPoll,
	patchSyncTimeout,
}

func RegisterMetrics(registry metrics.RegistererGatherer) {
	registry.MustRegister(allMetrics...)
}

func ResetMetrics() {
	metricscommon.ResetMetrics(allMetrics)
}

func PatchSyncPoll() prometheus.Observer {
	return patchSyncPoll.WithLabelValues()
}

func PatchSyncTimeout() prometheus.Counter {
	return patchSyncTimeout.WithLabelValues()
}
