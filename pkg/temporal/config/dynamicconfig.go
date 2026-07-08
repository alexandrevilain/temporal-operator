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
	"bytes"
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
	constraints := map[string]any{}

	if cv.Constraints.Namespace != "" {
		constraints["namespace"] = cv.Constraints.Namespace
	}

	if cv.Constraints.NamespaceID != "" {
		constraints["namespaceid"] = cv.Constraints.NamespaceID
	}

	if cv.Constraints.TaskQueueName != "" {
		constraints["taskqueuename"] = cv.Constraints.TaskQueueName
	}

	// TaskQueueType == tasktype
	// See: https://github.com/temporalio/temporal/blob/v1.19.1/common/dynamicconfig/file_based_client.go#L366
	if cv.Constraints.TaskQueueType != "" {
		constraints["tasktype"] = cv.Constraints.TaskQueueType
	}

	// TaskType == historytasktype
	// See: https://github.com/temporalio/temporal/blob/v1.19.1/common/dynamicconfig/file_based_client.go#L379
	if cv.Constraints.TaskType != "" {
		constraints["historytasktype"] = cv.Constraints.TaskType
	}

	if cv.Constraints.ShardID != 0 {
		constraints["shardid"] = cv.Constraints.ShardID
	}

	// Decode the raw JSON value using a decoder with UseNumber so that JSON
	// numbers are preserved as json.Number instead of being coerced to float64.
	// Without this, an integer like 2097152 becomes float64(2097152), which
	// yaml.v3 later marshals in scientific notation (2.097152e+06). Temporal's
	// file based dynamic config client then fails to parse it for settings that
	// expect an integer.
	decoder := json.NewDecoder(bytes.NewReader(cv.Value.Raw))
	decoder.UseNumber()

	var value any
	err := decoder.Decode(&value)
	if err != nil {
		return YamlConstrainedValue{}, err
	}

	return YamlConstrainedValue{
		Constraints: constraints,
		Value:       normalizeJSONNumbers(value),
	}, nil
}

// normalizeJSONNumbers recursively walks a value decoded from JSON with
// json.Decoder.UseNumber and converts every json.Number into a concrete int or
// float64. Integers are converted to int to match the type yaml.v3 produces
// when it unmarshals the config map back, keeping the reconciliation deep-equal
// comparison stable.
func normalizeJSONNumbers(value any) any {
	switch v := value.(type) {
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i)
		}
		if f, err := v.Float64(); err == nil {
			return f
		}
		return v.String()
	case map[string]any:
		for key, val := range v {
			v[key] = normalizeJSONNumbers(val)
		}
		return v
	case []any:
		for i, val := range v {
			v[i] = normalizeJSONNumbers(val)
		}
		return v
	default:
		return value
	}
}
