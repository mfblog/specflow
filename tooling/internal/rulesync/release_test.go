package rulesync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
)

func TestConsumersReadsCurrentFrontmatterRuleRefs(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/stable"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/units/stable"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/units/candidate"))

	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `agent` | `yes` | `no` | `stable` | `unit_fork` | stable |",
		"| `unit` | `trace` | `no` | `yes` | `candidate` | `unit_check` | candidate |",
	}, "\n")+"\n")
	writeUnitSpecWithRuleRefs(t, repoRoot, "stable", "agent", []string{"s_b_rule_demo@0.2.0"})
	writeUnitSpecWithRuleRefs(t, repoRoot, "candidate", "trace", []string{"s_b_rule_demo@0.2.0"})
	writeStableSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: stable
rule_version: 0.2.0
---

# Shared
`)

	result, err := Consumers(repoRoot, ConsumerOptions{RuleID: "shared_demo"})
	if err != nil {
		t.Fatalf("Consumers: %v", err)
	}
	if len(result.Consumers) != 2 {
		t.Fatalf("expected two consumers, got %+v", result.Consumers)
	}
	if result.Consumers[0].Object != "agent" || result.Consumers[1].Object != "trace" {
		t.Fatalf("unexpected consumers: %+v", result.Consumers)
	}
}

func TestReleaseVersionUpdatesCandidateAndAutoForksStableConsumer(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/stable"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/units/stable"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/units/stable/appendix"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/units/candidate"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/units/candidate/appendix"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_plans/active"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit"))

	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `agent` | `yes` | `no` | `stable` | `unit_fork` | stable |",
		"| `unit` | `trace` | `no` | `yes` | `candidate` | `unit_plan` | candidate |",
	}, "\n")+"\n")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_agent.md"), strings.Join([]string{
		"---",
		"id: agent",
		"layer: stable",
		"version: 0.1.0",
		"rule_refs:",
		"  - s_b_rule_demo@0.1.0",
		"---",
		"",
		"# agent",
		"",
		"Roles appendix: [`docs/specs/units/stable/appendix/s_unit_agent_roles.md`](./appendix/s_unit_agent_roles.md).",
		"",
		"## Testability / Acceptance Criteria",
		"",
		"acceptance_item_set:",
		"  - id: agent.acceptance",
		"    target: agent behavior is accepted.",
		"    verification_surface: internal_flow",
		"    implementation_surface: AgentCore/internal/agent",
		"    verification_method: Go test for agent behavior.",
		"    pass_condition: agent behavior passes the declared checks.",
		"    not_runnable_yet: no",
		"",
	}, "\n"))
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/appendix/s_unit_agent_roles.md"), strings.Join([]string{
		"---",
		"unit: agent",
		"layer: stable",
		"spec_version_ref: s_unit_agent@0.1.0",
		"---",
		"",
		"# Agent Roles",
		"",
		"See [`s_unit_agent.md`](../s_unit_agent.md).",
		"Non-reference text prefix_s_unit_agent.md_suffix stays unchanged.",
		"",
	}, "\n"))
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/appendix/c_unit_agent_old.md"), "stale")
	writeUnitSpecWithRuleRefs(t, repoRoot, "candidate", "trace", []string{"s_b_rule_demo@0.1.0"})
	writeStableSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: stable
rule_version: 0.2.0
---

# Shared
`)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/trace.md"), "check")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/trace.md"), "plan")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit/trace.md"), "verify")

	result, err := ReleaseVersion(repoRoot, ReleaseVersionOptions{
		RuleID:  "shared_demo",
		FromRef: "s_b_rule_demo@0.1.0",
		ToRef:   "s_b_rule_demo@0.2.0",
	})
	if err != nil {
		t.Fatalf("ReleaseVersion: %v", err)
	}
	if len(result.StableForked) != 1 || result.StableForked[0] != "unit:agent" {
		t.Fatalf("expected stable agent fork, got %+v", result.StableForked)
	}
	if len(result.CandidateUpdated) != 1 || result.CandidateUpdated[0] != "unit:trace" {
		t.Fatalf("expected candidate trace update, got %+v", result.CandidateUpdated)
	}
	if len(result.AppendixRetargeted) != 1 || result.AppendixRetargeted[0] != "docs/specs/units/candidate/appendix/c_unit_agent_roles.md" {
		t.Fatalf("expected agent appendix retarget, got %+v", result.AppendixRetargeted)
	}
	if len(result.AppendixRemoved) != 1 || result.AppendixRemoved[0] != "docs/specs/units/candidate/appendix/c_unit_agent_old.md" {
		t.Fatalf("expected stale agent appendix removal, got %+v", result.AppendixRemoved)
	}

	agentCandidate, err := os.ReadFile(filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_agent.md"))
	if err != nil {
		t.Fatalf("read candidate agent: %v", err)
	}
	if !strings.Contains(string(agentCandidate), "layer: candidate") ||
		!strings.Contains(string(agentCandidate), "version: 0.1.1") ||
		!strings.Contains(string(agentCandidate), "candidate_intent: change") ||
		!strings.Contains(string(agentCandidate), "  - s_b_rule_demo@0.2.0") {
		t.Fatalf("candidate agent was not correctly forked:\n%s", string(agentCandidate))
	}
	if !strings.Contains(string(agentCandidate), "./appendix/c_unit_agent_roles.md") ||
		strings.Contains(string(agentCandidate), "s_unit_agent_roles.md") {
		t.Fatalf("candidate agent appendix link was not retargeted:\n%s", string(agentCandidate))
	}
	agentAppendix, err := os.ReadFile(filepath.Join(repoRoot, "docs/specs/units/candidate/appendix/c_unit_agent_roles.md"))
	if err != nil {
		t.Fatalf("read candidate agent appendix: %v", err)
	}
	agentAppendixText := string(agentAppendix)
	if !strings.Contains(agentAppendixText, "layer: candidate") ||
		strings.Contains(agentAppendixText, "spec_version_ref:") ||
		!strings.Contains(agentAppendixText, "../c_unit_agent.md") ||
		strings.Contains(agentAppendixText, "../s_unit_agent.md") ||
		strings.Contains(agentAppendixText, "`s_unit_agent.md`") ||
		!strings.Contains(agentAppendixText, "prefix_s_unit_agent.md_suffix") {
		t.Fatalf("candidate agent appendix was not correctly retargeted:\n%s", agentAppendixText)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "docs/specs/units/candidate/appendix/c_unit_agent_old.md")); !os.IsNotExist(err) {
		t.Fatalf("expected stale candidate appendix to be removed, stat err=%v", err)
	}
	agentStable, err := os.ReadFile(filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_agent.md"))
	if err != nil {
		t.Fatalf("read stable agent: %v", err)
	}
	if !strings.Contains(string(agentStable), "  - s_b_rule_demo@0.1.0") {
		t.Fatalf("stable truth should remain untouched:\n%s", string(agentStable))
	}
	for _, relPath := range []string{
		"docs/specs/_check_result/unit/trace.md",
		"docs/specs/_plans/active/trace.md",
		"docs/specs/_verify_result/unit/trace.md",
	} {
		if _, err := os.Stat(filepath.Join(repoRoot, relPath)); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed, err=%v", relPath, err)
		}
	}
	if diagnostics := ValidateCurrentBindings(repoRoot, "s_b_rule_demo@0.1.0"); len(diagnostics) != 0 {
		t.Fatalf("expected current bindings to validate, got %+v", diagnostics)
	}
}

