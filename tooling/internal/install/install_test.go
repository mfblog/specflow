package install

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/buildrelease"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/toolingfreshness"
)

func TestDoctorPassesForFreshBinary(t *testing.T) {
	repoRoot := t.TempDir()
	liveFingerprint := setupDoctorRepo(t, repoRoot)
	writeFingerprintProbeBinary(t, repoRoot, liveFingerprint)

	result, err := Doctor(repoRoot)
	if err != nil {
		t.Fatalf("Doctor returned error: %v", err)
	}
	if len(result.Failures) != 0 {
		t.Fatalf("expected no failures, got %v", result.Failures)
	}
}

func TestDoctorFailsForStaleBinary(t *testing.T) {
	repoRoot := t.TempDir()
	_ = setupDoctorRepo(t, repoRoot)
	writeFingerprintProbeBinary(t, repoRoot, "stale-binary-fingerprint")

	result, err := Doctor(repoRoot)
	if err != nil {
		t.Fatalf("Doctor returned unexpected error: %v", err)
	}

	joined := strings.Join(result.Failures, "\n")
	if !strings.Contains(joined, "STALE specflow/tooling/bin/"+buildrelease.CurrentBinaryName()) {
		t.Fatalf("expected stale binary failure, got %v", result.Failures)
	}
	if !strings.Contains(joined, "STALE specflow/tooling/bin/"+buildrelease.CurrentReaderBinaryName()) {
		t.Fatalf("expected stale reader binary failure, got %v", result.Failures)
	}
}

func TestDoctorFailsWhenReaderWebAssetIsMissing(t *testing.T) {
	repoRoot := t.TempDir()
	liveFingerprint := setupDoctorRepo(t, repoRoot)
	writeFingerprintProbeBinary(t, repoRoot, liveFingerprint)
	if err := os.Remove(filepath.Join(repoRoot, "specflow/tooling/reader/web/app.js")); err != nil {
		t.Fatalf("Remove(app.js) failed: %v", err)
	}

	result, err := Doctor(repoRoot)
	if err != nil {
		t.Fatalf("Doctor returned unexpected error: %v", err)
	}

	joined := strings.Join(result.Failures, "\n")
	if !strings.Contains(joined, "MISSING specflow/tooling/reader/web/app.js") {
		t.Fatalf("expected missing reader web asset failure, got %v", result.Failures)
	}
}

func TestDoctorReportsSourceRepoReaderPath(t *testing.T) {
	repoRoot := t.TempDir()
	setupDoctorRepoAt(t, repoRoot, "", "tooling")
	if err := os.Remove(filepath.Join(repoRoot, "tooling/reader/web/app.js")); err != nil {
		t.Fatalf("Remove(app.js) failed: %v", err)
	}

	result, err := Doctor(repoRoot)
	if err != nil {
		t.Fatalf("Doctor returned unexpected error: %v", err)
	}

	joined := strings.Join(result.Failures, "\n")
	if !strings.Contains(joined, "MISSING tooling/reader/web/app.js") {
		t.Fatalf("expected source-repo reader asset failure, got %v", result.Failures)
	}
	if strings.Contains(joined, "MISSING specflow/tooling/reader/web/app.js") {
		t.Fatalf("doctor resolved the installed reader path in a source repo: %v", result.Failures)
	}
	if strings.Contains(joined, "MISSING docs/") {
		t.Fatalf("doctor checked installed-project destinations in a source repo: %v", result.Failures)
	}

}

func setupDoctorRepo(t *testing.T, repoRoot string) string {
	t.Helper()
	return setupDoctorRepoAt(t, repoRoot, "specflow", "specflow/tooling")
}

func setupDoctorRepoAt(t *testing.T, repoRoot, contentRoot, toolingRoot string) string {
	t.Helper()
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(toolingRoot), "manifest.tsv"), strings.Join([]string{
		"templates/AGENTS.md\tAGENTS.md\tframework",
	}, "\n")+"\n")
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(contentRoot), "templates/AGENTS.md"), "template\n==SPECFLOW:BEGIN==\nmanaged\n==SPECFLOW:END==\n")
	mustWriteFile(t, filepath.Join(repoRoot, "AGENTS.md"), "host\n==SPECFLOW:BEGIN==\nmanaged\n==SPECFLOW:END==\n")
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(toolingRoot), "go.mod"), "module github.com/Bingordinary/SpecFlow/specflow/tooling\n\ngo 1.22.2\n")
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(toolingRoot), "cmd/specflowctl/main.go"), "package main\n\nfunc main() {}\n")
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(toolingRoot), "cmd/specflow-reader/main.go"), "package main\n\nfunc main() {}\n")
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(toolingRoot), "internal/demo/demo.go"), "package demo\n\nfunc Value() string { return \"demo\" }\n")
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(toolingRoot), "reader/web/index.html"), "<!doctype html>\n")
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(toolingRoot), "reader/web/styles.css"), "body { color: #111; }\n")
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(toolingRoot), "reader/web/app.js"), "console.log('demo');\n")
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(toolingRoot), "reader/web/cytoscape.min.js"), "window.cytoscape = function() {};\n")
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(toolingRoot), "reader/web/mermaid.min.js"), "window.mermaid = { initialize() {}, run() {} };\n")

	fingerprint, _, err := toolingfreshness.LiveFingerprint(repoRoot)
	if err != nil {
		t.Fatalf("LiveFingerprint returned error: %v", err)
	}
	return fingerprint
}

func writeFingerprintProbeBinary(t *testing.T, repoRoot, fingerprint string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("script-based executable probe is not supported on windows")
	}
	script := "#!/usr/bin/env bash\nif [[ \"$1\" == \"" + toolingfreshness.HiddenBuildFingerprintCommand + "\" ]]; then\n  printf '%s\\n' \"" + fingerprint + "\"\n  exit 0\nfi\nexit 0\n"
	mustWriteExecutableFile(t, filepath.Join(repoRoot, "specflow/tooling/bin", buildrelease.CurrentBinaryName()), script)
	mustWriteExecutableFile(t, filepath.Join(repoRoot, "specflow/tooling/bin", buildrelease.CurrentReaderBinaryName()), script)
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

func mustWriteExecutableFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) failed: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile(%s) failed: %v", path, err)
	}
}
