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
	"errors"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
)

/*

TODO(wt) 'In-place Update of Pod Resources' implementation bug (still in Kube 1.32).
Note: there currently appears to be a bug in the 'In-place Update of Pod Resources' implementation whereby successful
resizes are restarted - this is specifically mitigated against within csaWaitStatus(). This sometimes (depending on the
timing of retrieving pods via kubectl) manifested in a CSA status that (correctly) stated that the resize had occurred,
but a Status.ContainerStatuses[].Resources disparity and associated test assertion failure(s) since Kube had almost
immediately restarted the resize at the point of retrieving the pod.

Example logs of such an event (restart marked with '<-- HERE'):

{
	"level": "debug",
	"namespace": "deployment-non-startup-admitted-flow-startup-probe",
	"name": "echo-server-69cdc45777-69648",
	"reconcileID": "54f3de56-55bc-4038-8cb1-e8d180a6b9fc",
	"targetname": "echo-server",
	"targetstates": {
		"startupProbe": "true",
		"readinessProbe": "false",
		"container": "terminated",
		"started": "false",
		"ready": "false",
		"resources": "poststartup",
		"statusResources": "unknown"
	},
	"time": 1698695785143,
	"message": "target container currently not running"
}
{
	"level": "info",
	"namespace": "deployment-non-startup-admitted-flow-startup-probe",
	"name": "echo-server-69cdc45777-69648",
	"reconcileID": "7fcaac3e-25a4-47c9-86ec-68989466c3cb",
	"targetname": "echo-server",
	"targetstates": {
		"startupProbe": "true",
		"readinessProbe": "false",
		"container": "running",
		"started": "false",
		"ready": "false",
		"resources": "poststartup",
		"statusResources": "containerresourcesmatch"
	},
	"time": 1698695786101,
	"message": "startup resources commanded"
}
{
	"level": "debug",
	"namespace": "deployment-non-startup-admitted-flow-startup-probe",
	"name": "echo-server-69cdc45777-69648",
	"reconcileID": "d6056528-c1f1-459c-a00c-4fd37699e0e9",
	"targetname": "echo-server",
	"targetstates": {
		"startupProbe": "true",
		"readinessProbe": "false",
		"container": "running",
		"started": "false",
		"ready": "false",
		"resources": "startup",
		"statusResources": "containerresourcesmismatch"
	},
	"time": 1698695786101,
	"message": "startup scale not yet completed - has been proposed"
}
{
	"level": "debug",
	"namespace": "deployment-non-startup-admitted-flow-startup-probe",
	"name": "echo-server-69cdc45777-69648",
	"reconcileID": "45721bfe-7503-4067-a32c-74fbfea2866e",
	"targetname": "echo-server",
	"targetstates": {
		"startupProbe": "true",
		"readinessProbe": "false",
		"container": "running",
		"started": "false",
		"ready": "false",
		"resources": "startup",
		"statusResources": "containerresourcesmismatch"
	},
	"time": 1698695788169,
	"message": "startup scale not yet completed - in progress"
}
{
	"level": "info",
	"namespace": "deployment-non-startup-admitted-flow-startup-probe",
	"name": "echo-server-69cdc45777-69648",
	"reconcileID": "b96b072a-6ff3-4cb4-a27e-af1666f04ef7",
	"targetname": "echo-server",
	"targetstates": {
		"startupProbe": "true",
		"readinessProbe": "false",
		"container": "running",
		"started": "false",
		"ready": "false",
		"resources": "startup",
		"statusResources": "containerresourcesmatch"
	},
	"time": 1698695789015,
	"message": "startup resources enacted"
}
{
	"level": "debug",
	"namespace": "deployment-non-startup-admitted-flow-startup-probe",
	"name": "echo-server-69cdc45777-69648",
	"reconcileID": "4057fe97-8540-4286-85dd-0548c6995877",
	"targetname": "echo-server",
	"targetstates": {
		"startupProbe": "true",
		"readinessProbe": "false",
		"container": "running",
		"started": "false",
		"ready": "false",
		"resources": "startup",
		"statusResources": "containerresourcesmismatch" <-- HERE
	},
	"time": 1698695789058,
	"message": "startup scale not yet completed - in progress" <-- HERE
}
{
	"level": "info",
	"namespace": "deployment-non-startup-admitted-flow-startup-probe",
	"name": "echo-server-69cdc45777-69648",
	"reconcileID": "bab2a5cd-809e-4139-bca6-db28c28c6a63",
	"targetname": "echo-server",
	"targetstates": {
		"startupProbe": "true",
		"readinessProbe": "false",
		"container": "running",
		"started": "false",
		"ready": "false",
		"resources": "startup",
		"statusResources": "containerresourcesmatch"
	},
	"time": 1698695789982,
	"message": "startup resources enacted"
}

*/

