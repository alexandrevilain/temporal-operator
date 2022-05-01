/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"go.temporal.io/server/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceSpec contains a temporal service specifications.
type ServiceSpec struct {
	//
	// +optional
	Port *int `json:"port"`
	// +optional
	MembershipPort *int `json:"membershipPort"`
	//+kubebuilder:validation:Minimum=1
	Replicas *int `json:"replicas"`
}

// TemporalServicesSpec contains all temporal services specifications.
type TemporalServicesSpec struct {
	// +optional
	Frontend *ServiceSpec `json:"frontend"`
	// +optional
	History *ServiceSpec `json:"history"`
	// +optional
	Matching *ServiceSpec `json:"matching"`
	// +optional
	Worker *ServiceSpec `json:"worker"`
}

// GetServiceSpec returns service spec from its name.
func (s *TemporalServicesSpec) GetServiceSpec(name string) (*ServiceSpec, error) {
	switch name {
	case common.FrontendServiceName:
		return s.Frontend, nil
	case common.HistoryServiceName:
		return s.History, nil
	case common.MatchingServiceName:
		return s.Matching, nil
	case common.WorkerServiceName:
		return s.Worker, nil
	default:
		return nil, fmt.Errorf("unkown service %s", name)
	}
}

// SecretKeyReference contains enough information to locate the referenced Kubernetes Secret object in the same
// namespace.
type SecretKeyReference struct {
	// Name of the Secret.
	// +required
	Name string `json:"name"`

	// Key in the Secret.
	// +optional
	Key string `json:"key,omitempty"`
}

// SQLSpec contains SQL datastore connections specifications.
type SQLSpec struct {
	// User is the username to be used for the connection.
	User string `json:"user"`
	// PluginName is the name of SQL plugin.
	PluginName string `json:"pluginName"`
	// DatabaseName is the name of SQL database to connect to.
	DatabaseName string `json:"databaseName"`
	// ConnectAddr is the remote addr of the database.
	ConnectAddr string `json:"connectAddr"`
	// ConnectProtocol is the protocol that goes with the ConnectAddr.
	// +optional
	ConnectProtocol string `json:"connectProtocol"`
	// ConnectAttributes is a set of key-value attributes to be sent as part of connect data_source_name url
	// +optional
	ConnectAttributes map[string]string `json:"connectAttributes"`
	// MaxConns the max number of connections to this datastore.
	// +optional
	MaxConns int `json:"maxConns"`
	// MaxIdleConns is the max number of idle connections to this datastore.
	// +optional
	MaxIdleConns int `json:"maxIdleConns"`
	// MaxConnLifetime is the maximum time a connection can be alive
	// +optional
	MaxConnLifetime time.Duration `json:"maxConnLifetime"`
	// TaskScanPartitions is the number of partitions to sequentially scan during ListTaskQueue operations.
	// +optional
	TaskScanPartitions int `json:"taskScanPartitions"`
}

// DatastoreTLSSpec contains datastore TLS connections specifications.
type DatastoreTLSSpec struct {
	Enabled bool `json:"bool"`
	// +optional
	CertFileRef *SecretKeyReference `json:"certFileRef"`
	// +optional
	KeyFileRef *SecretKeyReference `json:"keyFileRef"`
	// +optional
	CaFileRef              *SecretKeyReference `json:"caFileRef"`
	EnableHostVerification bool                `json:"enableHostVerification"`
	ServerName             string              `json:"serverName"`
}

type DatastoreType string

const (
	CassandraDatastore   DatastoreType = "cassandra"
	PostgresSQLDatastore DatastoreType = "postgresql"
	MySQLDatastore       DatastoreType = "mysql"
)

// TemporalDatastoreSpec contains temporal datastore specifications.
type TemporalDatastoreSpec struct {
	// Name is the name of the datatstore.
	// It should be unique and will be referenced within the persitence spec.
	// +required
	Name string `json:"name"`
	// SQL holds all connection parameters for SQL datastores.
	// +optional
	SQL *SQLSpec `json:"sql"`
	// PasswordSecret is the reference to the secret holding the password.
	// +required
	PasswordSecretRef SecretKeyReference `json:"passwordSecretRef"`
	// TLS is an optionnal s
	// +optional
	TLS *DatastoreTLSSpec `json:"tls"`
}

// Default sets default values on the datastore.
func (s *TemporalDatastoreSpec) Default() {
	if s.SQL != nil {
		if s.SQL.ConnectProtocol == "" {
			s.SQL.ConnectProtocol = "tcp"
		}
	}
}

func (s *TemporalDatastoreSpec) GetDatastoreType() (DatastoreType, error) {
	if s.SQL != nil {
		switch s.SQL.PluginName {
		case "postgres":
			return PostgresSQLDatastore, nil
		case "mysql":
			return MySQLDatastore, nil
		}
	}
	return DatastoreType(""), errors.New("can't specify datastore type from current spec")
}

// GetTLSKeyFileMountPath returns the client TLS cert mount path.
// It returns empty if the tls config is nil or if no secret key ref has been specified.
func (s *TemporalDatastoreSpec) GetTLSCertFileMountPath() string {
	if s.TLS == nil || s.TLS.CertFileRef == nil {
		return ""
	}

	return path.Join("/etc/tls/datastores", s.Name, "client.pem")
}

// GetTLSKeyFileMountPath returns the client TLS key mount path.
// It returns empty if the tls config is nil or if no secret key ref has been specified.
func (s *TemporalDatastoreSpec) GetTLSKeyFileMountPath() string {
	if s.TLS == nil || s.TLS.KeyFileRef == nil {
		return ""
	}
	return path.Join("/etc/tls/datastores", s.Name, "client.key")
}

