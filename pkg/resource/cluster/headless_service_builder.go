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

package cluster

import (
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	"github.com/alexandrevilain/temporal-operator/pkg/resource"
	"github.com/alexandrevilain/temporal-operator/pkg/resource/prometheus"
	"go.temporal.io/server/common/primitives"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type HeadlessServiceBuilder struct {
	serviceName string
	instance    *v1beta1.TemporalCluster
	scheme      *runtime.Scheme
	service     *v1beta1.ServiceSpec
}

func NewHeadlessServiceBuilder(serviceName string, instance *v1beta1.TemporalCluster, scheme *runtime.Scheme, service *v1beta1.ServiceSpec) resource.Builder {
	return &HeadlessServiceBuilder{
		serviceName: serviceName,
		instance:    instance,
		scheme:      scheme,
		service:     service,
	}
}

func (b *HeadlessServiceBuilder) Build() client.Object {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.instance.ChildResourceName(fmt.Sprintf("%s-headless", b.serviceName)),
			Namespace: b.instance.Namespace,
			Labels: metadata.Merge(
				metadata.GetLabels(b.instance.Name, b.serviceName, b.instance.Spec.Version, b.instance.Labels),
				metadata.HeadlessLabels(),
			),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}
}

func (b *HeadlessServiceBuilder) Update(object client.Object) error {
	service := object.(*corev1.Service)
	service.Labels = metadata.Merge(
		object.GetLabels(),
		metadata.GetLabels(b.instance.Name, b.serviceName, b.instance.Spec.Version, b.instance.Labels),
		metadata.HeadlessLabels(),
	)
	service.Annotations = metadata.Merge(
		object.GetAnnotations(),
		metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
	)
	service.Spec.Type = corev1.ServiceTypeClusterIP
	service.Spec.ClusterIP = corev1.ClusterIPNone
	service.Spec.Selector = metadata.LabelsSelector(b.instance.Name, b.serviceName)

	service.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "http-metrics",
			TargetPort: prometheus.MetricsPortName,
			Protocol:   corev1.ProtocolTCP,
			Port:       9090,
		},
	}

	if b.serviceName != string(primitives.WorkerService) {
		service.Spec.Ports = append(service.Spec.Ports,
			corev1.ServicePort{
				// Here "tcp-" is used instead of "grpc-" because temporal uses
				// pod-to-pod traffic over ip. Because no "Host" header is set,
				// istio can't create mTLS for gRPC.
				Name:       "tcp-rpc",
				TargetPort: intstr.FromString("rpc"),
				Protocol:   corev1.ProtocolTCP,
				Port:       int32(*b.service.Port),
			},
			corev1.ServicePort{
				Name:       "tcp-membership",
				TargetPort: intstr.FromString("membership"),
				Protocol:   corev1.ProtocolTCP,
				Port:       int32(*b.service.MembershipPort),
			},
		)
	}

	service.Spec.Ports = []corev1.ServicePort{

		{
			Name:       "http-metrics",
			TargetPort: prometheus.MetricsPortName,
			Protocol:   corev1.ProtocolTCP,
			Port:       9090,
		},
	}

	if err := controllerutil.SetControllerReference(b.instance, service, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}
	return nil
}
