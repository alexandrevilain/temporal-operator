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

package resource

import (
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/mtls/certmanager"
	"github.com/alexandrevilain/temporal-operator/pkg/temporal/persistence"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"go.temporal.io/server/common/cluster"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/dynamicconfig"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/metrics"
	"go.temporal.io/server/common/primitives"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ConfigmapBuilder struct {
	instance *v1beta1.TemporalCluster
	scheme   *runtime.Scheme
}

func NewConfigmapBuilder(instance *v1beta1.TemporalCluster, scheme *runtime.Scheme) *ConfigmapBuilder {
	return &ConfigmapBuilder{
		instance: instance,
		scheme:   scheme,
	}
}

func (b *ConfigmapBuilder) Build() (client.Object, error) {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName(ServiceConfig),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance.Name, ServiceConfig, b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}, nil
}

func (b *ConfigmapBuilder) buildDatastoreConfig(store *v1beta1.DatastoreSpec) (*config.DataStore, error) {
	cfg := &config.DataStore{}
	switch store.GetType() {
	case v1beta1.PostgresSQLDatastore,
		v1beta1.PostgresSQL12Datastore,
		v1beta1.MySQLDatastore,
		v1beta1.MySQL8Datastore:
		cfg.SQL = persistence.NewSQLConfigFromDatastoreSpec(store)
		cfg.SQL.Password = fmt.Sprintf("{{ .Env.%s }}", store.GetPasswordEnvVarName())
	case v1beta1.CassandraDatastore:
		cfg.Cassandra = persistence.NewCassandraConfigFromDatastoreSpec(store)
		cfg.Cassandra.Password = fmt.Sprintf("{{ .Env.%s }}", store.GetPasswordEnvVarName())
	case v1beta1.ElasticsearchDatastore:
		esCfg, err := persistence.NewElasticsearchConfigFromDatastoreSpec(store)
		if err != nil {
			return nil, fmt.Errorf("can't get elasticsearch config: %w", err)
		}
		cfg.Elasticsearch = esCfg
		cfg.Elasticsearch.Password = fmt.Sprintf("{{ .Env.%s }}", store.GetPasswordEnvVarName())
	case v1beta1.UnknownDatastore:
		return nil, errors.New("unknown datastore")
	}
	return cfg, nil
}

func (b *ConfigmapBuilder) buildPersistenceConfig() (*config.Persistence, error) {
	cfg := &config.Persistence{
		NumHistoryShards: b.instance.Spec.NumHistoryShards,
		DefaultStore:     b.instance.Spec.Persistence.DefaultStore.Name,
		VisibilityStore:  b.instance.Spec.Persistence.VisibilityStore.Name,
		DataStores:       map[string]config.DataStore{},
	}

	if b.instance.Spec.Persistence.AdvancedVisibilityStore != nil {
		cfg.AdvancedVisibilityStore = v1beta1.AdvancedVisibilityStoreName
	}

	for _, store := range b.instance.Spec.Persistence.GetDatastores() {
		storeConfig, err := b.buildDatastoreConfig(store)
		if err != nil {
			return nil, err
		}

		cfg.DataStores[store.Name] = *storeConfig
	}

	return cfg, nil
}

