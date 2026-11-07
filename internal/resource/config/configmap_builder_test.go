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
	"testing"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/resource/config"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/ptr"
)

func newCluster() *v1beta1.TemporalCluster {
	store := func(name, db string) *v1beta1.DatastoreSpec {
		return &v1beta1.DatastoreSpec{
			Name: name,
			SQL: &v1beta1.SQLSpec{
				PluginName:      "postgres12",
				User:            "temporal",
				DatabaseName:    db,
				ConnectAddr:     "db-rw:5432",
				ConnectProtocol: "tcp",
			},
			PasswordSecretRef: &v1beta1.SecretKeyReference{Name: "db-app", Key: "password"},
		}
	}
	c := &v1beta1.TemporalCluster{
		TypeMeta:   v1beta1.TemporalClusterTypeMeta,
		ObjectMeta: metav1.ObjectMeta{Name: "temporal", Namespace: "default"},
		Spec: v1beta1.TemporalClusterSpec{
			Version:          version.MustNewVersionFromString("1.28.4"),
			NumHistoryShards: 4,
			Persistence: v1beta1.TemporalPersistenceSpec{
				DefaultStore:    store("default", "temporal"),
				VisibilityStore: store("visibility", "temporal_visibility"),
			},
		},
	}
	c.Default()
	return c
}

func render(t *testing.T, c *v1beta1.TemporalCluster) string {
	t.Helper()
	scheme := runtime.NewScheme()
	utilruntime.Must(v1beta1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))

	b := config.NewConfigmapBuilder(c, scheme)
	cm := b.Build()
	require.NoError(t, b.Update(cm))
	return cm.(*corev1.ConfigMap).Data["config_template.yaml"]
}

func TestConfigmapBuilderClusterMetadata(t *testing.T) {
	type rendered struct {
		ClusterMetadata struct {
			EnableGlobalNamespace    bool   `yaml:"enableGlobalNamespace"`
			FailoverVersionIncrement int64  `yaml:"failoverVersionIncrement"`
			MasterClusterName        string `yaml:"masterClusterName"`
			CurrentClusterName       string `yaml:"currentClusterName"`
			ClusterInformation       map[string]struct {
				InitialFailoverVersion int64 `yaml:"initialFailoverVersion"`
			} `yaml:"clusterInformation"`
		} `yaml:"clusterMetadata"`
		DCRedirectionPolicy struct {
			Policy string `yaml:"policy"`
		} `yaml:"dcRedirectionPolicy"`
	}

	tests := map[string]struct {
		spec              *v1beta1.ClusterMetadataSpec
		wantGlobal        bool
		wantFailover      int64
		wantDCRedirection string
	}{
		"unset renders a standalone cluster": {
			spec:         nil,
			wantFailover: 1,
		},
		"empty block keeps the defaults": {
			spec:         &v1beta1.ClusterMetadataSpec{},
			wantFailover: 1,
		},
		"global namespaces with the default version": {
			spec:         &v1beta1.ClusterMetadataSpec{EnableGlobalNamespace: true},
			wantGlobal:   true,
			wantFailover: 1,
		},
		"all fields": {
			spec: &v1beta1.ClusterMetadataSpec{
				EnableGlobalNamespace:  true,
				InitialFailoverVersion: ptr.To[int64](2),
				DCRedirectionPolicy:    "selected-apis-forwarding",
			},
			wantGlobal:        true,
			wantFailover:      2,
			wantDCRedirection: "selected-apis-forwarding",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := newCluster()
			c.Spec.ClusterMetadata = tt.spec

			var got rendered
			require.NoError(t, yaml.Unmarshal([]byte(render(t, c)), &got))

			md := got.ClusterMetadata
			require.Equal(t, tt.wantGlobal, md.EnableGlobalNamespace)
			require.Equal(t, v1beta1.FailoverVersionIncrement, md.FailoverVersionIncrement)
			require.Equal(t, c.Name, md.MasterClusterName)
			require.Equal(t, c.Name, md.CurrentClusterName)
			require.Len(t, md.ClusterInformation, 1)
			require.Equal(t, tt.wantFailover, md.ClusterInformation[c.Name].InitialFailoverVersion)
			require.Equal(t, tt.wantDCRedirection, got.DCRedirectionPolicy.Policy)
		})
	}
}
