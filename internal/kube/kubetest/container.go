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
	"errors"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	DefaultContainerName = "container"
)

var (
	DefaultContainerCpuResizePolicy    = v1.NotRequired
	DefaultContainerMemoryResizePolicy = v1.NotRequired
)

// containerConfig holds configuration for generating a test container.
type containerConfig struct {
	name               string
	cpuRequests        string
	cpuLimits          string
	memoryRequests     string
	memoryLimits       string
	cpuResizePolicy    v1.ResourceResizeRestartPolicy
	memoryResizePolicy v1.ResourceResizeRestartPolicy
}

// TODO(wt) construct based on enabledResources
func newContainerConfig(enabledResources []v1.ResourceName, stateResources podcommon.StateResources) containerConfig {
	config := containerConfig{
		name:               DefaultContainerName,
		cpuResizePolicy:    DefaultContainerCpuResizePolicy,
		memoryResizePolicy: DefaultContainerMemoryResizePolicy,
	}

	switch stateResources {
	case podcommon.StateResourcesStartup:
		config.cpuRequests = PodAnnotationCpuStartupEnabled
		config.cpuLimits = PodAnnotationCpuStartupEnabled
		config.memoryRequests = PodAnnotationMemoryStartupEnabled
		config.memoryLimits = PodAnnotationMemoryStartupEnabled

	case podcommon.StateResourcesPostStartup:
		config.cpuRequests = PodAnnotationCpuPostStartupRequestsEnabled
		config.cpuLimits = PodAnnotationCpuPostStartupLimitsEnabled
		config.memoryRequests = PodAnnotationMemoryPostStartupRequestsEnabled
		config.memoryLimits = PodAnnotationMemoryPostStartupLimitsEnabled

	case podcommon.StateResourcesUnknown:
		config.cpuRequests = PodAnnotationCpuUnknown
		config.cpuLimits = PodAnnotationCpuUnknown
		config.memoryRequests = PodAnnotationMemoryUnknown
		config.memoryLimits = PodAnnotationMemoryUnknown

	default:
		panic(errors.New("invalid stateResources"))
	}

	return config
}

// container returns a test container from the supplied config.
func container(config containerConfig) *v1.Container {
	return &v1.Container{
		Name: config.name,
		Resources: v1.ResourceRequirements{
			Requests: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    resource.MustParse(config.cpuRequests),
				v1.ResourceMemory: resource.MustParse(config.memoryRequests),
			},
			Limits: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    resource.MustParse(config.cpuLimits),
				v1.ResourceMemory: resource.MustParse(config.memoryLimits),
			},
		},
		ResizePolicy: []v1.ContainerResizePolicy{
			{
				ResourceName:  v1.ResourceCPU,
				RestartPolicy: config.cpuResizePolicy,
			},
			{
				ResourceName:  v1.ResourceMemory,
				RestartPolicy: config.memoryResizePolicy,
			},
		},
	}
}
