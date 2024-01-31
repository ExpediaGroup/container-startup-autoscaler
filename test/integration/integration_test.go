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
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
)

/*

// TODO(wt) 'In-place Update of Pod Resources' implementation bug
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
		"allocatedResources": "unknown",
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
		"allocatedResources": "containerrequestsmatch",
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
		"allocatedResources": "containerrequestsmismatch",
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
		"allocatedResources": "containerrequestsmatch",
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
		"allocatedResources": "containerrequestsmatch",
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
		"allocatedResources": "containerrequestsmatch",
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
		"allocatedResources": "containerrequestsmatch",
		"statusResources": "containerresourcesmatch"
	},
	"time": 1698695789982,
	"message": "startup resources enacted"
}

*/

const (
	defaultTimeoutSecs = 60
)

var deleteNsPostTest = true

func TestMain(m *testing.M) {
	setStringConfig := func(env string, config *string) {
		envVal := os.Getenv(env)
		if envVal != "" {
			*config = envVal
		}
	}
	setBoolConfig := func(env string, config *bool) {
		envVal := os.Getenv(env)
		if envVal != "" {
			var err error
			*config, err = strconv.ParseBool(envVal)
			if err != nil {
				fmt.Println("(config)", env, "value is not a bool")
				os.Exit(1)
			}
		}
	}

	maxParallelism := "4"
	reuseCluster := false
	installMetricsServer := false
	keepCsa := false
	keepCluster := false

	setStringConfig("MAX_PARALLELISM", &maxParallelism)
	setBoolConfig("REUSE_CLUSTER", &reuseCluster)
	setBoolConfig("INSTALL_METRICS_SERVER", &installMetricsServer)
	setBoolConfig("KEEP_CSA", &keepCsa)
	setBoolConfig("KEEP_CLUSTER", &keepCluster)
	setBoolConfig("DELETE_NS_AFTER_TEST", &deleteNsPostTest)

	fmt.Println("(config) MAX_PARALLELISM:", maxParallelism)
	fmt.Println("(config) REUSE_CLUSTER:", reuseCluster)
	fmt.Println("(config) INSTALL_METRICS_SERVER:", installMetricsServer)
	fmt.Println("(config) KEEP_CSA:", keepCsa)
	fmt.Println("(config) KEEP_CLUSTER:", keepCluster)
	fmt.Println("(config) DELETE_NS_AFTER_TEST:", deleteNsPostTest)

	_ = flag.Set("test.parallel", maxParallelism)
	flag.Parse()
	if testing.Short() {
		fmt.Println("not running because short tests configured")
		os.Exit(0)
	}

	kindSetupCluster(reuseCluster, installMetricsServer)
	if err := csaRun(); err != nil {
		if !keepCsa {
			csaCleanUp()
		}
		if !keepCluster {
			kindCleanUpCluster()
		}
		fmt.Println(err)
		os.Exit(1)
	}

	exitVal := m.Run()
	if !keepCsa {
		csaCleanUp()
	}
	if !keepCluster {
		kindCleanUpCluster()
	}
	os.Exit(exitVal)
}

// Deployment ----------------------------------------------------------------------------------------------------------

func TestDeploymentNonStartupAdmittedFlowStartupProbe(t *testing.T) {
	t.Parallel()
	namespace := "deployment-non-startup-admitted-flow-startup-probe"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardPostStartup(namespace, int32(replicas), annotations)
			config.removeReadinessProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
	)
}

func TestDeploymentStartupAdmittedFlowStartupProbe(t *testing.T) {
	t.Parallel()
	namespace := "deployment-startup-admitted-flow-startup-probe"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardStartup(namespace, int32(replicas), annotations)
			config.removeReadinessProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
	)
}

func TestDeploymentNonStartupAdmittedFlowReadinessProbe(t *testing.T) {
	t.Parallel()
	namespace := "deployment-non-startup-admitted-flow-readiness-probe"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardPostStartup(namespace, int32(replicas), annotations)
			config.removeStartupProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
	)
}

func TestDeploymentStartupAdmittedFlowReadinessProbe(t *testing.T) {
	t.Parallel()
	namespace := "deployment-startup-admitted-flow-readiness-probe"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 2
			config := echoDeploymentConfigStandardStartup(namespace, int32(replicas), annotations)
			config.removeStartupProbes()
			return config.deploymentJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
	)
}

func TestDeploymentScaleWhenUnknownResources(t *testing.T) {
	t.Parallel()
	namespace := "deployment-scale-when-unknown-resources"
	maybeRegisterCleanup(t, namespace)

	_ = kubeDeleteNamespace(namespace)
	maybeLogErrAndFailNow(t, kubeCreateNamespace(namespace))

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
	maybeLogErrAndFailNow(t, kubeApplyYamlOrJsonResources(config.deploymentJson()))

	names, err := kubeGetPodNames(namespace, echoServerName)
	maybeLogErrAndFailNow(t, err)

	podStatusAnn, errs := csaWaitStatusAll(namespace, names, csaStatusMessageStartupEnacted, defaultTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}

	assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)

	podStatusAnn, errs = csaWaitStatusAll(namespace, names, csaStatusMessagePostStartupEnacted, defaultTimeoutSecs)
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

