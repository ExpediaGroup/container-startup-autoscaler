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

package kube

import (
	"context"
	"errors"
	"strings"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	metricsretry "github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/retry"
	"github.com/avast/retry-go/v4"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// kubeApiRetryOptions returns the configuration necessary to perform a retry for a Kube API invocation.
func kubeApiRetryOptions(ctx context.Context) []retry.Option {
	var opts []retry.Option

	// Don't retry if it's a 'not found' error or not recoverable.
	opts = append(opts, retry.RetryIf(func(err error) bool {
		return !kerrors.IsNotFound(err) && retry.IsRecoverable(err)
	}))

	// Log retry.
	opts = append(opts, retry.OnRetry(func(n uint, err error) {
		reason := "unknown"

		var stat kerrors.APIStatus
		var ok bool
		if stat, ok = err.(kerrors.APIStatus); !ok {
			stat, _ = errors.Unwrap(err).(kerrors.APIStatus)
		}

		if stat != nil && stat.Status().Reason != v1.StatusReasonUnknown {
			reason = strings.ToLower(string(stat.Status().Reason))
		}

		logging.Errorf(ctx, err, "(attempt %d) kube api call failed (reason: %s)", n+1, reason)
		metricsretry.Retry(reason).Inc()
	}))

	return opts
}
