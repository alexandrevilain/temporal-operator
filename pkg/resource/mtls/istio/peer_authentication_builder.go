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
	istioapisecurityv1beta1 "istio.io/api/security/v1beta1"
	istioapiv1beta1 "istio.io/api/type/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ resource.Builder = (*PeerAuthenticationBuilder)(nil)

type PeerAuthenticationBuilder struct {
	serviceName string
	instance    *v1beta1.TemporalCluster
	scheme      *runtime.Scheme
	service     *v1beta1.ServiceSpec
}

func NewPeerAuthenticationBuilder(serviceName string, instance *v1beta1.TemporalCluster, scheme *runtime.Scheme, service *v1beta1.ServiceSpec) *PeerAuthenticationBuilder {
	return &PeerAuthenticationBuilder{
		serviceName: serviceName,
		instance:    instance,
		scheme:      scheme,
		service:     service,
	}
}

func (b *PeerAuthenticationBuilder) Build() client.Object {
	return &istiosecurityv1beta1.PeerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName(b.serviceName),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance, b.serviceName, b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}
}

func (b *PeerAuthenticationBuilder) Enabled() bool {
	return b.instance.Spec.MTLS != nil && b.instance.Spec.MTLS.Provider == v1beta1.IstioMTLSProvider
}

func (b *PeerAuthenticationBuilder) Update(object client.Object) error {
	pa := object.(*istiosecurityv1beta1.PeerAuthentication)
	pa.Spec = istioapisecurityv1beta1.PeerAuthentication{
		Selector: &istioapiv1beta1.WorkloadSelector{
			MatchLabels: metadata.LabelsSelector(b.instance, b.serviceName),
		},
		Mtls: &istioapisecurityv1beta1.PeerAuthentication_MutualTLS{
			Mode: istioapisecurityv1beta1.PeerAuthentication_MutualTLS_STRICT,
		},
	}

	if err := controllerutil.SetControllerReference(b.instance, pa, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %w", err)
	}

	return nil
}

func (PeerAuthenticationBuilder) Equal(x, y *istioapisecurityv1beta1.PeerAuthentication) bool {
	return proto.Equal(x, y)
}