func TestMain(m *testing.M) {
	suppliedConfigInit()

	_ = flag.Set("test.parallel", suppliedConfig.maxParallelism)
	flag.Parse()
	if testing.Short() {
		logMessage(nil, "not running because short tests configured")
		os.Exit(0)
	}

	kindSetupCluster(nil)
	if err := csaRun(nil); err != nil {
		if !suppliedConfig.keepCsa {
			csaCleanUp(nil)
		}
		if !suppliedConfig.keepCluster {
			kindCleanUpCluster(nil)
		}
		logMessage(nil, err)
		os.Exit(1)
	}

	exitVal := m.Run()
	if !suppliedConfig.keepCsa {
		csaCleanUp(nil)
	}
	if !suppliedConfig.keepCluster {
		kindCleanUpCluster(nil)
	}
	os.Exit(exitVal)
}

// Deployment ----------------------------------------------------------------------------------------------------------

func TestDeploymentNonStartupAdmittedFlowStartupProbeAll(t *testing.T) {
	t.Parallel()
	namespace := "deployment-non-startup-admitted-flow-startup-probe-all"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsAllDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardPostStartup(namespace, int32(replicas), annotations)
			config.removeReadinessProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
	)
}

func TestDeploymentNonStartupAdmittedFlowStartupProbeCpu(t *testing.T) {
	t.Parallel()
	namespace := "deployment-non-startup-admitted-flow-startup-probe-cpu"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsCpuOnlyDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardPostStartup(namespace, int32(replicas), annotations)
			config.removeReadinessProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
	)
}

func TestDeploymentNonStartupAdmittedFlowStartupProbeMemory(t *testing.T) {
	t.Parallel()
	namespace := "deployment-non-startup-admitted-flow-startup-probe-memory"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsMemoryOnlyDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardPostStartup(namespace, int32(replicas), annotations)
			config.removeReadinessProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
	)
}

func TestDeploymentStartupAdmittedFlowStartupProbeAll(t *testing.T) {
	t.Parallel()
	namespace := "deployment-startup-admitted-flow-startup-probe-all"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsAllDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardStartup(namespace, int32(replicas), annotations)
			config.removeReadinessProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
	)
}

func TestDeploymentStartupAdmittedFlowStartupProbeCpu(t *testing.T) {
	t.Parallel()
	namespace := "deployment-startup-admitted-flow-startup-probe-cpu"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsCpuOnlyDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardStartup(namespace, int32(replicas), annotations)
			config.removeReadinessProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
	)
}

func TestDeploymentStartupAdmittedFlowStartupProbeMemory(t *testing.T) {
	t.Parallel()
	namespace := "deployment-startup-admitted-flow-startup-probe-memory"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsMemoryOnlyDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardStartup(namespace, int32(replicas), annotations)
			config.removeReadinessProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
	)
}

func TestDeploymentNonStartupAdmittedFlowReadinessProbeAll(t *testing.T) {
	t.Parallel()
	namespace := "deployment-non-startup-admitted-flow-readiness-probe-all"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsAllDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardPostStartup(namespace, int32(replicas), annotations)
			config.removeStartupProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
	)
}

func TestDeploymentNonStartupAdmittedFlowReadinessProbeCpu(t *testing.T) {
	t.Parallel()
	namespace := "deployment-non-startup-admitted-flow-readiness-probe-cpu"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsCpuOnlyDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardPostStartup(namespace, int32(replicas), annotations)
			config.removeStartupProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
	)
}

