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

package scaletest

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
)

var (
	RawResourcesCpuEnabled = scalecommon.RawResources{
		Startup:             kubetest.PodAnnotationCpuStartup,
		PostStartupRequests: kubetest.PodAnnotationCpuPostStartupRequests,
		PostStartupLimits:   kubetest.PodAnnotationCpuPostStartupLimits,
	}

	ResourcesCpuEnabled = scalecommon.Resources{
		Startup:             kubetest.PodCpuStartupEnabled,
		PostStartupRequests: kubetest.PodCpuPostStartupRequestsEnabled,
		PostStartupLimits:   kubetest.PodCpuPostStartupLimitsEnabled,
	}

	RawResourcesCpuDisabled = scalecommon.RawResources{
		Startup:             kubetest.PodAnnotationCpuStartup,
		PostStartupRequests: kubetest.PodAnnotationCpuStartup,
		PostStartupLimits:   kubetest.PodAnnotationCpuStartup,
	}

	ResourcesCpuDisabled = scalecommon.Resources{
		Startup:             kubetest.PodCpuStartupDisabled,
		PostStartupRequests: kubetest.PodCpuPostStartupRequestsDisabled,
		PostStartupLimits:   kubetest.PodCpuPostStartupLimitsDisabled,
	}

	RawResourcesMemoryEnabled = scalecommon.RawResources{
		Startup:             kubetest.PodAnnotationMemoryStartup,
		PostStartupRequests: kubetest.PodAnnotationMemoryPostStartupRequests,
		PostStartupLimits:   kubetest.PodAnnotationMemoryPostStartupLimits,
	}

	ResourcesMemoryEnabled = scalecommon.Resources{
		Startup:             kubetest.PodMemoryStartupEnabled,
		PostStartupRequests: kubetest.PodMemoryPostStartupRequestsEnabled,
		PostStartupLimits:   kubetest.PodMemoryPostStartupLimitsEnabled,
	}

	RawResourcesMemoryDisabled = scalecommon.RawResources{
		Startup:             kubetest.PodAnnotationMemoryStartup,
		PostStartupRequests: kubetest.PodAnnotationMemoryStartup,
		PostStartupLimits:   kubetest.PodAnnotationMemoryStartup,
	}

	ResourcesMemoryDisabled = scalecommon.Resources{
		Startup:             kubetest.PodMemoryStartupDisabled,
		PostStartupRequests: kubetest.PodMemoryPostStartupRequestsDisabled,
		PostStartupLimits:   kubetest.PodMemoryPostStartupLimitsDisabled,
	}

	RawResourcesCpuUnknown = scalecommon.RawResources{
		Startup:             kubetest.PodAnnotationCpuUnknown,
		PostStartupRequests: kubetest.PodAnnotationCpuUnknown,
		PostStartupLimits:   kubetest.PodAnnotationCpuUnknown,
	}

	ResourcesCpuUnknown = scalecommon.Resources{
		Startup:             kubetest.PodCpuUnknown,
		PostStartupRequests: kubetest.PodCpuUnknown,
		PostStartupLimits:   kubetest.PodCpuUnknown,
	}

	RawResourcesMemoryUnknown = scalecommon.RawResources{
		Startup:             kubetest.PodAnnotationMemoryUnknown,
		PostStartupRequests: kubetest.PodAnnotationMemoryUnknown,
		PostStartupLimits:   kubetest.PodAnnotationMemoryUnknown,
	}

	ResourcesMemoryUnknown = scalecommon.Resources{
		Startup:             kubetest.PodMemoryUnknown,
		PostStartupRequests: kubetest.PodMemoryUnknown,
		PostStartupLimits:   kubetest.PodMemoryUnknown,
	}
)
