package commands

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jarfernandez/check-image/internal/output"
	"github.com/jarfernandez/check-image/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCommand(t *testing.T) {
	assert.NotNil(t, userCmd)
	assert.Equal(t, "user image", userCmd.Use)
	assert.Contains(t, userCmd.Short, "user")

	// Test that it requires exactly 1 argument
	assert.NotNil(t, userCmd.Args)
	err := userCmd.Args(userCmd, []string{})
	assert.Error(t, err)

	err = userCmd.Args(userCmd, []string{"image"})
	assert.NoError(t, err)

	err = userCmd.Args(userCmd, []string{"image1", "image2"})
	assert.Error(t, err)
}

func TestUserCommandFlags(t *testing.T) {
	flag := userCmd.Flags().Lookup("user-policy")
	assert.NotNil(t, flag)
	assert.Equal(t, "", flag.DefValue)

	flag = userCmd.Flags().Lookup("min-uid")
	assert.NotNil(t, flag)
	assert.Equal(t, "0", flag.DefValue)

	flag = userCmd.Flags().Lookup("max-uid")
	assert.NotNil(t, flag)
	assert.Equal(t, "0", flag.DefValue)

	flag = userCmd.Flags().Lookup("blocked-users")
	assert.NotNil(t, flag)
	assert.Equal(t, "", flag.DefValue)

	flag = userCmd.Flags().Lookup("require-numeric")
	assert.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
}

func TestRunUser_BasicNonRoot(t *testing.T) {
	tests := []struct {
		name         string
		user         string
		expectedPass bool
		expectedMsg  string
	}{
		{
			name:         "non-root UID",
			user:         "1000",
			expectedPass: true,
			expectedMsg:  "Image user meets all requirements",
		},
		{
			name:         "non-root username",
			user:         "appuser",
			expectedPass: true,
			expectedMsg:  "Image user meets all requirements",
		},
		{
			name:         "non-root UID with GID",
			user:         "1000:1000",
			expectedPass: true,
			expectedMsg:  "Image user meets all requirements",
		},
		{
			name:         "root user",
			user:         "root",
			expectedPass: false,
			expectedMsg:  "Image user does not meet requirements",
		},
		{
			name:         "root UID 0",
			user:         "0",
			expectedPass: false,
			expectedMsg:  "Image user does not meet requirements",
		},
		{
			name:         "root UID 0 with GID",
			user:         "0:0",
			expectedPass: false,
			expectedMsg:  "Image user does not meet requirements",
		},
		{
			name:         "empty user defaults to root",
			user:         "",
			expectedPass: false,
			expectedMsg:  "Image user does not meet requirements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imageRef := createTestImage(t, testImageOptions{
				user:    tt.user,
				created: time.Now(),
			})

			result, err := runUser(context.Background(), imageRef, nil)
			require.NoError(t, err)

			assert.Equal(t, checkUser, result.Check)
			assert.Equal(t, imageRef, result.Image)
			assert.Equal(t, tt.expectedPass, result.Passed)
			assert.Equal(t, tt.expectedMsg, result.Message)

			details, ok := result.Details.(output.UserDetails)
			require.True(t, ok)
			assert.Equal(t, tt.user, details.User)
		})
	}
}

