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
	"net/url"
	"os"
	"strings"
	"time"

	appsv1alpha1 "github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/alexandrevilain/temporal-operator/tests/e2e/networking"
	"go.temporal.io/server/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func deployAndWaitForTemporalWithPostgres(ctx context.Context, cfg *envconf.Config, namespace, version string) (*appsv1alpha1.TemporalCluster, error) {
	// create the postgres
	err := deployAndWaitForPostgres(ctx, cfg, namespace)
	if err != nil {
		return nil, err
	}

	connectAddr := fmt.Sprintf("postgres.%s:5432", namespace) // create the temporal cluster
	temporalCluster := &appsv1alpha1.TemporalCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: namespace,
		},
		Spec: appsv1alpha1.TemporalClusterSpec{
			NumHistoryShards: 1,
			Version:          version,
			MTLS: &appsv1alpha1.MTLSSpec{
				Provider: appsv1alpha1.CertManagerMTLSProvider,
				Internode: &appsv1alpha1.InternodeMTLSSpec{
					Enabled: true,
				},
				Frontend: &appsv1alpha1.FrontendMTLSSpec{
					Enabled: true,
				},
			},
			Persistence: appsv1alpha1.TemporalPersistenceSpec{
				DefaultStore:    "default",
				VisibilityStore: "visibility",
			},
			Datastores: []appsv1alpha1.TemporalDatastoreSpec{
				{
					Name: "default",
					SQL: &appsv1alpha1.SQLSpec{
						User:            "temporal",
						PluginName:      "postgres",
						DatabaseName:    "temporal",
						ConnectAddr:     connectAddr,
						ConnectProtocol: "tcp",
					},
					PasswordSecretRef: appsv1alpha1.SecretKeyReference{
						Name: "postgres-password",
						Key:  "PASSWORD",
					},
				},
				{
					Name: "visibility",
					SQL: &appsv1alpha1.SQLSpec{
						User:            "temporal",
						PluginName:      "postgres",
						DatabaseName:    "temporal_visibility",
						ConnectAddr:     connectAddr,
						ConnectProtocol: "tcp",
					},
					PasswordSecretRef: appsv1alpha1.SecretKeyReference{
						Name: "postgres-password",
						Key:  "PASSWORD",
					},
				},
			},
		},
	}
	err = cfg.Client().Resources(namespace).Create(ctx, temporalCluster)
	if err != nil {
		return nil, err
	}

	return temporalCluster, nil

}

func klientToControllerRuntimeClient(k klient.Client) (client.Client, error) {
	return client.New(k.RESTConfig(), client.Options{})
}

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

func waitForDeployment(ctx context.Context, cfg *envconf.Config, dep *appsv1.Deployment) error {
	err := wait.For(
		conditions.New(cfg.Client().Resources()).ResourcesFound(&appsv1.DeploymentList{Items: []appsv1.Deployment{*dep}}),
		wait.WithTimeout(time.Minute*10),
	)
	if err != nil {
		return err
	}
	return wait.For(conditions.New(cfg.Client().Resources()).DeploymentConditionMatch(dep, appsv1.DeploymentAvailable, corev1.ConditionTrue), wait.WithTimeout(time.Minute*10))
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

func forwardPortToPod(cfg *rest.Config, pod *corev1.Pod, port int, stopCh <-chan struct{}, readyCh chan struct{}) error {
	stream := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", pod.Namespace, pod.Name)
	hostIP := strings.TrimLeft(cfg.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(cfg)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", port, 7233)}, stopCh, readyCh, stream.Out, stream.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
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
				{
					Key:      "app.kubernetes.io/version",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{temporalCluster.Spec.Version},
				},
			},
		},
	)
	if err != nil {
		return "", nil, err
	}

	podList := &corev1.PodList{}
	err = cfg.Client().Resources(temporalCluster.GetNamespace()).List(ctx, podList, resources.WithLabelSelector(selector.String()))
	if err != nil {
		return "", nil, err
	}

	if len(podList.Items) == 0 {
		return "", nil, errors.New("No frontend port found")
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

	go func() {
		err := forwardPortToPod(cfg.Client().RESTConfig(), &selectedPod, localPort, stopCh, readyCh)
		if err != nil {
			panic(err)
		}
	}()

	<-readyCh
	println("Port forwarding is ready to get traffic.")

	connectAddr := fmt.Sprintf("localhost:%d", localPort)
	return connectAddr, func() { close(stopCh) }, nil
}
