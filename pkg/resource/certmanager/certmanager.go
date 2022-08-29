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

import certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"

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
