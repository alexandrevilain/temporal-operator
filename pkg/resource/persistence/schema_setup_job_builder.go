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
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/istio"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/linkerd"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type SchemaJobBuilder struct {
	instance *v1alpha1.TemporalCluster
	scheme   *runtime.Scheme
	// name is the name of the job
	name string
	// action is the job action name label
	action string
	// command is the command the job should run
	command []string
}

func NewSchemaJobBuilder(instance *v1alpha1.TemporalCluster, scheme *runtime.Scheme, name, action string, command []string) *SchemaJobBuilder {
	return &SchemaJobBuilder{
		instance: instance,
		scheme:   scheme,
		name:     name,
		action:   action,
		command:  command,
	}
}

func (b *SchemaJobBuilder) GetLabels() map[string]string {
	return metadata.Merge(
		metadata.GetLabels(b.instance.Name, b.name, b.instance.Spec.Version, b.instance.Labels),
		map[string]string{
			"operator.temporal.io/action": b.action,
		},
	)
}

func (b *SchemaJobBuilder) Build() (client.Object, error) {
	envVars := []corev1.EnvVar{
		{
			Name:  "TEMPORAL_CLI_ADDRESS",
			Value: fmt.Sprintf("%s:%d", b.instance.ChildResourceName("frontend"), *b.instance.Spec.Services.Frontend.Port),
		},
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "scripts",
			MountPath: "/etc/scripts",
		},
	}

	volumes := []corev1.Volume{
		{
			Name: "scripts",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: b.instance.ChildResourceName("schema-scripts"),
					},
					DefaultMode: pointer.Int32(0777),
				},
			},
		},
	}

	for _, datastore := range b.instance.Spec.Datastores {
		envVars = append(envVars,
			corev1.EnvVar{
				Name: datastore.GetPasswordEnvVarName(),
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: datastore.PasswordSecretRef.Name,
						},
						Key: datastore.PasswordSecretRef.Key,
					},
				},
			},
		)
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

	for _, datastore := range b.instance.Spec.Datastores {
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
								SecretName: datastore.TLS.CaFileRef.Name,
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
								SecretName: datastore.TLS.CaFileRef.Name,
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

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName(b.name),
			Namespace:   b.instance.Namespace,
			Labels:      b.GetLabels(),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: metadata.Merge(
						istio.GetLabels(b.instance),
						metadata.GetLabels(b.instance.Name, b.name, b.instance.Spec.Version, b.instance.Labels),
					),
					Annotations: metadata.Merge(
						linkerd.GetAnnotations(b.instance),
						istio.GetAnnotations(b.instance),
						metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
					),
				},
				Spec: corev1.PodSpec{
					RestartPolicy:    corev1.RestartPolicyOnFailure,
					ImagePullSecrets: b.instance.Spec.ImagePullSecrets,
					Containers: []corev1.Container{
						{
							Name:                     "schema-script-runner",
							Image:                    fmt.Sprintf("%s:%s", b.instance.Spec.AdminTools.Image, b.instance.Spec.Version),
							ImagePullPolicy:          corev1.PullAlways,
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: corev1.TerminationMessageReadFile,
							Command:                  append([]string{"/bin/sh", "-c"}, b.command...),
							Env:                      envVars,
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: pointer.Bool(false),
							},
							VolumeMounts: volumeMounts,
						},
					},
					TerminationGracePeriodSeconds: pointer.Int64(30),
					DNSPolicy:                     corev1.DNSClusterFirst,
					SecurityContext:               &corev1.PodSecurityContext{},
					SchedulerName:                 "default-scheduler",
					Volumes:                       volumes,
				},
			},
		},
	}, nil
}

func (b *SchemaJobBuilder) Update(object client.Object) error {
	job := object.(*batchv1.Job)
	if err := controllerutil.SetControllerReference(b.instance, job, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}
	return nil
}
