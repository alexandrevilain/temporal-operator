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
	"encoding/json"
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// PatchPodSpecWithOverride patches the provided pod spec with the provided pod spec override.
func PatchPodSpecWithOverride(spec, override *corev1.PodSpec) (*corev1.PodSpec, error) {
	if override == nil {
		return nil, nil
	}

	orginalSpec, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("can't marshal pod spec: %w", err)
	}

	overrideSpec, err := json.Marshal(override)
	if err != nil {
		return nil, fmt.Errorf("can't marshal pod spec override: %w", err)
	}

	patchedJSON, err := strategicpatch.StrategicMergePatch(orginalSpec, overrideSpec, corev1.PodSpec{})
	if err != nil {
		return nil, fmt.Errorf("can't patch pod spec: %w", err)
	}

	patchedSpec := &corev1.PodSpec{}
	err = json.Unmarshal(patchedJSON, patchedSpec)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal patched pod spec: %w", err)
	}

	return patchedSpec, nil
}

// ApplyPodTemplateOverrides applies the provided PodTemplateOverride to the provided PodTemplateSpec.
func ApplyPodTemplateSpecOverrides(podTemplate *corev1.PodTemplateSpec, override *v1beta1.PodTemplateSpecOverride) error {
	if override == nil {
		return nil
	}

	if override.ObjectMetaOverride != nil {
		if len(override.Labels) > 0 {
			podTemplate.Labels = metadata.Merge(podTemplate.Labels, override.Labels)
		}
		if len(override.Annotations) > 0 {
			podTemplate.Annotations = metadata.Merge(podTemplate.Annotations, override.Annotations)
		}
	}

	if override.Spec != nil {
		// Ensure override has required field initilized so that StrategicMergePatch will work as expected.
		if len(override.Spec.Containers) == 0 {
			override.Spec.Containers = []corev1.Container{}
		}

		if len(override.Spec.InitContainers) == 0 {
			override.Spec.InitContainers = []corev1.Container{}
		}

		patchedSpec, err := PatchPodSpecWithOverride(&podTemplate.Spec, override.Spec)
		if err != nil {
			return err
		}
		if patchedSpec != nil {
			podTemplate.Spec = *patchedSpec
		}
	}
	return nil
}

// ApplyDeploymentOverrides applies the provided DeploymentOverride to the provided Deployment.
func ApplyDeploymentOverrides(deployment *appsv1.Deployment, override *v1beta1.DeploymentOverride) error {
	if override == nil {
		return nil
	}

	if override.ObjectMetaOverride != nil {
		if len(override.Labels) > 0 {
			deployment.Labels = metadata.Merge(deployment.Labels, override.Labels)
		}

		if len(override.Annotations) > 0 {
			deployment.Annotations = metadata.Merge(deployment.Annotations, override.Annotations)
		}
	}

	if override.Spec != nil {
		err := ApplyPodTemplateSpecOverrides(&deployment.Spec.Template, override.Spec.Template)
		if err != nil {
			return err
		}
	}

	return nil
}
