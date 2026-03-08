package validator

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TypeValidator is a function that validates a string value for a specific type.
type TypeValidator func(value string) error

// emailRegex is compiled once at package init for performance.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// truncateValue truncates a value to 50 characters for error messages.
func truncateValue(value string) string {
	if len(value) > 50 {
		return value[:50] + "..."
	}
	return value
}

// typeValidators maps type names to their validation functions.
var typeValidators = map[string]TypeValidator{
	"string": func(value string) error {
		// Any non-empty string is valid
		return nil
	},

	"integer": func(value string) error {
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("value %q is not a valid integer", truncateValue(value))
		}
		return nil
	},

	"float": func(value string) error {
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("value %q is not a valid float", truncateValue(value))
		}
		return nil
	},

	"boolean": func(value string) error {
		lower := strings.ToLower(value)
		switch lower {
		case "true", "false", "1", "0", "yes", "no":
			return nil
		default:
			return fmt.Errorf("value %q is not a valid boolean, expected: true, false, 1, 0, yes, no", truncateValue(value))
		}
	},

	"url": func(value string) error {
		u, err := url.Parse(value)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("value %q is not a valid URL", truncateValue(value))
		}
		return nil
	},

	"email": func(value string) error {
		if !emailRegex.MatchString(value) {
			return fmt.Errorf("value %q is not a valid email address", truncateValue(value))
		}
		return nil
	},

	"port": func(value string) error {
		port, err := strconv.Atoi(value)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("value %q is not a valid port (must be 1-65535)", truncateValue(value))
		}
		return nil
	},

	"enum": func(value string) error {
		// Enum validation happens at a higher level with AllowedValues from schema.
		// The type validator itself always returns nil.
		return nil
	},

	"filepath": func(value string) error {
		if value == "" || strings.ContainsRune(value, '\x00') {
			return fmt.Errorf("value is not a valid file path")
		}
		return nil
	},

	"duration": func(value string) error {
		if _, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("value %q is not a valid duration (use Go duration format: 5s, 10m, 1h30m)", truncateValue(value))
		}
		return nil
	},
}

// ValidateType validates a value against the rules for a specific type.
func ValidateType(typeName string, value string) error {
	validator, ok := typeValidators[typeName]
	if !ok {
		return fmt.Errorf("unknown type: %s", typeName)
	}
	return validator(value)
}

// ValidateTypeSensitive validates a value against the rules for a specific type,
// and redacts the value in error messages if sensitive is true.
func ValidateTypeSensitive(typeName string, value string, sensitive bool) error {
	err := ValidateType(typeName, value)
	if err != nil && sensitive {
		// Redact the value from the error message
		return fmt.Errorf("%s", RedactMessage(err.Error(), value, sensitive))
	}
	return err
}
