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

package kubernetes_test

import (
	"strings"
	"testing"

	"github.com/alexandrevilain/temporal-operator/pkg/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
)

func TestIsDeploymentReady(t *testing.T) {
	tests := map[string]struct {
		deployment string
		expected   bool
	}{
		"deployment conditions does not contains Available type": {
			deployment: `apiVersion: apps/v1
kind: Deployment
metadata:
  generation: 1
  name: prod-frontend
  namespace: demo
status:
  availableReplicas: 1
  conditions:
  - lastTransitionTime: "2022-06-27T13:12:39Z"
    lastUpdateTime: "2022-06-27T13:12:49Z"
    message: ReplicaSet "prod-frontend-5548d8f78f" has successfully progressed.
    reason: NewReplicaSetAvailable
    status: "True"
    type: Progressing
  observedGeneration: 1
  readyReplicas: 1
  replicas: 1
  updatedReplicas: 1`,
			expected: false,
		},
		"deployment conditions contains Available type but status is false": {
			deployment: `apiVersion: apps/v1
kind: Deployment
metadata:
  generation: 1
  name: prod-frontend
  namespace: demo
status:
  availableReplicas: 1
  conditions:
  - lastTransitionTime: "2022-06-27T13:12:49Z"
    lastUpdateTime: "2022-06-27T13:12:49Z"
    message: Deployment has minimum availability.
    reason: MinimumReplicasAvailable
    status: "False"
    type: Available
  - lastTransitionTime: "2022-06-27T13:12:39Z"
    lastUpdateTime: "2022-06-27T13:12:49Z"
    message: ReplicaSet "prod-frontend-5548d8f78f" has successfully progressed.
    reason: NewReplicaSetAvailable
    status: "True"
    type: Progressing
  observedGeneration: 1
  readyReplicas: 1
  replicas: 1
  updatedReplicas: 1`,
			expected: false,
		},
		"deployment conditions contains Available type and status is true": {
			deployment: `apiVersion: apps/v1
kind: Deployment
metadata:
  generation: 1
  name: prod-frontend
  namespace: demo
status:
  availableReplicas: 1
  conditions:
  - lastTransitionTime: "2022-06-27T13:12:49Z"
    lastUpdateTime: "2022-06-27T13:12:49Z"
    message: Deployment has minimum availability.
    reason: MinimumReplicasAvailable
    status: "True"
    type: Available
  - lastTransitionTime: "2022-06-27T13:12:39Z"
    lastUpdateTime: "2022-06-27T13:12:49Z"
    message: ReplicaSet "prod-frontend-5548d8f78f" has successfully progressed.
    reason: NewReplicaSetAvailable
    status: "True"
    type: Progressing
  observedGeneration: 1
  readyReplicas: 1
  replicas: 1
  updatedReplicas: 1`,
			expected: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			obj := &appsv1.Deployment{}
			err := decoder.Decode(strings.NewReader(test.deployment), obj)
			require.NoError(tt, err)

			result := kubernetes.IsDeploymentReady(obj)
			assert.Equal(tt, test.expected, result)
		})
	}
}
