package reporter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ThousifXahamed079/gatekeeper/internal/validator"
)

// === NewReporter tests ===

func TestNewReporter_Text(t *testing.T) {
	r, err := NewReporter("text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := r.(*TextReporter); !ok {
		t.Errorf("expected *TextReporter, got %T", r)
	}
}

func TestNewReporter_JSON(t *testing.T) {
	r, err := NewReporter("json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := r.(*JSONReporter); !ok {
		t.Errorf("expected *JSONReporter, got %T", r)
	}
}

func TestNewReporter_GitHub(t *testing.T) {
	r, err := NewReporter("github")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := r.(*GitHubReporter); !ok {
		t.Errorf("expected *GitHubReporter, got %T", r)
	}
}

func TestNewReporter_Unknown(t *testing.T) {
	_, err := NewReporter("xml")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
	if !strings.Contains(err.Error(), "unknown output format") {
		t.Errorf("expected 'unknown output format' in error, got: %v", err)
	}
}

func TestNewReporter_Empty(t *testing.T) {
	_, err := NewReporter("")
	if err == nil {
		t.Fatal("expected error for empty format")
	}
}

// === NewReport tests ===

func TestNewReport_Counts(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusPass},
		{VarName: "VAR2", Status: validator.StatusPass},
		{VarName: "VAR3", Status: validator.StatusPass},
		{VarName: "VAR4", Status: validator.StatusFail},
		{VarName: "VAR5", Status: validator.StatusWarn},
	}

	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	if report.PassCount != 3 {
		t.Errorf("expected PassCount=3, got %d", report.PassCount)
	}
	if report.ErrorCount != 1 {
		t.Errorf("expected ErrorCount=1, got %d", report.ErrorCount)
	}
	if report.WarningCount != 1 {
		t.Errorf("expected WarningCount=1, got %d", report.WarningCount)
	}
	if report.TotalVars != 5 {
		t.Errorf("expected TotalVars=5, got %d", report.TotalVars)
	}
}

func TestNewReport_AllPass_Passed(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusPass},
		{VarName: "VAR2", Status: validator.StatusPass},
	}

	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	if !report.Passed() {
		t.Error("expected Passed() to return true")
	}
}

func TestNewReport_OneFail_NotPassed(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusPass},
		{VarName: "VAR2", Status: validator.StatusFail},
	}

	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	if report.Passed() {
		t.Error("expected Passed() to return false")
	}
}

func TestNewReport_Empty_Passed(t *testing.T) {
	results := []validator.ValidationResult{}

	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	if report.TotalVars != 0 {
		t.Errorf("expected TotalVars=0, got %d", report.TotalVars)
	}
	if !report.Passed() {
		t.Error("expected Passed() to return true for empty results")
	}
}

// === TextReporter tests ===

func TestTextReporter_AllPass(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusPass, Value: "value1", VarType: "string"},
		{VarName: "VAR2", Status: validator.StatusPass, Value: "value2", VarType: "port"},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS environment", results)

	var buf bytes.Buffer
	reporter := &TextReporter{}
	err := reporter.Render(report, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "PASSED") {
		t.Error("expected 'PASSED' in output")
	}
	// Should not have detail lines (no failures)
	if strings.Contains(output, "✗ VAR1") {
		t.Error("should not have failure detail lines when all pass")
	}
}

func TestTextReporter_Mixed(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusPass, Value: "ok", VarType: "string"},
		{VarName: "VAR2", Status: validator.StatusFail, Message: "missing required variable", VarType: "string"},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &TextReporter{}
	err := reporter.Render(report, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "FAILED") {
		t.Error("expected 'FAILED' in output")
	}
	if !strings.Contains(output, "✗") {
		t.Error("expected failure detail line with ✗")
	}
}

func TestTextReporter_SensitiveVar(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "SECRET", Status: validator.StatusPass, Value: "super-secret", Sensitive: true, VarType: "string"},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &TextReporter{}
	err := reporter.Render(report, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "super-secret") {
		t.Error("sensitive value should not appear in output")
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Error("expected '[REDACTED]' for sensitive value")
	}
}

func TestTextReporter_ZeroResults(t *testing.T) {
	results := []validator.ValidationResult{}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &TextReporter{}
	err := reporter.Render(report, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No variables defined") {
		t.Error("expected 'No variables defined' message")
	}
}

