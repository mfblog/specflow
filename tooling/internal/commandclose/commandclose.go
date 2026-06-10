package commandclose

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/commandpreflight"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/processcleanup"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/unitappendix"
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
	PromotionSummaryFile      string
}

type transition struct {
	Status             statusfile.ObjectStatus
	ValidationProcess  string
	ValidationDecision string
	CleanupKind        string
	CleanupMode        string
	FailureLayer       string
	Reason             string
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
			// unit_check may close when Next Command is unit_verify — this
			// handles spec re-validation during the implementation phase
			// (Next Command=unit_verify, Notes=pending_impl). The state
			// transition table already handles all outcomes correctly:
			//   pass         → unit_verify (no net change)
			//   fix_required → unit_check  (valid regression)
			if opts.Command == "unit_check" && before.NextCommand == "unit_verify" {
				if !strings.Contains(before.Notes, "pending_impl") {
					return Result{}, fmt.Errorf("unit_check re-validation requires Notes=pending_impl: actual Notes=%q", before.Notes)
				}
			} else if opts.Command == "unit_stable_verify" && before.ActiveLayer == "stable" && before.NextCommand != "unit_promote" {
				// unit_stable_verify may close for stable units regardless of
				// Next Command, per status.md "Valid Next Commands" allows
				// semantics — provided Next Command is not unit_promote.
				// The transition table handles all outcomes correctly.
			} else {
				return Result{}, fmt.Errorf("status next command mismatch: actual=%s expected=%s", before.NextCommand, opts.Command)
			}
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
		mismatches, err := validateProcess(opts.RepoRoot, opts.ObjectType, opts.Object, trans.ValidationProcess, trans.ValidationDecision)
		if err != nil {
			return result, err
		}
		result.ValidationMismatches = mismatches
	}
	if err := validateControlledStableVerifyForkIntent(opts); err != nil {
		return result, err
	}
	if err := validateUnitForkAppendixCoverage(opts); err != nil {
		return result, err
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
		if trans.CleanupMode == "unit_promote" {
			summaryFile, err := writeStablePromotionSummary(opts.RepoRoot, opts.ObjectType, opts.Object)
			if err != nil {
				return result, err
			}
			result.PromotionSummaryFile = summaryFile
		}
		updated, err := statusfile.UpsertObjectStatus(opts.RepoRoot, trans.Status, isCreateCommand(opts.Command))
		if err != nil {
			return result, err
		}
		result.StatusUpdated = updated
		cleanup, err := processcleanup.ApplyObjectSuccessCleanup(opts.RepoRoot, opts.ObjectType, opts.Object, trans.CleanupMode)
		result.SuccessCleanup = cleanup
		if err != nil {
			if result.StatusUpdated {
				return result, successCleanupAfterStatusError(opts, trans.CleanupMode, err)
			}
			return result, err
		}
	default:
		updated, err := statusfile.UpsertObjectStatus(opts.RepoRoot, trans.Status, isCreateCommand(opts.Command))
		if err != nil {
			return result, err
		}
		result.StatusUpdated = updated
	}

	return result, nil
}

func successCleanupAfterStatusError(opts Options, cleanupMode string, err error) error {
	return fmt.Errorf("success cleanup failed after status update; fix the filesystem blocker and rerun `specflowctl process cleanup-success --repo-root %s --object-type %s --object %s --mode %s`: %w", opts.RepoRoot, opts.ObjectType, opts.Object, cleanupMode, err)
}

func writeStablePromotionSummary(repoRoot, objectType, object string) (string, error) {
	verifyData, err := snapshot.LoadProcessSnapshot(repoRoot, objectType, object, "verify")
	if err != nil {
		return "", err
	}
	stableSnapshot, err := snapshot.RebuildObjectLayer(repoRoot, objectType, object, "stable")
	if err != nil {
		return "", fmt.Errorf("stable promotion summary requires current stable truth: %w", err)
	}
	summaryRef := snapshot.StablePromotionSummaryFilePath(objectType, object)
	summary := renderStablePromotionSummary(stableSnapshot, verifyData)
	summaryAbs := filepath.Join(repoRoot, filepath.FromSlash(summaryRef))
	if err := os.MkdirAll(filepath.Dir(summaryAbs), 0o755); err != nil {
		return "", fmt.Errorf("create stable promotion summary dir for %s: %w", summaryRef, err)
	}
	if err := os.WriteFile(summaryAbs, []byte(summary), 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", summaryRef, err)
	}
	return summaryRef, nil
}

