package commandclose

import (
	"fmt"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/commandpreflight"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/processcleanup"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

const (
	cleanupNone     = "none"
	cleanupFallback = "fallback"
	cleanupSuccess  = "success"
)

type Options struct {
	RepoRoot        string
	Command         string
	ObjectType      string
	Object          string
	Outcome         string
	Reason          string
	FailureLayer    string
	CandidateIntent string
	Notes           string
	StableBefore    string
	Apply           bool
}

type Result struct {
	Command                   string
	ObjectType                string
	Object                    string
	Outcome                   string
	Applied                   bool
	StatusBeforePresent       bool
	StatusBefore              statusfile.ObjectStatus
	StatusAfter               statusfile.ObjectStatus
	InputValidationAction     string
	InputValidatedProcesses   []commandpreflight.Process
	InputValidationMismatches []string
	ValidationAction          string
	CleanupAction             string
	FallbackCleanup           processcleanup.CleanupResult
	SuccessCleanup            processcleanup.SuccessCleanupResult
	StatusUpdated             bool
	ValidationMismatches      []string
}

type transition struct {
	Status            statusfile.ObjectStatus
	ValidationProcess string
	CleanupKind       string
	CleanupMode       string
	FailureLayer      string
	Reason            string
}

func Close(opts Options) (Result, error) {
	opts = normalizeOptions(opts)
	if err := validateRequiredOptions(opts); err != nil {
		return Result{}, err
	}

	before, present, err := lookupStatus(opts.RepoRoot, opts.ObjectType, opts.Object)
	if err != nil {
		return Result{}, err
	}
	if isCreateCommand(opts.Command) {
		if present {
			return Result{}, fmt.Errorf("%s %q is already registered in docs/specs/_status.md", opts.ObjectType, opts.Object)
		}
	} else {
		if !present {
			return Result{}, fmt.Errorf("%s %q is not registered in docs/specs/_status.md", opts.ObjectType, opts.Object)
		}
		if before.NextCommand != opts.Command {
			return Result{}, fmt.Errorf("status next command mismatch: actual=%s expected=%s", before.NextCommand, opts.Command)
		}
	}

	trans, err := determineTransition(opts, before, present)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Command:               opts.Command,
		ObjectType:            opts.ObjectType,
		Object:                opts.Object,
		Outcome:               opts.Outcome,
		Applied:               opts.Apply,
		StatusBeforePresent:   present,
		StatusBefore:          before,
		StatusAfter:           trans.Status,
		InputValidationAction: inputValidationAction(opts.Command, trans),
		ValidationAction:      validationAction(trans.ValidationProcess),
		CleanupAction:         cleanupAction(trans),
	}

	if shouldValidateInput(opts.Command, trans) {
		preflight := commandpreflight.Run(opts.RepoRoot, opts.Command, opts.ObjectType, opts.Object)
		result.InputValidatedProcesses = preflight.ValidatedProcesses
		if !preflight.MayContinue {
			result.InputValidationMismatches = collectPreflightDiagnostics(preflight)
			return result, fmt.Errorf("command close input preflight failed for %s %s/%s: %s", opts.Command, opts.ObjectType, opts.Object, strings.Join(result.InputValidationMismatches, "; "))
		}
	}

	if trans.ValidationProcess != "" {
		mismatches, err := validateProcess(opts.RepoRoot, opts.ObjectType, opts.Object, trans.ValidationProcess)
		if err != nil {
			return result, err
		}
		result.ValidationMismatches = mismatches
	}

	if !opts.Apply {
		return result, nil
	}

	switch trans.CleanupKind {
	case cleanupFallback:
		cleanup, err := processcleanup.ApplyObjectFallback(opts.RepoRoot, opts.ObjectType, opts.Object, opts.Command, trans.Reason, trans.FailureLayer)
		if err != nil {
			return result, err
		}
		result.FallbackCleanup = cleanup
		updated, err := statusfile.UpsertObjectStatus(opts.RepoRoot, trans.Status, false)
		if err != nil {
			return result, err
		}
		result.StatusUpdated = updated
	case cleanupSuccess:
		updated, err := statusfile.UpsertObjectStatus(opts.RepoRoot, trans.Status, isCreateCommand(opts.Command))
		if err != nil {
			return result, err
		}
		result.StatusUpdated = updated
		cleanup, err := processcleanup.ApplyObjectSuccessCleanup(opts.RepoRoot, opts.ObjectType, opts.Object, trans.CleanupMode)
		if err != nil {
			return result, err
		}
		result.SuccessCleanup = cleanup
	default:
		updated, err := statusfile.UpsertObjectStatus(opts.RepoRoot, trans.Status, isCreateCommand(opts.Command))
		if err != nil {
			return result, err
		}
		result.StatusUpdated = updated
	}

	return result, nil
}

