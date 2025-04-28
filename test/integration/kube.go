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
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"k8s.io/api/core/v1"
)

func kubePrintNodeInfo(t *testing.T) error {
	output, err := cmdRun(
		t,
		exec.Command("kubectl", "describe", "nodes", "--kubeconfig", kindKubeconfig),
		"",
		"unable to describe nodes",
		false,
		true,
	)
	if err != nil {
		return err
	}

	logMessage(t, "node information:")
	logMessage(t, output)
	return nil
}

func kubeCreateNamespace(t *testing.T, name string) error {
	_, err := cmdRun(
		t,
		exec.Command("kubectl", "create", "namespace", name, "--kubeconfig", kindKubeconfig),
		fmt.Sprintf("creating namespace '%s'...", name),
		fmt.Sprintf("unable to create namespace '%s'", name),
		false,
	)
	return err
}

func kubeDeleteNamespace(t *testing.T, name string) error {
	_, err := cmdRun(
		t,
		exec.Command("kubectl", "delete", "namespace", name, "--kubeconfig", kindKubeconfig),
		fmt.Sprintf("deleting namespace '%s'...", name),
		fmt.Sprintf("unable to delete namespace '%s'", name),
		false,
	)
	return err
}

func kubeApplyYamlOrJsonResources(t *testing.T, yamlOrJson string) error {
	cmd := exec.Command("kubectl", "apply", "-f", "-", "--kubeconfig", kindKubeconfig)
	cmd.Stdin = strings.NewReader(yamlOrJson)
	_, err := cmdRun(
		t,
		cmd,
		fmt.Sprintf("applying resources '%s'...", yamlOrJson),
		fmt.Sprintf("unable to apply resources '%s'", yamlOrJson),
		false,
	)
	return err
}

func kubeApplyKustomizeResources(t *testing.T, kPath string) error {
	_, err := cmdRun(
		t,
		exec.Command("kubectl", "apply", "-k", kPath, "--kubeconfig", kindKubeconfig),
		fmt.Sprintf("applying kustomize resources from '%s'...", kPath),
		fmt.Sprintf("unable to apply kustomize resources from '%s'...", kPath),
		false,
	)
	return err
}

func kubeGetPodNames(t *testing.T, namespace string, nameContains string, suppressInfo ...bool) ([]string, error) {
	output, err := cmdRun(
		t,
		exec.Command(
			"kubectl", "get", "pods",
			"-n", namespace,
			"--no-headers=true",
			"--output=custom-columns=NAME:.metadata.name",
			"--kubeconfig", kindKubeconfig,
		),
		fmt.Sprintf("getting pod names for namespace '%s', name contains: '%s'...", namespace, nameContains),
		fmt.Sprintf("unable to get pod names for namespace '%s', name contains: '%s'...", namespace, nameContains),
		false,
		suppressInfo...,
	)
	if err != nil {
		return nil, err
	}

	var ret []string
	for _, s := range strings.Split(output, "\n") {
		if strings.Contains(s, nameContains) {
			ret = append(ret, s)
		}
	}

	return ret, nil
}

func kubeGetPod(t *testing.T, namespace string, name string, suppressInfo ...bool) (*v1.Pod, error) {
	output, err := cmdRun(
		t,
		exec.Command(
			"kubectl", "get", "pod",
			name,
			"-n", namespace,
			"-o", "json",
			"--kubeconfig", kindKubeconfig,
		),
		fmt.Sprintf("getting pod '%s/%s'...", namespace, name),
		fmt.Sprintf("unable to get pod '%s/%s'", namespace, name),
		false,
		suppressInfo...,
	)
	if err != nil {
		return nil, err
	}

	pod := &v1.Pod{}
	_ = json.Unmarshal([]byte(output), pod)
	return pod, err
}

