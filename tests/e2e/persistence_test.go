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
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var (
	initialClusterVersion       = "1.18.5"
	firstUpgradeClusterVersion  = "1.19.1"
	secondUpgradeClusterVersion = "1.20.0"
)

func TestPersistence(t *testing.T) {
	tests := map[string]struct {
		deployDependencies func(ctx context.Context, cfg *envconf.Config, namespace string) error
		cluster            func(ctx context.Context, cfg *envconf.Config, namespace string) *v1beta1.TemporalCluster
	}{
		"postgresql persistence": {
			deployDependencies: deployAndWaitForPostgres,
			cluster: func(ctx context.Context, cfg *envconf.Config, namespace string) *v1beta1.TemporalCluster {
				connectAddr := fmt.Sprintf("postgres.%s:5432", namespace) // create the temporal cluster

				return &v1beta1.TemporalCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: namespace,
					},
					Spec: v1beta1.TemporalClusterSpec{
						NumHistoryShards:           1,
						JobTtlSecondsAfterFinished: &jobTtl,
						Version:                    version.MustNewVersionFromString(initialClusterVersion),
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
									PluginName:      "postgres",
									DatabaseName:    "temporal",
									ConnectAddr:     connectAddr,
									ConnectProtocol: "tcp",
								},
								PasswordSecretRef: v1beta1.SecretKeyReference{
									Name: "postgres-password",
									Key:  "PASSWORD",
								},
							},
							VisibilityStore: &v1beta1.DatastoreSpec{
								SQL: &v1beta1.SQLSpec{
									User:            "temporal",
									PluginName:      "postgres",
									DatabaseName:    "temporal_visibility",
									ConnectAddr:     connectAddr,
									ConnectProtocol: "tcp",
								},
								PasswordSecretRef: v1beta1.SecretKeyReference{
									Name: "postgres-password",
									Key:  "PASSWORD",
								},
							},
						},
					},
				}
			},
		},
		"mysql persistence": {
			deployDependencies: deployAndWaitForMySQL,
			cluster: func(ctx context.Context, cfg *envconf.Config, namespace string) *v1beta1.TemporalCluster {
				connectAddr := fmt.Sprintf("mysql.%s:3306", namespace)

				return &v1beta1.TemporalCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: namespace,
					},
					Spec: v1beta1.TemporalClusterSpec{
						NumHistoryShards:           1,
						JobTtlSecondsAfterFinished: &jobTtl,
						Version:                    version.MustNewVersionFromString(initialClusterVersion),
						Persistence: v1beta1.TemporalPersistenceSpec{
							DefaultStore: &v1beta1.DatastoreSpec{
								SQL: &v1beta1.SQLSpec{
									User:            "temporal",
									PluginName:      "mysql",
									DatabaseName:    "temporal",
									ConnectAddr:     connectAddr,
									ConnectProtocol: "tcp",
									ConnectAttributes: map[string]string{
										"tx_isolation": "READ-COMMITTED",
									},
								},
								PasswordSecretRef: v1beta1.SecretKeyReference{
									Name: "mysql-password",
									Key:  "PASSWORD",
								},
							},
							VisibilityStore: &v1beta1.DatastoreSpec{
								SQL: &v1beta1.SQLSpec{
									User:            "temporal",
									PluginName:      "mysql",
									DatabaseName:    "temporal_visibility",
									ConnectAddr:     connectAddr,
									ConnectProtocol: "tcp",
									ConnectAttributes: map[string]string{
										"tx_isolation": "READ-COMMITTED",
									},
								},
								PasswordSecretRef: v1beta1.SecretKeyReference{
									Name: "mysql-password",
									Key:  "PASSWORD",
								},
							},
						},
					},
				}
			},
		},
		"cassandra persistence": {
			deployDependencies: deployAndWaitForCassandra,
			cluster: func(ctx context.Context, cfg *envconf.Config, namespace string) *v1beta1.TemporalCluster {
				connectAddr := fmt.Sprintf("cassandra.%s", namespace)

				return &v1beta1.TemporalCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: namespace,
					},
					Spec: v1beta1.TemporalClusterSpec{
						NumHistoryShards:           1,
						JobTtlSecondsAfterFinished: &jobTtl,
						Version:                    version.MustNewVersionFromString(initialClusterVersion),
						Persistence: v1beta1.TemporalPersistenceSpec{
							DefaultStore: &v1beta1.DatastoreSpec{
								Cassandra: &v1beta1.CassandraSpec{
									Hosts:      []string{connectAddr},
									User:       "temporal",
									Keyspace:   "temporal",
									Datacenter: "datacenter1",
								},
								PasswordSecretRef: v1beta1.SecretKeyReference{
									Name: "cassandra-password",
									Key:  "PASSWORD",
								},
							},
							VisibilityStore: &v1beta1.DatastoreSpec{
								Cassandra: &v1beta1.CassandraSpec{
									Hosts:      []string{connectAddr},
									User:       "temporal",
									Keyspace:   "temporal_visibility",
									Datacenter: "datacenter1",
								},
								PasswordSecretRef: v1beta1.SecretKeyReference{
									Name: "cassandra-password",
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
		// Temporary mysql e2e tests disabling
		if name == "mysql persistence" {
			continue
		}
		// Temporary mysql e2e tests disabling

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
			Assess("Temporal cluster can handle workflows", AssertClusterCanHandleWorkflows()).
			Assess("Upgrade cluster to 1.19.x", AssertTemporalClusterCanBeUpgraded(firstUpgradeClusterVersion)).
			Assess("Temporal cluster ready after upgrade", AssertTemporalClusterReady()).
			Assess("Temporal cluster can handle workflows after upgrade", AssertClusterCanHandleWorkflows()).
			Assess("Upgrade cluster to 1.20.x", AssertTemporalClusterCanBeUpgraded(secondUpgradeClusterVersion)).
			Assess("Temporal cluster ready after upgrade", AssertTemporalClusterReady()).
			Assess("Temporal cluster can handle workflows after upgrade", AssertClusterCanHandleWorkflows()).
			Feature()

		featureTable = append(featureTable, feature)
	}

	testenv.TestInParallel(t, featureTable...)
}
