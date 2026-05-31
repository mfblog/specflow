package evaluationrequest

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
)

func TestCreateFreshnessTextDriftRequestRendersFreshnessReceipt(t *testing.T) {
	repoRoot := setupCandidateRequestRepo(t)
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeCheckProcess(t, repoRoot, expected)

	candidatePath := filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md")
	content := mustReadFile(t, candidatePath)
	writeFile(t, candidatePath, strings.Replace(content, "# Demo\n", "# Demo\n\nEditorial note only.\n", 1))

	result, err := Create(Options{
		RepoRoot:    repoRoot,
		ObjectType:  "unit",
		Object:      "demo",
		Pack:        PackFreshnessTextDriftReuse,
		ProcessKind: "check",
		Now:         time.Date(2026, 5, 30, 1, 2, 3, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if result.Validation.FreshnessImpact != snapshot.FreshnessTextDrift || result.Validation.EvidenceReuse != snapshot.EvidenceReusePendingReview {
		t.Fatalf("expected pending text drift, got impact=%s reuse=%s", result.Validation.FreshnessImpact, result.Validation.EvidenceReuse)
	}

	request := mustReadFile(t, filepath.Join(repoRoot, filepath.FromSlash(result.RequestFile)))
	for _, phrase := range []string{
		"## Allowed Inputs",
		"current truth or spec file.",
		"## Forbidden Inputs",
		"reuse claims when deterministic validation reports `semantic_drift`, `acceptance_drift`, `dependency_drift`, `schema_drift`, or `unknown_drift`.",
		"## Reviewer Output",
		"## Executor Receipt After Pass",
		"Only the executor writes this receipt into process evidence after receiving reviewer result `pass`.",
		"freshness_impact: text_drift",
		"evidence_reuse: accepted",
		"freshness_current_fingerprint: " + result.Validation.Expected.SpecFingerprint,
		"freshness_review_mode: independent",
		"freshness_reviewer_result: pass",
		"freshness_reviewer_context: minimal_context",
		"freshness_review_input_refs: freshness_text_drift_reuse;" + result.RequestFile,
		"freshness_review_findings: none",
	} {
		if !strings.Contains(request, phrase) {
			t.Fatalf("freshness request missing %q:\n%s", phrase, request)
		}
	}
	for _, forbidden := range []string{
		"\nevaluation_mode: independent\n",
		"\nreviewer_result: pass\n",
		"\nreview_input_refs: freshness_text_drift_reuse",
		"\nhuman_decision_refs: none\n",
		"framework/lifecycle/unit_check.md",
		"framework/lifecycle/unit_plan.md",
		"framework/lifecycle/unit_verify.md",
		"framework/lifecycle/unit_stable_verify.md",
	} {
		if strings.Contains(request, forbidden) {
			t.Fatalf("freshness request must not render ordinary receipt field %q:\n%s", forbidden, request)
		}
	}
}

func TestCreateStandardRequestsIncludeLifecycleOwnerRefs(t *testing.T) {
	tests := []struct {
		name         string
		pack         string
		lifecycleRef string
		allowed      string
		forbidden    string
		question     string
		setup        func(t *testing.T) string
	}{
		{
			name:         "check",
			pack:         PackUnitCheckPass,
			lifecycleRef: "framework/lifecycle/unit_check.md",
			allowed:      "candidate unit truth, candidate appendices owned by the unit, stable truth, and rules.",
			forbidden:    "implementation plan drafts.",
			question:     "Is the unit goal, responsibility, boundary, dependency truth, and rule binding explicit enough for planning?",
			setup: func(t *testing.T) string {
				t.Helper()
				repoRoot := setupCandidateRequestRepo(t)
				expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
				if err != nil {
					t.Fatalf("RebuildCurrentObject: %v", err)
				}
				writeCheckProcessWithoutReceipt(t, repoRoot, expected)
				return repoRoot
			},
		},
		{
			name:         "plan",
			pack:         PackUnitPlanPlanReady,
			lifecycleRef: "framework/lifecycle/unit_plan.md",
			allowed:      "active plan under review.",
			forbidden:    "implementation work not authorized by the active plan.",
			question:     "Does the plan cover every accepted acceptance item?",
			setup: func(t *testing.T) string {
				t.Helper()
				repoRoot := setupCandidateRequestRepoWithRefs(t)
				expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
				if err != nil {
					t.Fatalf("RebuildCurrentObject: %v", err)
				}
				writeFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), "# check\n")
				writePlanProcessWithoutReceipt(t, repoRoot, expected)
				return repoRoot
			},
		},
		{
			name:         "verify",
			pack:         PackUnitVerifyReadyToPromote,
			lifecycleRef: "framework/lifecycle/unit_verify.md",
			allowed:      "verify result under review.",
			forbidden:    "unrecorded executor claims that tests passed.",
			question:     "Does the verify result cover every executable acceptance item?",
			setup: func(t *testing.T) string {
				t.Helper()
				repoRoot := setupCandidateRequestRepoWithRefs(t)
				expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
				if err != nil {
					t.Fatalf("RebuildCurrentObject: %v", err)
				}
				writeFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), "# check\n")
				writePlanProcessWithoutReceipt(t, repoRoot, expected)
				writeVerifyProcessWithoutReceipt(t, repoRoot, expected)
				return repoRoot
			},
		},
		{
			name:         "stable verify",
			pack:         PackUnitStableVerifyAdvancing,
			lifecycleRef: "framework/lifecycle/unit_stable_verify.md",
			allowed:      "stable unit truth, stable appendices owned by the unit, rules, and repository mapping snapshot.",
			forbidden:    "executor preference for aligned, controlled repair, or controlled change outcomes.",
			question:     "Does current implementation align with stable truth, or does the stored decision correctly identify the controlled next step?",
			setup: func(t *testing.T) string {
				t.Helper()
				repoRoot := setupStableRequestRepo(t)
				expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
				if err != nil {
					t.Fatalf("RebuildCurrentObject: %v", err)
				}
				mapping, err := snapshot.BuildRepositoryMappingSnapshot(repoRoot)
				if err != nil {
					t.Fatalf("BuildRepositoryMappingSnapshot: %v", err)
				}
				writeStableVerifyProcessWithoutReceipt(t, repoRoot, expected, mapping)
				return repoRoot
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repoRoot := tc.setup(t)
			result, err := Create(Options{
				RepoRoot:   repoRoot,
				ObjectType: "unit",
				Object:     "demo",
				Pack:       tc.pack,
				Now:        time.Date(2026, 5, 30, 1, 2, 3, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("Create: %v", err)
			}
			if !containsString(result.ReviewInputRefs, tc.lifecycleRef) {
				t.Fatalf("expected lifecycle owner ref %s, got %+v", tc.lifecycleRef, result.ReviewInputRefs)
			}

			request := mustReadFile(t, filepath.Join(repoRoot, filepath.FromSlash(result.RequestFile)))
			if !strings.Contains(request, "- "+tc.lifecycleRef+"\n") {
				t.Fatalf("request file missing lifecycle owner ref %s:\n%s", tc.lifecycleRef, request)
			}
			for _, phrase := range []string{
				"## Allowed Inputs",
				tc.allowed,
				"## Forbidden Inputs",
				tc.forbidden,
				"## Evaluation Questions",
				tc.question,
				"## Reviewer Output",
				"pass | blocked | needs_human_decision",
				"## Executor Receipt After Pass",
				"Only the executor writes this receipt into process evidence after receiving reviewer result `pass`.",
			} {
				if !strings.Contains(request, phrase) {
					t.Fatalf("request file missing %q:\n%s", phrase, request)
				}
			}
		})
	}
}

func TestCreatePlanRequestIncludesSnapshotInputRefs(t *testing.T) {
	repoRoot := setupCandidateRequestRepoWithRefs(t)
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	writeFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), "# check\n")
	writePlanProcessWithoutReceipt(t, repoRoot, expected)

	result, err := Create(Options{
		RepoRoot:   repoRoot,
		ObjectType: "unit",
		Object:     "demo",
		Pack:       PackUnitPlanPlanReady,
		Now:        time.Date(2026, 5, 30, 1, 2, 3, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	for _, ref := range []string{
		"docs/specs/units/candidate/c_unit_demo.md",
		"docs/specs/units/stable/s_unit_demo.md",
		"docs/specs/units/candidate/appendix/c_unit_demo_evidence.md",
		"docs/specs/units/stable/s_unit_dependency.md",
		"docs/specs/rules/candidate/c_b_rule_demo.md",
		"docs/specs/_check_result/unit/demo.md",
		"docs/specs/_plans/active/demo.md",
	} {
		if !containsString(result.ReviewInputRefs, ref) {
			t.Fatalf("expected review input ref %s, got %+v", ref, result.ReviewInputRefs)
		}
	}
}

func TestCreateStableVerifyRequestIncludesProcessScalarRefs(t *testing.T) {
	repoRoot := setupStableRequestRepo(t)
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	mapping, err := snapshot.BuildRepositoryMappingSnapshot(repoRoot)
	if err != nil {
		t.Fatalf("BuildRepositoryMappingSnapshot: %v", err)
	}
	writeStableVerifyProcessWithoutReceipt(t, repoRoot, expected, mapping)

	result, err := Create(Options{
		RepoRoot:   repoRoot,
		ObjectType: "unit",
		Object:     "demo",
		Pack:       PackUnitStableVerifyAdvancing,
		Now:        time.Date(2026, 5, 30, 1, 2, 3, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	for _, ref := range []string{
		"docs/specs/_stable_verify_result/unit/demo.md",
		"docs/specs/repository_mapping.md",
		"docs/specs/units/stable/appendix/s_unit_demo_notes.md",
		"AgentCore/internal/demo",
		"go test ./...",
	} {
		if !containsString(result.ReviewInputRefs, ref) {
			t.Fatalf("expected review input ref %s, got %+v", ref, result.ReviewInputRefs)
		}
	}
	if !containsString(result.ReviewFileRefs, "docs/specs/_stable_verify_result/unit/demo.md") {
		t.Fatalf("expected stable verify process file in file refs, got %+v", result.ReviewFileRefs)
	}
	for _, ref := range []string{"AgentCore/internal/demo", "go test ./..."} {
		if !containsString(result.ReviewEvidenceRefs, ref) {
			t.Fatalf("expected evidence ref %s, got %+v", ref, result.ReviewEvidenceRefs)
		}
		if containsString(result.ReviewFileRefs, ref) {
			t.Fatalf("evidence ref %s must not be rendered as file ref, got %+v", ref, result.ReviewFileRefs)
		}
	}
	request := mustReadFile(t, filepath.Join(repoRoot, filepath.FromSlash(result.RequestFile)))
	fileRefs := sectionBetween(t, request, "## Review File Refs", "## Review Evidence Refs")
	evidenceRefs := sectionBetween(t, request, "## Review Evidence Refs", "## Evaluation Questions")
	for _, ref := range []string{"AgentCore/internal/demo", "go test ./..."} {
		if strings.Contains(fileRefs, ref) {
			t.Fatalf("evidence ref %s must not appear in Review File Refs:\n%s", ref, fileRefs)
		}
		if !strings.Contains(evidenceRefs, ref) {
			t.Fatalf("evidence ref %s missing from Review Evidence Refs:\n%s", ref, evidenceRefs)
		}
	}
	if !strings.Contains(request, "Use Review Evidence Refs only to judge whether the recorded evidence is sufficient and traceable; do not treat every evidence ref as a readable file.") {
		t.Fatalf("request missing evidence-ref role instruction:\n%s", request)
	}
}

func TestCreateStableVerifyAdvancingRejectsNonAdvancingDecision(t *testing.T) {
	repoRoot := setupStableRequestRepo(t)
	expected, err := snapshot.RebuildCurrentObject(repoRoot, "unit", "demo")
	if err != nil {
		t.Fatalf("RebuildCurrentObject: %v", err)
	}
	mapping, err := snapshot.BuildRepositoryMappingSnapshot(repoRoot)
	if err != nil {
		t.Fatalf("BuildRepositoryMappingSnapshot: %v", err)
	}
	writeStableVerifyProcessWithoutReceipt(t, repoRoot, expected, mapping)

	processPath := filepath.Join(repoRoot, "docs/specs/_stable_verify_result/unit/demo.md")
	body := mustReadFile(t, processPath)
	body = strings.Replace(body, "decision: aligned", "decision: small_repair_required", 1)
	body = strings.Replace(body, "allow_next: true", "allow_next: false", 1)
	body = strings.Replace(body, "next_command: unit_fork", "next_command: unit_stable_verify", 1)
	writeFile(t, processPath, body)

	_, err = Create(Options{
		RepoRoot:   repoRoot,
		ObjectType: "unit",
		Object:     "demo",
		Pack:       PackUnitStableVerifyAdvancing,
		Now:        time.Date(2026, 5, 30, 1, 2, 3, 0, time.UTC),
	})
	if err == nil || !strings.Contains(err.Error(), "requires stable_verify decision aligned") {
		t.Fatalf("expected non-advancing stable verify decision rejection, got %v", err)
	}
}

func setupCandidateRequestRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	writeStatus(t, repoRoot, "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | test |\n")
	writeFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), `---
id: demo
layer: candidate
version: 0.1.0
rule_refs: none
evidence_appendix_ref: none
---

# Demo

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
	writeRepositoryMapping(t, repoRoot)
	return repoRoot
}

func setupCandidateRequestRepoWithRefs(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	writeStatus(t, repoRoot, strings.Join([]string{
		"| `unit` | `demo` | `yes` | `yes` | `candidate` | `unit_plan` | test |",
		"| `unit` | `dependency` | `yes` | `no` | `stable` | `unit_fork` | test |",
	}, "\n")+"\n")
	writeFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_demo.md"), stableUnitSpec("demo", "1.0.0"))
	writeFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_dependency.md"), stableUnitSpec("dependency", "1.0.0"))
	writeFile(t, filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md"), `---
rule_id: demo_rule
rule_scope: bound
layer: candidate
rule_version: 0.1.0
---

# Demo Rule
`)
	writeFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/appendix/c_unit_demo_evidence.md"), `---
unit: demo
layer: candidate
---

# Evidence
`)
	writeFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), `---
id: demo
layer: candidate
version: 0.2.0
unit_refs:
  - s_unit_dependency@1.0.0
rule_refs:
  - c_b_rule_demo@0.1.0
evidence_appendix_ref: docs/specs/units/candidate/appendix/c_unit_demo_evidence.md
---

# Demo

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
	writeRepositoryMapping(t, repoRoot)
	return repoRoot
}

func setupStableRequestRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	writeStatus(t, repoRoot, "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_stable_verify` | test |\n")
	writeFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_demo.md"), stableUnitSpec("demo", "1.0.0"))
	writeFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/appendix/s_unit_demo_notes.md"), `---
unit: demo
layer: stable
---

# Demo Notes
`)
	writeRepositoryMapping(t, repoRoot)
	return repoRoot
}

func stableUnitSpec(unit, version string) string {
	return `---
id: ` + unit + `
layer: stable
version: ` + version + `
rule_refs: none
evidence_appendix_ref: none
---

# ` + unit + `

## Testability / Acceptance Criteria

acceptance_item_set:
  - id: ` + unit + `.core
    target: ` + unit + ` behavior is accepted.
    verification_surface: internal_flow
    implementation_surface: AgentCore/internal/` + unit + `
    verification_method: Go test for the ` + unit + ` behavior.
    pass_condition: The ` + unit + ` behavior passes under the declared checks.
    not_runnable_yet: no
`
}

func writeStatus(t *testing.T, repoRoot, rows string) {
	t.Helper()
	writeFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n"+rows)
}

func writeRepositoryMapping(t *testing.T, repoRoot string) {
	t.Helper()
	writeFile(t, filepath.Join(repoRoot, "docs/specs/repository_mapping.md"), `---
id: repository_mapping
version: 0.1.0
---

# Repository Mapping
`)
}

func writeCheckProcessWithoutReceipt(t *testing.T, repoRoot string, expected snapshot.Snapshot) {
	t.Helper()
	writeFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), "# check\n\n```yaml\n"+strings.Join([]string{
		"object_type: unit",
		"object_ref: demo",
		"gate: unit_check",
		"decision: pass",
		"allow_next: true",
		"next_command: unit_plan",
		"blocking_summary: none",
		"coverage_summary: current candidate",
		"truth_layer_ref: " + expected.TruthLayerRef,
		"truth_file_ref: " + expected.SpecFileRef,
		"truth_version_ref: " + expected.SpecVersionRef,
		"truth_fingerprint: " + expected.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + expected.AcceptanceBehaviorFingerprint,
		"acceptance_item_set:",
		renderAcceptanceItems(expected.AcceptanceItemSet),
		"unit_appendix_snapshot: none",
		"rule_snapshot: none",
	}, "\n")+"\n```\n")
}

