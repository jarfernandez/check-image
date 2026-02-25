package commands

import (
	"errors"
	"testing"

	"github.com/jarfernandez/check-image/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCheckCmd_Success(t *testing.T) {
	Result = ValidationSkipped
	t.Cleanup(func() { Result = ValidationSkipped })

	result := &output.CheckResult{
		Check:   "test",
		Image:   "nginx:latest",
		Passed:  true,
		Message: "ok",
	}
	err := runCheckCmd("test", func(string) (*output.CheckResult, error) {
		return result, nil
	}, "nginx:latest")
	require.NoError(t, err)
	assert.Equal(t, ValidationSucceeded, Result)
}

func TestRunCheckCmd_Failure(t *testing.T) {
	Result = ValidationSkipped
	t.Cleanup(func() { Result = ValidationSkipped })

	result := &output.CheckResult{
		Check:   "test",
		Image:   "nginx:latest",
		Passed:  false,
		Message: "not ok",
	}
	err := runCheckCmd("test", func(string) (*output.CheckResult, error) {
		return result, nil
	}, "nginx:latest")
	require.NoError(t, err)
	assert.Equal(t, ValidationFailed, Result)
}

func TestRunCheckCmd_RunError(t *testing.T) {
	err := runCheckCmd("mycheck", func(string) (*output.CheckResult, error) {
		return nil, errors.New("something went wrong")
	}, "nginx:latest")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "check mycheck operation failed")
	assert.Contains(t, err.Error(), "something went wrong")
}
