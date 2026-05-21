package reader

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSnapshotConnectsUnitSpecAndRule(t *testing.T) {
	repoRoot := createReaderRepo(t)

	snapshot := BuildSnapshot(repoRoot)

	unit := findObject(t, snapshot.Objects, "unit", "assistant")
	if unit.HumanState != "正在确认的设计" {
		t.Fatalf("expected human candidate state, got %q", unit.HumanState)
	}
	if unit.NextLabel != "检查设计是否足够支撑开发" {
		t.Fatalf("expected translated next command, got %q", unit.NextLabel)
	}
	repairUnit := findObject(t, snapshot.Objects, "unit", "tool")
	if repairUnit.NextIntent != "repair" || repairUnit.NextIntentLabel != "修复基线" {
		t.Fatalf("expected repair next intent, got intent=%q label=%q", repairUnit.NextIntent, repairUnit.NextIntentLabel)
	}
	repairCandidate := findObject(t, snapshot.Objects, "unit", "memory")
	if repairCandidate.NextIntent != "repair" || repairCandidate.NextIntentLabel != "修复基线" {
		t.Fatalf("expected candidate repair intent, got intent=%q label=%q", repairCandidate.NextIntent, repairCandidate.NextIntentLabel)
	}
	expectedTruthPaths := []string{
		"docs/specs/units/candidate/c_unit_assistant.md",
		"docs/specs/units/candidate/appendix/c_unit_assistant_evidence.md",
		"docs/specs/units/candidate/appendix/c_unit_assistant_prompt.md",
	}
	if !sourcePathsEqual(unit.TruthPaths, expectedTruthPaths) {
		t.Fatalf("unexpected truth paths: %+v", unit.TruthPaths)
	}
	if !stringSlicesEqual(unit.RuleRefs, []string{"b_rule_runtime_model", "b_rule_unregistered"}) {
		t.Fatalf("unexpected rule refs: %+v", unit.RuleRefs)
	}
	if !stringSlicesEqual(unit.UnitRefs, []string{"tool"}) {
		t.Fatalf("unexpected unit refs: %+v", unit.UnitRefs)
	}

	shared := findObject(t, snapshot.Objects, "rule", "b_rule_runtime_model")
	if countObjects(snapshot.Objects, "rule", "b_rule_runtime_model") != 1 {
		t.Fatalf("expected one runtime rule object, got %+v", snapshot.Objects)
	}
	if countObjects(snapshot.Objects, "rule", "shared_runtime_model") != 0 {
		t.Fatalf("mapping shorthand must merge into the formal rule object, got %+v", snapshot.Objects)
	}
	if len(shared.BoundObjects) != 1 || shared.BoundObjects[0] != "unit:assistant" {
		t.Fatalf("unexpected bound objects: %+v", shared.BoundObjects)
	}
	if !sourcePathsEqual(shared.TruthPaths, []string{"docs/specs/rules/candidate/c_b_rule_runtime_model.md"}) {
		t.Fatalf("unexpected shared truth paths: %+v", shared.TruthPaths)
	}
	if !hasEdge(snapshot.Edges, "unit:assistant", "file:docs/specs/units/candidate/c_unit_assistant.md", "described_by") {
		t.Fatalf("expected unit described_by edge, got %+v", snapshot.Edges)
	}
	if !hasEdge(snapshot.Edges, "unit:assistant", "shared:b_rule_runtime_model", "uses_shared") {
		t.Fatalf("expected unit uses_shared edge, got %+v", snapshot.Edges)
	}
	registryUnit := findRegistryItem(t, snapshot.Registry, "unit", "assistant")
	if !registryUnit.MappingRegistered || !registryUnit.StatusRegistered || !registryUnit.TruthRegistered || !registryUnit.ImplementationRegistered {
		t.Fatalf("expected assistant registry to be complete, got %+v", registryUnit)
	}
	if registryUnit.Result != "landed" || len(registryUnit.Issues) != 0 {
		t.Fatalf("expected assistant registry landed result, got %+v", registryUnit)
	}
	if len(registryUnit.UnitRefs) != 1 || registryUnit.UnitRefs[0] != "tool" {
		t.Fatalf("unexpected assistant unit refs: %+v", registryUnit.UnitRefs)
	}
	registryRule := findRegistryItem(t, snapshot.Registry, "rule", "b_rule_runtime_model")
	if !registryRule.MappingRegistered || !registryRule.TruthRegistered || registryRule.StatusRegistered {
		t.Fatalf("expected runtime rule registry without status registration, got %+v", registryRule)
	}
	if registryRule.RuleScope != "bound" || len(registryRule.BoundObjects) != 1 {
		t.Fatalf("expected runtime rule to be bound by current specs, got %+v", registryRule)
	}
	registryGlobalRule := findRegistryItem(t, snapshot.Registry, "rule", "g_rule_repository_baseline")
	if !registryGlobalRule.MappingRegistered || !registryGlobalRule.TruthRegistered || registryGlobalRule.StatusRegistered {
		t.Fatalf("expected global rule registry without status registration, got %+v", registryGlobalRule)
	}
	if registryGlobalRule.RuleScope != "global" || registryGlobalRule.Result != "planned" {
		t.Fatalf("expected global rule to be planned global registry item without direct code path, got %+v", registryGlobalRule)
	}
	registryPlannedRule := findRegistryItem(t, snapshot.Registry, "rule", "b_rule_future")
	if registryPlannedRule.Result != "planned" || registryPlannedRule.TruthRegistered {
		t.Fatalf("expected planned future rule without Spec file, got %+v", registryPlannedRule)
	}
	registryUnregisteredRule := findRegistryItem(t, snapshot.Registry, "rule", "b_rule_unregistered")
	if registryUnregisteredRule.Result != "unregistered_file" {
		t.Fatalf("expected unregistered rule file, got %+v", registryUnregisteredRule)
	}
	truthNode := findNode(t, snapshot.Nodes, "file:docs/specs/units/candidate/c_unit_assistant.md")
	if truthNode.Label != "assistant (candidate)" {
		t.Fatalf("expected candidate layer in truth node label, got %q", truthNode.Label)
	}
	if snapshot.CandidateRelations.RelationResult != "pass" {
		t.Fatalf("expected candidate relation snapshot to compute, got %+v", snapshot.CandidateRelations)
	}
	if !stringSlicesEqual(snapshot.CandidateRelations.ReadyCandidates, []string{"memory"}) {
		t.Fatalf("unexpected ready candidate relations: %+v", snapshot.CandidateRelations.ReadyCandidates)
	}
	if len(snapshot.CandidateRelations.BlockedCandidates) != 1 || snapshot.CandidateRelations.BlockedCandidates[0].Object != "assistant" {
		t.Fatalf("expected assistant to be blocked by candidate rule refs in fixture, got %+v", snapshot.CandidateRelations.BlockedCandidates)
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("Marshal(snapshot) returned error: %v", err)
	}
	text := string(data)
	for _, expected := range []string{
		`"diagnostics":[]`,
		`"implementation_paths":[]`,
		`"bound_objects":[]`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected JSON to contain %s, got %s", expected, text)
		}
	}
}

