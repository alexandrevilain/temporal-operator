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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1alpha1 "github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/pkg/cluster"
	"github.com/alexandrevilain/temporal-operator/pkg/persistence"
)

const (
	ownerKey  = ".metadata.controller"
	ownerKind = "TemporalCluster"
)

// TemporalClusterReconciler reconciles a TemporalCluster object
type TemporalClusterReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	PersistenceManager *persistence.Manager
}

//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps.alexandrevilain.dev,resources=temporalclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.alexandrevilain.dev,resources=temporalclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.alexandrevilain.dev,resources=temporalclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *TemporalClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	temporalCluster := &appsv1alpha1.TemporalCluster{}
	err := r.Get(ctx, req.NamespacedName, temporalCluster)
	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	} else if apierrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	// Check if the resource has been marked for deletion
	if !temporalCluster.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Deleting temporal cluster", "name", temporalCluster.Name)
		return ctrl.Result{}, nil
	}

	// Set defaults on unfilled fields.
	temporalCluster.Default()

	if err := r.reconcilePersistence(ctx, temporalCluster); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileResources(ctx, temporalCluster); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *TemporalClusterReconciler) reconcilePersistence(ctx context.Context, temporalCluster *appsv1alpha1.TemporalCluster) error {
	if err := r.PersistenceManager.RunDefaultStoreSchemaTasks(ctx, temporalCluster); err != nil {
		return err
	}

	if err := r.PersistenceManager.RunVisibilityStoreSchemaTasks(ctx, temporalCluster); err != nil {
		return err
	}

	return nil
}

func (r *TemporalClusterReconciler) reconcileResources(ctx context.Context, temporalCluster *appsv1alpha1.TemporalCluster) error {
	logger := log.FromContext(ctx)

	clusterBuilder := cluster.TemporalClusterBuilder{
		Instance: temporalCluster,
		Scheme:   r.Scheme,
	}

	builders, err := clusterBuilder.ResourceBuilders()
	if err != nil {
		return err
	}

	logger.Info("Retrieved builders", "count", len(builders))

	for _, builder := range builders {
		resource, err := builder.Build()
		if err != nil {
			return err
		}
		operationResult, err := controllerutil.CreateOrUpdate(ctx, r.Client, resource, func() error {
			return builder.Update(resource)
		})
		if err != nil {
			action := r.operationResultToAction(operationResult)
			msg := fmt.Sprintf("failed to %s %T %s", action, resource, resource.GetName())
			logger.Error(err, msg)
			return err
		}
		if operationResult != controllerutil.OperationResultNone {
			msg := fmt.Sprintf("%s %T %s", operationResult, resource, resource.GetName())
			logger.Info(msg)
		}
	}

	return nil
}

func (r *TemporalClusterReconciler) operationResultToAction(operationResult controllerutil.OperationResult) string {
	var action string
	switch operationResult {
	case controllerutil.OperationResultCreated:
		action = "create"
	case controllerutil.OperationResultUpdated:
		action = "update"
	}
	return action
}

// SetupWithManager sets up the controller with the Manager.
func (r *TemporalClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	for _, resource := range []client.Object{&appsv1.Deployment{}, &corev1.ConfigMap{}, &corev1.Service{}} {
		if err := mgr.GetFieldIndexer().IndexField(context.Background(), resource, ownerKey, addResourceToIndex); err != nil {
			return err
		}
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.TemporalCluster{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func addResourceToIndex(rawObj client.Object) []string {
	switch resourceObject := rawObj.(type) {
	case *appsv1.Deployment:
		owner := metav1.GetControllerOf(resourceObject)
		return validateAndGetOwner(owner)
	case *corev1.ConfigMap:
		owner := metav1.GetControllerOf(resourceObject)
		return validateAndGetOwner(owner)
	case *corev1.Service:
		owner := metav1.GetControllerOf(resourceObject)
		return validateAndGetOwner(owner)
	default:
		return nil
	}
}

func validateAndGetOwner(owner *metav1.OwnerReference) []string {
	if owner == nil {
		return nil
	}
	if owner.APIVersion != appsv1alpha1.GroupVersion.String() || owner.Kind != ownerKind {
		return nil
	}
	return []string{owner.Name}
}
