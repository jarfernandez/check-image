package commands

import (
	"fmt"
	"os"
	"testing"

	"github.com/jarfernandez/check-image/internal/imageutil"
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

// resetAuthState resets all authentication-related global state to defaults.
// Call this via t.Cleanup in any test that modifies auth state.
func resetAuthState(t *testing.T) {
	t.Helper()
	registryUsername = ""
	registryPassword = ""
	registryPasswordStdin = false
	imageutil.ResetKeychain()
	os.Unsetenv("CHECK_IMAGE_USERNAME")
	os.Unsetenv("CHECK_IMAGE_PASSWORD")
}

func TestRootCommandAuthFlags_Exist(t *testing.T) {
	usernameFlag := rootCmd.PersistentFlags().Lookup("username")
	require.NotNil(t, usernameFlag, "flag --username must exist")
	assert.Equal(t, "", usernameFlag.DefValue)

	passwordFlag := rootCmd.PersistentFlags().Lookup("password")
	require.NotNil(t, passwordFlag, "flag --password must exist")
	assert.Equal(t, "", passwordFlag.DefValue)

	passwordStdinFlag := rootCmd.PersistentFlags().Lookup("password-stdin")
	require.NotNil(t, passwordStdinFlag, "flag --password-stdin must exist")
	assert.Equal(t, "false", passwordStdinFlag.DefValue)
}

func TestRootCommandAuth_PasswordAndPasswordStdinMutuallyExclusive(t *testing.T) {
	t.Cleanup(func() { resetAuthState(t) })

	logLevel = "info"
	outputFormat = "text"
	registryUsername = "user"
	registryPassword = "pass"
	registryPasswordStdin = true

	err := rootCmd.PersistentPreRunE(rootCmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--password and --password-stdin are mutually exclusive")
}

func TestRootCommandAuth_UsernameWithoutPassword(t *testing.T) {
	t.Cleanup(func() { resetAuthState(t) })

	logLevel = "info"
	outputFormat = "text"
	registryUsername = "user"
	registryPassword = ""
	registryPasswordStdin = false

	err := rootCmd.PersistentPreRunE(rootCmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "registry password required when username is set")
}

func TestRootCommandAuth_PasswordWithoutUsername(t *testing.T) {
	t.Cleanup(func() { resetAuthState(t) })

	logLevel = "info"
	outputFormat = "text"
	registryUsername = ""
	registryPassword = "pass"
	registryPasswordStdin = false

	err := rootCmd.PersistentPreRunE(rootCmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "registry username required when password is set")
}

func TestRootCommandAuth_ValidCredentialsViaFlags(t *testing.T) {
	t.Cleanup(func() { resetAuthState(t) })

	logLevel = "info"
	outputFormat = "text"
	registryUsername = "myuser"
	registryPassword = "mypass"
	registryPasswordStdin = false

	err := rootCmd.PersistentPreRunE(rootCmd, []string{})
	require.NoError(t, err)

	// Verify the keychain was updated by checking it resolves the credentials
	auth, resolveErr := imageutil.ActiveKeychain().Resolve(nil)
	require.NoError(t, resolveErr)
	cfg, authErr := auth.Authorization()
	require.NoError(t, authErr)
	assert.Equal(t, "myuser", cfg.Username)
	assert.Equal(t, "mypass", cfg.Password)
}

func TestRootCommandAuth_ValidCredentialsViaEnvVars(t *testing.T) {
	t.Cleanup(func() { resetAuthState(t) })

	logLevel = "info"
	outputFormat = "text"
	registryUsername = ""
	registryPassword = ""
	registryPasswordStdin = false

	t.Setenv("CHECK_IMAGE_USERNAME", "envuser")
	t.Setenv("CHECK_IMAGE_PASSWORD", "envpass")

	err := rootCmd.PersistentPreRunE(rootCmd, []string{})
	require.NoError(t, err)

	auth, resolveErr := imageutil.ActiveKeychain().Resolve(nil)
	require.NoError(t, resolveErr)
	cfg, authErr := auth.Authorization()
	require.NoError(t, authErr)
	assert.Equal(t, "envuser", cfg.Username)
	assert.Equal(t, "envpass", cfg.Password)
}

func TestRootCommandAuth_FlagsTakePrecedenceOverEnvVars(t *testing.T) {
	t.Cleanup(func() { resetAuthState(t) })

	logLevel = "info"
	outputFormat = "text"
	registryUsername = "flaguser"
	registryPassword = "flagpass"
	registryPasswordStdin = false

	t.Setenv("CHECK_IMAGE_USERNAME", "envuser")
	t.Setenv("CHECK_IMAGE_PASSWORD", "envpass")

	err := rootCmd.PersistentPreRunE(rootCmd, []string{})
	require.NoError(t, err)

	auth, resolveErr := imageutil.ActiveKeychain().Resolve(nil)
	require.NoError(t, resolveErr)
	cfg, authErr := auth.Authorization()
	require.NoError(t, authErr)
	assert.Equal(t, "flaguser", cfg.Username)
	assert.Equal(t, "flagpass", cfg.Password)
}

func TestRootCommandAuth_EnvUsernameWithoutPassword(t *testing.T) {
	t.Cleanup(func() { resetAuthState(t) })

	logLevel = "info"
	outputFormat = "text"
	registryUsername = ""
	registryPassword = ""
	registryPasswordStdin = false

	t.Setenv("CHECK_IMAGE_USERNAME", "envuser")

	err := rootCmd.PersistentPreRunE(rootCmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "registry password required when username is set")
}

func TestRootCommandAuth_EnvPasswordWithoutUsername(t *testing.T) {
	t.Cleanup(func() { resetAuthState(t) })

	logLevel = "info"
	outputFormat = "text"
	registryUsername = ""
	registryPassword = ""
	registryPasswordStdin = false

	t.Setenv("CHECK_IMAGE_PASSWORD", "envpass")

	err := rootCmd.PersistentPreRunE(rootCmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "registry username required when password is set")
}

func TestRootCommandAuth_NoCredentials_UsesDefaultKeychain(t *testing.T) {
	t.Cleanup(func() { resetAuthState(t) })

	logLevel = "info"
	outputFormat = "text"
	registryUsername = ""
	registryPassword = ""
	registryPasswordStdin = false

	err := rootCmd.PersistentPreRunE(rootCmd, []string{})
	require.NoError(t, err)
	// When no credentials are set, the activeKeychain should remain as DefaultKeychain
	// (SetStaticCredentials is not called), so no error occurs
}

func TestRootCommandAuth_PasswordStdin_ReadsFromStdin(t *testing.T) {
	t.Cleanup(func() { resetAuthState(t) })

	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = origStdin
		r.Close()
	})

	// Write password to the write-end of the pipe then close it
	_, err = fmt.Fprintln(w, "stdinpass")
	require.NoError(t, err)
	w.Close()

	logLevel = "info"
	outputFormat = "text"
	registryUsername = "stdinuser"
	registryPassword = ""
	registryPasswordStdin = true

	err = rootCmd.PersistentPreRunE(rootCmd, []string{})
	require.NoError(t, err)

	auth, resolveErr := imageutil.ActiveKeychain().Resolve(nil)
	require.NoError(t, resolveErr)
	cfg, authErr := auth.Authorization()
	require.NoError(t, authErr)
	assert.Equal(t, "stdinuser", cfg.Username)
	assert.Equal(t, "stdinpass", cfg.Password) // trailing newline stripped
}

