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

func TestNewUpdates(t *testing.T) {
	updates := NewUpdates(NewConfigs(nil, nil))
	allUpdates := updates.AllUpdates()
	assert.Equal(t, 2, len(allUpdates))
	assert.Equal(t, v1.ResourceCPU, allUpdates[0].ResourceName())
	assert.Equal(t, v1.ResourceMemory, allUpdates[1].ResourceName())
}

func TestUpdatesApply(t *testing.T) {
}

func TestUpdatesRollback(t *testing.T) {
}

func TestUpdatesStatus(t *testing.T) {
}

func TestUpdatesString(t *testing.T) {
}
