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

package config

import (
	"go.temporal.io/server/common/cluster"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/dynamicconfig"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/telemetry"
)

// Config is the configuration for the Temporal server.
// It wraps server config but with support for older temporal versions.
// Starting from Temporal 1.25, in persistence, AdvancedVisibilityStore is no longer supported.
type Config struct {
	// Global is process-wide service-related configuration
	Global config.Global `yaml:"global"`
	// Persistence contains the configuration for temporal datastores
	Persistence Persistence `yaml:"persistence"`
	// Log is the logging config
	Log log.Config `yaml:"log"`
	// ClusterMetadata is the config containing all valid clusters and active cluster
	ClusterMetadata *cluster.Config `yaml:"clusterMetadata"`
	// DCRedirectionPolicy contains the frontend datacenter redirection policy
	DCRedirectionPolicy config.DCRedirectionPolicy `yaml:"dcRedirectionPolicy"`
	// Services is a map of service name to service config items
	Services map[string]config.Service `yaml:"services"`
	// Archival is the config for archival
	Archival config.Archival `yaml:"archival"`
	// PublicClient is config for connecting to temporal frontend
	PublicClient config.PublicClient `yaml:"publicClient"`
	// DynamicConfigClient is the config for setting up the file based dynamic config client
	// Filepath should be relative to the root directory
	DynamicConfigClient *dynamicconfig.FileBasedClientConfig `yaml:"dynamicConfigClient"`
	// NamespaceDefaults is the default config for every namespace
	NamespaceDefaults config.NamespaceDefaults `yaml:"namespaceDefaults"`
	// ExporterConfig allows the specification of process-wide OTEL exporters
	ExporterConfig telemetry.ExportConfig `yaml:"otel"`
}

type Persistence struct {
	config.Persistence `yaml:",inline"`
	// AdvancedVisibilityStore is the name of the datastore to be used for visibility records.
	AdvancedVisibilityStore string `yaml:"advancedVisibilityStore"`
}
