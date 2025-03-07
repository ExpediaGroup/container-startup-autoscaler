/*
Copyright 2025 Expedia Group, Inc.

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

package kube

// ContainerStatusNotPresentError is an error that indicates container status is not present.
type ContainerStatusNotPresentError struct{}

func NewContainerStatusNotPresentError() error {
	return ContainerStatusNotPresentError{}
}

func (e ContainerStatusNotPresentError) Error() string {
	return "container status not present"
}

// ContainerStatusResourcesNotPresentError is an error that indicates container status resources is not present.
type ContainerStatusResourcesNotPresentError struct{}

func NewContainerStatusResourcesNotPresentError() error {
	return ContainerStatusResourcesNotPresentError{}
}

func (e ContainerStatusResourcesNotPresentError) Error() string {
	return "container status resources not present"
}
