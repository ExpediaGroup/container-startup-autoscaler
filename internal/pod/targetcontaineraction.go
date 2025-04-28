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
	"errors"
	"fmt"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/controller/controllercommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"k8s.io/api/core/v1"
)

const eventReasonScaling = "Scaling"

// targetContainerAction is the default implementation of podcommon.TargetContainerAction.
type targetContainerAction struct {
	controllerConfig controllercommon.ControllerConfig
	status           podcommon.Status
	podHelper        kubecommon.PodHelper
}

func newTargetContainerAction(
	controllerConfig controllercommon.ControllerConfig,
	status podcommon.Status,
	podHelper kubecommon.PodHelper,
) *targetContainerAction {
	return &targetContainerAction{
		controllerConfig: controllerConfig,
		status:           status,
		podHelper:        podHelper,
	}
}

// Execute performs the appropriate action for the determined target container state.
func (a *targetContainerAction) Execute(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	targetContainer *v1.Container,
	scaleConfigs scalecommon.Configurations,
) error {
	if states.StartupProbe != podcommon.StateBoolTrue && states.StartupProbe != podcommon.StateBoolFalse {
		panic(fmt.Errorf("unsupported startup probe state '%s'", states.StartupProbe))
	}

	if states.ReadinessProbe != podcommon.StateBoolTrue && states.ReadinessProbe != podcommon.StateBoolFalse {
		panic(fmt.Errorf("unsupported readiness probe state '%s'", states.ReadinessProbe))
	}

	if states.Container != podcommon.StateContainerRunning {
		return a.containerNotRunningAction(ctx, states, pod, scaleConfigs)
	}

	if states.Started == podcommon.StateBoolUnknown {
		return a.startedUnknownAction(ctx)
	}

	if states.Ready == podcommon.StateBoolUnknown {
		return a.readyUnknownAction(ctx)
	}

	if states.Resources == podcommon.StateResourcesUnknown && !a.controllerConfig.ScaleWhenUnknownResources {
		return a.resUnknownAction(ctx, states, pod, targetContainer, scaleConfigs)
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
			return a.notStartedWithStartupResAction(ctx, states, pod, targetContainer, scaleConfigs)
		}
		return a.startedWithStartupResAction(ctx, states, pod, targetContainer, scaleConfigs)

	case podcommon.StateResourcesPostStartup:
		if !isStarted {
			return a.notStartedWithPostStartupResAction(ctx, states, pod, targetContainer, scaleConfigs)
		}
		return a.startedWithPostStartupResAction(ctx, states, pod, targetContainer, scaleConfigs)

	case podcommon.StateResourcesUnknown:
		if !isStarted {
			return a.notStartedWithUnknownResAction(ctx, states, pod, targetContainer, scaleConfigs)
		}
		return a.startedWithUnknownResAction(ctx, states, pod, targetContainer, scaleConfigs)
	}

	panic(errors.New("no action to invoke"))
}

// containerNotRunningAction only logs and updates status since the target container isn't currently running.
func (a *targetContainerAction) containerNotRunningAction(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	scaleConfigs scalecommon.Configurations,
) error {
	a.updateStatusAndLogInfo(
		ctx,
		logging.VInfo,
		pod,
		"target container currently not running",
		states, podcommon.StatusScaleStateNotApplicable,
		scaleConfigs,
		"",
	)
	return nil
}

// startedUnknownAction only logs and updates status since the target container's started status is currently unknown.
func (a *targetContainerAction) startedUnknownAction(ctx context.Context) error {
	logging.Infof(ctx, logging.VDebug, "target container started status currently unknown")
	return nil
}

// readyUnknownAction only logs and updates status since the target container's ready status is currently unknown.
func (a *targetContainerAction) readyUnknownAction(ctx context.Context) error {
	logging.Infof(ctx, logging.VDebug, "target container ready status currently unknown")
	return nil
}

// resUnknownAction updates status and returns an error since an unknown resource configuration has been applied to
// the target container.
func (a *targetContainerAction) resUnknownAction(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	targetContainer *v1.Container,
	scaleConfigs scalecommon.Configurations,
) error {
	msg := "unknown resources applied"
	a.updateStatus(ctx, pod, msg, states, podcommon.StatusScaleStateNotApplicable, scaleConfigs, "")
	return fmt.Errorf("%s (%s)", msg, a.containerResourceConfig(targetContainer, scaleConfigs))
}

// notStartedWithStartupResAction examines conditions and provides relevant feedback since the container is not ready
// with startup resources applied (although those resources might not yet be enacted).
func (a *targetContainerAction) notStartedWithStartupResAction(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	targetContainer *v1.Container,
	scaleConfigs scalecommon.Configurations,
) error {
	return a.processConfigEnacted(ctx, states, pod, targetContainer, scaleConfigs)
}

