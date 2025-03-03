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

package kube

import (
	"errors"
	"fmt"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// containerHelper is the default implementation of ContainerHelper.
type containerHelper struct{}

func NewContainerHelper() kubecommon.ContainerHelper {
	return containerHelper{}
}

// Get returns the container with the supplied containerName, from the supplied pod.
func (h containerHelper) Get(pod *v1.Pod, containerName string) (*v1.Container, error) {
	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			return &container, nil
		}
	}

	return &v1.Container{}, errors.New("container not present")
}

// HasStartupProbe reports whether container has a startup probe.
func (h containerHelper) HasStartupProbe(container *v1.Container) bool {
	return container.StartupProbe != nil
}

// HasReadinessProbe reports whether container has a readiness probe.
func (h containerHelper) HasReadinessProbe(container *v1.Container) bool {
	return container.ReadinessProbe != nil
}

// State returns the state of the container.
func (h containerHelper) State(pod *v1.Pod, container *v1.Container) (v1.ContainerState, error) {
	stat, err := h.status(pod, container)
	if err != nil {
		return v1.ContainerState{}, common.WrapErrorf(err, "unable to get container status")
	}

	return stat.State, nil
}

// IsStarted reports whether the container is started.
func (h containerHelper) IsStarted(pod *v1.Pod, container *v1.Container) (bool, error) {
	stat, err := h.status(pod, container)
	if err != nil {
		return false, common.WrapErrorf(err, "unable to get container status")
	}

	if stat.Started == nil {
		return false, nil
	}

	return *stat.Started, nil
}

// IsReady reports whether the container is ready.
func (h containerHelper) IsReady(pod *v1.Pod, container *v1.Container) (bool, error) {
	stat, err := h.status(pod, container)
	if err != nil {
		return false, common.WrapErrorf(err, "unable to get container status")
	}

	return stat.Ready, nil
}

// Requests returns requests for the supplied resourceName, from the supplied container.
func (h containerHelper) Requests(container *v1.Container, resourceName v1.ResourceName) resource.Quantity {
	if container.Resources.Requests == nil {
		return resource.Quantity{}
	}

	switch resourceName {
	case v1.ResourceCPU:
		return *container.Resources.Requests.Cpu()
	case v1.ResourceMemory:
		return *container.Resources.Requests.Memory()
	}

	panic(fmt.Errorf("resourceName '%s' not supported", resourceName))
}

// Limits returns limits for the supplied resourceName, from the supplied container.
func (h containerHelper) Limits(container *v1.Container, resourceName v1.ResourceName) resource.Quantity {
	if container.Resources.Limits == nil {
		return resource.Quantity{}
	}

	switch resourceName {
	case v1.ResourceCPU:
		return *container.Resources.Limits.Cpu()
	case v1.ResourceMemory:
		return *container.Resources.Limits.Memory()
	}

	panic(fmt.Errorf("resourceName '%s' not supported", resourceName))
}

// ResizePolicy returns the resource resize restart policy for the supplied resourceName, from the supplied container.
func (h containerHelper) ResizePolicy(
	container *v1.Container,
	resourceName v1.ResourceName,
) (v1.ResourceResizeRestartPolicy, error) {
	if container.ResizePolicy == nil {
		return "", errors.New("container resize policy is null - check kube version and feature gate configuration")
	}

	for _, resizePolicy := range container.ResizePolicy {
		if resizePolicy.ResourceName == resourceName {
			return resizePolicy.RestartPolicy, nil
		}
	}

	panic(fmt.Errorf("resourceName '%s' not supported", resourceName))
}

// CurrentRequests returns currently enacted requests for the supplied container and resourceName.
func (h containerHelper) CurrentRequests(
	pod *v1.Pod,
	container *v1.Container,
	resourceName v1.ResourceName,
) (resource.Quantity, error) {
	stat, err := h.status(pod, container)
	if err != nil {
		return resource.Quantity{}, common.WrapErrorf(err, "unable to get container status")
	}

	if stat.Resources == nil {
		return resource.Quantity{}, NewContainerStatusResourcesNotPresentError()
	}

	switch resourceName {
	case v1.ResourceCPU:
		return *stat.Resources.Requests.Cpu(), nil
	case v1.ResourceMemory:
		return *stat.Resources.Requests.Memory(), nil
	}

	panic(fmt.Errorf("resourceName '%s' not supported", resourceName))
}

// CurrentLimits returns currently enacted limits for the supplied container and resourceName, from the supplied
// pod.
func (h containerHelper) CurrentLimits(
	pod *v1.Pod,
	container *v1.Container,
	resourceName v1.ResourceName,
) (resource.Quantity, error) {
	stat, err := h.status(pod, container)
	if err != nil {
		return resource.Quantity{}, common.WrapErrorf(err, "unable to get container status")
	}

	if stat.Resources == nil {
		return resource.Quantity{}, NewContainerStatusResourcesNotPresentError()
	}

	switch resourceName {
	case v1.ResourceCPU:
		return *stat.Resources.Limits.Cpu(), nil
	case v1.ResourceMemory:
		return *stat.Resources.Limits.Memory(), nil
	}

	panic(fmt.Errorf("resourceName '%s' not supported", resourceName))
}

// status returns the container status for the supplied container.
func (h containerHelper) status(pod *v1.Pod, container *v1.Container) (*v1.ContainerStatus, error) {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == container.Name {
			return &containerStatus, nil
		}
	}

	return &v1.ContainerStatus{}, NewContainerStatusNotPresentError()
}
