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

package pod

import (
	"errors"
	"fmt"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// scaleConfig is the default implementation of podcommon.ScaleConfig.
type scaleConfig struct {
	targetContainerName        string
	cpuConfig                  podcommon.CpuConfig
	memoryConfig               podcommon.MemoryConfig
	hasAssignedFromAnnotations bool

	kubeHelper KubeHelper
}

func NewScaleConfig(kubeHelper KubeHelper) podcommon.ScaleConfig {
	return &scaleConfig{kubeHelper: kubeHelper}
}

// GetTargetContainerName returns the target container name.
func (c *scaleConfig) GetTargetContainerName() string {
	return c.targetContainerName
}

// GetCpuConfig returns CPU configuration.
func (c *scaleConfig) GetCpuConfig() podcommon.CpuConfig {
	return c.cpuConfig
}

// GetMemoryConfig returns memory configuration.
func (c *scaleConfig) GetMemoryConfig() podcommon.MemoryConfig {
	return c.memoryConfig
}

// StoreFromAnnotations retrieves annotation values from the supplied pod and stores them by assigning to struct
// fields. Should be invoked before invoking other methods.
func (c *scaleConfig) StoreFromAnnotations(pod *v1.Pod) error {
	// TODO(wt) want to support no cpu/memory limits specified for post-startup config? Can't yet as only guaranteed
	//  post-startup configuration is possible at the moment (pod QoS is immutable).
	annErrFmt := "unable to get '%s' annotation value"
	qParseErrFmt := "unable to parse '%s' annotation value ('%s')"

	value, err := c.kubeHelper.ExpectedAnnotationValueAs(pod, podcommon.AnnotationTargetContainerName, podcommon.TypeString)
	if err != nil {
		return common.WrapErrorf(err, annErrFmt, podcommon.AnnotationTargetContainerName)
	}
	targetContainerName := value.(string)

	value, err = c.kubeHelper.ExpectedAnnotationValueAs(pod, podcommon.AnnotationCpuStartup, podcommon.TypeString)
	if err != nil {
		return common.WrapErrorf(err, annErrFmt, podcommon.AnnotationCpuStartup)
	}
	cpuStartup, err := resource.ParseQuantity(value.(string))
	if err != nil {
		return common.WrapErrorf(err, qParseErrFmt, podcommon.AnnotationCpuStartup, value)
	}

	value, err = c.kubeHelper.ExpectedAnnotationValueAs(pod, podcommon.AnnotationCpuPostStartupRequests, podcommon.TypeString)
	if err != nil {
		return common.WrapErrorf(err, annErrFmt, podcommon.AnnotationCpuPostStartupRequests)
	}
	cpuPostStartupRequests, err := resource.ParseQuantity(value.(string))
	if err != nil {
		return common.WrapErrorf(err, qParseErrFmt, podcommon.AnnotationCpuPostStartupRequests, value)
	}

	value, err = c.kubeHelper.ExpectedAnnotationValueAs(pod, podcommon.AnnotationCpuPostStartupLimits, podcommon.TypeString)
	if err != nil {
		return common.WrapErrorf(err, annErrFmt, podcommon.AnnotationCpuPostStartupLimits)
	}
	cpuPostStartupLimits, err := resource.ParseQuantity(value.(string))
	if err != nil {
		return common.WrapErrorf(err, qParseErrFmt, podcommon.AnnotationCpuPostStartupLimits, value)
	}

	value, err = c.kubeHelper.ExpectedAnnotationValueAs(pod, podcommon.AnnotationMemoryStartup, podcommon.TypeString)
	if err != nil {
		return common.WrapErrorf(err, annErrFmt, podcommon.AnnotationMemoryStartup)
	}
	memoryStartup, err := resource.ParseQuantity(value.(string))
	if err != nil {
		return common.WrapErrorf(err, qParseErrFmt, podcommon.AnnotationMemoryStartup, value)
	}

	value, err = c.kubeHelper.ExpectedAnnotationValueAs(pod, podcommon.AnnotationMemoryPostStartupRequests, podcommon.TypeString)
	if err != nil {
		return common.WrapErrorf(err, annErrFmt, podcommon.AnnotationMemoryPostStartupRequests)
	}
	memoryPostStartupRequests, err := resource.ParseQuantity(value.(string))
	if err != nil {
		return common.WrapErrorf(err, qParseErrFmt, podcommon.AnnotationMemoryPostStartupRequests, value)
	}

	value, err = c.kubeHelper.ExpectedAnnotationValueAs(pod, podcommon.AnnotationMemoryPostStartupLimits, podcommon.TypeString)
	if err != nil {
		return common.WrapErrorf(err, annErrFmt, podcommon.AnnotationMemoryPostStartupLimits)
	}
	memoryPostStartupLimits, err := resource.ParseQuantity(value.(string))
	if err != nil {
		return common.WrapErrorf(err, qParseErrFmt, podcommon.AnnotationMemoryPostStartupLimits, value)
	}

	c.targetContainerName = targetContainerName
	c.cpuConfig = podcommon.NewCpuConfig(cpuStartup, cpuPostStartupRequests, cpuPostStartupLimits)
	c.memoryConfig = podcommon.NewMemoryConfig(memoryStartup, memoryPostStartupRequests, memoryPostStartupLimits)
	c.hasAssignedFromAnnotations = true

	return nil
}

// Validate validates the configuration.
func (c *scaleConfig) Validate() error {
	if !c.hasAssignedFromAnnotations {
		panic(errors.New("StoreFromAnnotations() hasn't been invoked first"))
	}

	if common.IsStringEmpty(c.targetContainerName) {
		return errors.New("target container name is empty")
	}

	// TODO(wt) only guaranteed post-startup configuration is possible at the moment (pod QoS is immutable)
	if !c.cpuConfig.PostStartupRequests.Equal(c.cpuConfig.PostStartupLimits) {
		return fmt.Errorf(
			"cpu post-startup requests (%s) must equal post-startup limits (%s) - change in qos class is not yet permitted by kube",
			c.cpuConfig.PostStartupRequests.String(),
			c.cpuConfig.PostStartupLimits.String(),
		)
	}

	// TODO(wt) only guaranteed post-startup configuration is possible at the moment (pod QoS is immutable)
	if !c.memoryConfig.PostStartupRequests.Equal(c.memoryConfig.PostStartupLimits) {
		return fmt.Errorf(
			"memory post-startup requests (%s) must equal post-startup limits (%s) - change in qos class is not yet permitted by kube",
			c.memoryConfig.PostStartupRequests.String(),
			c.memoryConfig.PostStartupLimits.String(),
		)
	}

	if c.cpuConfig.PostStartupRequests.Cmp(c.cpuConfig.Startup) == 1 {
		return fmt.Errorf(
			"cpu post-startup requests (%s) is greater than startup value (%s)",
			c.cpuConfig.PostStartupRequests.String(),
			c.cpuConfig.Startup.String(),
		)
	}

	if c.memoryConfig.PostStartupRequests.Cmp(c.memoryConfig.Startup) == 1 {
		return fmt.Errorf(
			"memory post-startup requests (%s) is greater than startup value (%s)",
			c.memoryConfig.PostStartupRequests.String(),
			c.memoryConfig.Startup.String(),
		)
	}

	// TODO(wt) reinstate once change in qos class is permitted by Kube
	//if c.cpuConfig.PostStartupLimits.Cmp(c.cpuConfig.PostStartupRequests) == -1 {
	//	return fmt.Errorf(
	//		"cpu post-startup limits (%s) is less than post-startup requests (%s)",
	//		c.cpuConfig.PostStartupLimits.String(),
	//		c.cpuConfig.PostStartupRequests.String(),
	//	)
	//}
	//
	//if c.memoryConfig.PostStartupLimits.Cmp(c.memoryConfig.PostStartupRequests) == -1 {
	//	return fmt.Errorf(
	//		"memory post-startup limits (%s) is less than post-startup requests (%s)",
	//		c.memoryConfig.PostStartupLimits.String(),
	//		c.memoryConfig.PostStartupRequests.String(),
	//	)
	//}

	return nil
}

// String returns a string representation of the configuration.
func (c *scaleConfig) String() string {
	return fmt.Sprintf(
		"cpu: %s/%s/%s, memory: %s/%s/%s",
		c.cpuConfig.Startup.String(), c.cpuConfig.PostStartupRequests.String(), c.cpuConfig.PostStartupLimits.String(),
		c.memoryConfig.Startup.String(), c.memoryConfig.PostStartupRequests.String(), c.memoryConfig.PostStartupLimits.String(),
	)
}
