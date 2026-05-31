package impactsync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/testfixtures"
)

func TestApplyInvalidatesCandidateObjectsAndCleansProcessFiles(t *testing.T) {
	repoRoot := t.TempDir()
	setupImpactRepo(t, repoRoot, strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_plan` | current round |",
	}, "\n")+"\n")
	for _, relPath := range []string{
		"docs/specs/_check_result/unit/demo.md",
		"docs/specs/_plans/active/demo.md",
		"docs/specs/_plans/draft/demo.md",
		"docs/specs/_verify_result/unit/demo.md",
	} {
		mustWriteImpactFile(t, filepath.Join(repoRoot, relPath), "# process\n")
	}

	result, err := Apply(repoRoot, Input{
		Modules: []ScopedModule{{
			Binding: ModuleBinding{
				Module:        "demo",
				ActiveLayer:   "candidate",
				NextCommand:   "unit_plan",
				BindingIssues: []string{"binding drift"},
			},
		}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if len(result.ModuleResults) != 1 || result.ModuleResults[0].NextCommand != "unit_check" || result.ModuleResults[0].Outcome != "invalidated" {
		t.Fatalf("unexpected module result: %+v", result.ModuleResults)
	}
	for _, relPath := range []string{
		"docs/specs/_check_result/unit/demo.md",
		"docs/specs/_plans/active/demo.md",
		"docs/specs/_plans/draft/demo.md",
		"docs/specs/_verify_result/unit/demo.md",
	} {
		if _, err := os.Stat(filepath.Join(repoRoot, relPath)); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be deleted, stat err=%v", relPath, err)
		}
	}

	statusData, err := os.ReadFile(filepath.Join(repoRoot, "docs/specs/_status.md"))
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	statusText := string(statusData)
	for _, expected := range []string{
		"| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | current round |",
	} {
		if !strings.Contains(statusText, expected) {
			t.Fatalf("status row %q not updated:\n%s", expected, statusText)
		}
	}
}

func TestApplyReroutesStableObjectsToVerifyCommands(t *testing.T) {
	repoRoot := t.TempDir()
	setupImpactRepo(t, repoRoot, strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | stable round |",
	}, "\n")+"\n")

	result, err := Apply(repoRoot, Input{
		Modules: []ScopedModule{{
			Binding: ModuleBinding{
				Module:        "demo",
				ActiveLayer:   "stable",
				NextCommand:   "unit_fork",
				BindingIssues: []string{"binding drift"},
			},
		}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if len(result.ModuleResults) != 1 || result.ModuleResults[0].NextCommand != "unit_stable_verify" || result.ModuleResults[0].Outcome != "rerouted" {
		t.Fatalf("unexpected module result: %+v", result.ModuleResults)
	}
	statusData, err := os.ReadFile(filepath.Join(repoRoot, "docs/specs/_status.md"))
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	statusText := string(statusData)
	for _, expected := range []string{
		"| `unit` | `demo` | `yes` | `no` | `stable` | `unit_stable_verify` | stable round |",
	} {
		if !strings.Contains(statusText, expected) {
			t.Fatalf("status row %q not updated:\n%s", expected, statusText)
		}
	}
}

func TestApplyUsesResolvedSharedInvalidationForStableObjects(t *testing.T) {
	repoRoot := t.TempDir()
	setupImpactRepo(t, repoRoot, strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | stable round |",
	}, "\n")+"\n")

	result, err := Apply(repoRoot, Input{
		Modules: []ScopedModule{{
			Binding: ModuleBinding{
				Module:      "demo",
				ActiveLayer: "stable",
				NextCommand: "unit_fork",
			},
			InvalidatingRuleRefs: []string{"s_b_rule_demo@1.0.0"},
		}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if len(result.ModuleResults) != 1 || result.ModuleResults[0].FallbackReasonCode != "rule_drift" {
		t.Fatalf("unexpected module result: %+v", result.ModuleResults)
	}
}

func TestApplyKeepsCandidateModuleWhenCallerAllowsSharedSnapshotMismatch(t *testing.T) {
	repoRoot := t.TempDir()
	allowedFileRef := setupImpactModuleSharedRepo(t, repoRoot)

	mustWriteImpactFile(t, filepath.Join(repoRoot, allowedFileRef), strings.Join([]string{
		"---",
		"rule_id: shared_demo",
		"rule_scope: bound",
		"layer: candidate",
		"rule_version: 0.1.0",
		"bound_objects:",
		"  - unit:demo",
		"---",
		"",
		"# Shared",
		"",
		"Body changed.",
		"",
	}, "\n"))

	result, err := Apply(repoRoot, Input{
		Modules: []ScopedModule{{
			Binding: ModuleBinding{
				Module:      "demo",
				ActiveLayer: "candidate",
				NextCommand: "unit_plan",
			},
			AllowedSharedSnapshotMismatchFileRefs: []string{"docs/specs/rules/candidate/c_b_rule_demo.md"},
		}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if len(result.ModuleResults) != 1 {
		t.Fatalf("expected one module result, got %+v", result.ModuleResults)
	}
	moduleResult := result.ModuleResults[0]
	if moduleResult.Outcome != "unchanged" {
		t.Fatalf("expected unchanged outcome, got %+v", moduleResult)
	}
	if moduleResult.NextCommand != "unit_plan" {
		t.Fatalf("expected next command unit_plan, got %+v", moduleResult)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md")); err != nil {
		t.Fatalf("expected process file to remain, stat err=%v", err)
	}
}

func TestApplyReportsFreshnessReviewRequiredWithoutCleanup(t *testing.T) {
	repoRoot := t.TempDir()
	setupImpactModuleSharedRepo(t, repoRoot)
	replaceImpactCandidateSpecText(t, repoRoot, "# Demo\n", "# Demo\n\nEditorial note only.\n")

	result, err := Apply(repoRoot, Input{
		Modules: []ScopedModule{{
			Binding: ModuleBinding{
				Module:      "demo",
				ActiveLayer: "candidate",
				NextCommand: "unit_plan",
			},
		}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	moduleResult := result.ModuleResults[0]
	if moduleResult.Outcome != "freshness_review_required" || moduleResult.FailureLayer != "freshness_layer" {
		t.Fatalf("expected freshness review result, got %+v", moduleResult)
	}
	if moduleResult.StatusUpdated || len(moduleResult.DeletedFiles) != 0 {
		t.Fatalf("freshness review must be non-destructive, got %+v", moduleResult)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md")); err != nil {
		t.Fatalf("expected process file to remain, stat err=%v", err)
	}
}

func TestApplyKeepsCandidateModuleWithAcceptedTextDrift(t *testing.T) {
	repoRoot := t.TempDir()
	setupImpactModuleSharedRepo(t, repoRoot)
	oldSnap, err := snapshot.RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	replaceImpactCandidateSpecText(t, repoRoot, "# Demo\n", "# Demo\n\nEditorial note only.\n")
	currentSnap, err := snapshot.RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent after edit: %v", err)
	}
	mustWriteImpactFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), renderImpactCheckProcessSnapshotWithFreshness(oldSnap, currentSnap))

	result, err := Apply(repoRoot, Input{
		Modules: []ScopedModule{{
			Binding: ModuleBinding{
				Module:      "demo",
				ActiveLayer: "candidate",
				NextCommand: "unit_plan",
			},
		}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	moduleResult := result.ModuleResults[0]
	if moduleResult.Outcome != "unchanged" || moduleResult.NextCommand != "unit_plan" {
		t.Fatalf("expected unchanged accepted text drift, got %+v", moduleResult)
	}
}

func TestApplyFallbackStillRunsForSemanticDrift(t *testing.T) {
	repoRoot := t.TempDir()
	setupImpactModuleSharedRepo(t, repoRoot)
	replaceImpactCandidateSpecText(t, repoRoot, "pass_condition: demo behavior passes the declared checks.", "pass_condition: demo behavior now requires changed checks.")

	result, err := Apply(repoRoot, Input{
		Modules: []ScopedModule{{
			Binding: ModuleBinding{
				Module:      "demo",
				ActiveLayer: "candidate",
				NextCommand: "unit_plan",
			},
		}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	moduleResult := result.ModuleResults[0]
	if moduleResult.Outcome != "invalidated" || moduleResult.NextCommand != "unit_check" {
		t.Fatalf("expected semantic drift fallback, got %+v", moduleResult)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md")); !os.IsNotExist(err) {
		t.Fatalf("expected process file cleanup for semantic drift, stat err=%v", err)
	}
}

func TestApplyKeepsCandidateModuleWhenPlanUsesPlanContract(t *testing.T) {
	repoRoot := t.TempDir()
	setupImpactModuleSharedRepo(t, repoRoot)

	snap, err := snapshot.RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	mustWriteImpactFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/demo.md"), renderImpactPlanProcessSnapshot(snap))

	result, err := Apply(repoRoot, Input{
		Modules: []ScopedModule{{
			Binding: ModuleBinding{
				Module:      "demo",
				ActiveLayer: "candidate",
				NextCommand: "unit_verify",
			},
		}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if len(result.ModuleResults) != 1 {
		t.Fatalf("expected one module result, got %+v", result.ModuleResults)
	}
	moduleResult := result.ModuleResults[0]
	if moduleResult.Outcome != "unchanged" || moduleResult.NextCommand != "unit_verify" {
		t.Fatalf("expected unchanged module with valid plan contract, got %+v", moduleResult)
	}
}

func TestApplyUsesPlanDriftReasonForInvalidPlanEvidence(t *testing.T) {
	repoRoot := t.TempDir()
	setupImpactModuleSharedRepo(t, repoRoot)

	snap, err := snapshot.RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	body := strings.Replace(renderImpactPlanProcessSnapshot(snap), "    coverage: covered", "", 1)
	mustWriteImpactFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/demo.md"), body)

	result, err := Apply(repoRoot, Input{
		Modules: []ScopedModule{{
			Binding: ModuleBinding{
				Module:      "demo",
				ActiveLayer: "candidate",
				NextCommand: "unit_verify",
			},
		}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	moduleResult := result.ModuleResults[0]
	if moduleResult.Outcome != "invalidated" || moduleResult.FailureLayer != "plan_layer" ||
		moduleResult.FallbackReasonCode != "plan_drift" || moduleResult.NextCommand != "unit_plan" {
		t.Fatalf("expected plan_drift fallback to unit_plan, got %+v", moduleResult)
	}
}

func setupImpactRepo(t *testing.T, repoRoot, statusContent string) {
	t.Helper()
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/_plans/active"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/_plans/draft"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit"))
	mustWriteImpactFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), statusContent)
}

func setupImpactModuleSharedRepo(t *testing.T, repoRoot string) string {
	t.Helper()
	mustMkdirImpactAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateDir)))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/rules/candidate"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/_plans/active"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/_plans/draft"))
	mustWriteImpactFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_plan` | current round |",
	}, "\n")+"\n")

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteImpactFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs:",
		"   - c_b_rule_demo@0.1.0",
		"",
	}, "\n"))

	sharedPath := filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md")
	mustWriteImpactFile(t, sharedPath, strings.Join([]string{
		"---",
		"rule_id: shared_demo",
		"rule_scope: bound",
		"layer: candidate",
		"rule_version: 0.1.0",
		"bound_objects:",
		"  - unit:demo",
		"---",
		"",
		"# Shared",
		"",
		"Body stays the same.",
		"",
	}, "\n"))

	snap, err := snapshot.RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	mustWriteImpactFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), renderImpactCheckProcessSnapshot(snap))
	return "docs/specs/rules/candidate/c_b_rule_demo.md"
}

