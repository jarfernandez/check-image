package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/jarfernandez/check-image/cmd/check-image/commands"
	"github.com/jarfernandez/check-image/internal/output"
	"github.com/stretchr/testify/assert"
)

// errWriter is an io.Writer that always returns an error, used to exercise
// the defensive error-handling branches in exitResult.
type errWriter struct{}

func (e errWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write error")
}

func TestExitResult_ValidationSucceeded(t *testing.T) {
	var buf bytes.Buffer
	exitCode := exitResult(commands.ExecuteResult{
		Validation: commands.ValidationSucceeded,
		Format:     output.FormatText,
	}, &buf)

	assert.Equal(t, 0, exitCode, "Should return exit code 0 for success")
	assert.Contains(t, buf.String(), "Validation succeeded", "Should print success message")
}

func TestExitResult_ValidationFailed(t *testing.T) {
	var buf bytes.Buffer
	exitCode := exitResult(commands.ExecuteResult{
		Validation: commands.ValidationFailed,
		Format:     output.FormatText,
	}, &buf)

	assert.Equal(t, 1, exitCode, "Should return exit code 1 for failure")
	assert.Contains(t, buf.String(), "Validation failed", "Should print failure message")
}

func TestExitResult_ValidationSkipped(t *testing.T) {
	var buf bytes.Buffer
	exitCode := exitResult(commands.ExecuteResult{
		Validation: commands.ValidationSkipped,
		Format:     output.FormatText,
	}, &buf)

	assert.Equal(t, 0, exitCode, "Should return exit code 0 for skipped validation")
	assert.NotContains(t, buf.String(), "Validation succeeded", "Should not print success message")
	assert.NotContains(t, buf.String(), "Validation failed", "Should not print failure message")
}

func TestExitResult_ExecutionError(t *testing.T) {
	var buf bytes.Buffer
	exitCode := exitResult(commands.ExecuteResult{
		Validation: commands.ExecutionError,
		Format:     output.FormatText,
	}, &buf)

	assert.Equal(t, 2, exitCode, "Should return exit code 2 for execution error")
	assert.Contains(t, buf.String(), "Execution error", "Should print execution error message")
}

func TestExitResult_OutputFormat(t *testing.T) {
	tests := []struct {
		name           string
		validation     commands.ValidationResult
		expectedExit   int
		expectedOutput string
		shouldContain  bool
	}{
		{
			name:           "Success prints to stdout",
			validation:     commands.ValidationSucceeded,
			expectedExit:   0,
			expectedOutput: "Validation succeeded\n",
			shouldContain:  true,
		},
		{
			name:           "Failure prints to stdout",
			validation:     commands.ValidationFailed,
			expectedExit:   1,
			expectedOutput: "Validation failed\n",
			shouldContain:  true,
		},
		{
			name:           "Skipped produces no output",
			validation:     commands.ValidationSkipped,
			expectedExit:   0,
			expectedOutput: "",
			shouldContain:  false,
		},
		{
			name:           "Execution error prints to stdout",
			validation:     commands.ExecutionError,
			expectedExit:   2,
			expectedOutput: "Execution error\n",
			shouldContain:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			exitCode := exitResult(commands.ExecuteResult{
				Validation: tt.validation,
				Format:     output.FormatText,
			}, &buf)

			assert.Equal(t, tt.expectedExit, exitCode)
			if tt.shouldContain {
				assert.Equal(t, tt.expectedOutput, buf.String())
			} else {
				assert.Empty(t, buf.String())
			}
		})
	}
}

func TestExitResult_JSONMode(t *testing.T) {
	tests := []struct {
		name         string
		validation   commands.ValidationResult
		expectedExit int
	}{
		{
			name:         "ValidationFailed in JSON mode returns exit 1",
			validation:   commands.ValidationFailed,
			expectedExit: 1,
		},
		{
			name:         "ValidationSucceeded in JSON mode returns exit 0",
			validation:   commands.ValidationSucceeded,
			expectedExit: 0,
		},
		{
			name:         "ExecutionError in JSON mode returns exit 2",
			validation:   commands.ExecutionError,
			expectedExit: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			exitCode := exitResult(commands.ExecuteResult{
				Validation: tt.validation,
				Format:     output.FormatJSON,
			}, &buf)

			assert.Equal(t, tt.expectedExit, exitCode)
		})
	}
}

func TestExitResult_JSONMode_SuppressesTextOutput(t *testing.T) {
	tests := []struct {
		name       string
		validation commands.ValidationResult
	}{
		{"ValidationSucceeded suppresses text", commands.ValidationSucceeded},
		{"ValidationFailed suppresses text", commands.ValidationFailed},
		{"ExecutionError suppresses text", commands.ExecutionError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			exitResult(commands.ExecuteResult{
				Validation: tt.validation,
				Format:     output.FormatJSON,
			}, &buf)

			assert.Empty(t, buf.String(), "JSON mode should not print status messages")
		})
	}
}

// TestExitResult_WriteError exercises the defensive fmt.Fprintf(os.Stderr, ...) branches
// inside exitResult that are reached when writing to stdout fails. The correct exit code
// must still be returned even when the write fails.
func TestExitResult_WriteError(t *testing.T) {
	tests := []struct {
		name         string
		validation   commands.ValidationResult
		expectedCode int
	}{
		{
			name:         "ExecutionError still returns 2 on write failure",
			validation:   commands.ExecutionError,
			expectedCode: 2,
		},
		{
			name:         "ValidationFailed still returns 1 on write failure",
			validation:   commands.ValidationFailed,
			expectedCode: 1,
		},
		{
			name:         "ValidationSucceeded still returns 0 on write failure",
			validation:   commands.ValidationSucceeded,
			expectedCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := exitResult(commands.ExecuteResult{
				Validation: tt.validation,
				Format:     output.FormatText,
			}, errWriter{})

			assert.Equal(t, tt.expectedCode, code)
		})
	}
}
