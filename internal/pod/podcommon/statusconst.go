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

package podcommon

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/metricscommon"
	"github.com/pkg/errors"
)

// StatusScaleState indicates the scale state for status purposes.
type StatusScaleState string

const (
	// StatusScaleStateNotApplicable indicates the scale state is not applicable.
	StatusScaleStateNotApplicable StatusScaleState = "notapplicable"

	// StatusScaleStateUpCommanded indicates scaling up commanded.
	StatusScaleStateUpCommanded StatusScaleState = "upcommanded"

	// StatusScaleStateUpEnacted indicates scaling up enacted.
	StatusScaleStateUpEnacted StatusScaleState = "upenacted"

	// StatusScaleStateUpFailed indicates scaling up failed.
	StatusScaleStateUpFailed StatusScaleState = "upfailed"

	// StatusScaleStateDownCommanded indicates scaling down commanded.
	StatusScaleStateDownCommanded StatusScaleState = "downcommanded"

	// StatusScaleStateDownEnacted indicates scaling down enacted.
	StatusScaleStateDownEnacted StatusScaleState = "downenacted"

	// StatusScaleStateDownFailed indicates scaling down failed.
	StatusScaleStateDownFailed StatusScaleState = "downfailed"

	// StatusScaleStateUnknownCommanded indicates scaling in an unknown direction commanded.
	StatusScaleStateUnknownCommanded StatusScaleState = "unknowncommanded"
)

// Direction returns the scale direction.
func (s StatusScaleState) Direction() metricscommon.Direction {
	switch s {
	case StatusScaleStateUpCommanded, StatusScaleStateUpEnacted, StatusScaleStateUpFailed:
		return metricscommon.DirectionUp
	case StatusScaleStateDownCommanded, StatusScaleStateDownEnacted, StatusScaleStateDownFailed:
		return metricscommon.DirectionDown
	}

	panic(errors.Errorf("'%s' not supported", s))
}
