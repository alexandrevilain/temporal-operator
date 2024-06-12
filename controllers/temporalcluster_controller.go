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

	"go.temporal.io/server/common/primitives"
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/discovery"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/alexandrevilain/controller-tools/pkg/hash"
	"github.com/alexandrevilain/controller-tools/pkg/patch"
	"github.com/alexandrevilain/controller-tools/pkg/resource"
	"github.com/alexandrevilain/temporal-operator/internal/resource/admintools"
	"github.com/alexandrevilain/temporal-operator/internal/resource/base"
	"github.com/alexandrevilain/temporal-operator/internal/resource/config"
	"github.com/alexandrevilain/temporal-operator/internal/resource/mtls/certmanager"
	"github.com/alexandrevilain/temporal-operator/internal/resource/mtls/istio"
	"github.com/alexandrevilain/temporal-operator/internal/resource/prometheus"
	"github.com/alexandrevilain/temporal-operator/internal/resource/ui"
	"github.com/alexandrevilain/temporal-operator/pkg/status"
)

const (
	ownerKey = ".metadata.controller"
)

// TemporalClusterReconciler reconciles a Cluster object.
type TemporalClusterReconciler struct {
	Base

	AvailableAPIs *discovery.AvailableAPIs
}

//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=get;create;patch
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="cert-manager.io",resources=certificates;issuers,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="security.istio.io",resources=peerauthentications,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="networking.istio.io",resources=destinationrules,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="monitoring.coreos.com",resources=servicemonitors,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=temporal.io,resources=temporalclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=temporal.io,resources=temporalclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=temporal.io,resources=temporalclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *TemporalClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	logger := log.FromContext(ctx)

	logger.Info("Starting reconciliation")

	cluster := &v1beta1.TemporalCluster{}
	err := r.Get(ctx, req.NamespacedName, cluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Check if the resource has been marked for deletion
	if !cluster.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("Deleting temporal cluster", "name", cluster.Name)
		return reconcile.Result{}, nil
	}

	patchHelper, err := patch.NewHelper(cluster, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}

	defer func() {
		// Always attempt to Patch the Cluster object and status after each reconciliation.
		err := patchHelper.Patch(ctx, cluster)
		if err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	// Check the ready condition
	cond, exists := v1beta1.GetTemporalClusterReadyCondition(cluster)
	if !exists || cond.ObservedGeneration != cluster.GetGeneration() {
		v1beta1.SetTemporalClusterReady(cluster, metav1.ConditionUnknown, v1beta1.ProgressingReason, "")
	}

	if requeueAfter, err := r.reconcilePersistence(ctx, cluster); err != nil || requeueAfter > 0 {
		if err != nil {
			logger.Error(err, "Can't reconcile persistence")
			if requeueAfter == 0 {
				requeueAfter = 2 * time.Second
			}
			return r.handleErrorWithRequeue(cluster, v1beta1.PersistenceReconciliationFailedReason, err, requeueAfter)
		}
		if requeueAfter > 0 {
			return reconcile.Result{RequeueAfter: requeueAfter}, nil
		}
	}

	if err := r.reconcileResources(ctx, cluster); err != nil {
		logger.Error(err, "Can't reconcile resources")
		return r.handleErrorWithRequeue(cluster, v1beta1.ResourcesReconciliationFailedReason, err, 2*time.Second)
	}

	return r.handleSuccess(cluster)
}

func (r *TemporalClusterReconciler) reconcileResources(ctx context.Context, temporalCluster *v1beta1.TemporalCluster) error {
	// reconcile configmap first, then compute its hash.
	configMapObject, err := r.Reconciler.ReconcileBuilder(ctx,
		temporalCluster,
		config.NewConfigmapBuilder(temporalCluster, r.Scheme))
	if err != nil {
		return fmt.Errorf("can't reconcile configmap: %w", err)
	}

	configMap, ok := configMapObject.(*corev1.ConfigMap)
	if !ok {
		return errors.New("can't cast configmap object to *corev1.ConfigMap")
	}

	configHash, err := hash.Sha256(configMap.Data)
	if err != nil {
		return fmt.Errorf("can't compute configmap hash: %w", err)
	}

	builders, err := r.resourceBuilders(temporalCluster, configHash)
	if err != nil {
		return err
	}

	objects, err := r.Reconciler.ReconcileBuilders(ctx, temporalCluster, builders)
	if err != nil {
		return err
	}

	statuses, err := status.ReconciledObjectsToServiceStatuses(temporalCluster, objects)
	if err != nil {
		return err
	}

	for _, status := range statuses {
		temporalCluster.Status.AddServiceStatus(status)
	}

	if status.ObservedVersionMatchesDesiredVersion(temporalCluster) {
		temporalCluster.Status.Version = temporalCluster.Spec.Version.String()
	}

	if status.IsClusterReady(temporalCluster) {
		v1beta1.SetTemporalClusterReady(temporalCluster, metav1.ConditionTrue, v1beta1.ServicesReadyReason, "")
	} else {
		v1beta1.SetTemporalClusterReady(temporalCluster, metav1.ConditionFalse, v1beta1.ServicesNotReadyReason, "")
	}

	return nil
}

