package schema

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseFile reads and parses a schema file from the given path.
// It validates the schema before returning it.
func ParseFile(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("schema file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	return ParseBytes(data)
}

// ParseBytes parses a schema from raw bytes.
// It validates the schema before returning it.
func ParseBytes(data []byte) (*Schema, error) {
	var s Schema
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse schema file: %w", err)
	}

	if err := s.Validate(); err != nil {
		return nil, err
	}

	return &s, nil
}

// schemaFileNames defines the priority order for schema file discovery.
var schemaFileNames = []string{
	".gatekeeper.yaml",
	".gatekeeper.yml",
	"gatekeeper.yaml",
}

// FindSchemaFile searches for a schema file in the current directory.
// It returns the path to the first found file.
func FindSchemaFile() (string, error) {
	for _, name := range schemaFileNames {
		if _, err := os.Stat(name); err == nil {
			return name, nil
		}
	}
	return "", errors.New("no schema file found. Create a .gatekeeper.yaml or run 'gatekeeper init'")
}
