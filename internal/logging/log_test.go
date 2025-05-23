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

package logging

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	ccontext "github.com/ExpediaGroup/container-startup-autoscaler/internal/context"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/go-logr/logr"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type args struct {
	ctx    context.Context
	v      V
	err    error
	format string
	args   []any
}

type test struct {
	name            string
	args            args
	wantPanicErrMsg string
	wantLogRxConfig wantLogRxConfig
}

type wantLogRxConfig struct {
	wantLevelRx      string
	wantMsgRx        string
	wantTargetNameRx string
	wantTargetStates bool
}

const (
	testCtxKeyBuffer = "buffer"

	testV         = VTrace
	testAddCaller = true

	testTargetContainerName = "name"
	testErrorMsg            = "error"
	testFormat              = "format %s"
	testArg                 = "arg"

	wantInfoLevelRx  = "info"
	wantDebugLevelRx = "debug"
	wantTraceLevelRx = "trace"
	wantErrorLevelRx = "error"

	wantFormatArgsMsgRx = "format arg"

	fatalSuffixRx = " \\(fatal\\)"
)

func init() {
	exitOnFatal = false
	Init(DefaultW, testV, testAddCaller)
}

func TestInit(t *testing.T) {
	t.Run("AddCallerTrue", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		Init(buffer, VDebug, true)
		assert.Equal(t, int(zerolog.DebugLevel), int(zerolog.GlobalLevel()))
		assert.Equal(t, VDebug, CurrentV)
		Infof(nil, VDebug, "test")
		assert.Contains(t, buffer.String(), fmt.Sprintf("\"%s\":", zerolog.CallerFieldName))

		Init(DefaultW, testV, testAddCaller) // Reset.
	})

	t.Run("AddCallerFalse", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		Init(buffer, VDebug, false)
		assert.Equal(t, int(zerolog.DebugLevel), int(zerolog.GlobalLevel()))
		assert.Equal(t, VDebug, CurrentV)
		Infof(nil, VDebug, "test")
		assert.NotContains(t, buffer.String(), fmt.Sprintf("\"%s\":", zerolog.CallerFieldName))

		Init(DefaultW, testV, testAddCaller) // Reset.
	})
}

func TestErrorf(t *testing.T) {
	tests := []test{
		{
			"ValidateFormatPanic",
			args{
				nil,
				VInfo,
				errors.New(testErrorMsg),
				" ",
				nil,
			},
			"format is empty",
			wantLogRxConfig{},
		},
		{
			"NoPodInfo",
			args{
				testContextNoPodInfo(),
				VInfo,
				errors.New(testErrorMsg),
				testFormat,
				[]any{testArg},
			},
			"",
			wantLogRxConfig{
				wantErrorLevelRx,
				wantFormatArgsMsgRx,
				"",
				false,
			},
		},
		{
			"PodInfo",
			args{
				testContextPodInfo(),
				VInfo,
				errors.New(testErrorMsg),
				testFormat,
				[]any{testArg},
			},
			"",
			wantLogRxConfig{
				wantErrorLevelRx,
				wantFormatArgsMsgRx,
				testTargetContainerName,
				true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, tt, func() { Errorf(tt.args.ctx, tt.args.err, tt.args.format, tt.args.args...) })
		})
	}
}

func TestFatalf(t *testing.T) {
	tests := []test{
		{
			"ValidateFormatPanic",
			args{
				nil,
				VInfo,
				errors.New(testErrorMsg),
				" ",
				nil,
			},
			"format is empty",
			wantLogRxConfig{},
		},
		{
			"NoPodInfo",
			args{
				testContextNoPodInfo(),
				VInfo,
				errors.New(testErrorMsg),
				testFormat,
				[]any{testArg},
			},
			"",
			wantLogRxConfig{
				wantErrorLevelRx,
				wantFormatArgsMsgRx + fatalSuffixRx,
				"",
				false,
			},
		},
		{
			"PodInfo",
			args{
				testContextPodInfo(),
				VInfo,
				errors.New(testErrorMsg),
				testFormat,
				[]any{testArg},
			},
			"",
			wantLogRxConfig{
				wantErrorLevelRx,
				wantFormatArgsMsgRx + fatalSuffixRx,
				testTargetContainerName,
				true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, tt, func() { Fatalf(tt.args.ctx, tt.args.err, tt.args.format, tt.args.args...) })
		})
	}
}

