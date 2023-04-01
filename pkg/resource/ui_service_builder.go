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

const UIServicePort = 8080

type UIServiceBuilder struct {
	instance *v1beta1.TemporalCluster
	scheme   *runtime.Scheme
}

func NewUIServiceBuilder(instance *v1beta1.TemporalCluster, scheme *runtime.Scheme) *UIServiceBuilder {
	return &UIServiceBuilder{
		instance: instance,
		scheme:   scheme,
	}
}

func (b *UIServiceBuilder) Build() (client.Object, error) {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.instance.ChildResourceName("ui"),
			Namespace: b.instance.Namespace,
		},
	}, nil
}

func (b *UIServiceBuilder) Update(object client.Object) error {
	service := object.(*corev1.Service)
	service.Labels = object.GetLabels()
	service.Annotations = object.GetAnnotations()
	service.Spec.Type = corev1.ServiceTypeClusterIP
	service.Spec.Selector = metadata.LabelsSelector(b.instance.Name, "ui")
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
	return nil
}
