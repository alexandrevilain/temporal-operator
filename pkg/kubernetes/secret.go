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

package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type SecretCopier struct {
	client.Client
	scheme *runtime.Scheme
}

func NewSecretCopier(c client.Client, scheme *runtime.Scheme) *SecretCopier {
	return &SecretCopier{
		Client: c,
		scheme: scheme,
	}
}

func (c *SecretCopier) Copy(ctx context.Context, owner client.Object, original client.ObjectKey, destinationNS string) error {
	secret := &corev1.Secret{}
	err := c.Get(ctx, original, secret)
	if err != nil {
		return fmt.Errorf("can't retrieve original secret: %w", err)
	}

	secretMeta := metav1.ObjectMeta{
		Name:        secret.GetName(),
		Namespace:   destinationNS,
		Labels:      secret.Labels,
		Annotations: secret.Annotations,
	}

	destinationSecret := &corev1.Secret{}
	destinationSecret.ObjectMeta = secretMeta

	_, err = controllerutil.CreateOrUpdate(ctx, c.Client, destinationSecret, func() error {
		destinationSecret.Labels = secretMeta.Labels
		destinationSecret.Annotations = secretMeta.Annotations

		destinationSecret.Data = secret.Data
		destinationSecret.StringData = secret.StringData
		destinationSecret.Immutable = secret.Immutable
		destinationSecret.Type = secret.Type

		err = controllerutil.SetOwnerReference(owner, destinationSecret, c.scheme)
		if err != nil {
			return fmt.Errorf("failed setting controller reference: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("can't create or update destination secret: %w", err)
	}

	return nil
}
