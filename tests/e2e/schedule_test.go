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
	"go.temporal.io/api/workflowservice/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestScheduleCreation(t *testing.T) {
	var cluster *v1beta1.TemporalCluster

	scheduleFeature := features.New("schedule creation using CRD").
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
		Assess("Can create a TemporalNamespace", AssertCanCreateTemporalNamespace("default")).
		Assess("TemporalNamespace ready", AssertTemporalNamespaceReady()).
		Assess("Can create a temporal schedule", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			namespace := GetNamespaceForFeature(ctx)
			temporalNamespace := GetTemporalNamespaceForFeature(ctx)

			// create the temporal schedule
			temporalSchedule := &v1beta1.TemporalSchedule{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
				Spec: v1beta1.TemporalScheduleSpec{
					NamespaceRef: v1beta1.TemporalReference{
						Name: temporalNamespace.GetName(),
					},
					Schedule: v1beta1.Schedule{
						Action: v1beta1.ScheduleAction{
							Workflow: v1beta1.ScheduleWorkflowAction{
								TaskQueue:    "queue",
								WorkflowType: "workflow",
							},
						},
						Spec: v1beta1.ScheduleSpec{
							Intervals: []v1beta1.ScheduleIntervalSpec{
								{
									Every: metav1.Duration{Duration: time.Second},
								},
							},
						},
					},
				},
			}
			err := cfg.Client().Resources(namespace).Create(ctx, temporalSchedule)
			if err != nil {
				t.Fatal(err)
			}

			return SetTemporalScheduleForFeature(ctx, temporalSchedule)
		}).
		Assess("Schedule exists", func(ctx context.Context, tt *testing.T, cfg *envconf.Config) context.Context {
			temporalNamespace := GetTemporalNamespaceForFeature(ctx)
			temporalSchedule := GetTemporalScheduleForFeature(ctx)

			connectAddr, closePortForward, err := forwardPortToTemporalFrontend(ctx, cfg, t, cluster)
			if err != nil {
				t.Fatal(err)
			}
			defer closePortForward()

			client := cfg.Client().Resources().GetControllerRuntimeClient()

			clusterClient, err := temporal.GetClusterClient(ctx, client, cluster, temporal.WithHostPort(connectAddr))
			if err != nil {
				t.Fatal(err)
			}

			err = wait.For(func(ctx context.Context) (done bool, err error) {
				// If no error while describing the schedule, it works.
				_, err = clusterClient.WorkflowService().DescribeSchedule(ctx, &workflowservice.DescribeScheduleRequest{
					Namespace:  temporalNamespace.GetName(),
					ScheduleId: temporalSchedule.GetName(),
				})
				if err != nil {
					var scheduleNotFoundError *serviceerror.NotFound
					if errors.As(err, &scheduleNotFoundError) {
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
		Assess("Schedule can be deleted", func(ctx context.Context, tt *testing.T, cfg *envconf.Config) context.Context {
			namespace := GetNamespaceForFeature(ctx)
			temporalSchedule := GetTemporalScheduleForFeature(ctx)

			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				err := cfg.Client().Resources(namespace).Get(ctx, temporalSchedule.GetName(), temporalSchedule.GetNamespace(), temporalSchedule)
				if err != nil {
					return err
				}

				temporalSchedule.Spec.AllowDeletion = true
				return cfg.Client().Resources(namespace).Update(ctx, temporalSchedule)
			})
			if err != nil {
				t.Fatal(err)
			}

			// Wait for controller to set finalizer.
			err = wait.For(func(ctx context.Context) (done bool, err error) {
				err = cfg.Client().Resources(namespace).Get(ctx, temporalSchedule.GetName(), temporalSchedule.GetNamespace(), temporalSchedule)
				if err != nil {
					t.Fatal(err)
				}

				result := controllerutil.ContainsFinalizer(temporalSchedule, "deletion.finalizers.temporal.io")
				return result, nil
			}, wait.WithTimeout(2*time.Minute), wait.WithInterval(1*time.Second))
			if err != nil {
				t.Fatal(err)
			}

			err = cfg.Client().Resources(namespace).Delete(ctx, temporalSchedule)
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Feature()

	testenv.Test(t, scheduleFeature)
}

func TestScheduleDeletionWhenNamespaceDoesNotExist(rt *testing.T) {
	var temporalClusterName, temporalNamespaceName string

	feature := features.New("schedule can be deleted when temporal namespace does not exist").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			namespace := GetNamespaceForFeature(ctx)

			temporalClusterName = "does-not-exist"
			temporalNamespaceName = "does-not-exist"

			// create TemporalSchedule
			temporalSchedule := &v1beta1.TemporalSchedule{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
				Spec: v1beta1.TemporalScheduleSpec{
					NamespaceRef: v1beta1.TemporalReference{
						Name: temporalNamespaceName,
					},
					Schedule: v1beta1.Schedule{
						Action: v1beta1.ScheduleAction{
							Workflow: v1beta1.ScheduleWorkflowAction{
								TaskQueue:    "queue",
								WorkflowType: "test",
							},
						},
					},
				},
			}

			err := c.Client().Resources(namespace).Create(ctx, temporalSchedule)
			if err != nil {
				t.Fatal(err)
			}
			return SetTemporalScheduleForFeature(ctx, temporalSchedule)
		}).
		Assess("TemporalCluster does not exist", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var temporalCluster = &v1beta1.TemporalCluster{}
			err := c.Client().Resources().Get(ctx, temporalClusterName, GetNamespaceForFeature(ctx), temporalCluster)
			if err == nil {
				t.Fatalf("found cluster: %v", temporalCluster)
			}

			return ctx
		}).
		Assess("TemporalNamespace does not exist", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var temporalNamespace = &v1beta1.TemporalNamespace{}
			err := c.Client().Resources().Get(ctx, temporalNamespaceName, GetNamespaceForFeature(ctx), temporalNamespace)
			if err == nil {
				t.Fatalf("found namespace: %v", temporalNamespace)
			}

			return ctx
		}).
		Assess("TemporalSchedule can be deleted", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			err := c.Client().Resources(GetNamespaceForFeature(ctx)).Delete(ctx, GetTemporalScheduleForFeature(ctx))
			if err != nil {
				t.Fatalf("failed to delete schedule: %v", err)
			}
			return ctx
		}).
		Feature()

	testenv.Test(rt, feature)
}

