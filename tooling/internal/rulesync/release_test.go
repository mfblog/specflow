package rulesync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/units/candidate"))
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

func TestReleaseVersionAutoForksStableScenarioWithoutCandidateIntent(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/stable"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/scenarios/stable"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/scenarios/candidate"))

	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `scenario` | `checkout` | `yes` | `no` | `stable` | `scenario_fork` | stable |",
	}, "\n")+"\n")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/scenarios/stable/s_scenario_checkout.md"), strings.Join([]string{
		"---",
		"id: checkout",
		"layer: stable",
		"version: 0.2.0",
		"rule_refs:",
		"  - s_b_rule_demo@0.1.0",
		"---",
		"",
		"# checkout",
		"",
		"repository_mapping_ref: repository_mapping@0.1.0",
		"unit_refs: none",
		"",
		"## Testability / Acceptance Criteria",
		"",
		"acceptance_item_set:",
		"  - id: checkout.e2e",
		"    target: Checkout chain reaches the declared result.",
		"    verification_surface: integration",
		"    implementation_surface: AgentCore/runtime",
		"    verification_method: Scenario verification for checkout.",
		"    pass_condition: The checkout scenario reaches the declared result.",
		"    not_runnable_yet: no",
		"",
	}, "\n"))
	writeStableSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: stable
rule_version: 0.2.0
---

# Shared
`)

	result, err := ReleaseVersion(repoRoot, ReleaseVersionOptions{
		RuleID:  "shared_demo",
		FromRef: "s_b_rule_demo@0.1.0",
		ToRef:   "s_b_rule_demo@0.2.0",
	})
	if err != nil {
		t.Fatalf("ReleaseVersion: %v", err)
	}
	if len(result.StableForked) != 1 || result.StableForked[0] != "scenario:checkout" {
		t.Fatalf("expected stable checkout scenario fork, got %+v", result.StableForked)
	}

	scenarioCandidate, err := os.ReadFile(filepath.Join(repoRoot, "docs/specs/scenarios/candidate/c_scenario_checkout.md"))
	if err != nil {
		t.Fatalf("read candidate scenario: %v", err)
	}
	candidateText := string(scenarioCandidate)
	if !strings.Contains(candidateText, "layer: candidate") ||
		!strings.Contains(candidateText, "version: 0.2.1") ||
		!strings.Contains(candidateText, "source_basis: new_design") ||
		!strings.Contains(candidateText, "evidence_appendix_ref: none") ||
		!strings.Contains(candidateText, "  - s_b_rule_demo@0.2.0") {
		t.Fatalf("candidate scenario was not correctly forked:\n%s", candidateText)
	}
	if strings.Contains(candidateText, "candidate_intent:") || strings.Contains(candidateText, "repair_basis:") {
		t.Fatalf("scenario candidate must not receive unit candidate intent fields:\n%s", candidateText)
	}

	statusData, err := os.ReadFile(filepath.Join(repoRoot, "docs/specs/_status.md"))
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	if !strings.Contains(string(statusData), "| `scenario` | `checkout` | `yes` | `yes` | `candidate` | `scenario_check` |") {
		t.Fatalf("expected scenario to fall back to scenario_check after auto-fork:\n%s", string(statusData))
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
