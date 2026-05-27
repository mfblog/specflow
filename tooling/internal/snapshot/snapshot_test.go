package snapshot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/testfixtures"
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
candidate_intent: change
source_basis: new_design
evidence_appendix_ref: none
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
	if len(result.RuleSnapshot) != 2 {
		t.Fatalf("expected global and explicit rule snapshot entries, got %d", len(result.RuleSnapshot))
	}
	if result.RuleSnapshot[1].VersionRef != "c_b_rule_demo@0.2.0" {
		t.Fatalf("unexpected rule version ref: %s", result.RuleSnapshot[1].VersionRef)
	}
}

func TestRebuildCurrentCollectsStableUnitRefs(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.StableDir)))

	status := "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | note |\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_auth.md"), `---
id: auth
layer: stable
version: 1.0.0
---

# Auth
`)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), `---
id: demo
layer: candidate
version: 0.1.0
unit_refs:
  - s_unit_auth@1.0.0
---

# Demo

2. rule_refs: none
`+testAcceptanceSection)

	result, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	if len(result.UnitSnapshot) != 1 {
		t.Fatalf("expected one unit dependency snapshot, got %+v", result.UnitSnapshot)
	}
	if result.UnitSnapshot[0].ObjectRef != "auth" || result.UnitSnapshot[0].VersionRef != "s_unit_auth@1.0.0" {
		t.Fatalf("unexpected unit dependency snapshot: %+v", result.UnitSnapshot[0])
	}
}

func TestRebuildCurrentRejectsCandidateUnitRefs(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateDir)))

	status := "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | note |\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), `---
id: demo
layer: candidate
version: 0.1.0
unit_refs:
  - c_unit_auth@0.1.0
---

# Demo

2. rule_refs: none
`+testAcceptanceSection)

	_, err := RebuildCurrent(repoRoot, "demo")
	if err == nil || !strings.Contains(err.Error(), "unit_refs must reference stable units") {
		t.Fatalf("expected candidate unit_refs rejection, got %v", err)
	}
}

func TestRebuildCurrentCollectsEvidenceAppendixRef(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateAppendixDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/stable"))

	status := "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | note |\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)

	mainSpec := `---
id: demo
layer: candidate
version: 0.1.0
candidate_intent: change
source_basis: mixed
evidence_appendix_ref: docs/specs/units/candidate/appendix/c_unit_demo_evidence.md
---

# Demo

` + testAcceptanceSection + `
## Rule Alignment

2. rule_refs: none
`
	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), mainSpec)

	appendix := `---
unit: demo
layer: candidate
---

# Evidence
`
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateAppendixDir), "c_unit_demo_evidence.md"), appendix)

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
		t.Fatalf("expected one evidence appendix snapshot entry, got %d", len(result.ModuleAppendixSnapshot))
	}
	if result.ModuleAppendixSnapshot[0].FileRef != "docs/specs/units/candidate/appendix/c_unit_demo_evidence.md" {
		t.Fatalf("unexpected appendix file ref: %s", result.ModuleAppendixSnapshot[0].FileRef)
	}
}

func TestRebuildCurrentAcceptsRepairCandidateIntent(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/units/stable"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/units/candidate"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit"))

	status := "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n| `unit` | `demo` | `yes` | `yes` | `candidate` | `unit_check` | repair |\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_demo.md"), `---
id: demo
layer: stable
version: 0.1.0
---

# Demo

`+testAcceptanceSection)

	mainSpec := `---
id: demo
layer: candidate
version: 0.1.1
candidate_intent: repair
repair_basis: s_unit_demo@0.1.0
source_basis: new_design
evidence_appendix_ref: none
---

# Demo

## Repair Scope

1. Restore ` + "`demo.core`" + ` to the stable behavior recorded by ` + "`s_unit_demo@0.1.0`" + `.

` + testAcceptanceSection + `
## Rule Alignment

