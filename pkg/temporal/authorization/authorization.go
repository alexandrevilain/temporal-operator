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

package authorization

import (
	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"go.temporal.io/server/common/config"
)

// ToTemporalAuthorization transforms v1beta1.AuthorizationSpec to temporal's authorization config.
func ToTemporalAuthorization(authorization *v1beta1.AuthorizationSpec) config.Authorization {
	if authorization == nil {
		return config.Authorization{}
	}

	return config.Authorization{
		JWTKeyProvider: config.JWTKeyProvider{
			KeySourceURIs:   authorization.JWTKeyProvider.KeySourceURIs,
			RefreshInterval: authorization.JWTKeyProvider.RefreshInterval.Duration,
		},
		PermissionsClaimName: authorization.PermissionsClaimName,
		Authorizer:           authorization.Authorizer,
		ClaimMapper:          authorization.ClaimMapper,
	}
}
