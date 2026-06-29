package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNextCLI(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Test next without unit (usage output)
	if err := runNext([]string{"--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("next failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Usage:") {
		t.Fatalf("expected usage output, got %s", output)
	}
}

func TestPromoteFailsOnMissingUnit(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runPromote([]string{"--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for missing --unit flag")
	}
	output := stderr.String()
	if !strings.Contains(output, "Usage:") {
		t.Fatalf("expected usage in stderr, got: %s", output)
	}
}

func TestPromoteFailsOnMissingSpec(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runPromote([]string{"--unit", "nonexistent", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for missing spec")
	}
	output := stdout.String()
	if !strings.Contains(output, "FAILED") {
		t.Fatalf("expected FAILED result, got %s", output)
	}
	if !strings.Contains(output, "not found") {
		t.Fatalf("expected 'not found' message, got %s", output)
	}
}

func TestPromoteWithValidSpec(t *testing.T) {
	repoRoot := createCLITestRepo(t)

	// Create a valid candidate spec
	candidateDir := filepath.Join(repoRoot, "docs/specs/units/candidate")
	err := os.MkdirAll(candidateDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	specContent := `---
id: test_unit
layer: candidate
version: 1.0.0
unit_refs: none
rule_refs: none
---

## Description

Test unit for promote testing.

## Testability / Acceptance Criteria

acceptance_item_set:
  - id: test.check
    description: Test check passes.
    verification_type: testable
    verification_surface: internal
    implementation_surface: internal
    verification_method: test
    pass_condition: passes
    not_runnable_yet: no
`
	specPath := filepath.Join(candidateDir, "c_unit_test_unit.md")
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create repository_mapping.md
	mappingDir := filepath.Join(repoRoot, "docs/specs")
	os.MkdirAll(mappingDir, 0755)
	mappingContent := `| kind | id | registration_state | implementation_paths | spec_files | responsibility |
|-----|----|-------------------|---------------------|------------|---------------|
| unit | test_unit | planned | none | docs/specs/units/candidate/c_unit_test_unit.md | Test unit |
`
	mappingPath := filepath.Join(mappingDir, "repository_mapping.md")
	if err := os.WriteFile(mappingPath, []byte(mappingContent), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err = runPromote([]string{"--unit", "test_unit", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("promote failed: %v\nstderr=%s\nstdout=%s", err, stderr.String(), stdout.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "PASSED") {
		t.Fatalf("expected PASSED result, got %s", output)
	}
	if !strings.Contains(output, "Promoted:") {
		t.Fatalf("expected promotion action, got %s", output)
	}

	// Check stable file was created
	stablePath := filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_test_unit.md")
	if _, err := os.Stat(stablePath); os.IsNotExist(err) {
		t.Fatal("stable spec was not created after promote")
	}
}

func createCLITestRepo(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "specflowctl-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	// Create minimal framework structure
	frameworkDir := filepath.Join(dir, "framework")
	os.MkdirAll(frameworkDir, 0755)
	return dir
}
