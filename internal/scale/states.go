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
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"k8s.io/api/core/v1"
)

type States interface {
	IsStartupConfigAppliedAll(*v1.Container) bool
	IsPostStartupConfigAppliedAll(*v1.Container) bool
	IsAnyCurrentZeroAll(*v1.Pod, *v1.Container) (bool, error)
	DoesRequestsCurrentMatchSpecAll(*v1.Pod, *v1.Container) (bool, error)
	DoesLimitsCurrentMatchSpecAll(*v1.Pod, *v1.Container) (bool, error)

	StateFor(v1.ResourceName) State
	AllStates() []State
}

func NewStates(configs Configs, containerHelper kube.ContainerHelper) States {
	return &states{
		cpuState:    NewState(v1.ResourceCPU, configs.ConfigFor(v1.ResourceCPU), containerHelper),
		memoryState: NewState(v1.ResourceMemory, configs.ConfigFor(v1.ResourceMemory), containerHelper),
	}
}

type states struct {
	cpuState    State
	memoryState State
}

func (s *states) IsStartupConfigAppliedAll(container *v1.Container) bool {
	appliedAll := true

	for _, state := range s.AllStates() {
		applied := state.IsStartupConfigApplied(container)
		if applied != nil {
			appliedAll = appliedAll && *applied
		}
	}

	return appliedAll
}

func (s *states) IsPostStartupConfigAppliedAll(container *v1.Container) bool {
	appliedAll := true

	for _, state := range s.AllStates() {
		applied := state.IsPostStartupConfigApplied(container)
		if applied != nil {
			appliedAll = appliedAll && *applied
		}
	}

	return appliedAll
}

func (s *states) IsAnyCurrentZeroAll(pod *v1.Pod, container *v1.Container) (bool, error) {
	zeroAny := false

	for _, state := range s.AllStates() {
		zero, err := state.IsAnyCurrentZero(pod, container)
		if err != nil {
			return false, common.WrapErrorf(err, "unable to determine if any current %s is zero", state.ResourceName())
		}

		if zero != nil {
			zeroAny = true
			break
		}
	}

	return zeroAny, nil
}

func (s *states) DoesRequestsCurrentMatchSpecAll(pod *v1.Pod, container *v1.Container) (bool, error) {
	matchAll := true

	for _, state := range s.AllStates() {
		match, err := state.DoesRequestsCurrentMatchSpec(pod, container)
		if err != nil {
			return false, common.WrapErrorf(err, "unable to determine if current %s requests matches spec", state.ResourceName())
		}

		if match != nil {
			matchAll = matchAll && *match
		}
	}

	return matchAll, nil
}

func (s *states) DoesLimitsCurrentMatchSpecAll(pod *v1.Pod, container *v1.Container) (bool, error) {
	matchAll := true

	for _, state := range s.AllStates() {
		match, err := state.DoesLimitsCurrentMatchSpec(pod, container)
		if err != nil {
			return false, common.WrapErrorf(err, "unable to determine if current %s limits matches spec", state.ResourceName())
		}

		if match != nil {
			matchAll = matchAll && *match
		}
	}

	return matchAll, nil
}

func (s *states) StateFor(resourceName v1.ResourceName) State {
	switch resourceName {
	case v1.ResourceCPU:
		return s.cpuState
	case v1.ResourceMemory:
		return s.memoryState
	default:
		return nil
	}
}

func (s *states) AllStates() []State {
	return []State{s.cpuState, s.memoryState}
}
