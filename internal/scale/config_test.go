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

func TestNewConfig(t *testing.T) {
	resourceName := v1.ResourceCPU
	config := NewConfig(
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

func TestConfigResourceName(t *testing.T) {
	resourceName := v1.ResourceCPU
	config := &config{resourceName: resourceName}
	assert.Equal(t, v1.ResourceCPU, config.ResourceName())
}

func TestConfigIsEnabled(t *testing.T) {
	type fields struct {
		csaEnabled               bool
		hasStoredFromAnnotations bool
	}
	tests := []struct {
		name         string
		fields       fields
		wantPanicMsg string
		want         bool
	}{
		{
			name: "Panic",
			fields: fields{
				hasStoredFromAnnotations: false,
			},
			wantPanicMsg: "StoreFromAnnotations() hasn't been invoked first",
		},
		{
			name: "True",
			fields: fields{
				csaEnabled:               true,
				hasStoredFromAnnotations: true,
			},
			want: true,
		},
		{
			name: "False",
			fields: fields{
				csaEnabled:               false,
				hasStoredFromAnnotations: true,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config{
				csaEnabled:               tt.fields.csaEnabled,
				hasStoredFromAnnotations: tt.fields.hasStoredFromAnnotations,
				userEnabled:              true,
			}
			if tt.wantPanicMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicMsg, func() { c.IsEnabled() })
			} else {
				assert.Equal(t, tt.want, c.IsEnabled())
			}
		})
	}
}

func TestConfigResources(t *testing.T) {
	type fields struct {
		csaEnabled               bool
		hasStoredFromAnnotations bool
		resources                scalecommon.Resources
	}
	tests := []struct {
		name         string
		fields       fields
		wantPanicMsg string
		want         scalecommon.Resources
	}{
		{
			name: "Panic",
			fields: fields{
				hasStoredFromAnnotations: false,
			},
			wantPanicMsg: "StoreFromAnnotations() hasn't been invoked first",
		},
		{
			name: "Ok",
			fields: fields{
				hasStoredFromAnnotations: true,
				resources:                scaletest.ResourcesCpuStartupEnabled,
			},
			want: scaletest.ResourcesCpuStartupEnabled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config{
				csaEnabled:               true,
				hasStoredFromAnnotations: tt.fields.hasStoredFromAnnotations,
				userEnabled:              true,
				resources:                tt.fields.resources,
			}
			if tt.wantPanicMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicMsg, func() { c.Resources() })
			} else {
				assert.Equal(t, tt.want, c.Resources())
			}
		})
	}
}

