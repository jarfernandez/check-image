package imageutil

import (
	"testing"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticKeychain_Resolve(t *testing.T) {
	kc := &staticKeychain{registry: "ghcr.io", username: "user", password: "pass"}

	auth, err := kc.Resolve(mockResource("ghcr.io"))
	require.NoError(t, err)

	cfg, err := auth.Authorization()
	require.NoError(t, err)
	assert.Equal(t, "user", cfg.Username)
	assert.Equal(t, "pass", cfg.Password)
}

func TestStaticKeychain_ResolveMatchesRegistry(t *testing.T) {
	kc := &staticKeychain{registry: "ghcr.io", username: "user", password: "pass"}

	// Matching registry returns credentials
	auth, err := kc.Resolve(mockResource("ghcr.io"))
	require.NoError(t, err)
	cfg, err := auth.Authorization()
	require.NoError(t, err)
	assert.Equal(t, "user", cfg.Username)
	assert.Equal(t, "pass", cfg.Password)

	// Nil resource returns credentials (backward compat)
	authNil, err := kc.Resolve(nil)
	require.NoError(t, err)
	cfgNil, err := authNil.Authorization()
	require.NoError(t, err)
	assert.Equal(t, "user", cfgNil.Username)
}

func TestStaticKeychain_ResolveReturnsAnonymousForDifferentRegistry(t *testing.T) {
	kc := &staticKeychain{registry: "ghcr.io", username: "user", password: "pass"}

	// Different registry returns anonymous
	auth, err := kc.Resolve(mockResource("index.docker.io"))
	require.NoError(t, err)
	assert.Equal(t, authn.Anonymous, auth)

	// Another different registry
	auth2, err := kc.Resolve(mockResource("registry.example.com"))
	require.NoError(t, err)
	assert.Equal(t, authn.Anonymous, auth2)
}

func TestStaticKeychain_ResolveEmptyRegistryMatchesAll(t *testing.T) {
	// When registry is empty, credentials are returned for any resource
	// (backward-compatible fallback)
	kc := &staticKeychain{registry: "", username: "user", password: "pass"}

	auth1, err := kc.Resolve(mockResource("ghcr.io"))
	require.NoError(t, err)
	cfg1, err := auth1.Authorization()
	require.NoError(t, err)
	assert.Equal(t, "user", cfg1.Username)

	auth2, err := kc.Resolve(mockResource("index.docker.io"))
	require.NoError(t, err)
	cfg2, err := auth2.Authorization()
	require.NoError(t, err)
	assert.Equal(t, "user", cfg2.Username)
}

// mockResource implements authn.Resource for testing purposes.
type mockResource string

func (r mockResource) RegistryStr() string { return string(r) }
func (r mockResource) String() string      { return string(r) }

func TestSetStaticCredentials_SetsMultiKeychain(t *testing.T) {
	t.Cleanup(func() { ResetKeychain() })

	SetStaticCredentials("ghcr.io", "myuser", "mypass")

	// activeKeychain should no longer be DefaultKeychain
	assert.NotEqual(t, authn.DefaultKeychain, activeKeychain)

	// Should be resolvable and return the static credentials for the target registry
	auth, err := activeKeychain.Resolve(mockResource("ghcr.io"))
	require.NoError(t, err)

	cfg, err := auth.Authorization()
	require.NoError(t, err)
	assert.Equal(t, "myuser", cfg.Username)
	assert.Equal(t, "mypass", cfg.Password)
}

func TestSetStaticCredentials_OverridesPreviousCredentials(t *testing.T) {
	t.Cleanup(func() { ResetKeychain() })

	SetStaticCredentials("ghcr.io", "first", "pass1")
	SetStaticCredentials("ghcr.io", "second", "pass2")

	auth, err := activeKeychain.Resolve(mockResource("ghcr.io"))
	require.NoError(t, err)

	cfg, err := auth.Authorization()
	require.NoError(t, err)
	assert.Equal(t, "second", cfg.Username)
	assert.Equal(t, "pass2", cfg.Password)
}

func TestActiveKeychain_DefaultIsDockerKeychain(t *testing.T) {
	// Verify the package-level default hasn't been changed by other tests
	// (relies on test isolation via ResetKeychain in other tests)
	assert.Equal(t, authn.DefaultKeychain, activeKeychain)
}

func TestActiveKeychain_ReturnsCurrentKeychain(t *testing.T) {
	t.Cleanup(func() { ResetKeychain() })

	// Before setting credentials, ActiveKeychain returns DefaultKeychain
	assert.Equal(t, authn.DefaultKeychain, ActiveKeychain())

	SetStaticCredentials("ghcr.io", "user", "pass")

	// After setting credentials, ActiveKeychain returns the MultiKeychain
	kc := ActiveKeychain()
	assert.NotNil(t, kc)
	assert.NotEqual(t, authn.DefaultKeychain, kc)
}

func TestResetKeychain_RestoresToDefault(t *testing.T) {
	SetStaticCredentials("ghcr.io", "user", "pass")
	assert.NotEqual(t, authn.DefaultKeychain, activeKeychain)

	ResetKeychain()

	assert.Equal(t, authn.DefaultKeychain, activeKeychain)
	assert.Equal(t, authn.DefaultKeychain, ActiveKeychain())
}
