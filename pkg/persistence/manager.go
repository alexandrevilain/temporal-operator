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
	"net/http"
	"os"
	"path"

	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/internal/forked/go.temporal.io/server/tools/cassandra"
	"github.com/alexandrevilain/temporal-operator/internal/forked/go.temporal.io/server/tools/common/schema"
	tlog "github.com/alexandrevilain/temporal-operator/pkg/log"
	"github.com/blang/semver/v4"
	_ "go.temporal.io/server/common/persistence/sql/sqlplugin/mysql"      // needed to load mysql plugin
	_ "go.temporal.io/server/common/persistence/sql/sqlplugin/postgresql" // needed to load postgresql plugin
	esclient "go.temporal.io/server/common/persistence/visibility/store/elasticsearch/client"
	"go.temporal.io/server/tools/sql"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

	cassandraSchemaPath        = "cassandra"
	cassandraVersionSchemaPath = ""

	elasticsearchSchemaPath   = "elasticsearch"
	elasticsearchTemplateName = "temporal_visibility_v1_template"
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
	conn, err := m.getDatastoreConnection(ctx, store, cluster.Namespace)
	if err != nil {
		return err
	}
	defer conn.Close()

	setupTask := schema.NewSetupSchemaTask(conn, &schema.SetupConfig{
		InitialVersion:    "0.0",
		Overwrite:         false,
		DisableVersioning: false,
	}, tlog.NewTemporalLogFromContext(ctx))

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

// RunAdvancedVisibilityStoreTasks creates the index setting for the temporal index and the index itself.
func (m *Manager) RunAdvancedVisibilityStoreTasks(ctx context.Context, cluster *v1alpha1.TemporalCluster, store *v1alpha1.TemporalDatastoreSpec, templateVersion semver.Version) error {
	conn, err := m.getElasticsearchConnectionFromDatastoreSpec(ctx, store, cluster.Namespace)
	if err != nil {
		return fmt.Errorf("can' get elasticsearch connection from datastore spec: %w", err)
	}

	templateFilePath := m.computeAdvancedVisibilitySchemaDir(store.Elasticsearch.Version, templateVersion)
	log.FromContext(ctx).Info("Retrieved advanced visibility schema dir", "path", templateFilePath)

	content, err := os.ReadFile(templateFilePath)
	if err != nil {
		return fmt.Errorf("can'read template file: %w", err)
	}

	_, err = conn.IndexPutTemplate(ctx, elasticsearchTemplateName, string(content))
	if err != nil {
		return fmt.Errorf("can't put index template: %w", err)
	}

	if store.Elasticsearch.Indices.Visibility != "" {
		_, err = conn.CreateIndex(ctx, store.Elasticsearch.Indices.Visibility)
		if err != nil {
			return fmt.Errorf("can't create '%s' index: %w", store.Elasticsearch.Indices.Visibility, err)
		}
	}

	if store.Elasticsearch.Indices.SecondaryVisibility != "" {
		_, err = conn.CreateIndex(ctx, store.Elasticsearch.Indices.SecondaryVisibility)
		if err != nil {
			return fmt.Errorf("can't create '%s' index: %w", store.Elasticsearch.Indices.SecondaryVisibility, err)
		}
	}

	return nil
}

func (m *Manager) getSQLConnectionFromDatastoreSpec(ctx context.Context, store *v1alpha1.TemporalDatastoreSpec, namespace string) (*sql.Connection, error) {
	config := NewSQLConfigFromDatastoreSpec(store)

	var err error
	config.Password, err = m.getStorePassword(ctx, store, namespace)
	if err != nil {
		return nil, err
	}

	config.ConnectAddr = "localhost:5432"

	return sql.NewConnection(config)
}

func (m *Manager) getDatastoreConnection(ctx context.Context, store *v1alpha1.TemporalDatastoreSpec, namespace string) (schema.DB, error) {
	datastoreType, err := store.GetDatastoreType()
	if err != nil {
		return nil, err
	}

	switch datastoreType {
	case v1alpha1.MySQLDatastore, v1alpha1.PostgresSQLDatastore:
		return m.getSQLConnectionFromDatastoreSpec(ctx, store, namespace)
	case v1alpha1.CassandraDatastore:
		return m.getCassandraConnectionFromDatastoreSpec(ctx, store, namespace)
	default:
		return nil, fmt.Errorf("unknown datastore type: %s", datastoreType)
	}
}

