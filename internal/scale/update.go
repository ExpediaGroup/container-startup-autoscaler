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

package scale

import (
	"errors"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// update is the default implementation of scalecommon.Update.
type update struct {
	resourceName v1.ResourceName
	config       scalecommon.Configuration
}

func NewUpdate(resourceName v1.ResourceName, config scalecommon.Configuration) scalecommon.Update {
	return &update{
		resourceName: resourceName,
		config:       config,
	}
}

// ResourceName returns the resource name that the update relates to.
func (u *update) ResourceName() v1.ResourceName {
	return u.resourceName
}

// StartupPodMutationFunc returns a function that mutates a pod to apply startup resources for the resource.
func (u *update) StartupPodMutationFunc(container *v1.Container) func(pod *v1.Pod) error {
	if !u.config.IsEnabled() {
		return func(pod *v1.Pod) error {
			return nil
		}
	}

	return func(pod *v1.Pod) error {
		err := u.setResources(
			pod,
			container,
			u.config.Resources().Startup,
			u.config.Resources().Startup,
		)
		if err != nil {
			return common.WrapErrorf(err, "unable to set %s startup resources", u.resourceName)
		}

		return nil
	}
}

// PostStartupPodMutationFunc returns a function that mutates a pod to apply post-startup resources for the resource.
func (u *update) PostStartupPodMutationFunc(container *v1.Container) func(pod *v1.Pod) error {
	if !u.config.IsEnabled() {
		return func(pod *v1.Pod) error {
			return nil
		}
	}

	return func(pod *v1.Pod) error {
		err := u.setResources(
			pod,
			container,
			u.config.Resources().PostStartupRequests,
			u.config.Resources().PostStartupLimits,
		)
		if err != nil {
			return common.WrapErrorf(err, "unable to set %s post-startup resources", u.resourceName)
		}

		return nil
	}
}

// setResources sets resources within the supplied pod.
func (u *update) setResources(
	pod *v1.Pod,
	container *v1.Container,
	requests resource.Quantity,
	limits resource.Quantity,
) error {
	var containerToMutate *v1.Container
	for _, ctr := range pod.Spec.Containers {
		if ctr.Name == container.Name {
			containerToMutate = &ctr
			break
		}
	}
	if containerToMutate == nil {
		return errors.New("container not present")
	}

	containerToMutate.Resources.Requests[u.resourceName] = requests
	containerToMutate.Resources.Limits[u.resourceName] = limits

	return nil
}
