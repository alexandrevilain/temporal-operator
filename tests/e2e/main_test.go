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
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2/klogr"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/support/kind"
	"sigs.k8s.io/e2e-framework/third_party/helm"
)

var (
	testenv             env.Environment
	jobTTL              int32 = 60
	workerProcessJobTTL int32 = 300
	replicas            int32 = 1
	listAddress               = "0.0.0.0:9090"
	logger                    = klogr.New()
)

func TestMain(m *testing.M) {
	log.SetLogger(logger)

	os.Exit(testMainRun(m))
}

func testMainRun(m *testing.M) int {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	clusterLogsOutPath := path.Join(wd, "..", "..", "out", "tests", "e2e")

	kubernetesVersion := os.Getenv("KUBERNETES_VERSION")
	if kubernetesVersion == "" {
		kubernetesVersion = "v1.26.0"
	}
	kindImage := fmt.Sprintf("kindest/node:%s", kubernetesVersion)

	operatorImagePath := os.Getenv("OPERATOR_IMAGE_PATH")
	exampleWorkerProcessImagePath := os.Getenv("WORKER_PROCESS_IMAGE_PATH")

	kindClusterName := envconf.RandomName("temporal", 16)
	runID := envconf.RandomName("ns", 4)

	cfg, err := envconf.NewFromFlags()
	if err != nil {
		panic(err)
	}

	setupError := func(err error) error {
		logger.Error(err, "setup error")

		cluster := kind.NewCluster(kindClusterName)
		if err := cluster.ExportLogs(clusterLogsOutPath); err != nil {
			logger.Error(err, "can't export cluster logs")
		}

		return err
	}

	testenv = env.
		NewWithConfig(cfg).
		// Create the cluster
		Setup(
			envfuncs.CreateKindClusterWithConfig(kindClusterName, kindImage, "kind-config.yaml"),
			envfuncs.LoadImageArchiveToCluster(kindClusterName, operatorImagePath),
			envfuncs.LoadImageArchiveToCluster(kindClusterName, exampleWorkerProcessImagePath),
			envfuncs.SetupCRDs("../../out/release/artifacts", "*.crds.yaml"),
		).
		// Add the operators crds to the client scheme.
		Setup(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
			fmt.Printf("KUBECONFIG=%s\n", c.KubeconfigFile())

			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				return ctx, err
			}
			err = v1beta1.AddToScheme(r.GetScheme())
			if err != nil {
				return ctx, err
			}
			return ctx, nil
		}).
		// Deploy cert-manager and ECK.
		Setup(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
			manager := helm.New(c.KubeconfigFile())
			err := manager.RunRepo(helm.WithArgs("add", "jetstack", "https://charts.jetstack.io"))
			if err != nil {
				return ctx, setupError(fmt.Errorf("failed to add cert-manager helm chart repo: %w", err))
			}
			err = manager.RunRepo(helm.WithArgs("add", "elastic", "https://helm.elastic.co"))
			if err != nil {
				return ctx, setupError(fmt.Errorf("failed to add elastic helm chart repo: %w", err))
			}
			err = manager.RunRepo(helm.WithArgs("update"))
			if err != nil {
				return ctx, setupError(fmt.Errorf("failed to upgrade helm repo: %w", err))
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
				return ctx, setupError(fmt.Errorf("failed to install cert-manager chart: %w", err))
			}

			err = manager.RunInstall(
				helm.WithName("elastic-operator"),
				helm.WithNamespace("elastic-system"),
				helm.WithReleaseName("elastic/eck-operator"),
				helm.WithVersion("v2.8.0"),
				helm.WithArgs("--create-namespace"),
				helm.WithWait(),
				helm.WithTimeout("10m"),
			)
			if err != nil {
				return ctx, setupError(fmt.Errorf("failed to install eck-operator chart: %w", err))
			}

			return ctx, nil
		}).
		// Deploy the operator and wait for it.
		Setup(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
			objects, err := decoder.DecodeAllFiles(ctx, os.DirFS("../../out/release/artifacts"), "temporal-operator.yaml")
			if err != nil {
				return ctx, setupError(fmt.Errorf("can't decode operator files: %w", err))
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
				err := c.Client().Resources().Create(ctx, obj)
				if err != nil {
					return ctx, setupError(fmt.Errorf("can't create operator resources: %w", err))
				}
			}

			err = wait.For(conditions.New(c.Client().Resources()).DeploymentConditionMatch(operatorDeploy, appsv1.DeploymentAvailable, corev1.ConditionTrue), wait.WithTimeout(time.Minute*1))
			return ctx, err
		}).
		Finish(
			envfuncs.ExportKindClusterLogs(kindClusterName, clusterLogsOutPath),
			envfuncs.TeardownCRDs("../../out/release/artifacts", "*.crds.yaml"),
			envfuncs.DestroyKindCluster(kindClusterName),
		).
		BeforeEachFeature(func(ctx context.Context, cfg *envconf.Config, t *testing.T, f features.Feature) (context.Context, error) {
			return createNSForTest(ctx, cfg, t, f, runID)
		}).
		AfterEachFeature(func(ctx context.Context, cfg *envconf.Config, t *testing.T, f features.Feature) (context.Context, error) {
			return deleteNSForTest(ctx, cfg, t, f, runID)
		})

	return testenv.Run(m)
}

// createNSForTest creates a random namespace with the runID as a prefix. It is stored in the context
// so that the deleteNSForTest routine can look it up and delete it.
func createNSForTest(ctx context.Context, cfg *envconf.Config, t *testing.T, f features.Feature, runID string) (context.Context, error) {
	ns := envconf.RandomName(runID, 10)
	ctx = SetNamespaceForFeature(ctx, ns)

	t.Logf("Creating namespace %s for feature  \"%s\" in test %s", ns, f.Name(), t.Name())

	return ctx, cfg.Client().Resources().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	})
}

// deleteNSForTest looks up the namespace corresponding to the given test and deletes it.
func deleteNSForTest(ctx context.Context, cfg *envconf.Config, t *testing.T, f features.Feature, runID string) (context.Context, error) {
	ns := GetNamespaceForFeature(ctx)

	t.Logf("Deleting namespace %s for feature \"%s\" in test %s", ns, f.Name(), t.Name())
	return ctx, nil

	// return ctx, cfg.Client().Resources().Delete(ctx, &corev1.Namespace{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name: ns,
	// 	},
	// })
}
