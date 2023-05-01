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

package istio_test

import (
	"testing"

	"github.com/alexandrevilain/temporal-operator/internal/resource/mtls/istio"
	"github.com/stretchr/testify/assert"
	istioapinetworkingv1beta1 "istio.io/api/networking/v1beta1"
)

func TestDestinationRuleEqual(t *testing.T) {
	tests := map[string]struct {
		a     *istioapinetworkingv1beta1.DestinationRule
		b     *istioapinetworkingv1beta1.DestinationRule
		equal bool
	}{
		"equal": {
			a: &istioapinetworkingv1beta1.DestinationRule{
				Host: "host",
				TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
					Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
						Mode: istioapinetworkingv1beta1.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
			},
			b: &istioapinetworkingv1beta1.DestinationRule{
				Host: "host",
				TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
					Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
						Mode: istioapinetworkingv1beta1.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
			},
			equal: true,
		},
		"host differs": {
			a: &istioapinetworkingv1beta1.DestinationRule{
				Host: "host-a",
				TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
					Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
						Mode: istioapinetworkingv1beta1.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
			},
			b: &istioapinetworkingv1beta1.DestinationRule{
				Host: "host-b",
				TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
					Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
						Mode: istioapinetworkingv1beta1.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
			},
			equal: false,
		},
		"traffic policy differs": {
			a: &istioapinetworkingv1beta1.DestinationRule{
				Host: "host",
				TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
					Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
						Mode: istioapinetworkingv1beta1.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
			},
			b: &istioapinetworkingv1beta1.DestinationRule{
				Host: "host",
				TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
					Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
						Mode: istioapinetworkingv1beta1.ClientTLSSettings_MUTUAL,
					},
				},
			},
			equal: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			b := &istio.DestinationRuleBuilder{}
			result := b.Equal(test.a, test.b)
			assert.Equal(tt, test.equal, result)
		})
	}
}
