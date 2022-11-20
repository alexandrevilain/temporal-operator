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
	// Reference to the temporal cluster the worker will connect to.
	ClusterRef *TemporalClusterReference `json:"clusterRef"`
	// Version defines the worker process version.
	// +optional
	Version string `json:"version"`
	// Image defines the temporal worker docker image the instance should run.
	Image string `json:"image"`
	// JobTtlSecondsAfterFinished is amount of time to keep job pods after jobs are completed.
	// Defaults to 300 seconds.
	// +optional
	//+kubebuilder:default:=300
	//+kubebuilder:validation:Minimum=1
	JobTtlSecondsAfterFinished *int32 `json:"jobTtlSecondsAfterFinished"`
	// Number of desired replicas. Default to 1.
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas *int32 `json:"replicas"`
	// Image pull policy for determining how to pull worker process images.
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
	// An optional list of references to secrets in the same namespace
	// to use for pulling temporal images from registries.
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// TemporalNamespace that worker will poll.
	TemporalNamespace string `json:"temporalNamespace"`
	// Builder is the configuration for building a TemporalWorkerProcess
	Builder *TemporalWorkerProcessBuilder `json:"builder,omitempty"`
}

// Reference to TemporalCluster
type TemporalClusterReference struct {
	// The name of the TemporalCluster to reference.
	Name string `json:"name,omitempty"`
	// The namespace of the TemporalCluster to reference.
	// Defaults to the namespace of the requested resource if omitted.
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`
}

// TemporalWorkerProcessStatus defines the observed state of TemporalWorkerProcess
type TemporalWorkerProcessStatus struct {
	// Created indicates if the worker process image was created.
	Created bool `json:"created"`
	// Ready defines if the worker process is ready.
	Ready bool `json:"ready"`
	// Version is the version of the image that will be used to build worker image.
	Version string `json:"version"`
	// BuildAttempt is the build attempt number of a given version
	// +required
	BuildAttempt *int32 `json:"attempt"`
	// Conditions represent the latest available observations of the worker process state.
	Conditions []metav1.Condition `json:"conditions"`
}

type TemporalWorkerProcessBuilder struct {
	// Enabled defines if the operator should build the temporal worker process.
	// +required
	Enabled bool `json:"enabled"`
	// Version is the version of the image that will be used to build worker image.
	Version string `json:"version,omitempty"`
	// BuildAttempt is the build attempt number of a given version
	BuildAttempt *int32 `json:"attempt,omitempty"`
	// Image is the image that will be used to build worker image.
	Image string `json:"image,omitempty"`
	//BuildDir is the location of where the sources will be built.
	BuildDir string `json:"buildDir,omitempty"`
	// GitRepository specifies how to connect to Git source control.
	GitRepository *GitRepositorySpec `json:"gitRepository,omitempty"`
	// BuildRegistry specifies how to connect to container registry.
	BuildRegistry *ContainerRegistryConfig `json:"buildRegistry,omitempty"`
}

// GetPasswordEnvVarName crafts the environment variable name for the datastore.
func (s *TemporalWorkerProcessBuilder) GetBuildRepoPasswordEnvVarName() string {
	return "TEMPORAL_WORKER_BUILDER_REPO_PASSWORD"
}

func (w *TemporalWorkerProcessBuilder) BuilderEnabled() bool {
	return w != nil && w.Enabled
}

// GitRepositorySpec specifies the required configuration to produce an
// Artifact for a Git repository.
type GitRepositorySpec struct {
	// URL specifies the Git repository URL, it can be an HTTP/S or SSH address.
	// +kubebuilder:validation:Pattern="^(http|https|ssh)://.*$"
	// +required
	URL string `json:"url"`
	// Reference specifies the Git reference to resolve and monitor for
	// changes, defaults to the 'master' branch.
	// +optional
	Reference *GitRepositoryRef `json:"reference,omitempty"`
}

// GitRepositoryRef specifies the Git reference to resolve and checkout.
type GitRepositoryRef struct {
	// Branch to check out, defaults to 'main' if no other field is defined.
	// +optional
	Branch string `json:"branch,omitempty"`
}

// ContainerRegistryConfig specifies the parameters to connect with desired container repository.
type ContainerRegistryConfig struct {
	// Repository is the fqdn to the image repo.
	// +required
	Repository string `json:"repository"`
	// Username is the username for the container repo.
	// +required
	Username string `json:"username"`
	// PasswordSecret is the reference to the secret holding the docker repo password.
	// +required
	PasswordSecretRef SecretKeyReference `json:"passwordSecretRef"`
}

// AddWorkerProcessStatus adds the provided worker process status.
func (s *TemporalWorkerProcessStatus) AddWorkerDeploymentStatus(status *TemporalWorkerProcessStatus) {
	s.Ready = status.Ready
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

// ChildResourceName returns child resource name using the worker processes name.
func (w *TemporalWorkerProcess) ChildResourceName(resource string) string {
	return fmt.Sprintf("%s-%s", w.Name, resource)
}

func init() {
	SchemeBuilder.Register(&TemporalWorkerProcess{}, &TemporalWorkerProcessList{})
}
