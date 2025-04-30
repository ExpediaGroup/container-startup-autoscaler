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

package controller

import (
	"testing"
	"time"

	csaevent "github.com/ExpediaGroup/container-startup-autoscaler/internal/event"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/event/eventcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/reconciler"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/component-base/metrics/testutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestPredicateCreateFunc(t *testing.T) {
	namespace := "namespace"
	name := "name"
	podEventCh := csaevent.DefaultPodEventPublisher.Subscribe(
		namespace,
		name,
		[]eventcommon.PodEventType{eventcommon.PodEventTypeCreate},
	)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	evt := event.TypedCreateEvent[*v1.Pod]{Object: pod}
	assert.True(t, predicateCreateFunc(evt))
	select {
	case podEvent := <-podEventCh:
		assert.Equal(t, eventcommon.PodEventTypeCreate, podEvent.EventType)
		assert.Same(t, pod, podEvent.Pod)
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("event not generated")
	}
}

func TestPredicateDeleteFunc(t *testing.T) {
	namespace := "namespace"
	name := "name"
	podEventCh := csaevent.DefaultPodEventPublisher.Subscribe(
		namespace,
		name,
		[]eventcommon.PodEventType{eventcommon.PodEventTypeDelete},
	)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	evt := event.TypedDeleteEvent[*v1.Pod]{Object: pod}
	assert.False(t, predicateDeleteFunc(evt))
	select {
	case podEvent := <-podEventCh:
		assert.Equal(t, eventcommon.PodEventTypeDelete, podEvent.EventType)
		assert.Same(t, pod, podEvent.Pod)
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("event not generated")
	}
}

func TestPredicateUpdateFunc(t *testing.T) {
	t.Run("ResourceVersionSame", func(t *testing.T) {
		oldPod, newPod := &v1.Pod{}, &v1.Pod{}
		oldPod.ResourceVersion, newPod.ResourceVersion = "1", "1"
		evt := event.TypedUpdateEvent[*v1.Pod]{
			ObjectOld: oldPod,
			ObjectNew: newPod,
		}
		assert.False(t, predicateUpdateFunc(evt))
	})

	t.Run("Deletion", func(t *testing.T) {
		oldPod, newPod := &v1.Pod{}, &v1.Pod{}
		oldPod.ResourceVersion, newPod.ResourceVersion = "1", "2"
		now := metav1.Now()
		newPod.DeletionTimestamp = &now
		evt := event.TypedUpdateEvent[*v1.Pod]{
			ObjectOld: oldPod,
			ObjectNew: newPod,
		}
		assert.False(t, predicateUpdateFunc(evt))
	})

	t.Run("StatusMissingOldNew", func(t *testing.T) {
		oldPod, newPod := &v1.Pod{}, &v1.Pod{}
		oldPod.ResourceVersion, newPod.ResourceVersion = "1", "2"
		evt := event.TypedUpdateEvent[*v1.Pod]{
			ObjectOld: oldPod,
			ObjectNew: newPod,
		}
		assert.True(t, predicateUpdateFunc(evt))
	})

	t.Run("OnlyStatusChanged", func(t *testing.T) {
		oldPod, newPod := &v1.Pod{}, &v1.Pod{}
		oldPod.ResourceVersion, newPod.ResourceVersion = "1", "2"
		oldPod.Annotations = map[string]string{kubecommon.AnnotationStatus: "test1"}
		newPod.Annotations = map[string]string{kubecommon.AnnotationStatus: "test2"}
		evt := event.TypedUpdateEvent[*v1.Pod]{
			ObjectOld: oldPod,
			ObjectNew: newPod,
		}
		assert.False(t, predicateUpdateFunc(evt))
		metricVal, _ := testutil.GetCounterMetricValue(reconciler.SkippedOnlyStatusChange())
		assert.Equal(t, float64(1), metricVal)
	})

	t.Run("NonStatusChanged", func(t *testing.T) {
		oldPod, newPod := &v1.Pod{}, &v1.Pod{}
		oldPod.ResourceVersion, newPod.ResourceVersion = "1", "2"
		oldPod.Annotations = map[string]string{kubecommon.AnnotationStatus: "test1"}
		newPod.Annotations = map[string]string{kubecommon.AnnotationStatus: "test1"}
		oldPod.ObjectMeta.Name, oldPod.ObjectMeta.Name = "test1", "test2"
		evt := event.TypedUpdateEvent[*v1.Pod]{
			ObjectOld: oldPod,
			ObjectNew: newPod,
		}
		assert.True(t, predicateUpdateFunc(evt))
	})

	t.Run("Subscriber", func(t *testing.T) {
		namespace := "namespace"
		name := "name"
		podEventCh := csaevent.DefaultPodEventPublisher.Subscribe(
			namespace,
			name,
			[]eventcommon.PodEventType{eventcommon.PodEventTypeUpdate},
		)
		oldPod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            name,
				Namespace:       namespace,
				ResourceVersion: "1",
			},
		}
		newPod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            name,
				Namespace:       namespace,
				ResourceVersion: "1",
			},
		}
		evt := event.TypedUpdateEvent[*v1.Pod]{
			ObjectOld: oldPod,
			ObjectNew: newPod,
		}
		_ = predicateUpdateFunc(evt)
		select {
		case podEvent := <-podEventCh:
			assert.Equal(t, eventcommon.PodEventTypeUpdate, podEvent.EventType)
			assert.Same(t, newPod, podEvent.Pod)
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("event not generated")
		}
	})
}

func TestPredicateGenericFunc(t *testing.T) {
	assert.False(t, predicateGenericFunc(event.TypedGenericEvent[*v1.Pod]{}))
}
