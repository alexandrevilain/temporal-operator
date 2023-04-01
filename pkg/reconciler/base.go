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

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/discovery"
	"github.com/alexandrevilain/temporal-operator/pkg/resource"
	"github.com/alexandrevilain/temporal-operator/pkg/resourceset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Base struct {
	client.Client
	Scheme        *runtime.Scheme
	Recorder      record.EventRecorder
	AvailableAPIs *discovery.AvailableAPIs
}

// LogAndRecordOperationResult logs and records an event for the provided CreateOrUpdate operation result.
func (r *Base) LogAndRecordOperationResult(ctx context.Context, owner, resource runtime.Object, operationResult controllerutil.OperationResult, err error) {
	logger := log.FromContext(ctx)

	var (
		action string
		reason string
	)
	switch operationResult {
	case controllerutil.OperationResultCreated:
		action = "create"
		reason = "RessourceCreate"
	case controllerutil.OperationResultUpdated, controllerutil.OperationResultUpdatedStatus, controllerutil.OperationResultUpdatedStatusOnly:
		action = "update"
		reason = "ResourceUpdate"
	case controllerutil.OperationResult("deleted"):
		action = "delete"
		reason = "ResourceDelete"
	case controllerutil.OperationResultNone:
		fallthrough
	default:
		return
	}

	if err == nil {
		msg := fmt.Sprintf("%sd resource %s of type %T", action, resource.(metav1.Object).GetName(), resource.(metav1.Object))
		reason := fmt.Sprintf("%sSuccess", reason)
		logger.Info(msg)
		r.Recorder.Event(owner, corev1.EventTypeNormal, reason, msg)
	}

	if err != nil {
		msg := fmt.Sprintf("failed to %s resource %s of Type %T", action, resource.(metav1.Object).GetName(), resource.(metav1.Object))
		reason := fmt.Sprintf("%sError", reason)
		logger.Error(err, msg)
		r.Recorder.Event(owner, corev1.EventTypeWarning, reason, msg)
	}
}

func (r *Base) ReconcileResources(ctx context.Context, owner runtime.Object, builder resourceset.Builder) ([]*v1beta1.ServiceStatus, time.Duration, error) {
	logger := log.FromContext(ctx)

	builders, err := builder.ResourceBuilders()
	if err != nil {
		return nil, 0, err
	}

	logger.Info("Retrieved builders", "count", len(builders))

	statuses, requeueAfter, err := r.ReconcileBuilders(ctx, owner, builders)
	if err != nil {
		logger.Error(err, "Can't reconcile builders")
		return nil, requeueAfter, err
	}

	if requeueAfter > 0 {
		return statuses, requeueAfter, nil
	}

	pruners := builder.ResourcePruners()

	if len(pruners) > 0 {
		logger.Info("Retrieved pruners", "count", len(pruners))
	}

	err = r.ReconcilePruners(ctx, owner, pruners)
	if err != nil {
		logger.Error(err, "Can't reconcile pruners")
		return nil, 0, err
	}

	return statuses, 0, nil
}

func (r *Base) ReconcileBuilders(ctx context.Context, owner runtime.Object, builders []resource.Builder) ([]*v1beta1.ServiceStatus, time.Duration, error) {
	logger := log.FromContext(ctx)

	statuses := []*v1beta1.ServiceStatus{}

	for _, builder := range builders {
		if comparer, ok := builder.(resource.Comparer); ok {
			err := equality.Semantic.AddFunc(comparer.Equal)
			if err != nil {
				return statuses, 0, err
			}
		}

		dependentBuilder, hasDependents := builder.(resource.DependentBuilder)
		if hasDependents {
			requeueAfter, err := dependentBuilder.EnsureDependencies(ctx, r.Client)
			if err != nil {
				return statuses, 0, err
			}
			if requeueAfter > 0 {
				return statuses, requeueAfter, err
			}
		}

		res, err := builder.Build()
		if err != nil {
			return statuses, 0, err
		}

		operationResult, err := controllerutil.CreateOrUpdate(ctx, r.Client, res, func() error {
			return builder.Update(res)
		})

		r.LogAndRecordOperationResult(ctx, owner, res, operationResult, err)
		if err != nil {
			return statuses, 0, err
		}

		reporter, ok := builder.(resource.StatusReporter)
		if !ok {
			continue
		}

		serviceStatus, err := reporter.ReportServiceStatus(ctx, r.Client)
		if err != nil {
			return statuses, 0, err
		}

		logger.Info("Reporting service status", "service", serviceStatus.Name)

		statuses = append(statuses, serviceStatus)
	}

	return statuses, 0, nil
}

func (r *Base) ReconcilePruners(ctx context.Context, owner runtime.Object, pruners []resource.Pruner) error {
	for _, pruner := range pruners {
		resource, err := pruner.Build()
		if err != nil {
			return err
		}
		err = r.Delete(ctx, resource)
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
		}
		r.LogAndRecordOperationResult(ctx, owner, resource, controllerutil.OperationResult("deleted"), err)
	}
	return nil
}
