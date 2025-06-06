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
	"fmt"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// configuration is the default implementation of scalecommon.Configuration.
type configuration struct {
	resourceName                      v1.ResourceName
	annotationStartupName             string
	annotationPostStartupRequestsName string
	annotationPostStartupLimitsName   string
	csaEnabled                        bool
	podHelper                         kubecommon.PodHelper
	containerHelper                   kubecommon.ContainerHelper

	hasStored    bool
	hasValidated bool
	userEnabled  bool
	rawResources scalecommon.RawResources
	resources    scalecommon.Resources
}

func NewConfiguration(
	resourceName v1.ResourceName,
	annotationStartupName string,
	annotationPostStartupRequestsName string,
	annotationPostStartupLimitsName string,
	csaEnabled bool,
	podHelper kubecommon.PodHelper,
	containerHelper kubecommon.ContainerHelper,
) scalecommon.Configuration {
	return &configuration{
		resourceName:                      resourceName,
		annotationStartupName:             annotationStartupName,
		annotationPostStartupRequestsName: annotationPostStartupRequestsName,
		annotationPostStartupLimitsName:   annotationPostStartupLimitsName,
		csaEnabled:                        csaEnabled,
		podHelper:                         podHelper,
		containerHelper:                   containerHelper,
	}
}

// ResourceName returns the resource name that the configuration relates to.
func (c *configuration) ResourceName() v1.ResourceName {
	return c.resourceName
}

// IsEnabled returns whether this configuration is enabled. Only true if both enabled by CSA and user. Panics if
// StoreFromAnnotations has not first been invoked.
func (c *configuration) IsEnabled() bool {
	c.checkStored()
	return c.csaEnabled && c.userEnabled
}

// Resources returns scalecommon.Resources stored from annotations. Panics if StoreFromAnnotations and Validate have
// not first been invoked.
func (c *configuration) Resources() scalecommon.Resources {
	c.checkStored()
	c.checkValidated()
	return c.resources
}

// StoreFromAnnotations parses and stores configuration from annotations within the supplied pod. Does nothing if not
// enabled by CSA or user.
func (c *configuration) StoreFromAnnotations(pod *v1.Pod) error {
	if !c.csaEnabled {
		c.hasStored = true
		return nil
	}

	hasStartupAnn, _ := c.podHelper.HasAnnotation(pod, c.annotationStartupName)
	hasPostStartupRequestsAnn, _ := c.podHelper.HasAnnotation(pod, c.annotationPostStartupRequestsName)
	hasPostStartupLimitsAnn, _ := c.podHelper.HasAnnotation(pod, c.annotationPostStartupLimitsName)

	if !hasStartupAnn && !hasPostStartupRequestsAnn && !hasPostStartupLimitsAnn {
		c.userEnabled = false
		c.hasStored = true
		return nil
	}

	startup, postStartupRequests, postStartupLimits := "", "", ""
	annErrFmt := "unable to get '%s' annotation value"

	if hasStartupAnn {
		value, err := c.podHelper.ExpectedAnnotationValueAs(pod, c.annotationStartupName, kubecommon.DataTypeString)
		if err != nil {
			return common.WrapErrorf(err, annErrFmt, c.annotationStartupName)
		}
		startup = value.(string)
	}

	if hasPostStartupRequestsAnn {
		value, err := c.podHelper.ExpectedAnnotationValueAs(pod, c.annotationPostStartupRequestsName, kubecommon.DataTypeString)
		if err != nil {
			return common.WrapErrorf(err, annErrFmt, c.annotationPostStartupRequestsName)
		}
		postStartupRequests = value.(string)
	}

	if hasPostStartupLimitsAnn {
		value, err := c.podHelper.ExpectedAnnotationValueAs(pod, c.annotationPostStartupLimitsName, kubecommon.DataTypeString)
		if err != nil {
			return common.WrapErrorf(err, annErrFmt, c.annotationPostStartupLimitsName)
		}
		postStartupLimits = value.(string)
	}

	c.rawResources = scalecommon.NewRawResources(startup, postStartupRequests, postStartupLimits)
	c.userEnabled = true // But subject to later validation.
	c.hasStored = true
	return nil
}