func writeCheckProcess(t *testing.T, repoRoot string, expected snapshot.Snapshot) {
	t.Helper()
	writeFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), "# check\n\n```yaml\n"+strings.Join([]string{
		"object_type: unit",
		"object_ref: demo",
		"gate: unit_check",
		"decision: pass",
		"allow_next: true",
		"next_command: unit_plan",
		"blocking_summary: none",
		"coverage_summary: current candidate",
		"truth_layer_ref: " + expected.TruthLayerRef,
		"truth_file_ref: " + expected.SpecFileRef,
		"truth_version_ref: " + expected.SpecVersionRef,
		"truth_fingerprint: " + expected.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + expected.AcceptanceBehaviorFingerprint,
		"acceptance_item_set:",
		renderAcceptanceItems(expected.AcceptanceItemSet),
		"unit_appendix_snapshot: none",
		"rule_snapshot: none",
		"evaluation_mode: independent",
		"reviewer_result: pass",
		"reviewer_context: minimal_context",
		"review_input_refs: " + requestReviewInputRefsForTest(expected.Object, PackUnitCheckPass, expected.SpecFileRef),
		"review_findings: none",
		"human_decision_refs: none",
	}, "\n")+"\n```\n")
}

func requestReviewInputRefsForTest(object, pack string, refs ...string) string {
	requestFile := filepath.ToSlash(filepath.Join("docs/specs/_independent_evaluation/requests/unit", object, pack+".md"))
	return strings.Join(append([]string{pack, requestFile}, refs...), ";")
}

