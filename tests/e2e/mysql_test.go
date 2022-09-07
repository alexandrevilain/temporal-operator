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

	appsv1alpha1 "github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestWithMySQLPersistence(t *testing.T) {
	mysqlFeature := features.New("MySQL for persistence").
		Setup(func(ctx context.Context, tt *testing.T, cfg *envconf.Config) context.Context {
			namespace := GetNamespaceForTest(ctx, t)
			t.Logf("using %s", namespace)

			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			// create the database
			err = deployAndWaitForMySQL(ctx, cfg, namespace)
			if err != nil {
				t.Fatal(err)
			}

			connectAddr := fmt.Sprintf("mysql.%s:3306", namespace)

			// create the temporal cluster
			temporalCluster := &appsv1alpha1.TemporalCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: namespace,
				},
				Spec: appsv1alpha1.TemporalClusterSpec{
					NumHistoryShards: 1,
					Version:          "1.16.3",
					Persistence: appsv1alpha1.TemporalPersistenceSpec{
						DefaultStore:    "default",
						VisibilityStore: "visibility",
					},
					Datastores: []appsv1alpha1.TemporalDatastoreSpec{
						{
							Name: "default",
							SQL: &appsv1alpha1.SQLSpec{
								User:            "temporal",
								PluginName:      "mysql",
								DatabaseName:    "temporal",
								ConnectAddr:     connectAddr,
								ConnectProtocol: "tcp",
								ConnectAttributes: map[string]string{
									"tx_isolation": "READ-COMMITTED",
								},
							},
							PasswordSecretRef: appsv1alpha1.SecretKeyReference{
								Name: "mysql-password",
								Key:  "PASSWORD",
							},
						},
						{
							Name: "visibility",
							SQL: &appsv1alpha1.SQLSpec{
								User:            "temporal",
								PluginName:      "mysql",
								DatabaseName:    "temporal_visibility",
								ConnectAddr:     connectAddr,
								ConnectProtocol: "tcp",
								ConnectAttributes: map[string]string{
									"tx_isolation": "READ-COMMITTED",
								},
							},
							PasswordSecretRef: appsv1alpha1.SecretKeyReference{
								Name: "mysql-password",
								Key:  "PASSWORD",
							},
						},
					},
				},
			}
			err = client.Resources(namespace).Create(ctx, temporalCluster)
			if err != nil {
				t.Fatal(err)
			}
			return context.WithValue(ctx, "cluster", temporalCluster)
		}).
		Assess("Temporal cluster created", AssertClusterReady()).
		Assess("Temporal cluster can handle workflows", AssertClusterCanHandleWorkflows()).
		Assess("Upgrade cluster", AssertClusterCanBeUpgraded("1.17.5")).
		Assess("Temporal cluster ready after upgrade", AssertClusterReady()).
		Assess("Temporal cluster can handle workflows after upgrade", AssertClusterCanHandleWorkflows()).
		Feature()

	testenv.Test(t, mysqlFeature)
}
