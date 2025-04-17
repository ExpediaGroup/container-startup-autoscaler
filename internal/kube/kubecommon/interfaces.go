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

package kubecommon

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
)

// PodHelper performs operations relating to Kube pods.
type PodHelper interface {
	Get(
		ctx context.Context,
		name types.NamespacedName,
	) (bool, *v1.Pod, error)

	Patch(
		ctx context.Context,
		pod *v1.Pod,
		podMutationFuncs []func(*v1.Pod) error,
		patchResize bool,
		mustSyncCache bool,
	) (*v1.Pod, error)

	HasAnnotation(
		pod *v1.Pod,
		name string,
	) (bool, string)

	ExpectedLabelValueAs(
		pod *v1.Pod,
		name string,
		as DataType,
	) (any, error)

	ExpectedAnnotationValueAs(
		pod *v1.Pod,
		name string,
		as DataType,
	) (any, error)

	IsContainerInSpec(
		pod *v1.Pod,
		containerName string,
	) bool

	ResizeConditions(
		pod *v1.Pod,
	) []v1.PodCondition
}

// ContainerHelper performs operations relating to Kube containers.
type ContainerHelper interface {
	Get(
		pod *v1.Pod,
		containerName string,
	) (*v1.Container, error)

	HasStartupProbe(
		container *v1.Container,
	) bool

	HasReadinessProbe(
		container *v1.Container,
	) bool

	State(
		pod *v1.Pod,
		container *v1.Container,
	) (v1.ContainerState, error)

	IsStarted(
		pod *v1.Pod,
		container *v1.Container,
	) (bool, error)

	IsReady(
		pod *v1.Pod,
		container *v1.Container,
	) (bool, error)

	Requests(
		container *v1.Container,
		resourceName v1.ResourceName,
	) resource.Quantity

	Limits(
		container *v1.Container,
		resourceName v1.ResourceName,
	) resource.Quantity

	ResizePolicy(
		container *v1.Container,
		resourceName v1.ResourceName,
	) (v1.ResourceResizeRestartPolicy, error)

	CurrentRequests(
		pod *v1.Pod,
		container *v1.Container,
		resourceName v1.ResourceName,
	) (resource.Quantity, error)

	CurrentLimits(
		pod *v1.Pod,
		container *v1.Container,
		resourceName v1.ResourceName,
	) (resource.Quantity, error)
}
