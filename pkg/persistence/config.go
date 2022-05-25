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
	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"go.temporal.io/server/common/auth"
	"go.temporal.io/server/common/config"
)

// NewSQLconfigFromDatastoreSpec creates a new instance of a temporal SQL config from the provided TemporalDatastoreSpec.
func NewSQLconfigFromDatastoreSpec(spec *v1alpha1.TemporalDatastoreSpec) *config.SQL {
	cfg := &config.SQL{
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
	}
	if spec.TLS != nil {
		cfg.TLS = &auth.TLS{
			Enabled:                cfg.TLS.Enabled,
			CertFile:               spec.GetTLSCertFileMountPath(),
			KeyFile:                spec.GetTLSKeyFileMountPath(),
			CaFile:                 spec.GetTLSCaFileMountPath(),
			EnableHostVerification: cfg.TLS.EnableHostVerification,
			ServerName:             cfg.TLS.ServerName,
		}
	}
	return cfg
}
