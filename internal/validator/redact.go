package validator

import "strings"

// RedactedPlaceholder is the string used to replace sensitive values.
const RedactedPlaceholder = "[REDACTED]"

// RedactValue returns a redacted placeholder if sensitive is true,
// otherwise returns the original value (truncated if too long).
func RedactValue(value string, sensitive bool) string {
	if sensitive {
		return RedactedPlaceholder
	}
	return truncateValue(value)
}

// RedactMessage replaces any occurrence of the value in a message
// with the redacted placeholder if sensitive is true.
func RedactMessage(message, value string, sensitive bool) string {
	if !sensitive || value == "" {
		return message
	}
	return strings.ReplaceAll(message, value, RedactedPlaceholder)
}

// FormatValueForError formats a value for inclusion in error messages,
// respecting sensitivity settings.
func FormatValueForError(value string, sensitive bool) string {
	if sensitive {
		return "value"
	}
	return "value " + `"` + truncateValue(value) + `"`
}
