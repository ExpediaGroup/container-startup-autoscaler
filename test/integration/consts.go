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

var k8sVersionToImage = map[string]map[string]string{
	"1.31": {
		"amd64": "kindest/node:v1.31.0@sha256:919a65376fd11b67df05caa2e60802ad5de2fca250c9fe0c55b0dce5c9591af3",
		"arm64": "kindest/node:v1.31.0@sha256:0ccfb11dc66eae4abc20c30ee95687bab51de8aeb04e325e1c49af0890646548",
	},
	"1.30": {
		"amd64": "kindest/node:v1.30.4@sha256:34cb98a38a57a3357fde925a41d61232bbbbeb411b45a25c0d766635d6c3b975",
		"arm64": "kindest/node:v1.30.4@sha256:6becd630a18e77730e31f3833f0b129bbcc9c09ee49c3b88429b3c1fdc30bfc4",
	},
	"1.29": {
		"amd64": "kindest/node:v1.29.8@sha256:b69a150f9951ef41158ec76de381a920df2be3582fd16fc19cf4757eef0dded9",
		"arm64": "kindest/node:v1.29.8@sha256:0d5623800cf6290edbc1007ca8a33a5f7e2ad92b41dc7022b4d20a66447db23c",
	},
	"1.28": {
		"amd64": "kindest/node:v1.28.13@sha256:d97df9fff48099bf9a94c92fdc39adde65bec2aa1d011f84233b96172c1003c9",
		"arm64": "kindest/node:v1.28.13@sha256:ddef612bb93a9aa3a989f9d3d4e01c0a7c4d866a4b949264146c182cd202d738",
	},
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
	echoServerNonTargetContainerCpuRequests    = "50m"
	echoServerNonTargetContainerCpuLimits      = "50m"
	echoServerNonTargetContainerMemoryRequests = "150M"
	echoServerNonTargetContainerMemoryLimits   = "150M"
)

const (
	echoServerDefaultProbeInitialDelaySeconds = 15
	echoServerProbePeriodSeconds              = 1
	echoServerProbeFailureThreshold           = echoServerDefaultProbeInitialDelaySeconds
)

// Tests ---------------------------------------------------------------------------------------------------------------
const (
	testsDefaultWaitStatusTimeoutSecs = echoServerDefaultProbeInitialDelaySeconds * 2
)
