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

	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/gocql/gocql"
	"github.com/gosimple/slug"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"go.temporal.io/server/common/primitives"
	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LogSpec contains the temporal logging configuration.
type LogSpec struct {
	// Stdout is true if the output needs to goto standard out; default is stderr.
	// +optional
	// +kubebuilder:default=true
	Stdout *bool `json:"stdout"`
	// Level is the desired log level; see colocated zap_logger.go::parseZapLevel()
	// +optional
	// +kubebuilder:validation:Enum=debug;info;warn;error;dpanic;panic;fatal
	// +kubebuilder:default=info
	Level string `json:"level"`
	// OutputFile is the path to the log output file.
	// +optional
	OutputFile string `json:"outputFile"`
	// Format determines the format of each log file printed to the output.
	// Use "console" if you want stack traces to appear on multiple lines.
	// +kubebuilder:validation:Enum=json;console
	// +kubebuilder:default=json
	// +optional
	Format string `json:"format"`
	// Development determines whether the logger is run in Development (== Test) or in
	// Production mode.  Default is Production.  Production-stage disables panics from
	// DPanic logging.
	// +kubebuilder:default=false
	// +optional
	Development bool `json:"development"`
}

// ServiceSpec contains a temporal service specifications.
type ServiceSpec struct {
	// Port defines a custom gRPC port for the service.
	// Default values are:
	// 7233 for Frontend service
	// 7234 for History service
	// 7235 for Matching service
	// 7239 for Worker service
	// +optional
	Port *int32 `json:"port"`
	// MembershipPort defines a custom membership port for the service.
	// Default values are:
	// 6933 for Frontend service
	// 6934 for History service
	// 6935 for Matching service
	// 6939 for Worker service
	// +optional
	MembershipPort *int32 `json:"membershipPort"`
	// HTTPPort defines a custom http port for the service.
	// Default values are:
	// 7243 for Frontend service
	// +optional
	HTTPPort *int32 `json:"httpPort"`
	// Number of desired replicas for the service. Default to 1.
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas *int32 `json:"replicas"`
	// Compute Resources required by this service.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Overrides adds some overrides to the resources deployed for the service.
	// Those overrides takes precedence over spec.services.overrides.
	// +optional
	Overrides *ServiceSpecOverride `json:"overrides,omitempty"`
	// InitContainers adds a list of init containers to the service's deployment.
	// +optional
	InitContainers []corev1.Container `json:"initContainers,omitempty"`
	// ServiceAccountOverride
}

// InternalFrontendServiceSpec contains temporal internal frontend service specifications.
type InternalFrontendServiceSpec struct {
	ServiceSpec `json:",inline"`
	// Enabled defines if we want to spawn the internal frontend service.
	// +optional
	// +kubebuilder:default:=false
	Enabled bool `json:"enabled,omitempty"`
}

func (s *InternalFrontendServiceSpec) IsEnabled() bool {
	return s != nil && s.Enabled
}

// ServicesSpec contains all temporal services specifications.
type ServicesSpec struct {
	// Frontend service custom specifications.
	// +optional
	Frontend *ServiceSpec `json:"frontend,omitempty"`
	// Internal Frontend service custom specifications.
	// Only compatible with temporal >= 1.20.0
	// +optional
	InternalFrontend *InternalFrontendServiceSpec `json:"internalFrontend,omitempty"`
	// History service custom specifications.
	// +optional
	History *ServiceSpec `json:"history,omitempty"`
	// Matching service custom specifications.
	// +optional
	Matching *ServiceSpec `json:"matching,omitempty"`
	// Worker service custom specifications.
	// +optional
	Worker *ServiceSpec `json:"worker,omitempty"`
	// Overrides adds some overrides to the resources deployed for all temporal services services.
	// Those overrides can be customized per service using spec.services.<serviceName>.overrides.
	// +optional
	Overrides *ServiceSpecOverride `json:"overrides,omitempty"`
}

// GetServiceSpec returns service spec from its name.
func (s *ServicesSpec) GetServiceSpec(name primitives.ServiceName) (*ServiceSpec, error) {
	switch name {
	case primitives.FrontendService:
		return s.Frontend, nil
	case primitives.InternalFrontendService:
		if s.InternalFrontend == nil {
			return &ServiceSpec{}, nil
		}
		return &s.InternalFrontend.ServiceSpec, nil
	case primitives.HistoryService:
		return s.History, nil
	case primitives.MatchingService:
		return s.Matching, nil
	case primitives.WorkerService:
		return s.Worker, nil
	case primitives.AllServices, primitives.ServerService, primitives.UnitTestService:
		fallthrough
	default:
		return nil, fmt.Errorf("unknown service %s", name)
	}
}

// ServiceSpecOverride provides the ability to override the generated manifests of a temporal service.
type ServiceSpecOverride struct {
	// Override configuration for the temporal service Deployment.
	Deployment *DeploymentOverride `json:"deployment,omitempty"`
}

