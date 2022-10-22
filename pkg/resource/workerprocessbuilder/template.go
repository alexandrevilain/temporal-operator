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

package workerprocessbuilder

import (
	"text/template"

	"github.com/lithammer/dedent"
)

const (
	DefaultWorkerBuilderTemplate = "default-worker-builder.sh"
)

type createWorkerBuilder struct {
	GitRepo                 string
	GitBranch               string
	BuildDir                string
	Image                   string
	BuildRepo               string
	BuildRepoUsername       string
	BuildRepoPasswordEnvVar string
}

var (
	templates = map[string]*template.Template{}

	templatesContent = map[string]string{
		DefaultWorkerBuilderTemplate: dedent.Dedent(`
			#!/bin/sh
			dnf install -y git
			mkdir -p /app
			cd app
			git clone --single-branch --branch {{ .GitBranch }} {{ .GitRepo }}
			cd {{ .BuildDir }}
			podman build -t {{ .BuildRepo }}/{{ .Image }} .
			podman login {{ .BuildRepo }} --username {{ .BuildRepoUsername }} --password ${{ .BuildRepoPasswordEnvVar }}
			podman push {{ .BuildRepo }}/{{ .Image }}
		`),
	}
)

func init() {
	for name, content := range templatesContent {
		templates[name] = template.Must(template.New(name).Parse(content))
		template.Must(templates[name].Parse(content))
	}
}
