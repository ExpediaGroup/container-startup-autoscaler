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

package update

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource/config"
	v1 "k8s.io/api/core/v1"
)

type cpuScaleUpdate struct {
	anyScaleUpdate

	scaleConfig config.ScaleConfig
}

func NewCpuScaleUpdate(scaleConfig config.ScaleConfig) *cpuScaleUpdate {
	return &cpuScaleUpdate{
		scaleConfig: scaleConfig,
	}
}

func (c *cpuScaleUpdate) ResourceType() scaleresource.ResourceType {
	return scaleresource.ResourceTypeCpu
}

func (c *cpuScaleUpdate) SetStartupResources(
	pod *v1.Pod,
	container *v1.Container,
	clonePod bool,
) (*v1.Pod, error) {
	if !c.scaleConfig.IsEnabled() {
		return pod, nil
	}

	newPod, err := c.setResources(
		pod,
		container,
		v1.ResourceCPU,
		c.scaleConfig.Resources().Startup,
		c.scaleConfig.Resources().Startup,
		clonePod,
	)
	if err != nil {
		return nil, common.WrapErrorf(err, "unable to set cpu startup resources")
	}

	return newPod, nil
}

func (c *cpuScaleUpdate) SetPostStartupResources(
	pod *v1.Pod,
	container *v1.Container,
	clonePod bool,
) (*v1.Pod, error) {
	if !c.scaleConfig.IsEnabled() {
		return pod, nil
	}

	newPod, err := c.setResources(
		pod,
		container,
		v1.ResourceCPU,
		c.scaleConfig.Resources().PostStartupRequests,
		c.scaleConfig.Resources().PostStartupLimits,
		clonePod,
	)
	if err != nil {
		return nil, common.WrapErrorf(err, "unable to set cpu post-startup resources")
	}

	return newPod, nil
}
