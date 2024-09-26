package version

// PersistenceJobTag returns the tag of the persistence job image for the given version.
// It's required as 1.24.x had really bad image tagging.
func PersistenceJobTag(version *Version) string {
	// Particular case for >= 1.24.0 but < 1.25.0
	if version.GreaterOrEqual(V1_24_0) && version.LessThan(V1_25_0) {
		return "1.24.2-tctl-1.18.1-cli-0.13.2" //
	}

	return version.String()
}
