package entrysync

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInspectOnlyReadsRegisteredEntrySection(t *testing.T) {
	repoRoot := t.TempDir()
	writeInstalledLayoutMarker(t, repoRoot)

	regDir := filepath.Join(repoRoot, "specflow/framework/operations")
	if err := os.MkdirAll(regDir, 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	registry := `# Entry Routing

## Entry File Registration

Registered entry index files: ` + "`AGENTS.md`, `GEMINI.md`, `CLAUDE.md`" + `.

## Implementation Classification

Risk levels guide routing.
`
	if err := os.WriteFile(filepath.Join(regDir, "entry_routing.md"), []byte(registry), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}

	block := "==SPECFLOW:BEGIN==\nmanaged\n==SPECFLOW:END==\n"
	for _, name := range []string{"AGENTS.md", "GEMINI.md", "CLAUDE.md"} {
		if err := os.WriteFile(filepath.Join(repoRoot, name), []byte(block), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	inspection, err := Inspect(repoRoot)
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if !inspection.Consistent {
		t.Fatalf("expected inspection to be consistent")
	}
	if len(inspection.RegisteredFiles) != 3 {
		t.Fatalf("expected 3 registered files, got %d: %v", len(inspection.RegisteredFiles), inspection.RegisteredFiles)
	}
}

func TestInspectSuggestsOnlyCurrentRoundChangedRegisteredFile(t *testing.T) {
	repoRoot := t.TempDir()
	writeInstalledLayoutMarker(t, repoRoot)

	regDir := filepath.Join(repoRoot, "specflow/framework/operations")
	if err := os.MkdirAll(regDir, 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	registry := `## Entry File Registration

Registered entry index files: ` + "`AGENTS.md`, `GEMINI.md`, `CLAUDE.md`" + `.
`
	if err := os.WriteFile(filepath.Join(regDir, "entry_routing.md"), []byte(registry), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}

	for name, content := range map[string]string{
		"AGENTS.md": "==SPECFLOW:BEGIN==\nmanaged agents\n==SPECFLOW:END==\n",
		"GEMINI.md": "==SPECFLOW:BEGIN==\nmanaged base\n==SPECFLOW:END==\n",
		"CLAUDE.md": "==SPECFLOW:BEGIN==\nmanaged base\n==SPECFLOW:END==\n",
	} {
		if err := os.WriteFile(filepath.Join(repoRoot, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	initGitRepo(t, repoRoot)

	if err := os.WriteFile(filepath.Join(repoRoot, "AGENTS.md"), []byte("==SPECFLOW:BEGIN==\nmanaged changed this round\n==SPECFLOW:END==\n"), 0o644); err != nil {
		t.Fatalf("rewrite AGENTS.md: %v", err)
	}

	inspection, err := Inspect(repoRoot)
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if inspection.Consistent {
		t.Fatalf("expected inspection to be inconsistent")
	}
	if inspection.SuggestedSource != "AGENTS.md" {
		t.Fatalf("expected AGENTS.md as suggested source, got %q", inspection.SuggestedSource)
	}
	if len(inspection.CurrentRoundChanged) != 1 || inspection.CurrentRoundChanged[0] != "AGENTS.md" {
		t.Fatalf("unexpected current-round changed files: %v", inspection.CurrentRoundChanged)
	}
}

func TestInspectTreatsUntrackedRegisteredEntryFileAsCurrentRoundChanged(t *testing.T) {
	repoRoot := t.TempDir()
	writeInstalledLayoutMarker(t, repoRoot)

	regDir := filepath.Join(repoRoot, "specflow/framework/operations")
	if err := os.MkdirAll(regDir, 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	initialRegistry := `## Entry File Registration

Registered entry index files: ` + "`AGENTS.md`, `GEMINI.md`, `CLAUDE.md`" + `.
`
	if err := os.WriteFile(filepath.Join(regDir, "entry_routing.md"), []byte(initialRegistry), 0o644); err != nil {
		t.Fatalf("write initial registry: %v", err)
	}

	for _, name := range []string{"AGENTS.md", "GEMINI.md", "CLAUDE.md"} {
		if err := os.WriteFile(filepath.Join(repoRoot, name), []byte("==SPECFLOW:BEGIN==\nmanaged base\n==SPECFLOW:END==\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	initGitRepo(t, repoRoot)

	updatedRegistry := `## Entry File Registration

Registered entry index files: ` + "`AGENTS.md`, `GEMINI.md`, `CLAUDE.md`, `GUIDE.md`" + `.
`
	if err := os.WriteFile(filepath.Join(regDir, "entry_routing.md"), []byte(updatedRegistry), 0o644); err != nil {
		t.Fatalf("write updated registry: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "GUIDE.md"), []byte("==SPECFLOW:BEGIN==\nmanaged guide\n==SPECFLOW:END==\n"), 0o644); err != nil {
		t.Fatalf("write GUIDE.md: %v", err)
	}

	inspection, err := Inspect(repoRoot)
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if inspection.Consistent {
		t.Fatalf("expected inspection to be inconsistent")
	}
	if inspection.SuggestedSource != "GUIDE.md" {
		t.Fatalf("expected GUIDE.md as suggested source, got %q", inspection.SuggestedSource)
	}
	if len(inspection.CurrentRoundChanged) != 1 || inspection.CurrentRoundChanged[0] != "GUIDE.md" {
		t.Fatalf("unexpected current-round changed files: %v", inspection.CurrentRoundChanged)
	}
}

func TestInspectSourceRepoUsesTemplateEntryFiles(t *testing.T) {
	repoRoot := t.TempDir()
	writeSourceLayoutMarker(t, repoRoot)

	regDir := filepath.Join(repoRoot, "framework/operations")
	if err := os.MkdirAll(regDir, 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	registry := `## Entry File Registration

Registered entry index files: ` + "`AGENTS.md`, `GEMINI.md`, `CLAUDE.md`" + `.
`
	if err := os.WriteFile(filepath.Join(regDir, "entry_routing.md"), []byte(registry), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}

	block := "==SPECFLOW:BEGIN==\nmanaged\n==SPECFLOW:END==\n"
	if err := os.MkdirAll(filepath.Join(repoRoot, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir templates dir: %v", err)
	}
	for _, name := range []string{"AGENTS.md", "GEMINI.md", "CLAUDE.md"} {
		if err := os.WriteFile(filepath.Join(repoRoot, "templates", name), []byte(block), 0o644); err != nil {
			t.Fatalf("write template %s: %v", name, err)
		}
	}

	inspection, err := Inspect(repoRoot)
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if !inspection.Consistent {
		t.Fatalf("expected source template inspection to be consistent")
	}
	for _, expected := range []string{"templates/AGENTS.md", "templates/CLAUDE.md", "templates/GEMINI.md"} {
		if !contains(inspection.RegisteredFiles, expected) {
			t.Fatalf("expected source registered file %s, got %+v", expected, inspection.RegisteredFiles)
		}
	}
}

func TestSyncSourceRepoAcceptsLogicalRegisteredSource(t *testing.T) {
	repoRoot := t.TempDir()
	writeSourceLayoutMarker(t, repoRoot)

	regDir := filepath.Join(repoRoot, "framework/operations")
	if err := os.MkdirAll(regDir, 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	registry := `## Entry File Registration

Registered entry index files: ` + "`AGENTS.md`, `GEMINI.md`" + `.
`
	if err := os.WriteFile(filepath.Join(regDir, "entry_routing.md"), []byte(registry), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repoRoot, "templates"), 0o755); err != nil {
		t.Fatalf("mkdir templates dir: %v", err)
	}
	agentsBlock := "==SPECFLOW:BEGIN==\nmanaged agents\n==SPECFLOW:END==\n"
	geminiBlock := "==SPECFLOW:BEGIN==\nmanaged gemini\n==SPECFLOW:END==\n"
	if err := os.WriteFile(filepath.Join(repoRoot, "templates/AGENTS.md"), []byte(agentsBlock), 0o644); err != nil {
		t.Fatalf("write template AGENTS.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "templates/GEMINI.md"), []byte(geminiBlock), 0o644); err != nil {
		t.Fatalf("write template GEMINI.md: %v", err)
	}

	result, err := Sync(repoRoot, "AGENTS.md")
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if result.Source != "templates/AGENTS.md" {
		t.Fatalf("expected normalized source, got %q", result.Source)
	}
	if len(result.UpdatedFiles) != 1 || result.UpdatedFiles[0] != "templates/GEMINI.md" {
		t.Fatalf("unexpected updated files: %+v", result.UpdatedFiles)
	}
}

func writeInstalledLayoutMarker(t *testing.T, repoRoot string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(repoRoot, "specflow/tooling"), 0o755); err != nil {
		t.Fatalf("mkdir installed tooling marker: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "specflow/tooling", "manifest.tsv"), []byte("path\tsha256\n"), 0o644); err != nil {
		t.Fatalf("write installed tooling marker: %v", err)
	}
}

func writeSourceLayoutMarker(t *testing.T, repoRoot string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(repoRoot, "tooling"), 0o755); err != nil {
		t.Fatalf("mkdir source tooling marker: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "tooling", "manifest.tsv"), []byte("path\tsha256\n"), 0o644); err != nil {
		t.Fatalf("write source tooling marker: %v", err)
	}
}

func initGitRepo(t *testing.T, repoRoot string) {
	t.Helper()
	runGit(t, repoRoot, "init")
	runGit(t, repoRoot, "config", "user.name", "SpecFlow Test")
	runGit(t, repoRoot, "config", "user.email", "specflow@example.com")
	runGit(t, repoRoot, "add", ".")
	runGit(t, repoRoot, "commit", "-m", "init")
}

func runGit(t *testing.T, repoRoot string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}
