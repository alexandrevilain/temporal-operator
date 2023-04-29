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

package istio

import (
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	"github.com/alexandrevilain/temporal-operator/pkg/resource"
	"google.golang.org/protobuf/proto"
	istioapinetworkingv1beta1 "istio.io/api/networking/v1beta1"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ resource.Builder = (*DestinationRuleBuilder)(nil)

type DestinationRuleBuilder struct {
	serviceName string
	instance    *v1beta1.TemporalCluster
	scheme      *runtime.Scheme
	service     *v1beta1.ServiceSpec
}

func NewDestinationRuleBuilder(serviceName string, instance *v1beta1.TemporalCluster, scheme *runtime.Scheme, service *v1beta1.ServiceSpec) *DestinationRuleBuilder {
	return &DestinationRuleBuilder{
		serviceName: serviceName,
		instance:    instance,
		scheme:      scheme,
		service:     service,
	}
}

func (b *DestinationRuleBuilder) Build() client.Object {
	return &istionetworkingv1beta1.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName(b.serviceName),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance, b.serviceName, b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}
}

func (b *DestinationRuleBuilder) Enabled() bool {
	return b.instance.Spec.MTLS != nil && b.instance.Spec.MTLS.Provider == v1beta1.IstioMTLSProvider
}

func (b *DestinationRuleBuilder) Update(object client.Object) error {
	pa := object.(*istionetworkingv1beta1.DestinationRule)
	pa.Spec = istioapinetworkingv1beta1.DestinationRule{
		Host: fmt.Sprintf("%s.%s.svc.cluster.local", b.instance.ChildResourceName(fmt.Sprintf("%s-headless", b.serviceName)), b.instance.Namespace),
		TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
			Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
				Mode: istioapinetworkingv1beta1.ClientTLSSettings_ISTIO_MUTUAL,
			},
		},
	}

	if err := controllerutil.SetControllerReference(b.instance, pa, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %w", err)
	}

	return nil
}

func (DestinationRuleBuilder) Equal(x, y *istioapinetworkingv1beta1.DestinationRule) bool {
	return proto.Equal(x, y)
}