// DeploymentOverride provides the ability to override a Deployment.
type DeploymentOverride struct {
	*ObjectMetaOverride `json:"metadata,omitempty"`
	// Specification of the desired behavior of the Deployment.
	// +optional
	Spec      *DeploymentOverrideSpec `json:"spec,omitempty"`
	JSONPatch *apiextensionsv1.JSON   `json:"jsonPatch,omitempty"`
}

// DeploymentOverrideSpec provides the ability to override a Deployment Spec.
// It's a subset of fields included in k8s.io/api/apps/v1.DeploymentSpec.
type DeploymentOverrideSpec struct {
	// Template describes the pods that will be created.
	// +optional
	Template *PodTemplateSpecOverride `json:"template,omitempty"`
}

// PodTemplateSpecOverride provides the ability to override a pod template spec.
// It's a subset of the fields included in k8s.io/api/core/v1.PodTemplateSpec.
type PodTemplateSpecOverride struct {
	*ObjectMetaOverride `json:"metadata,omitempty"`

	// Specification of the desired behavior of the pod.
	// +optional
	Spec *apiextensionsv1.JSON `json:"spec,omitempty"`
}

// ObjectMetaOverride provides the ability to override an object metadata.
// It's a subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta.
type ObjectMetaOverride struct {
	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
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
	// +kubebuilder:validation:Enum=postgres;postgres12;mysql;mysql8
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
	ConnectAttributes map[string]string `json:"connectAttributes,omitempty"`
	// MaxConns the max number of connections to this datastore.
	// +optional
	MaxConns int `json:"maxConns"`
	// MaxIdleConns is the max number of idle connections to this datastore.
	// +optional
	MaxIdleConns int `json:"maxIdleConns"`
	// MaxConnLifetime is the maximum time a connection can be alive
	// +optional
	MaxConnLifetime metav1.Duration `json:"maxConnLifetime"`
	// TaskScanPartitions is the number of partitions to sequentially scan during ListTaskQueue operations.
	// +optional
	TaskScanPartitions int `json:"taskScanPartitions"`
	// GCPServiceAccount is the service account to use to authenticate with GCP CloudSQL.
	// +optional
	GCPServiceAccount *string `json:"gcpServiceAccount,omitempty"`
}

// DatastoreTLSSpec contains datastore TLS connections specifications.
type DatastoreTLSSpec struct {
	// Enabled defines if the cluster should use a TLS connection to connect to the datastore.
	Enabled bool `json:"enabled"`
	// CertFileRef is a reference to a secret containing the cert file.
	// +optional
	CertFileRef *SecretKeyReference `json:"certFileRef,omitempty"`
	// KeyFileRef is a reference to a secret containing the key file.
	// +optional
	KeyFileRef *SecretKeyReference `json:"keyFileRef,omitempty"`
	// CaFileRef is a reference to a secret containing the ca file.
	// +optional
	CaFileRef *SecretKeyReference `json:"caFileRef,omitempty"`
	// EnableHostVerification defines if the hostname should be verified when connecting to the datastore.
	EnableHostVerification bool `json:"enableHostVerification"`
	// ServerName the datastore should present.
	// +optional
	ServerName string `json:"serverName"`
}

// ElasticsearchIndices holds index names.
type ElasticsearchIndices struct {
	// Visibility defines visibility's index name.
	// +kubebuilder:default=temporal_visibility_v1
	Visibility string `json:"visibility"`
	// SecondaryVisibility defines secondary visibility's index name.
	// +optional
	SecondaryVisibility string `json:"secondaryVisibility"`
}

// ElasticsearchSpec contains Elasticsearch datastore connections specifications.
type ElasticsearchSpec struct {
	// Version defines the elasticsearch version.
	// +kubebuilder:default=v7
	// +kubebuilder:validation:Pattern=`^v(6|7|8)$`
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
	// +optional
	Datacenter string `json:"datacenter"`
	// MaxConns is the max number of connections to this datastore for a single keyspace.
	// +optional
	MaxConns int `json:"maxConns"`
	// ConnectTimeout is a timeout for initial dial to cassandra server.
	// +optional
	ConnectTimeout *metav1.Duration `json:"connectTimeout"`
	// Consistency configuration.
	// +optional
	Consistency *CassandraConsistencySpec `json:"consistency,omitempty"`
	// DisableInitialHostLookup instructs the gocql client to connect only using the supplied hosts.
	// +optional
	DisableInitialHostLookup bool `json:"disableInitialHostLookup"`
}

type DatastoreType string

const (
	CassandraDatastore     DatastoreType = "cassandra"
	PostgresSQLDatastore   DatastoreType = "postgres"
	PostgresSQL12Datastore DatastoreType = "postgres12"
	MySQLDatastore         DatastoreType = "mysql"
	MySQL8Datastore        DatastoreType = "mysql8"
	ElasticsearchDatastore DatastoreType = "elasticsearch"
	UnknownDatastore       DatastoreType = "unknown"
)

