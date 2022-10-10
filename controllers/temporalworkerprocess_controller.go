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
	"github.com/alexandrevilain/temporal-operator/pkg/resource"
	"github.com/alexandrevilain/temporal-operator/pkg/workerprocess"
)

// TemporalWorkerProcessReconciler reconciles a TemporalWorkerProcess object
type TemporalWorkerProcessReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=temporal.io,resources=temporalworkerprocesses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=temporal.io,resources=temporalworkerprocesses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=temporal.io,resources=temporalworkerprocesses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TemporalWorkerProcess object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *TemporalWorkerProcessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Starting reconciliation")

	worker := &v1beta1.TemporalWorkerProcess{}
	err := r.Get(ctx, req.NamespacedName, worker)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Check if the resource has been marked for deletion
	if !worker.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Deleting temporal worker process", "name", worker.Name)
		return reconcile.Result{}, nil
	}

	if err := r.reconcileResources(ctx, worker); err != nil {
		logger.Error(err, "Can't reconcile resources")
		return r.handleErrorWithRequeue(ctx, worker, v1beta1.ResourcesReconciliationFailedReason, err, 2*time.Second)
	}

	return r.handleSuccess(ctx, worker)
}

// SetupWithManager sets up the controller with the Manager.
func (r *TemporalWorkerProcessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&temporaliov1beta1.TemporalWorkerProcess{}).
		Complete(r)
}

func (r *TemporalWorkerProcessReconciler) handleErrorWithRequeue(ctx context.Context, worker *v1beta1.TemporalWorkerProcess, reason string, err error, requeueAfter time.Duration) (ctrl.Result, error) {
	if reason == "" {
		reason = v1beta1.ReconcileErrorReason
	}
	v1beta1.SetTemporalWorkerProcessReconcileError(worker, metav1.ConditionTrue, reason, err.Error())
	err = r.updateClusterStatus(ctx, worker)
	return reconcile.Result{RequeueAfter: requeueAfter}, err
}

func (r *TemporalWorkerProcessReconciler) handleError(ctx context.Context, worker *v1beta1.TemporalWorkerProcess, reason string, err error) (ctrl.Result, error) {
	return r.handleErrorWithRequeue(ctx, worker, reason, err, 0)
}

func (r *TemporalWorkerProcessReconciler) updateClusterStatus(ctx context.Context, cluster *v1beta1.TemporalWorkerProcess) error {
	err := r.Status().Update(ctx, cluster)
	if err != nil {
		return err
	}
	return nil
}

func (r *TemporalWorkerProcessReconciler) reconcileResources(ctx context.Context, temporalWorkerProcess *v1beta1.TemporalWorkerProcess) error {
	logger := log.FromContext(ctx)

	workerProcessBuilder := workerprocess.ClusterBuilder{
		Instance: temporalWorkerProcess,
		Scheme:   r.Scheme,
	}

	builders, err := workerProcessBuilder.ResourceBuilders()
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
		r.logAndRecordOperationResult(ctx, temporalWorkerProcess, res, operationResult, err)
		if err != nil {
			return err
		}
	}

	return r.updateClusterStatus(ctx, temporalWorkerProcess)
}

func (r *TemporalWorkerProcessReconciler) logAndRecordOperationResult(ctx context.Context, worker *v1beta1.TemporalWorkerProcess, resource runtime.Object, operationResult controllerutil.OperationResult, err error) {
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
		r.Recorder.Event(worker, corev1.EventTypeNormal, reason, msg)
	}

	if err != nil {
		msg := fmt.Sprintf("failed to %s resource %s of Type %T", action, resource.(metav1.Object).GetName(), resource.(metav1.Object))
		reason := fmt.Sprintf("%sError", reason)
		logger.Error(err, msg)
		r.Recorder.Event(worker, corev1.EventTypeWarning, reason, msg)
	}
}

func (r *TemporalWorkerProcessReconciler) handleSuccess(ctx context.Context, worker *v1beta1.TemporalWorkerProcess) (ctrl.Result, error) {
	return r.handleSuccessWithRequeue(ctx, worker, 0)
}

func (r *TemporalWorkerProcessReconciler) handleSuccessWithRequeue(ctx context.Context, worker *v1beta1.TemporalWorkerProcess, requeueAfter time.Duration) (ctrl.Result, error) {
	v1beta1.SetTemporalWorkerProcessReconcileSuccess(worker, metav1.ConditionTrue, v1beta1.ReconcileSuccessReason, "")
	err := r.updateClusterStatus(ctx, worker)
	return reconcile.Result{RequeueAfter: requeueAfter}, err
}