func writeVerifyProcessWithoutReceipt(t *testing.T, repoRoot string, expected snapshot.Snapshot) {
	t.Helper()
	activePlanFingerprint := requestFileFingerprint(t, repoRoot, snapshot.ActivePlanFilePath(expected.Object))
	writeFile(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit/demo.md"), "# verify\n\n```yaml\n"+strings.Join([]string{
		"object_type: unit",
		"object_ref: demo",
		"gate: unit_verify",
		"decision: pass",
		"allow_next: true",
		"next_command: unit_promote",
		"blocking_summary: none",
		"coverage_summary: current candidate",
		"truth_layer_ref: " + expected.TruthLayerRef,
		"truth_file_ref: " + expected.SpecFileRef,
		"truth_version_ref: " + expected.SpecVersionRef,
		"truth_fingerprint: " + expected.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + expected.AcceptanceBehaviorFingerprint,
		"acceptance_item_set:",
		renderAcceptanceItems(expected.AcceptanceItemSet),
		"unit_appendix_snapshot:",
		renderAppendix(expected.ModuleAppendixSnapshot),
		"verification_scope_ref: current candidate",
		"active_plan_file_ref: " + snapshot.ActivePlanFilePath(expected.Object),
		"active_plan_fingerprint: " + activePlanFingerprint,
		"rule_snapshot:",
		renderRules(expected.RuleSnapshot),
		"acceptance_item_evidence_matrix:",
		"  - id: demo.core",
		"    status: pass",
		"retirement_evidence_matrix: none",
	}, "\n")+"\n```\n")
}

