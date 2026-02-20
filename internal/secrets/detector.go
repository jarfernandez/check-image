package secrets

import (
	"archive/tar"
	"fmt"
	cr "github.com/google/go-containerregistry/pkg/v1"
	log "github.com/sirupsen/logrus"
	"io"
	"path/filepath"
	"slices"
	"strings"
)

// EnvVarFinding represents a sensitive environment variable finding
type EnvVarFinding struct {
	Name        string
	Description string
}

// FileFinding represents a sensitive file finding
type FileFinding struct {
	Path        string
	LayerIndex  int
	Description string
}

// DetectionResult holds all findings from secrets detection
type DetectionResult struct {
	EnvVarFindings []EnvVarFinding
	FileFindings   []FileFinding
}

// CheckEnvironmentVariables scans environment variables for sensitive patterns
func CheckEnvironmentVariables(envVars []string, policy *Policy) []EnvVarFinding {
	if !policy.CheckEnvVars {
		return nil
	}

	var findings []EnvVarFinding
	patterns := policy.GetEnvPatterns()

	for _, envVar := range envVars {
		// Environment variables are in "KEY=VALUE" format
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 0 {
			continue
		}

		varName := parts[0]

		// Check if this variable is in the exclusion list
		if isExcluded(varName, policy.ExcludedEnvVars) {
			log.Debugf("Skipping excluded environment variable: %s", varName)
			continue
		}

		// Check if the variable name matches any sensitive patterns (case-insensitive)
		varNameLower := strings.ToLower(varName)
		for _, pattern := range patterns {
			patternLower := strings.ToLower(pattern)
			if strings.Contains(varNameLower, patternLower) {
				findings = append(findings, EnvVarFinding{
					Name:        varName,
					Description: "sensitive pattern detected",
				})
				log.Debugf("Found sensitive environment variable: %s (matches pattern: %s)", varName, pattern)
				break
			}
		}
	}

	return findings
}

// CheckFilesInLayers scans all layers for files matching sensitive patterns
func CheckFilesInLayers(image cr.Image, policy *Policy) ([]FileFinding, error) {
	if !policy.CheckFiles {
		return nil, nil
	}

	layers, err := image.Layers()
	if err != nil {
		return nil, fmt.Errorf("error getting image layers: %w", err)
	}

	var allFindings []FileFinding
	seenPaths := make(map[string]bool) // Deduplication across layers

	for i, layer := range layers {
		log.Debugf("Scanning layer %d/%d", i+1, len(layers))

		findings, err := scanLayer(layer, i, policy)
		if err != nil {
			log.Warnf("Error scanning layer %d: %v", i, err)
			continue
		}

		// Add findings, deduplicating by path
		for _, finding := range findings {
			if !seenPaths[finding.Path] {
				allFindings = append(allFindings, finding)
				seenPaths[finding.Path] = true
			}
		}
	}

	return allFindings, nil
}

// scanLayer scans a single layer for sensitive files
func scanLayer(layer cr.Layer, layerIndex int, policy *Policy) ([]FileFinding, error) {
	rc, err := layer.Uncompressed()
	if err != nil {
		return nil, fmt.Errorf("error uncompressing layer: %w", err)
	}
	defer func() {
		if closeErr := rc.Close(); closeErr != nil {
			log.Warnf("failed to close layer reader: %v", closeErr)
		}
	}()

	var findings []FileFinding
	tarReader := tar.NewReader(rc)
	patterns := policy.GetFilePatterns()

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading tar: %w", err)
		}

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Check if path should be excluded
		if isPathExcluded(header.Name, policy.ExcludedPaths) {
			log.Debugf("Skipping excluded path: %s", header.Name)
			continue
		}

		// Check if file matches any sensitive patterns
		if matchesPattern, description := matchesFilePattern(header.Name, patterns); matchesPattern {
			findings = append(findings, FileFinding{
				Path:        header.Name,
				LayerIndex:  layerIndex,
				Description: description,
			})
			log.Debugf("Found sensitive file in layer %d: %s (%s)", layerIndex, header.Name, description)
		}
	}

	return findings, nil
}

// isExcluded checks if a value is in the exclusion list (case-sensitive)
func isExcluded(value string, exclusionList []string) bool {
	return slices.Contains(exclusionList, value)
}

// isPathExcluded checks if a path matches any exclusion patterns
func isPathExcluded(path string, excludedPatterns []string) bool {
	for _, pattern := range excludedPatterns {
		// Support both exact matches and glob patterns
		if before, ok := strings.CutSuffix(pattern, "/**"); ok {
			// Directory prefix match
			prefix := before
			if strings.HasPrefix(path, prefix+"/") || path == prefix {
				return true
			}
		} else {
			// Exact match or glob pattern
			matched, err := filepath.Match(pattern, path)
			if err == nil && matched {
				return true
			}
			// Also try matching against the full path
			matched, err = filepath.Match(pattern, filepath.Base(path))
			if err == nil && matched {
				return true
			}
		}
	}
	return false
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

		// Handle path-based patterns (e.g., .aws/credentials)
		if strings.Contains(pattern, "/") {
			if strings.Contains(path, pattern) || strings.HasSuffix(path, pattern) {
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
