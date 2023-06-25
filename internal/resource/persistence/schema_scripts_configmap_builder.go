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
	"bytes"
	"fmt"
	"net"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/elliotchance/orderedmap/v2"
	"go.temporal.io/server/tools/common/schema"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Schema string

const (
	DefaultSchema    Schema = "default"
	VisibilitySchema Schema = "visibility"

	defaultSchemaPath    = "temporal"
	visibilitySchemaPath = "visibility"

	postgreSQLSchemaPath          = "postgresql"
	postgreSQLVersionSchemaPath   = "v96"
	postgreSQL12VersionSchemaPath = "v12"

	mysqlSchemaPath         = "mysql"
	mysqlVersionSchemaPath  = "v57"
	mysql8VersionSchemaPath = "v8"

	cassandraSchemaPath        = "cassandra"
	cassandraVersionSchemaPath = ""

	elasticsearchSchemaPath        = "elasticsearch"
	elasticsearchVersionSchemaPath = ""
)

type SchemaScriptsConfigmapBuilder struct {
	instance *v1beta1.TemporalCluster
	scheme   *runtime.Scheme
}

func NewSchemaScriptsConfigmapBuilder(instance *v1beta1.TemporalCluster, scheme *runtime.Scheme) *SchemaScriptsConfigmapBuilder {
	return &SchemaScriptsConfigmapBuilder{
		instance: instance,
		scheme:   scheme,
	}
}

func (b *SchemaScriptsConfigmapBuilder) Build() client.Object {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName("schema-scripts"),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance, "schema-scripts", b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}
}

func (b *SchemaScriptsConfigmapBuilder) Enabled() bool {
	return true
}

func (b *SchemaScriptsConfigmapBuilder) computeSchemaDir(storeType v1beta1.DatastoreType, targetSchema Schema) string {
	storeSchemaPath := ""
	storeVersionSchemaPath := ""
	switch storeType {
	case v1beta1.PostgresSQLDatastore:
		storeSchemaPath = postgreSQLSchemaPath
		storeVersionSchemaPath = postgreSQLVersionSchemaPath
	case v1beta1.PostgresSQL12Datastore:
		storeSchemaPath = postgreSQLSchemaPath
		storeVersionSchemaPath = postgreSQL12VersionSchemaPath
	case v1beta1.MySQLDatastore:
		storeSchemaPath = mysqlSchemaPath
		storeVersionSchemaPath = mysqlVersionSchemaPath
	case v1beta1.MySQL8Datastore:
		storeSchemaPath = mysqlSchemaPath
		storeVersionSchemaPath = mysql8VersionSchemaPath
	case v1beta1.CassandraDatastore:
		storeSchemaPath = cassandraSchemaPath
		storeVersionSchemaPath = cassandraVersionSchemaPath
	case v1beta1.ElasticsearchDatastore:
		storeSchemaPath = elasticsearchSchemaPath
		storeVersionSchemaPath = elasticsearchVersionSchemaPath
	case v1beta1.UnknownDatastore:
		storeSchemaPath = ""
		storeVersionSchemaPath = ""
	}

	tagetSchemaPath := defaultSchemaPath
	if targetSchema == VisibilitySchema {
		tagetSchemaPath = visibilitySchemaPath
	}

	return path.Join("/etc/temporal/schema", storeSchemaPath, storeVersionSchemaPath, tagetSchemaPath, "versioned")
}

func (b *SchemaScriptsConfigmapBuilder) argsMapToString(m *orderedmap.OrderedMap[string, string]) string {
	cmd := []string{}
	for el := m.Front(); el != nil; el = el.Next() {
		if el.Value != "" {
			cmd = append(cmd, fmt.Sprintf("--%s=%s", el.Key, el.Value))
		} else {
			cmd = append(cmd, fmt.Sprintf("--%s", el.Key))
		}
	}
	return strings.Join(cmd, " ")
}

