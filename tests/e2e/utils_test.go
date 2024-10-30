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
	"net/http"
	"net/http/httputil"
	"strings"
	"testing"
	"time"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	kubernetesutil "github.com/alexandrevilain/temporal-operator/tests/e2e/util/kubernetes"
	"github.com/alexandrevilain/temporal-operator/tests/e2e/util/networking"
	"go.temporal.io/server/common/primitives"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"

	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const doesNotExistName = "does-not-exist"

var defaultVersion = version.MustNewVersionFromString("1.24.3")

func deployAndWaitForTemporalWithPostgres(ctx context.Context, cfg *envconf.Config, namespace string) (*v1beta1.TemporalCluster, error) {
	// create the postgres
	err := deployAndWaitForPostgres(ctx, cfg, namespace)
	if err != nil {
		return nil, err
	}

	pluginName := "postgres12"
	if defaultVersion.GreaterOrEqual(version.V1_24_0) {
		pluginName = "postgres12"
	}

	connectAddr := fmt.Sprintf("postgres.%s:5432", namespace) // create the temporal cluster
	cluster := &v1beta1.TemporalCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: namespace,
		},
		Spec: v1beta1.TemporalClusterSpec{
			NumHistoryShards:           1,
			JobTTLSecondsAfterFinished: &jobTTL,
			Version:                    defaultVersion,
			Metrics: &v1beta1.MetricsSpec{
				Enabled: true,
				Prometheus: &v1beta1.PrometheusSpec{
					ListenAddress: listAddress,
				},
			},
			MTLS: &v1beta1.MTLSSpec{
				Provider: v1beta1.CertManagerMTLSProvider,
				Internode: &v1beta1.InternodeMTLSSpec{
					Enabled: true,
				},
				Frontend: &v1beta1.FrontendMTLSSpec{
					Enabled: true,
				},
			},
			Persistence: v1beta1.TemporalPersistenceSpec{
				DefaultStore: &v1beta1.DatastoreSpec{
					SQL: &v1beta1.SQLSpec{
						User:            "temporal",
						PluginName:      pluginName,
						DatabaseName:    "temporal",
						ConnectAddr:     connectAddr,
						ConnectProtocol: "tcp",
					},
					PasswordSecretRef: &v1beta1.SecretKeyReference{
						Name: "postgres-password",
						Key:  "PASSWORD",
					},
				},
				VisibilityStore: &v1beta1.DatastoreSpec{
					SQL: &v1beta1.SQLSpec{
						User:            "temporal",
						PluginName:      pluginName,
						DatabaseName:    "temporal_visibility",
						ConnectAddr:     connectAddr,
						ConnectProtocol: "tcp",
					},
					PasswordSecretRef: &v1beta1.SecretKeyReference{
						Name: "postgres-password",
						Key:  "PASSWORD",
					},
				},
			},
		},
	}
	err = cfg.Client().Resources(namespace).Create(ctx, cluster)
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

func deployAndWaitForMySQL(ctx context.Context, cfg *envconf.Config, namespace string) error {
	return deployAndWaitFor(ctx, cfg, "mysql", namespace)
}

func deployAndWaitForPostgres(ctx context.Context, cfg *envconf.Config, namespace string) error {
	return deployAndWaitFor(ctx, cfg, "postgres", namespace)
}

func deployAndWaitForElasticSearch(ctx context.Context, cfg *envconf.Config, namespace string) error {
	err := deployTestManifest(ctx, cfg, "elasticsearch", namespace)
	if err != nil {
		return err
	}

	es := &unstructured.Unstructured{}
	es.SetAPIVersion("elasticsearch.k8s.elastic.co/v1")
	es.SetKind("Elasticsearch")
	es.SetName("elasticsearch")
	es.SetNamespace(namespace)

	cond := conditions.New(cfg.Client().Resources()).ResourceMatch(es, func(object k8s.Object) bool {
		o := object.(*unstructured.Unstructured)
		val, found, err := unstructured.NestedString(o.UnstructuredContent(), "status", "health")
		if err != nil {
			return false
		}
		return val == "green" && found
	})

	err = wait.For(cond, wait.WithTimeout(time.Minute*10))
	if err != nil {
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(
		&metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "elasticsearch.k8s.elastic.co/cluster-name",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"elasticsearch"},
				},
			},
		},
	)
	if err != nil {
		return err
	}

	secret := &corev1.Secret{}
	err = cfg.Client().Resources(namespace).Get(ctx, "elasticsearch-es-elastic-user", namespace, secret)
	if err != nil {
		return err
	}

	password, ok := secret.Data["elastic"]
	if !ok {
		return errors.New("can't get elasticsearch user")
	}

	connectAddr, closePortForward, err := forwardPortToPod(ctx, cfg, &testing.T{}, namespace, selector, 9200)
	if err != nil {
		return err
	}

	defer closePortForward()

	body := `
	{
		"persistent": {
		  "cluster": {
			"routing": {
			  "allocation.disk.threshold_enabled": false
			}
		  }
		}
	}`

	url := fmt.Sprintf("http://%s/_cluster/settings", connectAddr)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("elastic", string(password))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	content, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}

	fmt.Printf("%s \n", content)
	return nil
}

