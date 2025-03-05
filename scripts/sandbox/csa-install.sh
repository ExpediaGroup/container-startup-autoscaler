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

extra_ca_cert_path=""

while [ $# -gt 0 ]; do
  case "$1" in
    --extra-ca-cert-path=*)
      extra_ca_cert_path="${1#*=}"
      ;;
    *)
      echo "Unrecognized argument: $1. Supported: --extra-ca-cert-path (optional)."
      exit 1
  esac
  shift
done

source config/vars.sh
source csa-uninstall.sh

# shellcheck disable=SC2154
if [ -z "$(docker images --filter "reference=$kind_node_docker_tag" --format '{{.Repository}}:{{.Tag}}')" ]; then
  if [ -n "$extra_ca_cert_path" ]; then
    if [ ! -e "$extra_ca_cert_path" ]; then
      echo "File supplied via --extra-ca-cert-path doesn't exist."
      exit 1
    fi

    default_kind_base_image=$(kind build node-image --help | sed -n 's/.*--base-image.*default ".*\(kindest[^"]*\)".*/\1/p')
    if [ -z "$default_kind_base_image" ]; then
      echo "Unable to locate default base image."
      exit 1
    fi

    # Gets overwritten if original base image tag is used, so alter.
    built_kind_base_image_tag="$default_kind_base_image-extracacert"

    temp_dir=$(mktemp -d)
    copied_extra_ca_cert_filename="extra-ca-cert.crt"

    cp "$extra_ca_cert_path" "$temp_dir/$copied_extra_ca_cert_filename"
    docker build \
           -f extracacert/Dockerfile \
           -t "$built_kind_base_image_tag" \
           --build-arg "BASE_IMAGE=$default_kind_base_image" \
           --build-arg "EXTRA_CA_CERT_FILENAME=$copied_extra_ca_cert_filename" \
           "$temp_dir"
    rm -rf "$temp_dir"

    kind build node-image \
         --type release "$kind_kube_version" \
         --base-image "$built_kind_base_image_tag" \
         --image "$kind_node_docker_tag"
  else
    kind build node-image \
         --type release "$kind_kube_version" \
         --image "$kind_node_docker_tag"
  fi
fi

# shellcheck disable=SC2154
kind create cluster \
     --name="$kind_cluster_name" \
     --config=config/kind.yaml \
     --image="$kind_node_docker_tag"

# shellcheck disable=SC2154
kind get kubeconfig --name "$kind_cluster_name" > "$kind_kubeconfig"

# shellcheck disable=SC2154
docker pull "$metrics_server_docker_image_tag"
kind load docker-image "$metrics_server_docker_image_tag" --name "$kind_cluster_name"
kubectl apply -k config/metricsserver --kubeconfig "$kind_kubeconfig"

# shellcheck disable=SC2154
docker pull "$echo_server_docker_image_tag"
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
