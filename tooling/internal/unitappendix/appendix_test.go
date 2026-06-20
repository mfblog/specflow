package unitappendix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanRequiresOwnedAppendixFrontmatter(t *testing.T) {
	repoRoot := t.TempDir()
	writeTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/appendix/c_unit_demo_prompt.md"), strings.Join([]string{
		"---",
		"unit: other",
		"layer: candidate",
		"---",
		"",
		"# Prompt",
	}, "\n")+"\n")

	_, err := Scan(repoRoot, "unit", "demo", "candidate")
	if err == nil || !strings.Contains(err.Error(), "frontmatter.unit mismatch") {
		t.Fatalf("expected frontmatter unit mismatch, got %v", err)
	}
}

func TestCandidateCoverageAllowsExtraCandidateAppendix(t *testing.T) {
	repoRoot := t.TempDir()
	writeTestFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/appendix/s_unit_demo_prompt.md"), strings.Join([]string{
		"---",
		"unit: demo",
		"layer: stable",
		"---",
		"",
		"# Stable Prompt",
	}, "\n")+"\n")
	writeTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/appendix/c_unit_demo_prompt.md"), strings.Join([]string{
		"---",
		"unit: demo",
		"layer: candidate",
		"---",
		"",
		"# Candidate Prompt",
	}, "\n")+"\n")
	writeTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/appendix/c_unit_demo_extra.md"), strings.Join([]string{
		"---",
		"unit: demo",
		"layer: candidate",
		"---",
		"",
		"# Extra",
	}, "\n")+"\n")

	if err := ValidateCandidateCoverage(repoRoot, "unit", "demo"); err != nil {
		t.Fatalf("ValidateCandidateCoverage: %v", err)
	}
}

func TestCandidateCoverageAllowsExemptStableAppendix(t *testing.T) {
	repoRoot := t.TempDir()
	writeTestFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/appendix/s_unit_demo_prompt.md"), strings.Join([]string{
		"---",
		"unit: demo",
		"layer: stable",
		"status: exempt",
		"---",
		"",
		"# Exempt Stable Prompt",
	}, "\n")+"\n")

	if err := ValidateCandidateCoverage(repoRoot, "unit", "demo"); err != nil {
		t.Fatalf("ValidateCandidateCoverage with exempt stable appendix: %v", err)
	}
}

func TestCandidateCoverageRejectsMissingCandidateAppendix(t *testing.T) {
	repoRoot := t.TempDir()
	writeTestFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/appendix/s_unit_demo_prompt.md"), strings.Join([]string{
		"---",
		"unit: demo",
		"layer: stable",
		"---",
		"",
		"# Stable Prompt",
	}, "\n")+"\n")

	err := ValidateCandidateCoverage(repoRoot, "unit", "demo")
	if err == nil || !strings.Contains(err.Error(), "missing candidate appendix") {
		t.Fatalf("expected missing candidate appendix, got %v", err)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
