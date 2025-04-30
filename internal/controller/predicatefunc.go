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

package controller

import (
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/common"
	csaevent "github.com/ExpediaGroup/container-startup-autoscaler/internal/event"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/event/eventcommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/kube/kubecommon"
	"github.com/ExpediaGroup/container-startup-autoscaler/internal/metrics/reconciler"
	"k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

/*
	Informer cache sync and re-syncing notes
	----------------------------------------

	- Upon startup, the informer populates the local object store (cache) with the relevant objects via the Kube API,
	  each of which always results in an 'add' event (predicateCreateFunc() is invoked).

	- The informer periodically re-syncs so to mitigate any bugs either in controller-runtime or this controller i.e.
	  not requeuing an object via the reconciler when it should have been. This results in an 'update' event for each
	  object in the local cache (predicateUpdateFunc() is invoked) - the objects are NOT re-obtained via the Kube API.
	  Deletions are NOT re-synced.

	Predicate function notes
	------------------------

	- The predicate functions below determine whether events should be reconciled (per event type). The general premise
	  is to minimise (on a best-effort basis) the amount of unnecessary work performed by the reconciler by filtering
	  out events that are known to not require reconciling.

	- Filtering out does not guarantee that such events won't be reconciled since the reconciler itself *always
	  retrieves the latest version of the pod*, which actually may have changed since being filtered out here. In
	  addition, some reconciles are requeued after a period of time - after which the latest version of the pod is again
	  retrieved. Reconciling events that otherwise would have been filtered out here will still operate correctly.

	- See https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/predicate for information on controller-runtime
	  predicates.

	- All predicate functions publish events to an event publisher, which reconcile logic can subscribe to if there's a
	  need to observe events during operation.
*/

// predicateCreateFunc returns whether create events should be reconciled.
func predicateCreateFunc(event event.TypedCreateEvent[*v1.Pod]) bool {
	csaevent.DefaultPodEventPublisher.Publish(
		nil,
		eventcommon.NewPodEvent(eventcommon.PodEventTypeCreate, event.Object),
	)

	// Never filter.
	return true
}

// predicateDeleteFunc returns whether delete events should be reconciled.
func predicateDeleteFunc(event event.TypedDeleteEvent[*v1.Pod]) bool {
	csaevent.DefaultPodEventPublisher.Publish(
		nil,
		eventcommon.NewPodEvent(eventcommon.PodEventTypeDelete, event.Object),
	)

	// Don't need to reconcile deletes.
	return false
}

// predicateUpdateFunc returns whether update events should be reconciled.
func predicateUpdateFunc(event event.TypedUpdateEvent[*v1.Pod]) bool {
	oldPod := event.ObjectOld
	newPod := event.ObjectNew

	csaevent.DefaultPodEventPublisher.Publish(
		nil,
		eventcommon.NewPodEvent(eventcommon.PodEventTypeUpdate, newPod),
	)

	if oldPod.ResourceVersion == newPod.ResourceVersion {
		// Shouldn't really find ourselves here...
		return false
	}

	// Don't reconcile pods that are deleting.
	if !newPod.DeletionTimestamp.IsZero() {
		return false
	}

	_, hasOldAnnStatusString := oldPod.Annotations[kubecommon.AnnotationStatus]
	_, hasNewAnnStatusString := newPod.Annotations[kubecommon.AnnotationStatus]

	// Reconcile if controller status not present in old and new (something non-status has changed).
	if !hasOldAnnStatusString && !hasNewAnnStatusString {
		return true
	}

	// Don't reconcile pods that *only* have an updated controller status.

	// Nested as deep struct comparison is expensive.
	if common.AreStructsEqual(oldPod.TypeMeta, newPod.TypeMeta) {
		// TypeMeta hasn't changed.

		if common.AreStructsEqual(oldPod.Spec, newPod.Spec) {
			// Spec hasn't changed.

			if common.AreStructsEqual(oldPod.Status, newPod.Status) {
				// Only ObjectMeta has changed.

				// Remove variable fields from old and new to compare ObjectMeta - if same only status has been updated.
				oldPodCopy, newPodCopy := oldPod.DeepCopy(), newPod.DeepCopy()
				oldPodCopy.ResourceVersion, newPodCopy.ResourceVersion = "", ""
				oldPodCopy.ManagedFields, newPodCopy.ManagedFields = nil, nil
				delete(oldPodCopy.Annotations, kubecommon.AnnotationStatus)
				delete(newPodCopy.Annotations, kubecommon.AnnotationStatus)

				if common.AreStructsEqual(oldPodCopy.ObjectMeta, newPodCopy.ObjectMeta) {
					reconciler.SkippedOnlyStatusChange().Inc()
					return false
				}
			}
		}
	}

	return true
}

// predicateGenericFunc returns whether generic events should be reconciled.
func predicateGenericFunc(_ event.TypedGenericEvent[*v1.Pod]) bool {
	return false
}
