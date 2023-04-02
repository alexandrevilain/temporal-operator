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

package main

import (
	"log"

	"crypto/tls"
	"crypto/x509"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/alexandrevilain/temporal-operator/examples/worker-process/helloworld"
)

func main() {
	clientOptions := client.Options{
		HostPort:  os.Getenv("TEMPORAL_HOST_URL"),
		Namespace: os.Getenv("TEMPORAL_NAMESPACE"),
	}

	if os.Getenv("TEMPORAL_MTLS_TLS_CERT") != "" && os.Getenv("TEMPORAL_MTLS_TLS_KEY") != "" {
		caCert, err := os.ReadFile(os.Getenv("TEMPORAL_MTLS_TLS_CA"))
		if err != nil {
			log.Fatalln("failed reading server CA's certificate", err)
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			log.Fatalln("failed to add server CA's certificate", err)
		}

		cert, err := tls.LoadX509KeyPair(os.Getenv("TEMPORAL_MTLS_TLS_CERT"), os.Getenv("TEMPORAL_MTLS_TLS_KEY"))
		if err != nil {
			log.Fatalln("Unable to load certs", err)
		}

		var serverName string
		if os.Getenv("TEMPORAL_MTLS_TLS_ENABLE_HOST_VERIFICATION") == "true" {
			serverName = os.Getenv("TEMPORAL_MTLS_TLS_SERVER_NAME")
		}

		clientOptions.ConnectionOptions = client.ConnectionOptions{
			TLS: &tls.Config{
				RootCAs:      certPool,
				Certificates: []tls.Certificate{cert},
				ServerName:   serverName,
			},
		}
	}

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "hello-world", worker.Options{})

	w.RegisterWorkflow(helloworld.Workflow)
	w.RegisterActivity(helloworld.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