func TestDeploymentNonStartupAdmittedFlowReadinessProbeMemory(t *testing.T) {
	t.Parallel()
	namespace := "deployment-non-startup-admitted-flow-readiness-probe-memory"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsMemoryOnlyDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardPostStartup(namespace, int32(replicas), annotations)
			config.removeStartupProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
	)
}

func TestDeploymentStartupAdmittedFlowReadinessProbeAll(t *testing.T) {
	t.Parallel()
	namespace := "deployment-startup-admitted-flow-readiness-probe-all"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsAllDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardStartup(namespace, int32(replicas), annotations)
			config.removeStartupProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
	)
}

func TestDeploymentStartupAdmittedFlowReadinessProbeCpu(t *testing.T) {
	t.Parallel()
	namespace := "deployment-startup-admitted-flow-readiness-probe-cpu"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsCpuOnlyDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardStartup(namespace, int32(replicas), annotations)
			config.removeStartupProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
	)
}

func TestDeploymentStartupAdmittedFlowReadinessProbeMemory(t *testing.T) {
	t.Parallel()
	namespace := "deployment-startup-admitted-flow-readiness-probe-memory"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsMemoryOnlyDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardStartup(namespace, int32(replicas), annotations)
			config.removeStartupProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
	)
}

func TestDeploymentScaleWhenUnknownResourcesAll(t *testing.T) {
	t.Parallel()
	namespace := "deployment-scale-when-unknown-resources-all"
	maybeRegisterCleanup(t, namespace)

	_ = kubeDeleteNamespace(t, namespace)
	maybeLogErrAndFailNow(t, kubeCreateNamespace(t, namespace))

	annotations := csaQuantityAnnotations{
		cpuStartup:                "200m",
		cpuPostStartupRequests:    "50m",
		cpuPostStartupLimits:      "50m",
		memoryStartup:             "200M",
		memoryPostStartupRequests: "150M",
		memoryPostStartupLimits:   "150M",
	}

	config := echoDeploymentConfigStandard(
		namespace,
		2,
		annotations,
		"175m", "175m",
		"175M", "175M",
		echoServerDefaultProbeInitialDelaySeconds,
	)
	config.removeReadinessProbes()
	maybeLogErrAndFailNow(t, kubeApplyYamlOrJsonResources(t, config.deploymentJson()))

	names, err := kubeGetPodNames(t, namespace, echoServerName)
	maybeLogErrAndFailNow(t, err)

	podStatusAnn, errs := csaWaitStatusAll(t, namespace, names, csaStatusMessageStartupEnacted, testsDefaultWaitStatusTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}

	assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)

	podStatusAnn, errs = csaWaitStatusAll(t, namespace, names, csaStatusMessagePostStartupEnacted, testsDefaultWaitStatusTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}

	assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)

	ensureEvents(
		t,
		csaEventReasonScaling,
		[]string{
			csaStatusMessageStartupCommandedUnknownRes, csaStatusMessageStartupEnacted,
			csaStatusMessagePostStartupCommanded, csaStatusMessagePostStartupEnacted,
		},
		namespace,
		names,
	)
}

