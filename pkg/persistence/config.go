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
	"fmt"
	"net/url"
	"strings"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"go.temporal.io/server/common/auth"
	"go.temporal.io/server/common/config"
	esclient "go.temporal.io/server/common/persistence/visibility/store/elasticsearch/client"
)

// NewSQLconfigFromDatastoreSpec creates a new instance of a temporal SQL config from the provided DatastoreSpec.
func NewSQLConfigFromDatastoreSpec(spec *v1beta1.DatastoreSpec) *config.SQL {
	return &config.SQL{
		User:               spec.SQL.User,
		Password:           "",
		PluginName:         spec.SQL.PluginName,
		DatabaseName:       spec.SQL.DatabaseName,
		ConnectAddr:        spec.SQL.ConnectAddr,
		ConnectProtocol:    spec.SQL.ConnectProtocol,
		ConnectAttributes:  spec.SQL.ConnectAttributes,
		MaxConns:           spec.SQL.MaxConns,
		MaxIdleConns:       spec.SQL.MaxIdleConns,
		MaxConnLifetime:    spec.SQL.MaxConnLifetime,
		TaskScanPartitions: spec.SQL.TaskScanPartitions,
		TLS:                tlsConfigConfigFromDatastoreSpec(spec),
	}
}

// NewElasticsearchConfigFromDatastoreSpec creates a new instance of a temporal elasticsearch client config from the provided DatastoreSpec.
func NewElasticsearchConfigFromDatastoreSpec(spec *v1beta1.DatastoreSpec) (*esclient.Config, error) {
	parsedURL, err := url.Parse(spec.Elasticsearch.URL)
	if err != nil {
		return nil, fmt.Errorf("can't parse elasticsearch url: %w", err)
	}
	return &esclient.Config{
		Version:                      spec.Elasticsearch.Version,
		URL:                          *parsedURL,
		Username:                     spec.Elasticsearch.Username,
		Password:                     "",
		Indices:                      elasticsearchIndicesToMap(spec.Elasticsearch.Indices),
		LogLevel:                     spec.Elasticsearch.LogLevel,
		CloseIdleConnectionsInterval: spec.Elasticsearch.CloseIdleConnectionsInterval.Duration,
		EnableSniff:                  spec.Elasticsearch.EnableSniff,
		EnableHealthcheck:            spec.Elasticsearch.EnableSniff,
	}, nil
}

func elasticsearchIndicesToMap(indices v1beta1.ElasticsearchIndices) map[string]string {
	result := map[string]string{}
	if indices.Visibility != "" {
		result[esclient.VisibilityAppName] = indices.Visibility
	}
	if indices.SecondaryVisibility != "" {
		result[esclient.SecondaryVisibilityAppName] = indices.SecondaryVisibility
	}
	return result
}

// NewCassandraConfigFromDatastoreSpec creates a new instance of a temporal cassandra config from the provided DatastoreSpec.
func NewCassandraConfigFromDatastoreSpec(spec *v1beta1.DatastoreSpec) *config.Cassandra {
	cfg := &config.Cassandra{
		Hosts:                    strings.Join(spec.Cassandra.Hosts, ","),
		Port:                     spec.Cassandra.Port,
		User:                     spec.Cassandra.User,
		Password:                 "",
		Keyspace:                 spec.Cassandra.Keyspace,
		Datacenter:               spec.Cassandra.Datacenter,
		MaxConns:                 spec.Cassandra.MaxConns,
		TLS:                      tlsConfigConfigFromDatastoreSpec(spec),
		Consistency:              &config.CassandraStoreConsistency{},
		DisableInitialHostLookup: spec.Cassandra.DisableInitialHostLookup,
	}

	if spec.Cassandra.ConnectTimeout != nil {
		cfg.ConnectTimeout = spec.Cassandra.ConnectTimeout.Duration
	}

	if spec.Cassandra.Consistency != nil {
		cfg.Consistency = &config.CassandraStoreConsistency{
			Default: &config.CassandraConsistencySettings{},
		}

		if spec.Cassandra.Consistency.Consistency != nil {
			cfg.Consistency.Default.Consistency = spec.Cassandra.Consistency.Consistency.String()
		}

		if spec.Cassandra.Consistency.SerialConsistency != nil {
			cfg.Consistency.Default.Consistency = spec.Cassandra.Consistency.SerialConsistency.String()
		}
	}
	return cfg
}

func tlsConfigConfigFromDatastoreSpec(spec *v1beta1.DatastoreSpec) *auth.TLS {
	if spec.TLS == nil {
		return nil
	}
	return &auth.TLS{
		Enabled:                spec.TLS.Enabled,
		CertFile:               spec.GetTLSCertFileMountPath(),
		KeyFile:                spec.GetTLSKeyFileMountPath(),
		CaFile:                 spec.GetTLSCaFileMountPath(),
		EnableHostVerification: spec.TLS.EnableHostVerification,
		ServerName:             spec.TLS.ServerName,
	}
}
