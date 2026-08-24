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
	"time"

	"github.com/alexandrevilain/controller-tools/pkg/patch"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
)

func createTestNamespace(ctx context.Context, name string, finalizers ...string) *v1beta1.TemporalNamespace {
	namespace := &v1beta1.TemporalNamespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  "default",
			Finalizers: finalizers,
		},
		Spec: v1beta1.TemporalNamespaceSpec{
			ClusterRef:      v1beta1.ObjectReference{Name: "test-cluster"},
			RetentionPeriod: &metav1.Duration{Duration: time.Hour},
			AllowDeletion:   true,
		},
	}
	Expect(k8sClient.Create(ctx, namespace)).To(Succeed())

	return namespace
}

// createTestCluster reports the cluster ready so namespace reconciliations get past the readiness gate.
func createTestCluster(ctx context.Context, name string) *v1beta1.TemporalCluster {
	cluster := &v1beta1.TemporalCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: v1beta1.TemporalClusterSpec{
			NumHistoryShards: 1,
			Persistence: v1beta1.TemporalPersistenceSpec{
				DefaultStore:    &v1beta1.DatastoreSpec{},
				VisibilityStore: &v1beta1.DatastoreSpec{},
			},
		},
	}
	Expect(k8sClient.Create(ctx, cluster)).To(Succeed())

	v1beta1.SetTemporalClusterReady(cluster, metav1.ConditionTrue, v1beta1.ServicesReadyReason, "")
	Expect(k8sClient.Status().Update(ctx, cluster)).To(Succeed())

	return cluster
}

// deleteAfterGetClient stands in for an object that goes away between a reconciliation's Get and
// its finalizer patch.
type deleteAfterGetClient struct {
	client.Client
	target client.ObjectKey
}

func (c *deleteAfterGetClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if err := c.Client.Get(ctx, key, obj, opts...); err != nil {
		return err
	}

	if key != c.target {
		return nil
	}

	return c.Client.Delete(ctx, obj)
}

func getTestNamespace(ctx context.Context, key client.ObjectKey) *v1beta1.TemporalNamespace {
	namespace := &v1beta1.TemporalNamespace{}
	Expect(k8sClient.Get(ctx, key, namespace)).To(Succeed())

	return namespace
}

// deleteInForeground makes the API server append foregroundDeletion to the object's finalizers.
func deleteInForeground(ctx context.Context, obj client.Object) {
	Expect(k8sClient.Delete(ctx, obj, client.PropagationPolicy(metav1.DeletePropagationForeground))).To(Succeed())
}

// runGarbageCollector strips foregroundDeletion by hand: envtest runs no controller-manager to do it.
func runGarbageCollector(ctx context.Context, key client.ObjectKey) {
	namespace := getTestNamespace(ctx, key)

	patched := namespace.DeepCopy()
	Expect(controllerutil.RemoveFinalizer(patched, metav1.FinalizerDeleteDependents)).To(BeTrue())
	Expect(k8sClient.Patch(ctx, patched, client.MergeFrom(namespace))).To(Succeed())
}