func normalizeOptions(opts Options) Options {
	opts.RepoRoot = strings.TrimSpace(opts.RepoRoot)
	opts.Command = strings.TrimSpace(opts.Command)
	opts.ObjectType = strings.TrimSpace(opts.ObjectType)
	opts.Object = strings.TrimSpace(opts.Object)
	opts.Outcome = strings.TrimSpace(opts.Outcome)
	opts.Reason = strings.TrimSpace(opts.Reason)
	opts.FailureLayer = strings.TrimSpace(opts.FailureLayer)
	opts.CandidateIntent = strings.TrimSpace(opts.CandidateIntent)
	opts.Notes = strings.TrimSpace(opts.Notes)
	opts.StableBefore = strings.TrimSpace(opts.StableBefore)
	return opts
}

func validateRequiredOptions(opts Options) error {
	if opts.RepoRoot == "" {
		return fmt.Errorf("repo root is required")
	}
	if opts.Command == "" || opts.ObjectType == "" || opts.Object == "" || opts.Outcome == "" {
		return fmt.Errorf("command, object-type, object, and outcome are required")
	}
	if opts.ObjectType != "unit" && opts.ObjectType != "scenario" {
		return fmt.Errorf("object-type must be unit or scenario")
	}
	if err := validateCommandMatchesObjectType(opts.Command, opts.ObjectType); err != nil {
		return err
	}
	if err := validateOutcomeFlags(opts); err != nil {
		return err
	}
	return nil
}

func validateCommandMatchesObjectType(command, objectType string) error {
	switch {
	case strings.HasPrefix(command, "unit_"):
		if objectType != "unit" {
			return fmt.Errorf("command %q requires object-type unit", command)
		}
	case strings.HasPrefix(command, "scenario_"):
		if objectType != "scenario" {
			return fmt.Errorf("command %q requires object-type scenario", command)
		}
	default:
		return fmt.Errorf("unsupported command %q", command)
	}
	return nil
}

func validateOutcomeFlags(opts Options) error {
	if opts.CandidateIntent != "" && opts.CandidateIntent != "repair" && opts.CandidateIntent != "change" {
		return fmt.Errorf("candidate-intent must be repair or change")
	}
	if opts.StableBefore != "" && opts.StableBefore != "yes" && opts.StableBefore != "no" {
		return fmt.Errorf("stable-before must be yes or no")
	}
	if opts.StableBefore != "" && opts.Outcome != "promotion_recovered" {
		return fmt.Errorf("stable-before is only accepted for promotion_recovered")
	}

	switch {
	case opts.Command == "unit_stable_verify" && opts.Outcome == "controlled_repair_required":
		if opts.CandidateIntent == "" {
			return fmt.Errorf("controlled_repair_required requires --candidate-intent repair")
		}
		if opts.CandidateIntent != "repair" {
			return fmt.Errorf("controlled_repair_required requires --candidate-intent repair, got %q", opts.CandidateIntent)
		}
	case opts.Command == "unit_stable_verify" && opts.Outcome == "controlled_change_required":
		if opts.CandidateIntent == "" {
			return fmt.Errorf("controlled_change_required requires --candidate-intent change")
		}
		if opts.CandidateIntent != "change" {
			return fmt.Errorf("controlled_change_required requires --candidate-intent change, got %q", opts.CandidateIntent)
		}
	case strings.HasSuffix(opts.Command, "_promote") && opts.Outcome == "promotion_recovered":
		if opts.StableBefore == "" {
			return fmt.Errorf("promotion_recovered requires --stable-before yes|no")
		}
	}
	if requiresExplicitTruthFallbackReason(opts) && opts.Reason == "" {
		return fmt.Errorf("%s/%s requires --reason", opts.Command, opts.Outcome)
	}
	if opts.Command == "unit_plan" && opts.Outcome == "truth_fallback" && opts.Reason != "" && opts.Reason != "truth_incomplete" {
		return fmt.Errorf("unit_plan/truth_fallback requires --reason truth_incomplete, got %q", opts.Reason)
	}
	return nil
}

