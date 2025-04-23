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

import (
	"fmt"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewContainerHelper(t *testing.T) {
	assert.Empty(t, NewContainerHelper())
}

func TestContainerHelperGet(t *testing.T) {
	type args struct {
		pod  *v1.Pod
		name string
	}
	tests := []struct {
		name       string
		args       args
		wantErrMsg string
		wantName   string
	}{
		{
			"ContainerNotPresent",
			args{
				&v1.Pod{},
				kubetest.DefaultContainerName,
			},
			"container not present",
			"",
		},
		{
			"Ok",
			args{
				kubetest.NewPodBuilder().Build(),
				kubetest.DefaultContainerName,
			},
			"",
			kubetest.DefaultContainerName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewContainerHelper()

			got, err := h.Get(tt.args.pod, tt.args.name)
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantName, got.Name)
		})
	}
}

func TestContainerHelperHasStartupProbe(t *testing.T) {
	type args struct {
		container *v1.Container
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"True",
			args{kubetest.NewContainerBuilder().StartupProbe(true).Build()},
			true,
		},
		{
			"False",
			args{kubetest.NewContainerBuilder().Build()},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewContainerHelper()
			assert.Equal(t, tt.want, h.HasStartupProbe(tt.args.container))
		})
	}
}

func TestContainerHelperHasReadinessProbe(t *testing.T) {
	type args struct {
		container *v1.Container
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"True",
			args{kubetest.NewContainerBuilder().ReadinessProbe(true).Build()},
			true,
		},
		{
			"False",
			args{kubetest.NewContainerBuilder().Build()},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewContainerHelper()
			assert.Equal(t, tt.want, h.HasReadinessProbe(tt.args.container))
		})
	}
}

