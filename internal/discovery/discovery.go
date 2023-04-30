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

	"github.com/alexandrevilain/controller-tools/pkg/discovery"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/go-logr/logr"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

// AvailableAPIs holds available apis in the cluster.
type AvailableAPIs struct {
	Istio              bool
	CertManager        bool
	PrometheusOperator bool
}

// FindAvailableAPIs searches for available well-known APIs in the cluster.
func FindAvailableAPIs(logger logr.Logger, mgr discovery.Manager) (*AvailableAPIs, error) {
	resources := &AvailableAPIs{}

	var err error
	resources.CertManager, err = mgr.AreObjectsSupported(&certmanagerv1.Issuer{}, &certmanagerv1.Certificate{})
	if err != nil {
		return nil, fmt.Errorf("can't determine if cert-manager is available: %w", err)
	}

	resources.Istio, err = mgr.AreObjectsSupported(&istiosecurityv1beta1.PeerAuthentication{}, &istionetworkingv1beta1.DestinationRule{})
	if err != nil {
		return nil, fmt.Errorf("can't determine if istio is available: %w", err)
	}

	resources.PrometheusOperator, err = mgr.AreObjectsSupported(&monitoringv1.ServiceMonitor{})
	if err != nil {
		return nil, fmt.Errorf("can't determine if prometheus-operator is available: %w", err)
	}

	logResourceAvailability(logger, "cert-manager", resources.CertManager)
	logResourceAvailability(logger, "istio", resources.Istio)
	logResourceAvailability(logger, "prometheus-operator", resources.PrometheusOperator)

	return resources, nil
}

func logResourceAvailability(logger logr.Logger, apiName string, found bool) {
	var msg string
	if found {
		msg = fmt.Sprintf("Found %s installation in the cluster, features requiring %s are enabled", apiName, apiName)
	} else {
		msg = fmt.Sprintf("Unable to find %s installation in the cluster, features requiring %s are disabled", apiName, apiName)
	}
	logger.Info(msg)
}
