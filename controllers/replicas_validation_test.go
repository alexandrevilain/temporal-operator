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

package controllers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// 0 replicas stops a service without deleting the cluster; a negative count is
// still rejected by the API server.
var _ = Describe("Service replicas", func() {
	newCluster := func(name string, replicas int64) *unstructured.Unstructured {
		u := &unstructured.Unstructured{}
		u.SetAPIVersion("temporal.io/v1beta1")
		u.SetKind("TemporalCluster")
		u.SetNamespace("default")
		u.SetName(name)
		u.Object["spec"] = map[string]any{
			"numHistoryShards": int64(1),
			"persistence": map[string]any{
				"defaultStore":    map[string]any{},
				"visibilityStore": map[string]any{},
			},
			"services": map[string]any{
				"history": map[string]any{"replicas": replicas},
			},
		}
		return u
	}

	DescribeTable("spec.services.history.replicas",
		func(name string, replicas int64, valid bool) {
			err := k8sClient.Create(context.Background(), newCluster(name, replicas))
			if valid {
				Expect(err).NotTo(HaveOccurred())
				return
			}
			Expect(apierrors.IsInvalid(err)).To(BeTrue(), "%v", err)
			Expect(err.Error()).To(ContainSubstring("spec.services.history.replicas"))
		},
		Entry("one", "replicas-one", int64(1), true),
		Entry("zero", "replicas-zero", int64(0), true),
		Entry("negative", "replicas-negative", int64(-1), false),
	)
})
