package commands

import (
	"testing"

	"github.com/jarfernandez/check-image/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderResult_TextMode(t *testing.T) {
	tests := []struct {
		name          string
		result        *output.CheckResult
		expectedParts []string
	}{
		{
			name: "Age check",
			result: &output.CheckResult{
				Check:  "age",
				Image:  "nginx:latest",
				Passed: true,
				Details: output.AgeDetails{
					CreatedAt: "2024-01-15",
					AgeDays:   30.5,
				},
				Message: "Image is recent",
			},
			expectedParts: []string{"Checking age", "nginx:latest", "2024-01-15", "30 days", "Image is recent"},
		},
		{
			name: "Size check",
			result: &output.CheckResult{
				Check:  "size",
				Image:  "alpine:latest",
				Passed: true,
				Details: output.SizeDetails{
					LayerCount: 1,
					MaxLayers:  20,
					Layers: []output.LayerInfo{
						{Index: 0, Bytes: 7654321},
					},
					TotalBytes: 7654321,
					TotalMB:    7.65,
					MaxSizeMB:  500,
				},
				Message: "Image size is acceptable",
			},
			expectedParts: []string{"Checking size", "alpine:latest", "Number of layers: 1", "7654321 bytes", "7.65 MB"},
		},
		{
			name: "Ports check",
			result: &output.CheckResult{
				Check:  "ports",
				Image:  "nginx:latest",
				Passed: false,
				Details: output.PortsDetails{
					ExposedPorts:      []int{80, 443, 8080},
					AllowedPorts:      []int{80, 443},
					UnauthorizedPorts: []int{8080},
				},
				Message: "Some ports are not allowed",
			},
			expectedParts: []string{"Checking ports", "nginx:latest", "Exposed ports:", "- 80", "- 8080", "not in the allowed list"},
		},
		{
			name: "Registry check",
			result: &output.CheckResult{
				Check:  "registry",
				Image:  "docker.io/nginx:latest",
				Passed: true,
				Details: output.RegistryDetails{
					Registry: "docker.io",
					Skipped:  false,
				},
				Message: "Registry is trusted",
			},
			expectedParts: []string{"Checking registry", "docker.io/nginx:latest", "Image registry: docker.io", "Registry is trusted"},
		},
		{
			name: "Root user check",
			result: &output.CheckResult{
				Check:   "root-user",
				Image:   "nginx:latest",
				Passed:  true,
				Message: "Image runs as non-root user",
			},
			expectedParts: []string{"Checking if image nginx:latest", "non-root user", "Image runs as non-root user"},
		},
		{
			name: "Secrets check",
			result: &output.CheckResult{
				Check:  "secrets",
				Image:  "myapp:latest",
				Passed: false,
				Details: output.SecretsDetails{
					EnvVarFindings: []output.EnvVarFinding{
						{Name: "API_KEY", Description: "Contains 'key'"},
					},
					FileFindings: []output.FileFinding{
						{LayerIndex: 0, Path: "/root/.ssh/id_rsa", Description: "SSH private key"},
					},
					TotalFindings: 2,
					EnvVarCount:   1,
					FileCount:     1,
				},
				Message: "Secrets found",
			},
			expectedParts: []string{"Checking secrets", "myapp:latest", "Environment Variables:", "API_KEY", "Files with Sensitive Patterns:", "id_rsa", "Total findings: 2"},
		},
		{
			name: "Labels check",
			result: &output.CheckResult{
				Check:  "labels",
				Image:  "nginx:latest",
				Passed: false,
				Details: output.LabelsDetails{
					RequiredLabels: []output.RequiredLabelCheck{
						{Name: "maintainer"},
						{Name: "version", Pattern: "^v?\\d+\\.\\d+\\.\\d+$"},
						{Name: "vendor", Value: "MyCompany"},
					},
					ActualLabels: map[string]string{
						"maintainer": "John Doe",
						"version":    "1.2",
					},
					MissingLabels: []string{"vendor"},
					InvalidLabels: []output.InvalidLabelDetail{
						{Name: "version", ActualValue: "1.2", ExpectedPattern: "^v?\\d+\\.\\d+\\.\\d+$", Reason: "does not match pattern"},
					},
				},
				Message: "Image does not meet label requirements",
			},
			expectedParts: []string{"Checking labels", "nginx:latest", "Required labels:", "maintainer (existence check)", "version (pattern:", "vendor (exact:", "Actual labels", "Missing labels", "Invalid labels"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set text output mode
			OutputFmt = output.FormatText

			captured := captureStdout(t, func() {
				err := renderResult(tt.result)
				require.NoError(t, err)
			})

			for _, part := range tt.expectedParts {
				assert.Contains(t, captured, part)
			}
		})
	}
}

