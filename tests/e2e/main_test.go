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
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/third_party/helm"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
)

var testenv env.Environment
var jobTtl int32 = 60
var replicas int32 = 1
var port int = 7233
var listAddress string = "0.0.0.0:9090"

func TestMain(m *testing.M) {
	kindVersion := os.Getenv("KUBERNETES_VERSION")
	if kindVersion == "" {
		kindVersion = "v1.23.4"
	}
	kindImage := fmt.Sprintf("kindest/node:%s", kindVersion)

	operatorImagePath := os.Getenv("OPERATOR_IMAGE_PATH")

	kindClusterName := envconf.RandomName("temporal", 16)
	runID := envconf.RandomName("ns", 4)

	cfg, err := envconf.NewFromFlags()
	if err != nil {
		log.Fatalf("envconf failed: %s", err)
	}

	testenv = env.
		NewWithConfig(cfg).
		// Create the cluster
		Setup(
			envfuncs.CreateKindClusterWithConfig(kindClusterName, kindImage, "kind-config.yaml"),
			envfuncs.LoadImageArchiveToCluster(kindClusterName, operatorImagePath),
			envfuncs.SetupCRDs("../../out/release/artifacts", "*.crds.yaml"),
		).
		// Add the operators crds to the client scheme.
		Setup(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
			fmt.Printf("KUBECONFIG=%s\n", c.KubeconfigFile())

			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				return ctx, err
			}
			v1beta1.AddToScheme(r.GetScheme())
			return ctx, nil
		}).
		// Deploy cert-manager.
		Setup(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
			manager := helm.New(c.KubeconfigFile())
			err := manager.RunRepo(helm.WithArgs("add", "jetstack", "https://charts.jetstack.io"))
			if err != nil {
				return ctx, fmt.Errorf("failed to add cert-manager helm chart repo: %w", err)
			}
			err = manager.RunRepo(helm.WithArgs("update"))
			if err != nil {
				return ctx, fmt.Errorf("failed to upgrade helm repo: %w", err)
			}
			err = manager.RunInstall(
				helm.WithName("cert-manager"),
				helm.WithNamespace("cert-manager"),
				helm.WithReleaseName("jetstack/cert-manager"),
				helm.WithVersion("v1.8.2"),
				helm.WithArgs("--create-namespace"),
				helm.WithArgs("--set", "installCRDs=true"),
				helm.WithWait(),
				helm.WithTimeout("10m"),
			)
			if err != nil {
				return ctx, fmt.Errorf("failed to install cert-manager chart: %w", err)
			}
			return ctx, nil
		}).
		// Deploy the operator and wait for it.
		Setup(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
			objects, err := decoder.DecodeAllFiles(ctx, os.DirFS("../../out/release/artifacts"), "temporal-operator.yaml")
			if err != nil {
				return ctx, err
			}

			var operatorDeploy *appsv1.Deployment
			for _, obj := range objects {
				deploy, ok := obj.(*appsv1.Deployment)
				if ok {
					operatorDeploy = deploy
					for i, container := range deploy.Spec.Template.Spec.Containers {
						if strings.Contains(container.Image, "ghcr.io/alexandrevilain/temporal-operator") {
							deploy.Spec.Template.Spec.Containers[i].Image = "temporal-operator"
							deploy.Spec.Template.Spec.Containers[i].ImagePullPolicy = "IfNotPresent"
						}
					}
				}
				c.Client().Resources().Create(ctx, obj)
			}

			err = wait.For(conditions.New(c.Client().Resources()).DeploymentConditionMatch(operatorDeploy, appsv1.DeploymentAvailable, v1.ConditionTrue), wait.WithTimeout(time.Minute*1))
			return ctx, err
		}).
		Finish(
			envfuncs.TeardownCRDs("../../out/release/artifacts", "*.crds.yaml"),
			envfuncs.DestroyKindCluster(kindClusterName),
		).
		BeforeEachFeature(func(ctx context.Context, cfg *envconf.Config, t *testing.T, f features.Feature) (context.Context, error) {
			return createNSForTest(ctx, cfg, t, f, runID)
		}).
		AfterEachFeature(func(ctx context.Context, cfg *envconf.Config, t *testing.T, f features.Feature) (context.Context, error) {
			return deleteNSForTest(ctx, cfg, t, f, runID)
		})

	os.Exit(testenv.Run(m))
}

// createNSForTest creates a random namespace with the runID as a prefix. It is stored in the context
// so that the deleteNSForTest routine can look it up and delete it.
func createNSForTest(ctx context.Context, cfg *envconf.Config, t *testing.T, f features.Feature, runID string) (context.Context, error) {
	ns := envconf.RandomName(runID, 10)
	ctx = SetNamespaceForFeature(ctx, ns)

	t.Logf("Creating namespace %s for feature  \"%s\" in test %s", ns, f.Name(), t.Name())

	return ctx, cfg.Client().Resources().Create(ctx, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	})
}

// deleteNSForTest looks up the namespace corresponding to the given test and deletes it.
func deleteNSForTest(ctx context.Context, cfg *envconf.Config, t *testing.T, f features.Feature, runID string) (context.Context, error) {
	ns := GetNamespaceForFeature(ctx)

	t.Logf("Deleting namespace %s for feature \"%s\" in test %s", ns, f.Name(), t.Name())

	return ctx, cfg.Client().Resources().Delete(ctx, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	})
}

type (
	clusterContextKey   string
	namespaceContextKey string
)

var (
	clusterKey   clusterContextKey   = "cluster"
	namespaceKey namespaceContextKey = "namespace"
)

func GetNamespaceForFeature(ctx context.Context) string {
	return ctx.Value(namespaceKey).(string)
}

func SetNamespaceForFeature(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, namespaceKey, namespace)
}
