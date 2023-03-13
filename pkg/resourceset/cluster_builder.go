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

package resourceset

import (
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/discovery"
	"github.com/alexandrevilain/temporal-operator/pkg/resource"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/admintools"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/cluster"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/mtls/certmanager"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/mtls/istio"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/prometheus"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/ui"
	"github.com/alexandrevilain/temporal-operator/pkg/resourceset/state"
	"go.temporal.io/server/common/primitives"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type ClusterBuilder struct {
	Instance      *v1beta1.TemporalCluster
	Scheme        *runtime.Scheme
	AvailableAPIs *discovery.AvailableAPIs
}

func (b *ClusterBuilder) getClusterServices() []primitives.ServiceName {
	services := []primitives.ServiceName{
		primitives.FrontendService,
		primitives.HistoryService,
		primitives.MatchingService,
		primitives.WorkerService,
	}

	if b.Instance.Spec.Services.InternalFrontend.IsEnabled() {
		services = append(services, primitives.InternalFrontendService)
	}

	return services
}

func (b *ClusterBuilder) ResourceBuilders() ([]resource.Builder, error) {
	builders := []resource.Builder{
		cluster.NewConfigmapBuilder(b.Instance, b.Scheme),
		cluster.NewFrontendServiceBuilder(b.Instance, b.Scheme),
	}

	for _, service := range b.getClusterServices() {
		specs, err := b.Instance.Spec.Services.GetServiceSpec(service)
		if err != nil {
			return nil, err
		}

		serviceName := string(service)

		builders = append(builders, cluster.NewServiceAccountBuilder(serviceName, b.Instance, b.Scheme, specs))
		builders = append(builders, cluster.NewDeploymentBuilder(serviceName, b.Instance, b.Scheme, specs))
		builders = append(builders, cluster.NewHeadlessServiceBuilder(serviceName, b.Instance, b.Scheme, specs))

		if b.Instance.Spec.MTLS != nil && b.Instance.Spec.MTLS.Provider == v1beta1.IstioMTLSProvider {
			builders = append(builders, istio.NewPeerAuthenticationBuilder(serviceName, b.Instance, b.Scheme, specs))
			builders = append(builders, istio.NewDestinationRuleBuilder(serviceName, b.Instance, b.Scheme, specs))
		}

		if b.Instance.Spec.Metrics.IsEnabled() &&
			b.Instance.Spec.Metrics.Prometheus != nil &&
			b.Instance.Spec.Metrics.Prometheus.ScrapeConfig != nil &&
			b.Instance.Spec.Metrics.Prometheus.ScrapeConfig.ServiceMonitor != nil &&
			b.Instance.Spec.Metrics.Prometheus.ScrapeConfig.ServiceMonitor.Enabled {
			builders = append(builders, prometheus.NewServiceMonitorBuilder(serviceName, b.Instance, b.Scheme, specs))
		}
	}

	if b.Instance.Spec.DynamicConfig != nil {
		builders = append(builders, cluster.NewDynamicConfigmapBuilder(b.Instance, b.Scheme))
	}

	if b.Instance.MTLSWithCertManagerEnabled() {
		builders = append(builders,
			certmanager.NewMTLSBootstrapIssuerBuilder(b.Instance, b.Scheme),
			certmanager.NewMTLSRootCACertificateBuilder(b.Instance, b.Scheme),
			certmanager.NewMTLSRootCAIssuerBuilder(b.Instance, b.Scheme),
		)

		if b.Instance.Spec.MTLS.InternodeEnabled() {
			builders = append(builders,
				certmanager.NewMTLSInternodeIntermediateCACertificateBuilder(b.Instance, b.Scheme),
				certmanager.NewMTLSInternodeIntermediateCAIssuerBuilder(b.Instance, b.Scheme),
				certmanager.NewMTLSInternodeCertificateBuilder(b.Instance, b.Scheme),
			)
		}

		if b.Instance.Spec.MTLS.FrontendEnabled() {
			builders = append(builders,
				certmanager.NewMTLSFrontendIntermediateCACertificateBuilder(b.Instance, b.Scheme),
				certmanager.NewMTLSFrontendIntermediateCAIssuerBuilder(b.Instance, b.Scheme),
				certmanager.NewMTLSFrontendCertificateBuilder(b.Instance, b.Scheme),
			)

			if !b.Instance.Spec.Services.InternalFrontend.IsEnabled() {
				builders = append(builders, certmanager.NewWorkerFrontendClientCertificateBuilder(b.Instance, b.Scheme))
			}
		}
	}

	if b.Instance.Spec.UI != nil && b.Instance.Spec.UI.Enabled {
		builders = append(builders,
			ui.NewDeploymentBuilder(b.Instance, b.Scheme),
			ui.NewServiceBuilder(b.Instance, b.Scheme),
		)
		if b.Instance.Spec.UI.Ingress != nil {
			builders = append(builders, ui.NewIngressBuilder(b.Instance, b.Scheme))
		}

		if b.Instance.MTLSWithCertManagerEnabled() && b.Instance.Spec.MTLS.FrontendEnabled() {
			builders = append(builders, ui.NewFrontendClientCertificateBuilder(b.Instance, b.Scheme))
		}
	}

	if b.Instance.Spec.AdminTools != nil && b.Instance.Spec.AdminTools.Enabled {
		builders = append(builders, admintools.NewDeploymentBuilder(b.Instance, b.Scheme))

		if b.Instance.MTLSWithCertManagerEnabled() && b.Instance.Spec.MTLS.FrontendEnabled() {
			builders = append(builders, admintools.NewFrontendClientCertificateBuilder(b.Instance, b.Scheme))
		}
	}

	return builders, nil
}

func (b *ClusterBuilder) ResourcePruners() ([]resource.Pruner, error) {
	builders, err := b.ResourceBuilders()
	if err != nil {
		return nil, err
	}
	genericBuilderFactories := []resource.GenericBuilderFactory{
		// Default cluster builders
		cluster.NewConfigmapBuilder,
		cluster.NewFrontendServiceBuilder,
		// Dynamic config
		cluster.NewDynamicConfigmapBuilder,
		// UI
		ui.NewDeploymentBuilder,
		ui.NewServiceBuilder,
		ui.NewIngressBuilder,
		ui.NewFrontendClientCertificateBuilder,
		// Admin tools
		admintools.NewDeploymentBuilder,
		admintools.NewFrontendClientCertificateBuilder,
		// mTLS using cert-manager
		certmanager.NewMTLSBootstrapIssuerBuilder,
		certmanager.NewMTLSRootCACertificateBuilder,
		certmanager.NewMTLSRootCAIssuerBuilder,
		certmanager.NewMTLSInternodeIntermediateCACertificateBuilder,
		certmanager.NewMTLSInternodeIntermediateCAIssuerBuilder,
		certmanager.NewMTLSInternodeCertificateBuilder,
		certmanager.NewMTLSFrontendIntermediateCACertificateBuilder,
		certmanager.NewMTLSFrontendIntermediateCAIssuerBuilder,
		certmanager.NewMTLSFrontendCertificateBuilder,
		certmanager.NewWorkerFrontendClientCertificateBuilder,
	}

	serviceSpecificBuilderFactories := []resource.ServiceSpecificBuilderFactory{
		cluster.NewServiceAccountBuilder,
		cluster.NewDeploymentBuilder,
		cluster.NewHeadlessServiceBuilder,
		// mTLS using istio
		istio.NewPeerAuthenticationBuilder,
		istio.NewDestinationRuleBuilder,
		// Monitoring using prometheus operator
		prometheus.NewServiceMonitorBuilder,
	}

	desiredState := state.NewDesired(b.Scheme)

	for _, builder := range builders {
		resource := builder.Build()
		if err := desiredState.Add(resource); err != nil {
			return nil, err
		}
	}

	allPruners := []resource.Pruner{}

	for _, genericBuilderFactory := range genericBuilderFactories {
		allPruners = append(allPruners, genericBuilderFactory(b.Instance, b.Scheme))
	}

	for _, service := range b.getClusterServices() {
		serviceName := string(service)
		for _, serviceSpecificBuilderFactory := range serviceSpecificBuilderFactories {
			allPruners = append(allPruners, serviceSpecificBuilderFactory(serviceName, b.Instance, b.Scheme, nil))
		}
	}

	pruners := []resource.Pruner{}

	for _, pruner := range allPruners {
		resource := pruner.Build()

		resourceGVK, err := apiutil.GVKForObject(resource, b.Scheme)
		if err != nil {
			return pruners, fmt.Errorf("can't get object GVK: %w", err)
		}

		if !b.AvailableAPIs.Available(resourceGVK) {
			continue
		}

		if ok, err := desiredState.Has(resource); err != nil {
			return nil, err
		} else if !ok {
			pruners = append(pruners, pruner)
		}
	}

	return pruners, nil
}
