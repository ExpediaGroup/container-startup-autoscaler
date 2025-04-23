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

package scale

import (
	"errors"
	"strings"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scaletest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewConfiguration(t *testing.T) {
	config := NewConfiguration(
		v1.ResourceCPU,
		"annotationStartupName",
		"annotationPostStartupRequestsName",
		"annotationPostStartupLimitsName",
		true,
		nil,
		nil,
	)
	expected := &configuration{
		resourceName:                      v1.ResourceCPU,
		annotationStartupName:             "annotationStartupName",
		annotationPostStartupRequestsName: "annotationPostStartupRequestsName",
		annotationPostStartupLimitsName:   "annotationPostStartupLimitsName",
		csaEnabled:                        true,
		podHelper:                         nil,
		containerHelper:                   nil,
	}
	assert.Equal(t, expected, config)
}

func TestConfigurationResourceName(t *testing.T) {
	resourceName := v1.ResourceCPU
	config := &configuration{resourceName: resourceName}
	assert.Equal(t, v1.ResourceCPU, config.ResourceName())
}

func TestConfigurationIsEnabled(t *testing.T) {
	type fields struct {
		csaEnabled   bool
		hasStored    bool
		hasValidated bool
	}
	tests := []struct {
		name         string
		fields       fields
		wantPanicMsg string
		want         bool
	}{
		{
			"PanicStoreFromAnnotations",
			fields{
				false,
				false,
				false,
			},
			"StoreFromAnnotations() hasn't been invoked first",
			false,
		},
		{
			"PanicValidate",
			fields{
				false,
				true,
				false,
			},
			"Validate() hasn't been invoked first",
			false,
		},
		{
			"True",
			fields{
				true,
				true,
				true,
			},
			"",
			true,
		},
		{
			"False",
			fields{
				false,
				true,
				true,
			},
			"",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configuration{
				csaEnabled:   tt.fields.csaEnabled,
				hasStored:    tt.fields.hasStored,
				hasValidated: tt.fields.hasValidated,
				userEnabled:  true,
			}
			if tt.wantPanicMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicMsg, func() { config.IsEnabled() })
			} else {
				assert.Equal(t, tt.want, config.IsEnabled())
			}
		})
	}
}

func TestConfigurationResources(t *testing.T) {
	type fields struct {
		csaEnabled   bool
		hasStored    bool
		hasValidated bool
		resources    scalecommon.Resources
	}
	tests := []struct {
		name         string
		fields       fields
		wantPanicMsg string
		want         scalecommon.Resources
	}{
		{
			"PanicStoreFromAnnotations",
			fields{
				false,
				false,
				false,
				scalecommon.Resources{},
			},
			"StoreFromAnnotations() hasn't been invoked first",
			scalecommon.Resources{},
		},
		{
			"PanicValidate",
			fields{
				false,
				true,
				false,
				scalecommon.Resources{},
			},
			"Validate() hasn't been invoked first",
			scalecommon.Resources{},
		},
		{
			"Ok",
			fields{
				false,
				true,
				true,
				scaletest.ResourcesCpuEnabled,
			},
			"",
			scaletest.ResourcesCpuEnabled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configuration{
				csaEnabled:   true,
				hasStored:    tt.fields.hasStored,
				hasValidated: tt.fields.hasValidated,
				userEnabled:  true,
				resources:    tt.fields.resources,
			}
			if tt.wantPanicMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicMsg, func() { config.Resources() })
			} else {
				assert.Equal(t, tt.want, config.Resources())
			}
		})
	}
}

