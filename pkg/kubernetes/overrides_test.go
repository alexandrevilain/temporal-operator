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

package kubernetes_test

import (
	"testing"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apimachineryresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestApplyDeploymentOverrides(t *testing.T) {
	tests := map[string]struct {
		original *appsv1.Deployment
		override *v1beta1.DeploymentOverride
		expected *appsv1.Deployment
	}{
		"works with nil override": {
			original: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			override: nil,
			expected: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
		},
		"add labels": {
			original: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						"a": "b",
					},
				},
			},
			override: &v1beta1.DeploymentOverride{
				ObjectMetaOverride: &v1beta1.ObjectMetaOverride{
					Labels: map[string]string{
						"c": "d",
					},
				},
			},
			expected: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						"a": "b",
						"c": "d",
					},
				},
			},
		},
		"add annotations": {
			original: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"a": "b",
					},
				},
			},
			override: &v1beta1.DeploymentOverride{
				ObjectMetaOverride: &v1beta1.ObjectMetaOverride{
					Annotations: map[string]string{
						"c": "d",
					},
				},
			},
			expected: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"a": "b",
						"c": "d",
					},
				},
			},
		},
		"add pod resources": {
			original: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"a": "b",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "testcontainer",
								},
							},
						},
					},
				},
			},
			override: &v1beta1.DeploymentOverride{
				ObjectMetaOverride: &v1beta1.ObjectMetaOverride{
					Annotations: map[string]string{
						"c": "d",
					},
				},
				Spec: &v1beta1.DeploymentOverrideSpec{
					Template: &v1beta1.PodTemplateSpecOverride{
						Spec: &apiextensionsv1.JSON{
							Raw: []byte(`{"containers":[{"name":"testcontainer","resources":{"limits":{"cpu":"100m"}}}]}`),
						},
					},
				},
			},
			expected: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"a": "b",
						"c": "d",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "testcontainer",
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU: apimachineryresource.MustParse("100m"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"add sidecar": {
			original: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"a": "b",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "testcontainer",
								},
							},
						},
					},
				},
			},
			override: &v1beta1.DeploymentOverride{
				Spec: &v1beta1.DeploymentOverrideSpec{
					Template: &v1beta1.PodTemplateSpecOverride{
						Spec: &apiextensionsv1.JSON{
							Raw: []byte(`{"containers":[{"name":"my-sidecar"}]}`),
						},
					},
				},
			},
			expected: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"a": "b",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "my-sidecar",
								},
								{
									Name: "testcontainer",
								},
							},
						},
					},
				},
			},
		},
		"add init container": {
			original: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"a": "b",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "testcontainer",
								},
							},
						},
					},
				},
			},
			override: &v1beta1.DeploymentOverride{
				Spec: &v1beta1.DeploymentOverrideSpec{
					Template: &v1beta1.PodTemplateSpecOverride{
						Spec: &apiextensionsv1.JSON{
							Raw: []byte(`{"initContainers":[{"name":"my-init","resources":{"limits":{"cpu":"50m"}}}]}`),
						},
					},
				},
			},
			expected: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"a": "b",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Name: "my-init",
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU: apimachineryresource.MustParse("50m"),
										},
									},
								},
							},
							Containers: []corev1.Container{
								{
									Name: "testcontainer",
								},
							},
						},
					},
				},
			},
		},
		"replace liveness probe": {
			original: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"a": "b",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "test",
									LivenessProbe: &corev1.Probe{
										ProbeHandler: corev1.ProbeHandler{
											TCPSocket: &corev1.TCPSocketAction{
												Port: intstr.FromString("rpc"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			override: &v1beta1.DeploymentOverride{
				Spec: &v1beta1.DeploymentOverrideSpec{
					Template: &v1beta1.PodTemplateSpecOverride{
						Spec: &apiextensionsv1.JSON{
							Raw: []byte(`{"containers":[{"name":"test","livenessProbe":{"$patch":"replace","tcpSocket":null,"exec":{"command":["echo","hi"]}}}]}`),
						},
					},
				},
			},
			expected: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"a": "b",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "test",
									LivenessProbe: &corev1.Probe{
										ProbeHandler: corev1.ProbeHandler{
											Exec: &corev1.ExecAction{
												Command: []string{"echo", "hi"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"add env var to existing env": {
			original: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "test",
									Env: []corev1.EnvVar{
										{
											Name: "a",
											ValueFrom: &corev1.EnvVarSource{
												ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: "test",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			override: &v1beta1.DeploymentOverride{
				JSONPatch: &apiextensionsv1.JSON{
					Raw: []byte(`[{"op":"add", "path":"/spec/template/spec/containers/0/env/-", "value":{"name":"b","value":"c"}}]`),
				},
			},
			expected: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "test",
									Env: []corev1.EnvVar{
										{
											Name: "a",
											ValueFrom: &corev1.EnvVarSource{
												ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: "test",
													},
												},
											},
										},
										{
											Name:  "b",
											Value: "c",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"add secret volume to existing volumes": {
			original: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "test",
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "a",
											ReadOnly:  true,
											MountPath: "/a",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "a",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			override: &v1beta1.DeploymentOverride{
				JSONPatch: &apiextensionsv1.JSON{
					Raw: []byte(`[
						{"op": "add", "path": "/spec/template/spec/containers/0/volumeMounts/-", "value": {"name": "b", "readOnly": true, "mountPath": "/b"}}, 
						{"op": "add", "path": "/spec/template/spec/volumes/-", "value": {"name": "b", "secret": {"secretName": "test"}}}
					]`),
				},
			},
			expected: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "test",
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "a",
											ReadOnly:  true,
											MountPath: "/a",
										},
										{
											Name:      "b",
											ReadOnly:  true,
											MountPath: "/b",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "a",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "test",
											},
										},
									},
								},
								{
									Name: "b",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: "test",
										},
									},
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
			err := kubernetes.ApplyDeploymentOverrides(test.original, test.override)
			require.NoError(tt, err)
			if !equality.Semantic.DeepEqual(test.original, test.expected) {
				tt.Logf("expected: %+v", test.expected)
				tt.Logf("actual: %+v", test.original)
				assert.True(tt, false)
			}
		})
	}
}

func TestApplyServiceOverrides(t *testing.T) {
	tests := map[string]struct {
		original *corev1.Service
		override *v1beta1.ObjectMetaOverride
		expected *corev1.Service
	}{
		"works with nil override": {
			original: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"a": "b",
					},
					Labels: map[string]string{
						"c": "d",
					},
				},
			},
			override: nil,
			expected: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"a": "b",
					},
					Labels: map[string]string{
						"c": "d",
					},
				},
			},
		},
		"add both": {
			original: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"a": "b",
					},
					Labels: map[string]string{
						"c": "d",
					},
				},
			},
			override: &v1beta1.ObjectMetaOverride{
				Annotations: map[string]string{
					"1": "2",
				},
				Labels: map[string]string{
					"3": "4",
				},
			},
			expected: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"a": "b",
						"1": "2",
					},
					Labels: map[string]string{
						"c": "d",
						"3": "4",
					},
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			err := kubernetes.ApplyServiceOverrides(test.original, test.override)
			require.NoError(tt, err)

			assert.True(tt, equality.Semantic.DeepEqual(test.original, test.expected))
		})
	}
}

