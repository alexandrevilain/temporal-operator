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

const (
	CreateCassandraTemplate  = "create-cassandra.sh"
	CreateDatabaseTemplate   = "create-database.sh"
	SetupSchemaTemplate      = "setup-schema.sh"
	UpdateSchemaTemplate     = "update-schema.sh"
	SetupAdvancedVisibility  = "setup-advanced-visibility.sh"
	UpdateAdvancedVisibility = "update-advanced-visibility.sh"
)

var (
	templates = map[string]*template.Template{}

	templatesContent = map[string]string{
		CreateCassandraTemplate: dedent.Dedent(`
			#!/bin/bash
			{{ .Tool }} {{ .ConnectionArgs }} create-Keyspace -k {{ .KeyspaceName }}
		`),
		CreateDatabaseTemplate: dedent.Dedent(`
			#!/bin/bash
			{{ .Tool }} {{ .ConnectionArgs }}  create-database -database {{ .DatabaseName }}
		`),
		SetupSchemaTemplate: dedent.Dedent(`
			#!/bin/bash
			{{ .Tool }} {{ .ConnectionArgs }} setup-schema -v {{ .InitialVersion }}
			{{ template "scripts" . }}
		`),
		UpdateSchemaTemplate: dedent.Dedent(`
			#!/bin/bash
			{{ .Tool }} {{ .ConnectionArgs }} update-schema -d {{ .SchemaDir }}
			{{ template "scripts" . }}
		`),
		SetupAdvancedVisibility: dedent.Dedent(`
			#!/bin/bash

			curl --fail --user "{{ .Username }}":"${{ .PasswordEnvVar }}" -X PUT "{{ .URL }}/_cluster/settings" -H "Content-Type: application/json" --data-binary @/etc/temporal/schema/elasticsearch/visibility/cluster_settings_{{ .Version }}.json --write-out "\n"
			curl --fail --user "{{ .Username }}":"${{ .PasswordEnvVar }}" -X PUT "{{ .URL }}/_template/temporal_visibility_v1_template" -H "Content-Type: application/json" --data-binary @/etc/temporal/schema/elasticsearch/visibility/index_template_{{ .Version }}.json --write-out "\n"
			# No --fail here because create index is not idempotent operation.
			curl --user "{{ .Username }}":"${{ .PasswordEnvVar }}" -X PUT "{{ .URL }}/{{ .Indices.Visibility }}" --write-out "\n"
			{{ if .Indices.SecondaryVisibility }}
			curl --user "{{ .Username }}":"${{ .PasswordEnvVar }}" -X PUT "{{ .URL }}/{{ .Indices.SecondaryVisibility }}" --write-out "\n"
			{{ end }}
			{{ template "scripts" . }}
		`),
		UpdateAdvancedVisibility: dedent.Dedent(`
			#!/bin/bash

			do_upgrade() {
				desired_version=$1
				es_version="{{ .Version }}"

				case $desired_version in
					v2)
						echo "Upgrading to schema v2"

						# Extracted from: 
						# https://github.com/temporalio/temporal/blob/v1.17.5/schema/elasticsearch/visibility/versioned/v2/upgrade.sh

						case $es_version in
							v6) date_type='date'       ; doc_type='/_doc' ;;
							*)  date_type='date_nanos' ; doc_type=''      ;;
						esac

						new_mapping='
						{
							"properties": {
								"TemporalScheduledStartTime": {
								"type": "'$date_type'"
								},
								"TemporalScheduledById": {
								"type": "keyword"
								},
								"TemporalSchedulePaused": {
								"type": "boolean"
								}
							}
						}
						'

						curl --silent --user "{{ .Username }}":"${{ .PasswordEnvVar }}" -X PUT "{{ .URL }}/{{ .Indices.Visibility }}${doc_type}/_mapping" -H "Content-Type: application/json" --data-binary "$new_mapping" | jq
						;;
					v3)
						echo "Upgrading to schema v3"
						
						new_mapping='
						{
							"properties": {
								"TemporalNamespaceDivision": {
								"type": "keyword"
								}
							}
						}
						'

						curl --silent --user "{{ .Username }}":"${{ .PasswordEnvVar }}" -X PUT "{{ .URL }}/{{ .Indices.Visibility }}/_mapping" -H "Content-Type: application/json" --data-binary "$new_mapping" | jq
						;;
				esac
			}

			# Get the expected schema version from the current simlink pointing to the versionned.
			expected_version=$(realpath /etc/temporal/schema/elasticsearch/visibility/index_template_v7.json | sed -e 's/.*versioned\/\(.*\)\/index_template_v7.json.*/\1/')
			current_version=""
			current_version_found=false

			# Get the current_mapping value in elasticsearch.
			current_mapping=$(curl --silent --user "{{ .Username }}":"${{ .PasswordEnvVar }}" {{ .URL }}/{{ .Indices.Visibility }})

			# Guess current mapping version
			# v0 does not have the "ExecutionDuration" property
			is_v0=$(echo $current_mapping | jq -r '.temporal_visibility_v1_dev.mappings.properties | has("ExecutionDuration") | not')
			if [ $is_v0 == "true" ]; then
				echo "Can't do upgrade from v0 schema, version needing advanced visibility schema v1 are not supported by the operator"
				exit 1;
			fi

			# v1 does not have the "TemporalScheduledById" property
			is_v1=$(echo $current_mapping | jq -r '.temporal_visibility_v1_dev.mappings.properties | has("TemporalScheduledById") | not')
			if [ $is_v1 == "true" ]; then
				if [ $current_version_found = false ]; then
					current_version_found=true
					current_version="v1"
				fi
			fi

			# v2 does not have the "TemporalNamespaceDivision" property
			is_v2=$(echo $current_mapping | jq -r '.temporal_visibility_v1_dev.mappings.properties | has("TemporalNamespaceDivision") | not')
			if [ $is_v2 == "true" ]; then
				if [ $current_version_found = false ]; then
					current_version_found=true
					current_version="v2"
				fi
			fi

			# v3 has the "TemporalNamespaceDivision"
			is_v3=$(echo $current_mapping | jq -r '.temporal_visibility_v1_dev.mappings.properties | has("TemporalNamespaceDivision")')
			if [ $is_v3 == "true" ]; then
				if [ $current_version_found = false ]; then
					current_version_found=true
					current_version="v3"
				fi
			fi

			echo "Expected schema version: $expected_version"
			echo "Current schema version: $current_version"

			current_version_int=$(echo $current_version | sed 's/^.//')
			expected_version_int=$(echo $expected_version | sed 's/^.//')

			if [ $current_version_int -eq $expected_version_int ]; then
				echo "Current schema version is already at the expected version"
				exit 0
			fi

			if [ $current_version_int -gt $expected_version_int ]; then
				echo "Current schema version is already to a newer version"
				exit 0
			fi

			echo "Schema version upgrade is needed"

			# If the current version is v1, the script only supports to update to v2. 
			expected_next_version=$(( current_version_int + 1))

			if [ $expected_next_version -ne $expected_version_int ]; then
				echo "Can't do Elasticsearch schema upgrade for no-following version numbers. (eg. from v1 to v2, but not from v1 to v3)"
				exit 1
			fi

			do_upgrade $expected_version

			until curl --silent --user "{{ .Username }}":"${{ .PasswordEnvVar }}" "{{ .URL }}/_cluster/health/{{ .Indices.Visibility }}" | jq --exit-status '.status=="green" | .'; do
				echo "Waiting for Elasticsearch index {{ .Indices.Visibility }} become green."
				sleep 1
			done
			{{ template "scripts" . }}
		`),
	}
)

