package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/ThousifXahamed079/gatekeeper/internal/validator"
)

// GitHubReporter renders validation results as GitHub Actions workflow commands.
type GitHubReporter struct{}

// Render writes the GitHub Actions annotations to w.
func (g *GitHubReporter) Render(report Report, w io.Writer) error {
	// Output annotations for failures and warnings
	for _, r := range report.Results {
		switch r.Status {
		case validator.StatusFail:
			fmt.Fprintf(w, "::error title=Gatekeeper::%s: %s\n", r.VarName, escapeGitHubMessage(r.Message))
		case validator.StatusWarn:
			fmt.Fprintf(w, "::warning title=Gatekeeper::%s: %s\n", r.VarName, escapeGitHubMessage(r.Message))
		case validator.StatusPass:
			// No output for passing results
		}
	}

	// Summary line
	if report.ErrorCount > 0 {
		fmt.Fprintf(w, "::error::Gatekeeper validation failed: %d error(s), %d warning(s)\n",
			report.ErrorCount, report.WarningCount)
	} else if report.WarningCount > 0 {
		fmt.Fprintf(w, "::warning::Gatekeeper validation passed with %d warning(s)\n",
			report.WarningCount)
	} else {
		fmt.Fprintf(w, "::notice::Gatekeeper validation passed: all %d variable(s) valid\n",
			report.TotalVars)
	}

	return nil
}

// escapeGitHubMessage escapes special characters for GitHub Actions annotations.
// Reference: https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions
func escapeGitHubMessage(s string) string {
	// Order matters: escape % first since it's used in the escape sequences
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "\r", "%0D")
	s = strings.ReplaceAll(s, "\n", "%0A")
	return s
}
