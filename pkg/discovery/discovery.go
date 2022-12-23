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
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

// AvailableAPIs holds available apis in the cluster.
type AvailableAPIs struct {
	Istio              bool
	CertManager        bool
	PrometheusOperator bool
}

type check struct {
	name          string
	groupVersion  string
	resource      string
	registerValue func(*AvailableAPIs, bool)
}

// FindAvailableAPIs searches for available well-known APIs in the cluster.
func FindAvailableAPIs(logger logr.Logger, config *rest.Config) (*AvailableAPIs, error) {
	resources := &AvailableAPIs{}

	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("can't create discovery client: %w", err)
	}

	checks := []check{
		{
			name:         "cert-manager",
			groupVersion: certmanagerv1.SchemeGroupVersion.String(),
			resource:     "certificates",
			registerValue: func(a *AvailableAPIs, found bool) {
				a.CertManager = found
			},
		},
		{
			name:         "istio",
			groupVersion: istionetworkingv1beta1.SchemeGroupVersion.String(),
			resource:     "destinationrules",
			registerValue: func(a *AvailableAPIs, found bool) {
				a.Istio = found
			},
		},
		{
			name:         "prometheus operator",
			groupVersion: monitoringv1.SchemeGroupVersion.String(),
			resource:     "servicemonitors",
			registerValue: func(a *AvailableAPIs, found bool) {
				a.PrometheusOperator = found
			},
		},
	}

	for _, check := range checks {
		found, err := k8sutil.IsAPIGroupVersionResourceSupported(client, check.groupVersion, check.resource)
		if err != nil {
			return nil, err
		}
		var msg string
		if found {
			msg = fmt.Sprintf("Found %s installation in the cluster, features requiring %s are enabled", check.name, check.name)
		} else {
			msg = fmt.Sprintf("Unable to find %s installation in the cluster, features requiring %s are disabled", check.name, check.name)
		}
		logger.Info(msg)
		check.registerValue(resources, found)
	}

	return resources, nil
}
