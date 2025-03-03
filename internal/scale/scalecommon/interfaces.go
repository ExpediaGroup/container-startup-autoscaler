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

package scalecommon

import (
	v1 "k8s.io/api/core/v1"
)

type Config interface {
	ResourceName() v1.ResourceName
	IsEnabled() bool
	Resources() Resources
	StoreFromAnnotations(*v1.Pod) error
	Validate(*v1.Container) error
	String() string
}

type Configs interface {
	TargetContainerName(*v1.Pod) (string, error)
	StoreFromAnnotationsAll(*v1.Pod) error
	ValidateAll(*v1.Container) error
	ValidateCollection() error

	ConfigFor(v1.ResourceName) Config
	AllConfigs() []Config
	AllEnabledConfigs() []Config
	AllEnabledConfigsResourceNames() []v1.ResourceName

	String() string
}

type State interface {
	ResourceName() v1.ResourceName
	IsStartupConfigApplied(*v1.Container) *bool
	IsPostStartupConfigApplied(*v1.Container) *bool
	IsAnyCurrentZero(*v1.Pod, *v1.Container) (*bool, error)
	DoesRequestsCurrentMatchSpec(*v1.Pod, *v1.Container) (*bool, error)
	DoesLimitsCurrentMatchSpec(*v1.Pod, *v1.Container) (*bool, error)
}

type States interface {
	IsStartupConfigAppliedAll(*v1.Container) bool
	IsPostStartupConfigAppliedAll(*v1.Container) bool
	IsAnyCurrentZeroAll(*v1.Pod, *v1.Container) (bool, error)
	DoesRequestsCurrentMatchSpecAll(*v1.Pod, *v1.Container) (bool, error)
	DoesLimitsCurrentMatchSpecAll(*v1.Pod, *v1.Container) (bool, error)

	StateFor(v1.ResourceName) State
	AllStates() []State
}

type Update interface {
	ResourceName() v1.ResourceName
	StartupPodMutationFunc(*v1.Container) func(pod *v1.Pod) error
	PostStartupPodMutationFunc(*v1.Container) func(pod *v1.Pod) error
}

type Updates interface {
	StartupPodMutationFuncAll(*v1.Container) []func(pod *v1.Pod) error
	PostStartupPodMutationFuncAll(*v1.Container) []func(pod *v1.Pod) error

	UpdateFor(v1.ResourceName) Update
	AllUpdates() []Update
}
