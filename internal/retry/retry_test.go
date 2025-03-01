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

package retry

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contexttest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/avast/retry-go/v4"
	"github.com/stretchr/testify/assert"
)

func TestStandardRetryConfig(t *testing.T) {
	got := StandardRetryConfig(contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build())
	assert.Len(t, got, 6)
}

func TestDoStandardRetry(t *testing.T) {
	ctx := contexttest.NewCtxBuilder(contexttest.NewCustomCtxConfig()).
		LogBuffer(nil).
		StandardRetryAttempts(3).
		StandardRetryDelaySecs(1).
		Build()

	start := time.Now()
	err := DoStandardRetry(ctx, func() error { return errors.New("") })
	assert.NotNil(t, err)
	assert.GreaterOrEqual(t, time.Since(start).Milliseconds(), int64(2000))
}

func TestDoStandardRetryWithMoreOpts(t *testing.T) {
	buffer := &bytes.Buffer{}
	ctx := contexttest.NewCtxBuilder(contexttest.NewCustomCtxConfig()).
		LogBuffer(buffer).
		StandardRetryAttempts(3).
		StandardRetryDelaySecs(1).
		Build()
	opt := retry.OnRetry(func(n uint, err error) {
		logging.Errorf(ctx, err, "test")
	})
	start := time.Now()
	err := DoStandardRetryWithMoreOpts(ctx, func() error { return errors.New("") }, []retry.Option{opt})
	assert.NotNil(t, err)
	assert.GreaterOrEqual(t, time.Since(start).Milliseconds(), int64(2000))
	assert.Contains(t, buffer.String(), "test")
}
