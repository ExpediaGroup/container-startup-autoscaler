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

package controller

import (
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/reconciler"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/component-base/metrics/testutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestPredicateCreateFunc(t *testing.T) {
	assert.True(t, PredicateCreateFunc(event.CreateEvent{}))
}

func TestPredicateDeleteFunc(t *testing.T) {
	assert.False(t, PredicateDeleteFunc(event.DeleteEvent{}))
}

func TestPredicateUpdateFunc(t *testing.T) {
	t.Run("ResourceVersionSame", func(t *testing.T) {
		oldPod, newPod := &v1.Pod{}, &v1.Pod{}
		oldPod.ResourceVersion, newPod.ResourceVersion = "1", "1"
		evt := event.UpdateEvent{
			ObjectOld: oldPod,
			ObjectNew: newPod,
		}
		assert.False(t, PredicateUpdateFunc(evt))
	})

	t.Run("Deletion", func(t *testing.T) {
		oldPod, newPod := &v1.Pod{}, &v1.Pod{}
		oldPod.ResourceVersion, newPod.ResourceVersion = "1", "2"
		now := metav1.Now()
		newPod.DeletionTimestamp = &now
		evt := event.UpdateEvent{
			ObjectOld: oldPod,
			ObjectNew: newPod,
		}
		assert.False(t, PredicateUpdateFunc(evt))
	})

	t.Run("StatusMissingOldNew", func(t *testing.T) {
		oldPod, newPod := &v1.Pod{}, &v1.Pod{}
		oldPod.ResourceVersion, newPod.ResourceVersion = "1", "2"
		evt := event.UpdateEvent{
			ObjectOld: oldPod,
			ObjectNew: newPod,
		}
		assert.True(t, PredicateUpdateFunc(evt))
	})

	t.Run("OnlyStatusChanged", func(t *testing.T) {
		oldPod, newPod := &v1.Pod{}, &v1.Pod{}
		oldPod.ResourceVersion, newPod.ResourceVersion = "1", "2"
		oldPod.Annotations = map[string]string{podcommon.AnnotationStatus: "test1"}
		newPod.Annotations = map[string]string{podcommon.AnnotationStatus: "test2"}
		evt := event.UpdateEvent{
			ObjectOld: oldPod,
			ObjectNew: newPod,
		}
		assert.False(t, PredicateUpdateFunc(evt))
		metricVal, _ := testutil.GetCounterMetricValue(reconciler.SkippedOnlyStatusChange())
		assert.Equal(t, float64(1), metricVal)
	})

	t.Run("NonStatusChanged", func(t *testing.T) {
		oldPod, newPod := &v1.Pod{}, &v1.Pod{}
		oldPod.ResourceVersion, newPod.ResourceVersion = "1", "2"
		oldPod.Annotations = map[string]string{podcommon.AnnotationStatus: "test1"}
		newPod.Annotations = map[string]string{podcommon.AnnotationStatus: "test1"}
		oldPod.ObjectMeta.Name, oldPod.ObjectMeta.Name = "test1", "test2"
		evt := event.UpdateEvent{
			ObjectOld: oldPod,
			ObjectNew: newPod,
		}
		assert.True(t, PredicateUpdateFunc(evt))
	})
}

func TestPredicateGenericFunc(t *testing.T) {
	assert.False(t, PredicateGenericFunc(event.GenericEvent{}))
}
