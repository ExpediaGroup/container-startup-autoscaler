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

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestNewUpdates(t *testing.T) {
	updates := NewUpdates(NewConfigurations(nil, nil))
	allUpdates := updates.AllUpdates()
	assert.Equal(t, 2, len(allUpdates))
	assert.Equal(t, v1.ResourceCPU, allUpdates[0].ResourceName())
	assert.Equal(t, v1.ResourceMemory, allUpdates[1].ResourceName())
}

func TestStartupPodMutationFuncAll(t *testing.T) {
	updates := &updates{
		cpuUpdate: &update{
			config: &configuration{
				hasStored:    true,
				hasValidated: true,
			},
		},
		memoryUpdate: &update{
			config: &configuration{
				hasStored:    true,
				hasValidated: true,
			},
		},
	}
	allFuncs := updates.StartupPodMutationFuncAll(&v1.Container{})
	assert.Equal(t, 2, len(allFuncs))
}

func TestPostStartupPodMutationFuncAll(t *testing.T) {
	updates := &updates{
		cpuUpdate: &update{
			config: &configuration{
				hasStored:    true,
				hasValidated: true,
			},
		},
		memoryUpdate: &update{
			config: &configuration{
				hasStored:    true,
				hasValidated: true,
			},
		},
	}
	allFuncs := updates.PostStartupPodMutationFuncAll(&v1.Container{})
	assert.Equal(t, 2, len(allFuncs))
}

func TestUpdateFor(t *testing.T) {
	type fields struct {
		cpuUpdate    scalecommon.Update
		memoryUpdate scalecommon.Update
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
				&update{resourceName: v1.ResourceCPU},
				&update{resourceName: v1.ResourceMemory},
			},
			args{v1.ResourceCPU},
			false,
			v1.ResourceCPU,
		},
		{
			"Memory",
			fields{
				&update{resourceName: v1.ResourceCPU},
				&update{resourceName: v1.ResourceMemory},
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
			updates := &updates{
				cpuUpdate:    tt.fields.cpuUpdate,
				memoryUpdate: tt.fields.memoryUpdate,
			}
			got := updates.UpdateFor(tt.args.resourceName)
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, tt.wantResourceName, got.ResourceName())
			}
		})
	}
}

func TestAllUpdates(t *testing.T) {
	updates := &updates{
		cpuUpdate:    &update{resourceName: v1.ResourceCPU},
		memoryUpdate: &update{resourceName: v1.ResourceMemory},
	}
	allUpdates := updates.AllUpdates()
	assert.Equal(t, 2, len(allUpdates))
	assert.Equal(t, v1.ResourceCPU, allUpdates[0].ResourceName())
	assert.Equal(t, v1.ResourceMemory, allUpdates[1].ResourceName())
}
