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
		wantName   string
		wantErrMsg string
	}{
		{
			name: "ContainerNotPresent",
			args: args{
				pod:  &v1.Pod{},
				name: kubetest.DefaultContainerName,
			},
			wantName:   "",
			wantErrMsg: "container not present",
		},
		{
			name: "Ok",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(
					podcommon.StateBoolFalse,
					podcommon.StateBoolFalse,
					true,
					true,
				)).Build(),
				name: kubetest.DefaultContainerName,
			},
			wantName: kubetest.DefaultContainerName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewContainerHelper()

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
			name: "True",
			args: args{
				container: kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).
					StartupProbe().
					Build(),
			},
			want: true,
		},
		{
			name: "False",
			args: args{
				container: kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
			},
			want: false,
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
			name: "True",
			args: args{
				container: kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).
					ReadinessProbe().
					Build(),
			},
			want: true,
		},
		{
			name: "False",
			args: args{
				container: kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
			},
			want: false,
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
		want       v1.ContainerState
		wantErrMsg string
	}{
		{
			name: "UnableToGetContainerStatus",
			args: args{
				pod:       &v1.Pod{},
				container: &v1.Container{},
			},
			want:       v1.ContainerState{},
			wantErrMsg: "unable to get container status",
		},
		{
			name: "Ok",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(
					podcommon.StateBoolFalse,
					podcommon.StateBoolFalse,
					true,
					true,
				)).Build(),
				container: kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
			},
			want: kubetest.DefaultPodStatusContainerState,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewContainerHelper()

			got, err := h.State(tt.args.pod, tt.args.container)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
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
		want       bool
		wantErrMsg string
	}{
		{
			name: "UnableToGetContainerStatus",
			args: args{
				pod:       &v1.Pod{},
				container: &v1.Container{},
			},
			want:       false,
			wantErrMsg: "unable to get container status",
		},
		{
			name: "True",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(
					podcommon.StateBoolTrue,
					podcommon.StateBoolTrue,
					true,
					true,
				)).Build(),
				container: kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
			},
			want: true,
		},
		{
			name: "FalseNotNil",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(
					podcommon.StateBoolFalse,
					podcommon.StateBoolFalse,
					true,
					true,
				)).Build(),
				container: kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
			},
			want: false,
		},
		{
			name: "FalseNil",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(
					podcommon.StateBoolFalse,
					podcommon.StateBoolFalse,
					true,
					true,
				)).NilContainerStatusStarted().Build(),
				container: kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewContainerHelper()

			got, err := h.IsStarted(tt.args.pod, tt.args.container)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
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
		want       bool
		wantErrMsg string
	}{
		{
			name: "UnableToGetContainerStatus",
			args: args{
				pod:       &v1.Pod{},
				container: &v1.Container{},
			},
			want:       false,
			wantErrMsg: "unable to get container status",
		},
		{
			name: "True",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(
					podcommon.StateBoolTrue,
					podcommon.StateBoolTrue,
					true,
					true,
				)).Build(),
				container: kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
			},
			want: true,
		},
		{
			name: "False",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(
					podcommon.StateBoolFalse,
					podcommon.StateBoolFalse,
					true,
					true,
				)).Build(),
				container: kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewContainerHelper()

			got, err := h.IsReady(tt.args.pod, tt.args.container)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
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
		want            resource.Quantity
		wantPanicErrMsg string
	}{
		{
			name: "NilRequests",
			args: args{
				container: kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).
					NilRequests().
					Build(),
				resourceName: v1.ResourceCPU,
			},
			want: resource.Quantity{},
		},
		{
			name: "Cpu",
			args: args{
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceCPU,
			},
			want: kubetest.PodAnnotationCpuStartupEnabledQuantity,
		},
		{
			name: "Memory",
			args: args{
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceMemory,
			},
			want: kubetest.PodAnnotationMemoryStartupEnabledQuantity,
		},
		{
			name: "ResourceNameNotSupported",
			args: args{
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceConfigMaps,
			},
			wantPanicErrMsg: fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
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
		want            resource.Quantity
		wantPanicErrMsg string
	}{
		{
			name: "NilLimits",
			args: args{
				container: kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).
					NilLimits().
					Build(),
				resourceName: v1.ResourceCPU,
			},
			want: resource.Quantity{},
		},
		{
			name: "Cpu",
			args: args{
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceCPU,
			},
			want: kubetest.PodAnnotationCpuStartupEnabledQuantity,
		},
		{
			name: "Memory",
			args: args{
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceMemory,
			},
			want: kubetest.PodAnnotationMemoryStartupEnabledQuantity,
		},
		{
			name: "ResourceNameNotSupported",
			args: args{
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceConfigMaps,
			},
			wantPanicErrMsg: fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
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
		want            v1.ResourceResizeRestartPolicy
		wantErrMsg      string
		wantPanicErrMsg string
	}{
		{
			name: "ContainerResizePolicyNull",
			args: args{
				container: kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).
					NilResizePolicy().
					Build(),
				resourceName: v1.ResourceCPU,
			},
			want:       v1.ResourceResizeRestartPolicy(""),
			wantErrMsg: "container resize policy is null",
		},
		{
			name: "Ok",
			args: args{
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceCPU,
			},
			want: v1.NotRequired,
		},
		{
			name: "ResourceNameNotSupported",
			args: args{
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceConfigMaps,
			},
			wantPanicErrMsg: fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
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
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
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
		want            resource.Quantity
		wantErrMsg      string
		wantPanicErrMsg string
	}{
		{
			name: "UnableToGetContainerStatus",
			args: args{
				pod:          &v1.Pod{},
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceCPU,
			},
			want:       resource.Quantity{},
			wantErrMsg: "unable to get container status",
		},
		{
			name: "StatusResourcesNil",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(
					podcommon.StateBoolFalse,
					podcommon.StateBoolFalse,
					true,
					true,
				)).NilContainerStatusResources().Build(),
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceCPU,
			},
			want:       resource.Quantity{},
			wantErrMsg: "container status resources not present",
		},
		{
			name: "Cpu",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(
					podcommon.StateBoolFalse,
					podcommon.StateBoolFalse,
					true,
					true,
				)).Build(),
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceCPU,
			},
			want: kubetest.PodAnnotationCpuStartupEnabledQuantity,
		},
		{
			name: "Memory",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(
					podcommon.StateBoolFalse,
					podcommon.StateBoolFalse,
					true,
					true,
				)).Build(),
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceMemory,
			},
			want: kubetest.PodAnnotationMemoryStartupEnabledQuantity,
		},
		{
			name: "ResourceNameNotSupported",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(
					podcommon.StateBoolFalse,
					podcommon.StateBoolFalse,
					true,
					true,
				)).Build(),
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceConfigMaps,
			},
			wantPanicErrMsg: fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
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
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
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
		want            resource.Quantity
		wantErrMsg      string
		wantPanicErrMsg string
	}{
		{
			name: "UnableToGetContainerStatus",
			args: args{
				pod:          &v1.Pod{},
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceCPU,
			},
			want:       resource.Quantity{},
			wantErrMsg: "unable to get container status",
		},
		{
			name: "StatusResourcesNil",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(
					podcommon.StateBoolFalse,
					podcommon.StateBoolFalse,
					true,
					true,
				)).NilContainerStatusResources().Build(),
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceCPU,
			},
			want:       resource.Quantity{},
			wantErrMsg: "container status resources not present",
		},
		{
			name: "Cpu",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(
					podcommon.StateBoolFalse,
					podcommon.StateBoolFalse,
					true,
					true,
				)).Build(),
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceCPU,
			},
			want: kubetest.PodAnnotationCpuStartupEnabledQuantity,
		},
		{
			name: "Memory",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(
					podcommon.StateBoolFalse,
					podcommon.StateBoolFalse,
					true,
					true,
				)).Build(),
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceMemory,
			},
			want: kubetest.PodAnnotationMemoryStartupEnabledQuantity,
		},
		{
			name: "ResourceNameNotSupported",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(
					podcommon.StateBoolFalse,
					podcommon.StateBoolFalse,
					true,
					true,
				)).Build(),
				container:    kubetest.NewContainerBuilder(kubetest.NewStartupContainerConfig(true, true)).Build(),
				resourceName: v1.ResourceConfigMaps,
			},
			wantPanicErrMsg: fmt.Sprintf("resourceName '%s' not supported", v1.ResourceConfigMaps),
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
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
