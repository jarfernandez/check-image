package secrets

import (
	"archive/tar"
	"fmt"
	"io"
	"strings"

	cr "github.com/google/go-containerregistry/pkg/v1"
	log "github.com/sirupsen/logrus"

	"github.com/jarfernandez/check-image/internal/output"
)

// CheckEnvironmentVariables scans environment variables for sensitive patterns
func CheckEnvironmentVariables(envVars []string, policy *Policy) []output.EnvVarFinding {
	if !policy.CheckEnvVars {
		return nil
	}

	var findings []output.EnvVarFinding
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
				findings = append(findings, output.EnvVarFinding{
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
func CheckFilesInLayers(image cr.Image, policy *Policy) ([]output.FileFinding, error) {
	if !policy.CheckFiles {
		return nil, nil
	}

	layers, err := image.Layers()
	if err != nil {
		return nil, fmt.Errorf("error getting image layers: %w", err)
	}

	var allFindings []output.FileFinding
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
func scanLayer(layer cr.Layer, layerIndex int, policy *Policy) ([]output.FileFinding, error) {
	rc, err := layer.Uncompressed()
	if err != nil {
		return nil, fmt.Errorf("error uncompressing layer: %w", err)
	}
	defer func() {
		if closeErr := rc.Close(); closeErr != nil {
			log.Warnf("failed to close layer reader: %v", closeErr)
		}
	}()

	var findings []output.FileFinding
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
			findings = append(findings, output.FileFinding{
				Path:        header.Name,
				LayerIndex:  layerIndex,
				Description: description,
			})
			log.Debugf("Found sensitive file in layer %d: %s (%s)", layerIndex, header.Name, description)
		}
	}

	return findings, nil
}
