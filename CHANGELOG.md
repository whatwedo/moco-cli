# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.1] - 2026-06-07

### Added

- Homebrew installation via the `whatwedo/homebrew-moco` tap
  (`brew install whatwedo/moco`).

### Changed

- `--version` and help now show the program as `moco-cli`.
- Binaries installed with `go install` report the module version via build info.

## [0.1.0] - 2026-06-07

### Added

- Initial release: a command-line client covering all MOCO API v1 endpoints.
- Commands generated from the OpenAPI spec, grouped by tag with verb subcommands
  (`moco projects list`, `moco projects get <id>`, …).
- Authentication via `--endpoint`/`--token` flags or `MOCO_ENDPOINT`/`MOCO_TOKEN`
  environment variables.
- JSON output (`--output json|raw`), suitable for piping into `jq`.
- German command help texts.

[Unreleased]: https://github.com/whatwedo/moco-cli/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/whatwedo/moco-cli/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/whatwedo/moco-cli/releases/tag/v0.1.0
