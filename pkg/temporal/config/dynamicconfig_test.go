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
	"gopkg.in/yaml.v3"
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
						Value:       int(5),
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
						Value: int(5),
					},
				},
			},
		},
		"combined constraints": {
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
						Value: int(5),
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
						Value: int(5),
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
						Value: int(5),
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

// TestDynamicConfigToYamlDynamicConfigLargeInteger ensures that large integer
// values are marshaled as plain integers and not in floating-point scientific
// notation (e.g. 2.097152e+06). Temporal's file based dynamic config client
// parses the resulting YAML and fails to load a setting when the number is
// rendered as a float for a setting that expects an integer.
func TestDynamicConfigToYamlDynamicConfigLargeInteger(t *testing.T) {
	dc := &v1beta1.DynamicConfigSpec{
		Values: map[string][]v1beta1.ConstrainedValue{
			"limit.blobSize.error": {
				{
					Value: &apiextensionsv1.JSON{Raw: []byte(`2097152`)},
				},
			},
		},
	}

	result, err := config.DynamicConfigToYamlDynamicConfig(dc)
	require.NoError(t, err)

	out, err := yaml.Marshal(result)
	require.NoError(t, err)

	assert.Contains(t, string(out), "value: 2097152")
	assert.NotContains(t, string(out), "2.097152e+06")
}

// TestDynamicConfigToYamlDynamicConfigNestedLargeInteger ensures large integers
// nested inside object/array values are also rendered as plain integers, since
// json.Unmarshal into an any coerces every number (including nested ones) to
// float64.
func TestDynamicConfigToYamlDynamicConfigNestedLargeInteger(t *testing.T) {
	dc := &v1beta1.DynamicConfigSpec{
		Values: map[string][]v1beta1.ConstrainedValue{
			"history.defaultActivityRetryPolicy": {
				{
					Value: &apiextensionsv1.JSON{Raw: []byte(`{"MaximumInterval": 2097152, "Sizes": [1048576, 4194304]}`)},
				},
			},
		},
	}

	result, err := config.DynamicConfigToYamlDynamicConfig(dc)
	require.NoError(t, err)

	out, err := yaml.Marshal(result)
	require.NoError(t, err)

	assert.Contains(t, string(out), "MaximumInterval: 2097152")
	assert.Contains(t, string(out), "- 1048576")
	assert.Contains(t, string(out), "- 4194304")
	assert.NotContains(t, string(out), "e+06")
}
