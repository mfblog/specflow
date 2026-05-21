package checkwork

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInitCreatesUnitWorkStateWithBaselineSkeleton(t *testing.T) {
	repoRoot := createCheckWorkRepo(t)
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)

	result, err := Init(repoRoot, "unit", "demo", now)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if !result.Created {
		t.Fatalf("expected created result: %+v", result)
	}
	content := readCheckWorkFile(t, repoRoot)
	for _, sliceID := range []string{
		"goal_and_responsibility",
		"dependency_truth_surface",
		"main_flow_and_state",
		"boundary_and_protocol",
		"data_artifact_and_output",
		"error_edge_and_gap",
		"acceptance_and_testability",
		"implementation_handoff",
		"goal_to_acceptance_convergence",
		"flow_to_boundary_convergence",
		"dependency_truth_convergence",
		"output_to_acceptance_convergence",
	} {
		if !strings.Contains(content, "| "+sliceID+" |") {
			t.Fatalf("expected baseline slice %s in work state:\n%s", sliceID, content)
		}
	}
	validation := Validate(repoRoot, "unit", "demo", now)
	if !validation.Valid {
		t.Fatalf("expected valid work state, diagnostics=%v", validation.Diagnostics)
	}
}

func TestValidateRejectsIllegalStatusAndMissingDynamicParent(t *testing.T) {
	repoRoot := createCheckWorkRepo(t)
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	if _, err := Init(repoRoot, "unit", "demo", now); err != nil {
		t.Fatalf("Init: %v", err)
	}

	path := checkWorkPath(repoRoot)
	content := readCheckWorkFile(t, repoRoot)
	content = strings.Replace(content, "| status | in_progress |", "| status | not_a_status |", 1)
	mustWriteCheckWorkFile(t, path, content)
	validation := Validate(repoRoot, "unit", "demo", now)
	if validation.Valid || !containsDiagnostic(validation.Diagnostics, "invalid work status") {
		t.Fatalf("expected invalid status diagnostic, got valid=%t diagnostics=%v", validation.Valid, validation.Diagnostics)
	}

	content = strings.Replace(content, "| status | not_a_status |", "| status | in_progress |", 1)
	dynamicTable := "| " + strings.Join(sliceColumns, " | ") + " |\n" +
		"|" + strings.Repeat("---|", len(sliceColumns)) + "\n" +
		"| dynamic_gap | dynamic | local | pending | Check uncovered gap. | new owner conflict | missing_parent | docs/specs/units/candidate/c_unit_demo.md | abc | none | none | pending | agent records a result | review slice dynamic_gap |\n"
	content = strings.Replace(content, "## Dynamic Slices\n\nnone\n", "## Dynamic Slices\n\n"+dynamicTable, 1)
	mustWriteCheckWorkFile(t, path, content)
	validation = Validate(repoRoot, "unit", "demo", now)
	if validation.Valid || !containsDiagnostic(validation.Diagnostics, "dynamic slice parent_slice_id must reference an existing slice") {
		t.Fatalf("expected dynamic parent diagnostic, got valid=%t diagnostics=%v", validation.Valid, validation.Diagnostics)
	}
}

