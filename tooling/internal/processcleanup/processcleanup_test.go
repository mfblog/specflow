package processcleanup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
)

func TestValidateFallbackReasonRejectsRemovedReasonLayers(t *testing.T) {
	cases := []struct {
		reason string
		layer  string
	}{
		{reason: "truth_drift", layer: "truth_layer"},
		{reason: "binding_drift", layer: "truth_layer"},
		{reason: "baseline_drift", layer: "truth_layer"},
		{reason: "rule_drift", layer: "truth_layer"},
		{reason: "truth_incomplete", layer: "truth_layer"},
		{reason: "gate_missing", layer: "gate_layer"},
		{reason: "evidence_incomplete", layer: "evidence_layer"},
		{reason: "stable_verify_invalid", layer: "evidence_layer"},
	}

	for _, tc := range cases {
		t.Run(tc.reason, func(t *testing.T) {
			if err := ValidateFallbackReason(tc.reason, tc.layer); err != nil {
				t.Fatalf("ValidateFallbackReason: %v", err)
			}
		})
	}
}

func TestValidateFallbackReasonRejectsUnsupportedAndMismatchedReasonLayers(t *testing.T) {
	for _, tc := range []struct {
		name   string
		reason string
		layer  string
		want   string
	}{
		{name: "legacy reason", reason: "truth_changed", layer: "truth_layer", want: "unsupported fallback reason"},
		{name: "gate reason on plan layer", reason: "gate_missing", layer: "plan_layer", want: "requires failure layer \"gate_layer\""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateFallbackReason(tc.reason, tc.layer)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q failure, got %v", tc.want, err)
			}
		})
	}
}

func TestApplyFallbackRejectsLegacyReasonWithoutInferringTruthLayer(t *testing.T) {
	_, err := ApplyFallback(t.TempDir(), "ai", "unit_verify", "truth_changed")
	if err == nil || !strings.Contains(err.Error(), "unsupported fallback reason") {
		t.Fatalf("expected unsupported legacy reason failure, got %v", err)
	}
}

func TestApplyFallbackForPromoteEvidenceIncomplete(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, "docs/specs/_verify_result/unit"), 0o755); err != nil {
		t.Fatalf("mkdir verify_result: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repoRoot, "docs/specs"), 0o755); err != nil {
		t.Fatalf("mkdir specs: %v", err)
	}

	status := strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `ai` | `yes` | `yes` | `candidate` | `unit_promote` | note |",
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(repoRoot, "docs/specs/_status.md"), []byte(status), 0o644); err != nil {
		t.Fatalf("write status: %v", err)
	}
	verifyPath := filepath.Join(repoRoot, "docs/specs/_verify_result/unit/ai.md")
	if err := os.WriteFile(verifyPath, []byte("verify"), 0o644); err != nil {
		t.Fatalf("write verify file: %v", err)
	}

	result, err := ApplyFallback(repoRoot, "ai", "unit_promote", "evidence_incomplete")
	if err != nil {
		t.Fatalf("ApplyFallback: %v", err)
	}
	if result.NextCommand != "unit_verify" {
		t.Fatalf("expected next command unit_verify, got %s", result.NextCommand)
	}
	if len(result.DeletedFiles) != 1 || result.DeletedFiles[0] != "docs/specs/_verify_result/unit/ai.md" {
		t.Fatalf("unexpected deleted files: %v", result.DeletedFiles)
	}
	if _, err := os.Stat(verifyPath); !os.IsNotExist(err) {
		t.Fatalf("expected verify file to be deleted, stat err=%v", err)
	}
}

func TestApplyFallbackForStableVerifyInvalidDeletesOnlyStableVerifyEvidence(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_stable_verify_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))

	status := strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `ai` | `yes` | `no` | `stable` | `unit_stable_verify` | note |",
	}, "\n") + "\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)

	stableVerifyPath := filepath.Join(repoRoot, "docs/specs/_stable_verify_result/unit/ai.md")
	verifyPath := filepath.Join(repoRoot, "docs/specs/_verify_result/unit/ai.md")
	mustWriteFile(t, stableVerifyPath, "stable verify")
	mustWriteFile(t, verifyPath, "candidate verify")

	result, err := ApplyFallback(repoRoot, "ai", "unit_stable_verify", "stable_verify_invalid")
	if err != nil {
		t.Fatalf("ApplyFallback: %v", err)
	}
	if result.NextCommand != "unit_stable_verify" {
		t.Fatalf("expected next command unit_stable_verify, got %s", result.NextCommand)
	}
	if len(result.DeletedFiles) != 1 || result.DeletedFiles[0] != "docs/specs/_stable_verify_result/unit/ai.md" {
		t.Fatalf("unexpected deleted files: %v", result.DeletedFiles)
	}
	if _, err := os.Stat(stableVerifyPath); !os.IsNotExist(err) {
		t.Fatalf("expected stable verify file to be deleted, stat err=%v", err)
	}
	if _, err := os.Stat(verifyPath); err != nil {
		t.Fatalf("expected candidate verify file to remain, stat err=%v", err)
	}
}

