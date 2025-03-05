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

package retry

import (
	"context"
	"time"

	ccontext "github.com/ExpediaGroup/container-startup-autoscaler/internal/context"
	"github.com/avast/retry-go/v4"
)

var baseConfig = []retry.Option{
	retry.DelayType(retry.FixedDelay),
	retry.RetryIf(retry.IsRecoverable),
	retry.LastErrorOnly(true),
}

// StandardRetryConfig returns the configuration necessary to perform a standard retry.
func StandardRetryConfig(ctx context.Context) []retry.Option {
	newConfig := append(baseConfig, retry.Attempts(uint(ccontext.StandardRetryAttempts(ctx))))
	newConfig = append(newConfig, retry.Delay(time.Duration(ccontext.StandardRetryDelaySecs(ctx))*time.Second))
	return append(newConfig, retry.Context(ctx))
}

// DoStandardRetry performs a standard retry for the supplied function.
func DoStandardRetry(ctx context.Context, function retry.RetryableFunc) error {
	return retry.Do(function, StandardRetryConfig(ctx)...)
}

// DoStandardRetryWithMoreOpts performs a standard retry for the supplied function, with additional options.
func DoStandardRetryWithMoreOpts(ctx context.Context, function retry.RetryableFunc, moreOpts []retry.Option) error {
	opts := append(StandardRetryConfig(ctx), moreOpts...)
	return retry.Do(function, opts...)
}
