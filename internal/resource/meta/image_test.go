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

package meta_test

import (
	"testing"

	"github.com/alexandrevilain/temporal-operator/internal/resource/meta"
	"github.com/stretchr/testify/assert"
)

func TestContainerImage(t *testing.T) {
	const (
		repository = "temporalio/server"
		digest     = "sha256:459aa877ef82c9a7f0ce688f2451ec8e576142c4efefecb2988d2f5f8b12699e"
	)

	for name, tc := range map[string]struct {
		defaultTag  string
		tagOverride string
		digest      string
		expected    string
	}{
		"default tag only": {
			defaultTag: "1.28.2",
			expected:   "temporalio/server:1.28.2",
		},
		"tag override replaces default": {
			defaultTag:  "1.28.2",
			tagOverride: "1.28.2-custom",
			expected:    "temporalio/server:1.28.2-custom",
		},
		"digest pins default tag": {
			defaultTag: "1.28.2",
			digest:     digest,
			expected:   "temporalio/server:1.28.2@" + digest,
		},
		"tag override and digest combined": {
			defaultTag:  "1.28.2",
			tagOverride: "1.28.2-custom",
			digest:      digest,
			expected:    "temporalio/server:1.28.2-custom@" + digest,
		},
	} {
		t.Run(name, func(t *testing.T) {
			got := meta.ContainerImage(repository, tc.defaultTag, tc.tagOverride, tc.digest)
			assert.Equal(t, tc.expected, got)
		})
	}
}
