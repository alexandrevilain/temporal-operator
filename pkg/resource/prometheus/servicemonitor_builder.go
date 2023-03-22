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

package prometheus

import (
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ServiceMonitorBuilder struct {
	serviceName string
	instance    *v1beta1.TemporalCluster
	scheme      *runtime.Scheme
	service     *v1beta1.ServiceSpec
}

func NewServiceMonitorBuilder(serviceName string, instance *v1beta1.TemporalCluster, scheme *runtime.Scheme, service *v1beta1.ServiceSpec) *ServiceMonitorBuilder {
	return &ServiceMonitorBuilder{
		serviceName: serviceName,
		instance:    instance,
		scheme:      scheme,
		service:     service,
	}
}

func (b *ServiceMonitorBuilder) Build() (client.Object, error) {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName(b.serviceName),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance.Name, b.serviceName, b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}, nil
}

func (b *ServiceMonitorBuilder) applySpecOverride(sm *monitoringv1.ServiceMonitor, specOverride *monitoringv1.ServiceMonitorSpec) error {
	if specOverride == nil {
		return nil
	}

	specOverrideCopy := specOverride.DeepCopy()

	// Clean non-overridable fields.
	specOverrideCopy.Endpoints = []monitoringv1.Endpoint{}
	specOverrideCopy.NamespaceSelector = monitoringv1.NamespaceSelector{}
	specOverrideCopy.Selector = metav1.LabelSelector{}

	patchedSpec, err := PatchServiceMonitorSpecWithOverride(&sm.Spec, specOverrideCopy)
	if err != nil {
		return err
	}
	if patchedSpec != nil {
		sm.Spec = *patchedSpec
	}

	return nil
}

func (b *ServiceMonitorBuilder) Update(object client.Object) error {
	sm := object.(*monitoringv1.ServiceMonitor)

	sm.Spec = monitoringv1.ServiceMonitorSpec{
		NamespaceSelector: monitoringv1.NamespaceSelector{
			MatchNames: []string{
				b.instance.Namespace,
			},
		},
		Selector: metav1.LabelSelector{
			MatchLabels: metadata.Merge(
				metadata.LabelsSelector(b.instance.GetName(), b.serviceName),
				metadata.HeadlessLabels(),
			),
		},
		Endpoints: []monitoringv1.Endpoint{
			{
				TargetPort:           &MetricsPortName,
				MetricRelabelConfigs: b.instance.Spec.Metrics.Prometheus.ScrapeConfig.ServiceMonitor.MetricRelabelConfigs,
			},
		},
	}

	if err := b.applySpecOverride(sm, b.instance.Spec.Metrics.Prometheus.ScrapeConfig.ServiceMonitor.Override); err != nil {
		return fmt.Errorf("failed applying service monitor spec override: %w", err)
	}

	if err := controllerutil.SetControllerReference(b.instance, sm, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %w", err)
	}

	return nil
}
