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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
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
func (r *TemporalClusterClientReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	namespacedName := types.NamespacedName{Namespace: req.Namespace, Name: clusterClient.Spec.ClusterRef.Name}
	cluster := &v1beta1.TemporalCluster{}
	err = r.Get(ctx, namespacedName, cluster)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !(cluster.MTLSWithCertManagerEnabled() && cluster.Spec.MTLS.FrontendEnabled()) {
		return reconcile.Result{Requeue: false}, errors.New("mTLS for frontend not enabled using cert-manager for the cluster, can't create a client")
	}

	builder := certmanager.NewGenericFrontendClientCertificateBuilder(cluster, r.Scheme, clusterClient.GetName())

	res, err := builder.Build()
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, res, func() error {
		err := builder.Update(res)
		if err != nil {
			return err
		}
		err = controllerutil.SetControllerReference(clusterClient, res, r.Scheme)
		if err != nil {
			return fmt.Errorf("failed setting controller reference: %v", err)
		}
		return nil
	})
	if err != nil {
		return reconcile.Result{}, err
	}

	certificate := res.(*certmanagerv1.Certificate)

	clusterClient.Status.ServerName = cluster.Spec.MTLS.Frontend.ServerName(cluster.ServerName())
	clusterClient.Status.SecretRef = corev1.LocalObjectReference{
		Name: certificate.Spec.SecretName,
	}

	err = r.Client.Status().Update(ctx, clusterClient)
	return reconcile.Result{}, err
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

	return controller.Complete(r)
}
