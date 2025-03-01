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

package contexttest

import (
	"bytes"
	"context"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contextcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/logging"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/tonglil/buflogr"
)

const KeyUuid = "uuid"

// ctxConfig holds configuration for generating a test context.
type ctxConfig struct {
	logBuffer              *bytes.Buffer
	standardRetryAttempts  int
	standardRetryDelaySecs int
}

func NewCustomCtxConfig() ctxConfig {
	return ctxConfig{}
}

func NewNoRetryCtxConfig(logBuffer *bytes.Buffer) ctxConfig {
	return ctxConfig{
		logBuffer:              logBuffer,
		standardRetryAttempts:  1,
		standardRetryDelaySecs: 0,
	}
}

func NewOneRetryCtxConfig(logBuffer *bytes.Buffer) ctxConfig {
	return ctxConfig{
		logBuffer:              logBuffer,
		standardRetryAttempts:  2,
		standardRetryDelaySecs: 0,
	}
}

// ctx returns a test context from the supplied config.
func ctx(config ctxConfig) context.Context {
	var c context.Context

	if config.logBuffer == nil {
		logging.Init(logging.DefaultW, logging.DefaultV, logging.DefaultAddCaller)
		c = logr.NewContext(context.TODO(), logging.Logger)
	} else {
		c = logr.NewContext(context.TODO(), buflogr.NewWithBuffer(config.logBuffer))
	}

	c = context.WithValue(c, KeyUuid, uuid.New().String())
	c = context.WithValue(c, contextcommon.KeyStandardRetryAttempts, config.standardRetryAttempts)
	c = context.WithValue(c, contextcommon.KeyStandardRetryDelaySecs, config.standardRetryDelaySecs)
	return c
}
