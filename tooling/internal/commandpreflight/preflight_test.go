package commandpreflight

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunUnitPlanDoesNotTreatCheckWorkAsGate(t *testing.T) {
	repoRoot := t.TempDir()
	mustWritePreflightFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_plan` | note |",
	}, "\n")+"\n")
	mustWritePreflightFile(t, filepath.Join(repoRoot, "docs/specs/_check_work/unit/demo.md"), "# check work\n")

	result := Run(repoRoot, "unit_plan", "unit", "demo")
	if result.MayContinue {
		t.Fatalf("unit_plan must not continue from _check_work alone: %+v", result)
	}
	if result.FailureLayer != "gate_layer" || result.RecommendedNextCommand != "unit_check" {
		t.Fatalf("expected gate fallback to unit_check, got %+v", result)
	}
	if len(result.ValidatedProcesses) != 1 || result.ValidatedProcesses[0].ProcessFile != "docs/specs/_check_result/unit/demo.md" {
		t.Fatalf("expected preflight to require _check_result, got %+v", result.ValidatedProcesses)
	}
}

func mustWritePreflightFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
