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

package certmanager

import (
	"fmt"
	"path"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

func addPrefix(prefix, name string) string {
	return fmt.Sprintf("%s_%s", prefix, name)
}

// GetTLSEnvironmentVariables returns needed env vars for enabling TLS connection for temporal tools.
// To support the whole range of temporal tools, the caller should provide an envPrefix which prefixes all TLS env vars.
func GetTLSEnvironmentVariables(instance *v1beta1.Cluster, envPrefix, certsMountPath string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  addPrefix(envPrefix, "TLS_CA"),
			Value: path.Join(certsMountPath, TLSCA),
		},
		{
			Name:  addPrefix(envPrefix, "TLS_CERT"),
			Value: path.Join(certsMountPath, TLSCert),
		},
		{
			Name:  addPrefix(envPrefix, "TLS_KEY"),
			Value: path.Join(certsMountPath, TLSKey),
		},
		{
			Name:  addPrefix(envPrefix, "TLS_ENABLE_HOST_VERIFICATION"),
			Value: "true",
		},
		{
			Name:  addPrefix(envPrefix, "TLS_SERVER_NAME"),
			Value: instance.Spec.MTLS.Frontend.ServerName(instance.ServerName()),
		},
	}
}
