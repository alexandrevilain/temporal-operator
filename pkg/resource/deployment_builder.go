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

	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	"github.com/alexandrevilain/temporal-operator/pkg/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type DeploymentBuilder struct {
	serviceName string
	instance    *v1alpha1.TemporalCluster
	scheme      *runtime.Scheme
	service     *v1alpha1.ServiceSpec
}

func NewDeploymentBuilder(serviceName string, instance *v1alpha1.TemporalCluster, scheme *runtime.Scheme, service *v1alpha1.ServiceSpec) *DeploymentBuilder {
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
		metadata.GetLabels(b.instance.Name, b.serviceName, b.instance.Spec.Version, b.instance.Labels),
		object.GetLabels(),
	)
	deployment.Annotations = metadata.Merge(
		metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		object.GetAnnotations(),
	)

	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: metadata.LabelsSelector(b.instance.Name, b.serviceName),
	}
	deployment.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      metadata.GetLabels(b.instance.Name, b.serviceName, b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}

	serviceContainer := corev1.Container{
		Name:                     b.serviceName,
		Image:                    fmt.Sprintf("%s:%s", b.instance.Spec.Image, b.instance.Spec.Version),
		ImagePullPolicy:          corev1.PullAlways,
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: pointer.Bool(false),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "rpc",
				ContainerPort: int32(*b.service.Port),
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          "metrics",
				ContainerPort: 9090,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
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
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "config",
				MountPath: "/etc/temporal/config/config_template.yaml",
				SubPath:   "config_template.yaml",
			},
		},
	}

	for _, datastore := range b.instance.Spec.Datastores {
		serviceContainer.Env = append(serviceContainer.Env,
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
			})
	}
	if b.instance.MTLSEnabled() {
		if b.instance.Spec.MTLS.InternodeEnabled() {
			serviceContainer.VolumeMounts = append(serviceContainer.VolumeMounts,
				corev1.VolumeMount{
					Name:      "internode-intermediate-ca",
					MountPath: b.instance.Spec.MTLS.Internode.GetIntermediateCACertificateMountPath(),
				},
				corev1.VolumeMount{
					Name:      "internode-certificate",
					MountPath: b.instance.Spec.MTLS.Internode.GetCertificateMountPath(),
				},
			)
		}
		if b.instance.Spec.MTLS.FrontendEnabled() {
			serviceContainer.VolumeMounts = append(serviceContainer.VolumeMounts,
				corev1.VolumeMount{
					Name:      "frontend-intermediate-ca",
					MountPath: b.instance.Spec.MTLS.Frontend.GetIntermediateCACertificateMountPath(),
				},
				corev1.VolumeMount{
					Name:      "frontend-certificate",
					MountPath: b.instance.Spec.MTLS.Frontend.GetCertificateMountPath(),
				},
				corev1.VolumeMount{
					Name:      "worker-certificate",
					MountPath: b.instance.Spec.MTLS.Frontend.GetWorkerCertificateMountPath(),
				},
			)
		}
	}

	deployment.Spec.Replicas = b.service.Replicas

	deployment.Spec.Template.Spec = corev1.PodSpec{
		ImagePullSecrets: b.instance.Spec.ImagePullSecrets,
		Containers: []corev1.Container{
			serviceContainer,
		},
		RestartPolicy:                 corev1.RestartPolicyAlways,
		TerminationGracePeriodSeconds: pointer.Int64(30),
		DNSPolicy:                     corev1.DNSClusterFirst,
		SchedulerName:                 "default-scheduler",
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser:    pointer.Int64(1000),
			RunAsGroup:   pointer.Int64(1000),
			FSGroup:      pointer.Int64(1000),
			RunAsNonRoot: pointer.Bool(true),
		},
		Volumes: []corev1.Volume{
			{
				Name: "config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: b.instance.ChildResourceName(ServiceConfig),
						},
						DefaultMode: pointer.Int32(420),
					},
				},
			},
		},
	}

	if b.instance.MTLSEnabled() {
		if b.instance.Spec.MTLS.InternodeEnabled() {
			deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes,
				corev1.Volume{
					Name: "internode-intermediate-ca",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: b.instance.ChildResourceName("internode-intermediate-ca-certificate"),
						},
					},
				},
				corev1.Volume{
					Name: "internode-certificate",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: b.instance.ChildResourceName("internode-certificate"),
						},
					},
				},
			)
		}

		if b.instance.Spec.MTLS.FrontendEnabled() {
			deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes,
				corev1.Volume{
					Name: "frontend-intermediate-ca",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: b.instance.ChildResourceName("frontend-intermediate-ca-certificate"),
						},
					},
				},
				corev1.Volume{
					Name: "frontend-certificate",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: b.instance.ChildResourceName("frontend-certificate"),
						},
					},
				},
				corev1.Volume{
					Name: "worker-certificate",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: b.instance.ChildResourceName("worker-certificate"),
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

func (b *DeploymentBuilder) ReportServiceStatus(ctx context.Context, c client.Client) (*v1alpha1.ServiceStatus, error) {
	deploy := &appsv1.Deployment{}
	err := c.Get(ctx, types.NamespacedName{
		Name:      b.instance.ChildResourceName(b.serviceName),
		Namespace: b.instance.Namespace,
	}, deploy)
	if err != nil {
		return nil, err
	}
	val, ok := deploy.Labels["app.kubernetes.io/version"]
	if !ok {
		return nil, errors.New("can't determine service version from deployment labels")
	}

	return &v1alpha1.ServiceStatus{
		Name:    b.serviceName,
		Version: val,
		Ready:   kubernetes.IsDeploymentReady(deploy),
	}, nil
}
