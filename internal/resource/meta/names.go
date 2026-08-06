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

package meta

// Service components.
const (
	FrontendService      = "frontend"
	ServiceConfig        = "config"
	ServiceDynamicConfig = "dynamicconfig"
)

// Additionals services.
const (
	ServiceUIName     = "ui"
	ServiceAdminTools = "admintools"
)

// Server config file location.
//
// These three values have to agree: the rendered config is stored in the config
// ConfigMap under ConfigFileName, mounted into the server container at
// ConfigFilePath, and — for Temporal >= 1.30, which no longer has a fixed
// built-in location — found by the server through the
// TEMPORAL_SERVER_CONFIG_FILE_PATH environment variable. If they drift apart the
// server exits at startup with "could not read config file", so they are
// defined once here rather than repeated at each use.
const (
	ConfigFileName = "config_template.yaml"
	ConfigMountDir = "/etc/temporal/config"
	ConfigFilePath = ConfigMountDir + "/" + ConfigFileName
)
