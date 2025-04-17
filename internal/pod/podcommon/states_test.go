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

	"github.com/stretchr/testify/assert"
)

func TestNewStates(t *testing.T) {
	s := NewStates(
		StateBoolUnknown,
		StateBoolUnknown,
		StateContainerUnknown,
		StateBoolUnknown,
		StateBoolUnknown,
		StateResourcesUnknown,
		StateStatusResourcesUnknown,
		NewResizeState(StateResizeNotStartedOrCompleted, ""),
	)
	assert.Equal(t, StateBoolUnknown, s.StartupProbe)
	assert.Equal(t, StateBoolUnknown, s.ReadinessProbe)
	assert.Equal(t, StateContainerUnknown, s.Container)
	assert.Equal(t, StateBoolUnknown, s.Started)
	assert.Equal(t, StateBoolUnknown, s.Ready)
	assert.Equal(t, StateResourcesUnknown, s.Resources)
	assert.Equal(t, StateStatusResourcesUnknown, s.StatusResources)
	assert.Equal(t, NewResizeState(StateResizeNotStartedOrCompleted, ""), s.Resize)
}

func TestNewStatesAllUnknown(t *testing.T) {
	s := NewStatesAllUnknown()
	assert.Equal(t, StateBoolUnknown, s.StartupProbe)
	assert.Equal(t, StateBoolUnknown, s.ReadinessProbe)
	assert.Equal(t, StateContainerUnknown, s.Container)
	assert.Equal(t, StateBoolUnknown, s.Started)
	assert.Equal(t, StateBoolUnknown, s.Ready)
	assert.Equal(t, StateResourcesUnknown, s.Resources)
	assert.Equal(t, StateStatusResourcesUnknown, s.StatusResources)
	assert.Equal(t, NewResizeState(StateResizeUnknown, ""), s.Resize)
}

func TestNewResizeState(t *testing.T) {
	rs := NewResizeState(
		StateResizeUnknown,
		"test",
	)
	assert.Equal(t, StateResizeUnknown, rs.State)
	assert.Equal(t, "test", rs.Message)
}
