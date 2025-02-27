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

package podcommon

// TODO(wt) tests fixed in this file
import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestNewStatusAnnotation(t *testing.T) {
	statAnn := NewStatusAnnotation(
		"status",
		States{},
		StatusAnnotationScale{},
		"lastUpdated",
	)
	assert.Equal(t, "status", statAnn.Status)
	assert.Equal(t, States{}, statAnn.States)
	assert.Equal(t, StatusAnnotationScale{}, statAnn.Scale)
	assert.Equal(t, "lastUpdated", statAnn.LastUpdated)
}

func TestStatusAnnotationJson(t *testing.T) {
	j := NewStatusAnnotation(
		"status",
		NewStates("1", "2", "3", "4", "5", "6", "7"),
		NewStatusAnnotationScale([]v1.ResourceName{v1.ResourceCPU}, "lastCommanded", "lastEnacted", "lastFailed"),
		"lastUpdated",
	).Json()
	assert.Equal(
		t,
		"{\"status\":\"status\","+
			"\"states\":{\"startupProbe\":\"1\",\"readinessProbe\":\"2\",\"container\":\"3\",\"started\":\"4\",\"ready\":\"5\",\"resources\":\"6\",\"statusResources\":\"7\"},"+
			"\"scale\":{\"enabledForResources\":[\"cpu\"],\"lastCommanded\":\"lastCommanded\",\"lastEnacted\":\"lastEnacted\",\"lastFailed\":\"lastFailed\"},"+
			"\"lastUpdated\":\"lastUpdated\"}",
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
			fields{Status: "status"},
			args{to: StatusAnnotation{Status: "status"}},
			true,
		},
		{
			"TrueLastUpdatedDifferent",
			fields{Status: "status"},
			args{to: StatusAnnotation{
				Status:      "status",
				LastUpdated: "lastUpdated",
			}},
			true,
		},
		{
			"False",
			fields{Status: "status1"},
			args{to: StatusAnnotation{Status: "status2"}},
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
		assert.Contains(t, err.Error(), "unable to unmarshal")
		assert.Equal(t, StatusAnnotation{}, got)
	})

	t.Run("Ok", func(t *testing.T) {
		got, err := StatusAnnotationFromString(
			"{\"status\":\"status\"," +
				"\"states\":{\"startupProbe\":\"1\",\"readinessProbe\":\"2\",\"container\":\"3\",\"started\":\"4\",\"ready\":\"5\",\"resources\":\"6\",\"statusResources\":\"7\"}," +
				"\"scale\":{\"enabledForResources\":[\"cpu\"],\"lastCommanded\":\"lastCommanded\",\"lastEnacted\":\"lastEnacted\",\"lastFailed\":\"lastFailed\"}," +
				"\"lastUpdated\":\"lastUpdated\"}",
		)
		assert.Nil(t, err)
		assert.Equal(
			t,
			NewStatusAnnotation(
				"status",
				NewStates("1", "2", "3", "4", "5", "6", "7"),
				NewStatusAnnotationScale([]v1.ResourceName{v1.ResourceCPU}, "lastCommanded", "lastEnacted", "lastFailed"),
				"lastUpdated",
			),
			got,
		)
	})
}

func TestNewStatusAnnotationScale(t *testing.T) {
	statAnn := NewStatusAnnotationScale([]v1.ResourceName{v1.ResourceCPU}, "lastCommanded", "lastEnacted", "lastFailed")
	assert.Equal(t, []v1.ResourceName{v1.ResourceCPU}, statAnn.EnabledForResources)
	assert.Equal(t, "lastCommanded", statAnn.LastCommanded)
	assert.Equal(t, "lastEnacted", statAnn.LastEnacted)
	assert.Equal(t, "lastFailed", statAnn.LastFailed)
}

func TestNewEmptyStatusAnnotationScale(t *testing.T) {
	statAnn := NewEmptyStatusAnnotationScale([]v1.ResourceName{v1.ResourceCPU})
	assert.Equal(t, []v1.ResourceName{v1.ResourceCPU}, statAnn.EnabledForResources)
	assert.Empty(t, statAnn.LastCommanded)
	assert.Empty(t, statAnn.LastEnacted)
	assert.Empty(t, statAnn.LastFailed)
}
