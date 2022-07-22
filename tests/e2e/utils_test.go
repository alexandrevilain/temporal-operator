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

package e2e

import (
	"context"
	"errors"
	"fmt"
	"time"

	appsv1alpha1 "github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/tests/e2e/networking"
	"github.com/anthhub/forwarder"
	"go.temporal.io/server/common"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func deployAndWaitForMySQL(ctx context.Context, cfg *envconf.Config, namespace string) error {
	return deployAndWaitFor(ctx, cfg, "mysql", namespace)
}

func deployAndWaitForPostgres(ctx context.Context, cfg *envconf.Config, namespace string) error {
	return deployAndWaitFor(ctx, cfg, "postgres", namespace)
}

func deployAndWaitForCassandra(ctx context.Context, cfg *envconf.Config, namespace string) error {
	name := "cassandra"
	err := deployTestManifest(ctx, cfg, name, namespace)
	if err != nil {
		return err
	}

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-0", name), Namespace: namespace},
	}

	return wait.For(conditions.New(cfg.Client().Resources()).PodReady(&pod), wait.WithTimeout(10*time.Minute))
}

func deployAndWaitFor(ctx context.Context, cfg *envconf.Config, name, namespace string) error {
	err := deployTestManifest(ctx, cfg, name, namespace)
	if err != nil {
		return err
	}

	dep := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}

	// wait for the deployment to become available
	return waitForDeployment(ctx, cfg, &dep)
}

func deployTestManifest(ctx context.Context, cfg *envconf.Config, name, namespace string) error {
	path := fmt.Sprintf("testdata/%s", name)
	return decoder.ApplyWithManifestDir(ctx, cfg.Client().Resources(namespace), path, "*", []resources.CreateOption{}, decoder.MutateNamespace(namespace))
}

func waitForDeployment(ctx context.Context, cfg *envconf.Config, dep *appsv1.Deployment) error {
	err := wait.For(
		conditions.New(cfg.Client().Resources()).ResourcesFound(&appsv1.DeploymentList{Items: []appsv1.Deployment{*dep}}),
		wait.WithTimeout(time.Minute*10),
	)
	if err != nil {
		return err
	}
	return wait.For(conditions.New(cfg.Client().Resources()).DeploymentConditionMatch(dep, appsv1.DeploymentAvailable, v1.ConditionTrue), wait.WithTimeout(time.Minute*10))
}

// waitForTemporalCluster waits for the temporal cluster's components to be up and running (reporting Ready condition).
func waitForTemporalCluster(ctx context.Context, cfg *envconf.Config, temporalCluster *appsv1alpha1.TemporalCluster) error {
	cond := conditions.New(cfg.Client().Resources()).ResourceMatch(temporalCluster, func(object k8s.Object) bool {
		for _, condition := range object.(*appsv1alpha1.TemporalCluster).Status.Conditions {
			if condition.Type == appsv1alpha1.ReadyCondition && condition.Status == metav1.ConditionTrue {
				return true
			}
		}
		return false
	})
	return wait.For(cond, wait.WithTimeout(time.Minute*10))
}

func waitForTemporalClusterClient(ctx context.Context, cfg *envconf.Config, temporalClusterClient *appsv1alpha1.TemporalClusterClient) error {
	cond := conditions.New(cfg.Client().Resources()).ResourceMatch(temporalClusterClient, func(object k8s.Object) bool {
		return object.(*appsv1alpha1.TemporalClusterClient).Status.SecretRef.Name != ""
	})
	return wait.For(cond, wait.WithTimeout(time.Minute*10))
}

func forwardPortToTemporalFrontend(ctx context.Context, cfg *envconf.Config, temporalCluster *appsv1alpha1.TemporalCluster) (string, func(), error) {
	selector, err := metav1.LabelSelectorAsSelector(
		&metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "app.kubernetes.io/name",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{temporalCluster.GetName()},
				},
				{
					Key:      "app.kubernetes.io/component",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{common.FrontendServiceName},
				},
			},
		},
	)
	if err != nil {
		return "", nil, err
	}

	podList := &v1.PodList{}
	err = cfg.Client().Resources(temporalCluster.GetNamespace()).List(ctx, podList, resources.WithLabelSelector(selector.String()))
	if err != nil {
		return "", nil, err
	}

	if len(podList.Items) == 0 {
		return "", nil, errors.New("No frontend port found")
	}

	port, err := networking.GetFreePort()
	if err != nil {
		return "", nil, err
	}

	ret, err := forwarder.WithRestConfig(ctx, []*forwarder.Option{
		{
			LocalPort:  port,
			RemotePort: 7233,
			Namespace:  podList.Items[0].GetNamespace(),
			PodName:    podList.Items[0].GetName(),
		},
	}, cfg.Client().RESTConfig())
	if err != nil {
		return "", nil, err
	}
	_, err = ret.Ready()
	if err != nil {
		return "", nil, err
	}

	connectAddr := fmt.Sprintf("localhost:%d", port)

	return connectAddr, func() { ret.Close() }, nil
}