var SQLDataStores = []DatastoreType{MySQLDatastore, MySQL8Datastore, PostgresSQLDatastore, PostgresSQL12Datastore}

const (
	DefaultStoreName             = "default"
	VisibilityStoreName          = "visibility"
	SecondaryVisibilityStoreName = "secondaryVisibility"
	AdvancedVisibilityStoreName  = "advancedVisibility"
)

// DatastoreSpec contains temporal datastore specifications.
type DatastoreSpec struct {
	// Name is the name of the datastore.
	// It should be unique and will be referenced within the persistence spec.
	// Defaults to "default" for default sore, "visibility" for visibility store,
	// "secondaryVisibility" for secondary visibility store and
	// "advancedVisibility" for advanced visibility store.
	// +optional
	Name string `json:"name"`
	// SQL holds all connection parameters for SQL datastores.
	// +optional
	SQL *SQLSpec `json:"sql,omitempty"`
	// Elasticsearch holds all connection parameters for Elasticsearch datastores.
	// +optional
	Elasticsearch *ElasticsearchSpec `json:"elasticsearch,omitempty"`
	// Cassandra holds all connection parameters for Cassandra datastore.
	// Note that cassandra is now deprecated for visibility store.
	// +optional
	Cassandra *CassandraSpec `json:"cassandra,omitempty"`
	// PasswordSecret is the reference to the secret holding the password.
	// +optional
	PasswordSecretRef *SecretKeyReference `json:"passwordSecretRef,omitempty"`
	// TLS is an optional option to connect to the datastore using TLS.
	// +optional
	TLS *DatastoreTLSSpec `json:"tls,omitempty"`
	// SkipCreate instructs the operator to skip creating the database for SQL datastores or to skip creating keyspace for Cassandra. Use this option if your database or keyspace has already been provisioned by an administrator.
	// +optional
	SkipCreate bool `json:"skipCreate"`
}

// LowerCaseName returns the datastore name in lower case.
func (s *DatastoreSpec) LowerCaseName() string {
	return strings.ToLower(s.Name)
}

