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

package controllers

import (
	"github.com/alexandrevilain/controller-tools/pkg/discovery"
	"github.com/alexandrevilain/controller-tools/pkg/reconciler"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Base struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	Jobs       *reconciler.JosbReconciler
	Reconciler *reconciler.Reconciler
}

func New(crclient client.Client, scheme *runtime.Scheme, recorder record.EventRecorder, discoveryMgr discovery.Manager) Base {
	return Base{
		Client:   crclient,
		Scheme:   scheme,
		Recorder: recorder,
		Jobs: &reconciler.JosbReconciler{
			Client:   crclient,
			Scheme:   scheme,
			Recorder: recorder,
		},
		Reconciler: &reconciler.Reconciler{
			Client:    crclient,
			Scheme:    scheme,
			Recorder:  recorder,
			Discovery: discoveryMgr,
		},
	}
}
