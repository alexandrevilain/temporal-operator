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

package workerprocess

import (
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ClusterClientBuilder struct {
	instance *v1beta1.TemporalWorkerProcess
	cluster  *v1beta1.TemporalCluster
	scheme   *runtime.Scheme
}

func NewClusterClientBuilder(instance *v1beta1.TemporalWorkerProcess, cluster *v1beta1.TemporalCluster, scheme *runtime.Scheme) *ClusterClientBuilder {
	return &ClusterClientBuilder{
		instance: instance,
		cluster:  cluster,
		scheme:   scheme,
	}
}

func (b *ClusterClientBuilder) Build() (client.Object, error) {
	return &v1beta1.TemporalClusterClient{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName("cluster-client"),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetVersionStringLabels(b.instance.Name, "cluster-client", b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}, nil
}

func (b *ClusterClientBuilder) Update(object client.Object) error {
	clusterClient := object.(*v1beta1.TemporalClusterClient)

	clusterClient.Labels = metadata.Merge(
		object.GetLabels(),
		metadata.GetVersionStringLabels(b.instance.Name, "cluster-client", b.instance.Spec.Version, b.instance.Labels),
	)

	clusterClient.Spec.ClusterRef = v1beta1.TemporalClusterReference{
		Name:      b.cluster.GetName(),
		Namespace: b.cluster.GetNamespace(),
	}

	if err := controllerutil.SetControllerReference(b.instance, clusterClient, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %w", err)
	}

	return nil
}
