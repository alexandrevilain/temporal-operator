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

package cluster

import (
	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/pkg/resource"
	"go.temporal.io/server/common"
	"k8s.io/apimachinery/pkg/runtime"
)

type TemporalClusterBuilder struct {
	Instance *v1alpha1.TemporalCluster
	Scheme   *runtime.Scheme
}

func (b *TemporalClusterBuilder) ResourceBuilders() ([]resource.Builder, error) {
	builders := []resource.Builder{
		resource.NewConfigmapBuilder(b.Instance, b.Scheme),
		resource.NewFrontendServiceBuilder(b.Instance, b.Scheme),
	}

	for _, serviceName := range []string{
		common.FrontendServiceName,
		common.HistoryServiceName,
		common.MatchingServiceName,
		common.WorkerServiceName,
	} {
		specs, err := b.Instance.Spec.Services.GetServiceSpec(serviceName)
		if err != nil {
			return nil, err
		}
		builders = append(builders, resource.NewDeploymentBuilder(serviceName, b.Instance, b.Scheme, specs))
		builders = append(builders, resource.NewHeadlessServiceBuilder(serviceName, b.Instance, b.Scheme, specs))
	}

	return builders, nil
}