func TestRunUser_WithPolicy(t *testing.T) {
	tests := []struct {
		name           string
		user           string
		policy         *user.Policy
		expectedPass   bool
		violationRules []string
	}{
		{
			name:         "UID in range passes",
			user:         "1000",
			policy:       &user.Policy{MinUID: ptrUintCmd(1000), MaxUID: ptrUintCmd(65534)},
			expectedPass: true,
		},
		{
			name:           "UID below minimum fails",
			user:           "500",
			policy:         &user.Policy{MinUID: ptrUintCmd(1000)},
			expectedPass:   false,
			violationRules: []string{"min-uid"},
		},
		{
			name:           "UID above maximum fails",
			user:           "70000",
			policy:         &user.Policy{MaxUID: ptrUintCmd(65534)},
			expectedPass:   false,
			violationRules: []string{"max-uid"},
		},
		{
			name:           "blocked username fails",
			user:           "daemon",
			policy:         &user.Policy{BlockedUsers: []string{"daemon", "nobody"}},
			expectedPass:   false,
			violationRules: []string{"blocked-user"},
		},
		{
			name:           "require numeric with username fails",
			user:           "appuser",
			policy:         &user.Policy{RequireNumeric: ptrBoolCmd(true)},
			expectedPass:   false,
			violationRules: []string{"require-numeric"},
		},
		{
			name:         "require numeric with UID passes",
			user:         "1000",
			policy:       &user.Policy{RequireNumeric: ptrBoolCmd(true)},
			expectedPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imageRef := createTestImage(t, testImageOptions{
				user:    tt.user,
				created: time.Now(),
			})

			result, err := runUser(context.Background(), imageRef, tt.policy)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedPass, result.Passed)

			details, ok := result.Details.(output.UserDetails)
			require.True(t, ok)

			if tt.violationRules != nil {
				require.Len(t, details.Violations, len(tt.violationRules))
				for i, rule := range tt.violationRules {
					assert.Equal(t, rule, details.Violations[i].Rule)
				}
			} else {
				assert.Empty(t, details.Violations)
			}
		})
	}
}

func TestRunUser_DetailsIncludePolicyConstraints(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		user:    "1000",
		created: time.Now(),
	})

	policy := &user.Policy{
		MinUID:         ptrUintCmd(1000),
		MaxUID:         ptrUintCmd(65534),
		BlockedUsers:   []string{"daemon"},
		RequireNumeric: ptrBoolCmd(true),
	}

	result, err := runUser(context.Background(), imageRef, policy)
	require.NoError(t, err)

	details, ok := result.Details.(output.UserDetails)
	require.True(t, ok)

	require.NotNil(t, details.MinUID)
	assert.Equal(t, uint(1000), *details.MinUID)
	require.NotNil(t, details.MaxUID)
	assert.Equal(t, uint(65534), *details.MaxUID)
	assert.Equal(t, []string{"daemon"}, details.BlockedUsers)
	assert.True(t, details.RequireNumeric)
}

func TestRunUser_DetailsOmitPolicyWhenNil(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		user:    "1000",
		created: time.Now(),
	})

	result, err := runUser(context.Background(), imageRef, nil)
	require.NoError(t, err)

	details, ok := result.Details.(output.UserDetails)
	require.True(t, ok)

	assert.Nil(t, details.MinUID)
	assert.Nil(t, details.MaxUID)
	assert.Nil(t, details.BlockedUsers)
	assert.False(t, details.RequireNumeric)
}

func TestRunUser_UserPartAndGroupPart(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{
		user:    "1000:2000",
		created: time.Now(),
	})

	result, err := runUser(context.Background(), imageRef, nil)
	require.NoError(t, err)

	details, ok := result.Details.(output.UserDetails)
	require.True(t, ok)

	assert.Equal(t, "1000:2000", details.User)
	assert.Equal(t, "1000", details.UserPart)
	assert.Equal(t, "2000", details.GroupPart)
	assert.True(t, details.IsNumeric)
	require.NotNil(t, details.UID)
	assert.Equal(t, uint64(1000), *details.UID)
}

func TestRunUser_InvalidImage(t *testing.T) {
	_, err := runUser(context.Background(), "nonexistent:image", nil)
	require.Error(t, err)
}

