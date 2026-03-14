package imageutil

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	cr "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	log "github.com/sirupsen/logrus"
)

const (
	maxRetries    = 3
	retryBaseWait = 1 * time.Second
)

// remoteTransport is the HTTP transport used for remote registry calls.
// It applies timeouts to prevent hanging on unresponsive registries.
var remoteTransport http.RoundTripper = &http.Transport{
	Proxy:                 http.ProxyFromEnvironment,
	DialContext:           (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
	TLSHandshakeTimeout:   15 * time.Second,
	ResponseHeaderTimeout: 30 * time.Second,
	ForceAttemptHTTP2:     true,
}

// getLocalImageFn and getRemoteImageFn are package-level variables used for
// retrieving images from the daemon and remote registry respectively.
// They can be overridden in tests to avoid daemon and network access.
var getLocalImageFn = GetLocalImage
var getRemoteImageFn = GetRemoteImage

// daemonImageFn wraps daemon.Image and can be overridden in tests to avoid
// requiring a live Docker daemon socket.
var daemonImageFn = daemon.Image

// GetImageRegistry extracts the registry from the image name
func GetImageRegistry(imageName string) (string, error) {
	ref, err := ParseReference(imageName)
	if err != nil {
		return "", err
	}

	// Non-registry transports don't have a registry
	if ref.Transport != TransportDaemonRegistry {
		return "", fmt.Errorf("registry not applicable for %s transport", ref.Transport)
	}

	// Parse as regular image reference
	parsedRef, err := name.ParseReference(ref.Path)
	if err != nil {
		return "", fmt.Errorf("error parsing the reference: %w", err)
	}

	return parsedRef.Context().RegistryStr(), nil
}

// GetLocalImage retrieves the local image from a reference name
func GetLocalImage(ctx context.Context, imageName string) (cr.Image, error) {
	ref, err := name.ParseReference(imageName)
	if err != nil {
		return nil, fmt.Errorf("error parsing the reference: %w", err)
	}

	image, err := daemonImageFn(ref, daemon.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("error retrieving the local image: %w", err)
	}

	return image, nil
}

// GetRemoteImage retrieves the remote image from a reference name.
// Transient errors (network timeouts, HTTP 429/5xx) are retried up to
// maxRetries times with exponential backoff.
func GetRemoteImage(ctx context.Context, imageName string) (cr.Image, error) {
	ref, err := name.ParseReference(imageName)
	if err != nil {
		return nil, fmt.Errorf("error parsing the reference: %w", err)
	}

	img, err := retryWithBackoff(ctx, maxRetries, retryBaseWait, func() (cr.Image, error) {
		return remote.Image(ref,
			remote.WithAuthFromKeychain(activeKeychain),
			remote.WithTransport(remoteTransport),
			remote.WithContext(ctx))
	})
	if err != nil {
		return nil, fmt.Errorf("error retrieving the remote image: %w", err)
	}
	return img, nil
}

// retryWithBackoff calls fn up to attempts+1 times, backing off exponentially
// between failures. Returns immediately on success or non-retryable error
// (unwrapped, so the caller can re-wrap with context). After exhausting all
// attempts returns an error of the form "after N attempts: <lastErr>".
// Context cancellation during a backoff sleep terminates the loop immediately.
func retryWithBackoff(ctx context.Context, attempts int, baseWait time.Duration, fn func() (cr.Image, error)) (cr.Image, error) {
	var lastErr error
	for attempt := 0; attempt <= attempts; attempt++ {
		img, err := fn()
		if err == nil {
			return img, nil
		}
		lastErr = err

		if !isRetryableError(err) {
			return nil, err
		}

		if attempt < attempts {
			backoff := baseWait * (1 << uint(attempt))
			log.WithFields(log.Fields{"attempt": attempt + 1, "max": attempts + 1, "error": err, "retry_in": backoff.String()}).Debug("Retrying after transient error")
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	return nil, fmt.Errorf("after %d attempts: %w", attempts+1, lastErr)
}

// isRetryableError returns true for transient network errors and HTTP 429/5xx
// status codes that are safe to retry.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Network-level errors (timeouts, DNS failures, connection resets)
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// HTTP status codes that indicate transient server issues.
	// go-containerregistry returns *transport.Error for HTTP responses, which
	// carries the status code as a typed integer field. Using a type assertion
	// is precise and version-stable — substring matching on err.Error() is
	// fragile because unrelated error messages can accidentally contain "500"
	// (e.g. digests like sha256:5001a... or port numbers like 5004).
	var tErr *transport.Error
	if errors.As(err, &tErr) {
		switch tErr.StatusCode {
		case 429, 500, 502, 503, 504:
			return true
		}
	}

	return false
}

// GetDockerArchiveImage retrieves an image from a Docker tarball (docker save format).
// A non-empty tag is required; use the format docker-archive:/path.tar:tag.
func GetDockerArchiveImage(tarballPath string, tag string) (cr.Image, error) {
	if tag == "" {
		return nil, fmt.Errorf("docker-archive transport requires a tag (e.g., docker-archive:/path.tar:tag)")
	}

	parsedTag, err := name.NewTag(tag)
	if err != nil {
		return nil, fmt.Errorf("error parsing tag %s: %w", tag, err)
	}

	// Load image from tarball
	image, err := tarball.ImageFromPath(tarballPath, &parsedTag)
	if err != nil {
		return nil, fmt.Errorf("error loading docker archive from %s: %w", tarballPath, err)
	}

	return image, nil
}

// GetOCIArchiveImage retrieves an image from an OCI tarball.
// The caller must call the returned cleanup function when done with the image
// to remove the temporary directory created during extraction.
func GetOCIArchiveImage(tarballPath string, reference string) (cr.Image, func(), error) {
	// OCI archives need to be extracted to a temporary directory first
	// then loaded using the OCI layout functions.
	// v1.Image is lazy, so the temp dir must remain on disk until the caller
	// is done accessing the image; cleanup is the caller's responsibility.
	tempDir, err := extractOCIArchive(tarballPath)
	if err != nil {
		return nil, func() {}, fmt.Errorf("error extracting OCI archive: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(tempDir) }

	img, err := GetOCILayoutImage(tempDir, reference)
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}
	return img, cleanup, nil
}

// GetImage retrieves the image using transport-aware reference parsing.
// The caller must call the returned cleanup function when done with the image.
// For all transports except oci-archive, cleanup does nothing.
func GetImage(ctx context.Context, imageName string) (cr.Image, func(), error) {
	ref, err := ParseReference(imageName)
	if err != nil {
		return nil, func() {}, err
	}

	switch ref.Transport {
	case TransportOCI:
		// OCI layout directory - no fallback
		reference := ref.Digest
		if reference == "" {
			reference = ref.Tag
		}
		if reference == "" {
			return nil, func() {}, fmt.Errorf("oci transport requires tag or digest")
		}
		img, err := GetOCILayoutImage(ref.Path, reference)
		if err != nil {
			return nil, func() {}, err
		}
		return img, func() {}, nil

	case TransportOCIArchive:
		// OCI tarball - extract and load; cleanup removes the temp dir
		reference := ref.Digest
		if reference == "" {
			reference = ref.Tag
		}
		if reference == "" {
			return nil, func() {}, fmt.Errorf("oci-archive transport requires tag or digest")
		}
		return GetOCIArchiveImage(ref.Path, reference)

	case TransportDockerArchive:
		// Docker tarball - load directly.
		img, err := GetDockerArchiveImage(ref.Path, ref.Tag)
		if err != nil {
			return nil, func() {}, err
		}
		return img, func() {}, nil

	case TransportDaemonRegistry:
		// Default mode: try local daemon, fall back to remote registry
		image, err := getLocalImageFn(ctx, ref.Path)
		if err == nil {
			return image, func() {}, nil
		}
		// Honour context cancellation: do not attempt the remote fallback
		// if the context was cancelled while the daemon call was in progress.
		if ctx.Err() != nil {
			return nil, func() {}, ctx.Err()
		}
		image, err = getRemoteImageFn(ctx, ref.Path)
		if err != nil {
			return nil, func() {}, err
		}
		return image, func() {}, nil

	default:
		return nil, func() {}, fmt.Errorf("unsupported transport: %s", ref.Transport)
	}
}

// GetImageConfig retrieves the configuration file of a given container image
func GetImageConfig(image cr.Image) (*cr.ConfigFile, error) {
	config, err := image.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("error retrieving the image configuration: %w", err)
	}

	return config, nil
}

// GetImageAndConfig retrieves both the image and its configuration file given an image name.
// The caller must call the returned cleanup function when done with the image.
// For all transports except oci-archive, cleanup does nothing.
func GetImageAndConfig(ctx context.Context, imageName string) (cr.Image, *cr.ConfigFile, func(), error) {
	image, cleanup, err := GetImage(ctx, imageName)
	if err != nil {
		return nil, nil, func() {}, err
	}

	config, err := GetImageConfig(image)
	if err != nil {
		cleanup()
		return nil, nil, func() {}, err
	}

	return image, config, cleanup, nil
}
