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
	resourceName := v1.ResourceCPU
	config := NewConfiguration(
		resourceName,
		"",
		"",
		"",
		true,
		nil,
		nil,
	)
	assert.Equal(t, resourceName, config.ResourceName())
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
			name: "PanicStoreFromAnnotations",
			fields: fields{
				hasStored: false,
			},
			wantPanicMsg: "StoreFromAnnotations() hasn't been invoked first",
		},
		{
			name: "PanicValidate",
			fields: fields{
				hasStored:    true,
				hasValidated: false,
			},
			wantPanicMsg: "Validate() hasn't been invoked first",
		},
		{
			name: "True",
			fields: fields{
				csaEnabled:   true,
				hasStored:    true,
				hasValidated: true,
			},
			want: true,
		},
		{
			name: "False",
			fields: fields{
				csaEnabled:   false,
				hasStored:    true,
				hasValidated: true,
			},
			want: false,
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
			name: "PanicStoreFromAnnotations",
			fields: fields{
				hasStored: false,
			},
			wantPanicMsg: "StoreFromAnnotations() hasn't been invoked first",
		},
		{
			name: "PanicValidate",
			fields: fields{
				hasStored:    true,
				hasValidated: false,
			},
			wantPanicMsg: "Validate() hasn't been invoked first",
		},
		{
			name: "Ok",
			fields: fields{
				hasStored:    true,
				hasValidated: true,
				resources:    scaletest.ResourcesCpuEnabled,
			},
			want: scaletest.ResourcesCpuEnabled,
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
			name:             "NotCsaEnabled",
			fields:           fields{csaEnabled: false},
			wantErrMsg:       "",
			wantHasStored:    true,
			wantRawResources: scalecommon.RawResources{},
		},
		{
			name: "UnableToGetStartupAnnotationValue",
			fields: fields{
				annotationStartupName: scalecommon.AnnotationCpuStartup,
				csaEnabled:            true,
				podHelper: kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) {
					m.On("ExpectedAnnotationValueAs", mock.Anything, mock.Anything, mock.Anything).
						Return("", errors.New(""))
					m.HasAnnotationDefault()
				}),
			},
			wantErrMsg:       "unable to get '" + scalecommon.AnnotationCpuStartup + "' annotation value",
			wantHasStored:    false,
			wantRawResources: scalecommon.RawResources{},
		},
		{
			name: "UnableToGetPostStartupRequestsAnnotationValue",
			fields: fields{
				annotationStartupName:             scalecommon.AnnotationCpuStartup,
				annotationPostStartupRequestsName: scalecommon.AnnotationCpuPostStartupRequests,
				csaEnabled:                        true,
				podHelper: kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) {
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
			wantErrMsg:       "unable to get '" + scalecommon.AnnotationCpuPostStartupRequests + "' annotation value",
			wantHasStored:    false,
			wantRawResources: scalecommon.RawResources{},
		},
		{
			name: "UnableToGetPostStartupLimitsAnnotationValue",
			fields: fields{
				annotationStartupName:             scalecommon.AnnotationCpuStartup,
				annotationPostStartupRequestsName: scalecommon.AnnotationCpuPostStartupRequests,
				annotationPostStartupLimitsName:   scalecommon.AnnotationCpuPostStartupLimits,
				csaEnabled:                        true,
				podHelper: kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) {
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
			wantErrMsg:       "unable to get '" + scalecommon.AnnotationCpuPostStartupLimits + "' annotation value",
			wantHasStored:    false,
			wantRawResources: scalecommon.RawResources{},
		},
		{
			name: "Ok",
			fields: fields{
				annotationStartupName:             scalecommon.AnnotationCpuStartup,
				annotationPostStartupRequestsName: scalecommon.AnnotationCpuPostStartupRequests,
				annotationPostStartupLimitsName:   scalecommon.AnnotationCpuPostStartupLimits,
				csaEnabled:                        true,
				podHelper:                         kubetest.NewMockPodHelper(nil),
			},
			wantErrMsg:       "",
			wantHasStored:    true,
			wantRawResources: scaletest.RawResourcesCpuEnabled,
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
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.wantHasStored, config.hasStored)
			assert.Equal(t, tt.wantRawResources, config.rawResources)
		})
	}
}