func (b *SchemaScriptsConfigmapBuilder) getSQLArgs(spec *v1beta1.DatastoreSpec) (*orderedmap.OrderedMap[string, string], error) {
	host, port, err := net.SplitHostPort(spec.SQL.ConnectAddr)
	if err != nil {
		return nil, fmt.Errorf("can't parse host port: %w", err)
	}

	args := orderedmap.NewOrderedMap[string, string]()
	args.Set(schema.CLIOptEndpoint, host)                                             // --endpoint
	args.Set(schema.CLIOptPort, port)                                                 // --port
	args.Set(schema.CLIOptUser, spec.SQL.User)                                        // --user
	args.Set(schema.CLIOptPassword, fmt.Sprintf("$%s", spec.GetPasswordEnvVarName())) // --password
	args.Set(schema.CLIOptDatabase, spec.SQL.DatabaseName)                            // --database
	args.Set(schema.CLIOptPluginName, spec.SQL.PluginName)                            // --plugin
	// TODO(alexandrevilain): support schema.CLIOptTimeout

	if len(spec.SQL.ConnectAttributes) > 0 {
		attributes := url.Values{}
		for k, v := range spec.SQL.ConnectAttributes {
			attributes.Add(k, v)
		}
		args.Set(schema.CLIOptConnectAttributes, attributes.Encode())
	}

	return args, nil
}

func (b *SchemaScriptsConfigmapBuilder) getCassandraArgs(spec *v1beta1.DatastoreSpec) (*orderedmap.OrderedMap[string, string], error) {
	// schema.CLIOptReplicationFactor because it's set at the keyspace creation
	// the script doesn't create the keyspace.
	args := orderedmap.NewOrderedMap[string, string]()
	args.Set(schema.CLIOptEndpoint, strings.Join(spec.Cassandra.Hosts, ","))
	args.Set(schema.CLIOptPort, strconv.Itoa(spec.Cassandra.Port))
	args.Set(schema.CLIOptUser, spec.Cassandra.User)
	args.Set(schema.CLIOptPassword, fmt.Sprintf("$%s", spec.GetPasswordEnvVarName()))
	args.Set(schema.CLIOptKeyspace, spec.Cassandra.Keyspace)
	args.Set(schema.CLIOptDatacenter, spec.Cassandra.Datacenter)
	if spec.Cassandra.Consistency != nil {
		args.Set(schema.CLIOptConsistency, spec.Cassandra.Consistency.Consistency.String())
	}
	// TODO(alexandrevilain): support schema.CLIOptTimeout
	// TODO(alexandrevilain): support schema.CLIOptAddressTranslator & schema.CLIOptAddressTranslatorOptions

	if spec.Cassandra.DisableInitialHostLookup {
		args.Set(schema.CLIFlagDisableInitialHostLookup, "")
	}

	return args, nil
}

func (b *SchemaScriptsConfigmapBuilder) getStoreArgs(spec *v1beta1.DatastoreSpec) (*orderedmap.OrderedMap[string, string], error) {
	var args *orderedmap.OrderedMap[string, string]
	var err error

	switch spec.GetType() {
	case v1beta1.CassandraDatastore:
		args, err = b.getCassandraArgs(spec)
		if err != nil {
			return nil, err
		}
	case v1beta1.PostgresSQLDatastore,
		v1beta1.PostgresSQL12Datastore,
		v1beta1.MySQLDatastore,
		v1beta1.MySQL8Datastore:
		args, err = b.getSQLArgs(spec)
		if err != nil {
			return nil, err
		}
	case v1beta1.ElasticsearchDatastore, v1beta1.UnknownDatastore:
		return nil, fmt.Errorf("unsupported datastore: %s", spec.GetType())
	}

	if spec.TLS != nil && spec.TLS.Enabled {
		args.Set(schema.CLIFlagEnableTLS, "")
		if spec.TLS.CaFileRef != nil {
			args.Set(schema.CLIFlagTLSCaFile, spec.GetTLSCaFileMountPath())
		}

		if spec.TLS.CertFileRef != nil {
			args.Set(schema.CLIFlagTLSCertFile, spec.GetTLSCertFileMountPath())
		}

		if spec.TLS.KeyFileRef != nil {
			args.Set(schema.CLIFlagTLSKeyFile, spec.GetTLSKeyFileMountPath())
		}

		if !spec.TLS.EnableHostVerification {
			args.Set(schema.CLIFlagTLSDisableHostVerification, "")
		}

		if spec.TLS.ServerName != "" {
			args.Set(schema.CLIFlagTLSHostName, spec.TLS.ServerName)
		}
	}

	return args, nil
}

