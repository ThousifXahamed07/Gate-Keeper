// Package cli provides CLI-related constants and utilities for Gatekeeper.
package cli

// Exit codes for the Gatekeeper CLI.
// These are the only exit codes that Gatekeeper will ever return.
const (
	// ExitSuccess indicates all validations passed.
	ExitSuccess = 0

	// ExitValidation indicates one or more validation failures.
	// This is also returned when --strict is set and warnings exist.
	ExitValidation = 1

	// ExitSchemaError indicates the schema file could not be parsed or is invalid.
	ExitSchemaError = 2
)

// ExitCodeDescription returns a human-readable description of an exit code.
func ExitCodeDescription(code int) string {
	switch code {
	case ExitSuccess:
		return "all validations passed"
	case ExitValidation:
		return "one or more validation failures"
	case ExitSchemaError:
		return "schema file could not be parsed or is invalid"
	default:
		return "unknown exit code"
	}
}
