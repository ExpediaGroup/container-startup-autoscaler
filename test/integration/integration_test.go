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
	"flag"
	"os"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
)

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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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
		cpuStartup:                "150m",
		cpuPostStartupRequests:    "100m",
		cpuPostStartupLimits:      "100m",
		memoryStartup:             "150M",
		memoryPostStartupRequests: "100M",
		memoryPostStartupLimits:   "100M",
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

	assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)

	podStatusAnn, errs = csaWaitStatusAll(t, namespace, names, csaStatusMessagePostStartupEnacted, testsDefaultWaitStatusTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}

	assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)

	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessageStartupCommandedUnknownRes, namespace, names)
	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessageStartupEnacted, namespace, names)
	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessagePostStartupCommanded, namespace, names)
	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessagePostStartupEnacted, namespace, names)
	assertEventCount(t, 4, namespace, names)
}

func TestDeploymentScaleWhenUnknownResourcesCpu(t *testing.T) {
	t.Parallel()
	namespace := "deployment-scale-when-unknown-resources-cpu"
	maybeRegisterCleanup(t, namespace)

	_ = kubeDeleteNamespace(t, namespace)
	maybeLogErrAndFailNow(t, kubeCreateNamespace(t, namespace))

	annotations := csaQuantityAnnotations{
		cpuStartup:             "150m",
		cpuPostStartupRequests: "100m",
		cpuPostStartupLimits:   "100m",
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

	assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)

	podStatusAnn, errs = csaWaitStatusAll(t, namespace, names, csaStatusMessagePostStartupEnacted, testsDefaultWaitStatusTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}

	assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)

	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessageStartupCommandedUnknownRes, namespace, names)
	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessageStartupEnacted, namespace, names)
	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessagePostStartupCommanded, namespace, names)
	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessagePostStartupEnacted, namespace, names)
	assertEventCount(t, 4, namespace, names)
}

func TestDeploymentScaleWhenUnknownResourcesMemory(t *testing.T) {
	t.Parallel()
	namespace := "deployment-scale-when-unknown-resources-memory"
	maybeRegisterCleanup(t, namespace)

	_ = kubeDeleteNamespace(t, namespace)
	maybeLogErrAndFailNow(t, kubeCreateNamespace(t, namespace))

	annotations := csaQuantityAnnotations{
		memoryStartup:             "150M",
		memoryPostStartupRequests: "100M",
		memoryPostStartupLimits:   "100M",
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

	assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)

	podStatusAnn, errs = csaWaitStatusAll(t, namespace, names, csaStatusMessagePostStartupEnacted, testsDefaultWaitStatusTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}

	assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)

	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessageStartupCommandedUnknownRes, namespace, names)
	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessageStartupEnacted, namespace, names)
	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessagePostStartupCommanded, namespace, names)
	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessagePostStartupEnacted, namespace, names)
	assertEventCount(t, 4, namespace, names)
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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, true, false)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, true, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
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
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertStartupEnacted(t, annotations, podStatusAnn, false, true, true)
		},
		func(t *testing.T, annotations csaQuantityAnnotations, podStatusAnn map[*v1.Pod]podcommon.StatusAnnotation) {
			assertPostStartupEnacted(t, annotations, podStatusAnn, false, true)
		},
	)
}

// Failure -------------------------------------------------------------------------------------------------------------

func TestValidationFailure(t *testing.T) {
	t.Parallel()
	namespace := "validation-failure"
	maybeRegisterCleanup(t, namespace)

	_ = kubeDeleteNamespace(t, namespace)
	maybeLogErrAndFailNow(t, kubeCreateNamespace(t, namespace))

	annotations := csaQuantityAnnotations{
		cpuStartup:             "100m",
		cpuPostStartupRequests: "150m",
		cpuPostStartupLimits:   "150m",
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
		require.Contains(t, statusAnn.Status, "cpu post-startup requests (150m) is greater than startup value (100m)")
		require.NotEmpty(t, statusAnn.LastUpdated)
		expectedScale := podcommon.StatusAnnotationScale{
			EnabledForResources: []v1.ResourceName{v1.ResourceCPU},
			LastCommanded:       "",
			LastEnacted:         "",
			LastFailed:          "",
		}
		require.Equal(t, expectedScale, statusAnn.Scale)
	}
}

func TestScaleFailureInfeasible(t *testing.T) {
	t.Parallel()
	namespace := "scale-failure-infeasible"
	maybeRegisterCleanup(t, namespace)

	_ = kubeDeleteNamespace(t, namespace)
	maybeLogErrAndFailNow(t, kubeCreateNamespace(t, namespace))

	annotations := csaQuantityAnnotations{
		cpuStartup:             "10000",
		cpuPostStartupRequests: "150m",
		cpuPostStartupLimits:   "150m",
	}

	config := echoDeploymentConfigStandardPostStartup(namespace, 2, annotations)
	maybeLogErrAndFailNow(t, kubeApplyYamlOrJsonResources(t, config.deploymentJson()))

	names, err := kubeGetPodNames(t, namespace, echoServerName)
	maybeLogErrAndFailNow(t, err)

	podStatusAnn, errs := csaWaitStatusAll(t, namespace, names, csaStatusMessageInfeasible, testsDefaultWaitStatusTimeoutSecs)
	if len(errs) > 0 {
		maybeLogErrAndFailNow(t, errs[len(errs)-1])
	}

	for _, statusAnn := range podStatusAnn {
		require.Contains(t, statusAnn.Status, "Startup scale failed - infeasible (")
		require.NotEmpty(t, statusAnn.LastUpdated)
		require.Equal(t, []v1.ResourceName{v1.ResourceCPU}, statusAnn.Scale.EnabledForResources)
		require.NotEmpty(t, statusAnn.Scale.LastCommanded)
		require.Empty(t, statusAnn.Scale.LastEnacted)
		require.NotEmpty(t, statusAnn.Scale.LastFailed)
	}

	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessageStartupCommanded, namespace, names)
	assertEvent(t, kubeEventTypeWarning, csaEventReasonScaling, csaStatusMessageInfeasible, namespace, names)
	assertEventCount(t, 2, namespace, names)
}

// Helpers -------------------------------------------------------------------------------------------------------------

func testWorkflow(
	t *testing.T,
	namespace string,
	annotations csaQuantityAnnotations,
	workloadJsonReplicasFunc func(csaQuantityAnnotations) (string, int),
	assertStartupEnactedFunc func(*testing.T, csaQuantityAnnotations, map[*v1.Pod]podcommon.StatusAnnotation),
	assertPostStartupEnactedFunc func(*testing.T, csaQuantityAnnotations, map[*v1.Pod]podcommon.StatusAnnotation),
	assertStartupEnactedRestartFunc func(*testing.T, csaQuantityAnnotations, map[*v1.Pod]podcommon.StatusAnnotation),
	assertPostStartupEnactedRestartFunc func(*testing.T, csaQuantityAnnotations, map[*v1.Pod]podcommon.StatusAnnotation),
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

	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessageStartupCommanded, namespace, names)
	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessageStartupEnacted, namespace, names)
	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessagePostStartupCommanded, namespace, names)
	assertEvent(t, kubeEventTypeNormal, csaEventReasonScaling, csaStatusMessagePostStartupEnacted, namespace, names)
	assertEventCount(t, 4, namespace, names)
}

func maybeRegisterCleanup(t *testing.T, namespace string) {
	if suppliedConfig.deleteNsPostTest {
		t.Cleanup(func() {
			_ = kubeDeleteNamespace(t, namespace)
		})
	}
}