func TestDeploymentScaleWhenUnknownResourcesCpu(t *testing.T) {
	t.Parallel()
	namespace := "deployment-scale-when-unknown-resources-cpu"
	maybeRegisterCleanup(t, namespace)

	_ = kubeDeleteNamespace(t, namespace)
	maybeLogErrAndFailNow(t, kubeCreateNamespace(t, namespace))

	annotations := csaQuantityAnnotations{
		cpuStartup:             "200m",
		cpuPostStartupRequests: "50m",
		cpuPostStartupLimits:   "50m",
	}

	config := echoDeploymentConfigStandard(
		namespace,
		2,
		annotations,
		"175m", "175m",
		echoServerMemoryDisabledRequests, echoServerMemoryDisabledLimits,
		echoServerDefaultProbeInitialDelaySeconds,
	)
	config.removeReadinessProbes()
	maybeLogErrAndFailNow(t, kubeApplyYamlOrJsonResources(t, config.deploymentJson()))

	names, err := kubeGetPodNames(t, namespace, echoServerName)
	maybeLogErrAndFailNow(t, err)

	podStatusAnn, errs := csaWaitStatusAll(t, namespace, names, csaStatusMessageStartupEnacted, testsDefaultWaitStatusTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}

	assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)

	podStatusAnn, errs = csaWaitStatusAll(t, namespace, names, csaStatusMessagePostStartupEnacted, testsDefaultWaitStatusTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}

	assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)

	ensureEvents(
		t,
		csaEventReasonScaling,
		[]string{
			csaStatusMessageStartupCommandedUnknownRes, csaStatusMessageStartupEnacted,
			csaStatusMessagePostStartupCommanded, csaStatusMessagePostStartupEnacted,
		},
		namespace,
		names,
	)
}

func TestDeploymentScaleWhenUnknownResourcesMemory(t *testing.T) {
	t.Parallel()
	namespace := "deployment-scale-when-unknown-resources-memory"
	maybeRegisterCleanup(t, namespace)

	_ = kubeDeleteNamespace(t, namespace)
	maybeLogErrAndFailNow(t, kubeCreateNamespace(t, namespace))

	annotations := csaQuantityAnnotations{
		memoryStartup:             "200M",
		memoryPostStartupRequests: "150M",
		memoryPostStartupLimits:   "150M",
	}

	config := echoDeploymentConfigStandard(
		namespace,
		2,
		annotations,
		echoServerCpuDisabledRequests, echoServerCpuDisabledLimits,
		"175M", "175M",
		echoServerDefaultProbeInitialDelaySeconds,
	)
	config.removeReadinessProbes()
	maybeLogErrAndFailNow(t, kubeApplyYamlOrJsonResources(t, config.deploymentJson()))

	names, err := kubeGetPodNames(t, namespace, echoServerName)
	maybeLogErrAndFailNow(t, err)

	podStatusAnn, errs := csaWaitStatusAll(t, namespace, names, csaStatusMessageStartupEnacted, testsDefaultWaitStatusTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}

	assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)

	podStatusAnn, errs = csaWaitStatusAll(t, namespace, names, csaStatusMessagePostStartupEnacted, testsDefaultWaitStatusTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}

	assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)

	ensureEvents(
		t,
		csaEventReasonScaling,
		[]string{
			csaStatusMessageStartupCommandedUnknownRes, csaStatusMessageStartupEnacted,
			csaStatusMessagePostStartupCommanded, csaStatusMessagePostStartupEnacted,
		},
		namespace,
		names,
	)
}

// StatefulSet ---------------------------------------------------------------------------------------------------------

func TestStatefulSetFlowStartupProbeAll(t *testing.T) {
	t.Parallel()
	namespace := "statefulset-flow-startup-probe-all"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsAllDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 1 // Can only test with 1 replica since pods are started sequentially (after become ready).
			config := echoStatefulSetConfigStandardPostStartup(namespace, int32(replicas), annotations)
			config.removeReadinessProbes()
			return config.statefulSetJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
	)
}

func TestStatefulSetFlowReadinessProbeAll(t *testing.T) {
	t.Parallel()
	namespace := "statefulset-flow-readiness-probe-all"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsAllDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 1
			config := echoStatefulSetConfigStandardPostStartup(namespace, int32(replicas), annotations)
			config.removeStartupProbes()
			return config.statefulSetJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
	)
}

// DaemonSet -----------------------------------------------------------------------------------------------------------

func TestDaemonSetFlowStartupProbeAll(t *testing.T) {
	t.Parallel()
	namespace := "daemonset-flow-startup-probe-all"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsAllDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			config := echoDaemonSetConfigStandardPostStartup(namespace, annotations)
			config.removeReadinessProbes()
			return config.daemonSetJson(), 1 // 1 node.
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
	)
}

