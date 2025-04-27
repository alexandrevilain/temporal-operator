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
