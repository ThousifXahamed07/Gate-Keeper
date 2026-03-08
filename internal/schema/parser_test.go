package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFile_ValidComplete(t *testing.T) {
	content := `
version: "1"
groups:
  - name: database
    description: Database configuration
  - name: app
    description: Application settings
vars:
  - name: DATABASE_URL
    type: url
    required: true
    description: The database connection string
    sensitive: true
    group: database
  - name: PORT
    type: port
    default: "8080"
    example: "3000"
  - name: LOG_LEVEL
    type: enum
    allowed_values: [debug, info, warn, error]
    group: app
  - name: TIMEOUT
    type: duration
    pattern: "^[0-9]+[smh]$"
  - name: DEBUG
    type: boolean
  - name: MAX_CONNECTIONS
    type: integer
  - name: RATE_LIMIT
    type: float
  - name: CONFIG_PATH
    type: filepath
  - name: ADMIN_EMAIL
    type: email
  - name: APP_NAME
    type: string
`
	dir := t.TempDir()
	path := filepath.Join(dir, ".gatekeeper.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	schema, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if schema.Version != "1" {
		t.Errorf("Version = %q, want %q", schema.Version, "1")
	}
	if len(schema.Groups) != 2 {
		t.Errorf("len(Groups) = %d, want 2", len(schema.Groups))
	}
	if len(schema.Vars) != 10 {
		t.Errorf("len(Vars) = %d, want 10", len(schema.Vars))
	}
}

func TestParseFile_MinimalSchema(t *testing.T) {
	content := `
version: "1"
vars:
  - name: MY_VAR
    type: string
`
	dir := t.TempDir()
	path := filepath.Join(dir, ".gatekeeper.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	schema, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if schema.Version != "1" {
		t.Errorf("Version = %q, want %q", schema.Version, "1")
	}
	if len(schema.Vars) != 1 {
		t.Errorf("len(Vars) = %d, want 1", len(schema.Vars))
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/path/.gatekeeper.yaml")
	if err == nil {
		t.Fatal("ParseFile() expected error for missing file")
	}
	if !strings.Contains(err.Error(), "schema file not found") {
		t.Errorf("error = %q, want to contain 'schema file not found'", err.Error())
	}
}

func TestParseFile_InvalidYAML(t *testing.T) {
	content := `
version: "1"
vars:
  - name: BAD
    type: [this is not valid yaml
`
	dir := t.TempDir()
	path := filepath.Join(dir, ".gatekeeper.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("ParseFile() expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "failed to parse schema file") {
		t.Errorf("error = %q, want to contain 'failed to parse schema file'", err.Error())
	}
}

func TestParseFile_ValidationError(t *testing.T) {
	content := `
version: "1"
vars:
  - name: MY_VAR
    type: unknown_type
`
	dir := t.TempDir()
	path := filepath.Join(dir, ".gatekeeper.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("ParseFile() expected validation error")
	}
	if !strings.Contains(err.Error(), "unknown type 'unknown_type'") {
		t.Errorf("error = %q, want to contain \"unknown type 'unknown_type'\"", err.Error())
	}
}

func TestParseBytes_Valid(t *testing.T) {
	content := []byte(`
version: "1"
vars:
  - name: TEST_VAR
    type: string
`)
	schema, err := ParseBytes(content)
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}
	if schema.Version != "1" {
		t.Errorf("Version = %q, want %q", schema.Version, "1")
	}
}

func TestFindSchemaFile_Primary(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	if err := os.WriteFile(".gatekeeper.yaml", []byte("version: \"1\"\nvars: []"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	path, err := FindSchemaFile()
	if err != nil {
		t.Fatalf("FindSchemaFile() error = %v", err)
	}
	if path != ".gatekeeper.yaml" {
		t.Errorf("path = %q, want %q", path, ".gatekeeper.yaml")
	}
}

func TestFindSchemaFile_Fallback(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	if err := os.WriteFile(".gatekeeper.yml", []byte("version: \"1\"\nvars: []"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	path, err := FindSchemaFile()
	if err != nil {
		t.Fatalf("FindSchemaFile() error = %v", err)
	}
	if path != ".gatekeeper.yml" {
		t.Errorf("path = %q, want %q", path, ".gatekeeper.yml")
	}
}

func TestFindSchemaFile_NoFile(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	_, err := FindSchemaFile()
	if err == nil {
		t.Fatal("FindSchemaFile() expected error when no file exists")
	}
	if !strings.Contains(err.Error(), "no schema file found") {
		t.Errorf("error = %q, want to contain 'no schema file found'", err.Error())
	}
}
