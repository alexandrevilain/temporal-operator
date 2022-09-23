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
	"github.com/alexandrevilain/temporal-operator/pkg/temporal"
	"github.com/alexandrevilain/temporal-operator/tests/e2e/temporal/teststarter"
	"github.com/alexandrevilain/temporal-operator/tests/e2e/temporal/testworker"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestWithmTLSEnabled(t *testing.T) {
	var cluster *v1beta1.Cluster
	var clusterClient *v1beta1.ClusterClient

	mTLSCertManagerFeature := features.New("mTLS enabled with cert-manager").
		Setup(func(ctx context.Context, tt *testing.T, cfg *envconf.Config) context.Context {
			namespace := GetNamespaceForFeature(ctx)

			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			// create the postgres
			err = deployAndWaitForPostgres(ctx, cfg, namespace)
			if err != nil {
				t.Fatal(err)
			}

			connectAddr := fmt.Sprintf("postgres.%s:5432", namespace)

			// create the temporal cluster
			cluster = &v1beta1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: namespace,
				},
				Spec: v1beta1.ClusterSpec{
					NumHistoryShards: 1,
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
						DefaultStore:    "default",
						VisibilityStore: "visibility",
					},
					Datastores: []v1beta1.TemporalDatastoreSpec{
						{
							Name: "default",
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
						{
							Name: "visibility",
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
			err = client.Resources(namespace).Create(ctx, cluster)
			if err != nil {
				t.Fatal(err)
			}

			return context.WithValue(ctx, clusterKey, cluster)
		}).
		Assess("Temporal cluster created", AssertClusterReady()).
		Assess("Can create a temporal cluster namespace", func(ctx context.Context, tt *testing.T, cfg *envconf.Config) context.Context {
			namespace := GetNamespaceForFeature(ctx)

			// create the temporal cluster client
			clusterClient = &v1beta1.ClusterClient{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
				Spec: v1beta1.ClusterClientSpec{
					ClusterRef: corev1.LocalObjectReference{
						Name: cluster.GetName(),
					},
				},
			}
			err := cfg.Client().Resources(namespace).Create(ctx, clusterClient)
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Temporal cluster client created", func(ctx context.Context, tt *testing.T, cfg *envconf.Config) context.Context {
			err := waitForClusterClient(ctx, cfg, clusterClient)
			if err != nil {
				t.Fatal(err)
			}
			return ctx

		}).
		Assess("Temporal cluster can handle workflows", func(ctx context.Context, tt *testing.T, cfg *envconf.Config) context.Context {
			namespace := GetNamespaceForFeature(ctx)

			connectAddr, closePortForward, err := forwardPortToTemporalFrontend(ctx, cfg, t, cluster)
			if err != nil {
				t.Fatal(err)
			}
			defer closePortForward()

			t.Logf("Temporal frontend addr: %s", connectAddr)

			client, err := klientToControllerRuntimeClient(cfg.Client())
			if err != nil {
				t.Fatal(err)
			}

			clientSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterClient.Status.SecretRef.Name,
					Namespace: namespace,
				},
			}

			list := &corev1.SecretList{Items: []corev1.Secret{*clientSecret}}

			err = wait.For(conditions.New(cfg.Client().Resources(namespace)).ResourcesFound(list))
			if err != nil {
				t.Fatal(err)
			}

			err = cfg.Client().Resources(namespace).Get(ctx, clusterClient.Status.SecretRef.Name, namespace, clientSecret)
			if err != nil {
				t.Fatal(err)
			}

			tlsCfg, err := temporal.GetTlSConfigFromSecret(clientSecret)
			if err != nil {
				t.Fatal(err)
			}
			tlsCfg.ServerName = clusterClient.Status.ServerName

			clusterClient, err := temporal.GetClusterClient(ctx, client, cluster, temporal.WithHostPort(connectAddr), temporal.WithTLSConfig(tlsCfg))
			if err != nil {
				t.Fatal(err)
			}

			w, err := testworker.NewWorker(clusterClient)
			if err != nil {
				t.Fatal(err)
			}

			t.Log("Starting test worker")
			w.Start()
			defer w.Stop()

			t.Logf("Starting workflow")
			err = teststarter.NewStarter(clusterClient).StartGreetingWorkflow()
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Feature()

	testenv.Test(t, mTLSCertManagerFeature)
}