func renderStablePromotionSummary(stableSnapshot snapshot.Snapshot, verifyData snapshot.ProcessSnapshotData) string {
	var builder strings.Builder
	builder.WriteString("# Stable Promotion Summary\n\n")
	builder.WriteString("```yaml\n")
	builder.WriteString("object_type: " + stableSnapshot.ObjectType + "\n")
	builder.WriteString("object_ref: " + stableSnapshot.Object + "\n")
	builder.WriteString("stable_truth_file_ref: " + stableSnapshot.SpecFileRef + "\n")
	builder.WriteString("stable_truth_version_ref: " + stableSnapshot.SpecVersionRef + "\n")
	builder.WriteString("stable_truth_fingerprint: " + stableSnapshot.SpecFingerprint + "\n")
	builder.WriteString("promotion_verify_result_ref: " + verifyData.ProcessFile + "\n")
	appendPromotionAcceptanceItemSet(&builder, verifyData.AcceptanceItemSet)
	appendPromotionCoverageSummary(&builder, verifyData.AcceptanceEvidence)
	appendPromotionEvidenceRefs(&builder, promotionEvidenceRefs(verifyData))
	builder.WriteString("```\n")
	return builder.String()
}

func appendPromotionAcceptanceItemSet(builder *strings.Builder, items []snapshot.AcceptanceItemEntry) {
	builder.WriteString("acceptance_item_set:\n")
	if len(items) == 0 {
		builder.WriteString("  - id: none\n")
		return
	}
	for _, item := range items {
		builder.WriteString("  - id: " + item.ID + "\n")
	}
}

func appendPromotionCoverageSummary(builder *strings.Builder, entries []snapshot.AcceptanceEvidenceEntry) {
	builder.WriteString("acceptance_item_coverage_summary:\n")
	if len(entries) == 0 {
		builder.WriteString("  - id: none\n")
		builder.WriteString("    status: not_checked\n")
		builder.WriteString("    evidence_refs: none\n")
		return
	}
	for _, entry := range entries {
		builder.WriteString("  - id: " + entry.ID + "\n")
		builder.WriteString("    status: " + entry.Status + "\n")
		evidenceRefs := strings.TrimSpace(entry.EvidenceRefs)
		if evidenceRefs == "" {
			evidenceRefs = "none"
		}
		builder.WriteString("    evidence_refs: " + evidenceRefs + "\n")
	}
}

func appendPromotionEvidenceRefs(builder *strings.Builder, refs []string) {
	builder.WriteString("key_evidence_source_refs:\n")
	for _, ref := range refs {
		builder.WriteString("  - " + ref + "\n")
	}
}

func promotionEvidenceRefs(verifyData snapshot.ProcessSnapshotData) []string {
	refs := []string{}
	for _, key := range []string{"evidence_refs", "review_input_refs", "human_decision_refs"} {
		value := strings.TrimSpace(verifyData.Scalars[key])
		if value == "" || value == "none" {
			continue
		}
		refs = appendUniqueString(refs, value)
	}
	for _, entry := range verifyData.AcceptanceEvidence {
		for _, ref := range strings.Split(entry.EvidenceRefs, ";") {
			ref = strings.TrimSpace(ref)
			if ref == "" || ref == "none" {
				continue
			}
			refs = appendUniqueString(refs, ref)
		}
	}
	for _, entry := range verifyData.RetirementEvidence {
		for _, ref := range strings.Split(entry.EvidenceRefs, ";") {
			ref = strings.TrimSpace(ref)
			if ref == "" || ref == "none" {
				continue
			}
			refs = appendUniqueString(refs, ref)
		}
	}
	if len(refs) == 0 {
		refs = append(refs, verifyData.ProcessFile)
	}
	return refs
}

