# Changelog

All notable changes to this project will be documented in this file. This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.0](https://github.com/jarfernandez/check-image/compare/v0.4.0...v0.5.0) (2026-02-14)


### Features

* Add `--output`/`-o` flag with JSON support ([#45](https://github.com/jarfernandez/check-image/issues/45)) ([436389b](https://github.com/jarfernandez/check-image/commit/436389b62d673df874b0400987a2f173b7a5607d))

## [0.4.0](https://github.com/jarfernandez/check-image/compare/v0.3.0...v0.4.0) (2026-02-11)


### Features

* add `--fail-fast` flag to `all` command ([#43](https://github.com/jarfernandez/check-image/issues/43)) ([52e4863](https://github.com/jarfernandez/check-image/commit/52e4863afb5d95509d132e73c2c3f1c938e51aa1))

## [0.3.0](https://github.com/jarfernandez/check-image/compare/v0.2.0...v0.3.0) (2026-02-10)


### Features

* add `all` command to run all validation checks at once ([#41](https://github.com/jarfernandez/check-image/issues/41)) ([8fac20e](https://github.com/jarfernandez/check-image/commit/8fac20e26aae87e4d76cf4d06d27b51a36d64a3e))

## [0.2.0](https://github.com/jarfernandez/check-image/compare/v0.1.1...v0.2.0) (2026-02-09)


### Features

* Update module path to use full GitHub URL ([#3](https://github.com/jarfernandez/check-image/issues/3)) ([42a9f22](https://github.com/jarfernandez/check-image/commit/42a9f2202d8971634c1004ce79ad94f41aaee1c0))


### Bug Fixes

* Chain GoReleaser as job within release-please workflow ([2a5e383](https://github.com/jarfernandez/check-image/commit/2a5e383dfb0ce497fb804604b1cb366b33db1e62))
* Configure release-please to use simple tag format without component prefix ([24b7a0c](https://github.com/jarfernandez/check-image/commit/24b7a0c6936ff932bc0522b256588cb238bb748d))
* Correct GoReleaser ldflags to use correct module path for version injection ([56e44ed](https://github.com/jarfernandez/check-image/commit/56e44edf9e15f2664af987ca4ed6ddbbcf1b4350))
* Update release-please config to v4 manifest format ([#5](https://github.com/jarfernandez/check-image/issues/5)) ([ea52291](https://github.com/jarfernandez/check-image/commit/ea522915a5d7871bedc3eeff45ff5e6e49b4db5f))


### Documentation

* Add pre-built binary installation instructions and clarify version behavior ([12737b3](https://github.com/jarfernandez/check-image/commit/12737b361473e7f278f2d079d3af54a9fa1f875f))
* Document release pipeline architecture in CLAUDE.md ([f17ffad](https://github.com/jarfernandez/check-image/commit/f17ffad008ec6e8e572ea489940e7ce8dfcfa64a))
* Update installation instructions with GitHub install method ([#4](https://github.com/jarfernandez/check-image/issues/4)) ([434d3ec](https://github.com/jarfernandez/check-image/commit/434d3ec5f1904767b715d5bc94e14e35b2648227))

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
