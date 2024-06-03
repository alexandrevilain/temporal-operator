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
	"errors"
	"fmt"
	"time"

	"github.com/alexandrevilain/controller-tools/pkg/patch"
	"go.temporal.io/api/serviceerror"
	temporalclient "go.temporal.io/sdk/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/temporal"
)

// TemporalScheduleReconciler reconciles a Schedule object.
type TemporalScheduleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=temporal.io,resources=temporalschedules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=temporal.io,resources=temporalschedules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=temporal.io,resources=temporalschedules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *TemporalScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	logger := log.FromContext(ctx)

	logger.Info("Starting reconciliation")

	schedule := &v1beta1.TemporalSchedule{}
	err := r.Get(ctx, req.NamespacedName, schedule)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	patchHelper, err := patch.NewHelper(schedule, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}

	defer func() {
		// Always attempt to Patch the Cluster object and status after each reconciliation.
		err := patchHelper.Patch(ctx, schedule)
		if err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	namespace := &v1beta1.TemporalNamespace{}
	err = r.Get(ctx, schedule.Spec.NamespaceRef.NamespacedName(schedule), namespace)
	if err != nil {
		if apierrors.IsNotFound(err) && !schedule.ObjectMeta.DeletionTimestamp.IsZero() {
			logger.Info("Namespace not found deleting schedule", "namespace", schedule.Spec.NamespaceRef.NamespacedName(schedule))
			// Two ways to get here:
			//  - TemporalNamespace has not been created yet. In this case, if the TemporalSchedule is deleted, no point in waiting for the TemporalNamespace to be healthy.
			//  - TemporalNamespace existed at some point, but now is deleted. In this case, the underlying schedule in the Temporal server is already gone.
			controllerutil.RemoveFinalizer(schedule, deletionFinalizer)
			return reconcile.Result{}, nil
		}

		return r.handleError(ctx, schedule, v1beta1.ReconcileErrorReason, "Namespace lookup", err)
	}

	if !namespace.IsReady() {
		logger.Info("Skipping schedule reconciliation until referenced namespace is ready")

		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	}

	cluster := &v1beta1.TemporalCluster{}
	err = r.Get(ctx, namespace.Spec.ClusterRef.NamespacedName(schedule), cluster)
	if err != nil {
		if apierrors.IsNotFound(err) && !schedule.ObjectMeta.DeletionTimestamp.IsZero() {
			logger.Info("Cluster not found deleting schedule", "cluster", namespace.Spec.ClusterRef.NamespacedName(schedule))
			// Two ways to get here:
			//  - TemporalCluster has not been created yet. In this case, if the TemporalSchedule is deleted, no point in waiting for the TemporalCluster to be healthy.
			//  - TemporalCluster existed at some point, but now is deleted. In this case, the underlying schedule in the Temporal server is already gone.
			controllerutil.RemoveFinalizer(schedule, deletionFinalizer)
			return reconcile.Result{}, nil
		}

		return r.handleError(ctx, schedule, v1beta1.ReconcileErrorReason, "Cluster lookup", err)
	}

	if !cluster.IsReady() {
		logger.Info("Skipping schedule reconciliation until referenced cluster is ready")

		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	}

	clientOpts := func(opt *temporalclient.Options) {
		opt.Namespace = schedule.Spec.NamespaceRef.Name
	}
	client, err := temporal.GetClusterClient(ctx, r.Client, cluster, clientOpts)
	if err != nil {
		err = fmt.Errorf("can't create cluster client: %w", err)
		return r.handleError(ctx, schedule, v1beta1.ReconcileErrorReason, "Creating cluster client", err)
	}
	defer client.Close()

	// Check if the resource has been marked for deletion
	if !schedule.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Deleting schedule")

		err := r.ensureScheduleDeleted(ctx, schedule, &client)
		if err != nil {
			return r.handleError(ctx, schedule, v1beta1.ReconcileErrorReason, "Deleting schedule", err)
		}
		return reconcile.Result{}, nil
	}

	// Ensure the schedule have a deletion marker if the AllowDeletion is set to true.
	r.ensureFinalizer(schedule)

	request, err := temporal.ScheduleToCreateScheduleRequest(cluster, schedule)
	if err != nil {
		return r.handleError(ctx, schedule, v1beta1.ReconcileErrorReason, "Constructing create schedule request", err)
	}

	_, err = client.WorkflowService().CreateSchedule(ctx, request)
	if err != nil {
		var scheduleAlreadyExistsError *serviceerror.WorkflowExecutionAlreadyStarted
		ok := errors.As(err, &scheduleAlreadyExistsError)
		if !ok {
			err = fmt.Errorf("can't create \"%s\" schedule: %w", schedule.GetName(), err)
			return r.handleError(ctx, schedule, v1beta1.ReconcileErrorReason, "Creating schedule", err)
		}

		request, err := temporal.ScheduleToUpdateScheduleRequest(schedule)
		if err != nil {
			return r.handleError(ctx, schedule, v1beta1.ReconcileErrorReason, "Constructing update schedule request", err)
		}

		_, err = client.WorkflowService().UpdateSchedule(ctx, request)
		if err != nil {
			return r.handleError(ctx, schedule, v1beta1.ReconcileErrorReason, "Updating schedule", err)
		}
	}

	logger.Info("Successfully reconciled schedule", "schedule", schedule.GetName())

	v1beta1.SetTemporalScheduleReady(schedule, metav1.ConditionTrue, v1beta1.TemporalScheduleCreatedReason, "Schedule successfully created")

	return r.handleSuccess(schedule)
}

