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

package testworker

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const Taskqueue = "greetings"

type Worker struct {
	client client.Client
	worker worker.Worker
}

func NewWorker(client client.Client) (*Worker, error) {
	w := &Worker{
		client: client,
	}

	w.worker = worker.New(w.client, Taskqueue, worker.Options{})
	w.worker.RegisterWorkflow(GreetingSample)
	w.worker.RegisterActivity(&Activities{
		Name:     "Temporal",
		Greeting: "Hello",
	})
	return w, nil
}

func (w *Worker) Start() error {
	return w.worker.Start()
}

func (w *Worker) Stop() {
	w.client.Close()
	w.worker.Stop()
}
