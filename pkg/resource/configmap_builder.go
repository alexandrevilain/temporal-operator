package resource

import (
	"fmt"
	"time"

	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	"github.com/alexandrevilain/temporal-operator/pkg/persistence"
	"go.temporal.io/server/common"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/log"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ConfigmapBuilder struct {
	instance *v1alpha1.TemporalCluster
	scheme   *runtime.Scheme
}

func NewConfigmapBuilder(instance *v1alpha1.TemporalCluster, scheme *runtime.Scheme) *ConfigmapBuilder {
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
			Labels:      metadata.GetLabels(b.instance.Name, ServiceConfig, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}, nil
}

func (b *ConfigmapBuilder) Update(object client.Object) error {
	configMap := object.(*corev1.ConfigMap)

	datastores := map[string]config.DataStore{}
	for _, store := range b.instance.Spec.Datastores {
		cfg := config.DataStore{}
		if store.SQL != nil {
			cfg.SQL = persistence.NewSQLconfigFromDatastoreSpec(&store)
			cfg.SQL.Password = fmt.Sprintf("{{ .Env.%s }}", store.GetPasswordEnvVarName())
		}
		datastores[store.Name] = cfg
	}

	temporalCfg := config.Config{
		Global: config.Global{
			Membership: config.Membership{
				MaxJoinDuration:  30 * time.Second,
				BroadcastAddress: "{{ default .Env.POD_IP \"0.0.0.0\" }}",
			},
			// TLS: config.RootTLS{
			// 	Internode: config.GroupTLS{
			// 		Client:           config.ClientTLS{},
			// 		Server:           config.ServerTLS{},
			// 		PerHostOverrides: map[string]config.ServerTLS{},
			// 	},
			// 	Frontend:         config.GroupTLS{},
			// 	SystemWorker:     config.WorkerTLS{},
			// 	RemoteClusters:   map[string]config.GroupTLS{},
			// 	ExpirationChecks: config.CertExpirationValidation{},
			// 	RefreshInterval:  0,
			// },
			// Metrics:       &metrics.Config{},
			// Authorization: config.Authorization{},
		},
		Persistence: config.Persistence{
			DefaultStore:            b.instance.Spec.Persistence.DefaultStore,
			VisibilityStore:         b.instance.Spec.Persistence.VisibilityStore,
			AdvancedVisibilityStore: b.instance.Spec.Persistence.AdvancedVisibilityStore,
			NumHistoryShards:        b.instance.Spec.NumHistoryShards,
			DataStores:              datastores,
		},
		Log: log.Config{
			Stdout: true,
			Level:  "info",
		},
		// ClusterMetadata: &cluster.Config{
		// 	EnableGlobalNamespace:    false,
		// 	FailoverVersionIncrement: 0,
		// 	MasterClusterName:        "",
		// 	CurrentClusterName:       "",
		// 	ClusterInformation:       map[string]cluster.ClusterInformation{},
		// },
		// DCRedirectionPolicy: config.DCRedirectionPolicy{
		// 	Policy: "",
		// 	ToDC:   "",
		// },
		Services: map[string]config.Service{
			common.FrontendServiceName: {
				RPC: config.RPC{
					GRPCPort:        *b.instance.Spec.Services.Frontend.Port,
					MembershipPort:  *b.instance.Spec.Services.Frontend.MembershipPort,
					BindOnLocalHost: false,
					BindOnIP:        "0.0.0.0",
				},
			},
			common.HistoryServiceName: {
				RPC: config.RPC{
					GRPCPort:        *b.instance.Spec.Services.History.Port,
					MembershipPort:  *b.instance.Spec.Services.History.MembershipPort,
					BindOnLocalHost: false,
					BindOnIP:        "0.0.0.0",
				},
			},
			common.MatchingServiceName: {
				RPC: config.RPC{
					GRPCPort:        *b.instance.Spec.Services.Matching.Port,
					MembershipPort:  *b.instance.Spec.Services.Matching.MembershipPort,
					BindOnLocalHost: false,
					BindOnIP:        "0.0.0.0",
				},
			},
			common.WorkerServiceName: {
				RPC: config.RPC{
					GRPCPort:        *b.instance.Spec.Services.Worker.Port,
					MembershipPort:  *b.instance.Spec.Services.Worker.MembershipPort,
					BindOnLocalHost: false,
					BindOnIP:        "0.0.0.0",
				},
			},
		},
		// Archival: config.Archival{
		// 	History:    config.HistoryArchival{},
		// 	Visibility: config.VisibilityArchival{},
		// },
		PublicClient: config.PublicClient{
			HostPort: fmt.Sprintf("%s:%d", b.instance.ChildResourceName("frontend"), b.instance.Spec.Services.Frontend.Port),
		},
		// DynamicConfigClient: &dynamicconfig.FileBasedClientConfig{
		// 	Filepath:     "",
		// 	PollInterval: 0,
		// },
		// NamespaceDefaults: config.NamespaceDefaults{
		// 	Archival: config.ArchivalNamespaceDefaults{
		// 		History: config.HistoryArchivalNamespaceDefaults{
		// 			State: "",
		// 			URI:   "",
		// 		},
		// 		Visibility: config.VisibilityArchivalNamespaceDefaults{
		// 			State: "",
		// 			URI:   "",
		// 		},
		// 	},
		// },
	}

	result, err := yaml.Marshal(temporalCfg)
	if err != nil {
		return fmt.Errorf("failed marshaling temporal config: %w", err)
	}

	configMap.Data = map[string]string{
		"config_template.yaml": string(result),
	}

	if err := controllerutil.SetControllerReference(b.instance, configMap, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}