type (
	baseData struct {
		MTLSProvider string
	}

	createDatabase struct {
		baseData
		Tool           string
		ConnectionArgs string
		DatabaseName   string
	}

	createKeyspace struct {
		baseData
		Tool           string
		ConnectionArgs string
		KeyspaceName   string
	}

	setupSchemaData struct {
		baseData
		Tool           string
		ConnectionArgs string
		InitialVersion string
	}

	updateSchemaData struct {
		baseData
		Tool           string
		ConnectionArgs string
		SchemaDir      string
	}

	advancedVisibilityData struct {
		baseData
		Version        string
		URL            string
		Username       string
		PasswordEnvVar string
		Indices        v1alpha1.ElasticsearchIndices
	}
)

var (
	proxyShutdownScriptsContent = dedent.Dedent(`
		{{- define "scripts" -}}
		{{- if eq .MTLSProvider "linkerd" -}}
		x=$?
		curl -X POST http://localhost:4191/shutdown
		exit $x
		{{- end -}}
		{{- if eq .MTLSProvider "istio" -}}
		x=$?
		curl -sf -XPOST http://127.0.0.1:15020/quitquitquit
		exit $x
		{{- end -}}
		{{- end -}}
	`)
)

func init() {
	for name, content := range templatesContent {
		templates[name] = template.Must(template.New(name).Parse(proxyShutdownScriptsContent))
		template.Must(templates[name].Parse(content))
	}
}
