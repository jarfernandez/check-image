package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{"default version", "dev", "dev"},
		{"custom version", "v1.0.0", "v1.0.0"},
		{"empty version", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalVersion := Version
			defer func() { Version = originalVersion }()

			Version = tt.version
			assert.Equal(t, tt.expected, Get())
		})
	}
}
