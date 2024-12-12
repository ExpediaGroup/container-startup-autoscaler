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

package controllercommon

import (
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/spf13/cobra"
)

const (
	FlagKubeConfigName    = "kubeconfig"
	FlagKubeConfigDesc    = "absolute path to the cluster kubeconfig file (uses in-cluster configuration if not supplied)"
	FlagKubeConfigDefault = ""

	FlagLeaderElectionEnabledName    = "leader-election-enabled"
	FlagLeaderElectionEnabledDesc    = "whether to enable leader election"
	FlagLeaderElectionEnabledDefault = true

	FlagLeaderElectionResourceNamespaceName    = "leader-election-resource-namespace"
	FlagLeaderElectionResourceNamespaceDesc    = "the namespace to create resources in if leader election is enabled (uses current namespace if not supplied)"
	FlagLeaderElectionResourceNamespaceDefault = ""

	FlagCacheSyncPeriodMinsName    = "cache-sync-period-mins"
	FlagCacheSyncPeriodMinsDesc    = "how frequently the informer should re-sync"
	FlagCacheSyncPeriodMinsDefault = 60

	FlagGracefulShutdownTimeoutSecsName    = "graceful-shutdown-timeout-secs"
	FlagGracefulShutdownTimeoutSecsDesc    = "how long to allow busy workers to complete upon shutdown"
	FlagGracefulShutdownTimeoutSecsDefault = 10

	FlagRequeueDurationSecsName    = "requeue-duration-secs"
	FlagRequeueDurationSecsDesc    = "how long to wait before requeuing a reconcile"
	FlagRequeueDurationSecsDefault = 1

	FlagMaxConcurrentReconcilesName    = "max-concurrent-reconciles"
	FlagMaxConcurrentReconcilesDesc    = "the maximum number of concurrent reconciles"
	FlagMaxConcurrentReconcilesDefault = 10

	FlagStandardRetryAttemptsName    = "standard-retry-attempts"
	FlagStandardRetryAttemptsDesc    = "the maximum number of attempts for a standard retry"
	FlagStandardRetryAttemptsDefault = 3

	FlagStandardRetryDelaySecsName    = "standard-retry-delay-secs"
	FlagStandardRetryDelaySecsDesc    = "the number of seconds to wait between standard retry attempts"
	FlagStandardRetryDelaySecsDefault = 1

	FlagScaleWhenUnknownResourcesName    = "scale-when-unknown-resources"
	FlagScaleWhenUnknownResourcesDesc    = "whether to scale when unknown resources (i.e. other than those specified within annotations) are encountered"
	FlagScaleWhenUnknownResourcesDefault = false

	FlagLogVName    = "log-v"
	FlagLogVDesc    = "log verbosity level (0: info, 1: debug, 2: trace) - 2 used if invalid"
	FlagLogVDefault = 0

	FlagLogAddCallerName    = "log-add-caller"
	FlagLogAddCallerDesc    = "whether to include the caller within logging output"
	FlagLogAddCallerDefault = false
)

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

