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

	"github.com/alexandrevilain/controller-tools/pkg/resource"
	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ resource.Builder = (*ServiceAccountBuilder)(nil)

type ServiceAccountBuilder struct {
	serviceName string
	instance    *v1beta1.TemporalCluster
	scheme      *runtime.Scheme
	service     *v1beta1.ServiceSpec
}

func NewServiceAccountBuilder(serviceName string, instance *v1beta1.TemporalCluster, scheme *runtime.Scheme, service *v1beta1.ServiceSpec) *ServiceAccountBuilder {
	return &ServiceAccountBuilder{
		serviceName: serviceName,
		instance:    instance,
		scheme:      scheme,
		service:     service,
	}
}

func (b *ServiceAccountBuilder) Build() client.Object {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName(b.serviceName),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance, b.serviceName, b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}
}

func (b *ServiceAccountBuilder) Enabled() bool {
	return isBuilderEnabled(b.instance, b.serviceName)
}

func (b *ServiceAccountBuilder) getArchivalAnnotations() map[string]string {
	if b.instance.Spec.Archival.IsEnabled() &&
		b.instance.Spec.Archival.Provider.S3 != nil &&
		b.instance.Spec.Archival.Provider.S3.RoleName != nil {
		return map[string]string{
			"eks.amazonaws.com/role-arn": *b.instance.Spec.Archival.Provider.S3.RoleName,
		}
	}

	return map[string]string{}
}

func (b *ServiceAccountBuilder) Update(object client.Object) error {
	sa := object.(*corev1.ServiceAccount)

	sa.Annotations = metadata.Merge(
		sa.Annotations,
		b.getArchivalAnnotations(),
	)

	if err := controllerutil.SetControllerReference(b.instance, sa, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %w", err)
	}

	return nil
}
