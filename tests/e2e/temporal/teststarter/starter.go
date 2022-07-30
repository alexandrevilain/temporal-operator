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

package teststarter

import (
	"context"

	"github.com/alexandrevilain/temporal-operator/tests/e2e/temporal/testworker"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

type Starter struct {
	client client.Client
}

func NewStarter(client client.Client) *Starter {
	return &Starter{
		client: client,
	}
}

func (s *Starter) StartGreetingWorkflow() error {
	workflowOptions := client.StartWorkflowOptions{
		ID:        "greetings_" + uuid.NewString(),
		TaskQueue: testworker.Taskqueue,
	}

	we, err := s.client.ExecuteWorkflow(context.Background(), workflowOptions, testworker.GreetingSample)
	if err != nil {
		return err
	}
	var result string
	return we.Get(context.Background(), &result)
}