func TestDaemonSetFlowReadinessProbeAll(t *testing.T) {
	t.Parallel()
	namespace := "daemonset-flow-readiness-probe-all"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		csaQuantityAnnotationsAllDefault,
		func(annotations csaQuantityAnnotations) (string, int) {
			config := echoDaemonSetConfigStandardPostStartup(namespace, annotations)
			config.removeStartupProbes()
			return config.daemonSetJson(), 1
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]pod.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
	)
}

// Failure -------------------------------------------------------------------------------------------------------------

func TestScaleFailure(t *testing.T) {
	// TODO(wt) is it possible to test this?
}

func TestValidationFailure(t *testing.T) {
	t.Parallel()
	namespace := "validation-failure"
	maybeRegisterCleanup(t, namespace)

	_ = kubeDeleteNamespace(t, namespace)
	maybeLogErrAndFailNow(t, kubeCreateNamespace(t, namespace))

	annotations := csaQuantityAnnotations{
		cpuStartup:                "50m",
		cpuPostStartupRequests:    "200m",
		cpuPostStartupLimits:      "200m",
		memoryStartup:             "200M",
		memoryPostStartupRequests: "150M",
		memoryPostStartupLimits:   "150M",
	}

	config := echoDeploymentConfigStandardStartup(namespace, 2, annotations)
	maybeLogErrAndFailNow(t, kubeApplyYamlOrJsonResources(t, config.deploymentJson()))

	names, err := kubeGetPodNames(t, namespace, echoServerName)
	maybeLogErrAndFailNow(t, err)

	podStatusAnn, errs := csaWaitStatusAll(t, namespace, names, csaStatusMessageValidationError, testsDefaultWaitStatusTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}

	for _, statusAnn := range podStatusAnn {
		assert.Contains(t, statusAnn.Status, "cpu post-startup requests (200m) is greater than startup value (50m)")
		require.NotEmpty(t, statusAnn.LastUpdated)

		require.Equal(t, podcommon.StateBoolUnknown, statusAnn.States.StartupProbe)
		require.Equal(t, podcommon.StateBoolUnknown, statusAnn.States.ReadinessProbe)
		require.Equal(t, podcommon.StateContainerUnknown, statusAnn.States.Container)
		require.Equal(t, podcommon.StateBoolUnknown, statusAnn.States.Started)
		require.Equal(t, podcommon.StateBoolUnknown, statusAnn.States.Ready)
		require.Equal(t, podcommon.StateResourcesUnknown, statusAnn.States.Resources)
		require.Equal(t, podcommon.StateStatusResourcesUnknown, statusAnn.States.StatusResources)

		require.Empty(t, statusAnn.Scale.LastCommanded)
		require.Empty(t, statusAnn.Scale.LastEnacted)
		require.Empty(t, statusAnn.Scale.LastFailed)
	}

	ensureEvents(t, csaEventReasonValidation, []string{csaStatusMessageValidationError}, namespace, names)
}

// Helpers -------------------------------------------------------------------------------------------------------------

