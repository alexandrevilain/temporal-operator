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

package status_test

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/status"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestObservedVersionMatchesDesiredVersion(t *testing.T) {
	tests := map[string]struct {
		cluster  *v1beta1.TemporalCluster
		expected bool
	}{
		"all services matches the desired version": {
			cluster: &v1beta1.TemporalCluster{
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.16.0"),
				},
				Status: v1beta1.TemporalClusterStatus{
					Services: []v1beta1.ServiceStatus{
						{
							Name:    "test",
							Version: "1.16.0",
						},
						{
							Name:    "test2",
							Version: "1.16.0",
						},
					},
				},
			},
			expected: true,
		},
		"services does not match the desired version": {
			cluster: &v1beta1.TemporalCluster{
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.16.0"),
				},
				Status: v1beta1.TemporalClusterStatus{
					Services: []v1beta1.ServiceStatus{
						{
							Name:    "test",
							Version: "1.15.0",
						},
						{
							Name:    "test2",
							Version: "1.16.0",
						},
					},
				},
			},
			expected: false,
		},
		"empty status": {
			cluster: &v1beta1.TemporalCluster{
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.16.0"),
				},
				Status: v1beta1.TemporalClusterStatus{},
			},
			expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			assert.Equal(tt, test.expected, status.ObservedVersionMatchesDesiredVersion(test.cluster))
		})
	}
}

func TestIsClusterReady(t *testing.T) {
	tests := map[string]struct {
		cluster  *v1beta1.TemporalCluster
		expected bool
	}{
		"all services are ready": {
			cluster: &v1beta1.TemporalCluster{
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.20.1"),
				},
				Status: v1beta1.TemporalClusterStatus{
					Services: []v1beta1.ServiceStatus{
						{
							Name:    "test",
							Version: "1.20.1",
							Ready:   true,
						},
						{
							Name:    "test2",
							Version: "1.20.1",
							Ready:   true,
						},
					},
				},
			},
			expected: true,
		},
		"one service not ready": {
			cluster: &v1beta1.TemporalCluster{
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.20.1"),
				},
				Status: v1beta1.TemporalClusterStatus{
					Services: []v1beta1.ServiceStatus{
						{
							Name:    "test",
							Version: "1.20.1",
							Ready:   true,
						},
						{
							Name:    "test2",
							Version: "1.20.1",
							Ready:   false,
						},
					},
				},
			},
			expected: false,
		},
		"empty status": {
			cluster: &v1beta1.TemporalCluster{
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.20.1"),
				},
				Status: v1beta1.TemporalClusterStatus{},
			},
			expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			assert.Equal(tt, test.expected, status.IsClusterReady(test.cluster))
		})
	}
}

func TestReconciledObjectsToServiceStatuses(t *testing.T) {
	tests := map[string]struct {
		cluster  *v1beta1.TemporalCluster
		objects  []client.Object
		expected []*v1beta1.ServiceStatus
	}{
		"empty object list": {
			cluster: &v1beta1.TemporalCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1beta1.TemporalClusterSpec{},
			},
			objects:  []client.Object{},
			expected: []*v1beta1.ServiceStatus{},
		},
		"frontend service ready without version": {
			cluster: &v1beta1.TemporalCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1beta1.TemporalClusterSpec{},
			},
			objects: []client.Object{
				&appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-frontend",
						Namespace: "default",
					},
					Status: appsv1.DeploymentStatus{
						ObservedGeneration: 1,
						UpdatedReplicas:    1,
						ReadyReplicas:      1,
						AvailableReplicas:  1,
						Replicas:           1,
						Conditions: []appsv1.DeploymentCondition{
							{
								Type:   appsv1.DeploymentAvailable,
								Status: corev1.ConditionTrue,
							},
							{
								Type:   appsv1.DeploymentProgressing,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			expected: []*v1beta1.ServiceStatus{
				{
					Name:    "frontend",
					Ready:   true,
					Version: "0.0.0",
				},
			},
		},
		"frontend service ready with version": {
			cluster: &v1beta1.TemporalCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1beta1.TemporalClusterSpec{},
			},
			objects: []client.Object{
				&appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-frontend",
						Namespace: "default",
						Labels: map[string]string{
							"app.kubernetes.io/version": "1.2.3",
						},
					},
					Status: appsv1.DeploymentStatus{
						ObservedGeneration: 1,
						UpdatedReplicas:    1,
						ReadyReplicas:      1,
						AvailableReplicas:  1,
						Replicas:           1,
						Conditions: []appsv1.DeploymentCondition{
							{
								Type:   appsv1.DeploymentAvailable,
								Status: corev1.ConditionTrue,
							},
							{
								Type:   appsv1.DeploymentProgressing,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			expected: []*v1beta1.ServiceStatus{
				{
					Name:    "frontend",
					Ready:   true,
					Version: "1.2.3",
				},
			},
		},
		"frontend service not ready with version": {
			cluster: &v1beta1.TemporalCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1beta1.TemporalClusterSpec{},
			},
			objects: []client.Object{
				&appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-frontend",
						Namespace: "default",
					},
					Status: appsv1.DeploymentStatus{
						ObservedGeneration: 1,
						UpdatedReplicas:    1,
						ReadyReplicas:      0,
						AvailableReplicas:  0,
						Replicas:           1,
					},
				},
			},
			expected: []*v1beta1.ServiceStatus{
				{
					Name:    "frontend",
					Ready:   false,
					Version: "0.0.0",
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			statuses, err := status.ReconciledObjectsToServiceStatuses(test.cluster, test.objects)
			assert.NoError(tt, err)

			assert.EqualValues(tt, test.expected, statuses)
		})
	}
}
