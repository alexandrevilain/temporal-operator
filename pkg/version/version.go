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

package version

import (
	"fmt"

	"github.com/alexandrevilain/temporal-operator/api/v1alpha1"
	"github.com/blang/semver/v4"
)

// SchemaVersions is temporal schemas versions by datastore type.
type SchemaVersions map[v1alpha1.DatastoreType]semver.Version

// Version represents a temporal version range.
type Version struct {
	Range          semver.Range
	SchemaVersions SchemaVersions
}

// SupportedVersions holds all supported temporal versions.
var SupportedVersions = []Version{
	{
		Range: semver.MustParseRange(">= 1.16.0"),
		SchemaVersions: SchemaVersions{
			v1alpha1.PostgresSQLDatastore: semver.MustParse("1.8.0"),
			v1alpha1.MySQLDatastore:       semver.MustParse("1.8.0"),
			v1alpha1.CassandraDatastore:   semver.MustParse("1.7.0"),
		},
	},
	{
		Range: semver.MustParseRange(">= 1.14.0 <1.16.0"),
		SchemaVersions: SchemaVersions{
			v1alpha1.PostgresSQLDatastore: semver.MustParse("1.7.0"),
			v1alpha1.MySQLDatastore:       semver.MustParse("1.7.0"),
			v1alpha1.CassandraDatastore:   semver.MustParse("1.6.0"),
		},
	},
	// Releases < 1.14 are not supported by this operator.
}

func findMatchingSupportedVersion(v semver.Version) (*Version, bool) {
	for _, version := range SupportedVersions {
		if version.Range(v) {
			return &version, true
		}
	}
	return nil, false
}

// ParseAndValidateTemporalVersion parses the provided version and determines if it's a supported one.
func ParseAndValidateTemporalVersion(v string) (semver.Version, error) {
	version, err := semver.Parse(v)
	if err != nil {
		return semver.Version{}, fmt.Errorf("can't parse version: %w", err)
	}

	_, found := findMatchingSupportedVersion(version)
	if !found {
		return semver.Version{}, fmt.Errorf("%s is not a supported version", v)
	}

	return version, nil
}

// GetExpectedSchemaVersions returns the SchemaVersion for the provided temporal version.
func GetExpectedSchemaVersions(v semver.Version) (SchemaVersions, error) {
	version, found := findMatchingSupportedVersion(v)
	if !found {
		return SchemaVersions{}, fmt.Errorf("%s is not a supported version", v)
	}
	return version.SchemaVersions, nil
}

// Parse is a utility function to parse the provided version.
func Parse(v string) (semver.Version, error) {
	version, err := semver.Parse(v)
	if err != nil {
		return semver.Version{}, fmt.Errorf("can't parse version: %w", err)
	}
	return version, nil
}
