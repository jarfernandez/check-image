package user

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadUserPolicy_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "policy.json")
	content := `{
		"min-uid": 1000,
		"max-uid": 65534,
		"blocked-users": ["daemon", "nobody"],
		"require-numeric": true
	}`
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	policy, err := LoadUserPolicy(path)
	require.NoError(t, err)
	require.NotNil(t, policy)

	require.NotNil(t, policy.MinUID)
	assert.Equal(t, uint(1000), *policy.MinUID)
	require.NotNil(t, policy.MaxUID)
	assert.Equal(t, uint(65534), *policy.MaxUID)
	assert.Equal(t, []string{"daemon", "nobody"}, policy.BlockedUsers)
	require.NotNil(t, policy.RequireNumeric)
	assert.True(t, *policy.RequireNumeric)
}

func TestLoadUserPolicy_YAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "policy.yaml")
	content := `min-uid: 500
max-uid: 60000
blocked-users:
  - bin
  - sys
require-numeric: false
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	policy, err := LoadUserPolicy(path)
	require.NoError(t, err)
	require.NotNil(t, policy)

	require.NotNil(t, policy.MinUID)
	assert.Equal(t, uint(500), *policy.MinUID)
	require.NotNil(t, policy.MaxUID)
	assert.Equal(t, uint(60000), *policy.MaxUID)
	assert.Equal(t, []string{"bin", "sys"}, policy.BlockedUsers)
	require.NotNil(t, policy.RequireNumeric)
	assert.False(t, *policy.RequireNumeric)
}

func TestLoadUserPolicy_PartialFields(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "policy.json")
	content := `{"min-uid": 1000}`
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	policy, err := LoadUserPolicy(path)
	require.NoError(t, err)
	require.NotNil(t, policy)

	require.NotNil(t, policy.MinUID)
	assert.Equal(t, uint(1000), *policy.MinUID)
	assert.Nil(t, policy.MaxUID)
	assert.Nil(t, policy.BlockedUsers)
	assert.Nil(t, policy.RequireNumeric)
}

func TestLoadUserPolicy_EmptyObject(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "policy.json")
	require.NoError(t, os.WriteFile(path, []byte(`{}`), 0600))

	policy, err := LoadUserPolicy(path)
	require.NoError(t, err)
	require.NotNil(t, policy)

	assert.Nil(t, policy.MinUID)
	assert.Nil(t, policy.MaxUID)
	assert.Nil(t, policy.BlockedUsers)
	assert.Nil(t, policy.RequireNumeric)
}

func TestLoadUserPolicy_NonexistentFile(t *testing.T) {
	_, err := LoadUserPolicy("/nonexistent/policy.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error reading user policy")
}

func TestLoadUserPolicy_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "policy.json")
	require.NoError(t, os.WriteFile(path, []byte(`{invalid}`), 0600))

	_, err := LoadUserPolicy(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}

func TestLoadUserPolicy_Stdin(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, p *Policy)
	}{
		{
			name:  "JSON from stdin",
			input: `{"min-uid": 2000, "require-numeric": true}`,
			validate: func(t *testing.T, p *Policy) {
				require.NotNil(t, p.MinUID)
				assert.Equal(t, uint(2000), *p.MinUID)
				require.NotNil(t, p.RequireNumeric)
				assert.True(t, *p.RequireNumeric)
			},
		},
		{
			name:  "YAML from stdin",
			input: "min-uid: 3000\nblocked-users:\n  - www-data\n",
			validate: func(t *testing.T, p *Policy) {
				require.NotNil(t, p.MinUID)
				assert.Equal(t, uint(3000), *p.MinUID)
				assert.Equal(t, []string{"www-data"}, p.BlockedUsers)
			},
		},
		{
			name:        "empty stdin",
			input:       "",
			wantErr:     true,
			errContains: "stdin is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			r, w, err := os.Pipe()
			require.NoError(t, err)
			os.Stdin = r

			go func() {
				_, _ = w.Write([]byte(tt.input))
				w.Close()
			}()

			policy, err := LoadUserPolicy("-")

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
			name:   "empty policy is valid",
			policy: Policy{},
		},
		{
			name:   "min-uid only",
			policy: Policy{MinUID: ptrUint(1000)},
		},
		{
			name:   "max-uid only",
			policy: Policy{MaxUID: ptrUint(65534)},
		},
		{
			name:   "min-uid equals max-uid",
			policy: Policy{MinUID: ptrUint(1000), MaxUID: ptrUint(1000)},
		},
		{
			name:   "min-uid less than max-uid",
			policy: Policy{MinUID: ptrUint(1000), MaxUID: ptrUint(65534)},
		},
		{
			name:        "min-uid exceeds max-uid",
			policy:      Policy{MinUID: ptrUint(65534), MaxUID: ptrUint(1000)},
			wantErr:     true,
			errContains: "min-uid (65534) must not exceed max-uid (1000)",
		},
		{
			name:   "min-uid zero with max-uid",
			policy: Policy{MinUID: ptrUint(0), MaxUID: ptrUint(65534)},
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
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoadUserPolicy_MinExceedsMax(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "policy.json")
	content := `{"min-uid": 65534, "max-uid": 1000}`
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	_, err := LoadUserPolicy(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "min-uid (65534) must not exceed max-uid (1000)")
}

func ptrUint(v uint) *uint {
	return &v
}
