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

package context

import (
	"context"
	"testing"
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/stretchr/testify/assert"
)

func TestWithStandardRetryAttempts(t *testing.T) {
	got := WithStandardRetryAttempts(context.TODO(), 1000)
	assert.Equal(t, 1000, got.Value(KeyStandardRetryAttempts).(int))
}

func TestStandardRetryAttempts(t *testing.T) {
	t.Run("NotSetPanic", func(t *testing.T) {
		assert.PanicsWithError(
			t,
			"standard retry attempts should have been previously set",
			func() { StandardRetryAttempts(context.TODO()) },
		)
	})

	t.Run("Ok", func(t *testing.T) {
		ctx := context.TODO()
		ctx = context.WithValue(ctx, KeyStandardRetryAttempts, 1000)
		assert.Equal(t, 1000, StandardRetryAttempts(ctx))
	})
}

func TestWithStandardRetryDelaySecs(t *testing.T) {
	got := WithStandardRetryDelaySecs(context.TODO(), 2000)
	assert.Equal(t, 2000, got.Value(KeyStandardRetryDelaySecs).(int))
}

func TestStandardRetryDelaySecs(t *testing.T) {
	t.Run("NotSetPanic", func(t *testing.T) {
		assert.PanicsWithError(
			t,
			"standard retry delay secs should have been previously set",
			func() { StandardRetryDelaySecs(context.TODO()) },
		)
	})

	t.Run("Ok", func(t *testing.T) {
		ctx := context.WithValue(context.TODO(), KeyStandardRetryDelaySecs, 2000)
		assert.Equal(t, 2000, StandardRetryDelaySecs(ctx))
	})
}

func TestWithTargetContainerName(t *testing.T) {
	got := WithTargetContainerName(context.TODO(), "test")
	assert.Equal(t, "test", got.Value(KeyTargetContainerName).(string))
}

func TestTargetContainerName(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "", TargetContainerName(context.TODO()))
	})

	t.Run("NotNil", func(t *testing.T) {
		ctx := context.WithValue(context.TODO(), KeyTargetContainerName, "test")
		assert.Equal(t, "test", TargetContainerName(ctx))
	})
}

func TestWithTargetContainerStates(t *testing.T) {
	states := podcommon.NewStatesAllUnknown()
	got := WithTargetContainerStates(context.TODO(), states)
	assert.Equal(t, states, got.Value(KeyTargetContainerStates).(podcommon.States))
}

func TestTargetContainerStates(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, podcommon.States{}, TargetContainerStates(context.TODO()))
	})

	t.Run("NotNil", func(t *testing.T) {
		states := podcommon.NewStatesAllUnknown()
		ctx := context.WithValue(context.TODO(), KeyTargetContainerStates, states)
		assert.Equal(t, states, TargetContainerStates(ctx))
	})
}

func TestWithTimeoutOverride(t *testing.T) {
	duration := 1 * time.Second
	got := WithTimeoutOverride(context.TODO(), duration)
	assert.Equal(t, duration, got.Value(KeyTimeoutOverride).(time.Duration))
}

func TestTimeoutOverride(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), TimeoutOverride(context.TODO()))
	})

	t.Run("NotNil", func(t *testing.T) {
		duration := 1 * time.Second
		ctx := context.WithValue(context.TODO(), KeyTimeoutOverride, duration)
		assert.Equal(t, duration, TimeoutOverride(ctx))
	})
}
