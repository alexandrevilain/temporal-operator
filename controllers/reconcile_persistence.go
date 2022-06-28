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
	"errors"
	"fmt"

	appsv1alpha1 "github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/blang/semver/v4"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// reconcilePersistence tries to reconcile the cluster persistence.
// If first checks if the schema status field for both of the default and visibility stores are empty. If empty it tries to setup the stores' schemas.
// Then it compares the current schema version (from the cluster's status) and determine if a schema upgrade is needed.
func (r *TemporalClusterReconciler) reconcilePersistence(ctx context.Context, temporalCluster *appsv1alpha1.TemporalCluster, clusterVersion semver.Version) error {
	logger := log.FromContext(ctx)

	defaultStore, found := temporalCluster.GetDefaultDatastore()
	if !found {
		return errors.New("default datastore not found")
	}

	visibilityStore, found := temporalCluster.GetVisibilityDatastore()
	if !found {
		return errors.New("visibility datastore not found")
	}

	if temporalCluster.Status.Persistence.DefaultStoreSchemaVersion == "" {
		logger.Info("Starting default store setup task")
		err := r.PersistenceManager.RunStoreSetupTask(ctx, temporalCluster, defaultStore)
		if err != nil {
			return err
		}
		temporalCluster.Status.Persistence.DefaultStoreSchemaVersion = "0.0.0"
	}

	if temporalCluster.Status.Persistence.VisibilityStoreSchemaVersion == "" {
		logger.Info("Starting visibility store setup task")
		err := r.PersistenceManager.RunStoreSetupTask(ctx, temporalCluster, visibilityStore)
		if err != nil {
			return err
		}
		temporalCluster.Status.Persistence.VisibilityStoreSchemaVersion = "0.0.0"
	}

	defaultStoreType, err := defaultStore.GetDatastoreType()
	if err != nil {
		return err
	}

	visibilityStoreType, err := visibilityStore.GetDatastoreType()
	if err != nil {
		return err
	}

	matchingVersion, ok := version.GetMatchingSupportedVersion(clusterVersion)
	if !ok {
		return errors.New("no matching version found")
	}

	expectedDefaultStoreSchemaVersion := matchingVersion.DefaultSchemaVersions[defaultStoreType]
	expectedVisibilityStoreSchemaVersion := matchingVersion.VisibilitySchemaVersion[visibilityStoreType]

	currentDefaultStoreSchemaVersion, err := version.Parse(temporalCluster.Status.Persistence.DefaultStoreSchemaVersion)
	if err != nil {
		return err
	}

	currentVisibilityStoreSchemaVersion, err := version.Parse(temporalCluster.Status.Persistence.VisibilityStoreSchemaVersion)
	if err != nil {
		return err
	}

	if expectedDefaultStoreSchemaVersion.GT(currentDefaultStoreSchemaVersion) {
		logger.Info("Starting default store update task")
		err := r.PersistenceManager.RunDefaultStoreUpdateTask(ctx, temporalCluster, defaultStore, expectedDefaultStoreSchemaVersion)
		if err != nil {
			return err
		}
		temporalCluster.Status.Persistence.DefaultStoreSchemaVersion = expectedDefaultStoreSchemaVersion.String()
	}

	if expectedVisibilityStoreSchemaVersion.GT(currentVisibilityStoreSchemaVersion) {
		logger.Info("Starting visibility store update task")
		err := r.PersistenceManager.RunVisibilityStoreUpdateTask(ctx, temporalCluster, visibilityStore, expectedVisibilityStoreSchemaVersion)
		if err != nil {
			return err
		}
		temporalCluster.Status.Persistence.VisibilityStoreSchemaVersion = expectedVisibilityStoreSchemaVersion.String()
	}

	// Reconcile advanced visibility store if enabled
	if temporalCluster.Spec.Persistence.AdvancedVisibilityStore != "" {
		expectedAdvancedVisibilityStoreSchemaVersion := matchingVersion.AdvancedVisibilitySchemaVersion[appsv1alpha1.ElasticsearchDatastore]

		currentAdvancedVisibilityStoreSchemaVersion := version.NullVersion
		if temporalCluster.Status.Persistence.AdvancedVisibilityStoreSchemaVersion != "" {
			currentAdvancedVisibilityStoreSchemaVersion, err = version.Parse(temporalCluster.Status.Persistence.AdvancedVisibilityStoreSchemaVersion)
			if err != nil {
				return fmt.Errorf("can't parse current advanced visibility schema version: %w", err)
			}
		}

		if expectedAdvancedVisibilityStoreSchemaVersion.GT(currentAdvancedVisibilityStoreSchemaVersion) {
			logger.Info("Starting advanced visibility store update task")

			advancedVisibilityStore, found := temporalCluster.GetAdvancedVisibilityDatastore()
			if !found {
				return errors.New("advanced visibility datastore not found")
			}

			err := r.PersistenceManager.RunAdvancedVisibilityStoreTasks(ctx, temporalCluster, advancedVisibilityStore, expectedAdvancedVisibilityStoreSchemaVersion)
			if err != nil {
				return err
			}
			temporalCluster.Status.Persistence.AdvancedVisibilityStoreSchemaVersion = expectedAdvancedVisibilityStoreSchemaVersion.String()
		}
	}

	return r.updateTemporalClusterStatus(ctx, temporalCluster)
}