func TestTextReporter_NoANSIInBuffer(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusPass, Value: "ok", VarType: "string"},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &TextReporter{}
	err := reporter.Render(report, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// bytes.Buffer is not a terminal, so no ANSI codes
	if strings.Contains(output, "\033[") {
		t.Error("should not have ANSI codes when writing to bytes.Buffer")
	}
}

func TestTextReporter_VarNameAlignment(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "A", Status: validator.StatusPass, Value: "ok", VarType: "string"},
		{VarName: "VERY_LONG_VARIABLE_NAME", Status: validator.StatusPass, Value: "ok", VarType: "string"},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &TextReporter{}
	err := reporter.Render(report, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Just check it renders without error - alignment is visual
	output := buf.String()
	if !strings.Contains(output, "A") {
		t.Error("expected short var name in output")
	}
	if !strings.Contains(output, "VERY_LONG_VARIABLE_NAME") {
		t.Error("expected long var name in output")
	}
}

// === JSONReporter tests ===

func TestJSONReporter_ValidJSON(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusPass, Value: "ok", VarType: "string", Source: "os"},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &JSONReporter{}
	err := reporter.Render(report, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output jsonOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}

func TestJSONReporter_FieldValues(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusPass, Value: "val1", VarType: "string", Source: "os"},
		{VarName: "VAR2", Status: validator.StatusFail, Message: "error", VarType: "port", Source: "env-file"},
	}
	report := NewReport("2.0.0", "test.yaml", "test source", results)

	var buf bytes.Buffer
	reporter := &JSONReporter{}
	err := reporter.Render(report, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output jsonOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if output.Version != "2.0.0" {
		t.Errorf("expected version '2.0.0', got %q", output.Version)
	}
	if output.Schema != "test.yaml" {
		t.Errorf("expected schema 'test.yaml', got %q", output.Schema)
	}
	if output.Summary.Total != 2 {
		t.Errorf("expected total=2, got %d", output.Summary.Total)
	}
	if output.Summary.Passed != 1 {
		t.Errorf("expected passed=1, got %d", output.Summary.Passed)
	}
	if output.Summary.Failed != 1 {
		t.Errorf("expected failed=1, got %d", output.Summary.Failed)
	}
}

func TestJSONReporter_AllPass_PassedTrue(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusPass},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &JSONReporter{}
	if err := reporter.Render(report, &buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	var output jsonOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !output.Passed {
		t.Error("expected passed=true")
	}
}

func TestJSONReporter_OneFail_PassedFalse(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusFail},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &JSONReporter{}
	if err := reporter.Render(report, &buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	var output jsonOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if output.Passed {
		t.Error("expected passed=false")
	}
}

func TestJSONReporter_SensitiveRedacted(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "SECRET", Status: validator.StatusPass, Value: "actual-secret", Sensitive: true},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &JSONReporter{}
	if err := reporter.Render(report, &buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	var output jsonOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if output.Results[0].Value != "***REDACTED***" {
		t.Errorf("expected '***REDACTED***', got %q", output.Results[0].Value)
	}
}

func TestJSONReporter_EmptyResults_NotNull(t *testing.T) {
	results := []validator.ValidationResult{}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &JSONReporter{}
	if err := reporter.Render(report, &buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.String()
	// Should have "results": [] not "results": null
	if strings.Contains(output, `"results": null`) {
		t.Error("results should be [] not null")
	}
	if !strings.Contains(output, `"results": []`) {
		t.Error("expected empty array for results")
	}
}

func TestJSONReporter_StatusStrings(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusPass},
		{VarName: "VAR2", Status: validator.StatusFail},
		{VarName: "VAR3", Status: validator.StatusWarn},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &JSONReporter{}
	if err := reporter.Render(report, &buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	var output jsonOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	expected := []string{"pass", "fail", "warn"}
	for i, exp := range expected {
		if output.Results[i].Status != exp {
			t.Errorf("expected status %q, got %q", exp, output.Results[i].Status)
		}
	}
}

func TestJSONReporter_NoTextOutsideJSON(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusPass},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &JSONReporter{}
	if err := reporter.Render(report, &buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if !strings.HasPrefix(output, "{") {
		t.Error("JSON should start with {")
	}
	if !strings.HasSuffix(output, "}") {
		t.Error("JSON should end with }")
	}
}

// === GitHubReporter tests ===

func TestGitHubReporter_FailAndWarn(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusFail, Message: "missing"},
		{VarName: "VAR2", Status: validator.StatusWarn, Message: "deprecated"},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &GitHubReporter{}
	err := reporter.Render(report, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	errorLines := 0
	warningLines := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "::error title=Gatekeeper::") {
			errorLines++
		}
		if strings.HasPrefix(line, "::warning title=Gatekeeper::") {
			warningLines++
		}
	}

	if errorLines != 1 {
		t.Errorf("expected 1 error annotation, got %d", errorLines)
	}
	if warningLines != 1 {
		t.Errorf("expected 1 warning annotation, got %d", warningLines)
	}
}

