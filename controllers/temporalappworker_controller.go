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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	temporaliov1beta1 "github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/appworker"
	"github.com/alexandrevilain/temporal-operator/pkg/resource"
)

// TemporalAppWorkerReconciler reconciles a TemporalAppWorker object
type TemporalAppWorkerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=temporal.io,resources=temporalappworkers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=temporal.io,resources=temporalappworkers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=temporal.io,resources=temporalappworkers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TemporalAppWorker object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *TemporalAppWorkerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Starting reconciliation")

	appworker := &v1beta1.TemporalAppWorker{}
	err := r.Get(ctx, req.NamespacedName, appworker)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Check if the resource has been marked for deletion
	if !appworker.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Deleting temporal cluster", "name", appworker.Name)
		return reconcile.Result{}, nil
	}

	// Set defaults on unfiled fields.
	updated := r.reconcileDefaults(ctx, appworker)
	if updated {
		err := r.Update(ctx, appworker)
		if err != nil {
			logger.Error(err, "Can't set cluster defaults")
			return r.handleError(ctx, appworker, "", err)
		}
		// As we updated the instance, another reconcile will be triggered.
		return reconcile.Result{}, nil
	}

	if err := r.reconcileResources(ctx, appworker); err != nil {
		logger.Error(err, "Can't reconcile resources")
		return r.handleErrorWithRequeue(ctx, appworker, v1beta1.ResourcesReconciliationFailedReason, err, 2*time.Second)
	}

	return r.handleSuccess(ctx, appworker)
}

// SetupWithManager sets up the controller with the Manager.
func (r *TemporalAppWorkerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&temporaliov1beta1.TemporalAppWorker{}).
		Complete(r)
}

func (r *TemporalAppWorkerReconciler) handleErrorWithRequeue(ctx context.Context, appworker *v1beta1.TemporalAppWorker, reason string, err error, requeueAfter time.Duration) (ctrl.Result, error) {
	if reason == "" {
		reason = v1beta1.ReconcileErrorReason
	}
	v1beta1.SetTemporalAppWorkerReconcileError(appworker, metav1.ConditionTrue, reason, err.Error())
	err = r.updateClusterStatus(ctx, appworker)
	return reconcile.Result{RequeueAfter: requeueAfter}, err
}

func (r *TemporalAppWorkerReconciler) handleError(ctx context.Context, appworker *v1beta1.TemporalAppWorker, reason string, err error) (ctrl.Result, error) {
	return r.handleErrorWithRequeue(ctx, appworker, reason, err, 0)
}

func (r *TemporalAppWorkerReconciler) updateClusterStatus(ctx context.Context, cluster *v1beta1.TemporalAppWorker) error {
	err := r.Status().Update(ctx, cluster)
	if err != nil {
		return err
	}
	return nil
}

func (r *TemporalAppWorkerReconciler) reconcileResources(ctx context.Context, temporalAppWorker *v1beta1.TemporalAppWorker) error {
	logger := log.FromContext(ctx)

	appWorkerBuilder := appworker.ClusterBuilder{
		Instance: temporalAppWorker,
		Scheme:   r.Scheme,
	}

	builders, err := appWorkerBuilder.ResourceBuilders()
	if err != nil {
		return err
	}

	logger.Info("Retrieved builders", "count", len(builders))

	for _, builder := range builders {
		if comparer, ok := builder.(resource.Comparer); ok {
			err := equality.Semantic.AddFunc(comparer.Equal)
			if err != nil {
				return err
			}
		}

		res, err := builder.Build()
		if err != nil {
			return err
		}

		operationResult, err := controllerutil.CreateOrUpdate(ctx, r.Client, res, func() error {
			return builder.Update(res)
		})
		r.logAndRecordOperationResult(ctx, temporalAppWorker, res, operationResult, err)
		if err != nil {
			return err
		}
	}

	return r.updateClusterStatus(ctx, temporalAppWorker)
}

func (r *TemporalAppWorkerReconciler) logAndRecordOperationResult(ctx context.Context, appworker *v1beta1.TemporalAppWorker, resource runtime.Object, operationResult controllerutil.OperationResult, err error) {
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
		r.Recorder.Event(appworker, corev1.EventTypeNormal, reason, msg)
	}

	if err != nil {
		msg := fmt.Sprintf("failed to %s resource %s of Type %T", action, resource.(metav1.Object).GetName(), resource.(metav1.Object))
		reason := fmt.Sprintf("%sError", reason)
		logger.Error(err, msg)
		r.Recorder.Event(appworker, corev1.EventTypeWarning, reason, msg)
	}
}

func (r *TemporalAppWorkerReconciler) handleSuccess(ctx context.Context, appworker *v1beta1.TemporalAppWorker) (ctrl.Result, error) {
	return r.handleSuccessWithRequeue(ctx, appworker, 0)
}

func (r *TemporalAppWorkerReconciler) handleSuccessWithRequeue(ctx context.Context, appworker *v1beta1.TemporalAppWorker, requeueAfter time.Duration) (ctrl.Result, error) {
	v1beta1.SetTemporalAppWorkerReconcileSuccess(appworker, metav1.ConditionTrue, v1beta1.ReconcileSuccessReason, "")
	err := r.updateClusterStatus(ctx, appworker)
	return reconcile.Result{RequeueAfter: requeueAfter}, err
}
