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
			cluster, err = deployAndWaitForTemporalWithPostgres(ctx, cfg, namespace, "1.23.0")
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

			err = wait.For(func(ctx context.Context) (done bool, err error) {
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
			err = wait.For(func(ctx context.Context) (done bool, err error) {
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
		// Assess("Namespace delete in temporal", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		// 	connectAddr, closePortForward, err := forwardPortToTemporalFrontend(ctx, cfg, t, cluster)
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}
		// 	defer closePortForward()

		// 	client := cfg.Client().Resources().GetControllerRuntimeClient()

		// 	tempClient, err := temporal.GetClusterClient(ctx, client, cluster, temporal.WithHostPort(connectAddr))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	// Wait for the client to return NamespaceNotFound error.
		// 	err = wait.For(func() (done bool, err error) {
		// 		list, err := tempClient.WorkflowService().ListNamespaces(ctx, &workflowservice.ListNamespacesRequest{
		// 			PageSize: 10,
		// 			NamespaceFilter: &namespace.NamespaceFilter{
		// 				IncludeDeleted: true,
		// 			},
		// 		})
		// 		if err != nil {
		// 			return false, err
		// 		}

		// 		t.Logf("Retrieved %d namespaces", len(list.Namespaces))

		// 		for _, namespace := range list.Namespaces {
		// 			if namespace.NamespaceInfo.Name == temporalNamespace.GetName() {
		// 				t.Logf("Found '%s' namespace: %s", temporalNamespace.GetName(), namespace.NamespaceInfo.GetState())
		// 				return namespace.NamespaceInfo.GetState() == enums.NAMESPACE_STATE_DELETED, nil
		// 			}
		// 		}

		// 		t.Logf("Namespace '%s' not found", temporalNamespace.GetName())

		// 		return true, nil
		// 	}, wait.WithTimeout(5*time.Minute), wait.WithInterval(5*time.Second))
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}

		// 	return ctx
		// }).
		Feature()

	testenv.Test(t, namespaceFature)
}

func TestNamespaceDeletionWhenClusterDoesNotExist(rt *testing.T) {
	var temporalClusterName string

	feature := features.New("namespace can be deleted when temporal cluster does not exist").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			namespace := GetNamespaceForFeature(ctx)

			temporalClusterName = "does-not-exist"

			// create TemporalNamespace
			temporalNamespace := &v1beta1.TemporalNamespace{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
				Spec: v1beta1.TemporalNamespaceSpec{
					ClusterRef: v1beta1.TemporalReference{
						Name: temporalClusterName,
					},
					RetentionPeriod: &metav1.Duration{Duration: 24 * time.Hour},
				},
			}
			err := c.Client().Resources(namespace).Create(ctx, temporalNamespace)
			if err != nil {
				t.Fatal(err)
			}
			return SetTemporalNamespaceForFeature(ctx, temporalNamespace)
		}).
		Assess("TemporalCluster does not exist", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var temporalCluster = &v1beta1.TemporalCluster{}
			err := c.Client().Resources().Get(ctx, temporalClusterName, GetNamespaceForFeature(ctx), temporalCluster)
			if err == nil {
				t.Fatalf("found cluster: %v", temporalCluster)
			}

			return ctx
		}).
		Assess("TemporalNamespace can be deleted", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			err := c.Client().Resources(GetNamespaceForFeature(ctx)).Delete(ctx, GetTemporalNamespaceForFeature(ctx))
			if err != nil {
				t.Fatalf("failed to delete namespace: %v", err)
			}
			return ctx
		}).
		Feature()

	testenv.Test(rt, feature)
}

func TestNamespaceDeletionWhenClusterDeleted(rt *testing.T) {
	feature := features.New("namespace can be deleted after a temporal cluster associated with it is also deleted").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// create TemporalCluster
			namespace := GetNamespaceForFeature(ctx)

			cluster, err := deployAndWaitForTemporalWithPostgres(ctx, c, namespace, "1.19.1")
			if err != nil {
				t.Fatal(err)
			}
			return SetTemporalClusterForFeature(ctx, cluster)
		}).
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// create TemporalNamespace
			namespace := GetNamespaceForFeature(ctx)
			cluster := GetTemporalClusterForFeature(ctx)

			temporalNamespace := &v1beta1.TemporalNamespace{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
				Spec: v1beta1.TemporalNamespaceSpec{
					ClusterRef: v1beta1.TemporalReference{
						Name: cluster.GetName(),
					},
					RetentionPeriod: &metav1.Duration{Duration: 24 * time.Hour},
				},
			}
			err := c.Client().Resources(namespace).Create(ctx, temporalNamespace)
			if err != nil {
				t.Fatal(err)
			}
			return SetTemporalNamespaceForFeature(ctx, temporalNamespace)
		}).
		Assess("TemporalCluster can be deleted", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			err := c.Client().Resources(GetNamespaceForFeature(ctx)).Delete(ctx, GetTemporalClusterForFeature(ctx))
			if err != nil {
				t.Fatalf("failed to delete: %v", err)
			}
			return ctx
		}).
		Assess("TemporalNamespace can be deleted", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			err := c.Client().Resources(GetNamespaceForFeature(ctx)).Delete(ctx, GetTemporalNamespaceForFeature(ctx))
			if err != nil {
				t.Fatalf("failed to delete: %v", err)
			}
			return ctx
		}).
		Feature()

	testenv.Test(rt, feature)
}