func requiresExplicitTruthFallbackReason(opts Options) bool {
	switch opts.Command {
	case "unit_plan", "unit_impl", "unit_verify", "scenario_verify":
		return opts.Outcome == "truth_fallback"
	default:
		return false
	}
}

func lookupStatus(repoRoot, objectType, object string) (statusfile.ObjectStatus, bool, error) {
	statuses, err := statusfile.LoadObjectStatuses(repoRoot)
	if err != nil {
		return statusfile.ObjectStatus{}, false, err
	}
	for _, status := range statuses {
		if status.ObjectType == objectType && status.Object == object {
			return status, true, nil
		}
	}
	return statusfile.ObjectStatus{}, false, nil
}

func determineTransition(opts Options, before statusfile.ObjectStatus, present bool) (transition, error) {
	current := before
	if !present {
		current = statusfile.ObjectStatus{ObjectType: opts.ObjectType, Object: opts.Object}
	}
	current.Notes = chooseNotes(opts.Notes, current.Notes)

	var trans transition
	switch opts.Command {
	case "unit_init":
		trans = exactOutcome(opts, current, "stable_created", status("unit", opts.Object, "yes", "no", "stable", "unit_fork", current.Notes))
	case "unit_new":
		trans = exactOutcome(opts, current, "candidate_created", status("unit", opts.Object, "no", "yes", "candidate", "unit_check", current.Notes))
	case "unit_stable_verify":
		trans = unitStableVerifyTransition(opts, current)
	case "unit_fork":
		trans = exactOutcome(opts, current, "candidate_created", status("unit", opts.Object, current.Stable, "yes", "candidate", "unit_check", current.Notes))
		trans.CleanupKind = cleanupSuccess
		trans.CleanupMode = "unit_fork"
	case "unit_check":
		trans = nextOnlyTransition(opts, current, map[string]string{
			"pass":         "unit_plan",
			"blocked":      "unit_check",
			"fix_required": "unit_check",
			"checkpoint":   "unit_check",
		})
		if opts.Outcome == "pass" {
			trans.ValidationProcess = "check"
		}
	case "unit_plan":
		trans = unitPlanTransition(opts, current)
	case "unit_impl":
		trans = unitImplTransition(opts, current)
	case "unit_verify":
		trans = unitVerifyTransition(opts, current)
	case "unit_promote":
		trans = unitPromoteTransition(opts, current)
	case "scenario_new":
		trans = exactOutcome(opts, current, "candidate_created", status("scenario", opts.Object, "no", "yes", "candidate", "scenario_check", current.Notes))
	case "scenario_stable_verify":
		trans = nextOnlyTransition(opts, current, map[string]string{
			"aligned":             "scenario_fork",
			"not_aligned":         "scenario_stable_verify",
			"evidence_incomplete": "scenario_stable_verify",
		})
	case "scenario_fork":
		trans = exactOutcome(opts, current, "candidate_created", status("scenario", opts.Object, current.Stable, "yes", "candidate", "scenario_check", current.Notes))
		trans.CleanupKind = cleanupSuccess
		trans.CleanupMode = "scenario_fork"
	case "scenario_check":
		trans = nextOnlyTransition(opts, current, map[string]string{
			"pass":         "scenario_verify",
			"blocked":      "scenario_check",
			"fix_required": "scenario_check",
		})
		if opts.Outcome == "pass" {
			trans.ValidationProcess = "check"
		}
	case "scenario_verify":
		trans = scenarioVerifyTransition(opts, current)
	case "scenario_promote":
		trans = scenarioPromoteTransition(opts, current)
	default:
		return transition{}, fmt.Errorf("unsupported command %q", opts.Command)
	}
	if trans.Status.ObjectType == "" {
		return transition{}, fmt.Errorf("unsupported outcome %q for command %q", opts.Outcome, opts.Command)
	}
	if opts.FailureLayer != "" && trans.FailureLayer == "" {
		return transition{}, fmt.Errorf("failure-layer is only accepted for fallback outcomes")
	}
	if opts.FailureLayer != "" && trans.FailureLayer != "" && opts.FailureLayer != trans.FailureLayer {
		return transition{}, fmt.Errorf("failure-layer %q does not match required layer %q for %s/%s", opts.FailureLayer, trans.FailureLayer, opts.Command, opts.Outcome)
	}
	if opts.Reason != "" {
		trans.Reason = opts.Reason
	}
	return trans, nil
}

