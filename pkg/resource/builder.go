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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Service components.
const (
	FrontendService      = "frontend"
	ServiceConfig        = "config"
	ServiceDynamicConfig = "dynamicconfig"
)

// Additionals services.
const (
	ServiceUIName     = "ui"
	ServiceAdminTools = "admintools"
)

type Builder interface {
	Build() client.Object
	Enabled() bool
	Update(client.Object) error
}

type DependentBuilder interface {
	Builder
	Dependencies() []*Dependency
}

type OwnerObject interface {
	client.Object
	SelectorLabels() map[string]string
}

// A Comparer provides a custom function to compare two resources returned
// by a Builder.
type Comparer interface {
	Equal()
}

type Status struct {
	GVK       schema.GroupVersionKind
	Name      string
	Namespace string
	Labels    map[string]string
	Ready     bool
}

type Dependency struct {
	Object    client.Object
	Name      string
	Namespace string
}