func TestRenderResult_TextMode_UnknownCheck(t *testing.T) {
	// The default branch surfaces check names that have no registered renderer
	// so that a missing case is immediately visible instead of producing silent
	// empty output.
	OutputFmt = output.FormatText

	result := &output.CheckResult{
		Check:   "unknown-future-check",
		Image:   "nginx:latest",
		Passed:  true,
		Message: "all good",
	}

	captured := captureStdout(t, func() {
		err := renderResult(result)
		require.NoError(t, err)
	})

	assert.Contains(t, captured, `no text renderer for check "unknown-future-check"`)
}

func TestRenderResult_TextMode_ErrorResult(t *testing.T) {
	// Regression test for F-02: a result with Error set has Details == nil.
	// renderResult must not panic with a bare type assertion.
	OutputFmt = output.FormatText

	result := &output.CheckResult{
		Check:   "age",
		Image:   "nginx:latest",
		Passed:  false,
		Message: "check failed with error: image not found",
		Error:   "image not found",
		// Details intentionally nil
	}

	captured := captureStdout(t, func() {
		err := renderResult(result)
		require.NoError(t, err)
	})

	assert.Contains(t, captured, "check failed with error: image not found")
}

func TestRenderResult_JSONMode(t *testing.T) {
	// Set JSON output mode
	OutputFmt = output.FormatJSON

	result := &output.CheckResult{
		Check:  "age",
		Image:  "nginx:latest",
		Passed: true,
		Details: output.AgeDetails{
			CreatedAt: "2024-01-15",
			AgeDays:   30.5,
		},
		Message: "Image is recent",
	}

	captured := captureStdout(t, func() {
		err := renderResult(result)
		require.NoError(t, err)
	})

	assert.Contains(t, captured, `"check": "age"`)
	assert.Contains(t, captured, `"image": "nginx:latest"`)
	assert.Contains(t, captured, `"passed": true`)
	assert.Contains(t, captured, `"message": "Image is recent"`)
}

func TestRenderAgeText_ValidImage(t *testing.T) {
	result := &output.CheckResult{
		Check:  "age",
		Image:  "nginx:latest",
		Passed: true,
		Details: output.AgeDetails{
			CreatedAt: "2024-01-15T10:30:00Z",
			AgeDays:   15.5,
		},
		Message: "Image is recent",
	}

	captured := captureStdout(t, func() {
		renderAgeText(result)
	})

	assert.Contains(t, captured, "Checking age of image nginx:latest")
	assert.Contains(t, captured, "Image creation date: 2024-01-15T10:30:00Z")
	assert.Contains(t, captured, "Image age: 16 days")
	assert.Contains(t, captured, "Image is recent")
}

func TestRenderAgeText_OldImage(t *testing.T) {
	result := &output.CheckResult{
		Check:  "age",
		Image:  "old-app:v1",
		Passed: false,
		Details: output.AgeDetails{
			CreatedAt: "2023-01-01T00:00:00Z",
			AgeDays:   400.0,
		},
		Message: "Image is too old",
	}

	captured := captureStdout(t, func() {
		renderAgeText(result)
	})

	assert.Contains(t, captured, "old-app:v1")
	assert.Contains(t, captured, "400 days")
	assert.Contains(t, captured, "Image is too old")
}

