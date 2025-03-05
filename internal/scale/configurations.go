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

// configurations is the default implementation of scalecommon.Configurations.
type configurations struct {
	cpuConfig    scalecommon.Configuration
	memoryConfig scalecommon.Configuration

	podHelper kubecommon.PodHelper
}

func NewConfigurations(podHelper kubecommon.PodHelper, containerHelper kubecommon.ContainerHelper) scalecommon.Configurations {
	return &configurations{
		cpuConfig: NewConfiguration(
			v1.ResourceCPU,
			scalecommon.AnnotationCpuStartup,
			scalecommon.AnnotationCpuPostStartupRequests,
			scalecommon.AnnotationCpuPostStartupLimits,
			true,
			podHelper,
			containerHelper,
		),
		memoryConfig: NewConfiguration(
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

// TargetContainerName returns the target container name applicable for this collection of configurations.
func (c *configurations) TargetContainerName(pod *v1.Pod) (string, error) {
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

// StoreFromAnnotationsAll invokes StoreFromAnnotations on each configuration within this collection.
func (c *configurations) StoreFromAnnotationsAll(pod *v1.Pod) error {
	for _, config := range c.AllConfigurations() {
		if err := config.StoreFromAnnotations(pod); err != nil {
			return err
		}
	}

	return nil
}

// ValidateAll invokes Validate on each configuration within this collection.
func (c *configurations) ValidateAll(container *v1.Container) error {
	for _, config := range c.AllConfigurations() {
		if err := config.Validate(container); err != nil {
			return err
		}
	}

	return nil
}

// ValidateCollection performs validation on the entire configuration collection.
func (c *configurations) ValidateCollection() error {
	atLeastOneEnabled := false

	for _, config := range c.AllConfigurations() {
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

// ConfigurationFor returns the configuration for the supplied resource name.
func (c *configurations) ConfigurationFor(resourceName v1.ResourceName) scalecommon.Configuration {
	switch resourceName {
	case v1.ResourceCPU:
		return c.cpuConfig
	case v1.ResourceMemory:
		return c.memoryConfig
	default:
		return nil
	}
}

// AllConfigurations returns all configurations within this collection.
func (c *configurations) AllConfigurations() []scalecommon.Configuration {
	return []scalecommon.Configuration{c.cpuConfig, c.memoryConfig}
}

// AllEnabledConfigurations returns all enabled configurations within this collection.
func (c *configurations) AllEnabledConfigurations() []scalecommon.Configuration {
	var enabledConfigs []scalecommon.Configuration

	for _, config := range c.AllConfigurations() {
		if config.IsEnabled() {
			enabledConfigs = append(enabledConfigs, config)
		}
	}

	return enabledConfigs
}

// AllEnabledConfigurationsResourceNames returns the resource names of all enabled configurations within this
// collection.
func (c *configurations) AllEnabledConfigurationsResourceNames() []v1.ResourceName {
	var enabledNames []v1.ResourceName

	for _, config := range c.AllConfigurations() {
		if config.IsEnabled() {
			enabledNames = append(enabledNames, config.ResourceName())
		}
	}

	return enabledNames
}

// String returns a string representation of all configurations within this collection.
func (c *configurations) String() string {
	var result string
	allConfigs := c.AllConfigurations()

	for i, config := range allConfigs {
		result += config.String()
		if i < len(allConfigs)-1 {
			result += ", "
		}
	}

	return result
}
