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

package ztemptestsfromdeleted

//import (
//	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
//	"github.com/stretchr/testify/mock"
//	"k8s.io/api/core/v1"
//)
//
//// MockScaleConfig is a generic mock for podcommon.ScaleConfig.
//type MockScaleConfig struct {
//	mock.Mock
//}
//
//func NewMockScaleConfig(configFunc func(*MockScaleConfig)) *MockScaleConfig {
//	mockConfig := &MockScaleConfig{}
//	configFunc(mockConfig)
//	return mockConfig
//}
//
//func (m *MockScaleConfig) GetTargetContainerName() string {
//	args := m.Called()
//	return args.String(0)
//}
//
//func (m *MockScaleConfig) GetCpuConfig() podcommon.CpuConfig {
//	args := m.Called()
//	return args.Get(0).(podcommon.CpuConfig)
//}
//
//func (m *MockScaleConfig) GetMemoryConfig() podcommon.MemoryConfig {
//	args := m.Called()
//	return args.Get(0).(podcommon.MemoryConfig)
//}
//
//func (m *MockScaleConfig) StoreFromAnnotations(pod *v1.Pod) error {
//	args := m.Called(pod)
//	return args.Error(0)
//}
//
//func (m *MockScaleConfig) Validate() error {
//	args := m.Called()
//	return args.Error(0)
//}
//
//func (m *MockScaleConfig) String() string {
//	args := m.Called()
//	return args.String(0)
//}
//
//func (m *MockScaleConfig) GetTargetContainerNameDefault() {
//	m.On("GetTargetContainerName").Return(DefaultContainerName)
//}
//
//func (m *MockScaleConfig) GetCpuConfigDefault() {
//	m.On("GetCpuConfig").Return(podcommon.NewCpuConfig(
//		MockDefaultCpuQuantity,
//		MockDefaultCpuQuantity,
//		MockDefaultCpuQuantity,
//	))
//}
//
//func (m *MockScaleConfig) GetMemoryConfigDefault() {
//	m.On("GetMemoryConfig").Return(podcommon.NewMemoryConfig(
//		MockDefaultMemoryQuantity,
//		MockDefaultMemoryQuantity,
//		MockDefaultMemoryQuantity,
//	))
//}
//
//func (m *MockScaleConfig) StoreFromAnnotationsDefault() {
//	m.On("StoreFromAnnotations", mock.Anything, mock.Anything).Return(nil)
//}
//
//func (m *MockScaleConfig) ValidateDefault() {
//	m.On("Validate").Return(nil)
//}
//
//func (m *MockScaleConfig) StringDefault() {
//	m.On("String").Return("")
//}
