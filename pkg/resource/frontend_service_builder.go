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
	"go.temporal.io/server/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type FrontendServiceBuilder struct {
	instance *v1alpha1.TemporalCluster
	scheme   *runtime.Scheme
}

func NewFrontendServiceBuilder(instance *v1alpha1.TemporalCluster, scheme *runtime.Scheme) *FrontendServiceBuilder {
	return &FrontendServiceBuilder{
		instance: instance,
		scheme:   scheme,
	}
}

func (b *FrontendServiceBuilder) Build() (client.Object, error) {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.instance.ChildResourceName("frontend"),
			Namespace: b.instance.Namespace,
		},
	}, nil
}

func (b *FrontendServiceBuilder) Update(object client.Object) error {
	service := object.(*corev1.Service)
	service.Labels = object.GetLabels()
	service.Annotations = object.GetAnnotations()
	service.Spec.Type = corev1.ServiceTypeClusterIP
	service.Spec.Selector = metadata.LabelsSelector(b.instance.Name, common.FrontendServiceName)
	service.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "grpc-rpc",
			Protocol:   corev1.ProtocolTCP,
			Port:       int32(*b.instance.Spec.Services.Frontend.Port),
			TargetPort: intstr.FromString("rpc"),
		},
		{
			Name:       "metrics",
			Protocol:   corev1.ProtocolTCP,
			Port:       9090,
			TargetPort: intstr.FromString("metrics"),
		},
	}

	if err := controllerutil.SetControllerReference(b.instance, service, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}
