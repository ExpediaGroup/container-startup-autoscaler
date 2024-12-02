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
	"fmt"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/metricscommon"
)

// StateBool indicates the state of a boolean-like state.
type StateBool string

const (
	StateBoolTrue    StateBool = "true"
	StateBoolFalse   StateBool = "false"
	StateBoolUnknown StateBool = "unknown"
)

// Bool returns a bool representation.
func (s StateBool) Bool() bool {
	return s == StateBoolTrue
}

// StateContainer indicates the state of a Kube container.
type StateContainer string

const (
	StateContainerRunning    StateContainer = "running"
	StateContainerWaiting    StateContainer = "waiting"
	StateContainerTerminated StateContainer = "terminated"
	StateContainerUnknown    StateContainer = "unknown"
)

// StateResources indicates what Kube container resources are set.
type StateResources string

const (
	StateResourcesStartup     StateResources = "startup"
	StateResourcesPostStartup StateResources = "poststartup"
	StateResourcesUnknown     StateResources = "unknown"
)

// Direction returns the scale direction.
func (s StateResources) Direction() metricscommon.Direction {
	switch s {
	case StateResourcesStartup:
		return metricscommon.DirectionUp
	case StateResourcesPostStartup:
		return metricscommon.DirectionDown
	}

	panic(fmt.Errorf("'%s' not supported", s))
}

// HumanReadable returns a string suitable to include within human-readable messages.
func (s StateResources) HumanReadable() string {
	switch s {
	case StateResourcesPostStartup:
		return "post-startup"
	default:
		return string(s)
	}
}

// StateStatusResources indicates the state of a Kube container's status resources.
type StateStatusResources string

const (
	// StateStatusResourcesIncomplete indicates status resources are incomplete.
	StateStatusResourcesIncomplete StateStatusResources = "incomplete"

	// StateStatusResourcesContainerResourcesMatch indicates status resources match container resources.
	StateStatusResourcesContainerResourcesMatch StateStatusResources = "containerresourcesmatch"

	// StateStatusResourcesContainerResourcesMismatch indicates status resources don't match container resources.
	StateStatusResourcesContainerResourcesMismatch StateStatusResources = "containerresourcesmismatch"

	// StateStatusResourcesUnknown indicates status resources are unknown.
	StateStatusResourcesUnknown StateStatusResources = "unknown"
)
