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
	"errors"
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
)

// WithStandardRetryAttempts adds or replaces KeyStandardRetryAttempts to/in ctx.
func WithStandardRetryAttempts(ctx context.Context, attempts int) context.Context {
	return context.WithValue(ctx, KeyStandardRetryAttempts, attempts)
}

// StandardRetryAttempts retrieves KeyStandardRetryAttempts from ctx.
func StandardRetryAttempts(ctx context.Context) int {
	value := ctx.Value(KeyStandardRetryAttempts)
	if value == nil {
		panic(errors.New("standard retry attempts should have been previously set"))
	}

	return ctx.Value(KeyStandardRetryAttempts).(int)
}

// WithStandardRetryDelaySecs adds or replaces KeyStandardRetryDelaySecs to/in ctx.
func WithStandardRetryDelaySecs(ctx context.Context, secs int) context.Context {
	return context.WithValue(ctx, KeyStandardRetryDelaySecs, secs)
}

// StandardRetryDelaySecs retrieves KeyStandardRetryDelaySecs from ctx.
func StandardRetryDelaySecs(ctx context.Context) int {
	value := ctx.Value(KeyStandardRetryDelaySecs)
	if value == nil {
		panic(errors.New("standard retry delay secs should have been previously set"))
	}

	return ctx.Value(KeyStandardRetryDelaySecs).(int)
}

// WithTargetContainerName adds or replaces KeyTargetContainerName to/in ctx.
func WithTargetContainerName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, KeyTargetContainerName, name)
}

// TargetContainerName retrieves KeyTargetContainerName from ctx.
func TargetContainerName(ctx context.Context) string {
	value := ctx.Value(KeyTargetContainerName)
	if value == nil {
		return ""
	}

	return ctx.Value(KeyTargetContainerName).(string)
}

// WithTargetContainerStates adds or replaces KeyTargetContainerStates to/in ctx.
func WithTargetContainerStates(ctx context.Context, states podcommon.States) context.Context {
	return context.WithValue(ctx, KeyTargetContainerStates, states)
}

// TargetContainerStates retrieves KeyTargetContainerStates from ctx.
func TargetContainerStates(ctx context.Context) podcommon.States {
	value := ctx.Value(KeyTargetContainerStates)
	if value == nil {
		return podcommon.States{}
	}

	return ctx.Value(KeyTargetContainerStates).(podcommon.States)
}

// WithTimeoutOverride adds or replaces KeyTimeoutOverride to/in ctx.
func WithTimeoutOverride(ctx context.Context, duration time.Duration) context.Context {
	return context.WithValue(ctx, KeyTimeoutOverride, duration)
}

// TimeoutOverride retrieves KeyTimeoutOverride from ctx.
func TimeoutOverride(ctx context.Context) time.Duration {
	value := ctx.Value(KeyTimeoutOverride)
	if value == nil {
		return time.Duration(0)
	}

	return ctx.Value(KeyTimeoutOverride).(time.Duration)
}
