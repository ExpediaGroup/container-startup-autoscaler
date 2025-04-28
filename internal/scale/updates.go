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

package scale

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"k8s.io/api/core/v1"
)

// updates is the default implementation of scalecommon.Updates.
type updates struct {
	cpuUpdate    scalecommon.Update
	memoryUpdate scalecommon.Update
}

func NewUpdates(configs scalecommon.Configurations) scalecommon.Updates {
	return &updates{
		cpuUpdate:    NewUpdate(v1.ResourceCPU, configs.ConfigurationFor(v1.ResourceCPU)),
		memoryUpdate: NewUpdate(v1.ResourceMemory, configs.ConfigurationFor(v1.ResourceMemory)),
	}
}

// StartupPodMutationFuncAll invokes StartupPodMutationFunc on each update within this collection and returns them.
func (u *updates) StartupPodMutationFuncAll(container *v1.Container) []func(*v1.Pod) (bool, func(*v1.Pod) bool, error) {
	var funcs []func(*v1.Pod) (bool, func(*v1.Pod) bool, error)

	for _, update := range u.AllUpdates() {
		funcs = append(funcs, update.StartupPodMutationFunc(container))
	}

	return funcs
}

// PostStartupPodMutationFuncAll invokes PostStartupPodMutationFunc on each update within this collection and returns
// them.
func (u *updates) PostStartupPodMutationFuncAll(container *v1.Container) []func(*v1.Pod) (bool, func(*v1.Pod) bool, error) {
	var funcs []func(*v1.Pod) (bool, func(*v1.Pod) bool, error)

	for _, update := range u.AllUpdates() {
		funcs = append(funcs, update.PostStartupPodMutationFunc(container))
	}

	return funcs
}

// UpdateFor returns the update for the supplied resource name.
func (u *updates) UpdateFor(resourceName v1.ResourceName) scalecommon.Update {
	switch resourceName {
	case v1.ResourceCPU:
		return u.cpuUpdate
	case v1.ResourceMemory:
		return u.memoryUpdate
	default:
		return nil
	}
}

// AllUpdates returns all updates within this collection.
func (u *updates) AllUpdates() []scalecommon.Update {
	return []scalecommon.Update{u.cpuUpdate, u.memoryUpdate}
}
