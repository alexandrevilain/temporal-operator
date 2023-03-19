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

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type GenericCAIssuerBuilder struct {
	instance *v1beta1.TemporalCluster
	scheme   *runtime.Scheme

	name       string
	secretName string
}

func (b *GenericCAIssuerBuilder) Build() client.Object {
	return &certmanagerv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.instance.ChildResourceName(b.name),
			Namespace: b.instance.Namespace,
		},
	}
}

func (b *GenericCAIssuerBuilder) Update(object client.Object) error {
	issuer := object.(*certmanagerv1.Issuer)
	issuer.Labels = object.GetLabels()
	issuer.Annotations = object.GetAnnotations()
	issuer.Spec.CA = &certmanagerv1.CAIssuer{
		SecretName: b.instance.ChildResourceName(b.secretName),
	}

	if err := controllerutil.SetControllerReference(b.instance, issuer, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}
	return nil
}
