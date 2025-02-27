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

package kube

// TODO(wt) tests fixed in this file
import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contexttest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/informercache"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/retry"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/pod/podcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
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

func TestNewPodHelper(t *testing.T) {
	c := fake.NewClientBuilder().Build()
	assert.NotNil(t, NewPodHelper(c))
}

func TestPodHelperGet(t *testing.T) {
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
			client: kubetest.ControllerRuntimeFakeClientWithKubeFake(
				func() *kubefake.Clientset { return kubefake.NewClientset() },
				func() interceptor.Funcs { return interceptor.Funcs{Get: kubetest.InterceptorFuncGetFail()} },
			),
			args: args{
				ctx:  contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				name: kubetest.DefaultPodNamespacedName,
			},
			wantFound:  false,
			wantPod:    &v1.Pod{},
			wantErrMsg: "unable to get pod",
		},
		{
			name: "NotFound",
			client: kubetest.ControllerRuntimeFakeClientWithKubeFake(
				func() *kubefake.Clientset { return kubefake.NewClientset() },
				func() interceptor.Funcs { return interceptor.Funcs{} },
			),
			args: args{
				ctx:  contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				name: kubetest.DefaultPodNamespacedName,
			},
			wantFound: false,
			wantPod:   &v1.Pod{},
		},
		{
			name: "Found",
			client: kubetest.ControllerRuntimeFakeClientWithKubeFake(
				func() *kubefake.Clientset {
					return kubefake.NewClientset(
						kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
					)
				},
				func() interceptor.Funcs { return interceptor.Funcs{} },
			),
			args: args{
				ctx:  contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				name: kubetest.DefaultPodNamespacedName,
			},
			wantFound: true,
			wantPod:   kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewPodHelper(tt.client)

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

func TestPodHelperPatch(t *testing.T) {
	t.Run("UnableToMutatePod", func(t *testing.T) {
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
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
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset() },
			func() interceptor.Funcs { return interceptor.Funcs{Patch: kubetest.InterceptorFuncPatchFail()} },
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
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset() },
			func() interceptor.Funcs {
				return interceptor.Funcs{
					Patch: kubetest.InterceptorFuncPatchFail(conflictErr),
					Get:   kubetest.InterceptorFuncGetFail(),
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
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset() },
			func() interceptor.Funcs {
				return interceptor.Funcs{
					Patch: kubetest.InterceptorFuncPatchFail(conflictErr),
					Get:   kubetest.InterceptorFuncGetFail(notFoundErr),
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
		pod := kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		podMutationFunc := func(pod *v1.Pod) (bool, *v1.Pod, error) {
			pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU] = cpuRequests
			pod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU] = cpuLimits
			pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory] = memoryRequests
			pod.Spec.Containers[0].Resources.Limits[v1.ResourceMemory] = memoryLimits
			return false, pod, nil
		}
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
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
		pod := kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		podMutationFunc := func(pod *v1.Pod) (bool, *v1.Pod, error) {
			pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU] = cpuRequests
			pod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU] = cpuLimits
			pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory] = memoryRequests
			pod.Spec.Containers[0].Resources.Limits[v1.ResourceMemory] = memoryLimits
			return true, pod, nil
		}
		conflictErr := kerrors.NewConflict(schema.GroupResource{}, "", errors.New(""))
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs {
				return interceptor.Funcs{SubResourcePatch: kubetest.InterceptorFuncSubResourcePatchFailFirstOnly(conflictErr)}
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
		pod := kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		podMutationFunc := func(pod *v1.Pod) (bool, *v1.Pod, error) {
			pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU] = cpuRequests
			pod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU] = cpuLimits
			pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory] = memoryRequests
			pod.Spec.Containers[0].Resources.Limits[v1.ResourceMemory] = memoryLimits
			return true, pod, nil
		}
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
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
		pod := kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		podMutationFunc := func(pod *v1.Pod) (bool, *v1.Pod, error) {
			pod.Annotations["test"] = "test"
			return true, pod, nil
		}
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
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

