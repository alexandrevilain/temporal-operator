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

package prometheus

import (
	"encoding/json"
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// PatchServiceMonitorSpecWithOverride patches the provided ServiceMonitor spec with the provided ServiceMonitor spec override.
func PatchServiceMonitorSpecWithOverride(spec *monitoringv1.ServiceMonitorSpec, override *monitoringv1.ServiceMonitorSpec) (*monitoringv1.ServiceMonitorSpec, error) {
	if override == nil {
		return nil, nil
	}

	orginalSpec, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("can't marshal service monitor spec: %w", err)
	}

	overrideSpec, err := json.Marshal(override)
	if err != nil {
		return nil, fmt.Errorf("can't marshal service monitor spec override: %w", err)
	}

	patchedJSON, err := strategicpatch.StrategicMergePatch(orginalSpec, overrideSpec, monitoringv1.ServiceMonitorSpec{})
	if err != nil {
		return nil, fmt.Errorf("can't patch service monitor spec: %w", err)
	}

	patchedSpec := &monitoringv1.ServiceMonitorSpec{}
	err = json.Unmarshal(patchedJSON, patchedSpec)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal patched service monitor spec: %w", err)
	}

	return patchedSpec, nil
}
