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

package pod

import (
	"fmt"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podtest"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewContainerKubeHelper(t *testing.T) {
	assert.Empty(t, newContainerKubeHelper())
}

func TestContainerKubeHelperGet(t *testing.T) {
	type args struct {
		pod  *v1.Pod
		name string
	}
	tests := []struct {
		name       string
		args       args
		wantName   string
		wantErrMsg string
	}{
		{
			name: "ContainerNotPresent",
			args: args{
				pod:  &v1.Pod{},
				name: podtest.DefaultContainerName,
			},
			wantName:   "",
			wantErrMsg: "container not present",
		},
		{
			name: "Ok",
			args: args{
				pod:  podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name: podtest.DefaultContainerName,
			},
			wantName: podtest.DefaultContainerName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newContainerKubeHelper()

			got, err := h.Get(tt.args.pod, tt.args.name)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.wantName, got.Name)
		})
	}
}

func TestContainerKubeHelperHasStartupProbe(t *testing.T) {
	type args struct {
		container *v1.Container
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "True",
			args: args{
				container: podtest.NewContainerBuilder(podtest.NewStartupContainerConfig()).StartupProbe().Build(),
			},
			want: true,
		},
		{
			name: "False",
			args: args{
				container: podtest.NewContainerBuilder(podtest.NewStartupContainerConfig()).Build(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newContainerKubeHelper()
			assert.Equal(t, tt.want, h.HasStartupProbe(tt.args.container))
		})
	}
}

func TestContainerKubeHelperHasReadinessProbe(t *testing.T) {
	type args struct {
		container *v1.Container
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "True",
			args: args{
				container: podtest.NewContainerBuilder(podtest.NewStartupContainerConfig()).ReadinessProbe().Build(),
			},
			want: true,
		},
		{
			name: "False",
			args: args{
				container: podtest.NewContainerBuilder(podtest.NewStartupContainerConfig()).Build(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newContainerKubeHelper()
			assert.Equal(t, tt.want, h.HasReadinessProbe(tt.args.container))
		})
	}
}

