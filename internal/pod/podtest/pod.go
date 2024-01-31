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

package podtest

import (
	"errors"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	PodAnnotationCpuStartup             = "3m"
	PodAnnotationCpuPostStartupRequests = "1m"
	PodAnnotationCpuPostStartupLimits   = "2m"

	PodAnnotationMemoryStartup             = "3M"
	PodAnnotationMemoryPostStartupRequests = "1M"
	PodAnnotationMemoryPostStartupLimits   = "2M"

	PodAnnotationCpuUnknown    = "999m"
	PodAnnotationMemoryUnknown = "999M"
)

var (
	PodAnnotationCpuStartupQuantity             = resource.MustParse(PodAnnotationCpuStartup)
	PodAnnotationCpuPostStartupRequestsQuantity = resource.MustParse(PodAnnotationCpuPostStartupRequests)
	PodAnnotationCpuPostStartupLimitsQuantity   = resource.MustParse(PodAnnotationCpuPostStartupLimits)

	PodAnnotationMemoryStartupQuantity             = resource.MustParse(PodAnnotationMemoryStartup)
	PodAnnotationMemoryPostStartupRequestsQuantity = resource.MustParse(PodAnnotationMemoryPostStartupRequests)
	PodAnnotationMemoryPostStartupLimitsQuantity   = resource.MustParse(PodAnnotationMemoryPostStartupLimits)

	PodAnnotationCpuUnknownQuantity    = resource.MustParse(PodAnnotationCpuUnknown)
	PodAnnotationMemoryUnknownQuantity = resource.MustParse(PodAnnotationMemoryUnknown)
)

const (
	DefaultPodNamespace                  = "namespace"
	DefaultPodName                       = "name"
	DefaultLabelEnabledValue             = "true"
	DefaultAnnotationTargetContainerName = DefaultContainerName
	DefaultStatusContainerName           = DefaultContainerName
)

var (
	DefaultPodNamespacedName = types.NamespacedName{
		Namespace: DefaultPodNamespace,
		Name:      DefaultPodName,
	}
	DefaultPodStatusContainerState = corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}
	DefaultPodStatusResize         = corev1.PodResizeStatus("")
)

// podConfig holds configuration for generating a test pod.
type podConfig struct {
	namespace                               string
	name                                    string
	labelEnabledValue                       string
	annotationTargetContainerName           string
	annotationCpuStartup                    string
	annotationCpuPostStartupRequests        string
	annotationCpuPostStartupLimits          string
	annotationMemoryStartup                 string
	annotationMemoryPostStartupRequests     string
	annotationMemoryPostStartupLimits       string
	statusContainerName                     string
	statusContainerState                    corev1.ContainerState
	statusContainerStarted                  bool
	statusContainerReady                    bool
	statusContainerAllocatedResourcesCpu    string
	statusContainerAllocatedResourcesMemory string
	statusContainerResourcesCpuRequests     string
	statusContainerResourcesCpuLimits       string
	statusContainerResourcesMemoryRequests  string
	statusContainerResourcesMemoryLimits    string
	statusResize                            corev1.PodResizeStatus
	containerConfig                         containerConfig
}

func NewStartupPodConfig(stateStarted podcommon.StateBool, stateReady podcommon.StateBool) podConfig {
	return newPodConfigForState(stateStarted, stateReady, podcommon.StateResourcesStartup)
}

func NewPostStartupPodConfig(stateStarted podcommon.StateBool, stateReady podcommon.StateBool) podConfig {
	return newPodConfigForState(stateStarted, stateReady, podcommon.StateResourcesPostStartup)
}

func NewUnknownPodConfig(stateStarted podcommon.StateBool, stateReady podcommon.StateBool) podConfig {
	return newPodConfigForState(stateStarted, stateReady, podcommon.StateResourcesUnknown)
}

