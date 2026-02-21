package version

import (
	"runtime"
	"strings"
)

var (
	// Version is the current version of check-image, injected at build time via ldflags.
	Version = "dev"
	// Commit is the short git commit hash, injected at build time via ldflags.
	Commit = "none"
	// BuildDate is the build timestamp in RFC3339 format, injected at build time via ldflags.
	BuildDate = "unknown"
)

// BuildInfo holds all version and build metadata.
type BuildInfo struct {
	Version   string
	Commit    string
	BuildDate string
	GoVersion string
	Platform  string
}

// Get returns the current version string.
func Get() string {
	return Version
}

// GetBuildInfo returns the full build information including version, commit, build date,
// Go version and target platform.
func GetBuildInfo() BuildInfo {
	v := strings.TrimSpace(Version)
	if v == "" {
		v = "dev"
	}
	return BuildInfo{
		Version:   v,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}
