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

type memoryScaleState struct {
	scaleConfig     config.ScaleConfig
	containerHelper kube.ContainerHelper
}

func NewMemoryScaleState(scaleConfig config.ScaleConfig, containerHelper kube.ContainerHelper) *memoryScaleState {
	return &memoryScaleState{
		scaleConfig:     scaleConfig,
		containerHelper: containerHelper,
	}
}

func (m *memoryScaleState) ResourceType() scaleresource.ResourceType {
	return scaleresource.ResourceTypeMemory
}

func (m *memoryScaleState) IsStartupConfigApplied(container *v1.Container) bool {
	if !m.scaleConfig.IsEnabled() {
		return true
	}

	requestsStartupApplied := m.containerHelper.Requests(container, v1.ResourceMemory).Equal(m.scaleConfig.Resources().Startup)
	limitsStartupApplied := m.containerHelper.Limits(container, v1.ResourceMemory).Equal(m.scaleConfig.Resources().Startup)
	return requestsStartupApplied && limitsStartupApplied
}

func (m *memoryScaleState) IsPostStartupConfigApplied(container *v1.Container) bool {
	if !m.scaleConfig.IsEnabled() {
		return true
	}

	postStartupRequestsApplied := m.containerHelper.Requests(container, v1.ResourceMemory).Equal(m.scaleConfig.Resources().PostStartupRequests)
	postStartupLimitsApplied := m.containerHelper.Limits(container, v1.ResourceMemory).Equal(m.scaleConfig.Resources().PostStartupLimits)
	return postStartupRequestsApplied && postStartupLimitsApplied
}

func (m *memoryScaleState) DoesRequestsCurrentMatchSpec(pod *v1.Pod, container *v1.Container) (bool, error) {
	if !m.scaleConfig.IsEnabled() {
		return true, nil
	}

	currentRequests, err := m.containerHelper.CurrentRequests(pod, container, v1.ResourceMemory)
	if err != nil {
		return false, common.WrapErrorf(err, "unable to get status resources memory requests")
	}

	requests := m.containerHelper.Requests(container, v1.ResourceMemory)
	return currentRequests.Equal(requests), nil
}

func (m *memoryScaleState) DoesLimitsCurrentMatchSpec(pod *v1.Pod, container *v1.Container) (bool, error) {
	if !m.scaleConfig.IsEnabled() {
		return true, nil
	}

	currentLimits, err := m.containerHelper.CurrentLimits(pod, container, v1.ResourceMemory)
	if err != nil {
		return false, common.WrapErrorf(err, "unable to get status resources memory limits")
	}

	limits := m.containerHelper.Limits(container, v1.ResourceMemory)
	return currentLimits.Equal(limits), nil
}
