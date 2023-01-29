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
	"context"
	"errors"
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	"github.com/alexandrevilain/temporal-operator/pkg/kubernetes"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/mtls/certmanager"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/persistence"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/prometheus"
	"go.temporal.io/server/common/primitives"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type DeploymentBuilder struct {
	serviceName string
	instance    *v1beta1.TemporalCluster
	scheme      *runtime.Scheme
	service     *v1beta1.ServiceSpec
}

func NewDeploymentBuilder(serviceName string, instance *v1beta1.TemporalCluster, scheme *runtime.Scheme, service *v1beta1.ServiceSpec) *DeploymentBuilder {
	return &DeploymentBuilder{
		serviceName: serviceName,
		instance:    instance,
		scheme:      scheme,
		service:     service,
	}
}

func (b *DeploymentBuilder) Build() (client.Object, error) {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName(b.serviceName),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance.Name, b.serviceName, b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}, nil
}

func (b *DeploymentBuilder) Update(object client.Object) error {
	deployment := object.(*appsv1.Deployment)
	deployment.Labels = metadata.Merge(
		object.GetLabels(),
		metadata.GetLabels(b.instance.Name, b.serviceName, b.instance.Spec.Version, b.instance.Labels),
	)
	deployment.Annotations = metadata.Merge(
		object.GetAnnotations(),
		metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
	)

	// worker has no grpc endpoint so omit liveness probe
	var livenessProbe *corev1.Probe
	if b.serviceName != primitives.WorkerService {
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
						Name: b.instance.ChildResourceName(ServiceConfig),
					},
					DefaultMode: pointer.Int32(corev1.ConfigMapVolumeSourceDefaultMode),
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
						Name: b.instance.ChildResourceName(ServiceDynamicConfig),
					},
					DefaultMode: pointer.Int32(corev1.ConfigMapVolumeSourceDefaultMode),
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "dynamicconfig",
			MountPath: "/etc/temporal/config/dynamic_config.yaml",
			SubPath:   "dynamic_config.yaml",
		})
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
							DefaultMode: pointer.Int32(corev1.SecretVolumeSourceDefaultMode),
						},
					},
				},
				corev1.Volume{
					Name: certmanager.InternodeCertificate,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  b.instance.ChildResourceName(certmanager.InternodeCertificate),
							DefaultMode: pointer.Int32(corev1.SecretVolumeSourceDefaultMode),
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
				corev1.VolumeMount{
					Name:      certmanager.WorkerFrontendClientCertificate,
					MountPath: b.instance.Spec.MTLS.Frontend.GetWorkerCertificateMountPath(),
				},
			)

			volumes = append(volumes,
				corev1.Volume{
					Name: certmanager.FrontendIntermediateCACertificate,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  b.instance.ChildResourceName(certmanager.FrontendIntermediateCACertificate),
							DefaultMode: pointer.Int32(corev1.SecretVolumeSourceDefaultMode),
						},
					},
				},
				corev1.Volume{
					Name: certmanager.FrontendCertificate,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  b.instance.ChildResourceName(certmanager.FrontendCertificate),
							DefaultMode: pointer.Int32(corev1.SecretVolumeSourceDefaultMode),
						},
					},
				},
				corev1.Volume{
					Name: certmanager.WorkerFrontendClientCertificate,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  b.instance.ChildResourceName(certmanager.WorkerFrontendClientCertificate),
							DefaultMode: pointer.Int32(corev1.SecretVolumeSourceDefaultMode),
						},
					},
				},
			)
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

	if b.instance.Spec.Metrics.MetricsEnabled() {
		if b.instance.Spec.Metrics.Prometheus != nil && b.instance.Spec.Metrics.Prometheus.ListenPort != nil {
			containerPorts = append(containerPorts, corev1.ContainerPort{
				Name:          prometheus.MetricsPortName.String(),
				ContainerPort: *b.instance.Spec.Metrics.Prometheus.ListenPort,
				Protocol:      corev1.ProtocolTCP,
			})
		}
	}

	deployment.Spec.Replicas = b.service.Replicas

	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: metadata.LabelsSelector(b.instance.Name, b.serviceName),
	}

	deployment.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: buildPodObjectMeta(b.instance, b.serviceName),
		Spec: corev1.PodSpec{
			ServiceAccountName:       b.instance.ChildResourceName(b.serviceName),
			DeprecatedServiceAccount: b.instance.ChildResourceName(b.serviceName),
			ImagePullSecrets:         b.instance.Spec.ImagePullSecrets,
			Containers: []corev1.Container{
				{
					Name:                     "service", // name "service" is here to simplify overrides
					Image:                    fmt.Sprintf("%s:%s", b.instance.Spec.Image, b.instance.Spec.Version),
					ImagePullPolicy:          corev1.PullAlways,
					TerminationMessagePath:   corev1.TerminationMessagePathDefault,
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					SecurityContext: &corev1.SecurityContext{
						AllowPrivilegeEscalation: pointer.Bool(false),
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
			RestartPolicy:                 corev1.RestartPolicyAlways,
			TerminationGracePeriodSeconds: pointer.Int64(30),
			DNSPolicy:                     corev1.DNSClusterFirst,
			SchedulerName:                 corev1.DefaultSchedulerName,
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser:    pointer.Int64(1000),
				RunAsGroup:   pointer.Int64(1000),
				FSGroup:      pointer.Int64(1000),
				RunAsNonRoot: pointer.Bool(true),
			},
			Volumes: volumes,
		},
	}

	if b.instance.Spec.Services.Overrides != nil && b.instance.Spec.Services.Overrides.Deployment != nil {
		err := ApplyDeploymentOverrides(deployment, b.instance.Spec.Services.Overrides.Deployment)
		if err != nil {
			return fmt.Errorf("can't apply deployment overrides: %v", err)
		}
	}

	if b.service.Overrides != nil && b.service.Overrides.Deployment != nil {
		err := ApplyDeploymentOverrides(deployment, b.service.Overrides.Deployment)
		if err != nil {
			return fmt.Errorf("failed applying deployment overrides: %v", err)
		}
	}

	if err := controllerutil.SetControllerReference(b.instance, deployment, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}

func (b *DeploymentBuilder) ReportServiceStatus(ctx context.Context, c client.Client) (*v1beta1.ServiceStatus, error) {
	status := &v1beta1.ServiceStatus{
		Name: b.serviceName,
	}

	deploy := &appsv1.Deployment{}

	namespacedName := types.NamespacedName{
		Name:      b.instance.ChildResourceName(b.serviceName),
		Namespace: b.instance.Namespace,
	}

	err := c.Get(ctx, namespacedName, deploy)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return status, nil
		}
		return status, err
	}

	var ok bool
	status.Version, ok = deploy.Labels["app.kubernetes.io/version"]
	if !ok {
		return nil, errors.New("can't determine service version from deployment labels")
	}

	status.Ready, err = kubernetes.IsDeploymentReady(deploy)
	if err != nil {
		return nil, fmt.Errorf("can't determine if deployment is ready: %w", err)
	}

	return status, nil
}
