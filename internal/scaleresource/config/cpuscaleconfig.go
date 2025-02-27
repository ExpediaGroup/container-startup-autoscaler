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

package config

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource/scaleresourcecommon"
	"k8s.io/api/core/v1"
)

type cpuScaleConfig struct {
	anyScaleConfig

	csaEnabled  bool
	userEnabled bool
	resources   Resources

	hasStoredFromAnnotations bool
	podHelper                kube.PodHelper
	containerHelper          kube.ContainerHelper
}

func NewCpuScaleConfig(
	csaEnabled bool,
	podHelper kube.PodHelper,
	containerHelper kube.ContainerHelper,
) *cpuScaleConfig {
	return &cpuScaleConfig{
		csaEnabled:      csaEnabled,
		podHelper:       podHelper,
		containerHelper: containerHelper,
	}
}

func (c *cpuScaleConfig) ResourceType() scaleresource.ResourceType {
	return scaleresource.ResourceTypeCpu
}

func (c *cpuScaleConfig) IsEnabled() bool {
	c.checkStored(c.hasStoredFromAnnotations)
	return c.csaEnabled && c.userEnabled
}

func (c *cpuScaleConfig) IsCsaEnabled() bool {
	return c.csaEnabled
}

func (c *cpuScaleConfig) IsUserEnabled() bool {
	c.checkStored(c.hasStoredFromAnnotations)
	return c.userEnabled
}

func (c *cpuScaleConfig) Resources() Resources {
	c.checkStored(c.hasStoredFromAnnotations)
	return c.resources
}

func (c *cpuScaleConfig) StoreFromAnnotations(pod *v1.Pod) error {
	if !c.csaEnabled {
		c.hasStoredFromAnnotations = true
		return nil
	}

	if c.hasNoResourceAnnotations(
		c.podHelper,
		pod,
		scaleresourcecommon.AnnotationCpuStartup,
		scaleresourcecommon.AnnotationCpuPostStartupRequests,
		scaleresourcecommon.AnnotationCpuPostStartupLimits,
	) {
		c.userEnabled = false
		c.hasStoredFromAnnotations = true
		return nil
	}

	startup, postStartupRequests, postStartupLimits, err := c.parseResourceAnnotations(
		c.podHelper,
		pod,
		scaleresourcecommon.AnnotationCpuStartup,
		scaleresourcecommon.AnnotationCpuPostStartupRequests,
		scaleresourcecommon.AnnotationCpuPostStartupLimits,
	)
	if err != nil {
		return err
	}

	c.userEnabled = true
	c.resources = newResources(startup, postStartupRequests, postStartupLimits)
	c.hasStoredFromAnnotations = true

	return nil
}

func (c *cpuScaleConfig) Validate(container *v1.Container) error {
	c.checkStored(c.hasStoredFromAnnotations)

	if !c.csaEnabled || !c.userEnabled {
		return nil
	}

	if err := c.validate(
		scaleresource.ResourceTypeCpu,
		v1.ResourceCPU,
		c.resources,
		container,
		c.containerHelper,
	); err != nil {
		return err
	}

	return nil
}

func (c *cpuScaleConfig) String() string {
	c.checkStored(c.hasStoredFromAnnotations)
	return c.string(c.ResourceType(), c.csaEnabled, c.userEnabled, c.resources)
}