// notStartedWithPostStartupResAction commands startup resources since the container is not ready but with post-startup
// resources applied. Happens if the container is restarted. Scaling up is done on a best-effort basis since there may
// not enough resources on the node to accommodate.
func (a *targetContainerAction) notStartedWithPostStartupResAction(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	targetContainer *v1.Container,
	scaleConfigs scalecommon.Configurations,
) error {
	resizeFuncs := scale.NewUpdates(scaleConfigs).StartupPodMutationFuncAll(targetContainer)
	newPod, err := a.podHelper.Patch(ctx, pod, resizeFuncs, true)
	if err != nil {
		return common.WrapErrorf(err, "unable to patch container resources")
	}

	a.updateStatusAndLogInfo(
		ctx,
		logging.VInfo,
		newPod,
		"startup resources commanded",
		states,
		podcommon.StatusScaleStateUpCommanded,
		scaleConfigs,
		"",
	)
	return nil
}

// startedWithStartupResAction commands post-startup resources since the container is ready but with startup resources
// applied.
func (a *targetContainerAction) startedWithStartupResAction(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	targetContainer *v1.Container,
	scaleConfigs scalecommon.Configurations,
) error {
	resizeFuncs := scale.NewUpdates(scaleConfigs).PostStartupPodMutationFuncAll(targetContainer)
	newPod, err := a.podHelper.Patch(ctx, pod, resizeFuncs, true)
	if err != nil {
		return common.WrapErrorf(err, "unable to patch container resources")
	}

	a.updateStatusAndLogInfo(
		ctx,
		logging.VInfo,
		newPod,
		"post-startup resources commanded",
		states,
		podcommon.StatusScaleStateDownCommanded,
		scaleConfigs,
		"",
	)
	return nil
}

// startedWithPostStartupResAction examines conditions and provides relevant feedback since the container is not ready
// with post-startup resources applied (although those resources might not yet be enacted).
func (a *targetContainerAction) startedWithPostStartupResAction(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	targetContainer *v1.Container,
	scaleConfigs scalecommon.Configurations,
) error {
	return a.processConfigEnacted(ctx, states, pod, targetContainer, scaleConfigs)
}

// notStartedWithUnknownResAction commands startup resources since the container is not ready but with unknown
// resources applied. Happens if an actor other than CSA modifies the container's resources; only executed if
// configuration flag is set.
func (a *targetContainerAction) notStartedWithUnknownResAction(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	targetContainer *v1.Container,
	scaleConfigs scalecommon.Configurations,
) error {
	resizeFuncs := scale.NewUpdates(scaleConfigs).StartupPodMutationFuncAll(targetContainer)
	newPod, err := a.podHelper.Patch(ctx, pod, resizeFuncs, true)
	if err != nil {
		return common.WrapErrorf(err, "unable to patch container resources")
	}

	a.updateStatusAndLogInfo(
		ctx,
		logging.VInfo,
		newPod,
		"startup resources commanded (unknown resources applied)",
		states,
		podcommon.StatusScaleStateUnknownCommanded,
		scaleConfigs,
		"",
	)
	return nil
}

// startedWithUnknownResAction commands post-startup resources since the container is ready but with unknown resources
// applied. Happens if an actor other than CSA modifies the container's resources; only executed if configuration flag
// is set.
func (a *targetContainerAction) startedWithUnknownResAction(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	targetContainer *v1.Container,
	scaleConfigs scalecommon.Configurations,
) error {
	resizeFuncs := scale.NewUpdates(scaleConfigs).PostStartupPodMutationFuncAll(targetContainer)
	newPod, err := a.podHelper.Patch(ctx, pod, resizeFuncs, true)
	if err != nil {
		return common.WrapErrorf(err, "unable to patch container resources")
	}

	a.updateStatusAndLogInfo(
		ctx,
		logging.VInfo,
		newPod,
		"post-startup resources commanded (unknown resources applied)",
		states,
		podcommon.StatusScaleStateUnknownCommanded,
		scaleConfigs,
		"",
	)
	return nil
}

