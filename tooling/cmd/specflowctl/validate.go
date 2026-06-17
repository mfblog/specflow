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

	var lastReason string
	constraintsFound := false

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

		constraintsFound = true
		r := filevalidation.ValidateWrite(phase, path, constraints)

		// Deny if any constraint denies
		if !r.Allowed {
			return validateResult{
				Allowed: r.Allowed,
				Reason:  fmt.Sprintf("unit=%s: %s", status.Object, r.Reason),
				Phase:   r.Phase,
				Path:    r.Path,
			}
		}
		lastReason = fmt.Sprintf("unit=%s: %s", status.Object, r.Reason)
	}

	if constraintsFound {
		// All constraints allowed the write
		return validateResult{
			Allowed: true,
			Reason:  lastReason,
			Phase:   phase,
			Path:    path,
		}
	}

	// No constraints found - apply defaults based on phase
	if phase == "pending_impl" || phase == "unit_impl" {
		// During implementation phase, deny spec/status/framework writes by default
		for _, denyPattern := range filevalidation.DefaultPendingImplDenyPatterns() {
			if filevalidation.MatchGlobPattern(denyPattern, path) {
				return validateResult{
					Allowed: false,
					Reason:  fmt.Sprintf("no constraints found; phase=%q matches default implementation-phase deny pattern %q", phase, denyPattern),
					Phase:   phase,
					Path:    path,
				}
			}
		}
	}
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

func writeCandidateFrontmatterResult(stdout io.Writer, result filevalidation.CandidateFrontmatterResult) {
	fmt.Fprintf(stdout, "valid: %t\n", result.Valid)
	fmt.Fprintf(stdout, "unit: %s\n", result.Unit)
	fmt.Fprintf(stdout, "diagnostic: %s\n", result.Diagnostic)
}

func writeValidateUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl validate write --path PATH --phase PHASE [--unit UNIT] [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl validate candidate-frontmatter --unit UNIT [--repo-root PATH]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Validate whether a file write is allowed in the current lifecycle phase.")
	fmt.Fprintln(w, "Checks constraints defined in docs/specs/_status.md.")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Validate candidate unit frontmatter consistency.")
	fmt.Fprintln(w, "Checks candidate_intent, source_basis, evidence_appendix_ref, repair_basis rules.")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --path PATH     File path to validate (required for 'write')")
	fmt.Fprintln(w, "  --phase PHASE   Current lifecycle phase (required for 'write')")
	fmt.Fprintln(w, "  --unit UNIT     Unit name (required for 'candidate-frontmatter'; optional for 'write')")
	fmt.Fprintln(w, "  --repo-root PATH Repository root directory (default: current directory)")
}