func exactOutcome(opts Options, current statusfile.ObjectStatus, expected string, after statusfile.ObjectStatus) transition {
	if opts.Outcome != expected {
		return transition{}
	}
	return transition{Status: after, CleanupKind: cleanupNone}
}

func nextOnlyTransition(opts Options, current statusfile.ObjectStatus, outcomes map[string]string) transition {
	next, ok := outcomes[opts.Outcome]
	if !ok {
		return transition{}
	}
	after := current
	after.NextCommand = next
	return transition{Status: after, CleanupKind: cleanupNone}
}

func unitStableVerifyTransition(opts Options, current statusfile.ObjectStatus) transition {
	switch opts.Outcome {
	case "aligned":
		return withNext(current, "unit_fork")
	case "small_repair_required", "evidence_incomplete", "truth_rejudge_required":
		return withNext(current, "unit_stable_verify")
	case "controlled_repair_required":
		if opts.CandidateIntent != "repair" {
			return transition{}
		}
		return withNext(current, "unit_fork")
	case "controlled_change_required":
		if opts.CandidateIntent != "change" {
			return transition{}
		}
		return withNext(current, "unit_fork")
	default:
		return transition{}
	}
}

func unitPlanTransition(opts Options, current statusfile.ObjectStatus) transition {
	switch opts.Outcome {
	case "plan_ready":
		return withNextAndValidation(current, "unit_impl", "plan")
	case "truth_fallback":
		return fallback(current, "unit_check", "truth_layer", opts.Reason)
	case "blocked", "decision_checkpoint":
		return withNext(current, "unit_plan")
	default:
		return transition{}
	}
}

func unitImplTransition(opts Options, current statusfile.ObjectStatus) transition {
	switch opts.Outcome {
	case "ready_for_verify":
		return withNext(current, "unit_verify")
	case "blocked":
		return withNext(current, "unit_impl")
	case "truth_fallback":
		return fallback(current, "unit_check", "truth_layer", opts.Reason)
	case "plan_fallback":
		return fallback(current, "unit_plan", "plan_layer", defaultReason(opts.Reason, "gate_missing"))
	case "gate_fallback":
		return fallback(current, "unit_check", "gate_layer", defaultReason(opts.Reason, "gate_missing"))
	default:
		return transition{}
	}
}

