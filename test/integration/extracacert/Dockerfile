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

ARG BASE_IMAGE=kindest/base:v20250214-acbabc1a

FROM ${BASE_IMAGE}
ARG EXTRA_CA_CERT_FILENAME
COPY ./${EXTRA_CA_CERT_FILENAME} /usr/local/share/ca-certificates/extra-ca-cert.crt
RUN update-ca-certificates
