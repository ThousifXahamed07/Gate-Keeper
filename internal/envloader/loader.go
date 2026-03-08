// Package envloader provides functionality to load environment variables
// from the OS and .env files.
package envloader

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// EnvEntry represents an environment variable with its value and source.
type EnvEntry struct {
	Value  string
	Source string // "os" or "env-file"
}

// EnvData holds environment variables and their sources.
// This is a convenience wrapper around maps for the validator.
type EnvData struct {
	Values  map[string]string // Variable name -> value
	Sources map[string]string // Variable name -> source ("os", "env-file")
}

// Load loads environment variables from the OS and optionally from an .env file.
// OS environment takes precedence over .env file values.
//
// Behavior:
//   - If skipEnvFile is true, only OS environment is loaded
//   - If envFilePath is empty and skipEnvFile is false, no .env file is loaded
//   - If .env file doesn't exist, it's silently skipped (normal in CI)
//   - If .env file exists but can't be parsed, returns an error
//   - OS values always override .env file values
func Load(envFilePath string, skipEnvFile bool) (*EnvData, error) {
	data := &EnvData{
		Values:  make(map[string]string),
		Sources: make(map[string]string),
	}

	// Load from .env file first (lower priority)
	if !skipEnvFile && envFilePath != "" {
		envVars, err := parseEnvFile(envFilePath)
		if err != nil {
			// Only return error if file exists but can't be parsed
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to parse env file: %w", err)
			}
			// File doesn't exist - that's OK, continue with OS env only
		} else {
			for key, value := range envVars {
				data.Values[key] = value
				data.Sources[key] = "env-file"
			}
		}
	}

	// Load from OS environment (higher priority - overwrites .env values)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			data.Values[key] = value
			data.Sources[key] = "os"
		}
	}

	return data, nil
}

// parseEnvFile parses a .env file and returns a map of key-value pairs.
//
// Supported syntax:
//   - KEY=VALUE (basic assignment)
//   - KEY="quoted value" (double quotes stripped)
//   - KEY='quoted value' (single quotes stripped)
//   - KEY=value with spaces (trailing whitespace trimmed)
//   - export KEY=VALUE (export prefix stripped)
//   - Empty lines are skipped
//   - Lines starting with # are treated as comments
//   - Lines without = are skipped with a warning to stderr
//   - Values with = signs are preserved: KEY=postgres://user:pass@host
//   - Inline comments are NOT stripped: KEY=value # comment → "value # comment"
//
// Not supported:
//   - Multiline values
//   - Variable expansion (${VAR} interpolation)
func parseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Handle \r\n line endings (scanner handles \n automatically)
		line = strings.TrimSuffix(line, "\r")

		// Trim leading/trailing whitespace
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip comment lines
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Handle "export KEY=VALUE" syntax
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
			line = strings.TrimSpace(line)
		}

		// Parse KEY=VALUE
		idx := strings.Index(line, "=")
		if idx == -1 {
			// Malformed line - warn and skip
			fmt.Fprintf(os.Stderr, "Warning: .env line %d: no '=' found, skipping: %s\n", lineNum, truncateLine(line))
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := line[idx+1:] // Don't trim value yet - we need to handle quotes first

		// Skip if key is empty
		if key == "" {
			continue
		}

		// Process the value: handle quotes or trim whitespace
		value = parseValue(value)

		result[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading env file: %w", err)
	}

	return result, nil
}

// parseValue processes an env file value, handling quotes and whitespace.
func parseValue(value string) string {
	// Trim leading whitespace only (trailing might be significant in quotes)
	value = strings.TrimLeft(value, " \t")

	if len(value) == 0 {
		return ""
	}

	// Handle double-quoted values
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		return value[1 : len(value)-1]
	}

	// Handle single-quoted values
	if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
		return value[1 : len(value)-1]
	}

	// Unquoted value - trim trailing whitespace
	return strings.TrimRight(value, " \t")
}

// truncateLine truncates a line for display in warnings.
func truncateLine(line string) string {
	if len(line) > 40 {
		return line[:40] + "..."
	}
	return line
}
