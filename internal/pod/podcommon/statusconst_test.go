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

package podcommon

import (
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/metricscommon"
	"github.com/stretchr/testify/assert"
)

func TestStatusScaleStateDirection(t *testing.T) {
	tests := []struct {
		name            string
		s               StatusScaleState
		wantPanicErrMsg string
		want            metricscommon.Direction
	}{
		{
			name: string(StatusScaleStateUpCommanded),
			s:    StatusScaleStateUpCommanded,
			want: metricscommon.DirectionUp,
		},
		{
			name: string(StatusScaleStateUpEnacted),
			s:    StatusScaleStateUpEnacted,
			want: metricscommon.DirectionUp,
		},
		{
			name: string(StatusScaleStateUpFailed),
			s:    StatusScaleStateUpFailed,
			want: metricscommon.DirectionUp,
		},
		{
			name: string(StatusScaleStateDownCommanded),
			s:    StatusScaleStateDownCommanded,
			want: metricscommon.DirectionDown,
		},
		{
			name: string(StatusScaleStateDownEnacted),
			s:    StatusScaleStateDownEnacted,
			want: metricscommon.DirectionDown,
		},
		{
			name: string(StatusScaleStateDownFailed),
			s:    StatusScaleStateDownFailed,
			want: metricscommon.DirectionDown,
		},
		{
			name:            "NotSupported",
			s:               StatusScaleStateNotApplicable,
			wantPanicErrMsg: "'notapplicable' not supported",
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
