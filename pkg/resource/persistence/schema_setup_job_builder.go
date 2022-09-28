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

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/mtls/istio"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/mtls/linkerd"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type SchemaJobBuilder struct {
	instance *v1beta1.Cluster
	scheme   *runtime.Scheme
	// name is the name of the job
	name string
	// command is the command the job should run
	command []string
}

func NewSchemaJobBuilder(instance *v1beta1.Cluster, scheme *runtime.Scheme, name string, command []string) *SchemaJobBuilder {
	return &SchemaJobBuilder{
		instance: instance,
		scheme:   scheme,
		name:     name,
		command:  command,
	}
}

func (b *SchemaJobBuilder) Build() (client.Object, error) {
	datastores := b.instance.Spec.Persistence.GetDatastores()

	envVars := []corev1.EnvVar{
		{
			Name:  "TEMPORAL_CLI_ADDRESS",
			Value: fmt.Sprintf("%s:%d", b.instance.ChildResourceName("frontend"), *b.instance.Spec.Services.Frontend.Port),
		},
	}
	envVars = append(envVars, GetDatastoresEnvironmentVariables(datastores)...)

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "scripts",
			MountPath: "/etc/scripts",
		},
	}

	volumeMounts = append(volumeMounts, GetDatastoresVolumeMounts(datastores)...)

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

	volumes = append(volumes, GetDatastoresVolumes(datastores)...)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName(b.name),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance.Name, b.name, b.instance.Spec.Version, b.instance.Labels),
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
							TerminationMessagePath:   corev1.TerminationMessagePathDefault,
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
					SchedulerName:                 corev1.DefaultSchedulerName,
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