// getElasticsearchConnectionFromDatastoreSpec returns the ES client connection from the store spec.
func (m *Manager) getElasticsearchConnectionFromDatastoreSpec(ctx context.Context, store *v1alpha1.TemporalDatastoreSpec, namespace string) (esclient.IntegrationTestsClient, error) {
	config, err := NewElasticsearchConfigFromDatastoreSpec(store)
	if err != nil {
		return nil, err
	}

	config.Password, err = m.getStorePassword(ctx, store, namespace)
	if err != nil {
		return nil, err
	}

	c, err := esclient.NewClient(config, &http.Client{}, tlog.NewTemporalLogFromContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("can't create elasticsearch client: %w", err)
	}

	return c.(esclient.IntegrationTestsClient), nil
}

func (m *Manager) getCassandraConnectionFromDatastoreSpec(ctx context.Context, store *v1alpha1.TemporalDatastoreSpec, namespace string) (*cassandra.CQLClient, error) {
	config := NewCassandraConfigFromDatastoreSpec(store)
	var err error
	config.Password, err = m.getStorePassword(ctx, store, namespace)
	if err != nil {
		return nil, err
	}

	cfg := &cassandra.CQLClientConfig{
		Hosts:                    config.Hosts,
		Port:                     config.Port,
		User:                     config.User,
		Password:                 config.Password,
		Keyspace:                 config.Keyspace,
		Timeout:                  int(config.ConnectTimeout),
		Datacenter:               config.Datacenter,
		TLS:                      config.TLS,
		DisableInitialHostLookup: config.DisableInitialHostLookup,
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 10
	}

	if config.Consistency != nil && config.Consistency.Default != nil {
		cfg.Consistency = config.Consistency.Default.Consistency
	}

	return cassandra.NewCQLClient(cfg, tlog.NewTemporalLogFromContext(ctx))
}

func (m *Manager) getStorePassword(ctx context.Context, store *v1alpha1.TemporalDatastoreSpec, namespace string) (string, error) {
	passwordSecret := &corev1.Secret{}
	err := m.Get(ctx, types.NamespacedName{Name: store.PasswordSecretRef.Name, Namespace: namespace}, passwordSecret)
	if err != nil {
		return "", err
	}
	password, ok := passwordSecret.Data[store.PasswordSecretRef.Key]
	if !ok {
		return "", fmt.Errorf("key '%s' not found in secret %s", store.PasswordSecretRef.Key, store.PasswordSecretRef.Name)
	}
	return string(password), nil
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
		storeSchemaPath = cassandraSchemaPath
		storeVersionSchemaPath = cassandraVersionSchemaPath
	}

	tagetSchemaPath := defaultSchemaPath
	if targetSchema == VisibilitySchema {
		tagetSchemaPath = visibilitySchemaPath
	}

	return path.Join(m.SchemaFilePath, storeSchemaPath, storeVersionSchemaPath, tagetSchemaPath, "versioned")
}

func (m *Manager) computeAdvancedVisibilitySchemaDir(esVersion string, templateVersion semver.Version) string {
	v := fmt.Sprintf("v%d", templateVersion.Major)
	file := fmt.Sprintf("index_template_%s.json", esVersion)
	return path.Join(m.SchemaFilePath, elasticsearchSchemaPath, visibilitySchemaPath, "versioned", v, file)
}

func (m *Manager) runUpdateSchemaTasks(ctx context.Context, cluster *v1alpha1.TemporalCluster, store *v1alpha1.TemporalDatastoreSpec, targetSchema Schema, targetVersion semver.Version) error {
	conn, err := m.getDatastoreConnection(ctx, store, cluster.Namespace)
	if err != nil {
		return err
	}
	defer conn.Close()

	datastoreType, err := store.GetDatastoreType()
	if err != nil {
		return err
	}

	updateTask := schema.NewUpdateSchemaTask(conn, &schema.UpdateConfig{
		TargetVersion: fmt.Sprintf("v%d.%d", targetVersion.Major, targetVersion.Minor),
		SchemaDir:     m.computeSchemaDir(datastoreType, targetSchema),
		IsDryRun:      false,
	}, tlog.NewTemporalLogFromContext(ctx))

	err = updateTask.Run()
	if err != nil {
		return err
	}

	return nil
}
