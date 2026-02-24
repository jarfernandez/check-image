package commands

// Default values for CLI flags shared between individual check commands and the all command.
const (
	defaultMaxAgeDays    uint = 90
	defaultMaxSizeMB     uint = 500
	defaultMaxLayerCount uint = 20
)

// imageArgFormatsDoc is the standard help paragraph describing the image argument
// transport formats. It is appended to every command's Long description so that
// adding a new transport only requires updating a single place.
const imageArgFormatsDoc = `The 'image' argument supports multiple formats:
  - Registry image (daemon with registry fallback): image:tag, registry/namespace/image:tag
  - OCI layout directory: oci:/path/to/layout:tag or oci:/path/to/layout@sha256:digest
  - OCI tarball: oci-archive:/path/to/image.tar:tag
  - Docker tarball: docker-archive:/path/to/image.tar:tag`
