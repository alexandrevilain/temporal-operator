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
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type cache struct {
	sync.RWMutex

	data map[schema.GroupVersionKind]bool
}

func newCache() *cache {
	return &cache{
		data: make(map[schema.GroupVersionKind]bool),
	}
}

func (c *cache) Get(gvk schema.GroupVersionKind) (bool, bool) {
	c.RLock()
	defer c.RUnlock()

	value, found := c.data[gvk]
	return value, found
}

func (c *cache) Set(gvk schema.GroupVersionKind, value bool) {
	c.Lock()
	defer c.Unlock()

	c.data[gvk] = value
}
