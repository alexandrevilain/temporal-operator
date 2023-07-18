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

package webhooks_test

import (
	"context"
	"testing"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/discovery"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/alexandrevilain/temporal-operator/webhooks"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
)

func TestDefault(t *testing.T) {
	tests := map[string]struct {
		initialObject  runtime.Object
		expectedObject runtime.Object
		expectedErr    string
	}{
		"default fields": {
			initialObject: &v1beta1.TemporalCluster{
				TypeMeta: v1beta1.TemporalClusterTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
			},
			expectedObject: func() runtime.Object {
				c := &v1beta1.TemporalCluster{
					TypeMeta: v1beta1.TemporalClusterTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name: "fake",
					},
				}
				c.Default()
				return c
			}(),
			expectedErr: "",
		},
		"deprecated fields: prometheus listen address": {
			initialObject: &v1beta1.TemporalCluster{
				TypeMeta: v1beta1.TemporalClusterTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec: v1beta1.TemporalClusterSpec{
					Metrics: &v1beta1.MetricsSpec{
						Enabled: true,
						Prometheus: &v1beta1.PrometheusSpec{
							ListenAddress: "localhost:8080",
						},
					},
				},
			},
			expectedObject: func() runtime.Object {
				c := &v1beta1.TemporalCluster{
					TypeMeta: v1beta1.TemporalClusterTypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name: "fake",
					},
					Spec: v1beta1.TemporalClusterSpec{
						Metrics: &v1beta1.MetricsSpec{
							Enabled: true,
							Prometheus: &v1beta1.PrometheusSpec{
								ListenPort: pointer.Int32(8080),
							},
						},
					},
				}
				c.Default()
				return c
			}(),
		},
		"bad port on listen address": {
			initialObject: &v1beta1.TemporalCluster{
				TypeMeta: v1beta1.TemporalClusterTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec: v1beta1.TemporalClusterSpec{
					Metrics: &v1beta1.MetricsSpec{
						Enabled: true,
						Prometheus: &v1beta1.PrometheusSpec{
							ListenAddress: "localhost:abc",
						},
					},
				},
			},
			expectedErr: "can't parse prometheus spec.metrics.prometheus.listenAddress port: strconv.ParseInt: parsing \"abc\": invalid syntax",
		},
		"no port on listen address": {
			initialObject: &v1beta1.TemporalCluster{
				TypeMeta: v1beta1.TemporalClusterTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec: v1beta1.TemporalClusterSpec{
					Metrics: &v1beta1.MetricsSpec{
						Enabled: true,
						Prometheus: &v1beta1.PrometheusSpec{
							ListenAddress: "localhost",
						},
					},
				},
			},
			expectedErr: "can't parse prometheus spec.metrics.prometheus.listenAddress: address localhost: missing port in address",
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			wh := &webhooks.TemporalClusterWebhook{}

			err := wh.Default(context.Background(), test.initialObject)
			if test.expectedErr != "" {
				assert.Error(tt, err)
				assert.Equal(tt, test.expectedErr, err.Error())
			} else {
				assert.NoError(tt, err)
				assert.EqualValues(tt, test.expectedObject, test.initialObject)
			}
		})
	}
}

