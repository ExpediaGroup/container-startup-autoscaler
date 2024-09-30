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
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
)

var kindKubeconfig string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	kindKubeconfig = fmt.Sprintf("%s%s.kube%sconfig-%s", home, pathSeparator, pathSeparator, kindClusterName)
}

func kindSetupCluster(t *testing.T) {
	hasExistingCluster := false

	output, _ := cmdRun(
		t,
		exec.Command("kind", "get", "clusters"),
		"getting existing kind clusters...",
		"unable to get existing kind clusters",
		true,
	)

	if output != "" {
		for _, s := range strings.Split(output, "\n") {
			if s == kindClusterName {
				hasExistingCluster = true
			}
		}
	}

	if !suppliedConfig.reuseCluster || !hasExistingCluster {
		if hasExistingCluster {
			kindCleanUpCluster(t)
		}

		kindNodeImage, err := kindImageFromKubeVersion(suppliedConfig.kubeVersion, runtime.GOARCH)
		if err != nil {
			logMessage(t, common.WrapErrorf(err, "unable to obtain kind image"))
			os.Exit(1)
		}
		logMessage(t, fmt.Sprintf("using kind node image '%s'", kindNodeImage))

		_, _ = cmdRun(
			t,
			exec.Command("kind", "create", "cluster",
				"--name", kindClusterName,
				"--config", pathAbsFromRel(kindConfigFileRelPath),
				"--image", kindNodeImage,
			),
			"creating kind cluster...",
			"unable to create kind cluster",
			true,
		)
	}

	output, _ = cmdRun(
		t,
		exec.Command("kind", "get", "kubeconfig", "--name", kindClusterName),
		"getting kind kubeconfig...",
		"unable to get kind kubeconfig",
		true,
	)

	if err := os.WriteFile(kindKubeconfig, []byte(output), 0644); err != nil {
		logMessage(t, common.WrapErrorf(err, "unable to write kubeconfig"))
		os.Exit(1)
	}

	if err := kubePrintNodeInfo(t); err != nil {
		logMessage(t, common.WrapErrorf(err, "unable to print kube node info"))
		os.Exit(1)
	}

	if suppliedConfig.installMetricsServer {
		_, _ = cmdRun(
			t,
			exec.Command("docker", "pull", metricsServerImageTag),
			"pulling metrics-server...",
			"unable to pull metrics-server",
			true,
		)

		_, _ = cmdRun(
			t,
			exec.Command("kind", "load", "docker-image", metricsServerImageTag, "--name", kindClusterName),
			"loading metrics-server into kind cluster...",
			"unable to load metrics-server into kind cluster",
			true,
		)

		if err := kubeApplyKustomizeResources(t, pathAbsFromRel(metricsServerKustomizeDirRelPath)); err != nil {
			logMessage(t, err)
			os.Exit(1)
		}

		err := kubeWaitResourceCondition(t, "kube-system", "k8s-app=metrics-server", "pod", "ready", metricsServerReadyTimeout)
		if err != nil {
			logMessage(t, err)
			os.Exit(1)
		}
	}

	_, _ = cmdRun(
		t,
		exec.Command("docker", "pull", echoServerDockerImageTag),
		"pulling echo-service...",
		"unable to pull echo-service",
		true,
	)

	_, _ = cmdRun(
		t,
		exec.Command("kind", "load", "docker-image", echoServerDockerImageTag, "--name", kindClusterName),
		"loading echo-service into kind cluster...",
		"unable to load echo-service into kind cluster",
		true,
	)
}

func kindCleanUpCluster(t *testing.T) {
	_, _ = cmdRun(
		t,
		exec.Command("kind", "delete", "cluster", "--name", kindClusterName),
		"deleting existing kind cluster...",
		"unable to delete existing kind cluster",
		false,
	)
}

func kindImageFromKubeVersion(kubeVersion, arch string) (string, error) {
	if archMap, found := k8sVersionToImage[kubeVersion]; found {
		if image, archFound := archMap[arch]; archFound {
			return image, nil
		}
		return "", fmt.Errorf("architecture '%s' not supported", arch)
	}

	return "", fmt.Errorf("kube version %s not supported", kubeVersion)
}