func kubeWaitPodsExist(t *testing.T, namespace string, nameContains string, count int, timeoutSecs int) error {
	logMessage(t, fmt.Sprintf(
		"waiting for %d pods (pod name contains '%s') to exist in namespace '%s'",
		count, nameContains, namespace,
	))

	started := time.Now()

	for {
		if int(time.Now().Sub(started).Seconds()) > timeoutSecs {
			return fmt.Errorf("waiting for %d pods (pod name contains '%s') to exist in namespace '%s' timed out",
				timeoutSecs, nameContains, namespace,
			)
		}

		pods, err := kubeGetPodNames(t, namespace, nameContains, true)
		if err != nil {
			return nil
		}

		if len(pods) == count {
			break
		}

		time.Sleep(csaStatusWaitMillis * time.Millisecond)
	}

	logMessage(t, fmt.Sprintf(
		"%d pods (pod name contains '%s') now exist in namespace '%s'",
		count, nameContains, namespace,
	))

	return nil
}

func kubeWaitResourceCondition(
	t *testing.T,
	namespace string,
	label string,
	resource string,
	condition string,
	timeout string,
) error {
	_, err := cmdRun(
		t,
		exec.Command(
			"kubectl",
			"wait",
			"--for=condition="+condition,
			resource,
			"-l", label,
			"-n", namespace,
			"--timeout="+timeout,
			"--kubeconfig", kindKubeconfig,
		),
		fmt.Sprintf(
			"waiting for condition '%s' on resource '%s' with label '%s' in namespace '%s'...",
			condition, resource, label, namespace,
		),
		fmt.Sprintf(
			"unable to wait for condition '%s' on resource '%s' with label '%s' in namespace '%s'",
			condition, resource, label, namespace,
		),
		false,
	)
	return err
}

func kubeGetEventMessages(t *testing.T, namespace string, podName string, eType string, reason string) ([]string, error) {
	output, err := cmdRun(
		t,
		exec.Command(
			"kubectl", "get", "events",
			"-n", namespace,
			fmt.Sprintf("--field-selector=involvedObject.name=%s,type=%s,reason=%s", podName, eType, reason),
			"--no-headers=true",
			"--output=custom-columns=MESSAGE:.message",
			"--kubeconfig", kindKubeconfig,
		),
		fmt.Sprintf("getting '%s'/'%s' events for pod '%s' in namespace '%s'...", eType, reason, podName, namespace),
		fmt.Sprintf("unable to get '%s'/'%s' events for pod '%s' in namespace '%s'...", eType, reason, podName, namespace),
		false,
	)
	if err != nil {
		return nil, err
	}

	splitOutput := strings.Split(output, "\n")
	for _, line := range splitOutput {
		logMessage(
			t,
			fmt.Sprintf("got '%s/%s' event message for pod '%s' in namespace '%s': %s", eType, reason, podName, namespace, line),
		)
	}

	return splitOutput, nil
}

func kubeGetCsaEventCount(t *testing.T, namespace string, podName string) (int, error) {
	output, err := cmdRun(
		t,
		exec.Command(
			"kubectl", "get", "events",
			"-n", namespace,
			fmt.Sprintf("--field-selector=involvedObject.name=%s,reason=%s", podName, csaEventReasonScaling),
			"--no-headers=true",
			"--output=custom-columns=MESSAGE:.message",
			"--kubeconfig", kindKubeconfig,
		),
		fmt.Sprintf("getting all csa events for pod '%s' in namespace '%s'...", podName, namespace),
		fmt.Sprintf("unable to get all csa events for pod '%s' in namespace '%s'...", podName, namespace),
		false,
	)
	if err != nil {
		return 0, err
	}

	return len(strings.Split(output, "\n")), nil
}

func kubeCauseContainerRestart(t *testing.T, containerId string) error {
	fixedContainerId := strings.ReplaceAll(containerId, "containerd://", "")

	_, err := cmdRun(
		t,
		exec.Command(
			"docker", "exec", "-i", kindClusterName+"-control-plane",
			"bash", "-c", "ctr -n k8s.io task kill -s SIGTERM "+fixedContainerId,
		),
		fmt.Sprintf("causing restart of container '%s'...", fixedContainerId),
		fmt.Sprintf("unable to cause restart of container '%s'...", fixedContainerId),
		false,
	)
	if err != nil {
		return err
	}

	return nil
}
