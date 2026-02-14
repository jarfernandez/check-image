package output

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Format
		wantErr bool
		errMsg  string
	}{
		{
			name:  "text format",
			input: "text",
			want:  FormatText,
		},
		{
			name:  "json format",
			input: "json",
			want:  FormatJSON,
		},
		{
			name:    "unsupported format",
			input:   "xml",
			wantErr: true,
			errMsg:  "unsupported output format",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errMsg:  "unsupported output format",
		},
		{
			name:    "uppercase JSON",
			input:   "JSON",
			wantErr: true,
			errMsg:  "unsupported output format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestRenderJSON(t *testing.T) {
	t.Run("renders check result", func(t *testing.T) {
		result := CheckResult{
			Check:   "age",
			Image:   "nginx:latest",
			Passed:  true,
			Message: "Image is less than 90 days old",
			Details: AgeDetails{
				CreatedAt: "2025-12-01T00:00:00Z",
				AgeDays:   75,
				MaxAge:    90,
			},
		}

		var buf bytes.Buffer
		err := RenderJSON(&buf, result)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `"check": "age"`)
		assert.Contains(t, output, `"passed": true`)
		assert.Contains(t, output, `"age-days": 75`)
	})

	t.Run("omits empty details", func(t *testing.T) {
		result := CheckResult{
			Check:   "age",
			Image:   "nginx:latest",
			Passed:  false,
			Message: "failed",
		}

		var buf bytes.Buffer
		err := RenderJSON(&buf, result)
		require.NoError(t, err)

		output := buf.String()
		assert.NotContains(t, output, `"details"`)
		assert.NotContains(t, output, `"error"`)
	})

	t.Run("renders all result", func(t *testing.T) {
		result := AllResult{
			Image:  "nginx:latest",
			Passed: false,
			Checks: []CheckResult{
				{Check: "age", Image: "nginx:latest", Passed: true, Message: "ok"},
				{Check: "root-user", Image: "nginx:latest", Passed: false, Message: "fail"},
			},
			Summary: Summary{
				Total:   6,
				Passed:  4,
				Failed:  1,
				Errored: 0,
				Skipped: []string{"registry"},
			},
		}

		var buf bytes.Buffer
		err := RenderJSON(&buf, result)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `"image": "nginx:latest"`)
		assert.Contains(t, output, `"passed": false`)
		assert.Contains(t, output, `"total": 6`)
		assert.Contains(t, output, `"registry"`)
	})

	t.Run("renders version result", func(t *testing.T) {
		result := VersionResult{Version: "v0.4.0"}

		var buf bytes.Buffer
		err := RenderJSON(&buf, result)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `"version": "v0.4.0"`)
	})
}
