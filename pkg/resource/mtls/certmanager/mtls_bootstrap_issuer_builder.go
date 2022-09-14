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

package certmanager

import (
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type MTLSBootstrapIssuerBuilder struct {
	instance *v1alpha1.TemporalCluster
	scheme   *runtime.Scheme
}

func NewMTLSBootstrapIssuerBuilder(instance *v1alpha1.TemporalCluster, scheme *runtime.Scheme) *MTLSBootstrapIssuerBuilder {
	return &MTLSBootstrapIssuerBuilder{
		instance: instance,
		scheme:   scheme,
	}
}

func (b *MTLSBootstrapIssuerBuilder) Build() (client.Object, error) {
	return &certmanagerv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.instance.ChildResourceName("bootstrap-issuer"),
			Namespace: b.instance.Namespace,
		},
	}, nil
}

func (b *MTLSBootstrapIssuerBuilder) Update(object client.Object) error {
	issuer := object.(*certmanagerv1.Issuer)
	issuer.Labels = object.GetLabels()
	issuer.Annotations = object.GetAnnotations()
	issuer.Spec.SelfSigned = &certmanagerv1.SelfSignedIssuer{}

	if err := controllerutil.SetControllerReference(b.instance, issuer, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}
	return nil
}
