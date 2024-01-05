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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct1 struct {
	structField testStruct2
}

type testStruct2 struct {
	stringField string
}

func TestIsStructEmpty(t *testing.T) {
	tests := []struct {
		name string
		s    any
		want bool
	}{
		{"NonPointerStructTrue", testStruct1{}, true},
		{"NonPointerStructFalse", testStruct1{structField: testStruct2{stringField: "test"}}, false},
		{"PointerStructTrue", &testStruct1{}, true},
		{"PointerStructFalse", &testStruct1{structField: testStruct2{stringField: "test"}}, false},
		{"NotStructFalse", 5, false},
		{"NilFalse", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsStructEmpty(tt.s))
		})
	}
}

func TestAreStructsEqual(t *testing.T) {
	tests := []struct {
		name string
		s1   any
		s2   any
		want bool
	}{
		{
			"NonPointerStructTrue",
			testStruct1{structField: testStruct2{stringField: "test"}},
			testStruct1{structField: testStruct2{stringField: "test"}},
			true,
		},
		{
			"NonPointerStructFalse",
			testStruct1{structField: testStruct2{stringField: "test1"}},
			testStruct1{structField: testStruct2{stringField: "test2"}},
			false,
		},
		{
			"PointerStructTrue",
			&testStruct1{structField: testStruct2{stringField: "test"}},
			&testStruct1{structField: testStruct2{stringField: "test"}},
			true,
		},
		{
			"PointerStructFalse",
			&testStruct1{structField: testStruct2{stringField: "test1"}},
			&testStruct1{structField: testStruct2{stringField: "test2"}},
			false,
		},
		{
			"NotStructFalse",
			5,
			5,
			false,
		},
		{
			"NilFalse",
			nil,
			nil,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, AreStructsEqual(tt.s1, tt.s2))
		})
	}
}
