package integration

import "os"

// Path
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
	"1.29": {
		"amd64": "kindest/node:v1.29.8@sha256:b69a150f9951ef41158ec76de381a920df2be3582fd16fc19cf4757eef0dded9",
		"arm64": "kindest/node:v1.29.8@sha256:0d5623800cf6290edbc1007ca8a33a5f7e2ad92b41dc7022b4d20a66447db23c",
	},
	"1.30": {
		"amd64": "kindest/node:v1.30.4@sha256:34cb98a38a57a3357fde925a41d61232bbbbeb411b45a25c0d766635d6c3b975",
		"arm64": "kindest/node:v1.30.4@sha256:6becd630a18e77730e31f3833f0b129bbcc9c09ee49c3b88429b3c1fdc30bfc4",
	},
	"1.31": {
		"amd64": "kindest/node:v1.31.0@sha256:919a65376fd11b67df05caa2e60802ad5de2fca250c9fe0c55b0dce5c9591af3",
		"arm64": "kindest/node:v1.31.0@sha256:0ccfb11dc66eae4abc20c30ee95687bab51de8aeb04e325e1c49af0890646548",
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
	csaStatusWaitMillis                            = 500
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
	echoServerDefaultProbeInitialDelaySeconds  = 120 // TODO(wt) enacting resources can sometimes take Kube upwards of 90s (Kube 1.29). Reduce this when addressed.
)

// Tests ---------------------------------------------------------------------------------------------------------------
const (
	testsDefaultWaitStatusTimeoutSecs = echoServerDefaultProbeInitialDelaySeconds + 30
)
