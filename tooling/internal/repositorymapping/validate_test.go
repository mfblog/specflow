package repositorymapping

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestValidateAcceptsObjectRegistry(t *testing.T) {
	repoRoot := t.TempDir()
	writeMapping(t, repoRoot, []string{
		"| unit | demo | landed | `AgentCore/internal/demo/**` | `docs/specs/units/candidate/c_unit_demo.md` | demo unit |",
		"| rule | b_rule_future | planned | none | none | future rule |",
		"| rule | b_rule_demo | planned | none | `docs/specs/rules/stable/s_b_rule_demo.md` | demo rule |",
	})
	writeFile(t, repoRoot, "AgentCore/internal/demo/service.go", "package demo\n")
	writeFile(t, repoRoot, "docs/specs/units/candidate/c_unit_demo.md", "# Demo\n")
	writeFile(t, repoRoot, "docs/specs/rules/stable/s_b_rule_demo.md", "# Rule\n")

	result, err := Validate(repoRoot)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if !result.Valid() {
		t.Fatalf("expected valid mapping, got diagnostics: %v", result.Diagnostics)
	}
}

func TestValidateAcceptsSourceTemplateBootstrap(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	templateRoot := filepath.Join(repoRoot, "templates")

	result, err := Validate(templateRoot)
	if err != nil {
		t.Fatalf("Validate template bootstrap returned error: %v", err)
	}
	if !result.Valid() {
		t.Fatalf("expected source template bootstrap mapping to validate, got diagnostics: %v", result.Diagnostics)
	}
}

func TestValidateRejectsMissingLandedImplementationPath(t *testing.T) {
	repoRoot := t.TempDir()
	writeMapping(t, repoRoot, []string{
		"| rule | b_rule_missing | landed | `AgentCore/internal/missing/**` | none | missing rule |",
	})

	result, err := Validate(repoRoot)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if result.Valid() || !containsDiagnostic(result.Diagnostics, "registered implementation path does not exist") {
		t.Fatalf("expected missing implementation diagnostic, got %v", result.Diagnostics)
	}
}

func TestValidateRejectsPlannedImplementationPath(t *testing.T) {
	repoRoot := t.TempDir()
	writeMapping(t, repoRoot, []string{
		"| rule | b_rule_future | planned | `AgentCore/internal/future/**` | none | future rule |",
	})

	result, err := Validate(repoRoot)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if result.Valid() || !containsDiagnostic(result.Diagnostics, "planned object must use implementation_paths=none") {
		t.Fatalf("expected planned implementation diagnostic, got %v", result.Diagnostics)
	}
}

func TestValidateRejectsInvalidRegistryRow(t *testing.T) {
	repoRoot := t.TempDir()
	writeMapping(t, repoRoot, []string{
		"| feature | demo | landed | `AgentCore/internal/demo/**` | none | invalid kind |",
		"| unit | planned_demo | waiting | none | none | invalid state |",
	})

	result, err := Validate(repoRoot)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if result.Valid() || !containsDiagnostic(result.Diagnostics, "invalid kind") || !containsDiagnostic(result.Diagnostics, "invalid registration_state") {
		t.Fatalf("expected invalid row diagnostics, got %v", result.Diagnostics)
	}
}

func TestValidateAcceptsLandedUnitWithSpecFile(t *testing.T) {
	repoRoot := t.TempDir()
	writeMapping(t, repoRoot, []string{
		"| unit | demo | landed | `AgentCore/internal/demo/**` | `docs/specs/units/candidate/c_unit_demo.md` | demo unit |",
	})
	writeFile(t, repoRoot, "AgentCore/internal/demo/service.go", "package demo\n")
	writeFile(t, repoRoot, "docs/specs/units/candidate/c_unit_demo.md", "# Demo\n")

	result, err := Validate(repoRoot)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if !result.Valid() {
		t.Fatalf("expected valid mapping for landed unit with spec file, got diagnostics: %v", result.Diagnostics)
	}
}

func writeMapping(t *testing.T, repoRoot string, rows []string) {
	t.Helper()
	content := []string{
		"# Repository Mapping",
		"",
		"## 2. Object Registry",
		"",
		"| kind | id | registration_state | implementation_paths | spec_files | responsibility |",
		"|---|---|---|---|---|---|",
	}
	content = append(content, rows...)
	content = append(content, "", "## 3. Boundary Rules", "")
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
