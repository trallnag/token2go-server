# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0),
and adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0).

## Unreleased

Nothing.

## [1.0.3](https://github.com/trallnag/token2go-server/compare/v1.0.2...v1.0.3) / 2023-03-05

### Changed

- Bumped minimum required Go version to 1.20. This affects release artifacts.
- Switched from ISC License (ISC) to Apache License (Apache-2.0).

## [1.0.2](https://github.com/trallnag/token2go-server/compare/v1.0.1...v1.0.2) / 2023-02-20

### Changed

- Switched from MIT License to functionally equivalent ISC License.
- Removed Darwin and Windows binaries from GitHub release artifacts.

## [1.0.1](https://github.com/trallnag/token2go-server/compare/v1.0.0...v1.0.1) / 2023-01-08

### Changed

- Added a timeout of 3 seconds to HTTP server
  ([GO-S2114](https://deepsource.io/directory/analyzers/go/issues/GO-S2114),
  [CWE-400](https://cwe.mitre.org/data/definitions/400.html)).

## [1.0.0](https://github.com/trallnag/token2go-server/compare/3d62e4caf205bdf26b12b3900e27540e6ebfbd2e...v1.0.0) / 2022-12-17

Initial stable release of the Token2go server. The API is ready for productive
usage and breaking changes are not expected in the near future.
