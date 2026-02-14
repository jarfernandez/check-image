package commands

import (
	"testing"

	"github.com/jarfernandez/check-image/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationResultConstants(t *testing.T) {
	// Ensure the constants have expected values (ordered by priority)
	assert.Equal(t, ValidationResult(0), ValidationSkipped)
	assert.Equal(t, ValidationResult(1), ValidationSucceeded)
	assert.Equal(t, ValidationResult(2), ValidationFailed)
	assert.Equal(t, ValidationResult(3), ExecutionError)
}

func TestUpdateResult(t *testing.T) {
	tests := []struct {
		name     string
		initial  ValidationResult
		update   ValidationResult
		expected ValidationResult
	}{
		{
			name:     "Skipped to Succeeded",
			initial:  ValidationSkipped,
			update:   ValidationSucceeded,
			expected: ValidationSucceeded,
		},
		{
			name:     "Succeeded to Failed",
			initial:  ValidationSucceeded,
			update:   ValidationFailed,
			expected: ValidationFailed,
		},
		{
			name:     "Failed to ExecutionError",
			initial:  ValidationFailed,
			update:   ExecutionError,
			expected: ExecutionError,
		},
		{
			name:     "ExecutionError stays over Succeeded",
			initial:  ExecutionError,
			update:   ValidationSucceeded,
			expected: ExecutionError,
		},
		{
			name:     "ExecutionError stays over Failed",
			initial:  ExecutionError,
			update:   ValidationFailed,
			expected: ExecutionError,
		},
		{
			name:     "Failed stays over Succeeded",
			initial:  ValidationFailed,
			update:   ValidationSucceeded,
			expected: ValidationFailed,
		},
		{
			name:     "Skipped stays when updating with Skipped",
			initial:  ValidationSkipped,
			update:   ValidationSkipped,
			expected: ValidationSkipped,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Result = tt.initial
			UpdateResult(tt.update)
			assert.Equal(t, tt.expected, Result)
		})
	}
}

func TestRootCommand(t *testing.T) {
	// Reset Result before test
	Result = ValidationSkipped

	// Test that root command exists and has correct properties
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "check-image", rootCmd.Use)
	assert.Contains(t, rootCmd.Short, "Validation of container images")
	assert.True(t, rootCmd.SilenceUsage)
}

func TestRootCommandLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{
			name:    "Valid log level - info",
			level:   "info",
			wantErr: false,
		},
		{
			name:    "Valid log level - debug",
			level:   "debug",
			wantErr: false,
		},
		{
			name:    "Valid log level - warn",
			level:   "warn",
			wantErr: false,
		},
		{
			name:    "Valid log level - error",
			level:   "error",
			wantErr: false,
		},
		{
			name:    "Invalid log level",
			level:   "invalid",
			wantErr: true,
		},
		{
			name:    "Empty log level",
			level:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore
			origFormat := outputFormat
			defer func() { outputFormat = origFormat }()
			outputFormat = "text"

			// Set log level
			logLevel = tt.level

			// Execute PersistentPreRunE
			err := rootCmd.PersistentPreRunE(rootCmd, []string{})

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid log level")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRootCommandOutputFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("output")
	assert.NotNil(t, flag)
	assert.Equal(t, "o", flag.Shorthand)
	assert.Equal(t, "text", flag.DefValue)
}

func TestRootCommandOutputFormat(t *testing.T) {
	tests := []struct {
		name       string
		format     string
		wantErr    bool
		wantFormat output.Format
	}{
		{
			name:       "text format",
			format:     "text",
			wantFormat: output.FormatText,
		},
		{
			name:       "json format",
			format:     "json",
			wantFormat: output.FormatJSON,
		},
		{
			name:    "invalid format",
			format:  "xml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origFormat := outputFormat
			origLogLevel := logLevel
			defer func() {
				outputFormat = origFormat
				logLevel = origLogLevel
			}()

			logLevel = "info"
			outputFormat = tt.format

			err := rootCmd.PersistentPreRunE(rootCmd, []string{})

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "unsupported output format")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantFormat, OutputFmt)
			}
		})
	}
}
