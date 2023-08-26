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
	"github.com/alexandrevilain/temporal-operator/internal/resource/persistence"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"k8s.io/apimachinery/pkg/runtime"
)

func sanitizeVersionToName(version *version.Version) string {
	return strings.ReplaceAll(version.String(), ".", "-")
}

func (r *TemporalClusterReconciler) reconcilePersistenceStatus(cluster *v1beta1.TemporalCluster) {
	if cluster.Status.Persistence == nil {
		cluster.Status.Persistence = new(v1beta1.TemporalPersistenceStatus)
	}

	if cluster.Status.Persistence.DefaultStore == nil {
		cluster.Status.Persistence.DefaultStore = new(v1beta1.DatastoreStatus)
	}

	if cluster.Status.Persistence.VisibilityStore == nil {
		cluster.Status.Persistence.VisibilityStore = new(v1beta1.DatastoreStatus)
	}

	if cluster.Status.Persistence.SecondaryVisibility == nil && cluster.Spec.Persistence.SecondaryVisibilityStore != nil {
		cluster.Status.Persistence.SecondaryVisibility = new(v1beta1.DatastoreStatus)
	}

	if cluster.Status.Persistence.AdvancedVisibilityStore == nil && cluster.Spec.Persistence.AdvancedVisibilityStore != nil {
		cluster.Status.Persistence.AdvancedVisibilityStore = new(v1beta1.DatastoreStatus)
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
				owner.(*v1beta1.TemporalCluster).Status.Persistence.DefaultStore.Created = true
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
				return nil
			},
		},
		{
			Name:    fmt.Sprintf("update-visibility-schema-v-%s", sanitizeVersionToName(cluster.Spec.Version)),
			Command: getDatabaseScriptCommand(persistence.UpdateVisibilitySchemaScript),
			Skip: func(owner runtime.Object) bool {
				c := owner.(*v1beta1.TemporalCluster)
				if c.Status.Persistence.VisibilityStore.SchemaVersion == nil {
					return false
				}
				return c.Status.Persistence.VisibilityStore.SchemaVersion.GreaterOrEqual(c.Spec.Version)
			},
			ReportSuccess: func(owner runtime.Object) error {
				c := owner.(*v1beta1.TemporalCluster)
				c.Status.Persistence.VisibilityStore.SchemaVersion = c.Spec.Version.DeepCopy()
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
					return owner.(*v1beta1.TemporalCluster).Status.Persistence.SecondaryVisibility.Created
				},
				ReportSuccess: func(owner runtime.Object) error {
					c := owner.(*v1beta1.TemporalCluster)
					c.Status.Persistence.SecondaryVisibility.Created = true
					return nil
				},
			},
			&reconciler.Job{
				Name:    "setup-secondary-visibility-schema",
				Command: getDatabaseScriptCommand(persistence.SetupSecondaryVisibilitySchemaScript),
				Skip: func(owner runtime.Object) bool {
					return owner.(*v1beta1.TemporalCluster).Status.Persistence.SecondaryVisibility.Setup
				},
				ReportSuccess: func(owner runtime.Object) error {
					c := owner.(*v1beta1.TemporalCluster)
					c.Status.Persistence.SecondaryVisibility.Setup = true
					return nil
				},
			},
			&reconciler.Job{
				Name:    fmt.Sprintf("update-secondary-visibility-schema-v-%s", sanitizeVersionToName(cluster.Spec.Version)),
				Command: getDatabaseScriptCommand(persistence.UpdateSecondaryVisibilitySchemaScript),
				Skip: func(owner runtime.Object) bool {
					c := owner.(*v1beta1.TemporalCluster)
					if c.Status.Persistence.SecondaryVisibility.SchemaVersion == nil {
						return false
					}
					return c.Status.Persistence.SecondaryVisibility.SchemaVersion.GreaterOrEqual(c.Spec.Version)
				},
				ReportSuccess: func(owner runtime.Object) error {
					c := owner.(*v1beta1.TemporalCluster)
					c.Status.Persistence.SecondaryVisibility.SchemaVersion = c.Spec.Version.DeepCopy()
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
