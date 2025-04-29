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
	syncTimeoutName = "sync_timeout"
)

var (
	syncTimeout = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: Subsystem,
		Name:      syncTimeoutName,
		Help:      "Number of informer cache sync timeouts after a pod mutation was performed via the Kube API (may result in inconsistent CSA status updates)",
	}, []string{})
)

// allMetrics must include all metrics defined above.
var allMetrics = []prometheus.Collector{
	syncTimeout,
}

func RegisterMetrics(registry metrics.RegistererGatherer) {
	registry.MustRegister(allMetrics...)
}

func ResetMetrics() {
	metricscommon.ResetMetrics(allMetrics)
}

func SyncTimeout() prometheus.Counter {
	return syncTimeout.WithLabelValues()
}
