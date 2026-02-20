package labels

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadLabelsPolicy_JSON(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, p *Policy)
	}{
		{
			name: "Valid policy with existence check",
			content: `{
				"required-labels": [
					{"name": "maintainer"}
				]
			}`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				require.Len(t, p.RequiredLabels, 1)
				assert.Equal(t, "maintainer", p.RequiredLabels[0].Name)
				assert.Empty(t, p.RequiredLabels[0].Value)
				assert.Empty(t, p.RequiredLabels[0].Pattern)
			},
		},
		{
			name: "Valid policy with exact value",
			content: `{
				"required-labels": [
					{"name": "org.opencontainers.image.vendor", "value": "MyCompany"}
				]
			}`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				require.Len(t, p.RequiredLabels, 1)
				assert.Equal(t, "org.opencontainers.image.vendor", p.RequiredLabels[0].Name)
				assert.Equal(t, "MyCompany", p.RequiredLabels[0].Value)
				assert.Empty(t, p.RequiredLabels[0].Pattern)
			},
		},
		{
			name: "Valid policy with pattern",
			content: `{
				"required-labels": [
					{"name": "version", "pattern": "^v?\\d+\\.\\d+\\.\\d+$"}
				]
			}`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				require.Len(t, p.RequiredLabels, 1)
				assert.Equal(t, "version", p.RequiredLabels[0].Name)
				assert.Empty(t, p.RequiredLabels[0].Value)
				assert.Equal(t, "^v?\\d+\\.\\d+\\.\\d+$", p.RequiredLabels[0].Pattern)
			},
		},
		{
			name: "Valid policy with multiple labels",
			content: `{
				"required-labels": [
					{"name": "maintainer"},
					{"name": "version", "pattern": "^v?\\d+\\.\\d+\\.\\d+$"},
					{"name": "vendor", "value": "Acme Inc"}
				]
			}`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.Len(t, p.RequiredLabels, 3)
			},
		},
		{
			name: "Empty required labels list",
			content: `{
				"required-labels": []
			}`,
			wantErr:     true,
			errContains: "at least one required label",
		},
		{
			name:        "No required labels field",
			content:     `{}`,
			wantErr:     true,
			errContains: "at least one required label",
		},
		{
			name: "Label with both value and pattern",
			content: `{
				"required-labels": [
					{"name": "version", "value": "1.0", "pattern": "^v?\\d+"}
				]
			}`,
			wantErr:     true,
			errContains: "both value and pattern",
		},
		{
			name: "Label without name",
			content: `{
				"required-labels": [
					{"value": "test"}
				]
			}`,
			wantErr:     true,
			errContains: "missing a name",
		},
		{
			name: "Duplicate label names",
			content: `{
				"required-labels": [
					{"name": "version"},
					{"name": "version", "pattern": "^v?\\d+"}
				]
			}`,
			wantErr:     true,
			errContains: "duplicate label name",
		},
		{
			name: "Invalid regex pattern",
			content: `{
				"required-labels": [
					{"name": "version", "pattern": "[invalid(regex"}
				]
			}`,
			wantErr:     true,
			errContains: "invalid pattern",
		},
		{
			name:        "Invalid JSON",
			content:     `{invalid json}`,
			wantErr:     true,
			errContains: "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			policyFile := filepath.Join(tmpDir, "policy.json")
			err := os.WriteFile(policyFile, []byte(tt.content), 0600)
			require.NoError(t, err)

			policy, err := LoadLabelsPolicy(policyFile)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, policy)
			if tt.validate != nil {
				tt.validate(t, policy)
			}
		})
	}
}

func TestLoadLabelsPolicy_YAML(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, p *Policy)
	}{
		{
			name: "Valid YAML with multiple labels",
			content: `required-labels:
  - name: maintainer
  - name: org.opencontainers.image.version
    pattern: "^v?\\d+\\.\\d+\\.\\d+$"
  - name: org.opencontainers.image.vendor
    value: "MyCompany"`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				require.Len(t, p.RequiredLabels, 3)
				assert.Equal(t, "maintainer", p.RequiredLabels[0].Name)
				assert.Equal(t, "org.opencontainers.image.version", p.RequiredLabels[1].Name)
				assert.Equal(t, "^v?\\d+\\.\\d+\\.\\d+$", p.RequiredLabels[1].Pattern)
				assert.Equal(t, "org.opencontainers.image.vendor", p.RequiredLabels[2].Name)
				assert.Equal(t, "MyCompany", p.RequiredLabels[2].Value)
			},
		},
		{
			name: "Valid YAML existence check only",
			content: `required-labels:
  - name: team
  - name: cost-center`,
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.Len(t, p.RequiredLabels, 2)
			},
		},
		{
			name: "Invalid YAML",
			content: `required-labels:
  - name: test
    invalid:
    - [unclosed`,
			wantErr:     true,
			errContains: "invalid YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			policyFile := filepath.Join(tmpDir, "policy.yaml")
			err := os.WriteFile(policyFile, []byte(tt.content), 0600)
			require.NoError(t, err)

			policy, err := LoadLabelsPolicy(policyFile)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, policy)
			if tt.validate != nil {
				tt.validate(t, policy)
			}
		})
	}
}

