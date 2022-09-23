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
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ClusterReconciler) reconcileSchemaScriptsConfigmap(ctx context.Context, cluster *v1beta1.Cluster) error {
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
	name    string
	command []string
}

func sanitizeVersionToName(version string) string {
	return strings.ReplaceAll(version, ".", "-")
}

// reconcilePersistence tries to reconcile the cluster persistence.
func (r *ClusterReconciler) reconcilePersistence(ctx context.Context, cluster *v1beta1.Cluster) (time.Duration, error) {
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
		},
		{
			name:    "create-visibility-database",
			command: []string{"/etc/scripts/create-visibility-database.sh"},
		},
		{
			name:    "setup-default-schema",
			command: []string{"/etc/scripts/setup-default-schema.sh"},
		},
		{
			name:    "setup-visibility-schema",
			command: []string{"/etc/scripts/setup-visibility-schema.sh"},
		},
		{
			name:    fmt.Sprintf("update-default-schema-v-%s", sanitizeVersionToName(cluster.Spec.Version)),
			command: []string{"/etc/scripts/update-default-schema.sh"},
		},
		{
			name:    fmt.Sprintf("update-visibility-schema-v-%s", sanitizeVersionToName(cluster.Spec.Version)),
			command: []string{"/etc/scripts/update-visibility-schema.sh"},
		},
	}
	if cluster.Spec.Persistence.AdvancedVisibilityStore != "" {
		jobs = append(jobs,
			job{
				name:    "setup-advanced-visibility",
				command: []string{"/etc/scripts/setup-advanced-visibility.sh"},
			},
			job{
				name:    fmt.Sprintf("update-advanced-visibility-v-%s", sanitizeVersionToName(cluster.Spec.Version)),
				command: []string{"/etc/scripts/update-advanced-visibility.sh"},
			})
	}

	for _, job := range jobs {
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
	}

	return 0, nil
}