func unitVerifyTransition(opts Options, current statusfile.ObjectStatus) transition {
	switch opts.Outcome {
	case "ready_to_promote":
		return withNextAndValidation(current, "unit_promote", "verify")
	case "implementation_deviation":
		return fallback(current, "unit_impl", "implementation_layer", defaultReason(opts.Reason, "implementation_deviation"))
	case "truth_fallback":
		return fallback(current, "unit_check", "truth_layer", opts.Reason)
	case "evidence_incomplete", "human_verify":
		return withNext(current, "unit_verify")
	default:
		return transition{}
	}
}

func unitPromoteTransition(opts Options, current statusfile.ObjectStatus) transition {
	switch opts.Outcome {
	case "promoted":
		after := status("unit", current.Object, "yes", "no", "stable", "unit_fork", current.Notes)
		return transition{Status: after, ValidationProcess: "verify", CleanupKind: cleanupSuccess, CleanupMode: "unit_promote"}
	case "promotion_recovered":
		stable, ok := stableBefore(opts)
		if !ok {
			return transition{}
		}
		return fallback(status("unit", current.Object, stable, "yes", "candidate", "unit_check", current.Notes), "unit_check", "truth_layer", defaultReason(opts.Reason, "truth_drift"))
	default:
		return promoteInvalidTransition(opts, current, "unit")
	}
}

func scenarioVerifyTransition(opts Options, current statusfile.ObjectStatus) transition {
	switch opts.Outcome {
	case "pass":
		return withNextAndValidation(current, "scenario_promote", "verify")
	case "truth_fallback":
		return fallback(current, "scenario_check", "truth_layer", opts.Reason)
	case "gate_fallback":
		return fallback(current, "scenario_check", "gate_layer", defaultReason(opts.Reason, "gate_missing"))
	case "evidence_incomplete":
		return fallback(current, "scenario_verify", "evidence_layer", defaultReason(opts.Reason, "evidence_incomplete"))
	case "blocked_by_affected_units":
		return withNext(current, "scenario_verify")
	default:
		return transition{}
	}
}

func scenarioPromoteTransition(opts Options, current statusfile.ObjectStatus) transition {
	switch opts.Outcome {
	case "promoted":
		after := status("scenario", current.Object, "yes", "no", "stable", "scenario_fork", current.Notes)
		return transition{Status: after, ValidationProcess: "verify", CleanupKind: cleanupSuccess, CleanupMode: "scenario_promote"}
	case "dependency_not_ready":
		return withNext(current, "scenario_promote")
	case "promotion_recovered":
		stable, ok := stableBefore(opts)
		if !ok {
			return transition{}
		}
		return fallback(status("scenario", current.Object, stable, "yes", "candidate", "scenario_check", current.Notes), "scenario_check", "truth_layer", defaultReason(opts.Reason, "truth_drift"))
	default:
		return promoteInvalidTransition(opts, current, "scenario")
	}
}

func promoteInvalidTransition(opts Options, current statusfile.ObjectStatus, objectType string) transition {
	if !strings.HasPrefix(opts.Outcome, "verify_invalid_") {
		return transition{}
	}
	suffix := strings.TrimPrefix(opts.Outcome, "verify_invalid_")
	nextCheck := objectType + "_check"
	nextVerify := objectType + "_verify"
	switch suffix {
	case "truth":
		return fallback(current, nextCheck, "truth_layer", defaultReason(opts.Reason, "truth_drift"))
	case "binding":
		return fallback(current, nextCheck, "truth_layer", defaultReason(opts.Reason, "binding_drift"))
	case "baseline":
		return fallback(current, nextCheck, "truth_layer", defaultReason(opts.Reason, "baseline_drift"))
	case "rule":
		return fallback(current, nextCheck, "truth_layer", defaultReason(opts.Reason, "rule_drift"))
	case "plan":
		if objectType != "unit" {
			return transition{}
		}
		return fallback(current, "unit_plan", "plan_layer", defaultReason(opts.Reason, "gate_missing"))
	case "implementation":
		if objectType != "unit" {
			return transition{}
		}
		return fallback(current, "unit_impl", "implementation_layer", defaultReason(opts.Reason, "implementation_deviation"))
	case "gate":
		return fallback(current, nextCheck, "gate_layer", defaultReason(opts.Reason, "gate_missing"))
	case "evidence":
		return fallback(current, nextVerify, "evidence_layer", defaultReason(opts.Reason, "evidence_incomplete"))
	default:
		return transition{}
	}
}

