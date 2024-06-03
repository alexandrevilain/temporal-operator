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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TemporalNamespaceArchivalSpec is a per-namespace archival configuration override.
type TemporalNamespaceArchivalSpec struct {
	// History is the config for this namespace history archival.
	// +optional
	History *ArchivalSpec `json:"history,omitempty"`
	// Visibility is the config for this namespace visibility archival.
	// +optional
	Visibility *ArchivalSpec `json:"visibility,omitempty"`
}

// TemporalNamespaceSpec defines the desired state of Namespace.
type TemporalNamespaceSpec struct {
	// Reference to the temporal cluster the namespace will be created.
	ClusterRef TemporalReference `json:"clusterRef"`
	// Namespace description.
	// +optional
	Description string `json:"description,omitempty"`
	// Namespace owner email.
	// +optional
	OwnerEmail string `json:"ownerEmail,omitempty"`
	// RetentionPeriod to apply on closed workflow executions.
	RetentionPeriod *metav1.Duration `json:"retentionPeriod"`
	// Data is a key-value map for any customized purpose.
	// +optional
	Data map[string]string `json:"data,omitempty"`
	// +optional
	SecurityToken string `json:"securityToken,omitempty"`
	// IsGlobalNamespace defines whether the namespace is a global namespace.
	// +optional
	IsGlobalNamespace bool `json:"isGlobalNamespace,omitempty"`
	// List of clusters names to which the namespace can fail over.
	// Only applicable if the namespace is a global namespace.
	// +optional
	Clusters []string `json:"clusters,omitempty"`
	// The name of active Temporal Cluster.
	// Only applicable if the namespace is a global namespace.
	// +optional
	ActiveClusterName string `json:"activeClusterName,omitempty"`
	// AllowDeletion makes the controller delete the Temporal namespace if the
	// CRD is deleted.
	// +optional
	AllowDeletion bool `json:"allowDeletion,omitempty"`
	// Archival is a per-namespace archival configuration.
	// If not set, the default cluster configuration is used.
	// +optional
	Archival *TemporalNamespaceArchivalSpec `json:"archival,omitempty"`
}

// TemporalNamespaceStatus defines the observed state of Namespace.
type TemporalNamespaceStatus struct {
	// Conditions represent the latest available observations of the Namespace state.
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// A TemporalNamespace creates a namespace in the targeted temporal cluster.
type TemporalNamespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemporalNamespaceSpec   `json:"spec,omitempty"`
	Status TemporalNamespaceStatus `json:"status,omitempty"`
}

// IsReady returns true if the TemporalNamespace's conditions reports it ready.
func (c *TemporalNamespace) IsReady() bool {
	for _, condition := range c.Status.Conditions {
		if condition.Type == ReadyCondition && condition.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

//+kubebuilder:object:root=true

// TemporalNamespaceList contains a list of Namespace.
type TemporalNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TemporalNamespace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TemporalNamespace{}, &TemporalNamespaceList{})
}