func TestConfigStoreFromAnnotations(t *testing.T) {
	type fields struct {
		annotationStartupName             string
		annotationPostStartupRequestsName string
		annotationPostStartupLimitsName   string
		csaEnabled                        bool
		podHelper                         kubecommon.PodHelper
	}
	tests := []struct {
		name                         string
		fields                       fields
		wantErrMsg                   string
		wantUserEnabled              bool
		wantHasStoredFromAnnotations bool
		wantResources                scalecommon.Resources
	}{
		{
			name:                         "NotCsaEnabled",
			fields:                       fields{csaEnabled: false},
			wantErrMsg:                   "",
			wantUserEnabled:              false,
			wantHasStoredFromAnnotations: true,
			wantResources:                scalecommon.Resources{},
		},
		{
			name: "NotUserEnabled",
			fields: fields{
				csaEnabled: true,
				podHelper: kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) {
					m.On("HasAnnotation", mock.Anything, mock.Anything).Return(false, "")
				}),
			},
			wantErrMsg:                   "",
			wantUserEnabled:              false,
			wantHasStoredFromAnnotations: true,
			wantResources:                scalecommon.Resources{},
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
			wantErrMsg:                   "unable to get '" + scalecommon.AnnotationCpuStartup + "' annotation value",
			wantUserEnabled:              false,
			wantHasStoredFromAnnotations: false,
			wantResources:                scalecommon.Resources{},
		},
		{
			name: "UnableToParseStartupAnnotationValue",
			fields: fields{
				annotationStartupName: scalecommon.AnnotationCpuStartup,
				csaEnabled:            true,
				podHelper: kubetest.NewMockPodHelper(func(m *kubetest.MockPodHelper) {
					m.On("ExpectedAnnotationValueAs", mock.Anything, mock.Anything, mock.Anything).
						Return("test", nil)
					m.HasAnnotationDefault()
				}),
			},
			wantErrMsg:                   "unable to parse '" + scalecommon.AnnotationCpuStartup + "' annotation value ('test')",
			wantUserEnabled:              false,
			wantHasStoredFromAnnotations: false,
			wantResources:                scalecommon.Resources{},
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
			wantErrMsg:                   "unable to get '" + scalecommon.AnnotationCpuPostStartupRequests + "' annotation value",
			wantUserEnabled:              false,
			wantHasStoredFromAnnotations: false,
			wantResources:                scalecommon.Resources{},
		},
		{
			name: "UnableToParsePostStartupRequestsAnnotationValue",
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
					).Return("test", nil)
					m.HasAnnotationDefault()
				}),
			},
			wantErrMsg:                   "unable to parse '" + scalecommon.AnnotationCpuPostStartupRequests + "' annotation value ('test')",
			wantUserEnabled:              false,
			wantHasStoredFromAnnotations: false,
			wantResources:                scalecommon.Resources{},
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
			wantErrMsg:                   "unable to get '" + scalecommon.AnnotationCpuPostStartupLimits + "' annotation value",
			wantUserEnabled:              false,
			wantHasStoredFromAnnotations: false,
			wantResources:                scalecommon.Resources{},
		},
		{
			name: "UnableToParsePostStartupLimitsAnnotationValue",
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
					).Return("test", nil)
					m.HasAnnotationDefault()
				}),
			},
			wantErrMsg:                   "unable to parse '" + scalecommon.AnnotationCpuPostStartupLimits + "' annotation value ('test')",
			wantUserEnabled:              false,
			wantHasStoredFromAnnotations: false,
			wantResources:                scalecommon.Resources{},
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
			wantErrMsg:                   "",
			wantUserEnabled:              true,
			wantHasStoredFromAnnotations: true,
			wantResources:                scaletest.ResourcesCpuStartupEnabled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config{
				annotationStartupName:             tt.fields.annotationStartupName,
				annotationPostStartupRequestsName: tt.fields.annotationPostStartupRequestsName,
				annotationPostStartupLimitsName:   tt.fields.annotationPostStartupLimitsName,
				csaEnabled:                        tt.fields.csaEnabled,
				podHelper:                         tt.fields.podHelper,
			}
			err := c.StoreFromAnnotations(&v1.Pod{})
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.wantUserEnabled, c.userEnabled)
			assert.Equal(t, tt.wantHasStoredFromAnnotations, c.hasStoredFromAnnotations)
			assert.Equal(t, tt.wantResources, c.resources)
		})
	}
}