func TestValidateCreate(t *testing.T) {
	tests := map[string]struct {
		object      runtime.Object
		wh          *webhooks.TemporalClusterWebhook
		expectedErr string
	}{
		"works with basic object": {
			object: &v1beta1.TemporalCluster{
				TypeMeta: v1beta1.TemporalClusterTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.18.4"),
				},
			},
			wh: &webhooks.TemporalClusterWebhook{
				AvailableAPIs: &discovery.AvailableAPIs{
					Istio:              true,
					CertManager:        true,
					PrometheusOperator: true,
				},
			},
		},
		"error with version not supported": {
			object: &v1beta1.TemporalCluster{
				TypeMeta: v1beta1.TemporalClusterTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("4560.18.4"),
				},
			},
			wh: &webhooks.TemporalClusterWebhook{
				AvailableAPIs: &discovery.AvailableAPIs{},
			},
			expectedErr: "TemporalCluster.temporal.io \"fake\" is invalid: spec.version: Forbidden: Unsupported temporal version",
		},
		"error with version marked as broken": {
			object: &v1beta1.TemporalCluster{
				TypeMeta: v1beta1.TemporalClusterTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.21.0"),
				},
			},
			wh: &webhooks.TemporalClusterWebhook{
				AvailableAPIs: &discovery.AvailableAPIs{},
			},
			expectedErr: "TemporalCluster.temporal.io \"fake\" is invalid: spec.version: Forbidden: version 1.21.0 is marked as broken by the operator, please upgrade to 1.21.1 (if allowed)",
		},
		"error when no cert manager and mTLS with cert-manager enabled": {
			object: &v1beta1.TemporalCluster{
				TypeMeta: v1beta1.TemporalClusterTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.18.4"),
					MTLS: &v1beta1.MTLSSpec{
						Provider: v1beta1.CertManagerMTLSProvider,
						Internode: &v1beta1.InternodeMTLSSpec{
							Enabled: true,
						},
						Frontend: &v1beta1.FrontendMTLSSpec{
							Enabled: true,
						},
					},
				},
			},
			wh: &webhooks.TemporalClusterWebhook{
				AvailableAPIs: &discovery.AvailableAPIs{
					Istio:              false,
					CertManager:        false,
					PrometheusOperator: false,
				},
			},
			expectedErr: "TemporalCluster.temporal.io \"fake\" is invalid: spec.mTLS.provider: Invalid value: \"cert-manager\": Can't use cert-manager as mTLS provider as it's not available in the cluster",
		},
		"error with old elastic search version": {
			object: &v1beta1.TemporalCluster{
				TypeMeta: v1beta1.TemporalClusterTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.18.4"),
					Persistence: v1beta1.TemporalPersistenceSpec{
						AdvancedVisibilityStore: &v1beta1.DatastoreSpec{
							Elasticsearch: &v1beta1.ElasticsearchSpec{
								Version: "v6",
							},
						},
					},
				},
			},
			wh: &webhooks.TemporalClusterWebhook{
				AvailableAPIs: &discovery.AvailableAPIs{
					Istio:              false,
					CertManager:        false,
					PrometheusOperator: false,
				},
			},
			expectedErr: "TemporalCluster.temporal.io \"fake\" is invalid: spec.persistence.advancedVisibilityStore.elasticsearch.version: Forbidden: temporal cluster version >= 1.18.0 doesn't support ElasticSearch v6",
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			_, err := test.wh.ValidateCreate(context.Background(), test.object)
			if test.expectedErr != "" {
				assert.Error(tt, err)
				// Here contains is used to allow partial error checks.
				// When version is wrong, it shows which versions are supported.
				// If using "Equal" it would requires us to update tests
				// for each new temporal version.
				assert.Contains(tt, err.Error(), test.expectedErr)
			} else {
				assert.NoError(tt, err)
			}
		})
	}
}

func TestValidateUpdate(t *testing.T) {
	tests := map[string]struct {
		oldlObject  runtime.Object
		newObject   runtime.Object
		expectedErr string
	}{
		"allowed upgrade": {
			oldlObject: &v1beta1.TemporalCluster{
				TypeMeta: v1beta1.TemporalClusterTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.18.4"),
				},
			},
			newObject: &v1beta1.TemporalCluster{
				TypeMeta: v1beta1.TemporalClusterTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.19.0"),
				},
			},
		},
		"version rollback": {
			oldlObject: &v1beta1.TemporalCluster{
				TypeMeta: v1beta1.TemporalClusterTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.19.0"),
				},
			},
			newObject: &v1beta1.TemporalCluster{
				TypeMeta: v1beta1.TemporalClusterTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec:   v1beta1.TemporalClusterSpec{Version: version.MustNewVersionFromString("1.18.4")},
				Status: v1beta1.TemporalClusterStatus{},
			},
			expectedErr: "TemporalCluster.temporal.io \"fake\" is invalid: spec.version: Forbidden: Unauthorized version upgrade. Only sequential version upgrades are allowed (from v1.n.x to v1.n+1.x)",
		},
		"not a sequential version update": {
			oldlObject: &v1beta1.TemporalCluster{
				TypeMeta: v1beta1.TemporalClusterTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.17.0"),
				},
			},
			newObject: &v1beta1.TemporalCluster{
				TypeMeta: v1beta1.TemporalClusterTypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake",
				},
				Spec:   v1beta1.TemporalClusterSpec{Version: version.MustNewVersionFromString("1.19.4")},
				Status: v1beta1.TemporalClusterStatus{},
			},
			expectedErr: "TemporalCluster.temporal.io \"fake\" is invalid: spec.version: Forbidden: Unauthorized version upgrade. Only sequential version upgrades are allowed (from v1.n.x to v1.n+1.x)",
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			wh := &webhooks.TemporalClusterWebhook{}
			_, err := wh.ValidateUpdate(context.Background(), test.oldlObject, test.newObject)
			if test.expectedErr != "" {
				assert.Error(tt, err)
				assert.Equal(tt, test.expectedErr, err.Error())
			} else {
				assert.NoError(tt, err)
			}
		})
	}
}