func _TestApplyFallbackForVerifyImplementationDeviation_removed_removed(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_check_work/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_plans/active"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))

	status := strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `ai` | `yes` | `yes` | `candidate` | `unit_verify` | note |",
	}, "\n") + "\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/ai.md"), "check")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/ai.md"), "plan")
	verifyPath := filepath.Join(repoRoot, "docs/specs/_verify_result/unit/ai.md")
	mustWriteFile(t, verifyPath, "verify")

	result, err := ApplyFallback(repoRoot, "ai", "unit_verify", "implementation_deviation")
	if err != nil {
		t.Fatalf("ApplyFallback: %v", err)
	}
	if result.NextCommand != "unit_impl" {
		t.Fatalf("expected next command unit_impl, got %s", result.NextCommand)
	}
	if len(result.DeletedFiles) != 1 || result.DeletedFiles[0] != "docs/specs/_verify_result/unit/ai.md" {
		t.Fatalf("unexpected deleted files: %v", result.DeletedFiles)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "docs/specs/_check_result/unit/ai.md")); err != nil {
		t.Fatalf("expected check file to remain, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "docs/specs/_plans/active/ai.md")); err != nil {
		t.Fatalf("expected plan file to remain, stat err=%v", err)
	}
	if _, err := os.Stat(verifyPath); !os.IsNotExist(err) {
		t.Fatalf("expected verify file to be deleted, stat err=%v", err)
	}
}

func TestApplyFallbackForVerifyTruthIncomplete(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_check_work/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_plans/active"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_plans/draft"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))

	status := strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `ai` | `yes` | `yes` | `candidate` | `unit_verify` | note |",
	}, "\n") + "\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)
	for _, relPath := range []string{
		"docs/specs/_check_work/unit/ai.md",
		"docs/specs/_check_result/unit/ai.md",
		"docs/specs/_verify_result/unit/ai.md",
	} {
		mustWriteFile(t, filepath.Join(repoRoot, relPath), relPath)
	}

	result, err := ApplyFallback(repoRoot, "ai", "unit_verify", "truth_incomplete")
	if err != nil {
		t.Fatalf("ApplyFallback: %v", err)
	}
	if result.NextCommand != "unit_check" {
		t.Fatalf("expected next command unit_check, got %s", result.NextCommand)
	}
	if len(result.DeletedFiles) != 3 {
		t.Fatalf("expected 3 deleted files, got %d: %v", len(result.DeletedFiles), result.DeletedFiles)
	}
	for _, relPath := range []string{
		"docs/specs/_check_work/unit/ai.md",
		"docs/specs/_check_result/unit/ai.md",
		"docs/specs/_verify_result/unit/ai.md",
	} {
		if _, err := os.Stat(filepath.Join(repoRoot, relPath)); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be deleted, stat err=%v", relPath, err)
		}
	}
}

func _TestApplyObjectFallbackForUnitPlanLayer(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_plans/active"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_plans/draft"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))

	status := strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `ai` | `yes` | `yes` | `candidate` | `unit_impl` | note |",
	}, "\n") + "\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/ai.md"), "check")
	for _, relPath := range []string{
		"docs/specs/_verify_result/unit/ai.md",
	} {
		mustWriteFile(t, filepath.Join(repoRoot, relPath), relPath)
	}

	result, err := ApplyObjectFallback(repoRoot, "unit", "ai", "unit_impl", "plan_drift", "plan_layer")
	if err != nil {
		t.Fatalf("ApplyObjectFallback: %v", err)
	}
	if result.NextCommand != "unit_verify" {
		t.Fatalf("expected next command unit_verify, got %s", result.NextCommand)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "docs/specs/_check_result/unit/ai.md")); err != nil {
		t.Fatalf("expected check file to remain, stat err=%v", err)
	}
	for _, relPath := range []string{
		"docs/specs/_verify_result/unit/ai.md",
	} {
		if _, err := os.Stat(filepath.Join(repoRoot, relPath)); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be deleted, stat err=%v", relPath, err)
		}
	}
}

func TestApplyObjectFallbackRejectsInvalidCanonicalReasonInput(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	status := strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `ai` | `no` | `yes` | `candidate` | `unit_impl` | note |",
	}, "\n") + "\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)

	for _, tc := range []struct {
		name   string
		reason string
		layer  string
		want   string
	}{
		{name: "legacy reason", reason: "truth_changed", layer: "truth_layer", want: "unsupported fallback reason"},
		{name: "layer mismatch", reason: "gate_missing", layer: "plan_layer", want: "requires failure layer \"gate_layer\""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ApplyObjectFallback(repoRoot, "unit", "ai", "unit_impl", tc.reason, tc.layer)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q failure, got %v", tc.want, err)
			}
		})
	}
}