2. rule_refs: none
`
	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), mainSpec)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	if expected.SpecVersionRef != "c_unit_demo@0.1.1" {
		t.Fatalf("unexpected spec version ref: %s", expected.SpecVersionRef)
	}

	writeCheckProcessFile(t, repoRoot, renderFormalCheckProcessBody(expected))
	result, err := ValidateProcessFile(repoRoot, "demo", "check")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected repair candidate snapshot to validate, got %+v", result)
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
candidate_intent: change
source_basis: new_design
evidence_appendix_ref: none
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
candidate_intent: change
source_basis: new_design
evidence_appendix_ref: none
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
candidate_intent: change
source_basis: new_design
evidence_appendix_ref: none
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

func TestValidateProcessFileRejectsNonPassCheckAndVerifyDecisions(t *testing.T) {
	for _, tc := range []struct {
		processKind string
		decision    string
	}{
		{processKind: "check", decision: "blocked"},
		{processKind: "check", decision: "fix_required"},
		{processKind: "verify", decision: "blocked"},
		{processKind: "verify", decision: "fix_required"},
	} {
		t.Run(tc.processKind+"/"+tc.decision, func(t *testing.T) {
			repoRoot := t.TempDir()
			setupSnapshotValidationRepo(t, repoRoot)
			if tc.processKind == "verify" {
				mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit"))
			}
			expected, err := RebuildCurrent(repoRoot, "demo")
			if err != nil {
				t.Fatalf("RebuildCurrent: %v", err)
			}

			body := renderFormalCheckProcessBody(expected)
			if tc.processKind == "verify" {
				body = renderFormalVerifyProcessBody(expected)
			}
			body = strings.Replace(body, "decision: pass", "decision: "+tc.decision, 1)
			if tc.processKind == "check" {
				writeCheckProcessFile(t, repoRoot, body)
			} else {
				mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit/demo.md"), "# verify\n\n```yaml\n"+body+"\n```\n")
			}

			result, err := ValidateProcessFile(repoRoot, "demo", tc.processKind)
			if err != nil {
				t.Fatalf("ValidateProcessFile: %v", err)
			}
			if result.Valid {
				t.Fatalf("expected invalid result, got valid")
			}
			want := "decision mismatch: actual=" + tc.decision + " expected=pass"
			if !containsMismatch(result.Mismatches, want) {
				t.Fatalf("expected %q mismatch, got %+v", want, result.Mismatches)
			}
		})
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
		"  - `fingerprint`: `" + expected.ModuleAppendixSnapshot[0].Fingerprint + "`",
		"- `rule_snapshot`: `none`",
		"- `evaluation_mode`: `independent`",
		"- `reviewer_result`: `pass`",
		"- `reviewer_context`: `minimal_context`",
		"- `review_input_refs`: `" + expected.SpecFileRef + "`",
		"- `review_findings`: `none`",
		"- `human_decision_refs`: `none`",
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

func TestValidateProcessFileRejectsUnsupportedAppendixSnapshotField(t *testing.T) {
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
---

# Appendix
`)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	unsupportedField := "label"

	writeCheckProcessFile(t, repoRoot, strings.Join([]string{
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
		renderAcceptanceItemSetForTest(expected.AcceptanceItemSet),
		"unit_appendix_snapshot:",
		"  - file_ref: " + expected.ModuleAppendixSnapshot[0].FileRef,
		"    " + unsupportedField + ": c_unit_demo_prompt",
		"    fingerprint: " + expected.ModuleAppendixSnapshot[0].Fingerprint,
		"rule_snapshot: none",
		renderIndependentEvaluationReceiptForTest(expected.SpecFileRef),
	}, "\n"))

	result, err := ValidateProcessFile(repoRoot, "demo", "check")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid result, got valid")
	}
	if !containsMismatch(result.Mismatches, "unsupported field: unit_appendix_snapshot."+unsupportedField) {
		t.Fatalf("expected unsupported appendix snapshot field mismatch, got %+v", result.Mismatches)
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

func TestValidateProcessFileClassifiesTextDriftRequiresFreshnessReview(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	writeCheckProcessFile(t, repoRoot, renderFormalCheckProcessBody(expected))
	replaceCandidateSpecText(t, repoRoot, "# Demo\n", "# Demo\n\nEditorial note only.\n")

	result, err := ValidateProcessFile(repoRoot, "demo", "check")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid result, got valid")
	}
	if result.FreshnessImpact != FreshnessTextDrift || result.EvidenceReuse != EvidenceReusePendingReview {
		t.Fatalf("expected pending text drift, got impact=%s reuse=%s mismatches=%+v", result.FreshnessImpact, result.EvidenceReuse, result.Mismatches)
	}
	if result.FailureLayer != "freshness_layer" || result.NextCommand != "" {
		t.Fatalf("expected non-rerouting freshness layer, got layer=%s next=%s", result.FailureLayer, result.NextCommand)
	}
	if !containsMismatch(result.Mismatches, "missing required freshness field: freshness_impact") {
		t.Fatalf("expected missing freshness receipt mismatch, got %+v", result.Mismatches)
	}
}

