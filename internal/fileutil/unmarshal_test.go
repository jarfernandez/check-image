package fileutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Name    string   `json:"name" yaml:"name"`
	Values  []string `json:"values" yaml:"values"`
	Enabled bool     `json:"enabled" yaml:"enabled"`
	Count   int      `json:"count" yaml:"count"`
}

func TestUnmarshalConfigFile_JSON(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		fileName    string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, cfg *testConfig)
	}{
		{
			name: "Valid JSON",
			content: `{
				"name": "test-config",
				"values": ["a", "b", "c"],
				"enabled": true,
				"count": 42
			}`,
			fileName: "config.json",
			wantErr:  false,
			validate: func(t *testing.T, cfg *testConfig) {
				assert.Equal(t, "test-config", cfg.Name)
				assert.Equal(t, []string{"a", "b", "c"}, cfg.Values)
				assert.True(t, cfg.Enabled)
				assert.Equal(t, 42, cfg.Count)
			},
		},
		{
			name:        "Invalid JSON - missing quote",
			content:     `{"name": "test, "enabled": true}`,
			fileName:    "config.json",
			wantErr:     true,
			errContains: "invalid JSON",
		},
		{
			name:        "Invalid JSON - trailing comma",
			content:     `{"name": "test",}`,
			fileName:    "config.json",
			wantErr:     true,
			errContains: "invalid JSON",
		},
		{
			name:     "Empty JSON object",
			content:  `{}`,
			fileName: "config.json",
			wantErr:  false,
			validate: func(t *testing.T, cfg *testConfig) {
				assert.Empty(t, cfg.Name)
				assert.Nil(t, cfg.Values)
				assert.False(t, cfg.Enabled)
				assert.Equal(t, 0, cfg.Count)
			},
		},
		{
			name:        "Not JSON at all",
			content:     `this is not json`,
			fileName:    "config.json",
			wantErr:     true,
			errContains: "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg testConfig
			err := UnmarshalConfigFile([]byte(tt.content), &cfg, tt.fileName)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, &cfg)
			}
		})
	}
}

func TestUnmarshalConfigFile_YAML(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		fileName    string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, cfg *testConfig)
	}{
		{
			name: "Valid YAML",
			content: `name: test-config
values:
  - a
  - b
  - c
enabled: true
count: 42`,
			fileName: "config.yaml",
			wantErr:  false,
			validate: func(t *testing.T, cfg *testConfig) {
				assert.Equal(t, "test-config", cfg.Name)
				assert.Equal(t, []string{"a", "b", "c"}, cfg.Values)
				assert.True(t, cfg.Enabled)
				assert.Equal(t, 42, cfg.Count)
			},
		},
		{
			name: "YAML with .yml extension",
			content: `name: yml-config
enabled: false`,
			fileName: "config.yml",
			wantErr:  false,
			validate: func(t *testing.T, cfg *testConfig) {
				assert.Equal(t, "yml-config", cfg.Name)
				assert.False(t, cfg.Enabled)
			},
		},
		{
			name:     "Empty YAML",
			content:  ``,
			fileName: "config.yaml",
			wantErr:  false,
			validate: func(t *testing.T, cfg *testConfig) {
				assert.Empty(t, cfg.Name)
			},
		},
		{
			name: "YAML with comments",
			content: `# This is a comment
name: commented-config
# Another comment
enabled: true`,
			fileName: "config.yaml",
			wantErr:  false,
			validate: func(t *testing.T, cfg *testConfig) {
				assert.Equal(t, "commented-config", cfg.Name)
				assert.True(t, cfg.Enabled)
			},
		},
		{
			name: "Invalid YAML - duplicate key",
			content: `name: test
name: duplicate`,
			fileName:    "config.yaml",
			wantErr:     true, // YAML v3 rejects duplicate keys
			errContains: "invalid YAML",
		},
		{
			name: "Invalid YAML - unclosed bracket",
			content: `values: [a, b, c
enabled: true`,
			fileName:    "config.yaml",
			wantErr:     true,
			errContains: "invalid YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg testConfig
			err := UnmarshalConfigFile([]byte(tt.content), &cfg, tt.fileName)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, &cfg)
			}
		})
	}
}

