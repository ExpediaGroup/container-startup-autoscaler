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

package podcommon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewCpuConfig(t *testing.T) {
	startup := resource.MustParse("1m")
	postStartupRequests := resource.MustParse("2m")
	postStartupLimits := resource.MustParse("3m")

	cConfig := NewCpuConfig(startup, postStartupRequests, postStartupLimits)
	assert.Equal(t, startup, cConfig.Startup)
	assert.Equal(t, postStartupRequests, cConfig.PostStartupRequests)
	assert.Equal(t, postStartupLimits, cConfig.PostStartupLimits)
}

func TestNewMemoryConfig(t *testing.T) {
	startup := resource.MustParse("1M")
	postStartupRequests := resource.MustParse("2M")
	postStartupLimits := resource.MustParse("3M")

	mConfig := NewMemoryConfig(startup, postStartupRequests, postStartupLimits)
	assert.Equal(t, startup, mConfig.Startup)
	assert.Equal(t, postStartupRequests, mConfig.PostStartupRequests)
	assert.Equal(t, postStartupLimits, mConfig.PostStartupLimits)
}
