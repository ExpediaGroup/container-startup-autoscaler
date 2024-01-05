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
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/metricscommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/reconciler"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/retry"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/scale"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAllMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	RegisterAllMetrics(registry, "test")

	gotReconciler, gotRetryKubeapi, gotScale := gotSubsystems(registry)
	assert.True(t, gotReconciler)
	assert.True(t, gotScale)
	assert.True(t, gotRetryKubeapi)
}

func TestUnregisterAllMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	RegisterAllMetrics(registry, "")
	UnregisterAllMetrics(registry)

	gotReconciler, gotRetryKubeapi, gotScale := gotSubsystems(registry)
	assert.False(t, gotReconciler)
	assert.False(t, gotScale)
	assert.False(t, gotRetryKubeapi)
}

func gotSubsystems(registry *prometheus.Registry) (bool, bool, bool) {
	descCh := make(chan *prometheus.Desc)
	doneCh := make(chan struct{})
	gotReconciler, gotRetryKubeapi, gotScale := false, false, false

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case desc := <-descCh:
				if strings.Contains(desc.String(), fmt.Sprintf("%s_%s", metricscommon.Namespace, reconciler.Subsystem)) {
					gotReconciler = true
				}

				if strings.Contains(desc.String(), fmt.Sprintf("%s_%s", metricscommon.Namespace, retry.SubsystemKubeApi)) {
					gotRetryKubeapi = true
				}

				if strings.Contains(desc.String(), fmt.Sprintf("%s_%s", metricscommon.Namespace, scale.Subsystem)) {
					gotScale = true
				}
			case <-doneCh:
				return
			}
		}
	}()

	registry.Describe(descCh)
	doneCh <- struct{}{}
	wg.Wait()
	return gotReconciler, gotRetryKubeapi, gotScale
}
