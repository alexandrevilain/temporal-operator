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

package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/alexandrevilain/temporal-operator/pkg/resource"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Job struct {
	Name          string
	Command       []string
	Skip          func(owner runtime.Object) bool
	ReportSuccess func(owner runtime.Object) error
}

type JobBuilderFactory func(owner runtime.Object, scheme *runtime.Scheme, name string, command []string) resource.Builder

type JosbReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (r *JosbReconciler) Reconcile(ctx context.Context, owner client.Object, builderFactory JobBuilderFactory, jobs []*Job) (time.Duration, error) {
	logger := log.FromContext(ctx)

	for _, job := range jobs {
		if job.Skip(owner) {
			continue
		}

		logger.Info("Checking for job", "name", job.Name)

		jobBuilder := builderFactory(owner, r.Scheme, job.Name, job.Command)

		expectedJob := jobBuilder.Build()

		matchingJob := &batchv1.Job{}
		err := r.Client.Get(ctx, types.NamespacedName{Name: expectedJob.GetName(), Namespace: expectedJob.GetNamespace()}, matchingJob)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// The job is not found, create it
				_, err := controllerutil.CreateOrUpdate(ctx, r.Client, expectedJob, func() error {
					return jobBuilder.Update(expectedJob)
				})
				if err != nil {
					return 0, err
				}
			} else {
				return 0, fmt.Errorf("can't get job: %w", err)
			}
		}

		if matchingJob.Status.Succeeded != 1 {
			logger.Info("Waiting for job to complete", "name", job.Name)

			// Requeue after 10 seconds
			return 10 * time.Second, nil
		}

		logger.Info("Job is finished", "name", job.Name)

		err = job.ReportSuccess(owner)
		if err != nil {
			return 0, fmt.Errorf("can't report job success: %w", err)
		}
	}
	return 0, nil
}
