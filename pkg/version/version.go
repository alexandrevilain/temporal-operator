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
	"errors"
	"fmt"

	"github.com/blang/semver/v4"
)

// SupportedVersionsRange holds all supported temporal versions.
var SupportedVersionsRange = semver.MustParseRange(">= 1.14.0 < 1.18.0")

// ParseAndValidateTemporalVersion parses the provided version and determines if it's a supported one.
func ParseAndValidateTemporalVersion(v string) (semver.Version, error) {
	version, err := semver.Parse(v)
	if err != nil {
		return semver.Version{}, fmt.Errorf("can't parse version: %w", err)
	}

	inRange := SupportedVersionsRange(version)
	if !inRange {
		return semver.Version{}, errors.New("provided version not in the supported range")
	}

	return version, nil
}
