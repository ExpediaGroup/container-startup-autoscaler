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

package kube

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContainerStatusNotPresentError(t *testing.T) {
	assert.Equal(t, ContainerStatusNotPresentError{}, NewContainerStatusNotPresentError())
}

func TestContainerStatusNotPresentErrorError(t *testing.T) {
	e := NewContainerStatusNotPresentError()
	assert.Equal(t, "container status not present", e.Error())
}

func TestNewContainerStatusResourcesNotPresentError(t *testing.T) {
	assert.Equal(t, ContainerStatusResourcesNotPresentError{}, NewContainerStatusResourcesNotPresentError())
}

func TestContainerStatusResourcesNotPresentErrorError(t *testing.T) {
	e := NewContainerStatusResourcesNotPresentError()
	assert.Equal(t, "container status resources not present", e.Error())
}
