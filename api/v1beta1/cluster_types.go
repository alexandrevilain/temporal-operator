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

package v1beta1

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/gosimple/slug"
	"go.temporal.io/server/common"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceSpec contains a temporal service specifications.
type ServiceSpec struct {
	// Port defines a custom gRPC port for the service.
	// Default values are:
	// 7233 for Frontend service
	// 7234 for History service
	// 7235 for Matching service
	// 7239 for Worker service
	// +optional
	Port *int `json:"port"`
	// Port defines a custom membership port for the service.
	// Default values are:
	// 6933 for Frontend service
	// 6934 for History service
	// 6935 for Matching service
	// 6939 for Worker service
	// +optional
	MembershipPort *int `json:"membershipPort"`
	// Number of desired replicas for the service. Default to 1.
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas *int32 `json:"replicas"`
}

// ServicesSpec contains all temporal services specifications.
type ServicesSpec struct {
	// Frontend service custom specifications.
	// +optional
	Frontend *ServiceSpec `json:"frontend"`
	// History service custom specifications.
	// +optional
	History *ServiceSpec `json:"history"`
	// Matching service custom specifications.
	// +optional
	Matching *ServiceSpec `json:"matching"`
	// Worker service custom specifications.
	// +optional
	Worker *ServiceSpec `json:"worker"`
}

// GetServiceSpec returns service spec from its name.
func (s *ServicesSpec) GetServiceSpec(name string) (*ServiceSpec, error) {
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
		return nil, fmt.Errorf("unknown service %s", name)
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
	// Enabled defines if the cluster should use a TLS connection to connect to the datastore.
	Enabled bool `json:"enabled"`
	// CertFileRef is a reference to a secret containing the cert file.
	// +optional
	CertFileRef *SecretKeyReference `json:"certFileRef"`
	// KeyFileRef is a reference to a secret containing the key file.
	// +optional
	KeyFileRef *SecretKeyReference `json:"keyFileRef"`
	// CaFileRef is a reference to a secret containing the ca file.
	// +optional
	CaFileRef *SecretKeyReference `json:"caFileRef"`
	// EnableHostVerification defines if the hostname should be verified when connecting to the datastore.
	EnableHostVerification bool `json:"enableHostVerification"`
	// ServerName the datastore should present.
	// +optional
	ServerName string `json:"serverName"`
}

// ElasticsearchIndices holds index names.
type ElasticsearchIndices struct {
	// Visibility defines visibility's index name.
	Visibility string `json:"visibility"`
	// SecondaryVisibility defines secondary visibility's index name.
	// +optional
	SecondaryVisibility string `json:"secondaryVisibility"`
}

// ElasticsearchSpec contains Elasticsearch datastore connections specifications.
type ElasticsearchSpec struct {
	// Version defines the elasticsearch version.
	// +kubebuilder:default=v7
	// +kubebuilder:validation:Pattern=`^v(6|7)$`
	Version string `json:"version"`
	// URL is the connection url to connect to the instance.
	// +kubebuilder:validation:Pattern=`^https?:\/\/.+$`
	URL string `json:"url"`
	// Username is the username to be used for the connection.
	Username string `json:"username"`
	// Indices holds visibility index names.
	Indices ElasticsearchIndices `json:"indices"`
	// LogLevel defines the temporal cluster's es client logger level.
	// +optional
	LogLevel string `json:"logLevel"`
	// CloseIdleConnectionsInterval is the max duration a connection stay open while idle.
	// +optional
	CloseIdleConnectionsInterval metav1.Duration `json:"closeIdleConnectionsInterval"`
	// EnableSniff enables or disables sniffer on the temporal cluster's es client.
	// +optional
	EnableSniff bool `json:"enableSniff"`
	// EnableHealthcheck enables or disables healthcheck on the temporal cluster's es client.
	// +optional
	EnableHealthcheck bool `json:"enableHealthcheck"`
}

// CassandraConsistencySpec sets the consistency level for regular & serial queries to Cassandra.
type CassandraConsistencySpec struct {
	// Consistency sets the default consistency level.
	// Values identical to gocql Consistency values. (defaults to LOCAL_QUORUM if not set).
	// +kubebuilder:validation:Enum=ANY;ONE;TWO;THREE;QUORUM;ALL;LOCAL_QUORUM;EACH_QUORUM;LOCAL_ONE
	// +optional
	Consistency *gocql.Consistency `json:"consistency"`
	// SerialConsistency sets the consistency for the serial prtion of queries. Values identical to gocql SerialConsistency values.
	// (defaults to LOCAL_SERIAL if not set)
	// +kubebuilder:validation:Enum=SERIAL;LOCAL_SERIAL
	// +optional
	SerialConsistency *gocql.SerialConsistency `json:"serialConsistency"`
}

