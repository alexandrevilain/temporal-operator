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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1alpha1 "github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/pkg/cluster"
	"github.com/alexandrevilain/temporal-operator/pkg/persistence"
	"github.com/alexandrevilain/temporal-operator/pkg/resource"
	"github.com/alexandrevilain/temporal-operator/pkg/status"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
)

const (
	ownerKey  = ".metadata.controller"
	ownerKind = "TemporalCluster"
)

// TemporalClusterReconciler reconciles a TemporalCluster object
type TemporalClusterReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	Recorder           record.EventRecorder
	PersistenceManager *persistence.Manager
}

//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps.alexandrevilain.dev,resources=temporalclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.alexandrevilain.dev,resources=temporalclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.alexandrevilain.dev,resources=temporalclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *TemporalClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Starting reconciliation")

	temporalCluster := &appsv1alpha1.TemporalCluster{}
	err := r.Get(ctx, req.NamespacedName, temporalCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Check if the resource has been marked for deletion
	if !temporalCluster.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Deleting temporal cluster", "name", temporalCluster.Name)
		return reconcile.Result{}, nil
	}

	// Set defaults on unfiled fields.
	updated := r.reconcileDefaults(ctx, temporalCluster)
	if updated {
		err := r.Update(ctx, temporalCluster)
		if err != nil {
			logger.Error(err, "Can't set cluster defaults")
			return r.handleError(ctx, temporalCluster, "", err)
		}
		// As we updated the instance, another reconcile will be triggered.
		return reconcile.Result{}, nil
	}
	// Validate that the cluster version is a supported one.
	clusterVersion, err := version.ParseAndValidateTemporalVersion(temporalCluster.Spec.Version)
	if err != nil {
		logger.Error(err, "Can't validate temporal version")
		return r.handleError(ctx, temporalCluster, appsv1alpha1.TemporalClusterValidationFailedReason, err)
	}

	logger.Info("Retrieved desired cluster version", "version", clusterVersion.String())

	appsv1alpha1.SetTemporalClusterReady(temporalCluster, metav1.ConditionUnknown, appsv1alpha1.ProgressingReason, "")
	err = r.updateTemporalClusterStatus(ctx, temporalCluster)
	if err != nil {
		return r.handleError(ctx, temporalCluster, "", err)
	}

	if err := r.reconcilePersistence(ctx, temporalCluster, clusterVersion); err != nil {
		logger.Error(err, "Can't reconcile persistence")
		return r.handleErrorWithRequeue(ctx, temporalCluster, appsv1alpha1.PersistenceReconciliationFailedReason, err, 2*time.Second)
	}

	if err := r.reconcileResources(ctx, temporalCluster); err != nil {
		logger.Error(err, "Can't reconcile resources")
		return r.handleErrorWithRequeue(ctx, temporalCluster, appsv1alpha1.ResourcesReconciliationFailedReason, err, 2*time.Second)
	}

	return r.handleSuccess(ctx, temporalCluster)
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
		r.logAndRecordOperationResult(ctx, temporalCluster, resource, operationResult, err)
		if err != nil {
			return err
		}
	}

	pruners := clusterBuilder.ResourcePruners()

	logger.Info("Retrieved pruners", "count", len(pruners))

	for _, pruner := range pruners {
		resource, err := pruner.Build()
		if err != nil {
			return err
		}
		err = r.Delete(ctx, resource)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
		}
		r.logAndRecordOperationResult(ctx, temporalCluster, resource, controllerutil.OperationResult("deleted"), err)
	}

	for _, builder := range builders {
		reporter, ok := builder.(resource.StatusReporter)
		if !ok {
			continue
		}

		serviceStatus, err := reporter.ReportServiceStatus(ctx, r.Client)
		if err != nil {
			return err
		}

		logger.Info("Reporting service status", "service", serviceStatus.Name)

		temporalCluster.Status.AddServiceStatus(serviceStatus)
	}

	if status.ObservedVersionMatchesDesiredVersion(temporalCluster) {
		temporalCluster.Status.Version = temporalCluster.Spec.Version
	}

	if status.IsClusterReady(temporalCluster) {
		appsv1alpha1.SetTemporalClusterReady(temporalCluster, metav1.ConditionTrue, appsv1alpha1.ServicesReadyReason, "")
	} else {
		appsv1alpha1.SetTemporalClusterReady(temporalCluster, metav1.ConditionFalse, appsv1alpha1.ServicesNotReadyReason, "")
	}

	return r.updateTemporalClusterStatus(ctx, temporalCluster)
}

