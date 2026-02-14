# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Check Image is a Go CLI tool for validating container images against security and operational standards. It validates images for size, age, exposed ports, registry trust, and security configurations (like non-root users).

## Build and Test Commands

### Building
Use `go install` instead of `go build` to install the binary to `GOBIN`:
```bash
go install ./cmd/check-image
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/imageutil
go test ./internal/registry
go test ./internal/secrets
```

### Running the CLI
```bash
# Show help
./check-image --help

# Set log level (trace, debug, info, warn, error, fatal, panic)
./check-image --log-level debug <command>
```

## Architecture

### Command Pattern
All validation commands follow a consistent pattern:
1. Commands are in `cmd/check-image/commands/` and use Cobra framework
2. Each `runX()` function returns `(*output.CheckResult, error)` — it never prints directly
3. The `RunE` handler in each command calls `renderResult()` to output text or JSON, then updates the global `Result` variable based on `result.Passed`
4. `Result` (`ValidationFailed`, `ValidationSucceeded`, or `ValidationSkipped`) is defined in `root.go` and drives the exit code in `main.go`

### Output Format
- Controlled by the `--output`/`-o` global flag (values: `text` default, `json`)
- `internal/output/format.go`: Defines `Format` type, `ParseFormat()`, and `RenderJSON()` helper
- `internal/output/results.go`: Result structs (`CheckResult`, `AgeDetails`, `SizeDetails`, `PortsDetails`, `RegistryDetails`, `RootUserDetails`, `SecretsDetails`, `AllResult`, `Summary`, `VersionResult`)
- `cmd/check-image/commands/render.go`: Text renderers for each check; `renderResult()` dispatches to JSON or text based on `OutputFmt`
- In JSON mode, `main.go` suppresses the final "Validation succeeded/failed" text message (it's already in the JSON)

### Image Retrieval Strategy
The `imageutil` package implements a transport-aware retrieval strategy with fallback support:
- **Transport Detection**: `ParseReference()` detects transport prefix (e.g., `oci:`, `oci-archive:`, `docker-archive:`)
- **OCI Layout Support**: `GetOCILayoutImage()` loads images from OCI layout directories (supports both tag and digest references)
- **OCI Archive Support**: `GetOCIArchiveImage()` loads images from OCI tarball archives
  - Extracts tarball to temporary directory using `extractOCIArchive()`
  - Validates paths to prevent path traversal attacks
  - Enforces 5GB decompression limit to prevent decompression bombs
  - Supports gzipped (.gz, .tgz) and uncompressed tarballs
  - Then uses OCI layout loader on extracted content
- **Docker Archive Support**: `GetDockerArchiveImage()` loads images from Docker tarball archives created with `docker save`
  - Uses `tarball.ImageFromPath()` from go-containerregistry
  - Supports tag-based image selection within multi-image archives
- **Default Behavior** (no transport prefix): `GetImage()` tries local Docker daemon first, then falls back to remote registry
- **Explicit Transports**: When a transport prefix is specified, only that source is attempted (no fallback)
- `GetLocalImage()` retrieves from Docker daemon
- `GetRemoteImage()` fetches from remote registry with keychain authentication
- All functions use `github.com/google/go-containerregistry` for image operations

**Supported Transport Syntax** (Skopeo-compatible):
- `oci:/path/to/layout:tag` - OCI layout directory with tag
- `oci:/path/to/layout@sha256:abc...` - OCI layout directory with digest
- `oci-archive:/path.tar:tag` - OCI tarball archive with tag
- `oci-archive:/path.tar@sha256:abc...` - OCI tarball archive with digest
- `docker-archive:/path.tar:tag` - Docker tarball archive (saved with `docker save`)
- `nginx:latest` or `docker.io/nginx:latest` - Default behavior (daemon → registry)

### Validation Commands

**size**: Validates image size and layer count
- Flags: `--max-size` (MB, default 500), `--max-layers` (default 20)
- Uses `GetRemoteImage()` directly (not the fallback pattern)

**age**: Validates image creation date
- Flags: `--max-age` (days, default 90)
- Reads `config.Created` timestamp from image config

**registry**: Validates image registry against a trust policy
- Flags: `--registry-policy` (required, JSON or YAML file)
- Policy format: specify either `trusted-registries` (allowlist) or `excluded-registries` (blocklist), but not both
- Allowlist mode: only registries in `trusted-registries` are allowed
- Blocklist mode: all registries except those in `excluded-registries` are allowed

**ports**: Validates exposed ports against an allowed list
- Flags: `--allowed-ports` (comma-separated list or `@file.json`/`@file.yaml`)
- File format: `{"allowed-ports": [80, 443]}`
- Parses ports from image config's `ExposedPorts` field (format: "8080/tcp")

**root-user**: Validates that image runs as non-root
- No flags
- Checks if `config.Config.User` is empty or "root"

**secrets**: Validates that image does not contain sensitive data (passwords, tokens, keys)
- Flags: `--secrets-policy` (optional, JSON or YAML file), `--skip-env-vars`, `--skip-files`
- Scans environment variables for sensitive patterns (case-insensitive matching for keywords like password, secret, token, key, etc.)
- Scans all image layers for files matching sensitive patterns (SSH keys, cloud credentials, password files, etc.)
- Uses `DefaultFilePatterns` map in `internal/secrets/policy.go` as single source of truth for patterns and descriptions
- Policy supports `excluded-paths`, `excluded-env-vars`, and custom patterns
- Works out-of-the-box with sensible defaults when no policy file is provided

**all**: Runs all validation checks on a container image at once
- Flags: `--config` (`-c`, config file), `--skip` (comma-separated checks to skip), `--fail-fast` (stop on first failure), plus all individual check flags (`--max-age`, `--max-size`, `--max-layers`, `--allowed-ports`, `--registry-policy`, `--secrets-policy`, `--skip-env-vars`, `--skip-files`)
- Precedence: CLI flags > config file values > defaults; `--skip` always wins
- Without `--config`: runs all 6 checks with defaults (except skipped)
- With `--config`: only runs checks present in the config file (except skipped)
- Uses `applyConfigValues()` with `cmd.Flags().Changed()` to respect CLI overrides
- Wrappers: `runPortsForAll()` calls `parseAllowedPorts()` before `runPorts()`; `runRegistryForAll()` skips gracefully when no `--registry-policy` is provided
- Continue-on-error (default): if a check returns an error, logs it, sets `Result = ValidationFailed`, and continues with the next check
- Fail-fast (`--fail-fast`): stops execution on the first check that fails (validation failure or execution error)

**version**: Shows the check-image version
- No flags (uses global `--output` flag for JSON support)
- Returns the version string from `internal/version.Version`
- Defaults to "dev" if no version is set
- In JSON mode: outputs `{"version": "v0.4.0"}`
- Version is injected at build time using ldflags: `-ldflags "-X check-image/internal/version.Version=v0.1.0"`

### Configuration Files
Sample configuration files are in `config/`:
- `allowed-ports.yaml` / `allowed-ports.json`: Allowed ports list
- `registry-policy.yaml` / `registry-policy.json`: Trusted registries list
- `config.yaml` / `config.json`: All-checks configuration (defines which checks to run and their parameters for the `all` command)
- `secrets-policy.yaml` / `secrets-policy.json`: Secrets detection policy with exclusions

Both JSON and YAML formats are supported throughout the tool. Format detection is by file extension (`.yaml`, `.yml` for YAML, otherwise JSON).

### Registry Policy Logic
In `internal/registry/policy.go`:
- Policy must specify either `trusted-registries` or `excluded-registries`, not both
- Allowlist mode (trusted-registries): only registries in the list are allowed
- Blocklist mode (excluded-registries): all registries except those in the list are allowed

### Secrets Detection Logic
In `internal/secrets/`:
- `policy.go`: Defines `DefaultFilePatterns` map as single source of truth for patterns and descriptions
- `detector.go`: Implements `CheckEnvironmentVariables()` and `CheckFilesInLayers()`
- Environment variable detection uses case-insensitive pattern matching against variable names
- File detection scans all layers (secrets in earlier layers remain in image history)
- Supports exclusion lists for both paths and environment variables to handle false positives
- Pattern descriptions consolidated in `DefaultFilePatterns` map to avoid duplication

### Logging
- Uses `logrus` with stderr output
- Timestamps formatted as "2006-01-02 15:04:05"
- Colors disabled when not running in a terminal
- Set level via `--log-level` flag on any command

### Release Pipeline
Single workflow in `.github/workflows/release-please.yml` with two chained jobs:

1. **release-please job**: Runs on every push to `main`. Creates/updates a release PR with changelog and version bump. When the release PR is merged, creates the git tag and GitHub release. Exports `releases_created` and `tag_name` as outputs.
2. **goreleaser job**: Depends on the release-please job. Only runs when `releases_created == 'true'`. Checks out the tag, builds binaries for linux/darwin/windows (amd64/arm64), and uploads them to the GitHub release via `mode: append`.

Both jobs must be in the same workflow because tags created by `GITHUB_TOKEN` do not trigger other workflows (GitHub limitation to prevent infinite loops).

**Configuration files**:
- `.github/release-please-config.json`: Release-please settings (release type, changelog sections)
- `.github/release-please-manifest.json`: Tracks the current version
- `.goreleaser.yml`: GoReleaser build configuration (platforms, archives). Changelog is disabled here; release-please handles it.

**Important**: Release-please tracks release PRs by the `autorelease: pending` label, not by title. When a release PR is successfully released, the label changes to `autorelease: tagged`. If this label gets stuck, release-please will abort with "untagged, merged release PRs outstanding".

**Commit types and releases**: Only `feat:`, `fix:`, `perf:`, and `refactor:` commits trigger version bumps. Use `ci:`, `chore:`, or `docs:` for changes that should not trigger releases (these are configured as `hidden: true` in the changelog sections).

## Go Project Rules

### GitHub Integration
- Use the GitHub CLI (`gh`) for all interactions with GitHub (issues, pull requests, comments).
- Use Conventional Commits format for all commit messages (e.g., `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).
- Do not add Claude as a co-author in commits (no `Co-Authored-By: Claude` lines).

### Go Best Practices

#### Project Structure
- Follow a standard Go project layout.
- Place the CLI entrypoint in `cmd/check-image/`.
- Keep `main.go` minimal and move logic to packages.
- Use `internal/` for non-public packages.
- Use `pkg/` only for reusable packages.

#### Package Design
- Keep packages small and focused.
- Avoid circular dependencies.
- Prefer flat package structures.

#### Naming
- Use clear and concise names.
- Exported identifiers must start with a capital letter.
- Unexported identifiers must start with a lowercase letter.
- Avoid stuttering in names.

#### Error Handling
- Handle all errors explicitly.
- Wrap errors with context using `fmt.Errorf("...: %w", err)`.
- Return errors instead of panicking.
- Write error messages in lowercase and without punctuation.

#### CLI Rules
- Use a single CLI framework consistently.
- Validate user input early and fail fast.
- Print user-facing errors to `stderr`.
- Return proper exit codes (`0` for success, non-zero for failure).

#### Interfaces
- Define interfaces at the point of use.
- Keep interfaces small.
- Accept interfaces and return concrete types.

#### Testing
- Use table-driven tests when appropriate.
- Test behavior, not implementation.
- Use the standard `testing` package with `testify` for assertions.
- All tests must be deterministic, fast, and isolated (no Docker daemon, registry, or network access required).
- Use in-memory images and temporary directories for testing.
- Comprehensive unit tests cover all commands and internal packages with 84.1% overall coverage.

#### Formatting and Tooling
- Format code with `gofmt`.
- Run `go vet` and `golangci-lint`.
- Keep `go.mod` tidy.
- Pre-commit hooks automatically enforce these requirements before each commit.
- See `.golangci.yml` for linter configuration (balanced settings).
- Install hooks with: `pre-commit install && pre-commit install --hook-type commit-msg`.

#### Documentation
- Explain *why*, not *what*.
- Comment only exported identifiers and complex logic.
- Keep documentation aligned with the code.

#### Dependencies
- Minimize external dependencies.
- Prefer the Go standard library.
- Avoid unnecessary dependency bloat.
