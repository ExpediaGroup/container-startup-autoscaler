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

// TODO(wt) Remaining for Kube 1.33:
// TODO(wt) Integration tests to include infeasible test (ridiculously large startup CPU), plus any others
// TODO(wt) Ensure docs state that target containers must explicitly state CPU/memory requests/limits
// TODO(wt) Ensure docs state 'last' timestamp behavior
// TODO(wt) Add note to docs around swap-enabled containers: https://github.com/kubernetes/kubernetes/pull/130831
// TODO(wt) Add note to docs around https://kubernetes.io/docs/tasks/configure-pod-container/assign-pod-level-resources/ - not supported by CSA (is it even supported by in-place?)
// TODO(wt) Need support for InPlacePodVerticalScalingExclusiveCPUs?
// TODO(wt) FYI: support has been for in-place vertical scaling of pods with sidecars
// TODO(wt) Update docs

package main

import (
	"os"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/controller"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/controller/controllercommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var controllerConfig = controllercommon.NewControllerConfig()

// main is the controller entry point. It configures flags for, and executes, the root command.
func main() {
	rootCmd := cobra.Command{
		Short: "Kubernetes container startup autoscaler",
		Run:   run,
	}

	controllerConfig.InitFlags(&rootCmd)
	if err := rootCmd.Execute(); err != nil {
		logging.Fatalf(nil, err, "unable to execute root command")
	}
}

// run is the root command work function. It obtains configuration, configures the controller-runtime manager,
// initializes the CSA controller and starts the controller-runtime manager.
func run(_ *cobra.Command, _ []string) {
	level := controllerConfig.LogV
	if level < 0 || level > 2 {
		level = int(logging.DefaultV)
	}
	logging.Init(logging.DefaultW, logging.V(level), controllerConfig.LogAddCaller)
	logging.Infof(nil, logging.VInfo, "starting %s...", controller.Name)
	controllerConfig.Log()

	if controllerConfig.KubeConfig != "" {
		if err := os.Setenv("KUBECONFIG", controllerConfig.KubeConfig); err != nil {
			logging.Fatalf(nil, err, "unable to set KUBECONFIG environment variable")
		}
	}

	cacheSyncPeriod := controllerConfig.CacheSyncPeriodMinsDuration()
	gracefulShutdownTimeout := controllerConfig.GracefulShutdownTimeoutSecsDuration()

	options := manager.Options{
		Cache: cache.Options{
			SyncPeriod: &cacheSyncPeriod,
			ByObject: map[client.Object]cache.ByObject{
				&v1.Pod{}: {
					// Restrict caching to pods that have enabled label to avoid caching everything.
					Label: labels.SelectorFromSet(labels.Set{kubecommon.LabelEnabled: "true"}),
				},
			},
		},
		GracefulShutdownTimeout: &gracefulShutdownTimeout,
		Logger:                  logging.Logger,
		Metrics:                 metricsserver.Options{BindAddress: controllerConfig.BindAddressMetrics},
		HealthProbeBindAddress:  controllerConfig.BindAddressProbes,
		PprofBindAddress:        controllerConfig.BindAddressPprof,
		LeaderElection:          controllerConfig.LeaderElectionEnabled,
		LeaderElectionNamespace: controllerConfig.LeaderElectionResourceNamespace,
		LeaderElectionID:        "csa-expediagroup-com",
	}

	// Uses KUBECONFIG env var if set, otherwise tries in-cluster config.
	restConfig, err := config.GetConfig()
	if err != nil {
		logging.Fatalf(nil, err, "unable to get rest config")
	}

	runtimeManager, err := manager.New(restConfig, options)
	if err != nil {
		logging.Fatalf(nil, err, "unable to create controller-runtime manager")
	}

	if err = runtimeManager.AddHealthzCheck("liveness", healthz.Ping); err != nil {
		logging.Fatalf(nil, err, "unable to add healthz check")
	}

	csaController := controller.NewController(controllerConfig, runtimeManager)

	if err = csaController.Initialize(); err != nil {
		logging.Fatalf(nil, err, "unable to initialize controller")
	}

	// Blocks.
	if err = runtimeManager.Start(signals.SetupSignalHandler()); err != nil {
		logging.Fatalf(nil, err, "unable to start controller-runtime manager")
	}
}
