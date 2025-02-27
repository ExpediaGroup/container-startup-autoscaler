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
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"k8s.io/api/core/v1"
)

type State interface {
	ResourceName() v1.ResourceName
	IsStartupConfigApplied(*v1.Container) bool
	IsPostStartupConfigApplied(*v1.Container) bool
	DoesRequestsCurrentMatchSpec(*v1.Pod, *v1.Container) (bool, error)
	DoesLimitsCurrentMatchSpec(*v1.Pod, *v1.Container) (bool, error)
}

type state struct {
	resourceName    v1.ResourceName
	config          Config
	containerHelper kube.ContainerHelper
}

func NewState(
	resourceName v1.ResourceName,
	config Config,
	containerHelper kube.ContainerHelper,
) *state {
	return &state{
		resourceName:    resourceName,
		config:          config,
		containerHelper: containerHelper,
	}
}

func (s *state) ResourceName() v1.ResourceName {
	return s.resourceName
}

func (s *state) IsStartupConfigApplied(container *v1.Container) bool {
	if !s.config.IsEnabled() {
		return true
	}

	startupRequestsApplied := s.containerHelper.Requests(container, s.resourceName).Equal(s.config.Resources().Startup)
	startupLimitsApplied := s.containerHelper.Limits(container, s.resourceName).Equal(s.config.Resources().Startup)
	return startupRequestsApplied && startupLimitsApplied
}

func (s *state) IsPostStartupConfigApplied(container *v1.Container) bool {
	if !s.config.IsEnabled() {
		return true
	}

	postStartupRequestsApplied := s.containerHelper.Requests(container, s.resourceName).Equal(s.config.Resources().PostStartupRequests)
	postStartupLimitsApplied := s.containerHelper.Limits(container, s.resourceName).Equal(s.config.Resources().PostStartupLimits)
	return postStartupRequestsApplied && postStartupLimitsApplied
}

func (s *state) DoesRequestsCurrentMatchSpec(pod *v1.Pod, container *v1.Container) (bool, error) {
	if !s.config.IsEnabled() {
		return true, nil
	}

	currentRequests, err := s.containerHelper.CurrentRequests(pod, container, s.resourceName)
	if err != nil {
		return false, common.WrapErrorf(err, "unable to get status resources %s requests", s.resourceName)
	}

	requests := s.containerHelper.Requests(container, s.resourceName)
	return currentRequests.Equal(requests), nil
}

func (s *state) DoesLimitsCurrentMatchSpec(pod *v1.Pod, container *v1.Container) (bool, error) {
	if !s.config.IsEnabled() {
		return true, nil
	}

	currentLimits, err := s.containerHelper.CurrentLimits(pod, container, s.resourceName)
	if err != nil {
		return false, common.WrapErrorf(err, "unable to get status resources %s limits", s.resourceName)
	}

	limits := s.containerHelper.Limits(container, s.resourceName)
	return currentLimits.Equal(limits), nil
}
