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

package scale

import (
	"k8s.io/api/core/v1"
)

type Updates interface {
	StartupPodMutationFuncAll(*v1.Container) []func(pod *v1.Pod) error
	PostStartupPodMutationFuncAll(*v1.Container) []func(pod *v1.Pod) error

	UpdateFor(v1.ResourceName) Update
	AllUpdates() []Update
}

func NewUpdates(configs Configs) Updates {
	return &updates{
		cpuUpdate:    NewUpdate(v1.ResourceCPU, configs.ConfigFor(v1.ResourceCPU)),
		memoryUpdate: NewUpdate(v1.ResourceMemory, configs.ConfigFor(v1.ResourceMemory)),
	}
}

type updates struct {
	cpuUpdate    Update
	memoryUpdate Update
}

func (u *updates) StartupPodMutationFuncAll(container *v1.Container) []func(pod *v1.Pod) error {
	var funcs []func(pod *v1.Pod) error

	for _, update := range u.AllUpdates() {
		funcs = append(funcs, update.StartupPodMutationFunc(container))
	}

	return funcs
}

func (u *updates) PostStartupPodMutationFuncAll(container *v1.Container) []func(pod *v1.Pod) error {
	var funcs []func(pod *v1.Pod) error

	for _, update := range u.AllUpdates() {
		funcs = append(funcs, update.PostStartupPodMutationFunc(container))
	}

	return funcs
}

func (u *updates) UpdateFor(resourceName v1.ResourceName) Update {
	switch resourceName {
	case v1.ResourceCPU:
		return u.cpuUpdate
	case v1.ResourceMemory:
		return u.memoryUpdate
	default:
		return nil
	}
}

func (u *updates) AllUpdates() []Update {
	return []Update{u.cpuUpdate, u.memoryUpdate}
}
