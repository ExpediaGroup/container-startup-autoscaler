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

package contexttest

import (
	"bytes"
)

const KeyUuid = "uuid"

type ctxConfig struct {
	logBuffer              *bytes.Buffer
	standardRetryAttempts  int
	standardRetryDelaySecs int
}

func NewCtxConfig() ctxConfig {
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
