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
	"github.com/alexandrevilain/temporal-operator/pkg/resource/mtls/certmanager"
	temporallog "github.com/alexandrevilain/temporal-operator/pkg/temporal/log"
	temporalclient "go.temporal.io/sdk/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// GetTlSConfigFromSecret returns a tls.Config from the provided secret.
// The secret should contain 3 keys: ca.crt, tls.crt and tls.key.
func GetTlSConfigFromSecret(secret *corev1.Secret) (*tls.Config, error) {
	caCrt, ok := secret.Data[certmanager.TLSCA]
	if !ok {
		return nil, errors.New("can't get ca.crt from client secret")
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCrt) {
		return nil, errors.New("failed to add server CA's certificate")
	}

	tlsCrt, ok := secret.Data[certmanager.TLSCert]
	if !ok {
		return nil, errors.New("can't get tls.crt from client secret")
	}

	tlsKey, ok := secret.Data[certmanager.TLSKey]
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

// GetClusterClientTLSConfig returns the tls configuration for the provided temporal cluster.
func GetClusterClientTLSConfig(ctx context.Context, client client.Client, cluster *v1alpha1.TemporalCluster) (*tls.Config, error) {
	secret := &corev1.Secret{}

	err := client.Get(ctx, types.NamespacedName{
		Name:      cluster.ChildResourceName(certmanager.FrontendCertificate),
		Namespace: cluster.GetNamespace(),
	}, secret)
	if err != nil {
		return nil, err
	}

	tlsConfig, err := GetTlSConfigFromSecret(secret)
	if err != nil {
		return nil, err
	}

	tlsConfig.ServerName = cluster.Spec.MTLS.Frontend.ServerName(cluster.ServerName())
	return tlsConfig, nil

}

func buildClusterClientOptions(ctx context.Context, client client.Client, cluster *v1alpha1.TemporalCluster, overrides ...ClientOption) (temporalclient.Options, error) {
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

	for _, override := range overrides {
		override(&opts)
	}

	return opts, nil
}

// ClientOption is an override option for temporal sdk client.
type ClientOption func(opts *temporalclient.Options)

// WithTLSConfig is overriding the client tls config.
func WithTLSConfig(cfg *tls.Config) ClientOption {
	return func(opts *temporalclient.Options) {
		opts.ConnectionOptions.TLS = cfg
	}
}

// WithHostPort is overriding the client host port.
func WithHostPort(hostPort string) ClientOption {
	return func(opts *temporalclient.Options) {
		opts.HostPort = hostPort
	}
}

// GetClusterClient returns a temporal sdk client for the provider temporal cluster.
func GetClusterClient(ctx context.Context, client client.Client, cluster *v1alpha1.TemporalCluster, overrides ...ClientOption) (temporalclient.Client, error) {
	opts, err := buildClusterClientOptions(ctx, client, cluster, overrides...)
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

// GetClusterNamespaceClient returns a temporal sdk namespace client for the provider temporal cluster.
func GetClusterNamespaceClient(ctx context.Context, client client.Client, cluster *v1alpha1.TemporalCluster, overrides ...ClientOption) (temporalclient.NamespaceClient, error) {
	opts, err := buildClusterClientOptions(ctx, client, cluster, overrides...)
	if err != nil {
		return nil, err
	}

	log.FromContext(ctx).V(1).Info("Connecting to temporal cluster", "address", opts.HostPort)

	return temporalclient.NewNamespaceClient(opts)
}
