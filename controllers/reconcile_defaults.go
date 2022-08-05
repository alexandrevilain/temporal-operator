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
	"reflect"
	"time"

	appsv1alpha1 "github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

const (
	defaultTemporalVersion = "1.17.2"
	defaultTemporalImage   = "temporalio/server"

	defaultTemporalUIVersion = "2.5.0"
	defaultTemporalUIImage   = "temporalio/ui"

	defaultTemporalAdmintoolsImage = "temporalio/admin-tools"
)

func (r *TemporalClusterReconciler) reconcileDefaults(ctx context.Context, temporalCluster *appsv1alpha1.TemporalCluster) bool {
	before := temporalCluster.DeepCopy()

	if temporalCluster.Spec.Version == "" {
		temporalCluster.Spec.Version = defaultTemporalVersion
	}
	if temporalCluster.Spec.Image == "" {
		temporalCluster.Spec.Image = defaultTemporalImage
	}
	if temporalCluster.Spec.Services == nil {
		temporalCluster.Spec.Services = new(appsv1alpha1.TemporalServicesSpec)
	}
	// Frontend specs
	if temporalCluster.Spec.Services.Frontend == nil {
		temporalCluster.Spec.Services.Frontend = new(appsv1alpha1.ServiceSpec)
	}
	if temporalCluster.Spec.Services.Frontend.Replicas == nil {
		temporalCluster.Spec.Services.Frontend.Replicas = pointer.Int32(1)
	}
	if temporalCluster.Spec.Services.Frontend.Port == nil {
		temporalCluster.Spec.Services.Frontend.Port = pointer.Int(7233)
	}
	if temporalCluster.Spec.Services.Frontend.MembershipPort == nil {
		temporalCluster.Spec.Services.Frontend.MembershipPort = pointer.Int(6933)
	}
	// History specs
	if temporalCluster.Spec.Services.History == nil {
		temporalCluster.Spec.Services.History = new(appsv1alpha1.ServiceSpec)
	}
	if temporalCluster.Spec.Services.History.Replicas == nil {
		temporalCluster.Spec.Services.History.Replicas = pointer.Int32(1)
	}
	if temporalCluster.Spec.Services.History.Port == nil {
		temporalCluster.Spec.Services.History.Port = pointer.Int(7234)
	}
	if temporalCluster.Spec.Services.History.MembershipPort == nil {
		temporalCluster.Spec.Services.History.MembershipPort = pointer.Int(6934)
	}
	// Matching specs
	if temporalCluster.Spec.Services.Matching == nil {
		temporalCluster.Spec.Services.Matching = new(appsv1alpha1.ServiceSpec)
	}
	if temporalCluster.Spec.Services.Matching.Replicas == nil {
		temporalCluster.Spec.Services.Matching.Replicas = pointer.Int32(1)
	}
	if temporalCluster.Spec.Services.Matching.Port == nil {
		temporalCluster.Spec.Services.Matching.Port = pointer.Int(7235)
	}
	if temporalCluster.Spec.Services.Matching.MembershipPort == nil {
		temporalCluster.Spec.Services.Matching.MembershipPort = pointer.Int(6935)
	}
	// Worker specs
	if temporalCluster.Spec.Services.Worker == nil {
		temporalCluster.Spec.Services.Worker = new(appsv1alpha1.ServiceSpec)
	}
	if temporalCluster.Spec.Services.Worker.Replicas == nil {
		temporalCluster.Spec.Services.Worker.Replicas = pointer.Int32(1)
	}
	if temporalCluster.Spec.Services.Worker.Port == nil {
		temporalCluster.Spec.Services.Worker.Port = pointer.Int(7239)
	}
	if temporalCluster.Spec.Services.Worker.MembershipPort == nil {
		temporalCluster.Spec.Services.Worker.MembershipPort = pointer.Int(6939)
	}

	for _, datastore := range temporalCluster.Spec.Datastores {
		if datastore.SQL != nil {
			if datastore.SQL.ConnectProtocol == "" {
				datastore.SQL.ConnectProtocol = "tcp"
			}
		}
	}

	if temporalCluster.Spec.Persistence.VisibilityStore == "" {
		temporalCluster.Spec.Persistence.VisibilityStore = temporalCluster.Spec.Persistence.DefaultStore
	}

	if temporalCluster.Spec.UI == nil {
		temporalCluster.Spec.UI = new(appsv1alpha1.TemporalUISpec)
	}

	if temporalCluster.Spec.UI.Version == "" {
		temporalCluster.Spec.UI.Version = defaultTemporalUIVersion
	}

	if temporalCluster.Spec.UI.Image == "" {
		temporalCluster.Spec.UI.Image = defaultTemporalUIImage
	}

	if temporalCluster.Spec.AdminTools == nil {
		temporalCluster.Spec.AdminTools = new(appsv1alpha1.TemporalAdminToolsSpec)
	}

	if temporalCluster.Spec.AdminTools.Image == "" {
		temporalCluster.Spec.AdminTools.Image = defaultTemporalAdmintoolsImage
	}

	if temporalCluster.MTLSWithCertManagerEnabled() {
		if temporalCluster.Spec.MTLS.RefreshInterval == nil {
			temporalCluster.Spec.MTLS.RefreshInterval = &metav1.Duration{Duration: time.Hour}
		}
		if temporalCluster.Spec.MTLS.CertificatesDuration == nil {
			temporalCluster.Spec.MTLS.CertificatesDuration = &appsv1alpha1.CertificatesDurationSpec{}
		}
		if temporalCluster.Spec.MTLS.CertificatesDuration.RootCACertificate == nil {
			temporalCluster.Spec.MTLS.CertificatesDuration.RootCACertificate = &metav1.Duration{Duration: time.Hour * 87600}
		}
		if temporalCluster.Spec.MTLS.CertificatesDuration.IntermediateCAsCertificates == nil {
			temporalCluster.Spec.MTLS.CertificatesDuration.IntermediateCAsCertificates = &metav1.Duration{Duration: time.Hour * 43830}
		}
		if temporalCluster.Spec.MTLS.CertificatesDuration.ClientCertificates == nil {
			temporalCluster.Spec.MTLS.CertificatesDuration.ClientCertificates = &metav1.Duration{Duration: time.Hour * 8766}
		}
		if temporalCluster.Spec.MTLS.CertificatesDuration.FrontendCertificate == nil {
			temporalCluster.Spec.MTLS.CertificatesDuration.FrontendCertificate = &metav1.Duration{Duration: time.Hour * 8766}
		}
		if temporalCluster.Spec.MTLS.CertificatesDuration.InternodeCertificate == nil {
			temporalCluster.Spec.MTLS.CertificatesDuration.InternodeCertificate = &metav1.Duration{Duration: time.Hour * 8766}
		}
	}

	return !reflect.DeepEqual(before.Spec, temporalCluster.Spec)
}