// processConfigEnacted examines conditions to determine if the previously commanded resources have been enacted.
// Some unfavourable conditions will yield an error. Logging and status are updated appropriately.
func (a *targetContainerAction) processConfigEnacted(
	ctx context.Context,
	states podcommon.States,
	pod *v1.Pod,
	targetContainer *v1.Container,
	scaleConfigs scalecommon.Configurations,
) error {
	switch states.Resize.State {
	case podcommon.StateResizeNotStartedOrCompleted:
		// Examine additional status later that will confirm whether not started or completed.

	case podcommon.StateResizeInProgress:
		a.updateStatusInProgressAndLogInfo(
			ctx,
			logging.VInfo,
			pod,
			states.Resources.HumanReadable()+" scale not yet completed - in progress",
			states,
			scaleConfigs,
		)
		return nil

	case podcommon.StateResizeDeferred:
		a.updateStatusAndLogInfo(
			ctx,
			logging.VInfo,
			pod,
			fmt.Sprintf("%s scale not yet completed - deferred (%s)", states.Resources.HumanReadable(), states.Resize.Message),
			states,
			podcommon.StatusScaleStateNotApplicable,
			scaleConfigs,
			"",
		)
		return nil

	case podcommon.StateResizeInfeasible:
		var scaleState podcommon.StatusScaleState

		switch states.Resources {
		case podcommon.StateResourcesPostStartup:
			scaleState = podcommon.StatusScaleStateDownFailed
		case podcommon.StateResourcesStartup:
			scaleState = podcommon.StatusScaleStateUpFailed
		}

		msg := fmt.Sprintf("%s scale failed - infeasible (%s)", states.Resources.HumanReadable(), states.Resize.Message)
		a.updateStatus(ctx, pod, msg, states, scaleState, scaleConfigs, "infeasible")
		return fmt.Errorf("%s (%s)", msg, a.containerResourceConfig(targetContainer, scaleConfigs))

	case podcommon.StateResizeError:
		var scaleState podcommon.StatusScaleState

		switch states.Resources {
		case podcommon.StateResourcesPostStartup:
			scaleState = podcommon.StatusScaleStateDownFailed
		case podcommon.StateResourcesStartup:
			scaleState = podcommon.StatusScaleStateUpFailed
		}

		msg := fmt.Sprintf("%s scale failed - error (%s)", states.Resources.HumanReadable(), states.Resize.Message)
		a.updateStatus(ctx, pod, msg, states, scaleState, scaleConfigs, "error")
		return fmt.Errorf("%s (%s)", msg, a.containerResourceConfig(targetContainer, scaleConfigs))

	default:
		panic(fmt.Errorf("unknown resize state '%s'", states.Resize.State))
	}

	// Resize is either not started or completed (StateResizeNotStartedOrCompleted), so examine StatusResources.
	switch states.StatusResources {
	case podcommon.StateStatusResourcesIncomplete:
		// Target container current CPU and/or memory resources are missing. Update status, log and return with the
		// expectation that the missing items become available in the future.
		a.updateStatusInProgressAndLogInfo(
			ctx,
			logging.VDebug,
			pod,
			"target container current cpu and/or memory resources currently missing",
			states,
			scaleConfigs,
		)
		return nil

	case podcommon.StateStatusResourcesContainerResourcesMatch: // Want this, but here so we can panic on default below.

	case podcommon.StateStatusResourcesContainerResourcesMismatch:
		// Target container current CPU and/or memory resources don't match target container's 'requests'. Update
		// status, log and return with the expectation that they match in the future.
		a.updateStatusInProgressAndLogInfo(
			ctx,
			logging.VDebug,
			pod,
			"target container current cpu and/or memory resources currently don't match target container's 'requests'",
			states,
			scaleConfigs,
		)
		return nil

	case podcommon.StateStatusResourcesUnknown:
		// Target container current CPU and/or memory resources are unknown. Update status, log and return with the
		// expectation that they become known in the future.
		a.updateStatusInProgressAndLogInfo(
			ctx,
			logging.VDebug,
			pod,
			"target container current cpu and/or memory resources currently unknown",
			states,
			scaleConfigs,
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

	a.updateStatusAndLogInfo(
		ctx,
		logging.VInfo,
		pod,
		states.Resources.HumanReadable()+" resources enacted",
		states,
		scaleState,
		scaleConfigs,
		"",
	)
	return nil
}

// containerResourceConfig returns a human-readable string representing current container resource configuration, for
// information purposes.
func (a *targetContainerAction) containerResourceConfig(
	targetContainer *v1.Container,
	scaleConfigs scalecommon.Configurations,
) string {
	return fmt.Sprintf(
		"target container resources: [%s], configurations: [%s]",
		targetContainer.Resources.String(),
		scaleConfigs.String(),
	)
}

// updateStatus updates status according to the supplied arguments. Errors are only logged so not to break flow.
func (a *targetContainerAction) updateStatus(
	ctx context.Context,
	pod *v1.Pod,
	status string,
	states podcommon.States,
	scaleState podcommon.StatusScaleState,
	scaleConfigs scalecommon.Configurations,
	failReason string,
) {
	_, err := a.status.Update(ctx, pod, status, states, scaleState, scaleConfigs, failReason)
	if err != nil {
		logging.Errorf(ctx, err, "unable to update status (will continue)")
	}
}

// updateStatusAndLogInfo updates status and logs an info message.
func (a *targetContainerAction) updateStatusAndLogInfo(
	ctx context.Context,
	v logging.V,
	pod *v1.Pod,
	message string,
	states podcommon.States,
	scaleState podcommon.StatusScaleState,
	scaleConfigs scalecommon.Configurations,
	failReason string,
) {
	a.updateStatus(ctx, pod, message, states, scaleState, scaleConfigs, failReason)
	logging.Infof(ctx, v, message)
}

// updateStatusInProgressAndLogInfo updates status when in progress and logs an info message.
func (a *targetContainerAction) updateStatusInProgressAndLogInfo(
	ctx context.Context,
	v logging.V,
	pod *v1.Pod,
	logMessage string,
	states podcommon.States,
	scaleConfigs scalecommon.Configurations,
) {
	a.updateStatus(
		ctx,
		pod,
		states.Resources.HumanReadable()+" scale not yet completed - in progress",
		states,
		podcommon.StatusScaleStateNotApplicable,
		scaleConfigs,
		"",
	)
	logging.Infof(ctx, v, logMessage)
}