func TestConfigurationStoreFromAnnotations(t *testing.T) {
	type fields struct {
		annotationStartupName             string
		annotationPostStartupRequestsName string
		annotationPostStartupLimitsName   string
		csaEnabled                        bool
		podHelper                         kubecommon.PodHelper
	}
	tests := []struct {
		name             string
		fields           fields
		wantErrMsg       string
		wantHasStored    bool
		wantRawResources scalecommon.RawResources
	}{
		{
			"NotCsaEnabled",
			fields{
				"",
				"",
				"",
				false,
				nil,
			},
			"",
			true,
			scalecommon.RawResources{},
		},
		{
			"UnableToGetStartupAnnotationValue",
			fields{
				scalecommon.AnnotationCpuStartup,
				"",
				"",
				true,
				kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) {
					m.On("ExpectedAnnotationValueAs", mock.Anything, mock.Anything, mock.Anything).
						Return("", errors.New(""))
					m.HasAnnotationDefault()
				}),
			},
			"unable to get '" + scalecommon.AnnotationCpuStartup + "' annotation value",
			false,
			scalecommon.RawResources{},
		},
		{
			"UnableToGetPostStartupRequestsAnnotationValue",
			fields{
				scalecommon.AnnotationCpuStartup,
				scalecommon.AnnotationCpuPostStartupRequests,
				"",
				true,
				kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) {
					m.On(
						"ExpectedAnnotationValueAs",
						mock.Anything,
						mock.MatchedBy(func(ann string) bool { return strings.Contains(ann, scalecommon.AnnotationCpuStartup) }),
						kubecommon.DataTypeString,
					).Return(kubetest.PodAnnotationCpuStartup, nil)
					m.On(
						"ExpectedAnnotationValueAs",
						mock.Anything,
						mock.MatchedBy(func(ann string) bool { return strings.Contains(ann, scalecommon.AnnotationCpuPostStartupRequests) }),
						kubecommon.DataTypeString,
					).Return("", errors.New(""))
					m.HasAnnotationDefault()
				}),
			},
			"unable to get '" + scalecommon.AnnotationCpuPostStartupRequests + "' annotation value",
			false,
			scalecommon.RawResources{},
		},
		{
			"UnableToGetPostStartupLimitsAnnotationValue",
			fields{
				scalecommon.AnnotationCpuStartup,
				scalecommon.AnnotationCpuPostStartupRequests,
				scalecommon.AnnotationCpuPostStartupLimits,
				true,
				kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) {
					m.On(
						"ExpectedAnnotationValueAs",
						mock.Anything,
						mock.MatchedBy(func(ann string) bool { return strings.Contains(ann, scalecommon.AnnotationCpuStartup) }),
						kubecommon.DataTypeString,
					).Return(kubetest.PodAnnotationCpuStartup, nil)
					m.On(
						"ExpectedAnnotationValueAs",
						mock.Anything,
						mock.MatchedBy(func(ann string) bool { return strings.Contains(ann, scalecommon.AnnotationCpuPostStartupRequests) }),
						kubecommon.DataTypeString,
					).Return(kubetest.PodAnnotationCpuPostStartupRequests, nil)
					m.On(
						"ExpectedAnnotationValueAs",
						mock.Anything,
						mock.MatchedBy(func(ann string) bool { return strings.Contains(ann, scalecommon.AnnotationCpuPostStartupLimits) }),
						kubecommon.DataTypeString,
					).Return("", errors.New(""))
					m.HasAnnotationDefault()
				}),
			},
			"unable to get '" + scalecommon.AnnotationCpuPostStartupLimits + "' annotation value",
			false,
			scalecommon.RawResources{},
		},
		{
			"Ok",
			fields{
				scalecommon.AnnotationCpuStartup,
				scalecommon.AnnotationCpuPostStartupRequests,
				scalecommon.AnnotationCpuPostStartupLimits,
				true,
				kubetest.NewMockPodHelper(nil),
			},
			"",
			true,
			scaletest.RawResourcesCpuEnabled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configuration{
				annotationStartupName:             tt.fields.annotationStartupName,
				annotationPostStartupRequestsName: tt.fields.annotationPostStartupRequestsName,
				annotationPostStartupLimitsName:   tt.fields.annotationPostStartupLimitsName,
				csaEnabled:                        tt.fields.csaEnabled,
				podHelper:                         tt.fields.podHelper,
			}
			err := config.StoreFromAnnotations(&v1.Pod{})
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantHasStored, config.hasStored)
			assert.Equal(t, tt.wantRawResources, config.rawResources)
		})
	}
}

