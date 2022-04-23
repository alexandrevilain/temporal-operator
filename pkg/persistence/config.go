package persistence

import (
	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"go.temporal.io/server/common/config"
)

// NewSQLconfigFromDatastoreSpec creates a new instance of a temporal SQL config from the provided TemporalDatastoreSpec
func NewSQLconfigFromDatastoreSpec(spec *v1alpha1.TemporalDatastoreSpec) *config.SQL {
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
		// TODO:
		// TLS:                &auth.TLS{},
	}
}
