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

package persistence

import (
	"context"
	"fmt"
	"path"

	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/internal/forked/go.temporal.io/server/tools/common/schema"
	"github.com/blang/semver/v4"
	temporallog "go.temporal.io/server/common/log"
	_ "go.temporal.io/server/common/persistence/sql/sqlplugin/mysql"      // needed to load mysql plugin
	_ "go.temporal.io/server/common/persistence/sql/sqlplugin/postgresql" // needed to load postgresql plugin
	"go.temporal.io/server/tools/sql"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Schema string

const (
	DefaultSchema    Schema = "default"
	VisibilitySchema Schema = "visibility"
)

const (
	defaultSchemaPath    = "temporal"
	visibilitySchemaPath = "visibility"

	postgreSQLSchemaPath        = "postgresql"
	postgreSQLVersionSchemaPath = "v96"

	mysqlSchemaPath        = "mysql"
	mysqlVersionSchemaPath = "v57"
)

// Manager handler persistence receonciliation.
type Manager struct {
	client.Client

	SchemaFilePath string
}

// NewManager creates a new instance of the persistence manager.
func NewManager(c client.Client, schemaFilePath string) *Manager {
	return &Manager{
		Client:         c,
		SchemaFilePath: schemaFilePath,
	}
}

// RunStoreSetupTask runs the setup schema task on the provided cluster's store.
func (m *Manager) RunStoreSetupTask(ctx context.Context, cluster *v1alpha1.TemporalCluster, store *v1alpha1.TemporalDatastoreSpec) error {
	conn, err := m.getSQLConnectionFromDatastoreSpec(ctx, store, cluster.Namespace)
	if err != nil {
		return err
	}
	defer conn.Close()

	setupTask := schema.NewSetupSchemaTask(conn, &schema.SetupConfig{
		InitialVersion:    "0.0",
		Overwrite:         false,
		DisableVersioning: false,
	}, temporallog.NewCLILogger())

	return setupTask.Run()
}

// RunDefaultStoreUpdateTask runs the update schema task on the provided cluster's default store.
func (m *Manager) RunDefaultStoreUpdateTask(ctx context.Context, cluster *v1alpha1.TemporalCluster, store *v1alpha1.TemporalDatastoreSpec, version semver.Version) error {
	return m.runUpdateSchemaTasks(ctx, cluster, store, DefaultSchema, version)
}

// RunVisibilityStoreUpdateTask runs the update schema task on the provided cluster's visibility store.
func (m *Manager) RunVisibilityStoreUpdateTask(ctx context.Context, cluster *v1alpha1.TemporalCluster, store *v1alpha1.TemporalDatastoreSpec, version semver.Version) error {
	return m.runUpdateSchemaTasks(ctx, cluster, store, VisibilitySchema, version)
}

func (m *Manager) getSQLConnectionFromDatastoreSpec(ctx context.Context, store *v1alpha1.TemporalDatastoreSpec, namespace string) (*sql.Connection, error) {
	config := NewSQLconfigFromDatastoreSpec(store)

	passwordSecret := &corev1.Secret{}
	err := m.Get(ctx, types.NamespacedName{Name: store.PasswordSecretRef.Name, Namespace: namespace}, passwordSecret)
	if err != nil {
		return nil, err
	}

	config.Password = string(passwordSecret.Data[store.PasswordSecretRef.Key])

	return sql.NewConnection(config)
}
func (m *Manager) computeSchemaDir(storeType v1alpha1.DatastoreType, targetSchema Schema) string {
	storeSchemaPath := ""
	storeVersionSchemaPath := ""
	switch storeType {
	case v1alpha1.PostgresSQLDatastore:
		storeSchemaPath = postgreSQLSchemaPath
		storeVersionSchemaPath = postgreSQLVersionSchemaPath
	case v1alpha1.MySQLDatastore:
		storeSchemaPath = mysqlSchemaPath
		storeVersionSchemaPath = mysqlVersionSchemaPath
	case v1alpha1.CassandraDatastore:
		panic("not supported yet")
	}

	tagetSchemaPath := defaultSchemaPath
	if targetSchema == VisibilitySchema {
		tagetSchemaPath = visibilitySchemaPath
	}

	return path.Join(m.SchemaFilePath, storeSchemaPath, storeVersionSchemaPath, tagetSchemaPath, "versioned")
}

func (m *Manager) runUpdateSchemaTasks(ctx context.Context, cluster *v1alpha1.TemporalCluster, store *v1alpha1.TemporalDatastoreSpec, targetSchema Schema, targetVersion semver.Version) error {
	conn, err := m.getSQLConnectionFromDatastoreSpec(ctx, store, cluster.Namespace)
	if err != nil {
		return err
	}
	defer conn.Close()

	datastoreType, err := store.GetDatastoreType()
	if err != nil {
		return err
	}

	updateTask := schema.NewUpdateSchemaTask(conn, &schema.UpdateConfig{
		DBName:        store.SQL.DatabaseName,
		TargetVersion: fmt.Sprintf("v%d.%d", targetVersion.Major, targetVersion.Minor),
		SchemaDir:     m.computeSchemaDir(datastoreType, targetSchema),
		IsDryRun:      false,
	}, temporallog.NewCLILogger())

	err = updateTask.Run()
	if err != nil {
		return err
	}

	return nil
}
