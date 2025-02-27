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

// TODO(wt) in this package, comment methods, add tests
// TODO(wt) check license header everywhere

package config

import (
	"errors"
	"fmt"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scaleresource"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type anyScaleConfig struct{}

func (a *anyScaleConfig) checkStored(hasStoredFromAnnotations bool) {
	if !hasStoredFromAnnotations {
		panic(errors.New("StoreFromAnnotations() hasn't been invoked first"))
	}
}

func (a *anyScaleConfig) hasNoResourceAnnotations(
	podHelper kube.PodHelper,
	pod *v1.Pod,
	annotationStartup string,
	annotationPostStartupRequests string,
	annotationPostStartupLimits string,
) bool {
	hasStartupAnn, _ := podHelper.HasAnnotation(pod, annotationStartup)
	hasPostStartupRequestsAnn, _ := podHelper.HasAnnotation(pod, annotationPostStartupRequests)
	hasPostStartupLimitsAnn, _ := podHelper.HasAnnotation(pod, annotationPostStartupLimits)

	return !hasStartupAnn && !hasPostStartupRequestsAnn && !hasPostStartupLimitsAnn
}

func (a *anyScaleConfig) parseResourceAnnotations(
	podHelper kube.PodHelper,
	pod *v1.Pod,
	annotationStartup string,
	annotationPostStartupRequests string,
	annotationPostStartupLimits string,
) (startup resource.Quantity, postStartupRequests resource.Quantity, postStartupLimits resource.Quantity, retErr error) {
	annErrFmt := "unable to get '%s' annotation value"
	qParseErrFmt := "unable to parse '%s' annotation value ('%s')"

	value, err := podHelper.ExpectedAnnotationValueAs(pod, annotationStartup, kubecommon.DataTypeString)
	if err != nil {
		retErr = common.WrapErrorf(err, annErrFmt, annotationStartup)
		return
	}
	startup, err = resource.ParseQuantity(value.(string))
	if err != nil {
		retErr = common.WrapErrorf(err, qParseErrFmt, annotationStartup, value)
		return
	}

	value, err = podHelper.ExpectedAnnotationValueAs(pod, annotationPostStartupRequests, kubecommon.DataTypeString)
	if err != nil {
		retErr = common.WrapErrorf(err, annErrFmt, annotationPostStartupRequests)
		return
	}
	postStartupRequests, err = resource.ParseQuantity(value.(string))
	if err != nil {
		retErr = common.WrapErrorf(err, qParseErrFmt, annotationPostStartupRequests, value)
		return
	}

	value, err = podHelper.ExpectedAnnotationValueAs(pod, annotationPostStartupLimits, kubecommon.DataTypeString)
	if err != nil {
		retErr = common.WrapErrorf(err, annErrFmt, annotationPostStartupLimits)
		return
	}
	postStartupLimits, err = resource.ParseQuantity(value.(string))
	if err != nil {
		retErr = common.WrapErrorf(err, qParseErrFmt, annotationPostStartupLimits, value)
		return
	}

	return
}

func (a *anyScaleConfig) validate(
	resourceType scaleresource.ResourceType,
	kubeResource v1.ResourceName,
	resources Resources,
	container *v1.Container,
	containerHelper kube.ContainerHelper,
) error {
	// TODO(wt) QoS class is currently immutable so post-startup resources must also be 'guaranteed'. See
	//  https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources#qos-class
	if !resources.PostStartupRequests.Equal(resources.PostStartupLimits) {
		return fmt.Errorf(
			"%s post-startup requests (%s) must equal post-startup limits (%s)",
			resourceType,
			resources.PostStartupRequests.String(),
			resources.PostStartupLimits.String(),
		)
	}

	if resources.PostStartupRequests.Cmp(resources.Startup) == 1 {
		return fmt.Errorf(
			"%s post-startup requests (%s) is greater than startup value (%s)",
			resourceType,
			resources.PostStartupRequests.String(),
			resources.Startup.String(),
		)
	}

	requests := containerHelper.Requests(container, kubeResource)
	if requests.IsZero() {
		return fmt.Errorf("target container does not specify %s requests", resourceType)
	}

	limits := containerHelper.Limits(container, kubeResource)
	if !requests.Equal(limits) {
		return fmt.Errorf(
			"target container %s requests (%s) must equal limits (%s)",
			resourceType, requests.String(), limits.String(),
		)
	}

	resizePolicy, err := containerHelper.ResizePolicy(container, kubeResource)
	if err != nil {
		return common.WrapErrorf(err, "unable to get target container %s resize policy", resourceType)
	}
	if resizePolicy != v1.NotRequired {
		return fmt.Errorf(
			"target container %s resize policy is not '%s' ('%s')",
			resourceType, v1.NotRequired, resizePolicy,
		)
	}

	return nil
}

func (a *anyScaleConfig) string(
	resourceType scaleresource.ResourceType,
	csaEnabled bool,
	userEnabled bool,
	resources Resources,
) string {
	if !csaEnabled || !userEnabled {
		return fmt.Sprintf("(%s) not enabled", resourceType)
	}

	return fmt.Sprintf(
		"(%s) startup: %s, post-startup requests: %s, post-startup limits: %s",
		resourceType,
		resources.Startup.String(),
		resources.PostStartupRequests.String(),
		resources.PostStartupLimits.String(),
	)
}