var _ = Describe("Finalizers", func() {
	ctx := context.Background()

	Context("When an object is deleted with foreground propagation", func() {
		It("gets foregroundDeletion appended to its finalizers by the API server", func() {
			namespace := createTestNamespace(ctx, "foreground-delete", deletionFinalizer)
			key := client.ObjectKeyFromObject(namespace)

			deleteInForeground(ctx, namespace)

			deleting := getTestNamespace(ctx, key)
			Expect(deleting.DeletionTimestamp.IsZero()).To(BeFalse())
			Expect(deleting.Finalizers).To(ConsistOf(deletionFinalizer, metav1.FinalizerDeleteDependents))
		})

		It("has the deferred whole-object patch rejected once the garbage collector has run", func() {
			namespace := createTestNamespace(ctx, "deferred-patch-rejected", deletionFinalizer)
			key := client.ObjectKeyFromObject(namespace)

			deleteInForeground(ctx, namespace)

			snapshot := getTestNamespace(ctx, key)
			Expect(snapshot.Finalizers).To(ContainElement(metav1.FinalizerDeleteDependents))

			patchHelper, err := patch.NewHelper(snapshot, k8sClient)
			Expect(err).NotTo(HaveOccurred())

			runGarbageCollector(ctx, key)

			Expect(controllerutil.RemoveFinalizer(snapshot, deletionFinalizer)).To(BeTrue())

			err = patchHelper.Patch(ctx, snapshot)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no new finalizers can be added if the object is being deleted"))
			Expect(err.Error()).To(ContainSubstring(metav1.FinalizerDeleteDependents))
		})
	})

	Context("When adding a finalizer", func() {
		It("does not write when the finalizer is already set", func() {
			namespace := createTestNamespace(ctx, "add-no-write-when-set", deletionFinalizer)
			key := client.ObjectKeyFromObject(namespace)

			snapshot := getTestNamespace(ctx, key)
			Expect(addDeletionFinalizer(ctx, k8sClient, snapshot)).To(Succeed())
			Expect(getTestNamespace(ctx, key).ResourceVersion).To(Equal(snapshot.ResourceVersion))
		})

		It("conflicts when the object is being deleted", func() {
			namespace := createTestNamespace(ctx, "add-conflicts-on-deleting-object")
			key := client.ObjectKeyFromObject(namespace)

			snapshot := getTestNamespace(ctx, key)
			deleteInForeground(ctx, namespace)

			err := addDeletionFinalizer(ctx, k8sClient, snapshot)
			Expect(apierrors.IsConflict(err)).To(BeTrue())
			Expect(getTestNamespace(ctx, key).Finalizers).To(ConsistOf(metav1.FinalizerDeleteDependents))
		})

		It("fails when the object is already deleted", func() {
			namespace := createTestNamespace(ctx, "add-on-deleted-object")
			key := client.ObjectKeyFromObject(namespace)

			snapshot := getTestNamespace(ctx, key)
			Expect(k8sClient.Delete(ctx, namespace)).To(Succeed())

			err := addDeletionFinalizer(ctx, k8sClient, snapshot)
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})

		It("fails the controller's finalizer step when the object is already deleted", func() {
			namespace := createTestNamespace(ctx, "add-failure-reaches-reconciliation")
			key := client.ObjectKeyFromObject(namespace)

			snapshot := getTestNamespace(ctx, key)
			Expect(k8sClient.Delete(ctx, namespace)).To(Succeed())

			reconciler := &TemporalNamespaceReconciler{Client: k8sClient}
			Expect(apierrors.IsNotFound(reconciler.ensureFinalizer(ctx, snapshot))).To(BeTrue())
		})

		It("ends the reconciliation cleanly when the object goes away mid-cycle", func() {
			createTestCluster(ctx, "test-cluster")

			namespace := createTestNamespace(ctx, "add-on-object-deleted-mid-reconciliation")
			key := client.ObjectKeyFromObject(namespace)

			reconciler := &TemporalNamespaceReconciler{
				Client: &deleteAfterGetClient{Client: k8sClient, target: key},
			}

			result, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: key})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))

			Expect(apierrors.IsNotFound(k8sClient.Get(ctx, key, &v1beta1.TemporalNamespace{}))).To(BeTrue())
		})
	})

	Context("When removing a finalizer", func() {
		It("drops the deletion finalizer and the object with it", func() {
			namespace := createTestNamespace(ctx, "remove-releases-object", deletionFinalizer)
			key := client.ObjectKeyFromObject(namespace)

			deleteInForeground(ctx, namespace)
			runGarbageCollector(ctx, key)

			Expect(removeDeletionFinalizer(ctx, k8sClient, getTestNamespace(ctx, key))).To(Succeed())

			err := k8sClient.Get(ctx, key, &v1beta1.TemporalNamespace{})
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})

		It("conflicts when foregroundDeletion was removed underneath the snapshot", func() {
			namespace := createTestNamespace(ctx, "remove-conflicts-on-stale-snapshot", deletionFinalizer)
			key := client.ObjectKeyFromObject(namespace)

			deleteInForeground(ctx, namespace)

			snapshot := getTestNamespace(ctx, key)
			runGarbageCollector(ctx, key)

			err := removeDeletionFinalizer(ctx, k8sClient, snapshot)
			Expect(apierrors.IsConflict(err)).To(BeTrue())
			Expect(getTestNamespace(ctx, key).Finalizers).To(ConsistOf(deletionFinalizer))
		})

		It("conflicts when foregroundDeletion was added underneath the snapshot", func() {
			namespace := createTestNamespace(ctx, "remove-keeps-foreground-deletion", deletionFinalizer)
			key := client.ObjectKeyFromObject(namespace)

			snapshot := getTestNamespace(ctx, key)
			deleteInForeground(ctx, namespace)

			err := removeDeletionFinalizer(ctx, k8sClient, snapshot)
			Expect(apierrors.IsConflict(err)).To(BeTrue())
			Expect(getTestNamespace(ctx, key).Finalizers).To(ConsistOf(deletionFinalizer, metav1.FinalizerDeleteDependents))
		})

		It("does not write when the finalizer is already gone", func() {
			namespace := createTestNamespace(ctx, "remove-no-write-when-absent")
			key := client.ObjectKeyFromObject(namespace)

			snapshot := getTestNamespace(ctx, key)
			Expect(removeDeletionFinalizer(ctx, k8sClient, snapshot)).To(Succeed())
			Expect(getTestNamespace(ctx, key).ResourceVersion).To(Equal(snapshot.ResourceVersion))
		})

		It("succeeds when the object is already deleted", func() {
			namespace := createTestNamespace(ctx, "gone-before-removal", deletionFinalizer)
			key := client.ObjectKeyFromObject(namespace)

			snapshot := getTestNamespace(ctx, key)

			Expect(k8sClient.Delete(ctx, namespace)).To(Succeed())
			Expect(removeDeletionFinalizer(ctx, k8sClient, getTestNamespace(ctx, key))).To(Succeed())

			// The stale snapshot still carries the finalizer, so this retry reaches a gone object.
			Expect(removeDeletionFinalizer(ctx, k8sClient, snapshot)).To(Succeed())
		})
	})
})