// CassandraSpec contains cassandra datastore connections specifications.
type CassandraSpec struct {
	// Hosts is a list of cassandra endpoints.
	Hosts []string `json:"hosts"`
	// Port is the cassandra port used for connection by gocql client.
	Port int `json:"port"`
	// User is the cassandra user used for authentication by gocql client.
	User string `json:"user"`
	// Keyspace is the cassandra keyspace.
	Keyspace string `json:"keyspace"`
	// Datacenter is the data center filter arg for cassandra.
	Datacenter string `json:"datacenter"`
	// MaxConns is the max number of connections to this datastore for a single keyspace.
	// +optional
	MaxConns int `json:"maxConns"`
	// ConnectTimeout is a timeout for initial dial to cassandra server.
	// +optional
	ConnectTimeout *metav1.Duration `json:"connectTimeout"`
	// Consistency configuration.
	// +optional
	Consistency *CassandraConsistencySpec `json:"consistency"`
	// DisableInitialHostLookup instructs the gocql client to connect only using the supplied hosts.
	// +optional
	DisableInitialHostLookup bool `json:"disableInitialHostLookup"`
}

type DatastoreType string

const (
	CassandraDatastore     DatastoreType = "cassandra"
	PostgresSQLDatastore   DatastoreType = "postgresql"
	MySQLDatastore         DatastoreType = "mysql"
	ElasticsearchDatastore DatastoreType = "elasticsearch"
	UnknownDatastore       DatastoreType = "unknown"
)

const (
	DefaultStoreName            = "default"
	VisibilityStoreName         = "visibility"
	AdvancedVisibilityStoreName = "advancedVisibility"
)

// DatastoreSpec contains temporal datastore specifications.
type DatastoreSpec struct {
	// Name is the name of the datatstore.
	// It should be unique and will be referenced within the persitence spec.
	// Defaults to "default" for default sore, "visibility" for visibility store and
	// "advancedVisibility" for advanced visibility store.
	// +optional
	Name string `json:"name"`
	// SQL holds all connection parameters for SQL datastores.
	// +optional
	SQL *SQLSpec `json:"sql"`
	// Elasticsearch holds all connection parameters for Elasticsearch datastores.
	// +optional
	Elasticsearch *ElasticsearchSpec `json:"elasticsearch"`
	// Cassandra holds all connection parameters for Cassandra datastore.
	// +optional
	Cassandra *CassandraSpec `json:"cassandra"`
	// PasswordSecret is the reference to the secret holding the password.
	// +required
	PasswordSecretRef SecretKeyReference `json:"passwordSecretRef"`
	// TLS is an optional option to connect to the datastore using TLS.
	// +optional
	TLS *DatastoreTLSSpec `json:"tls"`
}

func (s *DatastoreSpec) GetType() DatastoreType {
	if s.SQL != nil {
		switch s.SQL.PluginName {
		case "postgres":
			return PostgresSQLDatastore
		case "mysql":
			return MySQLDatastore
		}
	}
	if s.Elasticsearch != nil {
		return ElasticsearchDatastore
	}
	if s.Cassandra != nil {
		return CassandraDatastore
	}
	return UnknownDatastore
}

const (
	dataStoreTLSCertificateBasePath = "/etc/tls/datastores"
	DataStoreClientTLSCertFileName  = "client.pem"
	DataStoreClientTLSKeyFileName   = "client.key"
	DataStoreClientTLSCaFileName    = "ca.pem"
)

// GetTLSKeyFileMountPath returns the client TLS cert mount path.
// It returns empty if the tls config is nil or if no secret key ref has been specified.
func (s *DatastoreSpec) GetTLSCertFileMountPath() string {
	if s.TLS == nil || s.TLS.CertFileRef == nil {
		return ""
	}

	return path.Join(dataStoreTLSCertificateBasePath, s.Name, DataStoreClientTLSCertFileName)
}