func (r *TemporalClusterReconciler) resourceBuilders(temporalCluster *v1beta1.TemporalCluster, configHash string) ([]resource.Builder, error) {
	builders := []resource.Builder{
		base.NewFrontendServiceBuilder(temporalCluster, r.Scheme),
	}

	services := []primitives.ServiceName{
		primitives.FrontendService,
		primitives.HistoryService,
		primitives.MatchingService,
		primitives.WorkerService,
		primitives.InternalFrontendService,
	}

	for _, service := range services {
		specs, err := temporalCluster.Spec.Services.GetServiceSpec(service)
		if err != nil {
			return nil, err
		}

		serviceName := string(service)

		builders = append(builders, base.NewServiceAccountBuilder(serviceName, temporalCluster, r.Scheme))
		builders = append(builders, base.NewDeploymentBuilder(serviceName, temporalCluster, r.Scheme, specs, configHash))
		builders = append(builders, base.NewHeadlessServiceBuilder(serviceName, temporalCluster, r.Scheme, specs))

		builders = append(builders, istio.NewPeerAuthenticationBuilder(serviceName, temporalCluster, r.Scheme, specs))
		builders = append(builders, istio.NewDestinationRuleBuilder(serviceName, temporalCluster, r.Scheme, specs))
		builders = append(builders, prometheus.NewServiceMonitorBuilder(serviceName, temporalCluster, r.Scheme, specs))
	}

	builders = append(builders,
		base.NewDynamicConfigmapBuilder(temporalCluster, r.Scheme),
		// mTLS
		certmanager.NewMTLSBootstrapIssuerBuilder(temporalCluster, r.Scheme),
		certmanager.NewMTLSRootCACertificateBuilder(temporalCluster, r.Scheme),
		certmanager.NewMTLSRootCAIssuerBuilder(temporalCluster, r.Scheme),
		certmanager.NewMTLSInternodeIntermediateCACertificateBuilder(temporalCluster, r.Scheme),
		certmanager.NewMTLSInternodeIntermediateCAIssuerBuilder(temporalCluster, r.Scheme),
		certmanager.NewMTLSInternodeCertificateBuilder(temporalCluster, r.Scheme),
		certmanager.NewMTLSFrontendIntermediateCACertificateBuilder(temporalCluster, r.Scheme),
		certmanager.NewMTLSFrontendIntermediateCAIssuerBuilder(temporalCluster, r.Scheme),
		certmanager.NewMTLSFrontendCertificateBuilder(temporalCluster, r.Scheme),
		certmanager.NewWorkerFrontendClientCertificateBuilder(temporalCluster, r.Scheme),
		// UI:
		ui.NewDeploymentBuilder(temporalCluster, r.Scheme, configHash),
		ui.NewServiceBuilder(temporalCluster, r.Scheme),
		ui.NewIngressBuilder(temporalCluster, r.Scheme),
		ui.NewFrontendClientCertificateBuilder(temporalCluster, r.Scheme),
		// Admin tools:
		admintools.NewDeploymentBuilder(temporalCluster, r.Scheme, configHash),
		admintools.NewFrontendClientCertificateBuilder(temporalCluster, r.Scheme),
	)

	return builders, nil
}

func (r *TemporalClusterReconciler) handleSuccess(cluster *v1beta1.TemporalCluster) (ctrl.Result, error) {
	return r.handleSuccessWithRequeue(cluster, 0)
}

func (r *TemporalClusterReconciler) handleSuccessWithRequeue(cluster *v1beta1.TemporalCluster, requeueAfter time.Duration) (ctrl.Result, error) {
	v1beta1.SetTemporalClusterReconcileSuccess(cluster, metav1.ConditionTrue, v1beta1.ReconcileSuccessReason, "")
	return reconcile.Result{RequeueAfter: requeueAfter}, nil
}

