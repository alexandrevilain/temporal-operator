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

// TemporalClusterClientSpec defines the desired state of TemporalClusterClient
type TemporalClusterClientSpec struct {
	// Reference to the temporal cluster the client will get access to.
	TemporalClusterRef corev1.LocalObjectReference `json:"temporalClusterRef"`
}

// TemporalClusterClientStatus defines the observed state of TemporalClusterClient
type TemporalClusterClientStatus struct {
	// Reference to the Kubernetes Secret containing the certificate for the client.
	SecretRef corev1.LocalObjectReference `json:"secretRef"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TemporalClusterClient is the Schema for the temporalclusterclients API
type TemporalClusterClient struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemporalClusterClientSpec   `json:"spec,omitempty"`
	Status TemporalClusterClientStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TemporalClusterClientList contains a list of TemporalClusterClient
type TemporalClusterClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TemporalClusterClient `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TemporalClusterClient{}, &TemporalClusterClientList{})
}
