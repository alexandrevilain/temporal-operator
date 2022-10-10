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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TemporalWorkerProcessSpec defines the desired state of TemporalWorkerProcess
type TemporalWorkerProcessSpec struct {
	// Version defines the app worker version.
	Version string `json:"version"`
	// Image defines the temporal worker docker image the instance should run.
	Image string `json:"image"`
	// Number of desired replicas. Default to 1.
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas *int32 `json:"replicas"`
	// An optional list of references to secrets in the same namespace
	// to use for pulling temporal images from registries.
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// TemporalServer connection details.
	TemporalConnection *TemporalConnectionSpec `json:"temporalConnection"`
}

// TemporalConnectionSpec defines the attributes for connecting to a Temporal server.
type TemporalConnectionSpec struct {
	// FQDN of the temporal frontend service endpoint.
	URL string `json:"url"`
	// Port where the temporal frontend service is listening.
	Port *int `json:"port"`
	// Namespace that worker will poll.
	Namespace string `json:"namespace"`
}

// TemporalWorkerProcessStatus defines the observed state of TemporalWorkerProcess
type TemporalWorkerProcessStatus struct {
	// Conditions represent the latest available observations of the worker state.
	Conditions []metav1.Condition `json:"conditions"`
}

// +genclient
// +genclient:Namespaced
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type == 'Ready')].status"
// +kubebuilder:printcolumn:name="ReconcileSuccess",type="string",JSONPath=".status.conditions[?(@.type == 'ReconcileSuccess')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// TemporalWorkerProcess is the Schema for the temporalworkerprocesses API
type TemporalWorkerProcess struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemporalWorkerProcessSpec   `json:"spec,omitempty"`
	Status TemporalWorkerProcessStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TemporalWorkerProcessList contains a list of TemporalWorkerProcess
type TemporalWorkerProcessList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TemporalWorkerProcess `json:"items"`
}

// ChildResourceName returns child resource name using the app worker's name.
func (c *TemporalWorkerProcess) ChildResourceName(resource string) string {
	return fmt.Sprintf("%s-%s", c.Name, resource)
}

func init() {
	SchemeBuilder.Register(&TemporalWorkerProcess{}, &TemporalWorkerProcessList{})
}
