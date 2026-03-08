package validator

import (
	"strings"
	"testing"

	"github.com/ThousifXahamed079/gatekeeper/internal/schema"
)

func TestValidate_AllValid(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "PORT", Type: "port", Required: true},
			{Name: "HOST", Type: "string", Required: true},
			{Name: "DEBUG", Type: "boolean"},
		},
	}
	env := map[string]string{
		"PORT":  "8080",
		"HOST":  "localhost",
		"DEBUG": "true",
	}
	sources := map[string]string{
		"PORT":  "os",
		"HOST":  "os",
		"DEBUG": "os",
	}

	results := Validate(s, env, sources)

	for _, r := range results {
		if r.Status != StatusPass {
			t.Errorf("expected %s to pass, got status %v: %s", r.VarName, r.Status, r.Message)
		}
	}
}

// === Required field tests ===

func TestValidate_RequiredVarPresent(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "REQUIRED_VAR", Type: "string", Required: true},
		},
	}
	env := map[string]string{
		"REQUIRED_VAR": "some-value",
	}
	sources := map[string]string{
		"REQUIRED_VAR": "os",
	}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusPass {
		t.Errorf("expected StatusPass, got %v: %s", results[0].Status, results[0].Message)
	}
}

func TestValidate_MissingRequired(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "REQUIRED_VAR", Type: "string", Required: true},
		},
	}
	env := map[string]string{}
	sources := map[string]string{}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusFail {
		t.Errorf("expected StatusFail, got %v", results[0].Status)
	}
	if results[0].Source != "missing" {
		t.Errorf("expected source 'missing', got %q", results[0].Source)
	}
	if results[0].Message != "required variable is not set" {
		t.Errorf("expected message 'required variable is not set', got %q", results[0].Message)
	}
}

func TestValidate_RequiredMissingWithDefault(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "PORT", Type: "port", Required: true, Default: "3000"},
		},
	}
	env := map[string]string{}
	sources := map[string]string{}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusPass {
		t.Errorf("expected StatusPass, got %v: %s", results[0].Status, results[0].Message)
	}
	if results[0].Source != "default" {
		t.Errorf("expected source 'default', got %q", results[0].Source)
	}
	if results[0].Value != "3000" {
		t.Errorf("expected value '3000', got %q", results[0].Value)
	}
}

func TestValidate_OptionalMissingNoDefault(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "OPTIONAL_VAR", Type: "string", Required: false},
		},
	}
	env := map[string]string{}
	sources := map[string]string{}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusPass {
		t.Errorf("expected StatusPass, got %v: %s", results[0].Status, results[0].Message)
	}
	if results[0].Source != "missing" {
		t.Errorf("expected source 'missing', got %q", results[0].Source)
	}
}

func TestValidate_MissingOptionalWithDefault(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "PORT", Type: "port", Default: "3000"},
		},
	}
	env := map[string]string{}
	sources := map[string]string{}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusPass {
		t.Errorf("expected StatusPass, got %v: %s", results[0].Status, results[0].Message)
	}
	if results[0].Source != "default" {
		t.Errorf("expected source 'default', got %q", results[0].Source)
	}
	if results[0].Value != "3000" {
		t.Errorf("expected value '3000', got %q", results[0].Value)
	}
}

func TestValidate_RequiredPresentEmptyString(t *testing.T) {
	// Empty string "" is a VALID value - it means the var is set
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "EMPTY_VAR", Type: "string", Required: true},
		},
	}
	env := map[string]string{
		"EMPTY_VAR": "",
	}
	sources := map[string]string{
		"EMPTY_VAR": "os",
	}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusPass {
		t.Errorf("expected StatusPass (empty string is valid), got %v: %s", results[0].Status, results[0].Message)
	}
	if results[0].Source != "os" {
		t.Errorf("expected source 'os', got %q", results[0].Source)
	}
}

func TestValidate_OSOverridesEnvFile(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "MY_VAR", Type: "string", Required: true},
		},
	}
	// Simulate OS value winning over env-file
	env := map[string]string{
		"MY_VAR": "os-value",
	}
	sources := map[string]string{
		"MY_VAR": "os",
	}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != "os-value" {
		t.Errorf("expected value 'os-value', got %q", results[0].Value)
	}
	if results[0].Source != "os" {
		t.Errorf("expected source 'os', got %q", results[0].Source)
	}
}

// === Type validation tests ===

