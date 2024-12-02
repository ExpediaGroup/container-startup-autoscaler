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
	"errors"
	"fmt"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/controller/controllercommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/scale"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

const eventReasonScaling = "Scaling"

// TargetContainerAction performs actions based on target container state.
type TargetContainerAction interface {
	Execute(context.Context, podcommon.States, *v1.Pod, podcommon.ScaleConfig) error
}

// targetContainerAction is the default implementation of TargetContainerAction.
type targetContainerAction struct {
	controllerConfig    controllercommon.ControllerConfig
	recorder            record.EventRecorder
	status              Status
	kubeHelper          KubeHelper
	containerKubeHelper ContainerKubeHelper
}

func newTargetContainerAction(
	controllerConfig controllercommon.ControllerConfig,
	recorder record.EventRecorder,
	status Status,
	kubeHelper KubeHelper,
	containerKubeHelper ContainerKubeHelper,
) *targetContainerAction {
	return &targetContainerAction{
		controllerConfig:    controllerConfig,
		recorder:            recorder,
		status:              status,
		kubeHelper:          kubeHelper,
		containerKubeHelper: containerKubeHelper,
	}
}

// TODO(wt) might want to protect against resize flapping (startup -> post-startup -> startup ad nauseum) - disable for
//  a period of time with startup resources.

// Execute performs the appropriate action for the determined target container state.
func (a *targetContainerAction) Execute(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	config podcommon.ScaleConfig,
) error {
	if states.StartupProbe != podcommon.StateBoolTrue && states.StartupProbe != podcommon.StateBoolFalse {
		panic(fmt.Errorf("unsupported startup probe state '%s'", states.StartupProbe))
	}

	if states.ReadinessProbe != podcommon.StateBoolTrue && states.ReadinessProbe != podcommon.StateBoolFalse {
		panic(fmt.Errorf("unsupported readiness probe state '%s'", states.ReadinessProbe))
	}

	if states.Container != podcommon.StateContainerRunning {
		return a.containerNotRunningAction(ctx, states, pod)
	}

	if states.Started == podcommon.StateBoolUnknown {
		return a.startedUnknownAction(ctx, states, pod)
	}

	if states.Ready == podcommon.StateBoolUnknown {
		return a.readyUnknownAction(ctx, states, pod)
	}

	if states.Resources == podcommon.StateResourcesUnknown && !a.controllerConfig.ScaleWhenUnknownResources {
		return a.resUnknownAction(ctx, states, pod, config)
	}

	/*
		If the container specifies a startup probe, use only the 'started' signal of the container status to determine
		whether the container is started. Otherwise, use both the 'started' and 'ready' signals. It's preferable to
		have a startup probe defined since this unambiguously indicates whether a container is started whereas only a
		readiness probe may indicate other conditions that will cause unnecessary scaling (e.g. the readiness probe
		transiently failing post-startup).

		When only startup probe is present:
		- Container status 'started' is false when container is (re)started and true when startup probe succeeds.
		- Container status 'ready' is false when container is (re)started and true when 'started' is true.

		When only readiness probe is present:
		- Container status 'started' is false when container is (re)started and true when the container is running and
		  has passed the postStart lifecycle hook.
		- Container status 'ready' is false when container is (re)started and true when readiness probe succeeds.

		When both startup and readiness probes are present:
		- Container status 'started' is false when container is (re)started and true when startup probe succeeds.
		- Container status 'ready' is false when container is (re)started and true when readiness probe succeeds.
	*/
	var isStarted bool
	if states.StartupProbe.Bool() {
		isStarted = states.Started.Bool()
	} else if states.ReadinessProbe.Bool() {
		isStarted = states.Started.Bool() && states.Ready.Bool()
	} else {
		panic(errors.New("neither startup probe or readiness probe present"))
	}

	switch states.Resources {
	case podcommon.StateResourcesStartup:
		if !isStarted {
			return a.notStartedWithStartupResAction(ctx, states, pod, config)
		}
		return a.startedWithStartupResAction(ctx, states, pod, config)

	case podcommon.StateResourcesPostStartup:
		if !isStarted {
			return a.notStartedWithPostStartupResAction(ctx, states, pod, config)
		}
		return a.startedWithPostStartupResAction(ctx, states, pod, config)

	case podcommon.StateResourcesUnknown:
		if !isStarted {
			return a.notStartedWithUnknownResAction(ctx, states, pod, config)
		}
		return a.startedWithUnknownResAction(ctx, states, pod, config)
	}

	panic(errors.New("no action to invoke"))
}

