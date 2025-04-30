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

package event

import (
	"testing"
	"time"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/event/eventcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/stretchr/testify/assert"
)

func TestNewPodEventPublisher(t *testing.T) {
	assert.Equal(t, &podEventPublisher{}, newPodEventPublisher())
}

func TestPodEventPublisherSubscribe(t *testing.T) {
	publisher := newPodEventPublisher()
	ch := publisher.Subscribe(
		kubetest.DefaultPodNamespace,
		kubetest.DefaultPodName,
		[]eventcommon.PodEventType{eventcommon.PodEventTypeCreate},
	)
	pod := kubetest.NewPodBuilder().Build()

	publisher.Publish(nil, eventcommon.NewPodEvent(eventcommon.PodEventTypeCreate, pod))
	select {
	case <-ch:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("event not generated")
	}
}

func TestPodEventPublisherUnsubscribe(t *testing.T) {
	publisher := newPodEventPublisher()
	ch := publisher.Subscribe(
		kubetest.DefaultPodNamespace,
		kubetest.DefaultPodName,
		[]eventcommon.PodEventType{eventcommon.PodEventTypeCreate},
	)
	pod := kubetest.NewPodBuilder().Build()

	publisher.Publish(nil, eventcommon.NewPodEvent(eventcommon.PodEventTypeCreate, pod))
	select {
	case <-ch:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("event not generated")
	}

	publisher.Unsubscribe(ch)

	publisher.Publish(nil, eventcommon.NewPodEvent(eventcommon.PodEventTypeCreate, pod))
	select {
	case event := <-ch:
		assert.Empty(t, event) // Channel should be closed so expect empty.
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("empty not received on channel read")
	}
}

func TestPodEventPublisherPublish(t *testing.T) {
	t.Run("EventMatchesAllSubscribers", func(t *testing.T) {
		publisher := newPodEventPublisher()
		ch1 := publisher.Subscribe(
			kubetest.DefaultPodNamespace,
			kubetest.DefaultPodName,
			[]eventcommon.PodEventType{eventcommon.PodEventTypeCreate},
		)
		ch2 := publisher.Subscribe(
			kubetest.DefaultPodNamespace,
			kubetest.DefaultPodName,
			[]eventcommon.PodEventType{eventcommon.PodEventTypeCreate},
		)
		pod := kubetest.NewPodBuilder().Build()

		publisher.Publish(nil, eventcommon.NewPodEvent(eventcommon.PodEventTypeCreate, pod))
		gotCh1, gotCh2 := false, false
		select {
		case podEvent := <-ch1:
			assert.Equal(t, eventcommon.PodEventTypeCreate, podEvent.EventType)
			assert.Same(t, pod, podEvent.Pod)
			gotCh1 = true
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("subscriber 1 event not generated")
		}
		select {
		case podEvent := <-ch2:
			assert.Equal(t, eventcommon.PodEventTypeCreate, podEvent.EventType)
			assert.Same(t, pod, podEvent.Pod)
			gotCh2 = true
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("subscriber 2 event not generated")
		}
		assert.True(t, gotCh1)
		assert.True(t, gotCh2)
	})

	t.Run("EventMatchesSomeSubscribers", func(t *testing.T) {
		publisher := newPodEventPublisher()
		ch1 := publisher.Subscribe(
			kubetest.DefaultPodNamespace,
			kubetest.DefaultPodName,
			[]eventcommon.PodEventType{eventcommon.PodEventTypeUpdate},
		)
		ch2 := publisher.Subscribe(
			kubetest.DefaultPodNamespace,
			kubetest.DefaultPodName,
			[]eventcommon.PodEventType{eventcommon.PodEventTypeCreate},
		)
		ch3 := publisher.Subscribe(
			kubetest.DefaultPodNamespace,
			"",
			[]eventcommon.PodEventType{eventcommon.PodEventTypeCreate},
		)
		ch4 := publisher.Subscribe(
			"",
			kubetest.DefaultPodName,
			[]eventcommon.PodEventType{eventcommon.PodEventTypeCreate},
		)
		pod := kubetest.NewPodBuilder().Build()

		publisher.Publish(nil, eventcommon.NewPodEvent(eventcommon.PodEventTypeCreate, pod))
		gotCh1, gotCh2, gotCh3, gotCh4 := false, false, false, false

		select {
		case <-ch1:
			gotCh1 = true
		case <-time.After(500 * time.Millisecond):
		}

		select {
		case podEvent := <-ch2:
			assert.Equal(t, eventcommon.PodEventTypeCreate, podEvent.EventType)
			assert.Same(t, pod, podEvent.Pod)
			gotCh2 = true
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("subscriber 2 event not generated")
		}

		select {
		case <-ch3:
			gotCh3 = true
		case <-time.After(500 * time.Millisecond):
		}

		select {
		case <-ch4:
			gotCh4 = true
		case <-time.After(500 * time.Millisecond):
		}

		assert.False(t, gotCh1)
		assert.True(t, gotCh2)
		assert.False(t, gotCh3)
		assert.False(t, gotCh4)
	})

	t.Run("EventMatchesNoSubscribers", func(t *testing.T) {
		publisher := newPodEventPublisher()
		ch1 := publisher.Subscribe(
			kubetest.DefaultPodNamespace,
			kubetest.DefaultPodName,
			[]eventcommon.PodEventType{eventcommon.PodEventTypeUpdate},
		)
		ch2 := publisher.Subscribe(
			kubetest.DefaultPodNamespace,
			kubetest.DefaultPodName,
			[]eventcommon.PodEventType{eventcommon.PodEventTypeUpdate},
		)
		pod := kubetest.NewPodBuilder().Build()

		publisher.Publish(nil, eventcommon.NewPodEvent(eventcommon.PodEventTypeCreate, pod))
		gotCh1, gotCh2 := false, false
		select {
		case <-ch1:
			gotCh1 = true
		case <-time.After(500 * time.Millisecond):
		}
		select {
		case <-ch2:
			gotCh2 = true
		case <-time.After(500 * time.Millisecond):
		}
		assert.False(t, gotCh1)
		assert.False(t, gotCh2)
	})
}
