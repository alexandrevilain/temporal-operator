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

package persistence_test

import (
	"testing"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/resource/persistence"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestGetDatastoresEnvironmentVariables(t *testing.T) {
	tests := map[string]struct {
		datastores      []*v1beta1.DatastoreSpec
		expectedEnvVars []corev1.EnvVar
	}{
		"empty datastore list": {
			datastores:      []*v1beta1.DatastoreSpec{},
			expectedEnvVars: []corev1.EnvVar{},
		},
		"one datastore without secret key defined": {
			datastores: []*v1beta1.DatastoreSpec{
				{
					Name: "test",
					PasswordSecretRef: &v1beta1.SecretKeyReference{
						Name: "testSecret",
					},
				},
			},
			expectedEnvVars: []corev1.EnvVar{
				{
					Name: "TEMPORAL_TEST_DATASTORE_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "testSecret",
							},
							Key: "password",
						},
					},
				},
			},
		},
		"one datastore with secret key defined": {
			datastores: []*v1beta1.DatastoreSpec{
				{
					Name: "test",
					PasswordSecretRef: &v1beta1.SecretKeyReference{
						Name: "testSecret",
						Key:  "my-password",
					},
				},
			},
			expectedEnvVars: []corev1.EnvVar{
				{
					Name: "TEMPORAL_TEST_DATASTORE_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "testSecret",
							},
							Key: "my-password",
						},
					},
				},
			},
		},
		"two datastores": {
			datastores: []*v1beta1.DatastoreSpec{
				{
					Name: "test",
					PasswordSecretRef: &v1beta1.SecretKeyReference{
						Name: "testSecret",
						Key:  "password",
					},
				},
				{
					Name: "test-test",
					PasswordSecretRef: &v1beta1.SecretKeyReference{
						Name: "testSecret2",
						Key:  "my-password",
					},
				},
			},
			expectedEnvVars: []corev1.EnvVar{
				{
					Name: "TEMPORAL_TEST_DATASTORE_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "testSecret",
							},
							Key: "password",
						},
					},
				},
				{
					Name: "TEMPORAL_TEST-TEST_DATASTORE_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "testSecret2",
							},
							Key: "my-password",
						},
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			result := persistence.GetDatastoresEnvironmentVariables(test.datastores)
			assert.EqualValues(tt, test.expectedEnvVars, result)
		})
	}
}