func TestReleaseVersionRejectsControlledRepairStableConsumerBeforeWriting(t *testing.T) {
	repoRoot := setupReleaseStableAutoForkRepo(t, "unit_fork")
	writeReleaseStableVerifyProcess(t, repoRoot, "agent", "controlled_repair_required")
	writeReleaseStableSharedVersion(t, repoRoot, "0.2.0")

	statusBefore := readReleaseTestFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"))
	traceBefore := readReleaseTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_trace.md"))
	traceCheckPath := filepath.Join(repoRoot, "docs/specs/_check_result/unit/trace.md")
	traceCheckBefore := readReleaseTestFile(t, traceCheckPath)
	stableVerifyPath := releaseStableVerifyAbsPath(t, repoRoot, "agent")
	stableVerifyBefore := readReleaseTestFile(t, stableVerifyPath)

	result, err := ReleaseVersion(repoRoot, ReleaseVersionOptions{
		RuleID:  "shared_demo",
		FromRef: "s_b_rule_demo@0.1.0",
		ToRef:   "s_b_rule_demo@0.2.0",
	})
	if err == nil ||
		!strings.Contains(err.Error(), "controlled_repair_required") ||
		!strings.Contains(err.Error(), "candidate_intent=repair") {
		t.Fatalf("expected controlled repair rejection, got result=%+v err=%v", result, err)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_agent.md")); !os.IsNotExist(err) {
		t.Fatalf("candidate agent must not be written, stat err=%v", err)
	}
	if got := readReleaseTestFile(t, filepath.Join(repoRoot, "docs/specs/_status.md")); got != statusBefore {
		t.Fatalf("status must not change\nbefore:\n%s\nafter:\n%s", statusBefore, got)
	}
	if got := readReleaseTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_trace.md")); got != traceBefore {
		t.Fatalf("candidate trace must not change\nbefore:\n%s\nafter:\n%s", traceBefore, got)
	}
	if got := readReleaseTestFile(t, traceCheckPath); got != traceCheckBefore {
		t.Fatalf("trace process file must not change\nbefore:\n%s\nafter:\n%s", traceCheckBefore, got)
	}
	if got := readReleaseTestFile(t, stableVerifyPath); got != stableVerifyBefore {
		t.Fatalf("stable verify evidence must not change\nbefore:\n%s\nafter:\n%s", stableVerifyBefore, got)
	}
}

