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
	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
)

// WorkerProcessJob contains jobs needed to build a worker process image.
type WorkerProcessJob struct {
	Name          string
	Command       []string
	Skip          func(c *v1beta1.TemporalWorkerProcess) bool
	ReportSuccess func(c *v1beta1.TemporalWorkerProcess) error
}

func GetWorkerProcessJobs() []WorkerProcessJob {
	jobs := []WorkerProcessJob{
		{
			Name:    "build-worker-process",
			Command: []string{"/etc/scripts/build-worker-process.sh"},
			Skip: func(w *v1beta1.TemporalWorkerProcess) bool {
				return w.Status.Created
			},
			ReportSuccess: func(w *v1beta1.TemporalWorkerProcess) error {
				w.Status.Created = true
				return nil
			},
		},
	}

	return jobs
}
