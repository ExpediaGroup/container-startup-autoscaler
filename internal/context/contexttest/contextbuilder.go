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
)

// ctxBuilder builds a test context.
type ctxBuilder struct {
	config ctxConfig
}

func NewCtxBuilder(config ctxConfig) *ctxBuilder {
	return &ctxBuilder{config: config}
}

func (b *ctxBuilder) LogBuffer(logBuffer *bytes.Buffer) *ctxBuilder {
	b.config.logBuffer = logBuffer
	return b
}

func (b *ctxBuilder) StandardRetryAttempts(standardRetryAttempts int) *ctxBuilder {
	b.config.standardRetryAttempts = standardRetryAttempts
	return b
}

func (b *ctxBuilder) StandardRetryDelaySecs(standardRetryDelaySecs int) *ctxBuilder {
	b.config.standardRetryDelaySecs = standardRetryDelaySecs
	return b
}

func (b *ctxBuilder) Build() context.Context {
	// TODO(wt) move everything from context.go and remove context.go

	return ctx(b.config)
}