func TestRenderSizeText_UnderLimits(t *testing.T) {
	result := &output.CheckResult{
		Check:  "size",
		Image:  "alpine:latest",
		Passed: true,
		Details: output.SizeDetails{
			LayerCount: 1,
			MaxLayers:  20,
			Layers: []output.LayerInfo{
				{Index: 0, Bytes: 5000000},
			},
			TotalBytes: 5000000,
			TotalMB:    5.0,
			MaxSizeMB:  500,
		},
		Message: "Image size is acceptable",
	}

	captured := captureStdout(t, func() {
		renderSizeText(result)
	})

	assert.Contains(t, captured, "alpine:latest")
	assert.Contains(t, captured, "Number of layers: 1")
	assert.Contains(t, captured, "Layer 0: 5000000 bytes")
	assert.Contains(t, captured, "Total size: 5000000 bytes (5.00 MB)")
	assert.Contains(t, captured, "Image size is acceptable")
}

func TestRenderSizeText_ExceedsSizeLimit(t *testing.T) {
	result := &output.CheckResult{
		Check:  "size",
		Image:  "large-app:latest",
		Passed: false,
		Details: output.SizeDetails{
			LayerCount: 2,
			MaxLayers:  20,
			Layers: []output.LayerInfo{
				{Index: 0, Bytes: 300000000},
				{Index: 1, Bytes: 250000000},
			},
			TotalBytes: 550000000,
			TotalMB:    550.0,
			MaxSizeMB:  500,
		},
		Message: "Image exceeds maximum size",
	}

	captured := captureStdout(t, func() {
		renderSizeText(result)
	})

	assert.Contains(t, captured, "large-app:latest")
	assert.Contains(t, captured, "550000000 bytes (550.00 MB)")
	assert.Contains(t, captured, "Image exceeds maximum size")
}

func TestRenderSizeText_ExceedsLayersLimit(t *testing.T) {
	result := &output.CheckResult{
		Check:  "size",
		Image:  "many-layers:latest",
		Passed: false,
		Details: output.SizeDetails{
			LayerCount: 25,
			MaxLayers:  20,
			Layers:     []output.LayerInfo{}, // Simplified for test
			TotalBytes: 10000000,
			TotalMB:    10.0,
			MaxSizeMB:  500,
		},
		Message: "Image has too many layers",
	}

	captured := captureStdout(t, func() {
		renderSizeText(result)
	})

	assert.Contains(t, captured, "many-layers:latest")
	assert.Contains(t, captured, "Number of layers: 25")
	assert.Contains(t, captured, "Image has more than 20 layers")
	assert.Contains(t, captured, "Image has too many layers")
}

func TestRenderPortsText_AllAllowed(t *testing.T) {
	result := &output.CheckResult{
		Check:  "ports",
		Image:  "nginx:latest",
		Passed: true,
		Details: output.PortsDetails{
			ExposedPorts:      []int{80, 443},
			AllowedPorts:      []int{80, 443, 8080},
			UnauthorizedPorts: []int{},
		},
		Message: "All exposed ports are allowed",
	}

	captured := captureStdout(t, func() {
		renderPortsText(result)
	})

	assert.Contains(t, captured, "Checking ports of image nginx:latest")
	assert.Contains(t, captured, "Exposed ports:")
	assert.Contains(t, captured, "- 80")
	assert.Contains(t, captured, "- 443")
	assert.Contains(t, captured, "All exposed ports are allowed")
	assert.NotContains(t, captured, "not in the allowed list")
}

func TestRenderPortsText_SomeForbidden(t *testing.T) {
	result := &output.CheckResult{
		Check:  "ports",
		Image:  "app:latest",
		Passed: false,
		Details: output.PortsDetails{
			ExposedPorts:      []int{80, 22, 3306},
			AllowedPorts:      []int{80, 443},
			UnauthorizedPorts: []int{22, 3306},
		},
		Message: "Some ports are not allowed",
	}

	captured := captureStdout(t, func() {
		renderPortsText(result)
	})

	assert.Contains(t, captured, "Exposed ports:")
	assert.Contains(t, captured, "- 22")
	assert.Contains(t, captured, "- 3306")
	assert.Contains(t, captured, "The following ports are not in the allowed list:")
	assert.Contains(t, captured, "Some ports are not allowed")
}

func TestRenderPortsText_NoPorts(t *testing.T) {
	result := &output.CheckResult{
		Check:  "ports",
		Image:  "distroless:latest",
		Passed: true,
		Details: output.PortsDetails{
			ExposedPorts:      []int{},
			AllowedPorts:      []int{80, 443},
			UnauthorizedPorts: []int{},
		},
		Message: "",
	}

	captured := captureStdout(t, func() {
		renderPortsText(result)
	})

	assert.Contains(t, captured, "Checking ports of image distroless:latest")
	assert.Contains(t, captured, "No ports are exposed in this image")
	assert.NotContains(t, captured, "Exposed ports:")
}

