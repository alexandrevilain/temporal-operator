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
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/alexandrevilain/temporal-operator/pkg/discovery"
	"github.com/alexandrevilain/temporal-operator/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	kstatus "sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type BuildersReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Recorder  record.EventRecorder
	Discovery discovery.Manager
}

type Resource struct {
	builder resource.Builder
	current client.Object
	found   bool
}

// TODO(alexandrevilain): Add support for dependencies.
func (r *BuildersReconciler) Reconcile(ctx context.Context, owner client.Object, builders []resource.Builder) ([]*resource.Status, time.Duration, error) {
	logger := log.FromContext(ctx)

	resources, err := r.getResourcesFromBuilders(ctx, builders)
	if err != nil {
		return nil, 0, err
	}

	// Sort resources by their dependencies. The objective is to run builders without dependencies first
	// as they could be dependent for another builders.
	// We may have the need for a graph, but looks like it would work as expected for now.
	sort.Slice(resources, func(i, j int) bool {
		iDependentBuilder, iHasDependent := resources[i].builder.(resource.DependentBuilder)
		jDependentBuilder, jHasDependent := resources[j].builder.(resource.DependentBuilder)
		// If both have dependences, sort by dependencies length.
		if iHasDependent && jHasDependent {
			return len(iDependentBuilder.Dependencies()) < len(jDependentBuilder.Dependencies())
		}

		// If i has dependencies but not j, i should be after j.
		if iHasDependent && !jHasDependent {
			return false
		}
		// Otherwise i could be before j, it doesn't matter.
		return true
	})

	logger.Info("Reconciling resources", "count", len(resources))

	statuses := []*resource.Status{}

	for _, res := range resources {
		// If the builder isn't enabled, check if it needs to be deleted, then skip iteration.
		if !res.builder.Enabled() {
			if res.found {
				err := r.Client.Delete(ctx, res.current)
				r.logAndRecordOperationResult(ctx, owner, res.current, controllerutil.OperationResult("deleted"), err)
				if err != nil {
					return nil, 0, fmt.Errorf("can't delete resource: %w", err)
				}
			}

			continue
		}

		// The build can provide a custom compare function, ensure equality.Semantic knowns it.
		if comparer, ok := res.builder.(resource.Comparer); ok {
			err := equality.Semantic.AddFunc(comparer.Equal)
			if err != nil {
				return nil, 0, err
			}
		}

		if dependent, ok := res.builder.(resource.DependentBuilder); ok {
			for _, dependency := range dependent.Dependencies() {
				dependency := dependency
				objectKey := client.ObjectKey{
					Name:      dependency.Name,
					Namespace: dependency.Namespace,
				}
				logger.Info("Checking builder dependency",
					"builder", fmt.Sprintf("%T", res.builder),
					"dependency", objectKey,
				)
				err := r.Client.Get(ctx, objectKey, dependency.Object)
				if err != nil {
					if apierrors.IsNotFound(err) {
						return nil, 5 * time.Second, nil
					}
					return nil, 0, err
				}

				status, err := r.getResourceStatus(dependency.Object)
				if err != nil {
					return nil, 0, err
				}

				if !status.Ready {
					return nil, 5 * time.Second, nil
				}
			}
		}

		// Create case
		if !res.found && res.builder.Enabled() {
			res.current = res.builder.Build()
			err := res.builder.Update(res.current)
			if err != nil {
				return nil, 0, err
			}

			err = r.Client.Create(ctx, res.current)
			r.logAndRecordOperationResult(ctx, owner, res.current, controllerutil.OperationResultCreated, err)
			if err != nil {
				return nil, 0, err
			}
		}

		// Update case
		if res.found && res.builder.Enabled() {
			before := res.current.DeepCopyObject()
			err := res.builder.Update(res.current)
			if err != nil {
				return nil, 0, err
			}

			if !equality.Semantic.DeepEqual(before, res.current) {
				err = r.Client.Update(ctx, res.current)
				r.logAndRecordOperationResult(ctx, owner, res.current, controllerutil.OperationResultUpdated, err)
				if err != nil {
					return nil, 0, err
				}
			}
		}

		// Report status
		status, err := r.getResourceStatus(res.current)
		if err != nil {
			return nil, 0, err
		}

		statuses = append(statuses, status)
	}

	return statuses, 0, nil
}

func (r *BuildersReconciler) getResourceStatus(res client.Object) (*resource.Status, error) {
	udeploy, err := runtime.DefaultUnstructuredConverter.ToUnstructured(res)
	if err != nil {
		return nil, err
	}

	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(udeploy)

	status, err := kstatus.Compute(u)
	if err != nil {
		return nil, err
	}

	return &resource.Status{
		GVK:       res.GetObjectKind().GroupVersionKind(),
		Name:      res.GetName(),
		Namespace: res.GetNamespace(),
		Labels:    res.GetLabels(),
		Ready:     status.Status == kstatus.CurrentStatus,
	}, nil
}

func (r *BuildersReconciler) getType(gvk schema.GroupVersionKind) reflect.Type {
	for typename, reflectType := range r.Scheme.KnownTypes(gvk.GroupVersion()) {
		if typename == gvk.Kind {
			return reflectType
		}
	}
	return nil
}

func (r *BuildersReconciler) getResourcesFromBuilders(ctx context.Context, builders []resource.Builder) ([]*Resource, error) {
	logger := log.FromContext(ctx)

	result := []*Resource{}

	for _, builder := range builders {
		res := builder.Build()
		gvk, err := apiutil.GVKForObject(res, r.Scheme)
		if err != nil {
			return nil, err
		}

		objectType := r.getType(gvk)
		if objectType == nil {
			return nil, fmt.Errorf("can't get type for %s", gvk)
		}

		object, ok := reflect.New(objectType).Interface().(client.Object)
		if !ok {
			return nil, errors.New("can't create a new client.Object instance from known type")
		}

		supported, err := r.Discovery.IsGVKSupported(gvk)
		if err != nil {
			return nil, fmt.Errorf("can't determine if GVK \"%s\" is supported: %w", gvk.String(), err)
		}

		if !supported {
			logger.V(2).Info("Skipping resource due to unsupported by apiserver", "kind", gvk.Kind)
			continue
		}

		found := true
		err = r.Client.Get(ctx, client.ObjectKeyFromObject(res), object, &client.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				found = false
			} else {
				return nil, err
			}
		}

		result = append(result, &Resource{
			builder: builder,
			current: object,
			found:   found,
		})
	}

	return result, nil
}

// logAndRecordOperationResult logs and records an event for the provided object operation result.
func (r *BuildersReconciler) logAndRecordOperationResult(ctx context.Context, owner, resource runtime.Object, operationResult controllerutil.OperationResult, err error) {
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
