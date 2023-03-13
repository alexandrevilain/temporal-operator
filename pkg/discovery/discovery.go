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

package discovery

import (
	"fmt"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/go-logr/logr"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

var (
	istioCheckGVK              = schema.FromAPIVersionAndKind(istionetworkingv1beta1.SchemeGroupVersion.String(), "DestinationRule")
	certManagerCheckGVK        = schema.FromAPIVersionAndKind(certmanagerv1.SchemeGroupVersion.String(), "Certificate")
	prometheusOperatorCheckGVK = schema.FromAPIVersionAndKind(monitoringv1.SchemeGroupVersion.String(), "ServiceMonitor")
)

// AvailableAPIs holds available apis in the cluster.
type AvailableAPIs struct {
	scheme *runtime.Scheme
}

func (c *AvailableAPIs) Istio() bool {
	return c.Available(istioCheckGVK)
}

func (c *AvailableAPIs) CertManager() bool {
	return c.Available(certManagerCheckGVK)
}

func (c *AvailableAPIs) PrometheusOperator() bool {
	return c.Available(prometheusOperatorCheckGVK)
}

func (c *AvailableAPIs) Available(gvk schema.GroupVersionKind) bool {
	return c.scheme.Recognizes(gvk)
}

type check struct {
	name          string
	gv            schema.GroupVersion
	resource      string
	schemeBuilder runtime.SchemeBuilder
}

// FindAvailableAPIs searches for available well-known APIs in the cluster.
func FindAvailableAPIs(logger logr.Logger, config *rest.Config) (*AvailableAPIs, error) {
	resources := &AvailableAPIs{
		scheme: runtime.NewScheme(),
	}

	if err := clientgoscheme.AddToScheme(resources.scheme); err != nil {
		return nil, err
	}

	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("can't create discovery client: %w", err)
	}

	checks := []check{
		{
			name:     "cert-manager",
			gv:       certManagerCheckGVK.GroupVersion(),
			resource: "certificates",
			schemeBuilder: runtime.NewSchemeBuilder(
				certmanagerv1.AddToScheme,
			),
		},
		{
			name:     "istio",
			gv:       istioCheckGVK.GroupVersion(),
			resource: "destinationrules",
			schemeBuilder: runtime.NewSchemeBuilder(
				istiosecurityv1beta1.AddToScheme,
				istionetworkingv1beta1.AddToScheme,
			),
		},
		{
			name:     "prometheus operator",
			gv:       prometheusOperatorCheckGVK.GroupVersion(),
			resource: "servicemonitors",
			schemeBuilder: runtime.NewSchemeBuilder(
				monitoringv1.AddToScheme,
			),
		},
	}

	for _, check := range checks {
		found, err := k8sutil.IsAPIGroupVersionResourceSupported(client, check.gv.String(), check.resource)
		if err != nil {
			return nil, err
		}
		if !found {
			logger.Info(fmt.Sprintf("Unable to find %s installation in the cluster, features requiring %s are disabled", check.name, check.name))
			continue
		}

		logger.Info(fmt.Sprintf("Found %s installation in the cluster, features requiring %s are enabled", check.name, check.name))
		err = check.schemeBuilder.AddToScheme(resources.scheme)
		if err != nil {
			return nil, err
		}
	}

	return resources, nil
}
