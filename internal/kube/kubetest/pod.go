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
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	PodAnnotationCpuStartupEnabled             = "3m"
	PodAnnotationCpuPostStartupRequestsEnabled = "1m"
	PodAnnotationCpuPostStartupLimitsEnabled   = "2m"

	PodAnnotationCpuStartupDisabled             = PodAnnotationCpuStartupEnabled
	PodAnnotationCpuPostStartupRequestsDisabled = PodAnnotationCpuStartupEnabled
	PodAnnotationCpuPostStartupLimitsDisabled   = PodAnnotationCpuStartupEnabled

	PodAnnotationMemoryStartupEnabled             = "3M"
	PodAnnotationMemoryPostStartupRequestsEnabled = "1M"
	PodAnnotationMemoryPostStartupLimitsEnabled   = "2M"

	PodAnnotationMemoryStartupDisabled             = PodAnnotationMemoryStartupEnabled
	PodAnnotationMemoryPostStartupRequestsDisabled = PodAnnotationMemoryStartupEnabled
	PodAnnotationMemoryPostStartupLimitsDisabled   = PodAnnotationMemoryStartupEnabled

	PodAnnotationCpuUnknown    = "999m"
	PodAnnotationMemoryUnknown = "999M"
)

var (
	PodAnnotationCpuStartupEnabledQuantity             = resource.MustParse(PodAnnotationCpuStartupEnabled)
	PodAnnotationCpuPostStartupRequestsEnabledQuantity = resource.MustParse(PodAnnotationCpuPostStartupRequestsEnabled)
	PodAnnotationCpuPostStartupLimitsEnabledQuantity   = resource.MustParse(PodAnnotationCpuPostStartupLimitsEnabled)

	PodAnnotationCpuStartupDisabledQuantity             = PodAnnotationCpuStartupEnabledQuantity
	PodAnnotationCpuPostStartupRequestsDisabledQuantity = PodAnnotationCpuStartupEnabledQuantity
	PodAnnotationCpuPostStartupLimitsDisabledQuantity   = PodAnnotationCpuStartupEnabledQuantity

	PodAnnotationMemoryStartupEnabledQuantity             = resource.MustParse(PodAnnotationMemoryStartupEnabled)
	PodAnnotationMemoryPostStartupRequestsEnabledQuantity = resource.MustParse(PodAnnotationMemoryPostStartupRequestsEnabled)
	PodAnnotationMemoryPostStartupLimitsEnabledQuantity   = resource.MustParse(PodAnnotationMemoryPostStartupLimitsEnabled)

	PodAnnotationMemoryStartupDisabledQuantity             = PodAnnotationMemoryStartupEnabledQuantity
	PodAnnotationMemoryPostStartupRequestsDisabledQuantity = PodAnnotationMemoryStartupEnabledQuantity
	PodAnnotationMemoryPostStartupLimitsDisabledQuantity   = PodAnnotationMemoryStartupEnabledQuantity

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
	namespace                              string
	name                                   string
	labelEnabledValue                      string
	annotationTargetContainerName          string
	annotationCpuStartup                   string
	annotationCpuPostStartupRequests       string
	annotationCpuPostStartupLimits         string
	annotationMemoryStartup                string
	annotationMemoryPostStartupRequests    string
	annotationMemoryPostStartupLimits      string
	statusContainerName                    string
	statusContainerState                   corev1.ContainerState
	statusContainerStarted                 bool
	statusContainerReady                   bool
	statusContainerResourcesCpuRequests    string
	statusContainerResourcesCpuLimits      string
	statusContainerResourcesMemoryRequests string
	statusContainerResourcesMemoryLimits   string
	statusResize                           corev1.PodResizeStatus
	containerConfig                        containerConfig
}

// TODO(wt) construct based on enabledResources
func newPodConfig(
	enabledResources []corev1.ResourceName,
	stateResources podcommon.StateResources,
	stateStarted podcommon.StateBool,
	stateReady podcommon.StateBool,
) podConfig {
	config := podConfig{
		namespace:                           DefaultPodNamespace,
		name:                                DefaultPodName,
		labelEnabledValue:                   DefaultLabelEnabledValue,
		annotationTargetContainerName:       DefaultAnnotationTargetContainerName,
		annotationCpuStartup:                PodAnnotationCpuStartupEnabled,
		annotationCpuPostStartupRequests:    PodAnnotationCpuPostStartupRequestsEnabled,
		annotationCpuPostStartupLimits:      PodAnnotationCpuPostStartupLimitsEnabled,
		annotationMemoryStartup:             PodAnnotationMemoryStartupEnabled,
		annotationMemoryPostStartupRequests: PodAnnotationMemoryPostStartupRequestsEnabled,
		annotationMemoryPostStartupLimits:   PodAnnotationMemoryPostStartupLimitsEnabled,
		statusContainerName:                 DefaultStatusContainerName,
		statusContainerState:                DefaultPodStatusContainerState,
		statusResize:                        DefaultPodStatusResize,
		containerConfig:                     newContainerConfig(enabledResources, stateResources),
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
		config.statusContainerResourcesCpuRequests = PodAnnotationCpuStartupEnabled
		config.statusContainerResourcesCpuLimits = PodAnnotationCpuStartupEnabled
		config.statusContainerResourcesMemoryRequests = PodAnnotationMemoryStartupEnabled
		config.statusContainerResourcesMemoryLimits = PodAnnotationMemoryStartupEnabled

	case podcommon.StateResourcesPostStartup:
		config.statusContainerResourcesCpuRequests = PodAnnotationCpuPostStartupRequestsEnabled
		config.statusContainerResourcesCpuLimits = PodAnnotationCpuPostStartupLimitsEnabled
		config.statusContainerResourcesMemoryRequests = PodAnnotationMemoryPostStartupRequestsEnabled
		config.statusContainerResourcesMemoryLimits = PodAnnotationMemoryPostStartupLimitsEnabled

	case podcommon.StateResourcesUnknown:
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
				scalecommon.AnnotationTargetContainerName:       config.annotationTargetContainerName,
				scalecommon.AnnotationCpuStartup:                config.annotationCpuStartup,
				scalecommon.AnnotationCpuPostStartupRequests:    config.annotationCpuPostStartupRequests,
				scalecommon.AnnotationCpuPostStartupLimits:      config.annotationCpuPostStartupLimits,
				scalecommon.AnnotationMemoryStartup:             config.annotationMemoryStartup,
				scalecommon.AnnotationMemoryPostStartupRequests: config.annotationMemoryPostStartupRequests,
				scalecommon.AnnotationMemoryPostStartupLimits:   config.annotationMemoryPostStartupLimits,
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
