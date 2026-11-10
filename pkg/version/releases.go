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
)

// latestPatchReleases holds, for every minor version in SupportedVersionsRange,
// the newest patch release the operator steps through during an automatic
// upgrade (spec.upgrade.strategy: Automatic).
//
// The operator has no way to discover releases at runtime (clusters may pull
// from mirrors or be air-gapped), so this list is the release catalog. Every
// entry is a published GitHub release of temporalio/temporal with a matching
// temporalio/server image, and the admin-tools tag DefaultAdminToolTag returns
// for it exists on Docker Hub. When support for a new minor is added, append its
// latest patch here; when a newer patch of an existing minor ships, bump it.
// TestLatestPatchReleases enforces one non-broken entry per supported minor.
var latestPatchReleases = []*Version{
	MustNewVersionFromString("1.14.6"),
	MustNewVersionFromString("1.15.2"),
	MustNewVersionFromString("1.16.3"),
	MustNewVersionFromString("1.17.6"),
	MustNewVersionFromString("1.18.5"),
	MustNewVersionFromString("1.19.1"),
	MustNewVersionFromString("1.20.5"),
	MustNewVersionFromString("1.21.6"),
	MustNewVersionFromString("1.22.7"),
	MustNewVersionFromString("1.23.1"),
	MustNewVersionFromString("1.24.3"),
	MustNewVersionFromString("1.25.2"),
	MustNewVersionFromString("1.26.3"),
	MustNewVersionFromString("1.27.4"),
	MustNewVersionFromString("1.28.4"),
	MustNewVersionFromString("1.29.7"),
	MustNewVersionFromString("1.30.9"),
	MustNewVersionFromString("1.31.5"),
	MustNewVersionFromString("1.32.2"),
}

// LatestPatchReleases returns a copy of the release catalog, oldest first.
func LatestPatchReleases() []*Version {
	out := make([]*Version, 0, len(latestPatchReleases))
	for _, v := range latestPatchReleases {
		out = append(out, v.DeepCopy())
	}
	return out
}

// LatestPatch returns the newest known patch release of major.minor.
func LatestPatch(major, minor uint64) (*Version, bool) {
	for _, v := range latestPatchReleases {
		if v.Major() == major && v.Minor() == minor {
			return v.DeepCopy(), true
		}
	}
	return nil, false
}

// LatestKnownRelease returns the newest release in the catalog. It is the
// version spec.upgrade.target: Latest resolves to.
func LatestKnownRelease() *Version {
	return latestPatchReleases[len(latestPatchReleases)-1].DeepCopy()
}

// NextUpgradeStep returns the version to roll out next when moving a cluster
// that runs `from` towards `target`, honouring Temporal's rule that the server
// is upgraded one minor version at a time.
//
// When target is at most one minor ahead of from (including patch-only
// changes), the step is target itself. Otherwise it is the newest known,
// non-broken patch of the next minor.
func NextUpgradeStep(from, target *Version) (*Version, error) {
	if from == nil || from.Version == nil || target == nil || target.Version == nil {
		return nil, errors.New("both the running and the target version are required")
	}
	if target.Major() != from.Major() {
		return nil, fmt.Errorf("can't upgrade across major versions (from %s to %s)", from, target)
	}
	if target.LessThan(from) {
		return nil, fmt.Errorf("target version %s is lower than the running version %s", target, from)
	}
	if target.Minor() <= from.Minor()+1 {
		return target.DeepCopy(), nil
	}

	next, ok := LatestPatch(from.Major(), from.Minor()+1)
	if !ok {
		return nil, fmt.Errorf("no known release for %d.%d, can't upgrade from %s to %s", from.Major(), from.Minor()+1, from, target)
	}
	if IsBrokenRelease(next) {
		// The catalog is guarded by tests, this is a last line of defence.
		return nil, fmt.Errorf("catalog release %s is marked as broken", next)
	}
	return next, nil
}

// UpgradePath returns every version the cluster will run when moving from
// `from` to `target`, in order, excluding from and including target. It is
// empty when from already equals target.
func UpgradePath(from, target *Version) ([]*Version, error) {
	path := []*Version{}
	current := from
	// Each step advances at least one minor version (or lands on target), so
	// the path is bounded by the catalog size; the bound guards against a
	// future bug that fails to advance.
	for range len(latestPatchReleases) + 2 {
		if current.Equal(target.Version) {
			return path, nil
		}
		next, err := NextUpgradeStep(current, target)
		if err != nil {
			return nil, err
		}
		path = append(path, next)
		current = next
	}
	return nil, fmt.Errorf("can't compute an upgrade path from %s to %s", from, target)
}
