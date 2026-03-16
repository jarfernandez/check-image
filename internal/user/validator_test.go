package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseUser(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		userPart  string
		groupPart string
		isNumeric bool
		uid       *uint64
	}{
		{
			name: "empty string",
		},
		{
			name:      "username only",
			input:     "appuser",
			userPart:  "appuser",
			isNumeric: false,
		},
		{
			name:      "numeric UID only",
			input:     "1000",
			userPart:  "1000",
			isNumeric: true,
			uid:       new(uint64(1000)),
		},
		{
			name:      "UID zero",
			input:     "0",
			userPart:  "0",
			isNumeric: true,
			uid:       new(uint64(0)),
		},
		{
			name:      "UID with GID",
			input:     "1000:1000",
			userPart:  "1000",
			groupPart: "1000",
			isNumeric: true,
			uid:       new(uint64(1000)),
		},
		{
			name:      "username with group",
			input:     "appuser:appgroup",
			userPart:  "appuser",
			groupPart: "appgroup",
			isNumeric: false,
		},
		{
			name:      "root username",
			input:     "root",
			userPart:  "root",
			isNumeric: false,
		},
		{
			name:      "root with group",
			input:     "root:root",
			userPart:  "root",
			groupPart: "root",
			isNumeric: false,
		},
		{
			name:      "UID zero with named group",
			input:     "0:somegroup",
			userPart:  "0",
			groupPart: "somegroup",
			isNumeric: true,
			uid:       new(uint64(0)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := ParseUser(tt.input)
			assert.Equal(t, tt.input, info.Raw)
			assert.Equal(t, tt.userPart, info.UserPart)
			assert.Equal(t, tt.groupPart, info.GroupPart)
			assert.Equal(t, tt.isNumeric, info.IsNumeric)
			if tt.uid != nil {
				require.NotNil(t, info.UID)
				assert.Equal(t, *tt.uid, *info.UID)
			} else {
				assert.Nil(t, info.UID)
			}
		})
	}
}

func TestValidateUser_BasicNonRoot(t *testing.T) {
	tests := []struct {
		name       string
		user       string
		passed     bool
		violations []string // expected rule names
	}{
		{
			name:       "empty string defaults to root",
			user:       "",
			passed:     false,
			violations: []string{"non-empty"},
		},
		{
			name:       "explicit root",
			user:       "root",
			passed:     false,
			violations: []string{"non-root"},
		},
		{
			name:       "UID 0",
			user:       "0",
			passed:     false,
			violations: []string{"non-root-uid"},
		},
		{
			name:       "UID 0 with GID",
			user:       "0:0",
			passed:     false,
			violations: []string{"non-root-uid"},
		},
		{
			name:       "UID 0 with named group",
			user:       "0:somegroup",
			passed:     false,
			violations: []string{"non-root-uid"},
		},
		{
			name:   "non-root UID",
			user:   "1000",
			passed: true,
		},
		{
			name:   "non-root username",
			user:   "appuser",
			passed: true,
		},
		{
			name:   "non-root UID with GID",
			user:   "1000:1000",
			passed: true,
		},
		{
			name:   "non-root username with group",
			user:   "appuser:appgroup",
			passed: true,
		},
		{
			name:   "UID 1",
			user:   "1",
			passed: true,
		},
		{
			name:   "large UID",
			user:   "65534",
			passed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateUser(tt.user, nil)
			assert.Equal(t, tt.passed, result.Passed)
			if tt.violations != nil {
				require.Len(t, result.Violations, len(tt.violations))
				for i, rule := range tt.violations {
					assert.Equal(t, rule, result.Violations[i].Rule)
				}
			} else {
				assert.Empty(t, result.Violations)
			}
		})
	}
}

