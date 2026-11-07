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

// The failover version increment is 10, so a cluster's initial failover version is 1 to 9.
var _ = Describe("Cluster metadata", func() {
	newCluster := func(name string, clusterMetadata map[string]any) *unstructured.Unstructured {
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
			"clusterMetadata": clusterMetadata,
		}
		return u
	}

	DescribeTable("spec.clusterMetadata",
		func(name string, clusterMetadata map[string]any, invalidField string) {
			err := k8sClient.Create(context.Background(), newCluster(name, clusterMetadata))
			if invalidField == "" {
				Expect(err).NotTo(HaveOccurred())
				return
			}
			Expect(apierrors.IsInvalid(err)).To(BeTrue(), "%v", err)
			Expect(err.Error()).To(ContainSubstring(invalidField))
		},
		Entry("empty", "md-empty", map[string]any{}, ""),
		Entry("lowest version", "md-version-1", map[string]any{"enableGlobalNamespace": true, "initialFailoverVersion": int64(1)}, ""),
		Entry("highest version", "md-version-9", map[string]any{"enableGlobalNamespace": true, "initialFailoverVersion": int64(9)}, ""),
		Entry("redirection policy", "md-policy", map[string]any{"dcRedirectionPolicy": "all-apis-forwarding"}, ""),
		Entry("version 0", "md-version-0", map[string]any{"initialFailoverVersion": int64(0)}, "spec.clusterMetadata.initialFailoverVersion"),
		Entry("version 10", "md-version-10", map[string]any{"initialFailoverVersion": int64(10)}, "spec.clusterMetadata.initialFailoverVersion"),
		Entry("unknown policy", "md-policy-bad", map[string]any{"dcRedirectionPolicy": "forward-everything"}, "spec.clusterMetadata.dcRedirectionPolicy"),
	)
})