func TestValidate_WrongType(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "PORT", Type: "port", Required: true},
		},
	}
	env := map[string]string{
		"PORT": "not-a-port",
	}
	sources := map[string]string{
		"PORT": "os",
	}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusFail {
		t.Errorf("expected StatusFail, got %v", results[0].Status)
	}
}

// === Pattern validation tests ===

func TestValidate_PatternHttpsMatch(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "URL", Type: "string", Required: true, Pattern: "^https://"},
		},
	}
	env := map[string]string{
		"URL": "https://example.com",
	}
	sources := map[string]string{
		"URL": "os",
	}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusPass {
		t.Errorf("expected StatusPass, got %v: %s", results[0].Status, results[0].Message)
	}
}

func TestValidate_PatternHttpsNoMatch(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "URL", Type: "string", Required: true, Pattern: "^https://"},
		},
	}
	env := map[string]string{
		"URL": "http://example.com",
	}
	sources := map[string]string{
		"URL": "os",
	}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusFail {
		t.Errorf("expected StatusFail, got %v", results[0].Status)
	}
}

func TestValidate_PatternCurrencyCode(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"USD", false},
		{"EUR", false},
		{"usd", true},  // lowercase fails
		{"US", true},   // too short
		{"USDD", true}, // too long
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			s := &schema.Schema{
				Version: "1",
				Vars: []schema.Var{
					{Name: "CURRENCY", Type: "string", Required: true, Pattern: "^[A-Z]{3}$"},
				},
			}
			env := map[string]string{
				"CURRENCY": tt.value,
			}
			sources := map[string]string{
				"CURRENCY": "os",
			}

			results := Validate(s, env, sources)

			if len(results) != 1 {
				t.Fatalf("expected 1 result, got %d", len(results))
			}
			if tt.wantErr && results[0].Status != StatusFail {
				t.Errorf("expected StatusFail for value %q, got %v", tt.value, results[0].Status)
			}
			if !tt.wantErr && results[0].Status != StatusPass {
				t.Errorf("expected StatusPass for value %q, got %v: %s", tt.value, results[0].Status, results[0].Message)
			}
		})
	}
}

func TestValidate_PatternSensitiveNoValueInError(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "SECRET", Type: "string", Required: true, Sensitive: true, Pattern: "^secret-"},
		},
	}
	env := map[string]string{
		"SECRET": "not-matching-value",
	}
	sources := map[string]string{
		"SECRET": "os",
	}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusFail {
		t.Errorf("expected StatusFail, got %v", results[0].Status)
	}
	// Sensitive var - error message should NOT contain the actual value
	if strings.Contains(results[0].Message, "not-matching-value") {
		t.Errorf("sensitive var value should not appear in error message: %s", results[0].Message)
	}
	if !strings.Contains(results[0].Message, "does not match required pattern") {
		t.Errorf("error message should mention pattern mismatch: %s", results[0].Message)
	}
}

func TestValidate_PatternNonSensitiveValueInError(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "CODE", Type: "string", Required: true, Sensitive: false, Pattern: "^[A-Z]{3}$"},
		},
	}
	env := map[string]string{
		"CODE": "abc",
	}
	sources := map[string]string{
		"CODE": "os",
	}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusFail {
		t.Errorf("expected StatusFail, got %v", results[0].Status)
	}
	// Non-sensitive var - error message SHOULD contain the actual value
	if !strings.Contains(results[0].Message, "abc") {
		t.Errorf("non-sensitive var value should appear in error message: %s", results[0].Message)
	}
}

func TestValidate_DefaultValuePatternMismatch(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "CODE", Type: "string", Default: "bad", Pattern: "^[A-Z]{3}$"},
		},
	}
	env := map[string]string{}
	sources := map[string]string{}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// Default value "bad" doesn't match pattern "^[A-Z]{3}$"
	if results[0].Status != StatusFail {
		t.Errorf("expected StatusFail (bad default), got %v: %s", results[0].Status, results[0].Message)
	}
}

func TestValidate_NoPattern(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "VAR", Type: "string", Required: true},
		},
	}
	env := map[string]string{
		"VAR": "anything-goes",
	}
	sources := map[string]string{
		"VAR": "os",
	}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusPass {
		t.Errorf("expected StatusPass (no pattern to check), got %v: %s", results[0].Status, results[0].Message)
	}
}

