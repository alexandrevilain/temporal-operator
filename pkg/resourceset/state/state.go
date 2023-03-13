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

package state

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type Desired struct {
	scheme   *runtime.Scheme
	resoures map[string]map[string]struct{}
}

func NewDesired(scheme *runtime.Scheme) *Desired {
	return &Desired{
		scheme:   scheme,
		resoures: map[string]map[string]struct{}{},
	}
}

func (d *Desired) Add(obj client.Object) error {
	resourceGVK, err := apiutil.GVKForObject(obj, d.scheme)
	if err != nil {
		return fmt.Errorf("can't get object GVK: %w", err)
	}
	_, ok := d.resoures[resourceGVK.String()]
	if !ok {
		d.resoures[resourceGVK.String()] = map[string]struct{}{}
	}
	key := fmt.Sprintf("%s.%s", obj.GetNamespace(), obj.GetName())
	d.resoures[resourceGVK.String()][key] = struct{}{}
	return nil
}

// Has returns whenever the desired state has the provided object.
func (d *Desired) Has(obj client.Object) (bool, error) {
	resourceGVK, err := apiutil.GVKForObject(obj, d.scheme)
	if err != nil {
		return false, fmt.Errorf("can't get object GVK: %w", err)
	}
	gvk, ok := d.resoures[resourceGVK.String()]
	if !ok {
		return false, nil
	}
	key := fmt.Sprintf("%s.%s", obj.GetNamespace(), obj.GetName())
	_, ok = gvk[key]
	return ok, nil
}
