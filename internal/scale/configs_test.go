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
			want: kubetest.DefaultContainerName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &configs{
				podHelper: tt.fields.podHelper,
			}
			got, err := c.TargetContainerName(&v1.Pod{})
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
	t.Run("Error", func(t *testing.T) {
		mockConfigGood := scaletest.NewMockConfig(nil)
		mockConfigBad := scaletest.NewMockConfig(func(m *scaletest.MockConfig) {
			m.On("StoreFromAnnotations", mock.Anything).Return(errors.New(""))
		})
		configs := &configs{
			cpuConfig:    mockConfigGood,
			memoryConfig: mockConfigBad,
		}

		err := configs.StoreFromAnnotationsAll(&v1.Pod{})
		assert.NotNil(t, err)
	})

	t.Run("Ok", func(t *testing.T) {
		mockConfig := scaletest.NewMockConfig(nil)
		configs := &configs{
			cpuConfig:    mockConfig,
			memoryConfig: mockConfig,
		}

		err := configs.StoreFromAnnotationsAll(&v1.Pod{})
		assert.Nil(t, err)
	})
}

func TestConfigsValidateAll(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		mockConfigGood := scaletest.NewMockConfig(nil)
		mockConfigBad := scaletest.NewMockConfig(func(m *scaletest.MockConfig) {
			m.On("Validate", mock.Anything).Return(errors.New(""))
		})
		configs := &configs{
			cpuConfig:    mockConfigGood,
			memoryConfig: mockConfigBad,
		}

		err := configs.ValidateAll(&v1.Container{})
		assert.NotNil(t, err)
	})

	t.Run("Ok", func(t *testing.T) {
		mockConfig := scaletest.NewMockConfig(nil)
		configs := &configs{
			cpuConfig:    mockConfig,
			memoryConfig: mockConfig,
		}

		err := configs.ValidateAll(&v1.Container{})
		assert.Nil(t, err)
	})
}

func TestConfigsValidateCollection(t *testing.T) {
	t.Run("NoResourcesAreConfiguredForScaling", func(t *testing.T) {
		mockConfig := scaletest.NewMockConfig(func(m *scaletest.MockConfig) {
			m.On("IsEnabled").Return(false)
		})
		configs := &configs{
			cpuConfig:    mockConfig,
			memoryConfig: mockConfig,
		}

		err := configs.ValidateCollection()
		assert.Contains(t, err.Error(), "no resources are configured for scaling")
	})

	t.Run("Ok", func(t *testing.T) {
		mockConfig := scaletest.NewMockConfig(nil)
		configs := &configs{
			cpuConfig:    mockConfig,
			memoryConfig: mockConfig,
		}

		err := configs.ValidateCollection()
		assert.Nil(t, err)
	})
}

func TestConfigsConfigFor(t *testing.T) {
	// TODO(wt) continue here
}

func TestConfigsAllConfigs(t *testing.T) {
}

func TestConfigsAllEnabledConfigs(t *testing.T) {
}

func TestConfigsAllEnabledConfigsResourceNames(t *testing.T) {
}

func TestConfigsString(t *testing.T) {
}
