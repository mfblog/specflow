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
	expectedTruthPaths := []string{
		"docs/specs/units/candidate/c_unit_assistant.md",
		"docs/specs/units/candidate/appendix/c_unit_assistant_evidence.md",
		"docs/specs/units/candidate/appendix/c_unit_assistant_prompt.md",
	}
	if !sourcePathsEqual(unit.TruthPaths, expectedTruthPaths) {
		t.Fatalf("unexpected truth paths: %+v", unit.TruthPaths)
	}
	if len(unit.RuleRefs) != 1 || unit.RuleRefs[0] != "b_rule_runtime_model" {
		t.Fatalf("unexpected rule refs: %+v", unit.RuleRefs)
	}

	shared := findObject(t, snapshot.Objects, "rule", "b_rule_runtime_model")
	if len(shared.BoundObjects) != 1 || shared.BoundObjects[0] != "unit:assistant" {
		t.Fatalf("unexpected bound objects: %+v", shared.BoundObjects)
	}
	if !hasEdge(snapshot.Edges, "unit:assistant", "file:docs/specs/units/candidate/c_unit_assistant.md", "described_by") {
		t.Fatalf("expected unit described_by edge, got %+v", snapshot.Edges)
	}
	if !hasEdge(snapshot.Edges, "unit:assistant", "shared:b_rule_runtime_model", "uses_shared") {
		t.Fatalf("expected unit uses_shared edge, got %+v", snapshot.Edges)
	}
	truthNode := findNode(t, snapshot.Nodes, "file:docs/specs/units/candidate/c_unit_assistant.md")
	if truthNode.Label != "assistant (candidate)" {
		t.Fatalf("expected candidate layer in truth node label, got %q", truthNode.Label)
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
	}, "\n")+"\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/repository_mapping.md"), strings.Join([]string{
		"# Repository Mapping",
		"",
		"### 2.1 Current Units",
		"",
		"1. `assistant`",
		"   - assistant prompt responsibility",
		"",
		"### 2.3 Current Rules",
		"",
		"1. `runtime_model`",
		"   - runtime model rule",
		"",
		"### 4.5 Rule Truth Paths",
		"",
		"1. `runtime_model`",
		"   - `docs/specs/rules/candidate/c_b_rule_runtime_model.md`",
		"",
		"### 4.6 Unit Truth Rules And Implementation Paths",
		"",
		"1. `assistant`",
		"   - `truth_surface_rule`: `unit_default`",
		"   - `implementation_surface`",
		"     - `CLI/internal/assistant/**`",
	}, "\n")+"\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_g_rule_repository_baseline.md"), "# Global Rules\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_assistant.md"), strings.Join([]string{
		"---",
		"id: assistant",
		"layer: candidate",
		"version: 0.1.0",
		"evidence_appendix_ref: docs/specs/units/candidate/appendix/c_unit_assistant_evidence.md",
		"---",
		"",
		"# Assistant",
		"",
		"1. `rule_refs`:",
		"   - `c_b_rule_runtime_model@0.1.0`",
		"2. Prompt details live in [`c_unit_assistant_prompt.md`](./appendix/c_unit_assistant_prompt.md).",
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

func writeReaderTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) failed: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) failed: %v", path, err)
	}
}
