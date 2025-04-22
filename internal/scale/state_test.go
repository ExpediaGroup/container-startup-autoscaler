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
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewState(t *testing.T) {
	resourceName := v1.ResourceCPU
	state := NewState(resourceName, nil, nil)
	assert.Equal(t, resourceName, state.ResourceName())
}

func TestStateResourceName(t *testing.T) {
	resourceName := v1.ResourceCPU
	state := &state{resourceName: resourceName}
	assert.Equal(t, v1.ResourceCPU, state.ResourceName())
}

func TestStateIsStartupConfigApplied(t *testing.T) {
	type fields struct {
		config          scalecommon.Configuration
		containerHelper kubecommon.ContainerHelper
	}
	tests := []struct {
		name   string
		fields fields
		want   *bool
	}{
		{
			name: "NotEnabled",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  false,
				},
			},
			want: nil,
		},
		{
			name: "True",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  true,
					resources:    scaletest.ResourcesCpuEnabled,
				},
				containerHelper: kubetest.NewMockContainerHelper(nil),
			},
			want: func() *bool { b := true; return &b }(),
		},
		{
			name: "False",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  true,
					resources:    scaletest.ResourcesCpuEnabled,
				},
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("Requests", mock.Anything, mock.Anything).Return(kubetest.PodCpuPostStartupRequestsEnabled)
					m.LimitsDefault()
				}),
			},
			want: func() *bool { b := false; return &b }(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &state{
				resourceName:    v1.ResourceCPU,
				config:          tt.fields.config,
				containerHelper: tt.fields.containerHelper,
			}
			assert.Equal(t, tt.want, s.IsStartupConfigurationApplied(&v1.Container{}))
		})
	}
}

func TestStateIsPostStartupConfigApplied(t *testing.T) {
	type fields struct {
		config          scalecommon.Configuration
		containerHelper kubecommon.ContainerHelper
	}
	tests := []struct {
		name   string
		fields fields
		want   *bool
	}{
		{
			name: "NotEnabled",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  false,
				},
			},
			want: nil,
		},
		{
			name: "True",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  true,
					resources:    scaletest.ResourcesCpuEnabled,
				},
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("Requests", mock.Anything, mock.Anything).Return(kubetest.PodCpuPostStartupRequestsEnabled)
					m.On("Limits", mock.Anything, mock.Anything).Return(kubetest.PodCpuPostStartupLimitsEnabled)
				}),
			},
			want: func() *bool { b := true; return &b }(),
		},
		{
			name: "False",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  true,
					resources:    scaletest.ResourcesCpuEnabled,
				},
				containerHelper: kubetest.NewMockContainerHelper(nil),
			},
			want: func() *bool { b := false; return &b }(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &state{
				resourceName:    v1.ResourceCPU,
				config:          tt.fields.config,
				containerHelper: tt.fields.containerHelper,
			}
			assert.Equal(t, tt.want, s.IsPostStartupConfigurationApplied(&v1.Container{}))
		})
	}

}