func TestGetDatastoresVolumes(t *testing.T) {
	tests := map[string]struct {
		datastores      []*v1beta1.DatastoreSpec
		expectedEnvVars []corev1.Volume
	}{
		"empty datastore list": {
			datastores:      []*v1beta1.DatastoreSpec{},
			expectedEnvVars: []corev1.Volume{},
		},
		"datastore list without TLS": {
			datastores: []*v1beta1.DatastoreSpec{
				{
					Name: "test",
				},
				{
					Name: "test2",
				},
			},
			expectedEnvVars: []corev1.Volume{},
		},
		"datastore list with ca file reference filed": {
			datastores: []*v1beta1.DatastoreSpec{
				{
					Name: "test",
					TLS: &v1beta1.DatastoreTLSSpec{
						Enabled: true,
						CaFileRef: &v1beta1.SecretKeyReference{
							Name: "secret",
						},
					},
				},
				{
					Name: "test2",
					TLS: &v1beta1.DatastoreTLSSpec{
						Enabled: true,
						CaFileRef: &v1beta1.SecretKeyReference{
							Name: "secret2",
							Key:  "my-custom-key",
						},
					},
				},
			},
			expectedEnvVars: []corev1.Volume{
				{
					Name: "test-tls-ca-file",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "secret",
							Items: []corev1.KeyToPath{
								{
									Key:  "ca.pem",
									Path: v1beta1.DataStoreClientTLSCaFileName,
								},
							},
						},
					},
				},
				{
					Name: "test2-tls-ca-file",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "secret2",
							Items: []corev1.KeyToPath{
								{
									Key:  "my-custom-key",
									Path: v1beta1.DataStoreClientTLSCaFileName,
								},
							},
						},
					},
				},
			},
		},
		"datastore list with cert file reference filed": {
			datastores: []*v1beta1.DatastoreSpec{
				{
					Name: "test",
					TLS: &v1beta1.DatastoreTLSSpec{
						Enabled: true,
						CertFileRef: &v1beta1.SecretKeyReference{
							Name: "secret",
						},
					},
				},
				{
					Name: "test2",
					TLS: &v1beta1.DatastoreTLSSpec{
						Enabled: true,
						CertFileRef: &v1beta1.SecretKeyReference{
							Name: "secret2",
							Key:  "my-custom-key",
						},
					},
				},
			},
			expectedEnvVars: []corev1.Volume{
				{
					Name: "test-tls-cert-file",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "secret",
							Items: []corev1.KeyToPath{
								{
									Key:  "client.pem",
									Path: v1beta1.DataStoreClientTLSCertFileName,
								},
							},
						},
					},
				},
				{
					Name: "test2-tls-cert-file",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "secret2",
							Items: []corev1.KeyToPath{
								{
									Key:  "my-custom-key",
									Path: v1beta1.DataStoreClientTLSCertFileName,
								},
							},
						},
					},
				},
			},
		},
		"datastore list with key file reference filed": {
			datastores: []*v1beta1.DatastoreSpec{
				{
					Name: "test",
					TLS: &v1beta1.DatastoreTLSSpec{
						Enabled: true,
						KeyFileRef: &v1beta1.SecretKeyReference{
							Name: "secret",
						},
					},
				},
				{
					Name: "test2",
					TLS: &v1beta1.DatastoreTLSSpec{
						Enabled: true,
						KeyFileRef: &v1beta1.SecretKeyReference{
							Name: "secret2",
							Key:  "my-custom-key",
						},
					},
				},
			},
			expectedEnvVars: []corev1.Volume{
				{
					Name: "test-tls-key-file",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "secret",
							Items: []corev1.KeyToPath{
								{
									Key:  "client.key",
									Path: v1beta1.DataStoreClientTLSKeyFileName,
								},
							},
						},
					},
				},
				{
					Name: "test2-tls-key-file",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "secret2",
							Items: []corev1.KeyToPath{
								{
									Key:  "my-custom-key",
									Path: v1beta1.DataStoreClientTLSKeyFileName,
								},
							},
						},
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			result := persistence.GetDatastoresVolumes(test.datastores)
			assert.EqualValues(tt, test.expectedEnvVars, result)
		})
	}
}

func TestGetDatastoresVolumeMounts(t *testing.T) {
	tests := map[string]struct {
		datastores      []*v1beta1.DatastoreSpec
		expectedEnvVars []corev1.VolumeMount
	}{
		"empty datastore list": {
			datastores:      []*v1beta1.DatastoreSpec{},
			expectedEnvVars: []corev1.VolumeMount{},
		},
		"datastore list without TLS": {
			datastores: []*v1beta1.DatastoreSpec{
				{
					Name: "test",
				},
				{
					Name: "test2",
				},
			},
			expectedEnvVars: []corev1.VolumeMount{},
		},
		"datastore list with ca file reference filed": {
			datastores: []*v1beta1.DatastoreSpec{
				{
					Name: "test",
					TLS: &v1beta1.DatastoreTLSSpec{
						Enabled: true,
						CaFileRef: &v1beta1.SecretKeyReference{
							Name: "secret",
						},
					},
				},
			},
			expectedEnvVars: []corev1.VolumeMount{
				{
					Name:      "test-tls-ca-file",
					MountPath: "/etc/tls/datastores/ca/test",
				},
			},
		},
		"datastore list with cert file reference filed": {
			datastores: []*v1beta1.DatastoreSpec{
				{
					Name: "test",
					TLS: &v1beta1.DatastoreTLSSpec{
						Enabled: true,
						CertFileRef: &v1beta1.SecretKeyReference{
							Name: "secret",
						},
					},
				},
			},
			expectedEnvVars: []corev1.VolumeMount{
				{
					Name:      "test-tls-cert-file",
					MountPath: "/etc/tls/datastores/cert/test",
				},
			},
		},
		"datastore list with key file reference filed": {
			datastores: []*v1beta1.DatastoreSpec{
				{
					Name: "test",
					TLS: &v1beta1.DatastoreTLSSpec{
						Enabled: true,
						KeyFileRef: &v1beta1.SecretKeyReference{
							Name: "secret",
						},
					},
				},
			},
			expectedEnvVars: []corev1.VolumeMount{
				{
					Name:      "test-tls-key-file",
					MountPath: "/etc/tls/datastores/key/test",
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			result := persistence.GetDatastoresVolumeMounts(test.datastores)
			assert.EqualValues(tt, test.expectedEnvVars, result)
		})
	}
}
