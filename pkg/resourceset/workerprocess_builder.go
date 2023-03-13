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

package resourceset

import (
	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/resource"
	workerprocessresource "github.com/alexandrevilain/temporal-operator/pkg/resource/workerprocess"
	"k8s.io/apimachinery/pkg/runtime"
)

type WorkerProcessBuilder struct {
	Instance *v1beta1.TemporalWorkerProcess
	Scheme   *runtime.Scheme
	Cluster  *v1beta1.TemporalCluster
}

func (b *WorkerProcessBuilder) ResourceBuilders() ([]resource.Builder, error) {
	builders := []resource.Builder{}

	if b.Cluster.MTLSWithCertManagerEnabled() && b.Cluster.Spec.MTLS.FrontendEnabled() {
		builders = append(builders,
			workerprocessresource.NewClusterClientBuilder(b.Instance, b.Cluster, b.Scheme),
		)
	}

	builders = append(builders, workerprocessresource.NewDeploymentBuilder(b.Instance, b.Cluster, b.Scheme))

	return builders, nil
}

func (b *WorkerProcessBuilder) ResourcePruners() ([]resource.Pruner, error) {
	return []resource.Pruner{}, nil
}
