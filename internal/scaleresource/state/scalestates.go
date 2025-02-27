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

package state

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource/config"
	"k8s.io/api/core/v1"
)

type ScaleStates interface {
	IsStartupConfigAppliedAll(*v1.Container) bool
	IsPostStartupConfigAppliedAll(*v1.Container) bool
	DoesRequestsCurrentMatchSpecAll(*v1.Pod, *v1.Container) (bool, error)
	DoesLimitsCurrentMatchSpecAll(*v1.Pod, *v1.Container) (bool, error)

	ScaleStateFor(scaleresource.ResourceType) ScaleState
	AllScaleStates() []ScaleState
}

func NewScaleStates(scaleConfigs config.ScaleConfigs, containerHelper kube.ContainerHelper) ScaleStates {
	return &scaleStates{
		cpuScaleState:    NewCpuScaleState(scaleConfigs.ScaleConfigFor(scaleresource.ResourceTypeCpu), containerHelper),
		memoryScaleState: NewMemoryScaleState(scaleConfigs.ScaleConfigFor(scaleresource.ResourceTypeMemory), containerHelper),
	}
}

type scaleStates struct {
	cpuScaleState    ScaleState
	memoryScaleState ScaleState
}

func (s *scaleStates) IsStartupConfigAppliedAll(container *v1.Container) bool {
	appliedAll := true

	for _, scaleState := range s.AllScaleStates() {
		appliedAll = appliedAll && scaleState.IsStartupConfigApplied(container)
	}

	return appliedAll
}

func (s *scaleStates) IsPostStartupConfigAppliedAll(container *v1.Container) bool {
	appliedAll := true

	for _, scaleState := range s.AllScaleStates() {
		appliedAll = appliedAll && scaleState.IsPostStartupConfigApplied(container)
	}

	return appliedAll
}

func (s *scaleStates) DoesRequestsCurrentMatchSpecAll(pod *v1.Pod, container *v1.Container) (bool, error) {
	matchAll := true

	for _, scaleState := range s.AllScaleStates() {
		match, err := scaleState.DoesRequestsCurrentMatchSpec(pod, container)
		if err != nil {
			return false, common.WrapErrorf(err, "unable to determine if current %s requests matches spec", scaleState.ResourceType())
		}

		matchAll = matchAll && match
	}

	return matchAll, nil
}

func (s *scaleStates) DoesLimitsCurrentMatchSpecAll(pod *v1.Pod, container *v1.Container) (bool, error) {
	matchAll := true

	for _, scaleState := range s.AllScaleStates() {
		match, err := scaleState.DoesLimitsCurrentMatchSpec(pod, container)
		if err != nil {
			return false, common.WrapErrorf(err, "unable to determine if current %s limits matches spec", scaleState.ResourceType())
		}

		matchAll = matchAll && match
	}

	return matchAll, nil
}

func (s *scaleStates) ScaleStateFor(resourceType scaleresource.ResourceType) ScaleState {
	switch resourceType {
	case scaleresource.ResourceTypeCpu:
		return s.cpuScaleState
	case scaleresource.ResourceTypeMemory:
		return s.memoryScaleState
	default:
		return nil
	}
}

func (s *scaleStates) AllScaleStates() []ScaleState {
	return []ScaleState{s.cpuScaleState, s.memoryScaleState}
}
