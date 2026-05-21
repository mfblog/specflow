package relationgraph

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCandidateReferenceBlocksDependentCandidate(t *testing.T) {
	repoRoot := relationRepo(t)
	writeRelationStatus(t, repoRoot, []string{
		"| `unit` | `alpha` | `yes` | `yes` | `candidate` | `unit_check` | alpha |",
		"| `unit` | `beta` | `yes` | `yes` | `candidate` | `unit_check` | beta |",
	})
	writeCandidateUnit(t, repoRoot, "alpha", "0.2.0", "", "[Beta](./c_unit_beta.md)")
	writeCandidateUnit(t, repoRoot, "beta", "0.1.0", "", "No candidate dependencies.")

	result := Build(repoRoot)

	if !reflect.DeepEqual(result.ReadyCandidates, []string{"beta"}) {
		t.Fatalf("ready candidates = %#v", result.ReadyCandidates)
	}
	blocked := findBlockedCandidate(t, result.BlockedCandidates, "alpha")
	if !reflect.DeepEqual(blocked.BlockedBy, []string{"unit:beta"}) {
		t.Fatalf("blocked_by = %#v", blocked.BlockedBy)
	}
}

func TestCandidateOrderUsesStableTopologicalOrder(t *testing.T) {
	repoRoot := relationRepo(t)
	writeRelationStatus(t, repoRoot, []string{
		"| `unit` | `alpha` | `yes` | `yes` | `candidate` | `unit_check` | alpha |",
		"| `unit` | `beta` | `yes` | `yes` | `candidate` | `unit_check` | beta |",
		"| `unit` | `gamma` | `yes` | `yes` | `candidate` | `unit_check` | gamma |",
	})
	writeCandidateUnit(t, repoRoot, "alpha", "0.3.0", "", "`c_unit_beta@0.2.0`")
	writeCandidateUnit(t, repoRoot, "beta", "0.2.0", "", "`c_unit_gamma@0.1.0`")
	writeCandidateUnit(t, repoRoot, "gamma", "0.1.0", "", "No candidate dependencies.")

	result := Build(repoRoot)

	if !reflect.DeepEqual(result.CandidateOrder, []string{"gamma", "beta", "alpha"}) {
		t.Fatalf("candidate order = %#v", result.CandidateOrder)
	}
	if !reflect.DeepEqual(result.ReadyCandidates, []string{"gamma"}) {
		t.Fatalf("ready candidates = %#v", result.ReadyCandidates)
	}
}

func TestCandidateCycleBlocksPreflight(t *testing.T) {
	repoRoot := relationRepo(t)
	writeRelationStatus(t, repoRoot, []string{
		"| `unit` | `alpha` | `yes` | `yes` | `candidate` | `unit_check` | alpha |",
		"| `unit` | `beta` | `yes` | `yes` | `candidate` | `unit_check` | beta |",
	})
	writeCandidateUnit(t, repoRoot, "alpha", "0.2.0", "", "`c_unit_beta@0.2.0`")
	writeCandidateUnit(t, repoRoot, "beta", "0.2.0", "", "`c_unit_alpha@0.2.0`")

	result := Build(repoRoot)
	preflight := CandidatePreflight(repoRoot, "alpha")

	if result.RelationResult != "fail" {
		t.Fatalf("relation result = %q", result.RelationResult)
	}
	if len(result.CandidateCycles) != 1 {
		t.Fatalf("candidate cycles = %#v", result.CandidateCycles)
	}
	if preflight.MayContinue {
		t.Fatalf("preflight should block: %#v", preflight)
	}
}

func TestStableUnitRefsCycleIsDiagnosticOnly(t *testing.T) {
	repoRoot := relationRepo(t)
	writeRelationStatus(t, repoRoot, []string{
		"| `unit` | `alpha` | `yes` | `yes` | `candidate` | `unit_check` | alpha |",
		"| `unit` | `beta` | `yes` | `no` | `stable` | `unit_fork` | beta |",
	})
	writeCandidateUnit(t, repoRoot, "alpha", "0.2.0", "  - s_unit_beta@0.1.0", "No candidate dependencies.")
	writeStableUnit(t, repoRoot, "beta", "0.1.0", "  - s_unit_alpha@0.1.0")

	result := Build(repoRoot)

	if !reflect.DeepEqual(result.ReadyCandidates, []string{"alpha"}) {
		t.Fatalf("ready candidates = %#v", result.ReadyCandidates)
	}
	if len(result.CandidateCycles) != 0 {
		t.Fatalf("stable cycle must not become candidate cycle: %#v", result.CandidateCycles)
	}
	if !hasDiagnosticPrefix(result.Diagnostics, "stable_reference_cycle:") {
		t.Fatalf("expected stable cycle diagnostic, got %#v", result.Diagnostics)
	}
}

