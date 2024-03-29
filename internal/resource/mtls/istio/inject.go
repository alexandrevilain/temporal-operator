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

package istio

import (
	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
)

// GetLabels returns istio labels to enable proxy injection if the provided Cluster
// instance has mTLS enabled using istio.
func GetLabels(instance *v1beta1.TemporalCluster) map[string]string {
	if instance.Spec.MTLS != nil && instance.Spec.MTLS.Provider == v1beta1.IstioMTLSProvider {
		return map[string]string{
			"sidecar.istio.io/inject": "true",
		}
	}
	return map[string]string{}
}

// GetAnnotations returns istio annotations to delay application startup until the pod proxy is ready to accept traffic.
// Returned only if the provided Cluster instance has mTLS enabled using istio.
func GetAnnotations(instance *v1beta1.TemporalCluster) map[string]string {
	if instance.Spec.MTLS != nil && instance.Spec.MTLS.Provider == v1beta1.IstioMTLSProvider {
		return map[string]string{
			"proxy.istio.io/config": `{ "holdApplicationUntilProxyStarts": true }`,
		}
	}
	return map[string]string{}
}
