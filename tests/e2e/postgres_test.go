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

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestWithPostgresPersistence(t *testing.T) {
	pgFeature := features.New("PostgreSQL for persistence").
		Setup(func(ctx context.Context, tt *testing.T, cfg *envconf.Config) context.Context {
			namespace := GetNamespaceForTest(ctx, t)
			t.Logf("using %s", namespace)

			var err error
			temporalCluster, err := deployAndWaitForTemporalWithPostgres(ctx, cfg, namespace, "1.16.0")
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

	testenv.Test(t, pgFeature)
}
