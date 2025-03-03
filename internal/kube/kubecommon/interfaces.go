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

package kubecommon

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
)

// PodHelper performs operations relating to Kube pods.
type PodHelper interface {
	Get(context.Context, types.NamespacedName) (bool, *v1.Pod, error)
	Patch(context.Context, *v1.Pod, []func(*v1.Pod) error, bool, bool) (*v1.Pod, error)
	HasAnnotation(pod *v1.Pod, name string) (bool, string)
	ExpectedLabelValueAs(*v1.Pod, string, DataType) (any, error)
	ExpectedAnnotationValueAs(*v1.Pod, string, DataType) (any, error)
	IsContainerInSpec(*v1.Pod, string) bool
	ResizeStatus(*v1.Pod) v1.PodResizeStatus
}

// ContainerHelper performs operations relating to Kube containers.
type ContainerHelper interface {
	Get(*v1.Pod, string) (*v1.Container, error)
	HasStartupProbe(*v1.Container) bool
	HasReadinessProbe(*v1.Container) bool
	State(*v1.Pod, *v1.Container) (v1.ContainerState, error)
	IsStarted(*v1.Pod, *v1.Container) (bool, error)
	IsReady(*v1.Pod, *v1.Container) (bool, error)
	Requests(*v1.Container, v1.ResourceName) resource.Quantity
	Limits(*v1.Container, v1.ResourceName) resource.Quantity
	ResizePolicy(*v1.Container, v1.ResourceName) (v1.ResourceResizeRestartPolicy, error)
	CurrentRequests(*v1.Pod, *v1.Container, v1.ResourceName) (resource.Quantity, error)
	CurrentLimits(*v1.Pod, *v1.Container, v1.ResourceName) (resource.Quantity, error)
}