func (b *SchemaScriptsConfigmapBuilder) Update(object client.Object) error {
	configMap := object.(*corev1.ConfigMap)

	defaultStore := b.instance.Spec.Persistence.DefaultStore
	visibilityStore := b.instance.Spec.Persistence.VisibilityStore

	defaultStoreType := defaultStore.GetType()
	visibilityStoreType := visibilityStore.GetType()

	defaultStoreArgs, err := b.getStoreArgs(defaultStore)
	if err != nil {
		return fmt.Errorf("can't create default store args: %w", err)
	}

	visibilityStoreArgs, err := b.getStoreArgs(visibilityStore)
	if err != nil {
		return fmt.Errorf("can't create visibility store setup args: %w", err)
	}

	baseData := baseData{}

	if b.instance.Spec.MTLS != nil {
		baseData.MTLSProvider = string(b.instance.Spec.MTLS.Provider)
	}

	createDatabaseTemplateKey := CreateDatabaseTemplate
	if b.instance.Spec.Version.GreaterOrEqual(version.V1_18_0) {
		createDatabaseTemplateKey = CreateDatabaseTemplateV1_18
	}

	defaultStoreTool := "temporal-sql-tool"
	var renderedCreateDefaultDatabase bytes.Buffer
	var renderedCreateVisibilityDatabase bytes.Buffer
	if defaultStoreType == v1beta1.CassandraDatastore {
		// Fix for https://github.com/temporalio/temporal/blob/master/tools/cassandra/main.go#L70
		// Which requires an env var set.
		defaultStoreTool = "CASSANDRA_PORT=9042 temporal-cassandra-tool"

		err = templates[CreateCassandraTemplate].Execute(&renderedCreateDefaultDatabase, createKeyspace{
			baseData:       baseData,
			Tool:           defaultStoreTool,
			ConnectionArgs: b.argsMapToString(defaultStoreArgs),
			KeyspaceName:   defaultStore.Cassandra.Keyspace,
		})
		if err != nil {
			return fmt.Errorf("can't render create-default-keyspace.sh: %w", err)
		}
	} else {
		err = templates[createDatabaseTemplateKey].Execute(&renderedCreateDefaultDatabase, createDatabase{
			baseData:       baseData,
			Tool:           defaultStoreTool,
			ConnectionArgs: b.argsMapToString(defaultStoreArgs),
			DatabaseName:   defaultStore.SQL.DatabaseName,
		})
		if err != nil {
			return fmt.Errorf("can't render create-default-database.sh: %w", err)
		}
	}

	visibilityStoreTool := "temporal-sql-tool"
	if visibilityStoreType == v1beta1.CassandraDatastore {
		visibilityStoreTool = "CASSANDRA_PORT=9042 temporal-cassandra-tool"

		err = templates[CreateCassandraTemplate].Execute(&renderedCreateVisibilityDatabase, createKeyspace{
			baseData:       baseData,
			Tool:           visibilityStoreTool,
			ConnectionArgs: b.argsMapToString(defaultStoreArgs),
			KeyspaceName:   visibilityStore.Cassandra.Keyspace,
		})
		if err != nil {
			return fmt.Errorf("can't render create-visibility-keyspace.sh: %w", err)
		}
	} else {
		err = templates[createDatabaseTemplateKey].Execute(&renderedCreateVisibilityDatabase, createDatabase{
			baseData:       baseData,
			Tool:           visibilityStoreTool,
			ConnectionArgs: b.argsMapToString(visibilityStoreArgs),
			DatabaseName:   visibilityStore.SQL.DatabaseName,
		})
		if err != nil {
			return fmt.Errorf("can't render create-visibility-database.sh: %w", err)
		}
	}

	var renderedSetupDefaultSchema bytes.Buffer
	err = templates[SetupSchemaTemplate].Execute(&renderedSetupDefaultSchema, setupSchemaData{
		baseData:       baseData,
		Tool:           defaultStoreTool,
		ConnectionArgs: b.argsMapToString(defaultStoreArgs),
		InitialVersion: "0.0",
	})
	if err != nil {
		return fmt.Errorf("can't render setup-default-schema.sh: %w", err)
	}

	var renderedUpdateDefaultSchema bytes.Buffer
	err = templates[UpdateSchemaTemplate].Execute(&renderedUpdateDefaultSchema, updateSchemaData{
		baseData:       baseData,
		Tool:           defaultStoreTool,
		ConnectionArgs: b.argsMapToString(defaultStoreArgs),
		SchemaDir:      b.computeSchemaDir(defaultStoreType, DefaultSchema),
	})
	if err != nil {
		return fmt.Errorf("can't render update-default-schema.sh: %w", err)
	}

	var renderedSetupVisibilitySchema bytes.Buffer
	err = templates[SetupSchemaTemplate].Execute(&renderedSetupVisibilitySchema, setupSchemaData{
		baseData:       baseData,
		Tool:           visibilityStoreTool,
		ConnectionArgs: b.argsMapToString(visibilityStoreArgs),
		InitialVersion: "0.0",
	})
	if err != nil {
		return fmt.Errorf("can't render setup-visibility-schema.sh: %w", err)
	}

	var renderedUpdateVisibilitySchema bytes.Buffer
	err = templates[UpdateSchemaTemplate].Execute(&renderedUpdateVisibilitySchema, updateSchemaData{
		baseData:       baseData,
		Tool:           visibilityStoreTool,
		ConnectionArgs: b.argsMapToString(visibilityStoreArgs),
		SchemaDir:      b.computeSchemaDir(visibilityStoreType, VisibilitySchema),
	})
	if err != nil {
		return fmt.Errorf("can't render update-visibility-schema.sh: %w", err)
	}

	configMap.Data = map[string]string{
		"create-default-database.sh":    renderedCreateDefaultDatabase.String(),
		"create-visibility-database.sh": renderedCreateVisibilityDatabase.String(),
		"setup-default-schema.sh":       renderedSetupDefaultSchema.String(),
		"setup-visibility-schema.sh":    renderedSetupVisibilitySchema.String(),
		"update-default-schema.sh":      renderedUpdateDefaultSchema.String(),
		"update-visibility-schema.sh":   renderedUpdateVisibilitySchema.String(),
	}

	advancedVisibilityStore := b.instance.Spec.Persistence.AdvancedVisibilityStore
	if advancedVisibilityStore != nil {
		version := advancedVisibilityStore.Elasticsearch.Version
		if version == "v8" {
			// For now, when elasticsearch version 8 is specified, it uses v7 schema.
			// See: https://github.com/temporalio/temporal/tree/v1.20.3/schema/elasticsearch/visibility
			version = "v7"
		}
		data := advancedVisibilityData{
			baseData:       baseData,
			Version:        version,
			URL:            advancedVisibilityStore.Elasticsearch.URL,
			Username:       advancedVisibilityStore.Elasticsearch.Username,
			PasswordEnvVar: advancedVisibilityStore.GetPasswordEnvVarName(),
			Indices:        advancedVisibilityStore.Elasticsearch.Indices,
		}
		var renderedSetupAdvancedVisibility bytes.Buffer
		err = templates[SetupAdvancedVisibility].Execute(&renderedSetupAdvancedVisibility, data)
		if err != nil {
			return fmt.Errorf("can't render setup-advanced-visibility.sh: %w", err)
		}

		var renderedUpdateAdvancedVisibility bytes.Buffer
		err = templates[UpdateAdvancedVisibility].Execute(&renderedUpdateAdvancedVisibility, data)
		if err != nil {
			return fmt.Errorf("can't render setup-advanced-visibility.sh: %w", err)
		}

		configMap.Data["setup-advanced-visibility.sh"] = renderedSetupAdvancedVisibility.String()
		configMap.Data["update-advanced-visibility.sh"] = renderedUpdateAdvancedVisibility.String()
	}

	if err := controllerutil.SetControllerReference(b.instance, configMap, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %w", err)
	}

	return nil
}