func (c *ControllerConfig) InitFlags(command *cobra.Command) {
	command.Flags().StringVar(
		&c.KubeConfig,
		FlagKubeConfigName, FlagKubeConfigDefault, FlagKubeConfigDesc,
	)

	command.Flags().BoolVar(
		&c.LeaderElectionEnabled,
		FlagLeaderElectionEnabledName, FlagLeaderElectionEnabledDefault, FlagLeaderElectionEnabledDesc,
	)

	command.Flags().StringVar(
		&c.LeaderElectionResourceNamespace,
		FlagLeaderElectionResourceNamespaceName, FlagLeaderElectionResourceNamespaceDefault, FlagLeaderElectionResourceNamespaceDesc,
	)

	command.Flags().IntVar(
		&c.CacheSyncPeriodMins,
		FlagCacheSyncPeriodMinsName, FlagCacheSyncPeriodMinsDefault, FlagCacheSyncPeriodMinsDesc,
	)

	command.Flags().IntVar(
		&c.GracefulShutdownTimeoutSecs,
		FlagGracefulShutdownTimeoutSecsName, FlagGracefulShutdownTimeoutSecsDefault, FlagGracefulShutdownTimeoutSecsDesc,
	)

	command.Flags().IntVar(
		&c.RequeueDurationSecs,
		FlagRequeueDurationSecsName, FlagRequeueDurationSecsDefault, FlagRequeueDurationSecsDesc,
	)

	command.Flags().IntVar(
		&c.MaxConcurrentReconciles,
		FlagMaxConcurrentReconcilesName, FlagMaxConcurrentReconcilesDefault, FlagMaxConcurrentReconcilesDesc,
	)

	command.Flags().IntVar(
		&c.StandardRetryAttempts,
		FlagStandardRetryAttemptsName, FlagStandardRetryAttemptsDefault, FlagStandardRetryAttemptsDesc,
	)

	command.Flags().IntVar(
		&c.StandardRetryDelaySecs,
		FlagStandardRetryDelaySecsName, FlagStandardRetryDelaySecsDefault, FlagStandardRetryDelaySecsDesc,
	)

	command.Flags().BoolVar(
		&c.ScaleWhenUnknownResources,
		FlagScaleWhenUnknownResourcesName, FlagScaleWhenUnknownResourcesDefault, FlagScaleWhenUnknownResourcesDesc,
	)

	command.Flags().IntVar(
		&c.LogV,
		FlagLogVName, FlagLogVDefault, FlagLogVDesc,
	)

	command.Flags().BoolVar(
		&c.LogAddCaller,
		FlagLogAddCallerName, FlagLogAddCallerDefault, FlagLogAddCallerDesc,
	)
}

func (c *ControllerConfig) Log() {
	logging.Infof(nil, logging.VInfo, "(config) %s: %s", FlagKubeConfigName, c.KubeConfig)
	logging.Infof(nil, logging.VInfo, "(config) %s: %t", FlagLeaderElectionEnabledName, c.LeaderElectionEnabled)
	logging.Infof(nil, logging.VInfo, "(config) %s: %s", FlagLeaderElectionResourceNamespaceName, c.LeaderElectionResourceNamespace)
	logging.Infof(nil, logging.VInfo, "(config) %s: %d", FlagCacheSyncPeriodMinsName, c.CacheSyncPeriodMins)
	logging.Infof(nil, logging.VInfo, "(config) %s: %d", FlagGracefulShutdownTimeoutSecsName, c.GracefulShutdownTimeoutSecs)
	logging.Infof(nil, logging.VInfo, "(config) %s: %d", FlagRequeueDurationSecsName, c.RequeueDurationSecs)
	logging.Infof(nil, logging.VInfo, "(config) %s: %d", FlagMaxConcurrentReconcilesName, c.MaxConcurrentReconciles)
	logging.Infof(nil, logging.VInfo, "(config) %s: %d", FlagStandardRetryAttemptsName, c.StandardRetryAttempts)
	logging.Infof(nil, logging.VInfo, "(config) %s: %d", FlagStandardRetryDelaySecsName, c.StandardRetryDelaySecs)
	logging.Infof(nil, logging.VInfo, "(config) %s: %t", FlagScaleWhenUnknownResourcesName, c.ScaleWhenUnknownResources)
	logging.Infof(nil, logging.VInfo, "(config) %s: %d", FlagLogVName, c.LogV)
	logging.Infof(nil, logging.VInfo, "(config) %s: %t", FlagLogAddCallerName, c.LogAddCaller)
}

func (c *ControllerConfig) CacheSyncPeriodMinsDuration() time.Duration {
	return time.Duration(c.CacheSyncPeriodMins) * time.Minute
}

func (c *ControllerConfig) GracefulShutdownTimeoutSecsDuration() time.Duration {
	return time.Duration(c.GracefulShutdownTimeoutSecs) * time.Second
}

func (c *ControllerConfig) RequeueDurationSecsDuration() time.Duration {
	return time.Duration(c.RequeueDurationSecs) * time.Second
}
