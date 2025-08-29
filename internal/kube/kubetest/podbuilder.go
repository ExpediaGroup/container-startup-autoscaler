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
	"errors"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodBuilder struct {
	enabledResources            []v1.ResourceName
	resourcesState              podcommon.StateResources
	stateStarted                podcommon.StateBool
	stateReady                  podcommon.StateBool
	resizeConditions            []v1.PodCondition
	qosClass                    v1.PodQOSClass
	additionalLabels            map[string]string
	additionalAnnotations       map[string]string
	nilContainerStatusStarted   bool
	nilContainerStatusResources bool

	containerCustomizerFunc func(*ContainerBuilder)
}

func NewPodBuilder() *PodBuilder {
	b := &PodBuilder{}
	b.EnabledResourcesAll()
	b.ResourcesState(podcommon.StateResourcesStartup)
	b.StateStarted(podcommon.StateBoolFalse)
	b.StateReady(podcommon.StateBoolFalse)
	b.ResizeConditions()
	b.QOSClass(v1.PodQOSGuaranteed)
	return b
}

func (b *PodBuilder) EnabledResources(enabledResources []v1.ResourceName) *PodBuilder {
	b.enabledResources = enabledResources
	return b
}

func (b *PodBuilder) EnabledResourcesAll() *PodBuilder {
	b.enabledResources = []v1.ResourceName{v1.ResourceCPU, v1.ResourceMemory}
	return b
}

func (b *PodBuilder) EnabledResourcesNone() *PodBuilder {
	b.enabledResources = []v1.ResourceName{}
	return b
}

func (b *PodBuilder) ResourcesState(resourcesState podcommon.StateResources) *PodBuilder {
	b.resourcesState = resourcesState
	return b
}

func (b *PodBuilder) StateStarted(stateStarted podcommon.StateBool) *PodBuilder {
	b.stateStarted = stateStarted
	return b
}

func (b *PodBuilder) StateReady(stateReady podcommon.StateBool) *PodBuilder {
	b.stateReady = stateReady
	return b
}

func (b *PodBuilder) ResizeConditions(conditions ...v1.PodCondition) *PodBuilder {
	b.resizeConditions = conditions
	return b
}

func (b *PodBuilder) ResizeConditionsNotStartedOrCompletedNoConditions() *PodBuilder {
	b.resizeConditions = PodResizeConditionsNotStartedOrCompletedNoConditions
	return b
}

func (b *PodBuilder) ResizeConditionsInProgress() *PodBuilder {
	b.resizeConditions = PodResizeConditionsInProgress
	return b
}

func (b *PodBuilder) ResizeConditionsDeferred(message string) *PodBuilder {
	b.resizeConditions = PodResizeConditionsDeferred
	b.resizeConditions[0].Message = message
	return b
}

func (b *PodBuilder) ResizeConditionsInfeasible(message string) *PodBuilder {
	b.resizeConditions = PodResizeConditionsInfeasible
	b.resizeConditions[0].Message = message
	return b
}

func (b *PodBuilder) ResizeConditionsError(message string) *PodBuilder {
	b.resizeConditions = PodResizeConditionsError
	b.resizeConditions[0].Message = message
	return b
}

func (b *PodBuilder) ResizeConditionsUnknownPending() *PodBuilder {
	b.resizeConditions = PodResizeConditionsUnknownPending
	return b
}

func (b *PodBuilder) ResizeConditionsUnknownConditions() *PodBuilder {
	b.resizeConditions = PodResizeConditionsUnknownConditions
	return b
}

func (b *PodBuilder) QOSClass(qosClass v1.PodQOSClass) *PodBuilder {
	b.qosClass = qosClass
	return b
}

func (b *PodBuilder) QOSClassNotPresent() *PodBuilder {
	b.qosClass = ""
	return b
}

func (b *PodBuilder) AdditionalLabels(labels map[string]string) *PodBuilder {
	b.additionalLabels = labels
	return b
}

func (b *PodBuilder) AdditionalAnnotations(annotations map[string]string) *PodBuilder {
	b.additionalAnnotations = annotations
	return b
}

func (b *PodBuilder) NilContainerStatusStarted(nilContainerStatusStarted bool) *PodBuilder {
	b.nilContainerStatusStarted = nilContainerStatusStarted
	return b
}

func (b *PodBuilder) NilContainerStatusResources(nilContainerStatusResources bool) *PodBuilder {
	b.nilContainerStatusResources = nilContainerStatusResources
	return b
}

func (b *PodBuilder) ContainerCustomizerFunc(containerCustomizerFunc func(*ContainerBuilder)) *PodBuilder {
	b.containerCustomizerFunc = containerCustomizerFunc
	return b
}

func (b *PodBuilder) Build() *v1.Pod {
	p := b.pod()

	if b.nilContainerStatusStarted {
		p.Status.ContainerStatuses[0].Started = nil
	}

	if b.nilContainerStatusResources {
		p.Status.ContainerStatuses[0].Resources = nil
	}

	return p
}

func (b *PodBuilder) pod() *v1.Pod {
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
					State:   v1.ContainerState{Running: &v1.ContainerStateRunning{}},
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
			Conditions: b.resizeConditions,
			QOSClass:   b.qosClass,
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
