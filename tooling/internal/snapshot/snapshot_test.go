package snapshot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
)

const testAcceptanceSection = `## Testability / Acceptance Criteria

acceptance_item_set:
  - id: demo.core
    target: Demo behavior is accepted.
    verification_surface: internal_flow
    implementation_surface: AgentCore/internal/demo
    verification_method: Go test for the demo behavior.
    pass_condition: The demo behavior passes under the declared checks.
    not_runnable_yet: no
`

func TestRebuildCurrentCollectsAppendixAndSharedSnapshot(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateAppendixDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/candidate"))

	status := "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | note |\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)

	mainSpec := `---
id: demo
layer: candidate
version: 0.1.0
---

# Demo

` + testAcceptanceSection + `
See [appendix](./appendix/c_unit_demo_prompt.md).

## Rule Alignment

2. ` + "`rule_refs`:" + `
   - ` + "`c_b_rule_demo@0.2.0`" + `
`
	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), mainSpec)

	appendix := `---
unit: demo
layer: candidate
spec_version_ref: c_unit_demo@0.1.0
---

# Appendix
`
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateAppendixDir), "c_unit_demo_prompt.md"), appendix)

	shared := `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.2.0
bound_objects:
  - unit:demo
---

# Shared
`
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md"), shared)

	system := `---
rule_id: g_rule_repository_baseline
rule_scope: global
layer: stable
rule_version: 1.1.0
bound_objects: all_units
---

# Rule
`
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_g_rule_repository_baseline.md"), system)

	result, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	if result.SpecFileRef != mainSpecRef {
		t.Fatalf("unexpected spec file ref: %s", result.SpecFileRef)
	}
	if result.SpecVersionRef != "c_unit_demo@0.1.0" {
		t.Fatalf("unexpected spec version ref: %s", result.SpecVersionRef)
	}
	if len(result.ModuleAppendixSnapshot) != 1 {
		t.Fatalf("expected one appendix snapshot entry, got %d", len(result.ModuleAppendixSnapshot))
	}
	if result.ModuleAppendixSnapshot[0].AppendixRef != "c_unit_demo_prompt@c_unit_demo@0.1.0" {
		t.Fatalf("unexpected appendix ref: %s", result.ModuleAppendixSnapshot[0].AppendixRef)
	}
	if len(result.RuleSnapshot) != 2 {
		t.Fatalf("expected global and explicit rule snapshot entries, got %d", len(result.RuleSnapshot))
	}
	if result.RuleSnapshot[1].VersionRef != "c_b_rule_demo@0.2.0" {
		t.Fatalf("unexpected rule version ref: %s", result.RuleSnapshot[1].VersionRef)
	}
}

func TestRebuildCurrentCollectsEquivalentAppendixSubdirAndPlainFieldNames(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateDir), "support"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/candidate"))

	status := "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | note |\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)

	mainSpec := `---
id: demo
layer: candidate
version: 0.1.0
---

# Demo

` + testAcceptanceSection + `
See [support](./support/c_unit_demo_prompt.md).

## Rule Alignment

2. rule_refs:
   - c_b_rule_demo@0.2.0
`
	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), mainSpec)

	appendix := `---
unit: demo
layer: candidate
spec_version_ref: c_unit_demo@0.1.0
---

# Appendix
`
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateDir), "support", "c_unit_demo_prompt.md"), appendix)

	shared := `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.2.0
bound_objects:
  - unit:demo
---

# Shared
`
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md"), shared)

	system := `---
rule_id: g_rule_repository_baseline
rule_scope: global
layer: stable
rule_version: 1.1.0
bound_objects: all_units
---

# Rule
`
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_g_rule_repository_baseline.md"), system)

	result, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	if len(result.ModuleAppendixSnapshot) != 1 {
		t.Fatalf("expected one appendix snapshot entry, got %d", len(result.ModuleAppendixSnapshot))
	}
	if result.ModuleAppendixSnapshot[0].FileRef != "docs/specs/units/candidate/support/c_unit_demo_prompt.md" {
		t.Fatalf("unexpected appendix file ref: %s", result.ModuleAppendixSnapshot[0].FileRef)
	}
	if len(result.RuleSnapshot) != 2 || result.RuleSnapshot[1].VersionRef != "c_b_rule_demo@0.2.0" {
		t.Fatalf("unexpected rule snapshot: %+v", result.RuleSnapshot)
	}
}