// GetTLSKeyFileMountPath returns the client TLS key mount path.
// It returns empty if the tls config is nil or if no secret key ref has been specified.
func (s *DatastoreSpec) GetTLSKeyFileMountPath() string {
	if s.TLS == nil || s.TLS.KeyFileRef == nil {
		return ""
	}
	return path.Join(dataStoreTLSCertificateBasePath, s.Name, DataStoreClientTLSKeyFileName)
}

// GetTLSCaFileMountPath  returns the CA key mount path.
// It returns empty if the tls config is nil or if no secret key ref has been specified.
func (s *DatastoreSpec) GetTLSCaFileMountPath() string {
	if s.TLS == nil || s.TLS.CaFileRef == nil {
		return ""
	}
	return path.Join(dataStoreTLSCertificateBasePath, s.Name, DataStoreClientTLSCaFileName)
}

// GetPasswordEnvVarName crafts the environment variable name for the datastore.
func (s *DatastoreSpec) GetPasswordEnvVarName() string {
	storeName := slug.Make(s.Name)
	storeName = strings.ToUpper(storeName)
	return fmt.Sprintf("TEMPORAL_%s_DATASTORE_PASSWORD", storeName)
}

// TemporalPersistenceSpec contains temporal persistence specifications.
type TemporalPersistenceSpec struct {
	// DefaultStore holds the default datastore specs.
	DefaultStore *DatastoreSpec `json:"defaultStore"`
	// VisibilityStore holds the visibility datastore specs.
	VisibilityStore *DatastoreSpec `json:"visibilityStore"`
	// AdvancedVisibilityStore holds the avanced visibility datastore specs.
	// +optional
	AdvancedVisibilityStore *DatastoreSpec `json:"advancedVisibilityStore"`
}

func (p *TemporalPersistenceSpec) GetDatastores() []*DatastoreSpec {
	stores := []*DatastoreSpec{
		p.DefaultStore,
		p.VisibilityStore,
	}

	if p.AdvancedVisibilityStore != nil {
		stores = append(stores, p.AdvancedVisibilityStore)
	}

	return stores
}

// TemporalUIIngressSpec contains all configurations options for the UI ingress.
type TemporalUIIngressSpec struct {
	// Annotations allows custom annotations on the ingress ressource.
	Annotations map[string]string `json:"annotations,omitempty"`
	// IngressClassName is the name of the IngressClass the deployed ingress resource should use.
	IngressClassName *string `json:"ingressClassName,omitempty"`
	// Host is the list of host the ingress should use.
	Hosts []string `json:"hosts"`
	// TLS configuration.
	TLS []networkingv1.IngressTLS `json:"tls,omitempty" protobuf:"bytes,2,rep,name=tls"`
}

// TemporalUISpec defines parameters for the temporal UI within a Temporal cluster deployment.
type TemporalUISpec struct {
	// Enabled defines if the operator should deploy the web ui alongside the cluster.
	// +optional
	Enabled bool `json:"enabled"`
	// Version defines the temporal ui version the instance should run.
	// +optional
	Version string `json:"version"`
	// Image defines the temporal ui docker image the instance should run.
	// +optional
	Image string `json:"image"`
	// Ingress is an optional ingress configuration for the UI.
	// If lived empty, no ingress configuration will be created and the UI will only by available trough ClusterIP service.
	// +optional
	Ingress *TemporalUIIngressSpec `json:"ingress,omitempty"`
}

// TemporalUISpec defines parameters for the temporal admin tools within a Temporal cluster deployment.
// Note that deployed admin tools version is the same as the cluster's version.
type TemporalAdminToolsSpec struct {
	// Enabled defines if the operator should deploy the admin tools alongside the cluster.
	// +optional
	Enabled bool `json:"enabled"`
	// Image defines the temporal admin tools docker image the instance should run.
	// +optional
	Image string `json:"image"`
}

// MTLSProvider is the enum for support mTLS provider.
type MTLSProvider string

const (
	// CertManagerMTLSProvider uses cert-manager to manage mTLS certificate.
	CertManagerMTLSProvider MTLSProvider = "cert-manager"
	LinkerdMTLSProvider     MTLSProvider = "linkerd"
	IstioMTLSProvider       MTLSProvider = "istio"
)

// InternodeMTLSSpec defines parameters for the temporal encryption in transit with mTLS.
type FrontendMTLSSpec struct {
	// Enabled defines if the operator should enable mTLS for cluster's public endpoints.
	// +optional
	Enabled bool `json:"enabled"`
}

