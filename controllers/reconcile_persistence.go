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

package controllers

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/alexandrevilain/controller-tools/pkg/reconciler"
	"github.com/alexandrevilain/controller-tools/pkg/resource"
	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/resource/base"
	"github.com/alexandrevilain/temporal-operator/internal/resource/persistence"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"k8s.io/apimachinery/pkg/runtime"
)

func sanitizeVersionToName(version *version.Version) string {
	return strings.ReplaceAll(version.String(), ".", "-")
}

func datastoreTypeShortName(datastore v1beta1.DatastoreType) string {
	switch datastore {
	case v1beta1.CassandraDatastore:
		return "cass"
	case v1beta1.ElasticsearchDatastore:
		return "es"
	case v1beta1.PostgresSQLDatastore:
		return "pg"
	case v1beta1.PostgresSQL12Datastore:
		return "pg12"
	case v1beta1.MySQLDatastore:
		return "my"
	case v1beta1.MySQL8Datastore:
		return "my8"
	default:
		return ""
	}
}

// applyStatusDatastoreTypeDefaultValue sets default value on the status.persistence.[store].type for clusters created with operator =< 0.15.1.
// This field was added in the status type to run temporal-sql-tool's update-schema when SQL plugin is updated from postgres to postgres12.
// To run this the operator should track the datastore type when its created.
func (r *TemporalClusterReconciler) applyStatusDatastoreTypeDefaultValue(status *v1beta1.DatastoreStatus, datastore *v1beta1.DatastoreSpec) {
	if status == nil || datastore == nil {
		return
	}

	if status.Created && status.Setup && status.SchemaVersion != nil && status.Type == "" {
		status.Type = datastore.GetType()
	}
}

func (r *TemporalClusterReconciler) reconcilePersistenceStatus(cluster *v1beta1.TemporalCluster) {
	if cluster.Status.Persistence == nil {
		cluster.Status.Persistence = new(v1beta1.TemporalPersistenceStatus)
	}

	if cluster.Status.Persistence.DefaultStore == nil {
		cluster.Status.Persistence.DefaultStore = new(v1beta1.DatastoreStatus)
	}

	r.applyStatusDatastoreTypeDefaultValue(cluster.Status.Persistence.DefaultStore, cluster.Spec.Persistence.DefaultStore)

	if cluster.Status.Persistence.VisibilityStore == nil {
		cluster.Status.Persistence.VisibilityStore = new(v1beta1.DatastoreStatus)
	}

	r.applyStatusDatastoreTypeDefaultValue(cluster.Status.Persistence.VisibilityStore, cluster.Spec.Persistence.VisibilityStore)

	if cluster.Spec.Persistence.SecondaryVisibilityStore != nil {
		if cluster.Status.Persistence.SecondaryVisibilityStore == nil {
			cluster.Status.Persistence.SecondaryVisibilityStore = new(v1beta1.DatastoreStatus)
		}

		r.applyStatusDatastoreTypeDefaultValue(cluster.Status.Persistence.SecondaryVisibilityStore, cluster.Spec.Persistence.SecondaryVisibilityStore)
	}

	if cluster.Spec.Persistence.AdvancedVisibilityStore != nil {
		if cluster.Status.Persistence.AdvancedVisibilityStore == nil {
			cluster.Status.Persistence.AdvancedVisibilityStore = new(v1beta1.DatastoreStatus)
		}

		r.applyStatusDatastoreTypeDefaultValue(cluster.Status.Persistence.AdvancedVisibilityStore, cluster.Spec.Persistence.AdvancedVisibilityStore)
	}
}

func getDatabaseScriptCommand(script string) []string {
	return []string{path.Join("/etc/scripts", script)}
}

