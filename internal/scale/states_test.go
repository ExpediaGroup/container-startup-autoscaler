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

package scale

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestNewStates(t *testing.T) {
	states := NewStates(NewConfigs(nil, nil), nil)
	allStates := states.AllStates()
	assert.Equal(t, 2, len(allStates))
	assert.Equal(t, v1.ResourceCPU, allStates[0].ResourceName())
	assert.Equal(t, v1.ResourceMemory, allStates[1].ResourceName())
}

func TestStatesIsStartupConfigAppliedAll(t *testing.T) {
}

func TestStatesIsPostStartupConfigAppliedAll(t *testing.T) {
}

func TestStatesIsAnyCurrentZeroAll(t *testing.T) {
}

func TestStatesDoesRequestsCurrentMatchSpecAll(t *testing.T) {
}

func TestStatesDoesLimitsCurrentMatchSpecAll(t *testing.T) {
}

func TestStatesStateFor(t *testing.T) {
}

func TestStatesAllStates(t *testing.T) {
}
