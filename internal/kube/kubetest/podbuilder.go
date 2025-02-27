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
	"k8s.io/api/core/v1"
)

// podBuilder builds a test pod.
type podBuilder struct {
	config                      podConfig
	additionalLabels            map[string]string
	additionalAnnotations       map[string]string
	containerStatusState        *v1.ContainerState
	containerStatusResizeStatus *v1.PodResizeStatus
	nilContainerStatusStarted   bool
	nilContainerStatusResources bool
}

func NewPodBuilder(config podConfig) *podBuilder {
	return &podBuilder{config: config}
}

func (b *podBuilder) AdditionalLabels(labels map[string]string) *podBuilder {
	b.additionalLabels = labels
	return b
}

func (b *podBuilder) AdditionalAnnotations(annotations map[string]string) *podBuilder {
	b.additionalAnnotations = annotations
	return b
}

func (b *podBuilder) ContainerStatusState(state v1.ContainerState) *podBuilder {
	b.containerStatusState = &state
	return b
}

func (b *podBuilder) ContainerStatusResizeStatus(resizeStatus v1.PodResizeStatus) *podBuilder {
	b.containerStatusResizeStatus = &resizeStatus
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
	p := pod(b.config)

	for name, value := range b.additionalLabels {
		p.Labels[name] = value
	}

	for name, value := range b.additionalAnnotations {
		p.Annotations[name] = value
	}

	if b.containerStatusState != nil {
		p.Status.ContainerStatuses[0].State = *b.containerStatusState
	}

	if b.containerStatusResizeStatus != nil {
		p.Status.Resize = *b.containerStatusResizeStatus
	}

	if b.nilContainerStatusStarted {
		p.Status.ContainerStatuses[0].Started = nil
	}

	if b.nilContainerStatusResources {
		p.Status.ContainerStatuses[0].Resources = nil
	}

	return p
}
