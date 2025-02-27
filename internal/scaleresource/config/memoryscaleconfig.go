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
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource/scaleresourcecommon"
	"k8s.io/api/core/v1"
)

type memoryScaleConfig struct {
	anyScaleConfig

	csaEnabled  bool
	userEnabled bool
	resources   Resources

	hasStoredFromAnnotations bool
	podHelper                kube.PodHelper
	containerHelper          kube.ContainerHelper
}

func NewMemoryScaleConfig(
	csaEnabled bool,
	podHelper kube.PodHelper,
	containerHelper kube.ContainerHelper,
) *memoryScaleConfig {
	return &memoryScaleConfig{
		csaEnabled:      csaEnabled,
		podHelper:       podHelper,
		containerHelper: containerHelper,
	}
}

func (m *memoryScaleConfig) ResourceType() scaleresource.ResourceType {
	return scaleresource.ResourceTypeMemory
}

func (m *memoryScaleConfig) IsEnabled() bool {
	m.checkStored(m.hasStoredFromAnnotations)
	return m.csaEnabled && m.userEnabled
}

func (m *memoryScaleConfig) IsCsaEnabled() bool {
	return m.csaEnabled
}

func (m *memoryScaleConfig) IsUserEnabled() bool {
	m.checkStored(m.hasStoredFromAnnotations)
	return m.userEnabled
}

func (m *memoryScaleConfig) Resources() Resources {
	m.checkStored(m.hasStoredFromAnnotations)
	return m.resources
}

func (m *memoryScaleConfig) StoreFromAnnotations(pod *v1.Pod) error {
	if !m.csaEnabled {
		m.hasStoredFromAnnotations = true
		return nil
	}

	if m.hasNoResourceAnnotations(
		m.podHelper,
		pod,
		scaleresourcecommon.AnnotationMemoryStartup,
		scaleresourcecommon.AnnotationMemoryPostStartupRequests,
		scaleresourcecommon.AnnotationMemoryPostStartupLimits,
	) {
		m.userEnabled = false
		m.hasStoredFromAnnotations = true
		return nil
	}

	startup, postStartupRequests, postStartupLimits, err := m.parseResourceAnnotations(
		m.podHelper,
		pod,
		scaleresourcecommon.AnnotationMemoryStartup,
		scaleresourcecommon.AnnotationMemoryPostStartupRequests,
		scaleresourcecommon.AnnotationMemoryPostStartupLimits,
	)
	if err != nil {
		return common.WrapErrorf(err, "unable to parse resource annotations")
	}

	m.userEnabled = true
	m.resources = newResources(startup, postStartupRequests, postStartupLimits)
	m.hasStoredFromAnnotations = true

	return nil
}

func (m *memoryScaleConfig) Validate(container *v1.Container) error {
	m.checkStored(m.hasStoredFromAnnotations)

	if !m.csaEnabled || !m.userEnabled {
		return nil
	}

	if err := m.validate(
		scaleresource.ResourceTypeMemory,
		v1.ResourceMemory,
		m.resources,
		container,
		m.containerHelper,
	); err != nil {
		return err
	}

	return nil
}

func (m *memoryScaleConfig) String() string {
	m.checkStored(m.hasStoredFromAnnotations)
	return m.string(m.ResourceType(), m.csaEnabled, m.userEnabled, m.resources)
}
