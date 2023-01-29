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

package config_test

import (
	"testing"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/temporal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestDynamicConfigToYamlDynamicConfig(t *testing.T) {
	tests := map[string]struct {
		dyanmicConfig             *v1beta1.DynamicConfigSpec
		expectedYamlDynamicConfig config.YamlDynamicConfig
	}{
		"basic": {
			dyanmicConfig: &v1beta1.DynamicConfigSpec{
				Values: map[string][]v1beta1.ConstrainedValue{
					"matching.numTaskqueueReadPartitions": {
						{
							Value: &apiextensionsv1.JSON{Raw: []byte(`5`)},
						},
					},
				},
			},
			expectedYamlDynamicConfig: config.YamlDynamicConfig{
				"matching.numTaskqueueReadPartitions": {
					{
						Constraints: map[string]any{},
						Value:       float64(5),
					},
				},
			},
		},
		"namespace constaint": {
			dyanmicConfig: &v1beta1.DynamicConfigSpec{
				Values: map[string][]v1beta1.ConstrainedValue{
					"matching.numTaskqueueReadPartitions": {
						{
							Constraints: v1beta1.Constraints{
								Namespace: "accounting",
							},
							Value: &apiextensionsv1.JSON{Raw: []byte(`5`)},
						},
					},
				},
			},
			expectedYamlDynamicConfig: config.YamlDynamicConfig{
				"matching.numTaskqueueReadPartitions": {
					{
						Constraints: map[string]any{
							"namespace": "accounting",
						},
						Value: float64(5),
					},
				},
			},
		},
		"combined constaints": {
			dyanmicConfig: &v1beta1.DynamicConfigSpec{
				Values: map[string][]v1beta1.ConstrainedValue{
					"matching.numTaskqueueReadPartitions": {
						{
							Constraints: v1beta1.Constraints{
								NamespaceID:   "1234",
								TaskQueueName: "accounting-tq",
								ShardID:       int32(1),
							},
							Value: &apiextensionsv1.JSON{Raw: []byte(`5`)},
						},
					},
				},
			},
			expectedYamlDynamicConfig: config.YamlDynamicConfig{
				"matching.numTaskqueueReadPartitions": {
					{
						Constraints: map[string]any{
							"namespaceid":   "1234",
							"taskqueuename": "accounting-tq",
							"shardid":       int32(1),
						},
						Value: float64(5),
					},
				},
			},
		},
		"TaskQueueType constaint creates tasktype constaint": {
			dyanmicConfig: &v1beta1.DynamicConfigSpec{
				Values: map[string][]v1beta1.ConstrainedValue{
					"matching.numTaskqueueReadPartitions": {
						{
							Constraints: v1beta1.Constraints{
								TaskQueueType: "Workflow",
							},
							Value: &apiextensionsv1.JSON{Raw: []byte(`5`)},
						},
					},
				},
			},
			expectedYamlDynamicConfig: config.YamlDynamicConfig{
				"matching.numTaskqueueReadPartitions": {
					{
						Constraints: map[string]any{
							"tasktype": "Workflow",
						},
						Value: float64(5),
					},
				},
			},
		},
		"TaskType constaint creates tasktype historytasktype": {
			dyanmicConfig: &v1beta1.DynamicConfigSpec{
				Values: map[string][]v1beta1.ConstrainedValue{
					"matching.numTaskqueueReadPartitions": {
						{
							Constraints: v1beta1.Constraints{
								TaskType: "ActivityRetryTimer",
							},
							Value: &apiextensionsv1.JSON{Raw: []byte(`5`)},
						},
					},
				},
			},
			expectedYamlDynamicConfig: config.YamlDynamicConfig{
				"matching.numTaskqueueReadPartitions": {
					{
						Constraints: map[string]any{
							"historytasktype": "ActivityRetryTimer",
						},
						Value: float64(5),
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			result, err := config.DynamicConfigToYamlDynamicConfig(test.dyanmicConfig)
			require.NoError(tt, err)
			assert.EqualValues(tt, test.expectedYamlDynamicConfig, result)
		})
	}
}
