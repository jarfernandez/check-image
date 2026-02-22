# Check Image

Check Image is a Go-based CLI tool designed for validating container images. It ensures that images meet specific standards such as size, age, ports, and security configurations. This project follows Go conventions for command-line tools and is structured into `cmd` and `internal` directories.

## Table of Contents

- [Installation](#installation)
- [Docker](#docker)
- [GitHub Action](#github-action)
- [Usage](#usage)
- [Commands](#commands)
- [Configuration Files](#configuration-files)
- [Development](#development)
- [Testing](#testing)
- [CI/CD and Release Process](#cicd-and-release-process)
- [Contributing](#contributing)
- [License](#license)

## Installation

### Install from Pre-built Binaries (Recommended)

Download the latest release for your platform from the [releases page](https://github.com/jarfernandez/check-image/releases):

```bash
# Set the version you want to install (or use 'latest' tag from releases page)
VERSION=0.17.1 # x-release-please-version

# macOS (Apple Silicon)
curl -sL "https://github.com/jarfernandez/check-image/releases/download/v${VERSION}/check-image_${VERSION}_darwin_arm64.tar.gz" | tar xz
sudo mv check-image /usr/local/bin/

# macOS (Intel)
curl -sL "https://github.com/jarfernandez/check-image/releases/download/v${VERSION}/check-image_${VERSION}_darwin_amd64.tar.gz" | tar xz
sudo mv check-image /usr/local/bin/

# Linux (amd64)
curl -sL "https://github.com/jarfernandez/check-image/releases/download/v${VERSION}/check-image_${VERSION}_linux_amd64.tar.gz" | tar xz
sudo mv check-image /usr/local/bin/

# Linux (arm64)
curl -sL "https://github.com/jarfernandez/check-image/releases/download/v${VERSION}/check-image_${VERSION}_linux_arm64.tar.gz" | tar xz
sudo mv check-image /usr/local/bin/

# Windows (amd64)
# Download https://github.com/jarfernandez/check-image/releases/download/v${VERSION}/check-image_${VERSION}_windows_amd64.zip
# and extract to a directory in your PATH
```

Pre-built binaries include the correct version number (e.g., `check-image version --short` returns `v0.17.1`). <!-- x-release-please-version -->

### Install with Homebrew

**Requirements:** macOS or Linux with [Homebrew](https://brew.sh) installed.

```bash
brew tap jarfernandez/tap
brew install check-image
```

To upgrade to a new version:

```bash
brew upgrade check-image
```

### Install with Go

**Requirements:** Go 1.26 or newer

```bash
# Install the latest version
go install github.com/jarfernandez/check-image/cmd/check-image@latest

# Or install a specific version
go install github.com/jarfernandez/check-image/cmd/check-image@v0.17.1 # x-release-please-version
```

This will install the `check-image` binary to your `GOBIN` directory.

**Note:** Binaries installed with `go install` will show version as `dev` when running `check-image version`. This is expected behavior as `go install` compiles from source without version injection. For production use with correct version numbers, use pre-built binaries from releases.

### Install from Source

If you've cloned the repository, you can install it locally:

```bash
go install ./cmd/check-image
```

This is useful for development. The version will show as `dev`.

### Docker

Check Image is available as a multi-arch Docker image (linux/amd64, linux/arm64) from GitHub Container Registry:

```bash
docker pull ghcr.io/jarfernandez/check-image:latest
```

**Basic usage (validates remote registry images):**

```bash
# Check image age
docker run --rm ghcr.io/jarfernandez/check-image age nginx:latest --max-age 30

# Check image size
docker run --rm ghcr.io/jarfernandez/check-image size nginx:latest --max-size 100

# Check for root user
docker run --rm ghcr.io/jarfernandez/check-image root-user nginx:latest

# Run all checks with JSON output
docker run --rm ghcr.io/jarfernandez/check-image all nginx:latest -o json
```

**Using policy files via volume mounts:**

```bash
# Mount a local config directory
docker run --rm \
  -v "$(pwd)/config:/config:ro" \
  ghcr.io/jarfernandez/check-image registry nginx:latest \
  --registry-policy /config/registry-policy.json

# Run all checks with a config file
docker run --rm \
  -v "$(pwd)/config:/config:ro" \
  ghcr.io/jarfernandez/check-image all nginx:latest \
  --config /config/config.yaml
```

**Using with Docker socket (advanced):**

> **Warning:** Mounting the Docker socket grants the container full access to the Docker daemon, which is equivalent to root access on the host. Only use this in trusted environments.

```bash
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/jarfernandez/check-image age my-local-image:latest
```

Without the Docker socket mounted (the default), check-image automatically uses the remote registry to fetch image metadata. This is the recommended approach for CI/CD pipelines.

**Using a specific version:**

```bash
docker pull ghcr.io/jarfernandez/check-image:0.17.1 # x-release-please-version
```

## GitHub Action

Check Image is available as a GitHub Action for validating container images directly in your CI/CD workflows.

### Basic Usage

```yaml
- uses: jarfernandez/check-image@v0.17.1 # x-release-please-version
  with:
    image: nginx:latest
```

This runs all 10 checks with default settings. Checks that require additional configuration (registry, labels, platform) will report an error unless their configuration is provided.

### With a Config File

```yaml
- uses: jarfernandez/check-image@v0.17.1 # x-release-please-version
  with:
    image: myorg/myapp:${{ github.sha }}
    config: .check-image/config.yaml
```

The config file determines which checks to run and their parameters. See [All Checks Configuration Files](#all-checks-configuration-files) for the format.

### Running Specific Checks

```yaml
- uses: jarfernandez/check-image@v0.17.1 # x-release-please-version
  with:
    image: nginx:latest
    checks: age,size,root-user
    max-age: '30'
    max-size: '200'
```

### With Policy Files

```yaml
- uses: jarfernandez/check-image@v0.17.1 # x-release-please-version
  with:
    image: ghcr.io/myorg/app:latest
    registry-policy: policies/registry-policy.yaml
    labels-policy: policies/labels-policy.json
    skip: healthcheck
```

### Soft Failure

Use `continue-on-error` to prevent the action from failing the workflow:

```yaml
- uses: jarfernandez/check-image@v0.17.1 # x-release-please-version
  id: check
  continue-on-error: true
  with:
    image: nginx:latest
    config: .check-image/config.yaml

- name: Handle results
  if: steps.check.outputs.result == 'failed'
  run: echo "Image validation failed but continuing"
```

### Using JSON Output

The action captures full JSON output for programmatic use in subsequent steps:

```yaml
- uses: jarfernandez/check-image@v0.17.1 # x-release-please-version
  id: check
  continue-on-error: true
  with:
    image: nginx:latest
    checks: age,size

- name: Process results
  if: always()
  run: echo '${{ steps.check.outputs.json }}' | jq '.summary'
```

### Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `image` | Yes | - | Container image to validate |
| `config` | No | - | Path to config file for the `all` command |
| `checks` | No | - | Comma-separated list of checks to run (mutually exclusive with `skip`) |
| `skip` | No | - | Comma-separated list of checks to skip (mutually exclusive with `checks`) |
| `fail-fast` | No | `false` | Stop on first check failure |
| `max-age` | No | - | Maximum image age in days |
| `max-size` | No | - | Maximum image size in MB |
| `max-layers` | No | - | Maximum number of layers |
| `allowed-ports` | No | - | Comma-separated allowed ports or `@file` path |
| `allowed-platforms` | No | - | Comma-separated allowed platforms or `@file` path |
| `registry-policy` | No | - | Path to registry policy file |
| `labels-policy` | No | - | Path to labels policy file |
| `secrets-policy` | No | - | Path to secrets policy file |
| `skip-env-vars` | No | `false` | Skip environment variable checks |
| `skip-files` | No | `false` | Skip file system checks |
| `allow-shell-form` | No | `false` | Allow shell form for entrypoint or cmd |
| `log-level` | No | `info` | Log level |
| `version` | No | `0.17.1` <!-- x-release-please-version --> | check-image version to use |

### Outputs

| Output | Description |
|--------|-------------|
| `result` | Validation result: `passed`, `failed`, or `error` |
| `json` | Full JSON output from check-image |

The action also generates a **Step Summary** visible in the GitHub Actions UI with a results table, details of failed checks, and full JSON output.

### Requirements

The action downloads the check-image binary from GitHub Releases, so no additional dependencies are needed for validating remote registry images. To validate local Docker images (e.g., after `docker build`), Docker must be available on the runner — this is satisfied by default on `ubuntu-latest` runners.

## Usage

After installation, you can run the CLI tool:

```bash
check-image --help
```

### Image Reference Syntax

Check Image supports multiple image sources using a transport-based syntax (compatible with Skopeo):

**OCI Layout Directory** (with tag):
```bash
check-image age oci:/path/to/layout:latest
check-image size oci:./nginx-layout:v1.23 --max-size 100
```

**OCI Layout Directory** (with digest):
```bash
check-image age oci:/path/to/layout@sha256:abc123...
```

**OCI Archive** (tarball):
```bash
check-image age oci-archive:/path/to/image.tar:latest
check-image size oci-archive:./exported-image.tar:1.0 --max-size 100
```

**Docker Archive** (saved with `docker save`):
```bash
check-image age docker-archive:/path/to/saved.tar:nginx:latest
check-image size docker-archive:./backup.tar:myapp:2.0
```

**Default Behavior** (Docker daemon or remote registry):
```bash
check-image age nginx:latest
check-image size docker.io/nginx:latest --max-size 100

# Tag defaults to 'latest' if not specified
check-image age nginx
check-image registry ghcr.io/kubernetes-sigs/kind/node --registry-policy config/registry-policy.json
```

**Creating Archive Files:**

To create OCI archives:
```bash
# Using skopeo
skopeo copy docker://nginx:latest oci-archive:nginx.tar:latest

# Using podman
podman save --format oci-archive -o nginx.tar nginx:latest
```

To create Docker archives:
```bash
# Using docker
docker save -o nginx.tar nginx:latest

# Using podman (Docker-compatible format)
podman save --format docker-archive -o nginx.tar nginx:latest
```

**Important Notes:**
- When using explicit transport prefixes (`oci:`, `oci-archive:`, `docker-archive:`), only that source is attempted (no fallback)
- Without a transport prefix, Check Image tries the local Docker daemon first, then falls back to remote registry
- The `registry` command validation is skipped for non-registry transports (e.g., `oci:`)
- OCI archives are extracted to a temporary directory during processing (automatically cleaned up)
- Archive extraction includes security checks: path traversal protection and 5GB decompression limit

## Commands

The CLI supports various commands for validating container images. Each command is defined in the `cmd/check-image/commands` directory.

### Available Commands

#### `size`
Validates that the image size and layer count are within acceptable limits.

```bash
check-image size <image> --max-size <MB> --max-layers <count>
```

Options:
- `--max-size`: Maximum image size in MB (default: 500)
- `--max-layers`: Maximum number of layers (default: 20)

#### `age`
Validates that the image is not older than a specified number of days.

```bash
check-image age <image> --max-age <days>
```

Options:
- `--max-age`: Maximum image age in days (default: 90)

#### `registry`
Validates that the image registry is trusted based on a policy file.

```bash
check-image registry <image> --registry-policy <file>
```

Options:
- `--registry-policy`: Path to registry policy file (JSON or YAML, required)

Policy file supports either:
- `trusted-registries`: Allowlist of trusted registries
- `excluded-registries`: Blocklist of excluded registries

#### `ports`
Validates that the image does not expose unauthorized ports.

```bash
check-image ports <image> --allowed-ports <ports>
```

Options:
- `--allowed-ports`: Comma-separated list of allowed ports or `@<file>` with JSON/YAML array

#### `root-user`
Validates that the image runs as non-root user.

```bash
check-image root-user <image>
```

#### `healthcheck`
Validates that the image has a healthcheck defined.

```bash
check-image healthcheck <image>
```

The command checks that:
- A healthcheck section exists in the image configuration
- The healthcheck test command is not empty
- The healthcheck is not explicitly disabled (`NONE`)

#### `secrets`
Validates that the image does not contain sensitive data (passwords, tokens, keys).

```bash
check-image secrets <image> [--secrets-policy <file>] [--skip-env-vars] [--skip-files]
```

Options:
- `--secrets-policy`: Path to secrets policy file (JSON or YAML, optional)
- `--skip-env-vars`: Skip environment variable checks
- `--skip-files`: Skip file system checks

The command scans:
- Environment variables for sensitive patterns (password, secret, token, key, etc.)
- Files across all image layers for common secret files (SSH keys, cloud credentials, password files, etc.)

#### `entrypoint`
Validates that the image has a startup command defined (ENTRYPOINT or CMD) and uses exec form.

```bash
check-image entrypoint <image> [--allow-shell-form]
```

Options:
- `--allow-shell-form`: Allow shell form without failing (default: exec form required)

The command checks that:
- At least one of ENTRYPOINT or CMD is defined in the image configuration
- Neither uses shell form (`/bin/sh -c ...` or `/bin/bash -c ...`) unless `--allow-shell-form` is set

Exec form (`["nginx", "-g", "daemon off;"]`) is preferred over shell form because it avoids an intermediate shell process, ensures signals (e.g. SIGTERM) reach the real process directly, and eliminates unintended shell interpretation.

When `--allow-shell-form` is set and shell form is detected, the check passes and the result details include `"shell-form-allowed": true` for transparency.

#### `labels`
Validates that the image has required labels (OCI annotations) with correct values.

```bash
check-image labels <image> --labels-policy <file>
```

Options:
- `--labels-policy`: Path to labels policy file (JSON or YAML, required)

The command validates three types of requirements:
- **Existence check**: Label must be present with any value (only `name` specified)
- **Exact value match**: Label value must exactly match the specified string (`name` and `value`)
- **Pattern match**: Label value must match the regular expression (`name` and `pattern`)

#### `platform`
Validates that the image platform (OS and architecture) is in the allowed list.

```bash
check-image platform <image> --allowed-platforms <platforms>
```

Options:
- `--allowed-platforms`: Comma-separated list of allowed platforms or `@<file>` with JSON/YAML array (required)

The platform string is constructed as `OS/Architecture` (e.g., `linux/amd64`) or `OS/Architecture/Variant` for architectures with variants (e.g., `linux/arm/v7`). The check validates the resolved image's platform — the platform of the concrete image that would actually be executed, not a manifest index listing.

#### `all`
Runs all validation checks on a container image at once.

```bash
check-image all <image> [flags]
```

Options:
- `--config`, `-c`: Path to configuration file (JSON or YAML)
- `--include`: Comma-separated list of checks to run (age, size, ports, registry, root-user, healthcheck, secrets, labels, entrypoint, platform)
- `--skip`: Comma-separated list of checks to skip (age, size, ports, registry, root-user, healthcheck, secrets, labels, entrypoint, platform)
- `--max-age`, `-a`: Maximum age in days (default: 90)
- `--max-size`, `-m`: Maximum size in MB (default: 500)
- `--max-layers`, `-y`: Maximum number of layers (default: 20)
- `--allowed-ports`, `-p`: Comma-separated list of allowed ports or `@<file>`
- `--allowed-platforms`: Comma-separated list of allowed platforms or `@<file>`
- `--registry-policy`, `-r`: Registry policy file (JSON or YAML)
- `--labels-policy`: Labels policy file (JSON or YAML)
- `--secrets-policy`, `-s`: Secrets policy file (JSON or YAML)
- `--skip-env-vars`: Skip environment variable checks in secrets detection
- `--skip-files`: Skip file system checks in secrets detection
- `--allow-shell-form`: Allow shell form for entrypoint or cmd
- `--fail-fast`: Stop on first check failure (default: false)

Note: `--include` and `--skip` are mutually exclusive.

Precedence rules:
1. Without `--config`: all 10 checks run with defaults, except those in `--skip`
2. With `--config`: only checks present in the config file run, except those in `--skip`
3. `--include` overrides config file check selection (runs only specified checks)
4. CLI flags override config file values
5. `--include` and `--skip` always take precedence over the config file

#### `version`
Shows the check-image version with full build information.

```bash
check-image version
```

```
check-image version v0.12.1
commit:     a1b2c3d
built at:   2026-02-18T12:34:56Z
go version: go1.26.0
platform:   linux/amd64
```

Options:
- `--short`: Print only the version number

```bash
check-image version --short
```

```
v0.12.1
```

The version and build metadata are injected at build time using ldflags:
```bash
go build \
  -ldflags "-X github.com/jarfernandez/check-image/internal/version.Version=v0.1.0 \
            -X github.com/jarfernandez/check-image/internal/version.Commit=abc1234 \
            -X github.com/jarfernandez/check-image/internal/version.BuildDate=2026-02-18T12:34:56Z" \
  ./cmd/check-image
```

The Go version and platform are read from the Go runtime and do not require ldflags injection.

### Global Flags

All commands support:
- `--output`, `-o`: Output format: `text` (default), `json`
- `--log-level`: Set log level (trace, debug, info, warn, error, fatal, panic)
- `--username`: Registry username for authentication (env: `CHECK_IMAGE_USERNAME`)
- `--password`: Registry password or token (env: `CHECK_IMAGE_PASSWORD`). Caution: visible in process list — prefer `--password-stdin` or the env var.
- `--password-stdin`: Read the registry password from stdin. Cannot be combined with other flags that also read from stdin (`--config -`, `--allowed-ports @-`, etc.)

### Private Registry Authentication

Check Image supports three ways to provide credentials for private registries, applied with the following precedence:

1. **CLI flags** (`--username` / `--password` / `--password-stdin`) — highest priority
2. **Environment variables** (`CHECK_IMAGE_USERNAME` / `CHECK_IMAGE_PASSWORD`)
3. **`~/.docker/config.json`** and credential helpers (`authn.DefaultKeychain`) — already works without any changes

The same credentials are applied to all registry requests in that invocation. For per-registry credentials, configure Docker's credential helpers in `~/.docker/config.json` as usual.

**Using flags:**
```bash
check-image size my-registry.example.com/private-image:latest \
  --username myuser \
  --password mypassword
```

**Using `--password-stdin` (recommended — avoids password in process list):**
```bash
echo "mytoken" | check-image age my-registry.example.com/private-image:latest \
  --username myuser \
  --password-stdin

# Or read from a file
check-image root-user my-registry.example.com/private-image:latest \
  --username myuser \
  --password-stdin < ~/.secrets/registry-token
```

**Using environment variables (recommended for CI/CD):**
```bash
export CHECK_IMAGE_USERNAME=myuser
export CHECK_IMAGE_PASSWORD=mypassword
check-image healthcheck my-registry.example.com/private-image:latest
```

**Using `~/.docker/config.json` (already works — no flags needed):**
```bash
docker login my-registry.example.com
check-image secrets my-registry.example.com/private-image:latest
```

**Important notes:**
- `--password` and `--password-stdin` are mutually exclusive
- Specifying `--username` without a password (or vice versa) is an error
- `--password-stdin` reads the entire stdin input and strips trailing `\r\n` — it cannot be combined with other flags that also consume stdin (`--config -`, `--allowed-ports @-`, etc.)
- For GHCR (GitHub Container Registry) with a Personal Access Token, use `--username <github-username> --password-stdin` and pipe the PAT

### JSON Output

All commands support JSON output with `--output json` (or `-o json`). This is useful for scripting and CI/CD pipelines.

**Individual command:**
```bash
check-image age nginx:latest -o json
```
```json
{
  "check": "age",
  "image": "nginx:latest",
  "passed": true,
  "message": "Image is less than 90 days old",
  "details": {
    "created-at": "2025-12-01T00:00:00Z",
    "age-days": 75,
    "max-age": 90
  }
}
```

**All command:**
```bash
check-image all nginx:latest --skip registry,labels -o json
```
```json
{
  "image": "nginx:latest",
  "passed": false,
  "checks": [
    {
      "check": "age",
      "image": "nginx:latest",
      "passed": true,
      "message": "Image is less than 90 days old",
      "details": {
        "created-at": "2026-02-04T23:53:09Z",
        "age-days": 15.975077092155601,
        "max-age": 90
      }
    }
  ],
  "summary": {
    "total": 6,
    "passed": 3,
    "failed": 3,
    "errored": 0,
    "skipped": [
      "registry",
      "labels"
    ]
  }
}
```

**Version command (full):**
```bash
check-image version -o json
```
```json
{
  "version": "v0.12.1",
  "commit": "a1b2c3d",
  "built-at": "2026-02-18T12:34:56Z",
  "go-version": "go1.26.0",
  "platform": "linux/amd64"
}
```

**Version command (short):**
```bash
check-image version --short -o json
```
```json
{
  "version": "v0.12.1"
}
```

### Exit Codes

| Exit Code | Meaning | Example |
|-----------|---------|---------|
| 0 | Validation succeeded or no checks ran | Image passes all checks |
| 1 | Validation failed | Image is too old, runs as root, exposes unauthorized ports |
| 2 | Execution error | Invalid config file, image not found, invalid arguments |

In the `all` command, if some checks fail validation and others have execution errors, exit code 2 (execution error) takes precedence over exit code 1 (validation failure).

Usage in scripts:
```bash
check-image age nginx:latest --max-age 30
case $? in
  0) echo "Image passed validation" ;;
  1) echo "Image failed validation" ;;
  2) echo "Tool encountered an error" ;;
esac
```

## Configuration Files

The `config/` directory contains sample configuration files that can be used as templates:

### Allowed Ports Files
- `config/allowed-ports.json` - Sample allowed ports configuration in JSON format
- `config/allowed-ports.yaml` - Sample allowed ports configuration in YAML format

Example usage:
```bash
check-image ports nginx:latest --allowed-ports @config/allowed-ports.json
```

### Allowed Platforms Files
- `config/allowed-platforms.json` - Sample allowed platforms configuration in JSON format
- `config/allowed-platforms.yaml` - Sample allowed platforms configuration in YAML format

Example usage:
```bash
check-image platform nginx:latest --allowed-platforms @config/allowed-platforms.yaml
```

### Registry Policy Files
- `config/registry-policy.json` - Sample registry trust policy in JSON format
- `config/registry-policy.yaml` - Sample registry trust policy in YAML format

Example usage:
```bash
check-image registry nginx:latest --registry-policy config/registry-policy.json
```

### Secrets Policy Files
- `config/secrets-policy.json` - Sample secrets detection policy in JSON format
- `config/secrets-policy.yaml` - Sample secrets detection policy in YAML format

Example usage:
```bash
check-image secrets nginx:latest --secrets-policy config/secrets-policy.json
```

### All Checks Configuration Files
- `config/config.json` - Sample configuration for the `all` command in JSON format
- `config/config.yaml` - Sample configuration for the `all` command in YAML format

These files define which checks to run and their parameters. Only checks present in the file are executed.

Example usage:
```bash
check-image all nginx:latest -c config/config.yaml
```

### Reading Configuration from Stdin

All policy and configuration files support reading from standard input using the `-` syntax. This enables dynamic configuration from pipelines and scripts.

**Format Auto-detection:**
- When reading from stdin, the format (JSON or YAML) is automatically detected based on content
- JSON content starts with `{` or `[`
- Everything else is treated as YAML
- Maximum size limit: 10MB

**Supported commands:**

**Registry policy from stdin:**
```bash
echo '{"trusted-registries": ["docker.io", "ghcr.io"]}' | \
  check-image registry nginx:latest --registry-policy -

# Or with YAML
echo 'trusted-registries:
  - docker.io
  - ghcr.io' | \
  check-image registry nginx:latest --registry-policy -
```

**Secrets policy from stdin:**
```bash
cat secrets-policy.yaml | \
  check-image secrets nginx:latest --secrets-policy -
```

**Allowed ports from stdin:**
```bash
echo '{"allowed-ports": [80, 443]}' | \
  check-image ports nginx:latest --allowed-ports @-
```

**All command config from stdin:**
```bash
cat config/config.json | \
  check-image all nginx:latest --config -
```

**Pipeline examples:**
```bash
# Generate config dynamically and validate
jq -n '{"trusted-registries": ["docker.io"]}' | \
  check-image registry nginx:latest --registry-policy -

# Use environment-based configuration
envsubst < config-template.yaml | \
  check-image all nginx:latest --config -
```

### Inline Configuration

The `all` command configuration files support **inline policy embedding**, allowing you to define `registry-policy` and `secrets-policy` as objects directly in the config file instead of referencing separate files. This simplifies deployment by consolidating all configuration into a single file.

**Example files:**
- `config/config-inline.json` - Complete configuration with inline policies (JSON)
- `config/config-inline.yaml` - Complete configuration with inline policies (YAML)

**Inline vs File Reference:**

File-based approach (separate policy files):
```json
{
  "checks": {
    "registry": {
      "registry-policy": "config/registry-policy.json"
    },
    "secrets": {
      "secrets-policy": "config/secrets-policy.json"
    }
  }
}
```

Inline approach (embedded policies):
```json
{
  "checks": {
    "registry": {
      "registry-policy": {
        "trusted-registries": ["docker.io", "ghcr.io", "gcr.io"]
      }
    },
    "secrets": {
      "secrets-policy": {
        "check-env-vars": true,
        "check-files": true,
        "excluded-env-vars": ["PUBLIC_KEY"],
        "excluded-paths": ["/usr/share/**"]
      }
    }
  }
}
```

**Benefits:**
- **Simpler deployment**: Single file contains all configuration
- **Version control**: Easier to track policy changes alongside check parameters
- **Portability**: No need to manage multiple policy files
- **Flexibility**: Mix inline and file references in the same config

**Usage:**
```bash
# Use inline configuration
check-image all nginx:latest --config config/config-inline.yaml

# Inline config with CLI overrides
check-image all nginx:latest --config config/config-inline.json --max-age 30

# Mix with stdin
cat config/config-inline.yaml | check-image all nginx:latest --config -
```

**Note:** Both file paths (strings) and inline objects are supported. You can mix both approaches in the same configuration file based on your needs.

## Development

### Building from Source

To build the project from source:

```bash
git clone https://github.com/jarfernandez/check-image.git
cd check-image
go build -o check-image ./cmd/check-image
```

The binary will be created in the project root directory. You can then move it to a location in your `PATH` or run it directly with `./check-image`.

Alternatively, to install directly to your `GOBIN` directory:

```bash
go install ./cmd/check-image
```

### Pre-Commit Hooks

This project uses pre-commit hooks to enforce code quality and formatting standards before each commit. The hooks automatically run `gofmt`, `go vet`, `golangci-lint`, `go mod tidy`, execute tests with `go-test-mod`, and validate commit messages follow Conventional Commits format.

#### Installation

1. Install pre-commit framework:
   ```bash
   # macOS
   brew install pre-commit

   # Or with pip
   pip install pre-commit
   ```

2. Install golangci-lint:
   ```bash
   # macOS
   brew install golangci-lint

   # Or with go
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

3. (Optional) Install gosec for security scanning:
   ```bash
   go install github.com/securego/gosec/v2/cmd/gosec@latest
   ```

4. Install the pre-commit hooks:
   ```bash
   pre-commit install
   pre-commit install --hook-type commit-msg
   ```

#### Usage

The hooks run automatically on `git commit`. You can also:

- Run manually on all files:
  ```bash
  pre-commit run --all-files
  ```

- Run only tests:
  ```bash
  pre-commit run go-test-mod
  ```

- Skip hooks in emergencies (not recommended):
  ```bash
  git commit --no-verify
  ```

#### What Gets Checked

**Mandatory checks (will block commits):**
- File quality: trailing whitespace, end-of-file newlines, line endings
- Config validation: YAML and JSON syntax
- Go formatting: `gofmt`
- Go tidying: `go mod tidy`
- Go analysis: `go vet`
- Go linting: `golangci-lint` (see `.golangci.yml` for configuration)
- Go tests: `go test` via `go-test-mod`
- Commit message: Conventional Commits format validation

**Warning checks (informational only):**
- Security: `gosec` scans for security issues but doesn't block commits

### Project Structure

- `cmd/check-image/main.go`: The entry point of the application that initializes the CLI and executes commands.
- `cmd/check-image/commands/`: Contains individual command implementations using the `cobra` library.
- `internal/fileutil/`: Provides file reading utilities with support for JSON/YAML parsing and stdin input.
- `internal/imageutil/`: Provides utilities for interacting with container images, such as fetching images from local or remote sources and retrieving image configurations.
- `internal/labels/`: Handles label policy loading and validation for required OCI annotations.
- `internal/output/`: Defines output format types, result structs, and JSON rendering helpers.
- `internal/registry/`: Manages registry policies, including trusted and excluded registries.
- `internal/secrets/`: Handles secrets detection, including policy loading and scanning for sensitive data in environment variables and files.
- `internal/version/`: Manages the application version string, injected at build time via ldflags.
- `config/`: Contains sample configuration files for registry policies, allowed ports, labels, secrets detection, and all-checks configuration.
- `go.mod`: Defines the module and its dependencies.
- `go.sum`: Contains the checksums for module dependencies.

### External Dependencies

- `github.com/spf13/cobra`: For CLI command structure.
- `github.com/google/go-containerregistry`: For interacting with container registries.
- `github.com/sirupsen/logrus`: For logging.
- `github.com/mattn/go-isatty`: For terminal detection (controls log color output).
- `github.com/stretchr/testify`: For test assertions.
- `gopkg.in/yaml.v3`: For parsing YAML configuration files.

## Testing

The project has comprehensive unit tests with 92.2% overall coverage. All tests are deterministic, fast, and run without requiring Docker daemon, registry access, or network connectivity.

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/imageutil -v
go test ./internal/secrets -v

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Coverage

- **internal/version**: 100.0% coverage
- **internal/output**: 100.0% coverage
- **internal/labels**: 98.1% coverage
- **internal/registry**: 96.8% coverage
- **internal/secrets**: 95.8% coverage
- **internal/fileutil**: 90.8% coverage
- **internal/imageutil**: 82.5% coverage
- **cmd/check-image/commands**: 83.1% coverage
- **cmd/check-image**: 60.0% coverage

All tests are deterministic, fast, and run without requiring Docker daemon, registry access, or network connectivity. Tests use in-memory images, temporary directories, and OCI layout structures for validation.

## CI/CD and Release Process

This project uses GitHub Actions for continuous integration and automated releases.

### Continuous Integration

Every pull request and push to `main` automatically runs:

- **Tests**: Full test suite on Linux, macOS, and Windows
- **Linting**: `golangci-lint`, `gofmt`, `go vet`, and `go mod tidy` verification
- **Build Verification**: Cross-compilation for 5 platforms (Linux/macOS/Windows on amd64/arm64)
- **PR Title Validation**: Enforces Conventional Commits format (PRs only)
- **CodeQL Analysis**: Static security analysis for Go code (also runs on a weekly schedule)

All checks must pass before merging to `main`.

The GitHub Action is tested in a dedicated workflow (`test-action.yml`) that triggers on pull requests that modify `action.yml` or `entrypoint.sh`.

### Release Process

Releases are fully automated using [release-please](https://github.com/googleapis/release-please):

1. **Development**: Make changes and commit using [Conventional Commits](https://www.conventionalcommits.org/):
   ```bash
   git commit -m "feat: add new validation check"
   git commit -m "fix: resolve race condition in image loading"
   ```

2. **Merge to Main**: After PR approval and merge, release-please automatically:
   - Analyzes commits since the last release
   - Calculates the next version based on commit types:
     - `feat:` → minor version bump (0.1.0 → 0.2.0)
     - `fix:` → patch version bump (0.1.0 → 0.1.1)
     - `BREAKING CHANGE:` → major version bump (0.1.0 → 1.0.0)
   - Creates/updates a "Release PR" with updated CHANGELOG.md, README.md version references, and action.yml default version

3. **Release**: When the Release PR is merged, three jobs run in sequence:
   - **release-please job**: Creates the git tag and GitHub Release with the changelog
   - **goreleaser job**: Builds binaries for all platforms and uploads them to the release; version is injected via ldflags
   - **docker job**:
     - Lints `Dockerfile` with hadolint
     - Builds a single-arch image (`linux/amd64`) for Trivy security scanning (CRITICAL/HIGH vulnerabilities)
     - Validates the image with check-image itself (dogfooding: size, root-user, ports, secrets)
     - Builds and pushes a multi-arch image (`linux/amd64`, `linux/arm64`) to GHCR with semver tags (`major.minor.patch`, `major.minor`, `major`, `latest`)

### Supported Platforms

Releases include pre-built binaries for:
- Linux: amd64, arm64
- macOS: amd64, arm64
- Windows: amd64

### Commit Message Format

This project requires [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>: <description>

[optional body]

[optional footer]
```

**Allowed types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Test updates
- `chore`: Maintenance tasks
- `perf`: Performance improvements
- `ci`: CI/CD changes
- `build`: Build system changes
- `revert`: Revert previous commit

**Examples:**
```bash
git commit -m "feat: add support for OCI archives"
git commit -m "fix: handle missing environment variables"
git commit -m "docs: update installation instructions"
```

## Contributing

Contributions are welcome! Please ensure that:

- All new commands and utilities are covered by tests.
- Code follows Go's idiomatic error handling practices.
- Variable names are meaningful and descriptive.
- Functions are small and focused on a single responsibility.
- Table-driven tests are used for functions with multiple input scenarios.
- Commit messages follow Conventional Commits format.
- Pre-commit hooks pass before committing.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.
