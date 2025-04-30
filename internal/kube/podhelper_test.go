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

package kube

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contexttest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/event/eventcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/event/eventtest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubetest"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/informercache"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/retry"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/scale/scalecommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
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
	assert.Equal(t, &podHelper{client: c}, NewPodHelper(c))
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
		wantErrMsg string
		wantFound  bool
		wantPod    *v1.Pod
	}{
		{
			"UnableToGetPod",
			kubetest.ControllerRuntimeFakeClientWithKubeFake(
				func() *kubefake.Clientset { return kubefake.NewClientset() },
				func() interceptor.Funcs { return interceptor.Funcs{Get: kubetest.InterceptorFuncGetFail()} },
			),
			args{
				contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				kubetest.DefaultPodNamespacedName,
			},
			"unable to get pod",
			false,
			&v1.Pod{},
		},
		{
			"NotFound",
			kubetest.ControllerRuntimeFakeClientWithKubeFake(
				func() *kubefake.Clientset { return kubefake.NewClientset() },
				func() interceptor.Funcs { return interceptor.Funcs{} },
			),
			args{
				contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				kubetest.DefaultPodNamespacedName,
			},
			"",
			false,
			&v1.Pod{},
		},
		{
			"Found",
			kubetest.ControllerRuntimeFakeClientWithKubeFake(
				func() *kubefake.Clientset {
					return kubefake.NewClientset(kubetest.NewPodBuilder().Build())
				},
				func() interceptor.Funcs { return interceptor.Funcs{} },
			),
			args{
				contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
				kubetest.DefaultPodNamespacedName,
			},
			"",
			true,
			kubetest.NewPodBuilder().Build(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewPodHelper(tt.client)

			gotFound, gotPod, err := h.Get(tt.args.ctx, tt.args.name)
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
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
			nil,
			&v1.Pod{},
			[]func(*v1.Pod) (bool, func(*v1.Pod) bool, error){
				func(*v1.Pod) (bool, func(*v1.Pod) bool, error) { return false, nil, errors.New("") },
			},
			false,
		)
		assert.Nil(t, got)
		assert.ErrorContains(t, err, "unable to mutate pod")
	})

	t.Run("UnableToPatchPod", func(t *testing.T) {
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset() },
			func() interceptor.Funcs { return interceptor.Funcs{Patch: kubetest.InterceptorFuncPatchFail()} },
		))

		got, err := h.Patch(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			nil,
			&v1.Pod{},
			[]func(*v1.Pod) (bool, func(*v1.Pod) bool, error){
				func(*v1.Pod) (bool, func(*v1.Pod) bool, error) { return true, nil, nil },
			},
			false,
		)
		assert.Nil(t, got)
		assert.ErrorContains(t, err, "unable to patch pod")
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
			nil,
			&v1.Pod{},
			[]func(*v1.Pod) (bool, func(*v1.Pod) bool, error){
				func(*v1.Pod) (bool, func(*v1.Pod) bool, error) { return true, nil, nil },
			},
			false,
		)
		assert.Nil(t, got)
		assert.ErrorContains(t, err, "unable to get pod when resolving conflict")
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
			nil,
			&v1.Pod{},
			[]func(*v1.Pod) (bool, func(*v1.Pod) bool, error){
				func(*v1.Pod) (bool, func(*v1.Pod) bool, error) { return true, nil, nil },
			},
			false,
		)
		assert.Nil(t, got)
		assert.ErrorContains(t, err, "pod doesn't exist when resolving conflict")
	})

	t.Run("OkNoPatchResizeTrue", func(t *testing.T) {
		pod := kubetest.NewPodBuilder().Build()
		podMutationFunc1 := func(*v1.Pod) (bool, func(*v1.Pod) bool, error) { return false, nil, nil }
		podMutationFunc2 := func(*v1.Pod) (bool, func(*v1.Pod) bool, error) { return false, nil, nil }
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		))

		got, err := h.Patch(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			nil,
			pod,
			[]func(*v1.Pod) (bool, func(*v1.Pod) bool, error){podMutationFunc1, podMutationFunc2},
			true,
		)
		assert.NoError(t, err)
		assert.Equal(t, pod, got)
		assert.True(t, pod == got)
	})

	t.Run("OkNoChangeResizeTrue", func(t *testing.T) {
		pod := kubetest.NewPodBuilder().Build()
		podMutationFunc1 := func(mutatePod *v1.Pod) (bool, func(*v1.Pod) bool, error) {
			mutatePod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU] = kubetest.PodCpuPostStartupRequestsEnabled
			mutatePod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU] = kubetest.PodCpuPostStartupLimitsEnabled
			return true, nil, nil
		}
		podMutationFunc2 := func(mutatePod *v1.Pod) (bool, func(*v1.Pod) bool, error) {
			mutatePod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory] = kubetest.PodMemoryPostStartupRequestsEnabled
			mutatePod.Spec.Containers[0].Resources.Limits[v1.ResourceMemory] = kubetest.PodMemoryPostStartupLimitsEnabled
			return true, nil, nil
		}
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		))

		got, err := h.Patch(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			nil,
			pod,
			[]func(*v1.Pod) (bool, func(*v1.Pod) bool, error){podMutationFunc1, podMutationFunc2},
			true,
		)
		assert.NoError(t, err)
		assert.True(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceCPU].Equal(kubetest.PodCpuPostStartupRequestsEnabled))
		assert.True(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceCPU].Equal(kubetest.PodCpuPostStartupLimitsEnabled))
		assert.True(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceMemory].Equal(kubetest.PodMemoryPostStartupRequestsEnabled))
		assert.True(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceMemory].Equal(kubetest.PodMemoryPostStartupLimitsEnabled))

		// Ensure original pod isn't mutated
		assert.False(t, pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU].Equal(kubetest.PodCpuPostStartupRequestsEnabled))
		assert.False(t, pod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU].Equal(kubetest.PodCpuPostStartupLimitsEnabled))
		assert.False(t, pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory].Equal(kubetest.PodMemoryPostStartupRequestsEnabled))
		assert.False(t, pod.Spec.Containers[0].Resources.Limits[v1.ResourceMemory].Equal(kubetest.PodMemoryPostStartupLimitsEnabled))
	})

	t.Run("OkWithResolvedConflictResizeTrue", func(t *testing.T) {
		pod := kubetest.NewPodBuilder().Build()
		podMutationFunc1 := func(mutatePod *v1.Pod) (bool, func(*v1.Pod) bool, error) {
			mutatePod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU] = kubetest.PodCpuPostStartupRequestsEnabled
			mutatePod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU] = kubetest.PodCpuPostStartupLimitsEnabled
			return true, nil, nil
		}
		podMutationFunc2 := func(mutatePod *v1.Pod) (bool, func(*v1.Pod) bool, error) {
			mutatePod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory] = kubetest.PodMemoryPostStartupRequestsEnabled
			mutatePod.Spec.Containers[0].Resources.Limits[v1.ResourceMemory] = kubetest.PodMemoryPostStartupLimitsEnabled
			return true, nil, nil
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
			nil,
			pod,
			[]func(*v1.Pod) (bool, func(*v1.Pod) bool, error){podMutationFunc1, podMutationFunc2},
			true,
		)
		assert.NoError(t, err)
		assert.True(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceCPU].Equal(kubetest.PodCpuPostStartupRequestsEnabled))
		assert.True(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceCPU].Equal(kubetest.PodCpuPostStartupLimitsEnabled))
		assert.True(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceMemory].Equal(kubetest.PodMemoryPostStartupRequestsEnabled))
		assert.True(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceMemory].Equal(kubetest.PodMemoryPostStartupLimitsEnabled))
		afterMetricVal, _ := testutil.GetCounterMetricValue(retry.Retry(strings.ToLower(string(metav1.StatusReasonConflict))))
		assert.Equal(t, beforeMetricVal+1, afterMetricVal)
	})

	t.Run("OkWithoutConflictResizeTrue", func(t *testing.T) {
		pod := kubetest.NewPodBuilder().Build()
		podMutationFunc1 := func(mutatePod *v1.Pod) (bool, func(*v1.Pod) bool, error) {
			mutatePod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU] = kubetest.PodCpuPostStartupRequestsEnabled
			mutatePod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU] = kubetest.PodCpuPostStartupLimitsEnabled
			return true, nil, nil
		}
		podMutationFunc2 := func(mutatePod *v1.Pod) (bool, func(*v1.Pod) bool, error) {
			mutatePod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory] = kubetest.PodMemoryPostStartupRequestsEnabled
			mutatePod.Spec.Containers[0].Resources.Limits[v1.ResourceMemory] = kubetest.PodMemoryPostStartupLimitsEnabled
			return true, nil, nil
		}
		podMutationFunc3 := func(*v1.Pod) (bool, func(*v1.Pod) bool, error) { return false, nil, nil }
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		))

		got, err := h.Patch(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			nil,
			pod,
			[]func(*v1.Pod) (bool, func(*v1.Pod) bool, error){podMutationFunc1, podMutationFunc2, podMutationFunc3},
			true,
		)
		assert.NoError(t, err)
		assert.True(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceCPU].Equal(kubetest.PodCpuPostStartupRequestsEnabled))
		assert.True(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceCPU].Equal(kubetest.PodCpuPostStartupLimitsEnabled))
		assert.True(t, got.Spec.Containers[0].Resources.Requests[v1.ResourceMemory].Equal(kubetest.PodMemoryPostStartupRequestsEnabled))
		assert.True(t, got.Spec.Containers[0].Resources.Limits[v1.ResourceMemory].Equal(kubetest.PodMemoryPostStartupLimitsEnabled))

		// Ensure original pod isn't mutated
		assert.False(t, pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU].Equal(kubetest.PodCpuPostStartupRequestsEnabled))
		assert.False(t, pod.Spec.Containers[0].Resources.Limits[v1.ResourceCPU].Equal(kubetest.PodCpuPostStartupLimitsEnabled))
		assert.False(t, pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory].Equal(kubetest.PodMemoryPostStartupRequestsEnabled))
		assert.False(t, pod.Spec.Containers[0].Resources.Limits[v1.ResourceMemory].Equal(kubetest.PodMemoryPostStartupLimitsEnabled))
	})

	t.Run("OkWithoutConflictResizeFalse", func(t *testing.T) {
		pod := kubetest.NewPodBuilder().Build()
		podMutationFunc1 := func(*v1.Pod) (bool, func(*v1.Pod) bool, error) { return false, nil, nil }
		podMutationFunc2 := func(mutatePod *v1.Pod) (bool, func(*v1.Pod) bool, error) {
			mutatePod.Annotations["test"] = "test"
			return true, nil, nil
		}
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		))

		got, err := h.Patch(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			nil,
			pod,
			[]func(*v1.Pod) (bool, func(*v1.Pod) bool, error){podMutationFunc1, podMutationFunc2},
			false,
		)
		assert.NoError(t, err)
		assert.Equal(t, "test", got.Annotations["test"])

		// Ensure original pod isn't mutated
		_, gotAnn := pod.Annotations["test"]
		assert.False(t, gotAnn)
	})

	t.Run("OkShouldWaitForCacheUpdate", func(t *testing.T) {
		subscribeCalled := false
		unsubscribeCalled := false

		pod := kubetest.NewPodBuilder().Build()
		podMutationFunc := func(mutatePod *v1.Pod) (bool, func(*v1.Pod) bool, error) {
			mutatePod.Annotations["test"] = "test"
			waitCacheConditionsMetFunc := func(currentPod *v1.Pod) bool { return true }
			return true, waitCacheConditionsMetFunc, nil
		}
		ch := make(chan eventcommon.PodEvent, 1)
		chPod := kubetest.NewPodBuilder().Build()
		ch <- eventcommon.NewPodEvent(eventcommon.PodEventTypeUpdate, chPod)
		publisher := eventtest.NewMockPodEventPublisher(func(m *eventtest.MockPodEventPublisher) {
			m.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).
				Return(ch).
				Run(func(_ mock.Arguments) { subscribeCalled = true })
			m.On("Unsubscribe", mock.Anything).Run(func(_ mock.Arguments) { unsubscribeCalled = true })
		})
		h := NewPodHelper(kubetest.ControllerRuntimeFakeClientWithKubeFake(
			func() *kubefake.Clientset { return kubefake.NewClientset(pod) },
			func() interceptor.Funcs { return interceptor.Funcs{} },
		))

		got, err := h.Patch(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			publisher,
			pod,
			[]func(*v1.Pod) (bool, func(*v1.Pod) bool, error){podMutationFunc},
			false,
		)
		assert.NoError(t, err)
		assert.Same(t, chPod, got)
		assert.True(t, subscribeCalled)
		assert.True(t, unsubscribeCalled)
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
				kubetest.NewPodBuilder().AdditionalAnnotations(map[string]string{"testkey": "testvalue"}).Build(),
				"testkey",
			},
			true,
			"testvalue",
		},
		{
			"NotHas",
			args{
				kubetest.NewPodBuilder().AdditionalAnnotations(map[string]string{"testkey": "testvalue"}).Build(),
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
		wantPanicErrMsg string
		wantErrMsg      string
		want            any
	}{
		{
			"NotPresent",
			args{
				kubetest.NewPodBuilder().Build(),
				"test",
				kubecommon.DataTypeString,
			},
			"",
			fmt.Sprintf("%s '%s' not present", mapForLabel, "test"),
			nil,
		},
		{
			"UnableToParseValueAsBool",
			args{
				kubetest.NewPodBuilder().AdditionalLabels(map[string]string{"test": "test"}).Build(),
				"test",
				kubecommon.DataTypeBool,
			},
			"",
			fmt.Sprintf("unable to parse 'test' %s value 'test' as %s", mapForLabel, kubecommon.DataTypeBool),
			nil,
		},
		{
			"AsNotSupported",
			args{
				kubetest.NewPodBuilder().AdditionalLabels(map[string]string{"test": "test"}).Build(),
				"test",
				"test",
			},
			"as 'test' not supported",
			"",
			nil,
		},
		{
			"Ok",
			args{
				kubetest.NewPodBuilder().Build(),
				kubecommon.LabelEnabled,
				kubecommon.DataTypeBool,
			},
			"",
			"",
			true,
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
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
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
		wantPanicErrMsg string
		wantErrMsg      string
		want            any
	}{
		{
			"Ok",
			args{
				kubetest.NewPodBuilder().Build(),
				scalecommon.AnnotationCpuStartup,
				kubecommon.DataTypeString,
			},
			"",
			"",
			kubetest.PodAnnotationCpuStartup,
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
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
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
			"False",
			args{
				kubetest.NewPodBuilder().Build(),
				"",
			},
			false,
		},
		{
			"True",
			args{
				kubetest.NewPodBuilder().Build(),
				kubetest.DefaultContainerName,
			},
			true,
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

func TestPodHelperResizeConditions(t *testing.T) {
	t.Run("OkWithoutConditions", func(t *testing.T) {
		h := NewPodHelper(nil)
		pod := kubetest.NewPodBuilder().ResizeConditions().Build()

		got := h.ResizeConditions(pod)
		assert.Nil(t, got)
	})

	t.Run("OkWithConditions", func(t *testing.T) {
		h := NewPodHelper(nil)
		condition1 := v1.PodCondition{Type: v1.PodResizePending}
		condition2 := v1.PodCondition{Type: "othertype"}
		condition3 := v1.PodCondition{Type: v1.PodResizeInProgress}
		pod := kubetest.NewPodBuilder().ResizeConditions(condition1, condition2, condition3).Build()

		got := h.ResizeConditions(pod)
		assert.Equal(t, []v1.PodCondition{condition1, condition3}, got)
	})
}

func TestPodHelperQOSClass(t *testing.T) {
	t.Run("NotPresent", func(t *testing.T) {
		h := NewPodHelper(nil)
		pod := kubetest.NewPodBuilder().QOSClassNotPresent().Build()

		got, err := h.QOSClass(pod)
		assert.Error(t, err, "pod status qos class not present")
		assert.Equal(t, v1.PodQOSClass(""), got)
	})

	t.Run("Ok", func(t *testing.T) {
		h := NewPodHelper(nil)
		pod := kubetest.NewPodBuilder().Build()

		got, err := h.QOSClass(pod)
		assert.NoError(t, err)
		assert.Equal(t, v1.PodQOSGuaranteed, got)
	})
}

func TestPodHelperShouldWaitForCacheUpdate(t *testing.T) {
	t.Run("FalseZeroFuncs", func(t *testing.T) {
		h := podHelper{}

		got := h.shouldWaitForCacheUpdate(nil)
		assert.False(t, got)
	})

	t.Run("FalseAllNilFuncs", func(t *testing.T) {
		h := podHelper{}

		got := h.shouldWaitForCacheUpdate([]func(*v1.Pod) bool{nil, nil})
		assert.False(t, got)
	})

	t.Run("True", func(t *testing.T) {
		h := podHelper{}

		got := h.shouldWaitForCacheUpdate([]func(*v1.Pod) bool{func(pod *v1.Pod) bool { return true }})
		assert.True(t, got)
	})
}

func TestPodHelperWaitForCacheUpdate(t *testing.T) {
	t.Run("Panics", func(t *testing.T) {
		h := podHelper{}
		ch := make(chan eventcommon.PodEvent, 1)
		ch <- eventcommon.NewPodEvent(eventcommon.PodEventTypeCreate, &v1.Pod{})

		assert.PanicsWithError(
			t,
			"unexpected event type 'create'",
			func() {
				h.waitForCacheUpdate(contexttest.NewCtxBuilder(contexttest.NewCtxConfig()).Build(), nil, nil, ch)
			},
		)
	})

	t.Run("Ok", func(t *testing.T) {
		h := podHelper{}
		pod := kubetest.NewPodBuilder().Build()
		chPod := kubetest.NewPodBuilder().Build()
		ch := make(chan eventcommon.PodEvent, 1)
		ch <- eventcommon.NewPodEvent(eventcommon.PodEventTypeUpdate, chPod)
		funcs := []func(*v1.Pod) bool{
			func(currentPod *v1.Pod) bool {
				return *currentPod.Status.ContainerStatuses[0].Started == false
			},
			func(currentPod *v1.Pod) bool {
				return currentPod.Status.ContainerStatuses[0].Ready == false
			},
		}

		got := h.waitForCacheUpdate(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			pod,
			funcs,
			ch,
		)
		assert.Same(t, chPod, got)
	})

	t.Run("Timeout", func(t *testing.T) {
		h := podHelper{}
		pod := kubetest.NewPodBuilder().Build()
		chPod := kubetest.NewPodBuilder().Build()
		ch := make(chan eventcommon.PodEvent, 1)
		ch <- eventcommon.NewPodEvent(eventcommon.PodEventTypeUpdate, chPod)
		funcs := []func(*v1.Pod) bool{
			func(currentPod *v1.Pod) bool {
				return *currentPod.Status.ContainerStatuses[0].Started == false
			},
			func(currentPod *v1.Pod) bool {
				return currentPod.Status.ContainerStatuses[0].Ready == true
			},
		}

		beforeMetricVal, _ := testutil.GetCounterMetricValue(informercache.SyncTimeout())
		got := h.waitForCacheUpdate(
			contexttest.NewCtxBuilder(contexttest.NewNoRetryCtxConfig(nil)).Build(),
			pod,
			funcs,
			ch,
		)
		assert.Same(t, pod, got)
		afterMetricVal, _ := testutil.GetCounterMetricValue(informercache.SyncTimeout())
		assert.Equal(t, beforeMetricVal+1, afterMetricVal)
	})
}
