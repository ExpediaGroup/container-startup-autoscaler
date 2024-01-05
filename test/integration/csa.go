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
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

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

type csaQuantityAnnotations struct {
	cpuStartup                string
	cpuPostStartupRequests    string
	cpuPostStartupLimits      string
	memoryStartup             string
	memoryPostStartupRequests string
	memoryPostStartupLimits   string
}

func csaRun() error {
	csaCleanUp()

	_, err := cmdRun(
		exec.Command("docker", "build", "-t", csaDockerImageTag, rootAbsPath),
		"building csa...",
		"unable to build csa",
		false,
	)
	if err != nil {
		return err
	}

	_, err = cmdRun(
		exec.Command("kind", "load", "docker-image", csaDockerImageTag, "--name", kindClusterName),
		"loading csa into kind cluster...",
		"unable to load csa into kind cluster",
		false,
	)
	if err != nil {
		return err
	}

	_, err = cmdRun(
		exec.Command(
			"helm", "install",
			csaHelmName,
			pathAbsFromRel(csaHelmChartRelPath),
			"--create-namespace",
			"--namespace", csaHelmName,
			"--set-string", "csa.scaleWhenUnknownResources=true",
			"--set-string", "csa.logV=2",
			"--set-string", "container.image="+csaDockerImage,
			"--set-string", "container.tag="+csaDockerTag,
			"--wait",
			"--timeout", csaHelmTimeout,
			"--kubeconfig", kindKubeconfig,
		),
		"installing csa in kind cluster...",
		"unable to install csa in kind cluster",
		false,
	)
	if err != nil {
		return err
	}

	return nil
}

func csaCleanUp() {
	_, _ = cmdRun(
		exec.Command(
			"helm", "uninstall",
			csaHelmName,
			"--namespace", csaHelmName,
			"--kubeconfig", kindKubeconfig,
		),
		"uninstalling csa...",
		"unable to uninstall csa",
		false,
	)

	_ = kubeDeleteNamespace(csaHelmName)
}

func csaWaitStatus(
	podNamespace string,
	podName string,
	waitMsgContains string,
	timeoutSecs int,
) (*v1.Pod, podcommon.StatusAnnotation, error) {
	fmt.Println(fmt.Sprintf("waiting for csa status '%s' for pod '%s/%s'", waitMsgContains, podNamespace, podName))

	var retPod *v1.Pod
	retStatusAnn := podcommon.StatusAnnotation{}
	started := time.Now()
	getAgain := true

	for {
		if int(time.Now().Sub(started).Seconds()) > timeoutSecs {
			return retPod,
				retStatusAnn,
				errors.Errorf("waiting for csa status '%s' for pod '%s/%s' timed out", waitMsgContains, podNamespace, podName)
		}

		pod, err := kubeGetPod(podNamespace, podName, true)
		if err != nil {
			return retPod, retStatusAnn, err
		}

		statusAnnStr, exists := pod.Annotations[podcommon.AnnotationStatus]
		if !exists {
			fmt.Println(fmt.Sprintf("csa status for pod '%s/%s' doesn't yet exist", podNamespace, podName))
			time.Sleep(csaStatusWaitMillis * time.Millisecond)
			continue
		}

		statusAnn, err := podcommon.StatusAnnotationFromString(statusAnnStr)
		if err != nil {
			return retPod,
				retStatusAnn,
				errors.Wrapf(err, "unable to convert csa status for pod '%s/%s'", podNamespace, podName)
		}

		//fmt.Println(statusAnn.Json())

		if strings.Contains(statusAnn.Status, waitMsgContains) {
			// TODO(wt) 'In-place Update of Pod Resources' implementation bug
			//  See large comment at top of integration_test.go - need to re-get pod in case resize is restarted.
			//  Remove once fixed.
			if getAgain {
				getAgain = false
				time.Sleep(csaStatusWaitMillis * time.Millisecond)
				continue
			}

			retPod = pod
			retStatusAnn = statusAnn
			break
		}

		time.Sleep(csaStatusWaitMillis * time.Millisecond)
	}

	fmt.Println(fmt.Sprintf("got csa status message '%s' for pod '%s/%s'", waitMsgContains, podNamespace, podName))
	return retPod, retStatusAnn, nil
}

func csaWaitStatusAll(
	podNamespace string,
	podNames []string,
	waitMsgContains string,
	timeoutSecs int,
) (map[*v1.Pod]podcommon.StatusAnnotation, []error) {
	retMap := make(map[*v1.Pod]podcommon.StatusAnnotation)
	var retErrs []error
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, podName := range podNames {
		wg.Add(1)
		name := podName

		go func() {
			defer wg.Done()
			pod, statusAnn, err := csaWaitStatus(podNamespace, name, waitMsgContains, timeoutSecs)

			mutex.Lock()
			defer mutex.Unlock()
			if err != nil {
				retErrs = append(retErrs, err)
				return
			}

			retMap[pod] = statusAnn
		}()
	}

	wg.Wait()
	return retMap, retErrs
}
