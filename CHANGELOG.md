# Changelog

All notable changes to this project will be documented in this file. This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0](https://github.com/jarfernandez/check-image/compare/v0.2.0...v0.3.0) (2026-02-09)


### Features

* add version command ([64ecdb8](https://github.com/jarfernandez/check-image/commit/64ecdb8852dc171755f8b4ff72129b3af7f74395))
* **imageutil:** add local image retrieval ([010016c](https://github.com/jarfernandez/check-image/commit/010016c195a292fbda85329f1460f2d83ec07f40))
* **ports:** change allowed_ports to allowed-ports in config schema ([283df03](https://github.com/jarfernandez/check-image/commit/283df0351b4c2ed04a70d04c2c456912336b905d))
* **registry:** add command to validate trusted image registries ([0ca34c9](https://github.com/jarfernandez/check-image/commit/0ca34c9a4ccc4f53f7926387ff8cbff9934e3bb9))
* **registry:** change trusted_registries to trusted-registries and excluded_registries to excluded-registries in config schema ([eb8a9d4](https://github.com/jarfernandez/check-image/commit/eb8a9d404004e4244162037e6b82a68bb3608871))
* **registry:** enhance registry policy validation ([be77ebe](https://github.com/jarfernandez/check-image/commit/be77ebe8e6787411820bbdef6a99748f6f41efa4))
* **secrets:** add command to detect sensitive data in container images ([ad027cb](https://github.com/jarfernandez/check-image/commit/ad027cbc2d45152e18bb156a1a24f648088ac99b))
* support multiple image formats ([8e09297](https://github.com/jarfernandez/check-image/commit/8e0929776374536c435e4c400d4ed40f0e63c429))
* Update module path to use full GitHub URL ([#3](https://github.com/jarfernandez/check-image/issues/3)) ([9abb173](https://github.com/jarfernandez/check-image/commit/9abb173b1191f9e06b5865d53f8886fb6dc052ba))


### Bug Fixes

* add v prefix to version output in binaries ([324f08f](https://github.com/jarfernandez/check-image/commit/324f08f48630a89b265cb3562f85fe7c590bff88))
* Chain GoReleaser as job within release-please workflow ([9e4dc84](https://github.com/jarfernandez/check-image/commit/9e4dc84587bc618b9738712ba3b61effe6505e25))
* Configure release-please to use simple tag format without component prefix ([cb3fa85](https://github.com/jarfernandez/check-image/commit/cb3fa856c0c04fcc7eb5aa956f2f1ffddba46735))
* Correct GoReleaser ldflags to use correct module path for version injection ([7d9bd49](https://github.com/jarfernandez/check-image/commit/7d9bd4929b5b0852dbe936b48a7fae1db0ff9a2c))
* resolve security warnings detected by gosec ([54d7ae4](https://github.com/jarfernandez/check-image/commit/54d7ae45db268edbca3be3f4ce006167cdcff401))
* **size:** retrieve local image with remote fallback ([6b8151f](https://github.com/jarfernandez/check-image/commit/6b8151f3598251e72f5553dac85232aec8ffea01))
* Update release-please config to v4 manifest format ([#5](https://github.com/jarfernandez/check-image/issues/5)) ([dc54f9b](https://github.com/jarfernandez/check-image/commit/dc54f9b6a146dbde19b011abe2ff9e5ea328e198))


### Code Refactoring

* eliminate duplicate code across commands and policies ([399e980](https://github.com/jarfernandez/check-image/commit/399e980fbce6bb1aba3abac3c6aa3f05e24ceacf))


### Documentation

* Add CI/CD section to README table of contents ([1ab7465](https://github.com/jarfernandez/check-image/commit/1ab74657251a5e2c4136817bb27f31c82161ac5f))
* add CLAUDE.md with project guidelines and architecture ([f3953a2](https://github.com/jarfernandez/check-image/commit/f3953a26fbda690037750449acf013f2e8fe486a))
* add Contributor Covenant Code of Conduct ([fa014a5](https://github.com/jarfernandez/check-image/commit/fa014a53a49c68a70e29e71c834f1bcd4a8ffc3e))
* add initial CHANGELOG.md for v0.1.0 release ([4bc5de3](https://github.com/jarfernandez/check-image/commit/4bc5de3b9a56d691a5390d06a03405afa21e87f0))
* add MIT License ([66686c8](https://github.com/jarfernandez/check-image/commit/66686c88e643c45ac6a88d70b9bf264585424cc3))
* Add pre-built binary installation instructions and clarify version behavior ([cc4db72](https://github.com/jarfernandez/check-image/commit/cc4db722f5736476bc51d1957f9265fefc72eec0))
* add README ([31f07a2](https://github.com/jarfernandez/check-image/commit/31f07a2066ee7b3e22b8bb33e57f9dfbcf94c54a))
* Document release pipeline architecture in CLAUDE.md ([fb36f4c](https://github.com/jarfernandez/check-image/commit/fb36f4c20a4be80bd7290dc78a193ad02caa3126))
* enhance CLAUDE.md with detailed image retrieval strategy and testing guidelines ([cbcf114](https://github.com/jarfernandez/check-image/commit/cbcf11478aa69b215fe80b57b70f9bcaea19734e))
* standardize version tag format in examples ([4bbca7a](https://github.com/jarfernandez/check-image/commit/4bbca7a76005f3b158f59d0cf37a0ef0415cc105))
* update CHANGELOG.md with v0.1.1 release ([68a1702](https://github.com/jarfernandez/check-image/commit/68a1702a4dbd026b7083bea50f8b362b25dcd50f))
* update CLAUDE.md with coverage and version command ([bf2218c](https://github.com/jarfernandez/check-image/commit/bf2218c605a1d0f80e2f8f5b31cee9ed8c5afa90))
* update CLAUDE.md with secrets command documentation ([2b46ffe](https://github.com/jarfernandez/check-image/commit/2b46ffe7a56f877062f560d9b2cbc530aa469399))
* Update installation instructions with GitHub install method ([#4](https://github.com/jarfernandez/check-image/issues/4)) ([b7d4dc9](https://github.com/jarfernandez/check-image/commit/b7d4dc9fbd5cd17749b318da3773c5b3f28d5c9e))
* update README ([1ad79bb](https://github.com/jarfernandez/check-image/commit/1ad79bb7b1df557c2c5dc3da0c8316f17b30d3de))
* update README with comprehensive documentation ([962115b](https://github.com/jarfernandez/check-image/commit/962115bbed3d86885457685a8bfed3dc87ec7535))
* update README with coverage and version command ([3eb8b87](https://github.com/jarfernandez/check-image/commit/3eb8b877c33fb3f1f41be8aae5797bfe2d062e75))

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
