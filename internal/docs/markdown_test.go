package docs

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ThousifXahamed079/gatekeeper/internal/schema"
)

func TestOrganizeByGroup(t *testing.T) {
	tests := []struct {
		name           string
		schema         *schema.Schema
		expectedGroups int
		checkFunc      func(t *testing.T, result []GroupedVars)
	}{
		{
			name: "three groups with vars distributed",
			schema: &schema.Schema{
				Version: "1",
				Groups: []schema.Group{
					{Name: "Database", Description: "Database settings"},
					{Name: "API", Description: "API settings"},
					{Name: "Logging", Description: "Logging settings"},
				},
				Vars: []schema.Var{
					{Name: "DB_HOST", Group: "Database"},
					{Name: "API_KEY", Group: "API"},
					{Name: "DB_PORT", Group: "Database"},
					{Name: "LOG_LEVEL", Group: "Logging"},
				},
			},
			expectedGroups: 3,
			checkFunc: func(t *testing.T, result []GroupedVars) {
				// Check Database group (first)
				if result[0].Group.Name != "Database" {
					t.Errorf("expected first group to be Database, got %s", result[0].Group.Name)
				}
				if len(result[0].Vars) != 2 {
					t.Errorf("expected Database to have 2 vars, got %d", len(result[0].Vars))
				}
				// Vars should be in schema order
				if result[0].Vars[0].Name != "DB_HOST" || result[0].Vars[1].Name != "DB_PORT" {
					t.Error("vars not in schema order within Database group")
				}

				// Check API group (second)
				if result[1].Group.Name != "API" {
					t.Errorf("expected second group to be API, got %s", result[1].Group.Name)
				}
				if len(result[1].Vars) != 1 {
					t.Errorf("expected API to have 1 var, got %d", len(result[1].Vars))
				}

				// Check Logging group (third)
				if result[2].Group.Name != "Logging" {
					t.Errorf("expected third group to be Logging, got %s", result[2].Group.Name)
				}
				if len(result[2].Vars) != 1 {
					t.Errorf("expected Logging to have 1 var, got %d", len(result[2].Vars))
				}
			},
		},
		{
			name: "some vars have no group - appear under ungrouped",
			schema: &schema.Schema{
				Version: "1",
				Groups: []schema.Group{
					{Name: "Database", Description: "Database settings"},
				},
				Vars: []schema.Var{
					{Name: "DB_HOST", Group: "Database"},
					{Name: "RANDOM_VAR", Group: ""},
					{Name: "ANOTHER_VAR", Group: ""},
				},
			},
			expectedGroups: 2, // Database + Ungrouped
			checkFunc: func(t *testing.T, result []GroupedVars) {
				if result[0].Group.Name != "Database" {
					t.Errorf("expected first group to be Database, got %s", result[0].Group.Name)
				}
				if result[1].Group != nil {
					t.Error("expected second entry to be ungrouped (nil Group)")
				}
				if len(result[1].Vars) != 2 {
					t.Errorf("expected ungrouped to have 2 vars, got %d", len(result[1].Vars))
				}
			},
		},
		{
			name: "all vars have groups - no ungrouped section",
			schema: &schema.Schema{
				Version: "1",
				Groups: []schema.Group{
					{Name: "Database", Description: "Database settings"},
				},
				Vars: []schema.Var{
					{Name: "DB_HOST", Group: "Database"},
					{Name: "DB_PORT", Group: "Database"},
				},
			},
			expectedGroups: 1,
			checkFunc: func(t *testing.T, result []GroupedVars) {
				if len(result) != 1 {
					t.Errorf("expected 1 group, got %d", len(result))
				}
				if result[0].Group.Name != "Database" {
					t.Errorf("expected group to be Database, got %s", result[0].Group.Name)
				}
			},
		},
		{
			name: "no groups defined - all vars in flat list",
			schema: &schema.Schema{
				Version: "1",
				Groups:  []schema.Group{},
				Vars: []schema.Var{
					{Name: "VAR1"},
					{Name: "VAR2"},
				},
			},
			expectedGroups: 1,
			checkFunc: func(t *testing.T, result []GroupedVars) {
				if result[0].Group != nil {
					t.Error("expected no group (nil) for flat list")
				}
				if len(result[0].Vars) != 2 {
					t.Errorf("expected 2 vars, got %d", len(result[0].Vars))
				}
			},
		},
		{
			name: "group with no vars - still appears with empty vars",
			schema: &schema.Schema{
				Version: "1",
				Groups: []schema.Group{
					{Name: "Empty", Description: "This group has no vars"},
					{Name: "HasVars", Description: "This group has vars"},
				},
				Vars: []schema.Var{
					{Name: "VAR1", Group: "HasVars"},
				},
			},
			expectedGroups: 2,
			checkFunc: func(t *testing.T, result []GroupedVars) {
				// Empty group should still be present
				if result[0].Group.Name != "Empty" {
					t.Errorf("expected first group to be Empty, got %s", result[0].Group.Name)
				}
				if len(result[0].Vars) != 0 {
					t.Errorf("expected Empty group to have 0 vars, got %d", len(result[0].Vars))
				}
				// HasVars group should have the var
				if result[1].Group.Name != "HasVars" {
					t.Errorf("expected second group to be HasVars, got %s", result[1].Group.Name)
				}
				if len(result[1].Vars) != 1 {
					t.Errorf("expected HasVars to have 1 var, got %d", len(result[1].Vars))
				}
			},
		},
		{
			name: "vars within group maintain schema order",
			schema: &schema.Schema{
				Version: "1",
				Groups: []schema.Group{
					{Name: "Config", Description: "Configuration"},
				},
				Vars: []schema.Var{
					{Name: "FIRST", Group: "Config"},
					{Name: "SECOND", Group: "Config"},
					{Name: "THIRD", Group: "Config"},
				},
			},
			expectedGroups: 1,
			checkFunc: func(t *testing.T, result []GroupedVars) {
				vars := result[0].Vars
				if vars[0].Name != "FIRST" || vars[1].Name != "SECOND" || vars[2].Name != "THIRD" {
					t.Error("vars not in schema order")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := organizeByGroup(tt.schema)
			if len(result) != tt.expectedGroups {
				t.Errorf("expected %d groups, got %d", tt.expectedGroups, len(result))
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestGenerateMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		schema   *schema.Schema
		contains []string
		excludes []string
	}{
		{
			name: "basic schema with header",
			schema: &schema.Schema{
				Version: "1",
				Vars: []schema.Var{
					{Name: "API_KEY", Type: "string", Required: true, Description: "The API key"},
				},
			},
			contains: []string{
				"# Environment Variables",
				"Generated by [Gatekeeper]",
				"| Name | Type | Required | Default | Description |",
				"`API_KEY`",
				"`string`",
				"✅ Yes",
				"The API key",
			},
		},
		{
			name: "sensitive variable with lock icon",
			schema: &schema.Schema{
				Version: "1",
				Vars: []schema.Var{
					{Name: "SECRET", Type: "string", Sensitive: true},
				},
			},
			contains: []string{
				"`SECRET` 🔒",
			},
		},
		{
			name: "enum type with values",
			schema: &schema.Schema{
				Version: "1",
				Vars: []schema.Var{
					{Name: "LOG_LEVEL", Type: "enum", AllowedValues: []string{"debug", "info", "warn"}},
				},
			},
			contains: []string{
				"`enum(debug, info, warn)`",
			},
		},
		{
			name: "enum with more than 5 values shows ellipsis",
			schema: &schema.Schema{
				Version: "1",
				Vars: []schema.Var{
					{Name: "ENV", Type: "enum", AllowedValues: []string{"a", "b", "c", "d", "e", "f", "g"}},
				},
			},
			contains: []string{
				"`enum(a, b, c, ...)`",
			},
		},
		{
			name: "default value shown",
			schema: &schema.Schema{
				Version: "1",
				Vars: []schema.Var{
					{Name: "PORT", Type: "port", Default: "8080"},
				},
			},
			contains: []string{
				"`8080`",
			},
		},
		{
			name: "groups with headings",
			schema: &schema.Schema{
				Version: "1",
				Groups: []schema.Group{
					{Name: "Database", Description: "Database configuration"},
				},
				Vars: []schema.Var{
					{Name: "DB_HOST", Type: "string", Group: "Database"},
				},
			},
			contains: []string{
				"## Database",
				"Database configuration",
			},
		},
		{
			name: "ungrouped section when groups exist",
			schema: &schema.Schema{
				Version: "1",
				Groups: []schema.Group{
					{Name: "Database", Description: "Database configuration"},
				},
				Vars: []schema.Var{
					{Name: "DB_HOST", Type: "string", Group: "Database"},
					{Name: "RANDOM", Type: "string"},
				},
			},
			contains: []string{
				"## Database",
				"### Ungrouped",
			},
		},
		{
			name: "no ungrouped heading when no groups defined",
			schema: &schema.Schema{
				Version: "1",
				Groups: []schema.Group{},
				Vars: []schema.Var{
					{Name: "VAR1", Type: "string"},
				},
			},
			excludes: []string{
				"### Ungrouped",
			},
		},
		{
			name: "empty group shows message",
			schema: &schema.Schema{
				Version: "1",
				Groups: []schema.Group{
					{Name: "Empty", Description: "Empty group"},
				},
				Vars: []schema.Var{},
			},
			contains: []string{
				"## Empty",
				"*No variables in this group.*",
			},
		},
		{
			name: "required and non-required formatting",
			schema: &schema.Schema{
				Version: "1",
				Vars: []schema.Var{
					{Name: "REQ_VAR", Type: "string", Required: true},
					{Name: "OPT_VAR", Type: "string", Required: false},
				},
			},
			contains: []string{
				"✅ Yes",
				"| No |",
			},
		},
		{
			name: "missing description shows em dash",
			schema: &schema.Schema{
				Version: "1",
				Vars: []schema.Var{
					{Name: "NO_DESC", Type: "string"},
				},
			},
			contains: []string{
				"| — |",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := GenerateMarkdown(tt.schema, &buf)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()

			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got:\n%s", expected, output)
				}
			}

			for _, excluded := range tt.excludes {
				if strings.Contains(output, excluded) {
					t.Errorf("expected output to NOT contain %q, got:\n%s", excluded, output)
				}
			}
		})
	}
}

func TestFormatName(t *testing.T) {
	tests := []struct {
		name     string
		v        schema.Var
		expected string
	}{
		{
			name:     "regular variable",
			v:        schema.Var{Name: "API_KEY"},
			expected: "`API_KEY`",
		},
		{
			name:     "sensitive variable",
			v:        schema.Var{Name: "SECRET", Sensitive: true},
			expected: "`SECRET` 🔒",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatName(tt.v)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatType(t *testing.T) {
	tests := []struct {
		name     string
		v        schema.Var
		expected string
	}{
		{
			name:     "string type",
			v:        schema.Var{Type: "string"},
			expected: "`string`",
		},
		{
			name:     "enum with few values",
			v:        schema.Var{Type: "enum", AllowedValues: []string{"a", "b", "c"}},
			expected: "`enum(a, b, c)`",
		},
		{
			name:     "enum with exactly 5 values",
			v:        schema.Var{Type: "enum", AllowedValues: []string{"a", "b", "c", "d", "e"}},
			expected: "`enum(a, b, c, d, e)`",
		},
		{
			name:     "enum with more than 5 values",
			v:        schema.Var{Type: "enum", AllowedValues: []string{"a", "b", "c", "d", "e", "f"}},
			expected: "`enum(a, b, c, ...)`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatType(tt.v)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatRequired(t *testing.T) {
	tests := []struct {
		name     string
		required bool
		expected string
	}{
		{name: "required", required: true, expected: "✅ Yes"},
		{name: "not required", required: false, expected: "No"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRequired(schema.Var{Required: tt.required})
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatDefault(t *testing.T) {
	tests := []struct {
		name     string
		def      string
		expected string
	}{
		{name: "with default", def: "8080", expected: "`8080`"},
		{name: "no default", def: "", expected: "—"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDefault(schema.Var{Default: tt.def})
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatDescription(t *testing.T) {
	tests := []struct {
		name     string
		desc     string
		expected string
	}{
		{name: "with description", desc: "The API key", expected: "The API key"},
		{name: "empty description", desc: "", expected: "—"},
		{name: "description with pipe", desc: "A | B", expected: "A \\| B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDescription(schema.Var{Description: tt.desc})
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
