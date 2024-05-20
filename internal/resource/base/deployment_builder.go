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

package base

import (
	"fmt"
	"path/filepath"

	"github.com/alexandrevilain/controller-tools/pkg/resource"
	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	"github.com/alexandrevilain/temporal-operator/internal/resource/meta"
	"github.com/alexandrevilain/temporal-operator/internal/resource/mtls/certmanager"
	"github.com/alexandrevilain/temporal-operator/internal/resource/persistence"
	"github.com/alexandrevilain/temporal-operator/internal/resource/prometheus"
	"github.com/alexandrevilain/temporal-operator/pkg/kubernetes"
	"go.temporal.io/server/common/primitives"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ resource.Builder = (*DeploymentBuilder)(nil)

type DeploymentBuilder struct {
	serviceName string
	instance    *v1beta1.TemporalCluster
	scheme      *runtime.Scheme
	service     *v1beta1.ServiceSpec
	configHash  string
}

func NewDeploymentBuilder(serviceName string, instance *v1beta1.TemporalCluster, scheme *runtime.Scheme, service *v1beta1.ServiceSpec, configHash string) *DeploymentBuilder {
	return &DeploymentBuilder{
		serviceName: serviceName,
		instance:    instance,
		scheme:      scheme,
		service:     service,
		configHash:  configHash,
	}
}

func (b *DeploymentBuilder) Build() client.Object {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName(b.serviceName),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance, b.serviceName, b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}
}

func (b *DeploymentBuilder) Enabled() bool {
	return isBuilderEnabled(b.instance, b.serviceName)
}

