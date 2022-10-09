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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TemporalAppWorkerSpec defines the desired state of TemporalAppWorker
type TemporalAppWorkerSpec struct {
	// Version defines the app worker version.
	Version string `json:"version"`
	// Image defines the temporal ui docker image the instance should run.
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

// TemporalAppWorkerStatus defines the observed state of TemporalAppWorker
type TemporalAppWorkerStatus struct {
	// Conditions represent the latest available observations of the worker state.
	Conditions []metav1.Condition `json:"conditions"`
	// Number of desired replicas. Default to 1.
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas *int32 `json:"replicas"`
}

// +genclient
// +genclient:Namespaced
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type == 'Ready')].status"
// +kubebuilder:printcolumn:name="ReconcileSuccess",type="string",JSONPath=".status.conditions[?(@.type == 'ReconcileSuccess')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// TemporalAppWorker is the Schema for the temporalappworkers API
type TemporalAppWorker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemporalAppWorkerSpec   `json:"spec,omitempty"`
	Status TemporalAppWorkerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TemporalAppWorkerList contains a list of TemporalAppWorker
type TemporalAppWorkerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TemporalAppWorker `json:"items"`
}

// ChildResourceName returns child resource name using the app worker's name.
func (c *TemporalAppWorker) ChildResourceName(resource string) string {
	return fmt.Sprintf("%s-%s", c.Name, resource)
}

func init() {
	SchemeBuilder.Register(&TemporalAppWorker{}, &TemporalAppWorkerList{})
}
