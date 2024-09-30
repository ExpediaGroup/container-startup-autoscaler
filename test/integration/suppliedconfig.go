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

package integration

import (
	"fmt"
	"os"
	"strconv"
)

type suppliedConfigStruct struct {
	kubeVersion          string
	maxParallelism       string
	reuseCluster         bool
	installMetricsServer bool
	keepCsa              bool
	keepCluster          bool
	deleteNsPostTest     bool
}

var suppliedConfig = suppliedConfigStruct{
	kubeVersion:          "",
	maxParallelism:       "4",
	reuseCluster:         false,
	installMetricsServer: false,
	keepCsa:              false,
	keepCluster:          false,
	deleteNsPostTest:     true,
}

func suppliedConfigInit() {
	suppliedConfigSetString("KUBE_VERSION", &suppliedConfig.kubeVersion)
	suppliedConfigSetString("MAX_PARALLELISM", &suppliedConfig.maxParallelism)
	suppliedConfigSetBool("REUSE_CLUSTER", &suppliedConfig.reuseCluster)
	suppliedConfigSetBool("INSTALL_METRICS_SERVER", &suppliedConfig.installMetricsServer)
	suppliedConfigSetBool("KEEP_CSA", &suppliedConfig.keepCsa)
	suppliedConfigSetBool("KEEP_CLUSTER", &suppliedConfig.keepCluster)
	suppliedConfigSetBool("DELETE_NS_AFTER_TEST", &suppliedConfig.deleteNsPostTest)

	logMessage(nil, fmt.Sprintf("(config) KUBE_VERSION: %s", suppliedConfig.kubeVersion))
	logMessage(nil, fmt.Sprintf("(config) MAX_PARALLELISM: %s", suppliedConfig.maxParallelism))
	logMessage(nil, fmt.Sprintf("(config) REUSE_CLUSTER: %t", suppliedConfig.reuseCluster))
	logMessage(nil, fmt.Sprintf("(config) INSTALL_METRICS_SERVER: %t", suppliedConfig.installMetricsServer))
	logMessage(nil, fmt.Sprintf("(config) KEEP_CSA: %t", suppliedConfig.keepCsa))
	logMessage(nil, fmt.Sprintf("(config) KEEP_CLUSTER: %t", suppliedConfig.keepCluster))
	logMessage(nil, fmt.Sprintf("(config) DELETE_NS_AFTER_TEST: %t", suppliedConfig.deleteNsPostTest))
}

func suppliedConfigSetString(env string, config *string) {
	envVal := os.Getenv(env)

	if envVal == "" && *config == "" {
		// Require env unless defaulted via supplied.
		logMessage(nil, fmt.Sprintf("(config) '%s' value is required", env))
		os.Exit(1)
	}

	if envVal != "" {
		*config = envVal
	}
}

func suppliedConfigSetBool(env string, config *bool) {
	envVal := os.Getenv(env)

	if envVal == "" && config == nil {
		// Require env unless defaulted via supplied.
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