// containerNotRunningAction only logs and updates status since the target container isn't currently running.
func (a *targetContainerAction) containerNotRunningAction(ctx context.Context, states podcommon.States, pod *v1.Pod) error {
	a.logInfoAndUpdateStatus(
		ctx,
		logging.VDebug,
		states, podcommon.StatusScaleStateNotApplicable,
		pod,
		"target container currently not running",
	)
	return nil
}

// startedUnknownAction only logs and updates status since the target container's started status is currently unknown.
func (a *targetContainerAction) startedUnknownAction(ctx context.Context, states podcommon.States, pod *v1.Pod) error {
	a.logInfoAndUpdateStatus(
		ctx,
		logging.VDebug,
		states, podcommon.StatusScaleStateNotApplicable,
		pod,
		"target container started status currently unknown",
	)
	return nil
}

// readyUnknownAction only logs and updates status since the target container's ready status is currently unknown.
func (a *targetContainerAction) readyUnknownAction(ctx context.Context, states podcommon.States, pod *v1.Pod) error {
	a.logInfoAndUpdateStatus(
		ctx,
		logging.VDebug,
		states, podcommon.StatusScaleStateNotApplicable,
		pod,
		"target container ready status currently unknown",
	)
	return nil
}

// resUnknownAction updates status and returns an error since an unknown resource configuration has been applied to
// the target container.
func (a *targetContainerAction) resUnknownAction(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	config podcommon.ScaleConfig,
) error {
	msg := "unknown resources applied"
	a.updateStatus(ctx, states, podcommon.StatusScaleStateNotApplicable, pod, msg)
	return fmt.Errorf("%s (%s)", msg, a.containerResourceConfig(ctx, pod, config))
}

// notStartedWithStartupResAction examines conditions and provides relevant feedback since the container is not ready
// with startup resources applied (although those resources might not yet be enacted).
func (a *targetContainerAction) notStartedWithStartupResAction(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	config podcommon.ScaleConfig,
) error {
	return a.processConfigEnacted(ctx, states, pod, config)
}

// notStartedWithPostStartupResAction commands startup resources since the container is not ready but with post-startup
// resources applied. Happens if the container is restarted. Scaling up is done on a best-effort basis since there may
// not enough resources on the node to accommodate.
func (a *targetContainerAction) notStartedWithPostStartupResAction(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	config podcommon.ScaleConfig,
) error {
	msg := "startup resources commanded"
	_, err := a.kubeHelper.UpdateContainerResources(
		ctx,
		pod, config.GetTargetContainerName(),
		config.GetCpuConfig().Startup, config.GetCpuConfig().Startup,
		config.GetMemoryConfig().Startup, config.GetMemoryConfig().Startup,
		a.status.UpdateMutatePodFunc(ctx, msg, states, podcommon.StatusScaleStateUpCommanded),
	)
	if err != nil {
		return common.WrapErrorf(err, "unable to patch container resources")
	}

	logging.Infof(ctx, logging.VInfo, msg)
	a.normalEvent(pod, eventReasonScaling, msg)
	return nil
}

// startedWithStartupResAction commands post-startup resources since the container is ready but with startup resources
// applied.
func (a *targetContainerAction) startedWithStartupResAction(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	config podcommon.ScaleConfig,
) error {
	msg := "post-startup resources commanded"
	_, err := a.kubeHelper.UpdateContainerResources(
		ctx,
		pod, config.GetTargetContainerName(),
		config.GetCpuConfig().PostStartupRequests, config.GetCpuConfig().PostStartupLimits,
		config.GetMemoryConfig().PostStartupRequests, config.GetMemoryConfig().PostStartupLimits,
		a.status.UpdateMutatePodFunc(ctx, msg, states, podcommon.StatusScaleStateDownCommanded),
	)
	if err != nil {
		return common.WrapErrorf(err, "unable to patch container resources")
	}

	logging.Infof(ctx, logging.VInfo, msg)
	a.normalEvent(pod, eventReasonScaling, msg)
	return nil
}

// startedWithPostStartupResAction examines conditions and provides relevant feedback since the container is not ready
// with post-startup resources applied (although those resources might not yet be enacted).
func (a *targetContainerAction) startedWithPostStartupResAction(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	config podcommon.ScaleConfig,
) error {
	return a.processConfigEnacted(ctx, states, pod, config)
}

