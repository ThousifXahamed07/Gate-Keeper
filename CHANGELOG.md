# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-01-XX

### Added

- **Core CLI** (`gatekeeper check`)
  - Validate environment variables against YAML schema
  - Auto-detect `.gatekeeper.yaml` schema files
  - Load from OS environment and `.env` files
  - Output formats: text (default), JSON, GitHub Actions
  - Exit codes for CI/CD integration (0=success, 1=validation failure, 2=schema error)

- **Schema Definition** (`.gatekeeper.yaml`)
  - Groups for organizing related variables
  - Variable definitions with name, type, description
  - Required/optional field support
  - Default values
  - Sensitive value marking (redacted in output)
  - Pattern validation (regex)
  - Enum/allowed values

- **10 Supported Types**
  - `string` - Any string value
  - `integer` - Whole numbers (positive, negative, zero)
  - `float` - Decimal numbers
  - `boolean` - true/false, yes/no, 1/0, on/off
  - `url` - Valid HTTP/HTTPS URLs
  - `email` - Valid email addresses
  - `port` - Valid port numbers (1-65535)
  - `enum` - Constrained to allowed values
  - `filepath` - File system paths
  - `duration` - Go duration format (e.g., "30s", "5m", "1h")

- **Output Formats**
  - **Text**: Human-readable with ANSI colors (auto-detected)
  - **JSON**: Structured output for programmatic use
  - **GitHub**: Native `::error::` and `::warning::` annotations

- **CI/CD Integration**
  - Exit code 0: All validations passed
  - Exit code 1: One or more validation failures
  - Exit code 2: Schema parsing or configuration error
  - `--strict` flag: Treat warnings as errors

- **Security**
  - Sensitive values redacted in all output formats
  - Values marked with `sensitive: true` shown as `[REDACTED]`
  - Pattern validation errors don't expose actual values

- **Developer Experience**
  - `--no-env-file`: Skip `.env` file loading
  - `--schema`: Specify custom schema file path
  - `--env-file`: Specify custom `.env` file path
  - `gatekeeper version`: Show version information
  - `gatekeeper help`: Show usage information

### Technical Details

- Written in Go 1.21+
- Zero external runtime dependencies (stdlib only for core)
- Cross-platform support (Linux, macOS, Windows)
- Comprehensive test coverage
- GoReleaser for binary distribution

[Unreleased]: https://github.com/ThousifXahamed07/Gate-Keeper/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/ThousifXahamed07/Gate-Keeper/releases/tag/v0.1.0
