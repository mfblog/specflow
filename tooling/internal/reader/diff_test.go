package reader

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestReadAllowedSourceDiffForUnitCandidate(t *testing.T) {
	repoRoot := createReaderRepo(t)
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_memory.md"), strings.Join([]string{
		"---",
		"id: memory",
		"layer: stable",
		"version: 0.1.0",
		"---",
		"",
		"# Memory",
		"",
		"Stable paragraph.",
		"",
		"Removed paragraph.",
	}, "\n")+"\n")
	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_memory.md"), strings.Join([]string{
		"---",
		"id: memory",
		"layer: candidate",
		"version: 0.1.1",
		"candidate_intent: repair",
		"repair_basis: s_unit_memory@0.1.0",
		"source_basis: new_design",
		"evidence_appendix_ref: none",
		"---",
		"",
		"# Memory",
		"",
		"Candidate paragraph.",
	}, "\n")+"\n")

	diff, err := ReadAllowedSourceDiff(repoRoot, "docs/specs/units/candidate/c_unit_memory.md")
	if err != nil {
		t.Fatalf("ReadAllowedSourceDiff returned error: %v", err)
	}
	if !diff.Available {
		t.Fatalf("expected diff to be available, got %+v", diff)
	}
	if diff.StablePath != "docs/specs/units/stable/s_unit_memory.md" {
		t.Fatalf("unexpected stable path: %q", diff.StablePath)
	}
	if diff.Summary.Added == 0 || diff.Summary.Deleted == 0 || len(diff.Hunks) == 0 {
		t.Fatalf("expected added and deleted lines, got %+v", diff)
	}
	if !diffContainsLine(diff, "insert", "Candidate paragraph.") {
		t.Fatalf("expected inserted candidate paragraph, got %+v", diff.Hunks)
	}
	if !diffContainsLine(diff, "delete", "Removed paragraph.") {
		t.Fatalf("expected deleted stable paragraph, got %+v", diff.Hunks)
	}
}

func TestReadAllowedSourceDiffForUnitCandidateAppendix(t *testing.T) {
	repoRoot := createReaderRepo(t)

	diff, err := ReadAllowedSourceDiff(repoRoot, "docs/specs/units/candidate/appendix/c_unit_assistant_prompt.md")
	if err != nil {
		t.Fatalf("ReadAllowedSourceDiff appendix returned error: %v", err)
	}
	if !diff.Available {
		t.Fatalf("expected appendix diff to be available, got %+v", diff)
	}
	if diff.StablePath != "docs/specs/units/stable/appendix/s_unit_assistant_prompt.md" {
		t.Fatalf("unexpected stable appendix path: %q", diff.StablePath)
	}
}

func TestReadAllowedSourceDiffUnavailableCases(t *testing.T) {
	repoRoot := createReaderRepo(t)

	diff, err := ReadAllowedSourceDiff(repoRoot, "docs/specs/units/candidate/c_unit_memory.md")
	if err != nil {
		t.Fatalf("ReadAllowedSourceDiff missing stable returned error: %v", err)
	}
	if diff.Available || diff.Reason != "stable_missing" {
		t.Fatalf("expected stable_missing, got %+v", diff)
	}

	diff, err = ReadAllowedSourceDiff(repoRoot, "docs/specs/units/candidate/appendix/c_unit_assistant_unlinked.md")
	if err != nil {
		t.Fatalf("ReadAllowedSourceDiff appendix returned error: %v", err)
	}
	if diff.Available || diff.Reason != "stable_missing" {
		t.Fatalf("expected missing stable appendix unavailable result, got %+v", diff)
	}

	diff, err = ReadAllowedSourceDiff(repoRoot, "docs/specs/units/stable/s_unit_tool.md")
	if err != nil {
		t.Fatalf("ReadAllowedSourceDiff stable returned error: %v", err)
	}
	if diff.Available || diff.Reason != "not_unit_candidate_truth" {
		t.Fatalf("expected stable doc to be unavailable, got %+v", diff)
	}

	_, err = ReadAllowedSourceDiff(repoRoot, "../AGENTS.md")
	if err == nil || !strings.Contains(err.Error(), "escapes repo root") {
		t.Fatalf("expected path escape error, got %v", err)
	}
}

func diffContainsLine(diff SourceDiff, lineType, text string) bool {
	for _, hunk := range diff.Hunks {
		for _, line := range hunk.Lines {
			if line.Type == lineType && line.Text == text {
				return true
			}
		}
	}
	return false
}