// notStartedWithUnknownResAction commands startup resources since the container is not ready but with unknown
// resources applied. Happens if an actor other than CSA modifies the container's resources; only executed if
// configuration flag is set.
func (a *targetContainerAction) notStartedWithUnknownResAction(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	config podcommon.ScaleConfig,
) error {
	msg := "startup resources commanded (unknown resources applied)"
	_, err := a.kubeHelper.UpdateContainerResources(
		ctx,
		pod, config.GetTargetContainerName(),
		config.GetCpuConfig().Startup, config.GetCpuConfig().Startup,
		config.GetMemoryConfig().Startup, config.GetMemoryConfig().Startup,
		a.status.UpdateMutatePodFunc(ctx, msg, states, podcommon.StatusScaleStateUnknownCommanded),
	)
	if err != nil {
		return common.WrapErrorf(err, "unable to patch container resources")
	}

	logging.Infof(ctx, logging.VInfo, msg)
	a.normalEvent(pod, eventReasonScaling, msg)
	return nil
}

// startedWithUnknownResAction commands post-startup resources since the container is ready but with unknown resources
// applied. Happens if an actor other than CSA modifies the container's resources; only executed if configuration flag
// is set.
func (a *targetContainerAction) startedWithUnknownResAction(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	config podcommon.ScaleConfig,
) error {
	msg := "post-startup resources commanded (unknown resources applied)"
	_, err := a.kubeHelper.UpdateContainerResources(
		ctx,
		pod, config.GetTargetContainerName(),
		config.GetCpuConfig().PostStartupRequests, config.GetCpuConfig().PostStartupLimits,
		config.GetMemoryConfig().PostStartupRequests, config.GetMemoryConfig().PostStartupLimits,
		a.status.UpdateMutatePodFunc(ctx, msg, states, podcommon.StatusScaleStateUnknownCommanded),
	)
	if err != nil {
		return common.WrapErrorf(err, "unable to patch container resources")
	}

	logging.Infof(ctx, logging.VInfo, msg)
	a.normalEvent(pod, eventReasonScaling, msg)
	return nil
}

// processConfigEnacted examines conditions to determine if the previously commanded resources have been enacted.
// Some unfavourable conditions will yield an error. Logging and status are updated appropriately.
func (a *targetContainerAction) processConfigEnacted(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	config podcommon.ScaleConfig,
) error {
	// Examine resize status.
	switch a.kubeHelper.ResizeStatus(pod) {
	case "":
		// Empty means 'no pending resize' - assume it's completed and examine additional status later that will
		// confirm this.

	case v1.PodResizeStatusProposed:
		a.logInfoAndUpdateStatus(
			ctx,
			logging.VDebug,
			states, podcommon.StatusScaleStateNotApplicable,
			pod,
			states.Resources.HumanReadable()+" scale not yet completed - has been proposed",
		)
		return nil

	case v1.PodResizeStatusInProgress:
		a.logInfoAndUpdateStatus(
			ctx,
			logging.VDebug,
			states, podcommon.StatusScaleStateNotApplicable,
			pod,
			states.Resources.HumanReadable()+" scale not yet completed - in progress",
		)
		return nil

	case v1.PodResizeStatusDeferred:
		a.logInfoAndUpdateStatus(
			ctx,
			logging.VDebug,
			states, podcommon.StatusScaleStateNotApplicable,
			pod,
			states.Resources.HumanReadable()+" scale not yet completed - deferred",
		)
		return nil

	case v1.PodResizeStatusInfeasible:
		var scaleState podcommon.StatusScaleState

		switch states.Resources {
		case podcommon.StateResourcesPostStartup:
			scaleState = podcommon.StatusScaleStateDownFailed
		case podcommon.StateResourcesStartup:
			scaleState = podcommon.StatusScaleStateUpFailed
		}

		msg := states.Resources.HumanReadable() + " scale failed - infeasible"
		a.updateStatus(ctx, states, scaleState, pod, msg)
		scale.Failure(states.Resources.Direction(), "infeasible").Inc()
		a.warningEvent(pod, eventReasonScaling, msg)
		return fmt.Errorf("%s (%s)", msg, a.containerResourceConfig(ctx, pod, config))

	default:
		var scaleState podcommon.StatusScaleState

		switch states.Resources {
		case podcommon.StateResourcesPostStartup:
			scaleState = podcommon.StatusScaleStateDownFailed
		case podcommon.StateResourcesStartup:
			scaleState = podcommon.StatusScaleStateUpFailed
		}

		msg := states.Resources.HumanReadable() + " scale: unknown status"
		a.updateStatus(ctx, states, scaleState, pod, msg)
		scale.Failure(states.Resources.Direction(), "unknownstatus").Inc()
		a.warningEvent(pod, eventReasonScaling, msg)
		return fmt.Errorf("%s '%s'", msg, a.kubeHelper.ResizeStatus(pod))
	}

	// Resize is not pending, so examine StatusResources.
	switch states.StatusResources {
	case podcommon.StateStatusResourcesIncomplete:
		// Target container current CPU and/or memory resources are missing. Log and return with the expectation that
		// the missing items become available in the future.
		a.logInfoAndUpdateStatus(
			ctx,
			logging.VDebug,
			states, podcommon.StatusScaleStateNotApplicable,
			pod,
			"target container current cpu and/or memory resources currently missing",
		)
		return nil

	case podcommon.StateStatusResourcesContainerResourcesMatch: // Want this, but here so we can panic on default below.

	case podcommon.StateStatusResourcesContainerResourcesMismatch:
		// Target container current CPU and/or memory resources don't match target container's 'requests'. Log and
		// return with the expectation that they match in the future.
		a.logInfoAndUpdateStatus(
			ctx,
			logging.VDebug,
			states, podcommon.StatusScaleStateNotApplicable,
			pod,
			"target container current cpu and/or memory resources currently don't match target container's 'requests'",
		)
		return nil

	case podcommon.StateStatusResourcesUnknown:
		// Target container current CPU and/or memory resources are unknown. Log and return with the expectation that
		// they become known in the future.
		a.logInfoAndUpdateStatus(
			ctx,
			logging.VDebug,
			states, podcommon.StatusScaleStateNotApplicable,
			pod,
			"target container current cpu and/or memory resources currently unknown",
		)
		return nil

	default:
		panic(fmt.Errorf("unknown state '%s'", states.StatusResources))
	}

	// Desired state: target container resources correctly enacted.
	var scaleState podcommon.StatusScaleState

	switch states.Resources {
	case podcommon.StateResourcesPostStartup:
		scaleState = podcommon.StatusScaleStateDownEnacted
	case podcommon.StateResourcesStartup:
		scaleState = podcommon.StatusScaleStateUpEnacted
	}

	msg := states.Resources.HumanReadable() + " resources enacted"
	a.logInfoAndUpdateStatus(ctx, logging.VInfo, states, scaleState, pod, msg)
	a.normalEvent(pod, eventReasonScaling, msg)
	return nil
}

