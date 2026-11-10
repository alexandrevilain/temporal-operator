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
	"testing"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShellQuote(t *testing.T) {
	assert.Equal(t, `'foo'`, shellQuote("foo"))
	assert.Equal(t, `'a b'`, shellQuote("a b"))
	// embedded single quote is escaped
	assert.Equal(t, `'it'\''s'`, shellQuote("it's"))
}

// TestGetSQLArgs_PasswordCommand asserts that a datastore using an external
// passwordCommand renders the schema tool --password flag as a shell command
// substitution (resolved at runtime by the generated setup script), instead of
// referencing a password environment variable.
func TestGetSQLArgs_PasswordCommand(t *testing.T) {
	b := &SchemaScriptsConfigmapBuilder{}

	spec := &v1beta1.DatastoreSpec{
		Name: "default",
		SQL: &v1beta1.SQLSpec{
			User:         "temporal",
			PluginName:   "postgres12",
			DatabaseName: "temporal",
			ConnectAddr:  "postgres:5432",
			PasswordCommand: &v1beta1.SQLPasswordCommandSpec{
				Command: "/bin/sh",
				Args:    []string{"-c", "printf %s test"},
			},
		},
	}

	args, err := b.getSQLArgs(spec)
	require.NoError(t, err)

	rendered := b.argsMapToString(args)
	assert.Contains(t, rendered, `--password="$('/bin/sh' '-c' 'printf %s test')"`,
		"passwordCommand must render as a shell command substitution")
	assert.NotContains(t, rendered, "PASSWORD",
		"passwordCommand must not reference a password env var")
}

// TestGetSQLArgs_PasswordSecretRef keeps asserting the classic secret-based path
// still renders a password env var reference.
func TestGetSQLArgs_PasswordSecretRef(t *testing.T) {
	b := &SchemaScriptsConfigmapBuilder{}

	spec := &v1beta1.DatastoreSpec{
		Name: "default",
		SQL: &v1beta1.SQLSpec{
			User:         "temporal",
			PluginName:   "postgres12",
			DatabaseName: "temporal",
			ConnectAddr:  "postgres:5432",
		},
		PasswordSecretRef: &v1beta1.SecretKeyReference{Name: "postgres-password", Key: "PASSWORD"},
	}

	args, err := b.getSQLArgs(spec)
	require.NoError(t, err)

	rendered := b.argsMapToString(args)
	assert.Contains(t, rendered, "--password=\"$"+spec.GetPasswordEnvVarName()+"\"")
}

func esVisibilityStore() *v1beta1.DatastoreSpec {
	return &v1beta1.DatastoreSpec{
		Name: "visibility",
		Elasticsearch: &v1beta1.ElasticsearchSpec{
			URL:      "http://elasticsearch:9200",
			Username: "elastic",
			Indices:  v1beta1.ElasticsearchIndices{Visibility: "temporal_visibility_v1_dev"},
		},
		PasswordSecretRef: &v1beta1.SecretKeyReference{Name: "es-password", Key: "PASSWORD"},
	}
}

func esBuilder(v string) *SchemaScriptsConfigmapBuilder {
	return &SchemaScriptsConfigmapBuilder{
		instance: &v1beta1.TemporalCluster{
			Spec: v1beta1.TemporalClusterSpec{
				Version: version.MustNewVersionFromString(v),
			},
		},
	}
}

// TestESVisibility_Tool_Post130 asserts that on Temporal >= 1.30 the ES visibility
// setup/update scripts drive temporal-elasticsearch-tool (curl/jq were removed from
// the admin-tools image) instead of curl.
func TestESVisibility_Tool_Post130(t *testing.T) {
	b := esBuilder("1.30.5")
	store := esVisibilityStore()

	setup, err := b.GetStoreSetupTemplate(store)
	require.NoError(t, err)
	assert.Contains(t, setup, "temporal-elasticsearch-tool")
	assert.Contains(t, setup, "setup-schema")
	assert.Contains(t, setup, `create-index --index "temporal_visibility_v1_dev"`)
	assert.Contains(t, setup, `--endpoint="http://elasticsearch:9200"`)
	assert.Contains(t, setup, `--user="elastic"`)
	assert.Contains(t, setup, "--password=\"$"+store.GetPasswordEnvVarName()+"\"")
	assert.NotContains(t, setup, "curl")

	update, err := b.GetStoreUpdateTemplate(store, VisibilitySchema)
	require.NoError(t, err)
	assert.Contains(t, update, "temporal-elasticsearch-tool")
	assert.Contains(t, update, `update-schema --index "temporal_visibility_v1_dev"`)
	assert.NotContains(t, update, "curl")
}

// TestESVisibility_Curl_Pre130 asserts that on Temporal < 1.30 the legacy curl-based
// scripts are still generated (older admin-tools images ship curl and lack the tool).
func TestESVisibility_Curl_Pre130(t *testing.T) {
	b := esBuilder("1.29.7")
	store := esVisibilityStore()

	setup, err := b.GetStoreSetupTemplate(store)
	require.NoError(t, err)
	assert.Contains(t, setup, "curl")
	assert.Contains(t, setup, "_template")
	assert.NotContains(t, setup, "temporal-elasticsearch-tool")
}

