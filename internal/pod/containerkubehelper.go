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

package pod

import (
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ContainerKubeHelper performs operations relating to Kube containers.
type ContainerKubeHelper interface {
	Get(*v1.Pod, string) (*v1.Container, error)
	HasStartupProbe(container *v1.Container) bool
	HasReadinessProbe(container *v1.Container) bool
	State(*v1.Pod, string) (v1.ContainerState, error)
	IsStarted(*v1.Pod, string) (bool, error)
	IsReady(*v1.Pod, string) (bool, error)
	Requests(*v1.Container, v1.ResourceName) resource.Quantity
	Limits(*v1.Container, v1.ResourceName) resource.Quantity
	ResizePolicy(*v1.Container, v1.ResourceName) (v1.ResourceResizeRestartPolicy, error)
	AllocatedResources(*v1.Pod, string, v1.ResourceName) (resource.Quantity, error)
	CurrentRequests(*v1.Pod, string, v1.ResourceName) (resource.Quantity, error)
	CurrentLimits(*v1.Pod, string, v1.ResourceName) (resource.Quantity, error)
}

// containerKubeHelper is the default implementation of ContainerKubeHelper.
type containerKubeHelper struct{}

func newContainerKubeHelper() containerKubeHelper {
	return containerKubeHelper{}
}

// Get returns the container with the supplied containerName, from the supplied pod.
func (h containerKubeHelper) Get(pod *v1.Pod, containerName string) (*v1.Container, error) {
	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			return &container, nil
		}
	}

	return &v1.Container{}, errors.New("container not present")
}

// HasStartupProbe reports whether container has a startup probe.
func (h containerKubeHelper) HasStartupProbe(container *v1.Container) bool {
	return container.StartupProbe != nil
}

// HasReadinessProbe reports whether container has a readiness probe.
func (h containerKubeHelper) HasReadinessProbe(container *v1.Container) bool {
	return container.ReadinessProbe != nil
}

// State returns the container state of the container with the supplied containerName, from the supplied pod.
func (h containerKubeHelper) State(pod *v1.Pod, containerName string) (v1.ContainerState, error) {
	stat, err := h.status(pod, containerName)
	if err != nil {
		return v1.ContainerState{}, errors.Wrap(err, "unable to get container status")
	}

	return stat.State, nil
}

// IsStarted reports whether the container with the supplied containerName is started, from the supplied pod.
func (h containerKubeHelper) IsStarted(pod *v1.Pod, containerName string) (bool, error) {
	stat, err := h.status(pod, containerName)
	if err != nil {
		return false, errors.Wrap(err, "unable to get container status")
	}

	if stat.Started == nil {
		return false, nil
	}

	return *stat.Started, nil
}

// IsReady reports whether the container with the supplied containerName is ready, from the supplied pod.
func (h containerKubeHelper) IsReady(pod *v1.Pod, containerName string) (bool, error) {
	stat, err := h.status(pod, containerName)
	if err != nil {
		return false, errors.Wrap(err, "unable to get container status")
	}

	return stat.Ready, nil
}

// Requests returns requests for the supplied resourceName, from the supplied container.
func (h containerKubeHelper) Requests(container *v1.Container, resourceName v1.ResourceName) resource.Quantity {
	if container.Resources.Requests == nil {
		return resource.Quantity{}
	}

	switch resourceName {
	case v1.ResourceCPU:
		return *container.Resources.Requests.Cpu()
	case v1.ResourceMemory:
		return *container.Resources.Requests.Memory()
	}

	panic(errors.Errorf("resourceName '%s' not supported", resourceName))
}

// Limits returns limits for the supplied resourceName, from the supplied container.
func (h containerKubeHelper) Limits(container *v1.Container, resourceName v1.ResourceName) resource.Quantity {
	if container.Resources.Limits == nil {
		return resource.Quantity{}
	}

	switch resourceName {
	case v1.ResourceCPU:
		return *container.Resources.Limits.Cpu()
	case v1.ResourceMemory:
		return *container.Resources.Limits.Memory()
	}

	panic(errors.Errorf("resourceName '%s' not supported", resourceName))
}

// ResizePolicy returns the resource resize restart policy for the supplied resourceName, from the supplied container.
func (h containerKubeHelper) ResizePolicy(
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

	panic(errors.Errorf("resourceName '%s' not supported", resourceName))
}

// AllocatedResources returns allocated resources for the supplied containerName and resourceName, from the supplied
// pod.
func (h containerKubeHelper) AllocatedResources(
	pod *v1.Pod,
	containerName string,
	resourceName v1.ResourceName,
) (resource.Quantity, error) {
	stat, err := h.status(pod, containerName)
	if err != nil {
		return resource.Quantity{}, errors.Wrap(err, "unable to get container status")
	}

	if stat.AllocatedResources == nil {
		return resource.Quantity{}, NewContainerStatusAllocatedResourcesNotPresentError()
	}

	switch resourceName {
	case v1.ResourceCPU:
		return *stat.AllocatedResources.Cpu(), nil
	case v1.ResourceMemory:
		return *stat.AllocatedResources.Memory(), nil
	}

	panic(errors.Errorf("resourceName '%s' not supported", resourceName))
}

// CurrentRequests returns currently enacted requests for the supplied containerName and resourceName, from the
// supplied pod.
func (h containerKubeHelper) CurrentRequests(
	pod *v1.Pod,
	containerName string,
	resourceName v1.ResourceName,
) (resource.Quantity, error) {
	stat, err := h.status(pod, containerName)
	if err != nil {
		return resource.Quantity{}, errors.Wrap(err, "unable to get container status")
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

	panic(errors.Errorf("resourceName '%s' not supported", resourceName))
}

// CurrentLimits returns currently enacted limits for the supplied containerName and resourceName, from the supplied
// pod.
func (h containerKubeHelper) CurrentLimits(
	pod *v1.Pod,
	containerName string,
	resourceName v1.ResourceName,
) (resource.Quantity, error) {
	stat, err := h.status(pod, containerName)
	if err != nil {
		return resource.Quantity{}, errors.Wrap(err, "unable to get container status")
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

	panic(errors.Errorf("resourceName '%s' not supported", resourceName))
}

// status returns the container status for the supplied containerName, from the supplied pod.
func (h containerKubeHelper) status(pod *v1.Pod, containerName string) (*v1.ContainerStatus, error) {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == containerName {
			return &containerStatus, nil
		}
	}

	return &v1.ContainerStatus{}, NewContainerStatusNotPresentError()
}