func TestGitHubReporter_AllPass(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusPass},
		{VarName: "VAR2", Status: validator.StatusPass},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &GitHubReporter{}
	if err := reporter.Render(report, &buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.String()

	if strings.Contains(output, "::error") {
		t.Error("should not have ::error:: when all pass")
	}
	if strings.Contains(output, "::warning") {
		t.Error("should not have ::warning:: when all pass")
	}
	if !strings.Contains(output, "::notice::") {
		t.Error("expected ::notice:: for all pass")
	}
}

func TestGitHubReporter_AnnotationFormat(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "MY_VAR", Status: validator.StatusFail, Message: "is missing"},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &GitHubReporter{}
	if err := reporter.Render(report, &buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "::error title=Gatekeeper::MY_VAR: is missing") {
		t.Errorf("unexpected annotation format: %s", output)
	}
}

func TestGitHubReporter_EscapeNewline(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusFail, Message: "line1\nline2"},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &GitHubReporter{}
	if err := reporter.Render(report, &buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "%0A") {
		t.Error("newline should be escaped as %0A")
	}
	if strings.Contains(output, "\nline2") {
		t.Error("literal newline should not appear in message")
	}
}

func TestGitHubReporter_SummaryIsLast(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "VAR1", Status: validator.StatusFail, Message: "error"},
	}
	report := NewReport("1.0.0", "schema.yaml", "OS", results)

	var buf bytes.Buffer
	reporter := &GitHubReporter{}
	if err := reporter.Render(report, &buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	lastLine := lines[len(lines)-1]

	if !strings.HasPrefix(lastLine, "::error::Gatekeeper validation") {
		t.Errorf("last line should be summary, got: %s", lastLine)
	}
}

// === Integration test ===

func TestIntegration_AllReporters(t *testing.T) {
	results := []validator.ValidationResult{
		{VarName: "PORT", Status: validator.StatusPass, Value: "8080", VarType: "port", Required: true, Source: "os"},
		{VarName: "HOST", Status: validator.StatusPass, Value: "localhost", VarType: "string", Required: true, Source: "env-file"},
		{VarName: "DB_URL", Status: validator.StatusFail, Message: "required variable is not set", VarType: "url", Required: true, Source: "missing"},
		{VarName: "DEBUG", Status: validator.StatusFail, Message: "invalid boolean", VarType: "boolean", Required: false, Source: "os"},
		{VarName: "TIMEOUT", Status: validator.StatusWarn, Message: "deprecated, use CONNECT_TIMEOUT", VarType: "duration", Required: false, Source: "env-file"},
	}
	report := NewReport("0.1.0", ".gatekeeper.yaml", "OS environment + .env", results)

	formats := []string{"text", "json", "github"}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			reporter, err := NewReporter(format)
			if err != nil {
				t.Fatalf("failed to create reporter: %v", err)
			}

			var buf bytes.Buffer
			err = reporter.Render(report, &buf)
			if err != nil {
				t.Fatalf("render failed: %v", err)
			}

			output := buf.String()
			if output == "" {
				t.Error("output should not be empty")
			}

			// GitHub reporter only outputs failures and warnings, not passes
			// Text and JSON output all vars
			if format != "github" {
				for _, r := range results {
					if !strings.Contains(output, r.VarName) {
						t.Errorf("expected %s in output", r.VarName)
					}
				}
			} else {
				// For GitHub, check that failures and warnings are present
				for _, r := range results {
					if r.Status == validator.StatusFail || r.Status == validator.StatusWarn {
						if !strings.Contains(output, r.VarName) {
							t.Errorf("expected %s in output", r.VarName)
						}
					}
				}
			}
		})
	}
}

// === escapeGitHubMessage tests ===

func TestEscapeGitHubMessage(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"line1\nline2", "line1%0Aline2"},
		{"with\rcarriage", "with%0Dcarriage"},
		{"100%", "100%25"},
		{"multi\r\nline", "multi%0D%0Aline"},
		{"%\n%", "%25%0A%25"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapeGitHubMessage(tt.input)
			if got != tt.want {
				t.Errorf("escapeGitHubMessage(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
