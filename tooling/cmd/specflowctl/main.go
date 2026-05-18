package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Bingordinary/SpecFlow/tooling/internal/buildrelease"
	"github.com/Bingordinary/SpecFlow/tooling/internal/commandclose"
	"github.com/Bingordinary/SpecFlow/tooling/internal/commandpreflight"
	"github.com/Bingordinary/SpecFlow/tooling/internal/entrysync"
	"github.com/Bingordinary/SpecFlow/tooling/internal/install"
	"github.com/Bingordinary/SpecFlow/tooling/internal/processcleanup"
	"github.com/Bingordinary/SpecFlow/tooling/internal/projectstandards"
	"github.com/Bingordinary/SpecFlow/tooling/internal/repositorymapping"
	"github.com/Bingordinary/SpecFlow/tooling/internal/reviewrun"
	"github.com/Bingordinary/SpecFlow/tooling/internal/reviewscope"
	"github.com/Bingordinary/SpecFlow/tooling/internal/rulesync"
	"github.com/Bingordinary/SpecFlow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/tooling/internal/statusfile"
	"github.com/Bingordinary/SpecFlow/tooling/internal/toolingfreshness"
)

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
	if err := toolingfreshness.CheckProcess(args, cwd); err != nil {
		return err
	}

	if len(args) == 0 {
		writeRootUsage(stderr)
		return errors.New("missing command")
	}

	switch args[0] {
	case toolingfreshness.HiddenBuildFingerprintCommand:
		fmt.Fprintln(stdout, toolingfreshness.PrintBuildFingerprint())
		return nil
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "doctor":
		return runDoctor(args[1:], stdout, stderr)
	case "upgrade":
		return runUpgrade(args[1:], stdout, stderr)
	case "build-release":
		return runBuildRelease(args[1:], stdout, stderr)
	case "command":
		return runCommand(args[1:], stdout, stderr)
	case "entry":
		return runEntry(args[1:], stdout, stderr)
	case "registry":
		return runRegistry(args[1:], stdout, stderr)
	case "repository-mapping":
		return runRepositoryMapping(args[1:], stdout, stderr)
	case "review":
		return runReview(args[1:], stdout, stderr)
	case "process":
		return runProcess(args[1:], stdout, stderr)
	case "rule":
		return runRule(args[1:], stdout, stderr)
	case "snapshot":
		return runSnapshot(args[1:], stdout, stderr)
	case "status":
		return runStatus(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		writeRootUsage(stdout)
		return nil
	default:
		writeRootUsage(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runCommand(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		writeCommandUsage(stderr)
		return errors.New("missing command subcommand")
	}

	switch args[0] {
	case "close":
		fs := flag.NewFlagSet("command close", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		command := fs.String("command", "", "standard command name")
		objectType := fs.String("object-type", "", "formal object type: unit")
		object := fs.String("object", "", "formal object name")
		outcome := fs.String("outcome", "", "standard command outcome")
		reason := fs.String("reason", "", "fallback or diagnostic reason code")
		failureLayer := fs.String("failure-layer", "", "explicit fallback layer")
		candidateIntent := fs.String("candidate-intent", "", "controlled candidate intent: repair | change")
		notes := fs.String("notes", "", "status notes")
		stableBefore := fs.String("stable-before", "", "previous stable value for promotion recovery: yes | no")
		apply := fs.Bool("apply", false, "write status and cleanup process files")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*command) == "" || strings.TrimSpace(*objectType) == "" || strings.TrimSpace(*object) == "" || strings.TrimSpace(*outcome) == "" {
			writeCommandUsage(stderr)
			return errors.New("command, object-type, object, and outcome are required")
		}

		result, err := commandclose.Close(commandclose.Options{
			RepoRoot:        mustAbs(*repoRoot),
			Command:         *command,
			ObjectType:      *objectType,
			Object:          *object,
			Outcome:         *outcome,
			Reason:          *reason,
			FailureLayer:    *failureLayer,
			CandidateIntent: *candidateIntent,
			Notes:           *notes,
			StableBefore:    *stableBefore,
			Apply:           *apply,
		})
		if result.Command != "" {
			writeCommandCloseResult(stdout, result, err)
		}
		return err
	case "preflight":
		fs := flag.NewFlagSet("command preflight", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		command := fs.String("command", "", "standard command name")
		objectType := fs.String("object-type", "", "formal object type: unit")
		object := fs.String("object", "", "formal object name")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*command) == "" || strings.TrimSpace(*objectType) == "" || strings.TrimSpace(*object) == "" {
			writeCommandUsage(stderr)
			return errors.New("command, object-type, and object are required")
		}

		result := commandpreflight.Run(mustAbs(*repoRoot), *command, *objectType, *object)
		writeCommandPreflightResult(stdout, result)
		if !result.MayContinue {
			return errors.New("command preflight failed")
		}
		return nil
	case "-h", "--help", "help":
		writeCommandUsage(stdout)
		return nil
	default:
		writeCommandUsage(stderr)
		return fmt.Errorf("unknown command subcommand %q", args[0])
	}
}

func writeCommandPreflightResult(stdout io.Writer, result commandpreflight.Result) {
	preflightResult := "pass"
	if !result.MayContinue {
		preflightResult = "fail"
	}
	fmt.Fprintf(stdout, "preflight_result: %s\n", preflightResult)
	fmt.Fprintf(stdout, "command: %s\n", result.Command)
	fmt.Fprintf(stdout, "object_type: %s\n", result.ObjectType)
	fmt.Fprintf(stdout, "object: %s\n", result.Object)
	fmt.Fprintf(stdout, "may_continue: %t\n", result.MayContinue)
	fmt.Fprintf(stdout, "failure_layer: %s\n", noneIfEmpty(result.FailureLayer))
	fmt.Fprintf(stdout, "recommended_next_command: %s\n", noneIfEmpty(result.RecommendedNextCommand))
	fmt.Fprintln(stdout, "validated_processes:")
	if len(result.ValidatedProcesses) == 0 {
		fmt.Fprintln(stdout, "- none")
	} else {
		for _, process := range result.ValidatedProcesses {
			fmt.Fprintf(stdout, "- process: %s\n", process.ProcessKind)
			fmt.Fprintf(stdout, "  file: %s\n", noneIfEmpty(process.ProcessFile))
			fmt.Fprintf(stdout, "  result: %s\n", process.Result)
			fmt.Fprintf(stdout, "  failure_layer: %s\n", noneIfEmpty(process.FailureLayer))
			fmt.Fprintf(stdout, "  recommended_next_command: %s\n", noneIfEmpty(process.RecommendedNextCommand))
			for _, diagnostic := range process.Diagnostics {
				fmt.Fprintf(stdout, "  diagnostic: %s\n", diagnostic)
			}
		}
	}
	writeList(stdout, "diagnostics", result.Diagnostics)
}

func writeCommandCloseResult(stdout io.Writer, result commandclose.Result, closeErr error) {
	closeResult := "dry_run"
	if closeErr != nil {
		closeResult = "failed"
	} else if result.Applied {
		closeResult = "applied"
	}
	fmt.Fprintf(stdout, "command_close_result: %s\n", closeResult)
	fmt.Fprintf(stdout, "command: %s\n", result.Command)
	fmt.Fprintf(stdout, "object_type: %s\n", result.ObjectType)
	fmt.Fprintf(stdout, "object: %s\n", result.Object)
	fmt.Fprintf(stdout, "outcome: %s\n", result.Outcome)
	fmt.Fprintf(stdout, "apply: %t\n", result.Applied)
	fmt.Fprintf(stdout, "input_validation_action: %s\n", noneIfEmpty(result.InputValidationAction))
	fmt.Fprintf(stdout, "validation_action: %s\n", noneIfEmpty(result.ValidationAction))
	fmt.Fprintf(stdout, "cleanup_action: %s\n", noneIfEmpty(result.CleanupAction))
	fmt.Fprintf(stdout, "status_updated: %t\n", result.StatusUpdated)
	fmt.Fprintln(stdout, "status_before:")
	if result.StatusBeforePresent {
		writeCommandCloseStatus(stdout, result.StatusBefore)
	} else {
		fmt.Fprintln(stdout, "  present: false")
	}
	fmt.Fprintln(stdout, "status_after:")
	writeCommandCloseStatus(stdout, result.StatusAfter)
	writeCommandCloseInputProcesses(stdout, result.InputValidatedProcesses)
	writeList(stdout, "input_validation_mismatches", result.InputValidationMismatches)
	writeList(stdout, "validation_mismatches", result.ValidationMismatches)
	writeList(stdout, "fallback_deleted_files", result.FallbackCleanup.DeletedFiles)
	writeList(stdout, "fallback_missing_files", result.FallbackCleanup.MissingFiles)
	writeList(stdout, "success_deleted_files", result.SuccessCleanup.DeletedFiles)
	writeList(stdout, "success_missing_files", result.SuccessCleanup.MissingFiles)
}

func writeCommandCloseInputProcesses(stdout io.Writer, processes []commandpreflight.Process) {
	fmt.Fprintln(stdout, "input_validated_processes:")
	if len(processes) == 0 {
		fmt.Fprintln(stdout, "- none")
		return
	}
	for _, process := range processes {
		fmt.Fprintf(stdout, "- process: %s\n", process.ProcessKind)
		fmt.Fprintf(stdout, "  file: %s\n", noneIfEmpty(process.ProcessFile))
		fmt.Fprintf(stdout, "  result: %s\n", process.Result)
		fmt.Fprintf(stdout, "  failure_layer: %s\n", noneIfEmpty(process.FailureLayer))
		fmt.Fprintf(stdout, "  recommended_next_command: %s\n", noneIfEmpty(process.RecommendedNextCommand))
		for _, diagnostic := range process.Diagnostics {
			fmt.Fprintf(stdout, "  diagnostic: %s\n", diagnostic)
		}
	}
}

func writeCommandCloseStatus(stdout io.Writer, status statusfile.ObjectStatus) {
	present := status.ObjectType != "" || status.Object != ""
	fmt.Fprintf(stdout, "  present: %t\n", present)
	fmt.Fprintf(stdout, "  object_type: %s\n", noneIfEmpty(status.ObjectType))
	fmt.Fprintf(stdout, "  object: %s\n", noneIfEmpty(status.Object))
	fmt.Fprintf(stdout, "  stable: %s\n", noneIfEmpty(status.Stable))
	fmt.Fprintf(stdout, "  candidate: %s\n", noneIfEmpty(status.Candidate))
	fmt.Fprintf(stdout, "  active_layer: %s\n", noneIfEmpty(status.ActiveLayer))
	fmt.Fprintf(stdout, "  next_command: %s\n", noneIfEmpty(status.NextCommand))
	fmt.Fprintf(stdout, "  notes: %s\n", noneIfEmpty(status.Notes))
}

func runInit(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repoRoot := fs.String("repo-root", ".", "repository root")
	force := fs.Bool("force", false, "overwrite managed files")
	if err := fs.Parse(args); err != nil {
		return err
	}

	result, err := install.Init(mustAbs(*repoRoot), *force)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "specFlow init completed. copied=%d skipped=%d\n", result.Copied, result.Skipped)
	return nil
}

func runDoctor(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repoRoot := fs.String("repo-root", ".", "repository root")
	if err := fs.Parse(args); err != nil {
		return err
	}

	result, err := install.Doctor(mustAbs(*repoRoot))
	if err != nil {
		return err
	}
	for _, warning := range result.Warnings {
		fmt.Fprintln(stdout, warning)
	}
	if len(result.Failures) == 0 {
		fmt.Fprintln(stdout, "specFlow doctor passed")
		return nil
	}
	for _, failure := range result.Failures {
		fmt.Fprintln(stdout, failure)
	}
	return fmt.Errorf("specFlow doctor failed: %d issue(s)", len(result.Failures))
}

func runUpgrade(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("upgrade", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repoRoot := fs.String("repo-root", ".", "repository root")
	if err := fs.Parse(args); err != nil {
		return err
	}

	result, err := install.Upgrade(mustAbs(*repoRoot))
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "specFlow upgrade completed. updated=%d skipped=%d\n", result.Updated, result.Skipped)
	return nil
}

func runBuildRelease(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("build-release", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repoRoot := fs.String("repo-root", ".", "repository root")
	if err := fs.Parse(args); err != nil {
		return err
	}

	result, err := buildrelease.BuildAll(mustAbs(*repoRoot), nil)
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Built release binaries:")
	for _, target := range result.Targets {
		fmt.Fprintf(stdout, "- %s\n", target)
	}
	return nil
}

func runEntry(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		writeEntryUsage(stderr)
		return errors.New("missing entry subcommand")
	}

	switch args[0] {
	case "check":
		fs := flag.NewFlagSet("entry check", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}

		inspection, err := entrysync.Inspect(mustAbs(*repoRoot))
		if err != nil {
			return err
		}

		if inspection.Consistent {
			fmt.Fprintln(stdout, "Managed entry blocks are already consistent.")
			return nil
		}

		fmt.Fprintln(stdout, "Managed entry blocks are inconsistent.")
		if inspection.SuggestedSource != "" {
			fmt.Fprintf(stdout, "Suggested source: %s\n", inspection.SuggestedSource)
		}
		if len(inspection.CurrentRoundChanged) > 0 {
			fmt.Fprintln(stdout, "Registered entry files changed in current round:")
			for _, path := range inspection.CurrentRoundChanged {
				fmt.Fprintf(stdout, "- %s\n", path)
			}
		}
		return errors.New("entry managed blocks differ")
	case "sync":
		fs := flag.NewFlagSet("entry sync", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		source := fs.String("source", "", "registered source entry file")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}

		result, err := entrysync.Sync(mustAbs(*repoRoot), *source)
		if err != nil {
			return err
		}

		if len(result.UpdatedFiles) == 0 {
			if result.Source != "" {
				fmt.Fprintf(stdout, "Managed entry blocks already matched source: %s\n", result.Source)
			} else {
				fmt.Fprintln(stdout, "Managed entry blocks are already consistent.")
			}
			return nil
		}

		fmt.Fprintf(stdout, "Synced managed entry blocks from %s\n", result.Source)
		for _, path := range result.UpdatedFiles {
			fmt.Fprintf(stdout, "- %s\n", path)
		}
		return nil
	case "-h", "--help", "help":
		writeEntryUsage(stdout)
		return nil
	default:
		writeEntryUsage(stderr)
		return fmt.Errorf("unknown entry subcommand %q", args[0])
	}
}

func runRegistry(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		writeRegistryUsage(stderr)
		return errors.New("missing registry subcommand")
	}

	switch args[0] {
	case "validate":
		fs := flag.NewFlagSet("registry validate", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}

		result, err := projectstandards.ValidateRegistry(mustAbs(*repoRoot))
		if err != nil {
			return err
		}

		if len(result.Diagnostics) == 0 {
			fmt.Fprintf(stdout, "Project standards registry is valid. active_entries=%d\n", len(result.Entries))
			return nil
		}

		fmt.Fprintf(stdout, "Project standards registry is invalid. issues=%d\n", len(result.Diagnostics))
		for _, diagnostic := range result.Diagnostics {
			fmt.Fprintf(stdout, "- %s\n", diagnostic)
		}
		return errors.New("project standards registry validation failed")
	case "-h", "--help", "help":
		writeRegistryUsage(stdout)
		return nil
	default:
		writeRegistryUsage(stderr)
		return fmt.Errorf("unknown registry subcommand %q", args[0])
	}
}

