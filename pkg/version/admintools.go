package version

import (
	"fmt"
	"regexp"

	"github.com/Masterminds/semver/v3"
)

// DefaultAdminToolTag returns the tag of the admin tools image for the given version.
// It's required as 1.24.x had really bad image tagging.
func DefaultAdminToolTag(version *Version) string {
	// Particular case for >= 1.23.0 but < 1.24.0
	// Need this because 1.23.1 tag does not exist
	if version.GreaterOrEqual(V1_23_0) && version.LessThan(V1_24_0) {
		return "1.23.1.1-tctl-1.18.1-cli-0.12.0"
	}

	// Particular case for >= 1.24.0 but < 1.25.0
	if version.GreaterOrEqual(V1_24_0) && version.LessThan(V1_25_0) {
		return "1.24.2-tctl-1.18.1-cli-1.0.0"
	}

	// Particular case for >= 1.25 because the admin tools image tag doesn't
	// contains patch version (or it has the same bad naming as for 1.24.x).
	if version.GreaterOrEqual(V1_25_0) {
		return fmt.Sprintf("%d.%d", version.Major(), version.Minor())
	}

	return version.String()
}

// historicalDefaultAdminToolTags are fixed admin tools tags the operator has defaulted to for the
// 1.23 and 1.24 lines, in this release or an earlier one.
var historicalDefaultAdminToolTags = map[string]struct{}{
	"1.23.1.1-tctl-1.18.1-cli-0.12.0": {},
	"1.24.2-tctl-1.18.1-cli-0.13.2":   {},
	"1.24.2-tctl-1.18.1-cli-1.0.0":    {},
}

var majorMinorAdminToolTag = regexp.MustCompile(`^\d+\.\d+$`)

// IsKnownDefaultAdminToolTag reports whether tag has the shape of an admin tools tag the operator
// defaults to for some server version: one of the fixed tags used for the 1.23 and 1.24 lines, the
// "major.minor" form used from 1.25 on, or a plain server version below 1.23, which is what the
// operator used before the admin tools tagging became irregular.
//
// Automatic upgrades treat such a tag as "follow the server version": it is re-resolved through
// DefaultAdminToolTag for every step, instead of pinning the admin tools to the version the tag was
// copied from. Any other tag is a deliberate user choice and is rendered as is.
func IsKnownDefaultAdminToolTag(tag string) bool {
	if _, ok := historicalDefaultAdminToolTags[tag]; ok {
		return true
	}
	if majorMinorAdminToolTag.MatchString(tag) {
		return true
	}
	v, err := semver.StrictNewVersion(tag)
	if err != nil || v.Prerelease() != "" || v.Metadata() != "" {
		return false
	}
	return v.Major() == 1 && v.Minor() < 23
}
