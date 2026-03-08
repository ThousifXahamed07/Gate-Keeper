package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ThousifXahamed079/gatekeeper/internal/cli"
	"github.com/ThousifXahamed079/gatekeeper/internal/docs"
	"github.com/ThousifXahamed079/gatekeeper/internal/envloader"
	"github.com/ThousifXahamed079/gatekeeper/internal/schema"
	"github.com/ThousifXahamed079/gatekeeper/internal/validator"
)

var version = "v0.0.0-dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(cli.ExitSuccess)
	}

	switch os.Args[1] {
	case "check":
		os.Exit(runCheck(os.Args[2:]))
	case "docs":
		os.Exit(runDocs(os.Args[2:]))
	case "version":
		fmt.Println("gatekeeper " + version)
		os.Exit(cli.ExitSuccess)
	case "help", "--help", "-h":
		printUsage()
		os.Exit(cli.ExitSuccess)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(cli.ExitValidation)
	}
}

func printUsage() {
	fmt.Println("gatekeeper - The env validator that never lets bad config through.")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  gatekeeper <command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  check     Validate environment variables against schema")
	fmt.Println("  docs      Generate documentation from schema")
	fmt.Println("  version   Print version information")
	fmt.Println("  help      Show this help message")
	fmt.Println("")
	fmt.Println("Run 'gatekeeper <command> --help' for command-specific options.")
}

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	
	schemaPath := fs.String("schema", "", "Path to schema file (default: auto-detect)")
	envFile := fs.String("env-file", ".env", "Path to .env file")
	format := fs.String("format", "text", "Output format: text, json, github")
	noEnvFile := fs.Bool("no-env-file", false, "Skip loading .env file, only use OS environment")
	strict := fs.Bool("strict", false, "Treat warnings as errors")

	fs.Usage = func() {
		fmt.Println("Usage: gatekeeper check [options]")
		fmt.Println("")
		fmt.Println("Validate environment variables against the schema.")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return cli.ExitValidation
	}

	// Validate format
	if *format != "text" && *format != "json" && *format != "github" {
		fmt.Fprintf(os.Stderr, "Error: invalid format %q, must be one of: text, json, github\n", *format)
		return cli.ExitValidation
	}

	// Load schema
	var schemaFile string
	var err error
	
	if *schemaPath != "" {
		schemaFile = *schemaPath
	} else {
		schemaFile, err = schema.FindSchemaFile()
		if err != nil {
			printSchemaError(err.Error())
			return cli.ExitSchemaError
		}
	}

	s, err := schema.ParseFile(schemaFile)
	if err != nil {
		printSchemaError(err.Error())
		return cli.ExitSchemaError
	}

	// Load environment
	envFilePath := ""
	if !*noEnvFile {
		envFilePath = *envFile
	}
	
	envData, err := envloader.Load(envFilePath, *noEnvFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return cli.ExitValidation
	}

	// Validate
	results := validator.Validate(s, envData.Values, envData.Sources)

	// Count results
	var passed, failed, warned int
	for _, r := range results {
		switch r.Status {
		case validator.StatusPass:
			passed++
		case validator.StatusFail:
			failed++
		case validator.StatusWarn:
			warned++
		}
	}
	total := passed + failed + warned

	// Output results to stdout
	switch *format {
	case "json":
		printJSON(results)
	case "github":
		printGitHub(results)
	default:
		printText(results)
	}

	// Determine exit code and print summary to stderr
	if failed > 0 {
		printSummaryFail(failed, warned, total)
		return cli.ExitValidation
	}
	if *strict && warned > 0 {
		printSummaryFail(failed, warned, total)
		return cli.ExitValidation
	}
	
	printSummarySuccess(total)
	return cli.ExitSuccess
}

