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

package scalecommon

import "github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"

const (
	AnnotationTargetContainerName = podcommon.Namespace + "/target-container-name"

	AnnotationCpuStartup             = podcommon.Namespace + "/cpu-startup"
	AnnotationCpuPostStartupRequests = podcommon.Namespace + "/cpu-post-startup-requests"
	AnnotationCpuPostStartupLimits   = podcommon.Namespace + "/cpu-post-startup-limits"

	AnnotationMemoryStartup             = podcommon.Namespace + "/memory-startup"
	AnnotationMemoryPostStartupRequests = podcommon.Namespace + "/memory-post-startup-requests"
	AnnotationMemoryPostStartupLimits   = podcommon.Namespace + "/memory-post-startup-limits"
)
