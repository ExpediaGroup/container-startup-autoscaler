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

func TestNewUpdate(t *testing.T) {
	resourceName := v1.ResourceCPU
	update := NewUpdate(resourceName, nil)
	assert.Equal(t, resourceName, update.ResourceName())
}

func TestUpdateResourceName(t *testing.T) {
}

func TestUpdateStartupPodMutationFunc(t *testing.T) {
}

func TestUpdatePostStartupPodMutationFunc(t *testing.T) {
}