func (b *DeploymentBuilder) Update(object client.Object) error {
	deployment := object.(*appsv1.Deployment)
	deployment.Labels = metadata.Merge(
		object.GetLabels(),
		metadata.GetLabels(b.instance, b.serviceName, b.instance.Spec.Version, b.instance.Labels),
	)
	deployment.Annotations = metadata.Merge(
		object.GetAnnotations(),
		metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
	)

	// worker has no grpc endpoint so omit liveness probe
	var livenessProbe *corev1.Probe
	if b.serviceName != string(primitives.WorkerService) {
		livenessProbe = &corev1.Probe{
			InitialDelaySeconds: 150,
			TimeoutSeconds:      1,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			FailureThreshold:    3,
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromString("rpc"),
				},
			},
		}
	}

	envVars := []corev1.EnvVar{
		{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		},
		{
			Name:  "SERVICES",
			Value: b.serviceName,
		},
	}

	datastores := b.instance.Spec.Persistence.GetDatastores()

	envVars = append(envVars, persistence.GetDatastoresEnvironmentVariables(datastores)...)

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "config",
			MountPath: "/etc/temporal/config/config_template.yaml",
			SubPath:   "config_template.yaml",
		},
	}

	volumeMounts = append(volumeMounts, persistence.GetDatastoresVolumeMounts(datastores)...)

	volumes := []corev1.Volume{
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: b.instance.ChildResourceName(meta.ServiceConfig),
					},
					DefaultMode: ptr.To[int32](corev1.ConfigMapVolumeSourceDefaultMode),
				},
			},
		},
	}

	volumes = append(volumes, persistence.GetDatastoresVolumes(datastores)...)

	if b.instance.Spec.DynamicConfig != nil {
		volumes = append(volumes, corev1.Volume{
			Name: "dynamicconfig",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: b.instance.ChildResourceName(meta.ServiceDynamicConfig),
					},
					DefaultMode: ptr.To[int32](corev1.ConfigMapVolumeSourceDefaultMode),
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "dynamicconfig",
			MountPath: "/etc/temporal/config/dynamic_config.yaml",
			SubPath:   "dynamic_config.yaml",
		})
	}

	if b.instance.Spec.Archival.IsEnabled() {
		if b.instance.Spec.Archival.Provider.Kind() == v1beta1.S3ArchivalProviderKind &&
			b.instance.Spec.Archival.Provider.S3.Credentials != nil {
			envVars = append(envVars,
				corev1.EnvVar{
					Name: "AWS_ACCESS_KEY_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: b.instance.Spec.Archival.Provider.S3.Credentials.AccessKeyIDRef,
					},
				},
				corev1.EnvVar{
					Name: "AWS_SECRET_ACCESS_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: b.instance.Spec.Archival.Provider.S3.Credentials.SecretAccessKeyRef,
					},
				},
			)
		}

		if b.instance.Spec.Archival.Provider.Kind() == v1beta1.GCSArchivalProviderKind &&
			b.instance.Spec.Archival.Provider.GCS.CredentialsRef != nil {
			key := b.instance.Spec.Archival.Provider.GCS.CredentialsRef.Key
			if key == "" {
				key = "credentials.json"
			}
			volumes = append(volumes, corev1.Volume{
				Name: "archival",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: b.instance.Spec.Archival.Provider.GCS.CredentialsRef.Name,
						Items: []corev1.KeyToPath{
							{
								Key:  key,
								Path: filepath.Base(b.instance.Spec.Archival.Provider.GCS.CredentialsFileMountPath()),
							},
						},
						DefaultMode: ptr.To[int32](corev1.SecretVolumeSourceDefaultMode),
					},
				},
			})

			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      "archival",
				MountPath: filepath.Dir(b.instance.Spec.Archival.Provider.GCS.CredentialsFileMountPath()),
			})
		}
	}

	if b.instance.MTLSWithCertManagerEnabled() {
		if b.instance.Spec.MTLS.InternodeEnabled() {
			volumeMounts = append(volumeMounts,
				corev1.VolumeMount{
					Name:      certmanager.InternodeIntermediateCACertificate,
					MountPath: b.instance.Spec.MTLS.Internode.GetIntermediateCACertificateMountPath(),
				},
				corev1.VolumeMount{
					Name:      certmanager.InternodeCertificate,
					MountPath: b.instance.Spec.MTLS.Internode.GetCertificateMountPath(),
				},
			)

			volumes = append(volumes,
				corev1.Volume{
					Name: certmanager.InternodeIntermediateCACertificate,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  b.instance.ChildResourceName(certmanager.InternodeIntermediateCACertificate),
							DefaultMode: ptr.To[int32](corev1.SecretVolumeSourceDefaultMode),
						},
					},
				},
				corev1.Volume{
					Name: certmanager.InternodeCertificate,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  b.instance.ChildResourceName(certmanager.InternodeCertificate),
							DefaultMode: ptr.To[int32](corev1.SecretVolumeSourceDefaultMode),
						},
					},
				},
			)
		}
		if b.instance.Spec.MTLS.FrontendEnabled() {
			volumeMounts = append(volumeMounts,
				corev1.VolumeMount{
					Name:      certmanager.FrontendIntermediateCACertificate,
					MountPath: b.instance.Spec.MTLS.Frontend.GetIntermediateCACertificateMountPath(),
				},
				corev1.VolumeMount{
					Name:      certmanager.FrontendCertificate,
					MountPath: b.instance.Spec.MTLS.Frontend.GetCertificateMountPath(),
				},
			)

			volumes = append(volumes,
				corev1.Volume{
					Name: certmanager.FrontendIntermediateCACertificate,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  b.instance.ChildResourceName(certmanager.FrontendIntermediateCACertificate),
							DefaultMode: ptr.To[int32](corev1.SecretVolumeSourceDefaultMode),
						},
					},
				},
				corev1.Volume{
					Name: certmanager.FrontendCertificate,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  b.instance.ChildResourceName(certmanager.FrontendCertificate),
							DefaultMode: ptr.To[int32](corev1.SecretVolumeSourceDefaultMode),
						},
					},
				},
			)

			if !b.instance.Spec.Services.InternalFrontend.IsEnabled() {
				volumeMounts = append(volumeMounts, corev1.VolumeMount{
					Name:      certmanager.WorkerFrontendClientCertificate,
					MountPath: b.instance.Spec.MTLS.Frontend.GetWorkerCertificateMountPath(),
				})

				volumes = append(volumes, corev1.Volume{
					Name: certmanager.WorkerFrontendClientCertificate,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  b.instance.ChildResourceName(certmanager.WorkerFrontendClientCertificate),
							DefaultMode: ptr.To[int32](corev1.SecretVolumeSourceDefaultMode),
						},
					},
				})
			}
		}
	}

	containerPorts := []corev1.ContainerPort{
		{
			Name:          "rpc",
			ContainerPort: int32(*b.service.Port),
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "membership",
			ContainerPort: int32(*b.service.MembershipPort),
			Protocol:      corev1.ProtocolTCP,
		},
	}

	if b.instance.Spec.Metrics.IsEnabled() {
		if b.instance.Spec.Metrics.Prometheus != nil && b.instance.Spec.Metrics.Prometheus.ListenPort != nil {
			containerPorts = append(containerPorts, corev1.ContainerPort{
				Name:          prometheus.MetricsPortName.String(),
				ContainerPort: *b.instance.Spec.Metrics.Prometheus.ListenPort,
				Protocol:      corev1.ProtocolTCP,
			})
		}
	}

	if b.serviceName == string(primitives.FrontendService) && b.instance.Spec.Services.Frontend.HTTPPort != nil {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          "http",
			ContainerPort: int32(*b.instance.Spec.Services.Frontend.HTTPPort),
			Protocol:      corev1.ProtocolTCP,
		})
	}

	deployment.Spec.Replicas = b.service.Replicas

	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: metadata.LabelsSelector(b.instance, b.serviceName),
	}

	deployment.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: meta.BuildPodObjectMeta(b.instance, b.serviceName, b.configHash),
		Spec: corev1.PodSpec{
			ServiceAccountName:       b.instance.ChildResourceName(b.serviceName),
			DeprecatedServiceAccount: b.instance.ChildResourceName(b.serviceName),
			ImagePullSecrets:         b.instance.Spec.ImagePullSecrets,
			Containers: []corev1.Container{
				{
					Name:                     "service", // name "service" is here to simplify overrides
					Image:                    fmt.Sprintf("%s:%s", b.instance.Spec.Image, b.instance.Spec.Version),
					ImagePullPolicy:          corev1.PullIfNotPresent,
					Resources:                b.service.Resources,
					TerminationMessagePath:   corev1.TerminationMessagePathDefault,
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					SecurityContext: &corev1.SecurityContext{
						AllowPrivilegeEscalation: ptr.To(false),
						Capabilities: &corev1.Capabilities{
							Drop: []corev1.Capability{"ALL"},
						},
					},
					Ports:         containerPorts,
					LivenessProbe: livenessProbe,
					Env:           envVars,
					VolumeMounts:  volumeMounts,
				},
			},
			InitContainers:                b.service.InitContainers,
			RestartPolicy:                 corev1.RestartPolicyAlways,
			TerminationGracePeriodSeconds: ptr.To[int64](30),
			DNSPolicy:                     corev1.DNSClusterFirst,
			SchedulerName:                 corev1.DefaultSchedulerName,
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser:    ptr.To[int64](1000),
				RunAsGroup:   ptr.To[int64](1000),
				FSGroup:      ptr.To[int64](1000),
				RunAsNonRoot: ptr.To(true),
			},
			Volumes: volumes,
		},
	}

	if b.instance.Spec.Services.Overrides != nil && b.instance.Spec.Services.Overrides.Deployment != nil {
		err := kubernetes.ApplyDeploymentOverrides(deployment, b.instance.Spec.Services.Overrides.Deployment)
		if err != nil {
			return fmt.Errorf("can't apply deployment overrides: %w", err)
		}
	}

	if b.service.Overrides != nil && b.service.Overrides.Deployment != nil {
		err := kubernetes.ApplyDeploymentOverrides(deployment, b.service.Overrides.Deployment)
		if err != nil {
			return fmt.Errorf("failed applying deployment overrides: %w", err)
		}
	}

	if err := controllerutil.SetControllerReference(b.instance, deployment, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %w", err)
	}

	return nil
}