func TestValidateProcessFileAllowsAcceptedTextDrift(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	replaceCandidateSpecText(t, repoRoot, "# Demo\n", "# Demo\n\nEditorial note only.\n")
	current, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent after edit: %v", err)
	}
	body := strings.Join([]string{
		renderFormalCheckProcessBody(expected),
		renderFreshnessReceiptForTest(current.SpecFingerprint, current.SpecFileRef),
	}, "\n")
	writeCheckProcessFile(t, repoRoot, body)

	result, err := ValidateProcessFile(repoRoot, "demo", "check")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected accepted text drift, got invalid: impact=%s reuse=%s mismatches=%+v", result.FreshnessImpact, result.EvidenceReuse, result.Mismatches)
	}
	if result.FreshnessImpact != FreshnessTextDrift || result.EvidenceReuse != EvidenceReuseAccepted {
		t.Fatalf("expected accepted text drift, got impact=%s reuse=%s", result.FreshnessImpact, result.EvidenceReuse)
	}
}

func TestValidateProcessFileRejectsInvalidFreshnessReceipt(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	replaceCandidateSpecText(t, repoRoot, "# Demo\n", "# Demo\n\nEditorial note only.\n")
	current, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent after edit: %v", err)
	}

	for _, tc := range []struct {
		name        string
		old         string
		replacement string
		want        string
	}{
		{
			name:        "blocked reviewer",
			old:         "freshness_reviewer_result: pass",
			replacement: "freshness_reviewer_result: blocked",
			want:        "freshness_reviewer_result mismatch: actual=blocked expected=pass",
		},
		{
			name:        "review findings",
			old:         "freshness_review_findings: none",
			replacement: "freshness_review_findings: changed-meaning",
			want:        "freshness_review_findings mismatch: actual=changed-meaning expected=none",
		},
		{
			name:        "wrong current fingerprint",
			old:         "freshness_current_fingerprint: " + current.SpecFingerprint,
			replacement: "freshness_current_fingerprint: stale-fingerprint",
			want:        "freshness_current_fingerprint mismatch: actual=stale-fingerprint expected=" + current.SpecFingerprint,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := strings.Join([]string{
				renderFormalCheckProcessBody(expected),
				renderFreshnessReceiptForTest(current.SpecFingerprint, current.SpecFileRef),
			}, "\n")
			body = strings.Replace(body, tc.old, tc.replacement, 1)
			writeCheckProcessFile(t, repoRoot, body)

			result, err := ValidateProcessFile(repoRoot, "demo", "check")
			if err != nil {
				t.Fatalf("ValidateProcessFile: %v", err)
			}
			if result.Valid {
				t.Fatalf("expected invalid result, got valid")
			}
			if !containsMismatch(result.Mismatches, tc.want) {
				t.Fatalf("expected %q, got %+v", tc.want, result.Mismatches)
			}
		})
	}
}

