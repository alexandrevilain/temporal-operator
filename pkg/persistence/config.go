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
