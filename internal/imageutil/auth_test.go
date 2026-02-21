package imageutil

import (
	"testing"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticKeychain_Resolve(t *testing.T) {
	kc := &staticKeychain{username: "user", password: "pass"}

	auth, err := kc.Resolve(nil)
	require.NoError(t, err)

	cfg, err := auth.Authorization()
	require.NoError(t, err)
	assert.Equal(t, "user", cfg.Username)
	assert.Equal(t, "pass", cfg.Password)
}

func TestStaticKeychain_ResolveIgnoresResource(t *testing.T) {
	// staticKeychain applies the same credentials regardless of which registry
	// is being accessed — it resolves the same for any resource value
	kc := &staticKeychain{username: "user", password: "pass"}

	authForNil, err := kc.Resolve(nil)
	require.NoError(t, err)
	cfgNil, err := authForNil.Authorization()
	require.NoError(t, err)

	// Resolve with a non-nil resource (a different registry) — credentials must be the same
	dockerHub := mockResource("index.docker.io")
	authForHub, err := kc.Resolve(dockerHub)
	require.NoError(t, err)
	cfgHub, err := authForHub.Authorization()
	require.NoError(t, err)

	ghcr := mockResource("ghcr.io")
	authForGHCR, err := kc.Resolve(ghcr)
	require.NoError(t, err)
	cfgGHCR, err := authForGHCR.Authorization()
	require.NoError(t, err)

	// Same credentials regardless of resource
	assert.Equal(t, cfgNil.Username, cfgHub.Username)
	assert.Equal(t, cfgNil.Password, cfgHub.Password)
	assert.Equal(t, cfgNil.Username, cfgGHCR.Username)
	assert.Equal(t, cfgNil.Password, cfgGHCR.Password)
}

// mockResource implements authn.Resource for testing purposes.
type mockResource string

func (r mockResource) RegistryStr() string { return string(r) }
func (r mockResource) String() string      { return string(r) }

func TestSetStaticCredentials_SetsMultiKeychain(t *testing.T) {
	t.Cleanup(func() { ResetKeychain() })

	SetStaticCredentials("myuser", "mypass")

	// activeKeychain should no longer be DefaultKeychain
	assert.NotEqual(t, authn.DefaultKeychain, activeKeychain)

	// Should be resolvable and return the static credentials
	auth, err := activeKeychain.Resolve(nil)
	require.NoError(t, err)

	cfg, err := auth.Authorization()
	require.NoError(t, err)
	assert.Equal(t, "myuser", cfg.Username)
	assert.Equal(t, "mypass", cfg.Password)
}

func TestSetStaticCredentials_OverridesPreviousCredentials(t *testing.T) {
	t.Cleanup(func() { ResetKeychain() })

	SetStaticCredentials("first", "pass1")
	SetStaticCredentials("second", "pass2")

	auth, err := activeKeychain.Resolve(nil)
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

	SetStaticCredentials("user", "pass")

	// After setting credentials, ActiveKeychain returns the MultiKeychain
	kc := ActiveKeychain()
	assert.NotNil(t, kc)
	assert.NotEqual(t, authn.DefaultKeychain, kc)
}

func TestResetKeychain_RestoresToDefault(t *testing.T) {
	SetStaticCredentials("user", "pass")
	assert.NotEqual(t, authn.DefaultKeychain, activeKeychain)

	ResetKeychain()

	assert.Equal(t, authn.DefaultKeychain, activeKeychain)
	assert.Equal(t, authn.DefaultKeychain, ActiveKeychain())
}