func TestContainerKubeHelperState(t *testing.T) {
	type args struct {
		pod  *v1.Pod
		name string
	}
	tests := []struct {
		name       string
		args       args
		want       v1.ContainerState
		wantErrMsg string
	}{
		{
			name: "UnableToGetContainerStatus",
			args: args{
				pod:  &v1.Pod{},
				name: podtest.DefaultContainerName,
			},
			want:       v1.ContainerState{},
			wantErrMsg: "unable to get container status",
		},
		{
			name: "Ok",
			args: args{
				pod:  podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name: podtest.DefaultContainerName,
			},
			want: podtest.DefaultPodStatusContainerState,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newContainerKubeHelper()

			got, err := h.State(tt.args.pod, tt.args.name)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerKubeHelperIsStarted(t *testing.T) {
	type args struct {
		pod  *v1.Pod
		name string
	}
	tests := []struct {
		name       string
		args       args
		want       bool
		wantErrMsg string
	}{
		{
			name: "UnableToGetContainerStatus",
			args: args{
				pod:  &v1.Pod{},
				name: podtest.DefaultContainerName,
			},
			want:       false,
			wantErrMsg: "unable to get container status",
		},
		{
			name: "True",
			args: args{
				pod:  podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolTrue, podcommon.StateBoolTrue)).Build(),
				name: podtest.DefaultContainerName,
			},
			want: true,
		},
		{
			name: "FalseNotNil",
			args: args{
				pod:  podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name: podtest.DefaultContainerName,
			},
			want: false,
		},
		{
			name: "FalseNil",
			args: args{
				pod: podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
					NilContainerStatusStarted().
					Build(),
				name: podtest.DefaultContainerName,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newContainerKubeHelper()

			got, err := h.IsStarted(tt.args.pod, tt.args.name)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerKubeHelperIsReady(t *testing.T) {
	type args struct {
		pod  *v1.Pod
		name string
	}
	tests := []struct {
		name       string
		args       args
		want       bool
		wantErrMsg string
	}{
		{
			name: "UnableToGetContainerStatus",
			args: args{
				pod:  &v1.Pod{},
				name: podtest.DefaultContainerName,
			},
			want:       false,
			wantErrMsg: "unable to get container status",
		},
		{
			name: "True",
			args: args{
				pod:  podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolTrue, podcommon.StateBoolTrue)).Build(),
				name: podtest.DefaultContainerName,
			},
			want: true,
		},
		{
			name: "False",
			args: args{
				pod:  podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name: podtest.DefaultContainerName,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newContainerKubeHelper()

			got, err := h.IsReady(tt.args.pod, tt.args.name)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerKubeHelperRequests(t *testing.T) {
	type args struct {
		container    *v1.Container
		resourceName v1.ResourceName
	}
	tests := []struct {
		name            string
		args            args
		want            resource.Quantity
		wantPanicErrMsg string
	}{
		{
			name: "NilRequests",
			args: args{
				container: podtest.NewContainerBuilder(podtest.NewStartupContainerConfig()).
					NilRequests().Build(),
				resourceName: v1.ResourceCPU,
			},
			want: resource.Quantity{},
		},
		{
			name: "Cpu",
			args: args{
				container:    podtest.NewContainerBuilder(podtest.NewStartupContainerConfig()).Build(),
				resourceName: v1.ResourceCPU,
			},
			want: podtest.PodAnnotationCpuStartupQuantity,
		},
		{
			name: "Memory",
			args: args{
				container:    podtest.NewContainerBuilder(podtest.NewStartupContainerConfig()).Build(),
				resourceName: v1.ResourceMemory,
			},
			want: podtest.PodAnnotationMemoryStartupQuantity,
		},
		{
			name: "ResourceNameNotSupported",
			args: args{
				container:    podtest.NewContainerBuilder(podtest.NewStartupContainerConfig()).Build(),
				resourceName: v1.ResourceConfigMaps,
			},
			wantPanicErrMsg: fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newContainerKubeHelper()

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() { _ = h.Requests(tt.args.container, tt.args.resourceName) })
				return
			}

			got := h.Requests(tt.args.container, tt.args.resourceName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerKubeHelperLimits(t *testing.T) {
	type args struct {
		container    *v1.Container
		resourceName v1.ResourceName
	}
	tests := []struct {
		name            string
		args            args
		want            resource.Quantity
		wantPanicErrMsg string
	}{
		{
			name: "NilLimits",
			args: args{
				container: podtest.NewContainerBuilder(podtest.NewStartupContainerConfig()).
					NilLimits().Build(),
				resourceName: v1.ResourceCPU,
			},
			want: resource.Quantity{},
		},
		{
			name: "Cpu",
			args: args{
				container:    podtest.NewContainerBuilder(podtest.NewStartupContainerConfig()).Build(),
				resourceName: v1.ResourceCPU,
			},
			want: podtest.PodAnnotationCpuStartupQuantity,
		},
		{
			name: "Memory",
			args: args{
				container:    podtest.NewContainerBuilder(podtest.NewStartupContainerConfig()).Build(),
				resourceName: v1.ResourceMemory,
			},
			want: podtest.PodAnnotationMemoryStartupQuantity,
		},
		{
			name: "ResourceNameNotSupported",
			args: args{
				container:    podtest.NewContainerBuilder(podtest.NewStartupContainerConfig()).Build(),
				resourceName: v1.ResourceConfigMaps,
			},
			wantPanicErrMsg: fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newContainerKubeHelper()

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() { _ = h.Limits(tt.args.container, tt.args.resourceName) })
				return
			}

			got := h.Limits(tt.args.container, tt.args.resourceName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerKubeHelperResizePolicy(t *testing.T) {
	type args struct {
		container    *v1.Container
		resourceName v1.ResourceName
	}
	tests := []struct {
		name            string
		args            args
		want            v1.ResourceResizeRestartPolicy
		wantErrMsg      string
		wantPanicErrMsg string
	}{
		{
			name: "ContainerResizePolicyNull",
			args: args{
				container: podtest.NewContainerBuilder(podtest.NewStartupContainerConfig()).
					NilResizePolicy().Build(),
				resourceName: v1.ResourceCPU,
			},
			want:       v1.ResourceResizeRestartPolicy(""),
			wantErrMsg: "container resize policy is null",
		},
		{
			name: "Ok",
			args: args{
				container:    podtest.NewContainerBuilder(podtest.NewStartupContainerConfig()).Build(),
				resourceName: v1.ResourceCPU,
			},
			want: v1.NotRequired,
		},
		{
			name: "ResourceNameNotSupported",
			args: args{
				container:    podtest.NewContainerBuilder(podtest.NewStartupContainerConfig()).Build(),
				resourceName: v1.ResourceConfigMaps,
			},
			wantPanicErrMsg: fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newContainerKubeHelper()

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() { _, _ = h.ResizePolicy(tt.args.container, tt.args.resourceName) })
				return
			}

			got, err := h.ResizePolicy(tt.args.container, tt.args.resourceName)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerKubeHelperAllocatedResources(t *testing.T) {
	type args struct {
		pod          *v1.Pod
		name         string
		resourceName v1.ResourceName
	}
	tests := []struct {
		name            string
		args            args
		want            resource.Quantity
		wantErrMsg      string
		wantPanicErrMsg string
	}{
		{
			name: "UnableToGetContainerStatus",
			args: args{
				pod:          &v1.Pod{},
				name:         podtest.DefaultContainerName,
				resourceName: v1.ResourceCPU,
			},
			want:       resource.Quantity{},
			wantErrMsg: "unable to get container status",
		},
		{
			name: "NilAllocatedResources",
			args: args{

				pod: podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
					NilStatusAllocatedResources().Build(),
				name:         podtest.DefaultContainerName,
				resourceName: v1.ResourceCPU,
			},
			want:       resource.Quantity{},
			wantErrMsg: "container status allocated resources not present",
		},
		{
			name: "Cpu",
			args: args{
				pod:          podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name:         podtest.DefaultContainerName,
				resourceName: v1.ResourceCPU,
			},
			want: podtest.PodAnnotationCpuStartupQuantity,
		},
		{
			name: "Memory",
			args: args{
				pod:          podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name:         podtest.DefaultContainerName,
				resourceName: v1.ResourceMemory,
			},
			want: podtest.PodAnnotationMemoryStartupQuantity,
		},
		{
			name: "ResourceNameNotSupported",
			args: args{
				pod:          podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name:         podtest.DefaultContainerName,
				resourceName: v1.ResourceConfigMaps,
			},
			wantPanicErrMsg: fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newContainerKubeHelper()

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() {
					_, _ = h.AllocatedResources(tt.args.pod, tt.args.name, tt.args.resourceName)
				})
				return
			}

			got, err := h.AllocatedResources(tt.args.pod, tt.args.name, tt.args.resourceName)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerKubeHelperCurrentRequests(t *testing.T) {
	type args struct {
		pod          *v1.Pod
		name         string
		resourceName v1.ResourceName
	}
	tests := []struct {
		name            string
		args            args
		want            resource.Quantity
		wantErrMsg      string
		wantPanicErrMsg string
	}{
		{
			name: "UnableToGetContainerStatus",
			args: args{
				pod:          &v1.Pod{},
				name:         podtest.DefaultContainerName,
				resourceName: v1.ResourceCPU,
			},
			want:       resource.Quantity{},
			wantErrMsg: "unable to get container status",
		},
		{
			name: "StatusResourcesNil",
			args: args{
				pod: podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
					NilContainerStatusResources().Build(),
				name:         podtest.DefaultContainerName,
				resourceName: v1.ResourceCPU,
			},
			want:       resource.Quantity{},
			wantErrMsg: "container status resources not present",
		},
		{
			name: "Cpu",
			args: args{
				pod:          podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name:         podtest.DefaultContainerName,
				resourceName: v1.ResourceCPU,
			},
			want: podtest.PodAnnotationCpuStartupQuantity,
		},
		{
			name: "Memory",
			args: args{
				pod:          podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name:         podtest.DefaultContainerName,
				resourceName: v1.ResourceMemory,
			},
			want: podtest.PodAnnotationMemoryStartupQuantity,
		},
		{
			name: "ResourceNameNotSupported",
			args: args{
				pod:          podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name:         podtest.DefaultContainerName,
				resourceName: v1.ResourceConfigMaps,
			},
			wantPanicErrMsg: fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newContainerKubeHelper()

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() {
					_, _ = h.CurrentRequests(tt.args.pod, tt.args.name, tt.args.resourceName)
				})
				return
			}

			got, err := h.CurrentRequests(tt.args.pod, tt.args.name, tt.args.resourceName)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerKubeHelperCurrentLimits(t *testing.T) {
	type args struct {
		pod          *v1.Pod
		name         string
		resourceName v1.ResourceName
	}
	tests := []struct {
		name            string
		args            args
		want            resource.Quantity
		wantErrMsg      string
		wantPanicErrMsg string
	}{
		{
			name: "UnableToGetContainerStatus",
			args: args{
				pod:          &v1.Pod{},
				name:         podtest.DefaultContainerName,
				resourceName: v1.ResourceCPU,
			},
			want:       resource.Quantity{},
			wantErrMsg: "unable to get container status",
		},
		{
			name: "StatusResourcesNil",
			args: args{
				pod: podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
					NilContainerStatusResources().Build(),
				name:         podtest.DefaultContainerName,
				resourceName: v1.ResourceCPU,
			},
			want:       resource.Quantity{},
			wantErrMsg: "container status resources not present",
		},
		{
			name: "Cpu",
			args: args{
				pod:          podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name:         podtest.DefaultContainerName,
				resourceName: v1.ResourceCPU,
			},
			want: podtest.PodAnnotationCpuStartupQuantity,
		},
		{
			name: "Memory",
			args: args{
				pod:          podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name:         podtest.DefaultContainerName,
				resourceName: v1.ResourceMemory,
			},
			want: podtest.PodAnnotationMemoryStartupQuantity,
		},
		{
			name: "ResourceNameNotSupported",
			args: args{
				pod:          podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name:         podtest.DefaultContainerName,
				resourceName: v1.ResourceConfigMaps,
			},
			wantPanicErrMsg: fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newContainerKubeHelper()

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() {
					_, _ = h.CurrentLimits(tt.args.pod, tt.args.name, tt.args.resourceName)
				})
				return
			}

			got, err := h.CurrentLimits(tt.args.pod, tt.args.name, tt.args.resourceName)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
