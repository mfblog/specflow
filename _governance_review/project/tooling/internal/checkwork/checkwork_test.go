package checkwork

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInitCreatesUnitCheckChecklist(t *testing.T) {
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
	for _, itemID := range []string{
		"goal_and_responsibility",
		"dependency_truth_surface",
		"main_flow_and_state",
		"boundary_and_protocol",
		"data_artifact_and_output",
		"error_edge_and_gap",
		"acceptance_and_testability",
		"implementation_handoff",
	} {
		if !strings.Contains(content, "| "+itemID+" |") {
			t.Fatalf("expected checklist item %s in checklist file:\n%s", itemID, content)
		}
	}
	validation := Validate(repoRoot, "unit", "demo", now)
	if !validation.Valid {
		t.Fatalf("expected valid checklist, diagnostics=%v", validation.Diagnostics)
	}
}

func TestInitSourceRepoUsesLocalFrameworkInputs(t *testing.T) {
	repoRoot := createSourceCheckWorkRepo(t)
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)

	if _, err := Init(repoRoot, "unit", "demo", now); err != nil {
		t.Fatalf("Init source repo: %v", err)
	}
	content := readCheckWorkFile(t, repoRoot)
	if !strings.Contains(content, "framework/core/object_model.md") {
		t.Fatalf("expected source checklist to use local framework inputs:\n%s", content)
	}
	if strings.Contains(content, "specflow/framework/core/object_model.md") {
		t.Fatalf("source checklist must not use installed framework inputs:\n%s", content)
	}
}

func TestValidateRejectsIllegalStatusAndMissingChecklistItem(t *testing.T) {
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
	content = strings.Replace(content, "| goal_and_responsibility | pending |", "| unexpected_extra | pending |", 1)
	mustWriteCheckWorkFile(t, path, content)
	validation = Validate(repoRoot, "unit", "demo", now)
	if validation.Valid || !containsDiagnostic(validation.Diagnostics, "missing checklist item: goal_and_responsibility") {
		t.Fatalf("expected missing checklist diagnostic, got valid=%t diagnostics=%v", validation.Valid, validation.Diagnostics)
	}
}

func TestRefreshMarksClearChecklistItemsStale(t *testing.T) {
	repoRoot := createCheckWorkRepo(t)
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	if _, err := Init(repoRoot, "unit", "demo", now); err != nil {
		t.Fatalf("Init: %v", err)
	}

	path := checkWorkPath(repoRoot)
	content := readCheckWorkFile(t, repoRoot)
	content = strings.Replace(content, "| goal_and_responsibility | pending |", "| goal_and_responsibility | clear |", 1)
	mustWriteCheckWorkFile(t, path, content)

	candidatePath := filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md")
	mustWriteCheckWorkFile(t, candidatePath, demoCandidate("changed body"))

	result, err := Refresh(repoRoot, "unit", "demo", now.Add(time.Hour))
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if !containsString(result.StaleItems, "goal_and_responsibility") {
		t.Fatalf("expected goal checklist item stale, got %+v", result)
	}
	refreshed := readCheckWorkFile(t, repoRoot)
	if !strings.Contains(refreshed, "| goal_and_responsibility | stale |") {
		t.Fatalf("expected stale goal checklist item in file:\n%s", refreshed)
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
		t.Fatalf("expected new checklist after closed state, got %+v", result)
	}
	if len(result.DeletedFiles) != 1 || result.DeletedFiles[0].Reason != "closed_work_state" {
		t.Fatalf("expected closed checklist deletion, got %+v", result.DeletedFiles)
	}
}

func createCheckWorkRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	for _, dir := range []string{
		"docs/specs/units/candidate",
		"docs/specs",
		"specflow/framework/core",
		"specflow/framework/lifecycle",
		"specflow/framework/operations",
		"specflow/framework",
		"specflow/tooling",
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
		"specflow/framework/core/object_model.md",
		"specflow/framework/core/repository_mapping.md",
		"specflow/framework/lifecycle/unit_check.md",
		"specflow/framework/lifecycle/unit_verify.md",
		"specflow/framework/operations/implementation_change.md",
		"specflow/framework/process_snapshot_contract.md",
		"specflow/framework/candidate_handoff_contract.md",
		"specflow/framework/candidate_intent_policy.md",
	} {
		mustWriteCheckWorkFile(t, filepath.Join(repoRoot, filepath.FromSlash(relPath)), "# "+relPath+"\n")
	}
	mustWriteCheckWorkFile(t, filepath.Join(repoRoot, "specflow/tooling/manifest.tsv"), "templates/AGENTS.md\tAGENTS.md\tframework\n")
	return repoRoot
}

func createSourceCheckWorkRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	for _, dir := range []string{
		"docs/specs/units/candidate",
		"docs/specs",
		"framework/core",
		"framework/lifecycle",
		"framework/operations",
		"framework",
		"tooling",
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
		"framework/core/object_model.md",
		"framework/core/repository_mapping.md",
		"framework/lifecycle/unit_check.md",
		"framework/lifecycle/unit_verify.md",
		"framework/operations/implementation_change.md",
		"framework/process_snapshot_contract.md",
		"framework/candidate_handoff_contract.md",
		"framework/candidate_intent_policy.md",
	} {
		mustWriteCheckWorkFile(t, filepath.Join(repoRoot, filepath.FromSlash(relPath)), "# "+relPath+"\n")
	}
	mustWriteCheckWorkFile(t, filepath.Join(repoRoot, "tooling/manifest.tsv"), "templates/AGENTS.md\tAGENTS.md\tframework\n")
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
