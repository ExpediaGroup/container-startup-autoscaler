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

source config/vars.sh
source csa-uninstall.sh

# shellcheck disable=SC2154
kind create cluster \
     --name="$kind_cluster_name" \
     --config=config/kind.yaml \
     --image="$kind_image"

# shellcheck disable=SC2154
kind get kubeconfig --name "$kind_cluster_name" > "$kind_kubeconfig"

# shellcheck disable=SC2154
docker pull "$metrics_server_docker_image_tag"
kind load docker-image "$metrics_server_docker_image_tag" --name "$kind_cluster_name"
kubectl apply -k config/metricsserver --kubeconfig "$kind_kubeconfig"

# shellcheck disable=SC2154
docker pull "$echo_server_docker_image_tag"
# shellcheck disable=SC2154
kind load docker-image "$echo_server_docker_image_tag" --name "$kind_cluster_name"

# shellcheck disable=SC2154
docker build -t "$csa_docker_image_tag" ../../
kind load docker-image "$csa_docker_image_tag" --name "$kind_cluster_name"

# shellcheck disable=SC2154
helm install "$csa_helm_name" "../../charts/$csa_name" \
     --create-namespace \
     --namespace "$csa_helm_name" \
     --set-string "csa.logV=2" \
     --set-string "container.image=$csa_docker_image" \
     --set-string "container.tag=$csa_docker_tag" \
     --wait \
     --timeout "$csa_helm_timeout" \
     --kubeconfig "$kind_kubeconfig" \
