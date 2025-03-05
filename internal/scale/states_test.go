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

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scaletest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

func TestNewStates(t *testing.T) {
	states := NewStates(NewConfigurations(nil, nil), nil)
	allStates := states.AllStates()
	assert.Equal(t, 2, len(allStates))
	assert.Equal(t, v1.ResourceCPU, allStates[0].ResourceName())
	assert.Equal(t, v1.ResourceMemory, allStates[1].ResourceName())
}

func TestStatesIsStartupConfigAppliedAll(t *testing.T) {
	type fields struct {
		cpuState    scalecommon.State
		memoryState scalecommon.State
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "True",
			fields: fields{
				cpuState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("IsStartupConfigurationApplied", mock.Anything).Return(func() *bool { b := true; return &b }())
				}),
				memoryState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("IsStartupConfigurationApplied", mock.Anything).Return(func() *bool { b := true; return &b }())
				}),
			},
			want: true,
		},
		{
			name: "False1",
			fields: fields{
				cpuState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("IsStartupConfigurationApplied", mock.Anything).Return(func() *bool { b := false; return &b }())
				}),
				memoryState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("IsStartupConfigurationApplied", mock.Anything).Return(func() *bool { b := true; return &b }())
				}),
			},
			want: false,
		},
		{
			name: "False2",
			fields: fields{
				cpuState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("IsStartupConfigurationApplied", mock.Anything).Return(func() *bool { b := true; return &b }())
				}),
				memoryState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("IsStartupConfigurationApplied", mock.Anything).Return(func() *bool { b := false; return &b }())
				}),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &states{
				cpuState:    tt.fields.cpuState,
				memoryState: tt.fields.memoryState,
			}
			assert.Equal(t, tt.want, s.IsStartupConfigurationAppliedAll(&v1.Container{}))
		})
	}
}

func TestStatesIsPostStartupConfigAppliedAll(t *testing.T) {
	type fields struct {
		cpuState    scalecommon.State
		memoryState scalecommon.State
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "True",
			fields: fields{
				cpuState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("IsPostStartupConfigurationApplied", mock.Anything).Return(func() *bool { b := true; return &b }())
				}),
				memoryState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("IsPostStartupConfigurationApplied", mock.Anything).Return(func() *bool { b := true; return &b }())
				}),
			},
			want: true,
		},
		{
			name: "False1",
			fields: fields{
				cpuState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("IsPostStartupConfigurationApplied", mock.Anything).Return(func() *bool { b := false; return &b }())
				}),
				memoryState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("IsPostStartupConfigurationApplied", mock.Anything).Return(func() *bool { b := true; return &b }())
				}),
			},
			want: false,
		},
		{
			name: "False2",
			fields: fields{
				cpuState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("IsPostStartupConfigurationApplied", mock.Anything).Return(func() *bool { b := true; return &b }())
				}),
				memoryState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("IsPostStartupConfigurationApplied", mock.Anything).Return(func() *bool { b := false; return &b }())
				}),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &states{
				cpuState:    tt.fields.cpuState,
				memoryState: tt.fields.memoryState,
			}
			assert.Equal(t, tt.want, s.IsPostStartupConfigurationAppliedAll(&v1.Container{}))
		})
	}
}

