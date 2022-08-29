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

package debug

import (
	"fmt"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type DebugUpdatePredicate struct {
	log logr.Logger
	predicate.Funcs
}

func NewDebugUpdatePredicate() *DebugUpdatePredicate {
	return &DebugUpdatePredicate{log: log.Log.WithName("predicates").WithName("DebugUpdate")}
}

// Update implements default UpdateEvent filter for validating generation change.
func (d *DebugUpdatePredicate) Update(e event.UpdateEvent) bool {
	obj := fmt.Sprintf("%s/%s", e.ObjectNew.GetNamespace(), e.ObjectNew.GetName())
	diff, err := client.MergeFrom(e.ObjectOld).Data(e.ObjectNew)
	t := fmt.Sprintf("%T", e.ObjectNew.(metav1.Object))
	if err != nil {
		d.log.Info("error generating diff", "err", err, "obj", obj, "type", t)
	} else {
		d.log.Info("Update diff", "diff", string(diff), "obj", obj, "type", t)
	}
	return true
}