// TODO(wt) standardize on not setting test fields unless they deviate from zero value
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
			name:         "PanicStoreFromAnnotations",
			wantPanicMsg: "StoreFromAnnotations() hasn't been invoked first",
		},
		{
			name: "NotCsaEnabled",
			fields: fields{
				hasStored: true,
			},
			wantErrMsg:       "",
			wantHasValidated: true,
		},
		{
			name: "NotUserEnabled",
			fields: fields{
				csaEnabled:   true,
				hasStored:    true,
				rawResources: scalecommon.RawResources{},
			},
			wantErrMsg:       "",
			wantHasValidated: true,
		},
		{
			name: "AnnotationStartupNotPresent",
			fields: fields{
				csaEnabled:   true,
				hasStored:    true,
				rawResources: scalecommon.RawResources{PostStartupRequests: "1m"},
			},
			wantErrMsg: "annotation 'annotationStartupName' not present",
		},
		{
			name: "AnnotationPostStartupRequestsNotPresent",
			fields: fields{
				csaEnabled:   true,
				hasStored:    true,
				rawResources: scalecommon.RawResources{Startup: "1m"},
			},
			wantErrMsg: "annotation 'annotationPostStartupRequestsName' not present",
		},
		{
			name: "AnnotationPostStartupLimitsNotPresent",
			fields: fields{
				csaEnabled: true,
				hasStored:  true,
				rawResources: scalecommon.RawResources{
					Startup:             "1m",
					PostStartupRequests: "1m",
				},
			},
			wantErrMsg: "annotation 'annotationPostStartupLimitsName' not present",
		},
		{
			name: "UnableToParseStartupAnnotationValue",
			fields: fields{
				csaEnabled: true,
				hasStored:  true,
				rawResources: scalecommon.RawResources{
					Startup:             "invalid",
					PostStartupRequests: "1m",
					PostStartupLimits:   "1m",
				},
			},
			wantErrMsg: "unable to parse 'annotationStartupName' annotation value ('invalid')",
		},
		{
			name: "UnableToParsePostStartupRequestsAnnotationValue",
			fields: fields{
				csaEnabled: true,
				hasStored:  true,
				rawResources: scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "invalid",
					PostStartupLimits:   "1m",
				},
			},
			wantErrMsg: "unable to parse 'annotationPostStartupRequestsName' annotation value ('invalid')",
		},
		{
			name: "UnableToParsePostStartupLimitsAnnotationValue",
			fields: fields{
				csaEnabled: true,
				hasStored:  true,
				rawResources: scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "invalid",
				},
			},
			wantErrMsg: "unable to parse 'annotationPostStartupLimitsName' annotation value ('invalid')",
		},
		{
			name: "PostStartupRequestsMustEqualPostStartupLimits",
			fields: fields{
				csaEnabled: true,
				hasStored:  true,
				rawResources: scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "2m",
				},
			},
			wantErrMsg: "cpu post-startup requests (1m) must equal post-startup limits (2m)",
		},
		{
			name: "PostStartupRequestsGreaterThanStartupValue",
			fields: fields{
				csaEnabled: true,
				hasStored:  true,
				rawResources: scalecommon.RawResources{
					Startup:             "1m",
					PostStartupRequests: "2m",
					PostStartupLimits:   "2m",
				},
			},
			wantErrMsg: "cpu post-startup requests (2m) is greater than startup value (1m)",
		},
		{
			name: "TargetContainerDoesNotSpecifyRequests",
			fields: fields{
				csaEnabled: true,
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("Requests", mock.Anything, mock.Anything).Return(resource.Quantity{})
				}),
				hasStored: true,
				rawResources: scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "1m",
				},
			},
			wantErrMsg: "target container does not specify cpu requests",
		},
		{
			name: "TargetContainerDoesNotSpecifyLimits",
			fields: fields{
				csaEnabled: true,
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("Requests", mock.Anything, mock.Anything).Return(resource.MustParse("2m"))
					m.On("Limits", mock.Anything, mock.Anything).Return(resource.Quantity{})
				}),
				hasStored: true,
				rawResources: scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "1m",
				},
			},
			wantErrMsg: "target container does not specify cpu limits",
		},
		{
			name: "TargetContainerRequestsMustEqualLimits",
			fields: fields{
				csaEnabled: true,
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("Requests", mock.Anything, mock.Anything).Return(resource.MustParse("1m"))
					m.On("Limits", mock.Anything, mock.Anything).Return(resource.MustParse("2m"))
				}),
				hasStored: true,
				rawResources: scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "1m",
				},
			},
			wantErrMsg: "target container cpu requests (1m) must equal limits (2m)",
		},
		{
			name: "UnableToGetTargetContainerResizePolicy",
			fields: fields{
				csaEnabled: true,
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("ResizePolicy", mock.Anything, mock.Anything).
						Return(v1.ResourceResizeRestartPolicy(""), errors.New(""))
					m.RequestsDefault()
					m.LimitsDefault()
				}),
				hasStored: true,
				rawResources: scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "1m",
				},
			},
			wantErrMsg: "unable to get target container cpu resize policy",
		},
		{
			name: "TargetContainerResizePolicyIsNotNotRequired",
			fields: fields{
				csaEnabled: true,
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("ResizePolicy", mock.Anything, mock.Anything).
						Return(v1.RestartContainer, nil)
					m.RequestsDefault()
					m.LimitsDefault()
				}),
				hasStored: true,
				rawResources: scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "1m",
				},
			},
			wantErrMsg: "target container cpu resize policy is not 'NotRequired' ('RestartContainer')",
		},
		{
			name: "Ok",
			fields: fields{
				csaEnabled:      true,
				containerHelper: kubetest.NewMockContainerHelper(nil),
				hasStored:       true,
				rawResources: scalecommon.RawResources{
					Startup:             "2m",
					PostStartupRequests: "1m",
					PostStartupLimits:   "1m",
				},
			},
			wantErrMsg:       "",
			wantUserEnabled:  true,
			wantHasValidated: true,
			wantResources: scalecommon.Resources{
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
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				} else {
					assert.Nil(t, err)
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
			name: "PanicStoreFromAnnotations",
			fields: fields{
				hasStored: false,
			},
			wantPanicMsg: "StoreFromAnnotations() hasn't been invoked first",
		},
		{
			name: "PanicValidate",
			fields: fields{
				hasStored:    true,
				hasValidated: false,
			},
			wantPanicMsg: "Validate() hasn't been invoked first",
		},
		{
			name: "NotEnabled",
			fields: fields{
				csaEnabled:   false,
				hasStored:    true,
				hasValidated: true,
			},
			want: "(cpu) not enabled",
		},
		{
			name: "Enabled",
			fields: fields{
				csaEnabled:   true,
				hasStored:    true,
				hasValidated: true,
				resources:    scaletest.ResourcesCpuEnabled,
			},
			want: "(cpu) startup: " + kubetest.PodAnnotationCpuStartup +
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
