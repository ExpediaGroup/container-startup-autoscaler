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

package controllercommon

import (
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/spf13/cobra"
)

const (
	flagKubeConfigName    = "kubeconfig"
	flagKubeConfigDesc    = "absolute path to the cluster kubeconfig file (uses in-cluster configuration if not supplied)"
	flagKubeConfigDefault = ""

	flagLeaderElectionEnabledName    = "leader-election-enabled"
	flagLeaderElectionEnabledDesc    = "whether to enable leader election"
	flagLeaderElectionEnabledDefault = true

	flagLeaderElectionResourceNamespaceName    = "leader-election-resource-namespace"
	flagLeaderElectionResourceNamespaceDesc    = "the namespace to create resources in if leader election is enabled (uses current namespace if not supplied)"
	flagLeaderElectionResourceNamespaceDefault = ""

	flagCacheSyncPeriodMinsName    = "cache-sync-period-mins"
	flagCacheSyncPeriodMinsDesc    = "how frequently the informer should re-sync"
	flagCacheSyncPeriodMinsDefault = 60

	flagGracefulShutdownTimeoutSecsName    = "graceful-shutdown-timeout-secs"
	flagGracefulShutdownTimeoutSecsDesc    = "how long to allow busy workers to complete upon shutdown"
	flagGracefulShutdownTimeoutSecsDefault = 10

	flagRequeueDurationSecsName    = "requeue-duration-secs"
	flagRequeueDurationSecsDesc    = "how long to wait before requeuing a reconcile"
	flagRequeueDurationSecsDefault = 1

	flagMaxConcurrentReconcilesName    = "max-concurrent-reconciles"
	flagMaxConcurrentReconcilesDesc    = "the maximum number of concurrent reconciles"
	flagMaxConcurrentReconcilesDefault = 10

	flagStandardRetryAttemptsName    = "standard-retry-attempts"
	flagStandardRetryAttemptsDesc    = "the maximum number of attempts for a standard retry"
	flagStandardRetryAttemptsDefault = 3

	flagStandardRetryDelaySecsName    = "standard-retry-delay-secs"
	flagStandardRetryDelaySecsDesc    = "the number of seconds to wait between standard retry attempts"
	flagStandardRetryDelaySecsDefault = 1

	flagScaleWhenUnknownResourcesName    = "scale-when-unknown-resources"
	flagScaleWhenUnknownResourcesDesc    = "whether to scale when unknown resources (i.e. other than those specified within annotations) are encountered"
	flagScaleWhenUnknownResourcesDefault = false

	flagLogVName    = "log-v"
	flagLogVDesc    = "log verbosity level (0: info, 1: debug, 2: trace) - 2 used if invalid"
	flagLogVDefault = 0

	flagLogAddCallerName    = "log-add-caller"
	flagLogAddCallerDesc    = "whether to include the caller within logging output"
	flagLogAddCallerDefault = false
)

// ControllerConfig represents the configuration of the CSA controller.
type ControllerConfig struct {
	KubeConfig                      string
	LeaderElectionEnabled           bool
	LeaderElectionResourceNamespace string

	CacheSyncPeriodMins         int
	GracefulShutdownTimeoutSecs int
	RequeueDurationSecs         int
	MaxConcurrentReconciles     int
	StandardRetryAttempts       int
	StandardRetryDelaySecs      int
	ScaleWhenUnknownResources   bool
	LogV                        int
	LogAddCaller                bool

	BindAddressMetrics string
	BindAddressProbes  string
	BindAddressPprof   string
}

func NewControllerConfig() ControllerConfig {
	return ControllerConfig{
		BindAddressMetrics: ":8080",
		BindAddressProbes:  ":8081",
		BindAddressPprof:   ":8082",
	}
}