func (r *TemporalClusterReconciler) handleErrorWithRequeue(cluster *v1beta1.TemporalCluster, reason string, err error, requeueAfter time.Duration) (ctrl.Result, error) {
	r.Recorder.Event(cluster, corev1.EventTypeWarning, "ProcessingError", err.Error())
	if reason == "" {
		reason = v1beta1.ReconcileErrorReason
	}
	v1beta1.SetTemporalClusterReconcileError(cluster, metav1.ConditionTrue, reason, err.Error())
	return reconcile.Result{RequeueAfter: requeueAfter}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *TemporalClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	for _, resource := range []client.Object{&appsv1.Deployment{}, &corev1.ConfigMap{}, &corev1.Service{}, &corev1.ServiceAccount{}, &networkingv1.Ingress{}, &batchv1.Job{}} {
		if err := mgr.GetFieldIndexer().IndexField(context.Background(), resource, ownerKey, addResourceToIndex); err != nil {
			return err
		}
	}

	controller := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.TemporalCluster{}, builder.WithPredicates(predicate.Or(
			predicate.GenerationChangedPredicate{},
			predicate.LabelChangedPredicate{},
			predicate.AnnotationChangedPredicate{},
		))).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&networkingv1.Ingress{}).
		Owns(&batchv1.Job{})

	if r.AvailableAPIs.CertManager {
		controller = controller.
			Owns(&certmanagerv1.Issuer{}).
			Owns(&certmanagerv1.Certificate{})

		for _, resource := range []client.Object{&certmanagerv1.Issuer{}, &certmanagerv1.Certificate{}} {
			if err := mgr.GetFieldIndexer().IndexField(context.Background(), resource, ownerKey, addCertManagerResourceToIndex); err != nil {
				return err
			}
		}
	}

	if r.AvailableAPIs.Istio {
		controller = controller.
			Owns(&istiosecurityv1beta1.PeerAuthentication{}).
			Owns(&istionetworkingv1beta1.DestinationRule{})

		for _, resource := range []client.Object{&istiosecurityv1beta1.PeerAuthentication{}, &istionetworkingv1beta1.DestinationRule{}} {
			if err := mgr.GetFieldIndexer().IndexField(context.Background(), resource, ownerKey, addIstioResourceToIndex); err != nil {
				return err
			}
		}
	}

	if r.AvailableAPIs.PrometheusOperator {
		controller = controller.Owns(&monitoringv1.ServiceMonitor{})

		for _, resource := range []client.Object{&monitoringv1.ServiceMonitor{}} {
			if err := mgr.GetFieldIndexer().IndexField(context.Background(), resource, ownerKey, addPromtheusOperatorResourceToIndex); err != nil {
				return err
			}
		}
	}

	return controller.Complete(r)
}

func addResourceToIndex(rawObj client.Object) []string {
	switch resourceObject := rawObj.(type) {
	case *appsv1.Deployment,
		*corev1.ConfigMap,
		*corev1.Service,
		*corev1.ServiceAccount,
		*networkingv1.Ingress,
		*batchv1.Job:
		owner := metav1.GetControllerOf(resourceObject)
		return validateAndGetOwner(owner)
	default:
		return nil
	}
}

func addIstioResourceToIndex(rawObj client.Object) []string {
	switch resourceObject := rawObj.(type) {
	case *istiosecurityv1beta1.PeerAuthentication,
		*istionetworkingv1beta1.DestinationRule:
		owner := metav1.GetControllerOf(resourceObject)
		return validateAndGetOwner(owner)
	default:
		return nil
	}
}

func addCertManagerResourceToIndex(rawObj client.Object) []string {
	switch resourceObject := rawObj.(type) {
	case *certmanagerv1.Issuer,
		*certmanagerv1.Certificate:
		owner := metav1.GetControllerOf(resourceObject)
		return validateAndGetOwner(owner)
	default:
		return nil
	}
}

func addPromtheusOperatorResourceToIndex(rawObj client.Object) []string {
	switch resourceObject := rawObj.(type) {
	case *monitoringv1.ServiceMonitor:
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
	if owner.APIVersion != v1beta1.GroupVersion.String() || owner.Kind != v1beta1.TemporalClusterTypeMeta.Kind {
		return nil
	}
	return []string{owner.Name}
}
