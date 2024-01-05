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

package podcommon

import "k8s.io/apimachinery/pkg/api/resource"

// CpuConfig holds CPU configuration.
type CpuConfig struct {
	Startup             resource.Quantity
	PostStartupRequests resource.Quantity
	PostStartupLimits   resource.Quantity
}

func NewCpuConfig(
	startup resource.Quantity,
	postStartupRequests resource.Quantity,
	postStartupLimits resource.Quantity,
) CpuConfig {
	return CpuConfig{
		Startup:             startup,
		PostStartupRequests: postStartupRequests,
		PostStartupLimits:   postStartupLimits,
	}
}

// MemoryConfig holds memory configuration.
type MemoryConfig struct {
	Startup             resource.Quantity
	PostStartupRequests resource.Quantity
	PostStartupLimits   resource.Quantity
}

func NewMemoryConfig(
	startup resource.Quantity,
	postStartupRequests resource.Quantity,
	postStartupLimits resource.Quantity,
) MemoryConfig {
	return MemoryConfig{
		Startup:             startup,
		PostStartupRequests: postStartupRequests,
		PostStartupLimits:   postStartupLimits,
	}
}