// ServerName returns frontend servername for mTLS certificates.
func (FrontendMTLSSpec) ServerName(serverName string) string {
	return fmt.Sprintf("frontend.%s", serverName)
}

// GetIntermediateCACertificateMountPath returns the mount path for intermediate CA certificates.
func (FrontendMTLSSpec) GetIntermediateCACertificateMountPath() string {
	return "/etc/temporal/config/certs/client/ca"
}

// GetCertificateMountPath returns the mount path for the frontend certificate.
func (FrontendMTLSSpec) GetCertificateMountPath() string {
	return "/etc/temporal/config/certs/cluster/frontend"
}

// GetWorkerCertificateMountPath returns the mount path for the worker certificate.
func (FrontendMTLSSpec) GetWorkerCertificateMountPath() string {
	return "/etc/temporal/config/certs/cluster/worker"
}

// InternodeMTLSSpec defines parameters for the temporal encryption in transit with mTLS.
type InternodeMTLSSpec struct {
	// Enabled defines if the operator should enable mTLS for network between cluster nodes.
	// +optional
	Enabled bool `json:"enabled"`
}

// ServerName returns internode servername for mTLS certificates.
func (InternodeMTLSSpec) ServerName(serverName string) string {
	return fmt.Sprintf("internode.%s", serverName)
}

// GetIntermediateCACertificateMountPath returns the mount path for intermediate CA certificates.
func (InternodeMTLSSpec) GetIntermediateCACertificateMountPath() string {
	return "/etc/temporal/config/certs/cluster/ca"
}

// GetCertificateMountPath returns the mount path for the internode certificate.
func (InternodeMTLSSpec) GetCertificateMountPath() string {
	return "/etc/temporal/config/certs/cluster/internode"
}

// CertificatesDurationSpec defines parameters for the temporal mTLS certificates duration.
type CertificatesDurationSpec struct {
	// RootCACertificate is the 'duration' (i.e. lifetime) of the Root CA Certificate.
	// It defaults to 10 years.
	// +optional
	RootCACertificate *metav1.Duration `json:"rootCACertificate"`
	// IntermediateCACertificates is the 'duration' (i.e. lifetime) of the intermediate CAs Certificates.
	// It defaults to 5 years.
	// +optional
	IntermediateCAsCertificates *metav1.Duration `json:"intermediateCAsCertificates"`
	// ClientCertificates is the 'duration' (i.e. lifetime) of the client certificates.
	// It defaults to 1 year.
	// +optional
	ClientCertificates *metav1.Duration `json:"clientCertificates"`
	// FrontendCertificate is the 'duration' (i.e. lifetime) of the frontend certificate.
	// It defaults to 1 year.
	// +optional
	FrontendCertificate *metav1.Duration `json:"frontendCertificate"`
	// InternodeCertificate is the 'duration' (i.e. lifetime) of the internode certificate.
	// It defaults to 1 year.
	// +optional
	InternodeCertificate *metav1.Duration `json:"internodeCertificate"`
}

// MTLSSpec defines parameters for the temporal encryption in transit with mTLS.
type MTLSSpec struct {
	// Provider defines the tool used to manage mTLS certificates.
	// +kubebuilder:default=cert-manager
	// +kubebuilder:validation:Enum=cert-manager;linkerd;istio
	// +optional
	Provider MTLSProvider `json:"provider"`
	// Internode allows configuration of the internode traffic encryption.
	// Useless if mTLS provider is not cert-manager.
	// +optional
	Internode *InternodeMTLSSpec `json:"internode"`
	// Frontend allows configuration of the frontend's public endpoint traffic encryption.
	// Useless if mTLS provider is not cert-manager.
	// +optional
	Frontend *FrontendMTLSSpec `json:"frontend"`
	// CertificatesDuration allows configuration of maximum certificates lifetime.
	// Useless if mTLS provider is not cert-manager.
	// +optional
	CertificatesDuration *CertificatesDurationSpec `json:"certificatesDuration,omitempty"`
	// RefreshInterval defines interval between refreshes of certificates in the cluster components.
	// Defaults to 1 hour.
	// Useless if mTLS provider is not cert-manager.
	// +optional
	RefreshInterval *metav1.Duration `json:"refreshInterval"`
}

