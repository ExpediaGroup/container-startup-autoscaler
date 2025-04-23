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

package integration

import (
	"errors"
	"strings"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func assertStartupEnacted(
	t *testing.T,
	annotations csaQuantityAnnotations,
	podStatusAnn map[*v1.Pod]pod.StatusAnnotation,
	expectStartupProbe bool,
	expectReadinessProbe bool,
	//expectStatusCommandedEnactedEmpty bool, // TODO(wt)
	expectStatusScaleCommanded bool,
	expectStatusScaleEnacted bool,
) {
	if (!expectStartupProbe && !expectReadinessProbe) || (expectStartupProbe && expectReadinessProbe) {
		panic(errors.New("only one of expectStartupProbe/expectReadinessProbe must be true"))
	}

	for kubePod, statusAnn := range podStatusAnn {
		for _, c := range kubePod.Spec.Containers {
			var expectCpuR, expectCpuL, expectMemoryR, expectMemoryL string

			if c.Name == echoServerName {
				expectCpuR, expectCpuL = annotations.CpuStartupRequestsLimits()
				expectMemoryR, expectMemoryL = annotations.MemoryStartupRequestsLimits()
			} else if c.Name == echoServerNonTargetContainerName {
				expectCpuR, expectCpuL = echoServerNonTargetContainerCpuRequests, echoServerNonTargetContainerCpuLimits
				expectMemoryR, expectMemoryL = echoServerNonTargetContainerMemoryRequests, echoServerNonTargetContainerMemoryLimits
			} else {
				panic(errors.New("container name unrecognized"))
			}

			if expectStartupProbe {
				require.NotNil(t, c.StartupProbe)
			} else {
				require.Nil(t, c.StartupProbe)
			}
			if expectReadinessProbe {
				require.NotNil(t, c.ReadinessProbe)
			} else {
				require.Nil(t, c.ReadinessProbe)
			}
			cpuR := c.Resources.Requests[v1.ResourceCPU]
			require.Equal(t, expectCpuR, cpuR.String())
			cpuL := c.Resources.Limits[v1.ResourceCPU]
			require.Equal(t, expectCpuL, cpuL.String())
			memoryR := c.Resources.Requests[v1.ResourceMemory]
			require.Equal(t, expectMemoryR, memoryR.String())
			memoryL := c.Resources.Limits[v1.ResourceMemory]
			require.Equal(t, expectMemoryL, memoryL.String())
		}

		for _, s := range kubePod.Status.ContainerStatuses {
			var expectCpuR, expectCpuL, expectMemoryR, expectMemoryL string

			if s.Name == echoServerName {
				expectCpuR, expectCpuL = annotations.CpuStartupRequestsLimits()
				expectMemoryR, expectMemoryL = annotations.MemoryStartupRequestsLimits()

				// See comment in targetcontaineraction.go
				if expectStartupProbe {
					require.False(t, *s.Started)
				} else {
					require.True(t, *s.Started)
				}
				require.False(t, s.Ready)
			} else if s.Name == echoServerNonTargetContainerName {
				expectCpuR, expectCpuL = echoServerNonTargetContainerCpuRequests, echoServerNonTargetContainerCpuLimits
				expectMemoryR, expectMemoryL = echoServerNonTargetContainerMemoryRequests, echoServerNonTargetContainerMemoryLimits
			} else {
				panic(errors.New("container name unrecognized"))
			}

			require.NotNil(t, s.State.Running)
			cpuR := s.Resources.Requests[v1.ResourceCPU]
			require.Equal(t, expectCpuR, cpuR.String())
			cpuL := s.Resources.Limits[v1.ResourceCPU]
			require.Equal(t, expectCpuL, cpuL.String())
			memoryR := s.Resources.Requests[v1.ResourceMemory]
			require.Equal(t, expectMemoryR, memoryR.String())
			memoryL := s.Resources.Limits[v1.ResourceMemory]
			require.Equal(t, expectMemoryL, memoryL.String())
		}

		require.Equal(t, csaStatusMessageStartupEnacted, statusAnn.Status)
		require.NotEmpty(t, statusAnn.LastUpdated)

		require.Equal(t, expectStartupProbe, statusAnn.States.StartupProbe.Bool())
		require.Equal(t, expectReadinessProbe, statusAnn.States.ReadinessProbe.Bool())
		require.Equal(t, podcommon.StateContainerRunning, statusAnn.States.Container)
		if expectStartupProbe {
			require.Equal(t, podcommon.StateBoolFalse, statusAnn.States.Started)
		} else {
			require.Equal(t, podcommon.StateBoolTrue, statusAnn.States.Started)
		}
		require.Equal(t, podcommon.StateBoolFalse, statusAnn.States.Ready)
		require.Equal(t, podcommon.StateResourcesStartup, statusAnn.States.Resources)
		require.Equal(t, podcommon.StateStatusResourcesContainerResourcesMatch, statusAnn.States.StatusResources)

		if expectStatusScaleCommanded {
			require.NotEmpty(t, statusAnn.Scale.LastCommanded)
		} else {
			require.Empty(t, statusAnn.Scale.LastCommanded)
		}

		if expectStatusScaleEnacted {
			require.NotEmpty(t, statusAnn.Scale.LastEnacted)
		} else {
			require.Empty(t, statusAnn.Scale.LastEnacted)
		}

		require.Empty(t, statusAnn.Scale.LastFailed)

		// TODO(wt)
		//if expectStatusCommandedEnactedEmpty {
		//	require.Empty(t, statusAnn.Scale.LastCommanded)
		//	require.Empty(t, statusAnn.Scale.LastEnacted)
		//} else {
		//	require.NotEmpty(t, statusAnn.Scale.LastCommanded)
		//	require.NotEmpty(t, statusAnn.Scale.LastEnacted)
		//}
		//require.Empty(t, statusAnn.Scale.LastFailed)
	}
}

func assertPostStartupEnacted(
	t *testing.T,
	annotations csaQuantityAnnotations,
	podStatusAnn map[*v1.Pod]pod.StatusAnnotation,
	expectStartupProbe bool,
	expectReadinessProbe bool,
) {
	for kubePod, statusAnn := range podStatusAnn {
		for _, c := range kubePod.Spec.Containers {
			var expectCpuR, expectCpuL, expectMemoryR, expectMemoryL string

			if c.Name == echoServerName {
				expectCpuR, expectCpuL = annotations.CpuPostStartupRequestsLimits()
				expectMemoryR, expectMemoryL = annotations.MemoryPostStartupRequestsLimits()
			} else if c.Name == echoServerNonTargetContainerName {
				expectCpuR, expectCpuL = echoServerNonTargetContainerCpuRequests, echoServerNonTargetContainerCpuLimits
				expectMemoryR, expectMemoryL = echoServerNonTargetContainerMemoryRequests, echoServerNonTargetContainerMemoryLimits
			} else {
				panic(errors.New("container name unrecognized"))
			}

			if expectStartupProbe {
				require.NotNil(t, c.StartupProbe)
			} else {
				require.Nil(t, c.StartupProbe)
			}
			if expectReadinessProbe {
				require.NotNil(t, c.ReadinessProbe)
			} else {
				require.Nil(t, c.ReadinessProbe)
			}
			cpuR := c.Resources.Requests[v1.ResourceCPU]
			require.Equal(t, expectCpuR, cpuR.String())
			cpuL := c.Resources.Limits[v1.ResourceCPU]
			require.Equal(t, expectCpuL, cpuL.String())
			memoryR := c.Resources.Requests[v1.ResourceMemory]
			require.Equal(t, expectMemoryR, memoryR.String())
			memoryL := c.Resources.Limits[v1.ResourceMemory]
			require.Equal(t, expectMemoryL, memoryL.String())
		}

		for _, s := range kubePod.Status.ContainerStatuses {
			var expectCpuR, expectCpuL, expectMemoryR, expectMemoryL string

			if s.Name == echoServerName {
				expectCpuR, expectCpuL = annotations.CpuPostStartupRequestsLimits()
				expectMemoryR, expectMemoryL = annotations.MemoryPostStartupRequestsLimits()

				// See comment in targetcontaineraction.go
				require.True(t, *s.Started)
				require.True(t, s.Ready)
			} else if s.Name == echoServerNonTargetContainerName {
				expectCpuR, expectCpuL = echoServerNonTargetContainerCpuRequests, echoServerNonTargetContainerCpuLimits
				expectMemoryR, expectMemoryL = echoServerNonTargetContainerMemoryRequests, echoServerNonTargetContainerMemoryLimits
			} else {
				panic(errors.New("container name unrecognized"))
			}

			require.NotNil(t, s.State.Running)
			cpuR := s.Resources.Requests[v1.ResourceCPU]
			require.Equal(t, expectCpuR, cpuR.String())
			cpuL := s.Resources.Limits[v1.ResourceCPU]
			require.Equal(t, expectCpuL, cpuL.String())
			memoryR := s.Resources.Requests[v1.ResourceMemory]
			require.Equal(t, expectMemoryR, memoryR.String())
			memoryL := s.Resources.Limits[v1.ResourceMemory]
			require.Equal(t, expectMemoryL, memoryL.String())
		}

		require.Equal(t, csaStatusMessagePostStartupEnacted, statusAnn.Status)
		require.NotEmpty(t, statusAnn.LastUpdated)

		require.Equal(t, expectStartupProbe, statusAnn.States.StartupProbe.Bool())
		require.Equal(t, expectReadinessProbe, statusAnn.States.ReadinessProbe.Bool())
		require.Equal(t, podcommon.StateContainerRunning, statusAnn.States.Container)
		require.Equal(t, podcommon.StateBoolTrue, statusAnn.States.Started)
		require.Equal(t, podcommon.StateBoolTrue, statusAnn.States.Ready)
		require.Equal(t, podcommon.StateResourcesPostStartup, statusAnn.States.Resources)
		require.Equal(t, podcommon.StateStatusResourcesContainerResourcesMatch, statusAnn.States.StatusResources)

		require.NotEmpty(t, statusAnn.Scale.LastCommanded)
		require.NotEmpty(t, statusAnn.Scale.LastEnacted)
		require.Empty(t, statusAnn.Scale.LastFailed)
	}
}

func assertEvents(t *testing.T, reason string, substrs []string, namespace string, names []string) {
	for _, name := range names {
		messages, err := kubeGetEventMessages(t, namespace, name, reason)
		maybeLogErrAndFailNow(t, err)

		for _, substr := range substrs {
			gotMessage := false
			for _, message := range messages {
				if strings.Contains(message, substr) {
					gotMessage = true
					break
				}
			}

			assert.True(t, gotMessage)
		}
	}
}
