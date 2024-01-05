#!/usr/bin/env bash

# Copyright 2024 Expedia Group, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# kind -----------------------------------------------------------------------------------------------------------------

kind_cluster_name="csa-sandbox-cluster"

arch=$(uname -m)
case $arch in
  x86_64)
    # shellcheck disable=SC2034
    kind_image="kindest/node:v1.29.0@sha256:54a50c9354f11ce0aa56a85d2cacb1b950f85eab3fe1caf988826d1f89bf37eb"
    ;;
  arm64)
    # shellcheck disable=SC2034
    kind_image="kindest/node:v1.29.0@sha256:8ccbd8bc4d52c467f3c79eeeb434827c225600a1d7385a4b1c19d9e038c9e0c0"
    ;;
  *)
    echo "Error: architecture '$arch' not supported"
    exit 1
    ;;
esac

# shellcheck disable=SC2034
kind_kubeconfig="$HOME/.kube/config-$kind_cluster_name"
# shellcheck disable=SC2034
kind_container_name="$kind_cluster_name-control-plane"

# echo-server ----------------------------------------------------------------------------------------------------------

# shellcheck disable=SC2034
echo_server_docker_image_tag="ealen/echo-server:0.7.0"
# shellcheck disable=SC2034
echo_server_kube_namespace="echo-server"

# metrics-server -------------------------------------------------------------------------------------------------------

# shellcheck disable=SC2034
metrics_server_docker_image_tag="registry.k8s.io/metrics-server/metrics-server:v0.6.4"

# CSA ------------------------------------------------------------------------------------------------------------------

# shellcheck disable=SC2034
csa_name="container-startup-autoscaler"
csa_docker_image="csa"
csa_docker_tag="test"
# shellcheck disable=SC2034
csa_docker_image_tag="$csa_docker_image:$csa_docker_tag"
# shellcheck disable=SC2034
csa_helm_name="csa-sandbox"
# shellcheck disable=SC2034
csa_helm_timeout="60s"
# shellcheck disable=SC2034
csa_lease_name="csa-expediagroup-com"