func TestReleaseVersionAllowsControlledChangeStableConsumer(t *testing.T) {
	repoRoot := setupReleaseStableAutoForkRepo(t, "unit_fork")
	writeReleaseStableVerifyProcess(t, repoRoot, "agent", "controlled_change_required")
	writeReleaseStableSharedVersion(t, repoRoot, "0.2.0")

	result, err := ReleaseVersion(repoRoot, ReleaseVersionOptions{
		RuleID:  "shared_demo",
		FromRef: "s_b_rule_demo@0.1.0",
		ToRef:   "s_b_rule_demo@0.2.0",
	})
	if err != nil {
		t.Fatalf("ReleaseVersion: %v", err)
	}
	if len(result.StableForked) != 1 || result.StableForked[0] != "unit:agent" {
		t.Fatalf("expected stable agent fork, got %+v", result.StableForked)
	}
	agentCandidate := readReleaseTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_agent.md"))
	if !strings.Contains(agentCandidate, "candidate_intent: change") ||
		!strings.Contains(agentCandidate, "  - s_b_rule_demo@0.2.0") {
		t.Fatalf("candidate agent was not written as a change fork:\n%s", agentCandidate)
	}
	if _, err := os.Stat(releaseStableVerifyAbsPath(t, repoRoot, "agent")); !os.IsNotExist(err) {
		t.Fatalf("expected stable verify evidence to be cleaned up by unit_fork close, stat err=%v", err)
	}
	status := readReleaseTestFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"))
	if !strings.Contains(status, "| `unit` | `agent` | `yes` | `yes` | `candidate` | `unit_check` | Auto-forked by rule release-version from s_b_rule_demo@0.1.0 to s_b_rule_demo@0.2.0; rerun check. |") {
		t.Fatalf("agent status was not routed through unit_fork close:\n%s", status)
	}
}

