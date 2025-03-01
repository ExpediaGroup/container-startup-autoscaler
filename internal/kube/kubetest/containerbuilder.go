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

package kubetest

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"k8s.io/api/core/v1"
)

// ctxBuilder builds a test container.
type containerBuilder struct {
	enabledResources []v1.ResourceName
	resourcesState   podcommon.StateResources
	startupProbe     bool
	readinessProbe   bool
	nilResizePolicy  bool
	nilRequests      bool
	nilLimits        bool
}

func NewContainerBuilder() *containerBuilder {
	b := &containerBuilder{}
	b.enabledResources = []v1.ResourceName{v1.ResourceCPU, v1.ResourceMemory}
	b.resourcesState = podcommon.StateResourcesStartup
	return b
}

func (b *containerBuilder) EnabledResources(enabledResources []v1.ResourceName) *containerBuilder {
	b.enabledResources = enabledResources
	return b
}

func (b *containerBuilder) ResourcesStatePostStartup() *containerBuilder {
	b.resourcesState = podcommon.StateResourcesPostStartup
	return b
}

func (b *containerBuilder) StartupProbe() *containerBuilder {
	b.startupProbe = true
	return b
}

func (b *containerBuilder) ReadinessProbe() *containerBuilder {
	b.readinessProbe = true
	return b
}

func (b *containerBuilder) NilResizePolicy() *containerBuilder {
	b.nilResizePolicy = true
	return b
}

func (b *containerBuilder) NilRequests() *containerBuilder {
	b.nilRequests = true
	return b
}

func (b *containerBuilder) NilLimits() *containerBuilder {
	b.nilLimits = true
	return b
}

func (b *containerBuilder) Build() *v1.Container {
	// TODO(wt) move everything from container.go and remove container.go

	c := container(newContainerConfig(b.enabledResources, b.resourcesState))

	if b.startupProbe {
		c.StartupProbe = &v1.Probe{}
	}

	if b.readinessProbe {
		c.ReadinessProbe = &v1.Probe{}
	}

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
