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

# Based on https://www.thapaliya.com/en/writings/well-documented-makefiles/

.DEFAULT_GOAL:=help
SHELL:=/bin/bash
ROOT_DIR:=$(dir $(realpath $(lastword $(MAKEFILE_LIST))))

INT_TESTS_TIMEOUT=30m
HELM_TESTS_SNAPSHOT_DIR=${ROOT_DIR}charts/container-startup-autoscaler/tests/__snapshot__

.PHONY: help
help: ## Displays this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

## ------------------
## Test
## ------------------

.PHONY: test-run-unit
test-run-unit: ## Runs unit tests
	go test -count=1 ./internal/...

.PHONY: test-run-int
test-run-int: ## Runs integration tests for a specific major.minor version of Kube
	@if [ -z "${KUBE_VERSION}" ]; then \
		echo "KUBE_VERSION is required - run 'make test-run-int KUBE_VERSION=x.y'"; \
		exit 1; \
	fi
	go test -count=1 -timeout ${INT_TESTS_TIMEOUT} ./test/integration/...

.PHONY: test-run-int-verbose
test-run-int-verbose: ## Runs integration tests for a specific major.minor version of Kube, with verbose logging
	@if [ -z "${KUBE_VERSION}" ]; then \
		echo "KUBE_VERSION is required - run 'make test-run-int KUBE_VERSION=x.y'"; \
		exit 1; \
	fi
	go test -count=1 -timeout ${INT_TESTS_TIMEOUT} -v ./test/integration/...

.PHONY: test-run-helm
test-run-helm: ## Runs Helm tests
	@rm -rf ${HELM_TESTS_SNAPSHOT_DIR}
	@mkdir ${HELM_TESTS_SNAPSHOT_DIR}
	@chmod 777 ${HELM_TESTS_SNAPSHOT_DIR}
	docker run -t --rm -v ${ROOT_DIR}charts:/apps helmunittest/helm-unittest:3.12.3-0.3.5 container-startup-autoscaler
	@rm -rf ${HELM_TESTS_SNAPSHOT_DIR}

## ------------------
## Go Modules
## ------------------

.PHONY: go-modules-update
go-modules-update: ## Gets latest versions of all Go modules and updates go.mod
	go get -u ./...
	go mod tidy
