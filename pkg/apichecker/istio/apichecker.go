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
	"fmt"

	"github.com/alexandrevilain/temporal-operator/pkg/apichecker"
	istioapinetworkingv1beta1 "istio.io/api/networking/v1beta1"
	istioapisecurityv1beta1 "istio.io/api/security/v1beta1"
	istioapiv1beta1 "istio.io/api/type/v1beta1"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var fakeObjects = []client.Object{
	&istionetworkingv1beta1.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "fake-",
		},
		Spec: istioapinetworkingv1beta1.DestinationRule{
			Host: "fake",
			TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
				Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
					Mode: istioapinetworkingv1beta1.ClientTLSSettings_ISTIO_MUTUAL,
				},
			},
		},
	},
	&istiosecurityv1beta1.PeerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "fake-",
		},
		Spec: istioapisecurityv1beta1.PeerAuthentication{
			Selector: &istioapiv1beta1.WorkloadSelector{
				MatchLabels: map[string]string{
					"fake": "fake",
				},
			},
			Mtls: &istioapisecurityv1beta1.PeerAuthentication_MutualTLS{
				Mode: istioapisecurityv1beta1.PeerAuthentication_MutualTLS_STRICT,
			},
		},
	},
}

// NewAPIChecker creates a new ApiChecker.
func NewAPIChecker(restcfg *rest.Config, scheme *runtime.Scheme, namespace string) (*apichecker.APIChecker, error) {
	if err := istiosecurityv1beta1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("can't add istiosecurityv1beta1 to scheme: %w", err)
	}
	if err := istionetworkingv1beta1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("can't add istionetworkingv1beta1 to scheme: %w", err)
	}

	return apichecker.NewAPIChecker(restcfg, scheme, namespace, fakeObjects)
}
