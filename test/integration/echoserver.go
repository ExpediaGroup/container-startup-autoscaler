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

package integration

import (
	"strconv"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Deployment-----------------------------------------------------------------------------------------------------------

func echoDeploymentConfigStandardStartup(
	namespace string,
	replicas int32,
	annotations csaQuantityAnnotations,
) deploymentConfig {
	var cpuRequests, cpuLimits = annotations.CpuStartupRequestsLimits()
	var memoryRequests, memoryLimits = annotations.MemoryStartupRequestsLimits()

	return echoDeploymentConfigStandard(
		namespace,
		replicas,
		annotations,
		cpuRequests, cpuLimits,
		memoryRequests, memoryLimits,
		echoServerDefaultProbeInitialDelaySeconds,
	)
}

func echoDeploymentConfigStandardPostStartup(
	namespace string,
	replicas int32,
	annotations csaQuantityAnnotations,
) deploymentConfig {
	var cpuRequests, cpuLimits = annotations.CpuPostStartupRequestsLimits()
	var memoryRequests, memoryLimits = annotations.MemoryPostStartupRequestsLimits()

	return echoDeploymentConfigStandard(
		namespace,
		replicas,
		annotations,
		cpuRequests, cpuLimits,
		memoryRequests, memoryLimits,
		echoServerDefaultProbeInitialDelaySeconds,
	)
}

func echoDeploymentConfigStandard(
	namespace string,
	replicas int32,
	annotations csaQuantityAnnotations,
	cpuRequests string, cpuLimits string,
	memoryRequests string, memoryLimits string,
	probesInitialDelaySeconds int32,
) deploymentConfig {
	return deploymentConfig{
		namespace:   namespace,
		name:        echoServerName,
		replicas:    replicas,
		matchLabels: echoMatchLabels(),
		podConfig: podConfig{
			labels:      echoPodLabels(),
			annotations: echoPodAnnotations(annotations),
		},
		containerConfigs: echoContainerConfigs(
			cpuRequests, cpuLimits,
			memoryRequests, memoryLimits,
			probesInitialDelaySeconds,
		),
	}
}

// StatefulSet ---------------------------------------------------------------------------------------------------------

func echoStatefulSetConfigStandardStartup(
	namespace string,
	replicas int32,
	annotations csaQuantityAnnotations,
) statefulSetConfig {
	var cpuRequests, cpuLimits = annotations.CpuStartupRequestsLimits()
	var memoryRequests, memoryLimits = annotations.MemoryStartupRequestsLimits()

	return echoStatefulSetConfigStandard(
		namespace,
		replicas,
		annotations,
		cpuRequests, cpuLimits,
		memoryRequests, memoryLimits,
		echoServerDefaultProbeInitialDelaySeconds,
	)
}

func echoStatefulSetConfigStandardPostStartup(
	namespace string,
	replicas int32,
	annotations csaQuantityAnnotations,
) statefulSetConfig {
	var cpuRequests, cpuLimits = annotations.CpuPostStartupRequestsLimits()
	var memoryRequests, memoryLimits = annotations.MemoryPostStartupRequestsLimits()

	return echoStatefulSetConfigStandard(
		namespace,
		replicas,
		annotations,
		cpuRequests, cpuLimits,
		memoryRequests, memoryLimits,
		echoServerDefaultProbeInitialDelaySeconds,
	)
}

func echoStatefulSetConfigStandard(
	namespace string,
	replicas int32,
	annotations csaQuantityAnnotations,
	cpuRequests string, cpuLimits string,
	memoryRequests string, memoryLimits string,
	probesInitialDelaySeconds int32,
) statefulSetConfig {
	return statefulSetConfig{
		namespace:   namespace,
		name:        echoServerName,
		replicas:    replicas,
		matchLabels: echoMatchLabels(),
		podConfig: podConfig{
			labels:      echoPodLabels(),
			annotations: echoPodAnnotations(annotations),
		},
		containerConfigs: echoContainerConfigs(
			cpuRequests, cpuLimits,
			memoryRequests, memoryLimits,
			probesInitialDelaySeconds,
		),
	}
}

// DaemonSet -----------------------------------------------------------------------------------------------------------

func echoDaemonSetConfigStandardStartup(
	namespace string,
	annotations csaQuantityAnnotations,
) daemonSetConfig {
	var cpuRequests, cpuLimits = annotations.CpuStartupRequestsLimits()
	var memoryRequests, memoryLimits = annotations.MemoryStartupRequestsLimits()

	return echoDaemonSetConfigStandard(
		namespace,
		annotations,
		cpuRequests, cpuLimits,
		memoryRequests, memoryLimits,
		echoServerDefaultProbeInitialDelaySeconds,
	)
}

func echoDaemonSetConfigStandardPostStartup(
	namespace string,
	annotations csaQuantityAnnotations,
) daemonSetConfig {
	var cpuRequests, cpuLimits = annotations.CpuPostStartupRequestsLimits()
	var memoryRequests, memoryLimits = annotations.MemoryPostStartupRequestsLimits()

	return echoDaemonSetConfigStandard(
		namespace,
		annotations,
		cpuRequests, cpuLimits,
		memoryRequests, memoryLimits,
		echoServerDefaultProbeInitialDelaySeconds,
	)
}

func echoDaemonSetConfigStandard(
	namespace string,
	annotations csaQuantityAnnotations,
	cpuRequests string, cpuLimits string,
	memoryRequests string, memoryLimits string,
	probesInitialDelaySeconds int32,
) daemonSetConfig {
	return daemonSetConfig{
		namespace:   namespace,
		name:        echoServerName,
		matchLabels: echoMatchLabels(),
		podConfig: podConfig{
			labels:      echoPodLabels(),
			annotations: echoPodAnnotations(annotations),
		},
		containerConfigs: echoContainerConfigs(
			cpuRequests, cpuLimits,
			memoryRequests, memoryLimits,
			probesInitialDelaySeconds,
		),
	}
}

// Container -----------------------------------------------------------------------------------------------------------

func echoContainerConfigStandard(
	name string, port int32,
	cpuRequests string, cpuLimits string,
	memoryRequests string, memoryLimits string,
	probesInitialDelaySeconds int32,
) containerConfig {
	return containerConfig{
		name:          name,
		image:         echoServerDockerImageTag,
		containerPort: port,
		env: []v1.EnvVar{
			{
				Name:  "PORT",
				Value: strconv.Itoa(int(port)),
			},
		},
		resizePolicy: []v1.ContainerResizePolicy{
			{
				ResourceName:  v1.ResourceCPU,
				RestartPolicy: v1.NotRequired,
			},
			{
				ResourceName:  v1.ResourceMemory,
				RestartPolicy: v1.NotRequired,
			},
		},
		resources: v1.ResourceRequirements{
			Requests: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    resource.MustParse(cpuRequests),
				v1.ResourceMemory: resource.MustParse(memoryRequests),
			},
			Limits: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:    resource.MustParse(cpuLimits),
				v1.ResourceMemory: resource.MustParse(memoryLimits),
			},
		},
		startupProbe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/?echo_code=200",
					Port: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: port,
					},
				},
			},
			InitialDelaySeconds: probesInitialDelaySeconds,
			PeriodSeconds:       echoServerProbePeriodSeconds,
			FailureThreshold:    echoServerProbeFailureThreshold,
		},
		readinessProbe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/?echo_code=200",
					Port: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: port,
					},
				},
			},
			InitialDelaySeconds: probesInitialDelaySeconds,
			PeriodSeconds:       echoServerProbePeriodSeconds,
			FailureThreshold:    echoServerProbeFailureThreshold,
		},
	}
}