// GetType returns datastore type.
func (s *DatastoreSpec) GetType() DatastoreType {
	if s.SQL != nil {
		switch s.SQL.PluginName {
		case "postgres":
			return PostgresSQLDatastore
		case "postgres12":
			return PostgresSQL12Datastore
		case "mysql":
			return MySQLDatastore
		case "mysql8":
			return MySQL8Datastore
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

// IsSQL returns true if the datastore is an SQL datastore.
func (s *DatastoreSpec) IsSQL() bool {
	return slices.Contains(SQLDataStores, s.GetType())
}

const (
	dataStoreTLSCertificateBasePath = "/etc/tls/datastores"
	dataStoreTLSCAPrefix            = "ca"
	dataStoreTLSCertPrefix          = "cert"
	dataStoreTLSKeyPrefix           = "key"
	// DataStoreClientTLSCertFileName is the default client TLS cert file name.
	DataStoreClientTLSCertFileName = "client.pem"
	// DataStoreClientTLSKeyFileName is the default client TLS key file name.
	DataStoreClientTLSKeyFileName = "client.key"
	// DataStoreClientTLSCaFileName is the default client TLS ca file name.
	DataStoreClientTLSCaFileName = "ca.pem"
)

// GetTLSKeyFileMountPath returns the client TLS cert mount path.
// It returns empty if the tls config is nil or if no secret key ref has been specified.
func (s *DatastoreSpec) GetTLSCertFileMountPath() string {
	if s.TLS == nil || s.TLS.CertFileRef == nil {
		return ""
	}

	return path.Join(dataStoreTLSCertificateBasePath, dataStoreTLSCertPrefix, s.Name, DataStoreClientTLSCertFileName)
}

// GetTLSKeyFileMountPath returns the client TLS key mount path.
// It returns empty if the tls config is nil or if no secret key ref has been specified.
func (s *DatastoreSpec) GetTLSKeyFileMountPath() string {
	if s.TLS == nil || s.TLS.KeyFileRef == nil {
		return ""
	}
	return path.Join(dataStoreTLSCertificateBasePath, dataStoreTLSKeyPrefix, s.Name, DataStoreClientTLSKeyFileName)
}

// GetTLSCaFileMountPath  returns the CA key mount path.
// It returns empty if the tls config is nil or if no secret key ref has been specified.
func (s *DatastoreSpec) GetTLSCaFileMountPath() string {
	if s.TLS == nil || s.TLS.CaFileRef == nil {
		return ""
	}
	return path.Join(dataStoreTLSCertificateBasePath, dataStoreTLSCAPrefix, s.Name, DataStoreClientTLSCaFileName)
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
	// SecondaryVisibilityStore holds the secondary visibility datastore specs.
	// Feature only available for clusters >= 1.21.0.
	// +optional
	SecondaryVisibilityStore *DatastoreSpec `json:"secondaryVisibilityStore,omitempty"`
	// AdvancedVisibilityStore holds the advanced visibility datastore specs.
	// +optional
	AdvancedVisibilityStore *DatastoreSpec `json:"advancedVisibilityStore,omitempty"`
}

func (p *TemporalPersistenceSpec) GetDatastores() []*DatastoreSpec {
	stores := []*DatastoreSpec{
		p.DefaultStore,
		p.VisibilityStore,
	}

	if p.SecondaryVisibilityStore != nil {
		stores = append(stores, p.SecondaryVisibilityStore)
	}

	if p.AdvancedVisibilityStore != nil {
		stores = append(stores, p.AdvancedVisibilityStore)
	}

	return stores
}

func (p *TemporalPersistenceSpec) GetDatastoresMap() map[string]*DatastoreSpec {
	stores := map[string]*DatastoreSpec{
		"defaultStore":    p.DefaultStore,
		"visibilityStore": p.VisibilityStore,
	}

	if p.SecondaryVisibilityStore != nil {
		stores["secondaryVisibilityStore"] = p.SecondaryVisibilityStore
	}

	if p.AdvancedVisibilityStore != nil {
		stores["advancedVisibilityStore"] = p.AdvancedVisibilityStore
	}

	return stores
}

// TemporalUIIngressSpec contains all configurations options for the UI ingress.
type TemporalUIIngressSpec struct {
	// Annotations allows custom annotations on the ingress resource.
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
	// Number of desired replicas for the ui. Default to 1.
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas *int32 `json:"replicas"`
	// Compute Resources required by the ui.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Overrides adds some overrides to the resources deployed for the ui.
	// +optional
	Overrides *ServiceSpecOverride `json:"overrides,omitempty"`
	// Ingress is an optional ingress configuration for the UI.
	// If lived empty, no ingress configuration will be created and the UI will only by available trough ClusterIP service.
	// +optional
	Ingress *TemporalUIIngressSpec `json:"ingress,omitempty"`
	// Service is an optional service resource configuration for the UI.
	// +optional
	Service *ObjectMetaOverride `json:"service,omitempty"`
}

// TemporalAdminToolsSpec defines parameters for the temporal admin tools within a Temporal cluster deployment.
// Note that deployed admin tools version is the same as the cluster's version.
type TemporalAdminToolsSpec struct {
	// Enabled defines if the operator should deploy the admin tools alongside the cluster.
	// +optional
	Enabled bool `json:"enabled"`
	// Image defines the temporal admin tools docker image the instance should run.
	// +optional
	Image string `json:"image"`
	// Version defines the temporal admin tools version the instance should run.
	// +optional
	Version string `json:"version"`
	// Compute Resources required by the ui.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Overrides adds some overrides to the resources deployed for the ui.
	// +optional
	Overrides *ServiceSpecOverride `json:"overrides,omitempty"`
}

// MTLSProvider is the enum for support mTLS provider.
type MTLSProvider string

const (
	// CertManagerMTLSProvider uses cert-manager to manage mTLS certificate.
	CertManagerMTLSProvider MTLSProvider = "cert-manager"
	LinkerdMTLSProvider     MTLSProvider = "linkerd"
	IstioMTLSProvider       MTLSProvider = "istio"
)

// FrontendMTLSSpec defines parameters for the temporal encryption in transit with mTLS.
type FrontendMTLSSpec struct {
	// Enabled defines if the operator should enable mTLS for cluster's public endpoints.
	// +optional
	Enabled bool `json:"enabled"`
	// ExtraDNSNames is a list of additional DNS names associated with the TemporalCluster.
	// These DNS names can be used for accessing the TemporalCluster from external services.
	// The DNS names specified here will be added to the TLS certificate for secure communication.
	// +nullable
	ExtraDNSNames []string `json:"extraDnsNames,omitempty"`
}

// ServerName returns frontend servername for mTLS certificates.
func (FrontendMTLSSpec) ServerName(cluster *TemporalCluster) string {
	return fmt.Sprintf("%s.%s", cluster.ChildResourceName("frontend"), cluster.FQDNSuffix())
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
func (InternodeMTLSSpec) ServerName(cluster *TemporalCluster) string {
	return fmt.Sprintf("%s.%s", cluster.ChildResourceName("internode"), cluster.FQDNSuffix())
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
	RootCACertificate *metav1.Duration `json:"rootCACertificate"` //nolint:tagliatelle
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
	Internode *InternodeMTLSSpec `json:"internode,omitempty"`
	// Frontend allows configuration of the frontend's public endpoint traffic encryption.
	// Useless if mTLS provider is not cert-manager.
	// +optional
	Frontend *FrontendMTLSSpec `json:"frontend,omitempty"`
	// CertificatesDuration allows configuration of maximum certificates lifetime.
	// Useless if mTLS provider is not cert-manager.
	// +optional
	CertificatesDuration *CertificatesDurationSpec `json:"certificatesDuration,omitempty"`
	// RefreshInterval defines interval between refreshes of certificates in the cluster components.
	// Defaults to 1 hour.
	// Useless if mTLS provider is not cert-manager.
	// +optional
	RefreshInterval *metav1.Duration `json:"refreshInterval"`
	// RenewBefore is defines how long before the currently issued certificate's expiry
	// cert-manager should renew the certificate. The default is 2/3 of the
	// issued certificate's duration. Minimum accepted value is 5 minutes.
	// Useless if mTLS provider is not cert-manager.
	// +optional
	RenewBefore *metav1.Duration `json:"renewBefore,omitempty"`
	// PermissiveMetrics allows insecure HTTP requests to the metrics endpoint.
	// This is handy if the metrics collector does not support mTLS.
	// Useless if mTLS provider is not istio
	// +optional
	PermissiveMetrics bool `json:"permissiveMetrics"`
}

func (m *MTLSSpec) InternodeEnabled() bool {
	return m.Internode != nil && m.Internode.Enabled
}

func (m *MTLSSpec) FrontendEnabled() bool {
	return m.Frontend != nil && m.Frontend.Enabled
}

// PrometheusScrapeConfigServiceMonitor is the configuration for prometheus operator ServiceMonitor.
type PrometheusScrapeConfigServiceMonitor struct {
	// Enabled defines if the operator should create a ServiceMonitor for each services.
	// +optional
	Enabled bool `json:"enabled"`
	// Labels adds extra labels to the ServiceMonitor.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Override allows customization of the created ServiceMonitor.
	// All fields can be overwritten except "endpoints", "selector" and "namespaceSelector".
	// +optional
	Override *monitoringv1.ServiceMonitorSpec `json:"override,omitempty"`
	// MetricRelabelConfigs to apply to samples before ingestion.
	// +optional
	MetricRelabelConfigs []monitoringv1.RelabelConfig `json:"metricRelabelings,omitempty"`
}

// PrometheusScrapeConfig is the configuration for making prometheus scrape components metrics.
type PrometheusScrapeConfig struct {
	// Annotations defines if the operator should add prometheus scrape annotations to the services pods.
	// +optional
	Annotations bool `json:"annotations"`
	// +optional
	ServiceMonitor *PrometheusScrapeConfigServiceMonitor `json:"serviceMonitor,omitempty"`
}

// PrometheusSpec is the configuration for prometheus reporter.
type PrometheusSpec struct {
	// Deprecated. Address for prometheus to serve metrics from.
	// +optional
	// +deprecated
	ListenAddress string `json:"listenAddress"`
	// ListenPort for prometheus to serve metrics from.
	// +optional
	ListenPort *int32 `json:"listenPort,omitempty"`
	// ScrapeConfig is the prometheus scrape configuration.
	// +optional
	ScrapeConfig *PrometheusScrapeConfig `json:"scrapeConfig,omitempty"`
}

// MetricsSpec determines parameters for configuring metrics endpoints.
type MetricsSpec struct {
	// Enabled defines if the operator should enable metrics exposition on temporal components.
	Enabled bool `json:"enabled"`
	// ExcludeTags is a map from tag name string to tag values string list.
	// Each value present in keys will have relevant tag value replaced with "_tag_excluded_"
	// Each value in values list will white-list tag values to be reported as usual.
	// +optional
	ExcludeTags map[string][]string `json:"excludeTags,omitempty"`
	// PerUnitHistogramBoundaries defines the default histogram bucket boundaries.
	// Configuration of histogram boundaries for given metric unit.
	//
	// Supported values:
	// - "dimensionless"
	// - "milliseconds"
	// - "bytes"
	// +optional
	PerUnitHistogramBoundaries map[string][]string `json:"perUnitHistogramBoundaries,omitempty"`
	// Prefix sets the prefix to all outgoing metrics
	// +optional
	Prefix *string `json:"prefix,omitempty"`
	// Prometheus reporter configuration.
	// +optional
	Prometheus *PrometheusSpec `json:"prometheus,omitempty"`
}

func (m *MetricsSpec) IsEnabled() bool {
	return m != nil && m.Enabled
}

// Constraints is an alias for temporal's dynamicconfig.Constraints.
// It describes under what conditions a ConstrainedValue should be used.
type Constraints struct {
	// +optional
	Namespace string `json:"namespace"`
	// +optional
	NamespaceID string `json:"namespaceId"`
	// +optional
	TaskQueueName string `json:"taskQueueName"`
	// +optional
	TaskQueueType string `json:"taskQueueType"`
	// +optional
	ShardID int32 `json:"shardId"`
	// +optional
	TaskType string `json:"taskType"`
}

// ConstrainedValue is an alias for temporal's dynamicconfig.ConstrainedValue.
type ConstrainedValue struct {
	// Constraints describe under what conditions a ConstrainedValue should be used.
	// +optional
	Constraints Constraints `json:"constraints"`
	// Value is the value for the configuration key.
	// The type of the Value field depends on the key.
	// Acceptable types will be one of: int, float64, bool, string, map[string]any, time.Duration
	Value *apiextensionsv1.JSON `json:"value"`
	// Value json.RawMessage `json:"value"`
}

// DynamicConfigSpec is the configuration for temporal dynamic config.
type DynamicConfigSpec struct {
	// PollInterval defines how often the config should be updated by checking provided values.
	// Defaults to 10s.
	// +optional
	PollInterval *metav1.Duration `json:"pollInterval"`
	// Values contains all dynamic config keys and their constrained values.
	Values map[string][]ConstrainedValue `json:"values"`
}

// ClusterArchivalSpec is the configuration for cluster-wide archival config.
type ClusterArchivalSpec struct {
	// Enabled defines if the archival is enabled for the cluster.
	// +kubebuilder:default:=false
	// +optional
	Enabled bool `json:"enabled"`
	// Provider defines the archival provider for the cluster.
	// The same provider is used for both history and visibility,
	// but some config can be changed using spec.archival.[history|visibility].config.
	// +optional
	Provider *ArchivalProvider `json:"provider,omitempty"`
	// History is the default config for the history archival.
	// +optional
	History *ArchivalSpec `json:"history,omitempty"`
	// Visibility is the default config for visibility archival.
	// +optional
	Visibility *ArchivalSpec `json:"visibility,omitempty"`
}

func (s *ClusterArchivalSpec) IsEnabled() bool {
	return s != nil && s.Enabled
}

type ArchivalProviderKind string

const (
	FileStoreArchivalProviderKind ArchivalProviderKind = "filestore"
	S3ArchivalProviderKind        ArchivalProviderKind = "s3"
	GCSArchivalProviderKind       ArchivalProviderKind = "gcs"
	UnknownArchivalProviderKind   ArchivalProviderKind = "unknown"
)

// ArchivalProvider contains the config for archivers.
type ArchivalProvider struct {
	// +optional
	Filestore *FilestoreArchiver `json:"filestore,omitempty"`
	// +optional
	S3 *S3Archiver `json:"s3,omitempty"`
	// +optional
	GCS *GCSArchiver `json:"gcs,omitempty"`
}

func (p *ArchivalProvider) Kind() ArchivalProviderKind {
	if p.Filestore != nil {
		return FileStoreArchivalProviderKind
	}

	if p.S3 != nil {
		return S3ArchivalProviderKind
	}

	if p.GCS != nil {
		return GCSArchivalProviderKind
	}

	return UnknownArchivalProviderKind
}

// ArchivalSpec is the archival configuration for a particular persistence type (history or visibility).
type ArchivalSpec struct {
	// Enabled defines if the archival is enabled by default for all namespaces
	// or for a particular namespace (depends if it's for a TemporalCluster or a TemporalNamespace).
	// +kubebuilder:default:=false
	// +optional
	Enabled bool `json:"enabled"`
	// Paused defines if the archival is paused.
	// +kubebuilder:default:=false
	Paused bool `json:"paused"`
	// EnableRead allows temporal to read from the archived Event History.
	// +kubebuilder:default:=false
	EnableRead bool `json:"enableRead"`
	// Path is ...
	// +kubebuilder:validation:Required
	Path string `json:"path"`
}

// FilestoreArchiver is the file store archival provider configuration.
type FilestoreArchiver struct {
	// FilePermissions sets the file permissions of the archived files.
	// It's recommend to leave it empty and use the default value of "0666" to avoid read/write issues.
	// +kubebuilder:default:="0666"
	FilePermissions string `json:"filePermissions"`
	// DirPermissions sets the directory permissions of the archive directory.
	// It's recommend to leave it empty and use the default value of "0766" to avoid read/write issues.
	// +kubebuilder:default:="0766"
	DirPermissions string `json:"dirPermissions"`
}

// AuthorizationSpec defines the specifications for authorization in the temporal cluster. It contains fields
// that configure how JWT tokens are validated, how permissions are managed, and how claims are mapped.
type AuthorizationSpec struct {
	// JWTKeyProvider specifies the signing key provider used for validating JWT tokens.
	// +optional
	JWTKeyProvider AuthorizationSpecJWTKeyProvider `json:"jwtKeyProvider"`

	// PermissionsClaimName is the name of the claim within the JWT token that contains the user's permissions.
	// +optional
	PermissionsClaimName string `json:"permissionsClaimName"`

	// Authorizer defines the authorization mechanism to be used. It can be left as an empty string to
	// use a no-operation authorizer (noopAuthorizer), or set to "default" to use the temporal's default
	// authorizer (defaultAuthorizer).
	// +optional
	Authorizer string `json:"authorizer"`

	// ClaimMapper specifies the claim mapping mechanism used for handling JWT claims. Similar to the Authorizer,
	// it can be left as an empty string to use a no-operation claim mapper (noopClaimMapper), or set to "default"
	// to use the default JWT claim mapper (defaultJWTClaimMapper).
	// +optional
	ClaimMapper string `json:"claimMapper"`
}

// AuthorizationSpecJWTKeyProvider defines the configuration for a JWT key provider within the AuthorizationSpec.
// It specifies where to source the JWT keys from and how often they should be refreshed.
type AuthorizationSpecJWTKeyProvider struct {
	// KeySourceURIs is a list of URIs where the JWT signing keys can be obtained. These URIs are used by the
	// authorization system to fetch the public keys necessary for validating JWT tokens.
	// +optional
	KeySourceURIs []string `json:"keySourceURIs"`

	// RefreshInterval defines the time interval at which temporal should refresh the JWT signing keys from
	// the specified URIs.
	// +optional
	RefreshInterval *metav1.Duration `json:"refreshInterval"`
}

// S3Archiver is the S3 archival provider configuration.
type S3Archiver struct {
	// Region is the aws s3 region.
	// +kubebuilder:validation:Required
	Region string `json:"region"`
	// Use Endpoint if you want to use s3-compatible object storage.
	// +optional
	Endpoint *string `json:"endpoint,omitempty"`
	// Use RoleName if you want the temporal service account
	// to assume an AWS Identity and Access Management (IAM) role.
	// +optional
	RoleName *string `json:"roleName,omitempty"`
	// Use credentials if you want to use aws credentials from secret.
	// +optional
	Credentials *S3Credentials `json:"credentials,omitempty"`
	// Use s3ForcePathStyle if you want to use s3 path style.
	// +optional
	S3ForcePathStyle bool `json:"s3ForcePathStyle"`
}

type S3Credentials struct {
	// AccessKeyIDRef is the secret key selector containing AWS access key ID.
	// +kubebuilder:validation:Required
	AccessKeyIDRef *corev1.SecretKeySelector `json:"accessKeyIdRef"`
	// SecretAccessKeyRef is the secret key selector containing AWS secret access key.
	// +kubebuilder:validation:Required
	SecretAccessKeyRef *corev1.SecretKeySelector `json:"secretKeyRef"`
}

// GCSArchiver is the GCS archival provider configuration.
type GCSArchiver struct {
	// SecretAccessKeyRef is the secret key selector containing Google Cloud Storage credentials file.
	// +kubebuilder:validation:Required
	CredentialsRef *corev1.SecretKeySelector `json:"credentialsRef"`
}

func (GCSArchiver) CredentialsFileMountPath() string {
	return "/etc/archival/credentials.json"
}

// TemporalClusterSpec defines the desired state of Cluster.
type TemporalClusterSpec struct {
	// Image defines the temporal server docker image the cluster should use for each services.
	// +optional
	Image string `json:"image"`
	// Version defines the temporal version the cluster to be deployed.
	// This version impacts the underlying persistence schemas versions.
	// +optional
	Version *version.Version `json:"version"`
	// Log defines temporal cluster's logger configuration.
	// +optional
	Log *LogSpec `json:"log,omitempty"`
	// JobTTLSecondsAfterFinished is amount of time to keep job pods after jobs are completed.
	// Defaults to 300 seconds.
	// +optional
	//+kubebuilder:default:=300
	//+kubebuilder:validation:Minimum=1
	JobTTLSecondsAfterFinished *int32 `json:"jobTtlSecondsAfterFinished"`
	// JobResources allows set resources for setup/update jobs.
	// +optional
	JobResources corev1.ResourceRequirements `json:"jobResources,omitempty"`
	// JobInitContainers adds a list of init containers to the setup's jobs.
	// +optional
	JobInitContainers []corev1.Container `json:"jobInitContainers,omitempty"`
	// NumHistoryShards is the desired number of history shards.
	// This field is immutable.
	//+kubebuilder:validation:Minimum=1
	NumHistoryShards int32 `json:"numHistoryShards"`
	// Services allows customizations for each temporal services deployment.
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
	MTLS *MTLSSpec `json:"mTLS,omitempty"` //nolint:tagliatelle
	// Metrics allows configuration of scraping endpoints for stats. prometheus or m3.
	// +optional
	Metrics *MetricsSpec `json:"metrics,omitempty"`
	// DynamicConfig allows advanced configuration for the temporal cluster.
	// +optional
	DynamicConfig *DynamicConfigSpec `json:"dynamicConfig,omitempty"`
	// Archival allows Workflow Execution Event Histories and Visibility data backups for the temporal cluster.
	// +optional
	Archival *ClusterArchivalSpec `json:"archival,omitempty"`
	// Authorization allows authorization configuration for the temporal cluster.
	// +optional
	Authorization *AuthorizationSpec `json:"authorization,omitempty"`
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

// DatastoreStatus contains the current status of a datastore.
type DatastoreStatus struct {
	// Created indicates if the database or keyspace has been created.
	Created bool `json:"created"`
	// Setup indicates if tables have been set up.
	Setup bool `json:"setup"`
	// Type indicates the datastore type.
	// +optional
	Type DatastoreType `json:"type"`
	// SchemaVersion report the current schema version.
	// +optional
	SchemaVersion *version.Version `json:"schemaVersion,omitempty"`
}

// TemporalPersistenceStatus contains temporal persistence status.
type TemporalPersistenceStatus struct {
	// DefaultStore holds the default datastore status.
	DefaultStore *DatastoreStatus `json:"defaultStore"`
	// VisibilityStore holds the visibility datastore status.
	VisibilityStore *DatastoreStatus `json:"visibilityStore"`
	// SecondaryVisibilityStore holds the secondary visibility datastore status.
	// +optional
	SecondaryVisibilityStore *DatastoreStatus `json:"secondaryVisibilityStore"`
	// AdvancedVisibilityStore holds the advanced visibility datastore status.
	// +optional
	AdvancedVisibilityStore *DatastoreStatus `json:"advancedVisibilityStore,omitempty"`
}

// TemporalClusterStatus defines the observed state of Cluster.
type TemporalClusterStatus struct {
	// Version holds the current temporal version.
	Version string `json:"version,omitempty"`
	// Services holds all services statuses.
	Services []ServiceStatus `json:"services,omitempty"`
	// Persistence holds all datastores statuses.
	Persistence *TemporalPersistenceStatus `json:"persistence,omitempty"`
	// Conditions represent the latest available observations of the Cluster state.
	Conditions []metav1.Condition `json:"conditions"`
}

// AddServiceStatus adds the provided service status to the cluster's status.
func (s *TemporalClusterStatus) AddServiceStatus(status *ServiceStatus) {
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
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type == 'Ready')].status"
// +kubebuilder:printcolumn:name="ReconcileSuccess",type="string",JSONPath=".status.conditions[?(@.type == 'ReconcileSuccess')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:webhook:path=/validate-temporal-io-v1beta1-temporalcluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=temporal.io,resources=temporalclusters,verbs=create;update,versions=v1beta1,name=vtemporalc.kb.io,admissionReviewVersions=v1
// +kubebuilder:webhook:path=/mutate-temporal-io-v1beta1-temporalcluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=temporal.io,resources=temporalclusters,verbs=create;update,versions=v1beta1,name=mtemporalc.kb.io,admissionReviewVersions=v1

// TemporalCluster defines a temporal cluster deployment.
type TemporalCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the Temporal cluster.
	Spec TemporalClusterSpec `json:"spec,omitempty"`
	// Most recent observed status of the Temporal cluster.
	Status TemporalClusterStatus `json:"status,omitempty"`
}

func (c *TemporalCluster) SelectorLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":    c.GetName(),
		"app.kubernetes.io/part-of": "temporal",
	}
}

// ServerName returns cluster's server name.
func (c *TemporalCluster) ServerName() string {
	return fmt.Sprintf("%s.%s", c.Name, c.FQDNSuffix())
}

// FQDNSuffix returns the cluster's FQDN suffix.
func (c *TemporalCluster) FQDNSuffix() string {
	return fmt.Sprintf("%s.svc.cluster.local", c.Namespace)
}

// MTLSEnabled returns true if mTLS is enabled for internode or frontend using cert-manager.
func (c *TemporalCluster) MTLSWithCertManagerEnabled() bool {
	return c.Spec.MTLS != nil &&
		(c.Spec.MTLS.InternodeEnabled() || c.Spec.MTLS.FrontendEnabled()) &&
		c.Spec.MTLS.Provider == CertManagerMTLSProvider
}

// ChildResourceName returns child resource name using the cluster's name.
func (c *TemporalCluster) ChildResourceName(resource string) string {
	return fmt.Sprintf("%s-%s", c.Name, resource)
}

func (c *TemporalCluster) GetPublicClientAddress() string {
	// Use internal frontend if it's enabled, otherwise use regular frontend
	if c.Spec.Services != nil && c.Spec.Services.InternalFrontend.IsEnabled() {
		return fmt.Sprintf("%s.%s:%d", c.ChildResourceName("internal-frontend-headless"), c.GetNamespace(), *c.Spec.Services.InternalFrontend.Port)
	}
	return fmt.Sprintf("%s.%s:%d", c.ChildResourceName("frontend"), c.GetNamespace(), *c.Spec.Services.Frontend.Port)
}

// IsReady returns true if the TemporalCluster's conditions reports it ready.
func (c *TemporalCluster) IsReady() bool {
	for _, condition := range c.Status.Conditions {
		if condition.Type == ReadyCondition && condition.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

//+kubebuilder:object:root=true

// TemporalClusterList contains a list of Cluster.
type TemporalClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TemporalCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TemporalCluster{}, &TemporalClusterList{})
}
