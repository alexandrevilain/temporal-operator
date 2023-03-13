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

package cluster

import (
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	"github.com/alexandrevilain/temporal-operator/pkg/resource"
	"github.com/alexandrevilain/temporal-operator/pkg/temporal/config"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type DynamicConfigmapBuilder struct {
	instance *v1beta1.TemporalCluster
	scheme   *runtime.Scheme
}

func NewDynamicConfigmapBuilder(instance *v1beta1.TemporalCluster, scheme *runtime.Scheme) resource.Builder {
	return &DynamicConfigmapBuilder{
		instance: instance,
		scheme:   scheme,
	}
}

func (b *DynamicConfigmapBuilder) Build() client.Object {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName(ServiceDynamicConfig),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetLabels(b.instance.Name, ServiceDynamicConfig, b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}
}

func (b *DynamicConfigmapBuilder) Update(object client.Object) error {
	configMap := object.(*corev1.ConfigMap)

	currentValues := config.YamlDynamicConfig{}
	expectedValues, err := config.DynamicConfigToYamlDynamicConfig(b.instance.Spec.DynamicConfig)
	if err != nil {
		return fmt.Errorf("failed computing expected dynamic config: %w", err)
	}

	currentContent, ok := configMap.Data["dynamic_config.yaml"]
	if ok {
		err := yaml.Unmarshal([]byte(currentContent), &currentValues)
		if err != nil {
			return fmt.Errorf("failed unmarshaling current dynamic config: %w", err)
		}
	}

	var result string

	// Perform a deep equal to prevent any useless object updates.
	if equality.Semantic.DeepEqual(currentValues, expectedValues) {
		result = currentContent
	} else {
		expectedContent, err := yaml.Marshal(expectedValues)
		if err != nil {
			return fmt.Errorf("failed marshaling temporal config: %w", err)
		}
		result = string(expectedContent)
	}

	configMap.Data = map[string]string{
		"dynamic_config.yaml": result,
	}

	if err := controllerutil.SetControllerReference(b.instance, configMap, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}
