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

import "os"

// Path ----------------------------------------------------------------------------------------------------------------
const (
	pathSeparator        = string(os.PathSeparator)
	pathIntTestRelPath   = "test" + pathSeparator + "integration"
	pathConfigDirRelPath = pathIntTestRelPath + pathSeparator + "config"
)

// kind ----------------------------------------------------------------------------------------------------------------
const (
	kindClusterName       = "csa-int-cluster"
	kindConfigFileRelPath = pathConfigDirRelPath + pathSeparator + "kind.yaml"
)

var kubeVersionToFullVersion = map[string]string{
	"1.32": "v1.32.0",
	// Older versions are not supported by 'kind build node-image' as the server tgzs don't include the 'version' file
	// and fail.
}

// metrics-server ------------------------------------------------------------------------------------------------------
const (
	metricsServerImageTag            = "registry.k8s.io/metrics-server/metrics-server:v0.6.4"
	metricsServerKustomizeDirRelPath = pathConfigDirRelPath + pathSeparator + "metricsserver"
	metricsServerReadyTimeout        = "60s"
)

// CSA -----------------------------------------------------------------------------------------------------------------
const (
	csaDockerImage    = "csa"
	csaDockerTag      = "test"
	csaDockerImageTag = csaDockerImage + ":" + csaDockerTag
)

const (
	csaHelmChartRelPath = "charts" + pathSeparator + "container-startup-autoscaler"
	csaHelmName         = "csa-int"
	csaHelmTimeout      = "60s"
)

const (
	csaStatusWaitMillis                            = 1000
	csaStatusMessageStartupCommanded               = "Startup resources commanded"
	csaStatusMessageStartupCommandedUnknownRes     = "Startup resources commanded (unknown resources applied)"
	csaStatusMessagePostStartupCommanded           = "Post-startup resources commanded"
	csaStatusMessagePostStartupCommandedUnknownRes = "Post-startup resources commanded (unknown resources applied)"
	csaStatusMessageStartupEnacted                 = "Startup resources enacted"
	csaStatusMessagePostStartupEnacted             = "Post-startup resources enacted"
	csaStatusMessageValidationError                = "Validation error"
)

const (
	csaEventReasonScaling    = "Scaling"
	csaEventReasonValidation = "Validation"
)

// echo-server ---------------------------------------------------------------------------------------------------------
const (
	echoServerDockerImageTag = "ealen/echo-server:0.7.0"
	echoServerName           = "echo-server"
)

const (
	echoServerNonTargetContainerName           = echoServerName + "-non-target"
	echoServerNonTargetContainerCpuRequests    = "150m"
	echoServerNonTargetContainerCpuLimits      = "150m"
	echoServerNonTargetContainerMemoryRequests = "150M"
	echoServerNonTargetContainerMemoryLimits   = "150M"
)

const (
	echoServerCpuDisabledRequests    = echoServerNonTargetContainerCpuRequests
	echoServerCpuDisabledLimits      = echoServerNonTargetContainerCpuLimits
	echoServerMemoryDisabledRequests = echoServerNonTargetContainerMemoryRequests
	echoServerMemoryDisabledLimits   = echoServerNonTargetContainerMemoryLimits
)

const (
	echoServerDefaultProbeInitialDelaySeconds = 15
	echoServerProbePeriodSeconds              = 1
	echoServerProbeFailureThreshold           = echoServerDefaultProbeInitialDelaySeconds
)

// Annotations ---------------------------------------------------------------------------------------------------------
var (
	csaQuantityAnnotationsAllDefault = csaQuantityAnnotations{
		cpuStartup:                "150m",
		cpuPostStartupRequests:    "100m",
		cpuPostStartupLimits:      "100m",
		memoryStartup:             "150M",
		memoryPostStartupRequests: "100M",
		memoryPostStartupLimits:   "100M",
	}

	csaQuantityAnnotationsCpuOnlyDefault = csaQuantityAnnotations{
		cpuStartup:             "150m",
		cpuPostStartupRequests: "100m",
		cpuPostStartupLimits:   "100m",
	}

	csaQuantityAnnotationsMemoryOnlyDefault = csaQuantityAnnotations{
		memoryStartup:             "150M",
		memoryPostStartupRequests: "100M",
		memoryPostStartupLimits:   "100M",
	}
)

// Tests ---------------------------------------------------------------------------------------------------------------
const (
	testsDefaultWaitStatusTimeoutSecs = echoServerDefaultProbeInitialDelaySeconds * 2
)
