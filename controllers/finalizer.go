// Licensed to Alexandre VILAIN under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Alexandre VILAIN licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package controllers

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func addDeletionFinalizer(ctx context.Context, c client.Client, obj client.Object) error {
	return patchFinalizers(ctx, c, obj, func(patched client.Object) bool {
		return controllerutil.AddFinalizer(patched, deletionFinalizer)
	})
}

func removeDeletionFinalizer(ctx context.Context, c client.Client, obj client.Object) error {
	err := patchFinalizers(ctx, c, obj, func(patched client.Object) bool {
		return controllerutil.RemoveFinalizer(patched, deletionFinalizer)
	})

	// Dropping the last finalizer completes the deletion, so a retried removal finds the object gone.
	return client.IgnoreNotFound(err)
}

// patchFinalizers writes its own copy of obj, keeping finalizers out of the reconciliation's
// deferred patch: that patch is a JSON merge patch against a snapshot, so it would rewrite
// metadata.finalizers as a whole list and undo what the API server did to foregroundDeletion.
func patchFinalizers(ctx context.Context, c client.Client, obj client.Object, mutate func(client.Object) bool) error {
	patched := obj.DeepCopyObject().(client.Object)
	if !mutate(patched) {
		return nil
	}

	return c.Patch(ctx, patched, client.MergeFromWithOptions(obj, client.MergeFromWithOptimisticLock{}))
}
