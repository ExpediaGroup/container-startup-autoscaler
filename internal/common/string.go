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

import (
	"strings"
	"unicode"
)

// IsStringEmpty returns whether s is empty. All-spaces are reported empty.
func IsStringEmpty(s string) bool {
	return strings.ReplaceAll(s, " ", "") == ""
}

// CapitalizeFirstChar returns s with the first character capitalized.
func CapitalizeFirstChar(s string) string {
	if IsStringEmpty(s) {
		return s
	}

	out := []rune(s)
	out[0] = unicode.ToUpper(out[0])
	return string(out)
}
