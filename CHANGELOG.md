# Changelog

All notable changes to this project will be documented in this file. This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1](https://github.com/jarfernandez/check-image/releases/tag/v0.1.1) (2026-02-02)

### Bug Fixes

* Add v prefix to version output in binaries ([324f08f](https://github.com/jarfernandez/check-image/commit/324f08f48630a89b265cb3562f85fe7c590bff88))

## [0.1.0](https://github.com/jarfernandez/check-image/releases/tag/v0.1.0) (2026-02-02)

### Features

* Initial release of check-image CLI tool
* Image validation commands: size, age, registry, ports, root-user, secrets
* Support for multiple image sources (OCI layout, OCI archive, Docker archive, registry)
* Comprehensive test coverage (87.6%)
* CI/CD pipeline with automated releases

### Build System

* Add GitHub Actions workflows for automated CI/CD
* Multi-platform builds (Linux, macOS, Windows on amd64/arm64)
* Automated releases with semantic versioning using release-please
* GoReleaser configuration for multi-platform binary distribution