// GetTLSCaFileMountPath  returns the CA key mount path.
// It returns empty if the tls config is nil or if no secret key ref has been specified.
func (s *TemporalDatastoreSpec) GetTLSCaFileMountPath() string {
	if s.TLS == nil || s.TLS.CaFileRef == nil {
		return ""
	}
	return path.Join("/etc/tls/datastores", s.Name, "ca.pem")
}

// GetPasswordEnvVarName crafts the environment variable name for the datastore.
func (s *TemporalDatastoreSpec) GetPasswordEnvVarName() string {
	storeName := slug.Make(s.Name)
	storeName = strings.ToUpper(storeName)
	return fmt.Sprintf("TEMPORAL_%s_DATASTORE_PASSWORD", storeName)
}

// TemporalPersistenceSpec contains temporal persistence specifications.
type TemporalPersistenceSpec struct {
	// DefaultStore is the name of the default data store to use.
	DefaultStore string `json:"defaultStore"`
	// VisibilityStore is the name of the datastore to be used for visibility records.
	// If not set it defaults to the default store.
	// +optional
	VisibilityStore string `json:"visibilityStore"`
	// AdvancedVisibilityStore is the name of the datastore to be used for visibility records
	// +optional
	AdvancedVisibilityStore string `json:"advancedVisibilityStore"`
}

// TemporalClusterSpec defines the desired state of TemporalCluster.
type TemporalClusterSpec struct {
	// Image defines the temporal server image the instance should use.
	// +optional
	Image string `json:"image"`
	// Version defines the temporal version the instance should run.
	Version string `json:"version"`
	// NumHistoryShards is the desired number of history shards.
	// This field is immutable.
	//+kubebuilder:validation:Minimum=1
	NumHistoryShards int32                   `json:"numHistoryShards"`
	Services         TemporalServicesSpec    `json:"services"`
	Persistence      TemporalPersistenceSpec `json:"persistence"`
	Datastores       []TemporalDatastoreSpec `json:"datastores"`
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets"`
}

// TemporalClusterStatus defines the observed state of TemporalCluster.
type TemporalClusterStatus struct {
	// TODO(alexandrevilain): Use status
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TemporalCluster is the Schema for the temporalclusters API
type TemporalCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemporalClusterSpec   `json:"spec,omitempty"`
	Status TemporalClusterStatus `json:"status,omitempty"`
}

func (c *TemporalCluster) getDatastoreByName(name string) (*TemporalDatastoreSpec, bool) {
	for _, datastore := range c.Spec.Datastores {
		if datastore.Name == name {
			return &datastore, true
		}
	}
	return nil, false
}

func (c *TemporalCluster) GetDefaultDatastore() (*TemporalDatastoreSpec, bool) {
	return c.getDatastoreByName(c.Spec.Persistence.DefaultStore)
}

func (c *TemporalCluster) GetVisibilityDatastore() (*TemporalDatastoreSpec, bool) {
	return c.getDatastoreByName(c.Spec.Persistence.VisibilityStore)
}

// ChildResourceName returns child resource name using the cluster's name.
func (c *TemporalCluster) ChildResourceName(resource string) string {
	return fmt.Sprintf("%s-%s", c.Name, resource)
}

// Default sets default values on the temporal Cluster.
func (c *TemporalCluster) Default() {
	if c.Spec.Image == "" {
		c.Spec.Image = "temporalio/server"
	}
	// Frontend specs
	if c.Spec.Services.Frontend == nil {
		c.Spec.Services.Frontend = new(ServiceSpec)
	}
	if c.Spec.Services.Frontend.Port == nil {
		*c.Spec.Services.Frontend.Port = 7233
	}
	if c.Spec.Services.Frontend.MembershipPort == nil {
		*c.Spec.Services.Frontend.MembershipPort = 6933
	}
	// History specs
	if c.Spec.Services.History == nil {
		c.Spec.Services.History = new(ServiceSpec)
	}
	if c.Spec.Services.History.Port == nil {
		*c.Spec.Services.History.Port = 7234
	}
	if c.Spec.Services.History.MembershipPort == nil {
		*c.Spec.Services.History.MembershipPort = 6934
	}
	// Matching specs
	if c.Spec.Services.Matching == nil {
		c.Spec.Services.Matching = new(ServiceSpec)
	}
	if c.Spec.Services.Matching.Port == nil {
		*c.Spec.Services.Matching.Port = 7235
	}
	if c.Spec.Services.Matching.MembershipPort == nil {
		*c.Spec.Services.Matching.MembershipPort = 6935
	}
	// Worker specs
	if c.Spec.Services.Worker == nil {
		c.Spec.Services.Worker = new(ServiceSpec)
	}
	if c.Spec.Services.Worker.Port == nil {
		*c.Spec.Services.Worker.Port = 7239
	}
	if c.Spec.Services.Worker.MembershipPort == nil {
		*c.Spec.Services.Worker.MembershipPort = 6939
	}

	for _, datastore := range c.Spec.Datastores {
		datastore.Default()
	}

	if c.Spec.Persistence.VisibilityStore == "" {
		c.Spec.Persistence.VisibilityStore = c.Spec.Persistence.DefaultStore
	}
}

//+kubebuilder:object:root=true

// TemporalClusterList contains a list of TemporalCluster
type TemporalClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TemporalCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TemporalCluster{}, &TemporalClusterList{})
}
