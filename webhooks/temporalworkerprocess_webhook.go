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

package webhooks

import (
	"context"
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/discovery"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TemporalWorkerProcessWebhook provides endpoints to validate
// and set default fields values for TemporalWorkerProcess objects.
type TemporalWorkerProcessWebhook struct {
	AvailableAPIs *discovery.AvailableAPIs
	client.Client
}

func (w *TemporalWorkerProcessWebhook) getWorkerProcessFromRequest(obj runtime.Object) (*v1beta1.TemporalWorkerProcess, error) {
	wp, ok := obj.(*v1beta1.TemporalWorkerProcess)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected an TemporalWorkerProcess but got a %T", obj))
	}
	return wp, nil
}

func (w *TemporalWorkerProcessWebhook) getReferencedCluster(ctx context.Context, wp *v1beta1.TemporalWorkerProcess) (*v1beta1.TemporalCluster, error) {
	cluster := &v1beta1.TemporalCluster{}
	err := w.Get(ctx, wp.Spec.ClusterRef.NamespacedName(wp), cluster)
	if err != nil {
		return nil, fmt.Errorf("find referenced temporal cluster: %w", err)
	}

	return cluster, nil
}

func (w *TemporalWorkerProcessWebhook) aggregateWorkerProcessErrors(wp *v1beta1.TemporalWorkerProcess, errs field.ErrorList) error {
	if len(errs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		wp.GroupVersionKind().GroupKind(),
		wp.GetName(),
		errs,
	)
}

// Default ensures empty fields have their default value.
func (w *TemporalWorkerProcessWebhook) Default(ctx context.Context, obj runtime.Object) error {
	wp, err := w.getWorkerProcessFromRequest(obj)
	if err != nil {
		return err
	}

	wp.Default()

	return nil
}

func (w *TemporalWorkerProcessWebhook) validateWorkerProcess(workerprocess *v1beta1.TemporalWorkerProcess, cluster *v1beta1.TemporalCluster) field.ErrorList {
	var errs field.ErrorList

	return errs
}

// ValidateCreate ensures the user is creating a consistent temporal cluster.
func (w *TemporalWorkerProcessWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	workerprocess, err := w.getWorkerProcessFromRequest(obj)
	if err != nil {
		return err
	}

	cluster, err := w.getReferencedCluster(ctx, workerprocess)
	if err != nil {
		return err
	}

	errs := w.validateWorkerProcess(workerprocess, cluster)

	return w.aggregateWorkerProcessErrors(workerprocess, errs)
}

// ValidateUpdate validates TemporalWorkerProcess updates.
// It mainly check for sequential version upgrades.
func (w *TemporalWorkerProcessWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	oldWorkerProcess, err := w.getWorkerProcessFromRequest(oldObj)
	if err != nil {
		return err
	}

	newWorkerProcess, err := w.getWorkerProcessFromRequest(newObj)
	if err != nil {
		return err
	}

	if oldWorkerProcess.Spec.ClusterRef.NamespacedName(oldWorkerProcess) != newWorkerProcess.Spec.ClusterRef.NamespacedName(newWorkerProcess) {
		return w.aggregateWorkerProcessErrors(newWorkerProcess, field.ErrorList{
			field.Forbidden(
				field.NewPath("spec", "clusterRef"),
				"ClusterRef is immutable",
			),
		})
	}

	cluster, err := w.getReferencedCluster(ctx, newWorkerProcess)
	if err != nil {
		return err
	}

	errs := w.validateWorkerProcess(newWorkerProcess, cluster)

	return w.aggregateWorkerProcessErrors(newWorkerProcess, errs)
}

// ValidateDelete does nothing.
func (w *TemporalWorkerProcessWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	// No delete validation needed.
	return nil
}

func (w *TemporalWorkerProcessWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1beta1.TemporalWorkerProcess{}).
		WithDefaulter(w).
		WithValidator(w).
		Complete()
}
