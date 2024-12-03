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

package retry

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/metricscommon"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	SubsystemKubeApi = "retrykubeapi"
)

const (
	retryName = "retry"
)

var cName string

var (
	retry = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricscommon.Namespace,
		Subsystem: SubsystemKubeApi,
		Name:      retryName,
		Help:      "Number of Kube API retries (by reason)",
	}, []string{metricscommon.ControllerNameLabelName, metricscommon.ReasonLabelName})
)

// allMetrics must include all metrics defined above.
var allMetrics = []prometheus.Collector{
	retry,
}

func RegisterKubeApiMetrics(registry metrics.RegistererGatherer, controllerName string) {
	cName = controllerName
	registry.MustRegister(allMetrics...)
}

func ResetKubeApiMetrics() {
	metricscommon.ResetMetrics(allMetrics)
}

func Retry(reason string) prometheus.Counter {
	return retry.WithLabelValues(cName, reason)
}
