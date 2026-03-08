// Package reporter provides output formatting for validation results.
package reporter

import (
	"fmt"
	"io"

	"github.com/ThousifXahamed079/gatekeeper/internal/validator"
)

// Report contains all data needed to render validation results.
type Report struct {
	Version      string                     // CLI version, e.g. "0.1.0"
	SchemaFile   string                     // Path to schema file used
	EnvSource    string                     // Human-readable source, e.g. "OS environment + .env"
	Results      []validator.ValidationResult // Validation results
	ErrorCount   int                        // Count of StatusFail results
	WarningCount int                        // Count of StatusWarn results
	PassCount    int                        // Count of StatusPass results
	TotalVars    int                        // len(Results)
}

// NewReport creates a Report from validation results.
func NewReport(version, schemaFile, envSource string, results []validator.ValidationResult) Report {
	r := Report{
		Version:    version,
		SchemaFile: schemaFile,
		EnvSource:  envSource,
		Results:    results,
		TotalVars:  len(results),
	}

	for _, res := range results {
		switch res.Status {
		case validator.StatusPass:
			r.PassCount++
		case validator.StatusFail:
			r.ErrorCount++
		case validator.StatusWarn:
			r.WarningCount++
		}
	}

	return r
}

// Passed returns true if there are no validation errors.
func (r Report) Passed() bool {
	return r.ErrorCount == 0
}

// Reporter is the interface for rendering validation reports.
type Reporter interface {
	Render(report Report, w io.Writer) error
}

// NewReporter creates a Reporter for the specified format.
func NewReporter(format string) (Reporter, error) {
	switch format {
	case "text":
		return &TextReporter{}, nil
	case "json":
		return &JSONReporter{}, nil
	case "github":
		return &GitHubReporter{}, nil
	default:
		return nil, fmt.Errorf("unknown output format %q, supported formats: text, json, github", format)
	}
}
