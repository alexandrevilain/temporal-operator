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

	"github.com/alexandrevilain/temporal-operator/pkg/kubernetes/patch"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/reconciler"
	"github.com/alexandrevilain/temporal-operator/pkg/resource"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/workerprocessbuilder"
	"github.com/alexandrevilain/temporal-operator/pkg/resourceset"
	"github.com/alexandrevilain/temporal-operator/pkg/status"
)

// TemporalWorkerProcessReconciler reconciles a TemporalWorkerProcess object
type TemporalWorkerProcessReconciler struct {
	reconciler.Base
}

//+kubebuilder:rbac:groups=temporal.io,resources=temporalworkerprocesses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=temporal.io,resources=temporalworkerprocesses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=temporal.io,resources=temporalworkerprocesses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *TemporalWorkerProcessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
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
		logger.Info("Deleting worker process", "name", worker.Name)
		return reconcile.Result{}, nil
	}

	patchHelper, err := patch.NewHelper(worker, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}

	defer func() {
		// Always attempt to Patch the WorkerProcess object and status after each reconciliation.
		err := patchHelper.Patch(ctx, worker)
		if err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	// Experimental: ensure the workerprocess image is built by the temporal worker process builder.
	if requeueAfter, err := r.reconcileBuilder(ctx, worker); err != nil || requeueAfter > 0 {
		if err != nil {
			logger.Error(err, "Can't reconcile builder")
			if requeueAfter == 0 {
				requeueAfter = 2 * time.Second
			}
			return r.handleErrorWithRequeue(ctx, worker, v1beta1.ResourcesReconciliationFailedReason, err, requeueAfter)
		}
		if requeueAfter > 0 {
			return reconcile.Result{RequeueAfter: requeueAfter}, nil
		}
	}

	cluster := &v1beta1.TemporalCluster{}
	err = r.Get(ctx, worker.Spec.ClusterRef.NamespacedName(worker), cluster)
	if err != nil {
		logger.Error(err, "Can't find referenced temporal cluster")
		return r.handleError(ctx, worker, v1beta1.ReconcileErrorReason, err)
	}

	if requeueAfter, err := r.reconcileResources(ctx, worker, cluster); err != nil || requeueAfter > 0 {
		if err != nil {
			logger.Error(err, "Can't reconcile resources")
			if requeueAfter == 0 {
				requeueAfter = 2 * time.Second
			}
			return r.handleErrorWithRequeue(ctx, worker, v1beta1.ResourcesReconciliationFailedReason, err, requeueAfter)
		}
		if requeueAfter > 0 {
			return reconcile.Result{RequeueAfter: requeueAfter}, nil
		}
	}

	return r.handleSuccess(ctx, worker)
}

// Reconcile worker process builder configmaps.
func (r *TemporalWorkerProcessReconciler) reconcileWorkerScriptsConfigmap(ctx context.Context, worker *v1beta1.TemporalWorkerProcess) (time.Duration, error) {
	builders := []resource.Builder{
		workerprocessbuilder.NewJobScriptsConfigmapBuilder(worker, r.Scheme),
	}
	_, requeueAfter, err := r.ReconcileBuilders(ctx, worker, builders)
	return requeueAfter, err
}

func (r *TemporalWorkerProcessReconciler) reconcileBuilder(ctx context.Context, worker *v1beta1.TemporalWorkerProcess) (time.Duration, error) {
	if !worker.Spec.Builder.BuilderEnabled() {
		return 0, nil
	}

	var defaultBuildAttempt int32 = 1
	if worker.Status.BuildAttempt == nil {
		worker.Status.BuildAttempt = &defaultBuildAttempt
	}

	// First of all, ensure the configmap containing scripts is up-to-date
	if requeueAfter, err := r.reconcileWorkerScriptsConfigmap(ctx, worker); err != nil || requeueAfter > 0 {
		return requeueAfter, err
	}

	jobs := []*reconciler.Job{
		{
			Name:    "build-worker-process",
			Command: []string{"/etc/scripts/build-worker-process.sh"},
			Skip: func(owner runtime.Object) bool {

				if owner.(*v1beta1.TemporalWorkerProcess).Status.Version != owner.(*v1beta1.TemporalWorkerProcess).Spec.Version {
					return false
				}

				if pointer.Int32(*owner.(*v1beta1.TemporalWorkerProcess).Spec.Builder.BuildAttempt) != pointer.Int32(*owner.(*v1beta1.TemporalWorkerProcess).Status.BuildAttempt) {
					return false
				}

				return owner.(*v1beta1.TemporalWorkerProcess).Status.Created
			},
			ReportSuccess: func(owner runtime.Object) error {
				owner.(*v1beta1.TemporalWorkerProcess).Status.Created = true
				return nil
			},
		},
	}

	factory := func(owner runtime.Object, scheme *runtime.Scheme, name string, command []string) resource.Builder {
		worker := owner.(*v1beta1.TemporalWorkerProcess)
		return workerprocessbuilder.NewJobBuilder(worker, scheme, name, command)
	}

	if requeueAfter, err := r.ReconcileJobs(ctx, worker, factory, jobs); err != nil || requeueAfter > 0 {
		return requeueAfter, err
	}

	worker.Status.BuildAttempt = worker.Spec.Builder.BuildAttempt

	return 0, nil
}

