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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type WorkerProcessDeploymentBuilder struct {
	instance *v1beta1.TemporalWorkerProcess
	cluster  *v1beta1.TemporalCluster
	scheme   *runtime.Scheme
}

func NewWorkerProcessDeploymentBuilder(instance *v1beta1.TemporalWorkerProcess, cluster *v1beta1.TemporalCluster, scheme *runtime.Scheme) *WorkerProcessDeploymentBuilder {
	return &WorkerProcessDeploymentBuilder{
		instance: instance,
		cluster:  cluster,
		scheme:   scheme,
	}
}

func (b *WorkerProcessDeploymentBuilder) Build() (client.Object, error) {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName("worker"),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance.Name, "worker", b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}, nil
}

func (b *WorkerProcessDeploymentBuilder) Update(object client.Object) error {
	deployment := object.(*appsv1.Deployment)
	deployment.Labels = metadata.Merge(
		object.GetLabels(),
		metadata.GetLabels(b.instance.Name, "worker", b.instance.Spec.Version, b.instance.Labels),
	)
	deployment.Annotations = metadata.Merge(
		object.GetAnnotations(),
		metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
	)

	env := []corev1.EnvVar{
		{
			Name:  "TEMPORAL_HOST_URL",
			Value: b.cluster.GetPublicClientAddress(),
		},
		{
			Name:  "TEMPORAL_NAMESPACE",
			Value: b.instance.Spec.TemporalNamespace,
		},
	}

	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}

	deployment.Spec = appsv1.DeploymentSpec{
		Replicas: b.instance.Spec.Replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: metadata.LabelsSelector(b.instance.Name, "worker"),
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: buildWorkerProcessPodObjectMeta(b.instance, "worker"),
			Spec: corev1.PodSpec{
				ImagePullSecrets: b.instance.Spec.ImagePullSecrets,
				Containers: []corev1.Container{
					{
						Name:                     "worker",
						Image:                    fmt.Sprintf("%s:%s", b.instance.Spec.Image, b.instance.Spec.Version),
						ImagePullPolicy:          corev1.PullPolicy(b.instance.Spec.PullPolicy),
						TerminationMessagePath:   corev1.TerminationMessagePathDefault,
						TerminationMessagePolicy: corev1.TerminationMessageReadFile,
						Env:                      env,
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
						VolumeMounts: volumeMounts,
					},
				},
				RestartPolicy:                 corev1.RestartPolicyAlways,
				TerminationGracePeriodSeconds: pointer.Int64(30),
				DNSPolicy:                     corev1.DNSClusterFirst,
				SecurityContext:               &corev1.PodSecurityContext{},
				SchedulerName:                 corev1.DefaultSchedulerName,
				Volumes:                       volumes,
			},
		},
	}

	if err := controllerutil.SetControllerReference(b.instance, deployment, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}
