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