func writePlanProcessWithoutReceipt(t *testing.T, repoRoot string, expected snapshot.Snapshot) {
	t.Helper()
	writeFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/demo.md"), "# plan\n\n```yaml\n"+strings.Join([]string{
		"spec_file_ref: " + expected.SpecFileRef,
		"spec_version_ref: " + expected.SpecVersionRef,
		"spec_fingerprint: " + expected.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + expected.AcceptanceBehaviorFingerprint,
		"unit_appendix_snapshot:",
		renderAppendix(expected.ModuleAppendixSnapshot),
		"rule_snapshot:",
		renderRules(expected.RuleSnapshot),
		"acceptance_item_plan_coverage:",
		"  - id: demo.core",
		"    coverage: implementation slice and verification target",
		"retirement_targets: none",
	}, "\n")+"\n```\n")
}

func requestFileFingerprint(t *testing.T, repoRoot, fileRef string) string {
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

func writeStableVerifyProcessWithoutReceipt(t *testing.T, repoRoot string, expected snapshot.Snapshot, mapping snapshot.RepositoryMappingEntry) {
	t.Helper()
	writeFile(t, filepath.Join(repoRoot, "docs/specs/_stable_verify_result/unit/demo.md"), "# stable verify\n\n```yaml\n"+strings.Join([]string{
		"object_type: unit",
		"object_ref: demo",
		"gate: unit_stable_verify",
		"decision: aligned",
		"allow_next: true",
		"next_command: unit_fork",
		"blocking_summary: none",
		"coverage_summary: current stable implementation",
		"truth_layer_ref: stable",
		"truth_file_ref: " + expected.SpecFileRef,
		"truth_version_ref: " + expected.SpecVersionRef,
		"truth_fingerprint: " + expected.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + expected.AcceptanceBehaviorFingerprint,
		"repository_mapping_snapshot:",
		"  file_ref: " + mapping.FileRef,
		"  version_ref: " + mapping.VersionRef,
		"  fingerprint: " + mapping.Fingerprint,
		"acceptance_item_set:",
		renderAcceptanceItems(expected.AcceptanceItemSet),
		"unit_appendix_snapshot:",
		renderAppendix(expected.ModuleAppendixSnapshot),
		"unit_snapshot: none",
		"rule_snapshot: none",
		"acceptance_item_evidence_matrix:",
		"  - id: demo.core",
		"    status: pass",
		"implementation_surface_refs: AgentCore/internal/demo",
		"evidence_refs: go test ./...",
	}, "\n")+"\n```\n")
}

func renderAcceptanceItems(entries []snapshot.AcceptanceItemEntry) string {
	lines := []string{}
	for _, entry := range entries {
		lines = append(lines,
			"  - id: "+entry.ID,
			"    verification_surface: "+entry.VerificationSurface,
			"    not_runnable_yet: "+entry.NotRunnableYet,
		)
	}
	return strings.Join(lines, "\n")
}

func renderAppendix(entries []snapshot.AppendixEntry) string {
	if len(entries) == 0 {
		return "  none"
	}
	lines := []string{}
	for _, entry := range entries {
		lines = append(lines,
			"  - file_ref: "+entry.FileRef,
			"    fingerprint: "+entry.Fingerprint,
		)
	}
	return strings.Join(lines, "\n")
}

func renderRules(entries []snapshot.RuleEntry) string {
	if len(entries) == 0 {
		return "  none"
	}
	lines := []string{}
	for _, entry := range entries {
		lines = append(lines,
			"  - rule_id: "+entry.RuleID,
			"    layer: "+entry.Layer,
			"    file_ref: "+entry.FileRef,
			"    version_ref: "+entry.VersionRef,
			"    fingerprint: "+entry.Fingerprint,
		)
	}
	return strings.Join(lines, "\n")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func sectionBetween(t *testing.T, content, start, end string) string {
	t.Helper()
	startIndex := strings.Index(content, start)
	if startIndex < 0 {
		t.Fatalf("missing section start %q:\n%s", start, content)
	}
	startIndex += len(start)
	endIndex := strings.Index(content[startIndex:], end)
	if endIndex < 0 {
		t.Fatalf("missing section end %q after %q:\n%s", end, start, content)
	}
	return content[startIndex : startIndex+endIndex]
}
