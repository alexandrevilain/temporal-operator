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

package e2e

import (
	"context"

	"sigs.k8s.io/e2e-framework/support"
	"sigs.k8s.io/e2e-framework/support/kind"
)

// FixedKindProvider patches support/kind.Cluster so that it honors the WithImage option.
// TODO: remove when https://github.com/kubernetes-sigs/e2e-framework/pull/395 or similar is available in a e2e-framework release.
type FixedKindProvider struct {
	*kind.Cluster
	image string
}

func (k *FixedKindProvider) SetDefaults() support.E2EClusterProvider {
	k.Cluster.SetDefaults()
	return k
}

func (k *FixedKindProvider) WithName(name string) support.E2EClusterProvider {
	k.Cluster.WithName(name)
	return k
}

func (k *FixedKindProvider) WithVersion(version string) support.E2EClusterProvider {
	k.Cluster.WithVersion(version)
	return k
}

func (k *FixedKindProvider) WithPath(path string) support.E2EClusterProvider {
	k.Cluster.WithPath(path)
	return k
}

func (k *FixedKindProvider) WithOpts(opts ...support.ClusterOpts) support.E2EClusterProvider {
	k.Cluster.WithOpts(opts...)
	return k
}

// Ensure interface is implemented.
var _ support.E2EClusterProvider = &FixedKindProvider{}

func (k *FixedKindProvider) Create(ctx context.Context, args ...string) (string, error) {
	if k.image != "" {
		args = append(args, "--image", k.image)
	}
	return k.Cluster.Create(ctx, args...)
}
