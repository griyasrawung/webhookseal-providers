package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/griyasrawung/webhookseal-providers/internal/linter"
	"github.com/griyasrawung/webhookseal-providers/internal/runner"
	"github.com/griyasrawung/webhookseal-providers/internal/validator"
)

const version = "v0.1.0"

var providers = []string{"stripe", "github", "shopify", "slack", "twilio"}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}

func run(args []string, out io.Writer) int {
	if len(args) == 0 {
		printHelp(out)
		return 1
	}

	if args[0] == "--help" || args[0] == "-h" {
		printHelp(out)
		return 0
	}

	if args[0] == "--version" {
		fmt.Fprintln(out, version)
		return 0
	}

	root := "."
	schema := filepath.Join(root, "schemas", "provider-spec.schema.json")
	switch args[0] {
	case "validate-schema":
		fs := flag.NewFlagSet("validate-schema", flag.ExitOnError)
		all := fs.Bool("all", false, "validate all providers")
		provider := fs.String("provider", "", "provider id")
		_ = fs.Parse(args[1:])
		targets := resolveTargets(*all, *provider)
		valid := 0
		for _, p := range targets {
			err := validator.ValidateProviderSpec(schema, filepath.Join(root, "providers", p+".yaml"))
			if err != nil {
				fmt.Fprintf(out, "%s: invalid (%v)\n", p, err)
				continue
			}
			valid++
		}
		fmt.Fprintf(out, "%d/%d provider specs valid\n", valid, len(targets))
		return 0
	case "run-fixtures":
		fs := flag.NewFlagSet("run-fixtures", flag.ExitOnError)
		all := fs.Bool("all", false, "run all fixtures")
		provider := fs.String("provider", "", "provider id")
		_ = fs.Parse(args[1:])
		targets := resolveTargets(*all, *provider)
		pass, fail := 0, 0
		for _, p := range targets {
			p1, f1, err := runner.RunFixtureFile(runner.FixturePath(root, p))
			if err != nil {
				fmt.Fprintf(out, "%s: error (%v)\n", p, err)
				fail++
				continue
			}
			pass += p1
			fail += f1
		}
		fmt.Fprintf(out, "%d passed, %d failed\n", pass, fail)
		return 0
	case "lint":
		fs := flag.NewFlagSet("lint", flag.ExitOnError)
		all := fs.Bool("all", false, "lint all providers")
		provider := fs.String("provider", "", "provider id")
		strict := fs.Bool("strict", false, "treat warnings as failures")
		_ = fs.Parse(args[1:])
		targets := resolveTargets(*all, *provider)
		errs, warns := 0, 0
		for _, p := range targets {
			issues := linter.LintSpec(filepath.Join(root, "providers", p+".yaml"))
			for _, issue := range issues {
				fmt.Fprintf(out, "%s: %s\n", p, issue)
				switch issue.Severity {
				case linter.SeverityError:
					errs++
				case linter.SeverityWarning:
					warns++
				}
			}
		}
		fmt.Fprintf(out, "%d lint errors, %d lint warnings\n", errs, warns)
		if errs > 0 || (*strict && warns > 0) {
			return 1
		}
		return 0
	default:
		printHelp(out)
		return 1
	}
}

func resolveTargets(all bool, provider string) []string {
	if all {
		return providers
	}
	if strings.TrimSpace(provider) != "" {
		return []string{provider}
	}
	return providers
}

func printHelp(out io.Writer) {
	fmt.Fprintln(out, "webhookseal CLI")
	fmt.Fprintln(out, "commands: validate-schema, run-fixtures, lint")
}
