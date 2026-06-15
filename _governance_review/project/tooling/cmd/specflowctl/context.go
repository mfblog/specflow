package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/context"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/contextcard"
)

func runContext(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		writeContextUsage(stderr)
		return errors.New("missing context subcommand")
	}

	switch args[0] {
	case "collect":
		return runContextCollect(args[1:], stdout, stderr)
	case "card":
		return runContextCard(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		writeContextUsage(stdout)
		return nil
	default:
		writeContextUsage(stderr)
		return fmt.Errorf("unknown context subcommand %q", args[0])
	}
}

func runContextCollect(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("context collect", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repoRootPtr := fs.String("repo-root", ".", "repository root")
	flowPtr := fs.String("flow", "", "context flow: lifecycle")
	commandPtr := fs.String("command", "", "context command (e.g. unit_verify)")
	objectPtr := fs.String("object", "", "formal object name (e.g. auth)")
	formatPtr := fs.String("format", string(context.FormatPack), "output format: pack | refs")
	if err := fs.Parse(args); err != nil {
		return err
	}

	flow := strings.TrimSpace(*flowPtr)
	command := strings.TrimSpace(*commandPtr)
	object := strings.TrimSpace(*objectPtr)
	outputFormat := context.RenderFormat(strings.TrimSpace(*formatPtr))

	if flow == "" || command == "" {
		writeContextCollectUsage(stderr)
		return errors.New("flow and command are required")
	}

	if outputFormat != context.FormatPack && outputFormat != context.FormatRefs {
		writeContextCollectUsage(stderr)
		return fmt.Errorf("unsupported format %q; use pack or refs", *formatPtr)
	}

	absRoot, err := filepath.Abs(*repoRootPtr)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}

	var collector context.Collector
	switch flow {
	case "lifecycle":
		collector, err = context.NewLifecycleCollector(command)
	default:
		return fmt.Errorf("unsupported context flow %q", flow)
	}
	if err != nil {
		return err
	}

	pack, err := collector.Collect(absRoot, object)
	if err != nil {
		return err
	}

	if err := pack.Render(stdout, outputFormat); err != nil {
		return err
	}

	// Report summary to stderr
	essentials := 0
	missing := 0
	for _, f := range pack.Files {
		if f.Essential {
			essentials++
		}
		if !f.Exists {
			missing++
		}
	}
	if missing > 0 {
		fmt.Fprintf(stderr, "Warning: %d file(s) not found (out of %d total)\n", missing, len(pack.Files))
	}

	return nil
}

func runContextCard(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("context card", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repoRootPtr := fs.String("repo-root", ".", "repository root")
	typePtr := fs.String("object-type", "", "object type: unit | rule")
	objectPtr := fs.String("object", "", "object name (e.g. auth for unit, rule_001 for rule)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	objectType := strings.TrimSpace(*typePtr)
	objectName := strings.TrimSpace(*objectPtr)

	if objectType == "" || objectName == "" {
		fmt.Fprintln(stderr, "Usage:")
		fmt.Fprintln(stderr, "  specflowctl context card --object-type unit|rule --object NAME [--repo-root PATH]")
		return errors.New("--object-type and --object are required")
	}

	absRoot, err := filepath.Abs(*repoRootPtr)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}

	switch objectType {
	case "unit":
		card, err := contextcard.UnitCard(absRoot, objectName)
		if err != nil {
			return fmt.Errorf("generate unit card: %w", err)
		}
		_, err = fmt.Fprintln(stdout, card)
		return err
	case "rule":
		card, err := contextcard.RuleCard(absRoot, objectName)
		if err != nil {
			return fmt.Errorf("generate rule card: %w", err)
		}
		_, err = fmt.Fprintln(stdout, card)
		return err
	default:
		return fmt.Errorf("unsupported type %q; use unit or rule", objectType)
	}
}

func writeContextUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl context collect --flow lifecycle --command COMMAND [--object OBJECT] [--format pack|refs] [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl context card --object-type unit|rule --object NAME [--repo-root PATH]")
}

func writeContextCollectUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl context collect --flow lifecycle --command COMMAND [--object OBJECT] [--format pack|refs] [--repo-root PATH]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Flows:")
	fmt.Fprintln(w, "  lifecycle      Unit lifecycle commands (unit_init, unit_new, unit_fork, unit_check, unit_verify, unit_promote, unit_stable_verify, unit_advance)")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  specflowctl context collect --flow lifecycle --command unit_verify --object auth")
	fmt.Fprintln(w, "  specflowctl context collect --flow lifecycle --command unit_verify --object auth --format refs")
}