func appendUniqueString(items []string, item string) []string {
	for _, existing := range items {
		if existing == item {
			return items
		}
	}
	return append(items, item)
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
	if opts.ObjectType != "unit" {
		return fmt.Errorf("object-type must be unit")
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
	return nil
}

func requiresExplicitTruthFallbackReason(opts Options) bool {
	switch opts.Command {
	case "unit_verify":
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
			"pass":         "unit_verify",
			"blocked":      "unit_check",
			"fix_required": "unit_check",
			"checkpoint":   "unit_check",
		})
		if opts.Outcome == "pass" {
			trans.ValidationProcess = "check"
		}
	case "unit_verify":
		trans = unitVerifyTransition(opts, current)
	case "unit_promote":
		trans = unitPromoteTransition(opts, current)
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
	if trans.FailureLayer != "" {
		if err := processcleanup.ValidateFallbackReason(trans.Reason, trans.FailureLayer); err != nil {
			return transition{}, err
		}
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
		return withNextAndValidationDecision(current, "unit_fork", "stable_verify", opts.Outcome)
	case "small_repair_required", "evidence_incomplete", "truth_rejudge_required":
		return withNext(current, "unit_fork")
	case "controlled_repair_required":
		if opts.CandidateIntent != "repair" {
			return transition{}
		}
		return withNextAndValidationDecision(current, "unit_fork", "stable_verify", opts.Outcome)
	case "controlled_change_required":
		if opts.CandidateIntent != "change" {
			return transition{}
		}
		return withNextAndValidationDecision(current, "unit_fork", "stable_verify", opts.Outcome)
	default:
		return transition{}
	}
}



func unitVerifyTransition(opts Options, current statusfile.ObjectStatus) transition {
	switch opts.Outcome {
	case "ready_to_promote":
		return withNextAndValidation(current, "unit_promote", "verify")
	case "truth_fallback":
		return fallback(current, "unit_check", "truth_layer", opts.Reason)
	case "spec_issue":
		return withNext(current, "unit_check")
	case "evidence_incomplete", "human_verify", "impl_issue":
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

func withNextAndValidationDecision(current statusfile.ObjectStatus, next, process, decision string) transition {
	trans := withNextAndValidation(current, next, process)
	trans.ValidationDecision = decision
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

func validateProcess(repoRoot, objectType, object, process, expectedDecision string) ([]string, error) {
	result, err := snapshot.ValidateProcessFileForObject(repoRoot, objectType, object, process)
	if err != nil {
		return nil, err
	}
	if !result.Valid {
		return result.Mismatches, fmt.Errorf("required %s process is invalid for %s %s: %s", process, objectType, object, strings.Join(result.Mismatches, "; "))
	}
	if expectedDecision != "" {
		processData, err := snapshot.LoadProcessSnapshot(repoRoot, objectType, object, process)
		if err != nil {
			return nil, err
		}
		actualDecision := processData.Scalars["decision"]
		if actualDecision != expectedDecision {
			mismatch := fmt.Sprintf("decision mismatch: actual=%s expected=%s", actualDecision, expectedDecision)
			return []string{mismatch}, fmt.Errorf("required %s process decision is invalid for %s %s: %s", process, objectType, object, mismatch)
		}
	}
	return nil, nil
}

func validateControlledStableVerifyForkIntent(opts Options) error {
	if opts.Command != "unit_fork" || opts.Outcome != "candidate_created" {
		return nil
	}
	requirement, err := snapshot.StableVerifyCandidateIntentRequirement(opts.RepoRoot, opts.ObjectType, opts.Object)
	if err != nil {
		return err
	}
	if !requirement.Required {
		return nil
	}
	actualIntent, err := snapshot.CandidateIntentForObject(opts.RepoRoot, opts.ObjectType, opts.Object)
	if err != nil {
		return err
	}
	if actualIntent != requirement.RequiredIntent {
		return fmt.Errorf("candidate_intent mismatch for controlled stable verify decision %s: actual=%s expected=%s", requirement.Decision, emptyAsNone(actualIntent), requirement.RequiredIntent)
	}
	return nil
}

func validateUnitForkAppendixCoverage(opts Options) error {
	if opts.Command != "unit_fork" || opts.Outcome != "candidate_created" {
		return nil
	}
	return unitappendix.ValidateCandidateCoverage(opts.RepoRoot, opts.ObjectType, opts.Object)
}

func emptyAsNone(value string) string {
	if value == "" {
		return "none"
	}
	return value
}

func shouldValidateInput(command string, trans transition) bool {
	if trans.CleanupKind == cleanupFallback {
		return false
	}
	switch command {
	case "unit_verify", "unit_promote":
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
	case "unit_init", "unit_new":
		return true
	default:
		return false
	}
}
