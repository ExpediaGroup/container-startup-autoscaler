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

package common

import "reflect"

// IsStructEmpty returns whether s is (deeply) empty. Supports pointers. Returns false if s is not a struct.
func IsStructEmpty(s any) bool {
	value := reflect.ValueOf(s)

	if value.Kind() == reflect.Ptr {
		value = reflect.Indirect(reflect.ValueOf(s))
		s = value.Interface()
	}

	if value.Kind() == reflect.Struct {
		// Deep compare against a new empty version of the struct.
		return reflect.DeepEqual(s, reflect.New(reflect.TypeOf(s)).Elem().Interface())
	}

	return false
}

// AreStructsEqual returns whether s1 and s2 are (deeply) equal. Supports pointers. Returns false if s1 or s2 are not
// structs.
func AreStructsEqual(s1 any, s2 any) bool {
	value1 := reflect.ValueOf(s1)

	if value1.Kind() == reflect.Ptr {
		value1 = reflect.Indirect(reflect.ValueOf(s1))
		s1 = value1.Interface()
	}

	value2 := reflect.ValueOf(s2)

	if value2.Kind() == reflect.Ptr {
		value2 = reflect.Indirect(reflect.ValueOf(s2))
		s2 = value2.Interface()
	}

	if value1.Kind() != reflect.Struct || value2.Kind() != reflect.Struct {
		return false
	}

	return reflect.DeepEqual(s1, s2)
}
