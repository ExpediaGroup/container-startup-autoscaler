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

package kubetest

import (
	"context"
	"errors"
	"sync"

	"github.com/ExpediaGroup/container-startup-autoscaler/internal/context/contexttest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ctxUuidGetInvocs              = map[string]int{}
	ctxUuidPatchInvocs            = map[string]int{}
	ctxUuidSubResourcePatchInvocs = map[string]int{}
)

var (
	getMutex              sync.Mutex
	patchMutex            sync.Mutex
	subResourcePatchMutex sync.Mutex
)

// InterceptorFuncGetFail returns an interceptor get function that fails. Returns withError if supplied, otherwise an
// error with an empty message.
func InterceptorFuncGetFail(withError ...error) func(_ context.Context, _ client.WithWatch, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
	if len(withError) > 1 {
		panic(errors.New("only 0 or 1 errors can be supplied"))
	}

	return func(_ context.Context, _ client.WithWatch, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
		if len(withError) == 0 {
			return errors.New("")
		}
		return withError[0]
	}
}

// InterceptorFuncGetFailFirstOnly returns an interceptor get function that fails on the first invocation only. Returns
// withError if supplied, otherwise an error with an empty message.
func InterceptorFuncGetFailFirstOnly(withFirstError ...error) func(_ context.Context, _ client.WithWatch, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
	if len(withFirstError) > 1 {
		panic(errors.New("only 0 or 1 errors can be supplied"))
	}

	return func(ctx context.Context, _ client.WithWatch, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
		defer getMutex.Unlock()
		getMutex.Lock()

		uuid := ctx.Value(contexttest.KeyUuid).(string)
		current, got := ctxUuidGetInvocs[uuid]
		if !got {
			ctxUuidGetInvocs[uuid] = 1
			if len(withFirstError) == 0 {
				return errors.New("")
			}
			return withFirstError[0]
		}

		ctxUuidGetInvocs[uuid] = current + 1
		return nil
	}
}

// InterceptorFuncPatchFail returns an interceptor patch function that fails. Returns withError if supplied, otherwise
// an error with an empty message.
func InterceptorFuncPatchFail(withError ...error) func(_ context.Context, _ client.WithWatch, _ client.Object, _ client.Patch, _ ...client.PatchOption) error {
	if len(withError) > 1 {
		panic(errors.New("only 0 or 1 errors can be supplied"))
	}

	return func(_ context.Context, _ client.WithWatch, _ client.Object, _ client.Patch, _ ...client.PatchOption) error {
		if len(withError) == 0 {
			return errors.New("")
		}
		return withError[0]
	}
}

// InterceptorFuncPatchFailFirstOnly returns an interceptor patch function that fails on the first invocation only.
// Returns withError if supplied, otherwise an error with an empty message.
func InterceptorFuncPatchFailFirstOnly(withFirstError ...error) func(_ context.Context, _ client.WithWatch, _ client.Object, _ client.Patch, _ ...client.PatchOption) error {
	if len(withFirstError) > 1 {
		panic(errors.New("only 0 or 1 errors can be supplied"))
	}

	return func(ctx context.Context, _ client.WithWatch, _ client.Object, _ client.Patch, _ ...client.PatchOption) error {
		defer patchMutex.Unlock()
		patchMutex.Lock()

		uuid := ctx.Value(contexttest.KeyUuid).(string)
		current, got := ctxUuidPatchInvocs[uuid]
		if !got {
			ctxUuidPatchInvocs[uuid] = 1
			if len(withFirstError) == 0 {
				return errors.New("")
			}
			return withFirstError[0]
		}

		ctxUuidPatchInvocs[uuid] = current + 1
		return nil
	}
}

// InterceptorFuncSubResourcePatchFail returns an interceptor subresource patch function that fails. Returns withError
// if supplied, otherwise an error with an empty message.
func InterceptorFuncSubResourcePatchFail(withError ...error) func(_ context.Context, _ client.Client, _ string, _ client.Object, _ client.Patch, _ ...client.SubResourcePatchOption) error {
	if len(withError) > 1 {
		panic(errors.New("only 0 or 1 errors can be supplied"))
	}

	return func(_ context.Context, _ client.Client, _ string, _ client.Object, _ client.Patch, _ ...client.SubResourcePatchOption) error {
		if len(withError) == 0 {
			return errors.New("")
		}
		return withError[0]
	}
}

// InterceptorFuncSubResourcePatchFailFirstOnly returns an interceptor subresource patch function that fails on the
// first invocation only. Returns withError if supplied, otherwise an error with an empty message.
func InterceptorFuncSubResourcePatchFailFirstOnly(withFirstError ...error) func(_ context.Context, _ client.Client, _ string, _ client.Object, _ client.Patch, _ ...client.SubResourcePatchOption) error {
	return func(ctx context.Context, _ client.Client, _ string, _ client.Object, _ client.Patch, _ ...client.SubResourcePatchOption) error {
		defer subResourcePatchMutex.Unlock()
		subResourcePatchMutex.Lock()

		uuid := ctx.Value(contexttest.KeyUuid).(string)
		current, got := ctxUuidSubResourcePatchInvocs[uuid]
		if !got {
			ctxUuidSubResourcePatchInvocs[uuid] = 1
			if len(withFirstError) == 0 {
				return errors.New("")
			}
			return withFirstError[0]
		}

		ctxUuidSubResourcePatchInvocs[uuid] = current + 1
		return nil
	}
}