// TestESVisibility_EmptyUsername_NoBareUserFlag guards a silent-failure mode.
//
// argsMapToString renders an empty value as a bare "--user", and the tool's
// flag parser (urfave/cli v1 over the stdlib flag package) then takes the
// following token as that flag's value. For the setup script that token is the
// "setup-schema" subcommand itself, so the tool finds no command at all, prints
// its help and exits 0 — the schema job is recorded as successful while neither
// the index template nor the visibility index was ever created.
//
// An empty username is what an auth-less Elasticsearch needs and the CRD allows
// it (required, but with no minimum length), so this is reachable.
func TestESVisibility_EmptyUsername_NoBareUserFlag(t *testing.T) {
	b := esBuilder("1.30.5")
	store := esVisibilityStore()
	store.Elasticsearch.Username = ""
	store.PasswordSecretRef = nil

	setup, err := b.GetStoreSetupTemplate(store)
	require.NoError(t, err)
	assert.NotRegexp(t, `--user(\s|$)`, setup, "empty username must not render a bare --user flag")
	assert.NotRegexp(t, `--password(\s|$)`, setup, "absent password must not render a bare --password flag")
	assert.Contains(t, setup, "setup-schema")

	update, err := b.GetStoreUpdateTemplate(store, VisibilitySchema)
	require.NoError(t, err)
	assert.NotRegexp(t, `--user(\s|$)`, update)
}

// TestSQLAndCassandraArgs_EmptyUser_NoBareUserFlag covers the same rendering
// hazard on the other two datastore families.
func TestSQLAndCassandraArgs_EmptyUser_NoBareUserFlag(t *testing.T) {
	b := &SchemaScriptsConfigmapBuilder{}

	sqlArgs, err := b.getSQLArgs(&v1beta1.DatastoreSpec{
		Name: "default",
		SQL: &v1beta1.SQLSpec{
			User:         "",
			PluginName:   "postgres12",
			DatabaseName: "temporal",
			ConnectAddr:  "postgres:5432",
		},
	})
	require.NoError(t, err)
	assert.NotRegexp(t, `--user(\s|$)`, b.argsMapToString(sqlArgs))

	cassandraArgs := b.getCassandraArgs(&v1beta1.DatastoreSpec{
		Name:      "default",
		Cassandra: &v1beta1.CassandraSpec{Hosts: []string{"cassandra"}, Port: 9042, User: "", Keyspace: "temporal"},
	})
	assert.NotRegexp(t, `--user(\s|$)`, b.argsMapToString(cassandraArgs))
}

// TestGetStoreTool_ElasticsearchPre130 asserts the empty sentinel is preserved
// below 1.30. temporal-elasticsearch-tool does not exist in those admin-tools
// images, so a caller that reaches getStoreTool without checking the version
// must not be handed a binary name that cannot be executed.
func TestGetStoreTool_ElasticsearchPre130(t *testing.T) {
	assert.Empty(t, esBuilder("1.29.7").getStoreTool(v1beta1.ElasticsearchDatastore))
	assert.Equal(t, "temporal-elasticsearch-tool", esBuilder("1.30.5").getStoreTool(v1beta1.ElasticsearchDatastore))
}

// TestESVisibility_Tool_ShutdownFooterReachableOnFailure asserts the generated
// script cannot abort before the service-mesh shutdown footer.
//
// With "set -e", a failing temporal-elasticsearch-tool exits the script
// immediately and the footer never posts to the linkerd proxy's /shutdown. The
// sidecar then keeps the Job pod Running indefinitely rather than letting it
// fail and retry, and persistence reconciliation blocks on that job forever.
func TestESVisibility_Tool_ShutdownFooterReachableOnFailure(t *testing.T) {
	b := esBuilder("1.30.5")
	b.instance.Spec.MTLS = &v1beta1.MTLSSpec{Provider: v1beta1.LinkerdMTLSProvider}
	store := esVisibilityStore()

	for name, script := range map[string]func() (string, error){
		"setup":  func() (string, error) { return b.GetStoreSetupTemplate(store) },
		"update": func() (string, error) { return b.GetStoreUpdateTemplate(store, VisibilitySchema) },
	} {
		t.Run(name, func(t *testing.T) {
			rendered, err := script()
			require.NoError(t, err)

			assert.NotRegexp(t, `(?m)^\s*set -e`, rendered,
				"set -e would skip the sidecar shutdown footer when a step fails")
			assert.Contains(t, rendered, "localhost:4191/shutdown", "shutdown footer must be present")
			assert.Contains(t, rendered, "exit $x", "the tool's exit status must still be propagated")
			// The wget call must not swallow its own status either: both mesh
			// providers behave identically here.
			assert.NotContains(t, rendered, "|| true")
		})
	}
}

// TestESVisibility_Tool_StepsFailFast asserts the steps are chained so a failed
// setup-schema does not let create-index run (and report success) anyway.
func TestESVisibility_Tool_StepsFailFast(t *testing.T) {
	b := esBuilder("1.30.5")
	store := esVisibilityStore()
	store.Elasticsearch.Indices.SecondaryVisibility = "temporal_visibility_v1_dev_secondary"

	setup, err := b.GetStoreSetupTemplate(store)
	require.NoError(t, err)
	assert.Contains(t, setup, "setup-schema && \\")
	assert.Contains(t, setup, `create-index --index "temporal_visibility_v1_dev" && \`)
	assert.Contains(t, setup, `create-index --index "temporal_visibility_v1_dev_secondary"`)
}
