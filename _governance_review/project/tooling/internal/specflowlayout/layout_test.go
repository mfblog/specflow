package specflowlayout

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveSourceRepo(t *testing.T) {
	repoRoot := t.TempDir()
	writeMarker(t, filepath.Join(repoRoot, "tooling", "manifest.tsv"))

	layout, err := Resolve(repoRoot)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if layout.Kind != SourceRepo || layout.ToolingRoot != "tooling" {
		t.Fatalf("unexpected layout: %+v", layout)
	}
}

func TestResolveInstalledProject(t *testing.T) {
	repoRoot := t.TempDir()
	writeMarker(t, filepath.Join(repoRoot, "specflow", "tooling", "manifest.tsv"))

	layout, err := Resolve(repoRoot)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if layout.Kind != InstalledProject || layout.ToolingRoot != "specflow/tooling" {
		t.Fatalf("unexpected layout: %+v", layout)
	}
}

func TestResolveRejectsAmbiguousLayouts(t *testing.T) {
	repoRoot := t.TempDir()
	writeMarker(t, filepath.Join(repoRoot, "tooling", "manifest.tsv"))
	writeMarker(t, filepath.Join(repoRoot, "specflow", "tooling", "manifest.tsv"))

	_, err := Resolve(repoRoot)
	if err == nil || !strings.Contains(err.Error(), "ambiguous specFlow layout") {
		t.Fatalf("expected ambiguous layout error, got %v", err)
	}
}

func writeMarker(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(path, []byte("marker\n"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
}
