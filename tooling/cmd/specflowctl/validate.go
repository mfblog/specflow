package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/filevalidation"
)

func runValidate(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		writeValidateUsage(stderr)
		return errors.New("missing validate subcommand")
	}

	switch args[0] {
	case "write":
		fs := flag.NewFlagSet("validate write", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		path := fs.String("path", "", "file path to validate")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*path) == "" {
			writeValidateUsage(stderr)
			return errors.New("path is required")
		}

		result := validateWrite(mustAbs(*repoRoot), *path)
		writeValidateWriteResult(stdout, result)
		if !result.Allowed {
			return fmt.Errorf("write denied: %s", result.Reason)
		}
		return nil
	case "candidate-frontmatter":
		fs := flag.NewFlagSet("validate candidate-frontmatter", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		unitName := fs.String("unit", "", "unit name")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*unitName) == "" {
			writeValidateUsage(stderr)
			return errors.New("unit is required")
		}

		result := filevalidation.ValidateCandidateFrontmatter(mustAbs(*repoRoot), *unitName)
		writeCandidateFrontmatterResult(stdout, result)
		if !result.Valid {
			return fmt.Errorf("candidate frontmatter validation failed: %s", result.Diagnostic)
		}
		return nil
	case "-h", "--help", "help":
		writeValidateUsage(stdout)
		return nil
	default:
		writeValidateUsage(stderr)
		return fmt.Errorf("unknown validate subcommand %q", args[0])
	}
}

type validateResult struct {
	Allowed bool
	Reason  string
	Path    string
}

func validateWrite(repoRoot, path string) validateResult {
	normalizedPath := filepath.ToSlash(filepath.Clean(path))

	// Deny pattern: framework files are never writable via validate
	if strings.HasPrefix(normalizedPath, "framework/") {
		return validateResult{
			Allowed: false,
			Reason:  fmt.Sprintf("path %q is under framework/ and is not writable", path),
			Path:    path,
		}
	}

	// Deny pattern: stable spec files are not directly writable (use promote)
	if strings.HasPrefix(normalizedPath, "docs/specs/units/stable/") {
		return validateResult{
			Allowed: false,
			Reason:  fmt.Sprintf("path %q is under docs/specs/units/stable/; use promote to write stable specs", path),
			Path:    path,
		}
	}

	// Deny pattern: rule files are not directly writable (use rule governance flows)
	if strings.HasPrefix(normalizedPath, "docs/specs/rules/stable/") {
		return validateResult{
			Allowed: false,
			Reason:  fmt.Sprintf("path %q is under docs/specs/rules/stable/; use rule governance flows", path),
			Path:    path,
		}
	}

	// Candidate spec files are writable
	if strings.HasPrefix(normalizedPath, "docs/specs/units/candidate/") {
		return validateResult{
			Allowed: true,
			Reason:  fmt.Sprintf("path %q is a candidate spec file and is writable", path),
			Path:    path,
		}
	}

	// Candidate rule files are writable
	if strings.HasPrefix(normalizedPath, "docs/specs/rules/candidate/") {
		return validateResult{
			Allowed: true,
			Reason:  fmt.Sprintf("path %q is a candidate rule file and is writable", path),
			Path:    path,
		}
	}

	// Source code files are writable by default
	return validateResult{
		Allowed: true,
		Reason:  fmt.Sprintf("path %q is not governed by specFlow write restrictions", path),
		Path:    path,
	}
}

func writeValidateWriteResult(stdout io.Writer, result validateResult) {
	fmt.Fprintf(stdout, "allowed: %t\n", result.Allowed)
	fmt.Fprintf(stdout, "reason: %s\n", result.Reason)
	fmt.Fprintf(stdout, "path: %s\n", noneIfEmpty(result.Path))
}

func writeCandidateFrontmatterResult(stdout io.Writer, result filevalidation.CandidateFrontmatterResult) {
	fmt.Fprintf(stdout, "valid: %t\n", result.Valid)
	fmt.Fprintf(stdout, "unit: %s\n", result.Unit)
	fmt.Fprintf(stdout, "diagnostic: %s\n", result.Diagnostic)
}

func writeValidateUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl validate write --path PATH [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl validate candidate-frontmatter --unit UNIT [--repo-root PATH]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Validate whether a file write is allowed under current governance.")
	fmt.Fprintln(w, "Checks path against allowed write zones (candidate specs, source code).")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Validate candidate unit frontmatter consistency.")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --path PATH     File path to validate (required for 'write')")
	fmt.Fprintln(w, "  --unit UNIT     Unit name (required for 'candidate-frontmatter')")
	fmt.Fprintln(w, "  --repo-root PATH Repository root directory (default: current directory)")
}