// Helpers -------------------------------------------------------------------------------------------------------------

func echoMatchLabels() map[string]string {
	return map[string]string{"app": echoServerName}
}

func echoPodLabels() map[string]string {
	return map[string]string{
		"app":                   echoServerName,
		kubecommon.LabelEnabled: "true",
	}
}

func echoPodAnnotations(annotations csaQuantityAnnotations) map[string]string {
	ret := make(map[string]string)
	ret[scalecommon.AnnotationTargetContainerName] = echoServerName

	if annotations.cpuStartup != "" {
		ret[scalecommon.AnnotationCpuStartup] = annotations.cpuStartup
	}

	if annotations.cpuPostStartupRequests != "" {
		ret[scalecommon.AnnotationCpuPostStartupRequests] = annotations.cpuPostStartupRequests
	}

	if annotations.cpuPostStartupLimits != "" {
		ret[scalecommon.AnnotationCpuPostStartupLimits] = annotations.cpuPostStartupLimits
	}

	if annotations.memoryStartup != "" {
		ret[scalecommon.AnnotationMemoryStartup] = annotations.memoryStartup
	}

	if annotations.memoryPostStartupRequests != "" {
		ret[scalecommon.AnnotationMemoryPostStartupRequests] = annotations.memoryPostStartupRequests
	}

	if annotations.memoryPostStartupLimits != "" {
		ret[scalecommon.AnnotationMemoryPostStartupLimits] = annotations.memoryPostStartupLimits
	}

	return ret
}

func echoContainerConfigs(
	cpuRequests string, cpuLimits string,
	memoryRequests string, memoryLimits string,
	probesInitialDelaySeconds int32,
) []containerConfig {
	return []containerConfig{
		echoContainerConfigStandard(
			echoServerName,
			80,
			cpuRequests, cpuLimits,
			memoryRequests, memoryLimits,
			probesInitialDelaySeconds,
		),
		echoContainerConfigStandard(
			echoServerNonTargetContainerName,
			81,
			echoServerNonTargetContainerCpuRequests, echoServerNonTargetContainerCpuLimits,
			echoServerNonTargetContainerMemoryRequests, echoServerNonTargetContainerMemoryLimits,
			probesInitialDelaySeconds,
		),
	}
}
