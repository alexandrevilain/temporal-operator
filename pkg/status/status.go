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

package status

import "github.com/alexandrevilain/temporal-operator/api/v1beta1"

// ObservedVersionMatchesDesiredVersion returns true if all services status
// versions are matching the desired cluster version.
func ObservedVersionMatchesDesiredVersion(c *v1beta1.TemporalCluster) bool {
	if len(c.Status.Services) == 0 {
		return false
	}
	for _, serviceStatus := range c.Status.Services {
		if serviceStatus.Version != c.Spec.Version.String() {
			return false
		}
	}
	return true
}

// IsClusterReady returns true if all services status are in ready state.
func IsClusterReady(c *v1beta1.TemporalCluster) bool {
	if len(c.Status.Services) == 0 {
		return false
	}
	for _, serviceStatus := range c.Status.Services {
		if !serviceStatus.Ready || serviceStatus.Version != c.Spec.Version.String() {
			return false
		}
	}
	return true
}
