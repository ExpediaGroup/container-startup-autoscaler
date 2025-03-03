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

func TestNewConfigs(t *testing.T) {
	configs := NewConfigs(nil, nil)
	allConfigs := configs.AllConfigs()
	assert.Equal(t, 2, len(allConfigs))
	assert.Equal(t, v1.ResourceCPU, allConfigs[0].ResourceName())
	assert.Equal(t, v1.ResourceMemory, allConfigs[1].ResourceName())
}

func TestConfigsTargetContainerName(t *testing.T) {
}

func TestConfigsStoreFromAnnotationsAll(t *testing.T) {
}

func TestConfigsValidateAll(t *testing.T) {
}

func TestConfigsValidateCollection(t *testing.T) {
}

func TestConfigsConfigFor(t *testing.T) {
}

func TestConfigsAllConfigs(t *testing.T) {
}

func TestConfigsAllEnabledConfigs(t *testing.T) {
}

func TestConfigsAllEnabledConfigsResourceNames(t *testing.T) {
}

func TestConfigsString(t *testing.T) {
}
