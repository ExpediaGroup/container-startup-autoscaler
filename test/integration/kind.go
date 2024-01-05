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

	"github.com/pkg/errors"
)

const (
	kindClusterName       = "csa-int-cluster"
	kindNodeImagex8664    = "kindest/node:v1.29.0@sha256:54a50c9354f11ce0aa56a85d2cacb1b950f85eab3fe1caf988826d1f89bf37eb"
	kindNodeImageArm64    = "kindest/node:v1.29.0@sha256:8ccbd8bc4d52c467f3c79eeeb434827c225600a1d7385a4b1c19d9e038c9e0c0"
	kindConfigFileRelPath = configDirRelPath + pathSeparator + "kind.yaml"
)

const (
	metricsServerImageTag            = "registry.k8s.io/metrics-server/metrics-server:v0.6.4"
	metricsServerKustomizeDirRelPath = configDirRelPath + pathSeparator + "metricsserver"
	metricsServerReadyTimeout        = "60s"
)

var kindKubeconfig string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	kindKubeconfig = fmt.Sprintf("%s%s.kube%sconfig-%s", home, pathSeparator, pathSeparator, kindClusterName)
}

func kindSetupCluster(reuseCluster bool, installMetricsServer bool) {
	hasExistingCluster := false

	output, _ := cmdRun(
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

	if !reuseCluster || !hasExistingCluster {
		if hasExistingCluster {
			kindCleanUpCluster()
		}

		var kindNodeImage string

		switch runtime.GOARCH {
		case "amd64":
			kindNodeImage = kindNodeImagex8664
		case "arm64":
			kindNodeImage = kindNodeImageArm64
		default:
			fmt.Println(errors.Errorf("architecture '%s' not supported", runtime.GOARCH))
			os.Exit(1)
		}

		_, _ = cmdRun(
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
		exec.Command("kind", "get", "kubeconfig", "--name", kindClusterName),
		"getting kind kubeconfig...",
		"unable to get kind kubeconfig",
		true,
	)

	if err := os.WriteFile(kindKubeconfig, []byte(output), 0644); err != nil {
		fmt.Println(errors.Wrapf(err, "unable to write kubeconfig"))
		os.Exit(1)
	}

	if err := kubePrintNodeInfo(); err != nil {
		fmt.Println(errors.Wrapf(err, "unable to print kube node info"))
		os.Exit(1)
	}

	if installMetricsServer {
		_, _ = cmdRun(
			exec.Command("docker", "pull", metricsServerImageTag),
			"pulling metrics-server...",
			"unable to pull metrics-server",
			true,
		)

		_, _ = cmdRun(
			exec.Command("kind", "load", "docker-image", metricsServerImageTag, "--name", kindClusterName),
			"loading metrics-server into kind cluster...",
			"unable to load metrics-server into kind cluster",
			true,
		)

		if err := kubeApplyKustomizeResources(pathAbsFromRel(metricsServerKustomizeDirRelPath)); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err := kubeWaitResourceCondition("kube-system", "k8s-app=metrics-server", "pod", "ready", metricsServerReadyTimeout)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	_, _ = cmdRun(
		exec.Command("docker", "pull", echoServerDockerImageTag),
		"pulling echo-service...",
		"unable to pull echo-service",
		true,
	)

	_, _ = cmdRun(
		exec.Command("kind", "load", "docker-image", echoServerDockerImageTag, "--name", kindClusterName),
		"loading echo-service into kind cluster...",
		"unable to load echo-service into kind cluster",
		true,
	)
}

func kindCleanUpCluster() {
	_, _ = cmdRun(
		exec.Command("kind", "delete", "cluster", "--name", kindClusterName),
		"deleting existing kind cluster...",
		"unable to delete existing kind cluster",
		false,
	)
}
