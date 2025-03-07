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

package common

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapErrorf(t *testing.T) {
	t.Run("WithFormatString", func(t *testing.T) {
		err := WrapErrorf(errors.New("non-wrapped err message"), "format (%s)", "test")
		assert.Equal(t, "format (test): non-wrapped err message", err.Error())
	})

	t.Run("WithoutFormatString", func(t *testing.T) {
		err := WrapErrorf(errors.New("non-wrapped err message"), "format")
		assert.Equal(t, "format: non-wrapped err message", err.Error())
	})

	t.Run("ChainWithFormatString", func(t *testing.T) {
		errInner := WrapErrorf(errors.New("non-wrapped err message"), "errInner format (%s)", "errInner test")
		errOuter := WrapErrorf(errInner, "errOuter format (%s)", "errOuter test")
		assert.Equal(t, "errOuter format (errOuter test): errInner format (errInner test): non-wrapped err message", errOuter.Error())
	})

	t.Run("ChainWithoutFormatString", func(t *testing.T) {
		errInner := WrapErrorf(errors.New("non-wrapped err message"), "errInner format")
		errOuter := WrapErrorf(errInner, "errOuter format")
		assert.Equal(t, "errOuter format: errInner format: non-wrapped err message", errOuter.Error())
	})
}
