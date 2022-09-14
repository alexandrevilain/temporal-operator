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

package persistence

import (
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultPasswordSecretKey = "password"
)

// GetDatastoresEnvironmentVariables returns needed env vars for the provided datastores list.
func GetDatastoresEnvironmentVariables(datastores []v1alpha1.TemporalDatastoreSpec) []corev1.EnvVar {
	vars := []corev1.EnvVar{}
	for _, datastore := range datastores {
		key := datastore.PasswordSecretRef.Key
		if key == "" {
			key = defaultPasswordSecretKey
		}
		vars = append(vars,
			corev1.EnvVar{
				Name: datastore.GetPasswordEnvVarName(),
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: datastore.PasswordSecretRef.Name,
						},
						Key: key,
					},
				},
			},
		)
	}
	return vars
}

// GetDatastoresVolumes returns needed volume for the provided datastores list.
func GetDatastoresVolumes(datastores []v1alpha1.TemporalDatastoreSpec) []corev1.Volume {
	volumes := []corev1.Volume{}
	for _, datastore := range datastores {
		if datastore.TLS != nil && datastore.TLS.Enabled {
			if datastore.TLS.CaFileRef != nil {
				key := datastore.TLS.CaFileRef.Key
				if key == "" {
					key = v1alpha1.DataStoreClientTLSCaFileName
				}
				volumes = append(volumes,
					corev1.Volume{
						Name: fmt.Sprintf("%s-tls-ca-file", datastore.Name),
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: datastore.TLS.CaFileRef.Name,
								Items: []corev1.KeyToPath{
									{
										Key:  key,
										Path: datastore.GetTLSCaFileMountPath(),
									},
								},
							},
						},
					},
				)
			}
			if datastore.TLS.CertFileRef != nil {
				key := datastore.TLS.CertFileRef.Key
				if key == "" {
					key = v1alpha1.DataStoreClientTLSCertFileName
				}
				volumes = append(volumes,
					corev1.Volume{
						Name: fmt.Sprintf("%s-tls-cert-file", datastore.Name),
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: datastore.TLS.CertFileRef.Name,
								Items: []corev1.KeyToPath{
									{
										Key:  key,
										Path: datastore.GetTLSCertFileMountPath(),
									},
								},
							},
						},
					},
				)
			}

			if datastore.TLS.KeyFileRef != nil {
				key := datastore.TLS.KeyFileRef.Key
				if key == "" {
					key = v1alpha1.DataStoreClientTLSKeyFileName
				}
				volumes = append(volumes,
					corev1.Volume{
						Name: fmt.Sprintf("%s-tls-key-file", datastore.Name),
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: datastore.TLS.KeyFileRef.Name,
								Items: []corev1.KeyToPath{
									{
										Key:  key,
										Path: datastore.GetTLSKeyFileMountPath(),
									},
								},
							},
						},
					},
				)
			}
		}
	}

	return volumes
}

// GetDatastoresVolumeMounts returns needed volume mounts for the provided datastores list.
func GetDatastoresVolumeMounts(datastores []v1alpha1.TemporalDatastoreSpec) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{}
	for _, datastore := range datastores {
		if datastore.TLS != nil && datastore.TLS.Enabled {
			if datastore.TLS.CaFileRef != nil {
				volumeMounts = append(volumeMounts, corev1.VolumeMount{
					Name:      fmt.Sprintf("%s-tls-ca-file", datastore.Name),
					MountPath: datastore.GetTLSCaFileMountPath(),
				})
			}
			if datastore.TLS.CertFileRef != nil {
				volumeMounts = append(volumeMounts, corev1.VolumeMount{
					Name:      fmt.Sprintf("%s-tls-cert-file", datastore.Name),
					MountPath: datastore.GetTLSCertFileMountPath(),
				})
			}

			if datastore.TLS.KeyFileRef != nil {
				volumeMounts = append(volumeMounts, corev1.VolumeMount{
					Name:      fmt.Sprintf("%s-tls-key-file", datastore.Name),
					MountPath: datastore.GetTLSKeyFileMountPath(),
				})
			}
		}
	}
	return volumeMounts
}