func runRepositoryMapping(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		writeRepositoryMappingUsage(stderr)
		return errors.New("missing repository-mapping subcommand")
	}

	switch args[0] {
	case "validate":
		fs := flag.NewFlagSet("repository-mapping validate", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}

		result, err := repositorymapping.Validate(mustAbs(*repoRoot))
		if err != nil {
			return err
		}

		if result.Valid() {
			fmt.Fprintln(stdout, "Repository mapping is valid.")
			return nil
		}

		fmt.Fprintf(stdout, "Repository mapping is invalid. issues=%d\n", len(result.Diagnostics))
		for _, diagnostic := range result.Diagnostics {
			fmt.Fprintf(stdout, "- %s\n", diagnostic)
		}
		return errors.New("repository mapping validation failed")
	case "-h", "--help", "help":
		writeRepositoryMappingUsage(stdout)
		return nil
	default:
		writeRepositoryMappingUsage(stderr)
		return fmt.Errorf("unknown repository-mapping subcommand %q", args[0])
	}
}

func runReview(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		writeReviewUsage(stderr)
		return errors.New("missing review subcommand")
	}

	switch args[0] {
	case "scope":
		fs := flag.NewFlagSet("review scope", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		flow := fs.String("flow", reviewrun.FlowSpecFlowReview, "review flow")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := requireReviewFlow(*flow, stderr); err != nil {
			return err
		}
		return writeReviewScope(stdout, mustAbs(*repoRoot), *flow)
	case "collect-default-scope":
		fs := flag.NewFlagSet("review collect-default-scope", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		flow := fs.String("flow", "", "review flow")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := requireReviewFlow(*flow, stderr); err != nil {
			return err
		}

		return writeReviewScope(stdout, mustAbs(*repoRoot), *flow)
	case "run-init":
		fs := flag.NewFlagSet("review run-init", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		flow := fs.String("flow", "", "review flow")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := requireReviewFlow(*flow, stderr); err != nil {
			return err
		}
		result, err := reviewrun.Init(mustAbs(*repoRoot), *flow, time.Now().UTC())
		if err != nil {
			return err
		}
		if result.Created {
			fmt.Fprintf(stdout, "Review run-state created: %s\n", result.File)
			if len(result.DeletedFiles) > 0 {
				fmt.Fprintf(stdout, "Deleted run-state files (%d):\n", len(result.DeletedFiles))
				for _, deleted := range result.DeletedFiles {
					fmt.Fprintf(stdout, "- %s | reason=%s\n", deleted.File, deleted.Reason)
				}
			}
			return nil
		}
		fmt.Fprintf(stdout, "Review run-state reused: %s\n", result.File)
		return nil
	case "run-validate":
		fs := flag.NewFlagSet("review run-validate", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		flow := fs.String("flow", "", "review flow")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := requireReviewFlow(*flow, stderr); err != nil {
			return err
		}
		absRepoRoot := mustAbs(*repoRoot)
		file, err := reviewrun.FixedRunStateFile(absRepoRoot, *flow)
		if err != nil {
			return err
		}
		result := reviewrun.ValidateFile(absRepoRoot, *flow, file, time.Now().UTC())
		if result.Valid {
			fmt.Fprintf(stdout, "Review run-state is valid: %s\n", result.File)
			return nil
		}
		fmt.Fprintf(stdout, "Review run-state is invalid: %s\n", result.File)
		for _, diagnostic := range result.Diagnostics {
			fmt.Fprintf(stdout, "- %s\n", diagnostic)
		}
		return errors.New("review run-state validation failed")
	case "run-refresh":
		fs := flag.NewFlagSet("review run-refresh", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		flow := fs.String("flow", "", "review flow")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := requireReviewFlow(*flow, stderr); err != nil {
			return err
		}
		absRepoRoot := mustAbs(*repoRoot)
		file, err := reviewrun.FixedRunStateFile(absRepoRoot, *flow)
		if err != nil {
			return err
		}
		result, err := reviewrun.Refresh(absRepoRoot, *flow, file, time.Now().UTC())
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Review run-state refreshed: %s\n", result.File)
		fmt.Fprintf(stdout, "last_updated_at: %s\n", result.LastUpdatedAtUTC)
		writeList(stdout, "Changed fingerprint slices", result.ChangedSlices)
		writeList(stdout, "Stale slices", result.StaleSlices)
		writeList(stdout, "Missing inputs", result.MissingInputs)
		return nil
	case "run-touch":
		fs := flag.NewFlagSet("review run-touch", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		flow := fs.String("flow", "", "review flow")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if err := requireReviewFlow(*flow, stderr); err != nil {
			return err
		}
		absRepoRoot := mustAbs(*repoRoot)
		file, err := reviewrun.FixedRunStateFile(absRepoRoot, *flow)
		if err != nil {
			return err
		}
		result, err := reviewrun.Touch(absRepoRoot, *flow, file, time.Now().UTC())
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Review run-state touched: %s\n", result.File)
		fmt.Fprintf(stdout, "last_updated_at: %s\n", result.LastUpdatedAtUTC)
		return nil
	case "-h", "--help", "help":
		writeReviewUsage(stdout)
		return nil
	default:
		writeReviewUsage(stderr)
		return fmt.Errorf("unknown review subcommand %q", args[0])
	}
}

func runProcess(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		writeProcessUsage(stderr)
		return errors.New("missing process subcommand")
	}

	switch args[0] {
	case "cleanup-fallback":
		fs := flag.NewFlagSet("process cleanup-fallback", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		objectType := fs.String("object-type", "", "formal object type: unit")
		object := fs.String("object", "", "formal object name")
		fromCommand := fs.String("from-command", "", "origin command")
		reason := fs.String("reason", "", "fallback reason code")
		failureLayer := fs.String("failure-layer", "", "failure layer")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*objectType) == "" || strings.TrimSpace(*object) == "" || strings.TrimSpace(*fromCommand) == "" || strings.TrimSpace(*reason) == "" || strings.TrimSpace(*failureLayer) == "" {
			writeProcessUsage(stderr)
			return errors.New("object-type, object, from-command, reason, and failure-layer are required")
		}

		result, err := processcleanup.ApplyObjectFallback(mustAbs(*repoRoot), *objectType, *object, *fromCommand, *reason, *failureLayer)
		if err != nil {
			return err
		}

		fmt.Fprintf(stdout, "Applied fallback cleanup for %s %s\n", result.ObjectType, result.Object)
		fmt.Fprintf(stdout, "From command: %s\n", result.FromCommand)
		fmt.Fprintf(stdout, "Fallback reason: %s\n", result.Reason)
		fmt.Fprintf(stdout, "Failure layer: %s\n", result.FailureLayer)
		fmt.Fprintf(stdout, "Next Command: %s\n", result.NextCommand)
		writeList(stdout, "Deleted files", result.DeletedFiles)
		writeList(stdout, "Missing files", result.MissingFiles)
		fmt.Fprintf(stdout, "Status file updated: %t\n", result.StatusUpdated)
		return nil
	case "cleanup-success":
		fs := flag.NewFlagSet("process cleanup-success", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		objectType := fs.String("object-type", "", "formal object type: unit")
		object := fs.String("object", "", "formal object name")
		mode := fs.String("mode", "", "success cleanup mode")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*objectType) == "" || strings.TrimSpace(*object) == "" || strings.TrimSpace(*mode) == "" {
			writeProcessUsage(stderr)
			return errors.New("object-type, object, and mode are required")
		}

		result, err := processcleanup.ApplyObjectSuccessCleanup(mustAbs(*repoRoot), *objectType, *object, *mode)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Applied success cleanup for %s %s\n", result.ObjectType, result.Object)
		fmt.Fprintf(stdout, "Mode: %s\n", result.Mode)
		writeList(stdout, "Deleted files", result.DeletedFiles)
		writeList(stdout, "Missing files", result.MissingFiles)
		return nil
	case "-h", "--help", "help":
		writeProcessUsage(stdout)
		return nil
	default:
		writeProcessUsage(stderr)
		return fmt.Errorf("unknown process subcommand %q", args[0])
	}
}

func runRule(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		writeRuleUsage(stderr)
		return errors.New("missing rule subcommand")
	}

	switch args[0] {
	case "sync-impact":
		fs := flag.NewFlagSet("rule sync-impact", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		modules := fs.String("units", "", "comma-separated formal units")
		ruleRefs := fs.String("rule-refs", "", "comma-separated rule version refs")
		ruleIDs := fs.String("rule-ids", "", "comma-separated rule ids")
		stableLandingUnit := fs.String("stable-landing-unit", "", "formal unit whose same-round stable landing should not invalidate itself")
		stableLandingRuleRefs := fs.String("stable-landing-rule-refs", "", "comma-separated exact rule refs written by the same-round stable landing")
		retargetedUnits := fs.String("retargeted-units", "", "comma-separated candidate units retargeted to same-round stable landing rule refs")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}

		result, err := rulesync.SyncImpact(mustAbs(*repoRoot), rulesync.Options{
			Modules:               parseCSV(*modules),
			RuleRefs:              parseCSV(*ruleRefs),
			RuleIDs:               parseCSV(*ruleIDs),
			StableLandingModule:   strings.TrimSpace(*stableLandingUnit),
			StableLandingRuleRefs: parseCSV(*stableLandingRuleRefs),
			RetargetedUnits:       parseCSV(*retargetedUnits),
		})
		if err != nil {
			return err
		}

		writeList(stdout, "Scoped units", result.ScopedModules)
		writeList(stdout, "Scoped rule refs", result.ScopedRuleRefs)
		writeList(stdout, "Scoped rule ids", result.ScopedRuleIDs)
		fmt.Fprintf(stdout, "Stable landing unit: %s\n", noneIfEmpty(result.StableLandingModule))
		writeList(stdout, "Stable landing rule refs", result.StableLandingRuleRefs)
		writeList(stdout, "Retargeted units", result.RetargetedUnits)
		fmt.Fprintf(stdout, "Unit results (%d):\n", len(result.ModuleResults))
		if len(result.ModuleResults) == 0 {
			fmt.Fprintln(stdout, "- none")
		}
		for _, item := range result.ModuleResults {
			fmt.Fprintf(stdout, "- %s | layer=%s | outcome=%s | next=%s | reason=%s | status_updated=%t\n", item.Module, item.ActiveLayer, item.Outcome, item.NextCommand, noneIfEmpty(item.FallbackReasonCode), item.StatusUpdated)
			for _, diagnostic := range item.Diagnostics {
				fmt.Fprintf(stdout, "  diagnostic: %s\n", diagnostic)
			}
			for _, path := range item.DeletedFiles {
				fmt.Fprintf(stdout, "  deleted: %s\n", path)
			}
			for _, path := range item.MissingFiles {
				fmt.Fprintf(stdout, "  missing: %s\n", path)
			}
		}
		return nil
	case "consumers":
		fs := flag.NewFlagSet("rule consumers", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		ruleID := fs.String("rule-id", "", "rule id")
		ruleRef := fs.String("rule-ref", "", "exact rule version ref")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := rulesync.Consumers(mustAbs(*repoRoot), rulesync.ConsumerOptions{
			RuleID:  *ruleID,
			RuleRef: *ruleRef,
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Rule id: %s\n", noneIfEmpty(result.RuleID))
		fmt.Fprintf(stdout, "Rule ref: %s\n", noneIfEmpty(result.RuleRef))
		fmt.Fprintf(stdout, "Consumers (%d):\n", len(result.Consumers))
		if len(result.Consumers) == 0 {
			fmt.Fprintln(stdout, "- none")
		}
		for _, consumer := range result.Consumers {
			fmt.Fprintf(stdout, "- %s:%s | layer=%s | file=%s | refs=%s\n", consumer.ObjectType, consumer.Object, consumer.ActiveLayer, consumer.FileRef, strings.Join(defaultListValue(consumer.RuleRefs), ", "))
		}
		return nil
	case "release-version":
		fs := flag.NewFlagSet("rule release-version", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		ruleID := fs.String("rule-id", "", "rule id")
		fromRef := fs.String("from-ref", "", "old stable rule version ref")
		toRef := fs.String("to-ref", "", "new stable rule version ref")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := rulesync.ReleaseVersion(mustAbs(*repoRoot), rulesync.ReleaseVersionOptions{
			RuleID:  *ruleID,
			FromRef: *fromRef,
			ToRef:   *toRef,
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Released rule version: %s from %s to %s\n", result.RuleID, result.FromRef, result.ToRef)
		writeList(stdout, "Candidate current-layer objects updated", result.CandidateUpdated)
		writeList(stdout, "Stable current-layer objects forked", result.StableForked)
		writeList(stdout, "Appendix files retargeted", result.AppendixRetargeted)
		writeList(stdout, "Candidate appendices removed", result.AppendixRemoved)
		writeList(stdout, "Process files removed", result.ProcessFilesRemoved)
		writeList(stdout, "Synced units", result.Sync.ScopedModules)
		return nil
	case "-h", "--help", "help":
		writeRuleUsage(stdout)
		return nil
	default:
		writeRuleUsage(stderr)
		return fmt.Errorf("unknown rule subcommand %q", args[0])
	}
}

func runSnapshot(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		writeSnapshotUsage(stderr)
		return errors.New("missing snapshot subcommand")
	}

	switch args[0] {
	case "rebuild":
		fs := flag.NewFlagSet("snapshot rebuild", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		objectType := fs.String("object-type", "", "formal object type: unit")
		object := fs.String("object", "", "formal object name")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*objectType) == "" || strings.TrimSpace(*object) == "" {
			writeSnapshotUsage(stderr)
			return errors.New("object-type and object are required")
		}
		result, err := snapshot.RebuildCurrentObject(mustAbs(*repoRoot), *objectType, *object)
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, snapshot.Render(result))
		return nil
	case "validate-process":
		fs := flag.NewFlagSet("snapshot validate-process", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		objectType := fs.String("object-type", "", "formal object type: unit")
		object := fs.String("object", "", "formal object name")
		processKind := fs.String("process", "", "check | plan | verify")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*objectType) == "" || strings.TrimSpace(*object) == "" || strings.TrimSpace(*processKind) == "" {
			writeSnapshotUsage(stderr)
			return errors.New("object-type, object, and process are required")
		}
		result, err := snapshot.ValidateProcessFileForObject(mustAbs(*repoRoot), *objectType, *object, *processKind)
		if err != nil {
			return err
		}
		if result.Valid {
			fmt.Fprintf(stdout, "Process snapshot is valid. file=%s\n", result.ProcessFile)
			return nil
		}
		fmt.Fprintf(stdout, "Process snapshot is invalid. file=%s\n", result.ProcessFile)
		for _, mismatch := range result.Mismatches {
			fmt.Fprintf(stdout, "- %s\n", mismatch)
		}
		fmt.Fprintf(stdout, "Failure layer: %s\n", result.FailureLayer)
		if result.NextCommand != "" {
			fmt.Fprintf(stdout, "Recommended Next Command: %s\n", result.NextCommand)
		}
		return errors.New("process snapshot mismatch")
	case "-h", "--help", "help":
		writeSnapshotUsage(stdout)
		return nil
	default:
		writeSnapshotUsage(stderr)
		return fmt.Errorf("unknown snapshot subcommand %q", args[0])
	}
}

func runStatus(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		writeStatusUsage(stderr)
		return errors.New("missing status subcommand")
	}

	switch args[0] {
	case "set-object":
		fs := flag.NewFlagSet("status set-object", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		objectType := fs.String("type", "", "object type")
		object := fs.String("object", "", "formal object id")
		stable := fs.String("stable", "", "yes | no")
		candidate := fs.String("candidate", "", "yes | no")
		activeLayer := fs.String("active-layer", "", "stable | candidate")
		nextCommand := fs.String("next-command", "", "next command")
		notes := fs.String("notes", "", "notes text")
		create := fs.Bool("create", false, "create row when missing")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*objectType) == "" || strings.TrimSpace(*object) == "" || strings.TrimSpace(*stable) == "" || strings.TrimSpace(*candidate) == "" || strings.TrimSpace(*activeLayer) == "" || strings.TrimSpace(*nextCommand) == "" {
			writeStatusUsage(stderr)
			return errors.New("type, object, stable, candidate, active-layer, and next-command are required")
		}
		if *stable != "yes" && *stable != "no" {
			return fmt.Errorf("stable must be yes or no")
		}
		if *candidate != "yes" && *candidate != "no" {
			return fmt.Errorf("candidate must be yes or no")
		}
		if *activeLayer != "stable" && *activeLayer != "candidate" {
			return fmt.Errorf("active-layer must be stable or candidate")
		}

		updated, err := statusfile.UpsertObjectStatus(mustAbs(*repoRoot), statusfile.ObjectStatus{
			ObjectType:  strings.TrimSpace(*objectType),
			Object:      strings.TrimSpace(*object),
			Stable:      strings.TrimSpace(*stable),
			Candidate:   strings.TrimSpace(*candidate),
			ActiveLayer: strings.TrimSpace(*activeLayer),
			NextCommand: strings.TrimSpace(*nextCommand),
			Notes:       strings.TrimSpace(*notes),
		}, *create)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Status row upserted: %t\n", updated)
		fmt.Fprintf(stdout, "Object Type: %s\n", strings.TrimSpace(*objectType))
		fmt.Fprintf(stdout, "Object: %s\n", strings.TrimSpace(*object))
		fmt.Fprintf(stdout, "Stable: %s\n", strings.TrimSpace(*stable))
		fmt.Fprintf(stdout, "Candidate: %s\n", strings.TrimSpace(*candidate))
		fmt.Fprintf(stdout, "Active Layer: %s\n", strings.TrimSpace(*activeLayer))
		fmt.Fprintf(stdout, "Next Command: %s\n", strings.TrimSpace(*nextCommand))
		fmt.Fprintf(stdout, "Notes: %s\n", noneIfEmpty(strings.TrimSpace(*notes)))
		return nil
	case "set-unit":
		fs := flag.NewFlagSet("status set-unit", flag.ContinueOnError)
		fs.SetOutput(stderr)
		repoRoot := fs.String("repo-root", ".", "repository root")
		module := fs.String("unit", "", "formal unit name")
		stable := fs.String("stable", "", "yes | no")
		candidate := fs.String("candidate", "", "yes | no")
		activeLayer := fs.String("active-layer", "", "stable | candidate")
		nextCommand := fs.String("next-command", "", "next command")
		notes := fs.String("notes", "", "notes text")
		create := fs.Bool("create", false, "create row when missing")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*module) == "" || strings.TrimSpace(*stable) == "" || strings.TrimSpace(*candidate) == "" || strings.TrimSpace(*activeLayer) == "" || strings.TrimSpace(*nextCommand) == "" {
			writeStatusUsage(stderr)
			return errors.New("unit, stable, candidate, active-layer, and next-command are required")
		}
		if *stable != "yes" && *stable != "no" {
			return fmt.Errorf("stable must be yes or no")
		}
		if *candidate != "yes" && *candidate != "no" {
			return fmt.Errorf("candidate must be yes or no")
		}
		if *activeLayer != "stable" && *activeLayer != "candidate" {
			return fmt.Errorf("active-layer must be stable or candidate")
		}

		updated, err := statusfile.UpsertModuleStatus(mustAbs(*repoRoot), statusfile.ModuleStatus{
			Module:      strings.TrimSpace(*module),
			Stable:      strings.TrimSpace(*stable),
			Candidate:   strings.TrimSpace(*candidate),
			ActiveLayer: strings.TrimSpace(*activeLayer),
			NextCommand: strings.TrimSpace(*nextCommand),
			Notes:       strings.TrimSpace(*notes),
		}, *create)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Status row upserted: %t\n", updated)
		fmt.Fprintf(stdout, "Unit: %s\n", strings.TrimSpace(*module))
		fmt.Fprintf(stdout, "Stable: %s\n", strings.TrimSpace(*stable))
		fmt.Fprintf(stdout, "Candidate: %s\n", strings.TrimSpace(*candidate))
		fmt.Fprintf(stdout, "Active Layer: %s\n", strings.TrimSpace(*activeLayer))
		fmt.Fprintf(stdout, "Next Command: %s\n", strings.TrimSpace(*nextCommand))
		fmt.Fprintf(stdout, "Notes: %s\n", noneIfEmpty(strings.TrimSpace(*notes)))
		return nil
	case "-h", "--help", "help":
		writeStatusUsage(stdout)
		return nil
	default:
		writeStatusUsage(stderr)
		return fmt.Errorf("unknown status subcommand %q", args[0])
	}
}

func writeRootUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl <command> [subcommand] [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  init     Install specFlow files from manifest")
	fmt.Fprintln(w, "  doctor   Check installed specFlow structure")
	fmt.Fprintln(w, "  upgrade  Refresh framework-managed files")
	fmt.Fprintln(w, "  build-release Build platform binaries into specflow/tooling/bin")
	fmt.Fprintln(w, "  command  Run standard-command mechanical preflight checks and close commands")
	fmt.Fprintln(w, "  entry    Check or sync registered entry-file managed blocks")
	fmt.Fprintln(w, "  registry Validate docs/project_standards/_registry.md")
	fmt.Fprintln(w, "  repository-mapping Validate docs/specs/repository_mapping.md")
	fmt.Fprintln(w, "  review   Collect governance review scope or maintain run-state files")
	fmt.Fprintln(w, "  process  Execute deterministic fallback cleanup")
	fmt.Fprintln(w, "  rule     Execute deterministic rule-impact reconciliation helpers")
	fmt.Fprintln(w, "  snapshot Rebuild or compare process snapshot fields")
	fmt.Fprintln(w, "  status   Apply deterministic _status.md row writeback")
}

func writeCommandUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl command close --command COMMAND --object-type unit --object OBJECT --outcome OUTCOME [--reason CODE] [--failure-layer LAYER] [--candidate-intent repair|change] [--stable-before yes|no] [--notes TEXT] [--apply] [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl command preflight --command COMMAND --object-type unit --object OBJECT [--repo-root PATH]")
}

func writeEntryUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl entry check [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl entry sync [--repo-root PATH] [--source FILE]")
}

func writeRegistryUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl registry validate [--repo-root PATH]")
}

func writeRepositoryMappingUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl repository-mapping validate [--repo-root PATH]")
}

func writeReviewUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl review scope [--flow spec_flow_review|spec_flow_design_review] [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl review collect-default-scope --flow spec_flow_review|spec_flow_design_review [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl review run-init --flow spec_flow_review|spec_flow_design_review [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl review run-validate --flow spec_flow_review|spec_flow_design_review [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl review run-refresh --flow spec_flow_review|spec_flow_design_review [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl review run-touch --flow spec_flow_review|spec_flow_design_review [--repo-root PATH]")
}

func writeReviewScope(stdout io.Writer, repoRoot, flow string) error {
	var scope reviewscope.SpecFlowScope
	var err error
	switch flow {
	case reviewrun.FlowSpecFlowReview:
		scope, err = reviewscope.CollectDefaultSpecFlowScope(repoRoot)
	case reviewrun.FlowSpecFlowDesignReview:
		scope, err = reviewscope.CollectDefaultSpecFlowDesignScope(repoRoot)
	default:
		return fmt.Errorf("unsupported review flow %q", flow)
	}
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Review flow: %s\n", flow)
	fmt.Fprintf(stdout, "Review profile: %s\n", scope.Profile)
	writeList(stdout, "Framework guideline files", scope.FrameworkGuidelineFiles)
	writeList(stdout, "Command files", scope.CommandFiles)
	writeList(stdout, "Candidate intent files", scope.CandidateIntentFiles)
	writeList(stdout, "Guidance skill files", scope.GuidanceSkillFiles)
	writeList(stdout, "Rule-governance minimum files", scope.RuleGovernanceFiles)
	writeList(stdout, "Template governance files", scope.TemplateGovernanceFiles)
	writeList(stdout, "Template project-instance files", scope.TemplateProjectInstanceFiles)
	writeList(stdout, "Template entry files", scope.TemplateEntryFiles)
	writeList(stdout, "Project entry files", scope.ProjectEntryFiles)
	writeList(stdout, "Agent operability files", scope.AgentOperabilityFiles)
	writeList(stdout, "Project-instance compatibility files", scope.ProjectInstanceCompatibilityFiles)
	writeList(stdout, "Project registry files", scope.ProjectRegistryFiles)
	writeList(stdout, "Project registry diagnostics", scope.RegistryDiagnostics)
	writeList(stdout, "Tooling contract files", scope.ToolingContractFiles)
	writeList(stdout, "Tooling source files", scope.ToolingSourceFiles)
	if len(scope.ToolingScriptFiles) > 0 {
		writeList(stdout, "Tooling script files", scope.ToolingScriptFiles)
	}
	if len(scope.ToolingRuntimeFiles) > 0 {
		writeList(stdout, "Tooling runtime files", scope.ToolingRuntimeFiles)
	}
	writeList(stdout, "Active project-local governance-input files", scope.ActiveProjectStandardFiles)
	return nil
}

func requireReviewFlow(flow string, stderr io.Writer) error {
	flow = strings.TrimSpace(flow)
	if flow == "" {
		writeReviewUsage(stderr)
		return errors.New("flow is required")
	}
	for _, supported := range reviewrun.ConfiguredFlows() {
		if flow == supported {
			return nil
		}
	}
	writeReviewUsage(stderr)
	return fmt.Errorf("unsupported review flow %q", flow)
}

func writeProcessUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl process cleanup-fallback --object-type unit --object OBJECT --from-command COMMAND --reason CODE --failure-layer LAYER [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl process cleanup-success --object-type unit --object OBJECT --mode unit_fork|unit_promote [--repo-root PATH]")
}

func writeRuleUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl rule sync-impact (--rule-refs c_b_rule_x@0.1.0 | --rule-ids b_rule_x) [--units unit_a,unit_b] [--stable-landing-unit unit_a --stable-landing-rule-refs s_b_rule_x@1.0.0] [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl rule sync-impact --rule-refs c_b_rule_x@0.1.0,s_b_rule_x@0.1.0 --stable-landing-unit unit_a --stable-landing-rule-refs s_b_rule_x@0.1.0 --retargeted-units unit_b [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl rule consumers (--rule-id b_rule_x | --rule-ref s_b_rule_x@1.0.0) [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl rule release-version --rule-id b_rule_x --from-ref s_b_rule_x@0.3.0 --to-ref s_b_rule_x@0.4.0 [--repo-root PATH]")
}

func writeSnapshotUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl snapshot rebuild --object-type unit --object OBJECT [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl snapshot validate-process --object-type unit --object OBJECT --process check|plan|verify [--repo-root PATH]")
}

func writeStatusUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  specflowctl status set-object --type unit --object OBJECT --stable yes|no --candidate yes|no --active-layer stable|candidate --next-command COMMAND [--notes TEXT] [--create] [--repo-root PATH]")
	fmt.Fprintln(w, "  specflowctl status set-unit --unit UNIT --stable yes|no --candidate yes|no --active-layer stable|candidate --next-command COMMAND [--notes TEXT] [--create] [--repo-root PATH]")
}

func writeList(w io.Writer, title string, items []string) {
	fmt.Fprintf(w, "%s (%d):\n", title, len(items))
	if len(items) == 0 {
		fmt.Fprintln(w, "- none")
		return
	}
	for _, item := range items {
		fmt.Fprintf(w, "- %s\n", item)
	}
}

func mustAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}

func resolvePath(base, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(base, filepath.FromSlash(path))
}

func parseCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		result = append(result, part)
	}
	return result
}

func noneIfEmpty(value string) string {
	if strings.TrimSpace(value) == "" {
		return "none"
	}
	return value
}

func checkCommandForObjectType(objectType string) string {
	return "unit_check"
}

func defaultListValue(items []string) []string {
	if len(items) == 0 {
		return []string{"none"}
	}
	return items
}
