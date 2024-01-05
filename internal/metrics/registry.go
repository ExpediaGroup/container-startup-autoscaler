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

package metrics

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/reconciler"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/retry"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/scale"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// RegisterAllMetrics registers all metrics with the supplied registry.
func RegisterAllMetrics(registry metrics.RegistererGatherer, controllerName string) {
	reconciler.RegisterMetrics(registry, controllerName)
	retry.RegisterKubeApiMetrics(registry, controllerName)
	scale.RegisterMetrics(registry, controllerName)
}

// UnregisterAllMetrics unregisters all metrics within the supplied registry.
func UnregisterAllMetrics(registry metrics.RegistererGatherer) {
	reconciler.UnregisterMetrics(registry)
	retry.UnregisterKubeApiMetrics(registry)
	scale.UnregisterMetrics(registry)
}
