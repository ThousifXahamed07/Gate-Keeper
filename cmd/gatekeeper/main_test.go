package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain builds the binary before running tests
func TestMain(m *testing.M) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "gatekeeper_test", ".")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		panic("failed to build test binary: " + err.Error())
	}

	code := m.Run()

	// Cleanup
	os.Remove("gatekeeper_test")

	os.Exit(code)
}

func runGatekeeper(t *testing.T, args []string, env []string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command("./gatekeeper_test", args...)
	cmd.Env = append(os.Environ(), env...)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("failed to run command: %v", err)
	}

	return outBuf.String(), errBuf.String(), exitCode
}

func createTestSchema(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, ".gatekeeper.yaml")
	if err := os.WriteFile(schemaPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
	return schemaPath
}

func createTestEnvFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create .env: %v", err)
	}
	return envPath
}

// === Exit Code Tests ===

func TestExitCode_Success_AllVarsPass(t *testing.T) {
	schema := `version: "1"
vars:
  - name: TEST_VAR
    type: string
    required: true
`
	schemaPath := createTestSchema(t, schema)

	_, stderr, exitCode := runGatekeeper(t, []string{
		"check",
		"--schema", schemaPath,
		"--no-env-file",
	}, []string{"TEST_VAR=hello"})

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stderr, "✅ Gatekeeper") {
		t.Errorf("expected success summary in stderr, got: %s", stderr)
	}
}

func TestExitCode_Validation_RequiredVarMissing(t *testing.T) {
	schema := `version: "1"
vars:
  - name: REQUIRED_VAR
    type: string
    required: true
`
	schemaPath := createTestSchema(t, schema)

	// Run without setting REQUIRED_VAR
	_, stderr, exitCode := runGatekeeper(t, []string{
		"check",
		"--schema", schemaPath,
		"--no-env-file",
	}, []string{})

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "❌ Gatekeeper") {
		t.Errorf("expected failure summary in stderr, got: %s", stderr)
	}
}

func TestExitCode_SchemaError_MalformedSchema(t *testing.T) {
	// Create a malformed YAML file
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, ".gatekeeper.yaml")
	malformedContent := `version: "1"
vars:
  - name: TEST
    type: unknown_type_that_does_not_exist
`
	if err := os.WriteFile(schemaPath, []byte(malformedContent), 0644); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	_, stderr, exitCode := runGatekeeper(t, []string{
		"check",
		"--schema", schemaPath,
		"--no-env-file",
	}, []string{})

	if exitCode != 2 {
		t.Errorf("expected exit code 2, got %d", exitCode)
	}
	if !strings.Contains(stderr, "schema error") {
		t.Errorf("expected schema error in stderr, got: %s", stderr)
	}
}

func TestExitCode_SchemaError_FileNotFound(t *testing.T) {
	_, stderr, exitCode := runGatekeeper(t, []string{
		"check",
		"--schema", "/nonexistent/schema.yaml",
		"--no-env-file",
	}, []string{})

	if exitCode != 2 {
		t.Errorf("expected exit code 2, got %d", exitCode)
	}
	if !strings.Contains(stderr, "schema error") {
		t.Errorf("expected schema error in stderr, got: %s", stderr)
	}
}

func TestExitCode_Strict_WarningsAsErrors(t *testing.T) {
	// Note: This test assumes we have a way to generate warnings.
	// For now, we test that --strict flag is recognized.
	schema := `version: "1"
vars:
  - name: TEST_VAR
    type: string
    required: true
`
	schemaPath := createTestSchema(t, schema)

	_, _, exitCode := runGatekeeper(t, []string{
		"check",
		"--schema", schemaPath,
		"--no-env-file",
		"--strict",
	}, []string{"TEST_VAR=hello"})

	// With no warnings, should still pass
	if exitCode != 0 {
		t.Errorf("expected exit code 0 (no warnings), got %d", exitCode)
	}
}

// === Output Discipline Tests ===

