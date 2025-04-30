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
	"sync"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/controller/controllercommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	csametrics "github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod"
	"github.com/go-logr/logr"
	"k8s.io/api/core/v1"
	runtimecontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const Name = "csa"

var onceInstance sync.Once
var instance *controller

// controller represents the CSA controller itself.
type controller struct {
	controllerConfig controllercommon.ControllerConfig
	runtimeManager   manager.Manager

	onceInit sync.Once
}

func NewController(
	controllerConfig controllercommon.ControllerConfig,
	runtimeManager manager.Manager,
) *controller {
	onceInstance.Do(func() {
		instance = &controller{
			controllerConfig: controllerConfig,
			runtimeManager:   runtimeManager,
		}
	})

	return instance
}

// Initialize performs the tasks necessary to initialize the controller and register it with the controller-runtime
// manager. Will only be invoked once. runtimeController parameter is provided for test injection.
func (c *controller) Initialize(runtimeController ...runtimecontroller.Controller) error {
	var retErr error

	c.onceInit.Do(func() {
		reconciler := newContainerStartupAutoscalerReconciler(
			pod.NewPod(c.controllerConfig, c.runtimeManager.GetClient(), c.runtimeManager.GetEventRecorderFor(Name)),
			c.controllerConfig,
		)

		var actualRuntimeController runtimecontroller.Controller

		if len(runtimeController) == 0 {
			var err error
			actualRuntimeController, err = runtimecontroller.New(
				Name,
				c.runtimeManager,
				runtimecontroller.Options{
					MaxConcurrentReconciles: c.controllerConfig.MaxConcurrentReconciles,
					Reconciler:              reconciler,
					LogConstructor: func(req *reconcile.Request) logr.Logger {
						log := logging.Logger
						log = log.WithValues("controller", Name)

						if req != nil {
							log = log.WithValues(
								"podNamespace", req.Namespace,
								"name", req.Name,
							)
						}

						return log
					},
				},
			)
			if err != nil {
				retErr = common.WrapErrorf(err, "unable to create controller-runtime controller")
				return
			}
		} else {
			actualRuntimeController = runtimeController[0]
		}

		// Predicates are employed to filter out pod changes that are not necessary to reconcile.
		if err := actualRuntimeController.Watch(
			source.Kind(
				c.runtimeManager.GetCache(),
				&v1.Pod{},
				&handler.TypedEnqueueRequestForObject[*v1.Pod]{},
				predicate.TypedFuncs[*v1.Pod]{
					CreateFunc:  PredicateCreateFunc,
					DeleteFunc:  PredicateDeleteFunc,
					UpdateFunc:  PredicateUpdateFunc,
					GenericFunc: PredicateGenericFunc,
				},
			),
		); err != nil {
			retErr = common.WrapErrorf(err, "unable to watch pods")
			return
		}

		csametrics.RegisterAllMetrics(metrics.Registry)
	})

	return retErr
}
