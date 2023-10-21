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

package ui

import (
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	"github.com/alexandrevilain/temporal-operator/pkg/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const UIServicePort = 8080

type ServiceBuilder struct {
	instance *v1beta1.TemporalCluster
	scheme   *runtime.Scheme
}

func NewServiceBuilder(instance *v1beta1.TemporalCluster, scheme *runtime.Scheme) *ServiceBuilder {
	return &ServiceBuilder{
		instance: instance,
		scheme:   scheme,
	}
}

func (b *ServiceBuilder) Build() client.Object {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.instance.ChildResourceName("ui"),
			Namespace: b.instance.Namespace,
		},
	}
}

func (b *ServiceBuilder) Enabled() bool {
	return b.instance.Spec.UI != nil && b.instance.Spec.UI.Enabled
}

func (b *ServiceBuilder) Update(object client.Object) error {
	service := object.(*corev1.Service)
	service.Labels = object.GetLabels()
	service.Annotations = object.GetAnnotations()
	service.Spec.Type = corev1.ServiceTypeClusterIP
	service.Spec.Selector = metadata.LabelsSelector(b.instance, "ui")
	service.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "http",
			TargetPort: intstr.FromString("http"),
			Protocol:   corev1.ProtocolTCP,
			Port:       int32(UIServicePort),
		},
	}

	if err := controllerutil.SetControllerReference(b.instance, service, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %w", err)
	}

	if b.instance.Spec.UI.Overrides != nil && b.instance.Spec.UI.Overrides.Service != nil {
		err := kubernetes.ApplyServiceOverrides(service, b.instance.Spec.UI.Overrides.Service)
		if err != nil {
			return fmt.Errorf("failed applying service overrides: %w", err)
		}
	}
	return nil
}
