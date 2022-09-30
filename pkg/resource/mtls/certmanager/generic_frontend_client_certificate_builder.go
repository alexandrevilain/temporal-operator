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
	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagermeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GenericFrontendClientCertificateBuilder struct {
	instance *v1beta1.TemporalCluster
	scheme   *runtime.Scheme
	// name defines the name of the certificate
	name string
	// secretName is the name of the secrets holding the certificate
	secretName string
	// dnsName is the dns alt name of the certificate
	dnsName string
	// commonName is the common name to be used on the Certificate
	commonName string
}

func NewGenericFrontendClientCertificateBuilder(instance *v1beta1.TemporalCluster, scheme *runtime.Scheme, name string, secretName string, dnsName string, commonName string) *GenericFrontendClientCertificateBuilder {
	return &GenericFrontendClientCertificateBuilder{
		instance:   instance,
		scheme:     scheme,
		name:       name,
		secretName: secretName,
		dnsName:    dnsName,
		commonName: commonName,
	}
}

func (b *GenericFrontendClientCertificateBuilder) Build() (client.Object, error) {
	return &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.instance.ChildResourceName(b.name),
			Namespace: b.instance.Namespace,
		},
	}, nil
}

func (b *GenericFrontendClientCertificateBuilder) Update(object client.Object) error {
	certificate := object.(*certmanagerv1.Certificate)
	certificate.Labels = object.GetLabels()
	certificate.Annotations = object.GetAnnotations()
	certificate.Spec = certmanagerv1.CertificateSpec{
		SecretName: b.instance.ChildResourceName(b.secretName),
		CommonName: b.commonName,
		Duration:   b.instance.Spec.MTLS.CertificatesDuration.ClientCertificates,
		PrivateKey: &certmanagerv1.CertificatePrivateKey{
			RotationPolicy: certmanagerv1.RotationPolicyAlways,
			Encoding:       certmanagerv1.PKCS8,
			Algorithm:      certmanagerv1.RSAKeyAlgorithm,
			Size:           4096,
		},
		DNSNames: []string{
			b.dnsName,
		},
		IssuerRef: certmanagermeta.ObjectReference{
			Name: b.instance.ChildResourceName(frontendIntermediateCAIssuer),
			Kind: certmanagerv1.IssuerKind,
		},
	}

	return nil
}
