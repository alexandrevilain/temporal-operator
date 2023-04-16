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

package workerprocessbuilder

import (
	"bytes"
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/metadata"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Schema string

type JobScriptsConfigmapBuilder struct {
	instance *v1beta1.TemporalWorkerProcess
	scheme   *runtime.Scheme
}

func NewJobScriptsConfigmapBuilder(instance *v1beta1.TemporalWorkerProcess, scheme *runtime.Scheme) *JobScriptsConfigmapBuilder {
	return &JobScriptsConfigmapBuilder{
		instance: instance,
		scheme:   scheme,
	}
}

func (b *JobScriptsConfigmapBuilder) Build() client.Object {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        b.instance.ChildResourceName("builder-scripts"),
			Namespace:   b.instance.Namespace,
			Labels:      metadata.GetVersionStringLabels(b.instance, "builder-scripts", b.instance.Spec.Version, b.instance.Labels),
			Annotations: metadata.GetAnnotations(b.instance.Name, b.instance.Annotations),
		},
	}
}

func (b *JobScriptsConfigmapBuilder) Enabled() bool {
	return true
}

func (b *JobScriptsConfigmapBuilder) Update(object client.Object) error {
	configMap := object.(*corev1.ConfigMap)

	var renderedWorkerBuilder bytes.Buffer
	err := templates[DefaultWorkerBuilderTemplate].Execute(&renderedWorkerBuilder, createWorkerBuilder{
		GitRepo:                 b.instance.Spec.Builder.GitRepository.URL,
		GitBranch:               b.instance.Spec.Builder.GitRepository.Reference.Branch,
		BuildDir:                b.instance.Spec.Builder.BuildDir,
		Image:                   fmt.Sprintf("%s:%s", b.instance.Spec.Image, b.instance.Spec.Version),
		BuildRepo:               b.instance.Spec.Builder.BuildRegistry.Repository,
		BuildRepoUsername:       b.instance.Spec.Builder.BuildRegistry.Username,
		BuildRepoPasswordEnvVar: b.instance.Spec.Builder.GetBuildRepoPasswordEnvVarName(),
	})
	if err != nil {
		return fmt.Errorf("can't render default-worker-builder.sh: %w", err)
	}

	configMap.Data = map[string]string{
		"build-worker-process.sh": renderedWorkerBuilder.String(),
	}

	if err := controllerutil.SetControllerReference(b.instance, configMap, b.scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %w", err)
	}

	return nil
}
