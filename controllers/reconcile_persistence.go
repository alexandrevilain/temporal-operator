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
	"strings"
	"time"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/reconciler"
	"github.com/alexandrevilain/temporal-operator/pkg/resource"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/persistence"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"k8s.io/apimachinery/pkg/runtime"
)

func sanitizeVersionToName(version *version.Version) string {
	return strings.ReplaceAll(version.String(), ".", "-")
}

func (r *TemporalClusterReconciler) reconcileSchemaScriptsConfigmap(ctx context.Context, cluster *v1beta1.TemporalCluster) error {
	builders := []resource.Builder{
		persistence.NewSchemaScriptsConfigmapBuilder(cluster, r.Scheme),
	}
	_, _, err := r.ReconcileBuilders(ctx, cluster, builders)
	return err
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

	if cluster.Status.Persistence.AdvancedVisibilityStore == nil && cluster.Spec.Persistence.AdvancedVisibilityStore != nil {
		cluster.Status.Persistence.AdvancedVisibilityStore = new(v1beta1.DatastoreStatus)
	}
}

// reconcilePersistence tries to reconcile the cluster persistence.
func (r *TemporalClusterReconciler) reconcilePersistence(ctx context.Context, cluster *v1beta1.TemporalCluster) (time.Duration, error) {
	// First of all, ensure status fields are set.
	r.reconcilePersistenceStatus(cluster)
	// Ensure the configmap containing scripts is up-to-date
	err := r.reconcileSchemaScriptsConfigmap(ctx, cluster)
	if err != nil {
		return 0, fmt.Errorf("can't reconcile schema script configmap: %w", err)
	}

	// Then for each stores actions, check if the corresponding job is created and has successfully ran.
	jobs := []*reconciler.Job{
		{
			Name:    "create-default-database",
			Command: []string{"/etc/scripts/create-default-database.sh"},
			Skip: func(owner runtime.Object) bool {
				return owner.(*v1beta1.TemporalCluster).Status.Persistence.DefaultStore.Created
			},
			ReportSuccess: func(owner runtime.Object) error {
				owner.(*v1beta1.TemporalCluster).Status.Persistence.DefaultStore.Created = true
				return nil
			},
		},
		{
			Name:    "create-visibility-database",
			Command: []string{"/etc/scripts/create-visibility-database.sh"},
			Skip: func(owner runtime.Object) bool {
				return owner.(*v1beta1.TemporalCluster).Status.Persistence.VisibilityStore.Created
			},
			ReportSuccess: func(owner runtime.Object) error {
				c := owner.(*v1beta1.TemporalCluster)
				c.Status.Persistence.VisibilityStore.Created = true
				return nil
			},
		},
		{
			Name:    "setup-default-schema",
			Command: []string{"/etc/scripts/setup-default-schema.sh"},
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
			Command: []string{"/etc/scripts/setup-visibility-schema.sh"},
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
			Command: []string{"/etc/scripts/update-default-schema.sh"},
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
			Command: []string{"/etc/scripts/update-visibility-schema.sh"},
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
	if cluster.Spec.Persistence.AdvancedVisibilityStore != nil {
		jobs = append(jobs,
			&reconciler.Job{
				Name:    "setup-advanced-visibility",
				Command: []string{"/etc/scripts/setup-advanced-visibility.sh"},
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
				Name:    fmt.Sprintf("update-advanced-visibility-v-%s", sanitizeVersionToName(cluster.Spec.Version)),
				Command: []string{"/etc/scripts/update-advanced-visibility.sh"},
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

	return r.ReconcileJobs(ctx, cluster, factory, jobs)
}
