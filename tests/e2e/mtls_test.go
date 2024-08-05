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

package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestWithmTLSEnabled(t *testing.T) {
	tests := map[string]struct {
		deployDependencies func(ctx context.Context, cfg *envconf.Config, namespace string) error
		cluster            func(ctx context.Context, cfg *envconf.Config, namespace string) *v1beta1.TemporalCluster
	}{
		"mTLS enabled with cert-manager": {
			deployDependencies: deployAndWaitForPostgres,
			cluster: func(_ context.Context, _ *envconf.Config, namespace string) *v1beta1.TemporalCluster {
				connectAddr := fmt.Sprintf("postgres.%s:5432", namespace)
				return &v1beta1.TemporalCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: namespace,
					},
					Spec: v1beta1.TemporalClusterSpec{
						NumHistoryShards:           1,
						JobTTLSecondsAfterFinished: &jobTTL,
						MTLS: &v1beta1.MTLSSpec{
							Provider: v1beta1.CertManagerMTLSProvider,
							Internode: &v1beta1.InternodeMTLSSpec{
								Enabled: true,
							},
							Frontend: &v1beta1.FrontendMTLSSpec{
								Enabled: true,
							},
						},
						Persistence: v1beta1.TemporalPersistenceSpec{
							DefaultStore: &v1beta1.DatastoreSpec{
								SQL: &v1beta1.SQLSpec{
									User:            "temporal",
									PluginName:      "postgres12",
									DatabaseName:    "temporal",
									ConnectAddr:     connectAddr,
									ConnectProtocol: "tcp",
								},
								PasswordSecretRef: &v1beta1.SecretKeyReference{
									Name: "postgres-password",
									Key:  "PASSWORD",
								},
							},
							VisibilityStore: &v1beta1.DatastoreSpec{
								SQL: &v1beta1.SQLSpec{
									User:            "temporal",
									PluginName:      "postgres12",
									DatabaseName:    "temporal_visibility",
									ConnectAddr:     connectAddr,
									ConnectProtocol: "tcp",
								},
								PasswordSecretRef: &v1beta1.SecretKeyReference{
									Name: "postgres-password",
									Key:  "PASSWORD",
								},
							},
						},
					},
				}
			},
		},
		"mTLS enabled with cert-manager and internal frontend": {
			deployDependencies: deployAndWaitForPostgres,
			cluster: func(_ context.Context, _ *envconf.Config, namespace string) *v1beta1.TemporalCluster {
				connectAddr := fmt.Sprintf("postgres.%s:5432", namespace)
				return &v1beta1.TemporalCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: namespace,
					},
					Spec: v1beta1.TemporalClusterSpec{
						NumHistoryShards:           1,
						JobTTLSecondsAfterFinished: &jobTTL,
						MTLS: &v1beta1.MTLSSpec{
							Provider: v1beta1.CertManagerMTLSProvider,
							Internode: &v1beta1.InternodeMTLSSpec{
								Enabled: true,
							},
							Frontend: &v1beta1.FrontendMTLSSpec{
								Enabled: true,
							},
						},
						Services: &v1beta1.ServicesSpec{
							InternalFrontend: &v1beta1.InternalFrontendServiceSpec{
								Enabled: true,
							},
						},
						Persistence: v1beta1.TemporalPersistenceSpec{
							DefaultStore: &v1beta1.DatastoreSpec{
								SQL: &v1beta1.SQLSpec{
									User:            "temporal",
									PluginName:      "postgres12",
									DatabaseName:    "temporal",
									ConnectAddr:     connectAddr,
									ConnectProtocol: "tcp",
								},
								PasswordSecretRef: &v1beta1.SecretKeyReference{
									Name: "postgres-password",
									Key:  "PASSWORD",
								},
							},
							VisibilityStore: &v1beta1.DatastoreSpec{
								SQL: &v1beta1.SQLSpec{
									User:            "temporal",
									PluginName:      "postgres12",
									DatabaseName:    "temporal_visibility",
									ConnectAddr:     connectAddr,
									ConnectProtocol: "tcp",
								},
								PasswordSecretRef: &v1beta1.SecretKeyReference{
									Name: "postgres-password",
									Key:  "PASSWORD",
								},
							},
						},
					},
				}
			},
		},
	}

	featureTable := []features.Feature{}

	for name, testCase := range tests {
		test := testCase
		feature := features.New(name).
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				namespace := GetNamespaceForFeature(ctx)
				t.Logf("using namespace: %s", namespace)

				err := test.deployDependencies(ctx, cfg, namespace)
				if err != nil {
					t.Fatal(err)
				}

				cluster := test.cluster(ctx, cfg, namespace)
				if err != nil {
					t.Fatal(err)
				}

				err = cfg.Client().Resources().Create(ctx, cluster)
				if err != nil {
					t.Fatal(err)
				}

				return SetTemporalClusterForFeature(ctx, cluster)
			}).
			Assess("Temporal cluster created", AssertTemporalClusterReady()).
			Assess("Can create a TemporalNamespace", AssertCanCreateTemporalNamespace("default")).
			Assess("TemporalNamespace ready", AssertTemporalNamespaceReady()).
			Assess("Can create a TemporalClusterClient", AssertCanCreateTemporalClusterClient()).
			Assess("TemporalClusterClient ready", AssertTemporalClusterClientReady()).
			Assess("Temporal cluster with mTLS can handle workflows", AssertTemporalClusterWithMTLSCanHandleWorkflows()).
			Feature()

		featureTable = append(featureTable, feature)
	}

	testenv.TestInParallel(t, featureTable...)
}