// Validate performs validation against the stored configuration and supplied container. Panics if StoreFromAnnotations
// has not first been invoked.
func (c *configuration) Validate(container *v1.Container) error {
	c.checkStored()

	if !c.IsEnabled() {
		c.hasValidated = true
		return nil
	}

	annNotPresentErrFmt := "annotation '%s' not present"

	if c.rawResources.Startup == "" {
		return fmt.Errorf(annNotPresentErrFmt, c.annotationStartupName)
	}

	if c.rawResources.PostStartupRequests == "" {
		return fmt.Errorf(annNotPresentErrFmt, c.annotationPostStartupRequestsName)
	}

	if c.rawResources.PostStartupLimits == "" {
		return fmt.Errorf(annNotPresentErrFmt, c.annotationPostStartupLimitsName)
	}

	annParseErrFmt := "unable to parse '%s' annotation value ('%s')"

	startupQuantity, err := resource.ParseQuantity(c.rawResources.Startup)
	if err != nil {
		return common.WrapErrorf(err, annParseErrFmt, c.annotationStartupName, c.rawResources.Startup)
	}

	postStartupRequestsQuantity, err := resource.ParseQuantity(c.rawResources.PostStartupRequests)
	if err != nil {
		return common.WrapErrorf(err, annParseErrFmt, c.annotationPostStartupRequestsName, c.rawResources.PostStartupRequests)
	}

	postStartupLimitsQuantity, err := resource.ParseQuantity(c.rawResources.PostStartupLimits)
	if err != nil {
		return common.WrapErrorf(err, annParseErrFmt, c.annotationPostStartupLimitsName, c.rawResources.PostStartupLimits)
	}

	// TODO(wt) QoS class is currently immutable so post-startup resources must also be 'guaranteed'. See
	//  https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources#qos-class
	if !postStartupRequestsQuantity.Equal(postStartupLimitsQuantity) {
		return fmt.Errorf(
			"%s post-startup requests (%s) must equal post-startup limits (%s)",
			c.resourceName,
			c.rawResources.PostStartupRequests,
			c.rawResources.PostStartupLimits,
		)
	}

	if postStartupRequestsQuantity.Cmp(startupQuantity) == 1 {
		return fmt.Errorf(
			"%s post-startup requests (%s) is greater than startup value (%s)",
			c.resourceName,
			c.rawResources.PostStartupRequests,
			c.rawResources.Startup,
		)
	}

	requests := c.containerHelper.Requests(container, c.resourceName)
	if requests.IsZero() {
		return fmt.Errorf("target container does not specify %s requests", c.resourceName)
	}

	limits := c.containerHelper.Limits(container, c.resourceName)
	if limits.IsZero() {
		return fmt.Errorf("target container does not specify %s limits", c.resourceName)
	}

	if !requests.Equal(limits) {
		return fmt.Errorf(
			"target container %s requests (%s) must equal limits (%s)",
			c.resourceName, requests.String(), limits.String(),
		)
	}

	resizePolicy, err := c.containerHelper.ResizePolicy(container, c.resourceName)
	if err != nil {
		return common.WrapErrorf(err, "unable to get target container %s resize policy", c.resourceName)
	}
	if resizePolicy != v1.NotRequired {
		return fmt.Errorf(
			"target container %s resize policy is not '%s' ('%s')",
			c.resourceName, v1.NotRequired, resizePolicy,
		)
	}

	c.resources = scalecommon.NewResources(startupQuantity, postStartupRequestsQuantity, postStartupLimitsQuantity)
	c.hasValidated = true
	return nil
}

// String returns a string representation of the configuration. Panics if StoreFromAnnotations has not first been
// invoked.
func (c *configuration) String() string {
	c.checkStored()

	if !c.IsEnabled() {
		return fmt.Sprintf("(%s) not enabled", c.resourceName)
	}

	return fmt.Sprintf(
		"(%s) startup: %s, post-startup requests: %s, post-startup limits: %s",
		c.resourceName,
		c.rawResources.Startup,
		c.rawResources.PostStartupRequests,
		c.rawResources.PostStartupLimits,
	)
}

// checkStored panics if StoreFromAnnotations has not been invoked.
func (c *configuration) checkStored() {
	if !c.hasStored {
		panic(errors.New("StoreFromAnnotations() hasn't been invoked first"))
	}
}

// checkValidated panics if Validate has not been invoked.
func (c *configuration) checkValidated() {
	if !c.hasValidated {
		panic(errors.New("Validate() hasn't been invoked first"))
	}
}