func TestParseBlockedUsers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{name: "empty string", input: "", expected: nil},
		{name: "single user", input: "daemon", expected: []string{"daemon"}},
		{name: "multiple users", input: "daemon,nobody,www-data", expected: []string{"daemon", "nobody", "www-data"}},
		{name: "with spaces", input: " daemon , nobody ", expected: []string{"daemon", "nobody"}},
		{name: "trailing comma", input: "daemon,nobody,", expected: []string{"daemon", "nobody"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBlockedUsers(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// resetUserCmdFlags resets the Changed state of all user-related flags on
// userCmd. Cobra's pflag.Flag.Changed persists across subtests when using
// a shared command, so we must clear it explicitly.
func resetUserCmdFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"user-policy", "min-uid", "max-uid", "blocked-users", "require-numeric"} {
		f := userCmd.Flags().Lookup(name)
		if f != nil {
			f.Changed = false
		}
	}
	t.Cleanup(func() {
		for _, name := range []string{"user-policy", "min-uid", "max-uid", "blocked-users", "require-numeric"} {
			f := userCmd.Flags().Lookup(name)
			if f != nil {
				f.Changed = false
			}
		}
	})
}

func TestResolveUserPolicy(t *testing.T) {
	t.Run("no flags no policy returns nil", func(t *testing.T) {
		resetAllGlobals(t)
		resetUserCmdFlags(t)

		policy, err := resolveUserPolicy(userCmd)
		require.NoError(t, err)
		assert.Nil(t, policy)
	})

	t.Run("policy file only", func(t *testing.T) {
		resetAllGlobals(t)
		resetUserCmdFlags(t)

		tmpDir := t.TempDir()
		policyPath := filepath.Join(tmpDir, "policy.yaml")
		content := "min-uid: 1000\nmax-uid: 65534\n"
		require.NoError(t, os.WriteFile(policyPath, []byte(content), 0600))

		userPolicy = policyPath

		policy, err := resolveUserPolicy(userCmd)
		require.NoError(t, err)
		require.NotNil(t, policy)
		require.NotNil(t, policy.MinUID)
		assert.Equal(t, uint(1000), *policy.MinUID)
		require.NotNil(t, policy.MaxUID)
		assert.Equal(t, uint(65534), *policy.MaxUID)
	})

	t.Run("flags only create policy", func(t *testing.T) {
		resetAllGlobals(t)
		resetUserCmdFlags(t)

		require.NoError(t, userCmd.Flags().Set("min-uid", "500"))
		require.NoError(t, userCmd.Flags().Set("blocked-users", "daemon,nobody"))

		policy, err := resolveUserPolicy(userCmd)
		require.NoError(t, err)
		require.NotNil(t, policy)
		require.NotNil(t, policy.MinUID)
		assert.Equal(t, uint(500), *policy.MinUID)
		assert.Equal(t, []string{"daemon", "nobody"}, policy.BlockedUsers)
		// max-uid was not set, so should remain nil
		assert.Nil(t, policy.MaxUID)
	})

	t.Run("flags override policy file", func(t *testing.T) {
		resetAllGlobals(t)
		resetUserCmdFlags(t)

		tmpDir := t.TempDir()
		policyPath := filepath.Join(tmpDir, "policy.yaml")
		content := "min-uid: 2000\nmax-uid: 65534\nblocked-users:\n  - daemon\n"
		require.NoError(t, os.WriteFile(policyPath, []byte(content), 0600))

		userPolicy = policyPath
		require.NoError(t, userCmd.Flags().Set("min-uid", "500"))

		policy, err := resolveUserPolicy(userCmd)
		require.NoError(t, err)
		require.NotNil(t, policy)
		// min-uid overridden by CLI
		require.NotNil(t, policy.MinUID)
		assert.Equal(t, uint(500), *policy.MinUID)
		// max-uid from policy file
		require.NotNil(t, policy.MaxUID)
		assert.Equal(t, uint(65534), *policy.MaxUID)
		// blocked-users from policy file (not overridden)
		assert.Equal(t, []string{"daemon"}, policy.BlockedUsers)
	})

	t.Run("require-numeric flag", func(t *testing.T) {
		resetAllGlobals(t)
		resetUserCmdFlags(t)

		require.NoError(t, userCmd.Flags().Set("require-numeric", "true"))

		policy, err := resolveUserPolicy(userCmd)
		require.NoError(t, err)
		require.NotNil(t, policy)
		require.NotNil(t, policy.RequireNumeric)
		assert.True(t, *policy.RequireNumeric)
	})

	t.Run("max-uid flag only", func(t *testing.T) {
		resetAllGlobals(t)
		resetUserCmdFlags(t)

		require.NoError(t, userCmd.Flags().Set("max-uid", "65534"))

		policy, err := resolveUserPolicy(userCmd)
		require.NoError(t, err)
		require.NotNil(t, policy)
		require.NotNil(t, policy.MaxUID)
		assert.Equal(t, uint(65534), *policy.MaxUID)
		assert.Nil(t, policy.MinUID)
	})

	t.Run("invalid policy file returns error", func(t *testing.T) {
		resetAllGlobals(t)
		resetUserCmdFlags(t)

		userPolicy = "/nonexistent/policy.yaml"

		_, err := resolveUserPolicy(userCmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unable to load user policy")
	})

	t.Run("invalid min-uid > max-uid returns validation error", func(t *testing.T) {
		resetAllGlobals(t)
		resetUserCmdFlags(t)

		require.NoError(t, userCmd.Flags().Set("min-uid", "5000"))
		require.NoError(t, userCmd.Flags().Set("max-uid", "1000"))

		_, err := resolveUserPolicy(userCmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "min-uid")
	})
}

func TestRunUser_WithPolicyFile_YAML(t *testing.T) {
	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "policy.yaml")
	content := "min-uid: 1000\nmax-uid: 65534\nblocked-users:\n  - daemon\n"
	require.NoError(t, os.WriteFile(policyPath, []byte(content), 0600))

	policy, err := user.LoadUserPolicy(policyPath)
	require.NoError(t, err)

	imageRef := createTestImage(t, testImageOptions{user: "1500", created: time.Now()})
	result, err := runUser(context.Background(), imageRef, policy)
	require.NoError(t, err)
	assert.True(t, result.Passed)

	imageRef2 := createTestImage(t, testImageOptions{user: "500", created: time.Now()})
	result2, err := runUser(context.Background(), imageRef2, policy)
	require.NoError(t, err)
	assert.False(t, result2.Passed)
}

func TestRunUser_WithPolicyFile_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "policy.json")
	content := `{"min-uid": 1000, "blocked-users": ["nobody"]}`
	require.NoError(t, os.WriteFile(policyPath, []byte(content), 0600))

	policy, err := user.LoadUserPolicy(policyPath)
	require.NoError(t, err)

	imageRef := createTestImage(t, testImageOptions{user: "nobody", created: time.Now()})
	result, err := runUser(context.Background(), imageRef, policy)
	require.NoError(t, err)
	assert.False(t, result.Passed)

	details := result.Details.(output.UserDetails)
	require.Len(t, details.Violations, 1)
	assert.Equal(t, "blocked-user", details.Violations[0].Rule)
}

