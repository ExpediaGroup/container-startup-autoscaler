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

package kubetest

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	PodAnnotationCpuStartup             = "3m"
	PodAnnotationCpuPostStartupRequests = "1m"
	PodAnnotationCpuPostStartupLimits   = "2m"

	PodAnnotationMemoryStartup             = "3M"
	PodAnnotationMemoryPostStartupRequests = "1M"
	PodAnnotationMemoryPostStartupLimits   = "2M"

	PodAnnotationCpuUnknown    = "999m"
	PodAnnotationMemoryUnknown = "999M"
)

const (
	DefaultPodNamespace                  = "namespace"
	DefaultPodName                       = "name"
	DefaultPodResourceVersion            = "1"
	DefaultLabelEnabledValue             = "true"
	DefaultAnnotationTargetContainerName = DefaultContainerName
	DefaultStatusContainerName           = DefaultContainerName
)

var (
	DefaultPodNamespacedName = types.NamespacedName{
		Namespace: DefaultPodNamespace,
		Name:      DefaultPodName,
	}
	DefaultPodStatusContainerState     = corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}
	DefaultContainerStatusResizeStatus = corev1.PodResizeStatus("")
)
