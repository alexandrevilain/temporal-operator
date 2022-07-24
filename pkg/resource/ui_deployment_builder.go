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
	"github.com/alexandrevilain/temporal-operator/pkg/resource/linkerd"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type UIDeploymentBuilder struct {
	instance *v1alpha1.TemporalCluster
	scheme   *runtime.Scheme
}

func NewUIDeploymentBuilder(instance *v1alpha1.TemporalCluster, scheme *runtime.Scheme) *UIDeploymentBuilder {
	return &UIDeploymentBuilder{
		instance: instance,
		scheme:   scheme,
	}
}

func (b *UIDeploymentBuilder) Build() (client.Object, error) {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName("ui"),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance.Name, "ui", b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}, nil
}

func (b *UIDeploymentBuilder) Update(object client.Object) error {
	deployment := object.(*appsv1.Deployment)
	deployment.Labels = metadata.Merge(
		metadata.GetLabels(b.instance.Name, "ui", b.instance.Spec.Version, b.instance.Labels),
		object.GetLabels(),
	)
	deployment.Annotations = metadata.Merge(
		metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		object.GetAnnotations(),
	)

	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: metadata.LabelsSelector(b.instance.Name, "ui"),
	}
	deployment.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: metadata.GetLabels(b.instance.Name, "ui", b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.Merge(
				linkerd.GetAnnotations(b.instance),
				metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
			),
		},
	}

	container := corev1.Container{
		Name:                     "ui",
		Image:                    fmt.Sprintf("%s:%s", b.instance.Spec.UI.Image, b.instance.Spec.UI.Version),
		ImagePullPolicy:          corev1.PullAlways,
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: int32(8080),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "TEMPORAL_ADDRESS",
				Value: fmt.Sprintf("%s:%d", b.instance.ChildResourceName("frontend"), *b.instance.Spec.Services.Frontend.Port),
			},
			{
				Name:  "TEMPORAL_CORS_ORIGINS",
				Value: "",
			},
		},
	}

	if b.instance.MTLSWithCertManagerEnabled() && b.instance.Spec.MTLS.FrontendEnabled() {
		container.VolumeMounts = append(container.VolumeMounts,
			corev1.VolumeMount{
				Name:      "ui-mtls-certificate",
				MountPath: "/etc/temporal/config/certs/client/ui",
			},
		)
		container.Env = append(container.Env,
			corev1.EnvVar{
				Name:  "TEMPORAL_TLS_CA",
				Value: "/etc/temporal/config/certs/client/ui/ca.crt",
			},
			corev1.EnvVar{
				Name:  "TEMPORAL_TLS_CERT",
				Value: "/etc/temporal/config/certs/client/ui/tls.crt",
			},
			corev1.EnvVar{
				Name:  "TEMPORAL_TLS_KEY",
				Value: "/etc/temporal/config/certs/client/ui/tls.key",
			},
			corev1.EnvVar{
				Name:  "TEMPORAL_TLS_ENABLE_HOST_VERIFICATION",
				Value: "true",
			},
			corev1.EnvVar{
				Name:  "TEMPORAL_TLS_SERVER_NAME",
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
		SchedulerName:                 "default-scheduler",
		SecurityContext:               &corev1.PodSecurityContext{},
	}

	if b.instance.MTLSWithCertManagerEnabled() && b.instance.Spec.MTLS.FrontendEnabled() {
		if b.instance.Spec.MTLS.InternodeEnabled() {
			deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes,
				corev1.Volume{
					Name: "ui-mtls-certificate",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: b.instance.ChildResourceName("ui-mtls-certificate"),
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