func TestPodHelperUpdateContainerResources(t *testing.T) {
	t.Run("UnableToPatchPodResizeSubresource", func(t *testing.T) {
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset() },
			func() interceptor.Funcs {
				return interceptor.Funcs{SubResourcePatch: kubetest.InterceptorFuncSubResourcePatchFail()}
			},
		))

		got, err := h.UpdateContainerResources(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
			nil,
			false,
		)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "unable to patch pod resize subresource")
	})

	t.Run("UnableToApplyAdditionalPodMutations", func(t *testing.T) {
		pod := kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		))

		got, err := h.UpdateContainerResources(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			pod,
			func(pod *v1.Pod) (bool, *v1.Pod, error) {
				return false, nil, errors.New("")
			},
			false,
		)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "unable to patch pod additional mutations")
	})

	t.Run("Ok", func(t *testing.T) {
		pod := kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		))

		updatedPod := pod.DeepCopy()
		updatedContainer := &updatedPod.Spec.Containers[0]
		newCpuRequests, newCpuLimits := resource.MustParse("89998m"), resource.MustParse("99999m")
		newMemoryRequests, newMemoryLimits := resource.MustParse("89998M"), resource.MustParse("99999M")

		updatedContainer.Resources.Requests[v1.ResourceCPU] = newCpuRequests
		updatedContainer.Resources.Limits[v1.ResourceCPU] = newCpuLimits
		updatedContainer.Resources.Requests[v1.ResourceMemory] = newMemoryRequests
		updatedContainer.Resources.Limits[v1.ResourceMemory] = newMemoryLimits

		got, err := h.UpdateContainerResources(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			pod,
			func(pod *v1.Pod) (bool, *v1.Pod, error) {
				pod.Annotations["test"] = "test"
				return true, pod, nil
			},
			false,
		)
		assert.Nil(t, err)
		assert.True(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceCPU].Equal(newCpuRequests))
		assert.True(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceCPU].Equal(newCpuLimits))
		assert.True(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceMemory].Equal(newMemoryRequests))
		assert.True(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceMemory].Equal(newMemoryLimits))
		assert.Equal(t, "test", got.Annotations["test"])

		// Ensure pod isn't mutated
		_, gotAnn := pod.Annotations["test"]
		assert.False(t, gotAnn)
	})
}

func TestPodHelperHasAnnotation(t *testing.T) {
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
				kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
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
				kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
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
			h := NewPodHelper(nil)

			gotBool, gotString := h.HasAnnotation(tt.args.pod, tt.args.name)
			assert.Equal(t, tt.wantBool, gotBool)
			assert.Equal(t, tt.wantString, gotString)
		})
	}
}

func TestPodHelperExpectedLabelValueAs(t *testing.T) {
	type args struct {
		pod  *v1.Pod
		name string
		as   kubecommon.DataType
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
				pod:  kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name: "test",
				as:   kubecommon.DataTypeString,
			},
			want:       nil,
			wantErrMsg: fmt.Sprintf("%s '%s' not present", mapForLabel, "test"),
		},
		{
			name: "UnableToParseValueAsBool",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
					AdditionalLabels(map[string]string{"test": "test"}).Build(),
				name: "test",
				as:   kubecommon.DataTypeBool,
			},
			want:       nil,
			wantErrMsg: fmt.Sprintf("unable to parse 'test' %s value 'test' as %s", mapForLabel, kubecommon.DataTypeBool),
		},
		{
			name: "AsNotSupported",
			args: args{
				pod: kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
					AdditionalLabels(map[string]string{"test": "test"}).Build(),
				name: "test",
				as:   "test",
			},
			wantPanicErrMsg: "as 'test' not supported",
		},
		{
			name: "Ok",
			args: args{
				pod:  kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name: podcommon.LabelEnabled,
				as:   kubecommon.DataTypeBool,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewPodHelper(nil)

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

func TestPodHelperExpectedAnnotationValueAs(t *testing.T) {
	type args struct {
		pod  *v1.Pod
		name string
		as   kubecommon.DataType
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
				pod:  kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				name: scalecommon.AnnotationCpuStartup,
				as:   kubecommon.DataTypeString,
			},
			want: kubetest.PodAnnotationCpuStartup,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewPodHelper(nil)

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

func TestPodHelperIsContainerInSpec(t *testing.T) {
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
				pod:           kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				containerName: "",
			},
			want: false,
		},
		{
			name: "True",
			args: args{
				pod:           kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build(),
				containerName: kubetest.DefaultContainerName,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewPodHelper(nil)

			got := h.IsContainerInSpec(tt.args.pod, tt.args.containerName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPodHelperResizeStatus(t *testing.T) {
	h := NewPodHelper(nil)
	pod := kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).
		ContainerStatusResizeStatus(v1.PodResizeStatusInProgress).Build()

	got := h.ResizeStatus(pod)
	assert.Equal(t, v1.PodResizeStatusInProgress, got)
}

func TestPodHelperWaitForCacheUpdate(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		pod := kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		pod.ResourceVersion = "123"
		h := podHelper{
			client: kubetest.ControllerRuntimeFakeClientWithKubeFake(
				func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
				func() interceptor.Funcs { return interceptor.Funcs{} },
			),
		}

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
		pod := kubetest.NewPodBuilder(kubetest.NewStartupPodConfig(podcommon.StateBoolFalse, podcommon.StateBoolFalse)).Build()
		pod.ResourceVersion = "123"
		h := podHelper{
			client: kubetest.ControllerRuntimeFakeClientWithKubeFake(
				func() *kubefake.Clientset { return kubefake.NewClientset(&v1.Pod{}) },
				func() interceptor.Funcs { return interceptor.Funcs{} },
			),
		}

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
