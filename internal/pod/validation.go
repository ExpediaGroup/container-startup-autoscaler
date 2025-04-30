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

package pod

import (
	"context"
	"fmt"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/event/eventcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"k8s.io/api/core/v1"
)

// validation is the default implementation of podcommon.Validation.
type validation struct {
	status            podcommon.Status
	podHelper         kubecommon.PodHelper
	containerHelper   kubecommon.ContainerHelper
	podEventPublisher eventcommon.PodEventPublisher
}

func newValidation(
	status podcommon.Status,
	podHelper kubecommon.PodHelper,
	containerHelper kubecommon.ContainerHelper,
	podEventPublisher eventcommon.PodEventPublisher,
) *validation {
	return &validation{
		status:            status,
		podHelper:         podHelper,
		containerHelper:   containerHelper,
		podEventPublisher: podEventPublisher,
	}
}

// Validate performs core validation using the supplied pod.
func (v *validation) Validate(
	ctx context.Context,
	pod *v1.Pod,
	targetContainerName string,
	scaleConfigs scalecommon.Configurations,
) (*v1.Container, error) {
	// Double check enabled label (originally filtered for informer cache).
	enabled, err := v.podHelper.ExpectedLabelValueAs(pod, kubecommon.LabelEnabled, kubecommon.DataTypeBool)
	if err != nil {
		return nil, v.updateStatusAndGetError(ctx, pod, "unable to get pod enabled label value", err, scaleConfigs)
	}
	if !enabled.(bool) {
		return nil, v.updateStatusAndGetError(ctx, pod, "pod enabled label value is unexpectedly 'false'", nil, scaleConfigs)
	}

	// Ensure pod is not managed by a VPA (not currently compatible).
	for _, ann := range knownVpaAnnotations {
		has, _ := v.podHelper.HasAnnotation(pod, ann)
		if has {
			return nil, v.updateStatusAndGetError(
				ctx, pod,
				fmt.Sprintf("vpa not supported (pod has known '%s' vpa annotation)", ann),
				nil,
				scaleConfigs,
			)
		}
	}

	// Ensure target container is within pod spec.
	if !v.podHelper.IsContainerInSpec(pod, targetContainerName) {
		return nil, v.updateStatusAndGetError(ctx, pod, "target container not in pod spec", nil, scaleConfigs)
	}

	ctr, _ := v.containerHelper.Get(pod, targetContainerName)

	// Ensure at least one of startup or readiness probe is present in container.
	if !v.containerHelper.HasStartupProbe(ctr) && !v.containerHelper.HasReadinessProbe(ctr) {
		return nil, v.updateStatusAndGetError(ctx, pod, "target container does not specify startup probe or readiness probe", nil, scaleConfigs)
	}

	// All resources must be guaranteed in nature to avoid change in pod QoS class.
	qosClass, err := v.podHelper.QOSClass(pod)
	if err != nil {
		return nil, v.updateStatusAndGetError(ctx, pod, "unable to determine pod qos class", err, scaleConfigs)
	}

	if qosClass != v1.PodQOSGuaranteed {
		return nil, v.updateStatusAndGetError(ctx, pod, "pod qos class is not guaranteed", nil, scaleConfigs)
	}

	if err = scaleConfigs.ValidateAll(ctr); err != nil {
		return nil, v.updateStatusAndGetError(ctx, pod, err.Error(), nil, scaleConfigs)
	}

	if err = scaleConfigs.ValidateCollection(); err != nil {
		return nil, v.updateStatusAndGetError(ctx, pod, err.Error(), nil, scaleConfigs)
	}

	return ctr, nil
}

// updateStatusAndGetError updates status and returns a validation error. Status update errors are only logged so not
// to break flow.
func (v *validation) updateStatusAndGetError(
	ctx context.Context,
	pod *v1.Pod,
	errMessage string,
	cause error,
	scaleConfigs scalecommon.Configurations,
) error {
	ret := newValidationError(errMessage, cause)

	_, err := v.status.Update(
		ctx,
		v.podEventPublisher,
		pod,
		ret.Error(),
		podcommon.NewStatesAllUnknown(),
		podcommon.StatusScaleStateNotApplicable,
		scaleConfigs,
		"",
	)
	if err != nil {
		logging.Errorf(ctx, err, "unable to update status (will continue)")
	}

	return ret
}
