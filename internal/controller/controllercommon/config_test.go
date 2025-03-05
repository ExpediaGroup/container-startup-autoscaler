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
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewControllerConfig(t *testing.T) {
	config := NewControllerConfig()
	assert.NotEmpty(t, config.BindAddressMetrics)
	assert.NotEmpty(t, config.BindAddressProbes)
	assert.NotEmpty(t, config.BindAddressPprof)
}

func TestControllerConfigInitFlags(t *testing.T) {
	t.Run("AllDefaults", func(t *testing.T) {
		config := ControllerConfig{}
		cmd := &cobra.Command{
			Run: func(_ *cobra.Command, _ []string) {
				assert.Equal(t, FlagKubeConfigDefault, config.KubeConfig)
				assert.Equal(t, FlagLeaderElectionEnabledDefault, config.LeaderElectionEnabled)
				assert.Equal(t, FlagLeaderElectionResourceNamespaceDefault, config.LeaderElectionResourceNamespace)
				assert.Equal(t, FlagCacheSyncPeriodMinsDefault, config.CacheSyncPeriodMins)
				assert.Equal(t, FlagGracefulShutdownTimeoutSecsDefault, config.GracefulShutdownTimeoutSecs)
				assert.Equal(t, FlagRequeueDurationSecsDefault, config.RequeueDurationSecs)
				assert.Equal(t, FlagMaxConcurrentReconcilesDefault, config.MaxConcurrentReconciles)
				assert.Equal(t, FlagStandardRetryAttemptsDefault, config.StandardRetryAttempts)
				assert.Equal(t, FlagStandardRetryDelaySecsDefault, config.StandardRetryDelaySecs)
				assert.Equal(t, FlagLogVDefault, config.LogV)
				assert.Equal(t, FlagLogAddCallerDefault, config.LogAddCaller)
			},
		}
		config.InitFlags(cmd)
		_ = cmd.Execute()
	})

	t.Run("OneSet", func(t *testing.T) {
		config := ControllerConfig{}
		cmd := &cobra.Command{
			Run: func(_ *cobra.Command, _ []string) {
				assert.Equal(t, "test", config.KubeConfig)
			},
		}
		config.InitFlags(cmd)
		cmd.SetArgs([]string{
			fmt.Sprintf("--%s=test", FlagKubeConfigName),
		})
		_ = cmd.Execute()
	})
}

func TestControllerConfigLog(t *testing.T) {
	buffer := &bytes.Buffer{}
	logging.Init(buffer, logging.VInfo, false)
	config := ControllerConfig{}
	cmd := &cobra.Command{
		Run: func(_ *cobra.Command, _ []string) {
			assert.Equal(t, 12, strings.Count(buffer.String(), "\n"))
		},
	}
	config.Log()
	_ = cmd.Execute()
}

func TestControllerConfigCacheSyncPeriodMinsDuration(t *testing.T) {
	config := ControllerConfig{CacheSyncPeriodMins: 1}
	assert.Equal(t, 1*time.Minute, config.CacheSyncPeriodMinsDuration())
}

func TestControllerConfigGracefulShutdownTimeoutSecsDuration(t *testing.T) {
	config := ControllerConfig{GracefulShutdownTimeoutSecs: 1}
	assert.Equal(t, 1*time.Second, config.GracefulShutdownTimeoutSecsDuration())
}

func TestControllerConfigRequeueDurationSecsDuration(t *testing.T) {
	config := ControllerConfig{RequeueDurationSecs: 1}
	assert.Equal(t, 1*time.Second, config.RequeueDurationSecsDuration())
}