func TestInfof(t *testing.T) {
	tests := []test{
		{
			"ValidateFormatPanic",
			args{
				nil,
				VInfo,
				nil,
				" ",
				nil,
			},
			"format is empty",
			wantLogRxConfig{},
		},
		{
			"VInfoNoPodInfo",
			args{
				testContextNoPodInfo(),
				VInfo,
				nil,
				testFormat,
				[]any{testArg},
			},
			"",
			wantLogRxConfig{
				wantInfoLevelRx,
				wantFormatArgsMsgRx,
				"",
				false,
			},
		},
		{
			"VInfoPodInfo",
			args{
				testContextPodInfo(),
				VInfo,
				nil,
				testFormat,
				[]any{testArg},
			},
			"",
			wantLogRxConfig{
				wantInfoLevelRx,
				wantFormatArgsMsgRx,
				testTargetContainerName,
				true,
			},
		},
		{
			"VDebugNoPodInfo",
			args{
				testContextNoPodInfo(),
				VDebug,
				nil,
				testFormat,
				[]any{testArg},
			},
			"",
			wantLogRxConfig{
				wantDebugLevelRx,
				wantFormatArgsMsgRx,
				"",
				false,
			},
		},
		{
			"VTraceNoPodInfo",
			args{
				testContextNoPodInfo(),
				VTrace,
				nil,
				testFormat,
				[]any{testArg},
			},
			"",
			wantLogRxConfig{
				wantTraceLevelRx,
				wantFormatArgsMsgRx,
				"",
				false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, tt, func() { Infof(tt.args.ctx, tt.args.v, tt.args.format, tt.args.args...) })
		})
	}
}

func TestInfofNotLoggedForLevel(t *testing.T) {
	buffer := &bytes.Buffer{}
	Init(buffer, testV, testAddCaller)
	Infof(testContextNoPodInfo(), VDebug, "test")
	assert.Empty(t, buffer.String())
}

func testContextPodInfo() context.Context {
	buffer := &bytes.Buffer{}
	Init(buffer, testV, testAddCaller)

	ctx := logr.NewContext(context.TODO(), Logger)
	ctx = context.WithValue(ctx, testCtxKeyBuffer, buffer)
	ctx = ccontext.WithTargetContainerName(ctx, testTargetContainerName)
	ctx = ccontext.WithTargetContainerStates(ctx, podcommon.NewStatesAllUnknown())
	return ctx
}

func testContextNoPodInfo() context.Context {
	buffer := &bytes.Buffer{}
	Init(buffer, testV, testAddCaller)

	ctx := logr.NewContext(context.TODO(), Logger)
	ctx = context.WithValue(ctx, testCtxKeyBuffer, buffer)
	return ctx
}

func runTest(
	t *testing.T,
	test test,
	logFunc func(),
) {
	if test.wantPanicErrMsg != "" {
		assert.PanicsWithError(t, test.wantPanicErrMsg, logFunc)
		return
	}

	logFunc()
	assertLog(t, test.args.ctx, test.wantLogRxConfig)
	Init(DefaultW, testV, testAddCaller) // Reset.
}

func assertLog(t *testing.T, ctx context.Context, config wantLogRxConfig) {
	log := ctx.Value(testCtxKeyBuffer).(*bytes.Buffer).String()

	assert.Regexp(t, "\"caller\":\".+?\"", log)
	assert.Regexp(t, "\"time\":", log)
	assert.Regexp(t, fmt.Sprintf("\"level\":\"%s\"", config.wantLevelRx), log)
	assert.Regexp(t, fmt.Sprintf("\"message\":\"%s\"", config.wantMsgRx), log)

	if config.wantTargetNameRx != "" {
		assert.Regexp(t, fmt.Sprintf("\"%s\":\"%s\"", keyTargetContainerName, config.wantTargetNameRx), log)
	}

	if config.wantTargetStates {
		assert.Regexp(t, fmt.Sprintf("\"%s\":\\{.+?\\}", keyTargetContainerStates), log)
	}
}
