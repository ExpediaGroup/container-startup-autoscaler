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
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contexttest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/informercache"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/retry"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podtest"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/component-base/metrics/testutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestNewKubeHelper(t *testing.T) {
	c := fake.NewClientBuilder().Build()
	h := newKubeHelper(c)
	assert.Equal(t, c, h.client)
}

func TestKubeHelperGet(t *testing.T) {
	type args struct {
		ctx  context.Context
		name types.NamespacedName
	}
	tests := []struct {
		name       string
		client     client.Client
		args       args
		wantFound  bool
		wantPod    *v1.Pod
		wantErrMsg string
	}{
		{
			name: "UnableToGetPod",
			client: podtest.ControllerRuntimeFakeClientWithKubeFake(
				func() *kubefake.Clientset { return kubefake.NewClientset() },
				func() interceptor.Funcs { return interceptor.Funcs{Get: podtest.InterceptorFuncGetFail()} },
			),
			args: args{
				ctx:  contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				name: podtest.DefaultPodNamespacedName,
			},
			wantFound:  false,
			wantPod:    &v1.Pod{},
			wantErrMsg: "unable to get pod",
		},
		{
			name: "NotFound",
			client: podtest.ControllerRuntimeFakeClientWithKubeFake(
				func() *kubefake.Clientset { return kubefake.NewClientset() },
				func() interceptor.Funcs { return interceptor.Funcs{} },
			),
			args: args{
				ctx:  contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				name: podtest.DefaultPodNamespacedName,
			},
			wantFound: false,
			wantPod:   &v1.Pod{},
		},
		{
			name: "Found",
			client: podtest.ControllerRuntimeFakeClientWithKubeFake(
				func() *kubefake.Clientset {
					return kubefake.NewClientset(
						podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
					)
				},
				func() interceptor.Funcs { return interceptor.Funcs{} },
			),
			args: args{
				ctx:  contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				name: podtest.DefaultPodNamespacedName,
			},
			wantFound: true,
			wantPod:   podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newKubeHelper(tt.client)

			gotFound, gotPod, err := h.Get(tt.args.ctx, tt.args.name)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.wantFound, gotFound)
			assert.Equal(t, tt.wantPod, gotPod)
		})
	}
}

