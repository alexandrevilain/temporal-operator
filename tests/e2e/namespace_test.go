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
	"errors"
	"testing"
	"time"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/temporal"
	"go.temporal.io/api/serviceerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestNamespaceCreation(t *testing.T) {
	var cluster *v1beta1.TemporalCluster
	var temporalNamespace *v1beta1.TemporalNamespace

	namespaceFature := features.New("namespace creation using CRD").
		Setup(func(ctx context.Context, tt *testing.T, cfg *envconf.Config) context.Context {
			namespace := GetNamespaceForFeature(ctx)

			var err error
			cluster, err = deployAndWaitForTemporalWithPostgres(ctx, cfg, namespace, "1.19.1")
			if err != nil {
				t.Fatal(err)
			}
			return SetTemporalClusterForFeature(ctx, cluster)
		}).
		Assess("Temporal cluster created", AssertTemporalClusterReady()).
		Assess("Can create a temporal namespace", func(ctx context.Context, tt *testing.T, cfg *envconf.Config) context.Context {
			namespace := GetNamespaceForFeature(ctx)

			// create the temporal cluster client
			temporalNamespace = &v1beta1.TemporalNamespace{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
				Spec: v1beta1.TemporalNamespaceSpec{
					ClusterRef: corev1.LocalObjectReference{
						Name: cluster.GetName(),
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
			connectAddr, closePortForward, err := forwardPortToTemporalFrontend(ctx, cfg, t, cluster)
			if err != nil {
				t.Fatal(err)
			}
			defer closePortForward()

			client := cfg.Client().Resources().GetControllerRuntimeClient()

			nsClient, err := temporal.GetClusterNamespaceClient(ctx, client, cluster, temporal.WithHostPort(connectAddr))
			if err != nil {
				t.Fatal(err)
			}

			err = wait.For(func() (done bool, err error) {
				// If no error while describing the namespace, it works.
				_, err = nsClient.Describe(ctx, temporalNamespace.GetName())
				if err != nil {
					var namespaceNotFoundError *serviceerror.NamespaceNotFound
					if errors.As(err, &namespaceNotFoundError) {
						return false, nil
					}

					return false, err
				}

				return true, nil
			}, wait.WithTimeout(5*time.Minute), wait.WithInterval(5*time.Second))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Namespace can be deleted", func(ctx context.Context, tt *testing.T, cfg *envconf.Config) context.Context {
			namespace := GetNamespaceForFeature(ctx)

			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				err := cfg.Client().Resources(namespace).Get(ctx, temporalNamespace.GetName(), temporalNamespace.GetNamespace(), temporalNamespace)
				if err != nil {
					return err
				}

				temporalNamespace.Spec.AllowDeletion = true
				return cfg.Client().Resources(namespace).Update(ctx, temporalNamespace)
			})
			if err != nil {
				t.Fatal(err)
			}

			// Wait for controller to set finalizer.
			err = wait.For(func() (done bool, err error) {
				err = cfg.Client().Resources(namespace).Get(ctx, temporalNamespace.GetName(), temporalNamespace.GetNamespace(), temporalNamespace)
				if err != nil {
					t.Fatal(err)
				}

				result := controllerutil.ContainsFinalizer(temporalNamespace, "deletion.finalizers.temporal.io")
				return result, nil
			}, wait.WithTimeout(2*time.Minute), wait.WithInterval(1*time.Second))
			if err != nil {
				t.Fatal(err)
			}

			err = cfg.Client().Resources(namespace).Delete(ctx, temporalNamespace)
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Namespace delete in temporal", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			connectAddr, closePortForward, err := forwardPortToTemporalFrontend(ctx, cfg, t, cluster)
			if err != nil {
				t.Fatal(err)
			}
			defer closePortForward()

			client := cfg.Client().Resources().GetControllerRuntimeClient()

			nsClient, err := temporal.GetClusterNamespaceClient(ctx, client, cluster, temporal.WithHostPort(connectAddr))
			if err != nil {
				t.Fatal(err)
			}

			// Wait for the client to return NamespaceNotFound error.
			err = wait.For(func() (done bool, err error) {
				_, err = nsClient.Describe(ctx, temporalNamespace.GetName())
				if err != nil {
					var namespaceNotFoundError *serviceerror.NamespaceNotFound
					if errors.As(err, &namespaceNotFoundError) {
						return true, nil
					}
				}

				return false, nil
			}, wait.WithTimeout(5*time.Minute), wait.WithInterval(5*time.Second))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Feature()

	testenv.Test(t, namespaceFature)
}
