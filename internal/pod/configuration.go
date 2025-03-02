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
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale"
	v1 "k8s.io/api/core/v1"
)

// Configuration performs operations relating to configuration.
type Configuration interface {
	Configure(*v1.Pod) (scale.Configs, error)
}

// configuration is the default implementation of Configuration.
type configuration struct {
	podHelper       kube.PodHelper
	containerHelper kube.ContainerHelper
}

func newConfiguration(
	podHelper kube.PodHelper,
	containerHelper kube.ContainerHelper,
) *configuration {
	return &configuration{
		podHelper:       podHelper,
		containerHelper: containerHelper,
	}
}

// Configure performs configuration tasks using the supplied pod.
func (c *configuration) Configure(pod *v1.Pod) (scale.Configs, error) {
	configs := scale.NewConfigs(c.podHelper, c.containerHelper)

	if err := configs.StoreFromAnnotationsAll(pod); err != nil {
		return nil, common.WrapErrorf(err, "unable to store configuration from annotations")
	}

	return configs, nil
}