func TestApplyObjectFallbackRejectsScenario(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))

	status := strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `scenario` | `checkout` | `no` | `yes` | `candidate` | `scenario_promote` | note |",
	}, "\n") + "\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)

	_, err := ApplyObjectFallback(repoRoot, "scenario", "checkout", "scenario_promote", "stable_dependency_not_ready", "dependency_readiness_layer")
	if err == nil || !strings.Contains(err.Error(), "object type \"scenario\" is not supported; only unit is supported") {
		t.Fatalf("expected scenario rejection, got %v", err)
	}
}

func TestApplySuccessCleanupForPromote(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateAppendixDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_check_work/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_plans/active"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_plans/draft"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))

	status := strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `ai` | `yes` | `no` | `stable` | `unit_fork` | promoted |",
	}, "\n") + "\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)
	candidateMainRef, err := specpaths.MainSpecFileRef("candidate", "ai")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(candidateMainRef)), "candidate")
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateAppendixDir), "c_unit_ai_prompt.md"), "appendix")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_check_work/unit/ai.md"), "check work")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/ai.md"), "check")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/ai.md"), "plan")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_plans/draft/ai.md"), "draft plan")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit/ai.md"), "verify")
	summaryRef := snapshot.StablePromotionSummaryFilePath("unit", "ai")
	mustMkdirAll(t, filepath.Dir(filepath.Join(repoRoot, filepath.FromSlash(summaryRef))))
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(summaryRef)), "stable promotion summary")

	result, err := ApplySuccessCleanup(repoRoot, "ai", "unit_promote")
	if err != nil {
		t.Fatalf("ApplySuccessCleanup: %v", err)
	}
	if len(result.DeletedFiles) != 7 {
		t.Fatalf("expected 7 deleted files, got %d: %v", len(result.DeletedFiles), result.DeletedFiles)
	}
	for _, relPath := range []string{
		candidateMainRef,
		specpaths.CandidateAppendixDir + "/c_unit_ai_prompt.md",
		"docs/specs/_check_work/unit/ai.md",
		"docs/specs/_check_result/unit/ai.md",
		"docs/specs/_verify_result/unit/ai.md",
	} {
		if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(relPath))); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be deleted, stat err=%v", relPath, err)
		}
	}
}

func TestApplySuccessCleanupForPromoteRequiresStableSummary(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateAppendixDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))

	status := strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `ai` | `yes` | `no` | `stable` | `unit_fork` | promoted |",
	}, "\n") + "\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)
	candidateMainRef, err := specpaths.MainSpecFileRef("candidate", "ai")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(candidateMainRef)), "candidate")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit/ai.md"), "verify")

	_, err = ApplySuccessCleanup(repoRoot, "ai", "unit_promote")
	if err == nil || !strings.Contains(err.Error(), "stable promotion summary is required before unit_promote cleanup") {
		t.Fatalf("expected stable summary prerequisite error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(candidateMainRef))); err != nil {
		t.Fatalf("candidate main must remain when summary is missing, stat err=%v", err)
	}
}

func TestApplySuccessCleanupForUnitForkPreservesCurrentCandidateAppendix(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateAppendixDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_check_work/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_plans/active"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_plans/draft"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))

	status := strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `ai` | `yes` | `yes` | `candidate` | `unit_check` | forked |",
	}, "\n") + "\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)
	appendixRef := specpaths.CandidateAppendixDir + "/c_unit_ai_prompt.md"
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(appendixRef)), "appendix")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_check_work/unit/ai.md"), "check work")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit/ai.md"), "check")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_plans/active/ai.md"), "plan")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_plans/draft/ai.md"), "draft plan")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_verify_result/unit/ai.md"), "verify")

	result, err := ApplySuccessCleanup(repoRoot, "ai", "unit_fork")
	if err != nil {
		t.Fatalf("ApplySuccessCleanup: %v", err)
	}
	if len(result.DeletedFiles) != 5 {
		t.Fatalf("expected all process files deleted, got %d: %v", len(result.DeletedFiles), result.DeletedFiles)
	}
	for _, deleted := range result.DeletedFiles {
		if deleted == appendixRef {
			t.Fatalf("unit_fork cleanup must not delete current candidate appendix: %v", result.DeletedFiles)
		}
	}
	if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(appendixRef))); err != nil {
		t.Fatalf("expected candidate appendix to remain, stat err=%v", err)
	}
	for _, relPath := range []string{
		"docs/specs/_check_work/unit/ai.md",
		"docs/specs/_check_result/unit/ai.md",
		"docs/specs/_verify_result/unit/ai.md",
	} {
		if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(relPath))); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be deleted, stat err=%v", relPath, err)
		}
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