func testWorkflow(
	t *testing.T,
	namespace string,
	annotations csaQuantityAnnotations,
	workloadJsonReplicasFunc func(csaQuantityAnnotations) (string, int),
	assertStartupEnactedFunc func(*testing.T, csaQuantityAnnotations, map[*v1.Pod]pod.StatusAnnotation),
	assertPostStartupEnactedFunc func(*testing.T, csaQuantityAnnotations, map[*v1.Pod]pod.StatusAnnotation),
	assertStartupEnactedRestartFunc func(*testing.T, csaQuantityAnnotations, map[*v1.Pod]pod.StatusAnnotation),
	assertPostStartupEnactedRestartFunc func(*testing.T, csaQuantityAnnotations, map[*v1.Pod]pod.StatusAnnotation),
) {
	_ = kubeDeleteNamespace(t, namespace)
	maybeLogErrAndFailNow(t, kubeCreateNamespace(t, namespace))

	workloadJson, replicas := workloadJsonReplicasFunc(annotations)
	maybeLogErrAndFailNow(t, kubeApplyYamlOrJsonResources(t, workloadJson))

	maybeLogErrAndFailNow(t, kubeWaitPodsExist(t, namespace, echoServerName, replicas, testsDefaultWaitStatusTimeoutSecs))

	names, err := kubeGetPodNames(t, namespace, echoServerName)
	maybeLogErrAndFailNow(t, err)

	// Startup resources enacted ---------------------------------------------------------------------------------------
	podStatusAnn, errs := csaWaitStatusAll(t, namespace, names, csaStatusMessageStartupEnacted, testsDefaultWaitStatusTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}
	assertStartupEnactedFunc(t, annotations, podStatusAnn)

	// Post-startup resources enacted ----------------------------------------------------------------------------------
	podStatusAnn, errs = csaWaitStatusAll(t, namespace, names, csaStatusMessagePostStartupEnacted, testsDefaultWaitStatusTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}
	assertPostStartupEnactedFunc(t, annotations, podStatusAnn)

	// Container restart startup and post-startup resources enacted ----------------------------------------------------
	for kubePod := range podStatusAnn {
		for _, status := range kubePod.Status.ContainerStatuses {
			if status.Name == echoServerName {
				maybeLogErrAndFailNow(t, kubeCauseContainerRestart(t, status.ContainerID))
			}
		}
	}

	podStatusAnn, errs = csaWaitStatusAll(t, namespace, names, csaStatusMessageStartupEnacted, testsDefaultWaitStatusTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}
	assertStartupEnactedRestartFunc(t, annotations, podStatusAnn)

	podStatusAnn, errs = csaWaitStatusAll(t, namespace, names, csaStatusMessagePostStartupEnacted, testsDefaultWaitStatusTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}
	assertPostStartupEnactedRestartFunc(t, annotations, podStatusAnn)

	ensureEvents(
		t,
		csaEventReasonScaling,
		[]string{
			csaStatusMessageStartupCommanded, csaStatusMessageStartupEnacted,
			csaStatusMessagePostStartupCommanded, csaStatusMessagePostStartupEnacted,
		},
		namespace,
		names,
	)
}