func TestKubeHelperPatch(t *testing.T) {
	t.Run("UnableToMutatePod", func(t *testing.T) {
		h := newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset() },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		))

		got, err := h.Patch(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			&v1.Pod{},
			func(pod *v1.Pod) (bool, *v1.Pod, error) { return false, nil, errors.New("") },
			false,
			false,
		)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "unable to mutate pod")
	})

	t.Run("UnableToPatchPod", func(t *testing.T) {
		h := newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset() },
			func() interceptor.Funcs { return interceptor.Funcs{Patch: podtest.InterceptorFuncPatchFail()} },
		))

		got, err := h.Patch(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			&v1.Pod{},
			func(pod *v1.Pod) (bool, *v1.Pod, error) { return true, pod, nil },
			false,
			false,
		)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "unable to patch pod")
	})

	t.Run("ConflictUnableToGetPod", func(t *testing.T) {
		conflictErr := kerrors.NewConflict(schema.GroupResource{}, "", errors.New(""))
		h := newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset() },
			func() interceptor.Funcs {
				return interceptor.Funcs{
					Patch: podtest.InterceptorFuncPatchFail(conflictErr),
					Get:   podtest.InterceptorFuncGetFail(),
				}
			},
		))

		got, err := h.Patch(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			&v1.Pod{},
			func(pod *v1.Pod) (bool, *v1.Pod, error) { return true, pod, nil },
			false,
			false,
		)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "unable to get pod when resolving conflict")
	})

	t.Run("ConflictPodDoesntExist", func(t *testing.T) {
		conflictErr := kerrors.NewConflict(schema.GroupResource{}, "", errors.New(""))
		notFoundErr := kerrors.NewNotFound(schema.GroupResource{}, "")
		h := newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset() },
			func() interceptor.Funcs {
				return interceptor.Funcs{
					Patch: podtest.InterceptorFuncPatchFail(conflictErr),
					Get:   podtest.InterceptorFuncGetFail(notFoundErr),
				}
			},
		))

		got, err := h.Patch(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			&v1.Pod{},
			func(pod *v1.Pod) (bool, *v1.Pod, error) { return true, pod, nil },
			false,
			false,
		)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "pod doesn't exist when resolving conflict")
	})

	t.Run("OkNoPatchResizeTrue", func(t *testing.T) {
		cpuRequests, cpuLimits := resource.MustParse("89998m"), resource.MustParse("99999m")
		memoryRequests, memoryLimits := resource.MustParse("89998M"), resource.MustParse("99999M")
		pod := podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		podMutationFunc := func(pod *v1.Pod) (bool, *v1.Pod, error) {
			pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU] = cpuRequests
			pod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU] = cpuLimits
			pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory] = memoryRequests
			pod.Spec.Containers[0].Resources.Limits[v1.ResourceMemory] = memoryLimits
			return false, pod, nil
		}
		h := newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		))

		got, err := h.Patch(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			pod,
			podMutationFunc,
			true,
			true,
		)
		assert.Nil(t, err)
		assert.False(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceCPU].Equal(cpuRequests))
		assert.False(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceCPU].Equal(cpuLimits))
		assert.False(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceMemory].Equal(memoryRequests))
		assert.False(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceMemory].Equal(memoryLimits))

		// Ensure original pod isn't mutated
		assert.False(t, pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU].Equal(cpuRequests))
		assert.False(t, pod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU].Equal(cpuLimits))
		assert.False(t, pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory].Equal(memoryRequests))
		assert.False(t, pod.Spec.Containers[0].Resources.Limits[v1.ResourceMemory].Equal(memoryLimits))
	})

	t.Run("OkWithResolvedConflictResizeTrue", func(t *testing.T) {
		cpuRequests, cpuLimits := resource.MustParse("89998m"), resource.MustParse("99999m")
		memoryRequests, memoryLimits := resource.MustParse("89998M"), resource.MustParse("99999M")
		pod := podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		podMutationFunc := func(pod *v1.Pod) (bool, *v1.Pod, error) {
			pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU] = cpuRequests
			pod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU] = cpuLimits
			pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory] = memoryRequests
			pod.Spec.Containers[0].Resources.Limits[v1.ResourceMemory] = memoryLimits
			return true, pod, nil
		}
		conflictErr := kerrors.NewConflict(schema.GroupResource{}, "", errors.New(""))
		h := newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs {
				return interceptor.Funcs{SubResourcePatch: podtest.InterceptorFuncSubResourcePatchFailFirstOnly(conflictErr)}
			},
		))

		beforeMetricVal, _ := testutil.GetCounterMetricValue(retry.Retry(strings.ToLower(string(metav1.StatusReasonConflict))))
		got, err := h.Patch(
			contexttest.NewCtxBuilder(contexttest.NewOneRetryCtxConfig(nil)).Build(),
			pod,
			podMutationFunc,
			true,
			true,
		)
		assert.Nil(t, err)
		assert.True(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceCPU].Equal(cpuRequests))
		assert.True(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceCPU].Equal(cpuLimits))
		assert.True(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceMemory].Equal(memoryRequests))
		assert.True(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceMemory].Equal(memoryLimits))
		afterMetricVal, _ := testutil.GetCounterMetricValue(retry.Retry(strings.ToLower(string(metav1.StatusReasonConflict))))
		assert.Equal(t, beforeMetricVal+1, afterMetricVal)
	})

	t.Run("OkWithoutConflictResizeTrue", func(t *testing.T) {
		cpuRequests, cpuLimits := resource.MustParse("89998m"), resource.MustParse("99999m")
		memoryRequests, memoryLimits := resource.MustParse("89998M"), resource.MustParse("99999M")
		pod := podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		podMutationFunc := func(pod *v1.Pod) (bool, *v1.Pod, error) {
			pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU] = cpuRequests
			pod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU] = cpuLimits
			pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory] = memoryRequests
			pod.Spec.Containers[0].Resources.Limits[v1.ResourceMemory] = memoryLimits
			return true, pod, nil
		}
		h := newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		))

		got, err := h.Patch(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			pod,
			podMutationFunc,
			true,
			true,
		)
		assert.Nil(t, err)
		assert.True(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceCPU].Equal(cpuRequests))
		assert.True(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceCPU].Equal(cpuLimits))
		assert.True(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceMemory].Equal(memoryRequests))
		assert.True(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceMemory].Equal(memoryLimits))

		// Ensure original pod isn't mutated
		assert.False(t, pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU].Equal(cpuRequests))
		assert.False(t, pod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU].Equal(cpuLimits))
		assert.False(t, pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory].Equal(memoryRequests))
		assert.False(t, pod.Spec.Containers[0].Resources.Limits[v1.ResourceMemory].Equal(memoryLimits))
	})

	t.Run("OkWithoutConflictResizeFalse", func(t *testing.T) {
		pod := podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		podMutationFunc := func(pod *v1.Pod) (bool, *v1.Pod, error) {
			pod.Annotations["test"] = "test"
			return true, pod, nil
		}
		h := newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		))

		got, err := h.Patch(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			pod,
			podMutationFunc,
			false,
			false,
		)
		assert.Nil(t, err)
		assert.Equal(t, "test", got.Annotations["test"])

		// Ensure original pod isn't mutated
		_, gotAnn := pod.Annotations["test"]
		assert.False(t, gotAnn)
	})
}

