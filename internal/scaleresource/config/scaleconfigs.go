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

package config

import (
	"errors"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource/scaleresourcecommon"
	"k8s.io/api/core/v1"
)

type ScaleConfigs interface {
	TargetContainerName(*v1.Pod) (string, error)
	StoreFromAnnotationsAll(*v1.Pod) error
	ValidateAll(*v1.Container) error
	ValidateCollection() error

	ScaleConfigFor(scaleresource.ResourceType) ScaleConfig
	AllScaleConfigs() []ScaleConfig
	AllEnabledScaleConfigs() []ScaleConfig
	AllEnabledScaleConfigsTypes() []scaleresource.ResourceType

	String() string
}

func NewScaleConfigs(podHelper kube.PodHelper, containerHelper kube.ContainerHelper) ScaleConfigs {
	return &scaleConfigs{
		cpuScaleConfig:    NewCpuScaleConfig(true, podHelper, containerHelper),
		memoryScaleConfig: NewMemoryScaleConfig(true, podHelper, containerHelper),
		podHelper:         podHelper,
	}
}

type scaleConfigs struct {
	cpuScaleConfig    ScaleConfig
	memoryScaleConfig ScaleConfig

	podHelper kube.PodHelper
}

func (s *scaleConfigs) TargetContainerName(pod *v1.Pod) (string, error) {
	value, err := s.podHelper.ExpectedAnnotationValueAs(
		pod,
		scaleresourcecommon.AnnotationTargetContainerName,
		kubecommon.DataTypeString,
	)
	if err != nil {
		return "", common.WrapErrorf(err, "unable to get '%s' annotation value", scaleresourcecommon.AnnotationTargetContainerName)
	}

	return value.(string), nil
}

func (s *scaleConfigs) StoreFromAnnotationsAll(pod *v1.Pod) error {
	for _, scaleConfig := range s.AllScaleConfigs() {
		if err := scaleConfig.StoreFromAnnotations(pod); err != nil {
			return err
		}
	}

	return nil
}

func (s *scaleConfigs) ValidateAll(container *v1.Container) error {
	for _, scaleConfig := range s.AllScaleConfigs() {
		if err := scaleConfig.Validate(container); err != nil {
			return err
		}
	}

	return nil
}

func (s *scaleConfigs) ValidateCollection() error {
	atLeastOneEnabled := false

	for _, scaleConfig := range s.AllScaleConfigs() {
		if scaleConfig.IsEnabled() {
			atLeastOneEnabled = true
			break
		}
	}

	if !atLeastOneEnabled {
		return errors.New("no resources are configured for scaling")
	}

	return nil
}

func (s *scaleConfigs) ScaleConfigFor(resourceType scaleresource.ResourceType) ScaleConfig {
	switch resourceType {
	case scaleresource.ResourceTypeCpu:
		return s.cpuScaleConfig
	case scaleresource.ResourceTypeMemory:
		return s.memoryScaleConfig
	default:
		return nil
	}
}

func (s *scaleConfigs) AllScaleConfigs() []ScaleConfig {
	return []ScaleConfig{s.cpuScaleConfig, s.memoryScaleConfig}
}

func (s *scaleConfigs) AllEnabledScaleConfigs() []ScaleConfig {
	var enabledConfigs []ScaleConfig

	for _, scaleConfig := range s.AllScaleConfigs() {
		if scaleConfig.IsEnabled() {
			enabledConfigs = append(enabledConfigs, scaleConfig)
		}
	}

	return enabledConfigs
}

func (s *scaleConfigs) AllEnabledScaleConfigsTypes() []scaleresource.ResourceType {
	var enabledConfigsTypes []scaleresource.ResourceType

	for _, scaleConfig := range s.AllScaleConfigs() {
		if scaleConfig.IsEnabled() {
			enabledConfigsTypes = append(enabledConfigsTypes, scaleConfig.ResourceType())
		}
	}

	return enabledConfigsTypes
}

func (s *scaleConfigs) String() string {
	var result string
	allScaleConfigs := s.AllScaleConfigs()

	for i, scaleConfig := range allScaleConfigs {
		result += scaleConfig.String()
		if i < len(allScaleConfigs)-1 {
			result += ", "
		}
	}

	return result
}
