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

package podcommon

import (
	"encoding/json"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	v1 "k8s.io/api/core/v1"
)

// StatusAnnotation holds status information that's serialized to JSON for status reporting.
type StatusAnnotation struct {
	Status      string                `json:"status"`
	Scale       StatusAnnotationScale `json:"scale"`
	LastUpdated string                `json:"lastUpdated"`
}

func NewStatusAnnotation(
	status string,
	scale StatusAnnotationScale,
	lastUpdated string,
) StatusAnnotation {
	return StatusAnnotation{
		status,
		scale,
		lastUpdated,
	}
}

func NewEmptyStatusAnnotation() StatusAnnotation {
	return StatusAnnotation{}
}

// Json returns a JSON string.
func (s StatusAnnotation) Json() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

// Equal returns whether this is to equal to another.
func (s StatusAnnotation) Equal(to StatusAnnotation) bool {
	// Ignore s.LastUpdated.
	return s.Status == to.Status && common.AreStructsEqual(s.Scale, to.Scale)
}

// StatusAnnotationFromString returns a status annotation from s.
func StatusAnnotationFromString(s string) (StatusAnnotation, error) {
	ret := &StatusAnnotation{}
	if err := json.Unmarshal([]byte(s), ret); err != nil {
		return *ret, common.WrapErrorf(err, "unable to unmarshal")
	}

	return *ret, nil
}

// StatusAnnotationScale holds scale-related information that's serialized to JSON for status reporting.
type StatusAnnotationScale struct {
	EnabledForResources []v1.ResourceName `json:"enabledForResources"`
	LastCommanded       string            `json:"lastCommanded"`
	LastEnacted         string            `json:"lastEnacted"`
	LastFailed          string            `json:"lastFailed"`
}

func NewStatusAnnotationScale(
	enabledForResources []v1.ResourceName,
	lastCommanded string,
	lastEnacted string,
	lastFailed string,
) StatusAnnotationScale {
	return StatusAnnotationScale{
		fixedEnabledForResources(enabledForResources),
		lastCommanded,
		lastEnacted,
		lastFailed,
	}
}

func NewEmptyStatusAnnotationScale(enabledForResources []v1.ResourceName) StatusAnnotationScale {
	return StatusAnnotationScale{
		EnabledForResources: fixedEnabledForResources(enabledForResources),
	}
}

// fixedEnabledForResources explicitly returns an empty slice if enabledForResources is nil, otherwise the original
// slice. This ensures that the JSON output is always an array type, rather than null.
func fixedEnabledForResources(enabledForResources []v1.ResourceName) []v1.ResourceName {
	if enabledForResources == nil {
		return []v1.ResourceName{}
	}

	return enabledForResources
}
