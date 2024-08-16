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
	"k8s.io/utils/ptr"
)

const (
	defaultTemporalVersion = "1.23.0"
	defaultTemporalImage   = "temporalio/server"

	defaultTemporalUIImage = "temporalio/ui"
	defaultTemporalUITag   = "2.25.0"

	defaultTemporalAdmintoolsImage = "temporalio/admin-tools"
	defaultTemporalAdmintoolsTag   = "1.23.1.1-tctl-1.18.1-cli-0.12.0"
)

// Default set default fields values.
func (s *DatastoreSpec) Default() {
	if s.SQL != nil {
		if s.SQL.ConnectProtocol == "" {
			s.SQL.ConnectProtocol = "tcp"
		}
	}

	if s.Cassandra != nil {
		if s.Cassandra.ConnectTimeout == nil {
			s.Cassandra.ConnectTimeout = &metav1.Duration{Duration: 10 * time.Second}
		}
	}

	if s.Elasticsearch != nil {
		if s.Elasticsearch.Indices.Visibility == "" {
			s.Elasticsearch.Indices.Visibility = "temporal_visibility_v1"
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
	if c.Spec.Tag == "" {
		c.Spec.Tag = c.Spec.Version.String()
	}

	if c.Spec.Log == nil {
		c.Spec.Log = new(LogSpec)
		if c.Spec.Log.Stdout == nil {
			c.Spec.Log.Stdout = ptr.To(true)
		}
		if c.Spec.Log.Level == "" {
			c.Spec.Log.Level = "info"
		}
		if c.Spec.Log.Format == "" {
			c.Spec.Log.Format = "json"
		}
	}

	if c.Spec.Services == nil {
		c.Spec.Services = new(ServicesSpec)
	}
	// Frontend specs
	if c.Spec.Services.Frontend == nil {
		c.Spec.Services.Frontend = new(ServiceSpec)
	}
	if c.Spec.Services.Frontend.Replicas == nil {
		c.Spec.Services.Frontend.Replicas = ptr.To[int32](1)
	}
	if c.Spec.Services.Frontend.Port == nil {
		c.Spec.Services.Frontend.Port = ptr.To(7233)
	}
	if c.Spec.Services.Frontend.MembershipPort == nil {
		c.Spec.Services.Frontend.MembershipPort = ptr.To(6933)
	}
	if c.Spec.Services.Frontend.HTTPPort == nil {
		c.Spec.Services.Frontend.HTTPPort = ptr.To(7243)
	}
	// Internal Frontend specs
	if c.Spec.Services.InternalFrontend.IsEnabled() {
		if c.Spec.Services.InternalFrontend.Replicas == nil {
			c.Spec.Services.InternalFrontend.Replicas = ptr.To[int32](1)
		}
		if c.Spec.Services.InternalFrontend.Port == nil {
			c.Spec.Services.InternalFrontend.Port = ptr.To(7236)
		}
		if c.Spec.Services.InternalFrontend.MembershipPort == nil {
			c.Spec.Services.InternalFrontend.MembershipPort = ptr.To(6936)
		}
		if c.Spec.Services.InternalFrontend.HTTPPort == nil {
			c.Spec.Services.InternalFrontend.HTTPPort = ptr.To(0)
		}
	}
	// History specs
	if c.Spec.Services.History == nil {
		c.Spec.Services.History = new(ServiceSpec)
	}
	if c.Spec.Services.History.Replicas == nil {
		c.Spec.Services.History.Replicas = ptr.To[int32](1)
	}
	if c.Spec.Services.History.Port == nil {
		c.Spec.Services.History.Port = ptr.To(7234)
	}
	if c.Spec.Services.History.MembershipPort == nil {
		c.Spec.Services.History.MembershipPort = ptr.To(6934)
	}
	if c.Spec.Services.History.HTTPPort == nil {
		c.Spec.Services.History.HTTPPort = ptr.To(0)
	}
	// Matching specs
	if c.Spec.Services.Matching == nil {
		c.Spec.Services.Matching = new(ServiceSpec)
	}
	if c.Spec.Services.Matching.Replicas == nil {
		c.Spec.Services.Matching.Replicas = ptr.To[int32](1)
	}
	if c.Spec.Services.Matching.Port == nil {
		c.Spec.Services.Matching.Port = ptr.To(7235)
	}
	if c.Spec.Services.Matching.MembershipPort == nil {
		c.Spec.Services.Matching.MembershipPort = ptr.To(6935)
	}
	if c.Spec.Services.Matching.HTTPPort == nil {
		c.Spec.Services.Matching.HTTPPort = ptr.To(0)
	}
	// Worker specs
	if c.Spec.Services.Worker == nil {
		c.Spec.Services.Worker = new(ServiceSpec)
	}
	if c.Spec.Services.Worker.Replicas == nil {
		c.Spec.Services.Worker.Replicas = ptr.To[int32](1)
	}
	if c.Spec.Services.Worker.Port == nil {
		c.Spec.Services.Worker.Port = ptr.To(7239)
	}
	if c.Spec.Services.Worker.MembershipPort == nil {
		c.Spec.Services.Worker.MembershipPort = ptr.To(6939)
	}
	if c.Spec.Services.Worker.HTTPPort == nil {
		c.Spec.Services.Worker.HTTPPort = ptr.To(0)
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

	if c.Spec.Persistence.SecondaryVisibilityStore != nil {
		if c.Spec.Persistence.SecondaryVisibilityStore.Name == "" {
			c.Spec.Persistence.SecondaryVisibilityStore.Name = SecondaryVisibilityStoreName
		}
		c.Spec.Persistence.SecondaryVisibilityStore.Default()
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

	if c.Spec.UI.Tag == "" {
		if c.Spec.UI.Version != "" {
			c.Spec.UI.Tag = c.Spec.UI.Version
		} else {
			c.Spec.UI.Tag = defaultTemporalUITag
		}
	}

	if c.Spec.UI.Image == "" {
		c.Spec.UI.Image = defaultTemporalUIImage
	}

	if c.Spec.UI.Replicas == nil {
		c.Spec.UI.Replicas = ptr.To[int32](1)
	}

	if c.Spec.AdminTools == nil {
		c.Spec.AdminTools = new(TemporalAdminToolsSpec)
	}

	if c.Spec.AdminTools.Image == "" {
		c.Spec.AdminTools.Image = defaultTemporalAdmintoolsImage
	}

	if c.Spec.AdminTools.Tag == "" {
		c.Spec.AdminTools.Tag = defaultTemporalAdmintoolsTag
	}

	if c.Spec.MTLS != nil {
		if c.Spec.MTLS.RefreshInterval == nil {
			c.Spec.MTLS.RefreshInterval = &metav1.Duration{Duration: time.Hour}
		}
	}

	if c.MTLSWithCertManagerEnabled() {
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

	if c.Spec.Metrics.IsEnabled() {
		if c.Spec.Metrics.Prometheus != nil {
			if c.Spec.Metrics.Prometheus.ListenPort == nil {
				c.Spec.Metrics.Prometheus.ListenPort = ptr.To[int32](9090)
			}
		}
	}

	if c.Spec.DynamicConfig != nil {
		if c.Spec.DynamicConfig.PollInterval == nil {
			c.Spec.DynamicConfig.PollInterval = &metav1.Duration{Duration: time.Minute * 10}
		}
	}
}