func TestLoadLabelsPolicy_Stdin(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
	}{
		{
			name: "Valid JSON from stdin",
			content: `{
				"required-labels": [
					{"name": "maintainer"}
				]
			}`,
			wantErr: false,
		},
		{
			name: "Valid YAML from stdin",
			content: `required-labels:
  - name: maintainer`,
			wantErr: false,
		},
		{
			name:        "Invalid content from stdin",
			content:     `invalid content`,
			wantErr:     true,
			errContains: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temp file and redirect stdin
			tmpDir := t.TempDir()
			stdinFile := filepath.Join(tmpDir, "stdin")
			err := os.WriteFile(stdinFile, []byte(tt.content), 0600)
			require.NoError(t, err)

			// Open the file and restore stdin after test
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			f, err := os.Open(stdinFile)
			require.NoError(t, err)
			defer f.Close()
			os.Stdin = f

			policy, err := LoadLabelsPolicy("-")
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, policy)
		})
	}
}

func TestLoadLabelsPolicy_FileErrors(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		errContains string
	}{
		{
			name:        "Nonexistent file",
			path:        "/nonexistent/path/policy.json",
			errContains: "error reading labels policy",
		},
		{
			name:        "Directory instead of file",
			path:        ".",
			errContains: "error reading labels policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := LoadLabelsPolicy(tt.path)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
			assert.Nil(t, policy)
		})
	}
}

func TestLoadLabelsPolicyFromObject(t *testing.T) {
	tests := []struct {
		name        string
		obj         any
		wantErr     bool
		errContains string
		validate    func(t *testing.T, p *Policy)
	}{
		{
			name: "Valid inline policy",
			obj: map[string]any{
				"required-labels": []any{
					map[string]any{"name": "maintainer"},
					map[string]any{"name": "version", "pattern": "^v?\\d+"},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				assert.Len(t, p.RequiredLabels, 2)
				assert.Equal(t, "maintainer", p.RequiredLabels[0].Name)
				assert.Equal(t, "version", p.RequiredLabels[1].Name)
			},
		},
		{
			name: "Inline policy with value",
			obj: map[string]any{
				"required-labels": []any{
					map[string]any{"name": "vendor", "value": "Acme"},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, p *Policy) {
				require.Len(t, p.RequiredLabels, 1)
				assert.Equal(t, "vendor", p.RequiredLabels[0].Name)
				assert.Equal(t, "Acme", p.RequiredLabels[0].Value)
			},
		},
		{
			name: "Empty required labels",
			obj: map[string]any{
				"required-labels": []any{},
			},
			wantErr:     true,
			errContains: "at least one required label",
		},
		{
			name: "Invalid pattern in inline policy",
			obj: map[string]any{
				"required-labels": []any{
					map[string]any{"name": "version", "pattern": "[invalid("},
				},
			},
			wantErr:     true,
			errContains: "invalid pattern",
		},
		{
			name:        "Invalid object type",
			obj:         "not an object",
			wantErr:     true,
			errContains: "invalid policy object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := LoadLabelsPolicyFromObject(tt.obj)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, policy)
			if tt.validate != nil {
				tt.validate(t, policy)
			}
		})
	}
}

func TestPolicyValidate(t *testing.T) {
	tests := []struct {
		name        string
		policy      Policy
		wantErr     bool
		errContains string
	}{
		{
			name: "Valid policy",
			policy: Policy{
				RequiredLabels: []LabelRequirement{
					{Name: "maintainer"},
				},
			},
			wantErr: false,
		},
		{
			name:        "Empty required labels",
			policy:      Policy{RequiredLabels: []LabelRequirement{}},
			wantErr:     true,
			errContains: "at least one required label",
		},
		{
			name: "Label without name",
			policy: Policy{
				RequiredLabels: []LabelRequirement{
					{Value: "test"},
				},
			},
			wantErr:     true,
			errContains: "missing a name",
		},
		{
			name: "Both value and pattern",
			policy: Policy{
				RequiredLabels: []LabelRequirement{
					{Name: "version", Value: "1.0", Pattern: "^v?\\d+"},
				},
			},
			wantErr:     true,
			errContains: "both value and pattern",
		},
		{
			name: "Invalid regex pattern",
			policy: Policy{
				RequiredLabels: []LabelRequirement{
					{Name: "version", Pattern: "[invalid("},
				},
			},
			wantErr:     true,
			errContains: "invalid pattern",
		},
		{
			name: "Duplicate label names",
			policy: Policy{
				RequiredLabels: []LabelRequirement{
					{Name: "version"},
					{Name: "maintainer"},
					{Name: "version", Pattern: "^v?\\d+"},
				},
			},
			wantErr:     true,
			errContains: "duplicate label name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}
