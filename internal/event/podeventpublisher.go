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
	"context"
	"sync"
	"time"

	ccontext "github.com/ExpediaGroup/container-startup-autoscaler/internal/context"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/event/eventcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
)

const (
	subscriberChannelBufferSize       = 10
	subscriberChannelWriteTimeoutSecs = 3
)

var DefaultPodEventPublisher = newPodEventPublisher()

// podEventPublisherSubscriber represents a configured subscriber of pod events.
type podEventPublisherSubscriber struct {
	ch         chan eventcommon.PodEvent
	namespace  string
	name       string
	eventTypes []eventcommon.PodEventType
}

// podEventPublisher is the default implementation of controllercommon.PodEventPublisher.
type podEventPublisher struct {
	mu          sync.RWMutex
	subscribers []podEventPublisherSubscriber
}

func newPodEventPublisher() eventcommon.PodEventPublisher {
	return &podEventPublisher{}
}

// Subscribe registers a subscriber for pod events. Returns a channel that will receive events matching the supplied
// configuration.
func (p *podEventPublisher) Subscribe(
	namespace string,
	name string,
	eventTypes []eventcommon.PodEventType,
) <-chan eventcommon.PodEvent {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan eventcommon.PodEvent, subscriberChannelBufferSize)
	p.subscribers = append(p.subscribers, podEventPublisherSubscriber{
		ch:         ch,
		namespace:  namespace,
		name:       name,
		eventTypes: eventTypes,
	})
	return ch
}

// Unsubscribe removes a subscriber from the publisher using the channel returned by Subscribe. The channel is closed.
func (p *podEventPublisher) Unsubscribe(ch <-chan eventcommon.PodEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, sub := range p.subscribers {
		if sub.ch == ch {
			p.subscribers = append(p.subscribers[:i], p.subscribers[i+1:]...)
			close(sub.ch)
			break
		}
	}
}

// Publish sends a pod event to all subscribers that match the event's namespace, name, and event type.
func (p *podEventPublisher) Publish(ctx context.Context, event eventcommon.PodEvent) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	timeoutDuration := subscriberChannelWriteTimeoutSecs * time.Second
	if ctx != nil {
		timeoutOverride := ccontext.TimeoutOverride(ctx)
		if timeoutOverride != 0 {
			timeoutDuration = timeoutOverride
			logging.Infof(ctx, logging.VInfo, "default event send timeout overridden")
		}
	}

	for _, sub := range p.subscribers {
		if sub.namespace == event.Pod.Namespace && sub.name == event.Pod.Name {
			for _, eventType := range sub.eventTypes {
				if eventType == event.EventType {
					select {
					case sub.ch <- event:
					case <-time.After(timeoutDuration):
						logging.Errorf(nil, nil, "timed out sending event to subscriber")
					}
					break
				}
			}
		}
	}
}
