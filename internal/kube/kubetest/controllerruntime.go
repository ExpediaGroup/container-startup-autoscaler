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
	kubefake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func ControllerRuntimeFakeClientWithKubeFake(
	fakeClientFunc func() *kubefake.Clientset,
	interceptorFuncsFunc func() interceptor.Funcs,
) client.WithWatch {
	return fake.NewClientBuilder().
		WithObjectTracker(fakeClientFunc().Tracker()).
		WithInterceptorFuncs(interceptorFuncsFunc()).
		Build()
}