func deployAndWaitForCassandra(ctx context.Context, cfg *envconf.Config, namespace string) error {
	name := "cassandra"
	err := deployTestManifest(ctx, cfg, name, namespace)
	if err != nil {
		return err
	}

	pod := corev1.Pod{
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

func waitForDeployment(_ context.Context, cfg *envconf.Config, dep *appsv1.Deployment) error {
	err := wait.For(
		conditions.New(cfg.Client().Resources()).ResourcesFound(&appsv1.DeploymentList{Items: []appsv1.Deployment{*dep}}),
		wait.WithTimeout(time.Minute*10),
	)
	if err != nil {
		return err
	}
	return wait.For(conditions.New(cfg.Client().Resources()).DeploymentConditionMatch(dep, appsv1.DeploymentAvailable, corev1.ConditionTrue), wait.WithTimeout(time.Minute*10))
}

// waitForCluster waits for the temporal cluster's components to be up and running (reporting Ready condition).
func waitForCluster(_ context.Context, cfg *envconf.Config, cluster *v1beta1.TemporalCluster) error {
	cond := conditions.New(cfg.Client().Resources()).ResourceMatch(cluster, func(object k8s.Object) bool {
		return object.(*v1beta1.TemporalCluster).IsReady()
	})
	return wait.For(cond, wait.WithTimeout(time.Minute*10))
}

func waitForClusterClient(_ context.Context, cfg *envconf.Config, clusterClient *v1beta1.TemporalClusterClient) error {
	cond := conditions.New(cfg.Client().Resources()).ResourceMatch(clusterClient, func(object k8s.Object) bool {
		return object.(*v1beta1.TemporalClusterClient).Status.SecretRef != nil &&
			object.(*v1beta1.TemporalClusterClient).Status.SecretRef.Name != ""
	})
	return wait.For(cond, wait.WithTimeout(time.Minute*10))
}

type testLogWriter struct {
	t *testing.T
}

func (t *testLogWriter) Write(p []byte) (n int, err error) {
	t.t.Logf("%s", p)
	return len(p), nil
}

func forwardPortToTemporalFrontend(ctx context.Context, cfg *envconf.Config, t *testing.T, cluster *v1beta1.TemporalCluster) (string, func(), error) {
	selector, err := metav1.LabelSelectorAsSelector(
		&metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "app.kubernetes.io/name",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{cluster.GetName()},
				},
				{
					Key:      "app.kubernetes.io/component",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{string(primitives.FrontendService)},
				},
				{
					Key:      "app.kubernetes.io/version",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{cluster.Spec.Version.String()},
				},
			},
		},
	)
	if err != nil {
		return "", nil, err
	}

	return forwardPortToPod(ctx, cfg, t, cluster.GetNamespace(), selector, 7233)
}

func forwardPortToPod(ctx context.Context, cfg *envconf.Config, t *testing.T, namespace string, selector labels.Selector, port int) (string, func(), error) {
	podList := &corev1.PodList{}
	err := cfg.Client().Resources(namespace).List(ctx, podList, resources.WithLabelSelector(selector.String()))
	if err != nil {
		return "", nil, err
	}

	if len(podList.Items) == 0 {
		return "", nil, errors.New("no frontend port found")
	}

	selectedPod := podList.Items[0]

	localPort, err := networking.GetFreePort()
	if err != nil {
		return "", nil, err
	}

	// stopCh control the port forwarding lifecycle. When it gets closed the
	// port forward will terminate
	stopCh := make(chan struct{}, 1)
	// readyCh communicate when the port forward is ready to get traffic
	readyCh := make(chan struct{})

	out := &testLogWriter{t}

	go func() {
		err := kubernetesutil.ForwardPortToPod(cfg.Client().RESTConfig(), &selectedPod, localPort, port, out, stopCh, readyCh)
		if err != nil {
			panic(err)
		}
	}()

	<-readyCh
	t.Log("Port forwarding is ready to get traffic.")

	connectAddr := fmt.Sprintf("localhost:%d", localPort)
	return connectAddr, func() { close(stopCh) }, nil
}
