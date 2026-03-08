package reporter

import (
	"encoding/json"
	"io"

	"github.com/ThousifXahamed079/gatekeeper/internal/validator"
)

// JSONReporter renders validation results as structured JSON.
type JSONReporter struct{}

// JSON output structures
type jsonOutput struct {
	Version string      `json:"version"`
	Schema  string      `json:"schema"`
	Passed  bool        `json:"passed"`
	Summary jsonSummary `json:"summary"`
	Results []jsonResult `json:"results"`
}

type jsonSummary struct {
	Total    int `json:"total"`
	Passed   int `json:"passed"`
	Failed   int `json:"failed"`
	Warnings int `json:"warnings"`
}

type jsonResult struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Source   string `json:"source"`
	Value    string `json:"value"`
	Message  string `json:"message"`
}

// Render writes the JSON report to w.
func (j *JSONReporter) Render(report Report, w io.Writer) error {
	output := jsonOutput{
		Version: report.Version,
		Schema:  report.SchemaFile,
		Passed:  report.Passed(),
		Summary: jsonSummary{
			Total:    report.TotalVars,
			Passed:   report.PassCount,
			Failed:   report.ErrorCount,
			Warnings: report.WarningCount,
		},
		Results: make([]jsonResult, 0, len(report.Results)),
	}

	for _, r := range report.Results {
		jr := jsonResult{
			Name:     r.VarName,
			Status:   statusToString(r.Status),
			Type:     r.VarType,
			Required: r.Required,
			Source:   r.Source,
			Value:    sanitizeValue(r.Value, r.Sensitive),
			Message:  r.Message,
		}
		
		// Default type to "string" if empty
		if jr.Type == "" {
			jr.Type = "string"
		}
		
		output.Results = append(output.Results, jr)
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	if err != nil {
		return err
	}

	// Trailing newline
	_, err = w.Write([]byte("\n"))
	return err
}

// statusToString converts a Status enum to its string representation.
func statusToString(s validator.Status) string {
	switch s {
	case validator.StatusPass:
		return "pass"
	case validator.StatusFail:
		return "fail"
	case validator.StatusWarn:
		return "warn"
	default:
		return "unknown"
	}
}

// sanitizeValue ensures sensitive values are redacted.
func sanitizeValue(value string, sensitive bool) string {
	if sensitive && value != "***REDACTED***" {
		return "***REDACTED***"
	}
	return value
}