func replaceImpactCandidateSpecText(t *testing.T, repoRoot, old, replacement string) {
	t.Helper()
	path := filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateDir), "c_unit_demo.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read candidate spec: %v", err)
	}
	updated := strings.Replace(string(content), old, replacement, 1)
	if updated == string(content) {
		t.Fatalf("candidate spec did not contain %q", old)
	}
	mustWriteImpactFile(t, path, updated)
}

func mustMkdirImpactAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteImpactFile(t *testing.T, path, content string) {
	t.Helper()
	content = withCandidateAcceptanceFixture(path, content)
	content = testfixtures.NormalizeSpecFlowContent(path, content)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func withCandidateAcceptanceFixture(path, content string) string {
	normalizedPath := filepath.ToSlash(path)
	if !strings.Contains(normalizedPath, "docs/specs/units/candidate/c_unit_") {
		return content
	}
	if strings.Contains(content, "acceptance_item_set:") {
		return content
	}
	object := strings.TrimSuffix(filepath.Base(path), ".md")
	object = strings.TrimPrefix(object, "c_unit_")
	lines := append([]string{
		strings.TrimRight(content, "\n"),
		"",
	}, acceptanceSectionFixtureLines(object)...)
	return strings.Join(lines, "\n") + "\n"
}

func acceptanceSectionFixtureLines(object string) []string {
	return []string{
		"## Testability / Acceptance Criteria",
		"",
		"acceptance_item_set:",
		"  - id: " + object + ".acceptance",
		"    target: " + object + " behavior is accepted.",
		"    verification_surface: internal_flow",
		"    implementation_surface: AgentCore/internal/" + object,
		"    verification_method: Go test for " + object + " behavior.",
		"    pass_condition: " + object + " behavior passes the declared checks.",
		"    not_runnable_yet: no",
	}
}

func renderImpactCheckProcessSnapshot(snap snapshot.Snapshot) string {
	lines := []string{
		"# check",
		"",
		"```yaml",
		"object_type: unit",
		"object_ref: " + snap.Module,
		"gate: unit_check",
		"decision: pass",
		"allow_next: true",
		"next_command: unit_plan",
		"blocking_summary: none",
		"coverage_summary: current candidate",
		"truth_layer_ref: " + snap.TruthLayerRef,
		"truth_file_ref: " + snap.SpecFileRef,
		"truth_version_ref: " + snap.SpecVersionRef,
		"truth_fingerprint: " + snap.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + snap.AcceptanceBehaviorFingerprint,
	}
	lines = append(lines, renderImpactAcceptanceItemSet(snap.AcceptanceItemSet)...)
	lines = append(lines,
		"unit_appendix_snapshot: none",
		"rule_snapshot:",
	)
	for _, entry := range snap.RuleSnapshot {
		lines = append(lines,
			"  - rule_id: "+entry.RuleID,
			"    layer: "+entry.Layer,
			"    file_ref: "+entry.FileRef,
			"    version_ref: "+entry.VersionRef,
			"    fingerprint: "+entry.Fingerprint,
		)
	}
	if len(snap.RuleSnapshot) == 0 {
		lines[len(lines)-1] = "rule_snapshot: none"
	}
	lines = append(lines, renderImpactIndependentEvaluationReceipt(snap.Object, "unit_check_pass", snap.SpecFileRef)...)
	lines = append(lines, "```", "")
	return strings.Join(lines, "\n")
}

func renderImpactCheckProcessSnapshotWithFreshness(oldSnap, currentSnap snapshot.Snapshot) string {
	body := renderImpactCheckProcessSnapshot(oldSnap)
	receipt := strings.Join([]string{
		"freshness_impact: text_drift",
		"evidence_reuse: accepted",
		"freshness_current_fingerprint: " + currentSnap.SpecFingerprint,
		"freshness_review_mode: independent",
		"freshness_reviewer_result: pass",
		"freshness_reviewer_context: minimal_context",
		"freshness_review_input_refs: " + impactReviewInputRefsForTest(currentSnap.Object, "freshness_text_drift_reuse", currentSnap.SpecFileRef),
		"freshness_review_findings: none",
	}, "\n")
	return strings.Replace(body, "\n```\n", "\n"+receipt+"\n```\n", 1)
}

func renderImpactAcceptanceItemSet(entries []snapshot.AcceptanceItemEntry) []string {
	if len(entries) == 0 {
		return []string{"acceptance_item_set: none"}
	}
	lines := []string{"acceptance_item_set:"}
	for _, entry := range entries {
		lines = append(lines,
			"  - id: "+entry.ID,
			"    verification_surface: "+entry.VerificationSurface,
			"    not_runnable_yet: "+entry.NotRunnableYet,
		)
	}
	return lines
}

func renderImpactPlanProcessSnapshot(snap snapshot.Snapshot) string {
	lines := []string{
		"# plan",
		"",
		"```yaml",
		"spec_file_ref: " + snap.SpecFileRef,
		"spec_version_ref: " + snap.SpecVersionRef,
		"spec_fingerprint: " + snap.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + snap.AcceptanceBehaviorFingerprint,
		"unit_appendix_snapshot: none",
		"rule_snapshot:",
		"  - rule_id: " + snap.RuleSnapshot[0].RuleID,
		"    layer: " + snap.RuleSnapshot[0].Layer,
		"    file_ref: " + snap.RuleSnapshot[0].FileRef,
		"    version_ref: " + snap.RuleSnapshot[0].VersionRef,
		"    fingerprint: " + snap.RuleSnapshot[0].Fingerprint,
	}
	lines = append(lines, renderImpactAcceptancePlanCoverage(snap.AcceptanceItemSet)...)
	lines = append(lines, "retirement_targets: none")
	lines = append(lines, renderImpactIndependentEvaluationReceipt(snap.Object, "unit_plan_plan_ready", snap.SpecFileRef)...)
	lines = append(lines,
		"```",
		"",
	)
	return strings.Join(lines, "\n")
}

func renderImpactIndependentEvaluationReceipt(object, pack, reviewInputRef string) []string {
	return []string{
		"evaluation_mode: independent",
		"reviewer_result: pass",
		"reviewer_context: minimal_context",
		"review_input_refs: " + impactReviewInputRefsForTest(object, pack, reviewInputRef),
		"review_findings: none",
		"human_decision_refs: none",
	}
}

func impactReviewInputRefsForTest(object, pack string, refs ...string) string {
	requestFile := filepath.ToSlash(filepath.Join("docs/specs/_independent_evaluation/requests/unit", object, pack+".md"))
	return strings.Join(append([]string{pack, requestFile}, refs...), ";")
}

func renderImpactAcceptancePlanCoverage(entries []snapshot.AcceptanceItemEntry) []string {
	if len(entries) == 0 {
		return []string{"acceptance_item_plan_coverage: none"}
	}
	lines := []string{"acceptance_item_plan_coverage:"}
	for _, entry := range entries {
		lines = append(lines,
			"  - id: "+entry.ID,
			"    coverage: covered",
		)
	}
	return lines
}

func renderImpactRuleSnapshot(entries []snapshot.RuleEntry) []string {
	if len(entries) == 0 {
		return []string{"rule_snapshot: none"}
	}
	lines := []string{"rule_snapshot:"}
	for _, entry := range entries {
		lines = append(lines,
			"  - rule_id: "+entry.RuleID,
			"    layer: "+entry.Layer,
			"    file_ref: "+entry.FileRef,
			"    version_ref: "+entry.VersionRef,
			"    fingerprint: "+entry.Fingerprint,
		)
	}
	return lines
}
