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

package integration

import (
	"fmt"
	"os"
	"strconv"
)

type suppliedConfigStruct struct {
	kubeVersion          string
	maxParallelism       string
	extraCaCertPath      string
	reuseCluster         bool
	installMetricsServer bool
	keepCsa              bool
	keepCluster          bool
	deleteNsPostTest     bool
}

var suppliedConfig = suppliedConfigStruct{
	kubeVersion:          "",
	maxParallelism:       "5",
	extraCaCertPath:      "",
	reuseCluster:         false,
	installMetricsServer: false,
	keepCsa:              false,
	keepCluster:          false,
	deleteNsPostTest:     true,
}

func suppliedConfigInit() {
	suppliedConfigSetString("KUBE_VERSION", &suppliedConfig.kubeVersion, true)
	suppliedConfigSetString("MAX_PARALLELISM", &suppliedConfig.maxParallelism, true)
	suppliedConfigSetString("EXTRA_CA_CERT_PATH", &suppliedConfig.extraCaCertPath, false)
	suppliedConfigSetBool("REUSE_CLUSTER", &suppliedConfig.reuseCluster, true)
	suppliedConfigSetBool("INSTALL_METRICS_SERVER", &suppliedConfig.installMetricsServer, true)
	suppliedConfigSetBool("KEEP_CSA", &suppliedConfig.keepCsa, true)
	suppliedConfigSetBool("KEEP_CLUSTER", &suppliedConfig.keepCluster, true)
	suppliedConfigSetBool("DELETE_NS_AFTER_TEST", &suppliedConfig.deleteNsPostTest, true)

	logMessage(nil, fmt.Sprintf("(config) KUBE_VERSION: "+suppliedConfig.kubeVersion))
	logMessage(nil, fmt.Sprintf("(config) MAX_PARALLELISM: "+suppliedConfig.maxParallelism))
	logMessage(nil, fmt.Sprintf("(config) EXTRA_CA_CERT_PATH: "+suppliedConfig.extraCaCertPath))
	logMessage(nil, fmt.Sprintf("(config) REUSE_CLUSTER: %t", suppliedConfig.reuseCluster))
	logMessage(nil, fmt.Sprintf("(config) INSTALL_METRICS_SERVER: %t", suppliedConfig.installMetricsServer))
	logMessage(nil, fmt.Sprintf("(config) KEEP_CSA: %t", suppliedConfig.keepCsa))
	logMessage(nil, fmt.Sprintf("(config) KEEP_CLUSTER: %t", suppliedConfig.keepCluster))
	logMessage(nil, fmt.Sprintf("(config) DELETE_NS_AFTER_TEST: %t", suppliedConfig.deleteNsPostTest))
}

func suppliedConfigSetString(env string, config *string, required bool) {
	envVal := os.Getenv(env)
	haveEnvOrDefault := envVal != "" || (config != nil && *config != "")

	if !haveEnvOrDefault && required {
		logMessage(nil, fmt.Sprintf("(config) '%s' value is required", env))
		os.Exit(1)
	}

	if envVal != "" {
		*config = envVal
	}
}

func suppliedConfigSetBool(env string, config *bool, required bool) {
	envVal := os.Getenv(env)
	haveEnvOrDefault := envVal != "" || config != nil

	if !haveEnvOrDefault && required {
		logMessage(nil, fmt.Sprintf("(config) '%s' value is required", env))
		os.Exit(1)
	}

	if envVal != "" {
		var err error
		*config, err = strconv.ParseBool(envVal)
		if err != nil {
			logMessage(nil, fmt.Sprintf("(config) '%s' value is not a bool", env))
			os.Exit(1)
		}
	}
}
