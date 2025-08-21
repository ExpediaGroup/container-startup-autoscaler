#!/usr/bin/env bash

# Copyright 2025 Expedia Group, Inc.
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

# shellcheck disable=SC2034
kind_kube_version="v1.34.0-rc.2" # TODO(wt) update to v1.34.0 when available
# shellcheck disable=SC2034
kind_node_docker_tag="kindest/node:$kind_kube_version"

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
