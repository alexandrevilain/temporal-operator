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

// +kubebuilder:object:generate=true
package version

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/Masterminds/semver/v3"
)

var (
	// SupportedVersionsRange holds all supported temporal versions.
	SupportedVersionsRange  = mustNewConstraint(">= 1.14.0 < 1.32.0")
	ForbiddenBrokenReleases = []*Version{
		// v1.21.0 is reported as broken, see: https://github.com/temporalio/temporal/releases/tag/v1.21.0
		MustNewVersionFromString("1.21.0"),
		// v1.21.1 is reported as broken, see: https://github.com/temporalio/temporal/releases/tag/v1.21.1
		MustNewVersionFromString("1.21.1"),
		// v1.24.0 is reported as broken, see: https://github.com/temporalio/temporal/releases/tag/v1.24.0
		MustNewVersionFromString("1.24.0"),
		// v1.27.0 is reported as broken, see: https://github.com/temporalio/temporal/releases/tag/v1.27.0
		MustNewVersionFromString("1.27.0"),
		// The releases below are retracted by upstream's own go.mod, see the
		// retract block in https://github.com/temporalio/temporal/blob/v1.31.2/go.mod
		// None of them has a published GitHub release or container image.
		// v1.26.0 was "published accidentally"; use v1.26.2+.
		MustNewVersionFromString("1.26.0"),
		// v1.26.1 "contains retractions only"; use v1.26.2+.
		MustNewVersionFromString("1.26.1"),
		// v1.30.0 is retracted; use v1.30.1+.
		MustNewVersionFromString("1.30.0"),
	}
	V1_18_0 = MustNewVersionFromString("1.18.0") //nolint:staticcheck,revive
	V1_20_0 = MustNewVersionFromString("1.20.0") //nolint:staticcheck,revive
	V1_21_0 = MustNewVersionFromString("1.21.0") //nolint:staticcheck,revive
	V1_22_0 = MustNewVersionFromString("1.22.0") //nolint:staticcheck,revive
	V1_23_0 = MustNewVersionFromString("1.23.0") //nolint:staticcheck,revive
	V1_24_0 = MustNewVersionFromString("1.24.0") //nolint:staticcheck,revive
	V1_25_0 = MustNewVersionFromString("1.25.0") //nolint:staticcheck,revive
	V1_30_0 = MustNewVersionFromString("1.30.0") //nolint:staticcheck,revive
	V1_31_0 = MustNewVersionFromString("1.31.0") //nolint:staticcheck,revive
)

// Version is a wrapper around semver.Version which supports correct
// marshaling to YAML and JSON. In particular, it marshals into strings.
// +kubebuilder:validation:Type=string
type Version struct {
	*semver.Version
}

// IsBrokenRelease reports whether v is one of the releases the operator refuses
// to run.
func IsBrokenRelease(v *Version) bool {
	for _, broken := range ForbiddenBrokenReleases {
		if v.Equal(broken.Version) {
			return true
		}
	}
	return false
}

// NextNonBrokenPatch returns the first patch release after v that is not itself
// marked broken.
//
// Broken releases can be consecutive: v1.26.0 and v1.26.1 are both retracted
// upstream. Suggesting a plain IncPatch would send users from 1.26.0 to 1.26.1,
// which the webhook rejects in turn.
func (v *Version) NextNonBrokenPatch() *Version {
	candidate := v
	// ForbiddenBrokenReleases is finite so this always terminates; the bound is
	// only a guard against a future list that fails to advance.
	for range len(ForbiddenBrokenReleases) + 1 {
		next := candidate.IncPatch()
		candidate = &Version{Version: &next}
		if !IsBrokenRelease(candidate) {
			break
		}
	}
	return candidate
}

// Validate checks if the current version is in the supported temporal cluster
// version range.
func (v *Version) Validate() error {
	inRange := SupportedVersionsRange.Check(v.Version)
	if !inRange {
		return errors.New("provided version not in the supported range")
	}
	return nil
}

// ToUnstructured implements the value.UnstructuredConverter interface.
func (v *Version) ToUnstructured() any {
	if v == nil {
		return nil
	}
	return v.Version.String()
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (v *Version) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	parsed, err := semver.NewVersion(str)
	if err != nil {
		return err
	}
	v.Version = parsed
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (v Version) MarshalJSON() ([]byte, error) {
	if v.Version == nil {
		return []byte("0.0.0"), nil
	}
	return json.Marshal(v.Version.String())
}

// GreaterOrEqual returns whenever version is greater or equal than the provided version.
func (v *Version) GreaterOrEqual(compare *Version) bool {
	return checkConstraint(v, ">= %s", compare)
}

// LessThan returns whenever version is less than the provided version.
func (v *Version) LessThan(compare *Version) bool {
	return checkConstraint(v, "< %s", compare)
}

// constraintCache memoizes compiled semver constraints, keyed by their textual
// form.
//
// GreaterOrEqual and LessThan are called repeatedly during a single reconcile —
// once per datastore, once per deployment builder, once per config section —
// and nearly always against the same handful of package-level version
// constants, so recompiling the identical constraint every time is pure waste.
var constraintCache sync.Map

func checkConstraint(v *Version, format string, compare *Version) bool {
	expr := fmt.Sprintf(format, compare.String())

	if cached, ok := constraintCache.Load(expr); ok {
		return cached.(*semver.Constraints).Check(v.Version)
	}

	c, err := semver.NewConstraint(expr)
	if err != nil {
		// compare always renders as a valid semver, so this is unreachable.
		// It is handled anyway because the previous code discarded the error
		// and would have dereferenced a nil *Constraints if that ever changed.
		return false
	}

	constraintCache.Store(expr, c)
	return c.Check(v.Version)
}

// UpgradeConstraint returns the Temporal Server upgrade constraint.
// Users should upgrade Temporal Server sequentially.
// The returned constraint ensures that, we're could only upgrade to upgrade from v1.n.x to v1.n+1.x.
func (v *Version) UpgradeConstraint() (*semver.Constraints, error) {
	incrementedMinor := v.IncMinor()
	constraint := fmt.Sprintf(">= %d.%d.%d <= %d.%d", v.Major(), v.Minor(), v.Patch(), v.Major(), incrementedMinor.Minor())
	return semver.NewConstraint(constraint)
}

// OpenAPISchemaType is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
//
// See: https://github.com/kubernetes/kube-openapi/tree/master/pkg/generators
func (Version) OpenAPISchemaType() []string { return []string{"string"} }

// OpenAPISchemaFormat is used by the kube-openapi generator when constructing
// the OpenAPI spec of this type.
func (Version) OpenAPISchemaFormat() string { return "" }

func NewVersionFromString(v string) (*Version, error) {
	version, err := semver.NewVersion(v)
	return &Version{Version: version}, err
}

func MustNewVersionFromString(v string) *Version {
	version, err := NewVersionFromString(v)
	if err != nil {
		panic(err)
	}
	return version
}

func mustNewConstraint(constraint string) *semver.Constraints {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		panic(err)
	}
	return c
}
