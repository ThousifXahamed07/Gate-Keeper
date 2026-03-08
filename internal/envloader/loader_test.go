package envloader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper to create a temp .env file with the given content
func createTempEnvFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp .env file: %v", err)
	}
	return envPath
}

// === parseEnvFile tests ===

func TestParseEnvFile_BasicKeyValue(t *testing.T) {
	content := "KEY=value"
	envPath := createTempEnvFile(t, content)

	result, err := parseEnvFile(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["KEY"] != "value" {
		t.Errorf("expected KEY=value, got KEY=%q", result["KEY"])
	}
}

func TestParseEnvFile_DoubleQuotedValue(t *testing.T) {
	content := `KEY="hello world"`
	envPath := createTempEnvFile(t, content)

	result, err := parseEnvFile(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["KEY"] != "hello world" {
		t.Errorf("expected 'hello world', got %q", result["KEY"])
	}
}

func TestParseEnvFile_SingleQuotedValue(t *testing.T) {
	content := `KEY='hello world'`
	envPath := createTempEnvFile(t, content)

	result, err := parseEnvFile(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["KEY"] != "hello world" {
		t.Errorf("expected 'hello world', got %q", result["KEY"])
	}
}

func TestParseEnvFile_CommentsSkipped(t *testing.T) {
	content := `# This is a comment
KEY=value
# Another comment`
	envPath := createTempEnvFile(t, content)

	result, err := parseEnvFile(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 key, got %d", len(result))
	}
	if result["KEY"] != "value" {
		t.Errorf("expected KEY=value, got KEY=%q", result["KEY"])
	}
}

func TestParseEnvFile_EmptyLinesSkipped(t *testing.T) {
	content := `
KEY1=value1

KEY2=value2

`
	envPath := createTempEnvFile(t, content)

	result, err := parseEnvFile(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result))
	}
}

func TestParseEnvFile_ExportPrefix(t *testing.T) {
	content := `export KEY=value
export ANOTHER="quoted"`
	envPath := createTempEnvFile(t, content)

	result, err := parseEnvFile(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["KEY"] != "value" {
		t.Errorf("expected KEY=value, got KEY=%q", result["KEY"])
	}
	if result["ANOTHER"] != "quoted" {
		t.Errorf("expected ANOTHER=quoted, got ANOTHER=%q", result["ANOTHER"])
	}
}

func TestParseEnvFile_InlineCommentsNotStripped(t *testing.T) {
	// Inline comments should NOT be stripped - too error-prone
	content := `KEY=value # this is not a comment`
	envPath := createTempEnvFile(t, content)

	result, err := parseEnvFile(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The entire "value # this is not a comment" should be the value (trailing space trimmed)
	expected := "value # this is not a comment"
	if result["KEY"] != expected {
		t.Errorf("expected %q, got %q", expected, result["KEY"])
	}
}

func TestParseEnvFile_EmptyValue(t *testing.T) {
	content := `KEY=`
	envPath := createTempEnvFile(t, content)

	result, err := parseEnvFile(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["KEY"] != "" {
		t.Errorf("expected empty value, got %q", result["KEY"])
	}
}

func TestParseEnvFile_ValueWithEqualsSign(t *testing.T) {
	content := `DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require`
	envPath := createTempEnvFile(t, content)

	result, err := parseEnvFile(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "postgres://user:pass@host:5432/db?sslmode=require"
	if result["DATABASE_URL"] != expected {
		t.Errorf("expected %q, got %q", expected, result["DATABASE_URL"])
	}
}

func TestParseEnvFile_MalformedLineSkipped(t *testing.T) {
	// Capture stderr to verify warning
	content := `KEY=value
MALFORMED_LINE
ANOTHER=ok`
	envPath := createTempEnvFile(t, content)

	result, err := parseEnvFile(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 keys (malformed skipped), got %d", len(result))
	}
	if result["KEY"] != "value" {
		t.Errorf("expected KEY=value, got KEY=%q", result["KEY"])
	}
	if result["ANOTHER"] != "ok" {
		t.Errorf("expected ANOTHER=ok, got ANOTHER=%q", result["ANOTHER"])
	}
}

func TestParseEnvFile_WindowsLineEndings(t *testing.T) {
	content := "KEY1=value1\r\nKEY2=value2\r\n"
	envPath := createTempEnvFile(t, content)

	result, err := parseEnvFile(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["KEY1"] != "value1" {
		t.Errorf("expected KEY1=value1, got KEY1=%q", result["KEY1"])
	}
	if result["KEY2"] != "value2" {
		t.Errorf("expected KEY2=value2, got KEY2=%q", result["KEY2"])
	}
}

func TestParseEnvFile_NoTrailingNewline(t *testing.T) {
	content := "KEY=value" // No trailing newline
	envPath := createTempEnvFile(t, content)

	result, err := parseEnvFile(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["KEY"] != "value" {
		t.Errorf("expected KEY=value, got KEY=%q", result["KEY"])
	}
}

func TestParseEnvFile_WhitespaceAroundEquals(t *testing.T) {
	content := `KEY = value`
	envPath := createTempEnvFile(t, content)

	result, err := parseEnvFile(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["KEY"] != "value" {
		t.Errorf("expected value, got %q", result["KEY"])
	}
}

func TestParseEnvFile_QuotedValueWithSpaces(t *testing.T) {
	content := `MESSAGE="  leading and trailing spaces  "`
	envPath := createTempEnvFile(t, content)

	result, err := parseEnvFile(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "  leading and trailing spaces  "
	if result["MESSAGE"] != expected {
		t.Errorf("expected %q, got %q", expected, result["MESSAGE"])
	}
}

func TestParseEnvFile_NonExistentFile(t *testing.T) {
	_, err := parseEnvFile("/nonexistent/path/.env")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected IsNotExist error, got %v", err)
	}
}

// === Load tests ===

func TestLoad_OSEnvOverridesEnvFile(t *testing.T) {
	content := `OVERRIDE_TEST=from-env-file`
	envPath := createTempEnvFile(t, content)

	// Set OS env
	os.Setenv("OVERRIDE_TEST", "from-os")
	defer os.Unsetenv("OVERRIDE_TEST")

	data, err := Load(envPath, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Values["OVERRIDE_TEST"] != "from-os" {
		t.Errorf("expected OS value 'from-os', got %q", data.Values["OVERRIDE_TEST"])
	}
	if data.Sources["OVERRIDE_TEST"] != "os" {
		t.Errorf("expected source 'os', got %q", data.Sources["OVERRIDE_TEST"])
	}
}

func TestLoad_MissingEnvFileIsNotError(t *testing.T) {
	// Load from a non-existent .env file should not error
	data, err := Load("/nonexistent/.env", false)
	if err != nil {
		t.Errorf("expected no error for missing .env file, got %v", err)
	}

	// Should still have OS env vars
	if data == nil {
		t.Fatal("expected non-nil EnvData")
	}
}

func TestLoad_SourceTracking(t *testing.T) {
	content := `FROM_FILE=yes`
	envPath := createTempEnvFile(t, content)

	// Ensure we have at least one OS var
	os.Setenv("FROM_OS_TEST", "yes")
	defer os.Unsetenv("FROM_OS_TEST")

	data, err := Load(envPath, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Sources["FROM_FILE"] != "env-file" {
		t.Errorf("expected source 'env-file' for FROM_FILE, got %q", data.Sources["FROM_FILE"])
	}
	if data.Sources["FROM_OS_TEST"] != "os" {
		t.Errorf("expected source 'os' for FROM_OS_TEST, got %q", data.Sources["FROM_OS_TEST"])
	}
}

func TestLoad_SkipEnvFile(t *testing.T) {
	content := `SHOULD_NOT_LOAD=yes`
	envPath := createTempEnvFile(t, content)

	data, err := Load(envPath, true) // skipEnvFile = true
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, exists := data.Values["SHOULD_NOT_LOAD"]; exists {
		t.Error("expected SHOULD_NOT_LOAD to not be loaded when skipEnvFile=true")
	}
}

func TestLoad_EmptyEnvFilePath(t *testing.T) {
	data, err := Load("", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have OS env vars only
	if data == nil {
		t.Fatal("expected non-nil EnvData")
	}
}

func TestLoad_EnvFileOnlyVars(t *testing.T) {
	// Test that vars only in .env file have correct source
	uniqueKey := "UNIQUE_ENV_FILE_VAR_" + strings.ReplaceAll(t.Name(), "/", "_")
	content := uniqueKey + "=from-file"
	envPath := createTempEnvFile(t, content)

	// Make sure this var is NOT in OS env
	os.Unsetenv(uniqueKey)

	data, err := Load(envPath, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Values[uniqueKey] != "from-file" {
		t.Errorf("expected value 'from-file', got %q", data.Values[uniqueKey])
	}
	if data.Sources[uniqueKey] != "env-file" {
		t.Errorf("expected source 'env-file', got %q", data.Sources[uniqueKey])
	}
}

// === parseValue tests ===

func TestParseValue(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"hello world"`, "hello world"},
		{`'hello world'`, "hello world"},
		{`hello`, "hello"},
		{`hello world`, "hello world"},
		{`  hello  `, "hello"},
		{`"  spaces  "`, "  spaces  "},
		{``, ""},
		{`""`, ""},
		{`''`, ""},
		{`"unmatched`, `"unmatched`},
		{`unmatched"`, `unmatched"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseValue(tt.input)
			if got != tt.want {
				t.Errorf("parseValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// === truncateLine tests ===

func TestTruncateLine(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"short", "short"},
		{"exactly forty characters long ok yes!!", "exactly forty characters long ok yes!!"},
		{"this is a very long line that exceeds forty characters", "this is a very long line that exceeds fo..."},
	}

	for _, tt := range tests {
		t.Run(tt.input[:min(10, len(tt.input))], func(t *testing.T) {
			got := truncateLine(tt.input)
			if got != tt.want {
				t.Errorf("truncateLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
