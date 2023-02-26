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

package patch

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/alexandrevilain/temporal-operator/pkg/kubernetes"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Helper is a utility for ensuring the proper patching of objects and their status.
// Inspired from: https://github.com/kubernetes-sigs/cluster-api/blob/v1.3.3/util/patch/patch.go
// But with simplified logic.
type Helper struct {
	client       client.Client
	beforeObject client.Object
}

// NewHelper returns a new patch helper.
func NewHelper(obj client.Object, crClient client.Client) (*Helper, error) {
	if kubernetes.IsNil(obj) {
		return nil, errors.New("provided object is nil")
	}

	return &Helper{
		client:       crClient,
		beforeObject: obj.DeepCopyObject().(client.Object),
	}, nil
}

// Patch will attempt to patch the provided resource and its status.
func (h *Helper) Patch(ctx context.Context, afterObject client.Object) error {
	if kubernetes.IsNil(afterObject) {
		return errors.New("provided object is nil")
	}

	before, beforeStatus, err := splitObjectAndStatus(h.beforeObject)
	if err != nil {
		return err
	}

	after, afterStatus, err := splitObjectAndStatus(afterObject)
	if err != nil {
		return err
	}

	var errs []error

	if !reflect.DeepEqual(before, after) {
		if err := h.client.Patch(ctx, afterObject, client.MergeFrom(h.beforeObject)); err != nil {
			errs = append(errs, fmt.Errorf("unable to patch object: %w", err))
		}
	}

	if beforeStatus != nil && afterStatus != nil && !reflect.DeepEqual(beforeStatus, afterStatus) {
		if err := h.client.Status().Patch(ctx, afterObject, client.MergeFrom(h.beforeObject)); err != nil {
			errs = append(errs, fmt.Errorf("unable to patch object status: %w", err))
		}
	}

	return kerrors.NewAggregate(errs)
}

// splitObjectAndStatus converts provided objects to unstructured object, and remove its status.
// It returns the object without its status, the status and an error if something went wrong.
func splitObjectAndStatus(obj client.Object) (map[string]interface{}, interface{}, error) {
	unstructedObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, nil, err
	}

	unstructedObjStatus, statusFound, err := unstructured.NestedFieldCopy(unstructedObj, "status")
	if err != nil {
		return nil, nil, err
	}

	if statusFound {
		unstructured.RemoveNestedField(unstructedObj, "status")
	}

	return unstructedObj, unstructedObjStatus, nil
}