func assertStartupEnacted(
	t *testing.T,
	annotations csaQuantityAnnotations,
	podStatusAnn map[*v1.Pod]pod.StatusAnnotation,
	expectStartupProbe bool,
	expectReadinessProbe bool,
	expectStatusCommandedEnactedEmpty bool,
) {
	if (!expectStartupProbe && !expectReadinessProbe) || (expectStartupProbe && expectReadinessProbe) {
		panic(errors.New("only one of expectStartupProbe/expectReadinessProbe must be true"))
	}

	for kubePod, statusAnn := range podStatusAnn {
		for _, c := range kubePod.Spec.Containers {
			var expectCpuR, expectCpuL, expectMemoryR, expectMemoryL string

			if c.Name == echoServerName {
				expectCpuR, expectCpuL = annotations.CpuStartupRequestsLimits()
				expectMemoryR, expectMemoryL = annotations.MemoryStartupRequestsLimits()
			} else if c.Name == echoServerNonTargetContainerName {
				expectCpuR, expectCpuL = echoServerNonTargetContainerCpuRequests, echoServerNonTargetContainerCpuLimits
				expectMemoryR, expectMemoryL = echoServerNonTargetContainerMemoryRequests, echoServerNonTargetContainerMemoryLimits
			} else {
				panic(errors.New("container name unrecognized"))
			}

			if expectStartupProbe {
				require.NotNil(t, c.StartupProbe)
			} else {
				require.Nil(t, c.StartupProbe)
			}
			if expectReadinessProbe {
				require.NotNil(t, c.ReadinessProbe)
			} else {
				require.Nil(t, c.ReadinessProbe)
			}
			cpuR := c.Resources.Requests[v1.ResourceCPU]
			require.Equal(t, expectCpuR, cpuR.String())
			cpuL := c.Resources.Limits[v1.ResourceCPU]
			require.Equal(t, expectCpuL, cpuL.String())
			memoryR := c.Resources.Requests[v1.ResourceMemory]
			require.Equal(t, expectMemoryR, memoryR.String())
			memoryL := c.Resources.Limits[v1.ResourceMemory]
			require.Equal(t, expectMemoryL, memoryL.String())
		}

		for _, s := range kubePod.Status.ContainerStatuses {
			var expectCpuR, expectCpuL, expectMemoryR, expectMemoryL string

			if s.Name == echoServerName {
				expectCpuR, expectCpuL = annotations.CpuStartupRequestsLimits()
				expectMemoryR, expectMemoryL = annotations.MemoryStartupRequestsLimits()

				// See comment in targetcontaineraction.go
				if expectStartupProbe {
					require.False(t, *s.Started)
				} else {
					require.True(t, *s.Started)
				}
				require.False(t, s.Ready)
			} else if s.Name == echoServerNonTargetContainerName {
				expectCpuR, expectCpuL = echoServerNonTargetContainerCpuRequests, echoServerNonTargetContainerCpuLimits
				expectMemoryR, expectMemoryL = echoServerNonTargetContainerMemoryRequests, echoServerNonTargetContainerMemoryLimits
			} else {
				panic(errors.New("container name unrecognized"))
			}

			require.NotNil(t, s.State.Running)
			cpuR := s.Resources.Requests[v1.ResourceCPU]
			require.Equal(t, expectCpuR, cpuR.String())
			cpuL := s.Resources.Limits[v1.ResourceCPU]
			require.Equal(t, expectCpuL, cpuL.String())
			memoryR := s.Resources.Requests[v1.ResourceMemory]
			require.Equal(t, expectMemoryR, memoryR.String())
			memoryL := s.Resources.Limits[v1.ResourceMemory]
			require.Equal(t, expectMemoryL, memoryL.String())
		}

		require.Equal(t, csaStatusMessageStartupEnacted, statusAnn.Status)
		require.NotEmpty(t, statusAnn.LastUpdated)

		require.Equal(t, expectStartupProbe, statusAnn.States.StartupProbe.Bool())
		require.Equal(t, expectReadinessProbe, statusAnn.States.ReadinessProbe.Bool())
		require.Equal(t, podcommon.StateContainerRunning, statusAnn.States.Container)
		if expectStartupProbe {
			require.Equal(t, podcommon.StateBoolFalse, statusAnn.States.Started)
		} else {
			require.Equal(t, podcommon.StateBoolTrue, statusAnn.States.Started)
		}
		require.Equal(t, podcommon.StateBoolFalse, statusAnn.States.Ready)
		require.Equal(t, podcommon.StateResourcesStartup, statusAnn.States.Resources)
		require.Equal(t, podcommon.StateStatusResourcesContainerResourcesMatch, statusAnn.States.StatusResources)

		if expectStatusCommandedEnactedEmpty {
			require.Empty(t, statusAnn.Scale.LastCommanded)
			require.Empty(t, statusAnn.Scale.LastEnacted)
		} else {
			require.NotEmpty(t, statusAnn.Scale.LastCommanded)
			require.NotEmpty(t, statusAnn.Scale.LastEnacted)
		}
		require.Empty(t, statusAnn.Scale.LastFailed)
	}
}

