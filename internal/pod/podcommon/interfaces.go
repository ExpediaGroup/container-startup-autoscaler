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

package podcommon

import (
	"context"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	v1 "k8s.io/api/core/v1"
)

// Configuration performs operations relating to configuration.
type Configuration interface {
	Configure(*v1.Pod) (scalecommon.Configs, error)
}

// Validation performs operations relating to validation.
type Validation interface {
	Validate(context.Context, *v1.Pod, string, scalecommon.Configs) (*v1.Container, error)
}

// TargetContainerState performs operations relating to determining target container state.
type TargetContainerState interface {
	States(context.Context, *v1.Pod, *v1.Container, scalecommon.Configs) (States, error)
}

// TargetContainerAction performs actions based on target container state.
type TargetContainerAction interface {
	Execute(context.Context, States, *v1.Pod, *v1.Container, scalecommon.Configs) error
}

// Status performs operations relating to controller status.
type Status interface {
	Update(context.Context, *v1.Pod, string, States, StatusScaleState, scalecommon.Configs) (*v1.Pod, error)
	PodMutationFunc(context.Context, string, States, StatusScaleState, scalecommon.Configs) func(pod *v1.Pod) error
}
