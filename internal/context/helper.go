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

package context

import (
	"context"
	"errors"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contextcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
)

// WithStandardRetryAttempts adds or replaces contextcommon.KeyStandardRetryAttempts to/in ctx.
func WithStandardRetryAttempts(ctx context.Context, attempts int) context.Context {
	return context.WithValue(ctx, contextcommon.KeyStandardRetryAttempts, attempts)
}

// StandardRetryAttempts retrieves contextcommon.KeyStandardRetryAttempts from ctx.
func StandardRetryAttempts(ctx context.Context) int {
	value := ctx.Value(contextcommon.KeyStandardRetryAttempts)
	if value == nil {
		panic(errors.New("standard retry attempts should have been previously set"))
	}

	return ctx.Value(contextcommon.KeyStandardRetryAttempts).(int)
}

// WithStandardRetryDelaySecs adds or replaces contextcommon.KeyStandardRetryDelaySecs to/in ctx.
func WithStandardRetryDelaySecs(ctx context.Context, secs int) context.Context {
	return context.WithValue(ctx, contextcommon.KeyStandardRetryDelaySecs, secs)
}

// StandardRetryDelaySecs retrieves contextcommon.KeyStandardRetryDelaySecs from ctx.
func StandardRetryDelaySecs(ctx context.Context) int {
	value := ctx.Value(contextcommon.KeyStandardRetryDelaySecs)
	if value == nil {
		panic(errors.New("standard retry delay secs should have been previously set"))
	}

	return ctx.Value(contextcommon.KeyStandardRetryDelaySecs).(int)
}

// WithTargetContainerName adds or replaces contextcommon.KeyTargetContainerName to/in ctx.
func WithTargetContainerName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, contextcommon.KeyTargetContainerName, name)
}

// TargetContainerName retrieves contextcommon.KeyTargetContainerName from ctx.
func TargetContainerName(ctx context.Context) string {
	value := ctx.Value(contextcommon.KeyTargetContainerName)
	if value == nil {
		return ""
	}

	return ctx.Value(contextcommon.KeyTargetContainerName).(string)
}

// WithTargetContainerStates adds or replaces contextcommon.KeyTargetContainerStates to/in ctx.
func WithTargetContainerStates(ctx context.Context, states podcommon.States) context.Context {
	return context.WithValue(ctx, contextcommon.KeyTargetContainerStates, states)
}

// TargetContainerStates retrieves contextcommon.KeyTargetContainerStates from ctx.
func TargetContainerStates(ctx context.Context) podcommon.States {
	value := ctx.Value(contextcommon.KeyTargetContainerStates)
	if value == nil {
		return podcommon.States{}
	}

	return ctx.Value(contextcommon.KeyTargetContainerStates).(podcommon.States)
}
