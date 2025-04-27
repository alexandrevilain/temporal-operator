package version

import "fmt"

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
