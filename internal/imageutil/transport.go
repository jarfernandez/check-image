package imageutil

import (
	"fmt"
	"strings"
)

// Transport represents the method of accessing a container image
type Transport string

const (
	// TransportDaemonRegistry represents the default behavior: try local daemon first, then fall back to remote registry
	TransportDaemonRegistry Transport = "daemon-registry"
	// TransportOCI represents accessing images from OCI layout directory
	TransportOCI Transport = "oci"
	// TransportOCIArchive represents accessing images from OCI tarball (future)
	TransportOCIArchive Transport = "oci-archive"
	// TransportDockerArchive represents accessing images from Docker tarball (future)
	TransportDockerArchive Transport = "docker-archive"
)

// ImageReference represents a parsed container image reference with transport information
type ImageReference struct {
	Transport Transport
	Path      string // For file/dir transports, or full reference for daemon/registry
	Tag       string // Optional tag
	Digest    string // Optional digest (mutually exclusive with Tag)
}

// ParseReference parses an image reference with optional transport prefix
// Examples:
//
//	nginx:latest                    → daemon/registry, "nginx:latest", "", ""
//	nginx                           → daemon/registry, "nginx:latest", "", "" (latest added by default)
//	oci:/path/to/layout:v1          → oci, "/path/to/layout", "v1", ""
//	oci:/path/to/layout@sha256:abc  → oci, "/path/to/layout", "", "sha256:abc"
//	oci-archive:./image.tar:latest  → oci-archive, "./image.tar", "latest", ""
func ParseReference(ref string) (*ImageReference, error) {
	// If no colon or @ is present, add :latest for daemon/registry references
	if !strings.Contains(ref, ":") && !strings.Contains(ref, "@") {
		ref += ":latest"
	}

	// Check for transport prefix
	if !strings.Contains(ref, ":") {
		return nil, fmt.Errorf("invalid reference format: %s", ref)
	}

	parts := strings.SplitN(ref, ":", 2)

	// Check if first part is a transport
	transport := Transport(parts[0])
	switch transport {
	case TransportOCI, TransportOCIArchive, TransportDockerArchive:
		// Has transport prefix - parse path and tag/digest
		remainder := parts[1]
		return parseTransportReference(transport, remainder)
	default:
		// No transport prefix - default to daemon/registry fallback
		return &ImageReference{
			Transport: TransportDaemonRegistry,
			Path:      ref,
		}, nil
	}
}

func parseTransportReference(transport Transport, remainder string) (*ImageReference, error) {
	// Parse path and tag/digest
	// Format: path:tag or path@digest

	// Check for digest first (contains @)
	if strings.Contains(remainder, "@") {
		parts := strings.SplitN(remainder, "@", 2)
		return &ImageReference{
			Transport: transport,
			Path:      parts[0],
			Digest:    parts[1],
		}, nil
	}

	// For tags, we need to find where the path ends and tag begins
	// Strategy: Find the first colon that's not part of a Windows drive letter (C:)
	// This assumes file paths don't contain colons (which is reasonable)
	firstColon := findPathTagSeparator(remainder)
	if firstColon == -1 {
		// No tag or digest
		return &ImageReference{
			Transport: transport,
			Path:      remainder,
		}, nil
	}

	return &ImageReference{
		Transport: transport,
		Path:      remainder[:firstColon],
		Tag:       remainder[firstColon+1:],
	}, nil
}

// findPathTagSeparator finds the colon that separates the file path from the tag
// It skips Windows drive letters (e.g., C:) and returns the index of the separator
func findPathTagSeparator(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			// Check if this is a Windows drive letter (single char followed by :)
			if i == 1 && len(s) > 2 {
				// Could be C:\path - skip this colon
				continue
			}
			// Found the separator
			return i
		}
	}
	return -1
}
