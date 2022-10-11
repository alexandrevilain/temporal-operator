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
	"github.com/alexandrevilain/temporal-operator/pkg/resource/persistence"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *TemporalClusterReconciler) reconcileSchemaScriptsConfigmap(ctx context.Context, cluster *v1beta1.TemporalCluster) error {
	schemaScriptConfigMapBuilder := persistence.NewSchemaScriptsConfigmapBuilder(cluster, r.Scheme)
	schemaScriptConfigMap, err := schemaScriptConfigMapBuilder.Build()
	if err != nil {
		return err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, schemaScriptConfigMap, func() error {
		return schemaScriptConfigMapBuilder.Update(schemaScriptConfigMap)
	})
	return err
}

type job struct {
	name          string
	command       []string
	skip          func(c *v1beta1.TemporalCluster) bool
	reportSuccess func(c *v1beta1.TemporalCluster) error
}

func sanitizeVersionToName(version *version.Version) string {
	return strings.ReplaceAll(version.String(), ".", "-")
}

// reconcilePersistence tries to reconcile the cluster persistence.
func (r *TemporalClusterReconciler) reconcilePersistence(ctx context.Context, cluster *v1beta1.TemporalCluster) (time.Duration, error) {
	logger := log.FromContext(ctx)

	// First of all, ensure the configmap containing scripts is up-to-date
	err := r.reconcileSchemaScriptsConfigmap(ctx, cluster)
	if err != nil {
		return 0, fmt.Errorf("can't reconcile schema script configmap: %w", err)
	}

	// Then for each stores actions, check if the corresponding job is created and has succesfully ran.
	jobs := []job{
		{
			name:    "create-default-database",
			command: []string{"/etc/scripts/create-default-database.sh"},
			skip: func(c *v1beta1.TemporalCluster) bool {
				return c.Status.Persistence.DefaultStore.Created
			},
			reportSuccess: func(c *v1beta1.TemporalCluster) error {
				c.Status.Persistence.DefaultStore.Created = true
				return nil
			},
		},
		{
			name:    "create-visibility-database",
			command: []string{"/etc/scripts/create-visibility-database.sh"},
			skip: func(c *v1beta1.TemporalCluster) bool {
				return c.Status.Persistence.VisibilityStore.Created
			},
			reportSuccess: func(c *v1beta1.TemporalCluster) error {
				c.Status.Persistence.VisibilityStore.Created = true
				return nil
			},
		},
		{
			name:    "setup-default-schema",
			command: []string{"/etc/scripts/setup-default-schema.sh"},
			skip: func(c *v1beta1.TemporalCluster) bool {
				return c.Status.Persistence.DefaultStore.Setup
			},
			reportSuccess: func(c *v1beta1.TemporalCluster) error {
				c.Status.Persistence.DefaultStore.Setup = true
				return nil
			},
		},
		{
			name:    "setup-visibility-schema",
			command: []string{"/etc/scripts/setup-visibility-schema.sh"},
			skip: func(c *v1beta1.TemporalCluster) bool {
				return c.Status.Persistence.VisibilityStore.Setup
			},
			reportSuccess: func(c *v1beta1.TemporalCluster) error {
				c.Status.Persistence.VisibilityStore.Setup = true
				return nil
			},
		},
		{
			name:    fmt.Sprintf("update-default-schema-v-%s", sanitizeVersionToName(cluster.Spec.Version)),
			command: []string{"/etc/scripts/update-default-schema.sh"},
			skip: func(c *v1beta1.TemporalCluster) bool {
				if c.Status.Persistence.DefaultStore.SchemaVersion == nil {
					return false
				}
				return c.Status.Persistence.DefaultStore.SchemaVersion.GreaterOrEqual(cluster.Spec.Version)
			},
			reportSuccess: func(c *v1beta1.TemporalCluster) error {
				c.Status.Persistence.DefaultStore.SchemaVersion = c.Spec.Version.DeepCopy()
				return nil
			},
		},
		{
			name:    fmt.Sprintf("update-visibility-schema-v-%s", sanitizeVersionToName(cluster.Spec.Version)),
			command: []string{"/etc/scripts/update-visibility-schema.sh"},
			skip: func(c *v1beta1.TemporalCluster) bool {
				if c.Status.Persistence.VisibilityStore.SchemaVersion == nil {
					return false
				}
				return c.Status.Persistence.VisibilityStore.SchemaVersion.GreaterOrEqual(c.Spec.Version)
			},
			reportSuccess: func(c *v1beta1.TemporalCluster) error {
				c.Status.Persistence.VisibilityStore.SchemaVersion = c.Spec.Version.DeepCopy()
				return nil
			},
		},
	}
	if cluster.Spec.Persistence.AdvancedVisibilityStore != nil {
		jobs = append(jobs,
			job{
				name:    "setup-advanced-visibility",
				command: []string{"/etc/scripts/setup-advanced-visibility.sh"},
				skip: func(c *v1beta1.TemporalCluster) bool {
					return c.Status.Persistence.AdvancedVisibilityStore.Setup
				},
				reportSuccess: func(c *v1beta1.TemporalCluster) error {
					c.Status.Persistence.AdvancedVisibilityStore.Setup = true
					return nil
				},
			},
			job{
				name:    fmt.Sprintf("update-advanced-visibility-v-%s", sanitizeVersionToName(cluster.Spec.Version)),
				command: []string{"/etc/scripts/update-advanced-visibility.sh"},
				skip: func(c *v1beta1.TemporalCluster) bool {
					if c.Status.Persistence.AdvancedVisibilityStore.SchemaVersion == nil {
						return false
					}
					return c.Status.Persistence.AdvancedVisibilityStore.SchemaVersion.GreaterOrEqual(c.Spec.Version)
				},
				reportSuccess: func(c *v1beta1.TemporalCluster) error {
					c.Status.Persistence.AdvancedVisibilityStore.SchemaVersion = c.Spec.Version.DeepCopy()
					return nil
				},
			})
	}

	for _, job := range jobs {
		if job.skip(cluster) {
			continue
		}

		logger.Info("Checking for persistence job", "name", job.name)
		expectedJobBuilder := persistence.NewSchemaJobBuilder(cluster, r.Scheme, job.name, job.command)

		expectedJob, err := expectedJobBuilder.Build()
		if err != nil {
			return 0, nil
		}

		matchingJob := &batchv1.Job{}
		err = r.Client.Get(ctx, types.NamespacedName{Name: expectedJob.GetName(), Namespace: expectedJob.GetNamespace()}, matchingJob)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// The job is not found, create it
				_, err := controllerutil.CreateOrUpdate(ctx, r.Client, expectedJob, func() error {
					return expectedJobBuilder.Update(expectedJob)
				})
				if err != nil {
					return 0, err
				}
			} else {
				return 0, fmt.Errorf("can't get job: %w", err)
			}
		}

		if matchingJob.Status.Succeeded != 1 {
			logger.Info("Waiting for persistence job to complete", "name", job.name)

			// Requeue after 10 seconds
			return 10 * time.Second, nil
		}

		logger.Info("Persistence job is finished", "name", job.name)

		err = job.reportSuccess(cluster)
		if err != nil {
			return 0, fmt.Errorf("can't report job success: %w", err)
		}
	}

	return 0, nil
}
