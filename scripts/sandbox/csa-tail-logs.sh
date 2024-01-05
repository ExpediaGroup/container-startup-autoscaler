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
lease_holder=$(kubectl get lease "$csa_lease_name" \
                       -n "$csa_helm_name" \
                       -o=jsonpath='{.spec.holderIdentity}' \
                       --kubeconfig "$kind_kubeconfig"
)
lease_holder_pod="${lease_holder%%_*}"

# shellcheck disable=SC2154
kubectl logs "$lease_holder_pod" \
        -f \
        --tail=-1 \
        -c "$csa_name" \
        -n "$csa_helm_name" \
        --kubeconfig "$kind_kubeconfig"
