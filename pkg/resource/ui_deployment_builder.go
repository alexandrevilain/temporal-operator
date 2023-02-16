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

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/mtls/certmanager"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	uiCertsMountPath = "/etc/temporal/config/certs/client/ui"
)

type UIDeploymentBuilder struct {
	instance *v1beta1.TemporalCluster
	scheme   *runtime.Scheme
}

func NewUIDeploymentBuilder(instance *v1beta1.TemporalCluster, scheme *runtime.Scheme) *UIDeploymentBuilder {
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
		object.GetLabels(),
		metadata.GetLabels(b.instance.Name, "ui", b.instance.Spec.Version, b.instance.Labels),
	)
	deployment.Annotations = metadata.Merge(
		object.GetAnnotations(),
		metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
	)

	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}

	env := []corev1.EnvVar{
		{
			Name:  "TEMPORAL_ADDRESS",
			Value: fmt.Sprintf("%s:%d", b.instance.ChildResourceName(FrontendService), *b.instance.Spec.Services.Frontend.Port),
		},
		{
			Name:  "TEMPORAL_UI_PORT",
			Value: "8080",
		},
	}

	if b.instance.MTLSWithCertManagerEnabled() && b.instance.Spec.MTLS.FrontendEnabled() {
		volumeMounts = append(volumeMounts,
			corev1.VolumeMount{
				Name:      certmanager.UIFrontendClientCertificate,
				MountPath: uiCertsMountPath,
			},
		)

		volumes = append(volumes,
			corev1.Volume{
				Name: certmanager.UIFrontendClientCertificate,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  b.instance.ChildResourceName(certmanager.UIFrontendClientCertificate),
						DefaultMode: pointer.Int32(corev1.SecretVolumeSourceDefaultMode),
					},
				},
			},
		)

		env = append(env, certmanager.GetTLSEnvironmentVariables(b.instance, "TEMPORAL", uiCertsMountPath)...)
	}

	deployment.Spec.Replicas = pointer.Int32(1)
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: metadata.LabelsSelector(b.instance.Name, "ui"),
	}
	deployment.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: buildPodObjectMeta(b.instance, "ui"),
		Spec: corev1.PodSpec{
			ImagePullSecrets: b.instance.Spec.ImagePullSecrets,
			Containers: []corev1.Container{
				{
					Name:                     "ui",
					Image:                    fmt.Sprintf("%s:%s", b.instance.Spec.UI.Image, b.instance.Spec.UI.Version),
					ImagePullPolicy:          corev1.PullAlways,
					TerminationMessagePath:   corev1.TerminationMessagePathDefault,
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: int32(8080),
							Protocol:      corev1.ProtocolTCP,
						},
					},
					Env:          env,
					VolumeMounts: volumeMounts,
				},
			},
			Volumes:                       volumes,
			RestartPolicy:                 corev1.RestartPolicyAlways,
			TerminationGracePeriodSeconds: pointer.Int64(30),
			DNSPolicy:                     corev1.DNSClusterFirst,
			SchedulerName:                 corev1.DefaultSchedulerName,
			SecurityContext:               &corev1.PodSecurityContext{},
		},
	}

	if err := controllerutil.SetControllerReference(b.instance, deployment, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}
