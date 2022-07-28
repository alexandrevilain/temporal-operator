package temporal

import (
	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"go.temporal.io/api/replication/v1"
	"go.temporal.io/api/workflowservice/v1"
)

func NamespaceSpecToRegisterNamespaceRequest(spec *v1alpha1.NamespaceSpec) *workflowservice.RegisterNamespaceRequest {
	re := &workflowservice.RegisterNamespaceRequest{
		Namespace:     spec.Name,
		Description:   spec.Description,
		OwnerEmail:    spec.OwnerEmail,
		Data:          spec.Data,
		SecurityToken: spec.SecurityToken,
		// Not supported yet:
		// HistoryArchivalState:    0,
		// HistoryArchivalUri:      "",
		// VisibilityArchivalState: 0,
		// VisibilityArchivalUri:   "",
	}

	if spec.RetentionPeriod != nil {
		re.WorkflowExecutionRetentionPeriod = &spec.RetentionPeriod.Duration
	}

	if spec.IsGlobalNamespace {
		re.IsGlobalNamespace = true

		if len(spec.Clusters) > 0 {
			re.Clusters = make([]*replication.ClusterReplicationConfig, 0, len(spec.Clusters))
			for _, name := range spec.Clusters {
				re.Clusters = append(re.Clusters, &replication.ClusterReplicationConfig{
					ClusterName: name,
				})
			}
		}

		re.ActiveClusterName = spec.ActiveClusterName
	}

	return re
}
