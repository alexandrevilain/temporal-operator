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

	CreateDefaultDatabaseScript             = "create-default-database.sh"
	SetupDefaultSchemaScript                = "setup-default-schema.sh"
	UpdateDefaultSchemaScript               = "update-default-schema.sh"
	CreateVisibilityDatabaseScript          = "create-visibility-database.sh"
	SetupVisibilitySchemaScript             = "setup-visibility-schema.sh"
	UpdateVisibilitySchemaScript            = "update-visibility-schema.sh"
	CreateSecondaryVisibilityDatabaseScript = "create-secondary-visibility-database.sh"
	SetupSecondaryVisibilitySchemaScript    = "setup-secondary-visibility-schema.sh"
	UpdateSecondaryVisibilitySchemaScript   = "update-secondary-visibility-schema.sh"
	CreateAdvancedVisibilityDatabaseScript  = "create-advanced-visibility-database.sh"
	SetupAdvancedVisibilitySchemaScript     = "setup-advanced-visibility-schema.sh"
	UpdateAdvancedVisibilitySchemaScript    = "update-advanced-visibility-schema.sh"

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

func (b *SchemaScriptsConfigmapBuilder) baseData() baseData {
	baseData := baseData{}

	if b.instance.Spec.MTLS != nil {
		baseData.MTLSProvider = string(b.instance.Spec.MTLS.Provider)
	}

	// Temporal >= 1.30 dropped curl from the admin-tools image; the MTLS
	// sidecar-shutdown footer must use busybox wget instead.
	baseData.UseWget = b.instance.Spec.Version.GreaterOrEqual(version.V1_30_0)

	return baseData
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
			cmd = append(cmd, fmt.Sprintf(`--%s="%s"`, el.Key, el.Value))
		} else {
			cmd = append(cmd, fmt.Sprintf("--%s", el.Key))
		}
	}
	return strings.Join(cmd, " ")
}

// shellCommand renders an external password command and its arguments as a
// single POSIX-shell string suitable for command substitution "$( ... )".
func shellCommand(pc *v1beta1.SQLPasswordCommandSpec) string {
	parts := make([]string, 0, len(pc.Args)+1)
	parts = append(parts, shellQuote(pc.Command))
	for _, a := range pc.Args {
		parts = append(parts, shellQuote(a))
	}
	return strings.Join(parts, " ")
}

// shellQuote wraps s in single quotes, escaping any embedded single quotes, so
// it is safe to embed in a shell command line.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func (b *SchemaScriptsConfigmapBuilder) getSQLArgs(spec *v1beta1.DatastoreSpec) (*orderedmap.OrderedMap[string, string], error) {
	host, port, err := net.SplitHostPort(spec.SQL.ConnectAddr)
	if err != nil {
		return nil, fmt.Errorf("can't parse host port: %w", err)
	}

	args := orderedmap.NewOrderedMap[string, string]()
	args.Set(schema.CLIOptEndpoint, host) // --endpoint
	args.Set(schema.CLIOptPort, port)     // --port
	// Omit --user when empty: a bare flag would swallow the next token as its
	// value. See getElasticsearchArgs for the full explanation.
	if spec.SQL.User != "" {
		args.Set(schema.CLIOptUser, spec.SQL.User) // --user
	}
	switch {
	case spec.PasswordSecretRef != nil:
		args.Set(schema.CLIOptPassword, fmt.Sprintf("$%s", spec.GetPasswordEnvVarName())) // --password
	case spec.SQL.PasswordCommand != nil:
		// The datastore resolves its password by running an external command
		// (Temporal >= 1.31). The schema tool has no equivalent flag, so the
		// generated shell script resolves it via command substitution at runtime.
		args.Set(schema.CLIOptPassword, fmt.Sprintf("$(%s)", shellCommand(spec.SQL.PasswordCommand))) // --password
	}
	args.Set(schema.CLIOptDatabase, spec.SQL.DatabaseName) // --database
	args.Set(schema.CLIOptPluginName, spec.SQL.PluginName) // --plugin
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

