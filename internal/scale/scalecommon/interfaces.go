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

package scalecommon

import (
	v1 "k8s.io/api/core/v1"
)

// Configuration performs configuration-related operations for a specific resource.
type Configuration interface {
	ResourceName() v1.ResourceName

	IsEnabled() bool

	Resources() Resources

	StoreFromAnnotations(
		pod *v1.Pod,
	) error

	Validate(
		container *v1.Container,
	) error

	String() string
}

// Configurations performs operations upon a Configuration collection.
type Configurations interface {
	TargetContainerName(
		pod *v1.Pod,
	) (string, error)

	StoreFromAnnotationsAll(
		pod *v1.Pod,
	) error

	ValidateAll(
		container *v1.Container,
	) error

	ValidateCollection() error

	ConfigurationFor(
		resourceName v1.ResourceName,
	) Configuration

	AllConfigurations() []Configuration

	AllEnabledConfigurations() []Configuration

	AllEnabledConfigurationsResourceNames() []v1.ResourceName

	String() string
}

// State performs state-related operations for a specific resource.
type State interface {
	ResourceName() v1.ResourceName

	IsStartupConfigurationApplied(
		container *v1.Container,
	) *bool

	IsPostStartupConfigurationApplied(
		container *v1.Container,
	) *bool

	IsAnyCurrentZero(
		pod *v1.Pod,
		container *v1.Container,
	) (*bool, error)

	DoesRequestsCurrentMatchSpec(
		pod *v1.Pod,
		container *v1.Container,
	) (*bool, error)

	DoesLimitsCurrentMatchSpec(
		pod *v1.Pod,
		container *v1.Container,
	) (*bool, error)
}

// States performs operations upon a State collection.
type States interface {
	IsStartupConfigurationAppliedAll(
		container *v1.Container,
	) bool

	IsPostStartupConfigurationAppliedAll(
		container *v1.Container,
	) bool

	IsAnyCurrentZeroAll(
		pod *v1.Pod,
		container *v1.Container,
	) (bool, error)

	DoesRequestsCurrentMatchSpecAll(
		pod *v1.Pod,
		container *v1.Container,
	) (bool, error)

	DoesLimitsCurrentMatchSpecAll(
		pod *v1.Pod,
		container *v1.Container,
	) (bool, error)

	StateFor(
		resourceName v1.ResourceName,
	) State

	AllStates() []State
}

// Update performs update-related operations for a specific resource.
type Update interface {
	ResourceName() v1.ResourceName

	StartupPodMutationFunc(
		container *v1.Container,
	) func(podToMutate *v1.Pod) (bool, func(currentPod *v1.Pod) bool, error)

	PostStartupPodMutationFunc(
		container *v1.Container,
	) func(podToMutate *v1.Pod) (bool, func(currentPod *v1.Pod) bool, error)
}

// Updates performs operations upon an Update collection.
type Updates interface {
	StartupPodMutationFuncAll(
		container *v1.Container,
	) []func(podToMutate *v1.Pod) (bool, func(currentPod *v1.Pod) bool, error)

	PostStartupPodMutationFuncAll(
		container *v1.Container,
	) []func(podToMutate *v1.Pod) (bool, func(currentPod *v1.Pod) bool, error)

	UpdateFor(
		resourceName v1.ResourceName,
	) Update

	AllUpdates() []Update
}
