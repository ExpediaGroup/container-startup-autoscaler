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

// States holds information related to the current state of the target container.
type States struct {
	StartupProbe       StateBool               `json:"startupProbe"`
	ReadinessProbe     StateBool               `json:"readinessProbe"`
	Container          StateContainer          `json:"container"`
	Started            StateBool               `json:"started"`
	Ready              StateBool               `json:"ready"`
	Resources          StateResources          `json:"resources"`
	AllocatedResources StateAllocatedResources `json:"allocatedResources"`
	StatusResources    StateStatusResources    `json:"statusResources"`
}

func NewStates(
	startupProbe StateBool,
	readinessProbe StateBool,
	stateContainer StateContainer,
	started StateBool,
	ready StateBool,
	stateResources StateResources,
	stateAllocatedResources StateAllocatedResources,
	stateStatusResources StateStatusResources,
) States {
	return States{
		StartupProbe:       startupProbe,
		ReadinessProbe:     readinessProbe,
		Container:          stateContainer,
		Started:            started,
		Ready:              ready,
		Resources:          stateResources,
		AllocatedResources: stateAllocatedResources,
		StatusResources:    stateStatusResources,
	}
}

func NewStatesAllUnknown() States {
	return States{
		StartupProbe:       StateBoolUnknown,
		ReadinessProbe:     StateBoolUnknown,
		Container:          StateContainerUnknown,
		Started:            StateBoolUnknown,
		Ready:              StateBoolUnknown,
		Resources:          StateResourcesUnknown,
		AllocatedResources: StateAllocatedResourcesUnknown,
		StatusResources:    StateStatusResourcesUnknown,
	}
}