func TestScheduleDeletionWhenNamespaceDeleted(rt *testing.T) {
	feature := features.New("schedule can be deleted after a temporal namespace associated with it is also deleted").
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
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// create TemporalSchedule
			namespace := GetNamespaceForFeature(ctx)
			temporalNamespace := GetTemporalNamespaceForFeature(ctx)

			// create the temporal schedule
			temporalSchedule := &v1beta1.TemporalSchedule{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
				Spec: v1beta1.TemporalScheduleSpec{
					NamespaceRef: v1beta1.TemporalReference{
						Name: temporalNamespace.GetName(),
					},
					Schedule: v1beta1.Schedule{
						Action: v1beta1.ScheduleAction{
							Workflow: v1beta1.ScheduleWorkflowAction{
								TaskQueue:    "queue",
								WorkflowType: "workflow",
							},
						},
					},
				},
			}
			err := c.Client().Resources(namespace).Create(ctx, temporalSchedule)
			if err != nil {
				t.Fatal(err)
			}

			return SetTemporalScheduleForFeature(ctx, temporalSchedule)
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
		Assess("TemporalSchedule can be deleted", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			err := c.Client().Resources(GetNamespaceForFeature(ctx)).Delete(ctx, GetTemporalScheduleForFeature(ctx))
			if err != nil {
				t.Fatalf("failed to delete: %v", err)
			}
			return ctx
		}).
		Feature()

	testenv.Test(rt, feature)
}
