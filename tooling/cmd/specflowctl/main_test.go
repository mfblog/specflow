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
		"specflow/framework/commands/unit_check.md",
		"specflow/framework/commands/unit_plan.md",
		"specflow/framework/process_snapshot_contract.md",
		"specflow/framework/slice_work_state_protocol.md",
		"specflow/framework/candidate_handoff_contract.md",
		"specflow/framework/spec_writing_guide.md",
		"specflow/framework/candidate_intent_policy.md",
	} {
		writeCLITestFile(t, filepath.Join(repoRoot, filepath.FromSlash(relPath)), "# "+filepath.Base(relPath)+"\n")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runProcess([]string{"check-work-init", "--object-type", "unit", "--object", "demo", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("check-work-init failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Check work-state created:") {
		t.Fatalf("expected created output, got %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if err := runProcess([]string{"check-work-validate", "--object-type", "unit", "--object", "demo", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("check-work-validate failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Check work-state is valid:") {
		t.Fatalf("expected valid output, got %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if err := runProcess([]string{"check-work-refresh", "--object-type", "unit", "--object", "demo", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("check-work-refresh failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Check work-state refreshed:") {
		t.Fatalf("expected refreshed output, got %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if err := runProcess([]string{"check-work-touch", "--object-type", "unit", "--object", "demo", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("check-work-touch failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Check work-state touched:") {
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

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := runUnit([]string{"release-version", "--unit", "assistant", "--from-ref", "s_unit_assistant@0.8.0", "--to-ref", "s_unit_assistant@0.9.0", "--repo-root", repoRoot}, &stdout, &stderr); err != nil {
		t.Fatalf("unit release-version failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Released unit version: assistant from s_unit_assistant@0.8.0 to s_unit_assistant@0.9.0") {
		t.Fatalf("expected release output, got %s", output)
	}
	if !strings.Contains(output, "- unit:agent") {
		t.Fatalf("expected updated unit output, got %s", output)
	}
	content := mustReadCLITestFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_agent.md"))
	if !strings.Contains(content, "  - s_unit_assistant@0.9.0") {
		t.Fatalf("expected updated unit_refs, got %s", content)
	}
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
	if !strings.Contains(output, "specflow/framework/candidate_intents/repair.md") {
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
	if !strings.Contains(output, "specflow/framework/candidate_intent_policy.md") {
		t.Fatalf("expected candidate intent policy in design collect-default-scope output, got %s", output)
	}
	if !strings.Contains(output, "specflow/framework/candidate_intents/repair.md") {
		t.Fatalf("expected repair intent standard in design collect-default-scope output, got %s", output)
	}
	if !strings.Contains(output, "specflow/framework/candidate_intents/change.md") {
		t.Fatalf("expected change intent standard in design collect-default-scope output, got %s", output)
	}
	if !strings.Contains(output, "specflow/framework/output_baseline.md") {
		t.Fatalf("expected output baseline in design collect-default-scope output, got %s", output)
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
		"unit_appendix_snapshot: none",
		"rule_snapshot: none",
		"acceptance_item_plan_coverage:",
		"  - id: demo.core",
		"    coverage: implementation slice and verification target",
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

func TestSnapshotValidateProcessRejectsScenarioPlanCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepo(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runSnapshot([]string{"validate-process", "--object-type", "scenario", "--object", "demo_flow", "--process", "plan", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "object type \"scenario\" is not supported; only unit is supported") {
		t.Fatalf("expected scenario rejection, got err=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
}

func TestCommandPreflightPassesWithValidUnitImplementationInputsCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepo(t)
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCLIUnitCheckProcess(t, repoRoot, expected)
	writeCLIUnitPlanProcess(t, repoRoot, expected)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runCommand([]string{"preflight", "--command", "unit_impl", "--object-type", "unit", "--object", "demo", "--repo-root", repoRoot}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("preflight failed: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	output := stdout.String()
	for _, expectedText := range []string{
		"preflight_result: pass",
		"may_continue: true",
		"- process: check",
		"- process: plan",
		"result: valid",
	} {
		if !strings.Contains(output, expectedText) {
			t.Fatalf("expected %q in output, got %s", expectedText, output)
		}
	}
}

func TestCommandPreflightUsesNormalizedSnapshotFingerprintCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepoWithStatus(t, "unit_plan")
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
	err = runCommand([]string{"preflight", "--command", "unit_plan", "--object-type", "unit", "--object", "demo", "--repo-root", repoRoot}, &stdout, &stderr)
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

func TestCommandPreflightRejectsRawHashWithoutCleanupSideEffectsCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepoWithStatus(t, "unit_plan")
	specPath := filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md")
	content, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	crlfContent := strings.ReplaceAll(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n", "\r\n")
	if err := os.WriteFile(specPath, []byte(crlfContent), 0o644); err != nil {
		t.Fatalf("rewrite spec with CRLF: %v", err)
	}
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	content, err = os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	rawHash := fmt.Sprintf("%x", sha256.Sum256(content))
	if rawHash == expected.SpecFingerprint {
		t.Fatalf("test fixture must produce a raw byte hash that differs from normalized fingerprint")
	}
	stale := expected
	stale.SpecFingerprint = rawHash
	writeCLIUnitCheckProcess(t, repoRoot, stale)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runCommand([]string{"preflight", "--command", "unit_plan", "--object-type", "unit", "--object", "demo", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected preflight failure, got stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "preflight_result: fail") || !strings.Contains(output, "may_continue: false") {
		t.Fatalf("expected fail output, got %s", output)
	}
	statusContent, err := os.ReadFile(filepath.Join(repoRoot, "docs/specs/_status.md"))
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	if !strings.Contains(string(statusContent), "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_plan` | test |") {
		t.Fatalf("preflight must not rewrite status, got %s", string(statusContent))
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md")); err != nil {
		t.Fatalf("preflight must not delete process file: %v", err)
	}
}

func TestCommandCloseRejectsStaleCandidateInputWithoutStatusUpdateCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepoWithStatus(t, "unit_plan")
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	stale := expected
	stale.SpecFingerprint = "stale-fingerprint"
	writeCLIUnitCheckProcess(t, repoRoot, stale)

	statusPath := filepath.Join(repoRoot, "docs/specs/_status.md")
	beforeStatus := mustReadCLITestFile(t, statusPath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runCommand([]string{"close", "--command", "unit_plan", "--object-type", "unit", "--object", "demo", "--outcome", "blocked", "--apply", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "command close input preflight failed") {
		t.Fatalf("expected input preflight failure, got err=%v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
	}
	output := stdout.String()
	for _, expectedText := range []string{
		"command_close_result: failed",
		"apply: true",
		"input_validation_action: command_preflight",
		"input_validated_processes:",
		"- process: check",
		"result: invalid",
		"input_validation_mismatches",
	} {
		if !strings.Contains(output, expectedText) {
			t.Fatalf("expected %q in output, got %s", expectedText, output)
		}
	}
	if afterStatus := mustReadCLITestFile(t, statusPath); afterStatus != beforeStatus {
		t.Fatalf("stale input must not rewrite status\nbefore=%s\nafter=%s", beforeStatus, afterStatus)
	}
}

func TestCommandPreflightReportsMissingProcessFileAsCommandLayerCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepoWithStatus(t, "unit_plan")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := runCommand([]string{"preflight", "--command", "unit_plan", "--object-type", "unit", "--object", "demo", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected preflight failure for missing check result, got stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	output := stdout.String()
	for _, expectedText := range []string{
		"preflight_result: fail",
		"may_continue: false",
		"failure_layer: gate_layer",
		"recommended_next_command: unit_check",
		"diagnostic: missing process file: docs/specs/_check_result/unit/demo.md",
	} {
		if !strings.Contains(output, expectedText) {
			t.Fatalf("expected %q in output, got %s", expectedText, output)
		}
	}
}

func TestCommandPreflightTreatsValidationExecutionErrorAsToolingGapCLI(t *testing.T) {
	repoRoot := createCLISnapshotRepoWithStatus(t, "unit_plan")
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCLIUnitCheckProcess(t, repoRoot, expected)
	specPath := filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md")
	if err := os.Remove(specPath); err != nil {
		t.Fatalf("remove candidate spec: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runCommand([]string{"preflight", "--command", "unit_plan", "--object-type", "unit", "--object", "demo", "--repo-root", repoRoot}, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected preflight failure for validation execution error, got stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	output := stdout.String()
	for _, expectedText := range []string{
		"preflight_result: fail",
		"may_continue: false",
		"failure_layer: tooling_gap",
		"recommended_next_command: none",
		"diagnostic: read docs/specs/units/candidate/c_unit_demo.md",
	} {
		if !strings.Contains(output, expectedText) {
			t.Fatalf("expected %q in output, got %s", expectedText, output)
		}
	}
	for _, forbiddenText := range []string{
		"failure_layer: gate_layer",
		"failure_layer: truth_layer",
	} {
		if strings.Contains(output, forbiddenText) {
			t.Fatalf("validation execution error must not produce cleanup layer %q, got %s", forbiddenText, output)
		}
	}
	statusContent, err := os.ReadFile(filepath.Join(repoRoot, "docs/specs/_status.md"))
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	if !strings.Contains(string(statusContent), "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_plan` | test |") {
		t.Fatalf("preflight must not rewrite status, got %s", string(statusContent))
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md")); err != nil {
		t.Fatalf("preflight must not delete process file: %v", err)
	}
}

func TestCommandCloseUnitStableVerifyControlledRepairDryRunAndApplyCLI(t *testing.T) {
	repoRoot := createCLITestRepo(t)
	writeCLIStatusRows(t, repoRoot, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_stable_verify` | test |\n")
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
	repoRoot := createCLISnapshotRepoWithStatus(t, "unit_impl")
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
		"spec_flow_review.md",
		"spec_flow_design_review.md",
		"spec_flow_migrate.md",
		"agent_operability_standard.md",
		"natural_language_routing.md",
		"advance_policy.md",
		"onboarding_decision_policy.md",
		"candidate_intent_policy.md",
		"command_policy.md",
		"implementation_change_policy.md",
		"checkpoint_protocol.md",
		"output_baseline.md",
		"tooling_execution_policy.md",
		"slice_work_state_protocol.md",
		"severity_policy.md",
		"spec_policy.md",
		"spec_writing_guide.md",
		"repository_mapping_policy.md",
		"candidate_handoff_contract.md",
		"downgrade_policy.md",
		"recovery_policy.md",
		"impact_sync_policy.md",
		"process_snapshot_contract.md",
		"entry_index_registry.md",
		"rule_new.md",
		"rule_extract.md",
		"rule_bind.md",
		"rule_topology.md",
		"rule_sync.md",
		"rule_escape.md",
	}
	for _, name := range frameworkFiles {
		writeCLITestFile(t, filepath.Join(repoRoot, "specflow/framework", name), "# "+name+"\n")
	}
	for _, relPath := range []string{
		"specflow/framework/candidate_intents/repair.md",
		"specflow/framework/candidate_intents/change.md",
	} {
		writeCLITestFile(t, filepath.Join(repoRoot, relPath), "# "+filepath.Base(relPath)+"\n")
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
		"specflow/templates/docs/specs/_check_work/README.md",
		"specflow/templates/docs/specs/_check_result/README.md",
		"specflow/templates/docs/specs/_plans/README.md",
		"specflow/templates/docs/specs/_plans/draft/README.md",
		"specflow/templates/docs/specs/_plans/active/README.md",
		"specflow/templates/docs/specs/_verify_result/README.md",
		"specflow/templates/docs/specs/_governance_review/README.md",
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
		"specflow/tooling/scripts/build_release.sh",
		"specflow/tooling/scripts/tooling_fingerprint.sh",
		"specflow/tooling/scripts/tooling_fingerprint.ps1",
		"specflow/tooling/go.mod",
		"specflow/tooling/manifest.tsv",
	} {
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

func createCLISnapshotRepo(t *testing.T) string {
	return createCLISnapshotRepoWithStatus(t, "unit_impl")
}

func createCLISnapshotRepoWithStatus(t *testing.T, unitNextCommand string) string {
	t.Helper()
	repoRoot := t.TempDir()
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
		"next_command: unit_plan",
		"blocking_summary: none",
		"coverage_summary: demo",
		"truth_layer_ref: " + snap.TruthLayerRef,
		"truth_file_ref: " + snap.SpecFileRef,
		"truth_version_ref: " + snap.SpecVersionRef,
		"truth_fingerprint: " + snap.SpecFingerprint,
		"unit_appendix_snapshot: none",
		"rule_snapshot: none",
		"acceptance_item_set:",
		"  - id: demo.core",
		"    verification_surface: internal_flow",
		"    not_runnable_yet: no",
	}, "\n")+"\n```\n")
}

func writeCLIUnitPlanProcess(t *testing.T, repoRoot string, snap snapshot.Snapshot) {
	t.Helper()
	writeCLITestFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/demo.md"), "# plan\n\n```yaml\n"+strings.Join([]string{
		"spec_file_ref: " + snap.SpecFileRef,
		"spec_version_ref: " + snap.SpecVersionRef,
		"spec_fingerprint: " + snap.SpecFingerprint,
		"unit_appendix_snapshot: none",
		"rule_snapshot: none",
		"acceptance_item_plan_coverage:",
		"  - id: demo.core",
		"    coverage: implementation slice and verification target",
	}, "\n")+"\n```\n")
}

func writeCLIUnitVerifyProcess(t *testing.T, repoRoot string, snap snapshot.Snapshot) {
	t.Helper()
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
		"acceptance_item_set:",
		"  - id: demo.core",
		"    verification_surface: internal_flow",
		"    not_runnable_yet: no",
		"unit_appendix_snapshot: none",
		"verification_scope_ref: current candidate",
		"rule_snapshot: none",
		"acceptance_item_evidence_matrix:",
		"  - id: demo.core",
		"    status: pass",
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
