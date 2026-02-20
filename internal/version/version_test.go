package version

import (
	"runtime"
	"strings"
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

func TestGetBuildInfo(t *testing.T) {
	tests := []struct {
		name              string
		version           string
		commit            string
		buildDate         string
		expectedVersion   string
		expectedCommit    string
		expectedBuildDate string
	}{
		{
			name:              "full build info",
			version:           "v1.2.3",
			commit:            "abc1234",
			buildDate:         "2026-02-18T12:34:56Z",
			expectedVersion:   "v1.2.3",
			expectedCommit:    "abc1234",
			expectedBuildDate: "2026-02-18T12:34:56Z",
		},
		{
			name:              "dev build defaults",
			version:           "dev",
			commit:            "none",
			buildDate:         "unknown",
			expectedVersion:   "dev",
			expectedCommit:    "none",
			expectedBuildDate: "unknown",
		},
		{
			name:              "empty version falls back to dev",
			version:           "",
			commit:            "abc1234",
			buildDate:         "2026-02-18T12:34:56Z",
			expectedVersion:   "dev",
			expectedCommit:    "abc1234",
			expectedBuildDate: "2026-02-18T12:34:56Z",
		},
		{
			name:              "whitespace version falls back to dev",
			version:           "  ",
			commit:            "abc1234",
			buildDate:         "2026-02-18T12:34:56Z",
			expectedVersion:   "dev",
			expectedCommit:    "abc1234",
			expectedBuildDate: "2026-02-18T12:34:56Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalVersion := Version
			originalCommit := Commit
			originalBuildDate := BuildDate
			defer func() {
				Version = originalVersion
				Commit = originalCommit
				BuildDate = originalBuildDate
			}()

			Version = tt.version
			Commit = tt.commit
			BuildDate = tt.buildDate

			info := GetBuildInfo()

			assert.Equal(t, tt.expectedVersion, info.Version)
			assert.Equal(t, tt.expectedCommit, info.Commit)
			assert.Equal(t, tt.expectedBuildDate, info.BuildDate)
			// GoVersion and Platform come from runtime — verify they are non-empty and sensible
			assert.True(t, strings.HasPrefix(info.GoVersion, "go"), "GoVersion should start with 'go', got: %s", info.GoVersion)
			assert.Equal(t, runtime.GOOS+"/"+runtime.GOARCH, info.Platform)
		})
	}
}