func TestReleaseVersionRejectsStableConsumerWhenNextCommandIsNotUnitForkBeforeWriting(t *testing.T) {
	repoRoot := setupReleaseStableAutoForkRepo(t, "unit_stable_verify")
	writeReleaseStableSharedVersion(t, repoRoot, "0.2.0")

	statusBefore := readReleaseTestFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"))
	traceBefore := readReleaseTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_trace.md"))

	result, err := ReleaseVersion(repoRoot, ReleaseVersionOptions{
		RuleID:  "shared_demo",
		FromRef: "s_b_rule_demo@0.1.0",
		ToRef:   "s_b_rule_demo@0.2.0",
	})
	if err == nil || !strings.Contains(err.Error(), "requires current Next Command unit_fork") {
		t.Fatalf("expected unit_fork next-command rejection, got result=%+v err=%v", result, err)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_agent.md")); !os.IsNotExist(err) {
		t.Fatalf("candidate agent must not be written, stat err=%v", err)
	}
	if got := readReleaseTestFile(t, filepath.Join(repoRoot, "docs/specs/_status.md")); got != statusBefore {
		t.Fatalf("status must not change\nbefore:\n%s\nafter:\n%s", statusBefore, got)
	}
	if got := readReleaseTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_trace.md")); got != traceBefore {
		t.Fatalf("candidate trace must not change\nbefore:\n%s\nafter:\n%s", traceBefore, got)
	}
}

func TestReleaseVersionRejectsFromRefFromDifferentRuleWithoutWritingConsumers(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/stable"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/units/candidate"))

	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `trace` | `no` | `yes` | `candidate` | `unit_plan` | candidate |",
	}, "\n")+"\n")
	writeUnitSpecWithRuleRefs(t, repoRoot, "candidate", "trace", []string{"s_b_rule_other@0.1.0"})
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_b_rule_other.md"), `---
rule_id: b_rule_other
rule_scope: bound
layer: stable
rule_version: 0.1.0
---

# Other
`)
	writeStableSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: stable
rule_version: 0.2.0
---

# Shared
`)
	before, err := os.ReadFile(filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_trace.md"))
	if err != nil {
		t.Fatalf("read candidate before release: %v", err)
	}

	result, err := ReleaseVersion(repoRoot, ReleaseVersionOptions{
		RuleID:  "shared_demo",
		FromRef: "s_b_rule_other@0.1.0",
		ToRef:   "s_b_rule_demo@0.2.0",
	})
	if err == nil || !strings.Contains(err.Error(), "from-ref \"s_b_rule_other@0.1.0\" and to-ref \"s_b_rule_demo@0.2.0\" must refer to the same rule file prefix") {
		t.Fatalf("expected from-ref rule mismatch, got result=%+v err=%v", result, err)
	}
	after, err := os.ReadFile(filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_trace.md"))
	if err != nil {
		t.Fatalf("read candidate after release: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("release-version must not rewrite consumers after from-ref rule mismatch\nbefore:\n%s\nafter:\n%s", string(before), string(after))
	}
}

func setupReleaseStableAutoForkRepo(t *testing.T, stableNextCommand string) string {
	t.Helper()
	repoRoot := t.TempDir()
	for _, dir := range []string{
		"docs/specs",
		"docs/specs/rules/stable",
		"docs/specs/units/stable",
		"docs/specs/units/candidate",
		"docs/specs/_check_result/unit",
		"docs/specs/_plans/active",
		"docs/specs/_stable_verify_result/unit",
		"docs/specs/_verify_result/unit",
	} {
		mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(dir)))
	}
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `agent` | `yes` | `no` | `stable` | `" + stableNextCommand + "` | stable |",
		"| `unit` | `trace` | `no` | `yes` | `candidate` | `unit_plan` | candidate |",
	}, "\n")+"\n")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_agent.md"), strings.Join([]string{
		"---",
		"id: agent",
		"layer: stable",
		"version: 0.1.0",
		"rule_refs:",
		"  - s_b_rule_demo@0.1.0",
		"---",
		"",
		"# agent",
		"",
		"## Testability / Acceptance Criteria",
		"",
		"acceptance_item_set:",
		"  - id: agent.acceptance",
		"    target: agent behavior is accepted.",
		"    verification_surface: internal_flow",
		"    implementation_surface: AgentCore/internal/agent",
		"    verification_method: Go test for agent behavior.",
		"    pass_condition: agent behavior passes the declared checks.",
		"    not_runnable_yet: no",
		"",
	}, "\n"))
	writeUnitSpecWithRuleRefs(t, repoRoot, "candidate", "trace", []string{"s_b_rule_demo@0.1.0"})
	writeReleaseStableSharedVersion(t, repoRoot, "0.1.0")
	writeRepositoryMappingFile(t, repoRoot, "0.1.0")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/trace.md"), "check")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/trace.md"), "plan")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit/trace.md"), "verify")
	return repoRoot
}