func TestKubeHelperUpdateContainerResources(t *testing.T) {
	t.Run("ContainerNotPresent", func(t *testing.T) {
		h := newKubeHelper(nil)

		got, err := h.UpdateContainerResources(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			&v1.Pod{},
			"",
			resource.Quantity{}, resource.Quantity{},
			resource.Quantity{}, resource.Quantity{},
			nil,
			false,
		)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "container not present")
	})

	t.Run("UnableToPatchPodResizeSubresource", func(t *testing.T) {
		h := newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset() },
			func() interceptor.Funcs {
				return interceptor.Funcs{SubResourcePatch: podtest.InterceptorFuncSubResourcePatchFail()}
			},
		))

		got, err := h.UpdateContainerResources(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
			podtest.DefaultContainerName,
			resource.Quantity{}, resource.Quantity{},
			resource.Quantity{}, resource.Quantity{},
			nil,
			false,
		)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "unable to patch pod resize subresource")
	})

	t.Run("UnableToApplyAdditionalPodMutations", func(t *testing.T) {
		pod := podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		cpuRequests, cpuLimits := resource.MustParse("89998m"), resource.MustParse("99999m")
		memoryRequests, memoryLimits := resource.MustParse("89998M"), resource.MustParse("99999M")
		h := newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		))

		got, err := h.UpdateContainerResources(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			pod,
			podtest.DefaultContainerName,
			cpuRequests, cpuLimits,
			memoryRequests, memoryLimits,
			func(pod *v1.Pod) (bool, *v1.Pod, error) {
				return false, nil, errors.New("")
			},
			false,
		)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "unable to patch pod additional mutations")
	})

	t.Run("Ok", func(t *testing.T) {
		pod := podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		cpuRequests, cpuLimits := resource.MustParse("89998m"), resource.MustParse("99999m")
		memoryRequests, memoryLimits := resource.MustParse("89998M"), resource.MustParse("99999M")
		h := newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		))

		got, err := h.UpdateContainerResources(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			pod,
			podtest.DefaultContainerName,
			cpuRequests, cpuLimits,
			memoryRequests, memoryLimits,
			func(pod *v1.Pod) (bool, *v1.Pod, error) {
				pod.Annotations["test"] = "test"
				return true, pod, nil
			},
			false,
		)
		assert.Nil(t, err)
		assert.True(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceCPU].Equal(cpuRequests))
		assert.True(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceCPU].Equal(cpuLimits))
		assert.True(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceMemory].Equal(memoryRequests))
		assert.True(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceMemory].Equal(memoryLimits))
		assert.Equal(t, "test", got.Annotations["test"])

		// Ensure pod isn't mutated
		assert.False(t, pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU].Equal(cpuRequests))
		assert.False(t, pod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU].Equal(cpuLimits))
		assert.False(t, pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory].Equal(memoryRequests))
		assert.False(t, pod.Spec.Containers[0].Resources.Limits[v1.ResourceMemory].Equal(memoryLimits))
		_, gotAnn := pod.Annotations["test"]
		assert.False(t, gotAnn)
	})
}

func TestKubeHelperHasAnnotation(t *testing.T) {
	type args struct {
		pod  *v1.Pod
		name string
	}
	tests := []struct {
		name       string
		args       args
		wantBool   bool
		wantString string
	}{
		{
			"Has",
			args{
				podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
					AdditionalAnnotations(map[string]string{"testkey": "testvalue"}).
					Build(),
				"testkey",
			},
			true,
			"testvalue",
		},
		{
			"NotHas",
			args{
				podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
					AdditionalAnnotations(map[string]string{"testkey": "testvalue"}).
					Build(),
				"notpresent",
			},
			false,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newKubeHelper(nil)

			gotBool, gotString := h.HasAnnotation(tt.args.pod, tt.args.name)
			assert.Equal(t, tt.wantBool, gotBool)
			assert.Equal(t, tt.wantString, gotString)
		})
	}
}

