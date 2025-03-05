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
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	csapod "github.com/ExpediaGroup/container-startup-autoscaler/internal/pod"
	"k8s.io/api/core/v1"
)

type csaQuantityAnnotations struct {
	cpuStartup                string
	cpuPostStartupRequests    string
	cpuPostStartupLimits      string
	memoryStartup             string
	memoryPostStartupRequests string
	memoryPostStartupLimits   string
}

func (c *csaQuantityAnnotations) IsCpuSpecified() bool {
	if c.cpuStartup != "" && c.cpuPostStartupRequests != "" && c.cpuPostStartupLimits != "" {
		return true
	}

	if c.cpuStartup == "" && c.cpuPostStartupRequests == "" && c.cpuPostStartupLimits == "" {
		return false
	}

	panic(errors.New("only some of all required cpu annotations are set"))
}

func (c *csaQuantityAnnotations) IsMemorySpecified() bool {
	if c.memoryStartup != "" && c.memoryPostStartupRequests != "" && c.memoryPostStartupLimits != "" {
		return true
	}

	if c.memoryStartup == "" && c.memoryPostStartupRequests == "" && c.memoryPostStartupLimits == "" {
		return false
	}

	panic(errors.New("only some of all required memory annotations are set"))
}

func (c *csaQuantityAnnotations) CpuStartupRequestsLimits() (string, string) {
	if c.IsCpuSpecified() {
		return c.cpuStartup, c.cpuStartup
	}

	return echoServerCpuDisabledRequests, echoServerCpuDisabledLimits
}

func (c *csaQuantityAnnotations) CpuPostStartupRequestsLimits() (string, string) {
	if c.IsCpuSpecified() {
		return c.cpuPostStartupRequests, c.cpuPostStartupLimits
	}

	return echoServerCpuDisabledRequests, echoServerCpuDisabledLimits
}

func (c *csaQuantityAnnotations) MemoryStartupRequestsLimits() (string, string) {
	if c.IsMemorySpecified() {
		return c.memoryStartup, c.memoryStartup
	}

	return echoServerMemoryDisabledRequests, echoServerMemoryDisabledLimits
}

func (c *csaQuantityAnnotations) MemoryPostStartupRequestsLimits() (string, string) {
	if c.IsMemorySpecified() {
		return c.memoryPostStartupRequests, c.memoryPostStartupLimits
	}

	return echoServerMemoryDisabledRequests, echoServerMemoryDisabledLimits
}

func csaRun(t *testing.T) error {
	csaCleanUp(t)

	_, err := cmdRun(
		t,
		exec.Command("docker", "build", "-t", csaDockerImageTag, rootAbsPath),
		"building csa...",
		"unable to build csa",
		false,
	)
	if err != nil {
		return err
	}

	_, err = cmdRun(
		t,
		exec.Command("kind", "load", "docker-image", csaDockerImageTag, "--name", kindClusterName),
		"loading csa into kind cluster...",
		"unable to load csa into kind cluster",
		false,
	)
	if err != nil {
		return err
	}

	_, err = cmdRun(
		t,
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

func csaCleanUp(t *testing.T) {
	_, _ = cmdRun(
		t,
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

	_ = kubeDeleteNamespace(nil, csaHelmName)
}

func csaWaitStatus(
	t *testing.T,
	podNamespace string,
	podName string,
	waitMsgContains string,
	timeoutSecs int,
) (*v1.Pod, csapod.StatusAnnotation, error) {
	logMessage(t, fmt.Sprintf("waiting for csa status '%s' for pod '%s/%s'", waitMsgContains, podNamespace, podName))

	var retPod *v1.Pod
	retStatusAnn := csapod.StatusAnnotation{}
	started := time.Now()
	lastStatusAnnJson := ""
	getAgain := true

	for {
		if int(time.Now().Sub(started).Seconds()) > timeoutSecs {
			return retPod,
				retStatusAnn,
				fmt.Errorf(
					"waiting for csa status '%s' for pod '%s/%s' timed out - last status '%s'",
					waitMsgContains, podNamespace, podName, lastStatusAnnJson,
				)
		}

		pod, err := kubeGetPod(t, podNamespace, podName, true)
		if err != nil {
			return retPod, retStatusAnn, err
		}

		statusAnnStr, exists := pod.Annotations[kubecommon.AnnotationStatus]
		if !exists {
			logMessage(t, fmt.Sprintf("csa status for pod '%s/%s' doesn't yet exist", podNamespace, podName))
			time.Sleep(csaStatusWaitMillis * time.Millisecond)
			continue
		}

		statusAnn, err := csapod.StatusAnnotationFromString(statusAnnStr)
		if err != nil {
			return retPod,
				retStatusAnn,
				common.WrapErrorf(err, "unable to convert csa status for pod '%s/%s'", podNamespace, podName)
		}

		lastStatusAnnJson = statusAnn.Json()
		logMessage(t, fmt.Sprintf("current csa status for pod '%s/%s': %s", podNamespace, podName, lastStatusAnnJson))

		if strings.Contains(statusAnn.Status, waitMsgContains) {
			// TODO(wt) 'In-place Update of Pod Resources' implementation bug (still in Kube 1.32).
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

	logMessage(t, fmt.Sprintf("got csa status message '%s' for pod '%s/%s'", waitMsgContains, podNamespace, podName))
	return retPod, retStatusAnn, nil
}

func csaWaitStatusAll(
	t *testing.T,
	podNamespace string,
	podNames []string,
	waitMsgContains string,
	timeoutSecs int,
) (map[*v1.Pod]csapod.StatusAnnotation, []error) {
	retMap := make(map[*v1.Pod]csapod.StatusAnnotation)
	var retErrs []error
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, podName := range podNames {
		wg.Add(1)
		name := podName

		go func() {
			defer wg.Done()
			pod, statusAnn, err := csaWaitStatus(t, podNamespace, name, waitMsgContains, timeoutSecs)

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
