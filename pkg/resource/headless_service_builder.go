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

func NewHeadlessServiceBuilder(serviceName string, instance *v1beta1.TemporalCluster, scheme *runtime.Scheme, service *v1beta1.ServiceSpec) *HeadlessServiceBuilder {
	return &HeadlessServiceBuilder{
		serviceName: serviceName,
		instance:    instance,
		scheme:      scheme,
		service:     service,
	}
}

func (b *HeadlessServiceBuilder) Build() (client.Object, error) {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.instance.ChildResourceName(fmt.Sprintf("%s-headless", b.serviceName)),
			Namespace: b.instance.Namespace,
		},
	}, nil
}

func (b *HeadlessServiceBuilder) Update(object client.Object) error {
	service := object.(*corev1.Service)
	service.Labels = object.GetLabels()
	service.Annotations = object.GetAnnotations()

	service.Spec = corev1.ServiceSpec{
		Type:      corev1.ServiceTypeClusterIP,
		ClusterIP: corev1.ClusterIPNone,
		Selector:  metadata.LabelsSelector(b.instance.Name, b.serviceName),
		Ports: []corev1.ServicePort{
			{
				// Here "tcp-" is used instead of "grpc-" because temporal uses
				// pod-to-pod traffic over ip. Because no "Host" header is set,
				// istio can't create mTLS for gRPC.
				Name:       "tcp-rpc",
				TargetPort: intstr.FromString("rpc"),
				Protocol:   corev1.ProtocolTCP,
				Port:       int32(*b.service.Port),
			},
			{
				Name:       "tcp-membership",
				TargetPort: intstr.FromString("membership"),
				Protocol:   corev1.ProtocolTCP,
				Port:       int32(*b.service.MembershipPort),
			},
			{
				Name:       "http-metrics",
				TargetPort: intstr.FromString("metrics"),
				Protocol:   corev1.ProtocolTCP,
				Port:       9090,
			},
		},
	}

	if err := controllerutil.SetControllerReference(b.instance, service, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}
	return nil
}
