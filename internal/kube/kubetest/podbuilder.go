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

// podBuilder builds a test pod.
type podBuilder struct {
	enabledResources            []v1.ResourceName
	resourcesState              podcommon.StateResources
	stateStarted                podcommon.StateBool
	stateReady                  podcommon.StateBool
	containerStatusState        v1.ContainerState
	containerStatusResizeStatus v1.PodResizeStatus
	additionalLabels            map[string]string
	additionalAnnotations       map[string]string
	nilContainerStatusStarted   bool
	nilContainerStatusResources bool
}

func NewPodBuilder() *podBuilder {
	b := &podBuilder{}
	b.enabledResources = []v1.ResourceName{v1.ResourceCPU, v1.ResourceMemory}
	b.resourcesState = podcommon.StateResourcesStartup
	b.stateStarted = podcommon.StateBoolFalse
	b.stateReady = podcommon.StateBoolFalse
	b.containerStatusState = v1.ContainerState{Running: &v1.ContainerStateRunning{}}
	b.containerStatusResizeStatus = ""
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

func (b *podBuilder) ContainerStatusStateWaiting() *podBuilder {
	b.containerStatusState = v1.ContainerState{Waiting: &v1.ContainerStateWaiting{}}
	return b
}

func (b *podBuilder) ContainerStatusStateTerminated() *podBuilder {
	b.containerStatusState = v1.ContainerState{Terminated: &v1.ContainerStateTerminated{}}
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

func (b *podBuilder) Build() *v1.Pod {
	// TODO(wt) move everything from pod.go and remove pod.go

	p := pod(newPodConfig(b.enabledResources, b.resourcesState, b.stateStarted, b.stateReady))

	p.Status.ContainerStatuses[0].State = b.containerStatusState
	p.Status.Resize = b.containerStatusResizeStatus

	for name, value := range b.additionalLabels {
		p.Labels[name] = value
	}

	for name, value := range b.additionalAnnotations {
		p.Annotations[name] = value
	}

	if b.nilContainerStatusStarted {
		p.Status.ContainerStatuses[0].Started = nil
	}

	if b.nilContainerStatusResources {
		p.Status.ContainerStatuses[0].Resources = nil
	}

	return p
}