func TestRenderPortsText_NoAllowedPorts(t *testing.T) {
	result := &output.CheckResult{
		Check:  "ports",
		Image:  "nginx:latest",
		Passed: false,
		Details: output.PortsDetails{
			ExposedPorts:      []int{80},
			AllowedPorts:      []int{},
			UnauthorizedPorts: []int{},
		},
		Message: "",
	}

	captured := captureStdout(t, func() {
		renderPortsText(result)
	})

	assert.Contains(t, captured, "Exposed ports:")
	assert.Contains(t, captured, "- 80")
	assert.Contains(t, captured, "No allowed ports were provided")
}

func TestRenderRegistryText_Trusted(t *testing.T) {
	result := &output.CheckResult{
		Check:  "registry",
		Image:  "docker.io/nginx:latest",
		Passed: true,
		Details: output.RegistryDetails{
			Registry: "docker.io",
			Skipped:  false,
		},
		Message: "Registry is trusted",
	}

	captured := captureStdout(t, func() {
		renderRegistryText(result)
	})

	assert.Contains(t, captured, "Checking registry of image docker.io/nginx:latest")
	assert.Contains(t, captured, "Image registry: docker.io")
	assert.Contains(t, captured, "Registry is trusted")
}

func TestRenderRegistryText_Untrusted(t *testing.T) {
	result := &output.CheckResult{
		Check:  "registry",
		Image:  "untrusted.io/app:latest",
		Passed: false,
		Details: output.RegistryDetails{
			Registry: "untrusted.io",
			Skipped:  false,
		},
		Message: "Registry is not trusted",
	}

	captured := captureStdout(t, func() {
		renderRegistryText(result)
	})

	assert.Contains(t, captured, "untrusted.io/app:latest")
	assert.Contains(t, captured, "Image registry: untrusted.io")
	assert.Contains(t, captured, "Registry is not trusted")
}

func TestRenderRegistryText_Skipped(t *testing.T) {
	result := &output.CheckResult{
		Check:  "registry",
		Image:  "oci:/local/path:tag",
		Passed: true,
		Details: output.RegistryDetails{
			Registry: "",
			Skipped:  true,
		},
		Message: "",
	}

	captured := captureStdout(t, func() {
		renderRegistryText(result)
	})

	assert.Contains(t, captured, "Checking registry of image oci:/local/path:tag")
	assert.Contains(t, captured, "Registry validation skipped (not applicable for this transport)")
	assert.NotContains(t, captured, "Image registry:")
}

func TestRenderRootUserText_NonRoot(t *testing.T) {
	result := &output.CheckResult{
		Check:   "root-user",
		Image:   "nginx:latest",
		Passed:  true,
		Message: "Image runs as non-root user",
	}

	captured := captureStdout(t, func() {
		renderRootUserText(result)
	})

	assert.Contains(t, captured, "Checking if image nginx:latest is configured to run as a non-root user")
	assert.Contains(t, captured, "Image runs as non-root user")
}

func TestRenderRootUserText_Root(t *testing.T) {
	result := &output.CheckResult{
		Check:   "root-user",
		Image:   "old-app:v1",
		Passed:  false,
		Message: "Image runs as root user",
	}

	captured := captureStdout(t, func() {
		renderRootUserText(result)
	})

	assert.Contains(t, captured, "Checking if image old-app:v1 is configured to run as a non-root user")
	assert.Contains(t, captured, "Image runs as root user")
}

func TestRenderSecretsText_NoSecrets(t *testing.T) {
	result := &output.CheckResult{
		Check:  "secrets",
		Image:  "clean-app:latest",
		Passed: true,
		Details: output.SecretsDetails{
			EnvVarFindings: []output.EnvVarFinding{},
			FileFindings:   []output.FileFinding{},
			TotalFindings:  0,
			EnvVarCount:    0,
			FileCount:      0,
		},
		Message: "No secrets found",
	}

	captured := captureStdout(t, func() {
		renderSecretsText(result)
	})

	assert.Contains(t, captured, "Checking secrets in image clean-app:latest")
	assert.Contains(t, captured, "Total findings: 0")
	assert.Contains(t, captured, "No secrets found")
	assert.NotContains(t, captured, "Environment Variables:")
	assert.NotContains(t, captured, "Files with Sensitive Patterns:")
}