// reconcilePersistence tries to reconcile the cluster persistence.
func (r *TemporalClusterReconciler) reconcilePersistence(ctx context.Context, cluster *v1beta1.TemporalCluster) (time.Duration, error) {
	// First of all, ensure status fields are set.
	r.reconcilePersistenceStatus(cluster)

	// Ensure the configmap containing scripts is up-to-date
	_, err := r.Reconciler.ReconcileBuilder(ctx, cluster, persistence.NewSchemaScriptsConfigmapBuilder(cluster, r.Scheme))
	if err != nil {
		return 0, fmt.Errorf("can't reconcile schema script configmap: %w", err)
	}

	// Ensure the serviceaccount used by jobs is up-to-date
	serviceAccountBuilder := base.NewServiceAccountBuilder(persistence.ServiceNameSuffix, cluster, r.Scheme, &v1beta1.ServiceSpec{})
	_, err = r.Reconciler.ReconcileBuilders(ctx, cluster, []resource.Builder{serviceAccountBuilder})
	if err != nil {
		return 0, fmt.Errorf("can't reconcile schema serviceaccount: %w", err)
	}

	// Then for each stores actions, check if the corresponding job is created and has successfully ran.
	jobs := []*reconciler.Job{
		{
			Name:    "create-default-database",
			Command: getDatabaseScriptCommand(persistence.CreateDefaultDatabaseScript),
			Skip: func(owner runtime.Object) bool {
				cluster := owner.(*v1beta1.TemporalCluster)
				return cluster.Spec.Persistence.DefaultStore.SkipCreate ||
					cluster.Status.Persistence.DefaultStore.Created
			},
			ReportSuccess: func(owner runtime.Object) error {
				c := owner.(*v1beta1.TemporalCluster)
				c.Status.Persistence.DefaultStore.Created = true
				return nil
			},
		},
		{
			Name:    "create-visibility-database",
			Command: getDatabaseScriptCommand(persistence.CreateVisibilityDatabaseScript),
			Skip: func(owner runtime.Object) bool {
				cluster := owner.(*v1beta1.TemporalCluster)
				return cluster.Spec.Persistence.VisibilityStore.SkipCreate ||
					cluster.Status.Persistence.VisibilityStore.Created
			},
			ReportSuccess: func(owner runtime.Object) error {
				c := owner.(*v1beta1.TemporalCluster)
				c.Status.Persistence.VisibilityStore.Created = true
				return nil
			},
		},
		{
			Name:    "setup-default-schema",
			Command: getDatabaseScriptCommand(persistence.SetupDefaultSchemaScript),
			Skip: func(owner runtime.Object) bool {
				return owner.(*v1beta1.TemporalCluster).Status.Persistence.DefaultStore.Setup
			},
			ReportSuccess: func(owner runtime.Object) error {
				c := owner.(*v1beta1.TemporalCluster)
				c.Status.Persistence.DefaultStore.Setup = true
				return nil
			},
		},
		{
			Name:    "setup-visibility-schema",
			Command: getDatabaseScriptCommand(persistence.SetupVisibilitySchemaScript),
			Skip: func(owner runtime.Object) bool {
				return owner.(*v1beta1.TemporalCluster).Status.Persistence.VisibilityStore.Setup
			},
			ReportSuccess: func(owner runtime.Object) error {
				c := owner.(*v1beta1.TemporalCluster)
				c.Status.Persistence.VisibilityStore.Setup = true
				return nil
			},
		},
		{
			Name:    fmt.Sprintf("update-default-schema-v-%s", sanitizeVersionToName(cluster.Spec.Version)),
			Command: getDatabaseScriptCommand(persistence.UpdateDefaultSchemaScript),
			Skip: func(owner runtime.Object) bool {
				c := owner.(*v1beta1.TemporalCluster)
				if c.Status.Persistence.DefaultStore.SchemaVersion == nil {
					return false
				}
				return c.Status.Persistence.DefaultStore.SchemaVersion.GreaterOrEqual(cluster.Spec.Version)
			},
			ReportSuccess: func(owner runtime.Object) error {
				c := owner.(*v1beta1.TemporalCluster)
				c.Status.Persistence.DefaultStore.SchemaVersion = c.Spec.Version.DeepCopy()
				c.Status.Persistence.DefaultStore.Type = c.Spec.Persistence.DefaultStore.GetType()
				return nil
			},
		},
		{
			Name:    fmt.Sprintf("update-visibility-schema-v-%s-%s", sanitizeVersionToName(cluster.Spec.Version), datastoreTypeShortName(cluster.Spec.Persistence.VisibilityStore.GetType())),
			Command: getDatabaseScriptCommand(persistence.UpdateVisibilitySchemaScript),
			Skip: func(owner runtime.Object) bool {
				c := owner.(*v1beta1.TemporalCluster)
				if c.Status.Persistence.VisibilityStore.SchemaVersion == nil {
					return false
				}
				if c.Status.Persistence.VisibilityStore.Type != c.Spec.Persistence.VisibilityStore.GetType() {
					return false
				}
				return c.Status.Persistence.VisibilityStore.SchemaVersion.GreaterOrEqual(c.Spec.Version)
			},
			ReportSuccess: func(owner runtime.Object) error {
				c := owner.(*v1beta1.TemporalCluster)
				c.Status.Persistence.VisibilityStore.SchemaVersion = c.Spec.Version.DeepCopy()
				c.Status.Persistence.VisibilityStore.Type = c.Spec.Persistence.VisibilityStore.GetType()
				return nil
			},
		},
	}

	if cluster.Spec.Persistence.SecondaryVisibilityStore != nil {
		jobs = append(jobs,
			&reconciler.Job{
				Name:    "create-secondary-visibility-database",
				Command: getDatabaseScriptCommand(persistence.CreateSecondaryVisibilityDatabaseScript),
				Skip: func(owner runtime.Object) bool {
					return owner.(*v1beta1.TemporalCluster).Status.Persistence.SecondaryVisibilityStore.Created
				},
				ReportSuccess: func(owner runtime.Object) error {
					c := owner.(*v1beta1.TemporalCluster)
					c.Status.Persistence.SecondaryVisibilityStore.Created = true
					return nil
				},
			},
			&reconciler.Job{
				Name:    "setup-secondary-visibility-schema",
				Command: getDatabaseScriptCommand(persistence.SetupSecondaryVisibilitySchemaScript),
				Skip: func(owner runtime.Object) bool {
					return owner.(*v1beta1.TemporalCluster).Status.Persistence.SecondaryVisibilityStore.Setup
				},
				ReportSuccess: func(owner runtime.Object) error {
					c := owner.(*v1beta1.TemporalCluster)
					c.Status.Persistence.SecondaryVisibilityStore.Setup = true
					return nil
				},
			},
			&reconciler.Job{
				Name:    fmt.Sprintf("update-2nd-visibility-schema-v-%s-%s", sanitizeVersionToName(cluster.Spec.Version), datastoreTypeShortName(cluster.Spec.Persistence.SecondaryVisibilityStore.GetType())),
				Command: getDatabaseScriptCommand(persistence.UpdateSecondaryVisibilitySchemaScript),
				Skip: func(owner runtime.Object) bool {
					c := owner.(*v1beta1.TemporalCluster)
					if c.Status.Persistence.SecondaryVisibilityStore.SchemaVersion == nil {
						return false
					}
					if c.Status.Persistence.VisibilityStore.Type != c.Spec.Persistence.VisibilityStore.GetType() {
						return false
					}
					return c.Status.Persistence.SecondaryVisibilityStore.SchemaVersion.GreaterOrEqual(c.Spec.Version)
				},
				ReportSuccess: func(owner runtime.Object) error {
					c := owner.(*v1beta1.TemporalCluster)
					c.Status.Persistence.SecondaryVisibilityStore.SchemaVersion = c.Spec.Version.DeepCopy()
					c.Status.Persistence.VisibilityStore.Type = c.Spec.Persistence.VisibilityStore.GetType()
					return nil
				},
			})
	}
	if cluster.Spec.Persistence.AdvancedVisibilityStore != nil {
		jobs = append(jobs,
			&reconciler.Job{
				Name:    "create-advanced-visibility-database",
				Command: getDatabaseScriptCommand(persistence.CreateAdvancedVisibilityDatabaseScript),
				Skip: func(owner runtime.Object) bool {
					return owner.(*v1beta1.TemporalCluster).Status.Persistence.AdvancedVisibilityStore.Created
				},
				ReportSuccess: func(owner runtime.Object) error {
					c := owner.(*v1beta1.TemporalCluster)
					c.Status.Persistence.AdvancedVisibilityStore.Created = true
					return nil
				},
			},
			&reconciler.Job{
				Name:    "setup-advanced-visibility-schema",
				Command: getDatabaseScriptCommand(persistence.SetupAdvancedVisibilitySchemaScript),
				Skip: func(owner runtime.Object) bool {
					return owner.(*v1beta1.TemporalCluster).Status.Persistence.AdvancedVisibilityStore.Setup
				},
				ReportSuccess: func(owner runtime.Object) error {
					c := owner.(*v1beta1.TemporalCluster)
					c.Status.Persistence.AdvancedVisibilityStore.Setup = true
					return nil
				},
			},
			&reconciler.Job{
				Name:    fmt.Sprintf("update-advanced-visibility-schema-v-%s", sanitizeVersionToName(cluster.Spec.Version)),
				Command: getDatabaseScriptCommand(persistence.UpdateAdvancedVisibilitySchemaScript),
				Skip: func(owner runtime.Object) bool {
					c := owner.(*v1beta1.TemporalCluster)
					if c.Status.Persistence.AdvancedVisibilityStore.SchemaVersion == nil {
						return false
					}
					return c.Status.Persistence.AdvancedVisibilityStore.SchemaVersion.GreaterOrEqual(c.Spec.Version)
				},
				ReportSuccess: func(owner runtime.Object) error {
					c := owner.(*v1beta1.TemporalCluster)
					c.Status.Persistence.AdvancedVisibilityStore.SchemaVersion = c.Spec.Version.DeepCopy()
					c.Status.Persistence.AdvancedVisibilityStore.Type = c.Spec.Persistence.AdvancedVisibilityStore.GetType()
					return nil
				},
			})
	}

	factory := func(owner runtime.Object, scheme *runtime.Scheme, name string, command []string) resource.Builder {
		cluster := owner.(*v1beta1.TemporalCluster)
		return persistence.NewSchemaJobBuilder(cluster, scheme, name, command)
	}

	return r.Jobs.Reconcile(ctx, cluster, factory, jobs)
}
