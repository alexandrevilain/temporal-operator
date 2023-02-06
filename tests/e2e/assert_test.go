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

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/temporal"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/alexandrevilain/temporal-operator/tests/e2e/temporal/teststarter"
	"github.com/alexandrevilain/temporal-operator/tests/e2e/temporal/testworker"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func AssertClusterReady() features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		cluster := ctx.Value(clusterKey).(*v1beta1.TemporalCluster)
		err := waitForCluster(ctx, cfg, cluster)
		if err != nil {
			t.Fatal(err)
		}
		return ctx
	}
}

func AssertClusterCanBeUpgraded(v string) features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		cluster := ctx.Value(clusterKey).(*v1beta1.TemporalCluster)

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
		cluster := ctx.Value(clusterKey).(*v1beta1.TemporalCluster)
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
		w.Start()
		defer w.Stop()

		t.Logf("Starting workflow")
		err = teststarter.NewStarter(clusterClient).StartGreetingWorkflow()
		if err != nil {
			t.Fatal(err)
		}

		return ctx
	}
}
