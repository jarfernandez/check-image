package commands

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetValidationResult(t *testing.T) {
	tests := []struct {
		name           string
		initialResult  ValidationResult
		passed         bool
		successMsg     string
		failureMsg     string
		expectedResult ValidationResult
		expectedOutput string
	}{
		{
			name:           "Pass when initial state is skipped",
			initialResult:  ValidationSkipped,
			passed:         true,
			successMsg:     "Validation passed",
			failureMsg:     "Validation failed",
			expectedResult: ValidationSucceeded,
			expectedOutput: "Validation passed\n",
		},
		{
			name:           "Fail when initial state is skipped",
			initialResult:  ValidationSkipped,
			passed:         false,
			successMsg:     "Validation passed",
			failureMsg:     "Validation failed",
			expectedResult: ValidationFailed,
			expectedOutput: "Validation failed\n",
		},
		{
			name:           "Pass preserves previous failure",
			initialResult:  ValidationFailed,
			passed:         true,
			successMsg:     "Validation passed",
			failureMsg:     "Validation failed",
			expectedResult: ValidationFailed,
			expectedOutput: "Validation passed\n",
		},
		{
			name:           "Fail overrides succeeded",
			initialResult:  ValidationSucceeded,
			passed:         false,
			successMsg:     "Validation passed",
			failureMsg:     "Validation failed",
			expectedResult: ValidationFailed,
			expectedOutput: "Validation failed\n",
		},
		{
			name:           "Pass when already succeeded",
			initialResult:  ValidationSucceeded,
			passed:         true,
			successMsg:     "Validation passed",
			failureMsg:     "Validation failed",
			expectedResult: ValidationSucceeded,
			expectedOutput: "Validation passed\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Set initial result
			Result = tt.initialResult

			// Call function
			SetValidationResult(tt.passed, tt.successMsg, tt.failureMsg)

			// Restore stdout
			_ = w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)

			// Assert
			assert.Equal(t, tt.expectedResult, Result)
			assert.Equal(t, tt.expectedOutput, buf.String())
		})
	}
}

func TestValidationResultConstants(t *testing.T) {
	// Ensure the constants have expected values
	assert.Equal(t, ValidationResult(0), ValidationFailed)
	assert.Equal(t, ValidationResult(1), ValidationSucceeded)
	assert.Equal(t, ValidationResult(2), ValidationSkipped)
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
