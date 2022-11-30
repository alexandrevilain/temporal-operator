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
	"context"
	"fmt"

	"github.com/alexandrevilain/temporal-operator/pkg/apichecker/istio"
	"github.com/alexandrevilain/temporal-operator/pkg/apichecker/prometheus"
	certmanager "github.com/cert-manager/cert-manager/pkg/util/cmapichecker"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

// AvailableAPIs holds available apis in the cluster.
type AvailableAPIs struct {
	Istio              bool
	CertManager        bool
	PrometheusOperator bool
}

// ResourcesConfig holds namespaces where to create resources to check if they exists.
type ResourcesConfig struct {
	IstioNamespace              string
	CertManagerNamespace        string
	PrometheusOperatorNamespace string
}

// FindAvailableAPIs searches for available well-known APIs in the cluster.
func FindAvailableAPIs(ctx context.Context, logger logr.Logger, mgr ctrl.Manager, cfg ResourcesConfig) (*AvailableAPIs, error) {
	resources := &AvailableAPIs{}

	cmAPIChecker, err := certmanager.New(mgr.GetConfig(), mgr.GetScheme(), cfg.CertManagerNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to create cert-manager api checker: %w", err)
	}

	err = cmAPIChecker.Check(ctx)
	if err != nil {
		logger.Info("Unable to find cert-manager installation in the cluster, features requiring cert-manager are disabled")
	} else {
		resources.CertManager = true
		logger.Info("Found cert-manager installation in the cluster, features requiring cert-manager are enabled")
	}

	istioAPIChecker, err := istio.NewAPIChecker(mgr.GetConfig(), mgr.GetScheme(), cfg.IstioNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to create istio api checker: %w", err)
	}

	err = istioAPIChecker.Check(ctx)
	if err != nil {
		logger.Info("Unable to find istio installation in the cluster, features requiring istio are disabled")
	} else {
		resources.Istio = true
		logger.Info("Found istio installation in the cluster, features requiring istio are enabled")
	}

	promOperatorAPIChecker, err := prometheus.NewAPIChecker(mgr.GetConfig(), mgr.GetScheme(), cfg.PrometheusOperatorNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to create prometheus operator api checker: %w", err)
	}

	err = promOperatorAPIChecker.Check(ctx)
	if err != nil {
		logger.Info("Unable to find prometheus operator installation in the cluster, features requiring prometheus operator are disabled")
	} else {
		resources.PrometheusOperator = true
		logger.Info("Found prometheus operator installation in the cluster, features requiring prometheus operator are enabled")
	}

	return resources, nil
}
