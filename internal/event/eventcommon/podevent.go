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

package eventcommon

import "k8s.io/api/core/v1"

// PodEventType indicates the type of pod event.
type PodEventType string

const (
	PodEventTypeCreate PodEventType = "create"
	PodEventTypeDelete PodEventType = "delete"
	PodEventTypeUpdate PodEventType = "update"
)

// PodEvent represents a pod event triggered via a Kubernetes watch.
type PodEvent struct {
	EventType PodEventType
	Pod       *v1.Pod
}

func NewPodEvent(
	eventType PodEventType,
	pod *v1.Pod,
) PodEvent {
	return PodEvent{
		EventType: eventType,
		Pod:       pod,
	}
}