func withNext(current statusfile.ObjectStatus, next string) transition {
	after := current
	after.NextCommand = next
	return transition{Status: after, CleanupKind: cleanupNone}
}

func withNextAndValidation(current statusfile.ObjectStatus, next, process string) transition {
	trans := withNext(current, next)
	trans.ValidationProcess = process
	return trans
}

func fallback(current statusfile.ObjectStatus, next, layer, reason string) transition {
	after := current
	after.NextCommand = next
	return transition{Status: after, CleanupKind: cleanupFallback, FailureLayer: layer, Reason: reason}
}

func status(objectType, object, stable, candidate, activeLayer, nextCommand, notes string) statusfile.ObjectStatus {
	return statusfile.ObjectStatus{
		ObjectType:  objectType,
		Object:      object,
		Stable:      stable,
		Candidate:   candidate,
		ActiveLayer: activeLayer,
		NextCommand: nextCommand,
		Notes:       notes,
	}
}

func stableBefore(opts Options) (string, bool) {
	switch opts.StableBefore {
	case "yes", "no":
		return opts.StableBefore, true
	default:
		return "", false
	}
}

func chooseNotes(notes, existing string) string {
	if notes != "" {
		return notes
	}
	return existing
}

func defaultReason(actual, fallbackValue string) string {
	if actual != "" {
		return actual
	}
	return fallbackValue
}

func validateProcess(repoRoot, objectType, object, process string) ([]string, error) {
	result, err := snapshot.ValidateProcessFileForObject(repoRoot, objectType, object, process)
	if err != nil {
		return nil, err
	}
	if !result.Valid {
		return result.Mismatches, fmt.Errorf("required %s process is invalid for %s %s: %s", process, objectType, object, strings.Join(result.Mismatches, "; "))
	}
	return nil, nil
}

func shouldValidateInput(command string, trans transition) bool {
	if trans.CleanupKind == cleanupFallback {
		return false
	}
	switch command {
	case "unit_plan", "unit_impl", "unit_verify", "unit_promote", "scenario_verify", "scenario_promote":
		return true
	default:
		return false
	}
}

func inputValidationAction(command string, trans transition) string {
	if !shouldValidateInput(command, trans) {
		return "none"
	}
	return "command_preflight"
}

func collectPreflightDiagnostics(result commandpreflight.Result) []string {
	var diagnostics []string
	for _, process := range result.ValidatedProcesses {
		for _, diagnostic := range process.Diagnostics {
			diagnostics = append(diagnostics, fmt.Sprintf("%s: %s", process.ProcessKind, diagnostic))
		}
	}
	if len(diagnostics) == 0 {
		diagnostics = append(diagnostics, result.Diagnostics...)
	}
	if len(diagnostics) == 0 && result.FailureLayer != "" {
		diagnostics = append(diagnostics, "failure_layer="+result.FailureLayer)
	}
	return diagnostics
}

func validationAction(process string) string {
	if process == "" {
		return "none"
	}
	return "validate_process:" + process
}

func cleanupAction(trans transition) string {
	switch trans.CleanupKind {
	case cleanupFallback:
		return fmt.Sprintf("fallback:%s:%s", trans.FailureLayer, trans.Reason)
	case cleanupSuccess:
		return "success:" + trans.CleanupMode
	default:
		return "none"
	}
}

func isCreateCommand(command string) bool {
	switch command {
	case "unit_init", "unit_new", "scenario_new":
		return true
	default:
		return false
	}
}