// InitFlags defines Cobra flags for the CSA controller.
func (c *ControllerConfig) InitFlags(command *cobra.Command) {
	command.Flags().StringVar(
		&c.KubeConfig,
		flagKubeConfigName, flagKubeConfigDefault, flagKubeConfigDesc,
	)

	command.Flags().BoolVar(
		&c.LeaderElectionEnabled,
		flagLeaderElectionEnabledName, flagLeaderElectionEnabledDefault, flagLeaderElectionEnabledDesc,
	)

	command.Flags().StringVar(
		&c.LeaderElectionResourceNamespace,
		flagLeaderElectionResourceNamespaceName, flagLeaderElectionResourceNamespaceDefault, flagLeaderElectionResourceNamespaceDesc,
	)

	command.Flags().IntVar(
		&c.CacheSyncPeriodMins,
		flagCacheSyncPeriodMinsName, flagCacheSyncPeriodMinsDefault, flagCacheSyncPeriodMinsDesc,
	)

	command.Flags().IntVar(
		&c.GracefulShutdownTimeoutSecs,
		flagGracefulShutdownTimeoutSecsName, flagGracefulShutdownTimeoutSecsDefault, flagGracefulShutdownTimeoutSecsDesc,
	)

	command.Flags().IntVar(
		&c.RequeueDurationSecs,
		flagRequeueDurationSecsName, flagRequeueDurationSecsDefault, flagRequeueDurationSecsDesc,
	)

	command.Flags().IntVar(
		&c.MaxConcurrentReconciles,
		flagMaxConcurrentReconcilesName, flagMaxConcurrentReconcilesDefault, flagMaxConcurrentReconcilesDesc,
	)

	command.Flags().IntVar(
		&c.StandardRetryAttempts,
		flagStandardRetryAttemptsName, flagStandardRetryAttemptsDefault, flagStandardRetryAttemptsDesc,
	)

	command.Flags().IntVar(
		&c.StandardRetryDelaySecs,
		flagStandardRetryDelaySecsName, flagStandardRetryDelaySecsDefault, flagStandardRetryDelaySecsDesc,
	)

	command.Flags().BoolVar(
		&c.ScaleWhenUnknownResources,
		flagScaleWhenUnknownResourcesName, flagScaleWhenUnknownResourcesDefault, flagScaleWhenUnknownResourcesDesc,
	)

	command.Flags().IntVar(
		&c.LogV,
		flagLogVName, flagLogVDefault, flagLogVDesc,
	)

	command.Flags().BoolVar(
		&c.LogAddCaller,
		flagLogAddCallerName, flagLogAddCallerDefault, flagLogAddCallerDesc,
	)
}

// Log logs the configuration of the CSA controller.
func (c *ControllerConfig) Log() {
	logging.Infof(nil, logging.VInfo, "(config) %s: %s", flagKubeConfigName, c.KubeConfig)
	logging.Infof(nil, logging.VInfo, "(config) %s: %t", flagLeaderElectionEnabledName, c.LeaderElectionEnabled)
	logging.Infof(nil, logging.VInfo, "(config) %s: %s", flagLeaderElectionResourceNamespaceName, c.LeaderElectionResourceNamespace)
	logging.Infof(nil, logging.VInfo, "(config) %s: %d", flagCacheSyncPeriodMinsName, c.CacheSyncPeriodMins)
	logging.Infof(nil, logging.VInfo, "(config) %s: %d", flagGracefulShutdownTimeoutSecsName, c.GracefulShutdownTimeoutSecs)
	logging.Infof(nil, logging.VInfo, "(config) %s: %d", flagRequeueDurationSecsName, c.RequeueDurationSecs)
	logging.Infof(nil, logging.VInfo, "(config) %s: %d", flagMaxConcurrentReconcilesName, c.MaxConcurrentReconciles)
	logging.Infof(nil, logging.VInfo, "(config) %s: %d", flagStandardRetryAttemptsName, c.StandardRetryAttempts)
	logging.Infof(nil, logging.VInfo, "(config) %s: %d", flagStandardRetryDelaySecsName, c.StandardRetryDelaySecs)
	logging.Infof(nil, logging.VInfo, "(config) %s: %t", flagScaleWhenUnknownResourcesName, c.ScaleWhenUnknownResources)
	logging.Infof(nil, logging.VInfo, "(config) %s: %d", flagLogVName, c.LogV)
	logging.Infof(nil, logging.VInfo, "(config) %s: %t", flagLogAddCallerName, c.LogAddCaller)
}

// CacheSyncPeriodMinsDuration returns the cache sync period in minutes as a time.Duration.
func (c *ControllerConfig) CacheSyncPeriodMinsDuration() time.Duration {
	return time.Duration(c.CacheSyncPeriodMins) * time.Minute
}

// GracefulShutdownTimeoutSecsDuration returns the graceful shutdown timeout in seconds as a time.Duration.
func (c *ControllerConfig) GracefulShutdownTimeoutSecsDuration() time.Duration {
	return time.Duration(c.GracefulShutdownTimeoutSecs) * time.Second
}

// RequeueDurationSecsDuration returns the requeue duration in seconds as a time.Duration.
func (c *ControllerConfig) RequeueDurationSecsDuration() time.Duration {
	return time.Duration(c.RequeueDurationSecs) * time.Second
}
