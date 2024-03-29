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

package admintools

import (
	"github.com/alexandrevilain/controller-tools/pkg/resource"
	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/resource/mtls/certmanager"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ resource.Builder = (*FrontendClientCertificateBuilder)(nil)

type FrontendClientCertificateBuilder struct {
	instance *v1beta1.TemporalCluster

	*certmanager.GenericFrontendClientCertificateBuilder
}

func NewFrontendClientCertificateBuilder(instance *v1beta1.TemporalCluster, scheme *runtime.Scheme) *FrontendClientCertificateBuilder {
	return &FrontendClientCertificateBuilder{
		instance:                                instance,
		GenericFrontendClientCertificateBuilder: certmanager.NewGenericFrontendClientCertificateBuilder(instance, scheme, "admintools"),
	}
}

func (b *FrontendClientCertificateBuilder) Enabled() bool {
	return b.instance.Spec.AdminTools != nil &&
		b.instance.Spec.AdminTools.Enabled &&
		b.instance.MTLSWithCertManagerEnabled() &&
		b.instance.Spec.MTLS.FrontendEnabled()
}
