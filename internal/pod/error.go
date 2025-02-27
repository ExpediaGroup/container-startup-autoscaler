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

package pod

import (
	"fmt"
)

// ValidationError is an error that indicates validation failed. It wraps another error.
type ValidationError struct {
	message string
	wrapped error
}

func NewValidationError(message string, toWrap error) error {
	if toWrap == nil {
		return ValidationError{message: message}
	}

	return ValidationError{
		message: message,
		wrapped: toWrap,
	}
}

func (e ValidationError) Error() string {
	if e.wrapped == nil {
		return "validation error: " + e.message
	}

	return fmt.Errorf("validation error: %s: %w", e.message, e.wrapped).Error()
}

func (e ValidationError) Unwrap() error {
	return e.wrapped
}
