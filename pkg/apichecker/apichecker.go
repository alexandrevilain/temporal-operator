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

package apichecker

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// APIChecker checks for needed resources in the cluster.
// Under-the-hood is uses a dry-run client to check if CRDs are available.
// APIChecker is inspired by cert-manager's apichecker.
// (https://github.com/cert-manager/cert-manager/blob/master/pkg/util/cmapichecker).
type APIChecker struct {
	client      client.Client
	fakeObjects []client.Object
}

// NewAPIChecker creates a new ApiChecker.
func NewAPIChecker(restcfg *rest.Config, scheme *runtime.Scheme, namespace string, fakeObjects []client.Object) (*APIChecker, error) {
	cl, err := client.New(restcfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("can't create client: %w", err)
	}
	return &APIChecker{
		client:      client.NewNamespacedClient(client.NewDryRunClient(cl), namespace),
		fakeObjects: fakeObjects,
	}, nil
}

// Check attempts to perform a dry-run create of a needed istio resources.
func (c *APIChecker) Check(ctx context.Context) error {
	for _, fakeObject := range c.fakeObjects {
		err := c.client.Create(ctx, fakeObject)
		if err != nil {
			return err
		}
	}
	return nil
}