func TestEvidenceAppendixCandidateReferenceDoesNotBlock(t *testing.T) {
	repoRoot := relationRepo(t)
	writeRelationStatus(t, repoRoot, []string{
		"| `unit` | `alpha` | `yes` | `yes` | `candidate` | `unit_check` | alpha |",
		"| `unit` | `beta` | `yes` | `yes` | `candidate` | `unit_check` | beta |",
	})
	writeCandidateUnit(t, repoRoot, "alpha", "0.2.0", "", "See evidence.", "docs/specs/units/candidate/appendix/c_unit_alpha_evidence.md")
	writeCandidateUnit(t, repoRoot, "beta", "0.1.0", "", "No candidate dependencies.")
	writeRelationFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/appendix/c_unit_alpha_evidence.md"), "Evidence references `c_unit_beta@0.1.0`.\n")

	result := Build(repoRoot)

	if !reflect.DeepEqual(result.ReadyCandidates, []string{"alpha", "beta"}) {
		t.Fatalf("ready candidates = %#v", result.ReadyCandidates)
	}
	if len(result.BlockedCandidates) != 0 {
		t.Fatalf("evidence references must not block: %#v", result.BlockedCandidates)
	}
	if len(result.ReferenceEdges) != 1 || result.ReferenceEdges[0].Blocking {
		t.Fatalf("expected non-blocking evidence edge, got %#v", result.ReferenceEdges)
	}
}

func findBlockedCandidate(t *testing.T, items []BlockedCandidate, object string) BlockedCandidate {
	t.Helper()
	for _, item := range items {
		if item.Object == object {
			return item
		}
	}
	t.Fatalf("blocked candidate %q not found in %#v", object, items)
	return BlockedCandidate{}
}

func relationRepo(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func writeRelationStatus(t *testing.T, repoRoot string, rows []string) {
	t.Helper()
	content := strings.Join(append([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
	}, rows...), "\n") + "\n"
	writeRelationFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), content)
}

func writeCandidateUnit(t *testing.T, repoRoot, object, version, unitRefs, body string, evidenceRefs ...string) {
	t.Helper()
	evidenceRef := "none"
	if len(evidenceRefs) > 0 {
		evidenceRef = evidenceRefs[0]
	}
	unitRefsBlock := "unit_refs: none"
	if strings.TrimSpace(unitRefs) != "" {
		unitRefsBlock = "unit_refs:\n" + unitRefs
	}
	content := strings.Join([]string{
		"---",
		"id: " + object,
		"layer: candidate",
		"version: " + version,
		"evidence_appendix_ref: " + evidenceRef,
		unitRefsBlock,
		"rule_refs: none",
		"---",
		"",
		"# " + object,
		"",
		body,
	}, "\n") + "\n"
	writeRelationFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_"+object+".md"), content)
}

func writeStableUnit(t *testing.T, repoRoot, object, version, unitRefs string) {
	t.Helper()
	unitRefsBlock := "unit_refs: none"
	if strings.TrimSpace(unitRefs) != "" {
		unitRefsBlock = "unit_refs:\n" + unitRefs
	}
	content := strings.Join([]string{
		"---",
		"id: " + object,
		"layer: stable",
		"version: " + version,
		unitRefsBlock,
		"rule_refs: none",
		"---",
		"",
		"# " + object,
	}, "\n") + "\n"
	writeRelationFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_"+object+".md"), content)
}

func writeRelationFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hasDiagnosticPrefix(diagnostics []string, prefix string) bool {
	for _, diagnostic := range diagnostics {
		if strings.HasPrefix(diagnostic, prefix) {
			return true
		}
	}
	return false
}
