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
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"k8s.io/api/core/v1"
)

// states is the default implementation of scalecommon.States.
type states struct {
	cpuState    scalecommon.State
	memoryState scalecommon.State
}

func NewStates(configs scalecommon.Configurations, containerHelper kubecommon.ContainerHelper) scalecommon.States {
	return &states{
		cpuState:    NewState(v1.ResourceCPU, configs.ConfigurationFor(v1.ResourceCPU), containerHelper),
		memoryState: NewState(v1.ResourceMemory, configs.ConfigurationFor(v1.ResourceMemory), containerHelper),
	}
}

// IsStartupConfigurationAppliedAll invokes IsStartupConfigurationApplied on each state within this collection and
// returns whether they all returned true.
func (s *states) IsStartupConfigurationAppliedAll(container *v1.Container) bool {
	appliedAll := true

	for _, state := range s.AllStates() {
		applied := state.IsStartupConfigurationApplied(container)
		if applied != nil {
			appliedAll = appliedAll && *applied
		}
	}

	return appliedAll
}

// IsPostStartupConfigurationAppliedAll invokes IsPostStartupConfigurationApplied on each state within this collection
// and returns whether they all returned true.
func (s *states) IsPostStartupConfigurationAppliedAll(container *v1.Container) bool {
	appliedAll := true

	for _, state := range s.AllStates() {
		applied := state.IsPostStartupConfigurationApplied(container)
		if applied != nil {
			appliedAll = appliedAll && *applied
		}
	}

	return appliedAll
}

// IsAnyCurrentZeroAll invokes IsAnyCurrentZero on each state within this collection and returns whether any returned
// true.
func (s *states) IsAnyCurrentZeroAll(pod *v1.Pod, container *v1.Container) (bool, error) {
	zeroAny := false

	for _, state := range s.AllStates() {
		zero, err := state.IsAnyCurrentZero(pod, container)
		if err != nil {
			return false, common.WrapErrorf(err, "unable to determine if any current %s is zero", state.ResourceName())
		}
		if zero != nil && *zero == true {
			zeroAny = true
			break
		}
	}

	return zeroAny, nil
}

// DoesRequestsCurrentMatchSpecAll invokes DoesRequestsCurrentMatchSpec on each state within this collection and
// returns whether they all returned true.
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

// DoesLimitsCurrentMatchSpecAll invokes DoesLimitsCurrentMatchSpec on each state within this collection and returns
// whether they all returned true.
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

// StateFor returns the state for the supplied resource name.
func (s *states) StateFor(resourceName v1.ResourceName) scalecommon.State {
	switch resourceName {
	case v1.ResourceCPU:
		return s.cpuState
	case v1.ResourceMemory:
		return s.memoryState
	default:
		return nil
	}
}

// AllStates returns all states within this collection.
func (s *states) AllStates() []scalecommon.State {
	return []scalecommon.State{s.cpuState, s.memoryState}
}