func TestValidate_RegexMismatch(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "CODE", Type: "string", Required: true, Pattern: "^[A-Z]{3}$"},
		},
	}
	env := map[string]string{
		"CODE": "abc123",
	}
	sources := map[string]string{
		"CODE": "os",
	}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusFail {
		t.Errorf("expected StatusFail, got %v", results[0].Status)
	}
}

func TestValidate_PatternValid(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "CODE", Type: "string", Required: true, Pattern: "^[A-Z]{3}$"},
		},
	}
	env := map[string]string{
		"CODE": "ABC",
	}
	sources := map[string]string{
		"CODE": "os",
	}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusPass {
		t.Errorf("expected StatusPass, got %v: %s", results[0].Status, results[0].Message)
	}
}

// === Enum validation tests ===

func TestValidate_EnumMismatch(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "ENV", Type: "enum", Required: true, AllowedValues: []string{"dev", "prod", "staging"}},
		},
	}
	env := map[string]string{
		"ENV": "testing",
	}
	sources := map[string]string{
		"ENV": "os",
	}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusFail {
		t.Errorf("expected StatusFail, got %v", results[0].Status)
	}
}

func TestValidate_EnumValid(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "LOG_LEVEL", Type: "enum", Required: true, AllowedValues: []string{"debug", "info", "warn", "error"}},
		},
	}
	env := map[string]string{
		"LOG_LEVEL": "info",
	}
	sources := map[string]string{
		"LOG_LEVEL": "os",
	}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusPass {
		t.Errorf("expected StatusPass, got %v: %s", results[0].Status, results[0].Message)
	}
}

// === AllowedValues for non-enum types ===

func TestValidate_AllowedValuesWithStringType(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "REGION", Type: "string", Required: true, AllowedValues: []string{"us-east-1", "us-west-2", "eu-west-1"}},
		},
	}

	t.Run("valid value", func(t *testing.T) {
		env := map[string]string{"REGION": "us-east-1"}
		sources := map[string]string{"REGION": "os"}
		results := Validate(s, env, sources)
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		if results[0].Status != StatusPass {
			t.Errorf("expected StatusPass, got %v: %s", results[0].Status, results[0].Message)
		}
	})

	t.Run("invalid value", func(t *testing.T) {
		env := map[string]string{"REGION": "ap-south-1"}
		sources := map[string]string{"REGION": "os"}
		results := Validate(s, env, sources)
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		if results[0].Status != StatusFail {
			t.Errorf("expected StatusFail, got %v", results[0].Status)
		}
		if !strings.Contains(results[0].Message, "not one of the allowed values") {
			t.Errorf("expected message about allowed values, got: %s", results[0].Message)
		}
	})
}

func TestValidate_AllowedValuesWithIntegerType(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "REPLICAS", Type: "integer", Required: true, AllowedValues: []string{"1", "3", "5"}},
		},
	}

	t.Run("valid value", func(t *testing.T) {
		env := map[string]string{"REPLICAS": "3"}
		sources := map[string]string{"REPLICAS": "os"}
		results := Validate(s, env, sources)
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		if results[0].Status != StatusPass {
			t.Errorf("expected StatusPass, got %v: %s", results[0].Status, results[0].Message)
		}
	})

	t.Run("invalid integer but valid type", func(t *testing.T) {
		env := map[string]string{"REPLICAS": "2"}
		sources := map[string]string{"REPLICAS": "os"}
		results := Validate(s, env, sources)
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		// Type validation passes, but AllowedValues fails
		if results[0].Status != StatusFail {
			t.Errorf("expected StatusFail (not in allowed values), got %v", results[0].Status)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		env := map[string]string{"REPLICAS": "not-a-number"}
		sources := map[string]string{"REPLICAS": "os"}
		results := Validate(s, env, sources)
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		// Type validation fails first
		if results[0].Status != StatusFail {
			t.Errorf("expected StatusFail (type validation), got %v", results[0].Status)
		}
		if !strings.Contains(results[0].Message, "not a valid integer") {
			t.Errorf("expected type error message, got: %s", results[0].Message)
		}
	})
}