func TestValidateProcessFileClassifiesSemanticDrift(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	writeCheckProcessFile(t, repoRoot, renderFormalCheckProcessBody(expected)+"\n"+renderFreshnessReceiptForTest("unused", expected.SpecFileRef))
	replaceCandidateSpecText(t, repoRoot, "pass_condition: The demo behavior passes under the declared checks.", "pass_condition: The demo behavior passes under new semantic checks.")

	result, err := ValidateProcessFile(repoRoot, "demo", "check")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid semantic drift, got valid")
	}
	if result.FreshnessImpact != FreshnessSemanticDrift || result.EvidenceReuse != EvidenceReuseNotEligible {
		t.Fatalf("expected semantic drift, got impact=%s reuse=%s mismatches=%+v", result.FreshnessImpact, result.EvidenceReuse, result.Mismatches)
	}
}

func TestValidateProcessFileClassifiesAcceptanceDrift(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	writeCheckProcessFile(t, repoRoot, renderFormalCheckProcessBody(expected))
	replaceCandidateSpecText(t, repoRoot, "id: demo.core", "id: demo.changed")

	result, err := ValidateProcessFile(repoRoot, "demo", "check")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid acceptance drift, got valid")
	}
	if result.FreshnessImpact != FreshnessAcceptanceDrift {
		t.Fatalf("expected acceptance drift, got impact=%s mismatches=%+v", result.FreshnessImpact, result.Mismatches)
	}
}

func TestValidateProcessFileClassifiesDependencyDrift(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	writeCheckProcessFile(t, repoRoot, renderFormalCheckProcessBody(expected))
	writeStableGlobalRuleForFreshnessTest(t, repoRoot)

	result, err := ValidateProcessFile(repoRoot, "demo", "check")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid dependency drift, got valid")
	}
	if result.FreshnessImpact != FreshnessDependencyDrift {
		t.Fatalf("expected dependency drift, got impact=%s mismatches=%+v", result.FreshnessImpact, result.Mismatches)
	}
}

func TestValidateProcessFileClassifiesUnknownDriftWithoutBehaviorFingerprint(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	body := strings.Replace(renderFormalCheckProcessBody(expected), "acceptance_behavior_fingerprint: "+expected.AcceptanceBehaviorFingerprint+"\n", "", 1)
	writeCheckProcessFile(t, repoRoot, body)
	replaceCandidateSpecText(t, repoRoot, "# Demo\n", "# Demo\n\nEditorial note only.\n")

	result, err := ValidateProcessFile(repoRoot, "demo", "check")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid unknown drift, got valid")
	}
	if result.FreshnessImpact != FreshnessUnknownDrift || result.EvidenceReuse != EvidenceReuseNotEligible {
		t.Fatalf("expected unknown drift, got impact=%s reuse=%s mismatches=%+v", result.FreshnessImpact, result.EvidenceReuse, result.Mismatches)
	}
}

func TestValidateProcessFileRejectsMissingIndependentEvaluationReceipt(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}

	body := strings.Replace(renderFormalCheckProcessBody(expected), "evaluation_mode: independent\n", "", 1)
	writeCheckProcessFile(t, repoRoot, body)

	result, err := ValidateProcessFile(repoRoot, "demo", "check")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid result, got valid")
	}
	if !containsMismatch(result.Mismatches, "missing required field: evaluation_mode") {
		t.Fatalf("expected missing evaluation_mode mismatch, got %+v", result.Mismatches)
	}
}

func TestValidateProcessFileRejectsBlockedIndependentReviewer(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}

	body := strings.Replace(renderFormalPlanProcessBody(expected), "reviewer_result: pass", "reviewer_result: blocked", 1)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/demo.md"), "# plan\n\n```yaml\n"+body+"\n```\n")

	result, err := ValidateProcessFile(repoRoot, "demo", "plan")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid result, got valid")
	}
	if !containsMismatch(result.Mismatches, "reviewer_result mismatch: actual=blocked expected=pass") {
		t.Fatalf("expected reviewer_result mismatch, got %+v", result.Mismatches)
	}
}

