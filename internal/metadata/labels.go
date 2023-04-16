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

package metadata

import (
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OwnerObject interface {
	client.Object
	SelectorLabels() map[string]string
}

// LabelsSelector returns service's default labels.
func LabelsSelector(owner OwnerObject, serviceName string) map[string]string {
	return Merge(
		owner.SelectorLabels(),
		map[string]string{
			"app.kubernetes.io/component": serviceName,
		},
	)
}

// GetLabels returns a Labels for a temporal service.
func GetLabels(owner OwnerObject, service string, version *version.Version, labels map[string]string) map[string]string {
	l := LabelsSelector(owner, service)
	l["app.kubernetes.io/version"] = version.String()
	for k, v := range labels {
		l[k] = v
	}
	return l
}

// GetLabels returns a Labels for a temporal service using string Version.
func GetVersionStringLabels(owner OwnerObject, service string, version string, labels map[string]string) map[string]string {
	l := LabelsSelector(owner, service)
	l["app.kubernetes.io/version"] = version
	for k, v := range labels {
		l[k] = v
	}
	return l
}

// HeadlessLabels returns labels to express that a service is headless.
func HeadlessLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/headless": "true",
	}
}
