// Licensed to Alexandre VILAIN under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Alexandre VILAIN licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
package version_test

import (
	"testing"

	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgradeConstraint(t *testing.T) {
	tests := map[string]struct {
		version        *version.Version
		upgradeVersion *version.Version
		expectedAllow  bool
	}{
		"go to next version release": {
			version:        version.MustNewVersionFromString("1.17.5"),
			upgradeVersion: version.MustNewVersionFromString("1.18.0"),
			expectedAllow:  true,
		},
		"go to next version, with latest patch": {
			version:        version.MustNewVersionFromString("1.17.5"),
			upgradeVersion: version.MustNewVersionFromString("1.18.3"),
			expectedAllow:  true,
		},
		"go to 2 next versions": {
			version:        version.MustNewVersionFromString("1.17.5"),
			upgradeVersion: version.MustNewVersionFromString("1.19.0"),
			expectedAllow:  false,
		},
		"go to 2 next versions, with latest patch": {
			version:        version.MustNewVersionFromString("1.17.5"),
			upgradeVersion: version.MustNewVersionFromString("1.19.4"),
			expectedAllow:  false,
		},
		"go to older version": {
			version:        version.MustNewVersionFromString("1.17.5"),
			upgradeVersion: version.MustNewVersionFromString("1.16.6"),
			expectedAllow:  false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			constraint, err := test.version.UpgradeConstraint()
			require.NoError(tt, err)

			result := constraint.Check(test.upgradeVersion.Version)
			assert.Equal(tt, test.expectedAllow, result)
		})
	}
}

func TestVersionGreaterOrEqual(t *testing.T) {
	tests := map[string]struct {
		version1 *version.Version
		version2 *version.Version
		expected bool
	}{
		"version is less than": {
			version1: version.MustNewVersionFromString("1.17.5"),
			version2: version.MustNewVersionFromString("1.18.0"),
			expected: false,
		},
		"version is equal": {
			version1: version.MustNewVersionFromString("1.17.5"),
			version2: version.MustNewVersionFromString("1.17.5"),
			expected: true,
		},
		"version is greater than": {
			version1: version.MustNewVersionFromString("1.17.5"),
			version2: version.MustNewVersionFromString("1.16.0"),
			expected: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			result := test.version1.GreaterOrEqual(test.version2)
			assert.Equal(tt, test.expected, result)
		})
	}
}

func TestVersionLessThan(t *testing.T) {
	tests := map[string]struct {
		version1 *version.Version
		version2 *version.Version
		expected bool
	}{
		"version is less than": {
			version1: version.MustNewVersionFromString("1.17.5"),
			version2: version.MustNewVersionFromString("1.18.0"),
			expected: true,
		},
		"version is equal": {
			version1: version.MustNewVersionFromString("1.17.5"),
			version2: version.MustNewVersionFromString("1.17.5"),
			expected: false,
		},
		"version is greater than": {
			version1: version.MustNewVersionFromString("1.17.5"),
			version2: version.MustNewVersionFromString("1.16.0"),
			expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			result := test.version1.LessThan(test.version2)
			assert.Equal(tt, test.expected, result)
		})
	}
}
