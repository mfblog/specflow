package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/filevalidation"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

const constraintsNotesPrefix = "constraints:"

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
		phase := fs.String("phase", "", "current lifecycle phase (e.g. pending_impl, unit_verify)")
		unit := fs.String("unit", "", "unit name (optional, restricts constraint check to one unit)")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*path) == "" || strings.TrimSpace(*phase) == "" {
			writeValidateUsage(stderr)
			return errors.New("path and phase are required")
		}

		result := validateWrite(mustAbs(*repoRoot), *path, *phase, *unit)
		writeValidateWriteResult(stdout, result)
		if !result.Allowed {
			return fmt.Errorf("write denied: %s", result.Reason)
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
	Phase   string
	Path    string
}

func validateWrite(repoRoot, path, phase, unit string) validateResult {
	// Load constraints from _status.md
	statuses, err := statusfile.LoadObjectStatuses(repoRoot)
	if err != nil {
		return validateResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot load status: %v", err),
			Phase:   phase,
			Path:    path,
		}
	}

	for _, status := range statuses {
		// If unit is specified, only check that unit
		if unit != "" && status.Object != unit {
			continue
		}
		// Skip rows without constraints in notes
		if !strings.HasPrefix(strings.TrimSpace(status.Notes), constraintsNotesPrefix) {
			continue
		}

		constraintsStr := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(status.Notes), constraintsNotesPrefix))
		constraints, err := filevalidation.ParseConstraints(constraintsStr)
		if err != nil {
			return validateResult{
				Allowed: false,
				Reason:  fmt.Sprintf("parse constraints for %s: %v", status.Object, err),
				Phase:   phase,
				Path:    path,
			}
		}

		r := filevalidation.ValidateWrite(phase, path, constraints)
		return validateResult{
			Allowed: r.Allowed,
			Reason:  fmt.Sprintf("unit=%s: %s", status.Object, r.Reason),
			Phase:   r.Phase,
			Path:    r.Path,
		}
	}

	// No constraints found - allow by default
	return validateResult{
		Allowed: true,
		Reason:  "no constraints found for this path; write permitted by default",
		Phase:   phase,
		Path:    path,
	}
}

func writeValidateWriteResult(stdout io.Writer, result validateResult) {
	fmt.Fprintf(stdout, "allowed: %t\n", result.Allowed)
	fmt.Fprintf(stdout, "reason: %s\n", result.Reason)
	fmt.Fprintf(stdout, "phase: %s\n", noneIfEmpty(result.Phase))
	fmt.Fprintf(stdout, "path: %s\n", noneIfEmpty(result.Path))
}

func writeValidateUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl validate write --path PATH --phase PHASE [--unit UNIT] [--repo-root PATH]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Validate whether a file write is allowed in the current lifecycle phase.")
	fmt.Fprintln(w, "Checks constraints defined in docs/specs/_status.md.")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --path PATH     File path to validate (required)")
	fmt.Fprintln(w, "  --phase PHASE   Current lifecycle phase: unit_check, pending_impl, unit_verify, etc. (required)")
	fmt.Fprintln(w, "  --unit UNIT     Unit name (optional, restricts check to one unit)")
	fmt.Fprintln(w, "  --repo-root PATH Repository root directory (default: current directory)")
}
