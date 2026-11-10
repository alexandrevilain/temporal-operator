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

package version_test

import (
	"testing"

	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLatestPatchReleases guards the release catalog automatic upgrades step through: exactly one
// non-broken entry per minor version of the supported range, oldest first, so a supported minor
// can never be skipped and a broken release never rolled out.
func TestLatestPatchReleases(t *testing.T) {
	releases := version.LatestPatchReleases()
	require.NotEmpty(t, releases)

	first := releases[0]
	assert.Equal(t, uint64(1), first.Major())
	assert.Equal(t, uint64(14), first.Minor(), "the catalog must start at the lowest supported minor")
	assert.False(t, version.SupportedVersionsRange.Check(version.MustNewVersionFromString("1.13.99").Version),
		"the supported range moved below the catalog; add the missing minors")

	for i, v := range releases {
		assert.NoError(t, v.Validate(), "%s is outside the supported range", v)
		assert.False(t, version.IsBrokenRelease(v), "%s is marked as broken", v)
		assert.Empty(t, v.Prerelease(), "%s is a pre-release", v)
		if i > 0 {
			prev := releases[i-1]
			assert.Equal(t, prev.Major(), v.Major())
			assert.Equal(t, prev.Minor()+1, v.Minor(), "minor versions must be contiguous: %s follows %s", v, prev)
		}
	}

	latest := version.LatestKnownRelease()
	assert.True(t, latest.Equal(releases[len(releases)-1].Version))
	next := latest.IncMinor()
	assert.False(t, version.SupportedVersionsRange.Check(&next),
		"the supported range includes %d.%d but the catalog does not; add its latest patch", next.Major(), next.Minor())
}

func TestLatestPatch(t *testing.T) {
	v, ok := version.LatestPatch(1, 26)
	require.True(t, ok)
	assert.Equal(t, "1.26.3", v.String())

	_, ok = version.LatestPatch(1, 99)
	assert.False(t, ok)

	// The returned value is a copy: mutating it must not alter the catalog.
	*v.Version = *version.MustNewVersionFromString("1.0.0").Version
	again, _ := version.LatestPatch(1, 26)
	assert.Equal(t, "1.26.3", again.String())
}

func TestNextUpgradeStep(t *testing.T) {
	tests := map[string]struct {
		from, target string
		expected     string
		expectedErr  string
	}{
		"patch upgrade lands on target":         {from: "1.24.1", target: "1.24.3", expected: "1.24.3"},
		"adjacent minor lands on target":        {from: "1.24.3", target: "1.25.0", expected: "1.25.0"},
		"same version":                          {from: "1.32.2", target: "1.32.2", expected: "1.32.2"},
		"far target steps to next latest patch": {from: "1.24.3", target: "1.32.2", expected: "1.25.2"},
		"skips retracted 1.26.0 and 1.26.1":     {from: "1.25.2", target: "1.28.0", expected: "1.26.3"},
		"skips broken 1.27.0":                   {from: "1.26.3", target: "1.28.0", expected: "1.27.4"},
		"downgrade is refused":                  {from: "1.25.2", target: "1.24.3", expectedErr: "lower than the running version"},
		"major change is refused":               {from: "1.25.2", target: "2.0.0", expectedErr: "across major versions"},
		"unknown intermediate minor":            {from: "1.32.2", target: "1.40.0", expectedErr: "no known release for 1.33"},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			got, err := version.NextUpgradeStep(version.MustNewVersionFromString(test.from), version.MustNewVersionFromString(test.target))
			if test.expectedErr != "" {
				require.Error(tt, err)
				assert.Contains(tt, err.Error(), test.expectedErr)
				return
			}
			require.NoError(tt, err)
			assert.Equal(tt, test.expected, got.String())
		})
	}
}

func TestUpgradePath(t *testing.T) {
	path, err := version.UpgradePath(version.MustNewVersionFromString("1.24.3"), version.MustNewVersionFromString("1.32.2"))
	require.NoError(t, err)

	got := make([]string, 0, len(path))
	for _, v := range path {
		got = append(got, v.String())
	}
	assert.Equal(t, []string{"1.25.2", "1.26.3", "1.27.4", "1.28.4", "1.29.7", "1.30.9", "1.31.5", "1.32.2"}, got)

	for _, v := range path {
		assert.False(t, version.IsBrokenRelease(v), "%s is broken", v)
	}

	// A target that is not the latest patch of its minor is honoured exactly.
	path, err = version.UpgradePath(version.MustNewVersionFromString("1.29.1"), version.MustNewVersionFromString("1.31.2"))
	require.NoError(t, err)
	require.Len(t, path, 2)
	assert.Equal(t, "1.30.9", path[0].String())
	assert.Equal(t, "1.31.2", path[1].String())

	path, err = version.UpgradePath(version.MustNewVersionFromString("1.31.2"), version.MustNewVersionFromString("1.31.2"))
	require.NoError(t, err)
	assert.Empty(t, path)

	_, err = version.UpgradePath(version.MustNewVersionFromString("1.31.2"), version.MustNewVersionFromString("1.30.0"))
	assert.Error(t, err)
}
