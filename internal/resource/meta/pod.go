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

package meta

import (
	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	"github.com/alexandrevilain/temporal-operator/internal/resource/mtls/istio"
	"github.com/alexandrevilain/temporal-operator/internal/resource/mtls/linkerd"
	"github.com/alexandrevilain/temporal-operator/internal/resource/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	configHashKey = "operator.temporal.io/config"
)

// BuildPodObjectMeta return ObjectMeta for the service (frontend, ui, admintools) of the provided Cluster.
func BuildPodObjectMeta(instance *v1beta1.TemporalCluster, service, configHash string) metav1.ObjectMeta {
	instanceAnnotations := metadata.FilterAnnotations(instance.Annotations, func(k, _ string) bool {
		return k != "kubectl.kubernetes.io/last-applied-configuration"
	})

	return metav1.ObjectMeta{
		Labels: metadata.Merge(
			istio.GetLabels(instance),
			metadata.GetLabels(instance, service, instance.Spec.Version, instance.Labels),
		),
		Annotations: metadata.Merge(
			linkerd.GetAnnotations(instance),
			istio.GetAnnotations(instance),
			prometheus.GetAnnotations(instance),
			metadata.GetAnnotations(instance.Name, instanceAnnotations),
			map[string]string{
				configHashKey: configHash,
			},
		),
	}
}
