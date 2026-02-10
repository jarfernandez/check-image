# Check Image

Check Image is a Go-based CLI tool designed for validating container images. It ensures that images meet specific standards such as size, age, ports, and security configurations. This project follows Go conventions for command-line tools and is structured into `cmd` and `internal` directories.

## Table of Contents

- [Installation](#installation)
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
VERSION=0.2.0

# macOS (Apple Silicon)
curl -sL https://github.com/jarfernandez/check-image/releases/download/v${VERSION}/check-image_${VERSION}_darwin_arm64.tar.gz | tar xz
sudo mv check-image /usr/local/bin/

# macOS (Intel)
curl -sL https://github.com/jarfernandez/check-image/releases/download/v${VERSION}/check-image_${VERSION}_darwin_amd64.tar.gz | tar xz
sudo mv check-image /usr/local/bin/

# Linux (amd64)
curl -sL https://github.com/jarfernandez/check-image/releases/download/v${VERSION}/check-image_${VERSION}_linux_amd64.tar.gz | tar xz
sudo mv check-image /usr/local/bin/

# Linux (arm64)
curl -sL https://github.com/jarfernandez/check-image/releases/download/v${VERSION}/check-image_${VERSION}_linux_arm64.tar.gz | tar xz
sudo mv check-image /usr/local/bin/

# Windows (amd64)
# Download https://github.com/jarfernandez/check-image/releases/download/v${VERSION}/check-image_${VERSION}_windows_amd64.zip
# and extract to a directory in your PATH
```

Pre-built binaries include the correct version number (e.g., `check-image version` returns `v0.2.0`).

### Install with Go

**Requirements:** Go 1.24 or newer

```bash
# Install the latest version
go install github.com/jarfernandez/check-image/cmd/check-image@latest

# Or install a specific version
go install github.com/jarfernandez/check-image/cmd/check-image@v0.2.0
```

This will install the `check-image` binary to your `GOBIN` directory.

**Note:** Binaries installed with `go install` will show version as `dev` when running `check-image version`. This is expected behavior as `go install` compiles from source without version injection. For production use with correct version numbers, use pre-built binaries from releases.

### Install from Source

If you've cloned the repository, you can install it locally:

```bash
go install ./cmd/check-image
```

This is useful for development. The version will show as `dev`.

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

#### `all`
Runs all validation checks on a container image at once.

```bash
check-image all <image> [flags]
```

Options:
- `--config`, `-c`: Path to configuration file (JSON or YAML)
- `--skip`: Comma-separated list of checks to skip (age, size, ports, registry, root-user, secrets)
- `--max-age`, `-a`: Maximum age in days (default: 90)
- `--max-size`, `-m`: Maximum size in MB (default: 500)
- `--max-layers`, `-y`: Maximum number of layers (default: 20)
- `--allowed-ports`, `-p`: Comma-separated list of allowed ports or `@<file>`
- `--registry-policy`, `-r`: Registry policy file (JSON or YAML)
- `--secrets-policy`, `-s`: Secrets policy file (JSON or YAML)
- `--skip-env-vars`: Skip environment variable checks in secrets detection
- `--skip-files`: Skip file system checks in secrets detection

Precedence rules:
1. Without `--config`: all 6 checks run with defaults, except those in `--skip`
2. With `--config`: only checks present in the config file run, except those in `--skip`
3. CLI flags override config file values
4. `--skip` always takes precedence over the config file

Examples:
```bash
# Run all checks with defaults
check-image all nginx:latest

# Run all checks with custom limits
check-image all nginx:latest --max-age 30 --max-size 200

# Skip specific checks
check-image all nginx:latest --skip registry,secrets

# Use a configuration file
check-image all nginx:latest -c config/config.yaml

# Config file with CLI overrides and skip
check-image all nginx:latest -c config/config.yaml --max-age 30 --skip secrets
```

#### `version`
Shows the check-image version.

```bash
check-image version
```

The version can be set at build time using ldflags:
```bash
go build -ldflags "-X github.com/jarfernandez/check-image/internal/version.Version=v0.1.0" ./cmd/check-image
```

### Global Flags

All commands support:
- `--log-level`: Set log level (trace, debug, info, warn, error, fatal, panic)

## Configuration Files

The `config/` directory contains sample configuration files that can be used as templates:

### Allowed Ports Files
- `config/allowed-ports.json` - Sample allowed ports configuration in JSON format
- `config/allowed-ports.yaml` - Sample allowed ports configuration in YAML format

Example usage:
```bash
check-image ports nginx:latest --allowed-ports @config/allowed-ports.json
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
- `internal/imageutil/imageutil.go`: Provides utilities for interacting with container images, such as fetching images from local or remote sources and retrieving image configurations.
- `internal/registry/policy.go`: Manages registry policies, including trusted and excluded registries.
- `internal/secrets/`: Handles secrets detection, including policy loading and scanning for sensitive data in environment variables and files.
- `config/`: Contains sample configuration files for registry policies, allowed ports, and secrets detection.
- `go.mod`: Defines the module and its dependencies.
- `go.sum`: Contains the checksums for module dependencies.

### External Dependencies

- `github.com/spf13/cobra`: For CLI command structure.
- `github.com/google/go-containerregistry`: For interacting with container registries.
- `gopkg.in/yaml.v3`: For parsing YAML configuration files.
- `github.com/sirupsen/logrus`: For logging.

## Testing

The project has comprehensive unit tests with 87.6% overall coverage. All tests are deterministic, fast, and run without requiring Docker daemon, registry access, or network connectivity.

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

- **internal/version**: 100% coverage
- **internal/registry**: 100% coverage
- **internal/secrets**: 96.5% coverage
- **cmd/check-image/commands**: 86.3% coverage
- **internal/fileutil**: 82.9% coverage
- **internal/imageutil**: 73.9% coverage
- **cmd/check-image**: 63.6% coverage

All tests are deterministic, fast, and run without requiring Docker daemon, registry access, or network connectivity. Tests use in-memory images, temporary directories, and OCI layout structures for validation.

## CI/CD and Release Process

This project uses GitHub Actions for continuous integration and automated releases.

### Continuous Integration

Every pull request and push to `main` automatically runs:

- **Tests**: Full test suite on Linux, macOS, and Windows
- **Linting**: `golangci-lint` with strict checks
- **Build Verification**: Cross-compilation for 5 platforms (Linux/macOS/Windows on amd64/arm64)
- **PR Title Validation**: Enforces Conventional Commits format

All checks must pass before merging to `main`.

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
   - Creates/updates a "Release PR" with updated CHANGELOG.md

3. **Release**: When the Release PR is merged:
   - Git tag is created (e.g., `v0.2.0`)
   - GoReleaser builds binaries for all platforms
   - GitHub Release is created with binaries and changelog
   - Version is injected into binaries via ldflags

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