func TestStatesIsAnyCurrentZeroAll(t *testing.T) {
	type fields struct {
		cpuState    scalecommon.State
		memoryState scalecommon.State
	}
	tests := []struct {
		name       string
		fields     fields
		wantErrMsg string
		want       bool
	}{
		{
			name: "UnableToDetermineIfAnyCurrentIsZero",
			fields: fields{
				cpuState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("IsAnyCurrentZero", mock.Anything, mock.Anything).
						Return(func() *bool { b := false; return &b }(), errors.New(""))
					m.ResourceNameDefault()
				}),
				memoryState: scaletest.NewMockState(nil),
			},
			wantErrMsg: "unable to determine if any current cpu is zero",
			want:       false,
		},
		{
			name: "True",
			fields: fields{
				cpuState: scaletest.NewMockState(nil),
				memoryState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("IsAnyCurrentZero", mock.Anything, mock.Anything).
						Return(func() *bool { b := true; return &b }(), nil)
					m.ResourceNameDefault()
				}),
			},
			wantErrMsg: "",
			want:       true,
		},
		{
			name: "False",
			fields: fields{
				cpuState:    scaletest.NewMockState(nil),
				memoryState: scaletest.NewMockState(nil),
			},
			wantErrMsg: "",
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &states{
				cpuState:    tt.fields.cpuState,
				memoryState: tt.fields.memoryState,
			}
			got, err := s.IsAnyCurrentZeroAll(&v1.Pod{}, &v1.Container{})
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStatesDoesRequestsCurrentMatchSpecAll(t *testing.T) {
	type fields struct {
		cpuState    scalecommon.State
		memoryState scalecommon.State
	}
	tests := []struct {
		name       string
		fields     fields
		wantErrMsg string
		want       bool
	}{
		{
			name: "UnableToDetermineIfCurrentRequestsMatchesSpec",
			fields: fields{
				cpuState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("DoesRequestsCurrentMatchSpec", mock.Anything, mock.Anything).
						Return(func() *bool { b := false; return &b }(), errors.New(""))
					m.ResourceNameDefault()
				}),
				memoryState: scaletest.NewMockState(nil),
			},
			wantErrMsg: "unable to determine if current cpu requests matches spec",
			want:       false,
		},
		{
			name: "True",
			fields: fields{
				cpuState:    scaletest.NewMockState(nil),
				memoryState: scaletest.NewMockState(nil),
			},
			wantErrMsg: "",
			want:       true,
		},
		{
			name: "False1",
			fields: fields{
				cpuState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("DoesRequestsCurrentMatchSpec", mock.Anything, mock.Anything).
						Return(func() *bool { b := false; return &b }(), nil)
					m.ResourceNameDefault()
				}),
				memoryState: scaletest.NewMockState(nil),
			},
			wantErrMsg: "",
			want:       false,
		},
		{
			name: "False2",
			fields: fields{
				cpuState: scaletest.NewMockState(nil),
				memoryState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("DoesRequestsCurrentMatchSpec", mock.Anything, mock.Anything).
						Return(func() *bool { b := false; return &b }(), nil)
					m.ResourceNameDefault()
				}),
			},
			wantErrMsg: "",
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &states{
				cpuState:    tt.fields.cpuState,
				memoryState: tt.fields.memoryState,
			}
			got, err := s.DoesRequestsCurrentMatchSpecAll(&v1.Pod{}, &v1.Container{})
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStatesDoesLimitsCurrentMatchSpecAll(t *testing.T) {
	type fields struct {
		cpuState    scalecommon.State
		memoryState scalecommon.State
	}
	tests := []struct {
		name       string
		fields     fields
		wantErrMsg string
		want       bool
	}{
		{
			name: "UnableToDetermineIfCurrentLimitsMatchesSpec",
			fields: fields{
				cpuState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("DoesLimitsCurrentMatchSpec", mock.Anything, mock.Anything).
						Return(func() *bool { b := false; return &b }(), errors.New(""))
					m.ResourceNameDefault()
				}),
				memoryState: scaletest.NewMockState(nil),
			},
			wantErrMsg: "unable to determine if current cpu limits matches spec",
			want:       false,
		},
		{
			name: "True",
			fields: fields{
				cpuState:    scaletest.NewMockState(nil),
				memoryState: scaletest.NewMockState(nil),
			},
			wantErrMsg: "",
			want:       true,
		},
		{
			name: "False1",
			fields: fields{
				cpuState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("DoesLimitsCurrentMatchSpec", mock.Anything, mock.Anything).
						Return(func() *bool { b := false; return &b }(), nil)
					m.ResourceNameDefault()
				}),
				memoryState: scaletest.NewMockState(nil),
			},
			wantErrMsg: "",
			want:       false,
		},
		{
			name: "False2",
			fields: fields{
				cpuState: scaletest.NewMockState(nil),
				memoryState: scaletest.NewMockState(func(m *scaletest.MockState) {
					m.On("DoesLimitsCurrentMatchSpec", mock.Anything, mock.Anything).
						Return(func() *bool { b := false; return &b }(), nil)
					m.ResourceNameDefault()
				}),
			},
			wantErrMsg: "",
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &states{
				cpuState:    tt.fields.cpuState,
				memoryState: tt.fields.memoryState,
			}
			got, err := s.DoesLimitsCurrentMatchSpecAll(&v1.Pod{}, &v1.Container{})
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStatesStateFor(t *testing.T) {
	type fields struct {
		cpuState    scalecommon.State
		memoryState scalecommon.State
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
				cpuState:    &state{resourceName: v1.ResourceCPU},
				memoryState: &state{resourceName: v1.ResourceMemory},
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
				cpuState:    &state{resourceName: v1.ResourceCPU},
				memoryState: &state{resourceName: v1.ResourceMemory},
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
			states := &states{
				cpuState:    tt.fields.cpuState,
				memoryState: tt.fields.memoryState,
			}
			got := states.StateFor(tt.args.resourceName)
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, tt.wantResourceName, got.ResourceName())
			}
		})
	}
}

func TestStatesAllStates(t *testing.T) {
	states := &states{
		cpuState:    &state{resourceName: v1.ResourceCPU},
		memoryState: &state{resourceName: v1.ResourceMemory},
	}
	allStates := states.AllStates()
	assert.Equal(t, 2, len(allStates))
	assert.Equal(t, v1.ResourceCPU, allStates[0].ResourceName())
	assert.Equal(t, v1.ResourceMemory, allStates[1].ResourceName())
}
