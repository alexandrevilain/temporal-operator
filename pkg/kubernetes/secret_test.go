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

package kubernetes_test

import (
	"context"
	"testing"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/kubernetes"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSecretCopier(t *testing.T) {
	tests := map[string]struct {
		original    client.Object
		owner       client.Object
		destination string
		expected    client.Object
		expectedErr string
	}{
		"works with secret in same namespace than owner": {
			original: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				StringData: map[string]string{
					"test": "test",
				},
			},
			owner: &v1beta1.TemporalCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fakecluster",
					Namespace: "default",
				},
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.20.0"),
				},
			},
			destination: "default",
			expected: &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "temporal.io/v1beta1",
							Kind:       "TemporalCluster",
							Name:       "fakecluster",
						},
					},
					ResourceVersion: "1",
				},
				StringData: map[string]string{
					"test": "test",
				},
			},
		},
		"error with cross namespace owner reference": {
			original: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				StringData: map[string]string{
					"test": "test",
				},
			},
			owner: &v1beta1.TemporalCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fakecluster",
					Namespace: "test",
				},
				Spec: v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.20.0"),
				},
			},
			destination: "default",
			expectedErr: "failed setting controller reference: cross-namespace owner references are disallowed",
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			utilruntime.Must(v1beta1.AddToScheme(scheme))
			utilruntime.Must(corev1.AddToScheme(scheme))

			fakeClient := fake.NewClientBuilder().WithObjects(test.original).WithScheme(scheme).Build()

			copier := kubernetes.NewSecretCopier(fakeClient, fakeClient.Scheme())
			err := copier.Copy(ctx, test.owner, client.ObjectKeyFromObject(test.original), test.destination)
			if test.expectedErr != "" {
				assert.Error(tt, err)
				assert.ErrorContains(tt, err, test.expectedErr)

				return
			}
			assert.NoError(tt, err)
			result := &corev1.Secret{}
			require.NoError(tt, fakeClient.Get(ctx, client.ObjectKey{Name: test.original.GetName(), Namespace: test.destination}, result))

			assert.Equal(tt, test.expected, result)
		})
	}
}
