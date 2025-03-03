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
	"errors"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"k8s.io/api/core/v1"
)

func NewConfigs(podHelper kubecommon.PodHelper, containerHelper kubecommon.ContainerHelper) scalecommon.Configs {
	return &configs{
		cpuConfig: NewConfig(
			v1.ResourceCPU,
			scalecommon.AnnotationCpuStartup,
			scalecommon.AnnotationCpuPostStartupRequests,
			scalecommon.AnnotationCpuPostStartupLimits,
			true,
			podHelper,
			containerHelper,
		),
		memoryConfig: NewConfig(
			v1.ResourceMemory,
			scalecommon.AnnotationMemoryStartup,
			scalecommon.AnnotationMemoryPostStartupRequests,
			scalecommon.AnnotationMemoryPostStartupLimits,
			true,
			podHelper,
			containerHelper,
		),
		podHelper: podHelper,
	}
}

type configs struct {
	cpuConfig    scalecommon.Config
	memoryConfig scalecommon.Config

	podHelper kubecommon.PodHelper
}

func (c *configs) TargetContainerName(pod *v1.Pod) (string, error) {
	value, err := c.podHelper.ExpectedAnnotationValueAs(
		pod,
		scalecommon.AnnotationTargetContainerName,
		kubecommon.DataTypeString,
	)
	if err != nil {
		return "", common.WrapErrorf(err, "unable to get '%s' annotation value", scalecommon.AnnotationTargetContainerName)
	}

	return value.(string), nil
}

func (c *configs) StoreFromAnnotationsAll(pod *v1.Pod) error {
	for _, config := range c.AllConfigs() {
		if err := config.StoreFromAnnotations(pod); err != nil {
			return err
		}
	}

	return nil
}

func (c *configs) ValidateAll(container *v1.Container) error {
	for _, config := range c.AllConfigs() {
		if err := config.Validate(container); err != nil {
			return err
		}
	}

	return nil
}

func (c *configs) ValidateCollection() error {
	atLeastOneEnabled := false

	for _, config := range c.AllConfigs() {
		if config.IsEnabled() {
			atLeastOneEnabled = true
			break
		}
	}

	if !atLeastOneEnabled {
		return errors.New("no resources are configured for scaling")
	}

	return nil
}

func (c *configs) ConfigFor(resourceName v1.ResourceName) scalecommon.Config {
	switch resourceName {
	case v1.ResourceCPU:
		return c.cpuConfig
	case v1.ResourceMemory:
		return c.memoryConfig
	default:
		return nil
	}
}

func (c *configs) AllConfigs() []scalecommon.Config {
	return []scalecommon.Config{c.cpuConfig, c.memoryConfig}
}

func (c *configs) AllEnabledConfigs() []scalecommon.Config {
	var enabledConfigs []scalecommon.Config

	for _, config := range c.AllConfigs() {
		if config.IsEnabled() {
			enabledConfigs = append(enabledConfigs, config)
		}
	}

	return enabledConfigs
}

func (c *configs) AllEnabledConfigsResourceNames() []v1.ResourceName {
	var enabledNames []v1.ResourceName

	for _, config := range c.AllConfigs() {
		if config.IsEnabled() {
			enabledNames = append(enabledNames, config.ResourceName())
		}
	}

	return enabledNames
}

func (c *configs) String() string {
	var result string
	allConfigs := c.AllConfigs()

	for i, config := range allConfigs {
		result += config.String()
		if i < len(allConfigs)-1 {
			result += ", "
		}
	}

	return result
}