func TestRebuildCurrentRejectsUnsortedRuleRefs(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/candidate"))

	status := "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | note |\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), `---
id: demo
layer: candidate
version: 0.1.0
---

# Demo

`+testAcceptanceSection+`
## Rule Alignment

2. rule_refs:
   - c_b_rule_zeta@0.1.0
   - c_b_rule_alpha@0.1.0
`)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_alpha.md"), `---
rule_id: shared_alpha
rule_scope: bound
layer: candidate
rule_version: 0.1.0
bound_objects:
  - unit:demo
---

# Shared Alpha
`)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_zeta.md"), `---
rule_id: shared_zeta
rule_scope: bound
layer: candidate
rule_version: 0.1.0
bound_objects:
  - unit:demo
---

# Shared Zeta
`)

	_, err = RebuildCurrent(repoRoot, "demo")
	if err == nil || !strings.Contains(err.Error(), "rule_refs must be sorted") {
		t.Fatalf("expected unsorted rule_refs error, got %v", err)
	}
}

func TestRebuildCurrentIgnoresRawAppendixPathLiteral(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateAppendixDir)))

	status := "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | note |\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), `---
	id: demo
	layer: candidate
	version: 0.1.0
---

# Demo

`+testAcceptanceSection+`
Use appendix path `+"`./appendix/c_unit_demo_prompt.md`"+` for detailed prompts.

## Rule Alignment

2. rule_refs: none
`)

	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateAppendixDir), "c_unit_demo_prompt.md"), `---
unit: demo
layer: candidate
spec_version_ref: c_unit_demo@0.1.0
---

	# Appendix
`)

	result, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	if len(result.ModuleAppendixSnapshot) != 0 {
		t.Fatalf("expected raw path literal to be ignored, got %+v", result.ModuleAppendixSnapshot)
	}
}

func TestRebuildCurrentRejectsRootDirectoryAppendixDrift(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateDir)))

	status := "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | note |\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), `---
id: demo
layer: candidate
version: 0.1.0
---

# Demo

`+testAcceptanceSection+`
See [support](./c_unit_demo_prompt.md).

## Rule Alignment

2. rule_refs: none
`)

	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateDir), "c_unit_demo_prompt.md"), `---
unit: demo
layer: candidate
spec_version_ref: c_unit_demo@0.1.0
---

# Drift
`)

	_, err = RebuildCurrent(repoRoot, "demo")
	if err == nil || !strings.Contains(err.Error(), "directory drift") {
		t.Fatalf("expected directory drift error, got %v", err)
	}
}

func TestValidateProcessFileRejectsMissingRequiredSnapshotField(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	writeCheckProcessFile(t, repoRoot, strings.Join([]string{
		"object_type: unit",
		"object_ref: demo",
		"gate: unit_check",
		"decision: pass",
		"allow_next: true",
		"next_command: unit_plan",
		"blocking_summary: none",
		"coverage_summary: current candidate",
		"truth_layer_ref: candidate",
		"truth_file_ref: docs/specs/units/candidate/c_unit_demo.md",
		"truth_version_ref: c_unit_demo@0.1.0",
		"unit_appendix_snapshot: none",
		"rule_snapshot: none",
	}, "\n"))

	result, err := ValidateProcessFile(repoRoot, "demo", "check")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid result, got valid")
	}
	if !containsMismatch(result.Mismatches, "missing required field: truth_fingerprint") {
		t.Fatalf("expected missing truth_fingerprint mismatch, got %+v", result.Mismatches)
	}
}

func TestValidateProcessFileAcceptsExplicitNoneSnapshots(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}

	writeCheckProcessFile(t, repoRoot, renderFormalCheckProcessBody(expected))

	result, err := ValidateProcessFile(repoRoot, "demo", "check")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid result, got mismatches %+v", result.Mismatches)
	}
}

func TestValidateProcessFileAcceptsSnapshotFieldsWithoutYAMLFence(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}

	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), strings.Join([]string{
		"# check",
		"",
		renderFormalCheckProcessBody(expected),
		"",
	}, "\n"))

	result, err := ValidateProcessFile(repoRoot, "demo", "check")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid result, got mismatches %+v", result.Mismatches)
	}
}

func TestValidateProcessFileAcceptsMarkdownBulletSnapshotFormat(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateAppendixDir)))

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), `---
id: demo
layer: candidate
version: 0.1.0
---

# Demo

`+testAcceptanceSection+`
See [appendix](./appendix/c_unit_demo_prompt.md).

## Rule Alignment

2. rule_refs: none
`)
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateAppendixDir), "c_unit_demo_prompt.md"), `---
unit: demo
layer: candidate
spec_version_ref: c_unit_demo@0.1.0
---

# Appendix
`)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}

	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), strings.Join([]string{
		"# demo unit_check snapshot",
		"",
		"## Check Result Snapshot",
		"",
		"- `object_type`: `unit`",
		"- `object_ref`: `" + expected.Module + "`",
		"- `gate`: `unit_check`",
		"- `decision`: `pass`",
		"- `allow_next`: `true`",
		"- `next_command`: `unit_plan`",
		"- `blocking_summary`: `none`",
		"- `coverage_summary`: `current candidate`",
		"- `truth_layer_ref`: `" + expected.TruthLayerRef + "`",
		"- `truth_file_ref`: `" + expected.SpecFileRef + "`",
		"- `truth_version_ref`: `" + expected.SpecVersionRef + "`",
		"- `truth_fingerprint`: `" + expected.SpecFingerprint + "`",
		"- `acceptance_item_set`:",
		"  - `id`: `demo.core`",
		"    `verification_surface`: `internal_flow`",
		"    `not_runnable_yet`: `no`",
		"- `unit_appendix_snapshot`:",
		"  - `file_ref`: `" + expected.ModuleAppendixSnapshot[0].FileRef + "`",
		"  - `appendix_ref`: `" + expected.ModuleAppendixSnapshot[0].AppendixRef + "`",
		"  - `fingerprint`: `" + expected.ModuleAppendixSnapshot[0].Fingerprint + "`",
		"- `rule_snapshot`: `none`",
		"",
	}, "\n"))

	result, err := ValidateProcessFile(repoRoot, "demo", "check")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid result, got mismatches %+v", result.Mismatches)
	}
}

func TestValidateProcessFileRejectsUnexpectedGate(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}

	writeCheckProcessFile(t, repoRoot, strings.Replace(renderFormalCheckProcessBody(expected), "gate: unit_check", "gate: unit_plan", 1))

	result, err := ValidateProcessFile(repoRoot, "demo", "check")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid result, got valid")
	}
	if !containsMismatch(result.Mismatches, "gate mismatch: actual=unit_plan expected=unit_check") {
		t.Fatalf("expected gate mismatch, got %+v", result.Mismatches)
	}
}

func TestValidateProcessFileRejectsAcceptanceItemSetDrift(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}

	body := strings.Replace(renderFormalCheckProcessBody(expected), "id: demo.core", "id: demo.changed", 1)
	writeCheckProcessFile(t, repoRoot, body)

	result, err := ValidateProcessFile(repoRoot, "demo", "check")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid result, got valid")
	}
	if !containsMismatch(result.Mismatches, "acceptance_item_set mismatch: actual=demo.changed|internal_flow|no expected=demo.core|internal_flow|no") {
		t.Fatalf("expected acceptance item set mismatch, got %+v", result.Mismatches)
	}
}

func TestValidateProcessFileAcceptsPlanSchemaWithoutGateFields(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}

	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/demo.md"), "# plan\n\n```yaml\n"+renderFormalPlanProcessBody(expected)+"\n```\n")

	result, err := ValidateProcessFile(repoRoot, "demo", "plan")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid plan result, got mismatches %+v", result.Mismatches)
	}
}

func TestValidateProcessFileRejectsPlanCoverageGap(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}

	body := strings.Replace(renderFormalPlanProcessBody(expected), "id: demo.core", "id: demo.other", 1)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/demo.md"), "# plan\n\n```yaml\n"+body+"\n```\n")

	result, err := ValidateProcessFile(repoRoot, "demo", "plan")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid result, got valid")
	}
	if !containsMismatch(result.Mismatches, "acceptance_item_plan_coverage unknown id: demo.other") {
		t.Fatalf("expected unknown coverage id mismatch, got %+v", result.Mismatches)
	}
	if !containsMismatch(result.Mismatches, "acceptance_item_plan_coverage missing id: demo.core") {
		t.Fatalf("expected missing coverage id mismatch, got %+v", result.Mismatches)
	}
}

func TestValidateProcessFileRejectsVerifyEvidenceGap(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit"))

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}

	body := strings.Replace(renderFormalVerifyProcessBody(expected), "status: pass", "status: skipped", 1)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit/demo.md"), "# verify\n\n```yaml\n"+body+"\n```\n")

	result, err := ValidateProcessFile(repoRoot, "demo", "verify")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid result, got valid")
	}
	if !containsMismatch(result.Mismatches, "acceptance_item_evidence_matrix invalid status for demo.core: skipped") {
		t.Fatalf("expected invalid evidence status mismatch, got %+v", result.Mismatches)
	}
}

func TestRebuildCurrentRejectsUnstructuredAcceptanceSection(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), `---
id: demo
layer: candidate
version: 0.1.0
---

# Demo

## Testability / Acceptance Criteria

1. The demo behavior works.

## Rule Alignment

2. rule_refs: none
`)

	_, err = RebuildCurrent(repoRoot, "demo")
	if err == nil || !strings.Contains(err.Error(), "acceptance section must define acceptance_item_set") {
		t.Fatalf("expected unstructured acceptance section error, got %v", err)
	}
}

func TestRebuildCurrentRejectsCandidateWithoutAcceptanceSection(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), `---
id: demo
layer: candidate
version: 0.1.0
---

# Demo

## Rule Alignment

2. rule_refs: none
`)

	_, err = RebuildCurrent(repoRoot, "demo")
	if err == nil || !strings.Contains(err.Error(), "main Spec must define Testability / Acceptance Criteria with acceptance_item_set") {
		t.Fatalf("expected missing acceptance section error, got %v", err)
	}
}

func TestRebuildCurrentAllowsHistoricalStableProseAcceptance(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.StableDir)))

	status := "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | note |\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)

	mainSpecRef, err := specpaths.MainSpecFileRef("stable", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), `---
id: demo
layer: stable
version: 1.0.0
---

# Demo

## Testability / Acceptance Criteria

1. The historical stable behavior works.

## Rule Alignment

2. rule_refs: none
`)

	result, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	if len(result.AcceptanceItemSet) != 0 {
		t.Fatalf("expected historical stable prose acceptance to produce no snapshot items, got %+v", result.AcceptanceItemSet)
	}
}

func TestRebuildCurrentRejectsInvalidAcceptanceSurface(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Replace(string(content), "verification_surface: internal_flow", "verification_surface: unit_test", 1))

	_, err = RebuildCurrent(repoRoot, "demo")
	if err == nil || !strings.Contains(err.Error(), "verification_surface for demo.core must be one of") {
		t.Fatalf("expected invalid verification_surface error, got %v", err)
	}
}

func TestRebuildCurrentRejectsEmptyRuleRefsList(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), `---
id: demo
layer: candidate
version: 0.1.0
---

# Demo

`+testAcceptanceSection+`
## Rule Alignment

2. rule_refs:
`)

	_, err = RebuildCurrent(repoRoot, "demo")
	if err == nil || !strings.Contains(err.Error(), "must not be an empty list") {
		t.Fatalf("expected empty-list error, got %v", err)
	}
}

func TestRebuildCurrentRejectsDuplicateRuleRefs(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/candidate"))

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), `---
id: demo
layer: candidate
version: 0.1.0
---

# Demo

`+testAcceptanceSection+`
## Rule Alignment

2. rule_refs:
   - c_b_rule_demo@0.1.0
   - c_b_rule_demo@0.1.0
`)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md"), `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.1.0
bound_objects:
  - unit:demo
---

# Shared
`)

	_, err = RebuildCurrent(repoRoot, "demo")
	if err == nil || !strings.Contains(err.Error(), "duplicate item") {
		t.Fatalf("expected duplicate-item error, got %v", err)
	}
}

func TestRebuildCurrentRejectsRuleVersionMismatch(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/candidate"))

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), `---
id: demo
layer: candidate
version: 0.1.0
---

# Demo

`+testAcceptanceSection+`
## Rule Alignment

2. rule_refs:
   - c_b_rule_demo@0.1.0
`)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md"), `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.2.0
bound_objects:
  - unit:demo
---

# Shared
`)

	_, err = RebuildCurrent(repoRoot, "demo")
	if err == nil || !strings.Contains(err.Error(), "does not match frontmatter rule_version") {
		t.Fatalf("expected shared-version mismatch error, got %v", err)
	}
}

func TestRebuildCurrentRejectsStableModuleBindingCandidateShared(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.StableDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/candidate"))

	status := "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | note |\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)

	mainSpecRef, err := specpaths.MainSpecFileRef("stable", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), `---
id: demo
layer: stable
version: 1.0.0
---

# Demo

## Rule Alignment

2. rule_refs:
   - c_b_rule_demo@0.1.0
`)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md"), `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.1.0
bound_objects:
  - unit:demo
---

# Shared
`)

	_, err = RebuildCurrent(repoRoot, "demo")
	if err == nil || !strings.Contains(err.Error(), "stable-layer object binding must use an s_ rule ref") {
		t.Fatalf("expected stable-layer binding error, got %v", err)
	}
}

func TestRebuildCurrentIncludesStableGlobalRule(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	system := `---
rule_id: g_rule_repository_baseline
rule_scope: global
layer: stable
rule_version: 1.1.0
bound_objects: all_units
---

# Rule
`
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_g_rule_repository_baseline.md"), system)

	result, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	if len(result.RuleSnapshot) != 1 {
		t.Fatalf("expected one stable global rule, got %d", len(result.RuleSnapshot))
	}
	if result.RuleSnapshot[0].VersionRef != "s_g_rule_repository_baseline@1.1.0" {
		t.Fatalf("unexpected global rule version ref: %s", result.RuleSnapshot[0].VersionRef)
	}
}

func setupSnapshotValidationRepo(t *testing.T, repoRoot string) {
	t.Helper()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_plans/active"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_plans/draft"))

	status := "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | note |\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)

	mainSpec := `---
id: demo
layer: candidate
version: 0.1.0
---

# Demo

## Rule Alignment

2. rule_refs: none

## Testability / Acceptance Criteria

acceptance_item_set:
  - id: demo.core
    target: Demo behavior is accepted.
    verification_surface: internal_flow
    implementation_surface: AgentCore/internal/demo
    verification_method: Go test for the demo behavior.
    pass_condition: The demo behavior passes under the declared checks.
    not_runnable_yet: no
`
	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), mainSpec)
}

func writeCheckProcessFile(t *testing.T, repoRoot, yamlBody string) {
	t.Helper()
	content := "# check\n\n```yaml\n" + yamlBody + "\n```\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md"), content)
}

func renderFormalCheckProcessBody(expected Snapshot) string {
	return strings.Join([]string{
		"object_type: unit",
		"object_ref: " + expected.Module,
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
		"acceptance_item_set:",
		renderAcceptanceItemSetForTest(expected.AcceptanceItemSet),
		"unit_appendix_snapshot: none",
		"rule_snapshot: none",
	}, "\n")
}

func renderFormalPlanProcessBody(expected Snapshot) string {
	return strings.Join([]string{
		"spec_file_ref: " + expected.SpecFileRef,
		"spec_version_ref: " + expected.SpecVersionRef,
		"spec_fingerprint: " + expected.SpecFingerprint,
		"unit_appendix_snapshot: none",
		"rule_snapshot: none",
		"acceptance_item_plan_coverage:",
		renderAcceptancePlanCoverageForTest(expected.AcceptanceItemSet),
	}, "\n")
}

func renderFormalVerifyProcessBody(expected Snapshot) string {
	return strings.Join([]string{
		"object_type: unit",
		"object_ref: " + expected.Module,
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
		"acceptance_item_set:",
		renderAcceptanceItemSetForTest(expected.AcceptanceItemSet),
		"unit_appendix_snapshot: none",
		"verification_scope_ref: current candidate",
		"rule_snapshot: none",
		"acceptance_item_evidence_matrix:",
		renderAcceptanceEvidenceMatrixForTest(expected.AcceptanceItemSet),
	}, "\n")
}

func renderAcceptanceItemSetForTest(entries []AcceptanceItemEntry) string {
	if len(entries) == 0 {
		return "  none"
	}
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

func renderAcceptancePlanCoverageForTest(entries []AcceptanceItemEntry) string {
	if len(entries) == 0 {
		return "  none"
	}
	lines := []string{}
	for _, entry := range entries {
		lines = append(lines,
			"  - id: "+entry.ID,
			"    coverage: implementation slice and verification target",
		)
	}
	return strings.Join(lines, "\n")
}

func renderAcceptanceEvidenceMatrixForTest(entries []AcceptanceItemEntry) string {
	if len(entries) == 0 {
		return "  none"
	}
	lines := []string{}
	for _, entry := range entries {
		status := "pass"
		if entry.NotRunnableYet == "yes" {
			status = "not_runnable_yet"
		}
		lines = append(lines,
			"  - id: "+entry.ID,
			"    status: "+status,
		)
	}
	return strings.Join(lines, "\n")
}

func containsMismatch(mismatches []string, target string) bool {
	for _, mismatch := range mismatches {
		if mismatch == target {
			return true
		}
	}
	return false
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
