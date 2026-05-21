package unitrelease

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseVersionUpdatesCandidateAndDeletesProcessFiles(t *testing.T) {
	repoRoot := t.TempDir()
	writeReleaseStatus(t, repoRoot,
		"| `unit` | `assistant` | `yes` | `no` | `stable` | `unit_fork` | test |\n"+
			"| `unit` | `trace` | `no` | `yes` | `candidate` | `unit_verify` | test |\n")
	writeReleaseUnit(t, repoRoot, "stable", "assistant", "0.9.0", nil)
	writeReleaseUnit(t, repoRoot, "candidate", "trace", "0.1.0", []string{"s_unit_assistant@0.8.0"})
	writeReleaseProcessFiles(t, repoRoot, "trace")

	result, err := ReleaseVersion(repoRoot, Options{
		Unit:    "assistant",
		FromRef: "s_unit_assistant@0.8.0",
		ToRef:   "s_unit_assistant@0.9.0",
	})
	if err != nil {
		t.Fatalf("ReleaseVersion returned error: %v", err)
	}
	if result.Noop {
		t.Fatalf("expected non-noop result")
	}
	if got := strings.Join(result.CandidateUpdated, ","); got != "unit:trace" {
		t.Fatalf("expected trace candidate update, got %q", got)
	}
	assertContains(t, readReleaseFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_trace.md")), "  - s_unit_assistant@0.9.0")
	assertNotExists(t, filepath.Join(repoRoot, "docs/specs/_check_work/unit/trace.md"))
	assertNotExists(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/trace.md"))
	assertNotExists(t, filepath.Join(repoRoot, "docs/specs/_plans/draft/trace.md"))
	assertNotExists(t, filepath.Join(repoRoot, "docs/specs/_plans/active/trace.md"))
	assertNotExists(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit/trace.md"))
	status := readReleaseFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"))
	assertContains(t, status, "| `unit` | `trace` | `no` | `yes` | `candidate` | `unit_check` | Retargeted by unit release-version from s_unit_assistant@0.8.0 to s_unit_assistant@0.9.0; rerun check. |")
}

func TestReleaseVersionUpdatesStableWithoutForking(t *testing.T) {
	repoRoot := t.TempDir()
	writeReleaseStatus(t, repoRoot,
		"| `unit` | `assistant` | `yes` | `no` | `stable` | `unit_fork` | test |\n"+
			"| `unit` | `agent` | `yes` | `no` | `stable` | `unit_fork` | test |\n")
	writeReleaseUnit(t, repoRoot, "stable", "assistant", "0.9.0", nil)
	writeReleaseUnit(t, repoRoot, "stable", "agent", "0.1.0", []string{"s_unit_assistant@0.8.0"})

	result, err := ReleaseVersion(repoRoot, Options{
		Unit:    "assistant",
		FromRef: "s_unit_assistant@0.8.0",
		ToRef:   "s_unit_assistant@0.9.0",
	})
	if err != nil {
		t.Fatalf("ReleaseVersion returned error: %v", err)
	}
	if got := strings.Join(result.StableUpdated, ","); got != "unit:agent" {
		t.Fatalf("expected agent stable update, got %q", got)
	}
	assertContains(t, readReleaseFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_agent.md")), "  - s_unit_assistant@0.9.0")
	assertNotExists(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_agent.md"))
	status := readReleaseFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"))
	assertContains(t, status, "| `unit` | `agent` | `yes` | `no` | `stable` | `unit_stable_verify` | Retargeted by unit release-version from s_unit_assistant@0.8.0 to s_unit_assistant@0.9.0; rerun stable verification. |")
}

func TestReleaseVersionNoopWhenOldRefIsAbsent(t *testing.T) {
	repoRoot := t.TempDir()
	writeReleaseStatus(t, repoRoot,
		"| `unit` | `assistant` | `yes` | `no` | `stable` | `unit_fork` | test |\n"+
			"| `unit` | `agent` | `yes` | `no` | `stable` | `unit_fork` | test |\n")
	writeReleaseUnit(t, repoRoot, "stable", "assistant", "0.9.0", nil)
	writeReleaseUnit(t, repoRoot, "stable", "agent", "0.1.0", []string{"s_unit_assistant@0.9.0"})

	result, err := ReleaseVersion(repoRoot, Options{
		Unit:    "assistant",
		FromRef: "s_unit_assistant@0.8.0",
		ToRef:   "s_unit_assistant@0.9.0",
	})
	if err != nil {
		t.Fatalf("ReleaseVersion returned error: %v", err)
	}
	if !result.Noop {
		t.Fatalf("expected noop result")
	}
	if len(result.MainSpecsUpdated) != 0 {
		t.Fatalf("expected no updated specs, got %v", result.MainSpecsUpdated)
	}
}

func TestReleaseVersionRejectsDifferentUnitRefs(t *testing.T) {
	repoRoot := t.TempDir()
	writeReleaseStatus(t, repoRoot, "| `unit` | `assistant` | `yes` | `no` | `stable` | `unit_fork` | test |\n")
	writeReleaseUnit(t, repoRoot, "stable", "assistant", "0.9.0", nil)

	_, err := ReleaseVersion(repoRoot, Options{
		Unit:    "assistant",
		FromRef: "s_unit_assistant@0.8.0",
		ToRef:   "s_unit_memory@0.4.1",
	})
	if err == nil || !strings.Contains(err.Error(), "unit \"assistant\"") {
		t.Fatalf("expected same-unit error, got %v", err)
	}
}

func TestReleaseVersionRejectsMissingToRef(t *testing.T) {
	repoRoot := t.TempDir()
	writeReleaseStatus(t, repoRoot, "| `unit` | `assistant` | `yes` | `no` | `stable` | `unit_fork` | test |\n")

	_, err := ReleaseVersion(repoRoot, Options{
		Unit:    "assistant",
		FromRef: "s_unit_assistant@0.8.0",
		ToRef:   "s_unit_assistant@0.9.0",
	})
	if err == nil || !strings.Contains(err.Error(), "read docs/specs/units/stable/s_unit_assistant.md") {
		t.Fatalf("expected missing to-ref error, got %v", err)
	}
}

func TestReleaseVersionRejectsDuplicateRetargetResult(t *testing.T) {
	repoRoot := t.TempDir()
	writeReleaseStatus(t, repoRoot,
		"| `unit` | `assistant` | `yes` | `no` | `stable` | `unit_fork` | test |\n"+
			"| `unit` | `agent` | `yes` | `no` | `stable` | `unit_fork` | test |\n")
	writeReleaseUnit(t, repoRoot, "stable", "assistant", "0.9.0", nil)
	writeReleaseUnit(t, repoRoot, "stable", "agent", "0.1.0", []string{"s_unit_assistant@0.8.0", "s_unit_assistant@0.9.0"})

	_, err := ReleaseVersion(repoRoot, Options{
		Unit:    "assistant",
		FromRef: "s_unit_assistant@0.8.0",
		ToRef:   "s_unit_assistant@0.9.0",
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate item") {
		t.Fatalf("expected duplicate post-update blocker, got %v", err)
	}
}

func TestValidateCurrentRefsReportsRemainingOldRef(t *testing.T) {
	repoRoot := t.TempDir()
	writeReleaseStatus(t, repoRoot,
		"| `unit` | `assistant` | `yes` | `no` | `stable` | `unit_fork` | test |\n"+
			"| `unit` | `agent` | `yes` | `no` | `stable` | `unit_fork` | test |\n")
	writeReleaseUnit(t, repoRoot, "stable", "assistant", "0.9.0", nil)
	writeReleaseUnit(t, repoRoot, "stable", "agent", "0.1.0", []string{"s_unit_assistant@0.8.0"})

	diagnostics := ValidateCurrentRefs(repoRoot, "s_unit_assistant@0.8.0")
	if len(diagnostics) != 1 || !strings.Contains(diagnostics[0], "forbidden unit ref s_unit_assistant@0.8.0 remains") {
		t.Fatalf("expected forbidden ref diagnostic, got %v", diagnostics)
	}
}

func writeReleaseStatus(t *testing.T, repoRoot, rows string) {
	t.Helper()
	writeReleaseFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), ""+
		"# Spec Status\n\n"+
		"## Formal Objects\n\n"+
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n"+
		"|---|---|---|---|---|---|---|\n"+
		rows)
}

func writeReleaseUnit(t *testing.T, repoRoot, layer, unit, version string, refs []string) {
	t.Helper()
	prefix := "s_unit_"
	dir := "stable"
	if layer == "candidate" {
		prefix = "c_unit_"
		dir = "candidate"
	}
	refBlock := "unit_refs: none"
	if len(refs) > 0 {
		refBlock = "unit_refs:\n"
		for _, ref := range refs {
			refBlock += "  - " + ref + "\n"
		}
		refBlock = strings.TrimSuffix(refBlock, "\n")
	}
	writeReleaseFile(t, filepath.Join(repoRoot, "docs/specs/units", dir, prefix+unit+".md"), strings.Join([]string{
		"---",
		"id: " + unit,
		"layer: " + layer,
		"version: " + version,
		refBlock,
		"rule_refs: none",
		"---",
		"",
		"# " + unit,
		"",
	}, "\n"))
}

func writeReleaseProcessFiles(t *testing.T, repoRoot, unit string) {
	t.Helper()
	for _, relPath := range []string{
		"docs/specs/_check_work/unit/" + unit + ".md",
		"docs/specs/_check_result/unit/" + unit + ".md",
		"docs/specs/_plans/draft/" + unit + ".md",
		"docs/specs/_plans/active/" + unit + ".md",
		"docs/specs/_verify_result/unit/" + unit + ".md",
	} {
		writeReleaseFile(t, filepath.Join(repoRoot, relPath), "# process\n")
	}
}

func writeReleaseFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readReleaseFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func assertContains(t *testing.T, content, want string) {
	t.Helper()
	if !strings.Contains(content, want) {
		t.Fatalf("expected content to contain %q\ncontent:\n%s", want, content)
	}
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected %s to be removed", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
}
