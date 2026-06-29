package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/reader"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/toolingfreshness"
)

const defaultRepoRoot = "../../.."

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Check for --skip-freshness to allow running against arbitrary directories
	skipFreshness := false
	remainingArgs := []string{}
	for _, arg := range args {
		if arg == "--skip-freshness" {
			skipFreshness = true
		} else {
			remainingArgs = append(remainingArgs, arg)
		}
	}
	args = remainingArgs

	if !skipFreshness {
		freshnessArgs := argsForFreshness(args)
		if err := toolingfreshness.CheckProcess(freshnessArgs, cwd); err != nil {
			fmt.Fprintf(stderr, "Warning: %v (use --skip-freshness to bypass)\n", err)
			return err
		}
	}
	if len(args) > 0 {
		switch args[0] {
		case toolingfreshness.HiddenBuildFingerprintCommand:
			fmt.Fprintln(stdout, toolingfreshness.PrintBuildFingerprint())
			return nil
		case "-h", "--help", "help":
			writeUsage(stdout)
			return nil
		}
	}
	options, err := parseOptions(args, stderr)
	if err != nil {
		return err
	}
	return reader.Serve(context.Background(), options, stdout)
}

func parseOptions(args []string, stderr io.Writer) (reader.ServeOptions, error) {
	fs := flag.NewFlagSet("specflow-reader", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repoRoot := fs.String("repo-root", defaultRepoRoot, "repository root")
	addr := fs.String("addr", "127.0.0.1:17863", "listen address")
	if err := fs.Parse(args); err != nil {
		return reader.ServeOptions{}, err
	}
	if fs.NArg() > 0 {
		return reader.ServeOptions{}, fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	return reader.ServeOptions{
		RepoRoot: mustAbs(*repoRoot),
		Addr:     *addr,
	}, nil
}

func writeUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflow-reader [--repo-root PATH] [--addr HOST:PORT]")
	fmt.Fprintf(w, "\nIf --repo-root is omitted, it defaults to %s from the current working directory.\n", defaultRepoRoot)
}

func argsForFreshness(args []string) []string {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help" || args[0] == toolingfreshness.HiddenBuildFingerprintCommand) {
		return args
	}
	if hasRepoRootFlag(args) {
		return args
	}
	freshnessArgs := append([]string{}, args...)
	freshnessArgs = append(freshnessArgs, "--repo-root", defaultRepoRoot)
	return freshnessArgs
}

func hasRepoRootFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--repo-root" || strings.HasPrefix(arg, "--repo-root=") {
			return true
		}
	}
	return false
}

func mustAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}