func TestReadAllowedSourceRejectsPathEscape(t *testing.T) {
	repoRoot := createReaderRepo(t)

	_, err := ReadAllowedSource(repoRoot, "../AGENTS.md")
	if err == nil || !strings.Contains(err.Error(), "escapes repo root") {
		t.Fatalf("expected path escape error, got %v", err)
	}

	source, err := ReadAllowedSource(repoRoot, "docs/specs/_status.md")
	if err != nil {
		t.Fatalf("ReadAllowedSource returned error: %v", err)
	}
	if source.Path != "docs/specs/_status.md" || !strings.Contains(source.Content, "assistant") {
		t.Fatalf("unexpected source: %+v", source)
	}
}

func findObject(t *testing.T, objects []ObjectView, kind, id string) ObjectView {
	t.Helper()
	for _, object := range objects {
		if object.Kind == kind && object.ID == id {
			return object
		}
	}
	t.Fatalf("object %s/%s not found in %+v", kind, id, objects)
	return ObjectView{}
}

func findNode(t *testing.T, nodes []GraphNode, id string) GraphNode {
	t.Helper()
	for _, node := range nodes {
		if node.ID == id {
			return node
		}
	}
	t.Fatalf("node %s not found in %+v", id, nodes)
	return GraphNode{}
}

func findRegistryItem(t *testing.T, items []RegistryItem, kind, id string) RegistryItem {
	t.Helper()
	for _, item := range items {
		if item.Kind == kind && item.ID == id {
			return item
		}
	}
	t.Fatalf("registry item %s/%s not found in %+v", kind, id, items)
	return RegistryItem{}
}

func countObjects(objects []ObjectView, kind, id string) int {
	count := 0
	for _, object := range objects {
		if object.Kind == kind && object.ID == id {
			count++
		}
	}
	return count
}

func countRegistryItems(items []RegistryItem, kind, id string) int {
	count := 0
	for _, item := range items {
		if item.Kind == kind && item.ID == id {
			count++
		}
	}
	return count
}

func hasEdge(edges []GraphEdge, from, to, kind string) bool {
	for _, edge := range edges {
		if edge.From == from && edge.To == to && edge.Kind == kind {
			return true
		}
	}
	return false
}

func createReaderRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `assistant` | `no` | `yes` | `candidate` | `unit_check` | note |",
		"| `unit` | `memory` | `yes` | `yes` | `candidate` | `unit_plan` | repair candidate in progress |",
		"| `unit` | `tool` | `yes` | `no` | `stable` | `unit_fork` | next fork must create candidate_intent=repair |",
	}, "\n")+"\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/repository_mapping.md"), strings.Join([]string{
		"# Repository Mapping",
		"",
		"## 2. Object Registry",
		"",
		"| kind | id | registration_state | implementation_paths | spec_files | responsibility |",
		"|---|---|---|---|---|---|",
		"| unit | assistant | landed | `CLI/internal/assistant/**` | `docs/specs/units/candidate/c_unit_assistant.md` | assistant prompt responsibility |",
		"| unit | tool | planned | none | `docs/specs/units/stable/s_unit_tool.md` | tool execution responsibility |",
		"| unit | memory | planned | none | `docs/specs/units/candidate/c_unit_memory.md` | memory responsibility |",
		"| rule | g_rule_repository_baseline | planned | none | `docs/specs/rules/stable/s_g_rule_repository_baseline.md` | global baseline |",
		"| rule | b_rule_runtime_model | planned | none | `docs/specs/rules/candidate/c_b_rule_runtime_model.md` | runtime model rule |",
		"| rule | b_rule_future | planned | none | none | future shared rule |",
		"",
		"### 4.6 Unit Truth Rules And Implementation Paths",
		"",
		"1. `assistant`",
		"   - `truth_surface_rule`: `unit_default`",
		"   - `implementation_surface`",
		"     - `CLI/internal/assistant/**`",
		"",
		"## 5. Rule Alignment",
		"",
		"1. `s_g_rule_repository_baseline@0.1.0`",
		"2. `c_b_rule_runtime_model@0.1.0`",
		"",
		"## 6. Drift Handling",
	}, "\n")+"\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_g_rule_repository_baseline.md"), strings.Join([]string{
		"---",
		"rule_id: g_rule_repository_baseline",
		"rule_scope: global",
		"layer: stable",
		"rule_version: 0.1.0",
		"---",
		"",
		"# Global Rules",
	}, "\n")+"\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "CLI/internal/assistant/prompt.go"), "package assistant\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_assistant.md"), strings.Join([]string{
		"---",
		"id: assistant",
		"layer: candidate",
		"version: 0.1.0",
		"evidence_appendix_ref: docs/specs/units/candidate/appendix/c_unit_assistant_evidence.md",
		"unit_refs:",
		"  - s_unit_tool@0.1.0",
		"rule_refs:",
		"  - c_b_rule_runtime_model@0.1.0",
		"  - c_b_rule_unregistered@0.1.0",
		"---",
		"",
		"# Assistant",
		"",
		"1. `rule_refs`:",
		"   - `c_b_rule_runtime_model@0.1.0`",
		"   - `c_b_rule_unregistered@0.1.0`",
		"2. Prompt details live in [`c_unit_assistant_prompt.md`](./appendix/c_unit_assistant_prompt.md).",
	}, "\n")+"\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_tool.md"), strings.Join([]string{
		"---",
		"id: tool",
		"layer: stable",
		"version: 0.1.0",
		"---",
		"",
		"# Tool",
	}, "\n")+"\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_memory.md"), strings.Join([]string{
		"---",
		"id: memory",
		"layer: candidate",
		"version: 0.1.1",
		"candidate_intent: repair",
		"repair_basis: s_unit_memory@0.1.0",
		"source_basis: new_design",
		"evidence_appendix_ref: none",
		"---",
		"",
		"# Memory",
	}, "\n")+"\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/appendix/c_unit_assistant_evidence.md"), strings.Join([]string{
		"# Assistant Evidence",
		"",
		"Evidence notes.",
	}, "\n")+"\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/appendix/c_unit_assistant_prompt.md"), strings.Join([]string{
		"# Assistant Prompt",
		"",
		"Prompt notes.",
	}, "\n")+"\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_runtime_model.md"), strings.Join([]string{
		"---",
		"rule_id: b_rule_runtime_model",
		"rule_scope: bound",
		"layer: candidate",
		"rule_version: 0.1.0",
		"bound_objects:",
		"  - unit:assistant",
		"---",
		"",
		"# Shared Runtime Model",
	}, "\n")+"\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_unregistered.md"), strings.Join([]string{
		"---",
		"rule_id: b_rule_unregistered",
		"rule_scope: bound",
		"layer: candidate",
		"rule_version: 0.1.0",
		"---",
		"",
		"# Unregistered Rule",
	}, "\n")+"\n")
	return repoRoot
}

func sourcePathsEqual(refs []SourceRef, expected []string) bool {
	if len(refs) != len(expected) {
		return false
	}
	for idx, expectedPath := range expected {
		if refs[idx].Path != expectedPath {
			return false
		}
	}
	return true
}

func stringSlicesEqual(values []string, expected []string) bool {
	if len(values) != len(expected) {
		return false
	}
	for idx, expectedValue := range expected {
		if values[idx] != expectedValue {
			return false
		}
	}
	return true
}

func writeReaderTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) failed: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) failed: %v", path, err)
	}
}
