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

package scale

import (
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scaletest"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewUpdate(t *testing.T) {
	u := NewUpdate(v1.ResourceCPU, nil)
	expected := &update{
		resourceName: v1.ResourceCPU,
		config:       nil,
	}
	assert.Equal(t, expected, u)
}

func TestUpdateResourceName(t *testing.T) {
	resourceName := v1.ResourceCPU
	update := &update{resourceName: resourceName}
	assert.Equal(t, v1.ResourceCPU, update.ResourceName())
}

func TestStartupPodMutationFunc(t *testing.T) {
	type fields struct {
		resourceName v1.ResourceName
		config       scalecommon.Configuration
	}
	type args struct {
		container *v1.Container
		funcPod   *v1.Pod
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantErrMsg      string
		wantShouldPatch bool
		wantRequests    resource.Quantity
		wantLimits      resource.Quantity
	}{
		{
			"NotEnabled",
			fields{
				v1.ResourceCPU,
				scaletest.NewMockConfiguration(func(m *scaletest.MockConfiguration) {
					m.On("IsEnabled").Return(false)
				}),
			},
			args{
				nil,
				kubetest.NewPodBuilder().ResourcesState(podcommon.StateResourcesPostStartup).Build(),
			},
			"",
			false,
			kubetest.PodCpuPostStartupRequestsEnabled,
			kubetest.PodCpuPostStartupLimitsEnabled,
		},

		{
			"ContainerNotPreset",
			fields{
				v1.ResourceCPU,
				scaletest.NewMockConfiguration(nil),
			},
			args{
				&v1.Container{Name: ""},
				kubetest.NewPodBuilder().ResourcesState(podcommon.StateResourcesPostStartup).Build(),
			},
			"container not present",
			false,
			kubetest.PodCpuPostStartupRequestsEnabled,
			kubetest.PodCpuPostStartupLimitsEnabled,
		},
		{
			"Ok",
			fields{
				v1.ResourceCPU,
				scaletest.NewMockConfiguration(nil),
			},
			args{
				&kubetest.NewPodBuilder().ResourcesState(podcommon.StateResourcesPostStartup).Build().Spec.Containers[0],
				kubetest.NewPodBuilder().ResourcesState(podcommon.StateResourcesPostStartup).Build(),
			},
			"",
			true,
			kubetest.PodCpuStartupEnabled,
			kubetest.PodCpuStartupEnabled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update := &update{
				resourceName: tt.fields.resourceName,
				config:       tt.fields.config,
			}
			mutationFunc := update.StartupPodMutationFunc(tt.args.container)
			got, conditionsMetFunc, err := mutationFunc(tt.args.funcPod)
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantShouldPatch, got)
			assert.Nil(t, conditionsMetFunc)
			assert.Equal(t, tt.wantRequests, tt.args.funcPod.Spec.Containers[0].Resources.Requests[tt.fields.resourceName])
			assert.Equal(t, tt.wantLimits, tt.args.funcPod.Spec.Containers[0].Resources.Limits[tt.fields.resourceName])
		})
	}
}

func TestPostStartupPodMutationFunc(t *testing.T) {
	type fields struct {
		resourceName v1.ResourceName
		config       scalecommon.Configuration
	}
	type args struct {
		container *v1.Container
		funcPod   *v1.Pod
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantErrMsg      string
		wantShouldPatch bool
		wantRequests    resource.Quantity
		wantLimits      resource.Quantity
	}{
		{
			"NotEnabled",
			fields{
				v1.ResourceCPU,
				scaletest.NewMockConfiguration(func(m *scaletest.MockConfiguration) {
					m.On("IsEnabled").Return(false)
				}),
			},
			args{
				nil,
				kubetest.NewPodBuilder().Build(),
			},
			"",
			false,
			kubetest.PodCpuStartupEnabled,
			kubetest.PodCpuStartupEnabled,
		},

		{
			"ContainerNotPreset",
			fields{
				v1.ResourceCPU,
				scaletest.NewMockConfiguration(nil),
			},
			args{
				&v1.Container{Name: ""},
				kubetest.NewPodBuilder().Build(),
			},
			"container not present",
			false,
			kubetest.PodCpuStartupEnabled,
			kubetest.PodCpuStartupEnabled,
		},
		{
			"Ok",
			fields{
				v1.ResourceCPU,
				scaletest.NewMockConfiguration(nil),
			},
			args{
				&kubetest.NewPodBuilder().Build().Spec.Containers[0],
				kubetest.NewPodBuilder().Build(),
			},
			"",
			true,
			kubetest.PodCpuPostStartupRequestsEnabled,
			kubetest.PodCpuPostStartupLimitsEnabled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update := &update{
				resourceName: tt.fields.resourceName,
				config:       tt.fields.config,
			}
			mutationFunc := update.PostStartupPodMutationFunc(tt.args.container)
			got, conditionsMetFunc, err := mutationFunc(tt.args.funcPod)
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantShouldPatch, got)
			assert.Nil(t, conditionsMetFunc)
			assert.Equal(t, tt.wantRequests, tt.args.funcPod.Spec.Containers[0].Resources.Requests[tt.fields.resourceName])
			assert.Equal(t, tt.wantLimits, tt.args.funcPod.Spec.Containers[0].Resources.Limits[tt.fields.resourceName])
		})
	}
}
