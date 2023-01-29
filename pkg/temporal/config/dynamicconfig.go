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

package config

import (
	"encoding/json"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
)

type YamlDynamicConfig map[string][]YamlConstrainedValue

type YamlConstrainedValue struct {
	Constraints map[string]any
	Value       any
}

func DynamicConfigToYamlDynamicConfig(dc *v1beta1.DynamicConfigSpec) (YamlDynamicConfig, error) {
	result := map[string][]YamlConstrainedValue{}

	for k, v := range dc.Values {
		yamlConstrainedValues := []YamlConstrainedValue{}
		for _, constrainedValue := range v {
			yamlConstrainedValue, err := constrainedValueToYamlConstrainedValue(&constrainedValue)
			if err != nil {
				return result, err
			}
			yamlConstrainedValues = append(yamlConstrainedValues, yamlConstrainedValue)
		}
		result[k] = yamlConstrainedValues
	}
	return result, nil
}

// constrainedValueToYamlConstrainedValue transform kubernetes CRD-style ConstrainedValue to temporal's YamlConstrainedValue.
// Key names are extracted from: https://github.com/temporalio/temporal/blob/v1.19.1/common/dynamicconfig/file_based_client.go#L344
func constrainedValueToYamlConstrainedValue(cv *v1beta1.ConstrainedValue) (YamlConstrainedValue, error) {
	constaints := map[string]any{}

	if cv.Constraints.Namespace != "" {
		constaints["namespace"] = cv.Constraints.Namespace
	}

	if cv.Constraints.NamespaceID != "" {
		constaints["namespaceid"] = cv.Constraints.NamespaceID
	}

	if cv.Constraints.TaskQueueName != "" {
		constaints["taskqueuename"] = cv.Constraints.TaskQueueName
	}

	// TaskQueueType == tasktype
	// See: https://github.com/temporalio/temporal/blob/v1.19.1/common/dynamicconfig/file_based_client.go#L366
	if cv.Constraints.TaskQueueType != "" {
		constaints["tasktype"] = cv.Constraints.TaskQueueType
	}

	// TaskType == historytasktype
	// See: https://github.com/temporalio/temporal/blob/v1.19.1/common/dynamicconfig/file_based_client.go#L379
	if cv.Constraints.TaskType != "" {
		constaints["historytasktype"] = cv.Constraints.TaskType
	}

	if cv.Constraints.ShardID != 0 {
		constaints["shardid"] = cv.Constraints.ShardID
	}

	var value any
	err := json.Unmarshal(cv.Value.Raw, &value)
	if err != nil {
		return YamlConstrainedValue{}, err
	}

	return YamlConstrainedValue{
		Constraints: constaints,
		Value:       value,
	}, nil
}
