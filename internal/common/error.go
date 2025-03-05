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

package common

import "fmt"

// WrapErrorf returns an error with the supplied format that wraps err. The supplied format is appended with ': %w',
// with %w as err.
func WrapErrorf(err error, format string, a ...any) error {
	wrapFormat := fmt.Sprintf("%s: %%w", format)
	return fmt.Errorf(wrapFormat, append(a, err)...)
}
