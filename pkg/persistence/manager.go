package persistence

import (
	"context"
	"errors"
	"path"

	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/internal/forked/go.temporal.io/server/tools/common/schema"
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

// RunDefaultStoreSchemaTasks runs the setup and update default schema tasks
// on the provided cluster's default store.
func (m *Manager) RunDefaultStoreSchemaTasks(ctx context.Context, cluster *v1alpha1.TemporalCluster) error {
	defaultStore, found := cluster.GetDefaultDatastore()
	if !found {
		return errors.New("default datastore not found")
	}

	conn, err := m.getSQLConnectionFromDatastoreSpec(ctx, defaultStore, cluster.Namespace)
	if err != nil {
		return err
	}
	defer conn.Close()

	return m.runDatabaseSchemaTasks(ctx, conn, defaultStore, DefaultSchema)
}

// RunVisibilityStoreSchemaTasks runs the setup and update visibility schema tasks
// on the provided cluster's visibility store.
func (m *Manager) RunVisibilityStoreSchemaTasks(ctx context.Context, cluster *v1alpha1.TemporalCluster) error {
	visibilityStore, found := cluster.GetVisibilityDatastore()
	if !found {
		return errors.New("visibility datastore not found")
	}

	conn, err := m.getSQLConnectionFromDatastoreSpec(ctx, visibilityStore, cluster.Namespace)
	if err != nil {
		return err
	}
	defer conn.Close()

	return m.runDatabaseSchemaTasks(ctx, conn, visibilityStore, VisibilitySchema)
}

func (m *Manager) getSQLConnectionFromDatastoreSpec(ctx context.Context, store *v1alpha1.TemporalDatastoreSpec, namespace string) (*sql.Connection, error) {
	config := NewSQLconfigFromDatastoreSpec(store)

	// TODO(alexandrevilain): remove this
	config.ConnectAddr = "localhost:5432"
	// TODO(alexandrevilain): remove this

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

func (m *Manager) runDatabaseSchemaTasks(ctx context.Context, conn *sql.Connection, store *v1alpha1.TemporalDatastoreSpec, targetSchema Schema) error {
	setupTask := schema.NewSetupSchemaTask(conn, &schema.SetupConfig{
		InitialVersion:    "0.0",
		Overwrite:         false,
		DisableVersioning: false,
	}, temporallog.NewNoopLogger())

	err := setupTask.Run()
	if err != nil {
		return err
	}

	datastoreType, err := store.GetDatastoreType()
	if err != nil {
		return err
	}

	updateTask := schema.NewUpdateSchemaTask(conn, &schema.UpdateConfig{
		DBName:        store.SQL.DatabaseName,
		TargetVersion: "",
		SchemaDir:     m.computeSchemaDir(datastoreType, targetSchema),
		IsDryRun:      false,
	}, temporallog.NewNoopLogger())

	err = updateTask.Run()
	if err != nil {
		return err
	}

	return nil
}
