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

package meta_test

import (
	"testing"

	"github.com/alexandrevilain/controller-tools/pkg/resource"
	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/resource/admintools"
	"github.com/alexandrevilain/temporal-operator/internal/resource/base"
	"github.com/alexandrevilain/temporal-operator/internal/resource/ui"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// `kubectl rollout restart` sets this annotation on the pod template. A reconcile
// that drops it rolls the new pods back, so every Deployment builder keeps it.
func TestDeploymentUpdateKeepsRolloutRestart(t *testing.T) {
	const key, at = "kubectl.kubernetes.io/restartedAt", "2026-11-05T10:00:00Z"

	scheme := runtime.NewScheme()
	require.NoError(t, v1beta1.AddToScheme(scheme))

	c := &v1beta1.TemporalCluster{
		TypeMeta:   v1beta1.TemporalClusterTypeMeta,
		ObjectMeta: metav1.ObjectMeta{Name: "prod", Namespace: "demo"},
		Spec: v1beta1.TemporalClusterSpec{
			Image:      "temporalio/server:1.28.4",
			UI:         &v1beta1.TemporalUISpec{Enabled: true},
			AdminTools: &v1beta1.TemporalAdminToolsSpec{Enabled: true},
			Persistence: v1beta1.TemporalPersistenceSpec{
				DefaultStore:    &v1beta1.DatastoreSpec{Name: "default", SQL: &v1beta1.SQLSpec{PluginName: "postgres12"}},
				VisibilityStore: &v1beta1.DatastoreSpec{Name: "visibility", SQL: &v1beta1.SQLSpec{PluginName: "postgres12"}},
			},
		},
	}
	c.Default()

	for name, b := range map[string]resource.Builder{
		"frontend":   base.NewDeploymentBuilder("frontend", c, scheme, c.Spec.Services.Frontend, "hash"),
		"ui":         ui.NewDeploymentBuilder(c, scheme, "hash"),
		"admintools": admintools.NewDeploymentBuilder(c, scheme, "hash"),
	} {
		t.Run(name, func(t *testing.T) {
			d := b.Build().(*appsv1.Deployment)
			d.Spec.Template.Annotations = map[string]string{key: at}
			require.NoError(t, b.Update(d))
			require.Equal(t, at, d.Spec.Template.Annotations[key])
		})
	}
}