func TestConfigurationValidate(t *testing.T) {
	type fields struct {
		csaEnabled      bool
		containerHelper kubecommon.ContainerHelper
		hasStored       bool
		rawResources    scalecommon.RawResources
	}
	tests := []struct {
		name             string
		fields           fields
		wantPanicMsg     string
		wantErrMsg       string
		wantUserEnabled  bool
		wantHasValidated bool
		wantResources    scalecommon.Resources
	}{
		{
			"PanicStoreFromAnnotations",
			fields{},
			"StoreFromAnnotations() hasn't been invoked first",
			"",
			false,
			false,
			scalecommon.Resources{},
		},
		{
			"NotCsaEnabled",
			fields{
				false,
				nil,
				true,
				scalecommon.RawResources{},
			},
			"",
			"",
			false,
			true,
			scalecommon.Resources{},
		},
		{
			"NotUserEnabled",
			fields{
				true,
				nil,
				true,
				scalecommon.RawResources{},
			},
			"",
			"",
			false,
			true,
			scalecommon.Resources{},
		},
		{
			"AnnotationStartupNotPresent",
			fields{
				true,
				nil,
				true,
				scalecommon.RawResources{PostStartupRequests: "1m"},
			},
			"",
			"annotation 'annotationStartupName' not present",
			false,
			false,
			scalecommon.Resources{},
		},
		{
			"AnnotationPostStartupRequestsNotPresent",
			fields{
				true,
				nil,
				true,
				scalecommon.RawResources{Startup: "1m"},
			},
			"",
			"annotation 'annotationPostStartupRequestsName' not present",
			false,
			false,
			scalecommon.Resources{},
		},
		{
			"AnnotationPostStartupLimitsNotPresent",
			fields{
				true,
				nil,
				true,
				scalecommon.RawResources{
					Startup:             "1m",
					PostStartupRequests: "1m",
				},
			},
			"",
			"annotation 'annotationPostStartupLimitsName' not present",
			false,
			false,
			scalecommon.Resources{},
		},
		{
			"UnableToParseStartupAnnotationValue",
			fields{
				true,
				nil,
				true,
				scalecommon.RawResources{
					Startup:             "invalid",
					PostStartupRequests: "1m",
					PostStartupLimits:   "1m",
				},
			},
			"",
			"unable to parse 'annotationStartupName' annotation value ('invalid')",
			false,
			false,
			scalecommon.Resources{},
		},
		{
			"UnableToParsePostStartupRequestsAnnotationValue",
			fields{
				true,
				nil,
				true,
				scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "invalid",
					PostStartupLimits:   "1m",
				},
			},
			"",
			"unable to parse 'annotationPostStartupRequestsName' annotation value ('invalid')",
			false,
			false,
			scalecommon.Resources{},
		},
		{
			"UnableToParsePostStartupLimitsAnnotationValue",
			fields{
				true,
				nil,
				true,
				scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "invalid",
				},
			},
			"",
			"unable to parse 'annotationPostStartupLimitsName' annotation value ('invalid')",
			false,
			false,
			scalecommon.Resources{},
		},
		{
			"PostStartupRequestsMustEqualPostStartupLimits",
			fields{
				true,
				nil,
				true,
				scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "2m",
				},
			},
			"",
			"cpu post-startup requests (1m) must equal post-startup limits (2m)",
			false,
			false,
			scalecommon.Resources{},
		},
		{
			"PostStartupRequestsGreaterThanStartupValue",
			fields{
				true,
				nil,
				true,
				scalecommon.RawResources{
					Startup:             "1m",
					PostStartupRequests: "2m",
					PostStartupLimits:   "2m",
				},
			},
			"",
			"cpu post-startup requests (2m) is greater than startup value (1m)",
			false,
			false,
			scalecommon.Resources{},
		},
		{
			"TargetContainerDoesNotSpecifyRequests",
			fields{
				true,
				kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("Requests", mock.Anything, mock.Anything).Return(resource.Quantity{})
				}),
				true,
				scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "1m",
				},
			},
			"",
			"target container does not specify cpu requests",
			false,
			false,
			scalecommon.Resources{},
		},
		{
			"TargetContainerDoesNotSpecifyLimits",
			fields{
				true,
				kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("Requests", mock.Anything, mock.Anything).Return(resource.MustParse("2m"))
					m.On("Limits", mock.Anything, mock.Anything).Return(resource.Quantity{})
				}),
				true,
				scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "1m",
				},
			},
			"",
			"target container does not specify cpu limits",
			false,
			false,
			scalecommon.Resources{},
		},
		{
			"TargetContainerRequestsMustEqualLimits",
			fields{
				true,
				kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("Requests", mock.Anything, mock.Anything).Return(resource.MustParse("1m"))
					m.On("Limits", mock.Anything, mock.Anything).Return(resource.MustParse("2m"))
				}),
				true,
				scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "1m",
				},
			},
			"",
			"target container cpu requests (1m) must equal limits (2m)",
			false,
			false,
			scalecommon.Resources{},
		},
		{
			"UnableToGetTargetContainerResizePolicy",
			fields{
				true,
				kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("ResizePolicy", mock.Anything, mock.Anything).
						Return(v1.ResourceResizeRestartPolicy(""), errors.New(""))
					m.RequestsDefault()
					m.LimitsDefault()
				}),
				true,
				scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "1m",
				},
			},
			"",
			"unable to get target container cpu resize policy",
			false,
			false,
			scalecommon.Resources{},
		},
		{
			"TargetContainerResizePolicyIsNotNotRequired",
			fields{
				true,
				kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("ResizePolicy", mock.Anything, mock.Anything).
						Return(v1.RestartContainer, nil)
					m.RequestsDefault()
					m.LimitsDefault()
				}),
				true,
				scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "1m",
				},
			},
			"",
			"target container cpu resize policy is not 'NotRequired' ('RestartContainer')",
			false,
			false,
			scalecommon.Resources{},
		},
		{
			"Ok",
			fields{
				true,
				kubetest.NewMockContainerHelper(nil),
				true,
				scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "1m",
				},
			},
			"",
			"",
			true,
			true,
			scalecommon.Resources{
				Startup:             resource.MustParse("2m"),
				PostStartupRequests: resource.MustParse("1m"),
				PostStartupLimits:   resource.MustParse("1m"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configuration{
				resourceName:                      v1.ResourceCPU,
				annotationStartupName:             "annotationStartupName",
				annotationPostStartupRequestsName: "annotationPostStartupRequestsName",
				annotationPostStartupLimitsName:   "annotationPostStartupLimitsName",
				csaEnabled:                        tt.fields.csaEnabled,
				containerHelper:                   tt.fields.containerHelper,
				hasStored:                         tt.fields.hasStored,
				rawResources:                      tt.fields.rawResources,
			}
			if tt.wantPanicMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicMsg, func() { _ = config.Validate(&v1.Container{}) })
			} else {
				err := config.Validate(&v1.Container{})
				if tt.wantErrMsg != "" {
					assert.ErrorContains(t, err, tt.wantErrMsg)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, tt.wantUserEnabled, config.userEnabled)
				assert.Equal(t, tt.wantHasValidated, config.hasValidated)
				assert.Equal(t, tt.wantResources, config.resources)
			}
		})
	}
}

