/*
Copyright 2025 Expedia Group, Inc.

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

package kubetest

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type ContainerBuilder struct {
	enabledResources []v1.ResourceName
	resourcesState   podcommon.StateResources
	startupProbe     bool
	readinessProbe   bool
	nilResizePolicy  bool
	nilRequests      bool
	nilLimits        bool
}

func NewContainerBuilder() *ContainerBuilder {
	b := &ContainerBuilder{}
	b.EnabledResourcesAll()
	b.ResourcesState(podcommon.StateResourcesStartup)
	return b
}

func (b *ContainerBuilder) EnabledResources(enabledResources []v1.ResourceName) *ContainerBuilder {
	b.enabledResources = enabledResources
	return b
}

func (b *ContainerBuilder) EnabledResourcesAll() *ContainerBuilder {
	b.enabledResources = []v1.ResourceName{v1.ResourceCPU, v1.ResourceMemory}
	return b
}

func (b *ContainerBuilder) EnabledResourcesNone() *ContainerBuilder {
	b.enabledResources = []v1.ResourceName{}
	return b
}

func (b *ContainerBuilder) ResourcesState(resourcesState podcommon.StateResources) *ContainerBuilder {
	b.resourcesState = resourcesState
	return b
}

func (b *ContainerBuilder) StartupProbe(startupProbe bool) *ContainerBuilder {
	b.startupProbe = startupProbe
	return b
}

func (b *ContainerBuilder) ReadinessProbe(readinessProbe bool) *ContainerBuilder {
	b.readinessProbe = readinessProbe
	return b
}

func (b *ContainerBuilder) NilResizePolicy(nilResizePolicy bool) *ContainerBuilder {
	b.nilResizePolicy = nilResizePolicy
	return b
}

func (b *ContainerBuilder) NilRequests(nilRequests bool) *ContainerBuilder {
	b.nilRequests = nilRequests
	return b
}

func (b *ContainerBuilder) NilLimits(nilLimits bool) *ContainerBuilder {
	b.nilLimits = nilLimits
	return b
}

func (b *ContainerBuilder) Build() *v1.Container {
	c := b.container()

	if b.nilResizePolicy {
		c.ResizePolicy = nil
	}

	if b.nilRequests {
		c.Resources.Requests = nil
	}

	if b.nilLimits {
		c.Resources.Limits = nil
	}

	return c
}

func (b *ContainerBuilder) container() *v1.Container {
	cpuRequests, cpuLimits, memoryRequests, memoryLimits := quantities(b.enabledResources, b.resourcesState)

	var startupProbe *v1.Probe
	if b.startupProbe {
		startupProbe = &v1.Probe{}
	}

	var readinessProbe *v1.Probe
	if b.readinessProbe {
		readinessProbe = &v1.Probe{}
	}

	return &v1.Container{
		Name: DefaultContainerName,
		Resources: v1.ResourceRequirements{
			Requests: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    cpuRequests,
				v1.ResourceMemory: memoryRequests,
			},
			Limits: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    cpuLimits,
				v1.ResourceMemory: memoryLimits,
			},
		},
		ResizePolicy: []v1.ContainerResizePolicy{
			{
				ResourceName:  v1.ResourceCPU,
				RestartPolicy: DefaultContainerCpuResizePolicy,
			},
			{
				ResourceName:  v1.ResourceMemory,
				RestartPolicy: DefaultContainerMemoryResizePolicy,
			},
		},
		StartupProbe:   startupProbe,
		ReadinessProbe: readinessProbe,
	}
}
