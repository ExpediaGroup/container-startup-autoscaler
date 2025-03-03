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

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// podBuilder builds a test pod.
type podBuilder struct {
	enabledResources            []v1.ResourceName
	resourcesState              podcommon.StateResources
	stateStarted                podcommon.StateBool
	stateReady                  podcommon.StateBool
	containerStatusResizeStatus v1.PodResizeStatus
	additionalLabels            map[string]string
	additionalAnnotations       map[string]string
	nilContainerStatusStarted   bool
	nilContainerStatusResources bool

	containerCustomizerFunc func(*containerBuilder)
}

func NewPodBuilder() *podBuilder {
	b := &podBuilder{}
	b.enabledResources = []v1.ResourceName{v1.ResourceCPU, v1.ResourceMemory}
	b.resourcesState = podcommon.StateResourcesStartup
	b.stateStarted = podcommon.StateBoolFalse
	b.stateReady = podcommon.StateBoolFalse
	b.containerStatusResizeStatus = DefaultContainerStatusResizeStatus
	return b
}

func (b *podBuilder) EnabledResources(enabledResources []v1.ResourceName) *podBuilder {
	b.enabledResources = enabledResources
	return b
}

func (b *podBuilder) ResourcesStatePostStartup() *podBuilder {
	b.resourcesState = podcommon.StateResourcesPostStartup
	return b
}

func (b *podBuilder) ResourcesStateUnknown() *podBuilder {
	b.resourcesState = podcommon.StateResourcesUnknown
	return b
}

func (b *podBuilder) StateStartedTrue() *podBuilder {
	b.stateStarted = podcommon.StateBoolTrue
	return b
}

func (b *podBuilder) StateStartedUnknown() *podBuilder {
	b.stateStarted = podcommon.StateBoolUnknown
	return b
}

func (b *podBuilder) StateReadyTrue() *podBuilder {
	b.stateReady = podcommon.StateBoolTrue
	return b
}

func (b *podBuilder) StateReadyUnknown() *podBuilder {
	b.stateReady = podcommon.StateBoolUnknown
	return b
}

func (b *podBuilder) ContainerStatusResizeStatusProposed() *podBuilder {
	b.containerStatusResizeStatus = v1.PodResizeStatusProposed
	return b
}

func (b *podBuilder) ContainerStatusResizeStatusInProgress() *podBuilder {
	b.containerStatusResizeStatus = v1.PodResizeStatusInProgress
	return b
}

func (b *podBuilder) ContainerStatusResizeStatusDeferred() *podBuilder {
	b.containerStatusResizeStatus = v1.PodResizeStatusDeferred
	return b
}

func (b *podBuilder) ContainerStatusResizeStatusInfeasible() *podBuilder {
	b.containerStatusResizeStatus = v1.PodResizeStatusInfeasible
	return b
}

func (b *podBuilder) AdditionalLabels(labels map[string]string) *podBuilder {
	b.additionalLabels = labels
	return b
}

func (b *podBuilder) AdditionalAnnotations(annotations map[string]string) *podBuilder {
	b.additionalAnnotations = annotations
	return b
}

func (b *podBuilder) NilContainerStatusStarted() *podBuilder {
	b.nilContainerStatusStarted = true
	return b
}

func (b *podBuilder) NilContainerStatusResources() *podBuilder {
	b.nilContainerStatusResources = true
	return b
}

func (b *podBuilder) ContainerCustomizerFunc(containerCustomizerFunc func(*containerBuilder)) *podBuilder {
	b.containerCustomizerFunc = containerCustomizerFunc
	return b
}

func (b *podBuilder) Build() *v1.Pod {
	p := b.pod()

	if b.nilContainerStatusStarted {
		p.Status.ContainerStatuses[0].Started = nil
	}

	if b.nilContainerStatusResources {
		p.Status.ContainerStatuses[0].Resources = nil
	}

	return p
}

