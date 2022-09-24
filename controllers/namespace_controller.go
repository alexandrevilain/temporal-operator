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
	"fmt"
	"time"

	"go.temporal.io/api/serviceerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/temporal"
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=temporal.io,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=temporal.io,resources=namespaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=temporal.io,resources=namespaces/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Starting reconciliation")

	namespace := &v1beta1.Namespace{}
	err := r.Get(ctx, req.NamespacedName, namespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	namespacedName := types.NamespacedName{Namespace: req.Namespace, Name: namespace.Spec.ClusterRef.Name}
	cluster := &v1beta1.Cluster{}
	err = r.Get(ctx, namespacedName, cluster)
	if err != nil {
		return r.handleError(ctx, namespace, v1beta1.ReconcileErrorReason, err)
	}

	// Check if the resource has been marked for deletion
	if !namespace.ObjectMeta.DeletionTimestamp.IsZero() {
		// Namespace deletion is not supported in temporal for now
		// See: https://github.com/temporalio/temporal/issues/1679
		return reconcile.Result{}, nil
	}

	client, err := temporal.GetClusterNamespaceClient(ctx, r.Client, cluster)
	if err != nil {
		err = fmt.Errorf("can't create cluster namespace client: %w", err)
		return r.handleError(ctx, namespace, v1beta1.ReconcileErrorReason, err)
	}
	defer client.Close()

	err = client.Register(ctx, temporal.NamespaceToRegisterNamespaceRequest(namespace))
	if err != nil {
		_, ok := err.(*serviceerror.NamespaceAlreadyExists)
		if !ok {
			err = fmt.Errorf("can't create \"%s\" namespace: %w", namespace.GetName(), err)
			return r.handleError(ctx, namespace, v1beta1.ReconcileErrorReason, err)
		}
	}

	logger.Info("Successfully reconciled namespace", "namespace", namespace.GetName())

	return r.handleSuccess(ctx, namespace)
}

func (r *NamespaceReconciler) handleSuccess(ctx context.Context, namespace *v1beta1.Namespace) (ctrl.Result, error) {
	return r.handleSuccessWithRequeue(ctx, namespace, 0)
}

func (r *NamespaceReconciler) handleError(ctx context.Context, namespace *v1beta1.Namespace, reason string, err error) (ctrl.Result, error) {
	return r.handleErrorWithRequeue(ctx, namespace, reason, err, 0)
}

func (r *NamespaceReconciler) handleSuccessWithRequeue(ctx context.Context, namespace *v1beta1.Namespace, requeueAfter time.Duration) (ctrl.Result, error) {
	v1beta1.SetNamespaceReconcileSuccess(namespace, metav1.ConditionTrue, v1beta1.ReconcileSuccessReason, "")
	err := r.updateNamespaceStatus(ctx, namespace)
	return reconcile.Result{RequeueAfter: requeueAfter}, err
}

func (r *NamespaceReconciler) handleErrorWithRequeue(ctx context.Context, namespace *v1beta1.Namespace, reason string, err error, requeueAfter time.Duration) (ctrl.Result, error) {
	if reason == "" {
		reason = v1beta1.ReconcileErrorReason
	}
	v1beta1.SetNamespaceReconcileError(namespace, metav1.ConditionTrue, reason, err.Error())
	err = r.updateNamespaceStatus(ctx, namespace)
	return reconcile.Result{RequeueAfter: requeueAfter}, err
}

func (r *NamespaceReconciler) updateNamespaceStatus(ctx context.Context, namespace *v1beta1.Namespace) error {
	err := r.Status().Update(ctx, namespace)
	if err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Namespace{}, builder.WithPredicates(predicate.Or(
			predicate.GenerationChangedPredicate{},
			predicate.LabelChangedPredicate{},
			predicate.AnnotationChangedPredicate{},
		))).
		Complete(r)
}
