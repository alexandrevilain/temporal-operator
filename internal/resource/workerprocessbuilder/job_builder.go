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

package workerprocessbuilder

import (
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type JobBuilder struct {
	instance *v1beta1.TemporalWorkerProcess
	scheme   *runtime.Scheme
	// name is the name of the job
	name string
	// command is the command the job should run
	command []string
}

func NewJobBuilder(instance *v1beta1.TemporalWorkerProcess, scheme *runtime.Scheme, name string, command []string) *JobBuilder {
	return &JobBuilder{
		instance: instance,
		scheme:   scheme,
		name:     name,
		command:  command,
	}
}

func (b *JobBuilder) Build() client.Object {
	envVars := []corev1.EnvVar{
		{
			Name: b.instance.Spec.Builder.GetBuildRepoPasswordEnvVarName(),
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: b.instance.Spec.Builder.BuildRegistry.PasswordSecretRef.Name,
					},
					Key: b.instance.Spec.Builder.BuildRegistry.PasswordSecretRef.Key,
				},
			},
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
						Name: b.instance.ChildResourceName("builder-scripts"),
					},
					DefaultMode: ptr.To[int32](0o777),
				},
			},
		},
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName(b.name),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetVersionStringLabels(b.instance, b.name, b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: b.instance.Spec.JobTTLSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy:    corev1.RestartPolicyOnFailure,
					ImagePullSecrets: b.instance.Spec.ImagePullSecrets,
					Containers: []corev1.Container{
						{
							Name:                     "builder-script-runner",
							Image:                    fmt.Sprintf("%s:%s", b.instance.Spec.Builder.Image, b.instance.Spec.Builder.Version),
							ImagePullPolicy:          corev1.PullAlways,
							TerminationMessagePath:   corev1.TerminationMessagePathDefault,
							TerminationMessagePolicy: corev1.TerminationMessageReadFile,
							Command:                  append([]string{"/bin/sh", "-c"}, b.command...),
							Env:                      envVars,
							SecurityContext: &corev1.SecurityContext{
								Privileged: ptr.To(true),
							},
							VolumeMounts: volumeMounts,
						},
					},
					TerminationGracePeriodSeconds: ptr.To[int64](30),
					DNSPolicy:                     corev1.DNSClusterFirst,
					SecurityContext:               &corev1.PodSecurityContext{},
					SchedulerName:                 corev1.DefaultSchedulerName,
					Volumes:                       volumes,
				},
			},
		},
	}
}

func (b *JobBuilder) Enabled() bool {
	return true
}

func (b *JobBuilder) Update(object client.Object) error {
	job := object.(*batchv1.Job)
	if err := controllerutil.SetOwnerReference(b.instance, job, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %w", err)
	}
	return nil
}