func TestConfigurationString(t *testing.T) {
	type fields struct {
		csaEnabled   bool
		hasStored    bool
		hasValidated bool
		resources    scalecommon.Resources
	}
	tests := []struct {
		name         string
		fields       fields
		wantPanicMsg string
		want         string
	}{
		{
			"PanicStoreFromAnnotations",
			fields{
				false,
				false,
				false,
				scalecommon.Resources{},
			},
			"StoreFromAnnotations() hasn't been invoked first",
			"",
		},
		{
			"PanicValidate",
			fields{
				false,
				true,
				false,
				scalecommon.Resources{},
			},
			"Validate() hasn't been invoked first",
			"",
		},
		{
			"NotEnabled",
			fields{
				false,
				true,
				true,
				scalecommon.Resources{},
			},
			"",
			"(cpu) not enabled",
		},
		{
			"Enabled",
			fields{
				true,
				true,
				true,
				scaletest.ResourcesCpuEnabled,
			},
			"",
			"(cpu) startup: " + kubetest.PodAnnotationCpuStartup +
				", post-startup requests: " + kubetest.PodAnnotationCpuPostStartupRequests +
				", post-startup limits: " + kubetest.PodAnnotationCpuPostStartupLimits,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configuration{
				resourceName: v1.ResourceCPU,
				csaEnabled:   tt.fields.csaEnabled,
				userEnabled:  true,
				hasStored:    tt.fields.hasStored,
				hasValidated: tt.fields.hasValidated,
				resources:    tt.fields.resources,
			}
			if tt.wantPanicMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicMsg, func() { _ = config.String() })
			} else {
				assert.Equal(t, tt.want, config.String())
			}
		})
	}
}
