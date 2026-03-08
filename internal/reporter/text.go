package reporter

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ThousifXahamed079/gatekeeper/internal/validator"
)

// ANSI colour codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
)

// TextReporter renders validation results in human-friendly coloured text.
type TextReporter struct{}

// Render writes the text report to w.
func (t *TextReporter) Render(report Report, w io.Writer) error {
	useColor := shouldUseColor(w)
	c := colorizer{enabled: useColor}

	// Header
	fmt.Fprintf(w, "%s\n", c.cyan(c.bold(fmt.Sprintf("🚪 Gatekeeper v%s — validating environment...", report.Version))))
	fmt.Fprintln(w)

	// Schema and source info
	fmt.Fprintf(w, "%s  %s (%d vars)\n", c.cyan("Schema:"), report.SchemaFile, report.TotalVars)
	fmt.Fprintf(w, "%s  %s\n", c.cyan("Source:"), report.EnvSource)
	fmt.Fprintln(w)

	// Handle empty results
	if report.TotalVars == 0 {
		fmt.Fprintln(w, "No variables defined in schema.")
		return nil
	}

	// Calculate max name length for alignment
	maxNameLen := 0
	for _, r := range report.Results {
		if len(r.VarName) > maxNameLen {
			maxNameLen = len(r.VarName)
		}
	}
	maxNameLen += 2 // padding

	// Per-variable lines
	for _, r := range report.Results {
		line := formatResultLine(r, maxNameLen, c)
		fmt.Fprintln(w, line)
	}

	fmt.Fprintln(w)

	// Separator
	separator := strings.Repeat("─", 48)
	fmt.Fprintln(w, c.dim(separator))

	// Result summary
	resultText := fmt.Sprintf("RESULT: %d error(s), %d warning(s) — ", report.ErrorCount, report.WarningCount)
	if report.Passed() {
		fmt.Fprintf(w, "%s%s\n", resultText, c.green(c.bold("PASSED")))
	} else {
		fmt.Fprintf(w, "%s%s\n", resultText, c.red(c.bold("FAILED")))
	}

	// Detail lines for failures and warnings
	if report.ErrorCount > 0 || report.WarningCount > 0 {
		fmt.Fprintln(w)

		// Failures first
		for _, r := range report.Results {
			if r.Status == validator.StatusFail {
				fmt.Fprintf(w, "  %s %s: %s\n", c.red("✗"), r.VarName, r.Message)
			}
		}

		// Then warnings
		for _, r := range report.Results {
			if r.Status == validator.StatusWarn {
				fmt.Fprintf(w, "  %s %s: %s\n", c.yellow("⚠"), r.VarName, r.Message)
			}
		}
	}

	return nil
}

func formatResultLine(r validator.ValidationResult, maxNameLen int, c colorizer) string {
	// Icon
	var icon string
	switch r.Status {
	case validator.StatusPass:
		icon = c.green("✅")
	case validator.StatusFail:
		icon = c.red("❌")
	case validator.StatusWarn:
		icon = c.yellow("⚠️ ") // trailing space for alignment
	}

	// Padded name
	paddedName := r.VarName + strings.Repeat(" ", maxNameLen-len(r.VarName))

	// Required label
	reqLabel := "optional"
	if r.Required {
		reqLabel = "required"
	}

	// Value or status message
	var valueOrStatus string
	switch r.Status {
	case validator.StatusPass:
		if r.Sensitive {
			valueOrStatus = "[REDACTED]"
		} else {
			valueOrStatus = truncateForDisplay(r.Value, 40)
		}
	case validator.StatusFail, validator.StatusWarn:
		valueOrStatus = truncateForDisplay(r.Message, 60)
	}

	// Type display
	typeDisplay := r.VarType
	if typeDisplay == "" {
		typeDisplay = "string"
	}

	line := fmt.Sprintf("%s %s — %s, %s  %s", icon, paddedName, typeDisplay, reqLabel, valueOrStatus)

	// Apply colour to the whole line based on status
	switch r.Status {
	case validator.StatusPass:
		return c.green(line)
	case validator.StatusFail:
		return c.red(line)
	case validator.StatusWarn:
		return c.yellow(line)
	}
	return line
}

func truncateForDisplay(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

// shouldUseColor determines if ANSI colors should be used.
func shouldUseColor(w io.Writer) bool {
	// Check FORCE_COLOR first (overrides everything)
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Check NO_COLOR
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if writing to a terminal
	return IsTerminal(w)
}

// IsTerminal checks if the writer is a terminal.
func IsTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// colorizer handles conditional ANSI colouring.
type colorizer struct {
	enabled bool
}

func (c colorizer) apply(code, s string) string {
	if !c.enabled {
		return s
	}
	return code + s + Reset
}

func (c colorizer) red(s string) string    { return c.apply(Red, s) }
func (c colorizer) green(s string) string  { return c.apply(Green, s) }
func (c colorizer) yellow(s string) string { return c.apply(Yellow, s) }
func (c colorizer) cyan(s string) string   { return c.apply(Cyan, s) }
func (c colorizer) bold(s string) string   { return c.apply(Bold, s) }
func (c colorizer) dim(s string) string    { return c.apply(Dim, s) }
