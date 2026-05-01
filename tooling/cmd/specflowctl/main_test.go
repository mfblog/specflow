package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewRunInitAndValidateCLI(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := runReview([]string{"run-init", "--flow", "spec_flow_review", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("run-init failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Review run-state created:") {
		t.Fatalf("expected created output, got %s", output)
	}
	file := strings.TrimSpace(strings.TrimPrefix(output, "Review run-state created:"))

	stdout.Reset()
	stderr.Reset()
	if err := runReview([]string{"run-validate", "--flow", "spec_flow_review", "--repo-root", repoRoot, "--file", file}, &stdout, &stderr); err != nil {
		t.Fatalf("run-validate failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Review run-state is valid:") {
		t.Fatalf("expected valid output, got %s", stdout.String())
	}
}

func TestReviewRunRequiresFlowCLI(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runReview([]string{"run-init", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "flow is required") {
		t.Fatalf("expected flow required error, got err=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
}

func TestReviewCollectDefaultScopePrintsToolingRuntimeFilesCLI(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := runReview([]string{"collect-default-scope", "--flow", "spec_flow_review", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("collect-default-scope failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Tooling runtime files") {
		t.Fatalf("expected tooling runtime heading, got %s", output)
	}
	if !strings.Contains(output, "specflow/tooling/reader/web/app.js") {
		t.Fatalf("expected reader app.js in collect-default-scope output, got %s", output)
	}
}

func TestRepositoryMappingValidateCLI(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), ""+
		"# Spec Status\n\n"+
		"## Formal Objects\n\n"+
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n"+
		"|---|---|---|---|---|---|---|\n"+
		"| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | test |\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/repository_mapping.md"), ""+
		"# Repository Mapping\n\n"+
		"### 2.1 Current Units\n\n"+
		"1. `demo`\n"+
		"   - demo unit\n\n"+
		"### 4.6 Unit Truth Rules And Implementation Paths\n\n"+
		"1. `demo`\n"+
		"   - `truth_surface_rule`: `unit_default`\n"+
		"   - `implementation_surface`\n"+
		"     - current no exclusive implementation surface\n\n"+
		"### 4.7 Conflict Rules\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runRepositoryMapping([]string{"validate", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("repository-mapping validate failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Repository mapping is valid.") {
		t.Fatalf("expected valid output, got %s", stdout.String())
	}
}

func TestDesignReviewRunInitAndValidateCLI(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := runReview([]string{"run-init", "--flow", "spec_flow_design_review", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("design run-init failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Review run-state created:") || !strings.Contains(output, "spec_flow_design_review") {
		t.Fatalf("expected design created output, got %s", output)
	}
	file := strings.TrimSpace(strings.TrimPrefix(output, "Review run-state created:"))

	stdout.Reset()
	stderr.Reset()
	if err := runReview([]string{"run-validate", "--flow", "spec_flow_design_review", "--repo-root", repoRoot, "--file", file}, &stdout, &stderr); err != nil {
		t.Fatalf("design run-validate failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Review run-state is valid:") {
		t.Fatalf("expected valid output, got %s", stdout.String())
	}
}

func createCLITestRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	frameworkFiles := []string{
		"spec_flow_review.md",
		"spec_flow_design_review.md",
		"spec_flow_migrate.md",
		"agent_operability_standard.md",
		"natural_language_routing.md",
		"onboarding_decision_policy.md",
		"command_policy.md",
		"implementation_change_policy.md",
		"checkpoint_protocol.md",
		"tooling_execution_policy.md",
		"severity_policy.md",
		"scenario_policy.md",
		"spec_policy.md",
		"repository_mapping_policy.md",
		"candidate_handoff_contract.md",
		"downgrade_policy.md",
		"recovery_policy.md",
		"git_policy.md",
		"impact_sync_policy.md",
		"process_snapshot_contract.md",
		"entry_index_registry.md",
		"project_standards_policy.md",
		"project_standard_create.md",
		"shared_new.md",
		"shared_extract.md",
		"shared_bind.md",
		"shared_topology.md",
		"shared_sync.md",
		"shared_escape.md",
	}
	for _, name := range frameworkFiles {
		writeCLITestFile(t, filepath.Join(repoRoot, "specflow/framework", name), "# "+name+"\n")
	}
	for _, relPath := range []string{
		"specflow/framework/skills/using-specflow-guidance/SKILL.md",
		"specflow/framework/skills/project-framing/SKILL.md",
		"specflow/framework/skills/scope-cutting/SKILL.md",
		"specflow/framework/skills/solution-design/SKILL.md",
		"specflow/framework/skills/design-quality-review/SKILL.md",
		"specflow/framework/skills/spec-writeback-guidance/SKILL.md",
	} {
		writeCLITestFile(t, filepath.Join(repoRoot, relPath), "# skill\n")
	}
	writeCLITestFile(t, filepath.Join(repoRoot, "specflow/framework/commands/unit_check.md"), "# unit_check\n")
	for _, relPath := range []string{
		"specflow/templates/docs/specs/_status.md",
		"specflow/templates/docs/specs/_check_result/README.md",
		"specflow/templates/docs/specs/_plans/README.md",
		"specflow/templates/docs/specs/_plans/draft/README.md",
		"specflow/templates/docs/specs/_plans/active/README.md",
		"specflow/templates/docs/specs/_verify_result/README.md",
		"specflow/templates/docs/specs/_governance_review/README.md",
		"specflow/templates/docs/project_standards/_registry.md",
		"specflow/templates/AGENTS.md",
		"specflow/templates/GEMINI.md",
		"specflow/templates/CLAUDE.md",
		"AGENTS.md",
		"GEMINI.md",
		"CLAUDE.md",
		"specflow/tooling/README.md",
		"specflow/tooling/cmd/specflowctl/main.go",
		"specflow/tooling/internal/demo/demo.go",
		"specflow/tooling/go.mod",
		"specflow/tooling/manifest.tsv",
	} {
		writeCLITestFile(t, filepath.Join(repoRoot, relPath), "# "+filepath.Base(relPath)+"\n")
	}
	writeCLIReaderWebFiles(t, repoRoot)
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/project_standards/_registry.md"), ""+
		"# Registry\n\n"+
		"## Active Standards\n\n"+
		"| standard_id | type | surface | file | consumed_by | applies_to | effect | conflict_rule | notes |\n"+
		"|---|---|---|---|---|---|---|---|---|\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), ""+
		"# Spec Status\n\n"+
		"## Formal Objects\n\n"+
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n"+
		"|---|---|---|---|---|---|---|\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/repository_mapping.md"), "# Repository Mapping\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/system_constraints.md"), "# System Constraints\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), "# Demo Candidate\n")
	return repoRoot
}

func writeCLITestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func writeCLIReaderWebFiles(t *testing.T, repoRoot string) {
	t.Helper()
	writeCLITestFile(t, filepath.Join(repoRoot, "specflow/tooling/reader/web/index.html"), "<!doctype html>\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "specflow/tooling/reader/web/styles.css"), "body { color: #111; }\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "specflow/tooling/reader/web/app.js"), "console.log('demo');\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "specflow/tooling/reader/web/cytoscape.min.js"), "window.cytoscape = function() {};\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "specflow/tooling/reader/web/mermaid.min.js"), "window.mermaid = { initialize() {}, run() {} };\n")
}
