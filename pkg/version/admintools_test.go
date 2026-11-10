package version_test

import (
	"testing"

	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/stretchr/testify/assert"
)

func TestDefaultAdminToolTag(t *testing.T) {
	tests := []struct {
		name     string
		version  *version.Version
		expected string
	}{
		{
			name:     "Version 1.23.0",
			version:  version.MustNewVersionFromString("1.23.0"),
			expected: "1.23.1.1-tctl-1.18.1-cli-0.12.0",
		},
		{
			name:     "Version 1.23.9",
			version:  version.MustNewVersionFromString("1.23.9"),
			expected: "1.23.1.1-tctl-1.18.1-cli-0.12.0",
		},
		{
			name:     "Version 1.24.1",
			version:  version.MustNewVersionFromString("1.24.1"),
			expected: "1.24.2-tctl-1.18.1-cli-1.0.0",
		},
		{
			name:     "Version 1.24.9",
			version:  version.MustNewVersionFromString("1.24.9"),
			expected: "1.24.2-tctl-1.18.1-cli-1.0.0",
		},
		{
			name:     "Version 1.25.0",
			version:  version.MustNewVersionFromString("1.25.0"),
			expected: "1.25",
		},
		{
			name:     "Version 1.25.5",
			version:  version.MustNewVersionFromString("1.25.5"),
			expected: "1.25",
		},
		{
			name:     "Version 1.26.0",
			version:  version.MustNewVersionFromString("1.26.0"),
			expected: "1.26",
		},
		{
			name:     "Version 1.10.0",
			version:  version.MustNewVersionFromString("1.10.0"),
			expected: "1.10.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, version.DefaultAdminToolTag(tt.version))
		})
	}
}

func TestIsKnownDefaultAdminToolTag(t *testing.T) {
	tests := map[string]bool{
		// Fixed tags of the 1.23 and 1.24 lines, current and historical.
		"1.23.1.1-tctl-1.18.1-cli-0.12.0": true,
		"1.24.2-tctl-1.18.1-cli-1.0.0":    true,
		"1.24.2-tctl-1.18.1-cli-0.13.2":   true,
		// major.minor form used from 1.25 on.
		"1.25": true,
		"1.32": true,
		// Plain server versions, as defaulted below 1.23.
		"1.22.7": true,
		"1.14.6": true,
		// Not operator defaults.
		"":                       false,
		"1.25.2":                 false,
		"latest":                 false,
		"1.24.2-tctl-1.18.1":     false,
		"1.22.7-custom":          false,
		"my-registry-tag":        false,
		"1.32.2-tctl-1.18.1-cli": false,
	}

	for tag, expected := range tests {
		t.Run(tag, func(tt *testing.T) {
			assert.Equal(tt, expected, version.IsKnownDefaultAdminToolTag(tag))
		})
	}

	// Every default the operator produces for a catalog release is recognised as one.
	for _, v := range version.LatestPatchReleases() {
		assert.True(t, version.IsKnownDefaultAdminToolTag(version.DefaultAdminToolTag(v)), "default tag for %s", v)
	}
}
