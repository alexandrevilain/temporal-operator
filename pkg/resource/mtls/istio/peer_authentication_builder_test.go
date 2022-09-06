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

	"github.com/alexandrevilain/temporal-operator/pkg/resource/mtls/istio"
	"github.com/stretchr/testify/assert"
	istioapisecurityv1beta1 "istio.io/api/security/v1beta1"
	istioapiv1beta1 "istio.io/api/type/v1beta1"
)

func TestPeerAuthenticationEqual(t *testing.T) {
	tests := map[string]struct {
		a     *istioapisecurityv1beta1.PeerAuthentication
		b     *istioapisecurityv1beta1.PeerAuthentication
		equal bool
	}{
		"equal": {
			a: &istioapisecurityv1beta1.PeerAuthentication{
				Selector: &istioapiv1beta1.WorkloadSelector{
					MatchLabels: map[string]string{
						"app": "test",
					},
				},
				Mtls: &istioapisecurityv1beta1.PeerAuthentication_MutualTLS{
					Mode: istioapisecurityv1beta1.PeerAuthentication_MutualTLS_STRICT,
				},
			},
			b: &istioapisecurityv1beta1.PeerAuthentication{
				Selector: &istioapiv1beta1.WorkloadSelector{
					MatchLabels: map[string]string{
						"app": "test",
					},
				},
				Mtls: &istioapisecurityv1beta1.PeerAuthentication_MutualTLS{
					Mode: istioapisecurityv1beta1.PeerAuthentication_MutualTLS_STRICT,
				},
			},
			equal: true,
		},
		"match labels differs": {
			a: &istioapisecurityv1beta1.PeerAuthentication{
				Selector: &istioapiv1beta1.WorkloadSelector{
					MatchLabels: map[string]string{
						"app": "test-a",
					},
				},
				Mtls: &istioapisecurityv1beta1.PeerAuthentication_MutualTLS{
					Mode: istioapisecurityv1beta1.PeerAuthentication_MutualTLS_STRICT,
				},
			},
			b: &istioapisecurityv1beta1.PeerAuthentication{
				Selector: &istioapiv1beta1.WorkloadSelector{
					MatchLabels: map[string]string{
						"app": "test-b",
					},
				},
				Mtls: &istioapisecurityv1beta1.PeerAuthentication_MutualTLS{
					Mode: istioapisecurityv1beta1.PeerAuthentication_MutualTLS_STRICT,
				},
			},
			equal: false,
		},
		"mTLS mode differs": {
			a: &istioapisecurityv1beta1.PeerAuthentication{
				Selector: &istioapiv1beta1.WorkloadSelector{
					MatchLabels: map[string]string{
						"app": "test",
					},
				},
				Mtls: &istioapisecurityv1beta1.PeerAuthentication_MutualTLS{
					Mode: istioapisecurityv1beta1.PeerAuthentication_MutualTLS_STRICT,
				},
			},
			b: &istioapisecurityv1beta1.PeerAuthentication{
				Selector: &istioapiv1beta1.WorkloadSelector{
					MatchLabels: map[string]string{
						"app": "test",
					},
				},
				Mtls: &istioapisecurityv1beta1.PeerAuthentication_MutualTLS{
					Mode: istioapisecurityv1beta1.PeerAuthentication_MutualTLS_PERMISSIVE,
				},
			},
			equal: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			b := &istio.PeerAuthenticationBuilder{}
			result := b.Equal(test.a, test.b)
			assert.Equal(tt, test.equal, result)
		})
	}
}