func (b *ConfigmapBuilder) Update(object client.Object) error {
	configMap := object.(*corev1.ConfigMap)

	persistenceConfig, err := b.buildPersistenceConfig()
	if err != nil {
		return fmt.Errorf("can't build persistence config: %w", err)
	}

	temporalCfg := config.Config{
		Global: config.Global{
			Membership: config.Membership{
				MaxJoinDuration:  30 * time.Second,
				BroadcastAddress: "{{ default .Env.POD_IP \"0.0.0.0\" }}",
			},
		},
		Persistence: *persistenceConfig,
		Log: log.Config{
			Stdout: true,
			Level:  "info",
		},
		ClusterMetadata: &cluster.Config{
			EnableGlobalNamespace:    false,
			FailoverVersionIncrement: 10,
			MasterClusterName:        b.instance.Name,
			CurrentClusterName:       b.instance.Name,
			ClusterInformation: map[string]cluster.ClusterInformation{
				b.instance.Name: {
					Enabled:                true,
					InitialFailoverVersion: 1,
					RPCAddress:             "127.0.0.1:7233",
				},
			},
		},
		Services: map[string]config.Service{
			string(primitives.FrontendService): {
				RPC: config.RPC{
					GRPCPort:        *b.instance.Spec.Services.Frontend.Port,
					MembershipPort:  *b.instance.Spec.Services.Frontend.MembershipPort,
					BindOnLocalHost: false,
					BindOnIP:        "0.0.0.0",
				},
			},
			string(primitives.HistoryService): {
				RPC: config.RPC{
					GRPCPort:        *b.instance.Spec.Services.History.Port,
					MembershipPort:  *b.instance.Spec.Services.History.MembershipPort,
					BindOnLocalHost: false,
					BindOnIP:        "0.0.0.0",
				},
			},
			string(primitives.MatchingService): {
				RPC: config.RPC{
					GRPCPort:        *b.instance.Spec.Services.Matching.Port,
					MembershipPort:  *b.instance.Spec.Services.Matching.MembershipPort,
					BindOnLocalHost: false,
					BindOnIP:        "0.0.0.0",
				},
			},
			string(primitives.WorkerService): {
				RPC: config.RPC{
					GRPCPort:        *b.instance.Spec.Services.Worker.Port,
					MembershipPort:  *b.instance.Spec.Services.Worker.MembershipPort,
					BindOnLocalHost: false,
					BindOnIP:        "0.0.0.0",
				},
			},
		},
	}

	if b.instance.Spec.Version.GreaterOrEqual(version.V1_20_0) {
		if b.instance.Spec.Services.InternalFrontend.IsEnabled() {
			temporalCfg.Services[string(primitives.InternalFrontendService)] = config.Service{
				RPC: config.RPC{
					GRPCPort:        *b.instance.Spec.Services.InternalFrontend.Port,
					MembershipPort:  *b.instance.Spec.Services.InternalFrontend.MembershipPort,
					BindOnLocalHost: false,
					BindOnIP:        "0.0.0.0",
				},
			}
		}
	}

	if !b.instance.Spec.Version.GreaterOrEqual(version.V1_18_0) {
		temporalCfg.PublicClient = config.PublicClient{
			HostPort: b.instance.GetPublicClientAddress(),
		}
	}

	if b.instance.Spec.DynamicConfig != nil {
		temporalCfg.DynamicConfigClient = &dynamicconfig.FileBasedClientConfig{
			Filepath:     "/etc/temporal/config/dynamic_config.yaml",
			PollInterval: b.instance.Spec.DynamicConfig.PollInterval.Duration,
		}
	}

	if b.instance.Spec.Metrics.IsEnabled() {
		temporalCfg.Global.Metrics = &metrics.Config{
			ClientConfig: metrics.ClientConfig{
				Tags: map[string]string{"type": "{{ .Env.SERVICES }}"},
			},
		}
		if b.instance.Spec.Metrics.Prometheus != nil && b.instance.Spec.Metrics.Prometheus.ListenPort != nil {
			temporalCfg.Global.Metrics.Prometheus = &metrics.PrometheusConfig{
				TimerType:     "histogram",
				ListenAddress: fmt.Sprintf("0.0.0.0:%d", *b.instance.Spec.Metrics.Prometheus.ListenPort),
			}
		}
	}

	if b.instance.MTLSWithCertManagerEnabled() {
		temporalCfg.Global.TLS = config.RootTLS{
			RefreshInterval:  b.instance.Spec.MTLS.RefreshInterval.Duration,
			ExpirationChecks: config.CertExpirationValidation{},
		}

		internodeMTLS := b.instance.Spec.MTLS.Internode
		if internodeMTLS == nil {
			internodeMTLS = &v1beta1.InternodeMTLSSpec{}
		}

		internodeIntermediateCAFilePath := path.Join(internodeMTLS.GetIntermediateCACertificateMountPath(), certmanager.TLSCA)
		internodeServerCertFilePath := path.Join(internodeMTLS.GetCertificateMountPath(), certmanager.TLSCert)
		internodeServerKeyFilePath := path.Join(internodeMTLS.GetCertificateMountPath(), certmanager.TLSKey)
		internodeClientTLS := config.ClientTLS{
			ServerName:              internodeMTLS.ServerName(b.instance.ServerName()),
			DisableHostVerification: false,
			RootCAFiles:             []string{internodeIntermediateCAFilePath},
			ForceTLS:                true,
		}

		if b.instance.Spec.MTLS.InternodeEnabled() {
			temporalCfg.Global.TLS.Internode = config.GroupTLS{
				Client: internodeClientTLS,
				Server: config.ServerTLS{
					CertFile: internodeServerCertFilePath,
					KeyFile:  internodeServerKeyFilePath,
					ClientCAFiles: []string{
						internodeIntermediateCAFilePath,
					},
					RequireClientAuth: true,
				},
			}

			// If internode mTLs is enabled and internal frontend is enabled,
			// use internode mTLS certificates for the worker TLS.
			if b.instance.Spec.Services.InternalFrontend.IsEnabled() {
				temporalCfg.Global.TLS.SystemWorker = config.WorkerTLS{
					Client:   internodeClientTLS,
					CertFile: internodeServerCertFilePath,
					KeyFile:  internodeServerKeyFilePath,
				}
			}
		}

		if b.instance.Spec.MTLS.FrontendEnabled() {
			frontendMTLS := b.instance.Spec.MTLS.Frontend
			frontendIntermediateCAFilePath := path.Join(frontendMTLS.GetIntermediateCACertificateMountPath(), certmanager.TLSCA)

			temporalCfg.Global.TLS.Frontend = config.GroupTLS{
				Server: config.ServerTLS{
					CertFile:          path.Join(frontendMTLS.GetCertificateMountPath(), certmanager.TLSCert),
					KeyFile:           path.Join(frontendMTLS.GetCertificateMountPath(), certmanager.TLSKey),
					RequireClientAuth: true,
					ClientCAFiles: []string{
						internodeIntermediateCAFilePath,
						frontendIntermediateCAFilePath,
					},
				},
				Client: config.ClientTLS{
					ServerName:              frontendMTLS.ServerName(b.instance.ServerName()),
					DisableHostVerification: false,
					RootCAFiles:             []string{frontendIntermediateCAFilePath},
					ForceTLS:                true,
				},
				PerHostOverrides: map[string]config.ServerTLS{},
			}

			// If the Frontend mTLS is enabled, and if the internal frontend with internode mTLS is not enabled, the system worker should use the Frontend mTLS.
			if !(b.instance.Spec.MTLS.InternodeEnabled() && b.instance.Spec.Services.InternalFrontend.IsEnabled()) {
				temporalCfg.Global.TLS.SystemWorker = config.WorkerTLS{
					CertFile: path.Join(frontendMTLS.GetWorkerCertificateMountPath(), certmanager.TLSCert),
					KeyFile:  path.Join(frontendMTLS.GetWorkerCertificateMountPath(), certmanager.TLSKey),
					Client: config.ClientTLS{
						ServerName:              frontendMTLS.ServerName(b.instance.ServerName()),
						DisableHostVerification: false,
						RootCAFiles:             []string{frontendIntermediateCAFilePath},
						ForceTLS:                true,
					},
				}
			}
		}
	}

	result, err := yaml.Marshal(temporalCfg)
	if err != nil {
		return fmt.Errorf("failed marshaling temporal config: %w", err)
	}

	configMap.Data = map[string]string{
		"config_template.yaml": string(result),
	}

	if err := controllerutil.SetControllerReference(b.instance, configMap, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %w", err)
	}

	return nil
}
