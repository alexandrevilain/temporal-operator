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

package apichecker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alexandrevilain/temporal-operator/pkg/apichecker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type fakeErrorClient struct {
	client.Client

	createError error
}

func (cl *fakeErrorClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if cl.createError != nil {
		return cl.createError
	}

	return cl.Client.Create(ctx, obj, opts...)
}

var (
	errNoKindDestinationrule = `error finding the scope of the object: failed to get restmapping: no matches for kind "Pod" in group ""`
)

func TestApiChecker(t *testing.T) {
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)

	emptyScheme := runtime.NewScheme()

	clientWithScheme := fake.NewClientBuilder().WithScheme(scheme).Build()

	tests := map[string]struct {
		client        client.Client
		objects       []client.Object
		expectedError string
	}{
		"resource not registered in scheme": {
			client: fake.NewClientBuilder().WithScheme(emptyScheme).Build(),
			objects: []client.Object{
				&corev1.Pod{},
			},
			expectedError: "no kind is registered for the type v1.Pod in scheme",
		},
		"api server returning error": {
			client: &fakeErrorClient{
				Client:      clientWithScheme,
				createError: errors.New(errNoKindDestinationrule),
			},
			objects: []client.Object{
				&corev1.Pod{},
			},
			expectedError: errNoKindDestinationrule,
		},
		"scheme and api allowing request": {
			client:        clientWithScheme,
			expectedError: "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			c := apichecker.NewAPICheckerForTesting(test.client, test.objects...)

			err := c.Check(context.Background())
			if test.expectedError != "" {
				require.Error(tt, err)
				assert.Contains(tt, err.Error(), test.expectedError)
			} else {
				assert.NoError(tt, err)
			}
		})
	}
}