func newPodConfigForState(
	stateStarted podcommon.StateBool,
	stateReady podcommon.StateBool,
	stateResources podcommon.StateResources,
) podConfig {
	config := podConfig{
		namespace:                           DefaultPodNamespace,
		name:                                DefaultPodName,
		labelEnabledValue:                   DefaultLabelEnabledValue,
		annotationTargetContainerName:       DefaultAnnotationTargetContainerName,
		annotationCpuStartup:                PodAnnotationCpuStartup,
		annotationCpuPostStartupRequests:    PodAnnotationCpuPostStartupRequests,
		annotationCpuPostStartupLimits:      PodAnnotationCpuPostStartupLimits,
		annotationMemoryStartup:             PodAnnotationMemoryStartup,
		annotationMemoryPostStartupRequests: PodAnnotationMemoryPostStartupRequests,
		annotationMemoryPostStartupLimits:   PodAnnotationMemoryPostStartupLimits,
		statusContainerName:                 DefaultStatusContainerName,
		statusContainerState:                DefaultPodStatusContainerState,
		statusResize:                        DefaultPodStatusResize,
		containerConfig:                     newContainerConfigForState(stateResources),
	}

	switch stateStarted {
	case podcommon.StateBoolTrue:
		config.statusContainerStarted = true
	case podcommon.StateBoolFalse:
		config.statusContainerStarted = false
	default:
		panic(errors.New("invalid stateStarted"))
	}

	switch stateReady {
	case podcommon.StateBoolTrue:
		config.statusContainerReady = true
	case podcommon.StateBoolFalse:
		config.statusContainerReady = false
	default:
		panic(errors.New("invalid stateReady"))
	}

	switch stateResources {
	case podcommon.StateResourcesStartup:
		config.statusContainerAllocatedResourcesCpu = PodAnnotationCpuStartup
		config.statusContainerAllocatedResourcesMemory = PodAnnotationMemoryStartup
		config.statusContainerResourcesCpuRequests = PodAnnotationCpuStartup
		config.statusContainerResourcesCpuLimits = PodAnnotationCpuStartup
		config.statusContainerResourcesMemoryRequests = PodAnnotationMemoryStartup
		config.statusContainerResourcesMemoryLimits = PodAnnotationMemoryStartup

	case podcommon.StateResourcesPostStartup:
		config.statusContainerAllocatedResourcesCpu = PodAnnotationCpuPostStartupRequests
		config.statusContainerAllocatedResourcesMemory = PodAnnotationMemoryPostStartupRequests
		config.statusContainerResourcesCpuRequests = PodAnnotationCpuPostStartupRequests
		config.statusContainerResourcesCpuLimits = PodAnnotationCpuPostStartupLimits
		config.statusContainerResourcesMemoryRequests = PodAnnotationMemoryPostStartupRequests
		config.statusContainerResourcesMemoryLimits = PodAnnotationMemoryPostStartupLimits

	case podcommon.StateResourcesUnknown:
		config.statusContainerAllocatedResourcesCpu = PodAnnotationCpuUnknown
		config.statusContainerAllocatedResourcesMemory = PodAnnotationMemoryUnknown
		config.statusContainerResourcesCpuRequests = PodAnnotationCpuUnknown
		config.statusContainerResourcesCpuLimits = PodAnnotationCpuUnknown
		config.statusContainerResourcesMemoryRequests = PodAnnotationMemoryUnknown
		config.statusContainerResourcesMemoryLimits = PodAnnotationMemoryUnknown

	default:
		panic(errors.New("invalid stateResources"))
	}

	return config
}

// pod returns a test pod from the supplied config.
func pod(config podConfig) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: config.namespace,
			Name:      config.name,
			Labels: map[string]string{
				podcommon.LabelEnabled: config.labelEnabledValue,
			},
			Annotations: map[string]string{
				podcommon.AnnotationTargetContainerName:       config.annotationTargetContainerName,
				podcommon.AnnotationCpuStartup:                config.annotationCpuStartup,
				podcommon.AnnotationCpuPostStartupRequests:    config.annotationCpuPostStartupRequests,
				podcommon.AnnotationCpuPostStartupLimits:      config.annotationCpuPostStartupLimits,
				podcommon.AnnotationMemoryStartup:             config.annotationMemoryStartup,
				podcommon.AnnotationMemoryPostStartupRequests: config.annotationMemoryPostStartupRequests,
				podcommon.AnnotationMemoryPostStartupLimits:   config.annotationMemoryPostStartupLimits,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				*container(config.containerConfig),
			},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:    config.statusContainerName,
					State:   config.statusContainerState,
					Started: &config.statusContainerStarted,
					Ready:   config.statusContainerReady,
					AllocatedResources: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceCPU:    resource.MustParse(config.statusContainerAllocatedResourcesCpu),
						corev1.ResourceMemory: resource.MustParse(config.statusContainerAllocatedResourcesMemory),
					},
					Resources: &corev1.ResourceRequirements{
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse(config.statusContainerResourcesCpuRequests),
							corev1.ResourceMemory: resource.MustParse(config.statusContainerResourcesMemoryRequests),
						},
						Limits: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    resource.MustParse(config.statusContainerResourcesCpuLimits),
							corev1.ResourceMemory: resource.MustParse(config.statusContainerResourcesMemoryLimits),
						},
					},
				},
			},
			Resize: config.statusResize,
		},
	}
}
