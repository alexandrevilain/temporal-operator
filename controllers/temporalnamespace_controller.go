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

	appsv1alpha1 "github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/pkg/temporal"
)

// TemporalNamespaceReconciler reconciles a TemporalNamespace object
type TemporalNamespaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=temporal.io,resources=temporalnamespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=temporal.io,resources=temporalnamespaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=temporal.io,resources=temporalnamespaces/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *TemporalNamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Starting reconciliation")

	namespace := &appsv1alpha1.TemporalNamespace{}
	err := r.Get(ctx, req.NamespacedName, namespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	namespacedName := types.NamespacedName{Namespace: req.Namespace, Name: namespace.Spec.TemporalClusterRef.Name}
	temporalCluster := &appsv1alpha1.TemporalCluster{}
	err = r.Get(ctx, namespacedName, temporalCluster)
	if err != nil {
		return r.handleError(ctx, namespace, appsv1alpha1.ReconcileErrorReason, err)
	}

	// Check if the resource has been marked for deletion
	if !namespace.ObjectMeta.DeletionTimestamp.IsZero() {
		// Namespace deletion is not supported in temporal for now
		// See: https://github.com/temporalio/temporal/issues/1679
		return reconcile.Result{}, nil
	}

	client, err := temporal.GetClusterNamespaceClient(ctx, r.Client, temporalCluster)
	if err != nil {
		err = fmt.Errorf("can't create cluster namespace client: %w", err)
		return r.handleError(ctx, namespace, appsv1alpha1.ReconcileErrorReason, err)
	}
	defer client.Close()

	err = client.Register(ctx, temporal.NamespaceToRegisterNamespaceRequest(namespace))
	if err != nil {
		_, ok := err.(*serviceerror.NamespaceAlreadyExists)
		if !ok {
			err = fmt.Errorf("can't create \"%s\" namespace: %w", namespace.GetName(), err)
			return r.handleError(ctx, namespace, appsv1alpha1.ReconcileErrorReason, err)
		}
	}

	logger.Info("Successfully reconciled namespace", "namespace", namespace.GetName())

	return r.handleSuccess(ctx, namespace)
}

func (r *TemporalNamespaceReconciler) handleSuccess(ctx context.Context, namespace *appsv1alpha1.TemporalNamespace) (ctrl.Result, error) {
	return r.handleSuccessWithRequeue(ctx, namespace, 0)
}

func (r *TemporalNamespaceReconciler) handleError(ctx context.Context, namespace *appsv1alpha1.TemporalNamespace, reason string, err error) (ctrl.Result, error) {
	return r.handleErrorWithRequeue(ctx, namespace, reason, err, 0)
}

func (r *TemporalNamespaceReconciler) handleSuccessWithRequeue(ctx context.Context, namespace *appsv1alpha1.TemporalNamespace, requeueAfter time.Duration) (ctrl.Result, error) {
	appsv1alpha1.SetTemporalNamespaceReconcileSuccess(namespace, metav1.ConditionTrue, appsv1alpha1.ReconcileSuccessReason, "")
	err := r.updateTemporalNamespaceStatus(ctx, namespace)
	return reconcile.Result{RequeueAfter: requeueAfter}, err
}

func (r *TemporalNamespaceReconciler) handleErrorWithRequeue(ctx context.Context, namespace *appsv1alpha1.TemporalNamespace, reason string, err error, requeueAfter time.Duration) (ctrl.Result, error) {
	if reason == "" {
		reason = appsv1alpha1.ReconcileErrorReason
	}
	appsv1alpha1.SetTemporalNamespaceReconcileError(namespace, metav1.ConditionTrue, reason, err.Error())
	err = r.updateTemporalNamespaceStatus(ctx, namespace)
	return reconcile.Result{RequeueAfter: requeueAfter}, err
}

func (r *TemporalNamespaceReconciler) updateTemporalNamespaceStatus(ctx context.Context, namespace *appsv1alpha1.TemporalNamespace) error {
	err := r.Status().Update(ctx, namespace)
	if err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TemporalNamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.TemporalNamespace{}, builder.WithPredicates(predicate.Or(
			predicate.GenerationChangedPredicate{},
			predicate.LabelChangedPredicate{},
			predicate.AnnotationChangedPredicate{},
		))).
		Complete(r)
}
