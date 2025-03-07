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

package pod

import (
	"errors"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

func TestNewConfiguration(t *testing.T) {
	podHelper := kube.NewPodHelper(nil)
	containerHelper := kube.NewContainerHelper()
	configuration := newConfiguration(podHelper, containerHelper)
	assert.Equal(t, podHelper, configuration.podHelper)
	assert.Equal(t, containerHelper, configuration.containerHelper)
}

func TestConfigurationConfigure(t *testing.T) {
	t.Run("UnableToStoreConfigurationFromAnnotations", func(t *testing.T) {
		mockPodHelper := kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) {
			m.On("ExpectedAnnotationValueAs", mock.Anything, mock.Anything, mock.Anything).
				Return("", errors.New(""))
			m.HasAnnotationDefault()
		})

		configuration := newConfiguration(mockPodHelper, nil)
		configs, err := configuration.Configure(&v1.Pod{})
		assert.Contains(t, err.Error(), "unable to store configuration from annotations")
		assert.Nil(t, configs)
	})

	t.Run("Ok", func(t *testing.T) {
		mockPodHelper := kubetest.NewMockPodHelper(nil)
		mockContainerHelper := kubetest.NewMockContainerHelper(nil)

		configuration := newConfiguration(mockPodHelper, mockContainerHelper)
		configs, err := configuration.Configure(&v1.Pod{})
		assert.Nil(t, err)
		assert.NotNil(t, configs)
	})
}
