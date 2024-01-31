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

package logging

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	ccontext "github.com/ExpediaGroup/container-startup-autoscaler/internal/context"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type V int

const (
	VInfo  V = 0
	VDebug V = 1
	VTrace V = 2
)

const (
	DefaultV         = VTrace
	DefaultAddCaller = true
)

var (
	DefaultW = os.Stdout
	Logger   logr.Logger
	CurrentV = DefaultV
)

var (
	zLogger     zerolog.Logger
	exitOnFatal = true
)

// init configures the logger with default settings.
func init() {
	configureLogger(os.Stdout, DefaultV, DefaultAddCaller)
}

// Init (re)configures the logger with the supplied settings.
func Init(w io.Writer, v V, addCaller bool) {
	configureLogger(w, v, addCaller)
}

// configureLogger actually configures the logger with the supplied settings.
func configureLogger(w io.Writer, v V, addCaller bool) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerologr.SetMaxV(int(v)) // See https://github.com/go-logr/zerologr#implementation-details.
	zerologr.VerbosityFieldName = ""

	zLogger = zerolog.New(w)
	zLoggerCtx := zLogger.With()
	if addCaller {
		zLoggerCtx = zLoggerCtx.Caller()
	}
	zLogger = zLoggerCtx.Timestamp().Logger()
	Logger = zerologr.New(&zLogger)
	CurrentV = v
}

// Errorf logs err with a formatted message.
func Errorf(ctx context.Context, err error, format string, args ...any) {
	validateFormat(format)
	configuredLogger(ctx).Error(err, buildMessage(format, args, false))
}

// Fatalf logs err with a formatted message and exits with a non-0 return code.
func Fatalf(ctx context.Context, err error, format string, args ...any) {
	validateFormat(format)
	configuredLogger(ctx).Error(err, buildMessage(format, args, true))
	if exitOnFatal {
		os.Exit(1)
	}
}

// Infof logs information with a formatted message, at the indicated v.
func Infof(ctx context.Context, v V, format string, args ...any) {
	validateFormat(format)
	configuredLogger(ctx).V(int(v)).Info(buildMessage(format, args, false))
}

// validateFormat ensures format is correct and panics if not.
func validateFormat(format string) {
	if strings.ReplaceAll(format, " ", "") == "" {
		panic(errors.New("format is empty"))
	}
}

// buildMessage returns a message after applying formatting and other configuration.
func buildMessage(format string, args []any, isFatal bool) string {
	msg := format

	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	if isFatal {
		msg = fmt.Sprintf("%s (fatal)", msg)
	}

	return msg
}

// buildMessage returns a logger that's been configured with additional values.
func configuredLogger(ctx context.Context) logr.Logger {
	logger := Logger

	if ctx != nil {
		// Get the previously configured log from context (set by controller-runtime) so we can add more stuff.
		logger = log.FromContext(ctx)

		targetName := ccontext.TargetContainerName(ctx)
		if targetName != "" {
			logger = logger.WithValues(KeyTargetContainerName, targetName)
		}

		states := ccontext.TargetContainerStates(ctx)
		if states != (podcommon.States{}) {
			logger = logger.WithValues(KeyTargetContainerStates, states)
		}
	}

	logger = logger.WithCallDepth(1)

	return logger
}
