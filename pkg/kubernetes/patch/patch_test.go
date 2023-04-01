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

package patch_test

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/kubernetes/patch"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestHelperPatch(t *testing.T) {
	tests := map[string]struct {
		object        client.Object
		updateObject  func(client.Object)
		validatePatch func(*testing.T, client.Object)
		expectedErr   string
	}{
		"works with finalizers removed": {
			object: &v1beta1.TemporalCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "test-namespace",
					Finalizers: []string{
						"my.test/finalizer",
					},
				},
			},
			updateObject: func(o client.Object) {
				o.(*v1beta1.TemporalCluster).Finalizers = []string{}
			},
			validatePatch: func(tt *testing.T, o client.Object) {
				assert.Empty(tt, o.GetFinalizers())
			},
		},
		"works with finalizers added": {
			object: &v1beta1.TemporalCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "test-namespace",
				},
			},
			updateObject: func(o client.Object) {
				o.(*v1beta1.TemporalCluster).Finalizers = []string{"my.test/finalizer"}
			},
			validatePatch: func(tt *testing.T, o client.Object) {
				assert.Len(tt, o.GetFinalizers(), 1)
				assert.Equal(tt, o.GetFinalizers(), []string{"my.test/finalizer"})
			},
		},
		"works with only status update": {
			object: &v1beta1.TemporalCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "test-namespace",
				},
			},
			updateObject: func(o client.Object) {
				o.(*v1beta1.TemporalCluster).Status = v1beta1.TemporalClusterStatus{
					Version: "1.20.0",
				}
			},
			validatePatch: func(tt *testing.T, o client.Object) {
				assert.Equal(tt, o.(*v1beta1.TemporalCluster).Status.Version, "1.20.0")
			},
		},
		"works with only spec update": {
			object: &v1beta1.TemporalCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "test-namespace",
				},
			},
			updateObject: func(o client.Object) {
				o.(*v1beta1.TemporalCluster).Spec = v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.20.0"),
				}
			},
			validatePatch: func(tt *testing.T, o client.Object) {
				assert.Equal(tt, o.(*v1beta1.TemporalCluster).Spec.Version.String(), "1.20.0")
			},
		},
		"works with both spec and status update": {
			object: &v1beta1.TemporalCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "test-namespace",
				},
			},
			updateObject: func(o client.Object) {
				o.(*v1beta1.TemporalCluster).Spec = v1beta1.TemporalClusterSpec{
					Version: version.MustNewVersionFromString("1.20.0"),
				}
				o.(*v1beta1.TemporalCluster).Status = v1beta1.TemporalClusterStatus{
					Version: "1.20.0",
				}
			},
			validatePatch: func(tt *testing.T, o client.Object) {
				assert.Equal(tt, o.(*v1beta1.TemporalCluster).Spec.Version.String(), "1.20.0")
				assert.Equal(tt, o.(*v1beta1.TemporalCluster).Status.Version, "1.20.0")
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			scheme := runtime.NewScheme()
			utilruntime.Must(v1beta1.AddToScheme(scheme))

			ctx := context.Background()
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			err := fakeClient.Create(ctx, test.object)
			require.NoError(tt, err)

			h, err := patch.NewHelper(test.object, fakeClient)
			if err != nil {
				require.NoError(tt, err)
			}
			test.updateObject(test.object)

			patchErr := h.Patch(ctx, test.object)
			if test.expectedErr != "" {
				assert.Error(tt, patchErr)
				assert.EqualError(tt, patchErr, test.expectedErr)
			} else {
				after := &v1beta1.TemporalCluster{}
				err = fakeClient.Get(ctx, client.ObjectKeyFromObject(test.object), after)
				require.NoError(tt, err)

				test.validatePatch(tt, after)
			}
		})
	}
}
