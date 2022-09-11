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

package kubernetes

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kstatus "sigs.k8s.io/cli-utils/pkg/kstatus/status"
)

// IsDeploymentReady returns whenever the provided deployment is ready.
func IsDeploymentReady(deploy *appsv1.Deployment) (bool, error) {
	udeploy, err := runtime.DefaultUnstructuredConverter.ToUnstructured(deploy)
	if err != nil {
		return false, err
	}

	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(udeploy)

	result, err := kstatus.Compute(u)
	if err != nil {
		return false, err
	}

	return result.Status == kstatus.CurrentStatus, nil
}
