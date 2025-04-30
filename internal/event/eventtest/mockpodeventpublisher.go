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

package eventtest

import (
	"context"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/event/eventcommon"
	"github.com/stretchr/testify/mock"
)

type MockPodEventPublisher struct {
	mock.Mock
}

func NewMockPodEventPublisher(configFunc func(*MockPodEventPublisher)) *MockPodEventPublisher {
	m := &MockPodEventPublisher{}
	if configFunc != nil {
		configFunc(m)
	} else {
		m.AllDefaults()
	}

	return m
}

func (m *MockPodEventPublisher) Subscribe(
	namespace string,
	name string,
	eventTypes []eventcommon.PodEventType,
) <-chan eventcommon.PodEvent {
	args := m.Called(namespace, name, eventTypes)
	return args.Get(0).(chan eventcommon.PodEvent)
}

func (m *MockPodEventPublisher) Unsubscribe(ch <-chan eventcommon.PodEvent) {
	m.Called(ch)
}

func (m *MockPodEventPublisher) Publish(ctx context.Context, event eventcommon.PodEvent) {
	m.Called(ctx, event)
}

func (m *MockPodEventPublisher) SubscribeDefault() {
	m.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(make(chan eventcommon.PodEvent, 1))
}

func (m *MockPodEventPublisher) UnsubscribeDefault() {
	m.On("Unsubscribe", mock.Anything).Return()
}

func (m *MockPodEventPublisher) PublishDefault() {
	m.On("Publish", mock.Anything).Return()
}

func (m *MockPodEventPublisher) AllDefaults() {
	m.SubscribeDefault()
	m.UnsubscribeDefault()
	m.PublishDefault()
}
