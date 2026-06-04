package commandclose

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/processcleanup"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/testfixtures"
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
		{name: "unit_stable_verify aligned", opts: Options{Command: "unit_stable_verify", ObjectType: "unit", Object: "demo", Outcome: "aligned"}, current: unitStableStatus("unit_stable_verify"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "unit_fork", wantValidation: "stable_verify", wantCleanupKind: cleanupNone},
		{name: "unit_stable_verify small_repair_required", opts: Options{Command: "unit_stable_verify", ObjectType: "unit", Object: "demo", Outcome: "small_repair_required"}, current: unitStableStatus("unit_stable_verify"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "unit_stable_verify", wantCleanupKind: cleanupNone},
		{name: "unit_stable_verify evidence_incomplete", opts: Options{Command: "unit_stable_verify", ObjectType: "unit", Object: "demo", Outcome: "evidence_incomplete"}, current: unitStableStatus("unit_stable_verify"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "unit_stable_verify", wantCleanupKind: cleanupNone},
		{name: "unit_stable_verify truth_rejudge_required", opts: Options{Command: "unit_stable_verify", ObjectType: "unit", Object: "demo", Outcome: "truth_rejudge_required"}, current: unitStableStatus("unit_stable_verify"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "unit_stable_verify", wantCleanupKind: cleanupNone},
		{name: "unit_stable_verify controlled_repair_required", opts: Options{Command: "unit_stable_verify", ObjectType: "unit", Object: "demo", Outcome: "controlled_repair_required", CandidateIntent: "repair"}, current: unitStableStatus("unit_stable_verify"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "unit_fork", wantValidation: "stable_verify", wantCleanupKind: cleanupNone},
		{name: "unit_stable_verify controlled_change_required", opts: Options{Command: "unit_stable_verify", ObjectType: "unit", Object: "demo", Outcome: "controlled_change_required", CandidateIntent: "change"}, current: unitStableStatus("unit_stable_verify"), present: true, wantStable: "yes", wantCandidate: "no", wantActiveLayer: "stable", wantNext: "unit_fork", wantValidation: "stable_verify", wantCleanupKind: cleanupNone},
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
		{name: "unit_impl plan_fallback", opts: Options{Command: "unit_impl", ObjectType: "unit", Object: "demo", Outcome: "plan_fallback"}, current: unitCandidateStatus("unit_impl"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_plan", wantCleanupKind: cleanupFallback, wantFailureLayer: "plan_layer", wantReason: "plan_drift"},
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
		{name: "unit_promote verify_invalid_plan", opts: Options{Command: "unit_promote", ObjectType: "unit", Object: "demo", Outcome: "verify_invalid_plan"}, current: unitCandidateStatus("unit_promote"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_plan", wantCleanupKind: cleanupFallback, wantFailureLayer: "plan_layer", wantReason: "plan_drift"},
		{name: "unit_promote verify_invalid_implementation", opts: Options{Command: "unit_promote", ObjectType: "unit", Object: "demo", Outcome: "verify_invalid_implementation"}, current: unitCandidateStatus("unit_promote"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_impl", wantCleanupKind: cleanupFallback, wantFailureLayer: "implementation_layer", wantReason: "implementation_deviation"},
		{name: "unit_promote verify_invalid_gate", opts: Options{Command: "unit_promote", ObjectType: "unit", Object: "demo", Outcome: "verify_invalid_gate"}, current: unitCandidateStatus("unit_promote"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupFallback, wantFailureLayer: "gate_layer", wantReason: "gate_missing"},
		{name: "unit_promote verify_invalid_evidence", opts: Options{Command: "unit_promote", ObjectType: "unit", Object: "demo", Outcome: "verify_invalid_evidence"}, current: unitCandidateStatus("unit_promote"), present: true, wantStable: "no", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_verify", wantCleanupKind: cleanupFallback, wantFailureLayer: "evidence_layer", wantReason: "evidence_incomplete"},
		{name: "unit_promote promotion_recovered", opts: Options{Command: "unit_promote", ObjectType: "unit", Object: "demo", Outcome: "promotion_recovered", StableBefore: "yes"}, current: unitCandidateStatus("unit_promote"), present: true, wantStable: "yes", wantCandidate: "yes", wantActiveLayer: "candidate", wantNext: "unit_check", wantCleanupKind: cleanupFallback, wantFailureLayer: "truth_layer", wantReason: "truth_drift"},
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

func TestCloseUnitForkBlocksMissingCandidateAppendixCoverage(t *testing.T) {
	repoRoot := commandCloseTestRepo(t, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | test |\n")
	writeCommandCloseCandidateSpecWithIntent(t, repoRoot, "change")
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/appendix/s_unit_demo_prompt.md"), `---
unit: demo
layer: stable
---

# Stable Appendix
`)

	_, err := Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_fork",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "candidate_created",
		Apply:      true,
	})
	if err == nil || !strings.Contains(err.Error(), "missing candidate appendix") {
		t.Fatalf("expected missing candidate appendix error, got %v", err)
	}
	status, lookupErr := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if lookupErr != nil {
		t.Fatalf("LookupObjectStatus: %v", lookupErr)
	}
	if status.Candidate != "no" || status.ActiveLayer != "stable" || status.NextCommand != "unit_fork" {
		t.Fatalf("failed fork close must not update status, got %+v", status)
	}

	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/appendix/c_unit_demo_prompt.md"), `---
unit: demo
layer: candidate
---

# Candidate Appendix
`)
	result, err := Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_fork",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "candidate_created",
	})
	if err != nil {
		t.Fatalf("Close dry-run after candidate appendix exists: %v", err)
	}
	if result.StatusAfter.Candidate != "yes" || result.CleanupAction != "success:unit_fork" {
		t.Fatalf("unexpected dry-run result: %+v", result)
	}
}

func TestClosePromoteWritesStablePromotionSummaryBeforeCleanup(t *testing.T) {
	repoRoot := commandCloseSnapshotRepo(t, "| `unit` | `demo` | `yes` | `yes` | `candidate` | `unit_promote` | test |\n")
	writeCommandCloseStableSpec(t, repoRoot)
	candidateSnapshot, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCommandCloseUnitPlanProcessWithExtra(t, repoRoot, candidateSnapshot, "")
	writeCommandCloseUnitVerifyProcess(t, repoRoot, candidateSnapshot)

	result, err := Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_promote",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "promoted",
		Apply:      true,
	})
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
	if len(result.InputValidatedProcesses) != 2 ||
		result.InputValidatedProcesses[0].ProcessKind != "plan" ||
		result.InputValidatedProcesses[1].ProcessKind != "verify" {
		t.Fatalf("unit_promote must validate plan and verify, got %+v", result.InputValidatedProcesses)
	}
	summaryRef := "docs/specs/_verify_result/stable/unit/demo.md"
	if result.PromotionSummaryFile != summaryRef {
		t.Fatalf("expected promotion summary %s, got %q", summaryRef, result.PromotionSummaryFile)
	}
	summary, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(summaryRef)))
	if err != nil {
		t.Fatalf("read promotion summary: %v", err)
	}
	for _, want := range []string{
		"stable_truth_file_ref: docs/specs/units/stable/s_unit_demo.md",
		"stable_truth_version_ref: s_unit_demo@1.0.0",
		"promotion_verify_result_ref: docs/specs/_verify_result/unit/demo.md",
		"acceptance_item_coverage_summary:",
		"    status: pass",
		"    evidence_refs: go test ./...",
		"key_evidence_source_refs:",
		"  - go test ./...",
	} {
		if !strings.Contains(string(summary), want) {
			t.Fatalf("promotion summary missing %q:\n%s", want, string(summary))
		}
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "docs/specs/_verify_result/unit/demo.md")); !os.IsNotExist(err) {
		t.Fatalf("candidate verify result should be deleted after summary, stat err=%v", err)
	}
}

func TestClosePromoteCleanupFailureReportsRetryAfterStatusUpdate(t *testing.T) {
	repoRoot := commandCloseSnapshotRepo(t, "| `unit` | `demo` | `yes` | `yes` | `candidate` | `unit_promote` | test |\n")
	writeCommandCloseStableSpec(t, repoRoot)
	candidateSnapshot, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCommandCloseUnitPlanProcessWithExtra(t, repoRoot, candidateSnapshot, "")
	writeCommandCloseUnitVerifyProcess(t, repoRoot, candidateSnapshot)

	blockingDir := filepath.Join(repoRoot, "docs/specs/units/candidate/appendix/c_unit_demo_blocked.md")
	if err := os.MkdirAll(blockingDir, 0o755); err != nil {
		t.Fatalf("mkdir blocking appendix dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(blockingDir, "child.txt"), []byte("blocked"), 0o644); err != nil {
		t.Fatalf("write blocking appendix child: %v", err)
	}

	result, err := Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_promote",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "promoted",
		Apply:      true,
	})
	if err == nil || !strings.Contains(err.Error(), "success cleanup failed after status update") || !strings.Contains(err.Error(), "process cleanup-success") {
		t.Fatalf("expected retryable cleanup failure after status update, got %v", err)
	}
	if !result.StatusUpdated {
		t.Fatalf("expected status update before cleanup failure")
	}
	if result.PromotionSummaryFile != "docs/specs/_verify_result/stable/unit/demo.md" {
		t.Fatalf("expected promotion summary to be written before cleanup failure, got %q", result.PromotionSummaryFile)
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.Candidate != "no" || status.ActiveLayer != "stable" || status.NextCommand != "unit_fork" {
		t.Fatalf("status must reflect the completed promotion before retry cleanup: %+v", status)
	}

	if err := os.Remove(filepath.Join(blockingDir, "child.txt")); err != nil {
		t.Fatalf("remove blocking appendix child: %v", err)
	}
	cleanup, err := processcleanup.ApplyObjectSuccessCleanup(repoRoot, "unit", "demo", "unit_promote")
	if err != nil {
		t.Fatalf("retry cleanup-success: %v", err)
	}
	blockingRef := "docs/specs/units/candidate/appendix/c_unit_demo_blocked.md"
	if !containsString(cleanup.DeletedFiles, blockingRef) {
		t.Fatalf("expected retry cleanup to remove %s, got %+v", blockingRef, cleanup)
	}
}

func TestClosePromoteRejectsMissingStableTruthBeforeCleanup(t *testing.T) {
	repoRoot := commandCloseSnapshotRepo(t, "| `unit` | `demo` | `yes` | `yes` | `candidate` | `unit_promote` | test |\n")
	candidateSnapshot, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCommandCloseUnitPlanProcessWithExtra(t, repoRoot, candidateSnapshot, "")
	writeCommandCloseUnitVerifyProcess(t, repoRoot, candidateSnapshot)

	_, err = Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_promote",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "promoted",
		Apply:      true,
	})
	if err == nil || !strings.Contains(err.Error(), "stable promotion summary requires current stable truth") {
		t.Fatalf("expected stable truth blocker, got %v", err)
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.ActiveLayer != "candidate" || status.NextCommand != "unit_promote" {
		t.Fatalf("promotion summary failure must not advance status: %+v", status)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "docs/specs/_verify_result/unit/demo.md")); err != nil {
		t.Fatalf("candidate verify result must remain after failed close, stat err=%v", err)
	}
}

func TestCloseRejectsUnitPlanReadyWhenInputCheckMissing(t *testing.T) {
	repoRoot := commandCloseSnapshotRepo(t, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_plan` | test |\n")
	_, err := Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_plan",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "plan_ready",
		Apply:      true,
	})
	if err == nil || !strings.Contains(err.Error(), "command close input preflight failed") || !strings.Contains(err.Error(), "missing process file") {
		t.Fatalf("expected input preflight failure for missing check, got %v", err)
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.NextCommand != "unit_plan" {
		t.Fatalf("missing input check must not advance status, got %+v", status)
	}
}

func TestCloseRejectsUnitCheckPassWithoutIndependentReceipt(t *testing.T) {
	repoRoot := commandCloseSnapshotRepo(t, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | test |\n")
	snap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), "# check\n\n```yaml\n"+strings.Join([]string{
		"object_type: unit",
		"object_ref: demo",
		"gate: unit_check",
		"decision: pass",
		"allow_next: true",
		"next_command: unit_plan",
		"blocking_summary: none",
		"coverage_summary: demo",
		"truth_layer_ref: " + snap.TruthLayerRef,
		"truth_file_ref: " + snap.SpecFileRef,
		"truth_version_ref: " + snap.SpecVersionRef,
		"truth_fingerprint: " + snap.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + snap.AcceptanceBehaviorFingerprint,
		"unit_appendix_snapshot: none",
		"unit_snapshot: none",
		"rule_snapshot: none",
		"acceptance_item_set:",
		"  - id: demo.core",
		"    verification_surface: internal_flow",
		"    not_runnable_yet: no",
	}, "\n")+"\n```\n")

	_, err = Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_check",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "pass",
		Apply:      true,
	})
	if err == nil || !strings.Contains(err.Error(), "missing required field: evaluation_mode") {
		t.Fatalf("expected missing independent receipt failure, got %v", err)
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.NextCommand != "unit_check" {
		t.Fatalf("missing independent receipt must not advance status, got %+v", status)
	}
}

func TestCloseUnitCheckPassUsesCurrentEvidenceDespiteOldCheckWork(t *testing.T) {
	for _, checkWorkStatus := range []string{"blocked_on_finding", "closed_fix_required"} {
		t.Run(checkWorkStatus, func(t *testing.T) {
			repoRoot := commandCloseSnapshotRepo(t, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | test |\n")
			snap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
			if err != nil {
				t.Fatalf("RebuildCurrentObject: %v", err)
			}
			writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/_check_work/unit/demo.md"), strings.Join([]string{
				"# stale check work",
				"",
				"```yaml",
				"work_flow: unit_check",
				"object_type: unit",
				"object_ref: demo",
				"status: " + checkWorkStatus,
				"```",
				"",
			}, "\n"))
			writeCommandCloseUnitCheckProcess(t, repoRoot, snap)

			result, err := Close(Options{
				RepoRoot:   repoRoot,
				Command:    "unit_check",
				ObjectType: "unit",
				Object:     "demo",
				Outcome:    "pass",
				Apply:      true,
			})
			if err != nil {
				t.Fatalf("Close: %v", err)
			}
			if result.ValidationAction != "validate_process:check" {
				t.Fatalf("expected check validation action, got %s", result.ValidationAction)
			}
			status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
			if err != nil {
				t.Fatalf("LookupObjectStatus: %v", err)
			}
			if status.NextCommand != "unit_plan" {
				t.Fatalf("current valid check evidence must advance to unit_plan, got %+v", status)
			}
		})
	}
}

func TestCloseUnitPlanReadyAcceptsAcceptedTextDriftEvidence(t *testing.T) {
	repoRoot := commandCloseSnapshotRepo(t, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_plan` | test |\n")
	oldSnap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	replaceCommandCloseCandidateSpecText(t, repoRoot, "# Demo\n", "# Demo\n\nEditorial note only.\n")
	currentSnap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject after edit: %v", err)
	}
	writeCommandCloseUnitCheckProcessWithFreshness(t, repoRoot, oldSnap, currentSnap)
	writeCommandCloseUnitPlanProcessWithFreshness(t, repoRoot, oldSnap, currentSnap, "accepted")

	result, err := Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_plan",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "plan_ready",
		Apply:      true,
	})
	if err != nil {
		t.Fatalf("Close: %v result=%+v", err, result)
	}
	if result.StatusAfter.NextCommand != "unit_impl" {
		t.Fatalf("expected next command unit_impl, got %+v", result.StatusAfter)
	}
	if len(result.InputValidatedProcesses) != 1 ||
		result.InputValidatedProcesses[0].FreshnessImpact != snapshot.FreshnessTextDrift ||
		result.InputValidatedProcesses[0].EvidenceReuse != snapshot.EvidenceReuseAccepted {
		t.Fatalf("expected accepted text drift input validation, got %+v", result.InputValidatedProcesses)
	}
}

func TestCloseUnitImplAcceptsAcceptedTextDriftInputs(t *testing.T) {
	repoRoot := commandCloseSnapshotRepo(t, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_impl` | test |\n")
	oldSnap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	replaceCommandCloseCandidateSpecText(t, repoRoot, "# Demo\n", "# Demo\n\nEditorial note only.\n")
	currentSnap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject after edit: %v", err)
	}
	writeCommandCloseUnitCheckProcessWithFreshness(t, repoRoot, oldSnap, currentSnap)
	writeCommandCloseUnitPlanProcessWithFreshness(t, repoRoot, oldSnap, currentSnap, "accepted")

	result, err := Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_impl",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "ready_for_verify",
		Apply:      true,
	})
	if err != nil {
		t.Fatalf("Close: %v result=%+v", err, result)
	}
	if result.StatusAfter.NextCommand != "unit_verify" {
		t.Fatalf("expected next command unit_verify, got %+v", result.StatusAfter)
	}
	if len(result.InputValidatedProcesses) != 2 {
		t.Fatalf("expected check and plan validation, got %+v", result.InputValidatedProcesses)
	}
	for _, process := range result.InputValidatedProcesses {
		if process.FreshnessImpact != snapshot.FreshnessTextDrift || process.EvidenceReuse != snapshot.EvidenceReuseAccepted {
			t.Fatalf("expected accepted text drift process, got %+v", process)
		}
	}
}

func TestCloseRejectsTextDriftWithoutAcceptedFreshnessReceipt(t *testing.T) {
	for _, tc := range []struct {
		name     string
		planMode string
		want     string
	}{
		{
			name:     "missing receipt",
			planMode: "missing",
			want:     "missing required freshness field: freshness_impact",
		},
		{
			name:     "blocked reviewer",
			planMode: "blocked",
			want:     "freshness_reviewer_result mismatch: actual=blocked expected=pass",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repoRoot := commandCloseSnapshotRepo(t, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_plan` | test |\n")
			oldSnap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
			if err != nil {
				t.Fatalf("RebuildCurrentObject: %v", err)
			}
			replaceCommandCloseCandidateSpecText(t, repoRoot, "# Demo\n", "# Demo\n\nEditorial note only.\n")
			currentSnap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
			if err != nil {
				t.Fatalf("RebuildCurrentObject after edit: %v", err)
			}
			writeCommandCloseUnitCheckProcessWithFreshness(t, repoRoot, oldSnap, currentSnap)
			writeCommandCloseUnitPlanProcessWithFreshness(t, repoRoot, oldSnap, currentSnap, tc.planMode)

			_, err = Close(Options{
				RepoRoot:   repoRoot,
				Command:    "unit_plan",
				ObjectType: "unit",
				Object:     "demo",
				Outcome:    "plan_ready",
				Apply:      true,
			})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q failure, got %v", tc.want, err)
			}
			status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
			if err != nil {
				t.Fatalf("LookupObjectStatus: %v", err)
			}
			if status.NextCommand != "unit_plan" {
				t.Fatalf("failed text drift close must not advance status, got %+v", status)
			}
		})
	}
}

func TestCloseRejectsUnitImplReadyWhenInputPlanMissing(t *testing.T) {
	repoRoot := commandCloseSnapshotRepo(t, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_impl` | test |\n")
	snap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCommandCloseUnitCheckProcess(t, repoRoot, snap)

	_, err = Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_impl",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "ready_for_verify",
		Apply:      true,
	})
	if err == nil || !strings.Contains(err.Error(), "plan: missing process file") {
		t.Fatalf("expected input preflight failure for missing plan, got %v", err)
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.NextCommand != "unit_impl" {
		t.Fatalf("missing input plan must not advance status, got %+v", status)
	}
}

func TestCloseRejectsUnitVerifyReadyWhenInputCheckOrPlanMissing(t *testing.T) {
	t.Run("missing check", func(t *testing.T) {
		repoRoot := commandCloseSnapshotRepo(t, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_verify` | test |\n")
		_, err := Close(Options{
			RepoRoot:   repoRoot,
			Command:    "unit_verify",
			ObjectType: "unit",
			Object:     "demo",
			Outcome:    "ready_to_promote",
			Apply:      true,
		})
		if err == nil || !strings.Contains(err.Error(), "check: missing process file") {
			t.Fatalf("expected input preflight failure for missing check, got %v", err)
		}
	})

	t.Run("missing plan", func(t *testing.T) {
		repoRoot := commandCloseSnapshotRepo(t, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_verify` | test |\n")
		snap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
		if err != nil {
			t.Fatalf("RebuildCurrentObject: %v", err)
		}
		writeCommandCloseUnitCheckProcess(t, repoRoot, snap)

		_, err = Close(Options{
			RepoRoot:   repoRoot,
			Command:    "unit_verify",
			ObjectType: "unit",
			Object:     "demo",
			Outcome:    "ready_to_promote",
			Apply:      true,
		})
		if err == nil || !strings.Contains(err.Error(), "plan: missing process file") {
			t.Fatalf("expected input preflight failure for missing plan, got %v", err)
		}
	})
}

func TestCloseRejectsUnitVerifyReadyWithWeakAcceptanceEvidence(t *testing.T) {
	repoRoot := commandCloseSnapshotRepo(t, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_verify` | test |\n")
	snap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCommandCloseUnitCheckProcess(t, repoRoot, snap)
	writeCommandCloseUnitPlanProcessWithExtra(t, repoRoot, snap, "")
	writeCommandCloseUnitVerifyProcess(t, repoRoot, snap)

	verifyPath := filepath.Join(repoRoot, "docs/specs/_verify_result/unit/demo.md")
	content, err := os.ReadFile(verifyPath)
	if err != nil {
		t.Fatalf("read verify: %v", err)
	}
	weak := strings.Replace(string(content), "    evidence_refs: go test ./...\n", "", 1)
	writeCommandCloseTestFile(t, verifyPath, weak)

	_, err = Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_verify",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "ready_to_promote",
		Apply:      true,
	})
	if err == nil || !strings.Contains(err.Error(), "acceptance_item_evidence_matrix invalid: each item must include id, status, and evidence_refs") {
		t.Fatalf("expected weak acceptance evidence failure, got %v", err)
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.NextCommand != "unit_verify" {
		t.Fatalf("weak verify evidence must not advance status, got %+v", status)
	}
}

func TestCloseRejectsUnitStableVerifyAlignedWhenEvidenceMissing(t *testing.T) {
	repoRoot := commandCloseStableSnapshotRepo(t, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_stable_verify` | test |\n")

	_, err := Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_stable_verify",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "aligned",
		Apply:      true,
	})
	if err == nil || !strings.Contains(err.Error(), "read docs/specs/_stable_verify_result/unit/demo.md") {
		t.Fatalf("expected missing stable verify evidence error, got %v", err)
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.NextCommand != "unit_stable_verify" {
		t.Fatalf("missing stable verify evidence must not advance status, got %+v", status)
	}
}

func TestCloseRejectsUnitStableVerifyAlignedWithoutIndependentReceipt(t *testing.T) {
	repoRoot := commandCloseStableSnapshotRepo(t, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_stable_verify` | test |\n")
	snap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCommandCloseStableVerifyProcess(t, repoRoot, snap, "aligned")
	processPath := filepath.Join(repoRoot, "docs/specs/_stable_verify_result/unit/demo.md")
	content, err := os.ReadFile(processPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	writeCommandCloseTestFile(t, processPath, strings.Replace(string(content), "evaluation_mode: independent\n", "", 1))

	_, err = Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_stable_verify",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "aligned",
		Apply:      true,
	})
	if err == nil || !strings.Contains(err.Error(), "missing required field: evaluation_mode") {
		t.Fatalf("expected missing independent receipt failure, got %v", err)
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.NextCommand != "unit_stable_verify" {
		t.Fatalf("missing independent receipt must not advance status, got %+v", status)
	}
}

func TestCloseUnitStableVerifyControlledRepairConsumesMatchingEvidence(t *testing.T) {
	repoRoot := commandCloseStableSnapshotRepo(t, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_stable_verify` | test |\n")
	snap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCommandCloseStableVerifyProcess(t, repoRoot, snap, "controlled_repair_required")

	result, err := Close(Options{
		RepoRoot:        repoRoot,
		Command:         "unit_stable_verify",
		ObjectType:      "unit",
		Object:          "demo",
		Outcome:         "controlled_repair_required",
		CandidateIntent: "repair",
		Apply:           true,
	})
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
	if result.ValidationAction != "validate_process:stable_verify" {
		t.Fatalf("expected stable verify validation, got %s", result.ValidationAction)
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.NextCommand != "unit_fork" {
		t.Fatalf("expected unit_fork after controlled repair, got %+v", status)
	}
}

func TestCloseUnitForkRequiresRepairIntentAfterControlledStableVerify(t *testing.T) {
	repoRoot := commandCloseStableSnapshotRepo(t, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | test |\n")
	snap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCommandCloseStableVerifyProcess(t, repoRoot, snap, "controlled_repair_required")
	writeCommandCloseCandidateSpecWithIntent(t, repoRoot, "repair")

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
	if result.CleanupAction != "success:unit_fork" {
		t.Fatalf("expected unit_fork success cleanup, got %s", result.CleanupAction)
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.Candidate != "yes" || status.ActiveLayer != "candidate" || status.NextCommand != "unit_check" {
		t.Fatalf("unexpected status after controlled repair fork: %+v", status)
	}
}

func TestCloseRejectsUnitForkWrongIntentAfterControlledStableVerify(t *testing.T) {
	repoRoot := commandCloseStableSnapshotRepo(t, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | test |\n")
	snap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCommandCloseStableVerifyProcess(t, repoRoot, snap, "controlled_repair_required")
	writeCommandCloseCandidateSpecWithIntent(t, repoRoot, "change")

	_, err = Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_fork",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "candidate_created",
		Apply:      true,
	})
	if err == nil || !strings.Contains(err.Error(), "candidate_intent mismatch for controlled stable verify decision controlled_repair_required: actual=change expected=repair") {
		t.Fatalf("expected controlled repair candidate intent mismatch, got %v", err)
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.Candidate != "no" || status.ActiveLayer != "stable" || status.NextCommand != "unit_fork" {
		t.Fatalf("failed controlled fork close must not advance status, got %+v", status)
	}
}

func TestCloseUnitForkRequiresChangeIntentAfterControlledStableVerify(t *testing.T) {
	repoRoot := commandCloseStableSnapshotRepo(t, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | test |\n")
	snap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCommandCloseStableVerifyProcess(t, repoRoot, snap, "controlled_change_required")
	writeCommandCloseCandidateSpecWithIntent(t, repoRoot, "change")

	_, err = Close(Options{
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
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.Candidate != "yes" || status.ActiveLayer != "candidate" || status.NextCommand != "unit_check" {
		t.Fatalf("unexpected status after controlled change fork: %+v", status)
	}
}

func TestCloseRejectsUnitStableVerifyDecisionMismatch(t *testing.T) {
	repoRoot := commandCloseStableSnapshotRepo(t, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_stable_verify` | test |\n")
	snap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCommandCloseStableVerifyProcess(t, repoRoot, snap, "controlled_change_required")

	_, err = Close(Options{
		RepoRoot:   repoRoot,
		Command:    "unit_stable_verify",
		ObjectType: "unit",
		Object:     "demo",
		Outcome:    "aligned",
		Apply:      true,
	})
	if err == nil || !strings.Contains(err.Error(), "decision mismatch: actual=controlled_change_required expected=aligned") {
		t.Fatalf("expected stable verify decision mismatch, got %v", err)
	}
}

func TestCloseRejectsScenarioObjectType(t *testing.T) {
	repoRoot := commandCloseSnapshotRepo(t, "| `scenario` | `checkout` | `no` | `yes` | `candidate` | `scenario_promote` | test |\n")
	_, err := Close(Options{
		RepoRoot:   repoRoot,
		Command:    "scenario_promote",
		ObjectType: "scenario",
		Object:     "checkout",
		Outcome:    "dependency_not_ready",
		Apply:      true,
	})
	if err == nil || !strings.Contains(err.Error(), "object-type must be unit") {
		t.Fatalf("expected scenario object type rejection, got %v", err)
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
	for _, command := range []string{"unit_plan", "unit_impl", "unit_verify"} {
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

func TestDetermineTransitionRejectsUnsupportedAndMismatchedFallbackReasons(t *testing.T) {
	for _, tc := range []struct {
		name string
		opts Options
		want string
	}{
		{
			name: "legacy truth reason",
			opts: Options{Command: "unit_impl", ObjectType: "unit", Object: "demo", Outcome: "truth_fallback", Reason: "truth_changed"},
			want: "unsupported fallback reason",
		},
		{
			name: "gate reason on plan fallback",
			opts: Options{Command: "unit_impl", ObjectType: "unit", Object: "demo", Outcome: "plan_fallback", Reason: "gate_missing"},
			want: "requires failure layer \"gate_layer\"",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := determineTransition(tc.opts, unitCandidateStatus(tc.opts.Command), true)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q failure, got %v", tc.want, err)
			}
		})
	}
}

func unitStableStatus(next string) statusfile.ObjectStatus {
	return statusfile.ObjectStatus{ObjectType: "unit", Object: "demo", Stable: "yes", Candidate: "no", ActiveLayer: "stable", NextCommand: next, Notes: "test"}
}

func unitCandidateStatus(next string) statusfile.ObjectStatus {
	return statusfile.ObjectStatus{ObjectType: "unit", Object: "demo", Stable: "no", Candidate: "yes", ActiveLayer: "candidate", NextCommand: next, Notes: "test"}
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

func commandCloseSnapshotRepo(t *testing.T, rows string) string {
	t.Helper()
	repoRoot := commandCloseTestRepo(t, rows)
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), `---
id: demo
layer: candidate
version: 0.1.0
---

# Demo

rule_refs: none

## Testability / Acceptance Criteria

acceptance_item_set:
  - id: demo.core
    target: Demo behavior is accepted.
    verification_surface: internal_flow
    implementation_surface: AgentCore/internal/demo
    verification_method: Go test for the demo behavior.
    pass_condition: The demo behavior passes under the declared checks.
    not_runnable_yet: no
`)
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/repository_mapping.md"), `---
id: repository_mapping
version: 0.1.0
---

# Repository Mapping
`)
	return repoRoot
}

func commandCloseStableSnapshotRepo(t *testing.T, rows string) string {
	t.Helper()
	repoRoot := commandCloseTestRepo(t, rows)
	writeCommandCloseStableSpec(t, repoRoot)
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/repository_mapping.md"), `---
id: repository_mapping
version: 0.1.0
---

# Repository Mapping
`)
	return repoRoot
}

func writeCommandCloseStableSpec(t *testing.T, repoRoot string) {
	t.Helper()
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_demo.md"), `---
id: demo
layer: stable
version: 1.0.0
---

# Demo

rule_refs: none

## Testability / Acceptance Criteria

acceptance_item_set:
  - id: demo.core
    target: Demo behavior is accepted.
    verification_surface: internal_flow
    implementation_surface: AgentCore/internal/demo
    verification_method: Go test for the demo behavior.
    pass_condition: The demo behavior passes under the declared checks.
    not_runnable_yet: no
`)
}

func replaceCommandCloseCandidateSpecText(t *testing.T, repoRoot, old, replacement string) {
	t.Helper()
	path := filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read candidate spec: %v", err)
	}
	updated := strings.Replace(string(content), old, replacement, 1)
	if updated == string(content) {
		t.Fatalf("candidate spec did not contain %q", old)
	}
	writeCommandCloseTestFile(t, path, updated)
}

func writeCommandCloseUnitCheckProcessWithFreshness(t *testing.T, repoRoot string, oldSnap, currentSnap snapshot.Snapshot) {
	t.Helper()
	writeCommandCloseUnitCheckProcessWithExtra(t, repoRoot, oldSnap, renderCommandCloseFreshnessReceipt(currentSnap))
}

func writeCommandCloseUnitCheckProcessWithExtra(t *testing.T, repoRoot string, snap snapshot.Snapshot, extra string) {
	t.Helper()
	lines := []string{
		"object_type: unit",
		"object_ref: demo",
		"gate: unit_check",
		"decision: pass",
		"allow_next: true",
		"next_command: unit_plan",
		"blocking_summary: none",
		"coverage_summary: demo",
		"truth_layer_ref: " + snap.TruthLayerRef,
		"truth_file_ref: " + snap.SpecFileRef,
		"truth_version_ref: " + snap.SpecVersionRef,
		"truth_fingerprint: " + snap.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + snap.AcceptanceBehaviorFingerprint,
		"unit_appendix_snapshot: none",
		"unit_snapshot: none",
		"rule_snapshot: none",
		"acceptance_item_set:",
		"  - id: demo.core",
		"    verification_surface: internal_flow",
		"    not_runnable_yet: no",
		"evaluation_mode: independent",
		"reviewer_result: pass",
		"reviewer_context: minimal_context",
		"review_input_refs: " + commandCloseReviewInputRefsForTest(snap.Object, "unit_check_pass", snap.SpecFileRef),
		"review_findings: none",
		"human_decision_refs: none",
	}
	if strings.TrimSpace(extra) != "" {
		lines = append(lines, extra)
	}
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), "# check\n\n```yaml\n"+strings.Join(lines, "\n")+"\n```\n")
}

func writeCommandCloseUnitPlanProcessWithFreshness(t *testing.T, repoRoot string, oldSnap, currentSnap snapshot.Snapshot, mode string) {
	t.Helper()
	extra := ""
	if mode == "accepted" || mode == "blocked" {
		extra = renderCommandCloseFreshnessReceipt(currentSnap)
		if mode == "blocked" {
			extra = strings.Replace(extra, "freshness_reviewer_result: pass", "freshness_reviewer_result: blocked", 1)
		}
	}
	writeCommandCloseUnitPlanProcessWithExtra(t, repoRoot, oldSnap, extra)
}

func writeCommandCloseUnitPlanProcessWithExtra(t *testing.T, repoRoot string, snap snapshot.Snapshot, extra string) {
	t.Helper()
	lines := []string{
		"spec_file_ref: " + snap.SpecFileRef,
		"spec_version_ref: " + snap.SpecVersionRef,
		"spec_fingerprint: " + snap.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + snap.AcceptanceBehaviorFingerprint,
		"stable_candidate_diff_refs: " + commandCloseStableCandidateDiffRefsForTest(repoRoot, snap),
		"implementation_gap_refs: docs/specs/repository_mapping.md",
		"unit_appendix_snapshot: none",
		"unit_snapshot: none",
		"rule_snapshot: none",
		"acceptance_item_plan_coverage:",
		"  - id: demo.core",
		"    coverage: implementation slice and verification target",
		"retirement_targets: none",
		"planned_change_scope:",
		"  - id: pcs.core",
		"    basis_refs: " + snap.SpecFileRef,
		"    acceptance_item_ids: demo.core",
		"    implementation_refs: docs/specs/repository_mapping.md",
		"    verification_action: verify package-aware delta",
		"package_constraint_review: pass",
		"package_constraint_refs: " + snap.SpecFileRef,
		"package_constraint_summary: current package constraints reviewed for this delta",
		"evaluation_mode: independent",
		"reviewer_result: pass",
		"reviewer_context: minimal_context",
		"review_input_refs: " + commandCloseReviewInputRefsForTest(snap.Object, "unit_plan_plan_ready", snap.SpecFileRef),
		"review_findings: none",
		"human_decision_refs: none",
	}
	if strings.TrimSpace(extra) != "" {
		lines = append(lines, extra)
	}
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/demo.md"), "# plan\n\n```yaml\n"+strings.Join(lines, "\n")+"\n```\n")
}

func renderCommandCloseFreshnessReceipt(currentSnap snapshot.Snapshot) string {
	return strings.Join([]string{
		"freshness_impact: text_drift",
		"evidence_reuse: accepted",
		"freshness_current_fingerprint: " + currentSnap.SpecFingerprint,
		"freshness_review_mode: independent",
		"freshness_reviewer_result: pass",
		"freshness_reviewer_context: minimal_context",
		"freshness_review_input_refs: " + commandCloseReviewInputRefsForTest(currentSnap.Object, "freshness_text_drift_reuse", currentSnap.SpecFileRef),
		"freshness_review_findings: none",
	}, "\n")
}

func commandCloseStableCandidateDiffRefsForTest(repoRoot string, snap snapshot.Snapshot) string {
	stableRef := "docs/specs/units/stable/s_unit_" + snap.Object + ".md"
	if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(stableRef))); err == nil {
		return stableRef + ";" + snap.SpecFileRef
	}
	return "none"
}

func commandCloseReviewInputRefsForTest(object, pack string, refs ...string) string {
	requestFile := filepath.ToSlash(filepath.Join("docs/specs/_independent_evaluation/requests/unit", object, pack+".md"))
	return strings.Join(append([]string{pack, requestFile}, refs...), ";")
}

func writeCommandCloseUnitCheckProcess(t *testing.T, repoRoot string, snap snapshot.Snapshot) {
	t.Helper()
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), "# check\n\n```yaml\n"+strings.Join([]string{
		"object_type: unit",
		"object_ref: demo",
		"gate: unit_check",
		"decision: pass",
		"allow_next: true",
		"next_command: unit_plan",
		"blocking_summary: none",
		"coverage_summary: demo",
		"truth_layer_ref: " + snap.TruthLayerRef,
		"truth_file_ref: " + snap.SpecFileRef,
		"truth_version_ref: " + snap.SpecVersionRef,
		"truth_fingerprint: " + snap.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + snap.AcceptanceBehaviorFingerprint,
		"unit_appendix_snapshot: none",
		"unit_snapshot: none",
		"rule_snapshot: none",
		"acceptance_item_set:",
		"  - id: demo.core",
		"    verification_surface: internal_flow",
		"    not_runnable_yet: no",
		"evaluation_mode: independent",
		"reviewer_result: pass",
		"reviewer_context: minimal_context",
		"review_input_refs: " + commandCloseReviewInputRefsForTest(snap.Object, "unit_check_pass", snap.SpecFileRef),
		"review_findings: none",
		"human_decision_refs: none",
	}, "\n")+"\n```\n")
}

func writeCommandCloseUnitVerifyProcess(t *testing.T, repoRoot string, snap snapshot.Snapshot) {
	t.Helper()
	activePlanFingerprint := commandCloseFileFingerprint(t, repoRoot, snapshot.ActivePlanFilePath(snap.Object))
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit/demo.md"), "# verify\n\n```yaml\n"+strings.Join([]string{
		"object_type: unit",
		"object_ref: demo",
		"gate: unit_verify",
		"decision: pass",
		"allow_next: true",
		"next_command: unit_promote",
		"blocking_summary: none",
		"coverage_summary: current candidate",
		"truth_layer_ref: " + snap.TruthLayerRef,
		"truth_file_ref: " + snap.SpecFileRef,
		"truth_version_ref: " + snap.SpecVersionRef,
		"truth_fingerprint: " + snap.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + snap.AcceptanceBehaviorFingerprint,
		"acceptance_item_set:",
		"  - id: demo.core",
		"    verification_surface: internal_flow",
		"    not_runnable_yet: no",
		"unit_appendix_snapshot: none",
		"unit_snapshot: none",
		"active_plan_file_ref: " + snapshot.ActivePlanFilePath(snap.Object),
		"active_plan_fingerprint: " + activePlanFingerprint,
		"rule_snapshot: none",
		"acceptance_item_evidence_matrix:",
		"  - id: demo.core",
		"    status: pass",
		"    evidence_refs: go test ./...",
		"retirement_evidence_matrix: none",
		"package_delta_verification:",
		"  - planned_change_scope_id: pcs.core",
		"    result: pass",
		"    evidence_refs: go test ./...",
		"evidence_refs: go test ./...",
		"evaluation_mode: independent",
		"reviewer_result: pass",
		"reviewer_context: minimal_context",
		"review_input_refs: " + commandCloseReviewInputRefsForTest(snap.Object, "unit_verify_ready_to_promote", snap.SpecFileRef),
		"review_findings: none",
		"human_decision_refs: none",
	}, "\n")+"\n```\n")
}

func commandCloseFileFingerprint(t *testing.T, repoRoot, fileRef string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(fileRef)))
	if err != nil {
		t.Fatalf("read %s: %v", fileRef, err)
	}
	text := strings.ReplaceAll(string(content), "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")
	text += "\n"
	sum := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", sum)
}

func writeCommandCloseStableVerifyProcess(t *testing.T, repoRoot string, snap snapshot.Snapshot, decision string) {
	t.Helper()
	mapping, err := snapshot.BuildRepositoryMappingSnapshot(repoRoot)
	if err != nil {
		t.Fatalf("BuildRepositoryMappingSnapshot: %v", err)
	}
	allowNext := "false"
	nextCommand := "unit_stable_verify"
	if decision == "aligned" || decision == "controlled_repair_required" || decision == "controlled_change_required" {
		allowNext = "true"
		nextCommand = "unit_fork"
	}
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/_stable_verify_result/unit/demo.md"), "# stable verify\n\n```yaml\n"+strings.Join([]string{
		"object_type: unit",
		"object_ref: demo",
		"gate: unit_stable_verify",
		"decision: " + decision,
		"allow_next: " + allowNext,
		"next_command: " + nextCommand,
		"blocking_summary: none",
		"coverage_summary: current stable implementation",
		"truth_layer_ref: stable",
		"truth_file_ref: " + snap.SpecFileRef,
		"truth_version_ref: " + snap.SpecVersionRef,
		"truth_fingerprint: " + snap.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + snap.AcceptanceBehaviorFingerprint,
		"repository_mapping_snapshot:",
		"  file_ref: " + mapping.FileRef,
		"  version_ref: " + mapping.VersionRef,
		"  fingerprint: " + mapping.Fingerprint,
		"acceptance_item_set:",
		"  - id: demo.core",
		"    verification_surface: internal_flow",
		"    not_runnable_yet: no",
		"unit_appendix_snapshot: none",
		"unit_snapshot: none",
		"rule_snapshot: none",
		"acceptance_item_evidence_matrix:",
		"  - id: demo.core",
		"    status: pass",
		"    evidence_refs: go test ./...",
		"implementation_surface_refs: AgentCore/internal/demo",
		"evidence_refs: go test ./...",
		"evaluation_mode: independent",
		"reviewer_result: pass",
		"reviewer_context: minimal_context",
		"review_input_refs: " + commandCloseReviewInputRefsForTest(snap.Object, "unit_stable_verify_advancing", snap.SpecFileRef),
		"review_findings: none",
		"human_decision_refs: none",
	}, "\n")+"\n```\n")
}

func writeCommandCloseCandidateSpecWithIntent(t *testing.T, repoRoot, intent string) {
	t.Helper()
	extraFrontmatter := []string{
		"candidate_intent: " + intent,
		"source_basis: new_design",
		"evidence_appendix_ref: none",
	}
	repairScope := ""
	if intent == "repair" {
		extraFrontmatter = append(extraFrontmatter, "repair_basis: s_unit_demo@1.0.0")
		repairScope = "\n## Repair Scope\n\nRestore `demo.core` to stable behavior.\n"
	}
	writeCommandCloseTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), `---
id: demo
layer: candidate
version: 1.0.1
`+strings.Join(extraFrontmatter, "\n")+`
---

# Demo
`+repairScope+`
## Testability / Acceptance Criteria

acceptance_item_set:
  - id: demo.core
    target: Demo behavior is accepted.
    verification_surface: internal_flow
    implementation_surface: AgentCore/internal/demo
    verification_method: Go test for the demo behavior.
    pass_condition: The demo behavior passes under the declared checks.
    not_runnable_yet: no
`)
}

func writeCommandCloseTestFile(t *testing.T, path, content string) {
	t.Helper()
	content = testfixtures.NormalizeSpecFlowContent(path, content)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