// ensureFinalizer ensures the deletion finalizer is set on the object if the user allowed schedule deletion using the CRD.
func (r *TemporalScheduleReconciler) ensureFinalizer(schedule *v1beta1.TemporalSchedule) {
	if schedule.ObjectMeta.DeletionTimestamp.IsZero() {
		if schedule.Spec.AllowDeletion {
			_ = controllerutil.AddFinalizer(schedule, deletionFinalizer)
		} else {
			_ = controllerutil.RemoveFinalizer(schedule, deletionFinalizer)
		}
	}
}

func (r *TemporalScheduleReconciler) ensureScheduleDeleted(ctx context.Context, schedule *v1beta1.TemporalSchedule, client *temporalclient.Client) error {
	logger := log.FromContext(ctx)

	if !controllerutil.ContainsFinalizer(schedule, deletionFinalizer) {
		return nil
	}

	_, err := (*client).WorkflowService().DeleteSchedule(ctx, temporal.ScheduleToDeleteScheduleRequest(schedule))
	if err != nil {
		println(fmt.Sprintf("%T: %+v", err, err))
		var scheduleNotFoundError *serviceerror.NotFound
		if errors.As(err, &scheduleNotFoundError) {
			logger.Info("try to delete but not found", "schedule", schedule.GetName())
		} else {
			return fmt.Errorf("can't delete \"%s\" schedule: %w", schedule.GetName(), err)
		}
	}

	_ = controllerutil.RemoveFinalizer(schedule, deletionFinalizer)
	return nil
}

func (r *TemporalScheduleReconciler) handleSuccess(schedule *v1beta1.TemporalSchedule) (ctrl.Result, error) {
	return r.handleSuccessWithRequeue(schedule, 0)
}

func (r *TemporalScheduleReconciler) handleError(ctx context.Context, schedule *v1beta1.TemporalSchedule, reason string, action string, err error) (ctrl.Result, error) { //nolint:unparam
	logger := log.FromContext(ctx)

	logger.Error(err, action)
	return r.handleErrorWithRequeue(schedule, reason, err, 0)
}

func (r *TemporalScheduleReconciler) handleSuccessWithRequeue(schedule *v1beta1.TemporalSchedule, requeueAfter time.Duration) (ctrl.Result, error) {
	v1beta1.SetTemporalScheduleReconcileSuccess(schedule, metav1.ConditionTrue, v1beta1.ReconcileSuccessReason, "")
	return reconcile.Result{RequeueAfter: requeueAfter}, nil
}

func (r *TemporalScheduleReconciler) handleErrorWithRequeue(schedule *v1beta1.TemporalSchedule, reason string, err error, requeueAfter time.Duration) (ctrl.Result, error) {
	if reason == "" {
		reason = v1beta1.ReconcileErrorReason
	}
	v1beta1.SetTemporalScheduleReconcileError(schedule, metav1.ConditionTrue, reason, err.Error())
	return reconcile.Result{RequeueAfter: requeueAfter}, nil
}

func (r *TemporalScheduleReconciler) namespaceToSchedulesMapfunc(ctx context.Context, o client.Object) []reconcile.Request {
	namespace, ok := o.(*v1beta1.TemporalNamespace)
	if !ok {
		return nil
	}

	temporalSchedules := &v1beta1.TemporalScheduleList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(namespaceRefField, namespace.GetName()),
	}

	err := r.Client.List(ctx, temporalSchedules, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	result := []reconcile.Request{}
	for _, schedule := range temporalSchedules.Items {
		schedule := schedule
		// As we're only indexing on spec.namespaceRef.Name, ensure that referenced schedule is watching the namespace's schedule.
		if schedule.Spec.NamespaceRef.NamespacedName(&schedule) != client.ObjectKeyFromObject(namespace) {
			continue
		}
		result = append(result, reconcile.Request{
			NamespacedName: client.ObjectKeyFromObject(&schedule),
		})
	}

	return result
}

// SetupWithManager sets up the controller with the Manager.
func (r *TemporalScheduleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1beta1.TemporalSchedule{}, clusterRefField, func(rawObj client.Object) []string {
		temporalSchedule := rawObj.(*v1beta1.TemporalSchedule)
		if temporalSchedule.Spec.NamespaceRef.Name == "" {
			return nil
		}
		return []string{temporalSchedule.Spec.NamespaceRef.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.TemporalSchedule{}, builder.WithPredicates(predicate.Or(
			predicate.GenerationChangedPredicate{},
			predicate.LabelChangedPredicate{},
			predicate.AnnotationChangedPredicate{},
		))).
		Watches(
			&v1beta1.TemporalNamespace{},
			handler.EnqueueRequestsFromMapFunc(r.namespaceToSchedulesMapfunc),
		).
		Complete(r)
}
