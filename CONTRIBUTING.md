# Contributing to Gatekeeper

Thank you for your interest in contributing to Gatekeeper! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## How to Contribute

### Reporting Bugs

1. **Search existing issues** to avoid duplicates
2. **Use the bug report template** when creating a new issue
3. **Include**:
   - Go version (`go version`)
   - Operating system and architecture
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant schema/config files (sanitized)

### Suggesting Features

1. **Search existing issues** to avoid duplicates
2. **Describe the use case** clearly
3. **Explain why** this would benefit users
4. **Consider** backward compatibility

### Pull Requests

1. **Fork** the repository
2. **Create a branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes** following the coding standards
4. **Write tests** for new functionality
5. **Run tests** before submitting:
   ```bash
   make test
   make lint
   ```
6. **Commit** with clear messages following conventional commits
7. **Push** and create a Pull Request

## Development Setup

### Prerequisites

- Go 1.21 or later
- Make
- golangci-lint (optional, for linting)

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/Gate-Keeper.git
cd Gate-Keeper

# Install dependencies
go mod download

# Run tests
make test

# Build
make build

# Run linter
make lint
```

### Project Structure

```
.
├── cmd/gatekeeper/      # CLI entrypoint
├── internal/
│   ├── cli/             # Exit codes and CLI utilities
│   ├── envloader/       # .env file parsing
│   ├── reporter/        # Output formatters (text/json/github)
│   ├── schema/          # YAML schema parsing
│   └── validator/       # Validation logic
├── examples/            # Example schemas
├── scripts/             # Build and utility scripts
└── .github/workflows/   # CI/CD pipelines
```

## Coding Standards

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Keep functions focused and small
- Write descriptive variable names
- Add comments for exported functions

### Testing

- Write table-driven tests
- Aim for >80% coverage
- Test edge cases and error paths
- Use descriptive test names

### Commits

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add new validation type
fix: handle empty string values correctly
docs: update README with examples
test: add edge case tests for URL validation
refactor: simplify validator pipeline
chore: update dependencies
```

### Documentation

- Update README for user-facing changes
- Add godoc comments for exported symbols
- Include examples in documentation

## Adding a New Validation Type

1. Add the type constant in `internal/schema/types.go`
2. Add to `SupportedTypes` map
3. Implement validator in `internal/validator/rules.go`
4. Register in `typeValidators` map
5. Add tests in `internal/validator/rules_test.go`
6. Update README with documentation

## Release Process

Releases are automated via GitHub Actions when a tag is pushed:

```bash
git tag v1.0.0
git push origin v1.0.0
```

GoReleaser handles cross-compilation and artifact publishing.

## Questions?

- Open an issue for general questions
- Check existing documentation and issues first
- Be patient - maintainers are volunteers

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
