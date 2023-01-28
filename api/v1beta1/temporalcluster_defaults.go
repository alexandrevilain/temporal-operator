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

package v1beta1

import (
	"time"

	"github.com/alexandrevilain/temporal-operator/pkg/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

const (
	defaultTemporalVersion = "1.17.4"
	defaultTemporalImage   = "temporalio/server"

	defaultTemporalUIVersion = "2.5.0"
	defaultTemporalUIImage   = "temporalio/ui"

	defaultTemporalAdmintoolsImage = "temporalio/admin-tools"
)

// Default set default fields values.
func (d *DatastoreSpec) Default() {
	if d.SQL != nil {
		if d.SQL.ConnectProtocol == "" {
			d.SQL.ConnectProtocol = "tcp"
		}
	}

	if d.Cassandra != nil {
		if d.Cassandra.ConnectTimeout == nil {
			d.Cassandra.ConnectTimeout = &metav1.Duration{Duration: 10 * time.Second}
		}
	}
}

// Default set default fields values.
func (c *TemporalCluster) Default() {
	if c.Spec.Version == nil {
		c.Spec.Version = version.MustNewVersionFromString(defaultTemporalVersion)
	}
	if c.Spec.Image == "" {
		c.Spec.Image = defaultTemporalImage
	}
	if c.Spec.Services == nil {
		c.Spec.Services = new(ServicesSpec)
	}
	// Frontend specs
	if c.Spec.Services.Frontend == nil {
		c.Spec.Services.Frontend = new(ServiceSpec)
	}
	if c.Spec.Services.Frontend.Replicas == nil {
		c.Spec.Services.Frontend.Replicas = pointer.Int32(1)
	}
	if c.Spec.Services.Frontend.Port == nil {
		c.Spec.Services.Frontend.Port = pointer.Int(7233)
	}
	if c.Spec.Services.Frontend.MembershipPort == nil {
		c.Spec.Services.Frontend.MembershipPort = pointer.Int(6933)
	}
	// History specs
	if c.Spec.Services.History == nil {
		c.Spec.Services.History = new(ServiceSpec)
	}
	if c.Spec.Services.History.Replicas == nil {
		c.Spec.Services.History.Replicas = pointer.Int32(1)
	}
	if c.Spec.Services.History.Port == nil {
		c.Spec.Services.History.Port = pointer.Int(7234)
	}
	if c.Spec.Services.History.MembershipPort == nil {
		c.Spec.Services.History.MembershipPort = pointer.Int(6934)
	}
	// Matching specs
	if c.Spec.Services.Matching == nil {
		c.Spec.Services.Matching = new(ServiceSpec)
	}
	if c.Spec.Services.Matching.Replicas == nil {
		c.Spec.Services.Matching.Replicas = pointer.Int32(1)
	}
	if c.Spec.Services.Matching.Port == nil {
		c.Spec.Services.Matching.Port = pointer.Int(7235)
	}
	if c.Spec.Services.Matching.MembershipPort == nil {
		c.Spec.Services.Matching.MembershipPort = pointer.Int(6935)
	}
	// Worker specs
	if c.Spec.Services.Worker == nil {
		c.Spec.Services.Worker = new(ServiceSpec)
	}
	if c.Spec.Services.Worker.Replicas == nil {
		c.Spec.Services.Worker.Replicas = pointer.Int32(1)
	}
	if c.Spec.Services.Worker.Port == nil {
		c.Spec.Services.Worker.Port = pointer.Int(7239)
	}
	if c.Spec.Services.Worker.MembershipPort == nil {
		c.Spec.Services.Worker.MembershipPort = pointer.Int(6939)
	}

	if c.Spec.Persistence.DefaultStore != nil {
		if c.Spec.Persistence.DefaultStore.Name == "" {
			c.Spec.Persistence.DefaultStore.Name = DefaultStoreName
		}
		c.Spec.Persistence.DefaultStore.Default()
	}

	if c.Spec.Persistence.VisibilityStore != nil {
		if c.Spec.Persistence.VisibilityStore.Name == "" {
			c.Spec.Persistence.VisibilityStore.Name = VisibilityStoreName
		}
		c.Spec.Persistence.VisibilityStore.Default()
	}

	if c.Spec.Persistence.AdvancedVisibilityStore != nil {
		if c.Spec.Persistence.AdvancedVisibilityStore.Name == "" {
			c.Spec.Persistence.AdvancedVisibilityStore.Name = AdvancedVisibilityStoreName
		}
		c.Spec.Persistence.AdvancedVisibilityStore.Default()
	}

	if c.Spec.UI == nil {
		c.Spec.UI = new(TemporalUISpec)
	}

	if c.Spec.UI.Version == "" {
		c.Spec.UI.Version = defaultTemporalUIVersion
	}

	if c.Spec.UI.Image == "" {
		c.Spec.UI.Image = defaultTemporalUIImage
	}

	if c.Spec.AdminTools == nil {
		c.Spec.AdminTools = new(TemporalAdminToolsSpec)
	}

	if c.Spec.AdminTools.Image == "" {
		c.Spec.AdminTools.Image = defaultTemporalAdmintoolsImage
	}

	if c.MTLSWithCertManagerEnabled() {
		if c.Spec.MTLS.RefreshInterval == nil {
			c.Spec.MTLS.RefreshInterval = &metav1.Duration{Duration: time.Hour}
		}
		if c.Spec.MTLS.CertificatesDuration == nil {
			c.Spec.MTLS.CertificatesDuration = &CertificatesDurationSpec{}
		}
		if c.Spec.MTLS.CertificatesDuration.RootCACertificate == nil {
			c.Spec.MTLS.CertificatesDuration.RootCACertificate = &metav1.Duration{Duration: time.Hour * 87600}
		}
		if c.Spec.MTLS.CertificatesDuration.IntermediateCAsCertificates == nil {
			c.Spec.MTLS.CertificatesDuration.IntermediateCAsCertificates = &metav1.Duration{Duration: time.Hour * 43830}
		}
		if c.Spec.MTLS.CertificatesDuration.ClientCertificates == nil {
			c.Spec.MTLS.CertificatesDuration.ClientCertificates = &metav1.Duration{Duration: time.Hour * 8766}
		}
		if c.Spec.MTLS.CertificatesDuration.FrontendCertificate == nil {
			c.Spec.MTLS.CertificatesDuration.FrontendCertificate = &metav1.Duration{Duration: time.Hour * 8766}
		}
		if c.Spec.MTLS.CertificatesDuration.InternodeCertificate == nil {
			c.Spec.MTLS.CertificatesDuration.InternodeCertificate = &metav1.Duration{Duration: time.Hour * 8766}
		}
	}

	if c.Spec.Metrics.MetricsEnabled() {
		if c.Spec.Metrics.Prometheus != nil {
			if c.Spec.Metrics.Prometheus.ListenPort == nil {
				c.Spec.Metrics.Prometheus.ListenPort = pointer.Int32(9090)
			}
		}
	}

	if c.Spec.DynamicConfig != nil {
		if c.Spec.DynamicConfig.PollInterval == nil {
			c.Spec.DynamicConfig.PollInterval = &metav1.Duration{Duration: time.Minute * 10}
		}
	}
}