func (r *TemporalClusterReconciler) handleSuccess(ctx context.Context, temporalCluster *appsv1alpha1.TemporalCluster) (ctrl.Result, error) {
	return r.handleSuccessWithRequeue(ctx, temporalCluster, 0)
}

func (r *TemporalClusterReconciler) handleError(ctx context.Context, temporalCluster *appsv1alpha1.TemporalCluster, reason string, err error) (ctrl.Result, error) {
	return r.handleErrorWithRequeue(ctx, temporalCluster, reason, err, 0)
}

func (r *TemporalClusterReconciler) handleSuccessWithRequeue(ctx context.Context, temporalCluster *appsv1alpha1.TemporalCluster, requeueAfter time.Duration) (ctrl.Result, error) {
	appsv1alpha1.SetTemporalClusterReconcileSuccess(temporalCluster, metav1.ConditionTrue, appsv1alpha1.ReconcileSuccessReason, "")
	err := r.updateTemporalClusterStatus(ctx, temporalCluster)
	return reconcile.Result{RequeueAfter: requeueAfter}, err
}

func (r *TemporalClusterReconciler) handleErrorWithRequeue(ctx context.Context, temporalCluster *appsv1alpha1.TemporalCluster, reason string, err error, requeueAfter time.Duration) (ctrl.Result, error) {
	r.Recorder.Event(temporalCluster, corev1.EventTypeWarning, "ProcessingError", err.Error())
	if reason == "" {
		reason = appsv1alpha1.ReconcileErrorReason
	}
	appsv1alpha1.SetTemporalClusterReconcileError(temporalCluster, metav1.ConditionTrue, reason, err.Error())
	err = r.updateTemporalClusterStatus(ctx, temporalCluster)
	return reconcile.Result{RequeueAfter: requeueAfter}, err
}

func (r *TemporalClusterReconciler) updateTemporalClusterStatus(ctx context.Context, temporalCluster *appsv1alpha1.TemporalCluster) error {
	err := r.Status().Update(ctx, temporalCluster)
	if err != nil {
		return err
	}
	return nil
}

func (r *TemporalClusterReconciler) logAndRecordOperationResult(ctx context.Context, temporalCluster *appsv1alpha1.TemporalCluster, resource runtime.Object, operationResult controllerutil.OperationResult, err error) {
	logger := log.FromContext(ctx)

	var (
		action string
		reason string
	)
	switch operationResult {
	case controllerutil.OperationResultCreated:
		action = "create"
		reason = "RessourceCreate"
	case controllerutil.OperationResultUpdated:
		action = "update"
		reason = "ResourceUpdate"
	case controllerutil.OperationResult("deleted"):
		action = "delete"
		reason = "ResourceDelete"
	default:
		return
	}

	if err == nil {
		msg := fmt.Sprintf("%sd resource %s of type %T", action, resource.(metav1.Object).GetName(), resource.(metav1.Object))
		reason := fmt.Sprintf("%sSucess", reason)
		logger.Info(msg)
		r.Recorder.Event(temporalCluster, corev1.EventTypeNormal, reason, msg)
	}

	if err != nil {
		msg := fmt.Sprintf("failed to %s resource %s of Type %T", action, resource.(metav1.Object).GetName(), resource.(metav1.Object))
		reason := fmt.Sprintf("%sError", reason)
		logger.Error(err, msg)
		r.Recorder.Event(temporalCluster, corev1.EventTypeWarning, reason, msg)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *TemporalClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	for _, resource := range []client.Object{&appsv1.Deployment{}, &corev1.ConfigMap{}, &corev1.Service{}} {
		if err := mgr.GetFieldIndexer().IndexField(context.Background(), resource, ownerKey, addResourceToIndex); err != nil {
			return err
		}
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.TemporalCluster{}, builder.WithPredicates(predicate.Or(
			predicate.GenerationChangedPredicate{},
			predicate.LabelChangedPredicate{},
			predicate.AnnotationChangedPredicate{},
		))).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&networkingv1.Ingress{}).
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
	case *networkingv1.Ingress:
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
