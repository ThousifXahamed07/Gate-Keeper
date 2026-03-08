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

> _Coming soon — usage examples will be added as features are implemented._

```bash
# Validate your environment against a schema
gatekeeper validate --schema .env.schema.yaml

# Generate a schema from an existing .env file
gatekeeper init --from .env
```

## Contributing

> _Coming soon — contribution guidelines will be added shortly._

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).
