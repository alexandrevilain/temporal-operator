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

package temporal

import (
	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"go.temporal.io/api/replication/v1"
	"go.temporal.io/api/workflowservice/v1"
)

func NamespaceToRegisterNamespaceRequest(namespace *v1alpha1.TemporalNamespace) *workflowservice.RegisterNamespaceRequest {
	re := &workflowservice.RegisterNamespaceRequest{
		Namespace:     namespace.GetName(),
		Description:   namespace.Spec.Description,
		OwnerEmail:    namespace.Spec.OwnerEmail,
		Data:          namespace.Spec.Data,
		SecurityToken: namespace.Spec.SecurityToken,
		// Not supported yet:
		// HistoryArchivalState:    0,
		// HistoryArchivalUri:      "",
		// VisibilityArchivalState: 0,
		// VisibilityArchivalUri:   "",
	}

	if namespace.Spec.RetentionPeriod != nil {
		re.WorkflowExecutionRetentionPeriod = &namespace.Spec.RetentionPeriod.Duration
	}

	if namespace.Spec.IsGlobalNamespace {
		re.IsGlobalNamespace = true

		if len(namespace.Spec.Clusters) > 0 {
			re.Clusters = make([]*replication.ClusterReplicationConfig, 0, len(namespace.Spec.Clusters))
			for _, name := range namespace.Spec.Clusters {
				re.Clusters = append(re.Clusters, &replication.ClusterReplicationConfig{
					ClusterName: name,
				})
			}
		}

		re.ActiveClusterName = namespace.Spec.ActiveClusterName
	}

	return re
}