func TestValidate_AllowedValuesWithPortType(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "PORT", Type: "port", Required: true, AllowedValues: []string{"80", "443", "8080"}},
		},
	}

	t.Run("valid port in allowed", func(t *testing.T) {
		env := map[string]string{"PORT": "443"}
		sources := map[string]string{"PORT": "os"}
		results := Validate(s, env, sources)
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		if results[0].Status != StatusPass {
			t.Errorf("expected StatusPass, got %v: %s", results[0].Status, results[0].Message)
		}
	})

	t.Run("valid port not in allowed", func(t *testing.T) {
		env := map[string]string{"PORT": "3000"}
		sources := map[string]string{"PORT": "os"}
		results := Validate(s, env, sources)
		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}
		if results[0].Status != StatusFail {
			t.Errorf("expected StatusFail, got %v", results[0].Status)
		}
	})
}

func TestValidate_AllowedValuesCaseSensitive(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "ENV", Type: "string", Required: true, AllowedValues: []string{"dev", "prod", "staging"}},
		},
	}

	t.Run("exact case match passes", func(t *testing.T) {
		env := map[string]string{"ENV": "dev"}
		sources := map[string]string{"ENV": "os"}
		results := Validate(s, env, sources)
		if results[0].Status != StatusPass {
			t.Errorf("expected StatusPass, got %v", results[0].Status)
		}
	})

	t.Run("different case fails", func(t *testing.T) {
		env := map[string]string{"ENV": "DEV"}
		sources := map[string]string{"ENV": "os"}
		results := Validate(s, env, sources)
		if results[0].Status != StatusFail {
			t.Errorf("expected StatusFail (case sensitive), got %v", results[0].Status)
		}
	})

	t.Run("mixed case fails", func(t *testing.T) {
		env := map[string]string{"ENV": "Dev"}
		sources := map[string]string{"ENV": "os"}
		results := Validate(s, env, sources)
		if results[0].Status != StatusFail {
			t.Errorf("expected StatusFail (case sensitive), got %v", results[0].Status)
		}
	})
}

func TestValidate_AllowedValuesSensitiveRedacted(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "SECRET_LEVEL", Type: "string", Required: true, Sensitive: true, AllowedValues: []string{"low", "medium", "high"}},
		},
	}

	env := map[string]string{"SECRET_LEVEL": "secret-value"}
	sources := map[string]string{"SECRET_LEVEL": "os"}
	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusFail {
		t.Errorf("expected StatusFail, got %v", results[0].Status)
	}
	// Sensitive var - error message should NOT contain the actual value
	if strings.Contains(results[0].Message, "secret-value") {
		t.Errorf("sensitive var value should not appear in error message: %s", results[0].Message)
	}
}

func TestValidate_TypeValidationSensitiveRedacted(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "SECRET_PORT", Type: "port", Required: true, Sensitive: true},
		},
	}

	env := map[string]string{"SECRET_PORT": "my-secret-invalid-port"}
	sources := map[string]string{"SECRET_PORT": "os"}
	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusFail {
		t.Errorf("expected StatusFail, got %v", results[0].Status)
	}
	// Sensitive var - error message should NOT contain the actual value
	if strings.Contains(results[0].Message, "my-secret-invalid-port") {
		t.Errorf("sensitive var value should not appear in type error message: %s", results[0].Message)
	}
}

// === Sensitive var tests ===

func TestValidate_SensitiveVar(t *testing.T) {
	s := &schema.Schema{
		Version: "1",
		Vars: []schema.Var{
			{Name: "SECRET", Type: "string", Required: true, Sensitive: true},
		},
	}
	env := map[string]string{
		"SECRET": "super-secret-value",
	}
	sources := map[string]string{
		"SECRET": "env-file",
	}

	results := Validate(s, env, sources)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Sensitive {
		t.Error("expected Sensitive=true")
	}
}

// === Helper function tests ===

func TestHasFailures(t *testing.T) {
	results := []ValidationResult{
		{Status: StatusPass},
		{Status: StatusFail},
	}
	if !HasFailures(results) {
		t.Error("HasFailures should return true")
	}

	results = []ValidationResult{
		{Status: StatusPass},
		{Status: StatusWarn},
	}
	if HasFailures(results) {
		t.Error("HasFailures should return false")
	}
}

func TestHasWarnings(t *testing.T) {
	results := []ValidationResult{
		{Status: StatusPass},
		{Status: StatusWarn},
	}
	if !HasWarnings(results) {
		t.Error("HasWarnings should return true")
	}

	results = []ValidationResult{
		{Status: StatusPass},
		{Status: StatusFail},
	}
	if HasWarnings(results) {
		t.Error("HasWarnings should return false")
	}
}
