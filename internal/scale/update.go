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

package scale

import (
	"errors"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type Update interface {
	ResourceName() v1.ResourceName
	SetStartupResources(*v1.Pod, *v1.Container, bool) (*v1.Pod, error)
	SetPostStartupResources(*v1.Pod, *v1.Container, bool) (*v1.Pod, error)
}

type update struct {
	resourceName v1.ResourceName
	config       Config
}

func NewUpdate(resourceName v1.ResourceName, config Config) *update {
	return &update{
		resourceName: resourceName,
		config:       config,
	}
}

func (u *update) ResourceName() v1.ResourceName {
	return u.resourceName
}

func (u *update) SetStartupResources(
	pod *v1.Pod,
	container *v1.Container,
	clonePod bool,
) (*v1.Pod, error) {
	if !u.config.IsEnabled() {
		return pod, nil
	}

	newPod, err := u.setResources(
		pod,
		container,
		u.config.Resources().Startup,
		u.config.Resources().Startup,
		clonePod,
	)
	if err != nil {
		return nil, common.WrapErrorf(err, "unable to set %s startup resources", u.resourceName)
	}

	return newPod, nil
}

func (u *update) SetPostStartupResources(
	pod *v1.Pod,
	container *v1.Container,
	clonePod bool,
) (*v1.Pod, error) {
	if !u.config.IsEnabled() {
		return pod, nil
	}

	newPod, err := u.setResources(
		pod,
		container,
		u.config.Resources().PostStartupRequests,
		u.config.Resources().PostStartupLimits,
		clonePod,
	)
	if err != nil {
		return nil, common.WrapErrorf(err, "unable to set %s post-startup resources", u.resourceName)
	}

	return newPod, nil
}

func (u *update) setResources(
	pod *v1.Pod,
	container *v1.Container,
	requests resource.Quantity,
	limits resource.Quantity,
	clonePod bool,
) (*v1.Pod, error) {
	var podToMutate *v1.Pod
	if clonePod {
		podToMutate = pod.DeepCopy()
	} else {
		podToMutate = pod
	}

	var containerToMutate *v1.Container
	for _, ctr := range podToMutate.Spec.Containers {
		if ctr.Name == container.Name {
			containerToMutate = &ctr
			break
		}
	}
	if containerToMutate == nil {
		return nil, errors.New("container not present")
	}

	containerToMutate.Resources.Requests[u.resourceName] = requests
	containerToMutate.Resources.Limits[u.resourceName] = limits

	return podToMutate, nil
}
