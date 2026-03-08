package schema

import (
	"errors"
	"fmt"
	"regexp"
)

// Schema represents the root of a .gatekeeper.yaml configuration.
type Schema struct {
	Version string  `yaml:"version"` // Schema version, must be "1"
	Groups  []Group `yaml:"groups"`  // Optional grouping of vars
	Vars    []Var   `yaml:"vars"`    // List of environment variable definitions
}

// Group allows for logical grouping of environment variables in the configuration.
type Group struct {
	Name        string `yaml:"name"`        // Group identifier
	Description string `yaml:"description"` // Human-readable description
}

// Var defines the constraints and metadata for a single environment variable.
type Var struct {
	Name          string   `yaml:"name"`           // ENV_VAR_NAME (required)
	Type          string   `yaml:"type"`           // One of the built-in types (required)
	Required      bool     `yaml:"required"`       // Whether the var must be set
	Default       string   `yaml:"default"`        // Default value if not set
	Description   string   `yaml:"description"`    // Human-readable description
	Pattern       string   `yaml:"pattern"`        // Regex pattern to validate against
	Example       string   `yaml:"example"`        // Example value for docs
	Sensitive     bool     `yaml:"sensitive"`      // If true, redact in all output
	AllowedValues []string `yaml:"allowed_values"` // For enum type
	Group         string   `yaml:"group"`          // Which group this var belongs to
}

// Supported primitive types for configuration values.
const (
	TypeString   = "string"
	TypeInteger  = "integer"
	TypeFloat    = "float"
	TypeBoolean  = "boolean"
	TypeURL      = "url"
	TypeEmail    = "email"
	TypePort     = "port"
	TypeEnum     = "enum"
	TypeFilepath = "filepath"
	TypeDuration = "duration"
)

// SupportedTypes holds the set of all valid variable types.
var SupportedTypes = map[string]bool{
	TypeString:   true,
	TypeInteger:  true,
	TypeFloat:    true,
	TypeBoolean:  true,
	TypeURL:      true,
	TypeEmail:    true,
	TypePort:     true,
	TypeEnum:     true,
	TypeFilepath: true,
	TypeDuration: true,
}

// Validate checks if the schema is structurally valid according to rules.
func (s *Schema) Validate() error {
	if s.Version != "1" {
		return fmt.Errorf("invalid schema version: %q, must be \"1\"", s.Version)
	}

	groupMap := make(map[string]bool)
	for _, g := range s.Groups {
		groupMap[g.Name] = true
	}

	seenVars := make(map[string]bool)
	for _, v := range s.Vars {
		if v.Name == "" {
			return errors.New("var has an empty name")
		}

		if seenVars[v.Name] {
			return fmt.Errorf("duplicate var name found: '%s'", v.Name)
		}
		seenVars[v.Name] = true

		if !SupportedTypes[v.Type] {
			return fmt.Errorf("var '%s': unknown type '%s', supported types are: string, integer, float, boolean, url, email, port, enum, filepath, duration", v.Name, v.Type)
		}

		if v.Type == TypeEnum && len(v.AllowedValues) == 0 {
			return fmt.Errorf("var '%s': type is 'enum' but allowed_values is empty", v.Name)
		}

		if v.Group != "" && !groupMap[v.Group] {
			return fmt.Errorf("var '%s': references undefined group '%s'", v.Name, v.Group)
		}

		if v.Pattern != "" {
			if _, err := regexp.Compile(v.Pattern); err != nil {
				return fmt.Errorf("var '%s': invalid regex pattern: %v", v.Name, err)
			}
		}
	}

	return nil
}
