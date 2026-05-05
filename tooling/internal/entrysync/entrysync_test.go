package entrysync

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInspectOnlyReadsRegisteredEntrySection(t *testing.T) {
	repoRoot := t.TempDir()

	registryDir := filepath.Join(repoRoot, "specflow/framework")
	if err := os.MkdirAll(registryDir, 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	registry := `# Entry Index Registry

## Registered Entry Index Files

- ` + "`AGENTS.md`" + `
- ` + "`GEMINI.md`" + `
- ` + "`CLAUDE.md`" + `

## Manual Sync Trigger

- ` + "`specflow/tooling/bin/specflowctl-linux-amd64 entry sync --source AGENTS.md`" + `
`
	if err := os.WriteFile(filepath.Join(registryDir, "entry_index_registry.md"), []byte(registry), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}

	block := "<!-- SPECFLOW:BEGIN -->\nmanaged\n<!-- SPECFLOW:END -->\n"
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

	registryDir := filepath.Join(repoRoot, "specflow/framework")
	if err := os.MkdirAll(registryDir, 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	registry := `# Entry Index Registry

## Registered Entry Index Files

- ` + "`AGENTS.md`" + `
- ` + "`GEMINI.md`" + `
- ` + "`CLAUDE.md`" + `
`
	if err := os.WriteFile(filepath.Join(registryDir, "entry_index_registry.md"), []byte(registry), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}

	for name, content := range map[string]string{
		"AGENTS.md": "<!-- SPECFLOW:BEGIN -->\nmanaged agents\n<!-- SPECFLOW:END -->\n",
		"GEMINI.md": "<!-- SPECFLOW:BEGIN -->\nmanaged base\n<!-- SPECFLOW:END -->\n",
		"CLAUDE.md": "<!-- SPECFLOW:BEGIN -->\nmanaged base\n<!-- SPECFLOW:END -->\n",
	} {
		if err := os.WriteFile(filepath.Join(repoRoot, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	initGitRepo(t, repoRoot)

	if err := os.WriteFile(filepath.Join(repoRoot, "AGENTS.md"), []byte("<!-- SPECFLOW:BEGIN -->\nmanaged changed this round\n<!-- SPECFLOW:END -->\n"), 0o644); err != nil {
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

	registryDir := filepath.Join(repoRoot, "specflow/framework")
	if err := os.MkdirAll(registryDir, 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	initialRegistry := `# Entry Index Registry

## Registered Entry Index Files

- ` + "`AGENTS.md`" + `
- ` + "`GEMINI.md`" + `
- ` + "`CLAUDE.md`" + `
`
	if err := os.WriteFile(filepath.Join(registryDir, "entry_index_registry.md"), []byte(initialRegistry), 0o644); err != nil {
		t.Fatalf("write initial registry: %v", err)
	}

	for _, name := range []string{"AGENTS.md", "GEMINI.md", "CLAUDE.md"} {
		if err := os.WriteFile(filepath.Join(repoRoot, name), []byte("<!-- SPECFLOW:BEGIN -->\nmanaged base\n<!-- SPECFLOW:END -->\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	initGitRepo(t, repoRoot)

	updatedRegistry := `# Entry Index Registry

## Registered Entry Index Files

- ` + "`AGENTS.md`" + `
- ` + "`GEMINI.md`" + `
- ` + "`CLAUDE.md`" + `
- ` + "`GUIDE.md`" + `
`
	if err := os.WriteFile(filepath.Join(registryDir, "entry_index_registry.md"), []byte(updatedRegistry), 0o644); err != nil {
		t.Fatalf("write updated registry: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "GUIDE.md"), []byte("<!-- SPECFLOW:BEGIN -->\nmanaged guide\n<!-- SPECFLOW:END -->\n"), 0o644); err != nil {
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