func TestRunUser_WithPolicyFile_Stdin(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdin = r

	go func() {
		_, _ = w.Write([]byte(`{"min-uid": 2000}`))
		w.Close()
	}()

	policy, err := user.LoadUserPolicy("-")
	require.NoError(t, err)

	imageRef := createTestImage(t, testImageOptions{user: "1000", created: time.Now()})
	result, err := runUser(context.Background(), imageRef, policy)
	require.NoError(t, err)
	assert.False(t, result.Passed)

	details := result.Details.(output.UserDetails)
	require.Len(t, details.Violations, 1)
	assert.Equal(t, "min-uid", details.Violations[0].Rule)
}

func TestRunUser_CLIOverridesPolicy(t *testing.T) {
	// Policy says min-uid: 2000, but CLI override lowers to 500
	basePolicy := &user.Policy{MinUID: ptrUintCmd(2000)}

	// Override min-uid to 500
	overridden := uint(500)
	basePolicy.MinUID = &overridden

	imageRef := createTestImage(t, testImageOptions{user: "1000", created: time.Now()})
	result, err := runUser(context.Background(), imageRef, basePolicy)
	require.NoError(t, err)
	assert.True(t, result.Passed)
}

func TestRunUser_JSONOutput(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{user: "500", created: time.Now()})
	policy := &user.Policy{MinUID: ptrUintCmd(1000), MaxUID: ptrUintCmd(65534)}

	result, err := runUser(context.Background(), imageRef, policy)
	require.NoError(t, err)

	// Marshal to JSON and verify structure
	data, err := json.Marshal(result)
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(data, &parsed))

	assert.Equal(t, "user", parsed["check"])
	assert.Equal(t, false, parsed["passed"])
	assert.Contains(t, parsed["message"], "does not meet requirements")

	details := parsed["details"].(map[string]any)
	assert.Equal(t, "500", details["user"])
	assert.Equal(t, "500", details["user-part"])
	assert.Equal(t, true, details["is-numeric"])
	assert.Equal(t, float64(500), details["uid"])
	assert.Equal(t, float64(1000), details["min-uid"])
	assert.Equal(t, float64(65534), details["max-uid"])

	violations := details["violations"].([]any)
	require.Len(t, violations, 1)
	v := violations[0].(map[string]any)
	assert.Equal(t, "min-uid", v["rule"])
	assert.Contains(t, v["message"], "below minimum")
}

