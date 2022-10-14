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

package resource

import (
	"context"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Service components.
const (
	FrontendService = "frontend"
	ServiceConfig   = "config"
)

// Additionals services.
const (
	ServiceUIName     = "ui"
	ServiceAdminTools = "admintools"
)

type Builder interface {
	Build() (client.Object, error)
	Update(client.Object) error
}

type Pruner interface {
	Build() (client.Object, error)
}

type StatusReporter interface {
	ReportServiceStatus(context.Context, client.Client) (*v1beta1.ServiceStatus, error)
}

type WorkerProcessDeploymentReporter interface {
	ReportWorkerDeploymentStatus(context.Context, client.Client) (bool, error)
}

// A Comparer provides a custom function to compare two resources returned
// by a Builder.
type Comparer interface {
	Equal()
}
