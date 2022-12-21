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

package v1beta1

import v1 "k8s.io/api/core/v1"

// Default set default fields values.
func (w *TemporalWorkerProcess) Default() {
	if w.Spec.Builder.BuilderEnabled() {
		if w.Spec.Builder.GitRepository.Reference == nil {
			w.Spec.Builder.GitRepository.Reference = new(GitRepositoryRef)
		}
		if w.Spec.Builder.GitRepository.Reference.Branch == "" {
			w.Spec.Builder.GitRepository.Reference.Branch = "main"
		}
		if w.Spec.PullPolicy == "" {
			w.Spec.PullPolicy = v1.PullAlways
		}
	}
}
