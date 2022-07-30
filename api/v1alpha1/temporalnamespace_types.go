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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TemporalNamespaceSpec defines the desired state of TemporalNamespace
type TemporalNamespaceSpec struct {
	// Reference to the temporal cluster the namespace will be created.
	TemporalClusterRef corev1.LocalObjectReference `json:"temporalClusterRef"`
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
}

// TemporalNamespaceStatus defines the observed state of TemporalNamespace
type TemporalNamespaceStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TemporalNamespace is the Schema for the temporalnamespaces API
type TemporalNamespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemporalNamespaceSpec   `json:"spec,omitempty"`
	Status TemporalNamespaceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TemporalNamespaceList contains a list of TemporalNamespace
type TemporalNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TemporalNamespace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TemporalNamespace{}, &TemporalNamespaceList{})
}
