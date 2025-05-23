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
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/controller/controllercommon"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

type mockController struct {
	mock.Mock
}

func newMockController(configFunc func(*mockController)) *mockController {
	m := &mockController{}
	configFunc(m)
	return m
}

func (m *mockController) Reconcile(_ context.Context, _ reconcile.Request) (reconcile.Result, error) {
	panic(errors.New("not supported"))
}

func (m *mockController) Watch(src source.TypedSource[reconcile.Request]) error {
	args := m.Called(src)
	return args.Error(0)
}

func (m *mockController) Start(_ context.Context) error {
	panic(errors.New("not supported"))
}

func (m *mockController) GetLogger() logr.Logger {
	panic(errors.New("not supported"))
}

// ---------------------------------------------------------------------------------------------------------------------

type mockRuntimeManager struct {
	mock.Mock
}

func newMockRuntimeManager(configFunc func(*mockRuntimeManager)) *mockRuntimeManager {
	m := &mockRuntimeManager{}
	configFunc(m)
	return m
}

func (m *mockRuntimeManager) GetHTTPClient() *http.Client {
	panic(errors.New("not supported"))
}

func (m *mockRuntimeManager) GetConfig() *rest.Config {
	panic(errors.New("not supported"))
}

func (m *mockRuntimeManager) GetCache() cache.Cache {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(cache.Cache)
}

func (m *mockRuntimeManager) GetScheme() *runtime.Scheme {
	panic(errors.New("not supported"))
}

func (m *mockRuntimeManager) GetClient() client.Client {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(client.Client)
}

func (m *mockRuntimeManager) GetFieldIndexer() client.FieldIndexer {
	panic(errors.New("not supported"))
}

func (m *mockRuntimeManager) GetEventRecorderFor(name string) record.EventRecorder {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(record.EventRecorder)
}

func (m *mockRuntimeManager) GetRESTMapper() meta.RESTMapper {
	panic(errors.New("not supported"))
}

func (m *mockRuntimeManager) GetAPIReader() client.Reader {
	panic(errors.New("not supported"))
}

func (m *mockRuntimeManager) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockRuntimeManager) Add(_ manager.Runnable) error {
	panic(errors.New("not supported"))
}

func (m *mockRuntimeManager) Elected() <-chan struct{} {
	panic(errors.New("not supported"))
}

func (m *mockRuntimeManager) AddMetricsServerExtraHandler(_ string, _ http.Handler) error {
	panic(errors.New("not supported"))
}

func (m *mockRuntimeManager) AddHealthzCheck(_ string, _ healthz.Checker) error {
	panic(errors.New("not supported"))
}

func (m *mockRuntimeManager) AddReadyzCheck(_ string, _ healthz.Checker) error {
	panic(errors.New("not supported"))
}
func (m *mockRuntimeManager) GetWebhookServer() webhook.Server {
	panic(errors.New("not supported"))
}

func (m *mockRuntimeManager) GetLogger() logr.Logger {
	panic(errors.New("not supported"))
}

func (m *mockRuntimeManager) GetControllerOptions() config.Controller {
	panic(errors.New("not supported"))
}

// ---------------------------------------------------------------------------------------------------------------------

func TestNewController(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		conf := controllercommon.ControllerConfig{KubeConfig: "test1"}
		runtimeManager := newMockRuntimeManager(func(*mockRuntimeManager) {})
		cont := NewController(conf, runtimeManager)
		expected := &Controller{
			controllerConfig: conf,
			runtimeManager:   runtimeManager,
		}
		assert.Equal(t, expected, cont)

		conf = controllercommon.ControllerConfig{KubeConfig: "test2"}
		cont = NewController(conf, runtimeManager)
		assert.Equal(t, "test1", cont.controllerConfig.KubeConfig)
	})
}

func TestControllerInitialize(t *testing.T) {
	tests := []struct {
		name                     string
		configManagerMockFunc    func(*mockRuntimeManager)
		configControllerMockFunc func(*mockController)
		wantErrMsg               string
	}{
		{
			"UnableToWatchPods",
			func(runtimeManager *mockRuntimeManager) {
				runtimeManager.On("GetClient").Return(nil)
				runtimeManager.On("GetEventRecorderFor", mock.Anything).Return(nil)
				runtimeManager.On("GetCache").Return(nil)
			},
			func(controller *mockController) {
				controller.On("Watch", mock.Anything, mock.Anything, mock.Anything).Return(errors.New(""))
			},
			"unable to watch pods",
		},
		{
			"Ok",
			func(runtimeManager *mockRuntimeManager) {
				runtimeManager.On("GetClient").Return(nil)
				runtimeManager.On("GetEventRecorderFor", mock.Anything).Return(nil)
				runtimeManager.On("GetCache").Return(nil)
				runtimeManager.On("Start", mock.Anything).Return(nil)
			},
			func(controller *mockController) {
				controller.On("Watch", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				controllerConfig: controllercommon.ControllerConfig{},
				runtimeManager:   newMockRuntimeManager(tt.configManagerMockFunc),
			}

			err := c.Initialize(newMockController(tt.configControllerMockFunc))
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
