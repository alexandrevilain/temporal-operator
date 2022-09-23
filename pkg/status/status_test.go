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

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/status"
	"github.com/stretchr/testify/assert"
)

func TestObservedVersionMatchesDesiredVersion(t *testing.T) {
	tests := map[string]struct {
		cluster  *v1beta1.Cluster
		expected bool
	}{
		"all services matches the desired version": {
			cluster: &v1beta1.Cluster{
				Spec: v1beta1.ClusterSpec{
					Version: "1.16.0",
				},
				Status: v1beta1.ClusterStatus{
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
			cluster: &v1beta1.Cluster{
				Spec: v1beta1.ClusterSpec{
					Version: "1.16.0",
				},
				Status: v1beta1.ClusterStatus{
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
			cluster: &v1beta1.Cluster{
				Spec: v1beta1.ClusterSpec{
					Version: "1.16.0",
				},
				Status: v1beta1.ClusterStatus{},
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
		cluster  *v1beta1.Cluster
		expected bool
	}{
		"all services are ready": {
			cluster: &v1beta1.Cluster{
				Status: v1beta1.ClusterStatus{
					Services: []v1beta1.ServiceStatus{
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
			cluster: &v1beta1.Cluster{
				Status: v1beta1.ClusterStatus{
					Services: []v1beta1.ServiceStatus{
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
			cluster: &v1beta1.Cluster{
				Status: v1beta1.ClusterStatus{},
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
