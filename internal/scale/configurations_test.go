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

func TestNewConfigurations(t *testing.T) {
	configs := NewConfigurations(nil, nil)
	allConfigs := configs.AllConfigurations()
	assert.Equal(t, 2, len(allConfigs))
	assert.Equal(t, v1.ResourceCPU, allConfigs[0].ResourceName())
	assert.Equal(t, v1.ResourceMemory, allConfigs[1].ResourceName())
}

func TestConfigurationsTargetContainerName(t *testing.T) {
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
			"UnableToGetTargetContainerNameAnnotationValue",
			fields{
				kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) {
					m.On("ExpectedAnnotationValueAs", mock.Anything, mock.Anything, mock.Anything).
						Return("", errors.New(""))
				}),
			},
			"unable to get '" + scalecommon.AnnotationTargetContainerName + "' annotation value",
			"",
		},
		{
			"Ok",
			fields{
				kubetest.NewMockPodHelper(nil),
			},
			"",
			kubetest.DefaultContainerName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := &configurations{
				podHelper: tt.fields.podHelper,
			}
			got, err := configs.TargetContainerName(&v1.Pod{})
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfigurationsStoreFromAnnotationsAll(t *testing.T) {
	type fields struct {
		cpuConfig    scalecommon.Configuration
		memoryConfig scalecommon.Configuration
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"Error",
			fields{
				scaletest.NewMockConfiguration(nil),
				scaletest.NewMockConfiguration(func(m *scaletest.MockConfiguration) {
					m.On("StoreFromAnnotations", mock.Anything).Return(errors.New(""))
				}),
			},
			true,
		},
		{
			"Ok",
			fields{
				scaletest.NewMockConfiguration(nil),
				scaletest.NewMockConfiguration(nil),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := &configurations{
				cpuConfig:    tt.fields.cpuConfig,
				memoryConfig: tt.fields.memoryConfig,
			}
			err := configs.StoreFromAnnotationsAll(&v1.Pod{})
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigurationsValidateAll(t *testing.T) {
	type fields struct {
		cpuConfig    scalecommon.Configuration
		memoryConfig scalecommon.Configuration
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"Error",
			fields{
				scaletest.NewMockConfiguration(nil),
				scaletest.NewMockConfiguration(func(m *scaletest.MockConfiguration) {
					m.On("Validate", mock.Anything).Return(errors.New(""))
				}),
			},
			true,
		},
		{
			"Ok",
			fields{
				scaletest.NewMockConfiguration(nil),
				scaletest.NewMockConfiguration(nil),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := &configurations{
				cpuConfig:    tt.fields.cpuConfig,
				memoryConfig: tt.fields.memoryConfig,
			}
			err := configs.ValidateAll(&v1.Container{})
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigurationsValidateCollection(t *testing.T) {
	type fields struct {
		cpuConfig    scalecommon.Configuration
		memoryConfig scalecommon.Configuration
	}
	tests := []struct {
		name       string
		fields     fields
		wantErrMsg string
	}{
		{
			"NoResourcesAreConfiguredForScaling",
			fields{
				scaletest.NewMockConfiguration(func(m *scaletest.MockConfiguration) {
					m.On("IsEnabled").Return(false)
				}),
				scaletest.NewMockConfiguration(func(m *scaletest.MockConfiguration) {
					m.On("IsEnabled").Return(false)
				}),
			},
			"no resources are configured for scaling",
		},
		{
			"Ok",
			fields{
				scaletest.NewMockConfiguration(nil),
				scaletest.NewMockConfiguration(nil),
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := &configurations{
				cpuConfig:    tt.fields.cpuConfig,
				memoryConfig: tt.fields.memoryConfig,
			}
			err := configs.ValidateCollection()
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigurationsConfigFor(t *testing.T) {
	type fields struct {
		cpuConfig    scalecommon.Configuration
		memoryConfig scalecommon.Configuration
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
			"Cpu",
			fields{
				&configuration{resourceName: v1.ResourceCPU},
				&configuration{resourceName: v1.ResourceMemory},
			},
			args{v1.ResourceCPU},
			false,
			v1.ResourceCPU,
		},
		{
			"Memory",
			fields{
				&configuration{resourceName: v1.ResourceCPU},
				&configuration{resourceName: v1.ResourceMemory},
			},
			args{v1.ResourceMemory},
			false,
			v1.ResourceMemory,
		},
		{
			"Default",
			fields{},
			args{v1.ResourceName("")},
			true,
			v1.ResourceName(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := &configurations{
				cpuConfig:    tt.fields.cpuConfig,
				memoryConfig: tt.fields.memoryConfig,
			}
			got := configs.ConfigurationFor(tt.args.resourceName)
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, tt.wantResourceName, got.ResourceName())
			}
		})
	}
}

func TestConfigurationsAllConfigs(t *testing.T) {
	configs := &configurations{
		cpuConfig:    &configuration{resourceName: v1.ResourceCPU},
		memoryConfig: &configuration{resourceName: v1.ResourceMemory},
	}
	allConfigs := configs.AllConfigurations()
	assert.Equal(t, 2, len(allConfigs))
	assert.Equal(t, v1.ResourceCPU, allConfigs[0].ResourceName())
	assert.Equal(t, v1.ResourceMemory, allConfigs[1].ResourceName())
}

func TestConfigurationsAllEnabledConfigs(t *testing.T) {
	configs := &configurations{
		cpuConfig: &configuration{
			resourceName: v1.ResourceCPU,
			csaEnabled:   false,
			hasStored:    true,
			hasValidated: true,
			userEnabled:  false,
		},
		memoryConfig: &configuration{
			resourceName: v1.ResourceMemory,
			csaEnabled:   true,
			hasStored:    true,
			hasValidated: true,
			userEnabled:  true,
		},
	}
	allConfigs := configs.AllEnabledConfigurations()
	assert.Equal(t, 1, len(allConfigs))
	assert.Equal(t, v1.ResourceMemory, allConfigs[0].ResourceName())
}

func TestConfigurationsAllEnabledConfigsResourceNames(t *testing.T) {
	configs := &configurations{
		cpuConfig: &configuration{
			resourceName: v1.ResourceCPU,
			csaEnabled:   false,
			hasStored:    true,
			hasValidated: true,
			userEnabled:  false,
		},
		memoryConfig: &configuration{
			resourceName: v1.ResourceMemory,
			csaEnabled:   true,
			hasStored:    true,
			hasValidated: true,
			userEnabled:  true,
		},
	}
	assert.Equal(t, []v1.ResourceName{v1.ResourceMemory}, configs.AllEnabledConfigurationsResourceNames())
}

func TestConfigurationsString(t *testing.T) {
	configs := &configurations{
		cpuConfig: scaletest.NewMockConfiguration(func(m *scaletest.MockConfiguration) {
			m.On("String").Return("cpuConfig")
		}),
		memoryConfig: scaletest.NewMockConfiguration(func(m *scaletest.MockConfiguration) {
			m.On("String").Return("memoryConfig")
		}),
	}
	assert.Equal(t, "cpuConfig, memoryConfig", configs.String())
}
