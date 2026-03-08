package secrets

import (
	"path/filepath"
	"slices"
	"strings"
)

// isExcluded checks if a value is in the exclusion list (case-sensitive)
func isExcluded(value string, exclusionList []string) bool {
	return slices.Contains(exclusionList, value)
}

// isPathExcluded checks if a path matches any exclusion patterns
func isPathExcluded(path string, excludedPatterns []string) bool {
	for _, pattern := range excludedPatterns {
		if isDirectoryPattern(path, pattern) || isGlobPattern(path, pattern) {
			return true
		}
	}
	return false
}

// isDirectoryPattern reports whether path falls under a directory exclusion pattern
// (patterns ending with "/**", e.g. "/usr/share/**").
func isDirectoryPattern(path, pattern string) bool {
	prefix, ok := strings.CutSuffix(pattern, "/**")
	if !ok {
		return false
	}
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

// isGlobPattern reports whether path matches an exact path or glob pattern,
// trying both the full path and the basename.
func isGlobPattern(path, pattern string) bool {
	if matched, err := filepath.Match(pattern, path); err == nil && matched {
		return true
	}
	matched, err := filepath.Match(pattern, filepath.Base(path))
	return err == nil && matched
}

// matchesFilePattern checks if a file path matches any sensitive patterns
func matchesFilePattern(path string, patterns []string) (bool, string) {
	basename := filepath.Base(path)

	for _, pattern := range patterns {
		// Try matching against basename first
		matched, err := filepath.Match(pattern, basename)
		if err == nil && matched {
			return true, describePattern(pattern)
		}

		// Try matching against full path
		matched, err = filepath.Match(pattern, path)
		if err == nil && matched {
			return true, describePattern(pattern)
		}

		// Handle path-based patterns (e.g., .aws/credentials).
		// Use HasSuffix with a directory separator to avoid false positives where
		// the pattern appears as a prefix of another filename (e.g., ".aws/credentials"
		// must not match ".aws/credentials-backup").
		if strings.Contains(pattern, "/") {
			if strings.HasSuffix(path, "/"+pattern) || strings.HasSuffix(path, pattern) {
				return true, describePattern(pattern)
			}
		}
	}

	return false, ""
}

// describePattern provides a human-readable description for a pattern
func describePattern(pattern string) string {
	if desc, ok := DefaultFilePatterns[pattern]; ok {
		return desc
	}
	return "sensitive file pattern"
}
