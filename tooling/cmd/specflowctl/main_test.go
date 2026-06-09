package main
import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/testfixtures"
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
	if !strings.Contains(filepath.ToSlash(output), "docs/specs/_governance_review/spec_flow_review.md") {
		t.Fatalf("expected fixed spec_flow_review run-state path, got %s", output)
	}
	stdout.Reset()
	stderr.Reset()
	if err := runReview([]string{"run-validate", "--flow", "spec_flow_review", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("run-validate failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Review run-state is valid:") {
		t.Fatalf("expected valid output, got %s", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	if err := runReview([]string{"run-refresh", "--flow", "spec_flow_review", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("run-refresh failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Review run-state refreshed:") {
		t.Fatalf("expected refreshed output, got %s", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	if err := runReview([]string{"run-touch", "--flow", "spec_flow_review", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("run-touch failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Review run-state touched:") {
		t.Fatalf("expected touched output, got %s", stdout.String())
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
func TestProcessCheckWorkCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepo(t)
	for _, relPath := range []string{
		"specflow/framework/core/object_model.md",
		"specflow/framework/core/repository_mapping.md",
		"specflow/framework/lifecycle/unit_check.md",
		"specflow/framework/operations/implementation_change.md",
		"specflow/framework/process_snapshot_contract.md",
		"specflow/framework/candidate_handoff_contract.md",
		"specflow/framework/candidate_intent.md",
	} {
		writeCLITestFile(t, filepath.Join(repoRoot, filepath.FromSlash(relPath)), "# "+filepath.Base(relPath)+"\n")
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runProcess([]string{"check-work-init", "--object-type", "unit", "--object", "demo", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("check-work-init failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Check checklist created:") {
		t.Fatalf("expected created output, got %s", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	if err := runProcess([]string{"check-work-validate", "--object-type", "unit", "--object", "demo", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("check-work-validate failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Check checklist is valid:") {
		t.Fatalf("expected valid output, got %s", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	if err := runProcess([]string{"check-work-refresh", "--object-type", "unit", "--object", "demo", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("check-work-refresh failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Check checklist refreshed:") {
		t.Fatalf("expected refreshed output, got %s", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()
	if err := runProcess([]string{"check-work-touch", "--object-type", "unit", "--object", "demo", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("check-work-touch failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Check checklist touched:") {
		t.Fatalf("expected touched output, got %s", stdout.String())
	}
}
func TestUnitReleaseVersionCLI(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, ""+
		"| `unit` | `assistant` | `yes` | `no` | `stable` | `unit_fork` | test |\n"+
		"| `unit` | `agent` | `yes` | `no` | `stable` | `unit_fork` | test |\n")
	writeCLIUnitReleaseSpec(t, repoRoot, "stable", "assistant", "0.9.0", nil)
	writeCLIUnitReleaseSpec(t, repoRoot, "stable", "agent", "0.1.0", []string{"s_unit_assistant@0.8.0"})
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/_stable_verify_result/unit/agent.md"), "# stable verify\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runUnit([]string{"release-version", "--unit", "assistant", "--from-ref", "s_unit_assistant@0.8.0", "--to-ref", "s_unit_assistant@0.9.0", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("unit release-version failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Released unit version: assistant from s_unit_assistant@0.8.0 to s_unit_assistant@0.9.0") {
		t.Fatalf("expected release output, got %s", output)
	}
	if !strings.Contains(output, "Stable current-layer units rerouted") {
		t.Fatalf("expected stable reroute heading, got %s", output)
	}
	if !strings.Contains(output, "- unit:agent") {
		t.Fatalf("expected rerouted unit output, got %s", output)
	}
	content := mustReadCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_agent.md"))
	if !strings.Contains(content, "  - s_unit_assistant@0.8.0") || strings.Contains(content, "  - s_unit_assistant@0.9.0") {
		t.Fatalf("stable unit_refs must not be rewritten, got %s", content)
	}
	assertCLITestFileNotExists(t, filepath.Join(repoRoot, "docs/specs/_stable_verify_result/unit/agent.md"))
}
func TestUnitReleaseVersionCLIRequiresInputs(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runUnit([]string{"release-version", "--unit", "assistant", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "unit, from-ref, and to-ref are required") {
		t.Fatalf("expected required input error, got err=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
}
func TestRuleSyncImpactDeletedRuleRefsCLI(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | test |\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_g_rule_repository_baseline.md"), strings.Join([]string{
		"---",
		"rule_id: g_rule_repository_baseline",
		"rule_scope: global",
		"layer: stable",
		"rule_version: 1.0.0",
		"---",
		"",
		"# Global Rules",
		"",
	}, "\n"))
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"rule_refs: none",
		"---",
		"",
		"# Demo",
		"",
	}, "\n"))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runRule([]string{"sync-impact", "--deleted-rule-refs", "c_b_rule_demo@0.1.0", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("rule sync-impact failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Deleted rule refs verified no-impact") || !strings.Contains(output, "- c_b_rule_demo@0.1.0") {
		t.Fatalf("expected deleted-rule no-impact output, got %s", output)
	}
	if !strings.Contains(output, "Unit results (0):") {
		t.Fatalf("expected no unit results, got %s", output)
	}
}
func TestReviewCollectDefaultScopePrintsToolingScriptAndRuntimeFilesCLI(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runReview([]string{"collect-default-scope", "--flow", "spec_flow_review", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("collect-default-scope failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Tooling script files") {
		t.Fatalf("expected tooling script heading, got %s", output)
	}
	if !strings.Contains(output, "Candidate intent files") {
		t.Fatalf("expected candidate intent heading, got %s", output)
	}
	if !strings.Contains(output, "specflow/framework/candidate_intent.md") {
		t.Fatalf("expected repair intent standard in collect-default-scope output, got %s", output)
	}
	if !strings.Contains(output, "specflow/tooling/scripts/tooling_fingerprint.sh") {
		t.Fatalf("expected shell fingerprint script in collect-default-scope output, got %s", output)
	}
	if !strings.Contains(output, "specflow/tooling/scripts/tooling_fingerprint.ps1") {
		t.Fatalf("expected PowerShell fingerprint script in collect-default-scope output, got %s", output)
	}
	if !strings.Contains(output, "specflow/tooling/scripts/build_release.sh") {
		t.Fatalf("expected build release script in collect-default-scope output, got %s", output)
	}
	for _, relPath := range currentCLIToolingScriptFiles() {
		if !strings.Contains(output, relPath) {
			t.Fatalf("expected tooling script in collect-default-scope output: %s, got %s", relPath, output)
		}
	}
	if !strings.Contains(output, "Tooling runtime files") {
		t.Fatalf("expected tooling runtime heading, got %s", output)
	}
	if !strings.Contains(output, "specflow/tooling/reader/web/app.js") {
		t.Fatalf("expected reader app.js in collect-default-scope output, got %s", output)
	}
	if !strings.Contains(output, "Template project-instance files") {
		t.Fatalf("expected template project-instance heading, got %s", output)
	}
	if !strings.Contains(output, "specflow/templates/docs/specs/repository_mapping.md") {
		t.Fatalf("expected repository mapping template in collect-default-scope output, got %s", output)
	}
	if !strings.Contains(output, "specflow/templates/docs/specs/rules/stable/s_g_rule_repository_baseline.md") {
		t.Fatalf("expected global rule template in collect-default-scope output, got %s", output)
	}
}
func TestReviewCollectDefaultDesignScopePrintsCandidateIntentFilesCLI(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runReview([]string{"collect-default-scope", "--flow", "spec_flow_design_review", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("collect-default-scope failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Candidate intent files") {
		t.Fatalf("expected candidate intent heading, got %s", output)
	}
	if !strings.Contains(output, "specflow/framework/candidate_intent.md") {
		t.Fatalf("expected candidate intent policy in design collect-default-scope output, got %s", output)
	}
	if !strings.Contains(output, "specflow/framework/candidate_intent.md") {
		t.Fatalf("expected repair intent standard in design collect-default-scope output, got %s", output)
	}
	if !strings.Contains(output, "specflow/framework/candidate_intent.md") {
		t.Fatalf("expected change intent standard in design collect-default-scope output, got %s", output)
	}

}
func TestReviewCollectDefaultScopePrintsSourceLayoutCLI(t *testing.T) {
	repoRoot := createCLISourceReviewRepo(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runReview([]string{"collect-default-scope", "--flow", "spec_flow_review", "--layout", "source", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("source collect-default-scope failed: %v\nstderr=%s", err, stderr.String())
	}
	output := stdout.String()
	for _, expected := range []string{
		"Review layout: source_repo",
		"Framework root: framework",
		"Project-instance compatibility mode: template_bootstrap",
		"framework/lifecycle/unit_check.md",
		"templates/docs/specs/_status.md",
		"Source repo entry example files (1):",
		"example.md",
		"Project entry files (0):",
		"tooling/go.mod",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in source scope output, got %s", expected, output)
		}
	}
	if strings.Contains(output, "specflow/framework/lifecycle/unit_check.md") {
		t.Fatalf("source scope output must not use installed framework paths, got %s", output)
	}
}
func TestReviewScopeAliasIsNotSupportedCLI(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runReview([]string{"scope"}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected review scope alias to be rejected")
	}
	if !strings.Contains(err.Error(), `unknown review subcommand "scope"`) {
		t.Fatalf("expected unknown subcommand error, got %v", err)
	}
	if strings.Contains(stderr.String(), "specflowctl review scope") {
		t.Fatalf("review usage must not advertise removed scope alias, got %s", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout for rejected alias, got %s", stdout.String())
	}
}
func TestEntryCheckSourceRepoUsesTemplateEntriesCLI(t *testing.T) {
	repoRoot := createCLISourceReviewRepo(t)
	writeCLITestFile(t, filepath.Join(repoRoot, "framework/operations/entry_routing.md"), `# Entry Routing
## Entry File Registration
Registered entry index files: `+"`AGENTS.md`, `GEMINI.md`, `CLAUDE.md`"+`.
`)
	block := "==SPECFLOW:BEGIN==\nmanaged\n==SPECFLOW:END==\n"
	for _, name := range []string{"AGENTS.md", "GEMINI.md", "CLAUDE.md"} {
		writeCLITestFile(t, filepath.Join(repoRoot, "templates", name), block)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runEntry([]string{"check", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("source entry check failed: %v\nstderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Managed entry blocks are already consistent.") {
		t.Fatalf("expected source entry check success, got stdout=%s stderr=%s", stdout.String(), stderr.String())
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
		"## 2. Object Registry\n\n"+
		"| kind | id | registration_state | implementation_paths | spec_files | responsibility |\n"+
		"|---|---|---|---|---|---|\n"+
		"| unit | demo | planned | none | `docs/specs/units/candidate/c_unit_demo.md` | demo unit |\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), "# Demo\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runRepositoryMapping([]string{"validate", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("repository-mapping validate failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Repository mapping is valid.") {
		t.Fatalf("expected valid output, got %s", stdout.String())
	}
}
func TestRelationCandidatesCLI(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIRelationRepo(t, repoRoot, false)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runRelation([]string{"candidates", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("relation candidates failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	output := stdout.String()
	for _, expected := range []string{
		"relation_result: pass",
		"ready_candidates (1):",
		"- beta",
		"blocked_candidates (1):",
		"- alpha | blocked_by=unit:beta",
		"candidate_cycles (0):",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, output)
		}
	}
}
func TestRelationCandidatePreflightCLIBlocksWhenUpstreamCandidateIsNotStable(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIRelationRepo(t, repoRoot, false)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runRelation([]string{"candidate-preflight", "--object", "alpha", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "candidate relation preflight failed") {
		t.Fatalf("expected preflight failure, got err=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
	output := stdout.String()
	for _, expected := range []string{
		"relation_result: fail",
		"object: alpha",
		"may_continue: false",
		"blocked_by (1):",
		"- unit:beta",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, output)
		}
	}
}
func TestRelationCandidatesCLIReportsCycles(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIRelationRepo(t, repoRoot, true)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runRelation([]string{"candidates", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("relation candidates should report cycles without command error: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	output := stdout.String()
	for _, expected := range []string{
		"relation_result: fail",
		"candidate_cycles (1):",
		"alpha -> beta -> alpha",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, output)
		}
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
	if !strings.Contains(filepath.ToSlash(output), "docs/specs/_governance_review/spec_flow_design_review.md") {
		t.Fatalf("expected fixed spec_flow_design_review run-state path, got %s", output)
	}
	stdout.Reset()
	stderr.Reset()
	if err := runReview([]string{"run-validate", "--flow", "spec_flow_design_review", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("design run-validate failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Review run-state is valid:") {
		t.Fatalf("expected valid output, got %s", stdout.String())
	}
}
func writeCLIRelationRepo(t *testing.T, repoRoot string, cycle bool) {
	t.Helper()
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), ""+
		"# Spec Status\n\n"+
		"## Formal Objects\n\n"+
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n"+
		"|---|---|---|---|---|---|---|\n"+
		"| `unit` | `alpha` | `yes` | `yes` | `candidate` | `unit_check` | alpha |\n"+
		"| `unit` | `beta` | `yes` | `yes` | `candidate` | `unit_check` | beta |\n")
	betaBody := "No candidate dependencies."
	if cycle {
		betaBody = "`c_unit_alpha@0.1.0`"
	}
	writeCLICandidateUnit(t, repoRoot, "alpha", "0.1.0", "`c_unit_beta@0.1.0`")
	writeCLICandidateUnit(t, repoRoot, "beta", "0.1.0", betaBody)
}
func writeCLICandidateUnit(t *testing.T, repoRoot, object, version, body string) {
	t.Helper()
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_"+object+".md"), strings.Join([]string{
		"---",
		"id: " + object,
		"layer: candidate",
		"version: " + version,
		"evidence_appendix_ref: none",
		"unit_refs: none",
		"rule_refs: none",
		"---",
		"",
		"# " + object,
		"",
		body,
	}, "\n")+"\n")
}
func TestSnapshotValidateProcessUsesObjectFlagsCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepo(t)
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/demo.md"), "# plan\n\n```yaml\n"+strings.Join([]string{
		"spec_file_ref: " + expected.SpecFileRef,
		"spec_version_ref: " + expected.SpecVersionRef,
		"spec_fingerprint: " + expected.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + expected.AcceptanceBehaviorFingerprint,
		"stable_candidate_diff_refs: none",
		"implementation_gap_refs: docs/specs/repository_mapping.md",
		"unit_appendix_snapshot: none",
		"unit_snapshot: none",
		"rule_snapshot: none",
		"acceptance_item_plan_coverage:",
		"  - id: demo.core",
		"    coverage: implementation slice and verification target",
		"retirement_targets: none",
		"planned_change_scope:",
		"  - id: pcs.core",
		"    basis_refs: " + expected.SpecFileRef,
		"    acceptance_item_ids: demo.core",
		"    implementation_refs: docs/specs/repository_mapping.md",
		"    verification_action: verify package-aware delta",
		"package_constraint_review: pass",
		"package_constraint_refs: " + expected.SpecFileRef,
		"package_constraint_summary: current package constraints reviewed for this delta",
		"evaluation_mode: independent",
		"reviewer_result: pass",
		"reviewer_context: minimal_context",
		"review_input_refs: " + cliReviewInputRefsForTest(expected.Object, "unit_verify_ready_to_promote", expected.SpecFileRef),
		"review_findings: none",
		"human_decision_refs: none",
	}, "\n")+"\n```\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runSnapshot([]string{"validate-process", "--object-type", "unit", "--object", "demo", "--process", "plan", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("validate-process failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Process snapshot is valid.") {
		t.Fatalf("expected valid output, got %s", stdout.String())
	}
}
func TestEvaluationRequestRejectsUnsupportedPackCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepo(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runEvaluation([]string{"request", "--object-type", "unit", "--object", "demo", "--pack", "unknown_pack", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "unsupported independent evaluation pack") {
		t.Fatalf("expected unsupported pack error, got err=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
}
func TestEvaluationRequestRejectsUnsupportedObjectTypeCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepo(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runEvaluation([]string{"request", "--object-type", "scenario", "--object", "demo_flow", "--pack", "unit_verify_ready_to_promote", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "object type \"scenario\" is not supported; only unit is supported") {
		t.Fatalf("expected unsupported object type error, got err=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
}
func TestEvaluationRequestCreatesFreshnessTextDriftHandoffCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepo(t)
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCLIUnitCheckProcess(t, repoRoot, expected)
	candidatePath := filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md")
	content, err := os.ReadFile(candidatePath)
	if err != nil {
		t.Fatalf("read candidate: %v", err)
	}
	writeCLITestFile(t, candidatePath, strings.Replace(string(content), "# Demo\n", "# Demo\n\nEditorial note only.\n", 1))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runEvaluation([]string{"request", "--object-type", "unit", "--object", "demo", "--pack", "freshness_text_drift_reuse", "--process", "check", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("freshness evaluation request failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "freshness_impact: text_drift") || !strings.Contains(output, "evidence_reuse: pending_review") {
		t.Fatalf("expected pending text drift output, got %s", output)
	}
	requestFile := filepath.Join(repoRoot, "docs/specs/_independent_evaluation/requests/unit/demo/freshness_text_drift_reuse.md")
	requestContent, err := os.ReadFile(requestFile)
	if err != nil {
		t.Fatalf("read freshness request: %v", err)
	}
	if !strings.Contains(string(requestContent), "Is the change only wording, formatting, or clarification") {
		t.Fatalf("freshness request missing review question:\n%s", string(requestContent))
	}
	for _, phrase := range []string{
		"## Reviewer Output",
		"pass | blocked | needs_human_decision",
		"## Executor Receipt After Pass",
		"freshness_impact: text_drift",
		"evidence_reuse: accepted",
		"freshness_review_mode: independent",
		"freshness_reviewer_result: pass",
		"freshness_review_input_refs: freshness_text_drift_reuse;docs/specs/_independent_evaluation/requests/unit/demo/freshness_text_drift_reuse.md",
		"freshness_review_findings: none",
	} {
		if !strings.Contains(string(requestContent), phrase) {
			t.Fatalf("freshness request missing %q:\n%s", phrase, string(requestContent))
		}
	}
	for _, forbidden := range []string{
		"\nevaluation_mode: independent\n",
		"\nreview_input_refs: freshness_text_drift_reuse",
	} {
		if strings.Contains(string(requestContent), forbidden) {
			t.Fatalf("freshness request must not contain ordinary receipt field %q:\n%s", forbidden, string(requestContent))
		}
	}
}
func TestEvaluationRequestRejectsFreshnessSemanticDriftCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepo(t)
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCLIUnitCheckProcess(t, repoRoot, expected)
	candidatePath := filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md")
	content, err := os.ReadFile(candidatePath)
	if err != nil {
		t.Fatalf("read candidate: %v", err)
	}
	writeCLITestFile(t, candidatePath, strings.Replace(string(content), "The demo behavior passes under the declared checks.", "The demo behavior passes with a different condition.", 1))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runEvaluation([]string{"request", "--object-type", "unit", "--object", "demo", "--pack", "freshness_text_drift_reuse", "--process", "check", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "freshness review request requires freshness_impact=text_drift") {
		t.Fatalf("expected semantic drift rejection, got err=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
}
func TestSnapshotValidateProcessAcceptsStableVerifyCLI(t *testing.T) {
	repoRoot := createCLIStableSnapshotRepo(t)
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCLIStableVerifyProcess(t, repoRoot, expected, "aligned")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runSnapshot([]string{"validate-process", "--object-type", "unit", "--object", "demo", "--process", "stable_verify", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("validate-process failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Process snapshot is valid.") {
		t.Fatalf("expected valid output, got %s", stdout.String())
	}
}
func TestSnapshotValidateProcessRejectsScenarioPlanCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepo(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runSnapshot([]string{"validate-process", "--object-type", "scenario", "--object", "demo_flow", "--process", "plan", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "object type \"scenario\" is not supported; only unit is supported") {
		t.Fatalf("expected scenario rejection, got err=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
}
func TestCommandPreflightUsesNormalizedSnapshotFingerprintCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepoWithStatus(t, "unit_verify")
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCLIUnitCheckProcess(t, repoRoot, expected)
	specPath := filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md")
	content, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	crlfContent := strings.ReplaceAll(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n", "\r\n")
	if err := os.WriteFile(specPath, []byte(crlfContent), 0o644); err != nil {
		t.Fatalf("rewrite spec with CRLF: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runCommand([]string{"preflight", "--command", "unit_verify", "--object-type", "unit", "--object", "demo", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("preflight should pass with normalized line endings: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "preflight_result: pass") {
		t.Fatalf("expected pass output, got %s", stdout.String())
	}
}
func TestCommandPreflightDoesNotValidateRuleBindingsCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepoWithStatus(t, "unit_check")
	specPath := filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md")
	content, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	updated := strings.Replace(string(content), "version: 0.1.0\n---", "version: 0.1.0\nrule_refs:\n  - c_b_rule_missing@0.1.0\n---", 1)
	updated = strings.Replace(updated, "\nrule_refs: none\n", "\n", 1)
	if err := os.WriteFile(specPath, []byte(updated), 0o644); err != nil {
		t.Fatalf("rewrite spec with unresolved rule ref: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runCommand([]string{"preflight", "--command", "unit_check", "--object-type", "unit", "--object", "demo", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("preflight should not validate rule bindings: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	output := stdout.String()
	for _, expectedText := range []string{
		"preflight_result: pass",
		"may_continue: true",
		"failure_layer: none",
	} {
		if !strings.Contains(output, expectedText) {
			t.Fatalf("expected %q in output, got %s", expectedText, output)
		}
	}
	for _, forbiddenText := range []string{
		"failure_layer: truth_layer",
		"read rule docs/specs/rules/candidate/c_b_rule_missing.md",
	} {
		if strings.Contains(output, forbiddenText) {
			t.Fatalf("preflight must not report rule binding validation %q, got %s", forbiddenText, output)
		}
	}
}
func TestCommandCloseUnitStableVerifyControlledRepairDryRunAndApplyCLI(t *testing.T) {
	repoRoot := createCLIStableSnapshotRepo(t)
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCLIStableVerifyProcess(t, repoRoot, expected, "controlled_repair_required")
	statusPath := filepath.Join(repoRoot, "docs/specs/_status.md")
	before, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runCommand([]string{"close", "--command", "unit_stable_verify", "--object-type", "unit", "--object", "demo", "--outcome", "controlled_repair_required", "--candidate-intent", "repair", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("dry-run close failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	output := stdout.String()
	for _, expectedText := range []string{
		"command_close_result: dry_run",
		"cleanup_action: none",
		"status_after:",
		"  next_command: unit_fork",
	} {
		if !strings.Contains(output, expectedText) {
			t.Fatalf("expected %q in output, got %s", expectedText, output)
		}
	}
	afterDryRun, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status after dry-run: %v", err)
	}
	if string(afterDryRun) != string(before) {
		t.Fatalf("dry-run must not rewrite status\nbefore=%s\nafter=%s", string(before), string(afterDryRun))
	}
	stdout.Reset()
	stderr.Reset()
	err = runCommand([]string{"close", "--command", "unit_stable_verify", "--object-type", "unit", "--object", "demo", "--outcome", "controlled_repair_required", "--candidate-intent", "repair", "--apply", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("apply close failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "command_close_result: applied") {
		t.Fatalf("expected applied output, got %s", stdout.String())
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.NextCommand != "unit_fork" {
		t.Fatalf("expected unit_fork after apply, got %+v", status)
	}
}
func TestCommandCloseUnitStableVerifyControlledRepairRequiresIntentCLI(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_stable_verify` | test |\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runCommand([]string{"close", "--command", "unit_stable_verify", "--object-type", "unit", "--object", "demo", "--outcome", "controlled_repair_required", "--apply", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "requires --candidate-intent repair") {
		t.Fatalf("expected missing candidate-intent failure, got err=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.NextCommand != "unit_stable_verify" {
		t.Fatalf("missing intent must not advance status, got %+v", status)
	}
}
func TestCommandCloseUnitStableVerifyControlledRepairRejectsWrongIntentCLI(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_stable_verify` | test |\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runCommand([]string{"close", "--command", "unit_stable_verify", "--object-type", "unit", "--object", "demo", "--outcome", "controlled_repair_required", "--candidate-intent", "change", "--apply", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "requires --candidate-intent repair") {
		t.Fatalf("expected wrong candidate-intent failure, got err=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.NextCommand != "unit_stable_verify" {
		t.Fatalf("wrong intent must not advance status, got %+v", status)
	}
}
func TestCommandCloseUnitCheckPassRejectsInvalidGateCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepoWithStatus(t, "unit_check")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runCommand([]string{"close", "--command", "unit_check", "--object-type", "unit", "--object", "demo", "--outcome", "pass", "--apply", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "read docs/specs/_check_result/unit/demo.md") {
		t.Fatalf("expected invalid check gate failure, got err=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "validation_action: validate_process:check") {
		t.Fatalf("expected validation action output, got %s", stdout.String())
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.NextCommand != "unit_check" {
		t.Fatalf("invalid gate must not advance status, got %+v", status)
	}
}
func TestCommandCloseUnitVerifyReadyToPromoteRejectsInvalidVerifyCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepoWithStatus(t, "unit_verify")
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCLIUnitCheckProcess(t, repoRoot, expected)
	writeCLIUnitPlanProcess(t, repoRoot, expected)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runCommand([]string{"close", "--command", "unit_verify", "--object-type", "unit", "--object", "demo", "--outcome", "ready_to_promote", "--apply", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "read docs/specs/_verify_result/unit/demo.md") {
		t.Fatalf("expected invalid verify result failure, got err=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "validation_action: validate_process:verify") {
		t.Fatalf("expected verify validation action output, got %s", stdout.String())
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("LookupObjectStatus: %v", err)
	}
	if status.NextCommand != "unit_verify" {
		t.Fatalf("invalid verify result must not advance status, got %+v", status)
	}
}
func TestCommandCloseUnitPromoteDryRunKeepsCandidateAndProcessFilesCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepoWithStatus(t, "unit_promote")
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCLIUnitPlanProcess(t, repoRoot, expected)
	writeCLIUnitVerifyProcess(t, repoRoot, expected)
	statusPath := filepath.Join(repoRoot, "docs/specs/_status.md")
	candidatePath := filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md")
	verifyPath := filepath.Join(repoRoot, "docs/specs/_verify_result/unit/demo.md")
	beforeStatus := mustReadCLITestFile(t, statusPath)
	beforeCandidate := mustReadCLITestFile(t, candidatePath)
	beforeVerify := mustReadCLITestFile(t, verifyPath)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runCommand([]string{"close", "--command", "unit_promote", "--object-type", "unit", "--object", "demo", "--outcome", "promoted", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("dry-run close failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	output := stdout.String()
	for _, expectedText := range []string{
		"command_close_result: dry_run",
		"input_validation_action: command_preflight",
		"input_validated_processes:",
		"- process: verify",
		"result: valid",
		"validation_action: validate_process:verify",
		"cleanup_action: success:unit_promote",
		"  candidate: no",
		"  next_command: unit_fork",
	} {
		if !strings.Contains(output, expectedText) {
			t.Fatalf("expected %q in output, got %s", expectedText, output)
		}
	}
	if afterStatus := mustReadCLITestFile(t, statusPath); afterStatus != beforeStatus {
		t.Fatalf("dry-run must not rewrite status\nbefore=%s\nafter=%s", beforeStatus, afterStatus)
	}
	if afterCandidate := mustReadCLITestFile(t, candidatePath); afterCandidate != beforeCandidate {
		t.Fatalf("dry-run must not rewrite candidate\nbefore=%s\nafter=%s", beforeCandidate, afterCandidate)
	}
	if afterVerify := mustReadCLITestFile(t, verifyPath); afterVerify != beforeVerify {
		t.Fatalf("dry-run must not rewrite verify result\nbefore=%s\nafter=%s", beforeVerify, afterVerify)
	}
}
func TestCommandCloseRejectsScenarioObjectTypeCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepoWithStatus(t, "unit_verify")
	writeCLIStatusRows(t, repoRoot, ""+
		"| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_impl` | test |\n"+
		"| `scenario` | `demo_flow` | `no` | `yes` | `candidate` | `scenario_promote` | test |\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runCommand([]string{"close", "--command", "scenario_promote", "--object-type", "scenario", "--object", "demo_flow", "--outcome", "dependency_not_ready", "--apply", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "object-type must be unit") {
		t.Fatalf("expected scenario object type rejection, got err=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
}
func TestCommandCloseDryRunDoesNotModifyTrackedLikeFilesCLI(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | test |\n")
	checkPath := filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md")
	writeCLITestFile(t, checkPath, "check must stay\n")
	statusPath := filepath.Join(repoRoot, "docs/specs/_status.md")
	beforeStatus, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	beforeCheck, err := os.ReadFile(checkPath)
	if err != nil {
		t.Fatalf("read check: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runCommand([]string{"close", "--command", "unit_fork", "--object-type", "unit", "--object", "demo", "--outcome", "candidate_created", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("dry-run close failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	afterStatus, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status after dry-run: %v", err)
	}
	afterCheck, err := os.ReadFile(checkPath)
	if err != nil {
		t.Fatalf("read check after dry-run: %v", err)
	}
	if string(afterStatus) != string(beforeStatus) || string(afterCheck) != string(beforeCheck) {
		t.Fatalf("dry-run must not modify files\nstatus before=%s\nstatus after=%s\ncheck before=%s\ncheck after=%s", string(beforeStatus), string(afterStatus), string(beforeCheck), string(afterCheck))
	}
}
func createCLITestRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	frameworkFiles := []string{
		"advance_policy.md",
		"spec_flow_review.md",
		"spec_flow_design_review.md",
		"agent_operability_standard.md",
		"onboarding_decision_policy.md",
		"candidate_intent.md",
		"tooling_execution_policy.md",
		"slice_work_state_protocol.md",
		"severity_policy.md",
		"spec_policy.md",
		"spec_writing_guide.md",
		"spec_authoring_baseline.md",
		"candidate_handoff_contract.md",
		"downgrade_policy.md",
		"process_snapshot_contract.md",
		"entry_index_registry.md",
	}
	for _, name := range frameworkFiles {
		writeCLITestFile(t, filepath.Join(repoRoot, "specflow/framework", name), "# "+name+"\n")
	}
	for _, relPath := range []string{
		"specflow/framework/core/adoption_modes.md",
		"specflow/framework/core/context_card.md",
		"specflow/framework/core/freshness.md",
		"specflow/framework/core/independent_evaluation.md",
		"specflow/framework/core/object_model.md",
		"specflow/framework/core/status.md",
		"specflow/framework/core/repository_mapping.md",
		"specflow/framework/core/lifecycle_authority.md",
		"specflow/framework/lifecycle/overview.md",
		"specflow/framework/lifecycle/unit_init_new_fork.md",
		"specflow/framework/lifecycle/unit_check.md",
		"specflow/framework/lifecycle/unit_verify.md",
		"specflow/framework/lifecycle/unit_promote.md",
		"specflow/framework/lifecycle/unit_stable_verify.md",
		"specflow/framework/lifecycle/recovery.md",
		"specflow/framework/governance/rule_system.md",
		"specflow/framework/governance/impact_sync.md",
		"specflow/framework/governance/review.md",
		"specflow/framework/governance/review_scope.md",
		"specflow/framework/governance/rules/rule_new.md",
		"specflow/framework/governance/rules/rule_extract.md",
		"specflow/framework/governance/rules/rule_bind.md",
		"specflow/framework/governance/rules/rule_topology.md",
		"specflow/framework/governance/rules/rule_sync.md",
		"specflow/framework/governance/rules/rule_escape.md",
		"specflow/framework/operations/entry_routing.md",
		"specflow/framework/operations/implementation_change.md",
		"specflow/framework/operations/output_standard.md",
		"specflow/framework/operations/migration.md",
	} {
		writeCLITestFile(t, filepath.Join(repoRoot, relPath), "# "+filepath.Base(relPath)+"\n")
	}
	for _, relPath := range []string{
		"specflow/framework/candidate_intent.md",
		"specflow/framework/candidate_intent.md",
	} {
		writeCLITestFile(t, filepath.Join(repoRoot, relPath), "# "+filepath.Base(relPath)+"\n")
	}
	for _, relPath := range []string{
		"specflow/framework/guidance/using-specflow-guidance/SKILL.md",
		"specflow/framework/guidance/project-framing/SKILL.md",
		"specflow/framework/guidance/scope-cutting/SKILL.md",
		"specflow/framework/guidance/solution-design/SKILL.md",
		"specflow/framework/guidance/design-quality-review/SKILL.md",
		"specflow/framework/guidance/spec-writeback-guidance/SKILL.md",
	} {
		writeCLITestFile(t, filepath.Join(repoRoot, relPath), "# skill\n")
	}
	for _, relPath := range append([]string{
		"specflow/templates/docs/specs/_status.md",
		"specflow/templates/docs/specs/_check_work/README.md",
		"specflow/templates/docs/specs/_check_result/README.md",
		"specflow/templates/docs/specs/_plans/README.md",
		"specflow/templates/docs/specs/_plans/draft/README.md",
		"specflow/templates/docs/specs/_plans/active/README.md",
		"specflow/templates/docs/specs/_verify_result/README.md",
		"specflow/templates/docs/specs/_stable_verify_result/README.md",
		"specflow/templates/docs/specs/_governance_review/README.md",
		"specflow/templates/docs/specs/_independent_evaluation/README.md",
		"specflow/templates/docs/specs/repository_mapping.md",
		"specflow/templates/docs/specs/rules/stable/s_g_rule_repository_baseline.md",
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
	}, currentCLIToolingScriptFiles()...) {
		writeCLITestFile(t, filepath.Join(repoRoot, relPath), "# "+filepath.Base(relPath)+"\n")
	}
	writeCLIReaderWebFiles(t, repoRoot)
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), ""+
		"# Spec Status\n\n"+
		"## Formal Objects\n\n"+
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n"+
		"|---|---|---|---|---|---|---|\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/repository_mapping.md"), "# Repository Mapping\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_g_rule_repository_baseline.md"), "# Global Rules\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), "# Demo Candidate\n")
	return repoRoot
}
func createCLISourceReviewRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	for _, name := range []string{
		"advance_policy.md",
		"spec_flow_review.md",
		"spec_flow_design_review.md",
		"agent_operability_standard.md",
		"onboarding_decision_policy.md",
		"candidate_intent.md",
		"tooling_execution_policy.md",
		"slice_work_state_protocol.md",
		"severity_policy.md",
		"spec_policy.md",
		"spec_writing_guide.md",
		"spec_authoring_baseline.md",
		"candidate_handoff_contract.md",
		"downgrade_policy.md",
		"process_snapshot_contract.md",
		"entry_index_registry.md",
	} {
		writeCLITestFile(t, filepath.Join(repoRoot, "framework", name), "# "+name+"\n")
	}
	for _, relPath := range []string{
		"framework/core/adoption_modes.md",
		"framework/core/context_card.md",
		"framework/core/freshness.md",
		"framework/core/independent_evaluation.md",
		"framework/core/object_model.md",
		"framework/core/status.md",
		"framework/core/repository_mapping.md",
		"framework/core/lifecycle_authority.md",
		"framework/lifecycle/overview.md",
		"framework/lifecycle/unit_init_new_fork.md",
		"framework/lifecycle/unit_check.md",
		"framework/lifecycle/unit_verify.md",
		"framework/lifecycle/unit_promote.md",
		"framework/lifecycle/unit_stable_verify.md",
		"framework/lifecycle/recovery.md",
		"framework/governance/rule_system.md",
		"framework/governance/impact_sync.md",
		"framework/governance/review.md",
		"framework/governance/review_scope.md",
		"framework/governance/rules/rule_new.md",
		"framework/governance/rules/rule_extract.md",
		"framework/governance/rules/rule_bind.md",
		"framework/governance/rules/rule_topology.md",
		"framework/governance/rules/rule_sync.md",
		"framework/governance/rules/rule_escape.md",
		"framework/operations/entry_routing.md",
		"framework/operations/implementation_change.md",
		"framework/operations/output_standard.md",
		"framework/operations/migration.md",
		"framework/candidate_intent.md",
		"framework/candidate_intent.md",
		"framework/guidance/using-specflow-guidance/SKILL.md",
		"framework/guidance/project-framing/SKILL.md",
		"framework/guidance/scope-cutting/SKILL.md",
		"framework/guidance/solution-design/SKILL.md",
		"framework/guidance/design-quality-review/SKILL.md",
		"framework/guidance/spec-writeback-guidance/SKILL.md",
		"templates/docs/specs/_status.md",
		"templates/docs/specs/_check_work/README.md",
		"templates/docs/specs/_check_result/README.md",
		"templates/docs/specs/_plans/README.md",
		"templates/docs/specs/_plans/draft/README.md",
		"templates/docs/specs/_plans/active/README.md",
		"templates/docs/specs/_verify_result/README.md",
		"templates/docs/specs/_stable_verify_result/README.md",
		"templates/docs/specs/_governance_review/README.md",
		"templates/docs/specs/_independent_evaluation/README.md",
		"templates/docs/specs/repository_mapping.md",
		"templates/docs/specs/rules/stable/s_g_rule_repository_baseline.md",
		"templates/AGENTS.md",
		"templates/GEMINI.md",
		"templates/CLAUDE.md",
		"example.md",
		"tooling/README.md",
		"tooling/cmd/specflowctl/main.go",
		"tooling/internal/demo/demo.go",
		"tooling/go.mod",
		"tooling/manifest.tsv",
		"tooling/reader/web/index.html",
		"tooling/reader/web/styles.css",
		"tooling/reader/web/app.js",
		"tooling/reader/web/cytoscape.min.js",
		"tooling/reader/web/mermaid.min.js",
	} {
		writeCLITestFile(t, filepath.Join(repoRoot, relPath), "# "+filepath.Base(relPath)+"\n")
	}
	for _, relPath := range []string{
		"tooling/scripts/build_release.sh",
		"tooling/scripts/install.ps1",
		"tooling/scripts/install.sh",
		"tooling/scripts/pull_with_release.ps1",
		"tooling/scripts/pull_with_release.sh",
		"tooling/scripts/push_with_release.ps1",
		"tooling/scripts/push_with_release.sh",
		"tooling/scripts/tooling_fingerprint.ps1",
		"tooling/scripts/tooling_fingerprint.sh",
	} {
		writeCLITestFile(t, filepath.Join(repoRoot, relPath), "# script\n")
	}
	return repoRoot
}
func createCLISnapshotRepo(t *testing.T) string {
	return createCLISnapshotRepoWithStatus(t, "unit_verify")
}
func currentCLIToolingScriptFiles() []string {
	return []string{
		"specflow/tooling/scripts/build_release.sh",
		"specflow/tooling/scripts/install.ps1",
		"specflow/tooling/scripts/install.sh",
		"specflow/tooling/scripts/pull_with_release.ps1",
		"specflow/tooling/scripts/pull_with_release.sh",
		"specflow/tooling/scripts/push_with_release.ps1",
		"specflow/tooling/scripts/push_with_release.sh",
		"specflow/tooling/scripts/tooling_fingerprint.ps1",
		"specflow/tooling/scripts/tooling_fingerprint.sh",
	}
}
func createCLISnapshotRepoWithStatus(t *testing.T, unitNextCommand string) string {
	t.Helper()
	repoRoot := t.TempDir()
	writeCLITestFile(t, filepath.Join(repoRoot, "specflow/tooling/manifest.tsv"), "templates/AGENTS.md\tAGENTS.md\tframework\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), ""+
		"# Spec Status\n\n"+
		"## Formal Objects\n\n"+
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n"+
		"|---|---|---|---|---|---|---|\n"+
		"| `unit` | `demo` | `no` | `yes` | `candidate` | `"+unitNextCommand+"` | test |\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), `---
id: demo
layer: candidate
version: 0.1.0
---
# Demo
rule_refs: none
## Testability / Acceptance Criteria
acceptance_item_set:
  - id: demo.core
    target: Demo behavior is accepted.
    verification_surface: internal_flow
    implementation_surface: AgentCore/internal/demo
    verification_method: Go test for the demo behavior.
    pass_condition: The demo behavior passes under the declared checks.
    not_runnable_yet: no
`)
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/repository_mapping.md"), `---
id: repository_mapping
version: 0.1.0
---
# Repository Mapping
`)
	return repoRoot
}
func createCLIStableSnapshotRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_stable_verify` | test |\n")
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_demo.md"), `---
id: demo
layer: stable
version: 1.0.0
---
# Demo
rule_refs: none
## Testability / Acceptance Criteria
acceptance_item_set:
  - id: demo.core
    target: Demo behavior is accepted.
    verification_surface: internal_flow
    implementation_surface: AgentCore/internal/demo
    verification_method: Go test for the demo behavior.
    pass_condition: The demo behavior passes under the declared checks.
    not_runnable_yet: no
`)
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/repository_mapping.md"), `---
id: repository_mapping
version: 0.1.0
---
# Repository Mapping
`)
	return repoRoot
}
func writeCLIStatusRows(t *testing.T, repoRoot, rows string) {
	t.Helper()
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), ""+
		"# Spec Status\n\n"+
		"## Formal Objects\n\n"+
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n"+
		"|---|---|---|---|---|---|---|\n"+
		rows)
}
func writeCLIUnitCheckProcess(t *testing.T, repoRoot string, snap snapshot.Snapshot) {
	t.Helper()
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), "# check\n\n```yaml\n"+strings.Join([]string{
		"object_type: unit",
		"object_ref: demo",
		"gate: unit_check",
		"decision: pass",
		"allow_next: true",
		"next_command: unit_check",
		"blocking_summary: none",
		"coverage_summary: demo",
		"truth_layer_ref: " + snap.TruthLayerRef,
		"truth_file_ref: " + snap.SpecFileRef,
		"truth_version_ref: " + snap.SpecVersionRef,
		"truth_fingerprint: " + snap.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + snap.AcceptanceBehaviorFingerprint,
		"unit_appendix_snapshot: none",
		"unit_snapshot: none",
		"rule_snapshot: none",
		"acceptance_item_set:",
		"  - id: demo.core",
		"    verification_surface: internal_flow",
		"    not_runnable_yet: no",
		"evaluation_mode: independent",
		"reviewer_result: pass",
		"reviewer_context: minimal_context",
		"review_input_refs: " + cliReviewInputRefsForTest(snap.Object, "unit_check_pass", snap.SpecFileRef),
		"review_findings: none",
		"human_decision_refs: none",
	}, "\n")+"\n```\n")
}
func cliReviewInputRefsForTest(object, pack string, refs ...string) string {
	requestFile := filepath.ToSlash(filepath.Join("docs/specs/_independent_evaluation/requests/unit", object, pack+".md"))
	return strings.Join(append([]string{pack, requestFile}, refs...), ";")
}
func writeCLIUnitPlanProcess(t *testing.T, repoRoot string, snap snapshot.Snapshot) {
	t.Helper()
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/demo.md"), "# plan\n\n```yaml\n"+strings.Join([]string{
		"spec_file_ref: " + snap.SpecFileRef,
		"spec_version_ref: " + snap.SpecVersionRef,
		"spec_fingerprint: " + snap.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + snap.AcceptanceBehaviorFingerprint,
		"stable_candidate_diff_refs: none",
		"implementation_gap_refs: docs/specs/repository_mapping.md",
		"unit_appendix_snapshot: none",
		"unit_snapshot: none",
		"rule_snapshot: none",
		"acceptance_item_plan_coverage:",
		"  - id: demo.core",
		"    coverage: implementation slice and verification target",
		"retirement_targets: none",
		"planned_change_scope:",
		"  - id: pcs.core",
		"    basis_refs: " + snap.SpecFileRef,
		"    acceptance_item_ids: demo.core",
		"    implementation_refs: docs/specs/repository_mapping.md",
		"    verification_action: verify package-aware delta",
		"package_constraint_review: pass",
		"package_constraint_refs: " + snap.SpecFileRef,
		"package_constraint_summary: current package constraints reviewed for this delta",
		"evaluation_mode: independent",
		"reviewer_result: pass",
		"reviewer_context: minimal_context",
		"review_input_refs: " + cliReviewInputRefsForTest(snap.Object, "unit_verify_ready_to_promote", snap.SpecFileRef),
		"review_findings: none",
		"human_decision_refs: none",
	}, "\n")+"\n```\n")
}
func writeCLIUnitVerifyProcess(t *testing.T, repoRoot string, snap snapshot.Snapshot) {
	t.Helper()
	activePlanFingerprint := cliFileFingerprint(t, repoRoot, snapshot.ActivePlanFilePath(snap.Object))
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit/demo.md"), "# verify\n\n```yaml\n"+strings.Join([]string{
		"object_type: unit",
		"object_ref: demo",
		"gate: unit_verify",
		"decision: pass",
		"allow_next: true",
		"next_command: unit_promote",
		"blocking_summary: none",
		"coverage_summary: current candidate",
		"truth_layer_ref: " + snap.TruthLayerRef,
		"truth_file_ref: " + snap.SpecFileRef,
		"truth_version_ref: " + snap.SpecVersionRef,
		"truth_fingerprint: " + snap.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + snap.AcceptanceBehaviorFingerprint,
		"acceptance_item_set:",
		"  - id: demo.core",
		"    verification_surface: internal_flow",
		"    not_runnable_yet: no",
		"unit_appendix_snapshot: none",
		"unit_snapshot: none",
		"active_plan_file_ref: " + snapshot.ActivePlanFilePath(snap.Object),
		"active_plan_fingerprint: " + activePlanFingerprint,
		"rule_snapshot: none",
		"acceptance_item_evidence_matrix:",
		"  - id: demo.core",
		"    status: pass",
		"    evidence_refs: go test ./...",
		"retirement_evidence_matrix: none",
		"package_delta_verification:",
		"  - planned_change_scope_id: pcs.core",
		"    result: pass",
		"    evidence_refs: go test ./...",
		"evaluation_mode: independent",
		"reviewer_result: pass",
		"reviewer_context: minimal_context",
		"review_input_refs: " + cliReviewInputRefsForTest(snap.Object, "unit_verify_ready_to_promote", snap.SpecFileRef),
		"review_findings: none",
		"human_decision_refs: none",
	}, "\n")+"\n```\n")
}
func cliFileFingerprint(t *testing.T, repoRoot, fileRef string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(fileRef)))
	if err != nil {
		t.Fatalf("read %s: %v", fileRef, err)
	}
	text := strings.ReplaceAll(string(content), "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")
	text += "\n"
	sum := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", sum)
}
func writeCLIStableVerifyProcess(t *testing.T, repoRoot string, snap snapshot.Snapshot, decision string) {
	t.Helper()
	mapping, err := snapshot.BuildRepositoryMappingSnapshot(repoRoot)
	if err != nil {
		t.Fatalf("BuildRepositoryMappingSnapshot: %v", err)
	}
	allowNext := "false"
	nextCommand := "unit_stable_verify"
	if decision == "aligned" || decision == "controlled_repair_required" || decision == "controlled_change_required" {
		allowNext = "true"
		nextCommand = "unit_fork"
	}
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/_stable_verify_result/unit/demo.md"), "# stable verify\n\n```yaml\n"+strings.Join([]string{
		"object_type: unit",
		"object_ref: demo",
		"gate: unit_stable_verify",
		"decision: " + decision,
		"allow_next: " + allowNext,
		"next_command: " + nextCommand,
		"blocking_summary: none",
		"coverage_summary: current stable implementation",
		"truth_layer_ref: stable",
		"truth_file_ref: " + snap.SpecFileRef,
		"truth_version_ref: " + snap.SpecVersionRef,
		"truth_fingerprint: " + snap.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + snap.AcceptanceBehaviorFingerprint,
		"repository_mapping_snapshot:",
		"  file_ref: " + mapping.FileRef,
		"  version_ref: " + mapping.VersionRef,
		"  fingerprint: " + mapping.Fingerprint,
		"acceptance_item_set:",
		"  - id: demo.core",
		"    verification_surface: internal_flow",
		"    not_runnable_yet: no",
		"unit_appendix_snapshot: none",
		"unit_snapshot: none",
		"rule_snapshot: none",
		"acceptance_item_evidence_matrix:",
		"  - id: demo.core",
		"    status: pass",
		"    evidence_refs: go test ./...",
		"implementation_surface_refs: AgentCore/internal/demo",
		"evidence_refs: go test ./...",
		"evaluation_mode: independent",
		"reviewer_result: pass",
		"reviewer_context: minimal_context",
		"review_input_refs: " + cliReviewInputRefsForTest(snap.Object, "unit_stable_verify_advancing", snap.SpecFileRef),
		"review_findings: none",
		"human_decision_refs: none",
	}, "\n")+"\n```\n")
}
func writeCLIUnitReleaseSpec(t *testing.T, repoRoot, layer, unit, version string, unitRefs []string) {
	t.Helper()
	dir := "stable"
	prefix := "s_unit_"
	if layer == "candidate" {
		dir = "candidate"
		prefix = "c_unit_"
	}
	unitRefsBlock := "unit_refs: none"
	if len(unitRefs) > 0 {
		lines := []string{"unit_refs:"}
		for _, ref := range unitRefs {
			lines = append(lines, "  - "+ref)
		}
		unitRefsBlock = strings.Join(lines, "\n")
	}
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units", dir, prefix+unit+".md"), strings.Join([]string{
		"---",
		"id: " + unit,
		"layer: " + layer,
		"version: " + version,
		unitRefsBlock,
		"rule_refs: none",
		"---",
		"",
		"# " + unit,
		"",
	}, "\n"))
}
func mustReadCLITestFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}
func assertCLITestFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected %s to be removed", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
}
func writeCLITestFile(t *testing.T, path, content string) {
	t.Helper()
	content = testfixtures.NormalizeSpecFlowContent(path, content)
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
