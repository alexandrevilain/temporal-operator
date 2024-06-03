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
	"testing"
	"time"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/temporal"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/alexandrevilain/temporal-operator/tests/e2e/temporal/teststarter"
	"github.com/alexandrevilain/temporal-operator/tests/e2e/temporal/testworker"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func AssertTemporalClusterReady() features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		cluster := GetTemporalClusterForFeature(ctx)

		err := waitForCluster(ctx, cfg, cluster)
		if err != nil {
			t.Fatal(err)
		}
		return ctx
	}
}

func AssertTemporalClusterCanBeUpgraded(v string) features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		cluster := GetTemporalClusterForFeature(ctx)

		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			err := cfg.Client().Resources(cluster.GetNamespace()).Get(ctx, cluster.GetName(), cluster.GetNamespace(), cluster)
			if err != nil {
				return err
			}

			// Set the new version
			cluster.Spec.Version = version.MustNewVersionFromString(v)

			err = cfg.Client().Resources(cluster.GetNamespace()).Update(ctx, cluster)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		return ctx
	}
}

func AssertClusterCanHandleWorkflows() features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		cluster := GetTemporalClusterForFeature(ctx)
		connectAddr, closePortForward, err := forwardPortToTemporalFrontend(ctx, cfg, t, cluster)
		if err != nil {
			t.Fatal(err)
		}
		defer closePortForward()

		t.Logf("Temporal frontend addr: %s", connectAddr)

		client := cfg.Client().Resources().GetControllerRuntimeClient()

		clusterClient, err := temporal.GetClusterClient(ctx, client, cluster, temporal.WithHostPort(connectAddr))
		if err != nil {
			t.Fatal(err)
		}

		w, err := testworker.NewWorker(clusterClient)
		if err != nil {
			t.Fatal(err)
		}

		t.Log("Starting test worker")
		err = w.Start()
		if err != nil {
			t.Fatal(err)
		}

		defer w.Stop()

		t.Logf("Starting workflow")
		err = teststarter.NewStarter(clusterClient).StartGreetingWorkflow()
		if err != nil {
			t.Fatal(err)
		}

		return ctx
	}
}

func AssertTemporalClusterWithMTLSCanHandleWorkflows() features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		namespace := GetNamespaceForFeature(ctx)
		cluster := GetTemporalClusterForFeature(ctx)
		clusterClient := GetTemporalClusterClientForFeature(ctx)

		connectAddr, closePortForward, err := forwardPortToTemporalFrontend(ctx, cfg, t, cluster)
		if err != nil {
			t.Fatal(err)
		}
		defer closePortForward()

		t.Logf("Temporal frontend addr: %s", connectAddr)

		client := cfg.Client().Resources().GetControllerRuntimeClient()

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

		temporalClient, err := temporal.GetClusterClient(ctx, client, cluster, temporal.WithHostPort(connectAddr), temporal.WithTLSConfig(tlsCfg))
		if err != nil {
			t.Fatal(err)
		}

		w, err := testworker.NewWorker(temporalClient)
		if err != nil {
			t.Fatal(err)
		}

		t.Log("Starting test worker")
		err = w.Start()
		if err != nil {
			t.Fatal(err)
		}

		defer w.Stop()

		t.Logf("Starting workflow")
		err = teststarter.NewStarter(temporalClient).StartGreetingWorkflow()
		if err != nil {
			t.Fatal(err)
		}

		return ctx
	}
}

func AssertCanCreateTemporalClusterClient() features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		namespace := GetNamespaceForFeature(ctx)
		cluster := GetTemporalClusterForFeature(ctx)

		// create the temporal cluster client
		clusterClient := &v1beta1.TemporalClusterClient{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
			Spec: v1beta1.TemporalClusterClientSpec{
				ClusterRef: v1beta1.TemporalReference{
					Name: cluster.GetName(),
				},
			},
		}
		err := cfg.Client().Resources(namespace).Create(ctx, clusterClient)
		if err != nil {
			t.Fatal(err)
		}

		return SetTemporalClusterClientForFeature(ctx, clusterClient)
	}
}

func AssertTemporalClusterClientReady() features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		clusterClient := GetTemporalClusterClientForFeature(ctx)

		err := waitForClusterClient(ctx, cfg, clusterClient)
		if err != nil {
			t.Fatal(err)
		}
		return ctx
	}
}

func AssertCanCreateTemporalNamespace(name string) features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		namespace := GetNamespaceForFeature(ctx)
		cluster := GetTemporalClusterForFeature(ctx)

		// create the temporal cluster client
		temporalNamespace := &v1beta1.TemporalNamespace{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			Spec: v1beta1.TemporalNamespaceSpec{
				ClusterRef: v1beta1.TemporalReference{
					Name: cluster.GetName(),
				},
				RetentionPeriod: &metav1.Duration{Duration: 24 * time.Hour},
			},
		}
		err := cfg.Client().Resources(namespace).Create(ctx, temporalNamespace)
		if err != nil {
			t.Fatal(err)
		}

		return SetTemporalNamespaceForFeature(ctx, temporalNamespace)
	}
}

func AssertTemporalNamespaceReady() features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		temporalNamespace := GetTemporalNamespaceForFeature(ctx)

		cond := conditions.New(cfg.Client().Resources()).ResourceMatch(temporalNamespace, func(object k8s.Object) bool {
			for _, condition := range object.(*v1beta1.TemporalNamespace).Status.Conditions {
				if condition.Type == v1beta1.ReadyCondition && condition.Status == metav1.ConditionTrue {
					return true
				}
			}
			return false
		})

		err := wait.For(cond, wait.WithTimeout(time.Minute*10))
		if err != nil {
			t.Fatal(err)
		}

		// Wait 15s, according to the Temporal documentation:
		// "Note that registering a Namespace takes up to 15 seconds to complete".
		time.Sleep(15 * time.Second)

		return ctx
	}
}