func (m *MTLSSpec) InternodeEnabled() bool {
	return m.Internode != nil && m.Internode.Enabled
}

func (m *MTLSSpec) FrontendEnabled() bool {
	return m.Frontend != nil && m.Frontend.Enabled
}

// ClusterSpec defines the desired state of Cluster.
type ClusterSpec struct {
	// Image defines the temporal server docker image the cluster should use for each services.
	// +optional
	Image string `json:"image"`
	// Version defines the temporal version the cluster to be deployed.
	// This version impacts the underlying persistence schemas versions.
	Version string `json:"version"`
	// JobTtlSecondsAfterFinished is amount of time to keep job pods after jobs are completed.
	// This field is immutable.
	//+kubebuilder:validation:Minimum=1
	JobTtlSecondsAfterFinished int32 `json:"jobTtlSecondsAfterFinished"`
	// NumHistoryShards is the desired number of history shards.
	// This field is immutable.
	//+kubebuilder:validation:Minimum=1
	NumHistoryShards int32 `json:"numHistoryShards"`
	// Services allows customizations for for each temporal services deployment.
	// +optional
	Services *ServicesSpec `json:"services,omitempty"`
	// Persistence defines temporal persistence configuration.
	Persistence TemporalPersistenceSpec `json:"persistence"`
	// An optional list of references to secrets in the same namespace
	// to use for pulling temporal images from registries.
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// UI allows configuration of the optional temporal web ui deployed alongside the cluster.
	// +optional
	UI *TemporalUISpec `json:"ui,omitempty"`
	// AdminTools allows configuration of the optional admin tool pod deployed alongside the cluster.
	// +optional
	AdminTools *TemporalAdminToolsSpec `json:"admintools,omitempty"`
	// MTLS allows configuration of the network traffic encryption for the cluster.
	// +optional
	MTLS *MTLSSpec `json:"mTLS,omitempty"`
}

// ServiceStatus reports a service status.
type ServiceStatus struct {
	// Name of the temporal service.
	Name string `json:"name"`
	// Current observed version of the service.
	Version string `json:"version"`
	// Ready defines if the service is ready.
	Ready bool `json:"ready"`
}

// ClusterStatus defines the observed state of Cluster.
type ClusterStatus struct {
	// Version holds the current temporal version.
	Version string `json:"version,omitempty"`
	// Services holds all services statuses.
	Services []ServiceStatus `json:"services,omitempty"`
	// Conditions represent the latest available observations of the Cluster state.
	Conditions []metav1.Condition `json:"conditions"`
}

// AddServiceStatus adds the provided service status to the cluster's status.
func (s *ClusterStatus) AddServiceStatus(status *ServiceStatus) {
	found := false
	for i, serviceStatus := range s.Services {
		if serviceStatus.Name == status.Name {
			s.Services[i].Version = status.Version
			s.Services[i].Ready = status.Ready
			found = true
		}
	}
	if !found {
		s.Services = append(s.Services, *status)
	}
}

// +genclient
// +genclient:Namespaced
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type == 'Ready')].status"
// +kubebuilder:printcolumn:name="ReconcileSuccess",type="string",JSONPath=".status.conditions[?(@.type == 'ReconcileSuccess')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Cluster defines a temporal cluster deployment.
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the Temporal cluster.
	Spec ClusterSpec `json:"spec,omitempty"`
	// Most recent observed status of the Temporal cluster.
	Status ClusterStatus `json:"status,omitempty"`
}

// ServerName returns cluster's server name.
func (c *Cluster) ServerName() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", c.Name, c.Namespace)
}

// MTLSEnabled returns true if mTLS is enabled for internode or frontend using cert-manager.
func (c *Cluster) MTLSWithCertManagerEnabled() bool {
	return c.Spec.MTLS != nil &&
		(c.Spec.MTLS.InternodeEnabled() || c.Spec.MTLS.FrontendEnabled()) &&
		c.Spec.MTLS.Provider == CertManagerMTLSProvider
}

// ChildResourceName returns child resource name using the cluster's name.
func (c *Cluster) ChildResourceName(resource string) string {
	return fmt.Sprintf("%s-%s", c.Name, resource)
}

func (c *Cluster) GetPublicClientAddress() string {
	return fmt.Sprintf("%s.%s:%d", c.ChildResourceName("frontend"), c.GetNamespace(), *c.Spec.Services.Frontend.Port)
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