func (b *SchemaScriptsConfigmapBuilder) getCassandraArgs(spec *v1beta1.DatastoreSpec) *orderedmap.OrderedMap[string, string] {
	// schema.CLIOptReplicationFactor because it's set at the keyspace creation
	// the script doesn't create the keyspace.
	args := orderedmap.NewOrderedMap[string, string]()
	args.Set(schema.CLIOptEndpoint, strings.Join(spec.Cassandra.Hosts, ","))
	args.Set(schema.CLIOptPort, strconv.Itoa(spec.Cassandra.Port))
	// Omit --user when empty: a bare flag would swallow the next token as its
	// value. See getElasticsearchArgs for the full explanation.
	if spec.Cassandra.User != "" {
		args.Set(schema.CLIOptUser, spec.Cassandra.User)
	}
	args.Set(schema.CLIOptPassword, fmt.Sprintf("$%s", spec.GetPasswordEnvVarName()))
	args.Set(schema.CLIOptKeyspace, spec.Cassandra.Keyspace)
	if spec.Cassandra.Datacenter != "" {
		args.Set(schema.CLIOptDatacenter, spec.Cassandra.Datacenter)
	}

	if spec.Cassandra.Consistency != nil {
		args.Set(schema.CLIOptConsistency, spec.Cassandra.Consistency.Consistency.String())
	}
	// TODO(alexandrevilain): support schema.CLIOptTimeout
	// TODO(alexandrevilain): support schema.CLIOptAddressTranslator & schema.CLIOptAddressTranslatorOptions

	if spec.Cassandra.DisableInitialHostLookup {
		args.Set(schema.CLIFlagDisableInitialHostLookup, "")
	}

	return args
}

// getElasticsearchArgs builds the connection flags for temporal-elasticsearch-tool
// (Temporal >= 1.30). TLS flags are appended by the shared block in getStoreArgs.
func (b *SchemaScriptsConfigmapBuilder) getElasticsearchArgs(spec *v1beta1.DatastoreSpec) *orderedmap.OrderedMap[string, string] {
	args := orderedmap.NewOrderedMap[string, string]()
	args.Set(schema.CLIOptEndpoint, spec.Elasticsearch.URL) // --endpoint
	// Only set --user when it has a value. argsMapToString renders an empty
	// value as a bare "--user", and the tool's flag parser (urfave/cli v1, on
	// top of the stdlib flag package) then takes the *following* token as the
	// flag's value. For "setup-schema" that swallows the subcommand itself:
	// the tool finds no command, prints its help and exits 0, so the job is
	// reported successful while neither the index template nor the index was
	// ever created. An empty username is what an auth-less Elasticsearch needs,
	// and the CRD permits it, so this is reachable.
	if spec.Elasticsearch.Username != "" {
		args.Set(schema.CLIOptUser, spec.Elasticsearch.Username) // --user
	}
	if spec.PasswordSecretRef != nil {
		args.Set(schema.CLIOptPassword, fmt.Sprintf("$%s", spec.GetPasswordEnvVarName())) // --password
	}
	return args
}

