package validator

import (
	"strings"
	"testing"
)

func TestValidateType(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		value    string
		wantErr  bool
		errMsg   string
	}{
		// string type - always valid
		{"string valid empty", "string", "", false, ""},
		{"string valid text", "string", "hello world", false, ""},
		{"string valid special", "string", "!@#$%^&*()", false, ""},

		// integer type
		{"integer valid zero", "integer", "0", false, ""},
		{"integer valid negative", "integer", "-1", false, ""},
		{"integer valid max", "integer", "2147483647", false, ""},
		{"integer valid positive", "integer", "42", false, ""},
		{"integer invalid float", "integer", "1.5", true, "not a valid integer"},
		{"integer invalid text", "integer", "abc", true, "not a valid integer"},
		{"integer invalid empty", "integer", "", true, "not a valid integer"},

		// float type
		{"float valid integer", "float", "42", false, ""},
		{"float valid decimal", "float", "3.14159", false, ""},
		{"float valid negative", "float", "-0.5", false, ""},
		{"float valid scientific", "float", "1.5e10", false, ""},
		{"float invalid text", "float", "abc", true, "not a valid float"},
		{"float invalid empty", "float", "", true, "not a valid float"},
		{"float invalid mixed", "float", "12.34.56", true, "not a valid float"},

		// boolean type
		{"boolean valid true", "boolean", "true", false, ""},
		{"boolean valid TRUE", "boolean", "TRUE", false, ""},
		{"boolean valid False", "boolean", "False", false, ""},
		{"boolean valid 1", "boolean", "1", false, ""},
		{"boolean valid 0", "boolean", "0", false, ""},
		{"boolean valid yes", "boolean", "yes", false, ""},
		{"boolean valid NO", "boolean", "NO", false, ""},
		{"boolean invalid maybe", "boolean", "maybe", true, "not a valid boolean"},
		{"boolean invalid empty", "boolean", "", true, "not a valid boolean"},
		{"boolean invalid number", "boolean", "2", true, "not a valid boolean"},

		// url type
		{"url valid https", "url", "https://example.com", false, ""},
		{"url valid http with path", "url", "http://example.com/path", false, ""},
		{"url valid ftp", "url", "ftp://files.host/path", false, ""},
		{"url invalid no scheme", "url", "example.com", true, "not a valid URL"},
		{"url invalid empty", "url", "", true, "not a valid URL"},
		{"url invalid just scheme", "url", "https://", true, "not a valid URL"},

		// email type
		{"email valid simple", "email", "test@example.com", false, ""},
		{"email valid plus", "email", "test+tag@example.com", false, ""},
		{"email valid subdomain", "email", "user@mail.example.co.uk", false, ""},
		{"email invalid no at", "email", "notanemail", true, "not a valid email"},
		{"email invalid no domain", "email", "test@", true, "not a valid email"},
		{"email invalid no user", "email", "@example.com", true, "not a valid email"},

		// port type
		{"port valid 1", "port", "1", false, ""},
		{"port valid 8080", "port", "8080", false, ""},
		{"port valid 65535", "port", "65535", false, ""},
		{"port valid 443", "port", "443", false, ""},
		{"port invalid 0", "port", "0", true, "not a valid port"},
		{"port invalid 65536", "port", "65536", true, "not a valid port"},
		{"port invalid negative", "port", "-1", true, "not a valid port"},
		{"port invalid text", "port", "http", true, "not a valid port"},

		// enum type - always valid at this level
		{"enum valid any", "enum", "anything", false, ""},
		{"enum valid empty", "enum", "", false, ""},
		{"enum valid number", "enum", "123", false, ""},

		// filepath type
		{"filepath valid simple", "filepath", "/path/to/file", false, ""},
		{"filepath valid relative", "filepath", "relative/path", false, ""},
		{"filepath valid windows", "filepath", "C:\\Users\\test", false, ""},
		{"filepath invalid empty", "filepath", "", true, "not a valid file path"},
		{"filepath invalid null", "filepath", "path\x00with\x00nulls", true, "not a valid file path"},

		// duration type
		{"duration valid seconds", "duration", "5s", false, ""},
		{"duration valid minutes", "duration", "10m30s", false, ""},
		{"duration valid hours", "duration", "1h", false, ""},
		{"duration valid complex", "duration", "2h30m45s", false, ""},
		{"duration valid milliseconds", "duration", "500ms", false, ""},
		{"duration invalid number only", "duration", "5", true, "not a valid duration"},
		{"duration invalid text", "duration", "abc", true, "not a valid duration"},
		{"duration invalid empty", "duration", "", true, "not a valid duration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateType(tt.typeName, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateType(%q, %q) error = %v, wantErr %v", tt.typeName, tt.value, err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateType(%q, %q) error = %v, want error containing %q", tt.typeName, tt.value, err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateType_UnknownType(t *testing.T) {
	err := ValidateType("unknown", "value")
	if err == nil {
		t.Fatal("ValidateType() expected error for unknown type")
	}
	if !strings.Contains(err.Error(), "unknown type") {
		t.Errorf("error = %q, want to contain 'unknown type'", err.Error())
	}
}

func TestTruncateValue(t *testing.T) {
	short := "short"
	if truncateValue(short) != short {
		t.Errorf("truncateValue(%q) = %q, want %q", short, truncateValue(short), short)
	}

	long := strings.Repeat("a", 100)
	truncated := truncateValue(long)
	if len(truncated) != 53 { // 50 chars + "..."
		t.Errorf("truncateValue() len = %d, want 53", len(truncated))
	}
	if !strings.HasSuffix(truncated, "...") {
		t.Errorf("truncateValue() should end with '...'")
	}
}

func TestValidateTypeSensitive(t *testing.T) {
	tests := []struct {
		name             string
		typeName         string
		value            string
		sensitive        bool
		wantErr          bool
		errContainsValue bool
	}{
		{
			name:             "non-sensitive integer error shows value",
			typeName:         "integer",
			value:            "not-an-int",
			sensitive:        false,
			wantErr:          true,
			errContainsValue: true,
		},
		{
			name:             "sensitive integer error hides value",
			typeName:         "integer",
			value:            "secret-not-an-int",
			sensitive:        true,
			wantErr:          true,
			errContainsValue: false,
		},
		{
			name:             "non-sensitive port error shows value",
			typeName:         "port",
			value:            "invalid-port",
			sensitive:        false,
			wantErr:          true,
			errContainsValue: true,
		},
		{
			name:             "sensitive port error hides value",
			typeName:         "port",
			value:            "secret-port",
			sensitive:        true,
			wantErr:          true,
			errContainsValue: false,
		},
		{
			name:             "valid value non-sensitive no error",
			typeName:         "integer",
			value:            "123",
			sensitive:        false,
			wantErr:          false,
			errContainsValue: false,
		},
		{
			name:             "valid value sensitive no error",
			typeName:         "integer",
			value:            "456",
			sensitive:        true,
			wantErr:          false,
			errContainsValue: false,
		},
		{
			name:             "sensitive email error hides value",
			typeName:         "email",
			value:            "secret@invalid",
			sensitive:        true,
			wantErr:          true,
			errContainsValue: false,
		},
		{
			name:             "sensitive url error hides value",
			typeName:         "url",
			value:            "not-a-valid-url",
			sensitive:        true,
			wantErr:          true,
			errContainsValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTypeSensitive(tt.typeName, tt.value, tt.sensitive)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTypeSensitive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				containsValue := strings.Contains(err.Error(), tt.value)
				if containsValue != tt.errContainsValue {
					if tt.errContainsValue {
						t.Errorf("error should contain value %q, got: %s", tt.value, err.Error())
					} else {
						t.Errorf("error should NOT contain value %q (sensitive), got: %s", tt.value, err.Error())
					}
				}
			}
		})
	}
}
