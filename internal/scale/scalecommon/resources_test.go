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

package scalecommon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewRawResources(t *testing.T) {
	resources := NewRawResources("3m", "1m", "2m")
	expected := RawResources{
		Startup:             "3m",
		PostStartupRequests: "1m",
		PostStartupLimits:   "2m",
	}
	assert.Equal(t, expected, resources)
}

func TestNewResources(t *testing.T) {
	resources := NewResources(resource.MustParse("3m"), resource.MustParse("1m"), resource.MustParse("2m"))
	expected := Resources{
		Startup:             resource.MustParse("3m"),
		PostStartupRequests: resource.MustParse("1m"),
		PostStartupLimits:   resource.MustParse("2m"),
	}
	assert.Equal(t, expected, resources)
}
