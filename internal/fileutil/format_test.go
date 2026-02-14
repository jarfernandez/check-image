package fileutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsYAML(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		isYAML bool
	}{
		{
			name:   "JSON object",
			data:   []byte(`{"key": "value"}`),
			isYAML: false,
		},
		{
			name:   "JSON array",
			data:   []byte(`["item1", "item2"]`),
			isYAML: false,
		},
		{
			name:   "JSON with whitespace",
			data:   []byte(`  { "key": "value" }  `),
			isYAML: false,
		},
		{
			name:   "JSON with newlines",
			data:   []byte("\n\n  { \"key\": \"value\" }"),
			isYAML: false,
		},
		{
			name:   "YAML simple",
			data:   []byte("key: value"),
			isYAML: true,
		},
		{
			name:   "YAML with whitespace",
			data:   []byte("  key: value  "),
			isYAML: true,
		},
		{
			name:   "YAML with newlines",
			data:   []byte("\n\nkey: value"),
			isYAML: true,
		},
		{
			name:   "YAML list",
			data:   []byte("- item1\n- item2"),
			isYAML: true,
		},
		{
			name:   "Empty data",
			data:   []byte(""),
			isYAML: false, // defaults to JSON
		},
		{
			name:   "Whitespace only",
			data:   []byte("   \n\t  "),
			isYAML: false, // defaults to JSON
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsYAML(tt.data)
			assert.Equal(t, tt.isYAML, result)
		})
	}
}

func TestHasYAMLExtension(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"config.yaml", true},
		{"config.yml", true},
		{"config.json", false},
		{"config", false},
		{"config.txt", false},
		{"/path/to/config.yaml", true},
		{"/path/to/config.yml", true},
		{"/path/to/config.json", false},
		{"config.YAML", false}, // case sensitive
		{"config.YML", false},  // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := HasYAMLExtension(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
