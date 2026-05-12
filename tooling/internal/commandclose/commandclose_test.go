package commandclose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

func TestDetermineTransitionCoversStandardOutcomes(t *testing.T) {
	cases := []struct {
		name             string
		opts             Options
		current          statusfile.ObjectStatus
		present          bool
		wantStable       string
		wantCandidate    string
		wantActiveLayer  string
		wantNext         string
		wantValidation   string
		wantCleanupKind  string
		wantCleanupMode  string
		wantFailureLayer string
		wantReason       string
	}{
		{name: "unit_init stable_created", opts: Options{Command: "unit_init", ObjectType: "unit", Object: "demo", Outcome: "stable_created"}, present: false, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "unit_fork", wantCleanupKind: cleanupNone},
		{name: "unit_new candidate_created", opts: Options{Command: "unit_new", ObjectType: "unit", Object: "demo", Outcome: "candidate_created"}, present: false, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupNone},
		{name: "unit_stable_verify aligned", opts: Options{Command: "unit_stable_verify", ObjectType: "unit", Object: "demo", Outcome: "aligned"}, current: unitStableStatus("unit_stable_verify"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "unit_fork", wantCleanupKind: cleanupNone},
		{name: "unit_stable_verify small_repair_required", opts: Options{Command: "unit_stable_verify", ObjectType: "unit", Object: "demo", Outcome: "small_repair_required"}, current: unitStableStatus("unit_stable_verify"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "unit_stable_verify", wantCleanupKind: cleanupNone},
		{name: "unit_stable_verify evidence_incomplete", opts: Options{Command: "unit_stable_verify", ObjectType: "unit", Object: "demo", Outcome: "evidence_incomplete"}, current: unitStableStatus("unit_stable_verify"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "unit_stable_verify", wantCleanupKind: cleanupNone},
		{name: "unit_stable_verify truth_rejudge_required", opts: Options{Command: "unit_stable_verify", ObjectType: "unit", Object: "demo", Outcome: "truth_rejudge_required"}, current: unitStableStatus("unit_stable_verify"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "unit_stable_verify", wantCleanupKind: cleanupNone},
		{name: "unit_stable_verify controlled_repair_required", opts: Options{Command: "unit_stable_verify", ObjectType: "unit", Object: "demo", Outcome: "controlled_repair_required", CandidateIntent: "repair"}, current: unitStableStatus("unit_stable_verify"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "unit_fork", wantCleanupKind: cleanupNone},
		{name: "unit_stable_verify controlled_change_required", opts: Options{Command: "unit_stable_verify", ObjectType: "unit", Object: "demo", Outcome: "controlled_change_required", CandidateIntent: "change"}, current: unitStableStatus("unit_stable_verify"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "unit_fork", wantCleanupKind: cleanupNone},
		{name: "unit_fork candidate_created", opts: Options{Command: "unit_fork", ObjectType: "unit", Object: "demo", Outcome: "candidate_created"}, current: unitStableStatus("unit_fork"), present: true, wantStable: "yes", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupSuccess, wantCleanupMode: "unit_fork"},
		{name: "unit_check pass", opts: Options{Command: "unit_check", ObjectType: "unit", Object: "demo", Outcome: "pass"}, current: unitCandidateStatus("unit_check"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_plan", wantValidation: "check", wantCleanupKind: cleanupNone},
		{name: "unit_check blocked", opts: Options{Command: "unit_check", ObjectType: "unit", Object: "demo", Outcome: "blocked"}, current: unitCandidateStatus("unit_check"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupNone},
		{name: "unit_check fix_required", opts: Options{Command: "unit_check", ObjectType: "unit", Object: "demo", Outcome: "fix_required"}, current: unitCandidateStatus("unit_check"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupNone},
		{name: "unit_check checkpoint", opts: Options{Command: "unit_check", ObjectType: "unit", Object: "demo", Outcome: "checkpoint"}, current: unitCandidateStatus("unit_check"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupNone},
		{name: "unit_plan plan_ready", opts: Options{Command: "unit_plan", ObjectType: "unit", Object: "demo", Outcome: "plan_ready"}, current: unitCandidateStatus("unit_plan"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_impl", wantValidation: "plan", wantCleanupKind: cleanupNone},
		{name: "unit_plan truth_fallback", opts: Options{Command: "unit_plan", ObjectType: "unit", Object: "demo", Outcome: "truth_fallback", Reason: "truth_incomplete"}, current: unitCandidateStatus("unit_plan"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupFallback, wantFailureLayer: "truth_layer", wantReason: "truth_incomplete"},
		{name: "unit_plan blocked", opts: Options{Command: "unit_plan", ObjectType: "unit", Object: "demo", Outcome: "blocked"}, current: unitCandidateStatus("unit_plan"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_plan", wantCleanupKind: cleanupNone},
		{name: "unit_plan decision_checkpoint", opts: Options{Command: "unit_plan", ObjectType: "unit", Object: "demo", Outcome: "decision_checkpoint"}, current: unitCandidateStatus("unit_plan"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_plan", wantCleanupKind: cleanupNone},
		{name: "unit_impl ready_for_verify", opts: Options{Command: "unit_impl", ObjectType: "unit", Object: "demo", Outcome: "ready_for_verify"}, current: unitCandidateStatus("unit_impl"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_verify", wantCleanupKind: cleanupNone},
		{name: "unit_impl blocked", opts: Options{Command: "unit_impl", ObjectType: "unit", Object: "demo", Outcome: "blocked"}, current: unitCandidateStatus("unit_impl"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_impl", wantCleanupKind: cleanupNone},
		{name: "unit_impl truth_fallback", opts: Options{Command: "unit_impl", ObjectType: "unit", Object: "demo", Outcome: "truth_fallback", Reason: "truth_drift"}, current: unitCandidateStatus("unit_impl"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupFallback, wantFailureLayer: "truth_layer", wantReason: "truth_drift"},
		{name: "unit_impl plan_fallback", opts: Options{Command: "unit_impl", ObjectType: "unit", Object: "demo", Outcome: "plan_fallback"}, current: unitCandidateStatus("unit_impl"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_plan", wantCleanupKind: cleanupFallback, wantFailureLayer: "plan_layer", wantReason: "gate_missing"},
		{name: "unit_impl gate_fallback", opts: Options{Command: "unit_impl", ObjectType: "unit", Object: "demo", Outcome: "gate_fallback"}, current: unitCandidateStatus("unit_impl"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupFallback, wantFailureLayer: "gate_layer", wantReason: "gate_missing"},
		{name: "unit_verify ready_to_promote", opts: Options{Command: "unit_verify", ObjectType: "unit", Object: "demo", Outcome: "ready_to_promote"}, current: unitCandidateStatus("unit_verify"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_promote", wantValidation: "verify", wantCleanupKind: cleanupNone},
		{name: "unit_verify implementation_deviation", opts: Options{Command: "unit_verify", ObjectType: "unit", Object: "demo", Outcome: "implementation_deviation"}, current: unitCandidateStatus("unit_verify"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_impl", wantCleanupKind: cleanupFallback, wantFailureLayer: "implementation_layer", wantReason: "implementation_deviation"},
		{name: "unit_verify truth_fallback", opts: Options{Command: "unit_verify", ObjectType: "unit", Object: "demo", Outcome: "truth_fallback", Reason: "truth_drift"}, current: unitCandidateStatus("unit_verify"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupFallback, wantFailureLayer: "truth_layer", wantReason: "truth_drift"},
		{name: "unit_verify evidence_incomplete", opts: Options{Command: "unit_verify", ObjectType: "unit", Object: "demo", Outcome: "evidence_incomplete"}, current: unitCandidateStatus("unit_verify"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_verify", wantCleanupKind: cleanupNone},
		{name: "unit_verify human_verify", opts: Options{Command: "unit_verify", ObjectType: "unit", Object: "demo", Outcome: "human_verify"}, current: unitCandidateStatus("unit_verify"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_verify", wantCleanupKind: cleanupNone},
		{name: "unit_promote promoted", opts: Options{Command: "unit_promote", ObjectType: "unit", Object: "demo", Outcome: "promoted"}, current: unitCandidateStatus("unit_promote"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "unit_fork", wantValidation: "verify", wantCleanupKind: cleanupSuccess, wantCleanupMode: "unit_promote"},
		{name: "unit_promote verify_invalid_truth", opts: Options{Command: "unit_promote", ObjectType: "unit", Object: "demo", Outcome: "verify_invalid_truth"}, current: unitCandidateStatus("unit_promote"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupFallback, wantFailureLayer: "truth_layer", wantReason: "truth_drift"},
		{name: "unit_promote verify_invalid_binding", opts: Options{Command: "unit_promote", ObjectType: "unit", Object: "demo", Outcome: "verify_invalid_binding"}, current: unitCandidateStatus("unit_promote"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupFallback, wantFailureLayer: "truth_layer", wantReason: "binding_drift"},
		{name: "unit_promote verify_invalid_baseline", opts: Options{Command: "unit_promote", ObjectType: "unit", Object: "demo", Outcome: "verify_invalid_baseline"}, current: unitCandidateStatus("unit_promote"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupFallback, wantFailureLayer: "truth_layer", wantReason: "baseline_drift"},
		{name: "unit_promote verify_invalid_rule", opts: Options{Command: "unit_promote", ObjectType: "unit", Object: "demo", Outcome: "verify_invalid_rule"}, current: unitCandidateStatus("unit_promote"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupFallback, wantFailureLayer: "truth_layer", wantReason: "rule_drift"},
		{name: "unit_promote verify_invalid_plan", opts: Options{Command: "unit_promote", ObjectType: "unit", Object: "demo", Outcome: "verify_invalid_plan"}, current: unitCandidateStatus("unit_promote"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_plan", wantCleanupKind: cleanupFallback, wantFailureLayer: "plan_layer", wantReason: "gate_missing"},
		{name: "unit_promote verify_invalid_implementation", opts: Options{Command: "unit_promote", ObjectType: "unit", Object: "demo", Outcome: "verify_invalid_implementation"}, current: unitCandidateStatus("unit_promote"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_impl", wantCleanupKind: cleanupFallback, wantFailureLayer: "implementation_layer", wantReason: "implementation_deviation"},
		{name: "unit_promote verify_invalid_gate", opts: Options{Command: "unit_promote", ObjectType: "unit", Object: "demo", Outcome: "verify_invalid_gate"}, current: unitCandidateStatus("unit_promote"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupFallback, wantFailureLayer: "gate_layer", wantReason: "gate_missing"},
		{name: "unit_promote verify_invalid_evidence", opts: Options{Command: "unit_promote", ObjectType: "unit", Object: "demo", Outcome: "verify_invalid_evidence"}, current: unitCandidateStatus("unit_promote"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_verify", wantCleanupKind: cleanupFallback, wantFailureLayer: "evidence_layer", wantReason: "evidence_incomplete"},
		{name: "unit_promote promotion_recovered", opts: Options{Command: "unit_promote", ObjectType: "unit", Object: "demo", Outcome: "promotion_recovered", StableBefore: "yes"}, current: unitCandidateStatus("unit_promote"), present: true, wantStable: "yes", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupFallback, wantFailureLayer: "truth_layer", wantReason: "truth_drift"},
		{name: "scenario_new candidate_created", opts: Options{Command: "scenario_new", ObjectType: "scenario", Object: "checkout", Outcome: "candidate_created"}, present: false, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "scenario_check", wantCleanupKind: cleanupNone},
		{name: "scenario_stable_verify aligned", opts: Options{Command: "scenario_stable_verify", ObjectType: "scenario", Object: "checkout", Outcome: "aligned"}, current: scenarioStableStatus("scenario_stable_verify"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "scenario_fork", wantCleanupKind: cleanupNone},
		{name: "scenario_stable_verify not_aligned", opts: Options{Command: "scenario_stable_verify", ObjectType: "scenario", Object: "checkout", Outcome: "not_aligned"}, current: scenarioStableStatus("scenario_stable_verify"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "scenario_stable_verify", wantCleanupKind: cleanupNone},
		{name: "scenario_stable_verify evidence_incomplete", opts: Options{Command: "scenario_stable_verify", ObjectType: "scenario", Object: "checkout", Outcome: "evidence_incomplete"}, current: scenarioStableStatus("scenario_stable_verify"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "scenario_stable_verify", wantCleanupKind: cleanupNone},
		{name: "scenario_fork candidate_created", opts: Options{Command: "scenario_fork", ObjectType: "scenario", Object: "checkout", Outcome: "candidate_created"}, current: scenarioStableStatus("scenario_fork"), present: true, wantStable: "yes", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "scenario_check", wantCleanupKind: cleanupSuccess, wantCleanupMode: "scenario_fork"},
		{name: "scenario_check pass", opts: Options{Command: "scenario_check", ObjectType: "scenario", Object: "checkout", Outcome: "pass"}, current: scenarioCandidateStatus("scenario_check"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "scenario_verify", wantValidation: "check", wantCleanupKind: cleanupNone},
		{name: "scenario_check blocked", opts: Options{Command: "scenario_check", ObjectType: "scenario", Object: "checkout", Outcome: "blocked"}, current: scenarioCandidateStatus("scenario_check"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "scenario_check", wantCleanupKind: cleanupNone},
		{name: "scenario_check fix_required", opts: Options{Command: "scenario_check", ObjectType: "scenario", Object: "checkout", Outcome: "fix_required"}, current: scenarioCandidateStatus("scenario_check"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "scenario_check", wantCleanupKind: cleanupNone},
		{name: "scenario_verify pass", opts: Options{Command: "scenario_verify", ObjectType: "scenario", Object: "checkout", Outcome: "pass"}, current: scenarioCandidateStatus("scenario_verify"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "scenario_promote", wantValidation: "verify", wantCleanupKind: cleanupNone},
		{name: "scenario_verify truth_fallback", opts: Options{Command: "scenario_verify", ObjectType: "scenario", Object: "checkout", Outcome: "truth_fallback", Reason: "truth_drift"}, current: scenarioCandidateStatus("scenario_verify"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "scenario_check", wantCleanupKind: cleanupFallback, wantFailureLayer: "truth_layer", wantReason: "truth_drift"},
		{name: "scenario_verify gate_fallback", opts: Options{Command: "scenario_verify", ObjectType: "scenario", Object: "checkout", Outcome: "gate_fallback"}, current: scenarioCandidateStatus("scenario_verify"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "scenario_check", wantCleanupKind: cleanupFallback, wantFailureLayer: "gate_layer", wantReason: "gate_missing"},
		{name: "scenario_verify evidence_incomplete", opts: Options{Command: "scenario_verify", ObjectType: "scenario", Object: "checkout", Outcome: "evidence_incomplete"}, current: scenarioCandidateStatus("scenario_verify"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "scenario_verify", wantCleanupKind: cleanupFallback, wantFailureLayer: "evidence_layer", wantReason: "evidence_incomplete"},
		{name: "scenario_verify blocked_by_affected_units", opts: Options{Command: "scenario_verify", ObjectType: "scenario", Object: "checkout", Outcome: "blocked_by_affected_units"}, current: scenarioCandidateStatus("scenario_verify"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "scenario_verify", wantCleanupKind: cleanupNone},
		{name: "scenario_promote promoted", opts: Options{Command: "scenario_promote", ObjectType: "scenario", Object: "checkout", Outcome: "promoted"}, current: scenarioCandidateStatus("scenario_promote"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "scenario_fork", wantValidation: "verify", wantCleanupKind: cleanupSuccess, wantCleanupMode: "scenario_promote"},
		{name: "scenario_promote dependency_not_ready", opts: Options{Command: "scenario_promote", ObjectType: "scenario", Object: "checkout", Outcome: "dependency_not_ready"}, current: scenarioCandidateStatus("scenario_promote"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "scenario_promote", wantCleanupKind: cleanupNone},
		{name: "scenario_promote promotion_recovered", opts: Options{Command: "scenario_promote", ObjectType: "scenario", Object: "checkout", Outcome: "promotion_recovered", StableBefore: "no"}, current: scenarioCandidateStatus("scenario_promote"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "scenario_check", wantCleanupKind: cleanupFallback, wantFailureLayer: "truth_layer", wantReason: "truth_drift"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			trans, err := determineTransition(tc.opts, tc.current, tc.present)
			if err != nil {
				t.Fatalf("determineTransition: %v", err)
			}
			if trans.Status.Stable != tc.wantStable || trans.Status.Candidate != tc.wantCandidate || trans.Status.ActiveLayer != tc.wantActiveLayer || trans.Status.NextCommand != tc.wantNext {
				t.Fatalf("status mismatch: got stable=%s candidate=%s active=%s next=%s", trans.Status.Stable, trans.Status.Candidate, trans.Status.ActiveLayer, trans.Status.NextCommand)
			}
			if trans.ValidationProcess != tc.wantValidation {
				t.Fatalf("validation mismatch: got %q want %q", trans.ValidationProcess, tc.wantValidation)
			}
			if trans.CleanupKind != tc.wantCleanupKind || trans.CleanupMode != tc.wantCleanupMode || trans.FailureLayer != tc.wantFailureLayer || trans.Reason != tc.wantReason {
				t.Fatalf("cleanup mismatch: got kind=%s mode=%s layer=%s reason=%s", trans.CleanupKind, trans.CleanupMode, trans.FailureLayer, trans.Reason)
			}
		})
	}
}

func TestCloseRejectsInvalidCurrentNextCommand(t *testing.T) {
	repoRoot := commandCloseTestRepo(t, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | test |\n")
	_, err := Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_plan",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "blocked",
	})
	if err == nil || !strings.Contains(err.Error(), "status next command mismatch") {
		t.Fatalf("expected next-command mismatch, got %v", err)
	}
}

func TestCloseRejectsUnsupportedOutcome(t *testing.T) {
	repoRoot := commandCloseTestRepo(t, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | test |\n")
	_, err := Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_check",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "unknown",
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported outcome") {
		t.Fatalf("expected unsupported outcome, got %v", err)
	}
}

func TestCloseRejectsMissingRequiredFlag(t *testing.T) {
	_, err := Close(Options{
		RepoRoot:   t.TempDir(),
		Command:    "unit_check",
		ObjectType: "unit",
		Object:     "demo",
	})
	if err == nil || !strings.Contains(err.Error(), "command, object-type, object, and outcome are required") {
		t.Fatalf("expected required flag error, got %v", err)
	}
}

func TestCloseApplyInvokesFallbackCleanup(t *testing.T) {
	repoRoot := commandCloseTestRepo(t, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_impl` | test |\n")
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), "check")
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/demo.md"), "plan")
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit/demo.md"), "verify")

	result, err := Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_impl",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "truth_fallback",
		Reason:     "truth_drift",
		Apply:      true,
	})
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
	if result.CleanupAction != "fallback:truth_layer:truth_drift" || len(result.FallbackCleanup.DeletedFiles) != 3 {
		t.Fatalf("expected fallback cleanup deletion, got action=%s deleted=%v", result.CleanupAction, result.FallbackCleanup.DeletedFiles)
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.NextCommand != "unit_check" {
		t.Fatalf("expected unit_check after fallback, got %s", status.NextCommand)
	}
}

func TestCloseApplyInvokesSuccessCleanup(t *testing.T) {
	repoRoot := commandCloseTestRepo(t, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | test |\n")
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), "check")
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/demo.md"), "plan")
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit/demo.md"), "verify")

	result, err := Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_fork",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "candidate_created",
		Apply:      true,
	})
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
	if result.CleanupAction != "success:unit_fork" || len(result.SuccessCleanup.DeletedFiles) != 3 {
		t.Fatalf("expected success cleanup deletion, got action=%s deleted=%v", result.CleanupAction, result.SuccessCleanup.DeletedFiles)
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.Candidate != "yes" || status.ActiveLayer != "candidate" || status.NextCommand != "unit_check" {
		t.Fatalf("unexpected status after success cleanup: %+v", status)
	}
}

func TestValidateOutcomeFlagsForControlledCandidateIntent(t *testing.T) {
	err := validateOutcomeFlags(Options{Command: "unit_stable_verify", Outcome: "controlled_repair_required"})
	if err == nil || !strings.Contains(err.Error(), "requires --candidate-intent repair") {
		t.Fatalf("expected missing repair intent error, got %v", err)
	}

	err = validateOutcomeFlags(Options{Command: "unit_stable_verify", Outcome: "controlled_repair_required", CandidateIntent: "change"})
	if err == nil || !strings.Contains(err.Error(), "requires --candidate-intent repair") {
		t.Fatalf("expected wrong repair intent error, got %v", err)
	}
}

func TestValidateOutcomeFlagsRequiresReasonForGenericTruthFallback(t *testing.T) {
	for _, command := range []string{"unit_plan", "unit_impl", "unit_verify", "scenario_verify"} {
		t.Run(command, func(t *testing.T) {
			err := validateOutcomeFlags(Options{Command: command, Outcome: "truth_fallback"})
			if err == nil || !strings.Contains(err.Error(), "requires --reason") {
				t.Fatalf("expected reason requirement for %s truth_fallback, got %v", command, err)
			}
		})
	}
}

func TestValidateOutcomeFlagsRequiresUnitPlanTruthIncompleteReason(t *testing.T) {
	err := validateOutcomeFlags(Options{Command: "unit_plan", Outcome: "truth_fallback", Reason: "truth_drift"})
	if err == nil || !strings.Contains(err.Error(), "requires --reason truth_incomplete") {
		t.Fatalf("expected unit_plan truth_fallback reason guard, got %v", err)
	}
}

func unitStableStatus(next string) statusfile.ObjectStatus {
	return statusfile.ObjectStatus{ObjectType: "unit", Object: "demo", Stable: "yes", Candidate: "no", ActiveLayer: "stable", NextCommand: next, Notes: "test"}
}

func unitCandidateStatus(next string) statusfile.ObjectStatus {
	return statusfile.ObjectStatus{ObjectType: "unit", Object: "demo", Stable: "no", Candidate: "yes", ActiveLayer: "candidate", NextCommand: next, Notes: "test"}
}

func scenarioStableStatus(next string) statusfile.ObjectStatus {
	return statusfile.ObjectStatus{ObjectType: "scenario", Object: "checkout", Stable: "yes", Candidate: "no", ActiveLayer: "stable", NextCommand: next, Notes: "test"}
}

func scenarioCandidateStatus(next string) statusfile.ObjectStatus {
	return statusfile.ObjectStatus{ObjectType: "scenario", Object: "checkout", Stable: "no", Candidate: "yes", ActiveLayer: "candidate", NextCommand: next, Notes: "test"}
}

func commandCloseTestRepo(t *testing.T, rows string) string {
	t.Helper()
	repoRoot := t.TempDir()
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), ""+
		"# Spec Status\n\n"+
		"## Formal Objects\n\n"+
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n"+
		"|---|---|---|---|---|---|---|\n"+
		rows)
	return repoRoot
}

func writeCommandCloseTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