func TestValidateProcessFileRejectsPassWithReviewFindings(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit"))

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}

	body := strings.Replace(renderFormalVerifyProcessBody(expected), "review_findings: none", "review_findings: missing-acceptance-boundary", 1)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit/demo.md"), "# verify\n\n```yaml\n"+body+"\n```\n")

	result, err := ValidateProcessFile(repoRoot, "demo", "verify")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid result, got valid")
	}
	if !containsMismatch(result.Mismatches, "review_findings mismatch: actual=missing-acceptance-boundary expected=none") {
		t.Fatalf("expected review_findings mismatch, got %+v", result.Mismatches)
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

func TestValidateProcessFileClassifiesPlanSchemaGapAsPlanLayer(t *testing.T) {
	repoRoot := t.TempDir()
	setupSnapshotValidationRepo(t, repoRoot)

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	body := strings.Replace(renderFormalPlanProcessBody(expected), "    coverage: implementation slice and verification target", "", 1)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/demo.md"), "# plan\n\n```yaml\n"+body+"\n```\n")

	result, err := ValidateProcessFileForObject(repoRoot, "unit", "demo", "plan")
	if err != nil {
		t.Fatalf("ValidateProcessFileForObject: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid result, got valid")
	}
	if result.FailureLayer != "plan_layer" || result.NextCommand != "unit_plan" {
		t.Fatalf("expected plan_layer/unit_plan, got %s/%s mismatches=%v", result.FailureLayer, result.NextCommand, result.Mismatches)
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

func TestValidateProcessFileAcceptsStableVerifyEvidence(t *testing.T) {
	repoRoot := t.TempDir()
	setupStableSnapshotValidationRepo(t, repoRoot)
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_stable_verify_result/unit"))

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	mapping, err := BuildRepositoryMappingSnapshot(repoRoot)
	if err != nil {
		t.Fatalf("BuildRepositoryMappingSnapshot: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_stable_verify_result/unit/demo.md"), "# stable verify\n\n```yaml\n"+renderFormalStableVerifyProcessBody(expected, mapping, "aligned")+"\n```\n")

	result, err := ValidateProcessFile(repoRoot, "demo", "stable_verify")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid result, got mismatches=%v", result.Mismatches)
	}
}

func TestValidateProcessFileRejectsStableVerifySnapshotDrift(t *testing.T) {
	repoRoot := t.TempDir()
	setupStableSnapshotValidationRepo(t, repoRoot)
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_stable_verify_result/unit"))

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	mapping, err := BuildRepositoryMappingSnapshot(repoRoot)
	if err != nil {
		t.Fatalf("BuildRepositoryMappingSnapshot: %v", err)
	}
	body := strings.Replace(renderFormalStableVerifyProcessBody(expected, mapping, "aligned"), "version_ref: repository_mapping@0.1.0", "version_ref: repository_mapping@0.0.9", 1)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_stable_verify_result/unit/demo.md"), "# stable verify\n\n```yaml\n"+body+"\n```\n")

	result, err := ValidateProcessFile(repoRoot, "demo", "stable_verify")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid result, got valid")
	}
	if result.NextCommand != "unit_stable_verify" {
		t.Fatalf("expected stable verify fallback, got %s", result.NextCommand)
	}
	if !containsMismatch(result.Mismatches, "repository_mapping_snapshot mismatch: actual=docs/specs/repository_mapping.md|repository_mapping@0.0.9|"+mapping.Fingerprint+" expected="+normalizeRepositoryMapping(mapping)) {
		t.Fatalf("expected repository mapping mismatch, got %+v", result.Mismatches)
	}
}

func TestValidateProcessFileRejectsStableVerifyAlignedEvidenceGap(t *testing.T) {
	repoRoot := t.TempDir()
	setupStableSnapshotValidationRepo(t, repoRoot)
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_stable_verify_result/unit"))

	expected, err := RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	mapping, err := BuildRepositoryMappingSnapshot(repoRoot)
	if err != nil {
		t.Fatalf("BuildRepositoryMappingSnapshot: %v", err)
	}
	body := strings.Replace(renderFormalStableVerifyProcessBody(expected, mapping, "aligned"), "status: pass", "status: partial", 1)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_stable_verify_result/unit/demo.md"), "# stable verify\n\n```yaml\n"+body+"\n```\n")

	result, err := ValidateProcessFile(repoRoot, "demo", "stable_verify")
	if err != nil {
		t.Fatalf("ValidateProcessFile: %v", err)
	}
	if result.Valid {
		t.Fatalf("expected invalid result, got valid")
	}
	if !containsMismatch(result.Mismatches, "stable_verify aligned evidence for demo.core must be pass") {
		t.Fatalf("expected aligned evidence pass mismatch, got %+v", result.Mismatches)
	}
}

func TestValidateProcessFileRejectsScenarioObjectType(t *testing.T) {
	repoRoot := t.TempDir()

	_, err := ValidateProcessFileForObject(repoRoot, "scenario", "checkout", "plan")
	if err == nil || !strings.Contains(err.Error(), "object type \"scenario\" is not supported; only unit is supported") {
		t.Fatalf("expected scenario rejection, got %v", err)
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
candidate_intent: change
source_basis: new_design
evidence_appendix_ref: none
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

func setupStableSnapshotValidationRepo(t *testing.T, repoRoot string) {
	t.Helper()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.StableDir)))

	status := "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n| `unit` | `demo` | `yes` | `no` | `stable` | `unit_stable_verify` | note |\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)

	mainSpec := `---
id: demo
layer: stable
version: 1.0.0
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
	mainSpecRef, err := specpaths.MainSpecFileRef("stable", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), mainSpec)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/repository_mapping.md"), `---
id: repository_mapping
version: 0.1.0
---

# Repository Mapping
`)
}

func replaceCandidateSpecText(t *testing.T, repoRoot, old, replacement string) {
	t.Helper()
	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	path := filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef))
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", mainSpecRef, err)
	}
	updated := strings.Replace(string(content), old, replacement, 1)
	if updated == string(content) {
		t.Fatalf("test fixture did not contain %q", old)
	}
	mustWriteFile(t, path, updated)
}

func writeStableGlobalRuleForFreshnessTest(t *testing.T, repoRoot string) {
	t.Helper()
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_g_rule_repository_baseline.md"), `---
rule_id: g_rule_repository_baseline
rule_scope: global
layer: stable
rule_version: 1.1.0
---

# Rule
`)
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
		"acceptance_behavior_fingerprint: " + expected.AcceptanceBehaviorFingerprint,
		"acceptance_item_set:",
		renderAcceptanceItemSetForTest(expected.AcceptanceItemSet),
		"unit_appendix_snapshot:",
		renderAppendixLinesForTest(expected.ModuleAppendixSnapshot),
		"rule_snapshot: none",
		renderIndependentEvaluationReceiptForTest(expected.SpecFileRef),
	}, "\n")
}

func renderFormalPlanProcessBody(expected Snapshot) string {
	return strings.Join([]string{
		"spec_file_ref: " + expected.SpecFileRef,
		"spec_version_ref: " + expected.SpecVersionRef,
		"spec_fingerprint: " + expected.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + expected.AcceptanceBehaviorFingerprint,
		"unit_appendix_snapshot:",
		renderAppendixLinesForTest(expected.ModuleAppendixSnapshot),
		"rule_snapshot: none",
		"acceptance_item_plan_coverage:",
		renderAcceptancePlanCoverageForTest(expected.AcceptanceItemSet),
		renderIndependentEvaluationReceiptForTest(expected.SpecFileRef),
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
		"acceptance_behavior_fingerprint: " + expected.AcceptanceBehaviorFingerprint,
		"acceptance_item_set:",
		renderAcceptanceItemSetForTest(expected.AcceptanceItemSet),
		"unit_appendix_snapshot: none",
		"verification_scope_ref: current candidate",
		"rule_snapshot: none",
		"acceptance_item_evidence_matrix:",
		renderAcceptanceEvidenceMatrixForTest(expected.AcceptanceItemSet),
		renderIndependentEvaluationReceiptForTest(expected.SpecFileRef),
	}, "\n")
}

func renderFormalStableVerifyProcessBody(expected Snapshot, mapping RepositoryMappingEntry, decision string) string {
	route := stableVerifyDecisions[decision]
	return strings.Join([]string{
		"object_type: unit",
		"object_ref: " + expected.Module,
		"gate: unit_stable_verify",
		"decision: " + decision,
		"allow_next: " + route.AllowNext,
		"next_command: " + route.NextCommand,
		"blocking_summary: none",
		"coverage_summary: current stable implementation",
		"truth_layer_ref: stable",
		"truth_file_ref: " + expected.SpecFileRef,
		"truth_version_ref: " + expected.SpecVersionRef,
		"truth_fingerprint: " + expected.SpecFingerprint,
		"acceptance_behavior_fingerprint: " + expected.AcceptanceBehaviorFingerprint,
		"repository_mapping_snapshot:",
		renderRepositoryMappingLinesForTest(mapping),
		"acceptance_item_set:",
		renderAcceptanceItemSetForTest(expected.AcceptanceItemSet),
		"unit_appendix_snapshot:",
		renderAppendixLinesForTest(expected.ModuleAppendixSnapshot),
		"unit_snapshot:",
		renderObjectSnapshotLinesForTest(expected.UnitSnapshot),
		"rule_snapshot:",
		renderSharedLinesForTest(expected.RuleSnapshot),
		"acceptance_item_evidence_matrix:",
		renderAcceptanceEvidenceMatrixForTest(expected.AcceptanceItemSet),
		"implementation_surface_refs: AgentCore/internal/demo",
		"evidence_refs: go test ./...",
		renderIndependentEvaluationReceiptForTest(expected.SpecFileRef),
	}, "\n")
}

func renderIndependentEvaluationReceiptForTest(reviewInputRef string) string {
	return strings.Join([]string{
		"evaluation_mode: independent",
		"reviewer_result: pass",
		"reviewer_context: minimal_context",
		"review_input_refs: " + reviewInputRef,
		"review_findings: none",
		"human_decision_refs: none",
	}, "\n")
}

func renderFreshnessReceiptForTest(currentFingerprint, reviewInputRef string) string {
	return strings.Join([]string{
		"freshness_impact: text_drift",
		"evidence_reuse: accepted",
		"freshness_current_fingerprint: " + currentFingerprint,
		"freshness_review_mode: independent",
		"freshness_reviewer_result: pass",
		"freshness_reviewer_context: minimal_context",
		"freshness_review_input_refs: " + reviewInputRef,
		"freshness_review_findings: none",
	}, "\n")
}

func renderRepositoryMappingLinesForTest(entry RepositoryMappingEntry) string {
	if entry.FileRef == "" && entry.VersionRef == "" && entry.Fingerprint == "" {
		return "  none"
	}
	return strings.Join([]string{
		"  file_ref: " + entry.FileRef,
		"  version_ref: " + entry.VersionRef,
		"  fingerprint: " + entry.Fingerprint,
	}, "\n")
}

func renderObjectSnapshotLinesForTest(entries []ObjectSnapshotEntry) string {
	if len(entries) == 0 {
		return "  none"
	}
	lines := []string{}
	for _, entry := range entries {
		lines = append(lines,
			"  - unit: "+entry.ObjectRef,
			"    layer: "+entry.Layer,
			"    file_ref: "+entry.FileRef,
			"    version_ref: "+entry.VersionRef,
			"    fingerprint: "+entry.Fingerprint,
		)
	}
	return strings.Join(lines, "\n")
}

func renderAppendixLinesForTest(entries []AppendixEntry) string {
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

func renderSharedLinesForTest(entries []RuleEntry) string {
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
	content = testfixtures.NormalizeSpecFlowContent(path, content)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
