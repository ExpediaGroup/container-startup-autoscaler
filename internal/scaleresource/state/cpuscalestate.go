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

type cpuScaleState struct {
	scaleConfig     config.ScaleConfig
	containerHelper kube.ContainerHelper
}

func NewCpuScaleState(scaleConfig config.ScaleConfig, containerHelper kube.ContainerHelper) *cpuScaleState {
	return &cpuScaleState{
		scaleConfig:     scaleConfig,
		containerHelper: containerHelper,
	}
}

func (c *cpuScaleState) ResourceType() scaleresource.ResourceType {
	return scaleresource.ResourceTypeCpu
}

func (c *cpuScaleState) IsStartupConfigApplied(container *v1.Container) bool {
	if !c.scaleConfig.IsEnabled() {
		return true
	}

	startupRequestsApplied := c.containerHelper.Requests(container, v1.ResourceCPU).Equal(c.scaleConfig.Resources().Startup)
	startupLimitsApplied := c.containerHelper.Limits(container, v1.ResourceCPU).Equal(c.scaleConfig.Resources().Startup)
	return startupRequestsApplied && startupLimitsApplied
}

func (c *cpuScaleState) IsPostStartupConfigApplied(container *v1.Container) bool {
	if !c.scaleConfig.IsEnabled() {
		return true
	}

	postStartupRequestsApplied := c.containerHelper.Requests(container, v1.ResourceCPU).Equal(c.scaleConfig.Resources().PostStartupRequests)
	postStartupLimitsApplied := c.containerHelper.Limits(container, v1.ResourceCPU).Equal(c.scaleConfig.Resources().PostStartupLimits)
	return postStartupRequestsApplied && postStartupLimitsApplied
}

func (c *cpuScaleState) DoesRequestsCurrentMatchSpec(pod *v1.Pod, container *v1.Container) (bool, error) {
	if !c.scaleConfig.IsEnabled() {
		return true, nil
	}

	currentRequests, err := c.containerHelper.CurrentRequests(pod, container, v1.ResourceCPU)
	if err != nil {
		return false, common.WrapErrorf(err, "unable to get status resources cpu requests")
	}

	requests := c.containerHelper.Requests(container, v1.ResourceCPU)
	return currentRequests.Equal(requests), nil
}

func (c *cpuScaleState) DoesLimitsCurrentMatchSpec(pod *v1.Pod, container *v1.Container) (bool, error) {
	if !c.scaleConfig.IsEnabled() {
		return true, nil
	}

	currentLimits, err := c.containerHelper.CurrentLimits(pod, container, v1.ResourceCPU)
	if err != nil {
		return false, common.WrapErrorf(err, "unable to get status resources cpu limits")
	}

	limits := c.containerHelper.Limits(container, v1.ResourceCPU)
	return currentLimits.Equal(limits), nil
}
