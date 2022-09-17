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
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
)

// internal issuers and certificates names.
const (
	bootstrapIssuer               = "bootstrap-issuer"
	frontendIntermediateCAIssuer  = "frontend-intermediate-ca-issuer"
	internodeIntermediateCAIssuer = "internode-intermediate-ca-issuer"

	rootCaIssuer      = "root-ca-issuer"
	rootCaCertificate = "root-ca-certificate"
)

const (
	// InternodeCertificate is the name of the certificate used for internode communications.
	InternodeCertificate = "internode-certificate"
	// FrontendCertificate is the name of the certificate used by the frontend.
	FrontendCertificate = "frontend-certificate"
	// WorkerCertificate is the name of the client certificate
	// used for by the worker for authenticating against the frontend.
	WorkerCertificate = "worker-certificate"

	// InternodeIntermediateCACertificate is the name of the intermediate CA certificate used to issue
	// internode certificates.
	InternodeIntermediateCACertificate = "internode-intermediate-ca-certificate"
	// FrontendIntermediateCACertificate is the name of the intermediate CA certificate used to issue
	// frontend certificates.
	FrontendIntermediateCACertificate = "frontend-intermediate-ca-certificate"

	// AdmintoolsFrontendClientCertificate is the name of the client certificate
	// used for by admin tools for authenticating against the frontend.
	AdmintoolsFrontendClientCertificate = "admintools-mtls-certificate"
	// UIFrontendClientCertificate is the name of the client certificate
	// used for by UI for authenticating against the frontend.
	UIFrontendClientCertificate = "ui-mtls-certificate"
)

const (
	TLSCA   = "ca.crt"
	TLSCert = "tls.crt"
	TLSKey  = "tls.key"
)

var (
	caCertificatesUsages = []certmanagerv1.KeyUsage{
		certmanagerv1.UsageDigitalSignature,
		certmanagerv1.UsageCRLSign,
		certmanagerv1.UsageCertSign,
	}
	caCertificatePrivateKey = &certmanagerv1.CertificatePrivateKey{
		RotationPolicy: certmanagerv1.RotationPolicyAlways,
		Encoding:       certmanagerv1.PKCS8,
		Algorithm:      certmanagerv1.RSAKeyAlgorithm,
		Size:           4096,
	}
)