func assertPostStartupEnacted(
	t *testing.T,
	annotations csaQuantityAnnotations,
	podStatusAnn map[*v1.Pod]pod.StatusAnnotation,
	expectStartupProbe bool,
	expectReadinessProbe bool,
) {
	for kubePod, statusAnn := range podStatusAnn {
		for _, c := range kubePod.Spec.Containers {
			var expectCpuR, expectCpuL, expectMemoryR, expectMemoryL string

			if c.Name == echoServerName {
				expectCpuR, expectCpuL = annotations.CpuPostStartupRequestsLimits()
				expectMemoryR, expectMemoryL = annotations.MemoryPostStartupRequestsLimits()
			} else if c.Name == echoServerNonTargetContainerName {
				expectCpuR, expectCpuL = echoServerNonTargetContainerCpuRequests, echoServerNonTargetContainerCpuLimits
				expectMemoryR, expectMemoryL = echoServerNonTargetContainerMemoryRequests, echoServerNonTargetContainerMemoryLimits
			} else {
				panic(errors.New("container name unrecognized"))
			}

			if expectStartupProbe {
				require.NotNil(t, c.StartupProbe)
			} else {
				require.Nil(t, c.StartupProbe)
			}
			if expectReadinessProbe {
				require.NotNil(t, c.ReadinessProbe)
			} else {
				require.Nil(t, c.ReadinessProbe)
			}
			cpuR := c.Resources.Requests[v1.ResourceCPU]
			require.Equal(t, expectCpuR, cpuR.String())
			cpuL := c.Resources.Limits[v1.ResourceCPU]
			require.Equal(t, expectCpuL, cpuL.String())
			memoryR := c.Resources.Requests[v1.ResourceMemory]
			require.Equal(t, expectMemoryR, memoryR.String())
			memoryL := c.Resources.Limits[v1.ResourceMemory]
			require.Equal(t, expectMemoryL, memoryL.String())
		}

		for _, s := range kubePod.Status.ContainerStatuses {
			var expectCpuR, expectCpuL, expectMemoryR, expectMemoryL string

			if s.Name == echoServerName {
				expectCpuR, expectCpuL = annotations.CpuPostStartupRequestsLimits()
				expectMemoryR, expectMemoryL = annotations.MemoryPostStartupRequestsLimits()

				// See comment in targetcontaineraction.go
				require.True(t, *s.Started)
				require.True(t, s.Ready)
			} else if s.Name == echoServerNonTargetContainerName {
				expectCpuR, expectCpuL = echoServerNonTargetContainerCpuRequests, echoServerNonTargetContainerCpuLimits
				expectMemoryR, expectMemoryL = echoServerNonTargetContainerMemoryRequests, echoServerNonTargetContainerMemoryLimits
			} else {
				panic(errors.New("container name unrecognized"))
			}

			require.NotNil(t, s.State.Running)
			cpuR := s.Resources.Requests[v1.ResourceCPU]
			require.Equal(t, expectCpuR, cpuR.String())
			cpuL := s.Resources.Limits[v1.ResourceCPU]
			require.Equal(t, expectCpuL, cpuL.String())
			memoryR := s.Resources.Requests[v1.ResourceMemory]
			require.Equal(t, expectMemoryR, memoryR.String())
			memoryL := s.Resources.Limits[v1.ResourceMemory]
			require.Equal(t, expectMemoryL, memoryL.String())
		}

		require.Equal(t, csaStatusMessagePostStartupEnacted, statusAnn.Status)
		require.NotEmpty(t, statusAnn.LastUpdated)

		require.Equal(t, expectStartupProbe, statusAnn.States.StartupProbe.Bool())
		require.Equal(t, expectReadinessProbe, statusAnn.States.ReadinessProbe.Bool())
		require.Equal(t, podcommon.StateContainerRunning, statusAnn.States.Container)
		require.Equal(t, podcommon.StateBoolTrue, statusAnn.States.Started)
		require.Equal(t, podcommon.StateBoolTrue, statusAnn.States.Ready)
		require.Equal(t, podcommon.StateResourcesPostStartup, statusAnn.States.Resources)
		require.Equal(t, podcommon.StateStatusResourcesContainerResourcesMatch, statusAnn.States.StatusResources)

		require.NotEmpty(t, statusAnn.Scale.LastCommanded)
		require.NotEmpty(t, statusAnn.Scale.LastEnacted)
		require.Empty(t, statusAnn.Scale.LastFailed)
	}
}

func ensureEvents(t *testing.T, reason string, substrs []string, namespace string, names []string) {
	for _, name := range names {
		messages, err := kubeGetEventMessages(t, namespace, name, reason)
		maybeLogErrAndFailNow(t, err)

		for _, substr := range substrs {
			gotMessage := false
			for _, message := range messages {
				if strings.Contains(message, substr) {
					gotMessage = true
					break
				}
			}

			assert.True(t, gotMessage)
		}
	}
}

func maybeRegisterCleanup(t *testing.T, namespace string) {
	if suppliedConfig.deleteNsPostTest {
		t.Cleanup(func() {
			_ = kubeDeleteNamespace(t, namespace)
		})
	}
}

func maybeLogErrAndFailNow(t *testing.T, err error) {
	if err != nil {
		logMessage(t, err)
		t.FailNow()
	}
}
