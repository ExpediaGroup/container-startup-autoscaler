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

package podcommon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestNewStatusAnnotation(t *testing.T) {
	statAnn := NewStatusAnnotation(
		"status",
		StatusAnnotationScale{},
		"lastUpdated",
	)
	expected := StatusAnnotation{
		Status:      "status",
		Scale:       StatusAnnotationScale{},
		LastUpdated: "lastUpdated",
	}
	assert.Equal(t, expected, statAnn)
}

func TestNewEmptyStatusAnnotation(t *testing.T) {
	assert.Equal(t, StatusAnnotation{}, NewEmptyStatusAnnotation())
}

func TestStatusAnnotationJson(t *testing.T) {
	j := NewStatusAnnotation(
		"status",
		NewStatusAnnotationScale([]v1.ResourceName{v1.ResourceCPU}, "1", "2", "3"),
		"4",
	).Json()
	assert.Equal(
		t,
		`{"status":"status",`+
			`"scale":{"enabledForResources":["cpu"],"lastCommanded":"1","lastEnacted":"2","lastFailed":"3"},`+
			`"lastUpdated":"4"}`,
		j,
	)
}

func TestStatusAnnotationEqual(t *testing.T) {
	type fields struct {
		Status string
	}
	type args struct {
		to StatusAnnotation
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"TrueLastUpdatedSame",
			fields{"status"},
			args{StatusAnnotation{Status: "status"}},
			true,
		},
		{
			"TrueLastUpdatedDifferent",
			fields{"status"},
			args{StatusAnnotation{
				Status:      "status",
				LastUpdated: "lastUpdated",
			}},
			true,
		},
		{
			"False",
			fields{"status1"},
			args{StatusAnnotation{Status: "status2"}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StatusAnnotation{Status: tt.fields.Status}
			assert.Equal(t, tt.want, s.Equal(tt.args.to))
		})
	}
}

func TestStatusAnnotationFromString(t *testing.T) {
	t.Run("UnableToUnmarshal", func(t *testing.T) {
		got, err := StatusAnnotationFromString("test")
		assert.ErrorContains(t, err, "unable to unmarshal")
		assert.Equal(t, StatusAnnotation{}, got)
	})

	t.Run("Ok", func(t *testing.T) {
		got, err := StatusAnnotationFromString(
			`{"status":"status",` +
				`"scale":{"enabledForResources":["cpu"],"lastCommanded":"1","lastEnacted":"2","lastFailed":"3"},` +
				`"lastUpdated":"4"}`,
		)
		assert.NoError(t, err)
		assert.Equal(
			t,
			NewStatusAnnotation(
				"status",
				NewStatusAnnotationScale([]v1.ResourceName{v1.ResourceCPU}, "1", "2", "3"),
				"4",
			),
			got,
		)
	})
}

func TestNewStatusAnnotationScale(t *testing.T) {
	statAnn := NewStatusAnnotationScale(
		[]v1.ResourceName{v1.ResourceCPU},
		"lastCommanded",
		"lastEnacted",
		"lastFailed",
	)
	expected := StatusAnnotationScale{
		EnabledForResources: []v1.ResourceName{v1.ResourceCPU},
		LastCommanded:       "lastCommanded",
		LastEnacted:         "lastEnacted",
		LastFailed:          "lastFailed",
	}
	assert.Equal(t, expected, statAnn)
}

func TestNewEmptyStatusAnnotationScale(t *testing.T) {
	statAnn := NewEmptyStatusAnnotationScale([]v1.ResourceName{v1.ResourceCPU})
	expected := StatusAnnotationScale{
		EnabledForResources: []v1.ResourceName{v1.ResourceCPU},
		LastCommanded:       "",
		LastEnacted:         "",
		LastFailed:          "",
	}
	assert.Equal(t, expected, statAnn)
}

func TestFixedEnabledForResources(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		got := fixedEnabledForResources(nil)
		assert.NotNil(t, got)
	})

	t.Run("NotNil", func(t *testing.T) {
		resources := []v1.ResourceName{v1.ResourceCPU, v1.ResourceMemory}
		got := fixedEnabledForResources(resources)
		assert.Equal(t, resources, got)
	})
}
