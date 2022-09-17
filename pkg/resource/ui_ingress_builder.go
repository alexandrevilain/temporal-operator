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
	"fmt"
	"net/url"
	"strings"

	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type UIIngressBuilder struct {
	instance *v1alpha1.TemporalCluster
	scheme   *runtime.Scheme
}

func NewUIIngressBuilder(instance *v1alpha1.TemporalCluster, scheme *runtime.Scheme) *UIIngressBuilder {
	return &UIIngressBuilder{
		instance: instance,
		scheme:   scheme,
	}
}

func (b *UIIngressBuilder) Build() (client.Object, error) {
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName("ui"),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance.Name, "ui", b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}, nil
}

// parseHost parses the provided ingress host.
// It parses the path, but it's useless for now has the UI does not support another path than "/".
func (b *UIIngressBuilder) parseHost(host string) *url.URL {
	result := &url.URL{}
	parts := strings.Split(host, "/")
	if len(parts) == 0 {
		return result
	}
	if len(parts) >= 1 {
		result.Host = parts[0]
		result.Path = "/" + strings.Join(parts[1:], "/")
	}
	return result
}

func (b *UIIngressBuilder) Update(object client.Object) error {
	ingress := object.(*networkingv1.Ingress)
	ingress.Labels = object.GetLabels()
	ingress.Annotations = metadata.Merge(object.GetAnnotations(), b.instance.Spec.UI.Ingress.Annotations)

	rules := make([]networkingv1.IngressRule, 0, len(b.instance.Spec.UI.Ingress.Hosts))

	for _, host := range b.instance.Spec.UI.Ingress.Hosts {
		parsedURL := b.parseHost(host)
		pathType := networkingv1.PathTypePrefix
		rules = append(rules, networkingv1.IngressRule{
			Host: parsedURL.Host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{
							Path:     "/", // Note that the generated source code for the UI uses hardcoded "/".
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: b.instance.ChildResourceName("ui"),
									Port: networkingv1.ServiceBackendPort{
										Number: UIServicePort,
									},
								},
							},
						},
					},
				},
			},
		})
	}

	ingress.Spec = networkingv1.IngressSpec{
		IngressClassName: b.instance.Spec.UI.Ingress.IngressClassName,
		Rules:            rules,
		TLS:              b.instance.Spec.UI.Ingress.TLS,
	}

	if err := controllerutil.SetControllerReference(b.instance, ingress, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}
	return nil
}
