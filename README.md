# 🚪 Gatekeeper

> **The env validator that never lets bad config through.**

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CI](https://github.com/ThousifXahamed079/gatekeeper/actions/workflows/ci.yml/badge.svg)](https://github.com/ThousifXahamed079/gatekeeper/actions/workflows/ci.yml)

---

> 🚧 **Under active development** — APIs and CLI flags may change without notice.

---

## Overview

Gatekeeper is a lightweight, fast, and extensible environment variable validator for modern applications. Define your expected env schema in a simple YAML file and let Gatekeeper ensure your configuration is correct **before** your app ever starts.

**Why Gatekeeper?**

- 🔒 Catch missing or malformed env vars at startup, not at runtime
- 📄 Define schemas in human-readable YAML
- ⚡ Zero-overhead — validate once, run with confidence
- 🧩 Extensible validation rules
- 📊 Clear, actionable error reports

## Installation

> _Coming soon — Gatekeeper is not yet published._

```bash
# Via go install
go install github.com/ThousifXahamed079/gatekeeper/cmd/gatekeeper@latest

# Or download a binary from the Releases page
```

## Quick Start

```bash
# Validate your environment against a schema
gatekeeper check --schema .gatekeeper.yaml

# Generate markdown documentation from schema
gatekeeper docs --schema .gatekeeper.yaml --out ENV.md

# Generate .env.example from schema
gatekeeper docs --schema .gatekeeper.yaml --format env-example --out .env.example
```

### Schema Example

Create a `.gatekeeper.yaml` file:

```yaml
version: "1"

groups:
  - name: Database
    description: Database connection settings

vars:
  - name: DATABASE_URL
    type: url
    required: true
    sensitive: true
    description: Database connection URL
    group: Database

  - name: DB_MAX_CONNECTIONS
    type: integer
    required: false
    default: "10"
    description: Maximum database connections
    group: Database
```

### Supported Types

| Type | Description | Example |
|------|-------------|---------|
| `string` | Any string value | `hello` |
| `integer` | Whole numbers | `42` |
| `float` | Decimal numbers | `3.14` |
| `boolean` | `true` or `false` | `true` |
| `url` | Valid URL | `https://example.com` |
| `email` | Valid email address | `user@example.com` |
| `port` | Port number (1-65535) | `8080` |
| `enum` | One of allowed values | `production` |
| `filepath` | File system path | `/var/log/app` |
| `duration` | Go duration string | `30s`, `5m`, `1h` |

## Commands

### `gatekeeper check`

Validate environment variables against the schema.

```bash
gatekeeper check [options]

Options:
  --schema <path>    Path to schema file (default: auto-detect)
  --env-file <path>  Path to .env file (default: .env)
  --format <format>  Output format: text, json, github (default: text)
  --no-env-file      Skip loading .env file, only use OS environment
  --strict           Treat warnings as errors
```

### `gatekeeper docs`

Generate documentation from the schema.

```bash
gatekeeper docs [options]

Options:
  --schema <path>   Path to schema file (default: auto-detect)
  --format <format> Output format: markdown, env-example (default: markdown)
  --out <path>      Output file path (default: stdout)
```

## Contributing

> _Coming soon — contribution guidelines will be added shortly._

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).
