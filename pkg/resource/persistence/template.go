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

package persistence

import (
	"text/template"

	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/lithammer/dedent"
)

var (
	setupSchema              *template.Template
	updateSchema             *template.Template
	setupAdvancedVisibility  *template.Template
	updateAdvancedVisibility *template.Template
)

type (
	setupSchemaData struct {
		Tool           string
		ConnectionArgs string
		InitialVersion string
	}

	updateSchemaData struct {
		Tool           string
		ConnectionArgs string
		SchemaDir      string
	}

	setupAdvancedVisibilityData struct {
		Version        string
		URL            string
		Username       string
		PasswordEnvVar string
		Indices        v1alpha1.ElasticsearchIndices
	}
	updateAdvancedVisibilityData struct{}
)

var (
	setupSchemaContent = dedent.Dedent(`
		#!/bin/bash
		{{ .Tool }} {{ .ConnectionArgs }} setup-schema -v {{ .InitialVersion }}
	`)

	updateSchemaContent = dedent.Dedent(`
		#!/bin/bash
		{{ .Tool }} {{ .ConnectionArgs }} update-schema -d {{ .SchemaDir }}
	`)
	setupAdvancedVisibilityContent = dedent.Dedent(`
		#!/bin/bash

		curl --fail --user "{{ .Username }}":"${{ .PasswordEnvVar }}" -X PUT "{{ .URL }}/_cluster/settings" -H "Content-Type: application/json" --data-binary @/etc/temporal/schema/elasticsearch/visibility/cluster_settings_{{ .Version }}.json --write-out "\n"
		curl --fail --user "{{ .Username }}":"${{ .PasswordEnvVar }}" -X PUT "{{ .URL }}/_template/temporal_visibility_v1_template" -H "Content-Type: application/json" --data-binary @/etc/temporal/schema/elasticsearch/visibility/index_template_{{ .Version }}.json --write-out "\n"
		# No --fail here because create index is not idempotent operation.
		curl --user "{{ .Username }}":"${{ .PasswordEnvVar }}" -X PUT "{{ .URL }}/{{ .Indices.Visibility }}" --write-out "\n"
		{{ if .Indices.SecondaryVisibility }}
		curl --user "{{ .Username }}":"${{ .PasswordEnvVar }}" -X PUT "{{ .URL }}/{{ .Indices.SecondaryVisibility }}" --write-out "\n"
		{{ end }}
	`)
	updateAdvancedVisibilityContent = dedent.Dedent(`
		#!/bin/bash

		# Try to guess current version by querying mapping.
		# If the mapping has "blabl" => So it's v2, otherwise if it has "bla" ist's v1
		# From that compute the folders where we need to run ./upgrade.sh

	`)
)

func init() {
	setupSchema = template.Must(
		template.
			New("setup.sh").
			Parse(setupSchemaContent))

	updateSchema = template.Must(
		template.
			New("update.sh").
			Parse(updateSchemaContent))

	setupAdvancedVisibility = template.Must(
		template.
			New("setup-advanced-visibility.sh").
			Parse(setupAdvancedVisibilityContent))

	updateAdvancedVisibility = template.Must(
		template.
			New("update-advanced-visibility.sh").
			Parse(updateAdvancedVisibilityContent))
}