// containerResourceConfig returns a human-readable string representing current container resource configuration, for
// information purposes.
func (a *targetContainerAction) containerResourceConfig(
	ctx context.Context,
	pod *v1.Pod,
	config podcommon.ScaleConfig,
) string {
	ctr, err := a.containerKubeHelper.Get(pod, config.GetTargetContainerName())
	if err != nil {
		logging.Errorf(ctx, err, "unable to get container to generate container resource config (will continue)")
		return "[unable to generate container resource config]"
	}

	cpuRequests := a.containerKubeHelper.Requests(ctr, v1.ResourceCPU)
	cpuLimits := a.containerKubeHelper.Limits(ctr, v1.ResourceCPU)
	memoryRequests := a.containerKubeHelper.Requests(ctr, v1.ResourceMemory)
	memoryLimits := a.containerKubeHelper.Limits(ctr, v1.ResourceMemory)

	return fmt.Sprintf(
		"container cpu requests/limits: %s/%s, "+
			"container memory requests/limits: %s/%s, "+
			"configuration values: [%s]",
		cpuRequests.String(), cpuLimits.String(), memoryRequests.String(), memoryLimits.String(),
		config.String(),
	)
}

// updateStatus updates status according to the supplied arguments. Errors are only logged so not to break flow.
func (a *targetContainerAction) updateStatus(
	ctx context.Context,
	states podcommon.States,
	scaleState podcommon.StatusScaleState,
	pod *v1.Pod,
	status string,
) {
	_, err := a.status.Update(ctx, pod, status, states, scaleState)
	if err != nil {
		logging.Errorf(ctx, err, "unable to update status (will continue)")
	}
}

// logInfoAndUpdateStatus logs an info message and updates status.
func (a *targetContainerAction) logInfoAndUpdateStatus(
	ctx context.Context,
	v logging.V,
	states podcommon.States,
	scaleState podcommon.StatusScaleState,
	pod *v1.Pod,
	message string,
) {
	logging.Infof(ctx, v, message)
	a.updateStatus(ctx, states, scaleState, pod, message)
}

// normalEvent yields a 'normal' Kube event for the supplied pod with the supplied reason and message.
func (a *targetContainerAction) normalEvent(pod *v1.Pod, reason string, message string) {
	a.recorder.Event(pod, v1.EventTypeNormal, reason, common.CapitalizeFirstChar(message))
}

// warningEvent yields a 'warning' Kube event for the supplied pod with the supplied reason and message.
func (a *targetContainerAction) warningEvent(pod *v1.Pod, reason string, message string) {
	a.recorder.Event(pod, v1.EventTypeWarning, reason, common.CapitalizeFirstChar(message))
}
