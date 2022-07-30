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

package temporal

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	temporallog "github.com/alexandrevilain/temporal-operator/pkg/temporal/log"
	temporalclient "go.temporal.io/sdk/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func GetTlSConfigFromSecret(ctx context.Context, secret *corev1.Secret) (*tls.Config, error) {
	caCrt, ok := secret.Data["ca.crt"]
	if !ok {
		return nil, errors.New("can't get ca.crt from client secret")
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCrt) {
		return nil, errors.New("failed to add server CA's certificate")
	}

	tlsCrt, ok := secret.Data["tls.crt"]
	if !ok {
		return nil, errors.New("can't get tls.crt from client secret")
	}

	tlsKey, ok := secret.Data["tls.key"]
	if !ok {
		return nil, errors.New("can't get tls.key from client secret")
	}

	clientCert, err := tls.X509KeyPair(tlsCrt, tlsKey)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{clientCert},
	}, nil
}

func GetClusterClientTLSConfig(ctx context.Context, client client.Client, cluster *v1alpha1.TemporalCluster) (*tls.Config, error) {
	secret := &corev1.Secret{}

	err := client.Get(ctx, types.NamespacedName{
		Name:      cluster.ChildResourceName("frontend-certificate"),
		Namespace: cluster.GetNamespace(),
	}, secret)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := GetTlSConfigFromSecret(ctx, secret)
	if err != nil {
		return nil, err
	}

	tlsConfig.ServerName = cluster.Spec.MTLS.Frontend.ServerName(cluster.ServerName())
	return tlsConfig, nil

}

func buildClusterClientOptions(ctx context.Context, client client.Client, cluster *v1alpha1.TemporalCluster) (temporalclient.Options, error) {
	opts := temporalclient.Options{
		HostPort: cluster.GetPublicClientAddress(),
		Logger:   temporallog.NewTemporalSDKLogFromContext(ctx),
	}
	if cluster.MTLSWithCertManagerEnabled() && cluster.Spec.MTLS.FrontendEnabled() {
		tlsConfig, err := GetClusterClientTLSConfig(ctx, client, cluster)
		if err != nil {
			return opts, fmt.Errorf("can't get cluster TLS config: %w", err)
		}
		opts.ConnectionOptions.TLS = tlsConfig
	}
	return opts, nil
}

func GetClusterClient(ctx context.Context, client client.Client, cluster *v1alpha1.TemporalCluster) (temporalclient.Client, error) {
	opts, err := buildClusterClientOptions(ctx, client, cluster)
	if err != nil {
		return nil, err
	}

	log.FromContext(ctx).V(1).Info("Connecting to temporal cluster", "address", opts.HostPort)

	c, err := temporalclient.Dial(opts)
	if err != nil {
		return nil, fmt.Errorf("can't create temporal client: %w", err)
	}
	return c, nil
}

func GetClusterNamespaceClient(ctx context.Context, client client.Client, cluster *v1alpha1.TemporalCluster) (temporalclient.NamespaceClient, error) {
	opts, err := buildClusterClientOptions(ctx, client, cluster)
	if err != nil {
		return nil, err
	}

	log.FromContext(ctx).V(1).Info("Connecting to temporal cluster", "address", opts.HostPort)

	return temporalclient.NewNamespaceClient(opts)
}
