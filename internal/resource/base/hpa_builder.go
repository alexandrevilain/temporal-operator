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
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ resource.Builder = (*HPABuilder)(nil)

type HPABuilder struct {
	serviceName string
	instance    *v1beta1.TemporalCluster
	scheme      *runtime.Scheme
	service     *v1beta1.ServiceSpec
}

func NewHPABuilder(serviceName string, instance *v1beta1.TemporalCluster, scheme *runtime.Scheme, service *v1beta1.ServiceSpec) *HPABuilder {
	return &HPABuilder{
		serviceName: serviceName,
		instance:    instance,
		scheme:      scheme,
		service:     service,
	}
}

func (b *HPABuilder) Build() client.Object {
	return &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName(b.serviceName),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance, b.serviceName, b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}
}

func (b *HPABuilder) Enabled() bool {
	return isBuilderEnabled(b.instance, b.serviceName) && b.service.IsAutoscalingEnabled()
}

func (b *HPABuilder) Update(object client.Object) error {
	hpa := object.(*autoscalingv2.HorizontalPodAutoscaler)
	hpa.Labels = metadata.Merge(
		object.GetLabels(),
		metadata.GetLabels(b.instance, b.serviceName, b.instance.Spec.Version, b.instance.Labels),
	)
	hpa.Annotations = metadata.Merge(
		object.GetAnnotations(),
		metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
	)

	autoscaling := b.service.Autoscaling

	hpa.Spec.ScaleTargetRef = autoscalingv2.CrossVersionObjectReference{
		Kind:       "Deployment",
		Name:       b.instance.ChildResourceName(b.serviceName),
		APIVersion: "apps/v1",
	}

	// Set min/max replicas with defaults
	if autoscaling.MinReplicas != nil {
		hpa.Spec.MinReplicas = autoscaling.MinReplicas
	} else {
		hpa.Spec.MinReplicas = ptr.To[int32](1)
	}

	if autoscaling.MaxReplicas != nil {
		hpa.Spec.MaxReplicas = *autoscaling.MaxReplicas
	} else {
		hpa.Spec.MaxReplicas = 10 // Default max replicas
	}

	// Set metrics
	if len(autoscaling.Metrics) > 0 {
		hpa.Spec.Metrics = autoscaling.Metrics
	} else {
		// Default metrics: 80% CPU and 70% memory utilization (industry standards)
		hpa.Spec.Metrics = []autoscalingv2.MetricSpec{
			{
				Type: autoscalingv2.ResourceMetricSourceType,
				Resource: &autoscalingv2.ResourceMetricSource{
					Name: corev1.ResourceCPU,
					Target: autoscalingv2.MetricTarget{
						Type:               autoscalingv2.UtilizationMetricType,
						AverageUtilization: ptr.To[int32](80),
					},
				},
			},
			{
				Type: autoscalingv2.ResourceMetricSourceType,
				Resource: &autoscalingv2.ResourceMetricSource{
					Name: corev1.ResourceMemory,
					Target: autoscalingv2.MetricTarget{
						Type:               autoscalingv2.UtilizationMetricType,
						AverageUtilization: ptr.To[int32](70),
					},
				},
			},
		}
	}

	// Set behavior if specified
	if autoscaling.Behavior != nil {
		hpa.Spec.Behavior = autoscaling.Behavior
	}

	if err := controllerutil.SetControllerReference(b.instance, hpa, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %w", err)
	}

	return nil
}
