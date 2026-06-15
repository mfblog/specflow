package statusfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateNextCommand(t *testing.T) {
	repoRoot := t.TempDir()
	statusPath := filepath.Join(repoRoot, "docs/specs")
	if err := os.MkdirAll(statusPath, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	content := strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `ai` | `yes` | `yes` | `candidate` | `unit_check` | note |",
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(statusPath, "_status.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write status: %v", err)
	}

	updated, err := UpdateNextCommand(repoRoot, "ai", "unit_verify")
	if err != nil {
		t.Fatalf("UpdateNextCommand: %v", err)
	}
	if !updated {
		t.Fatalf("expected update to be true")
	}

	data, err := os.ReadFile(filepath.Join(statusPath, "_status.md"))
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	if !strings.Contains(string(data), "| `unit` | `ai` | `yes` | `yes` | `candidate` | `unit_verify` | note |") {
		t.Fatalf("updated status row not found:\n%s", string(data))
	}
}

func TestUpsertModuleStatusCreatesNewRow(t *testing.T) {
	repoRoot := t.TempDir()
	statusPath := filepath.Join(repoRoot, "docs/specs")
	if err := os.MkdirAll(statusPath, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	content := strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `ai` | `yes` | `no` | `stable` | `unit_fork` | stable note |",
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(statusPath, "_status.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write status: %v", err)
	}

	updated, err := UpsertModuleStatus(repoRoot, ModuleStatus{
		Module:      "unit_new",
		Stable:      "no",
		Candidate:   "yes",
		ActiveLayer: "candidate",
		NextCommand: "unit_check",
		Notes:       "new note",
	}, true)
	if err != nil {
		t.Fatalf("UpsertModuleStatus: %v", err)
	}
	if !updated {
		t.Fatalf("expected update to be true")
	}

	data, err := os.ReadFile(filepath.Join(statusPath, "_status.md"))
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	if !strings.Contains(string(data), "| `unit` | `unit_new` | `no` | `yes` | `candidate` | `unit_check` | new note |") {
		t.Fatalf("created status row not found:\n%s", string(data))
	}
}

func TestUpsertModuleStatusRejectsUnsupportedNextCommand(t *testing.T) {
	repoRoot := t.TempDir()
	statusPath := filepath.Join(repoRoot, "docs/specs")
	if err := os.MkdirAll(statusPath, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	content := strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `ai` | `yes` | `no` | `stable` | `unit_fork` | stable note |",
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(statusPath, "_status.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write status: %v", err)
	}

	_, err := UpsertModuleStatus(repoRoot, ModuleStatus{
		Module:      "ai",
		Stable:      "yes",
		Candidate:   "no",
		ActiveLayer: "stable",
		NextCommand: "typo_command",
		Notes:       "stable note",
	}, false)
	if err == nil || !strings.Contains(err.Error(), "not a supported status value") {
		t.Fatalf("expected unsupported next command error, got %v", err)
	}
}

func TestUpsertObjectStatusRejectsUnsupportedObjectType(t *testing.T) {
	repoRoot := t.TempDir()
	statusPath := filepath.Join(repoRoot, "docs/specs")
	if err := os.MkdirAll(statusPath, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	content := strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(statusPath, "_status.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write status: %v", err)
	}

	_, err := UpsertObjectStatus(repoRoot, ObjectStatus{
		ObjectType:  "flow",
		Object:      "demo",
		Stable:      "no",
		Candidate:   "yes",
		ActiveLayer: "candidate",
		NextCommand: "scenario_check",
	}, true)
	if err == nil || !strings.Contains(err.Error(), "object type \"flow\" is not a supported status value") {
		t.Fatalf("expected unsupported object type error, got %v", err)
	}
}

func TestUpdateObjectNextCommandRejectsScenario(t *testing.T) {
	repoRoot := t.TempDir()
	statusPath := filepath.Join(repoRoot, "docs/specs")
	if err := os.MkdirAll(statusPath, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	content := strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `scenario` | `demo` | `no` | `yes` | `candidate` | `scenario_verify` | note |",
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(statusPath, "_status.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write status: %v", err)
	}

	updated, err := UpdateObjectNextCommand(repoRoot, "scenario", "demo", "scenario_check")
	if err == nil || !strings.Contains(err.Error(), "object type \"scenario\" is not a supported status value") {
		t.Fatalf("expected scenario rejection, got updated=%v err=%v", updated, err)
	}
}