func TestStateIsAnyCurrentZero(t *testing.T) {
	type fields struct {
		config          scalecommon.Configuration
		containerHelper kubecommon.ContainerHelper
	}
	tests := []struct {
		name       string
		fields     fields
		wantErrMsg string
		want       *bool
	}{
		{
			name: "NotEnabled",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  false,
				},
			},
			wantErrMsg: "",
			want:       nil,
		},
		{
			name: "UnableToGetCpuCurrentRequests",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  true,
				},
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("CurrentRequests", mock.Anything, mock.Anything, mock.Anything).
						Return(resource.Quantity{}, errors.New(""))
				}),
			},
			wantErrMsg: "unable to get cpu current requests",
			want:       func() *bool { b := false; return &b }(),
		},
		{
			name: "UnableToGetCpuCurrentLimits",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  true,
				},
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("CurrentLimits", mock.Anything, mock.Anything, mock.Anything).
						Return(resource.Quantity{}, errors.New(""))
					m.CurrentRequestsDefault()
				}),
			},
			wantErrMsg: "unable to get cpu current limits",
			want:       func() *bool { b := false; return &b }(),
		},
		{
			name: "True",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  true,
				},
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("CurrentRequests", mock.Anything, mock.Anything, mock.Anything).
						Return(resource.Quantity{}, nil)
					m.On("CurrentLimits", mock.Anything, mock.Anything, mock.Anything).
						Return(resource.Quantity{}, nil)
				}),
			},
			wantErrMsg: "",
			want:       func() *bool { b := true; return &b }(),
		},
		{
			name: "False",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  true,
				},
				containerHelper: kubetest.NewMockContainerHelper(nil),
			},
			wantErrMsg: "",
			want:       func() *bool { b := false; return &b }(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &state{
				resourceName:    v1.ResourceCPU,
				config:          tt.fields.config,
				containerHelper: tt.fields.containerHelper,
			}
			got, err := s.IsAnyCurrentZero(&v1.Pod{}, &v1.Container{})
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStateDoesRequestsCurrentMatchSpec(t *testing.T) {
	type fields struct {
		config          scalecommon.Configuration
		containerHelper kubecommon.ContainerHelper
	}
	tests := []struct {
		name       string
		fields     fields
		wantErrMsg string
		want       *bool
	}{
		{
			name: "NotEnabled",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  false,
				},
			},
			wantErrMsg: "",
			want:       nil,
		},
		{
			name: "UnableToGetCpuCurrentRequests",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  true,
				},
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("CurrentRequests", mock.Anything, mock.Anything, mock.Anything).
						Return(resource.Quantity{}, errors.New(""))
				}),
			},
			wantErrMsg: "unable to get cpu current requests",
			want:       func() *bool { b := false; return &b }(),
		},
		{
			name: "True",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  true,
				},
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("CurrentRequests", mock.Anything, mock.Anything, mock.Anything).
						Return(resource.Quantity{}, nil)
					m.On("Requests", mock.Anything, v1.ResourceCPU).Return(resource.Quantity{})
				}),
			},
			wantErrMsg: "",
			want:       func() *bool { b := true; return &b }(),
		},
		{
			name: "False",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  true,
				},
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("CurrentRequests", mock.Anything, mock.Anything, mock.Anything).
						Return(resource.Quantity{}, nil)
					m.On("Requests", mock.Anything, v1.ResourceCPU).Return(resource.MustParse("1m"))
				}),
			},
			wantErrMsg: "",
			want:       func() *bool { b := false; return &b }(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &state{
				resourceName:    v1.ResourceCPU,
				config:          tt.fields.config,
				containerHelper: tt.fields.containerHelper,
			}
			got, err := s.DoesRequestsCurrentMatchSpec(&v1.Pod{}, &v1.Container{})
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStateDoesLimitsCurrentMatchSpec(t *testing.T) {
	type fields struct {
		config          scalecommon.Configuration
		containerHelper kubecommon.ContainerHelper
	}
	tests := []struct {
		name       string
		fields     fields
		wantErrMsg string
		want       *bool
	}{
		{
			name: "NotEnabled",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  false,
				},
			},
			wantErrMsg: "",
			want:       nil,
		},
		{
			name: "UnableToGetCpuCurrentLimits",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  true,
				},
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("CurrentLimits", mock.Anything, mock.Anything, mock.Anything).
						Return(resource.Quantity{}, errors.New(""))
				}),
			},
			wantErrMsg: "unable to get cpu current limits",
			want:       func() *bool { b := false; return &b }(),
		},
		{
			name: "True",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  true,
				},
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("CurrentLimits", mock.Anything, mock.Anything, mock.Anything).
						Return(resource.Quantity{}, nil)
					m.On("Limits", mock.Anything, v1.ResourceCPU).Return(resource.Quantity{})
				}),
			},
			wantErrMsg: "",
			want:       func() *bool { b := true; return &b }(),
		},
		{
			name: "False",
			fields: fields{
				config: &configuration{
					csaEnabled:   true,
					hasStored:    true,
					hasValidated: true,
					userEnabled:  true,
				},
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("CurrentLimits", mock.Anything, mock.Anything, mock.Anything).
						Return(resource.Quantity{}, nil)
					m.On("Limits", mock.Anything, v1.ResourceCPU).Return(resource.MustParse("1m"))
				}),
			},
			wantErrMsg: "",
			want:       func() *bool { b := false; return &b }(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &state{
				resourceName:    v1.ResourceCPU,
				config:          tt.fields.config,
				containerHelper: tt.fields.containerHelper,
			}
			got, err := s.DoesLimitsCurrentMatchSpec(&v1.Pod{}, &v1.Container{})
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
