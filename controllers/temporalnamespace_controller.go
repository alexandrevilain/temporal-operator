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
	"strings"
	"time"

	"github.com/alexandrevilain/controller-tools/pkg/patch"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/serviceerror"
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

// TemporalNamespaceReconciler reconciles a Namespace object.
type TemporalNamespaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=temporal.io,resources=temporalnamespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=temporal.io,resources=temporalnamespaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=temporal.io,resources=temporalnamespaces/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *TemporalNamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	logger := log.FromContext(ctx)

	logger.Info("Starting reconciliation")

	namespace := &v1beta1.TemporalNamespace{}
	err := r.Get(ctx, req.NamespacedName, namespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	patchHelper, err := patch.NewHelper(namespace, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}

	defer func() {
		// Always attempt to Patch the Cluster object and status after each reconciliation.
		err := patchHelper.Patch(ctx, namespace)
		if err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	cluster := &v1beta1.TemporalCluster{}
	err = r.Get(ctx, namespace.Spec.ClusterRef.NamespacedName(namespace), cluster)
	if err != nil {
		if apierrors.IsNotFound(err) && !namespace.ObjectMeta.DeletionTimestamp.IsZero() {
			// Two ways to get here:
			//  - TemporalCluster has not been created yet. In this case, if the TemporalNamespace is deleted, no point in waiting for the TemporalCluster to be healthy.
			//  - TemporalCluster existed at some point, but now is deleted. In this case, the underlying namespace in the Temporal server is already gone.
			controllerutil.RemoveFinalizer(namespace, deletionFinalizer)
			return reconcile.Result{}, nil
		}
		return r.handleError(namespace, v1beta1.ReconcileErrorReason, err)
	}

	if !cluster.IsReady() {
		logger.Info("Skipping namespace reconciliation until referenced cluster is ready")

		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Check if the resource has been marked for deletion
	if !namespace.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Deleting namespace")

		err := r.ensureNamespaceDeleted(ctx, namespace, cluster)
		if err != nil {
			return r.handleError(namespace, v1beta1.ReconcileErrorReason, err)
		}
		return reconcile.Result{}, nil
	}

	// Ensure the namespace have a deletion marker if the AllowDeletion is set to true.
	r.ensureFinalizer(namespace)

	client, err := temporal.GetClusterNamespaceClient(ctx, r.Client, cluster)
	if err != nil {
		err = fmt.Errorf("can't create cluster namespace client: %w", err)
		return r.handleError(namespace, v1beta1.ReconcileErrorReason, err)
	}
	defer client.Close()

	err = client.Register(ctx, temporal.NamespaceToRegisterNamespaceRequest(cluster, namespace))
	if err != nil {
		var namespaceAlreadyExistsError *serviceerror.NamespaceAlreadyExists
		ok := errors.As(err, &namespaceAlreadyExistsError)
		if !ok {
			err = fmt.Errorf("can't create \"%s\" namespace: %w", namespace.GetName(), err)
			return r.handleError(namespace, v1beta1.ReconcileErrorReason, err)
		}
		err = client.Update(ctx, temporal.NamespaceToUpdateNamespaceRequest(cluster, namespace))
		if err != nil {
			return r.handleError(namespace, v1beta1.ReconcileErrorReason, err)
		}
	}

	err = r.reconcileCustomSearchAttributes(ctx, namespace, cluster)
	if err != nil {
		return r.handleError(namespace, v1beta1.ReconcileErrorReason, err)
	}

	logger.Info("Successfully reconciled namespace", "namespace", namespace.GetName())

	v1beta1.SetTemporalNamespaceReady(namespace, metav1.ConditionTrue, v1beta1.TemporalNamespaceCreatedReason, "Namespace successfully created")

	return r.handleSuccess(namespace)
}

// reconcileCustomSearchAttributes ensures that the custom search attributes on the Temporal server exactly match those defined in the spec
func (r *TemporalNamespaceReconciler) reconcileCustomSearchAttributes(ctx context.Context, namespace *v1beta1.TemporalNamespace, cluster *v1beta1.TemporalCluster) error {
	ns := namespace.GetName()

	client, err := temporal.GetClusterClient(ctx, r.Client, cluster)
	if err != nil {
		return err
	}

	// List search attributes currently on the Temporal server
	listReq := &operatorservice.ListSearchAttributesRequest{Namespace: ns}
	searchAttributesOnServer, err := client.OperatorService().ListSearchAttributes(ctx, listReq)
	if err != nil {
		return err
	}
	// just focus on the CUSTOM search attributes.
	customSearchAttributesOnServer := &searchAttributesOnServer.CustomAttributes // a pointer avoids unecessary copying for the sake of just a named variable

	// Note that the CustomSearchAttributes map data structure that is built using the Spec merely maps string->string.
	// To rigorously compare search attributes between the spec and the Temporal server, the types need to be consistent.
	// Therefore, we need to construct a string->enums.IndexedValueType map from the string->string map.
	customSearchAttributesInSpec := make(map[string]enums.IndexedValueType, len(namespace.Spec.CustomSearchAttributes))
	for searchAttrNameString, searchAttrTypeString := range namespace.Spec.CustomSearchAttributes {
		indexedValueType, err := searchAttributeTypeStringToEnum(searchAttrTypeString)
		if err != nil {
			return fmt.Errorf("unable to parse search attribute type %s: %w", searchAttrTypeString, err)
		}
		customSearchAttributesInSpec[searchAttrNameString] = indexedValueType
	}

	// Remove those custom search attributes from the Temporal server whose name does not exist in the Spec,
	// or whose name exists in the Spec but whose type doesn't match the type in the Spec.
	customSearchAttributesToRemove := make([]string, 0)
	for serverSearchAttrName, serverSearchAttrType := range *customSearchAttributesOnServer {
		specSearchAttrType, serverSearchAttrNameExistsInSpec := customSearchAttributesInSpec[serverSearchAttrName]
		if !serverSearchAttrNameExistsInSpec || serverSearchAttrType != specSearchAttrType {
			customSearchAttributesToRemove = append(customSearchAttributesToRemove, serverSearchAttrName)
		}
	}
	removeReq := &operatorservice.RemoveSearchAttributesRequest{
		Namespace:        ns,
		SearchAttributes: customSearchAttributesToRemove,
	}
	_, err = client.OperatorService().RemoveSearchAttributes(ctx, removeReq)
	if err != nil {
		return fmt.Errorf("unable to remove search attributes: %w", err)
	}

	// Create custom search attributes from the Spec which don't yet exist on the Temporal server.
	// If the Temporal server already has a custom search attribute with the same name but a different type,
	// then return an error.
	customSearchAttributesToCreate := make(map[string]enums.IndexedValueType)
	for specSearchAttrName, specSearchAttrType := range customSearchAttributesInSpec {
		serverSearchAttrType, specSearchAttrNameExistsOnServer := (*customSearchAttributesOnServer)[specSearchAttrName]
		if specSearchAttrNameExistsOnServer {
			if specSearchAttrType != serverSearchAttrType {
				return fmt.Errorf("search attribute %s already exists and has different type %s", specSearchAttrName, serverSearchAttrType.String())
			}
		} else {
			customSearchAttributesToCreate[specSearchAttrName] = specSearchAttrType
		}
	}
	createReq := &operatorservice.AddSearchAttributesRequest{
		Namespace:        ns,
		SearchAttributes: customSearchAttributesToCreate,
	}
	_, err = client.OperatorService().AddSearchAttributes(ctx, createReq)
	if err != nil {
		return fmt.Errorf("unable to add search attributes: %w", err)
	}


	return nil
}

// searchAttributeTypeStringToEnum returns the IndexedValueType for a given string or an error.
func searchAttributeTypeStringToEnum(search string) (enums.IndexedValueType, error) {
	for k, v := range enums.IndexedValueType_shorthandValue {
		if strings.EqualFold(search, k) {
			return enums.IndexedValueType(v), nil
		}
	}
	return enums.INDEXED_VALUE_TYPE_UNSPECIFIED, fmt.Errorf("unsupported search attribute type: %v", search)
}

// ensureFinalizer ensures the deletion finalizer is set on the object if the user allowed namespace deletion using the CRD.
func (r *TemporalNamespaceReconciler) ensureFinalizer(namespace *v1beta1.TemporalNamespace) {
	if namespace.ObjectMeta.DeletionTimestamp.IsZero() && namespace.Spec.AllowDeletion {
		_ = controllerutil.AddFinalizer(namespace, deletionFinalizer)
	}
}

func (r *TemporalNamespaceReconciler) ensureNamespaceDeleted(ctx context.Context, namespace *v1beta1.TemporalNamespace, cluster *v1beta1.TemporalCluster) error {
	logger := log.FromContext(ctx)

	if !controllerutil.ContainsFinalizer(namespace, deletionFinalizer) {
		return nil
	}

	client, err := temporal.GetClusterClient(ctx, r.Client, cluster)
	if err != nil {
		return fmt.Errorf("can't create cluster client: %w", err)
	}
	defer client.Close()

	_, err = client.OperatorService().DeleteNamespace(ctx, temporal.NamespaceToDeleteNamespaceRequest(namespace))
	if err != nil {
		var namespaceNotFoundError *serviceerror.NamespaceNotFound
		if errors.As(err, &namespaceNotFoundError) {
			logger.Info("try to delete but not found", "namespace", namespace.GetName())
		} else {
			return fmt.Errorf("can't delete \"%s\" namespace: %w", namespace.GetName(), err)
		}
	}

	_ = controllerutil.RemoveFinalizer(namespace, deletionFinalizer)
	return nil
}

func (r *TemporalNamespaceReconciler) handleSuccess(namespace *v1beta1.TemporalNamespace) (ctrl.Result, error) {
	return r.handleSuccessWithRequeue(namespace, 0)
}

func (r *TemporalNamespaceReconciler) handleError(namespace *v1beta1.TemporalNamespace, reason string, err error) (ctrl.Result, error) { //nolint:unparam
	return r.handleErrorWithRequeue(namespace, reason, err, 0)
}

func (r *TemporalNamespaceReconciler) handleSuccessWithRequeue(namespace *v1beta1.TemporalNamespace, requeueAfter time.Duration) (ctrl.Result, error) {
	v1beta1.SetTemporalNamespaceReconcileSuccess(namespace, metav1.ConditionTrue, v1beta1.ReconcileSuccessReason, "")
	return reconcile.Result{RequeueAfter: requeueAfter}, nil
}

func (r *TemporalNamespaceReconciler) handleErrorWithRequeue(namespace *v1beta1.TemporalNamespace, reason string, err error, requeueAfter time.Duration) (ctrl.Result, error) {
	if reason == "" {
		reason = v1beta1.ReconcileErrorReason
	}
	v1beta1.SetTemporalNamespaceReconcileError(namespace, metav1.ConditionTrue, reason, err.Error())
	return reconcile.Result{RequeueAfter: requeueAfter}, nil
}

func (r *TemporalNamespaceReconciler) clusterToNamespacesMapfunc(ctx context.Context, o client.Object) []reconcile.Request {
	cluster, ok := o.(*v1beta1.TemporalCluster)
	if !ok {
		return nil
	}

	temporalNamespaces := &v1beta1.TemporalNamespaceList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(clusterRefField, cluster.GetName()),
	}

	err := r.Client.List(ctx, temporalNamespaces, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	result := []reconcile.Request{}
	for _, namespace := range temporalNamespaces.Items {
		namespace := namespace
		// As we're only indexing on spec.clusterRef.Name, ensure that referenced namespace is watching the cluster's namespace.
		if namespace.Spec.ClusterRef.NamespacedName(&namespace) != client.ObjectKeyFromObject(cluster) {
			continue
		}
		result = append(result, reconcile.Request{
			NamespacedName: client.ObjectKeyFromObject(&namespace),
		})
	}

	return result
}

// SetupWithManager sets up the controller with the Manager.
func (r *TemporalNamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1beta1.TemporalNamespace{}, clusterRefField, func(rawObj client.Object) []string {
		temporalNamespace := rawObj.(*v1beta1.TemporalNamespace)
		if temporalNamespace.Spec.ClusterRef.Name == "" {
			return nil
		}
		return []string{temporalNamespace.Spec.ClusterRef.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.TemporalNamespace{}, builder.WithPredicates(predicate.Or(
			predicate.GenerationChangedPredicate{},
			predicate.LabelChangedPredicate{},
			predicate.AnnotationChangedPredicate{},
		))).
		Watches(
			&v1beta1.TemporalCluster{},
			handler.EnqueueRequestsFromMapFunc(r.clusterToNamespacesMapfunc),
		).
		Complete(r)
}
