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
	"context"
	"fmt"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

const eventReasonValidation = "Validation"

// Validation performs operations relating to validation.
type Validation interface {
	Validate(context.Context, *v1.Pod, podcommon.ScaleConfig, func(podcommon.ScaleConfig)) error
}

// validation is the default implementation of Validation.
type validation struct {
	recorder            record.EventRecorder
	status              Status
	kubeHelper          KubeHelper
	containerKubeHelper ContainerKubeHelper
}

func newValidation(
	recorder record.EventRecorder,
	status Status,
	kubeHelper KubeHelper,
	containerKubeHelper ContainerKubeHelper,
) *validation {
	return &validation{
		recorder:            recorder,
		status:              status,
		kubeHelper:          kubeHelper,
		containerKubeHelper: containerKubeHelper,
	}
}

// Validate performs core validation using the supplied pod. Populates (or repopulates) scaleConfigToPopulate;
// additional arbitrary code may be run immediately after the scale configuration is populated via
// afterScaleConfigPopulatedFunc.
func (v *validation) Validate(
	ctx context.Context,
	pod *v1.Pod,
	scaleConfigToPopulate podcommon.ScaleConfig,
	afterScaleConfigPopulatedFunc func(podcommon.ScaleConfig),
) error {
	// Double check enabled label (originally filtered for informer cache).
	enabled, err := v.kubeHelper.ExpectedLabelValueAs(pod, podcommon.LabelEnabled, podcommon.TypeBool)
	if err != nil {
		return v.updateStatusAndGetError(ctx, pod, "unable to get pod enabled label value", err)
	}
	if !enabled.(bool) {
		return v.updateStatusAndGetError(ctx, pod, "pod enabled label value is unexpectedly 'false'", nil)
	}

	// Ensure pod is not managed by a VPA (not currently compatible).
	for _, ann := range podcommon.KnownVpaAnnotations {
		has, _ := v.kubeHelper.HasAnnotation(pod, ann)
		if has {
			return v.updateStatusAndGetError(
				ctx, pod,
				fmt.Sprintf("vpa not supported (pod has known '%s' vpa annotation)", ann),
				nil,
			)
		}
	}

	// Get and validate configuration.
	err = scaleConfigToPopulate.StoreFromAnnotations(pod)
	if err != nil {
		return v.updateStatusAndGetError(ctx, pod, "unable to get annotation configuration values", err)
	}
	populatedScaleConfig := scaleConfigToPopulate

	if afterScaleConfigPopulatedFunc != nil {
		afterScaleConfigPopulatedFunc(populatedScaleConfig)
	}

	if err = populatedScaleConfig.Validate(); err != nil {
		return v.updateStatusAndGetError(ctx, pod, "unable to validate configuration values", err)
	}

	// Ensure target container is within pod spec.
	if !v.kubeHelper.IsContainerInSpec(pod, populatedScaleConfig.GetTargetContainerName()) {
		return v.updateStatusAndGetError(ctx, pod, "target container not in pod spec", nil)
	}

	ctr, _ := v.containerKubeHelper.Get(pod, populatedScaleConfig.GetTargetContainerName())

	// Ensure at least one of startup or readiness probe is present in container.
	if !v.containerKubeHelper.HasStartupProbe(ctr) && !v.containerKubeHelper.HasReadinessProbe(ctr) {
		return v.updateStatusAndGetError(ctx, pod, "target container does not specify startup probe or readiness probe", nil)
	}

	// Ensure container specifies requests for both CPU and memory.
	cpuRequests := v.containerKubeHelper.Requests(ctr, v1.ResourceCPU)
	if cpuRequests.IsZero() {
		return v.updateStatusAndGetError(ctx, pod, "target container does not specify cpu requests", nil)
	}

	memoryRequests := v.containerKubeHelper.Requests(ctr, v1.ResourceMemory)
	if memoryRequests.IsZero() {
		return v.updateStatusAndGetError(ctx, pod, "target container does not specify memory requests", nil)
	}

	// TODO(wt) only guaranteed configuration is possible at the moment (pod QoS is immutable)
	cpuLimits := v.containerKubeHelper.Limits(ctr, v1.ResourceCPU)
	if !cpuRequests.Equal(cpuLimits) {
		return v.updateStatusAndGetError(
			ctx, pod,
			fmt.Sprintf(
				"target container cpu requests (%s) must equal limits (%s) - change in qos class is not yet permitted by kube",
				cpuRequests.String(), cpuLimits.String(),
			),
			nil,
		)
	}

	// TODO(wt) only guaranteed configuration is possible at the moment (pod QoS is immutable)
	memoryLimits := v.containerKubeHelper.Limits(ctr, v1.ResourceMemory)
	if !memoryRequests.Equal(memoryLimits) {
		return v.updateStatusAndGetError(
			ctx, pod,
			fmt.Sprintf(
				"target container memory requests (%s) must equal limits (%s) - change in qos class is not yet permitted by kube",
				memoryRequests.String(), memoryLimits.String(),
			),
			nil,
		)
	}

	// Ensure correct resize configuration for both CPU and memory.
	cpuResizePolicy, err := v.containerKubeHelper.ResizePolicy(ctr, v1.ResourceCPU)
	if err != nil {
		return v.updateStatusAndGetError(ctx, pod, "unable to get target container cpu resize policy", err)
	}
	if cpuResizePolicy != v1.NotRequired {
		return v.updateStatusAndGetError(
			ctx, pod,
			fmt.Sprintf("target container cpu resize policy is not '%s' ('%s')", v1.NotRequired, cpuResizePolicy),
			nil,
		)
	}

	memoryResizePolicy, err := v.containerKubeHelper.ResizePolicy(ctr, v1.ResourceMemory)
	if err != nil {
		return v.updateStatusAndGetError(ctx, pod, "unable to get target container memory resize policy", err)
	}
	if memoryResizePolicy != v1.NotRequired {
		return v.updateStatusAndGetError(
			ctx, pod,
			fmt.Sprintf("target container memory resize policy is not '%s' ('%s')", v1.NotRequired, memoryResizePolicy),
			nil,
		)
	}

	return nil
}

// updateStatusAndGetError updates status and returns a validation error. Status update errors are only logged so not
// to break flow.
func (v *validation) updateStatusAndGetError(
	ctx context.Context,
	pod *v1.Pod,
	errMessage string,
	cause error,
) error {
	ret := NewValidationError(errMessage, cause)

	_, err := v.status.Update(ctx, pod, ret.Error(), podcommon.NewStatesAllUnknown(), podcommon.StatusScaleStateNotApplicable)
	if err != nil {
		logging.Errorf(ctx, err, "unable to update status (will continue)")
	}

	v.recorder.Event(pod, v1.EventTypeWarning, eventReasonValidation, common.CapitalizeFirstChar(ret.Error()))

	return ret
}
