package imageutil

import "github.com/google/go-containerregistry/pkg/authn"

// activeKeychain is the keychain used for remote registry authentication.
// It defaults to DefaultKeychain (Docker config, credential helpers) and can
// be overridden with explicit credentials via SetStaticCredentials.
var activeKeychain authn.Keychain = authn.DefaultKeychain

// staticKeychain provides fixed credentials for all registry requests.
type staticKeychain struct {
	username string
	password string
}

func (k *staticKeychain) Resolve(_ authn.Resource) (authn.Authenticator, error) {
	return authn.FromConfig(authn.AuthConfig{
		Username: k.username,
		Password: k.password,
	}), nil
}

// SetStaticCredentials configures explicit credentials that take priority over
// the default keychain (Docker config, credential helpers). The same credentials
// are applied to all registries.
func SetStaticCredentials(username, password string) {
	activeKeychain = authn.NewMultiKeychain(
		&staticKeychain{username: username, password: password},
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
