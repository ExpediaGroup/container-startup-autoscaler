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
	"errors"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("test", errors.New(""))
	assert.True(t, errors.As(err, &ValidationError{}))
}

func TestValidationErrorError(t *testing.T) {
	t.Run("Wrapped", func(t *testing.T) {
		err1 := errors.New("err1")
		err2 := common.WrapErrorf(err1, "err2")
		e := NewValidationError("err3", err2)
		assert.Equal(t, "validation error: err3: err2: err1", e.Error())
	})

	t.Run("NotWrapped", func(t *testing.T) {
		e := NewValidationError("test", nil)
		assert.Equal(t, "validation error: test", e.Error())
	})
}
