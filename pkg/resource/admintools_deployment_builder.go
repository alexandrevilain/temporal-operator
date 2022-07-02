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

package resource

import (
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type AdminToolsDeploymentBuilder struct {
	instance *v1alpha1.TemporalCluster
	scheme   *runtime.Scheme
}

func NewAdminToolsDeploymentBuilder(instance *v1alpha1.TemporalCluster, scheme *runtime.Scheme) *AdminToolsDeploymentBuilder {
	return &AdminToolsDeploymentBuilder{
		instance: instance,
		scheme:   scheme,
	}
}

func (b *AdminToolsDeploymentBuilder) Build() (client.Object, error) {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName("admintools"),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance.Name, "admintools", b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}, nil
}

func (b *AdminToolsDeploymentBuilder) Update(object client.Object) error {
	deployment := object.(*appsv1.Deployment)
	deployment.Labels = metadata.Merge(
		object.GetLabels(),
		metadata.GetLabels(b.instance.Name, "admintools", b.instance.Spec.Version, b.instance.Labels),
	)
	deployment.Annotations = metadata.Merge(
		object.GetAnnotations(),
		metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
	)

	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: metadata.LabelsSelector(b.instance.Name, "admintools"),
	}
	deployment.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      metadata.GetLabels(b.instance.Name, "admintools", b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}

	container := corev1.Container{
		Name:                     "admintools",
		Image:                    fmt.Sprintf("%s:%s", b.instance.Spec.AdminTools.Image, b.instance.Spec.Version),
		ImagePullPolicy:          corev1.PullAlways,
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
		Env: []corev1.EnvVar{
			{
				Name:  "TEMPORAL_CLI_ADDRESS",
				Value: fmt.Sprintf("%s:%d", b.instance.ChildResourceName("frontend"), *b.instance.Spec.Services.Frontend.Port),
			},
		},
		LivenessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"ls", "/"},
				},
			},
			InitialDelaySeconds: 5,
			TimeoutSeconds:      1,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			FailureThreshold:    3,
		},
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: pointer.Bool(false),
		},
	}

	if b.instance.MTLSEnabled() && b.instance.Spec.MTLS.FrontendEnabled() {
		container.VolumeMounts = append(container.VolumeMounts,
			corev1.VolumeMount{
				Name:      "admintools-mtls-certificate",
				MountPath: "/etc/temporal/config/certs/client/admintools",
			},
		)
		container.Env = append(container.Env,
			corev1.EnvVar{
				Name:  "TEMPORAL_CLI_TLS_CA",
				Value: "/etc/temporal/config/certs/client/admintools/ca.crt",
			},
			corev1.EnvVar{
				Name:  "TEMPORAL_CLI_TLS_CERT",
				Value: "/etc/temporal/config/certs/client/admintools/tls.crt",
			},
			corev1.EnvVar{
				Name:  "TEMPORAL_CLI_TLS_KEY",
				Value: "/etc/temporal/config/certs/client/admintools/tls.key",
			},
			corev1.EnvVar{
				Name:  "TEMPORAL_CLI_TLS_ENABLE_HOST_VERIFICATION",
				Value: "true",
			},
			corev1.EnvVar{
				Name:  "TEMPORAL_CLI_TLS_SERVER_NAME",
				Value: b.instance.Spec.MTLS.Frontend.ServerName(b.instance.ServerName()),
			},
		)
	}

	deployment.Spec.Template.Spec = corev1.PodSpec{
		ImagePullSecrets: b.instance.Spec.ImagePullSecrets,
		Containers: []corev1.Container{
			container,
		},
		RestartPolicy:                 corev1.RestartPolicyAlways,
		TerminationGracePeriodSeconds: pointer.Int64(30),
		DNSPolicy:                     corev1.DNSClusterFirst,
		SecurityContext:               &corev1.PodSecurityContext{},
		SchedulerName:                 "default-scheduler",
	}

	if b.instance.MTLSEnabled() && b.instance.Spec.MTLS.FrontendEnabled() {
		if b.instance.Spec.MTLS.InternodeEnabled() {
			deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes,
				corev1.Volume{
					Name: "admintools-mtls-certificate",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: b.instance.ChildResourceName("admintools-mtls-certificate"),
						},
					},
				},
			)
		}
	}

	if err := controllerutil.SetControllerReference(b.instance, deployment, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}