func TestRenderSecretsText_EnvVarsOnly(t *testing.T) {
	result := &output.CheckResult{
		Check:  "secrets",
		Image:  "app:latest",
		Passed: false,
		Details: output.SecretsDetails{
			EnvVarFindings: []output.EnvVarFinding{
				{Name: "API_KEY", Description: "Contains 'key'"},
				{Name: "DB_PASSWORD", Description: "Contains 'password'"},
			},
			FileFindings:  []output.FileFinding{},
			TotalFindings: 2,
			EnvVarCount:   2,
			FileCount:     0,
		},
		Message: "Secrets found in environment variables",
	}

	captured := captureStdout(t, func() {
		renderSecretsText(result)
	})

	assert.Contains(t, captured, "Environment Variables:")
	assert.Contains(t, captured, "API_KEY (Contains 'key')")
	assert.Contains(t, captured, "DB_PASSWORD (Contains 'password')")
	assert.Contains(t, captured, "Total findings: 2 (2 environment variables, 0 files)")
	assert.NotContains(t, captured, "Files with Sensitive Patterns:")
}

func TestRenderSecretsText_FilesOnly(t *testing.T) {
	result := &output.CheckResult{
		Check:  "secrets",
		Image:  "app:latest",
		Passed: false,
		Details: output.SecretsDetails{
			EnvVarFindings: []output.EnvVarFinding{},
			FileFindings: []output.FileFinding{
				{LayerIndex: 0, Path: "/root/.ssh/id_rsa", Description: "SSH private key"},
				{LayerIndex: 1, Path: "/etc/shadow", Description: "Shadow file"},
			},
			TotalFindings: 2,
			EnvVarCount:   0,
			FileCount:     2,
		},
		Message: "Secrets found in files",
	}

	captured := captureStdout(t, func() {
		renderSecretsText(result)
	})

	assert.Contains(t, captured, "Files with Sensitive Patterns:")
	assert.Contains(t, captured, "Layer 1:")
	assert.Contains(t, captured, "/root/.ssh/id_rsa (SSH private key)")
	assert.Contains(t, captured, "Layer 2:")
	assert.Contains(t, captured, "/etc/shadow (Shadow file)")
	assert.Contains(t, captured, "Total findings: 2 (0 environment variables, 2 files)")
	assert.NotContains(t, captured, "Environment Variables:")
}

func TestRenderSecretsText_Mixed(t *testing.T) {
	result := &output.CheckResult{
		Check:  "secrets",
		Image:  "app:latest",
		Passed: false,
		Details: output.SecretsDetails{
			EnvVarFindings: []output.EnvVarFinding{
				{Name: "SECRET_TOKEN", Description: "Contains 'secret'"},
			},
			FileFindings: []output.FileFinding{
				{LayerIndex: 0, Path: "/app/.env", Description: "Environment file"},
				{LayerIndex: 0, Path: "/app/.aws/credentials", Description: "AWS credentials"},
			},
			TotalFindings: 3,
			EnvVarCount:   1,
			FileCount:     2,
		},
		Message: "Secrets found",
	}

	captured := captureStdout(t, func() {
		renderSecretsText(result)
	})

	assert.Contains(t, captured, "Environment Variables:")
	assert.Contains(t, captured, "SECRET_TOKEN")
	assert.Contains(t, captured, "Files with Sensitive Patterns:")
	assert.Contains(t, captured, "Layer 1:")
	assert.Contains(t, captured, "/app/.env")
	assert.Contains(t, captured, "/app/.aws/credentials")
	assert.Contains(t, captured, "Total findings: 3 (1 environment variables, 2 files)")
}

