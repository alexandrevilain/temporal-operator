package testworker

import (
	"crypto/tls"

	"go.temporal.io/sdk/client"
)

type ClientOption func(opts *client.Options)

func WithTLSConfig(cfg *tls.Config) ClientOption {
	return func(opts *client.Options) {
		opts.ConnectionOptions.TLS = cfg
	}
}
