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
	"encoding/json"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment ----------------------------------------------------------------------------------------------------------

type deploymentConfig struct {
	namespace        string
	name             string
	replicas         int32
	matchLabels      map[string]string
	podConfig        podConfig
	containerConfigs []containerConfig
}

func (d deploymentConfig) removeStartupProbes() {
	for i := 0; i < len(d.containerConfigs); i++ {
		d.containerConfigs[i].startupProbe = nil
	}
}

func (d deploymentConfig) removeReadinessProbes() {
	for i := 0; i < len(d.containerConfigs); i++ {
		d.containerConfigs[i].readinessProbe = nil
	}
}

func (d deploymentConfig) deployment() *appsv1.Deployment {
	var containers []corev1.Container
	for _, config := range d.containerConfigs {
		containers = append(containers, config.container())
	}

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: d.namespace,
			Name:      d.name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &d.replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: d.matchLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      d.podConfig.labels,
					Annotations: d.podConfig.annotations,
				},
				Spec: corev1.PodSpec{Containers: containers},
			},
		},
	}
}

func (d deploymentConfig) deploymentJson() string {
	jsonBytes, _ := json.Marshal(d.deployment())
	return string(jsonBytes)
}

// StatefulSet ---------------------------------------------------------------------------------------------------------

type statefulSetConfig struct {
	namespace        string
	name             string
	replicas         int32
	matchLabels      map[string]string
	podConfig        podConfig
	containerConfigs []containerConfig
}

func (s statefulSetConfig) removeStartupProbes() {
	for i := 0; i < len(s.containerConfigs); i++ {
		s.containerConfigs[i].startupProbe = nil
	}
}

func (s statefulSetConfig) removeReadinessProbes() {
	for i := 0; i < len(s.containerConfigs); i++ {
		s.containerConfigs[i].readinessProbe = nil
	}
}

func (s statefulSetConfig) statefulSet() *appsv1.StatefulSet {
	var containers []corev1.Container
	for _, config := range s.containerConfigs {
		containers = append(containers, config.container())
	}

	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.namespace,
			Name:      s.name,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &s.replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: s.matchLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      s.podConfig.labels,
					Annotations: s.podConfig.annotations,
				},
				Spec: corev1.PodSpec{Containers: containers},
			},
		},
	}
}

func (s statefulSetConfig) statefulSetJson() string {
	jsonBytes, _ := json.Marshal(s.statefulSet())
	return string(jsonBytes)
}

// DaemonSet -----------------------------------------------------------------------------------------------------------

type daemonSetConfig struct {
	namespace        string
	name             string
	matchLabels      map[string]string
	podConfig        podConfig
	containerConfigs []containerConfig
}

func (d daemonSetConfig) removeStartupProbes() {
	for i := 0; i < len(d.containerConfigs); i++ {
		d.containerConfigs[i].startupProbe = nil
	}
}

func (d daemonSetConfig) removeReadinessProbes() {
	for i := 0; i < len(d.containerConfigs); i++ {
		d.containerConfigs[i].readinessProbe = nil
	}
}

func (d daemonSetConfig) daemonSet() *appsv1.DaemonSet {
	var containers []corev1.Container
	for _, config := range d.containerConfigs {
		containers = append(containers, config.container())
	}

	return &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "DaemonSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: d.namespace,
			Name:      d.name,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: d.matchLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      d.podConfig.labels,
					Annotations: d.podConfig.annotations,
				},
				Spec: corev1.PodSpec{Containers: containers},
			},
		},
	}
}

func (d daemonSetConfig) daemonSetJson() string {
	jsonBytes, _ := json.Marshal(d.daemonSet())
	return string(jsonBytes)
}

// Pod -----------------------------------------------------------------------------------------------------------------

type podConfig struct {
	labels      map[string]string
	annotations map[string]string
}

// Container -----------------------------------------------------------------------------------------------------------

type containerConfig struct {
	name           string
	image          string
	containerPort  int32
	env            []corev1.EnvVar
	resizePolicy   []corev1.ContainerResizePolicy
	resources      corev1.ResourceRequirements
	startupProbe   *corev1.Probe
	readinessProbe *corev1.Probe
}

func (c containerConfig) container() corev1.Container {
	return corev1.Container{
		Name:            c.name,
		Image:           c.image,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Ports: []corev1.ContainerPort{
			{ContainerPort: c.containerPort},
		},
		Env:            c.env,
		ResizePolicy:   c.resizePolicy,
		Resources:      c.resources,
		StartupProbe:   c.startupProbe,
		ReadinessProbe: c.readinessProbe,
	}
}
