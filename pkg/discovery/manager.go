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

package discovery

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type (
	// Manager provides resources discovery features to know if
	// objects are supported by the cluster its connected to.
	Manager interface {
		IsGVKSupported(gvk schema.GroupVersionKind) (bool, error)
		IsObjectSupported(obj client.Object) (bool, error)
		AreObjectsSupported(objs ...client.Object) (bool, error)
	}

	manager struct {
		scheme *runtime.Scheme
		client *discovery.DiscoveryClient
		cache  *cache
	}
)

// NewManager creates a new instance of a discovery manager.
func NewManager(config *rest.Config, scheme *runtime.Scheme) (Manager, error) {
	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("can't create discovery client: %w", err)
	}

	return &manager{
		client: client,
		scheme: scheme,
		cache:  newCache(),
	}, nil
}

// IsGVKSupported returns true if the provided GVK is supported by the cluster.
func (m *manager) IsGVKSupported(gvk schema.GroupVersionKind) (bool, error) {
	if value, found := m.cache.Get(gvk); found {
		return value, nil
	}

	gvr, _ := apimeta.UnsafeGuessKindToResource(gvk)

	return m.isGVRSupported(gvr)
}

// IsObjectSupported returns true if the provided object is supported by the cluster.
func (m *manager) IsObjectSupported(obj client.Object) (bool, error) {
	gvk, err := apiutil.GVKForObject(obj, m.scheme)
	if err != nil {
		return false, err
	}

	return m.IsGVKSupported(gvk)
}

// AreObjectsSupported returns true if all provided objects are supported by the cluster.
func (m *manager) AreObjectsSupported(objs ...client.Object) (bool, error) {
	for _, obj := range objs {
		supported, err := m.IsObjectSupported(obj)
		if err != nil {
			return supported, err
		}
		if !supported {
			return false, nil
		}
	}

	return true, nil
}

// isGVRSupported checks if given groupVersion and resource is supported by the cluster.
func (m *manager) isGVRSupported(gvr schema.GroupVersionResource) (bool, error) {
	apiResourceList, err := m.client.ServerResourcesForGroupVersion(gvr.GroupVersion().String())
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	for _, apiResource := range apiResourceList.APIResources {
		if gvr.Resource == apiResource.Name {
			return true, nil
		}
	}
	return false, nil
}