func (b *podBuilder) pod() *v1.Pod {
	var labels = make(map[string]string)
	labels[kubecommon.LabelEnabled] = DefaultLabelEnabledValue

	for name, value := range b.additionalLabels {
		labels[name] = value
	}

	var annotations = make(map[string]string)
	annotations[scalecommon.AnnotationTargetContainerName] = DefaultAnnotationTargetContainerName

	if containsResource(b.enabledResources, v1.ResourceCPU) {
		annotations[scalecommon.AnnotationCpuStartup] = PodAnnotationCpuStartup
		annotations[scalecommon.AnnotationCpuPostStartupRequests] = PodAnnotationCpuPostStartupRequests
		annotations[scalecommon.AnnotationCpuPostStartupLimits] = PodAnnotationCpuPostStartupLimits
	}

	if containsResource(b.enabledResources, v1.ResourceMemory) {
		annotations[scalecommon.AnnotationMemoryStartup] = PodAnnotationMemoryStartup
		annotations[scalecommon.AnnotationMemoryPostStartupRequests] = PodAnnotationMemoryPostStartupRequests
		annotations[scalecommon.AnnotationMemoryPostStartupLimits] = PodAnnotationMemoryPostStartupLimits
	}

	for name, value := range b.additionalAnnotations {
		annotations[name] = value
	}

	cpuRequests, cpuLimits, memoryRequests, memoryLimits := quantities(b.enabledResources, b.resourcesState)

	var stateStarted, stateReady bool

	switch b.stateStarted {
	case podcommon.StateBoolTrue:
		stateStarted = true
	case podcommon.StateBoolFalse:
		stateStarted = false
	default:
		panic(errors.New("invalid stateStarted"))
	}

	switch b.stateReady {
	case podcommon.StateBoolTrue:
		stateReady = true
	case podcommon.StateBoolFalse:
		stateReady = false
	default:
		panic(errors.New("invalid stateReady"))
	}

	builder := NewContainerBuilder().EnabledResources(b.enabledResources).ResourcesState(b.resourcesState)
	if b.containerCustomizerFunc != nil {
		b.containerCustomizerFunc(builder)
	}
	container := builder.Build()

	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       DefaultPodNamespace,
			Name:            DefaultPodName,
			Labels:          labels,
			Annotations:     annotations,
			ResourceVersion: DefaultPodResourceVersion,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{*container},
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:    DefaultStatusContainerName,
					State:   DefaultPodStatusContainerState,
					Started: &stateStarted,
					Ready:   stateReady,
					Resources: &v1.ResourceRequirements{
						Requests: map[v1.ResourceName]resource.Quantity{
							v1.ResourceCPU:    cpuRequests,
							v1.ResourceMemory: memoryRequests,
						},
						Limits: map[v1.ResourceName]resource.Quantity{
							v1.ResourceCPU:    cpuLimits,
							v1.ResourceMemory: memoryLimits,
						},
					},
				},
			},
			Resize: b.containerStatusResizeStatus,
		},
	}
}

func containsResource(resources []v1.ResourceName, resource v1.ResourceName) bool {
	for _, r := range resources {
		if r == resource {
			return true
		}
	}
	return false
}

func quantities(enabledResources []v1.ResourceName, resourcesState podcommon.StateResources) (
	cpuRequests resource.Quantity,
	cpuLimits resource.Quantity,
	memoryRequests resource.Quantity,
	memoryLimits resource.Quantity,
) {
	switch resourcesState {
	case podcommon.StateResourcesStartup:
		if containsResource(enabledResources, v1.ResourceCPU) {
			cpuRequests = PodCpuStartupEnabled
			cpuLimits = PodCpuStartupEnabled
		} else {
			cpuRequests = PodCpuStartupDisabled
			cpuLimits = PodCpuStartupDisabled
		}

		if containsResource(enabledResources, v1.ResourceCPU) {
			memoryRequests = PodMemoryStartupEnabled
			memoryLimits = PodMemoryStartupEnabled
		} else {
			memoryRequests = PodMemoryStartupDisabled
			memoryLimits = PodMemoryStartupDisabled
		}

	case podcommon.StateResourcesPostStartup:
		if containsResource(enabledResources, v1.ResourceCPU) {
			cpuRequests = PodCpuPostStartupRequestsEnabled
			cpuLimits = PodCpuPostStartupLimitsEnabled
		} else {
			cpuRequests = PodCpuPostStartupRequestsDisabled
			cpuLimits = PodCpuPostStartupLimitsDisabled
		}

		if containsResource(enabledResources, v1.ResourceCPU) {
			memoryRequests = PodMemoryPostStartupRequestsEnabled
			memoryLimits = PodMemoryPostStartupLimitsEnabled
		} else {
			memoryRequests = PodMemoryPostStartupRequestsDisabled
			memoryLimits = PodMemoryPostStartupLimitsDisabled
		}

	case podcommon.StateResourcesUnknown:
		cpuRequests = PodCpuUnknown
		cpuLimits = PodCpuUnknown
		memoryRequests = PodMemoryUnknown
		memoryLimits = PodMemoryUnknown

	default:
		panic(errors.New("invalid resourcesState"))
	}

	return
}