func runDocs(args []string) int {
	fs := flag.NewFlagSet("docs", flag.ExitOnError)

	schemaPath := fs.String("schema", "", "Path to schema file (default: auto-detect)")
	format := fs.String("format", "markdown", "Output format: markdown, env-example")
	outPath := fs.String("out", "", "Output file path (default: stdout)")

	fs.Usage = func() {
		fmt.Println("Usage: gatekeeper docs [options]")
		fmt.Println("")
		fmt.Println("Generate documentation from the schema.")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return cli.ExitValidation
	}

	// Validate format
	if *format != "markdown" && *format != "env-example" {
		fmt.Fprintf(os.Stderr, "Error: invalid format %q, must be one of: markdown, env-example\n", *format)
		return cli.ExitValidation
	}

	// Load schema
	var schemaFile string
	var err error

	if *schemaPath != "" {
		schemaFile = *schemaPath
	} else {
		schemaFile, err = schema.FindSchemaFile()
		if err != nil {
			printSchemaError(err.Error())
			return cli.ExitSchemaError
		}
	}

	s, err := schema.ParseFile(schemaFile)
	if err != nil {
		printSchemaError(err.Error())
		return cli.ExitSchemaError
	}

	// Determine output destination
	var output *os.File
	if *outPath != "" {
		output, err = os.Create(*outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot create output file: %s\n", err.Error())
			return cli.ExitValidation
		}
		defer output.Close()
	} else {
		output = os.Stdout
	}

	// Generate documentation
	switch *format {
	case "markdown":
		if err := docs.GenerateMarkdown(s, output); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			return cli.ExitValidation
		}
	case "env-example":
		if err := docs.GenerateEnvExample(s, output); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			return cli.ExitValidation
		}
	}

	// Print success message to stderr if writing to file
	if *outPath != "" {
		fmt.Fprintf(os.Stderr, "✅ Documentation written to %s\n", *outPath)
	}

	return cli.ExitSuccess
}

func printSchemaError(message string) {
	fmt.Fprintf(os.Stderr, "❌ Gatekeeper: schema error — %s\n", message)
}

func printSummarySuccess(total int) {
	fmt.Fprintf(os.Stderr, "✅ Gatekeeper: all %d variable(s) validated successfully\n", total)
}

func printSummaryFail(errors, warnings, total int) {
	fmt.Fprintf(os.Stderr, "❌ Gatekeeper: %d error(s), %d warning(s) in %d variable(s)\n", errors, warnings, total)
}

func printJSON(results []validator.ValidationResult) {
	type jsonResult struct {
		VarName   string `json:"var_name"`
		Value     string `json:"value,omitempty"`
		Expected  string `json:"expected,omitempty"`
		Status    string `json:"status"`
		Message   string `json:"message"`
		Sensitive bool   `json:"sensitive,omitempty"`
		Source    string `json:"source"`
	}

	output := struct {
		Results []jsonResult `json:"results"`
		Summary struct {
			Total  int `json:"total"`
			Passed int `json:"passed"`
			Failed int `json:"failed"`
			Warned int `json:"warned"`
		} `json:"summary"`
	}{}

	for _, r := range results {
		jr := jsonResult{
			VarName:   r.VarName,
			Expected:  r.Expected,
			Status:    r.Status.String(),
			Message:   r.Message,
			Sensitive: r.Sensitive,
			Source:    r.Source,
		}
		// Redact sensitive values
		if r.Sensitive && r.Value != "" {
			jr.Value = "[REDACTED]"
		} else {
			jr.Value = r.Value
		}
		output.Results = append(output.Results, jr)

		output.Summary.Total++
		switch r.Status {
		case validator.StatusPass:
			output.Summary.Passed++
		case validator.StatusFail:
			output.Summary.Failed++
		case validator.StatusWarn:
			output.Summary.Warned++
		}
	}

	data, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(data))
}

func printGitHub(results []validator.ValidationResult) {
	for _, r := range results {
		switch r.Status {
		case validator.StatusFail:
			fmt.Printf("::error::Variable %s: %s\n", r.VarName, r.Message)
		case validator.StatusWarn:
			fmt.Printf("::warning::Variable %s: %s\n", r.VarName, r.Message)
		}
	}
}

func printText(results []validator.ValidationResult) {
	fmt.Println("Gatekeeper Validation Results")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	for _, r := range results {
		var icon string
		switch r.Status {
		case validator.StatusPass:
			icon = "✅"
		case validator.StatusFail:
			icon = "❌"
		case validator.StatusWarn:
			icon = "⚠️"
		}

		value := r.Value
		if r.Sensitive && value != "" {
			value = "[REDACTED]"
		}

		fmt.Printf("%s %s\n", icon, r.VarName)
		fmt.Printf("   Status:  %s\n", r.Status.String())
		fmt.Printf("   Message: %s\n", r.Message)
		if value != "" {
			fmt.Printf("   Value:   %s\n", value)
		}
		if r.Source != "" {
			fmt.Printf("   Source:  %s\n", r.Source)
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("-", 50))
}
