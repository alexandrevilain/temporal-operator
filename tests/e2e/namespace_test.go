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

	appsv1alpha1 "github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/pkg/temporal"
	"go.temporal.io/api/serviceerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestNamespaceCreation(t *testing.T) {
	var temporalCluster *appsv1alpha1.TemporalCluster
	var temporalNamespace *appsv1alpha1.TemporalNamespace

	namespaceFature := features.New("namespace creation using CRD").
		Setup(func(ctx context.Context, tt *testing.T, cfg *envconf.Config) context.Context {
			namespace := GetNamespaceForTest(ctx, t)

			var err error
			temporalCluster, err = deployAndWaitForTemporalWithPostgres(ctx, cfg, namespace)
			if err != nil {
				t.Fatal(err)
			}
			return ctx
		}).
		Assess("Temporal cluster created", func(ctx context.Context, tt *testing.T, cfg *envconf.Config) context.Context {
			err := waitForTemporalCluster(ctx, cfg, temporalCluster)
			if err != nil {
				t.Fatal(err)
			}
			return ctx
		}).
		Assess("Can create a temporal namespace", func(ctx context.Context, tt *testing.T, cfg *envconf.Config) context.Context {
			namespace := GetNamespaceForTest(ctx, t)

			// create the temporal cluster client
			temporalNamespace = &appsv1alpha1.TemporalNamespace{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
				Spec: appsv1alpha1.TemporalNamespaceSpec{
					TemporalClusterRef: corev1.LocalObjectReference{
						Name: temporalCluster.GetName(),
					},
					RetentionPeriod: &metav1.Duration{Duration: 24 * time.Hour},
				},
			}
			err := cfg.Client().Resources(namespace).Create(ctx, temporalNamespace)
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Namespace exists", func(ctx context.Context, tt *testing.T, cfg *envconf.Config) context.Context {
			connectAddr, closePortForward, err := forwardPortToTemporalFrontend(ctx, cfg, temporalCluster)
			if err != nil {
				t.Fatal(err)
			}
			defer closePortForward()

			client, err := klientToControllerRuntimeClient(cfg.Client())
			if err != nil {
				t.Fatal(err)
			}

			nsClient, err := temporal.GetClusterNamespaceClient(ctx, client, temporalCluster, temporal.WithHostPort(connectAddr))
			if err != nil {
				t.Fatal(err)
			}

			wait.For(func() (done bool, err error) {
				// If no error while describing the namespace, it works.
				_, err = nsClient.Describe(ctx, temporalNamespace.GetName())
				if err != nil {
					_, ok := err.(*serviceerror.NamespaceNotFound)
					if ok {
						return false, nil
					}
					return false, err
				}

				return true, nil
			}, wait.WithTimeout(5*time.Minute), wait.WithInterval(5*time.Second))
			return ctx
		}).
		Feature()

	testenv.Test(t, namespaceFature)
}
