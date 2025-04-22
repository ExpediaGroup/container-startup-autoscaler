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

package scalecommon

import "k8s.io/apimachinery/pkg/api/resource"

// RawResources represents raw startup and post-started resources for a container.
type RawResources struct {
	Startup             string
	PostStartupRequests string
	PostStartupLimits   string
}

// TODO(wt) test
func NewRawResources(
	startup string,
	postStartupRequests string,
	postStartupLimits string,
) RawResources {
	return RawResources{
		Startup:             startup,
		PostStartupRequests: postStartupRequests,
		PostStartupLimits:   postStartupLimits,
	}
}

// Resources represents typed startup and post-started resources for a container.
type Resources struct {
	Startup             resource.Quantity
	PostStartupRequests resource.Quantity
	PostStartupLimits   resource.Quantity
}

func NewResources(
	startup resource.Quantity,
	postStartupRequests resource.Quantity,
	postStartupLimits resource.Quantity,
) Resources {
	return Resources{
		Startup:             startup,
		PostStartupRequests: postStartupRequests,
		PostStartupLimits:   postStartupLimits,
	}
}