func writeReleaseStableSharedVersion(t *testing.T, repoRoot, version string) {
	t.Helper()
	writeStableSharedFile(t, repoRoot, strings.Join([]string{
		"---",
		"rule_id: shared_demo",
		"rule_scope: bound",
		"layer: stable",
		"rule_version: " + version,
		"---",
		"",
		"# Shared",
		"",
	}, "\n"))
}

func writeReleaseStableVerifyProcess(t *testing.T, repoRoot, unit, decision string) {
	t.Helper()
	snap, err := snapshot.RebuildCurrentObject(repoRoot, "unit", unit)
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
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
	lines := []string{
		"object_type: unit",
		"object_ref: " + unit,
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
	}
	lines = append(lines, renderReleaseAcceptanceItems(snap.AcceptanceItemSet)...)
	lines = append(lines, renderReleaseAppendixSnapshot(snap.ModuleAppendixSnapshot)...)
	lines = append(lines, renderReleaseUnitSnapshot(snap.UnitSnapshot)...)
	lines = append(lines, renderReleaseRuleSnapshot(snap.RuleSnapshot)...)
	lines = append(lines, renderReleaseAcceptanceEvidence(snap.AcceptanceItemSet)...)
	lines = append(lines,
		"implementation_surface_refs: AgentCore/internal/"+unit,
		"evidence_refs: go test ./...",
		"evaluation_mode: independent",
		"reviewer_result: pass",
		"reviewer_context: minimal_context",
		"review_input_refs: "+rulesyncReviewInputRefsForTest(snap.Object, "unit_stable_verify_advancing", snap.SpecFileRef),
		"review_findings: none",
		"human_decision_refs: none",
	)
	processPath := releaseStableVerifyAbsPath(t, repoRoot, unit)
	mustWriteFile(t, processPath, "# stable verify\n\n```yaml\n"+strings.Join(lines, "\n")+"\n```\n")
}

func renderReleaseAcceptanceItems(entries []snapshot.AcceptanceItemEntry) []string {
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

func renderReleaseAppendixSnapshot(entries []snapshot.AppendixEntry) []string {
	if len(entries) == 0 {
		return []string{"unit_appendix_snapshot: none"}
	}
	lines := []string{"unit_appendix_snapshot:"}
	for _, entry := range entries {
		lines = append(lines,
			"  - file_ref: "+entry.FileRef,
			"    fingerprint: "+entry.Fingerprint,
		)
	}
	return lines
}

func renderReleaseUnitSnapshot(entries []snapshot.ObjectSnapshotEntry) []string {
	if len(entries) == 0 {
		return []string{"unit_snapshot: none"}
	}
	lines := []string{"unit_snapshot:"}
	for _, entry := range entries {
		lines = append(lines,
			"  - unit: "+entry.ObjectRef,
			"    layer: "+entry.Layer,
			"    file_ref: "+entry.FileRef,
			"    version_ref: "+entry.VersionRef,
			"    fingerprint: "+entry.Fingerprint,
		)
	}
	return lines
}

func renderReleaseRuleSnapshot(entries []snapshot.RuleEntry) []string {
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

func renderReleaseAcceptanceEvidence(entries []snapshot.AcceptanceItemEntry) []string {
	if len(entries) == 0 {
		return []string{"acceptance_item_evidence_matrix: none"}
	}
	lines := []string{"acceptance_item_evidence_matrix:"}
	for _, entry := range entries {
		lines = append(lines,
			"  - id: "+entry.ID,
			"    status: pass",
		)
	}
	return lines
}

func releaseStableVerifyAbsPath(t *testing.T, repoRoot, unit string) string {
	t.Helper()
	fileRef, err := snapshot.ProcessFilePath("unit", unit, "stable_verify")
	if err != nil {
		t.Fatalf("stable verify path: %v", err)
	}
	return filepath.Join(repoRoot, filepath.FromSlash(fileRef))
}

func readReleaseTestFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