func (r *TemporalWorkerProcessReconciler) reconcileResources(ctx context.Context, temporalWorkerProcess *v1beta1.TemporalWorkerProcess, temporalCluster *v1beta1.TemporalCluster) (time.Duration, error) {
	workerProcessBuilder := &resourceset.WorkerProcessBuilder{
		Instance: temporalWorkerProcess,
		Cluster:  temporalCluster,
		Scheme:   r.Scheme,
	}

	statuses, requeueAfter, err := r.ReconcileResources(ctx, temporalCluster, workerProcessBuilder)
	if err != nil {
		return requeueAfter, err
	}
	if requeueAfter > 0 {
		return requeueAfter, nil
	}

	if len(statuses) == 1 {
		temporalWorkerProcess.Status.Ready = statuses[0].Ready
	} else {
		temporalWorkerProcess.Status.Ready = false
	}

	if status.IsWorkerProcessReady(temporalWorkerProcess) {
		v1beta1.SetTemporalWorkerProcessReady(temporalWorkerProcess, metav1.ConditionTrue, v1beta1.ServicesReadyReason, "")
		temporalWorkerProcess.Status.Version = temporalWorkerProcess.Spec.Version
	} else {
		v1beta1.SetTemporalWorkerProcessReady(temporalWorkerProcess, metav1.ConditionFalse, v1beta1.ServicesNotReadyReason, "")
	}

	return 0, nil
}

func (r *TemporalWorkerProcessReconciler) handleErrorWithRequeue(ctx context.Context, worker *v1beta1.TemporalWorkerProcess, reason string, err error, requeueAfter time.Duration) (ctrl.Result, error) {
	if reason == "" {
		reason = v1beta1.ReconcileErrorReason
	}
	v1beta1.SetTemporalWorkerProcessReconcileError(worker, metav1.ConditionTrue, reason, err.Error())
	return reconcile.Result{RequeueAfter: requeueAfter}, err
}

func (r *TemporalWorkerProcessReconciler) handleError(ctx context.Context, worker *v1beta1.TemporalWorkerProcess, reason string, err error) (ctrl.Result, error) {
	return r.handleErrorWithRequeue(ctx, worker, reason, err, 0)
}

func (r *TemporalWorkerProcessReconciler) handleSuccess(ctx context.Context, worker *v1beta1.TemporalWorkerProcess) (ctrl.Result, error) {
	return r.handleSuccessWithRequeue(ctx, worker, 0)
}

func (r *TemporalWorkerProcessReconciler) handleSuccessWithRequeue(ctx context.Context, worker *v1beta1.TemporalWorkerProcess, requeueAfter time.Duration) (ctrl.Result, error) {
	v1beta1.SetTemporalWorkerProcessReconcileSuccess(worker, metav1.ConditionTrue, v1beta1.ReconcileSuccessReason, "")
	return reconcile.Result{RequeueAfter: requeueAfter}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TemporalWorkerProcessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	controller := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.TemporalWorkerProcess{}, builder.WithPredicates(predicate.Or(
			predicate.GenerationChangedPredicate{},
			predicate.LabelChangedPredicate{},
			predicate.AnnotationChangedPredicate{},
		))).
		Owns(&appsv1.Deployment{})

	return controller.Complete(r)
}
