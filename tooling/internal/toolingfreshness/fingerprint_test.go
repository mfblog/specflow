package toolingfreshness

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLiveFingerprintChangesWhenToolingSourceChanges(t *testing.T) {
	repoRoot := t.TempDir()
	writeToolingRepo(t, repoRoot)

	first, files, err := LiveFingerprint(repoRoot)
	if err != nil {
		t.Fatalf("LiveFingerprint returned error: %v", err)
	}
	if len(files) != 4 {
		t.Fatalf("expected 4 fingerprint input files, got %d", len(files))
	}

	mustWriteFile(t, filepath.Join(repoRoot, "specflow/tooling/internal/demo/demo.go"), "package demo\n\nfunc Value() string { return \"changed\" }\n")

	second, _, err := LiveFingerprint(repoRoot)
	if err != nil {
		t.Fatalf("LiveFingerprint after change returned error: %v", err)
	}
	if first == second {
		t.Fatalf("expected fingerprint to change after tooling source change")
	}
}

func TestLiveFingerprintChangesWhenManifestChanges(t *testing.T) {
	repoRoot := t.TempDir()
	writeToolingRepo(t, repoRoot)

	first, _, err := LiveFingerprint(repoRoot)
	if err != nil {
		t.Fatalf("LiveFingerprint returned error: %v", err)
	}

	mustWriteFile(t, filepath.Join(repoRoot, "specflow/tooling/manifest.tsv"), "templates/GEMINI.md\tGEMINI.md\tframework\n")

	second, _, err := LiveFingerprint(repoRoot)
	if err != nil {
		t.Fatalf("LiveFingerprint after manifest change returned error: %v", err)
	}
	if first == second {
		t.Fatalf("expected fingerprint to change after manifest change")
	}
}

func TestLiveFingerprintIgnoresNonToolingFiles(t *testing.T) {
	repoRoot := t.TempDir()
	writeToolingRepo(t, repoRoot)

	first, _, err := LiveFingerprint(repoRoot)
	if err != nil {
		t.Fatalf("LiveFingerprint returned error: %v", err)
	}

	mustWriteFile(t, filepath.Join(repoRoot, "specflow/tooling/README.md"), "# changed doc\n")

	second, _, err := LiveFingerprint(repoRoot)
	if err != nil {
		t.Fatalf("LiveFingerprint after doc change returned error: %v", err)
	}
	if first != second {
		t.Fatalf("expected fingerprint to stay unchanged after non-tooling file change")
	}
}

func TestLiveFingerprintIgnoresReaderAssetChanges(t *testing.T) {
	repoRoot := t.TempDir()
	writeToolingRepo(t, repoRoot)

	first, _, err := LiveFingerprint(repoRoot)
	if err != nil {
		t.Fatalf("LiveFingerprint returned error: %v", err)
	}

	mustWriteFile(t, filepath.Join(repoRoot, "specflow/tooling/reader/web/app.js"), "console.log('changed');\n")

	second, _, err := LiveFingerprint(repoRoot)
	if err != nil {
		t.Fatalf("LiveFingerprint after reader asset change returned error: %v", err)
	}
	if first != second {
		t.Fatalf("expected fingerprint to ignore reader asset change")
	}
}

func TestShellFingerprintScriptMatchesLiveFingerprint(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash is not available")
	}
	if err := exec.Command("bash", "-lc", "true").Run(); err != nil {
		t.Skipf("bash is not usable in this environment: %v", err)
	}

	repoRoot, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	want, _, err := LiveFingerprint(repoRoot)
	if err != nil {
		t.Fatalf("LiveFingerprint returned error: %v", err)
	}

	scriptPath := filepath.Join(repoRoot, "specflow", "tooling", "scripts", "tooling_fingerprint.sh")
	out, err := exec.Command("bash", scriptPath).CombinedOutput()
	if err != nil {
		t.Fatalf("tooling_fingerprint.sh failed: %v\n%s", err, string(out))
	}
	if got := strings.TrimSpace(string(out)); got != want {
		t.Fatalf("script fingerprint mismatch\ngot  %s\nwant %s", got, want)
	}

	shortOut, err := exec.Command("bash", scriptPath, "--short").CombinedOutput()
	if err != nil {
		t.Fatalf("tooling_fingerprint.sh --short failed: %v\n%s", err, string(shortOut))
	}
	if got, want := strings.TrimSpace(string(shortOut)), want[:12]; got != want {
		t.Fatalf("script short fingerprint mismatch\ngot  %s\nwant %s", got, want)
	}
}

func writeToolingRepo(t *testing.T, repoRoot string) {
	t.Helper()
	mustWriteFile(t, filepath.Join(repoRoot, "specflow/tooling/go.mod"), "module github.com/Bingordinary/SpecFlow/specflow/tooling\n\ngo 1.22.2\n")
	mustWriteFile(t, filepath.Join(repoRoot, "specflow/tooling/manifest.tsv"), "templates/AGENTS.md\tAGENTS.md\tframework\n")
	mustWriteFile(t, filepath.Join(repoRoot, "specflow/tooling/cmd/specflowctl/main.go"), "package main\n\nfunc main() {}\n")
	mustWriteFile(t, filepath.Join(repoRoot, "specflow/tooling/internal/demo/demo.go"), "package demo\n\nfunc Value() string { return \"demo\" }\n")
	mustWriteFile(t, filepath.Join(repoRoot, "specflow/tooling/reader/web/app.js"), "console.log('demo');\n")
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) failed: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) failed: %v", path, err)
	}
}
