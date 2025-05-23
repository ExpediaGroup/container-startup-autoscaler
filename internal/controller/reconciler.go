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

package controller

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	ccontext "github.com/ExpediaGroup/container-startup-autoscaler/internal/context"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/controller/controllercommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/reconciler"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod"
	cmap "github.com/orcaman/concurrent-map/v2"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// containerStartupAutoscalerReconciler is the reconcile.Reconciler implementation that Controller-runtime is
// configured to use.
type containerStartupAutoscalerReconciler struct {
	pod              *pod.Pod
	controllerConfig controllercommon.ControllerConfig
	reconcilingPods  cmap.ConcurrentMap[string, any]
	mutex            sync.Mutex
}

func newContainerStartupAutoscalerReconciler(
	pod *pod.Pod,
	controllerConfig controllercommon.ControllerConfig,
) *containerStartupAutoscalerReconciler {
	return &containerStartupAutoscalerReconciler{
		pod:              pod,
		controllerConfig: controllerConfig,
		reconcilingPods:  cmap.New[any](),
	}
}

// Reconcile implements controller-runtime's reconcile.Reconciler to reconcile pods that are marked as being eligible
// for startup scaling. It performs a number of tasks to validate configuration, determine the current state and
// action that state. The general premise is to examine the pod changes that led to the reconcile (post filtering) and
// take appropriate action; there will be circumstances when there is effectively no tangible action to take.
func (r *containerStartupAutoscalerReconciler) Reconcile(
	ctx context.Context,
	request reconcile.Request,
) (reconcile.Result, error) {
	namespacedName := request.NamespacedName.String()

	// Prevent concurrent reconciles for the same pod to avoid overlap — requeue if necessary. Although
	// controller-runtime currently guarantees this, this serves as a fail-safe in case its behavior changes in the
	// future. Map get/set must be done exclusively (atomically).
	r.mutex.Lock()

	_, exists := r.reconcilingPods.Get(namespacedName)
	if exists {
		r.mutex.Unlock()
		logging.Infof(ctx, logging.VDebug, "existing reconcile in progress (will requeue)")
		reconciler.ExistingInProgress().Inc()
		return reconcile.Result{RequeueAfter: r.controllerConfig.RequeueDurationSecsDuration()}, nil
	}

	r.reconcilingPods.Set(namespacedName, nil)
	r.mutex.Unlock()

	// Set context items for standard retry.
	ctx = ccontext.WithStandardRetryAttempts(ctx, r.controllerConfig.StandardRetryAttempts)
	ctx = ccontext.WithStandardRetryDelaySecs(ctx, r.controllerConfig.StandardRetryDelaySecs)

	defer r.reconcilingPods.Remove(namespacedName)

	// Get the pod. Note: the latest version of the pod is always retrieved, which may have changed since initial
	// filtering in predicatefunc.go or if a requeue is being processed. There is no affinity with pod version.
	// Reconcilation will still operate correctly in this case as current conditions are always examined.
	podExists, kubePod, err := r.pod.PodHelper.Get(ctx, request.NamespacedName)
	if err != nil {
		logging.Errorf(ctx, err, "unable to get pod (will requeue)")
		reconciler.Failure(reconciler.FailureReasonUnableToGetPod).Inc()
		return reconcile.Result{RequeueAfter: r.controllerConfig.RequeueDurationSecsDuration()}, nil
	}

	if !podExists {
		err = errors.New("pod doesn't exist (won't requeue)")
		logging.Errorf(ctx, err, err.Error())
		reconciler.Failure(reconciler.FailureReasonPodDoesNotExist).Inc()
		return reconcile.Result{}, reconcile.TerminalError(err)
	}

	// Marshal and log pod only if VTrace - expensive.
	if logging.CurrentV == logging.VTrace {
		var podJson []byte
		podJson, err = json.Marshal(kubePod)
		if err != nil {
			logging.Errorf(ctx, err, "unable to marshal pod to json for trace logging")
		} else {
			logging.Infof(ctx, logging.VTrace, "reconciling pod: %s", string(podJson))
		}
	}

	scaleConfigs, err := r.pod.Configuration.Configure(kubePod)
	if err != nil {
		msg := "unable to configure pod (won't requeue)"
		logging.Errorf(ctx, err, msg)
		reconciler.Failure(reconciler.FailureReasonConfiguration).Inc()
		return reconcile.Result{}, reconcile.TerminalError(common.WrapErrorf(err, msg))
	}

	targetContainerName, err := scaleConfigs.TargetContainerName(kubePod)
	if err != nil {
		msg := "unable to get target container name (won't requeue)"
		logging.Errorf(ctx, err, msg)
		reconciler.Failure(reconciler.FailureReasonConfiguration).Inc()
		return reconcile.Result{}, reconcile.TerminalError(common.WrapErrorf(err, msg))
	}

	ctx = ccontext.WithTargetContainerName(ctx, targetContainerName)

	var builder strings.Builder
	for i, scaleConfig := range scaleConfigs.AllConfigurations() {
		if i > 0 {
			builder.WriteString(" / ")
		}
		builder.WriteString(scaleConfig.String())
	}
	logging.Infof(ctx, logging.VDebug, "scale configurations: %s", builder.String())

	targetContainer, err := r.pod.Validation.Validate(ctx, kubePod, targetContainerName, scaleConfigs)
	if err != nil {
		msg := "unable to validate pod (won't requeue)"
		logging.Errorf(ctx, err, msg)
		reconciler.Failure(reconciler.FailureReasonValidation).Inc()
		return reconcile.Result{}, reconcile.TerminalError(common.WrapErrorf(err, msg))
	}

	// Determine target container states.
	states, err := r.pod.TargetContainerState.States(ctx, kubePod, targetContainer, scaleConfigs)
	if err != nil {
		msg := "unable to determine target container states (won't requeue)"
		logging.Errorf(ctx, err, msg)
		reconciler.Failure(reconciler.FailureReasonStatesDetermination).Inc()
		return reconcile.Result{}, reconcile.TerminalError(common.WrapErrorf(err, msg))
	}
	ctx = ccontext.WithTargetContainerStates(ctx, states)

	// Execute action for determined target container states.
	err = r.pod.TargetContainerAction.Execute(ctx, states, kubePod, targetContainer, scaleConfigs)
	if err != nil {
		msg := "unable to action target container states (won't requeue)"
		logging.Errorf(ctx, err, msg)
		reconciler.Failure(reconciler.FailureReasonStatesAction).Inc()
		return reconcile.Result{}, reconcile.TerminalError(common.WrapErrorf(err, msg))
	}

	return reconcile.Result{}, nil
}
