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

	appsv1alpha1 "github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/pkg/status"
	"github.com/stretchr/testify/assert"
)

func TestObservedVersionMatchesDesiredVersion(t *testing.T) {
	tests := map[string]struct {
		cluster  *appsv1alpha1.TemporalCluster
		expected bool
	}{
		"all services matches the desired version": {
			cluster: &appsv1alpha1.TemporalCluster{
				Spec: appsv1alpha1.TemporalClusterSpec{
					Version: "1.16.0",
				},
				Status: appsv1alpha1.TemporalClusterStatus{
					Services: []appsv1alpha1.ServiceStatus{
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
			cluster: &appsv1alpha1.TemporalCluster{
				Spec: appsv1alpha1.TemporalClusterSpec{
					Version: "1.16.0",
				},
				Status: appsv1alpha1.TemporalClusterStatus{
					Services: []appsv1alpha1.ServiceStatus{
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
			cluster: &appsv1alpha1.TemporalCluster{
				Spec: appsv1alpha1.TemporalClusterSpec{
					Version: "1.16.0",
				},
				Status: appsv1alpha1.TemporalClusterStatus{},
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
		cluster  *appsv1alpha1.TemporalCluster
		expected bool
	}{
		"all services are ready": {
			cluster: &appsv1alpha1.TemporalCluster{
				Status: appsv1alpha1.TemporalClusterStatus{
					Services: []appsv1alpha1.ServiceStatus{
						{
							Name:  "test",
							Ready: true,
						},
						{
							Name:  "test2",
							Ready: true,
						},
					},
				},
			},
			expected: true,
		},
		"one service not ready": {
			cluster: &appsv1alpha1.TemporalCluster{
				Status: appsv1alpha1.TemporalClusterStatus{
					Services: []appsv1alpha1.ServiceStatus{
						{
							Name:  "test",
							Ready: true,
						},
						{
							Name:  "test2",
							Ready: false,
						},
					},
				},
			},
			expected: false,
		},
		"empty status": {
			cluster: &appsv1alpha1.TemporalCluster{
				Status: appsv1alpha1.TemporalClusterStatus{},
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
