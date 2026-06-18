package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestNextCommandStateStableIdle(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | |\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_demo.md"), "# Demo\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runNext([]string{"--unit", "demo", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runNext failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()

	expected := []string{
		"TASK: Fork stable truth",
		"READS:",
		"  - docs/specs/units/stable/s_unit_demo.md",
		"WRITES:",
		"  - docs/specs/units/stable/s_unit_demo.md",
		"COMPLETION:",
		"command unit_fork",
		"forked",
	}
	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("expected %q in output, got:\n%s", s, output)
		}
	}
}

func TestNextCommandStateStableVerify(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_stable_verify` | |\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_demo.md"), "# Demo\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runNext([]string{"--unit", "demo", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runNext failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()

	expected := []string{
		"TASK: Verify current implementation matches stable truth",
		"READS:",
		"WRITES:",
		"  (none)",
		"COMPLETION:",
		"command unit_stable_verify",
		"aligned",
		"controlled_repair_required",
	}
	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("expected %q in output, got:\n%s", s, output)
		}
	}
}

func TestNextCommandStateCandidateCheck(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `yes` | `yes` | `candidate` | `unit_check` | |\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), "# Demo\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runNext([]string{"--unit", "demo", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runNext failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()

	expected := []string{
		"TASK: Verify candidate spec is clear",
		"COMPLETION:",
		"command unit_check",
		"pass",
		"blocked",
		"fix_required",
	}
	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("expected %q in output, got:\n%s", s, output)
		}
	}
}

func TestNextCommandStateCandidatePending(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `yes` | `yes` | `candidate` | `unit_verify` | `pending_impl` |\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), "# Demo\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runNext([]string{"--unit", "demo", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runNext failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()

	expected := []string{
		"TASK: Implement per candidate spec",
		"READS:",
		"WRITES:",
		"  - src/**",
		"  - tests/**",
		"BLOCKED: docs/specs/** (except repository_mapping.md), framework/**",
		"specflowctl next --unit demo",
	}
	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("expected %q in output, got:\n%s", s, output)
		}
	}
}

func TestNextCommandStateCandidateVerify(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `yes` | `yes` | `candidate` | `unit_verify` | |\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), "# Demo\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runNext([]string{"--unit", "demo", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runNext failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()

	expected := []string{
		"TASK: Verify implementation matches candidate spec",
		"COMPLETION:",
		"command unit_verify",
		"ready_to_promote",
		"spec_issue",
		"impl_issue",
	}
	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("expected %q in output, got:\n%s", s, output)
		}
	}
}

func TestNextCommandStateCandidatePromote(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `yes` | `yes` | `candidate` | `unit_promote` | |\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), "# Demo\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runNext([]string{"--unit", "demo", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runNext failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()

	expected := []string{
		"TASK: Promote verified candidate truth to stable",
		"COMPLETION:",
		"command unit_promote",
		"promoted",
	}
	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("expected %q in output, got:\n%s", s, output)
		}
	}
}

func TestNextCommandStateUnitNew(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_new` | |\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), "# Demo\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runNext([]string{"--unit", "demo", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runNext failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()

	expected := []string{
		"TASK: Create new candidate truth",
		"COMPLETION:",
		"command unit_new",
		"pass",
		"blocked",
	}
	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("expected %q in output, got:\n%s", s, output)
		}
	}
}

func TestNextCommandStateUnitInit(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `no` | `no` | `` | `unit_init` | |\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_demo.md"), "# Demo\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runNext([]string{"--unit", "demo", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runNext failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()

	expected := []string{
		"TASK: Init existing capability",
		"COMPLETION:",
		"command unit_init",
		"pass",
		"blocked",
	}
	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("expected %q in output, got:\n%s", s, output)
		}
	}
}

func TestNextCommandExplain(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `yes` | `yes` | `candidate` | `unit_check` | |\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), "# Demo\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runNext([]string{"--unit", "demo", "--explain", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runNext with --explain failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()

	expected := []string{
		"TASK: Verify candidate spec is clear",
		"=== EXPLAIN ===",
		"Lifecycle file: unit_check.md",
	}
	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("expected %q in output, got:\n%s", s, output)
		}
	}
}

func TestNextCommandExplainNoUnit(t *testing.T) {
	repoRoot := createCLITestRepo(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runNext([]string{"--explain", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for --explain without --unit, got nil")
	}
}

func TestNextCommandUnitNotFound(t *testing.T) {
	repoRoot := createCLITestRepo(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runNext([]string{"--unit", "nonexistent", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for nonexistent unit, got nil")
	}
}

func TestNextCommandListMode(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `yes` | `yes` | `candidate` | `unit_check` | |\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runNext([]string{"--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("runNext (list) failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()

	expected := []string{
		"Usage:",
		"--unit <name>",
	}
	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("expected %q in output, got:\n%s", s, output)
		}
	}
}
