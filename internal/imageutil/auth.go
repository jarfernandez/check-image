package imageutil

import "github.com/google/go-containerregistry/pkg/authn"

// activeKeychain is the keychain used for remote registry authentication.
// It defaults to DefaultKeychain (Docker config, credential helpers) and can
// be overridden with explicit credentials via SetStaticCredentials.
var activeKeychain authn.Keychain = authn.DefaultKeychain

// staticKeychain provides fixed credentials scoped to a specific registry.
// When registry is non-empty, credentials are only returned for requests
// targeting that registry; other registries receive anonymous authentication.
type staticKeychain struct {
	registry string
	username string
	password string
}

func (k *staticKeychain) Resolve(r authn.Resource) (authn.Authenticator, error) {
	// When a target registry is configured, only return credentials for
	// requests to that specific host. This prevents credentials from being
	// sent to unrelated registries during cross-registry pulls.
	if k.registry != "" && r != nil && r.RegistryStr() != k.registry {
		return authn.Anonymous, nil
	}
	return authn.FromConfig(authn.AuthConfig{
		Username: k.username,
		Password: k.password,
	}), nil
}

// SetStaticCredentials configures explicit credentials that take priority over
// the default keychain (Docker config, credential helpers). Credentials are
// scoped to the given registry hostname; requests to other registries fall
// through to the default keychain. If registry is empty, credentials are
// applied to all registries (backward-compatible fallback).
func SetStaticCredentials(registry, username, password string) {
	activeKeychain = authn.NewMultiKeychain(
		&staticKeychain{registry: registry, username: username, password: password},
		authn.DefaultKeychain,
	)
}

// ActiveKeychain returns the currently configured keychain.
// Primarily useful for testing and diagnostic purposes.
func ActiveKeychain() authn.Keychain {
	return activeKeychain
}

// ResetKeychain resets the active keychain back to the default (Docker config
// and credential helpers). Use this to clear credentials previously set with
// SetStaticCredentials.
func ResetKeychain() {
	activeKeychain = authn.DefaultKeychain
}