func TestRenderSecretsText_MultipleLayers(t *testing.T) {
	result := &output.CheckResult{
		Check:  "secrets",
		Image:  "app:latest",
		Passed: false,
		Details: output.SecretsDetails{
			EnvVarFindings: []output.EnvVarFinding{},
			FileFindings: []output.FileFinding{
				{LayerIndex: 0, Path: "/layer0/secret1", Description: "Secret 1"},
				{LayerIndex: 0, Path: "/layer0/secret2", Description: "Secret 2"},
				{LayerIndex: 2, Path: "/layer2/secret3", Description: "Secret 3"},
			},
			TotalFindings: 3,
			EnvVarCount:   0,
			FileCount:     3,
		},
		Message: "Secrets found",
	}

	captured := captureStdout(t, func() {
		renderSecretsText(result)
	})

	assert.Contains(t, captured, "Layer 1:")
	assert.Contains(t, captured, "/layer0/secret1")
	assert.Contains(t, captured, "/layer0/secret2")
	assert.Contains(t, captured, "Layer 3:")
	assert.Contains(t, captured, "/layer2/secret3")
}

func TestRenderSecretsText_SparseLayerIndices(t *testing.T) {
	// Findings in layers 0 and 20 with only 2 map entries.
	// The old len(layerMap)+10 loop (range 0..11) would silently drop layer 20.
	result := &output.CheckResult{
		Check:  "secrets",
		Image:  "app:latest",
		Passed: false,
		Details: output.SecretsDetails{
			EnvVarFindings: []output.EnvVarFinding{},
			FileFindings: []output.FileFinding{
				{LayerIndex: 0, Path: "/layer0/secret", Description: "First layer secret"},
				{LayerIndex: 20, Path: "/layer20/secret", Description: "High layer secret"},
			},
			TotalFindings: 2,
			EnvVarCount:   0,
			FileCount:     2,
		},
		Message: "Secrets found",
	}

	captured := captureStdout(t, func() {
		renderSecretsText(result)
	})

	assert.Contains(t, captured, "Layer 1:")
	assert.Contains(t, captured, "/layer0/secret (First layer secret)")
	assert.Contains(t, captured, "Layer 21:")
	assert.Contains(t, captured, "/layer20/secret (High layer secret)")
}

func TestRenderLabelsText_AllValid(t *testing.T) {
	result := &output.CheckResult{
		Check:  "labels",
		Image:  "nginx:latest",
		Passed: true,
		Details: output.LabelsDetails{
			RequiredLabels: []output.RequiredLabelCheck{
				{Name: "maintainer"},
				{Name: "version", Pattern: "^v?\\d+\\.\\d+\\.\\d+$"},
			},
			ActualLabels: map[string]string{
				"maintainer": "John Doe",
				"version":    "v1.2.3",
			},
			MissingLabels: []string{},
			InvalidLabels: []output.InvalidLabelDetail{},
		},
		Message: "All required labels are present and valid",
	}

	captured := captureStdout(t, func() {
		renderLabelsText(result)
	})

	assert.Contains(t, captured, "Checking labels of image nginx:latest")
	assert.Contains(t, captured, "Required labels:")
	assert.Contains(t, captured, "maintainer (existence check)")
	assert.Contains(t, captured, "version (pattern:")
	assert.Contains(t, captured, "Actual labels (2):")
	assert.Contains(t, captured, "maintainer: John Doe")
	assert.Contains(t, captured, "version: v1.2.3")
	assert.Contains(t, captured, "All required labels are present and valid")
	assert.NotContains(t, captured, "Missing labels")
	assert.NotContains(t, captured, "Invalid labels")
}

func TestRenderLabelsText_MissingLabels(t *testing.T) {
	result := &output.CheckResult{
		Check:  "labels",
		Image:  "app:latest",
		Passed: false,
		Details: output.LabelsDetails{
			RequiredLabels: []output.RequiredLabelCheck{
				{Name: "maintainer"},
				{Name: "version"},
			},
			ActualLabels:  map[string]string{},
			MissingLabels: []string{"maintainer", "version"},
			InvalidLabels: []output.InvalidLabelDetail{},
		},
		Message: "Missing required labels",
	}

	captured := captureStdout(t, func() {
		renderLabelsText(result)
	})

	assert.Contains(t, captured, "No labels found in image")
	assert.Contains(t, captured, "Missing labels (2):")
	assert.Contains(t, captured, "- maintainer")
	assert.Contains(t, captured, "- version")
	assert.Contains(t, captured, "Missing required labels")
}