func TestUnmarshalConfigFile_FormatDetection(t *testing.T) {
	yamlContent := `name: yaml-config
enabled: true`

	jsonContent := `{"name": "json-config", "enabled": false}`

	tests := []struct {
		name     string
		content  string
		fileName string
		validate func(t *testing.T, cfg *testConfig)
	}{
		{
			name:     "YAML extension .yaml",
			content:  yamlContent,
			fileName: "test.yaml",
			validate: func(t *testing.T, cfg *testConfig) {
				assert.Equal(t, "yaml-config", cfg.Name)
				assert.True(t, cfg.Enabled)
			},
		},
		{
			name:     "YAML extension .yml",
			content:  yamlContent,
			fileName: "test.yml",
			validate: func(t *testing.T, cfg *testConfig) {
				assert.Equal(t, "yaml-config", cfg.Name)
				assert.True(t, cfg.Enabled)
			},
		},
		{
			name:     "JSON extension .json",
			content:  jsonContent,
			fileName: "test.json",
			validate: func(t *testing.T, cfg *testConfig) {
				assert.Equal(t, "json-config", cfg.Name)
				assert.False(t, cfg.Enabled)
			},
		},
		{
			name:     "No extension defaults to JSON",
			content:  jsonContent,
			fileName: "config",
			validate: func(t *testing.T, cfg *testConfig) {
				assert.Equal(t, "json-config", cfg.Name)
			},
		},
		{
			name:     "Unknown extension defaults to JSON",
			content:  jsonContent,
			fileName: "config.txt",
			validate: func(t *testing.T, cfg *testConfig) {
				assert.Equal(t, "json-config", cfg.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg testConfig
			err := UnmarshalConfigFile([]byte(tt.content), &cfg, tt.fileName)
			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, &cfg)
			}
		})
	}
}

func TestUnmarshalConfigFile_TypeMismatch(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		fileName string
		wantErr  bool
	}{
		{
			name:     "JSON string where number expected",
			content:  `{"name": "test", "count": "not-a-number"}`,
			fileName: "config.json",
			wantErr:  true,
		},
		{
			name: "YAML string where bool expected",
			content: `name: test
enabled: not-a-bool`,
			fileName: "config.yaml",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg testConfig
			err := UnmarshalConfigFile([]byte(tt.content), &cfg, tt.fileName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUnmarshalConfigFile_Integration(t *testing.T) {
	// Test actual file reading flow
	tmpDir := t.TempDir()

	jsonFile := filepath.Join(tmpDir, "config.json")
	jsonContent := []byte(`{"name": "integration", "enabled": true}`)
	err := os.WriteFile(jsonFile, jsonContent, 0600)
	require.NoError(t, err)

	yamlFile := filepath.Join(tmpDir, "config.yaml")
	yamlContent := []byte("name: integration\nenabled: true")
	err = os.WriteFile(yamlFile, yamlContent, 0600)
	require.NoError(t, err)

	t.Run("JSON file", func(t *testing.T) {
		data, err := os.ReadFile(jsonFile)
		require.NoError(t, err)

		var cfg testConfig
		err = UnmarshalConfigFile(data, &cfg, jsonFile)
		require.NoError(t, err)
		assert.Equal(t, "integration", cfg.Name)
		assert.True(t, cfg.Enabled)
	})

	t.Run("YAML file", func(t *testing.T) {
		data, err := os.ReadFile(yamlFile)
		require.NoError(t, err)

		var cfg testConfig
		err = UnmarshalConfigFile(data, &cfg, yamlFile)
		require.NoError(t, err)
		assert.Equal(t, "integration", cfg.Name)
		assert.True(t, cfg.Enabled)
	})
}
