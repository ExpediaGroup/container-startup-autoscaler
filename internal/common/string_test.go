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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsStringEmpty(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"TrueEmpty", args{s: ""}, true},
		{"TrueWhitespace", args{s: "   "}, true},
		{"False", args{s: "test"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsStringEmpty(tt.args.s))
		})
	}
}

func TestCapitalizeFirstChar(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"OneWord", args{s: "test"}, "Test"},
		{"TwoWords", args{s: "test test"}, "Test test"},
		{"Empty", args{s: ""}, ""},
		{"EmptyWhitespace", args{s: " "}, " "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CapitalizeFirstChar(tt.args.s))
		})
	}
}
