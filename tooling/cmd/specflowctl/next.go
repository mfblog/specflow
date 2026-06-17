package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/next"
)

func runNext(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		// No args: list all units with state summary
		return runNextList(stdout, stderr)
	}

	fs := flag.NewFlagSet("next", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repoRootPtr := fs.String("repo-root", ".", "repository root")
	unitPtr := fs.String("unit", "", "unit name")
	explainPtr := fs.Bool("explain", false, "show full lifecycle context")
	if err := fs.Parse(args); err != nil {
		return err
	}

	unitName := strings.TrimSpace(*unitPtr)

	if unitName == "" {
		// Might have --explain but no --unit
		if *explainPtr {
			fmt.Fprintln(stderr, "Error: --explain requires --unit")
			writeNextUsage(stderr)
			return errors.New("--explain requires --unit")
		}
		return runNextList(stdout, stderr)
	}

	absRoot, err := filepath.Abs(*repoRootPtr)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}

	d, err := next.GetDirective(absRoot, unitName)
	if err != nil {
		return fmt.Errorf("get directive: %w", err)
	}

	if *explainPtr {
		output, err := next.ExplainDirective(absRoot, d)
		if err != nil {
			return fmt.Errorf("explain directive: %w", err)
		}
		_, err = fmt.Fprint(stdout, output)
		return err
	}

	_, err = fmt.Fprint(stdout, next.RenderDirective(d))
	return err
}

func runNextList(stdout, _ io.Writer) error {
	// List mode — show all units with their current state
	// Future: implement full list by reading _status.md
	// For now, direct user to use --unit
	fmt.Fprintln(stdout, "Usage: specflowctl next --unit <name>")
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "To get a directive for a specific unit, run:")
	fmt.Fprintln(stdout, "  specflowctl next --unit <unit_name>")
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "For full context with lifecycle reference:")
	fmt.Fprintln(stdout, "  specflowctl next --unit <unit_name> --explain")
	return nil
}

func writeNextUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl next --unit NAME [--explain] [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl next")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --unit NAME      Unit name to get directive for")
	fmt.Fprintln(w, "  --explain        Show full lifecycle context with references")
	fmt.Fprintln(w, "  --repo-root PATH Repository root path (default: .)")
}
