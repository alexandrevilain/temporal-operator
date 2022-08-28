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
	certmanagermeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type MTLSFrontendItermediateCACertificateBuilder struct {
	instance *v1alpha1.TemporalCluster
	scheme   *runtime.Scheme
}

func NewMTLSFrontendIntermediateCACertificateBuilder(instance *v1alpha1.TemporalCluster, scheme *runtime.Scheme) *MTLSFrontendItermediateCACertificateBuilder {
	return &MTLSFrontendItermediateCACertificateBuilder{
		instance: instance,
		scheme:   scheme,
	}
}

func (b *MTLSFrontendItermediateCACertificateBuilder) Build() (client.Object, error) {
	return &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.instance.ChildResourceName("frontend-intermediate-ca-certificate"),
			Namespace: b.instance.Namespace,
		},
	}, nil
}

func (b *MTLSFrontendItermediateCACertificateBuilder) Update(object client.Object) error {
	certificate := object.(*certmanagerv1.Certificate)
	certificate.Labels = object.GetLabels()
	certificate.Annotations = object.GetAnnotations()
	certificate.Spec = certmanagerv1.CertificateSpec{
		IsCA:       true,
		SecretName: b.instance.ChildResourceName("frontend-intermediate-ca-certificate"),
		CommonName: "Frontend intermediate CA certificate",
		Duration:   b.instance.Spec.MTLS.CertificatesDuration.IntermediateCAsCertificates,
		PrivateKey: &certmanagerv1.CertificatePrivateKey{
			RotationPolicy: certmanagerv1.RotationPolicyAlways,
			Encoding:       certmanagerv1.PKCS8,
			Algorithm:      certmanagerv1.RSAKeyAlgorithm,
			Size:           4096,
		},
		DNSNames: []string{
			b.instance.ServerName(),
		},
		IssuerRef: certmanagermeta.ObjectReference{
			Name: b.instance.ChildResourceName("root-ca-issuer"),
			Kind: certmanagerv1.IssuerKind,
		},
		Usages: []certmanagerv1.KeyUsage{
			certmanagerv1.UsageDigitalSignature,
			certmanagerv1.UsageCRLSign,
			certmanagerv1.UsageCertSign,
		},
	}

	if err := controllerutil.SetControllerReference(b.instance, certificate, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}
	return nil
}