func TestApplyPodTemplateSpecOverrides(t *testing.T) {
	tests := map[string]struct {
		original *corev1.PodTemplateSpec
		override *v1beta1.PodTemplateSpecOverride
		expected *corev1.PodTemplateSpec
	}{
		"works with nil overrides": {
			original: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			override: nil,
			expected: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
		},
		"add labels": {
			original: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						"a": "b",
					},
				},
			},
			override: &v1beta1.PodTemplateSpecOverride{
				ObjectMetaOverride: &v1beta1.ObjectMetaOverride{
					Labels: map[string]string{
						"c": "d",
					},
				},
			},
			expected: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						"a": "b",
						"c": "d",
					},
				},
			},
		},
		"add annotations": {
			original: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"a": "b",
					},
				},
			},
			override: &v1beta1.PodTemplateSpecOverride{
				ObjectMetaOverride: &v1beta1.ObjectMetaOverride{
					Annotations: map[string]string{
						"c": "d",
					},
				},
			},
			expected: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"a": "b",
						"c": "d",
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			err := kubernetes.ApplyPodTemplateSpecOverrides(test.original, test.override)
			require.NoError(tt, err)
			assert.True(tt, equality.Semantic.DeepEqual(test.original, test.expected))
		})
	}
}

func TestPatchPodSpecWithOverride(t *testing.T) {
	tests := map[string]struct {
		original *corev1.PodSpec
		override *corev1.PodSpec
		expected *corev1.PodSpec
	}{
		"works with nil overrides": {
			original: &corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "test",
					},
				},
			},
			override: nil,
			expected: &corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "test",
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			_, err := kubernetes.PatchPodSpecWithOverride(test.original, test.override)
			require.NoError(tt, err)
			assert.True(tt, equality.Semantic.DeepEqual(test.original, test.expected))
		})
	}
}
