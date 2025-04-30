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
	"fmt"
	"sync"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/metricscommon"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"k8s.io/component-base/metrics/testutil"
)

func TestRegisterMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	RegisterMetrics(registry)
	assert.Equal(t, len(allMetrics), len(descs(registry)))
}

func TestResetMetrics(t *testing.T) {
	SyncTimeout().Inc()
	value, _ := testutil.GetCounterMetricValue(SyncTimeout())
	assert.Equal(t, float64(1), value)
	ResetMetrics()

	value, _ = testutil.GetCounterMetricValue(SyncTimeout())
	assert.Equal(t, float64(0), value)
}

func TestPatchSyncTimeout(t *testing.T) {
	m := SyncTimeout()
	assert.Contains(
		t,
		m.Desc().String(),
		fmt.Sprintf("%s_%s_%s", metricscommon.Namespace, Subsystem, syncTimeoutName),
	)
}

func descs(registry *prometheus.Registry) []string {
	ch := make(chan *prometheus.Desc)
	done := make(chan struct{})
	var ret []string

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case desc := <-ch:
				ret = append(ret, desc.String())
			case <-done:
				return
			}
		}
	}()

	registry.Describe(ch)
	done <- struct{}{}
	wg.Wait()
	return ret
}
