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

package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/resource/config"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	temporalconfig "go.temporal.io/server/common/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func newSQLCluster(t *testing.T, v string) *v1beta1.TemporalCluster {
	t.Helper()

	store := func() *v1beta1.DatastoreSpec {
		return &v1beta1.DatastoreSpec{
			SQL: &v1beta1.SQLSpec{
				User:            "temporal",
				PluginName:      "postgres12",
				DatabaseName:    "temporal",
				ConnectAddr:     "postgres:5432",
				ConnectProtocol: "tcp",
			},
		}
	}

	cluster := &v1beta1.TemporalCluster{
		TypeMeta:   v1beta1.TemporalClusterTypeMeta,
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: v1beta1.TemporalClusterSpec{
			Version:          version.MustNewVersionFromString(v),
			NumHistoryShards: 1,
			Persistence: v1beta1.TemporalPersistenceSpec{
				DefaultStore:    store(),
				VisibilityStore: store(),
			},
		},
	}
	cluster.Spec.Persistence.DefaultStore.Name = "default"
	cluster.Spec.Persistence.VisibilityStore.Name = "visibility"
	cluster.Default()

	return cluster
}

func buildConfigTemplate(t *testing.T, cluster *v1beta1.TemporalCluster) string {
	t.Helper()

	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, v1beta1.AddToScheme(scheme))

	builder := config.NewConfigmapBuilder(cluster, scheme)
	obj := builder.Build()
	require.NoError(t, builder.Update(obj))

	cm, ok := obj.(*corev1.ConfigMap)
	require.True(t, ok, "expected a ConfigMap")

	tmpl, ok := cm.Data["config_template.yaml"]
	require.True(t, ok, "config_template.yaml key must be present")

	return tmpl
}

// TestConfigTemplating_Pre130 asserts that clusters older than 1.30 keep using
// the dockerize "{{ .Env.X }}" placeholders and do not carry the sprig
// enable-template header.
func TestConfigTemplating_Pre130(t *testing.T) {
	tmpl := buildConfigTemplate(t, newSQLCluster(t, "1.29.7"))

	assert.False(t, strings.HasPrefix(tmpl, "# enable-template"),
		"pre-1.30 config must not enable server-side templating")
	assert.Contains(t, tmpl, "{{ .Env.", "pre-1.30 config must use dockerize placeholders")
	assert.NotContains(t, tmpl, `{{ env "`, "pre-1.30 config must not use sprig env placeholders")
}

// TestConfigTemplating_Post130 asserts that clusters >= 1.30 emit the sprig
// enable-template header and env placeholders, and that the rendered template
// is accepted by the real Temporal Server config loader (which embeds sprig and
// removed dockerize in 1.30).
func TestConfigTemplating_Post130(t *testing.T) {
	tmpl := buildConfigTemplate(t, newSQLCluster(t, "1.30.5"))

	assert.True(t, strings.HasPrefix(tmpl, "# enable-template"),
		"1.30+ config must start with the enable-template header")
	assert.Contains(t, tmpl, `{{ env "`, "1.30+ config must use sprig env placeholders")
	assert.NotContains(t, tmpl, "{{ .Env.", "1.30+ config must not use dockerize placeholders")

	// Feed the generated template to the actual server config loader to prove it
	// parses, renders (sprig) and unmarshals under Temporal Server >= 1.30.
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(tmpl), 0o600))

	cfg, err := temporalconfig.Load(temporalconfig.WithConfigFile(path))
	require.NoError(t, err, "generated 1.30 config must load via the server config loader")
	assert.NotNil(t, cfg)
}

// TestConfigPasswordCommand asserts that a 1.31 cluster whose datastore uses an
// external passwordCommand renders it into the server config (and omits the
// static password placeholder), and that the result loads via the real server
// config loader.
func TestConfigPasswordCommand(t *testing.T) {
	cluster := newSQLCluster(t, "1.31.1")
	cluster.Spec.Persistence.DefaultStore.SQL.PasswordCommand = &v1beta1.SQLPasswordCommandSpec{
		Command: "/bin/echo",
		Args:    []string{"token"},
	}

	tmpl := buildConfigTemplate(t, cluster)

	assert.Contains(t, tmpl, "passwordCommand:", "passwordCommand must be rendered")
	assert.Contains(t, tmpl, "/bin/echo", "passwordCommand command must be rendered")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(tmpl), 0o600))

	cfg, err := temporalconfig.Load(temporalconfig.WithConfigFile(path))
	require.NoError(t, err, "generated 1.31 passwordCommand config must load")
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.Persistence.DataStores["default"].SQL)
	assert.NotNil(t, cfg.Persistence.DataStores["default"].SQL.PasswordCommand,
		"loaded config must carry the passwordCommand")
}
