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
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/controller/controllercommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Pod is a facade (and the only package-external entry point) for pod interaction and contains a number of services.
// It only exposes exported services methods via their corresponding interfaces.
type Pod struct {
	Configuration         podcommon.Configuration
	Validation            podcommon.Validation
	TargetContainerState  podcommon.TargetContainerState
	TargetContainerAction podcommon.TargetContainerAction
	Status                podcommon.Status
	PodHelper             kubecommon.PodHelper
	ContainerHelper       kubecommon.ContainerHelper
}

func NewPod(
	controllerConfig controllercommon.ControllerConfig,
	client client.Client,
	recorder record.EventRecorder,
) *Pod {
	podHelper := kube.NewPodHelper(client)
	containerHelper := kube.NewContainerHelper()
	stat := newStatus(podHelper)

	return &Pod{
		Configuration:         newConfiguration(podHelper, containerHelper),
		Validation:            newValidation(recorder, stat, podHelper, containerHelper),
		TargetContainerState:  newTargetContainerState(containerHelper),
		TargetContainerAction: newTargetContainerAction(controllerConfig, recorder, stat, podHelper),
		Status:                stat,
		PodHelper:             podHelper,
		ContainerHelper:       containerHelper,
	}
}