func TestRootCommandAuth_PasswordStdin_StripsTrailingNewlines(t *testing.T) {
	t.Cleanup(func() { resetAuthState(t) })

	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = origStdin
		r.Close()
	})

	// Write password with \r\n (Windows-style) line ending
	_, err = fmt.Fprint(w, "tokenvalue\r\n")
	require.NoError(t, err)
	w.Close()

	logLevel = "info"
	outputFormat = "text"
	registryUsername = "tokenuser"
	registryPassword = ""
	registryPasswordStdin = true

	err = rootCmd.PersistentPreRunE(rootCmd, []string{})
	require.NoError(t, err)

	auth, resolveErr := imageutil.ActiveKeychain().Resolve(nil)
	require.NoError(t, resolveErr)
	cfg, authErr := auth.Authorization()
	require.NoError(t, authErr)
	assert.Equal(t, "tokenvalue", cfg.Password)
}

func TestRootCommandAuth_MixedFlagAndEnv_UsernameFromFlag_PasswordFromEnv(t *testing.T) {
	t.Cleanup(func() { resetAuthState(t) })

	logLevel = "info"
	outputFormat = "text"
	registryUsername = "flaguser"
	registryPassword = ""
	registryPasswordStdin = false

	t.Setenv("CHECK_IMAGE_PASSWORD", "envpass")

	err := rootCmd.PersistentPreRunE(rootCmd, []string{})
	require.NoError(t, err)

	auth, resolveErr := imageutil.ActiveKeychain().Resolve(nil)
	require.NoError(t, resolveErr)
	cfg, authErr := auth.Authorization()
	require.NoError(t, authErr)
	assert.Equal(t, "flaguser", cfg.Username)
	assert.Equal(t, "envpass", cfg.Password)
}