func TestConfigValidate(t *testing.T) {
	type fields struct {
		csaEnabled               bool
		containerHelper          kubecommon.ContainerHelper
		hasStoredFromAnnotations bool
		resources                scalecommon.Resources
	}
	tests := []struct {
		name         string
		fields       fields
		wantPanicMsg string
		wantErrMsg   string
	}{
		{
			name: "Panics",
			fields: fields{
				hasStoredFromAnnotations: false,
			},
			wantPanicMsg: "StoreFromAnnotations() hasn't been invoked first",
		},
		{
			name: "NotEnabled",
			fields: fields{
				csaEnabled:               false,
				hasStoredFromAnnotations: true,
			},
			wantErrMsg: "",
		},
		{
			name: "PostStartupRequestsMustEqualPostStartupLimits",
			fields: fields{
				csaEnabled:               true,
				hasStoredFromAnnotations: true,
				resources: scalecommon.Resources{
					PostStartupRequests: resource.MustParse("1m"),
					PostStartupLimits:   resource.MustParse("2m"),
				},
			},
			wantErrMsg: "cpu post-startup requests (1m) must equal post-startup limits (2m)",
		},
		{
			name: "PostStartupRequestsGreaterThanStartupValue",
			fields: fields{
				csaEnabled:               true,
				hasStoredFromAnnotations: true,
				resources: scalecommon.Resources{
					Startup:             resource.MustParse("1m"),
					PostStartupRequests: resource.MustParse("2m"),
					PostStartupLimits:   resource.MustParse("2m"),
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
				hasStoredFromAnnotations: true,
				resources: scalecommon.Resources{
					Startup:             resource.MustParse("2m"),
					PostStartupRequests: resource.MustParse("1m"),
					PostStartupLimits:   resource.MustParse("1m"),
				},
			},
			wantErrMsg: "target container does not specify cpu requests",
		},
		{
			name: "TargetContainerRequestsMustEqualLimits",
			fields: fields{
				csaEnabled: true,
				containerHelper: kubetest.NewMockContainerHelper(func(m *kubetest.MockContainerHelper) {
					m.On("Requests", mock.Anything, mock.Anything).Return(resource.MustParse("1m"))
					m.On("Limits", mock.Anything, mock.Anything).Return(resource.MustParse("2m"))
				}),
				hasStoredFromAnnotations: true,
				resources: scalecommon.Resources{
					Startup:             resource.MustParse("2m"),
					PostStartupRequests: resource.MustParse("1m"),
					PostStartupLimits:   resource.MustParse("1m"),
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
				hasStoredFromAnnotations: true,
				resources: scalecommon.Resources{
					Startup:             resource.MustParse("2m"),
					PostStartupRequests: resource.MustParse("1m"),
					PostStartupLimits:   resource.MustParse("1m"),
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
				hasStoredFromAnnotations: true,
				resources: scalecommon.Resources{
					Startup:             resource.MustParse("2m"),
					PostStartupRequests: resource.MustParse("1m"),
					PostStartupLimits:   resource.MustParse("1m"),
				},
			},
			wantErrMsg: "target container cpu resize policy is not 'NotRequired' ('RestartContainer')",
		},
		{
			name: "Ok",
			fields: fields{
				csaEnabled:               true,
				containerHelper:          kubetest.NewMockContainerHelper(nil),
				hasStoredFromAnnotations: true,
				resources: scalecommon.Resources{
					Startup:             resource.MustParse("2m"),
					PostStartupRequests: resource.MustParse("1m"),
					PostStartupLimits:   resource.MustParse("1m"),
				},
			},
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config{
				resourceName:             v1.ResourceCPU,
				csaEnabled:               tt.fields.csaEnabled,
				containerHelper:          tt.fields.containerHelper,
				userEnabled:              true,
				hasStoredFromAnnotations: tt.fields.hasStoredFromAnnotations,
				resources:                tt.fields.resources,
			}
			if tt.wantPanicMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicMsg, func() { _ = c.Validate(&v1.Container{}) })
			} else {
				err := c.Validate(&v1.Container{})
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				} else {
					assert.Nil(t, err)
				}
			}
		})
	}
}

func TestConfigString(t *testing.T) {
	type fields struct {
		csaEnabled               bool
		hasStoredFromAnnotations bool
		resources                scalecommon.Resources
	}
	tests := []struct {
		name         string
		fields       fields
		wantPanicMsg string
		want         string
	}{
		{
			name: "Panics",
			fields: fields{
				hasStoredFromAnnotations: false,
			},
			wantPanicMsg: "StoreFromAnnotations() hasn't been invoked first",
		},
		{
			name: "NotEnabled",
			fields: fields{
				csaEnabled:               false,
				hasStoredFromAnnotations: true,
			},
			want: "(cpu) not enabled",
		},
		{
			name: "Enabled",
			fields: fields{
				csaEnabled:               true,
				hasStoredFromAnnotations: true,
				resources:                scaletest.ResourcesCpuStartupEnabled,
			},
			want: "(cpu) startup: " + kubetest.PodAnnotationCpuStartup +
				", post-startup requests: " + kubetest.PodAnnotationCpuPostStartupRequests +
				", post-startup limits: " + kubetest.PodAnnotationCpuPostStartupLimits,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config{
				resourceName:             v1.ResourceCPU,
				csaEnabled:               tt.fields.csaEnabled,
				userEnabled:              true,
				hasStoredFromAnnotations: tt.fields.hasStoredFromAnnotations,
				resources:                tt.fields.resources,
			}
			if tt.wantPanicMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicMsg, func() { _ = c.String() })
			} else {
				assert.Equal(t, tt.want, c.String())
			}
		})
	}
}
