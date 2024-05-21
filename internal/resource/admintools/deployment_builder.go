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

package admintools

import (
	"fmt"

	"github.com/alexandrevilain/controller-tools/pkg/resource"
	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	"github.com/alexandrevilain/temporal-operator/internal/resource/meta"
	"github.com/alexandrevilain/temporal-operator/internal/resource/mtls/certmanager"
	"github.com/alexandrevilain/temporal-operator/pkg/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ resource.Builder = (*DeploymentBuilder)(nil)

const (
	admintoolsCertsMountPath = "/etc/temporal/config/certs/client/admintools"
)

type DeploymentBuilder struct {
	instance   *v1beta1.TemporalCluster
	scheme     *runtime.Scheme
	configHash string
}

func NewDeploymentBuilder(instance *v1beta1.TemporalCluster, scheme *runtime.Scheme, configHash string) *DeploymentBuilder {
	return &DeploymentBuilder{
		instance:   instance,
		scheme:     scheme,
		configHash: configHash,
	}
}

func (b *DeploymentBuilder) Build() client.Object {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName("admintools"),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance, "admintools", b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}
}

func (b *DeploymentBuilder) Enabled() bool {
	return b.instance.Spec.AdminTools != nil && b.instance.Spec.AdminTools.Enabled
}

func (b *DeploymentBuilder) Update(object client.Object) error {
	deployment := object.(*appsv1.Deployment)
	deployment.Labels = metadata.Merge(
		object.GetLabels(),
		metadata.GetLabels(b.instance, "admintools", b.instance.Spec.Version, b.instance.Labels),
	)
	deployment.Annotations = metadata.Merge(
		object.GetAnnotations(),
		metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
	)

	address := fmt.Sprintf("%s:%d", b.instance.ChildResourceName(meta.FrontendService), *b.instance.Spec.Services.Frontend.Port)
	env := []corev1.EnvVar{
		{
			Name:  "TEMPORAL_CLI_ADDRESS", // tctl
			Value: address,
		},
		{
			Name:  "TEMPORAL_ADDRESS", // temporal
			Value: address,
		},
	}

	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}

	if b.instance.MTLSWithCertManagerEnabled() && b.instance.Spec.MTLS.FrontendEnabled() {
		volumeMounts = append(volumeMounts,
			corev1.VolumeMount{
				Name:      certmanager.AdmintoolsFrontendClientCertificate,
				MountPath: admintoolsCertsMountPath,
			},
		)
		volumes = append(volumes,
			corev1.Volume{
				Name: certmanager.AdmintoolsFrontendClientCertificate,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  b.instance.ChildResourceName(certmanager.AdmintoolsFrontendClientCertificate),
						DefaultMode: ptr.To[int32](corev1.SecretVolumeSourceDefaultMode),
					},
				},
			},
		)

		// Add tctl environment variables
		env = append(env, certmanager.GetTLSEnvironmentVariables(b.instance, "TEMPORAL_CLI", admintoolsCertsMountPath)...)
		// Add temporal cli environment variables (>= 0.9.0)
		env = append(env, certmanager.GetTLSEnvironmentVariables(b.instance, "TEMPORAL", admintoolsCertsMountPath)...)
	}

	deployment.Spec.Replicas = ptr.To[int32](1)

	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: metadata.LabelsSelector(b.instance, "admintools"),
	}

	deployment.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: meta.BuildPodObjectMeta(b.instance, "admintools", b.configHash),
		Spec: corev1.PodSpec{
			ImagePullSecrets: b.instance.Spec.ImagePullSecrets,
			Containers: []corev1.Container{
				{
					Name:                     "admintools",
					Image:                    fmt.Sprintf("%s:%s", b.instance.Spec.AdminTools.Image, b.instance.Spec.Version),
					ImagePullPolicy:          corev1.PullIfNotPresent,
					TerminationMessagePath:   corev1.TerminationMessagePathDefault,
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					Env:                      env,
					Resources:                b.instance.Spec.AdminTools.Resources,
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
						AllowPrivilegeEscalation: ptr.To(false),
					},
					VolumeMounts: volumeMounts,
				},
			},
			RestartPolicy:                 corev1.RestartPolicyAlways,
			TerminationGracePeriodSeconds: ptr.To[int64](30),
			DNSPolicy:                     corev1.DNSClusterFirst,
			SecurityContext:               &corev1.PodSecurityContext{},
			SchedulerName:                 corev1.DefaultSchedulerName,
			Volumes:                       volumes,
		},
	}

	if b.instance.Spec.AdminTools.Overrides != nil && b.instance.Spec.AdminTools.Overrides.Deployment != nil {
		err := kubernetes.ApplyDeploymentOverrides(deployment, b.instance.Spec.AdminTools.Overrides.Deployment)
		if err != nil {
			return fmt.Errorf("can't apply deployment overrides: %w", err)
		}
	}

	if err := controllerutil.SetControllerReference(b.instance, deployment, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %w", err)
	}

	return nil
}
