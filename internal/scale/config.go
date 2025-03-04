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

// TODO(wt) tests, comments
// TODO(wt) update go to latest 1.23
// TODO(wt) check all method comments - might be out of date now
// TODO(wt) ensure integration tests with only one of cpu/memory enabled
// TODO(wt) ensure docs up to date completely - cpu/memory now optional (but at least one required)

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

type config struct {
	resourceName                      v1.ResourceName
	annotationStartupName             string
	annotationPostStartupRequestsName string
	annotationPostStartupLimitsName   string
	csaEnabled                        bool
	podHelper                         kubecommon.PodHelper
	containerHelper                   kubecommon.ContainerHelper

	hasStoredFromAnnotations bool
	userEnabled              bool
	resources                scalecommon.Resources
}

func NewConfig(
	resourceName v1.ResourceName,
	annotationStartupName string,
	annotationPostStartupRequestsName string,
	annotationPostStartupLimitsName string,
	csaEnabled bool,
	podHelper kubecommon.PodHelper,
	containerHelper kubecommon.ContainerHelper,
) scalecommon.Config {
	return &config{
		resourceName:                      resourceName,
		annotationStartupName:             annotationStartupName,
		annotationPostStartupRequestsName: annotationPostStartupRequestsName,
		annotationPostStartupLimitsName:   annotationPostStartupLimitsName,
		csaEnabled:                        csaEnabled,
		podHelper:                         podHelper,
		containerHelper:                   containerHelper,
	}
}

func (c *config) ResourceName() v1.ResourceName {
	return c.resourceName
}

func (c *config) IsEnabled() bool {
	c.checkStored()
	return c.csaEnabled && c.userEnabled
}

func (c *config) Resources() scalecommon.Resources {
	c.checkStored()
	return c.resources
}

func (c *config) StoreFromAnnotations(pod *v1.Pod) error {
	if !c.csaEnabled {
		c.hasStoredFromAnnotations = true
		return nil
	}

	hasStartupAnn, _ := c.podHelper.HasAnnotation(pod, c.annotationStartupName)
	hasPostStartupRequestsAnn, _ := c.podHelper.HasAnnotation(pod, c.annotationPostStartupRequestsName)
	hasPostStartupLimitsAnn, _ := c.podHelper.HasAnnotation(pod, c.annotationPostStartupLimitsName)

	if !hasStartupAnn && !hasPostStartupRequestsAnn && !hasPostStartupLimitsAnn {
		c.userEnabled = false
		c.hasStoredFromAnnotations = true
		return nil
	}

	annErrFmt := "unable to get '%s' annotation value"
	qParseErrFmt := "unable to parse '%s' annotation value ('%s')"

	value, err := c.podHelper.ExpectedAnnotationValueAs(pod, c.annotationStartupName, kubecommon.DataTypeString)
	if err != nil {
		return common.WrapErrorf(err, annErrFmt, c.annotationStartupName)
	}
	startup, err := resource.ParseQuantity(value.(string))
	if err != nil {
		return common.WrapErrorf(err, qParseErrFmt, c.annotationStartupName, value)
	}

	value, err = c.podHelper.ExpectedAnnotationValueAs(pod, c.annotationPostStartupRequestsName, kubecommon.DataTypeString)
	if err != nil {
		return common.WrapErrorf(err, annErrFmt, c.annotationPostStartupRequestsName)
	}
	postStartupRequests, err := resource.ParseQuantity(value.(string))
	if err != nil {
		return common.WrapErrorf(err, qParseErrFmt, c.annotationPostStartupRequestsName, value)
	}

	value, err = c.podHelper.ExpectedAnnotationValueAs(pod, c.annotationPostStartupLimitsName, kubecommon.DataTypeString)
	if err != nil {
		return common.WrapErrorf(err, annErrFmt, c.annotationPostStartupLimitsName)
	}
	postStartupLimits, err := resource.ParseQuantity(value.(string))
	if err != nil {
		return common.WrapErrorf(err, qParseErrFmt, c.annotationPostStartupLimitsName, value)
	}

	c.userEnabled = true
	c.resources = scalecommon.NewResources(startup, postStartupRequests, postStartupLimits)
	c.hasStoredFromAnnotations = true

	return nil
}

func (c *config) Validate(container *v1.Container) error {
	c.checkStored()

	if !c.csaEnabled || !c.userEnabled {
		return nil
	}

	// TODO(wt) QoS class is currently immutable so post-startup resources must also be 'guaranteed'. See
	//  https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources#qos-class
	if !c.resources.PostStartupRequests.Equal(c.resources.PostStartupLimits) {
		return fmt.Errorf(
			"%s post-startup requests (%s) must equal post-startup limits (%s)",
			c.resourceName,
			c.resources.PostStartupRequests.String(),
			c.resources.PostStartupLimits.String(),
		)
	}

	if c.resources.PostStartupRequests.Cmp(c.resources.Startup) == 1 {
		return fmt.Errorf(
			"%s post-startup requests (%s) is greater than startup value (%s)",
			c.resourceName,
			c.resources.PostStartupRequests.String(),
			c.resources.Startup.String(),
		)
	}

	requests := c.containerHelper.Requests(container, c.resourceName)
	if requests.IsZero() {
		return fmt.Errorf("target container does not specify %s requests", c.resourceName)
	}

	limits := c.containerHelper.Limits(container, c.resourceName)
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

	return nil
}

func (c *config) String() string {
	c.checkStored()

	if !c.csaEnabled || !c.userEnabled {
		return fmt.Sprintf("(%s) not enabled", c.resourceName)
	}

	return fmt.Sprintf(
		"(%s) startup: %s, post-startup requests: %s, post-startup limits: %s",
		c.resourceName,
		c.resources.Startup.String(),
		c.resources.PostStartupRequests.String(),
		c.resources.PostStartupLimits.String(),
	)
}

func (c *config) checkStored() {
	if !c.hasStoredFromAnnotations {
		panic(errors.New("StoreFromAnnotations() hasn't been invoked first"))
	}
}
