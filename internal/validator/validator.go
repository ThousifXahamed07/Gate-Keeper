package validator

import (
	"fmt"
	"regexp"

	"github.com/ThousifXahamed079/gatekeeper/internal/schema"
)

// Status represents the result status of a validation check.
type Status int

const (
	// StatusPass indicates the validation passed.
	StatusPass Status = iota
	// StatusFail indicates the validation failed.
	StatusFail
	// StatusWarn indicates a warning (non-blocking).
	StatusWarn
)

// String returns a string representation of the status.
func (s Status) String() string {
	switch s {
	case StatusPass:
		return "pass"
	case StatusFail:
		return "fail"
	case StatusWarn:
		return "warn"
	default:
		return "unknown"
	}
}

// ValidationResult holds the result of validating a single environment variable.
type ValidationResult struct {
	VarName   string // Name of the environment variable
	Value     string // Actual value (will be redacted by reporter if sensitive)
	Expected  string // What was expected
	Status    Status // Pass, Fail, Warn
	Message   string // Human-readable result message
	Sensitive bool   // Whether to redact the value in output
	Source    string // "os", "env-file", "default", "missing"
	VarType   string // "string", "url", "port", etc.
	Required  bool   // Whether the variable is required
}

// Validate validates environment variables against the schema.
// env contains values from the environment (OS or .env file merged).
// sources tracks the origin of each value ("os", "env-file").
func Validate(s *schema.Schema, env map[string]string, sources map[string]string) []ValidationResult {
	var results []ValidationResult

	for _, v := range s.Vars {
		result := validateVar(v, env, sources)
		results = append(results, result)
	}

	return results
}

// validateVar validates a single variable against its schema definition.
func validateVar(v schema.Var, env map[string]string, sources map[string]string) ValidationResult {
	result := ValidationResult{
		VarName:   v.Name,
		Sensitive: v.Sensitive,
		VarType:   v.Type,
		Required:  v.Required,
	}

	value, exists := env[v.Name]
	source := sources[v.Name]

	// Check if the variable is set
	// Note: empty string "" is a valid value - only "not present in map" counts as missing
	if !exists {
		if v.Default != "" {
			// Use default value
			result.Value = v.Default
			result.Source = "default"
			result.Status = StatusPass
			result.Message = "using default value"
			// Still validate the default value
			value = v.Default
		} else if v.Required {
			// Required but missing
			result.Source = "missing"
			result.Status = StatusFail
			result.Message = "required variable is not set"
			result.Expected = "a value to be set"
			return result
		} else {
			// Optional and not set (no default)
			result.Source = "missing"
			result.Status = StatusPass
			result.Message = "optional variable is not set (no default)"
			return result
		}
	} else {
		result.Value = value
		result.Source = source
	}

	// Type validation (with sensitive value redaction)
	if err := ValidateTypeSensitive(v.Type, value, v.Sensitive); err != nil {
		result.Status = StatusFail
		result.Message = err.Error()
		result.Expected = fmt.Sprintf("a valid %s", v.Type)
		return result
	}

	// AllowedValues validation (works for enum type OR any type with allowed_values specified)
	if len(v.AllowedValues) > 0 {
		if !isValueInAllowed(value, v.AllowedValues) {
			result.Status = StatusFail
			if v.Sensitive {
				result.Message = "value is not one of the allowed values"
			} else {
				result.Message = fmt.Sprintf("value %q is not one of the allowed values", truncateValue(value))
			}
			result.Expected = fmt.Sprintf("one of: %v", v.AllowedValues)
			return result
		}
	}

	// Pattern validation
	if v.Pattern != "" {
		re, err := regexp.Compile(v.Pattern)
		if err != nil {
			// This shouldn't happen if schema.Validate() was called, but handle it defensively
			result.Status = StatusFail
			result.Message = fmt.Sprintf("internal error: invalid pattern '%s'", v.Pattern)
			return result
		}
		if !re.MatchString(value) {
			result.Status = StatusFail
			if v.Sensitive {
				// Don't reveal the value for sensitive vars
				result.Message = fmt.Sprintf("value does not match required pattern: %s", v.Pattern)
			} else {
				result.Message = fmt.Sprintf("value %q does not match required pattern: %s", truncateValue(value), v.Pattern)
			}
			result.Expected = fmt.Sprintf("value matching pattern: %s", v.Pattern)
			return result
		}
	}

	// All checks passed
	if result.Status != StatusPass {
		result.Status = StatusPass
	}
	if result.Message == "" {
		result.Message = "valid"
	}

	return result
}

// isValueInAllowed checks if a value is in the allowed values list.
func isValueInAllowed(value string, allowed []string) bool {
	for _, a := range allowed {
		if value == a {
			return true
		}
	}
	return false
}

// HasFailures returns true if any result has StatusFail.
func HasFailures(results []ValidationResult) bool {
	for _, r := range results {
		if r.Status == StatusFail {
			return true
		}
	}
	return false
}

// HasWarnings returns true if any result has StatusWarn.
func HasWarnings(results []ValidationResult) bool {
	for _, r := range results {
		if r.Status == StatusWarn {
			return true
		}
	}
	return false
}
