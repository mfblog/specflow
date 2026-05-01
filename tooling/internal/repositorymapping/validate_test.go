package repositorymapping

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateAcceptsRuleBasedUnitTruthPaths(t *testing.T) {
	repoRoot := t.TempDir()
	writeTestRepo(t, repoRoot, []statusRow{
		{objectType: "unit", object: "candidate_demo", stable: "no", candidate: "yes", activeLayer: "candidate", nextCommand: "unit_check"},
		{objectType: "unit", object: "stable_demo", stable: "yes", candidate: "no", activeLayer: "stable", nextCommand: "unit_fork"},
	})
	writeMapping(t, repoRoot, "unit_default", nil)
	writeFile(t, repoRoot, "docs/specs/units/candidate/c_unit_candidate_demo.md", "# Candidate\n")
	writeFile(t, repoRoot, "docs/specs/units/stable/s_unit_stable_demo.md", "# Stable\n")
	writeFile(t, repoRoot, "docs/specs/shared_contracts/stable/s_shared_demo.md", "# Shared\n")

	result, err := Validate(repoRoot)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if !result.Valid() {
		t.Fatalf("expected valid mapping, got diagnostics: %v", result.Diagnostics)
	}
}

func TestValidateRejectsLegacyConcreteTruthSurface(t *testing.T) {
	repoRoot := t.TempDir()
	writeTestRepo(t, repoRoot, []statusRow{
		{objectType: "unit", object: "demo", stable: "no", candidate: "yes", activeLayer: "candidate", nextCommand: "unit_check"},
	})
	writeMapping(t, repoRoot, "legacy", nil)
	writeFile(t, repoRoot, "docs/specs/units/candidate/c_unit_demo.md", "# Demo\n")
	writeFile(t, repoRoot, "docs/specs/shared_contracts/stable/s_shared_demo.md", "# Shared\n")

	result, err := Validate(repoRoot)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if result.Valid() {
		t.Fatal("expected legacy truth_surface diagnostics")
	}
	if !containsDiagnostic(result.Diagnostics, "deprecated truth_surface") || !containsDiagnostic(result.Diagnostics, "concrete unit truth path") {
		t.Fatalf("expected legacy diagnostics, got %v", result.Diagnostics)
	}
}

func TestValidateRejectsMissingStableTruthFile(t *testing.T) {
	repoRoot := t.TempDir()
	writeTestRepo(t, repoRoot, []statusRow{
		{objectType: "unit", object: "demo", stable: "yes", candidate: "no", activeLayer: "stable", nextCommand: "unit_fork"},
	})
	writeMapping(t, repoRoot, "unit_default", nil)
	writeFile(t, repoRoot, "docs/specs/shared_contracts/stable/s_shared_demo.md", "# Shared\n")

	result, err := Validate(repoRoot)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if result.Valid() {
		t.Fatal("expected missing stable file diagnostics")
	}
	if !containsDiagnostic(result.Diagnostics, "stable truth file does not exist") {
		t.Fatalf("expected missing stable diagnostic, got %v", result.Diagnostics)
	}
}

func TestValidateRejectsMissingSharedContractPath(t *testing.T) {
	repoRoot := t.TempDir()
	writeTestRepo(t, repoRoot, []statusRow{
		{objectType: "unit", object: "demo", stable: "no", candidate: "yes", activeLayer: "candidate", nextCommand: "unit_check"},
	})
	writeMapping(t, repoRoot, "unit_default", []string{"docs/specs/shared_contracts/stable/s_shared_missing.md"})
	writeFile(t, repoRoot, "docs/specs/units/candidate/c_unit_demo.md", "# Demo\n")

	result, err := Validate(repoRoot)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if result.Valid() {
		t.Fatal("expected missing shared contract diagnostics")
	}
	if !containsDiagnostic(result.Diagnostics, "shared contract path does not exist") {
		t.Fatalf("expected missing shared diagnostic, got %v", result.Diagnostics)
	}
}

type statusRow struct {
	objectType  string
	object      string
	stable      string
	candidate   string
	activeLayer string
	nextCommand string
}

func writeTestRepo(t *testing.T, repoRoot string, rows []statusRow) {
	t.Helper()
	lines := []string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
	}
	for _, row := range rows {
		lines = append(lines, "| `"+row.objectType+"` | `"+row.object+"` | `"+row.stable+"` | `"+row.candidate+"` | `"+row.activeLayer+"` | `"+row.nextCommand+"` | note |")
	}
	writeFile(t, repoRoot, "docs/specs/_status.md", strings.Join(lines, "\n")+"\n")
}

func writeMapping(t *testing.T, repoRoot, mode string, sharedPaths []string) {
	t.Helper()
	if sharedPaths == nil {
		sharedPaths = []string{"docs/specs/shared_contracts/stable/s_shared_demo.md"}
	}
	unitBlock := []string{
		"1. `demo`",
		"   - `truth_surface_rule`: `unit_default`",
	}
	if mode == "legacy" {
		unitBlock = []string{
			"1. `demo`",
			"   - `truth_surface`",
			"     - `docs/specs/units/candidate/c_unit_demo.md`",
		}
	}
	content := []string{
		"# Repository Mapping",
		"",
		"### 2.1 Current Units",
		"",
		"1. `demo`",
		"   - demo unit",
		"2. `candidate_demo`",
		"   - candidate demo unit",
		"3. `stable_demo`",
		"   - stable demo unit",
		"",
		"### 2.3 Current Shared Contracts",
		"",
		"1. `demo`",
		"   - demo shared",
		"",
		"### 4.5 Shared Contract Truth Paths",
		"",
		"1. `demo`",
	}
	for _, sharedPath := range sharedPaths {
		content = append(content, "   - `"+sharedPath+"`")
	}
	content = append(content,
		"",
		"### 4.6 Unit Truth Rules And Implementation Paths",
		"",
		"1. `stable` -> `docs/specs/units/stable/s_unit_{unit}.md`",
		"2. `candidate` -> `docs/specs/units/candidate/c_unit_{unit}.md`",
		"",
	)
	content = append(content, unitBlock...)
	content = append(content,
		"   - `implementation_surface`",
		"     - 当前未声明独占实现路径，后续结构真相更新时再绑定。",
		"2. `candidate_demo`",
		"   - `truth_surface_rule`: `unit_default`",
		"   - `implementation_surface`",
		"     - 当前未声明独占实现路径，后续结构真相更新时再绑定。",
		"3. `stable_demo`",
		"   - `truth_surface_rule`: `unit_default`",
		"   - `implementation_surface`",
		"     - 当前未声明独占实现路径，后续结构真相更新时再绑定。",
		"",
		"### 4.7 Conflict Rules",
	)
	writeFile(t, repoRoot, "docs/specs/repository_mapping.md", strings.Join(content, "\n")+"\n")
}

func writeFile(t *testing.T, repoRoot, relPath, content string) {
	t.Helper()
	path := filepath.Join(repoRoot, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func containsDiagnostic(diagnostics []string, pattern string) bool {
	for _, diagnostic := range diagnostics {
		if strings.Contains(diagnostic, pattern) {
			return true
		}
	}
	return false
}
