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

package update

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource/config"
	v1 "k8s.io/api/core/v1"
)

type ScaleUpdates interface {
	SetStartupResourcesAll(*v1.Pod, *v1.Container) (*v1.Pod, error)
	SetPostStartupResourcesAll(*v1.Pod, *v1.Container) (*v1.Pod, error)

	ScaleUpdateFor(scaleresource.ResourceType) ScaleUpdate
	AllScaleUpdates() []ScaleUpdate
}

func NewScaleUpdates(scaleConfigs config.ScaleConfigs) ScaleUpdates {
	return &scaleUpdates{
		cpuScaleUpdate:    NewCpuScaleUpdate(scaleConfigs.ScaleConfigFor(scaleresource.ResourceTypeCpu)),
		memoryScaleUpdate: NewMemoryScaleUpdate(scaleConfigs.ScaleConfigFor(scaleresource.ResourceTypeMemory)),
	}
}

type scaleUpdates struct {
	cpuScaleUpdate    ScaleUpdate
	memoryScaleUpdate ScaleUpdate
}

func (s *scaleUpdates) SetStartupResourcesAll(pod *v1.Pod, container *v1.Container) (*v1.Pod, error) {
	clonedPod := pod.DeepCopy()

	for _, scaleUpdate := range s.AllScaleUpdates() {
		_, err := scaleUpdate.SetStartupResources(clonedPod, container, false)
		if err != nil {
			return nil, common.WrapErrorf(err, "unable to set %s startup resources", scaleUpdate.ResourceType())
		}
	}

	return clonedPod, nil
}

func (s *scaleUpdates) SetPostStartupResourcesAll(pod *v1.Pod, container *v1.Container) (*v1.Pod, error) {
	clonedPod := pod.DeepCopy()

	for _, scaleUpdate := range s.AllScaleUpdates() {
		_, err := scaleUpdate.SetPostStartupResources(clonedPod, container, false)
		if err != nil {
			return nil, common.WrapErrorf(err, "unable to set %s post-startup resources", scaleUpdate.ResourceType())
		}
	}

	return clonedPod, nil
}

func (s *scaleUpdates) ScaleUpdateFor(resourceType scaleresource.ResourceType) ScaleUpdate {
	switch resourceType {
	case scaleresource.ResourceTypeCpu:
		return s.cpuScaleUpdate
	case scaleresource.ResourceTypeMemory:
		return s.memoryScaleUpdate
	default:
		return nil
	}
}

func (s *scaleUpdates) AllScaleUpdates() []ScaleUpdate {
	return []ScaleUpdate{s.cpuScaleUpdate, s.memoryScaleUpdate}
}
