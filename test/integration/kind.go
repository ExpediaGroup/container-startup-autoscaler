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
	"regexp"
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
	output, _ := cmdRun(
		t,
		exec.Command("kind", "get", "clusters"),
		"getting existing kind clusters...",
		"unable to get existing kind clusters",
		true,
	)

	hasExistingCluster := false
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

		kubeFullVersion, err := kubeFullVersionFor(suppliedConfig.kubeVersion)
		if err != nil {
			logMessage(t, common.WrapErrorf(err, "unable to obtain kube full version"))
			os.Exit(1)
		}

		dockerTag := "kindest/node:" + kubeFullVersion

		output, _ = cmdRun(
			t,
			exec.Command("docker", "images",
				"--filter", "reference="+dockerTag,
				"--format", "{{.Repository}}:{{.Tag}}",
			),
			"getting existing docker images...",
			"unable to get existing docker images",
			true,
		)

		if output == "" {
			if suppliedConfig.extraCaCertPath != "" {
				output, _ = cmdRun(
					t,
					exec.Command("kind", "build", "node-image", "--help"),
					"getting kind default node image...",
					"unable to get kind default node image",
					true,
				)

				defaultKindBaseImage := regexp.MustCompile(`kindest/base:v[0-9]+-[a-f0-9]+`).FindString(output)
				if defaultKindBaseImage == "" {
					logMessage(t, "unable to locate default base image")
					os.Exit(1)
				}

				builtKindBaseImageTag := defaultKindBaseImage + "-extracacert"

				tempDir, err := os.MkdirTemp("", "*")
				if err != nil {
					logMessage(t, common.WrapErrorf(err, "unable to create temporary directory"))
					os.Exit(1)
				}
				defer func(path string) {
					_ = os.RemoveAll(path)
				}(tempDir)

				copiedExtraCaCertFilename := "extra-ca-cert.crt"

				cert, err := os.ReadFile(suppliedConfig.extraCaCertPath)
				if err != nil {
					logMessage(t, common.WrapErrorf(err, "unable to read extra CA certificate file"))
					os.Exit(1)
				}

				if err := os.WriteFile(tempDir+pathSeparator+copiedExtraCaCertFilename, cert, 0644); err != nil {
					logMessage(t, common.WrapErrorf(err, "unable to write CA certificate file in temporary directory"))
					os.Exit(1)
				}

				_, _ = cmdRun(
					t,
					exec.Command("docker", "build",
						"-f", "extracacert/Dockerfile",
						"-t", builtKindBaseImageTag,
						"--build-arg", "BASE_IMAGE="+defaultKindBaseImage,
						"--build-arg", "EXTRA_CA_CERT_FILENAME="+copiedExtraCaCertFilename,
						tempDir,
					),
					"building kind base image...",
					"unable to build kind base image",
					true,
				)

				_, _ = cmdRun(
					t,
					exec.Command("kind", "build", "node-image",
						"--type", "release", kubeFullVersion,
						"--base-image", builtKindBaseImageTag,
						"--image", dockerTag,
					),
					"building kind node image...",
					"unable to build kind node image",
					true,
				)
			} else {
				_, _ = cmdRun(
					t,
					exec.Command("kind", "build", "node-image",
						"--type", "release", kubeFullVersion,
						"--image", dockerTag,
					),
					"building kind node image...",
					"unable to build kind node image",
					true,
				)
			}
		}

		_, _ = cmdRun(
			t,
			exec.Command("kind", "create", "cluster",
				"--name", kindClusterName,
				"--config", pathAbsFromRel(kindConfigFileRelPath),
				"--image", dockerTag,
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

func kubeFullVersionFor(kubeVersion string) (string, error) {
	if fullVersion, found := kubeVersionToFullVersion[kubeVersion]; found {
		return fullVersion, nil
	}

	return "", fmt.Errorf("kube version %s not supported", kubeVersion)
}
