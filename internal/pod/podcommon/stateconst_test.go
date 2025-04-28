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

package podcommon

import (
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/metricscommon"
	"github.com/stretchr/testify/assert"
)

func TestStateBoolBool(t *testing.T) {
	tests := []struct {
		name string
		s    StateBool
		want bool
	}{
		{
			string(StateBoolTrue),
			StateBoolTrue,
			true,
		},
		{
			string(StateBoolFalse),
			StateBoolFalse,
			false,
		},
		{
			string(StateBoolUnknown),
			StateBoolFalse,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.s.Bool())
		})
	}
}

func TestStateResourcesDirection(t *testing.T) {
	tests := []struct {
		name            string
		s               StateResources
		wantPanicErrMsg string
		want            metricscommon.Direction
	}{
		{
			string(StateResourcesStartup),
			StateResourcesStartup,
			"",
			metricscommon.DirectionUp,
		},
		{
			string(StateResourcesPostStartup),
			StateResourcesPostStartup,
			"",
			metricscommon.DirectionDown,
		},
		{
			"NotSupported",
			StateResources("test"),
			"'test' not supported",
			metricscommon.Direction(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() { tt.s.Direction() })
				return
			}

			assert.Equal(t, tt.want, tt.s.Direction())
		})
	}
}

func TestStateResourcesHumanReadable(t *testing.T) {
	tests := []struct {
		name string
		s    StateResources
		want string
	}{
		{
			string(StateResourcesPostStartup),
			StateResourcesPostStartup,
			"post-startup",
		},
		{
			string(StateResourcesStartup),
			StateResourcesStartup,
			string(StateResourcesStartup),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.s.HumanReadable())
		})
	}
}
