package reader

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSnapshotConnectsUnitSpecAndSharedContract(t *testing.T) {
	repoRoot := createReaderRepo(t)

	snapshot := BuildSnapshot(repoRoot)

	unit := findObject(t, snapshot.Objects, "unit", "assistant")
	if unit.HumanState != "正在确认的设计" {
		t.Fatalf("expected human candidate state, got %q", unit.HumanState)
	}
	if unit.NextLabel != "检查设计是否足够支撑开发" {
		t.Fatalf("expected translated next command, got %q", unit.NextLabel)
	}
	if len(unit.TruthPaths) != 1 || unit.TruthPaths[0].Path != "docs/specs/units/candidate/c_unit_assistant.md" {
		t.Fatalf("unexpected truth paths: %+v", unit.TruthPaths)
	}
	if len(unit.SharedRefs) != 1 || unit.SharedRefs[0] != "shared_runtime_model" {
		t.Fatalf("unexpected shared refs: %+v", unit.SharedRefs)
	}

	shared := findObject(t, snapshot.Objects, "shared_contract", "shared_runtime_model")
	if len(shared.BoundObjects) != 1 || shared.BoundObjects[0] != "unit:assistant" {
		t.Fatalf("unexpected bound objects: %+v", shared.BoundObjects)
	}
	if !hasEdge(snapshot.Edges, "unit:assistant", "file:docs/specs/units/candidate/c_unit_assistant.md", "described_by") {
		t.Fatalf("expected unit described_by edge, got %+v", snapshot.Edges)
	}
	if !hasEdge(snapshot.Edges, "unit:assistant", "shared:shared_runtime_model", "uses_shared") {
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
		"### 2.3 Current Shared Contracts",
		"",
		"1. `runtime_model`",
		"   - runtime model rule",
		"",
		"### 4.5 Shared Contract Truth Paths",
		"",
		"1. `runtime_model`",
		"   - `docs/specs/shared_contracts/candidate/c_shared_runtime_model.md`",
		"",
		"### 4.6 Unit Truth And Implementation Paths",
		"",
		"1. `assistant`",
		"   - `truth_surface`",
		"     - `docs/specs/units/candidate/c_unit_assistant.md`",
		"   - `implementation_surface`",
		"     - `CLI/internal/assistant/**`",
	}, "\n")+"\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/system_constraints.md"), "# System Constraints\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_assistant.md"), strings.Join([]string{
		"---",
		"id: assistant",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Assistant",
		"",
		"1. `shared_contract_refs`: `c_shared_runtime_model@0.1.0`",
	}, "\n")+"\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/shared_contracts/candidate/c_shared_runtime_model.md"), strings.Join([]string{
		"---",
		"shared_contract_id: shared_runtime_model",
		"layer: candidate",
		"shared_version: 0.1.0",
		"bound_objects:",
		"  - unit:assistant",
		"---",
		"",
		"# Shared Runtime Model",
	}, "\n")+"\n")
	return repoRoot
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