func TestRefreshMarksPassedSlicesStale(t *testing.T) {
	repoRoot := createCheckWorkRepo(t)
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	if _, err := Init(repoRoot, "unit", "demo", now); err != nil {
		t.Fatalf("Init: %v", err)
	}

	path := checkWorkPath(repoRoot)
	content := readCheckWorkFile(t, repoRoot)
	content = strings.Replace(content, "| goal_and_responsibility | baseline | local | pending |", "| goal_and_responsibility | baseline | local | passed |", 1)
	content = strings.Replace(content, "| goal_to_acceptance_convergence | baseline | cross_convergence | pending |", "| goal_to_acceptance_convergence | baseline | cross_convergence | passed |", 1)
	mustWriteCheckWorkFile(t, path, content)

	candidatePath := filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md")
	mustWriteCheckWorkFile(t, candidatePath, demoCandidate("changed body"))

	result, err := Refresh(repoRoot, "unit", "demo", now.Add(time.Hour))
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if !containsString(result.StaleSlices, "goal_and_responsibility") {
		t.Fatalf("expected goal slice stale, got %+v", result)
	}
	if !containsString(result.StaleSlices, "goal_to_acceptance_convergence") {
		t.Fatalf("expected cross slice stale, got %+v", result)
	}
	refreshed := readCheckWorkFile(t, repoRoot)
	if !strings.Contains(refreshed, "| goal_and_responsibility | baseline | local | stale |") {
		t.Fatalf("expected stale goal slice in file:\n%s", refreshed)
	}
}

func TestInitDoesNotReuseClosedWorkState(t *testing.T) {
	repoRoot := createCheckWorkRepo(t)
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	if _, err := Init(repoRoot, "unit", "demo", now); err != nil {
		t.Fatalf("Init: %v", err)
	}

	path := checkWorkPath(repoRoot)
	content := readCheckWorkFile(t, repoRoot)
	content = strings.Replace(content, "| status | in_progress |", "| status | closed_pass |", 1)
	mustWriteCheckWorkFile(t, path, content)

	result, err := Init(repoRoot, "unit", "demo", now.Add(time.Hour))
	if err != nil {
		t.Fatalf("Init closed: %v", err)
	}
	if !result.Created || result.Reused {
		t.Fatalf("expected new work state after closed state, got %+v", result)
	}
	if len(result.DeletedFiles) != 1 || result.DeletedFiles[0].Reason != "closed_work_state" {
		t.Fatalf("expected closed work-state deletion, got %+v", result.DeletedFiles)
	}
}

func createCheckWorkRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	for _, dir := range []string{
		"docs/specs/units/candidate",
		"docs/specs",
		"specflow/framework/commands",
		"specflow/framework",
	} {
		if err := os.MkdirAll(filepath.Join(repoRoot, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	mustWriteCheckWorkFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | note |",
	}, "\n")+"\n")
	mustWriteCheckWorkFile(t, filepath.Join(repoRoot, "docs/specs/repository_mapping.md"), "---\nid: repository_mapping\nversion: 0.1.0\n---\n# Repository Mapping\n")
	mustWriteCheckWorkFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), demoCandidate("initial body"))
	for _, relPath := range []string{
		"specflow/framework/commands/unit_check.md",
		"specflow/framework/commands/unit_plan.md",
		"specflow/framework/process_snapshot_contract.md",
		"specflow/framework/slice_work_state_protocol.md",
		"specflow/framework/candidate_handoff_contract.md",
		"specflow/framework/spec_writing_guide.md",
		"specflow/framework/candidate_intent_policy.md",
	} {
		mustWriteCheckWorkFile(t, filepath.Join(repoRoot, filepath.FromSlash(relPath)), "# "+relPath+"\n")
	}
	return repoRoot
}

func demoCandidate(body string) string {
	return strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"candidate_intent: change",
		"source_basis: new_design",
		"evidence_appendix_ref: none",
		"---",
		"",
		"# Demo",
		"",
		body,
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs: none",
	}, "\n") + "\n"
}

func checkWorkPath(repoRoot string) string {
	return filepath.Join(repoRoot, "docs/specs/_check_work/unit/demo.md")
}

func readCheckWorkFile(t *testing.T, repoRoot string) string {
	t.Helper()
	content, err := os.ReadFile(checkWorkPath(repoRoot))
	if err != nil {
		t.Fatalf("read check work: %v", err)
	}
	return string(content)
}

func mustWriteCheckWorkFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func containsDiagnostic(diagnostics []string, want string) bool {
	for _, diagnostic := range diagnostics {
		if strings.Contains(diagnostic, want) {
			return true
		}
	}
	return false
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