func TestRenderLabelsText_InvalidValues(t *testing.T) {
	result := &output.CheckResult{
		Check:  "labels",
		Image:  "app:latest",
		Passed: false,
		Details: output.LabelsDetails{
			RequiredLabels: []output.RequiredLabelCheck{
				{Name: "version", Pattern: "^v?\\d+\\.\\d+\\.\\d+$"},
				{Name: "vendor", Value: "MyCompany"},
			},
			ActualLabels: map[string]string{
				"vendor":  "OtherCompany",
				"version": "1.2",
			},
			MissingLabels: []string{},
			InvalidLabels: []output.InvalidLabelDetail{
				{Name: "version", ActualValue: "1.2", ExpectedPattern: "^v?\\d+\\.\\d+\\.\\d+$", Reason: "value \"1.2\" does not match pattern"},
				{Name: "vendor", ActualValue: "OtherCompany", ExpectedValue: "MyCompany", Reason: "expected \"MyCompany\""},
			},
		},
		Message: "Invalid label values",
	}

	captured := captureStdout(t, func() {
		renderLabelsText(result)
	})

	assert.Contains(t, captured, "Invalid labels (2):")
	assert.Contains(t, captured, "- version: value \"1.2\" does not match pattern")
	assert.Contains(t, captured, "- vendor: expected \"MyCompany\"")
	assert.Contains(t, captured, "Invalid label values")
}

func TestRenderLabelsText_MixedValidationModes(t *testing.T) {
	result := &output.CheckResult{
		Check:  "labels",
		Image:  "app:latest",
		Passed: true,
		Details: output.LabelsDetails{
			RequiredLabels: []output.RequiredLabelCheck{
				{Name: "maintainer"},
				{Name: "version", Pattern: "^v?\\d+\\.\\d+\\.\\d+$"},
				{Name: "vendor", Value: "MyCompany"},
			},
			ActualLabels: map[string]string{
				"maintainer": "John Doe",
				"vendor":     "MyCompany",
				"version":    "v1.2.3",
			},
			MissingLabels: []string{},
			InvalidLabels: []output.InvalidLabelDetail{},
		},
		Message: "All labels valid",
	}

	captured := captureStdout(t, func() {
		renderLabelsText(result)
	})

	assert.Contains(t, captured, "maintainer (existence check)")
	assert.Contains(t, captured, "version (pattern:")
	assert.Contains(t, captured, "vendor (exact:")
}

func TestRenderLabelsText_EmptyLabels(t *testing.T) {
	result := &output.CheckResult{
		Check:  "labels",
		Image:  "minimal:latest",
		Passed: false,
		Details: output.LabelsDetails{
			RequiredLabels: []output.RequiredLabelCheck{
				{Name: "maintainer"},
			},
			ActualLabels:  map[string]string{},
			MissingLabels: []string{"maintainer"},
			InvalidLabels: []output.InvalidLabelDetail{},
		},
		Message: "No labels found",
	}

	captured := captureStdout(t, func() {
		renderLabelsText(result)
	})

	assert.Contains(t, captured, "No labels found in image")
	assert.Contains(t, captured, "Missing labels (1):")
	assert.NotContains(t, captured, "Actual labels")
}

func TestRenderEntrypointText_WithBothEntrypointAndCmd(t *testing.T) {
	result := &output.CheckResult{
		Check:  "entrypoint",
		Image:  "nginx:latest",
		Passed: true,
		Details: output.EntrypointDetails{
			HasEntrypoint: true,
			ExecForm:      true,
			Entrypoint:    []string{"/docker-entrypoint.sh"},
			Cmd:           []string{"nginx", "-g", "daemon off;"},
		},
		Message: "Image has a valid exec-form entrypoint",
	}

	captured := captureStdout(t, func() {
		renderEntrypointText(result)
	})

	assert.Contains(t, captured, "Checking entrypoint of image nginx:latest")
	assert.Contains(t, captured, "Entrypoint:")
	assert.Contains(t, captured, "/docker-entrypoint.sh")
	assert.Contains(t, captured, "Cmd:")
	assert.Contains(t, captured, "nginx")
	assert.Contains(t, captured, "Image has a valid exec-form entrypoint")
}