func TestKubeHelperExpectedLabelValueAs(t *testing.T) {
	type args struct {
		pod  *v1.Pod
		name string
		as   podcommon.Type
	}
	tests := []struct {
		name            string
		args            args
		want            any
		wantPanicErrMsg string
		wantErrMsg      string
	}{
		{
			name: "NotPresent",
			args: args{
				pod:  podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name: "test",
				as:   podcommon.TypeString,
			},
			want:       nil,
			wantErrMsg: fmt.Sprintf("%s '%s' not present", mapForLabel, "test"),
		},
		{
			name: "UnableToParseValueAsBool",
			args: args{
				pod: podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
					AdditionalLabels(map[string]string{"test": "test"}).Build(),
				name: "test",
				as:   podcommon.TypeBool,
			},
			want:       nil,
			wantErrMsg: fmt.Sprintf("unable to parse 'test' %s value 'test' as %s", mapForLabel, podcommon.TypeBool),
		},
		{
			name: "AsNotSupported",
			args: args{
				pod: podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
					AdditionalLabels(map[string]string{"test": "test"}).Build(),
				name: "test",
				as:   "test",
			},
			wantPanicErrMsg: "as 'test' not supported",
		},
		{
			name: "Ok",
			args: args{
				pod:  podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name: podcommon.LabelEnabled,
				as:   podcommon.TypeBool,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newKubeHelper(nil)

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() { _, _ = h.ExpectedLabelValueAs(tt.args.pod, tt.args.name, tt.args.as) })
				return
			}

			got, err := h.ExpectedLabelValueAs(tt.args.pod, tt.args.name, tt.args.as)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestKubeHelperExpectedAnnotationValueAs(t *testing.T) {
	type args struct {
		pod  *v1.Pod
		name string
		as   podcommon.Type
	}
	tests := []struct {
		name            string
		args            args
		want            any
		wantPanicErrMsg string
		wantErrMsg      string
	}{
		{
			name: "Ok",
			args: args{
				pod:  podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name: podcommon.AnnotationCpuStartup,
				as:   podcommon.TypeString,
			},
			want: podtest.PodAnnotationCpuStartup,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newKubeHelper(nil)

			if tt.wantPanicErrMsg != "" {
				assert.PanicsWithError(t, tt.wantPanicErrMsg, func() { _, _ = h.ExpectedAnnotationValueAs(tt.args.pod, tt.args.name, tt.args.as) })
				return
			}

			got, err := h.ExpectedAnnotationValueAs(tt.args.pod, tt.args.name, tt.args.as)
			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestKubeHelperIsContainerInSpec(t *testing.T) {
	type args struct {
		pod           *v1.Pod
		containerName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "False",
			args: args{
				pod:           podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				containerName: "",
			},
			want: false,
		},
		{
			name: "True",
			args: args{
				pod:           podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				containerName: podtest.DefaultContainerName,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newKubeHelper(nil)

			got := h.IsContainerInSpec(tt.args.pod, tt.args.containerName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestKubeHelperResizeStatus(t *testing.T) {
	h := newKubeHelper(nil)
	pod := podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
		ContainerStatusResizeStatus(v1.PodResizeStatusInProgress).Build()

	got := h.ResizeStatus(pod)
	assert.Equal(t, v1.PodResizeStatusInProgress, got)
}

func TestKubeHelperWaitForCacheUpdate(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		pod := podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		pod.ResourceVersion = "123"
		h := newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		))

		beforeMetricVal, _ := testutil.GetHistogramMetricValue(informercache.SyncPoll())
		newPod := h.waitForCacheUpdate(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			pod,
		)
		assert.NotNil(t, newPod)
		afterMetricVal, _ := testutil.GetHistogramMetricValue(informercache.SyncPoll())
		assert.Equal(t, beforeMetricVal+1, afterMetricVal)
	})

	t.Run("Timeout", func(t *testing.T) {
		pod := podtest.NewPodBuilder(podtest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		pod.ResourceVersion = "123"
		h := newKubeHelper(podtest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(&v1.Pod{}) },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		))

		beforeMetricVal, _ := testutil.GetCounterMetricValue(informercache.SyncTimeout())
		newPod := h.waitForCacheUpdate(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			pod,
		)
		assert.Nil(t, newPod)
		afterMetricVal, _ := testutil.GetCounterMetricValue(informercache.SyncTimeout())
		assert.Equal(t, beforeMetricVal+1, afterMetricVal)
	})
}
