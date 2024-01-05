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

# shellcheck disable=SC2154
container_id=$(kubectl get pod \
                       -n "$echo_server_kube_namespace" \
                       -o=jsonpath='{.items[0].status.containerStatuses[0].containerID}' \
                       --kubeconfig "$kind_kubeconfig"
)

fixed_container_id=${container_id/containerd:\/\//}

# shellcheck disable=SC2154
docker exec -it "$kind_container_name" bash -c "ctr -n k8s.io task kill -s SIGTERM $fixed_container_id"