func TestOutput_ResultsToStdout_SummaryToStderr(t *testing.T) {
	schema := `version: "1"
vars:
  - name: TEST_VAR
    type: string
    required: true
`
	schemaPath := createTestSchema(t, schema)

	stdout, stderr, _ := runGatekeeper(t, []string{
		"check",
		"--schema", schemaPath,
		"--no-env-file",
	}, []string{"TEST_VAR=hello"})

	// Results should be in stdout
	if !strings.Contains(stdout, "TEST_VAR") {
		t.Errorf("expected TEST_VAR in stdout, got: %s", stdout)
	}

	// Summary should be in stderr
	if !strings.Contains(stderr, "Gatekeeper") {
		t.Errorf("expected summary in stderr, got: %s", stderr)
	}
}

func TestOutput_JSONFormat_PipingWorks(t *testing.T) {
	schema := `version: "1"
vars:
  - name: TEST_VAR
    type: string
    required: true
`
	schemaPath := createTestSchema(t, schema)

	stdout, stderr, _ := runGatekeeper(t, []string{
		"check",
		"--schema", schemaPath,
		"--no-env-file",
		"--format", "json",
	}, []string{"TEST_VAR=hello"})

	// stdout should be valid JSON
	if !strings.HasPrefix(strings.TrimSpace(stdout), "{") {
		t.Errorf("expected JSON in stdout, got: %s", stdout)
	}

	// stderr should have the summary (not mixed into JSON)
	if !strings.Contains(stderr, "Gatekeeper") {
		t.Errorf("expected summary in stderr, got: %s", stderr)
	}

	// stdout should NOT contain the summary line
	if strings.Contains(stdout, "Gatekeeper:") {
		t.Errorf("stdout should not contain summary line: %s", stdout)
	}
}

func TestOutput_GitHubFormat(t *testing.T) {
	schema := `version: "1"
vars:
  - name: MISSING_VAR
    type: string
    required: true
`
	schemaPath := createTestSchema(t, schema)

	stdout, _, exitCode := runGatekeeper(t, []string{
		"check",
		"--schema", schemaPath,
		"--no-env-file",
		"--format", "github",
	}, []string{})

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	// GitHub Actions error format
	if !strings.Contains(stdout, "::error::") {
		t.Errorf("expected GitHub Actions error format in stdout, got: %s", stdout)
	}
}

// === Summary Format Tests ===

func TestSummary_SuccessFormat(t *testing.T) {
	schema := `version: "1"
vars:
  - name: VAR1
    type: string
    required: true
  - name: VAR2
    type: string
    required: true
`
	schemaPath := createTestSchema(t, schema)

	_, stderr, _ := runGatekeeper(t, []string{
		"check",
		"--schema", schemaPath,
		"--no-env-file",
	}, []string{"VAR1=a", "VAR2=b"})

	if !strings.Contains(stderr, "✅ Gatekeeper: all 2 variable(s) validated successfully") {
		t.Errorf("expected success summary format, got: %s", stderr)
	}
}

func TestSummary_FailureFormat(t *testing.T) {
	schema := `version: "1"
vars:
  - name: REQUIRED_VAR
    type: string
    required: true
  - name: ANOTHER_REQUIRED
    type: string
    required: true
`
	schemaPath := createTestSchema(t, schema)

	_, stderr, _ := runGatekeeper(t, []string{
		"check",
		"--schema", schemaPath,
		"--no-env-file",
	}, []string{})

	// Should have "X error(s), Y warning(s) in Z variable(s)"
	if !strings.Contains(stderr, "❌ Gatekeeper:") {
		t.Errorf("expected failure summary format, got: %s", stderr)
	}
	if !strings.Contains(stderr, "error(s)") {
		t.Errorf("expected 'error(s)' in summary, got: %s", stderr)
	}
}

// === Version and Help Tests ===

func TestVersion(t *testing.T) {
	stdout, _, exitCode := runGatekeeper(t, []string{"version"}, []string{})

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "gatekeeper") {
		t.Errorf("expected version output, got: %s", stdout)
	}
}

func TestHelp(t *testing.T) {
	stdout, _, exitCode := runGatekeeper(t, []string{"help"}, []string{})

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("expected help output, got: %s", stdout)
	}
}

func TestUnknownCommand(t *testing.T) {
	_, stderr, exitCode := runGatekeeper(t, []string{"unknown"}, []string{})

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "Unknown command") {
		t.Errorf("expected 'Unknown command' in stderr, got: %s", stderr)
	}
}