func TestStatefulSetFlowStartupProbe(t *testing.T) {
	t.Parallel()
	namespace := "statefulset-flow-startup-probe"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 1 // Can only test with 1 replica since pods are started sequentially (after become ready).
			config := echoStatefulSetConfigStandardPostStartup(namespace, int32(replicas), annotations)
			config.removeReadinessProbes()
			return config.statefulSetJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
	)
}

func TestStatefulSetFlowReadinessProbe(t *testing.T) {
	t.Parallel()
	namespace := "statefulset-flow-readiness-probe"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		func(annotations csaQuantityAnnotations) (string, int) {
			replicas := 1
			config := echoStatefulSetConfigStandardPostStartup(namespace, int32(replicas), annotations)
			config.removeStartupProbes()
			return config.statefulSetJson(), replicas
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
	)
}

// DaemonSet -----------------------------------------------------------------------------------------------------------

func TestDaemonSetFlowStartupProbe(t *testing.T) {
	t.Parallel()
	namespace := "daemonset-flow-startup-probe"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		func(annotations csaQuantityAnnotations) (string, int) {
			config := echoDaemonSetConfigStandardPostStartup(namespace, annotations)
			config.removeReadinessProbes()
			return config.daemonSetJson(), 1 // 1 node.
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
	)
}