func (b *SchemaScriptsConfigmapBuilder) getStoreArgs(spec *v1beta1.DatastoreSpec) (*orderedmap.OrderedMap[string, string], error) {
	var args *orderedmap.OrderedMap[string, string]
	var err error

	switch spec.GetType() {
	case v1beta1.CassandraDatastore:
		args = b.getCassandraArgs(spec)
	case v1beta1.PostgresSQLDatastore,
		v1beta1.PostgresSQL12Datastore,
		v1beta1.MySQLDatastore,
		v1beta1.MySQL8Datastore:
		args, err = b.getSQLArgs(spec)
		if err != nil {
			return nil, err
		}
	case v1beta1.ElasticsearchDatastore:
		args = b.getElasticsearchArgs(spec)
	case v1beta1.UnknownDatastore:
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

func (b *SchemaScriptsConfigmapBuilder) getStoreTool(storeType v1beta1.DatastoreType) string {
	var tool string
	switch storeType {
	case v1beta1.PostgresSQLDatastore,
		v1beta1.PostgresSQL12Datastore,
		v1beta1.MySQLDatastore,
		v1beta1.MySQL8Datastore:
		tool = "temporal-sql-tool"
	case v1beta1.CassandraDatastore:
		// Fix for https://github.com/temporalio/temporal/blob/master/tools/cassandra/main.go#L70
		// Which requires an env var set.
		tool = "CASSANDRA_PORT=9042 temporal-cassandra-tool"
	case v1beta1.ElasticsearchDatastore:
		// temporal-elasticsearch-tool only exists in admin-tools >= 1.30. Older
		// images drive Elasticsearch through the curl/jq templates, which do not
		// reference a tool at all, so keep returning the empty sentinel there
		// rather than emitting a binary that is not present in the image.
		if b.esToolAvailable() {
			tool = "temporal-elasticsearch-tool"
		}
	case v1beta1.UnknownDatastore:
		tool = ""
	}
	return tool
}

// esToolAvailable reports whether the cluster's admin-tools image ships
// temporal-elasticsearch-tool. Temporal >= 1.30 removed curl and jq from that
// image and added the tool as the supported way to manage the visibility index.
func (b *SchemaScriptsConfigmapBuilder) esToolAvailable() bool {
	return b.instance.Spec.Version.GreaterOrEqual(version.V1_30_0)
}

func (b *SchemaScriptsConfigmapBuilder) getESVersion(es *v1beta1.ElasticsearchSpec) string {
	version := es.Version
	if version == "v8" {
		// For now, when elasticsearch version 8 is specified, it uses v7 schema.
		// See: https://github.com/temporalio/temporal/tree/v1.20.3/schema/elasticsearch/visibility
		version = "v7"
	}

	return version
}

func (b *SchemaScriptsConfigmapBuilder) renderTemplate(name string, data any) (string, error) {
	var result bytes.Buffer
	err := templates[name].Execute(&result, data)
	if err != nil {
		return "", fmt.Errorf("can't render %s: %w", name, err)
	}

	return result.String(), nil
}

func (b *SchemaScriptsConfigmapBuilder) GetStoreCreateTemplate(spec *v1beta1.DatastoreSpec) (string, error) {
	storeType := spec.GetType()
	if spec.SkipCreate || storeType == v1beta1.ElasticsearchDatastore {
		return b.renderTemplate(noOpTemplate, b.baseData())
	}

	args, err := b.getStoreArgs(spec)
	if err != nil {
		return "", fmt.Errorf("can't get store args: %w", err)
	}

	if storeType == v1beta1.CassandraDatastore {
		data := createKeyspace{
			baseData:       b.baseData(),
			Tool:           b.getStoreTool(storeType),
			ConnectionArgs: b.argsMapToString(args),
			KeyspaceName:   spec.Cassandra.Keyspace,
		}

		return b.renderTemplate(createCassandraTemplate, data)
	}

	createDatabaseTemplateKey := createDatabaseTemplate
	if b.instance.Spec.Version.GreaterOrEqual(version.V1_18_0) {
		createDatabaseTemplateKey = createDatabaseTemplateV1_18
	}

	data := createDatabase{
		baseData:       b.baseData(),
		Tool:           b.getStoreTool(storeType),
		ConnectionArgs: b.argsMapToString(args),
		DatabaseName:   spec.SQL.DatabaseName,
	}

	return b.renderTemplate(createDatabaseTemplateKey, data)
}

func (b *SchemaScriptsConfigmapBuilder) GetStoreSetupTemplate(spec *v1beta1.DatastoreSpec) (string, error) {
	storeType := spec.GetType()
	if storeType == v1beta1.ElasticsearchDatastore {
		// Temporal >= 1.30 uses temporal-elasticsearch-tool (curl/jq removed from the image).
		if b.esToolAvailable() {
			// getStoreArgs routes ES to getElasticsearchArgs and appends the shared
			// TLS flags, so the tool gets --tls/--tls-*-file when the store uses TLS.
			args, err := b.getStoreArgs(spec)
			if err != nil {
				return "", fmt.Errorf("can't get store args: %w", err)
			}
			return b.renderTemplate(setupESVisibilityTool, esToolData{
				baseData:       b.baseData(),
				Tool:           b.getStoreTool(storeType),
				ConnectionArgs: b.argsMapToString(args),
				Indices:        spec.Elasticsearch.Indices,
			})
		}
		data := esSchemaData{
			baseData:       b.baseData(),
			Version:        b.getESVersion(spec.Elasticsearch),
			URL:            spec.Elasticsearch.URL,
			Username:       spec.Elasticsearch.Username,
			PasswordEnvVar: spec.GetPasswordEnvVarName(),
			Indices:        spec.Elasticsearch.Indices,
		}
		return b.renderTemplate(setupESVisibility, data)
	}

	args, err := b.getStoreArgs(spec)
	if err != nil {
		return "", fmt.Errorf("can't get store args: %w", err)
	}

	data := setupSchemaData{
		baseData:       b.baseData(),
		Tool:           b.getStoreTool(storeType),
		ConnectionArgs: b.argsMapToString(args),
		InitialVersion: "0.0",
	}

	return b.renderTemplate(setupSchemaTemplate, data)
}

func (b *SchemaScriptsConfigmapBuilder) GetStoreUpdateTemplate(spec *v1beta1.DatastoreSpec, targetSchema Schema) (string, error) {
	storeType := spec.GetType()
	if storeType == v1beta1.ElasticsearchDatastore {
		// Temporal >= 1.30 uses temporal-elasticsearch-tool (curl/jq removed from the image).
		if b.esToolAvailable() {
			args, err := b.getStoreArgs(spec)
			if err != nil {
				return "", fmt.Errorf("can't get store args: %w", err)
			}
			return b.renderTemplate(updateESVisibilityTool, esToolData{
				baseData:       b.baseData(),
				Tool:           b.getStoreTool(storeType),
				ConnectionArgs: b.argsMapToString(args),
				Indices:        spec.Elasticsearch.Indices,
			})
		}
		data := esSchemaData{
			baseData:       b.baseData(),
			Version:        b.getESVersion(spec.Elasticsearch),
			URL:            spec.Elasticsearch.URL,
			Username:       spec.Elasticsearch.Username,
			PasswordEnvVar: spec.GetPasswordEnvVarName(),
			Indices:        spec.Elasticsearch.Indices,
		}
		return b.renderTemplate(updateESVisibility, data)
	}

	args, err := b.getStoreArgs(spec)
	if err != nil {
		return "", fmt.Errorf("can't get store args: %w", err)
	}

	data := updateSchemaData{
		baseData:       b.baseData(),
		Tool:           b.getStoreTool(storeType),
		ConnectionArgs: b.argsMapToString(args),
		SchemaDir:      b.computeSchemaDir(storeType, targetSchema),
	}

	return b.renderTemplate(updateSchemaTemplate, data)
}

func (b *SchemaScriptsConfigmapBuilder) Update(object client.Object) error {
	configMap := object.(*corev1.ConfigMap)
	configMap.Data = map[string]string{}

	var err error
	configMap.Data[CreateDefaultDatabaseScript], err = b.GetStoreCreateTemplate(b.instance.Spec.Persistence.DefaultStore)
	if err != nil {
		return err
	}

	configMap.Data[SetupDefaultSchemaScript], err = b.GetStoreSetupTemplate(b.instance.Spec.Persistence.DefaultStore)
	if err != nil {
		return err
	}

	configMap.Data[UpdateDefaultSchemaScript], err = b.GetStoreUpdateTemplate(b.instance.Spec.Persistence.DefaultStore, DefaultSchema)
	if err != nil {
		return err
	}

	configMap.Data[CreateVisibilityDatabaseScript], err = b.GetStoreCreateTemplate(b.instance.Spec.Persistence.VisibilityStore)
	if err != nil {
		return err
	}

	configMap.Data[SetupVisibilitySchemaScript], err = b.GetStoreSetupTemplate(b.instance.Spec.Persistence.VisibilityStore)
	if err != nil {
		return err
	}

	configMap.Data[UpdateVisibilitySchemaScript], err = b.GetStoreUpdateTemplate(b.instance.Spec.Persistence.VisibilityStore, VisibilitySchema)
	if err != nil {
		return err
	}

	secondaryVisibilityStore := b.instance.Spec.Persistence.SecondaryVisibilityStore
	if secondaryVisibilityStore != nil {
		configMap.Data[CreateSecondaryVisibilityDatabaseScript], err = b.GetStoreCreateTemplate(secondaryVisibilityStore)
		if err != nil {
			return err
		}

		configMap.Data[SetupSecondaryVisibilitySchemaScript], err = b.GetStoreSetupTemplate(secondaryVisibilityStore)
		if err != nil {
			return err
		}

		configMap.Data[UpdateSecondaryVisibilitySchemaScript], err = b.GetStoreUpdateTemplate(secondaryVisibilityStore, VisibilitySchema)
		if err != nil {
			return err
		}
	}

	advancedVisibilityStore := b.instance.Spec.Persistence.AdvancedVisibilityStore
	if advancedVisibilityStore != nil {
		configMap.Data[CreateAdvancedVisibilityDatabaseScript], err = b.GetStoreCreateTemplate(advancedVisibilityStore)
		if err != nil {
			return err
		}

		configMap.Data[SetupAdvancedVisibilitySchemaScript], err = b.GetStoreSetupTemplate(advancedVisibilityStore)
		if err != nil {
			return err
		}

		configMap.Data[UpdateAdvancedVisibilitySchemaScript], err = b.GetStoreUpdateTemplate(advancedVisibilityStore, VisibilitySchema)
		if err != nil {
			return err
		}
	}

	if err := controllerutil.SetControllerReference(b.instance, configMap, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %w", err)
	}

	return nil
}
