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

package pod

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("test", errors.New(""))
	assert.True(t, errors.Is(err, ValidationError{}))
}

func TestValidationErrorError(t *testing.T) {
	t.Run("NotNilCause", func(t *testing.T) {
		e := NewValidationError("test1", errors.New("test2"))
		assert.Equal(t, "validation error: test1: test2", e.Error())
	})

	t.Run("NilCause", func(t *testing.T) {
		e := NewValidationError("test", nil)
		assert.Equal(t, "validation error: test", e.Error())
	})
}

func TestValidationErrorCause(t *testing.T) {
	t.Run("NotNilCause", func(t *testing.T) {
		e := ValidationError{wrapped: errors.New("")}
		assert.NotNil(t, e.Cause())
	})

	t.Run("NilCause", func(t *testing.T) {
		e := ValidationError{wrapped: nil}
		assert.Nil(t, e.Cause())
	})
}

func TestNewContainerStatusNotPresentError(t *testing.T) {
	assert.True(t, errors.Is(NewContainerStatusNotPresentError(), ContainerStatusNotPresentError{}))
}

func TestContainerStatusNotPresentErrorError(t *testing.T) {
	e := NewContainerStatusNotPresentError()
	assert.Equal(t, "container status not present", e.Error())
}

func TestNewContainerStatusAllocatedResourcesNotPresentError(t *testing.T) {
	assert.True(t, errors.Is(NewContainerStatusAllocatedResourcesNotPresentError(), ContainerStatusAllocatedResourcesNotPresentError{}))
}

func TestContainerStatusAllocatedResourcesNotPresentErrorError(t *testing.T) {
	e := NewContainerStatusAllocatedResourcesNotPresentError()
	assert.Equal(t, "container status allocated resources not present", e.Error())
}

func TestNewContainerStatusResourcesNotPresentError(t *testing.T) {
	assert.True(t, errors.Is(NewContainerStatusResourcesNotPresentError(), ContainerStatusResourcesNotPresentError{}))
}

func TestContainerStatusResourcesNotPresentErrorError(t *testing.T) {
	e := NewContainerStatusResourcesNotPresentError()
	assert.Equal(t, "container status resources not present", e.Error())
}