func TestDaemonSetFlowReadinessProbe(t *testing.T) {
	t.Parallel()
	namespace := "daemonset-flow-readiness-probe"
	maybeRegisterCleanup(t, namespace)

	testWorkflow(
		t,
		namespace,
		func(annotations csaQuantityAnnotations) (string, int) {
			config := echoDaemonSetConfigStandardPostStartup(namespace, annotations)
			config.removeStartupProbes()
			return config.daemonSetJson(), 1
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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

	_ = kubeDeleteNamespace(namespace)
	maybeLogErrAndFailNow(t, kubeCreateNamespace(namespace))

	annotations := csaQuantityAnnotations{
		cpuStartup:                "50m",
		cpuPostStartupRequests:    "200m",
		cpuPostStartupLimits:      "200m",
		memoryStartup:             "200M",
		memoryPostStartupRequests: "150M",
		memoryPostStartupLimits:   "150M",
	}

	config := echoDeploymentConfigStandardStartup(namespace, 2, annotations)
	maybeLogErrAndFailNow(t, kubeApplyYamlOrJsonResources(config.deploymentJson()))

	names, err := kubeGetPodNames(namespace, echoServerName)
	maybeLogErrAndFailNow(t, err)

	podStatusAnn, errs := csaWaitStatusAll(namespace, names, csaStatusMessageValidationError, defaultTimeoutSecs)
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
		require.Equal(t, podcommon.StateAllocatedResourcesUnknown, statusAnn.States.AllocatedResources)
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
	workloadJsonReplicasFunc func(csaQuantityAnnotations) (string, int),
	assertStartupEnactedFunc func(*testing.T, csaQuantityAnnotations, map[*v1.Pod]podcommon.StatusAnnotation),
	assertPostStartupEnactedFunc func(*testing.T, csaQuantityAnnotations, map[*v1.Pod]podcommon.StatusAnnotation),
	assertStartupEnactedRestartFunc func(*testing.T, csaQuantityAnnotations, map[*v1.Pod]podcommon.StatusAnnotation),
	assertPostStartupEnactedRestartFunc func(*testing.T, csaQuantityAnnotations, map[*v1.Pod]podcommon.StatusAnnotation),
) {
	_ = kubeDeleteNamespace(namespace)
	maybeLogErrAndFailNow(t, kubeCreateNamespace(namespace))

	annotations := csaQuantityAnnotations{
		cpuStartup:                "200m",
		cpuPostStartupRequests:    "50m",
		cpuPostStartupLimits:      "50m",
		memoryStartup:             "200M",
		memoryPostStartupRequests: "150M",
		memoryPostStartupLimits:   "150M",
	}

	workloadJson, replicas := workloadJsonReplicasFunc(annotations)
	maybeLogErrAndFailNow(t, kubeApplyYamlOrJsonResources(workloadJson))

	maybeLogErrAndFailNow(t, kubeWaitPodsExist(namespace, echoServerName, replicas, defaultTimeoutSecs))

	names, err := kubeGetPodNames(namespace, echoServerName)
	maybeLogErrAndFailNow(t, err)

	// Startup resources enacted ---------------------------------------------------------------------------------------
	podStatusAnn, errs := csaWaitStatusAll(namespace, names, csaStatusMessageStartupEnacted, defaultTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}
	assertStartupEnactedFunc(t, annotations, podStatusAnn)

	// Post-startup resources enacted ----------------------------------------------------------------------------------
	podStatusAnn, errs = csaWaitStatusAll(namespace, names, csaStatusMessagePostStartupEnacted, defaultTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}
	assertPostStartupEnactedFunc(t, annotations, podStatusAnn)

	// Container restart startup and post-startup resources enacted ----------------------------------------------------
	for pod := range podStatusAnn {
		for _, status := range pod.Status.ContainerStatuses {
			if status.Name == echoServerName {
				maybeLogErrAndFailNow(t, kubeCauseContainerRestart(status.ContainerID))
			}
		}
	}

	podStatusAnn, errs = csaWaitStatusAll(namespace, names, csaStatusMessageStartupEnacted, defaultTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}
	assertStartupEnactedRestartFunc(t, annotations, podStatusAnn)

	podStatusAnn, errs = csaWaitStatusAll(namespace, names, csaStatusMessagePostStartupEnacted, defaultTimeoutSecs)
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
	podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation,
	expectStartupProbe bool,
	expectReadinessProbe bool,
	expectStatusCommandedEnactedEmpty bool,
) {
	if (!expectStartupProbe && !expectReadinessProbe) || (expectStartupProbe && expectReadinessProbe) {
		panic(errors.New("only one of expectStartupProbe/expectReadinessProbe must be true"))
	}

	for pod, statusAnn := range podStatusAnn {
		for _, c := range pod.Spec.Containers {
			expectCpuR, expectCpuL := annotations.cpuPostStartupRequests, annotations.cpuPostStartupLimits
			expectMemoryR, expectMemoryL := annotations.memoryPostStartupRequests, annotations.memoryPostStartupLimits

			if c.Name == echoServerName {
				expectCpuR, expectCpuL = annotations.cpuStartup, annotations.cpuStartup
				expectMemoryR, expectMemoryL = annotations.memoryStartup, annotations.memoryStartup
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

		for _, s := range pod.Status.ContainerStatuses {
			expectCpuA, expectMemoryA := annotations.cpuPostStartupRequests, annotations.memoryPostStartupRequests
			expectCpuR, expectCpuL := annotations.cpuPostStartupRequests, annotations.cpuPostStartupLimits
			expectMemoryR, expectMemoryL := annotations.memoryPostStartupRequests, annotations.memoryPostStartupLimits

			if s.Name == echoServerName {
				expectCpuA, expectMemoryA = annotations.cpuStartup, annotations.memoryStartup
				expectCpuR, expectCpuL = annotations.cpuStartup, annotations.cpuStartup
				expectMemoryR, expectMemoryL = annotations.memoryStartup, annotations.memoryStartup

				// See comment in targetcontaineraction.go
				if expectStartupProbe {
					require.False(t, *s.Started)
				} else {
					require.True(t, *s.Started)
				}
				require.False(t, s.Ready)
			}

			require.NotNil(t, s.State.Running)
			cpuA := s.AllocatedResources[v1.ResourceCPU]
			require.Equal(t, expectCpuA, cpuA.String())
			memoryA := s.AllocatedResources[v1.ResourceMemory]
			require.Equal(t, expectMemoryA, memoryA.String())
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
		require.Equal(t, podcommon.StateAllocatedResourcesContainerRequestsMatch, statusAnn.States.AllocatedResources)
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
	podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation,
	expectStartupProbe bool,
	expectReadinessProbe bool,
) {
	for pod, statusAnn := range podStatusAnn {
		for _, c := range pod.Spec.Containers {
			expectCpuR, expectCpuL := annotations.cpuPostStartupRequests, annotations.cpuPostStartupLimits
			expectMemoryR, expectMemoryL := annotations.memoryPostStartupRequests, annotations.memoryPostStartupLimits

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

		for _, s := range pod.Status.ContainerStatuses {
			expectCpuA, expectMemoryA := annotations.cpuPostStartupRequests, annotations.memoryPostStartupRequests
			expectCpuR, expectCpuL := annotations.cpuPostStartupRequests, annotations.cpuPostStartupLimits
			expectMemoryR, expectMemoryL := annotations.memoryPostStartupRequests, annotations.memoryPostStartupLimits

			if s.Name == echoServerName {
				// See comment in targetcontaineraction.go
				require.True(t, *s.Started)
				require.True(t, s.Ready)
			}

			require.NotNil(t, s.State.Running)
			cpuA := s.AllocatedResources[v1.ResourceCPU]
			require.Equal(t, expectCpuA, cpuA.String())
			memoryA := s.AllocatedResources[v1.ResourceMemory]
			require.Equal(t, expectMemoryA, memoryA.String())
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
		require.Equal(t, podcommon.StateAllocatedResourcesContainerRequestsMatch, statusAnn.States.AllocatedResources)
		require.Equal(t, podcommon.StateStatusResourcesContainerResourcesMatch, statusAnn.States.StatusResources)

		require.NotEmpty(t, statusAnn.Scale.LastCommanded)
		require.NotEmpty(t, statusAnn.Scale.LastEnacted)
		require.Empty(t, statusAnn.Scale.LastFailed)
	}
}

func ensureEvents(t *testing.T, reason string, substrs []string, namespace string, names []string) {
	for _, name := range names {
		messages, err := kubeGetEventMessages(namespace, name, reason)
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
	if deleteNsPostTest {
		t.Cleanup(func() {
			_ = kubeDeleteNamespace(namespace)
		})
	}
}

func maybeLogErrAndFailNow(t *testing.T, err error) {
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}