func TestRunUser_JSONOutput_Passing(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{user: "1000", created: time.Now()})

	result, err := runUser(context.Background(), imageRef, nil)
	require.NoError(t, err)

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(data, &parsed))

	assert.Equal(t, "user", parsed["check"])
	assert.Equal(t, true, parsed["passed"])

	details := parsed["details"].(map[string]any)
	assert.Equal(t, "1000", details["user"])
	assert.Equal(t, true, details["is-numeric"])
	// No violations key when empty (omitempty)
	assert.Nil(t, details["violations"])
	// No policy fields when no policy
	assert.Nil(t, details["min-uid"])
	assert.Nil(t, details["max-uid"])
}

func TestRunUser_TextOutput_WithViolations(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{user: "500", created: time.Now()})
	policy := &user.Policy{MinUID: ptrUintCmd(1000)}

	result, err := runUser(context.Background(), imageRef, policy)
	require.NoError(t, err)

	out := captureStdout(t, func() {
		renderUserText(result)
	})

	assert.Contains(t, out, "Checking USER directive of image")
	assert.Contains(t, out, "USER:")
	assert.Contains(t, out, "500")
	assert.Contains(t, out, "Violations:")
	assert.Contains(t, out, "below minimum 1000")
	assert.Contains(t, out, "does not meet requirements")
}

func TestRunUser_TextOutput_NotSet(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{user: "", created: time.Now()})

	result, err := runUser(context.Background(), imageRef, nil)
	require.NoError(t, err)

	out := captureStdout(t, func() {
		renderUserText(result)
	})

	assert.Contains(t, out, "(not set)")
	assert.Contains(t, out, "Violations:")
	assert.Contains(t, out, "not set (defaults to root)")
}

func TestRunUser_TextOutput_Passing(t *testing.T) {
	imageRef := createTestImage(t, testImageOptions{user: "1000", created: time.Now()})

	result, err := runUser(context.Background(), imageRef, nil)
	require.NoError(t, err)

	out := captureStdout(t, func() {
		renderUserText(result)
	})

	assert.Contains(t, out, "USER:")
	assert.Contains(t, out, "1000")
	assert.Contains(t, out, "meets all requirements")
	assert.NotContains(t, out, "Violations:")
}

func ptrUintCmd(v uint) *uint {
	return &v
}

func ptrBoolCmd(v bool) *bool {
	return &v
}