func TestValidateUser_WithPolicy(t *testing.T) {
	tests := []struct {
		name       string
		user       string
		policy     *Policy
		passed     bool
		violations []string // expected rule names
	}{
		{
			name:   "UID in range passes",
			user:   "1000",
			policy: &Policy{MinUID: new(uint(1000)), MaxUID: new(uint(65534))},
			passed: true,
		},
		{
			name:       "UID below minimum",
			user:       "500",
			policy:     &Policy{MinUID: new(uint(1000))},
			passed:     false,
			violations: []string{"min-uid"},
		},
		{
			name:       "UID above maximum",
			user:       "70000",
			policy:     &Policy{MaxUID: new(uint(65534))},
			passed:     false,
			violations: []string{"max-uid"},
		},
		{
			name:       "UID below minimum and above maximum",
			user:       "500",
			policy:     &Policy{MinUID: new(uint(1000)), MaxUID: new(uint(400))},
			passed:     false,
			violations: []string{"min-uid", "max-uid"},
		},
		{
			name:       "blocked username",
			user:       "daemon",
			policy:     &Policy{BlockedUsers: []string{"daemon", "nobody"}},
			passed:     false,
			violations: []string{"blocked-user"},
		},
		{
			name:   "non-blocked username",
			user:   "appuser",
			policy: &Policy{BlockedUsers: []string{"daemon", "nobody"}},
			passed: true,
		},
		{
			name:       "require numeric with username",
			user:       "appuser",
			policy:     &Policy{RequireNumeric: new(true)},
			passed:     false,
			violations: []string{"require-numeric"},
		},
		{
			name:   "require numeric with UID",
			user:   "1000",
			policy: &Policy{RequireNumeric: new(true)},
			passed: true,
		},
		{
			name:   "require numeric false with username",
			user:   "appuser",
			policy: &Policy{RequireNumeric: new(false)},
			passed: true,
		},
		{
			name:   "username not affected by UID range",
			user:   "appuser",
			policy: &Policy{MinUID: new(uint(1000)), MaxUID: new(uint(65534))},
			passed: true,
		},
		{
			name:       "username blocked and require-numeric",
			user:       "daemon",
			policy:     &Policy{BlockedUsers: []string{"daemon"}, RequireNumeric: new(true)},
			passed:     false,
			violations: []string{"require-numeric", "blocked-user"},
		},
		{
			name:   "UID at exact minimum boundary",
			user:   "1000",
			policy: &Policy{MinUID: new(uint(1000))},
			passed: true,
		},
		{
			name:   "UID at exact maximum boundary",
			user:   "65534",
			policy: &Policy{MaxUID: new(uint(65534))},
			passed: true,
		},
		{
			name:       "UID one below minimum",
			user:       "999",
			policy:     &Policy{MinUID: new(uint(1000))},
			passed:     false,
			violations: []string{"min-uid"},
		},
		{
			name:       "UID one above maximum",
			user:       "65535",
			policy:     &Policy{MaxUID: new(uint(65534))},
			passed:     false,
			violations: []string{"max-uid"},
		},
		{
			name:   "UID with group and policy passes",
			user:   "1000:1000",
			policy: &Policy{MinUID: new(uint(1000)), MaxUID: new(uint(65534))},
			passed: true,
		},
		{
			name:   "empty policy passes for non-root",
			user:   "1000",
			policy: &Policy{},
			passed: true,
		},
		{
			name:   "blocked list with no match",
			user:   "1000",
			policy: &Policy{BlockedUsers: []string{"daemon", "nobody"}},
			passed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateUser(tt.user, tt.policy)
			assert.Equal(t, tt.passed, result.Passed)
			if tt.violations != nil {
				require.Len(t, result.Violations, len(tt.violations))
				for i, rule := range tt.violations {
					assert.Equal(t, rule, result.Violations[i].Rule)
				}
			} else {
				assert.Empty(t, result.Violations)
			}
		})
	}
}

func TestValidateUser_AlwaysEnforcedRegardlessOfPolicy(t *testing.T) {
	// Even with a permissive policy, root checks are always enforced
	policy := &Policy{MinUID: new(uint(0)), MaxUID: new(uint(65534))}

	t.Run("empty user fails even with policy", func(t *testing.T) {
		result := ValidateUser("", policy)
		assert.False(t, result.Passed)
		require.Len(t, result.Violations, 1)
		assert.Equal(t, "non-empty", result.Violations[0].Rule)
	})

	t.Run("root user fails even with policy", func(t *testing.T) {
		result := ValidateUser("root", policy)
		assert.False(t, result.Passed)
		require.Len(t, result.Violations, 1)
		assert.Equal(t, "non-root", result.Violations[0].Rule)
	})

	t.Run("UID 0 fails even with min-uid 0", func(t *testing.T) {
		result := ValidateUser("0", policy)
		assert.False(t, result.Passed)
		require.Len(t, result.Violations, 1)
		assert.Equal(t, "non-root-uid", result.Violations[0].Rule)
	})
}

func TestValidateUser_ViolationMessages(t *testing.T) {
	t.Run("non-empty violation message", func(t *testing.T) {
		result := ValidateUser("", nil)
		require.Len(t, result.Violations, 1)
		assert.Equal(t, "user must not be root", result.Violations[0].Message)
	})

	t.Run("min-uid violation message", func(t *testing.T) {
		result := ValidateUser("500", &Policy{MinUID: new(uint(1000))})
		require.Len(t, result.Violations, 1)
		assert.Equal(t, "UID 500 is below minimum 1000", result.Violations[0].Message)
	})

	t.Run("max-uid violation message", func(t *testing.T) {
		result := ValidateUser("70000", &Policy{MaxUID: new(uint(65534))})
		require.Len(t, result.Violations, 1)
		assert.Equal(t, "UID 70000 is above maximum 65534", result.Violations[0].Message)
	})

	t.Run("blocked-user violation message", func(t *testing.T) {
		result := ValidateUser("daemon", &Policy{BlockedUsers: []string{"daemon"}})
		require.Len(t, result.Violations, 1)
		assert.Equal(t, `user "daemon" is in the blocked users list`, result.Violations[0].Message)
	})

	t.Run("require-numeric violation message", func(t *testing.T) {
		result := ValidateUser("appuser", &Policy{RequireNumeric: new(true)})
		require.Len(t, result.Violations, 1)
		assert.Equal(t, `user "appuser" must be a numeric UID`, result.Violations[0].Message)
	})
}
