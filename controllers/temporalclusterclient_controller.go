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
	"time"

	"github.com/alexandrevilain/temporal-operator/pkg/kubernetes/patch"
	certmanagerapiutil "github.com/cert-manager/cert-manager/pkg/api/util"
	certmanagermeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/kubernetes"
	"github.com/alexandrevilain/temporal-operator/pkg/reconciler"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/mtls/certmanager"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
)

// TemporalClusterClientReconciler reconciles a ClusterClient object
type TemporalClusterClientReconciler struct {
	reconciler.Base
}

//+kubebuilder:rbac:groups=temporal.io,resources=temporalclusterclients,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=temporal.io,resources=temporalclusterclients/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=temporal.io,resources=temporalclusterclients/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *TemporalClusterClientReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	logger := log.FromContext(ctx)

	logger.Info("Starting reconciliation")

	clusterClient := &v1beta1.TemporalClusterClient{}
	err := r.Get(ctx, req.NamespacedName, clusterClient)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Check if the resource has been marked for deletion
	if !clusterClient.ObjectMeta.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, nil
	}

	patchHelper, err := patch.NewHelper(clusterClient, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}

	defer func() {
		// Always attempt to Patch the ClusterClient object and status after each reconciliation.
		err := patchHelper.Patch(ctx, clusterClient)
		if err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	// Get referenced cluster.
	cluster := &v1beta1.TemporalCluster{}
	err = r.Get(ctx, clusterClient.Spec.ClusterRef.NamespacedName(clusterClient), cluster)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !(cluster.MTLSWithCertManagerEnabled() && cluster.Spec.MTLS.FrontendEnabled()) {
		return reconcile.Result{Requeue: false}, errors.New("mTLS for frontend not enabled using cert-manager for the cluster, can't create a client")
	}

	clusterClient.Status.ServerName = cluster.Spec.MTLS.Frontend.ServerName(cluster.ServerName())
	if clusterClient.Status.SecretRef == nil {
		clusterClient.Status.SecretRef = &corev1.LocalObjectReference{
			Name: "",
		}
	}

	builder := certmanager.NewGenericFrontendClientCertificateBuilder(cluster, r.Scheme, clusterClient.GetName())
	certificateObject, err := builder.Build()
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, certificateObject, func() error {
		return builder.Update(certificateObject)
	})
	if err != nil {
		return reconcile.Result{}, err
	}

	certificate := certificateObject.(*certmanagerv1.Certificate)

	condition := certmanagerapiutil.GetCertificateCondition(certificate, certmanagerv1.CertificateConditionReady)
	if condition == nil || condition.Status != certmanagermeta.ConditionTrue {
		logger.Info("Waiting for certificate to become ready, requeuing")
		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	}

	if clusterClient.GetNamespace() != cluster.GetNamespace() {
		originalSecret := client.ObjectKey{Namespace: certificate.GetNamespace(), Name: certificate.Spec.SecretName}
		err = kubernetes.NewSecretCopier(r.Client, r.Scheme).Copy(ctx, clusterClient, originalSecret, clusterClient.GetNamespace())
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	clusterClient.Status.SecretRef = &corev1.LocalObjectReference{
		Name: certificate.Spec.SecretName,
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TemporalClusterClientReconciler) SetupWithManager(mgr ctrl.Manager) error {
	controller := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.TemporalClusterClient{})

	if r.AvailableAPIs.CertManager {
		controller = controller.
			Owns(&certmanagerv1.Issuer{}).
			Owns(&certmanagerv1.Certificate{})
	}

	controller.Owns(&corev1.Secret{})

	return controller.Complete(r)
}
