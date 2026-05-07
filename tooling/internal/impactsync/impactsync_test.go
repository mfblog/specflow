package impactsync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
)

func TestApplyInvalidatesCandidateObjectsAndCleansProcessFiles(t *testing.T) {
	repoRoot := t.TempDir()
	setupImpactRepo(t, repoRoot, strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|---|",
		"| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_plan` | current round |",
		"| `scenario` | `demo` | `no` | `yes` | `candidate` | `scenario_verify` | current round |",
	}, "\n")+"\n")
	for _, relPath := range []string{
		"docs/specs/_check_result/unit/demo.md",
		"docs/specs/_plans/active/demo.md",
		"docs/specs/_plans/draft/demo.md",
		"docs/specs/_verify_result/unit/demo.md",
		"docs/specs/_check_result/scenario/demo.md",
		"docs/specs/_verify_result/scenario/demo.md",
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
		Flows: []ScopedObject{{
			Binding: ObjectBinding{
				ObjectType:    "scenario",
				Object:        "demo",
				ActiveLayer:   "candidate",
				NextCommand:   "scenario_verify",
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
	if len(result.FlowResults) != 1 || result.FlowResults[0].NextCommand != "scenario_check" || result.FlowResults[0].Outcome != "invalidated" {
		t.Fatalf("unexpected flow result: %+v", result.FlowResults)
	}
	for _, relPath := range []string{
		"docs/specs/_check_result/unit/demo.md",
		"docs/specs/_plans/active/demo.md",
		"docs/specs/_plans/draft/demo.md",
		"docs/specs/_verify_result/unit/demo.md",
		"docs/specs/_check_result/scenario/demo.md",
		"docs/specs/_verify_result/scenario/demo.md",
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
		"| `scenario` | `demo` | `no` | `yes` | `candidate` | `scenario_check` | current round |",
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
		"|---|---|---|---|---|---|---|---|",
		"| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | stable round |",
		"| `scenario` | `demo` | `yes` | `no` | `stable` | `scenario_fork` | stable round |",
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
		Flows: []ScopedObject{{
			Binding: ObjectBinding{
				ObjectType:    "scenario",
				Object:        "demo",
				ActiveLayer:   "stable",
				NextCommand:   "scenario_fork",
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
	if len(result.FlowResults) != 1 || result.FlowResults[0].NextCommand != "scenario_stable_verify" || result.FlowResults[0].Outcome != "rerouted" {
		t.Fatalf("unexpected flow result: %+v", result.FlowResults)
	}
	statusData, err := os.ReadFile(filepath.Join(repoRoot, "docs/specs/_status.md"))
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	statusText := string(statusData)
	for _, expected := range []string{
		"| `unit` | `demo` | `yes` | `no` | `stable` | `unit_stable_verify` | stable round |",
		"| `scenario` | `demo` | `yes` | `no` | `stable` | `scenario_stable_verify` | stable round |",
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
		"|---|---|---|---|---|---|---|---|",
		"| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | stable round |",
		"| `scenario` | `demo` | `yes` | `no` | `stable` | `scenario_fork` | stable round |",
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
		Flows: []ScopedObject{{
			Binding: ObjectBinding{
				ObjectType:  "scenario",
				Object:      "demo",
				ActiveLayer: "stable",
				NextCommand: "scenario_fork",
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
	if len(result.FlowResults) != 1 || result.FlowResults[0].FallbackReasonCode != "rule_drift" {
		t.Fatalf("unexpected flow result: %+v", result.FlowResults)
	}
}

func TestApplyUsesExplicitFallbackScopeForObjects(t *testing.T) {
	repoRoot := t.TempDir()
	setupImpactRepo(t, repoRoot, strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|---|",
		"| `scenario` | `demo` | `no` | `yes` | `candidate` | `scenario_verify` | current round |",
	}, "\n")+"\n")
	for _, relPath := range []string{
		"docs/specs/_check_result/scenario/demo.md",
		"docs/specs/_verify_result/scenario/demo.md",
	} {
		mustWriteImpactFile(t, filepath.Join(repoRoot, relPath), "# process\n")
	}

	result, err := Apply(repoRoot, Input{
		Flows: []ScopedObject{{
			Binding: ObjectBinding{
				ObjectType:  "scenario",
				Object:      "demo",
				ActiveLayer: "candidate",
				NextCommand: "scenario_verify",
			},
			ExplicitFallbackScope: true,
		}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if len(result.FlowResults) != 1 || result.FlowResults[0].FallbackReasonCode != "binding_drift" || result.FlowResults[0].NextCommand != "scenario_check" {
		t.Fatalf("unexpected flow result: %+v", result.FlowResults)
	}
}

func TestApplyClassifiesScenarioGateLayerAndKeepsVerify(t *testing.T) {
	repoRoot := t.TempDir()
	snap := setupImpactScenarioSnapshotRepo(t, repoRoot)
	checkPath := filepath.Join(repoRoot, "docs/specs/_check_result/scenario/checkout.md")
	verifyPath := filepath.Join(repoRoot, "docs/specs/_verify_result/scenario/checkout.md")
	mustWriteImpactFile(t, checkPath, strings.Replace(renderImpactScenarioCheckProcessSnapshot(snap), "coverage_summary: current candidate\n", "", 1))
	mustWriteImpactFile(t, verifyPath, renderImpactScenarioVerifyProcessSnapshot(snap, "pass"))

	result, err := Apply(repoRoot, Input{
		Flows: []ScopedObject{{
			Binding: ObjectBinding{
				ObjectType:  "scenario",
				Object:      "checkout",
				ActiveLayer: "candidate",
				NextCommand: "scenario_verify",
			},
			ValidateProcess: true,
		}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(result.FlowResults) != 1 {
		t.Fatalf("expected one flow result, got %+v", result.FlowResults)
	}
	flowResult := result.FlowResults[0]
	if flowResult.Outcome != "invalidated" || flowResult.FailureLayer != "gate_layer" || flowResult.NextCommand != "scenario_check" {
		t.Fatalf("expected gate_layer scenario_check, got %+v", flowResult)
	}
	if _, err := os.Stat(checkPath); !os.IsNotExist(err) {
		t.Fatalf("expected check file deleted, stat err=%v", err)
	}
	if _, err := os.Stat(verifyPath); err != nil {
		t.Fatalf("expected verify file to remain, stat err=%v", err)
	}
}

func TestApplyClassifiesScenarioEvidenceLayerAndKeepsCheck(t *testing.T) {
	repoRoot := t.TempDir()
	snap := setupImpactScenarioSnapshotRepo(t, repoRoot)
	checkPath := filepath.Join(repoRoot, "docs/specs/_check_result/scenario/checkout.md")
	verifyPath := filepath.Join(repoRoot, "docs/specs/_verify_result/scenario/checkout.md")
	mustWriteImpactFile(t, checkPath, renderImpactScenarioCheckProcessSnapshot(snap))
	mustWriteImpactFile(t, verifyPath, renderImpactScenarioVerifyProcessSnapshot(snap, "skipped"))

	result, err := Apply(repoRoot, Input{
		Flows: []ScopedObject{{
			Binding: ObjectBinding{
				ObjectType:  "scenario",
				Object:      "checkout",
				ActiveLayer: "candidate",
				NextCommand: "scenario_promote",
			},
			ValidateProcess: true,
		}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(result.FlowResults) != 1 {
		t.Fatalf("expected one flow result, got %+v", result.FlowResults)
	}
	flowResult := result.FlowResults[0]
	if flowResult.Outcome != "invalidated" || flowResult.FailureLayer != "evidence_layer" || flowResult.NextCommand != "scenario_verify" {
		t.Fatalf("expected evidence_layer scenario_verify, got %+v", flowResult)
	}
	if _, err := os.Stat(checkPath); err != nil {
		t.Fatalf("expected check file to remain, stat err=%v", err)
	}
	if _, err := os.Stat(verifyPath); !os.IsNotExist(err) {
		t.Fatalf("expected verify file deleted, stat err=%v", err)
	}
}

func TestApplyKeepsObjectsUnchangedWithoutFallbackInputs(t *testing.T) {
	repoRoot := t.TempDir()
	setupImpactRepo(t, repoRoot, strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|---|",
		"| `scenario` | `demo` | `no` | `yes` | `candidate` | `scenario_verify` | current round |",
	}, "\n")+"\n")

	result, err := Apply(repoRoot, Input{
		Flows: []ScopedObject{{
			Binding: ObjectBinding{
				ObjectType:  "scenario",
				Object:      "demo",
				ActiveLayer: "candidate",
				NextCommand: "scenario_verify",
			},
		}},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if len(result.FlowResults) != 1 || result.FlowResults[0].Outcome != "unchanged" {
		t.Fatalf("unexpected flow result: %+v", result.FlowResults)
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

func setupImpactRepo(t *testing.T, repoRoot, statusContent string) {
	t.Helper()
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/_check_result/scenario"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/_plans/active"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/_plans/draft"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/_verify_result/scenario"))
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

func setupImpactScenarioSnapshotRepo(t *testing.T, repoRoot string) snapshot.Snapshot {
	t.Helper()
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/scenarios/candidate"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/units/stable"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/rules/stable"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/_check_result/scenario"))
	mustMkdirImpactAll(t, filepath.Join(repoRoot, "docs/specs/_verify_result/scenario"))
	mustWriteImpactFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `ai` | `yes` | `no` | `stable` | `unit_fork` | stable dependency |",
		"| `scenario` | `checkout` | `no` | `yes` | `candidate` | `scenario_verify` | current round |",
	}, "\n")+"\n")
	mustWriteImpactFile(t, filepath.Join(repoRoot, "docs/specs/repository_mapping.md"), strings.Join([]string{
		"---",
		"version: 0.1.0",
		"---",
		"",
		"# Repository Mapping",
		"",
	}, "\n"))
	mustWriteImpactFile(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_g_rule_repository_baseline.md"), strings.Join([]string{
		"---",
		"rule_id: g_rule_repository_baseline",
		"rule_scope: global",
		"layer: stable",
		"rule_version: 1.0.0",
		"bound_objects: all",
		"---",
		"",
		"# Baseline",
		"",
	}, "\n"))
	mustWriteImpactFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_ai.md"), strings.Join([]string{
		"---",
		"id: ai",
		"layer: stable",
		"version: 1.0.0",
		"---",
		"",
		"# AI",
		"",
	}, "\n"))
	mustWriteImpactFile(t, filepath.Join(repoRoot, "docs/specs/scenarios/candidate/c_scenario_checkout.md"), strings.Join([]string{
		"---",
		"id: checkout",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Checkout",
		"",
		"## Bindings",
		"",
		"repository_mapping_ref: repository_mapping@0.1.0",
		"unit_refs:",
		"   - s_unit_ai@1.0.0",
		"rule_refs: none",
		"",
	}, "\n")+strings.Join(acceptanceSectionFixtureLines("checkout"), "\n")+"\n")

	snap, err := snapshot.RebuildCurrentObject(repoRoot, "scenario", "checkout")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	return snap
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
	lines = append(lines, "```", "")
	return strings.Join(lines, "\n")
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
		"unit_appendix_snapshot: none",
		"rule_snapshot:",
		"  - rule_id: " + snap.RuleSnapshot[0].RuleID,
		"    layer: " + snap.RuleSnapshot[0].Layer,
		"    file_ref: " + snap.RuleSnapshot[0].FileRef,
		"    version_ref: " + snap.RuleSnapshot[0].VersionRef,
		"    fingerprint: " + snap.RuleSnapshot[0].Fingerprint,
	}
	lines = append(lines, renderImpactAcceptancePlanCoverage(snap.AcceptanceItemSet)...)
	lines = append(lines,
		"```",
		"",
	)
	return strings.Join(lines, "\n")
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

func renderImpactScenarioCheckProcessSnapshot(snap snapshot.Snapshot) string {
	lines := []string{
		"# check",
		"",
		"```yaml",
		"object_type: scenario",
		"object_ref: " + snap.Object,
		"gate: scenario_check",
		"decision: pass",
		"allow_next: true",
		"next_command: scenario_verify",
		"blocking_summary: none",
		"coverage_summary: current candidate",
		"truth_layer_ref: " + snap.TruthLayerRef,
		"truth_file_ref: " + snap.SpecFileRef,
		"truth_version_ref: " + snap.SpecVersionRef,
		"truth_fingerprint: " + snap.SpecFingerprint,
	}
	lines = append(lines, renderImpactRepositoryMappingSnapshot(snap.RepositoryMapping)...)
	lines = append(lines, renderImpactObjectSnapshot("unit_snapshot", "unit", snap.UnitSnapshot)...)
	lines = append(lines, renderImpactRuleSnapshot(snap.RuleSnapshot)...)
	lines = append(lines, renderImpactAcceptanceItemSet(snap.AcceptanceItemSet)...)
	lines = append(lines, "```", "")
	return strings.Join(lines, "\n")
}

func renderImpactScenarioVerifyProcessSnapshot(snap snapshot.Snapshot, status string) string {
	lines := []string{
		"# verify",
		"",
		"```yaml",
		"object_type: scenario",
		"object_ref: " + snap.Object,
		"gate: scenario_verify",
		"decision: pass",
		"allow_next: true",
		"next_command: scenario_promote",
		"blocking_summary: none",
		"coverage_summary: current candidate",
		"truth_layer_ref: " + snap.TruthLayerRef,
		"truth_file_ref: " + snap.SpecFileRef,
		"truth_version_ref: " + snap.SpecVersionRef,
		"truth_fingerprint: " + snap.SpecFingerprint,
	}
	lines = append(lines, renderImpactRepositoryMappingSnapshot(snap.RepositoryMapping)...)
	lines = append(lines, renderImpactObjectSnapshot("unit_snapshot", "unit", snap.UnitSnapshot)...)
	lines = append(lines,
		"verification_scope_ref: current",
	)
	lines = append(lines, renderImpactRuleSnapshot(snap.RuleSnapshot)...)
	lines = append(lines, renderImpactAcceptanceItemSet(snap.AcceptanceItemSet)...)
	lines = append(lines, renderImpactAcceptanceEvidence(snap.AcceptanceItemSet, status)...)
	lines = append(lines, "```", "")
	return strings.Join(lines, "\n")
}

func renderImpactRepositoryMappingSnapshot(entry snapshot.RepositoryMappingEntry) []string {
	return []string{
		"repository_mapping_snapshot:",
		"  file_ref: " + entry.FileRef,
		"  version_ref: " + entry.VersionRef,
		"  fingerprint: " + entry.Fingerprint,
	}
}

func renderImpactObjectSnapshot(fieldName, objectField string, entries []snapshot.ObjectSnapshotEntry) []string {
	if len(entries) == 0 {
		return []string{fieldName + ": none"}
	}
	lines := []string{fieldName + ":"}
	for _, entry := range entries {
		lines = append(lines,
			"  - "+objectField+": "+entry.ObjectRef,
			"    layer: "+entry.Layer,
			"    file_ref: "+entry.FileRef,
			"    version_ref: "+entry.VersionRef,
			"    fingerprint: "+entry.Fingerprint,
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

func renderImpactAcceptanceEvidence(entries []snapshot.AcceptanceItemEntry, status string) []string {
	if len(entries) == 0 {
		return []string{"acceptance_item_evidence_matrix: none"}
	}
	lines := []string{"acceptance_item_evidence_matrix:"}
	for _, entry := range entries {
		lines = append(lines,
			"  - id: "+entry.ID,
			"    status: "+status,
		)
	}
	return lines
}