func TestContainerHelperState(t *testing.T) {
	type args struct {
		pod       *v1.Pod
		container *v1.Container
	}
	tests := []struct {
		name       string
		args       args
		wantErrMsg string
		want       v1.ContainerState
	}{
		{
			"UnableToGetContainerStatus",
			args{
				&v1.Pod{},
				&v1.Container{},
			},
			"unable to get container status",
			v1.ContainerState{},
		},
		{
			"Ok",
			args{
				kubetest.NewPodBuilder().Build(),
				kubetest.NewContainerBuilder().Build(),
			},
			"",
			v1.ContainerState{Running: &v1.ContainerStateRunning{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewContainerHelper()

			got, err := h.State(tt.args.pod, tt.args.container)
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerHelperIsStarted(t *testing.T) {
	type args struct {
		pod       *v1.Pod
		container *v1.Container
	}
	tests := []struct {
		name       string
		args       args
		wantErrMsg string
		want       bool
	}{
		{
			"UnableToGetContainerStatus",
			args{
				&v1.Pod{},
				&v1.Container{},
			},
			"unable to get container status",
			false,
		},
		{
			"True",
			args{
				kubetest.NewPodBuilder().StateStarted(podcommon.StateBoolTrue).StateReady(podcommon.StateBoolTrue).Build(),
				kubetest.NewContainerBuilder().Build(),
			},
			"",
			true,
		},
		{
			"FalseNotNil",
			args{
				kubetest.NewPodBuilder().Build(),
				kubetest.NewContainerBuilder().Build(),
			},
			"",
			false,
		},
		{
			"FalseNil",
			args{
				kubetest.NewPodBuilder().NilContainerStatusStarted(true).Build(),
				kubetest.NewContainerBuilder().Build(),
			},
			"",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewContainerHelper()

			got, err := h.IsStarted(tt.args.pod, tt.args.container)
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerHelperIsReady(t *testing.T) {
	type args struct {
		pod       *v1.Pod
		container *v1.Container
	}
	tests := []struct {
		name       string
		args       args
		wantErrMsg string
		want       bool
	}{
		{
			"UnableToGetContainerStatus",
			args{
				&v1.Pod{},
				&v1.Container{},
			},
			"unable to get container status",
			false,
		},
		{
			"True",
			args{
				kubetest.NewPodBuilder().StateStarted(podcommon.StateBoolTrue).StateReady(podcommon.StateBoolTrue).Build(),
				kubetest.NewContainerBuilder().Build(),
			},
			"",
			true,
		},
		{
			"False",
			args{
				kubetest.NewPodBuilder().Build(),
				kubetest.NewContainerBuilder().Build(),
			},
			"",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewContainerHelper()

			got, err := h.IsReady(tt.args.pod, tt.args.container)
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerHelperRequests(t *testing.T) {
	type args struct {
		container    *v1.Container
		resourceName v1.ResourceName
	}
	tests := []struct {
		name            string
		args            args
		wantPanicErrMsg string
		want            resource.Quantity
	}{
		{
			"NilRequests",
			args{
				kubetest.NewContainerBuilder().NilRequests(true).Build(),
				v1.ResourceCPU,
			},
			"",
			resource.Quantity{},
		},
		{
			"Cpu",
			args{
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceCPU,
			},
			"",
			kubetest.PodCpuStartupEnabled,
		},
		{
			"Memory",
			args{
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceMemory,
			},
			"",
			kubetest.PodMemoryStartupEnabled,
		},
		{
			"ResourceNameNotSupported",
			args{
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceConfigMaps,
			},
			fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
			resource.Quantity{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewContainerHelper()

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() { _ = h.Requests(tt.args.container, tt.args.resourceName) })
				return
			}

			got := h.Requests(tt.args.container, tt.args.resourceName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerHelperLimits(t *testing.T) {
	type args struct {
		container    *v1.Container
		resourceName v1.ResourceName
	}
	tests := []struct {
		name            string
		args            args
		wantPanicErrMsg string
		want            resource.Quantity
	}{
		{
			"NilLimits",
			args{
				kubetest.NewContainerBuilder().NilLimits(true).Build(),
				v1.ResourceCPU,
			},
			"",
			resource.Quantity{},
		},
		{
			"Cpu",
			args{
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceCPU,
			},
			"",
			kubetest.PodCpuStartupEnabled,
		},
		{
			"Memory",
			args{
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceMemory,
			},
			"",
			kubetest.PodMemoryStartupEnabled,
		},
		{
			"ResourceNameNotSupported",
			args{
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceConfigMaps,
			},
			fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
			resource.Quantity{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewContainerHelper()

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() { _ = h.Limits(tt.args.container, tt.args.resourceName) })
				return
			}

			got := h.Limits(tt.args.container, tt.args.resourceName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerHelperResizePolicy(t *testing.T) {
	type args struct {
		container    *v1.Container
		resourceName v1.ResourceName
	}
	tests := []struct {
		name            string
		args            args
		wantPanicErrMsg string
		wantErrMsg      string
		want            v1.ResourceResizeRestartPolicy
	}{
		{
			"ContainerResizePolicyNull",
			args{
				kubetest.NewContainerBuilder().NilResizePolicy(true).Build(),
				v1.ResourceCPU,
			},
			"",
			"container resize policy is null",
			v1.ResourceResizeRestartPolicy(""),
		},
		{
			"Ok",
			args{
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceCPU,
			},
			"",
			"",
			v1.NotRequired,
		},
		{
			"ResourceNameNotSupported",
			args{
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceConfigMaps,
			},
			fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
			"",
			v1.ResourceResizeRestartPolicy(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewContainerHelper()

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() { _, _ = h.ResizePolicy(tt.args.container, tt.args.resourceName) })
				return
			}

			got, err := h.ResizePolicy(tt.args.container, tt.args.resourceName)
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerHelperCurrentRequests(t *testing.T) {
	type args struct {
		pod          *v1.Pod
		container    *v1.Container
		resourceName v1.ResourceName
	}
	tests := []struct {
		name            string
		args            args
		wantPanicErrMsg string
		wantErrMsg      string
		want            resource.Quantity
	}{
		{
			"UnableToGetContainerStatus",
			args{
				&v1.Pod{},
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceCPU,
			},
			"",
			"unable to get container status",
			resource.Quantity{},
		},
		{
			"StatusResourcesNil",
			args{
				kubetest.NewPodBuilder().NilContainerStatusResources(true).Build(),
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceCPU,
			},
			"",
			"container status resources not present",
			resource.Quantity{},
		},
		{
			"Cpu",
			args{
				kubetest.NewPodBuilder().Build(),
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceCPU,
			},
			"",
			"",
			kubetest.PodCpuStartupEnabled,
		},
		{
			"Memory",
			args{
				kubetest.NewPodBuilder().Build(),
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceMemory,
			},
			"",
			"",
			kubetest.PodMemoryStartupEnabled,
		},
		{
			"ResourceNameNotSupported",
			args{
				kubetest.NewPodBuilder().Build(),
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceConfigMaps,
			},
			fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
			"",
			resource.Quantity{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewContainerHelper()

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() {
					_, _ = h.CurrentRequests(tt.args.pod, tt.args.container, tt.args.resourceName)
				})
				return
			}

			got, err := h.CurrentRequests(tt.args.pod, tt.args.container, tt.args.resourceName)
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainerHelperCurrentLimits(t *testing.T) {
	type args struct {
		pod          *v1.Pod
		container    *v1.Container
		resourceName v1.ResourceName
	}
	tests := []struct {
		name            string
		args            args
		wantPanicErrMsg string
		wantErrMsg      string
		want            resource.Quantity
	}{
		{
			"UnableToGetContainerStatus",
			args{
				&v1.Pod{},
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceCPU,
			},
			"",
			"unable to get container status",
			resource.Quantity{},
		},
		{
			"StatusResourcesNil",
			args{
				kubetest.NewPodBuilder().NilContainerStatusResources(true).Build(),
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceCPU,
			},
			"",
			"container status resources not present",
			resource.Quantity{},
		},
		{
			"Cpu",
			args{
				kubetest.NewPodBuilder().Build(),
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceCPU,
			},
			"",
			"",
			kubetest.PodCpuStartupEnabled,
		},
		{
			"Memory",
			args{
				kubetest.NewPodBuilder().Build(),
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceMemory,
			},
			"",
			"",
			kubetest.PodMemoryStartupEnabled,
		},
		{
			"ResourceNameNotSupported",
			args{
				kubetest.NewPodBuilder().Build(),
				kubetest.NewContainerBuilder().Build(),
				v1.ResourceConfigMaps,
			},
			fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
			"",
			resource.Quantity{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewContainerHelper()

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() {
					_, _ = h.CurrentLimits(tt.args.pod, tt.args.container, tt.args.resourceName)
				})
				return
			}

			got, err := h.CurrentLimits(tt.args.pod, tt.args.container, tt.args.resourceName)
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
