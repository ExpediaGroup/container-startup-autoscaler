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

package scale

import (
	"errors"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scaletest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

func TestNewConfigs(t *testing.T) {
	configs := NewConfigs(nil, nil)
	allConfigs := configs.AllConfigs()
	assert.Equal(t, 2, len(allConfigs))
	assert.Equal(t, v1.ResourceCPU, allConfigs[0].ResourceName())
	assert.Equal(t, v1.ResourceMemory, allConfigs[1].ResourceName())
}

func TestConfigsTargetContainerName(t *testing.T) {
	type fields struct {
		podHelper kubecommon.PodHelper
	}
	tests := []struct {
		name       string
		fields     fields
		wantErrMsg string
		want       string
	}{
		{
			name: "UnableToGetTargetContainerNameAnnotationValue",
			fields: fields{
				podHelper: kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) {
					m.On("ExpectedAnnotationValueAs", mock.Anything, mock.Anything, mock.Anything).
						Return("", errors.New(""))
				}),
			},
			wantErrMsg: "unable to get '" + scalecommon.AnnotationTargetContainerName + "' annotation value",
			want:       "",
		},
		{
			name: "Ok",
			fields: fields{
				podHelper: kubetest.NewMockPodHelper(nil),
			},
			wantErrMsg: "",
			want:       kubetest.DefaultContainerName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := &configs{
				podHelper: tt.fields.podHelper,
			}
			got, err := configs.TargetContainerName(&v1.Pod{})
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfigsStoreFromAnnotationsAll(t *testing.T) {
	type fields struct {
		cpuConfig    scalecommon.Config
		memoryConfig scalecommon.Config
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Error",
			fields: fields{
				cpuConfig: scaletest.NewMockConfig(nil),
				memoryConfig: scaletest.NewMockConfig(func(m *scaletest.MockConfig) {
					m.On("StoreFromAnnotations", mock.Anything).Return(errors.New(""))
				}),
			},
			wantErr: true,
		},
		{
			name: "Ok",
			fields: fields{
				cpuConfig:    scaletest.NewMockConfig(nil),
				memoryConfig: scaletest.NewMockConfig(nil),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := &configs{
				cpuConfig:    tt.fields.cpuConfig,
				memoryConfig: tt.fields.memoryConfig,
			}
			err := configs.StoreFromAnnotationsAll(&v1.Pod{})
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestConfigsValidateAll(t *testing.T) {
	type fields struct {
		cpuConfig    scalecommon.Config
		memoryConfig scalecommon.Config
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Error",
			fields: fields{
				cpuConfig: scaletest.NewMockConfig(nil),
				memoryConfig: scaletest.NewMockConfig(func(m *scaletest.MockConfig) {
					m.On("Validate", mock.Anything).Return(errors.New(""))
				}),
			},
			wantErr: true,
		},
		{
			name: "Ok",
			fields: fields{
				cpuConfig:    scaletest.NewMockConfig(nil),
				memoryConfig: scaletest.NewMockConfig(nil),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := &configs{
				cpuConfig:    tt.fields.cpuConfig,
				memoryConfig: tt.fields.memoryConfig,
			}
			err := configs.ValidateAll(&v1.Container{})
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestConfigsValidateCollection(t *testing.T) {
	type fields struct {
		cpuConfig    scalecommon.Config
		memoryConfig scalecommon.Config
	}
	tests := []struct {
		name       string
		fields     fields
		wantErrMsg string
	}{
		{
			name: "NoResourcesAreConfiguredForScaling",
			fields: fields{
				cpuConfig: scaletest.NewMockConfig(func(m *scaletest.MockConfig) {
					m.On("IsEnabled").Return(false)
				}),
				memoryConfig: scaletest.NewMockConfig(func(m *scaletest.MockConfig) {
					m.On("IsEnabled").Return(false)
				}),
			},
			wantErrMsg: "no resources are configured for scaling",
		},
		{
			name: "Ok",
			fields: fields{
				cpuConfig:    scaletest.NewMockConfig(nil),
				memoryConfig: scaletest.NewMockConfig(nil),
			},
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := &configs{
				cpuConfig:    tt.fields.cpuConfig,
				memoryConfig: tt.fields.memoryConfig,
			}
			err := configs.ValidateCollection()
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestConfigsConfigFor(t *testing.T) {
	type fields struct {
		cpuConfig    scalecommon.Config
		memoryConfig scalecommon.Config
	}
	type args struct {
		resourceName v1.ResourceName
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantNil          bool
		wantResourceName v1.ResourceName
	}{
		{
			name: "Cpu",
			fields: fields{
				cpuConfig:    &config{resourceName: v1.ResourceCPU},
				memoryConfig: &config{resourceName: v1.ResourceMemory},
			},
			args: args{
				resourceName: v1.ResourceCPU,
			},
			wantNil:          false,
			wantResourceName: v1.ResourceCPU,
		},
		{
			name: "Memory",
			fields: fields{
				cpuConfig:    &config{resourceName: v1.ResourceCPU},
				memoryConfig: &config{resourceName: v1.ResourceMemory},
			},
			args: args{
				resourceName: v1.ResourceMemory,
			},
			wantNil:          false,
			wantResourceName: v1.ResourceMemory,
		},
		{
			name:   "Default",
			fields: fields{},
			args: args{
				resourceName: v1.ResourceName(""),
			},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := &configs{
				cpuConfig:    tt.fields.cpuConfig,
				memoryConfig: tt.fields.memoryConfig,
			}
			got := configs.ConfigFor(tt.args.resourceName)
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, tt.wantResourceName, got.ResourceName())
			}
		})
	}
}

func TestConfigsAllConfigs(t *testing.T) {
	configs := &configs{
		cpuConfig:    &config{resourceName: v1.ResourceCPU},
		memoryConfig: &config{resourceName: v1.ResourceMemory},
	}
	allConfigs := configs.AllConfigs()
	assert.Equal(t, 2, len(allConfigs))
	assert.Equal(t, v1.ResourceCPU, allConfigs[0].ResourceName())
	assert.Equal(t, v1.ResourceMemory, allConfigs[1].ResourceName())
}

func TestConfigsAllEnabledConfigs(t *testing.T) {
	configs := &configs{
		cpuConfig: &config{
			resourceName:             v1.ResourceCPU,
			csaEnabled:               false,
			hasStoredFromAnnotations: true,
			userEnabled:              false,
		},
		memoryConfig: &config{
			resourceName:             v1.ResourceMemory,
			csaEnabled:               true,
			hasStoredFromAnnotations: true,
			userEnabled:              true,
		},
	}
	allConfigs := configs.AllEnabledConfigs()
	assert.Equal(t, 1, len(allConfigs))
	assert.Equal(t, v1.ResourceMemory, allConfigs[0].ResourceName())
}

func TestConfigsAllEnabledConfigsResourceNames(t *testing.T) {
	configs := &configs{
		cpuConfig: &config{
			resourceName:             v1.ResourceCPU,
			csaEnabled:               false,
			hasStoredFromAnnotations: true,
			userEnabled:              false,
		},
		memoryConfig: &config{
			resourceName:             v1.ResourceMemory,
			csaEnabled:               true,
			hasStoredFromAnnotations: true,
			userEnabled:              true,
		},
	}
	assert.Equal(t, []v1.ResourceName{v1.ResourceMemory}, configs.AllEnabledConfigsResourceNames())
}

func TestConfigsString(t *testing.T) {
	configs := &configs{
		cpuConfig: scaletest.NewMockConfig(func(m *scaletest.MockConfig) {
			m.On("String").Return("cpuConfig")
		}),
		memoryConfig: scaletest.NewMockConfig(func(m *scaletest.MockConfig) {
			m.On("String").Return("memoryConfig")
		}),
	}
	assert.Equal(t, "cpuConfig, memoryConfig", configs.String())
}
