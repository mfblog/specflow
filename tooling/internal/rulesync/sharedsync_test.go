package rulesync

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/testfixtures"
)

func TestSyncImpactKeepsCandidateWhenOnlyBoundModulesChanged(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupCandidateSharedRepo(t, repoRoot)

	writeSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.1.0
bound_objects:
  - unit:demo
  - unit:module_other
---

# Shared

Body stays the same.
`)

	result, err := SyncImpact(repoRoot, Options{
		RuleRefs:                     []string{sharedRef},
		BoundObjectsOnlyRuleFileRefs: []string{"docs/specs/rules/candidate/c_b_rule_demo.md"},
	})
	if err == nil || !strings.Contains(err.Error(), "bound_objects-only sync is no longer supported") {
		t.Fatalf("expected bound_objects-only rejection, got result=%+v err=%v", result, err)
	}
	checkPath := filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md")
	if _, err := os.Stat(checkPath); err != nil {
		t.Fatalf("expected process file to remain, stat err=%v", err)
	}
}

func TestSyncImpactKeepsExplicitModuleScopeWhenOnlyBoundModulesChanged(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupCandidateSharedRepo(t, repoRoot)

	writeSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.1.0
bound_objects:
  - unit:demo
  - unit:module_other
---

# Shared

Body stays the same.
`)

	result, err := SyncImpact(repoRoot, Options{
		Modules:                      []string{"demo"},
		RuleRefs:                     []string{sharedRef},
		BoundObjectsOnlyRuleFileRefs: []string{"docs/specs/rules/candidate/c_b_rule_demo.md"},
	})
	if err == nil || !strings.Contains(err.Error(), "bound_objects-only sync is no longer supported") {
		t.Fatalf("expected bound_objects-only rejection, got result=%+v err=%v", result, err)
	}
}

func TestSyncImpactInvalidatesCandidateWhenBoundModulesChangedWithoutExplicitDeclaration(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupCandidateSharedRepo(t, repoRoot)

	writeSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.1.0
bound_objects:
  - unit:demo
  - unit:module_other
---

# Shared

Body stays the same.
`)

	result, err := SyncImpact(repoRoot, Options{RuleRefs: []string{sharedRef}})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.ModuleResults) != 1 {
		t.Fatalf("expected one module result, got %d", len(result.ModuleResults))
	}
	moduleResult := result.ModuleResults[0]
	if moduleResult.Outcome != "unchanged" {
		t.Fatalf("expected unchanged outcome after bound_objects removal, got %+v", moduleResult)
	}
	if len(result.BoundObjectDrifts) != 0 {
		t.Fatalf("expected no bound_objects drift after metadata removal, got %+v", result.BoundObjectDrifts)
	}
}

func TestSyncImpactRejectsMissingExplicitScope(t *testing.T) {
	repoRoot := t.TempDir()
	setupCandidateSharedRepo(t, repoRoot)

	_, err := SyncImpact(repoRoot, Options{})
	if err == nil || !strings.Contains(err.Error(), "at least one of rule refs, rule ids, or deleted rule refs is required") {
		t.Fatalf("expected missing-scope error, got %v", err)
	}
}

func TestSyncImpactAcceptsTerminalDeletedRuleRefWithNoConsumers(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupCandidateSharedRepo(t, repoRoot)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs: none",
		"",
	}, "\n"))
	if err := os.Remove(filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md")); err != nil {
		t.Fatalf("remove rule file: %v", err)
	}

	result, err := SyncImpact(repoRoot, Options{DeletedRuleRefs: []string{sharedRef}})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.DeletedRuleRefs) != 1 || result.DeletedRuleRefs[0] != sharedRef {
		t.Fatalf("expected deleted rule ref output, got %+v", result.DeletedRuleRefs)
	}
	if len(result.ScopedModules) != 0 {
		t.Fatalf("expected no scoped modules, got %+v", result.ScopedModules)
	}
	if len(result.ModuleResults) != 0 {
		t.Fatalf("expected no module results, got %+v", result.ModuleResults)
	}
}

func TestSyncImpactRejectsTerminalDeletedRuleRefWithConsumer(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupCandidateSharedRepo(t, repoRoot)

	if err := os.Remove(filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md")); err != nil {
		t.Fatalf("remove rule file: %v", err)
	}

	_, err := SyncImpact(repoRoot, Options{DeletedRuleRefs: []string{sharedRef}})
	if err == nil || !strings.Contains(err.Error(), "is still referenced by current-layer unit rule_refs") {
		t.Fatalf("expected deleted rule ref consumer error, got %v", err)
	}
}

func TestSyncImpactInvalidatesCandidateOnSharedTruthDrift(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupCandidateSharedRepo(t, repoRoot)

	writeSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.1.0
bound_objects:
  - unit:demo
---

# Shared

Body changed.
`)

	result, err := SyncImpact(repoRoot, Options{RuleRefs: []string{sharedRef}})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.ModuleResults) != 1 {
		t.Fatalf("expected one module result, got %d", len(result.ModuleResults))
	}
	moduleResult := result.ModuleResults[0]
	if moduleResult.Outcome != "invalidated" {
		t.Fatalf("expected invalidated outcome, got %+v", moduleResult)
	}
	if moduleResult.FallbackReasonCode != "rule_drift" {
		t.Fatalf("expected rule_drift, got %s", moduleResult.FallbackReasonCode)
	}
	if moduleResult.NextCommand != "unit_check" {
		t.Fatalf("expected next command unit_check, got %s", moduleResult.NextCommand)
	}
	if !moduleResult.StatusUpdated {
		t.Fatalf("expected status update")
	}
	checkPath := filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md")
	if _, err := os.Stat(checkPath); !os.IsNotExist(err) {
		t.Fatalf("expected process file to be deleted, stat err=%v", err)
	}
	statusData, err := os.ReadFile(filepath.Join(repoRoot, "docs/specs/_status.md"))
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	if !strings.Contains(string(statusData), "| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_check` | current round |") {
		t.Fatalf("status row not updated:\n%s", string(statusData))
	}
}

func TestSyncImpactIncludesModulesStillBoundToDeletedSharedRef(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupCandidateSharedRepo(t, repoRoot)

	if err := os.Remove(filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md")); err != nil {
		t.Fatalf("remove rule file: %v", err)
	}

	result, err := SyncImpact(repoRoot, Options{RuleRefs: []string{sharedRef}})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.ModuleResults) != 1 {
		t.Fatalf("expected one module result, got %d", len(result.ModuleResults))
	}
	moduleResult := result.ModuleResults[0]
	if moduleResult.Module != "demo" {
		t.Fatalf("expected demo, got %+v", moduleResult)
	}
	if moduleResult.Outcome != "invalidated" {
		t.Fatalf("expected invalidated outcome, got %+v", moduleResult)
	}
	if moduleResult.FallbackReasonCode != "binding_drift" {
		t.Fatalf("expected binding_drift, got %s", moduleResult.FallbackReasonCode)
	}
}

func TestSyncImpactFailsClosedForSharedIDWhenBindingsPointToDeletedSharedRef(t *testing.T) {
	repoRoot := t.TempDir()
	setupCandidateSharedRepo(t, repoRoot)

	if err := os.Remove(filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md")); err != nil {
		t.Fatalf("remove rule file: %v", err)
	}

	_, err := SyncImpact(repoRoot, Options{RuleIDs: []string{"shared_demo"}})
	if err == nil {
		t.Fatalf("expected shared-id sync to fail closed when rule ref is unresolved")
	}
	if !strings.Contains(err.Error(), "cannot determine affected downstream objects safely") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSyncImpactReroutesStableModuleToStableVerify(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupStableSharedRepo(t, repoRoot)

	writeStableSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: stable
rule_version: 1.0.0
bound_objects:
  - unit:demo
---

# Shared

Body changed.
`)

	result, err := SyncImpact(repoRoot, Options{RuleRefs: []string{sharedRef}})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.ModuleResults) != 1 {
		t.Fatalf("expected one module result, got %d", len(result.ModuleResults))
	}
	moduleResult := result.ModuleResults[0]
	if moduleResult.Outcome != "rerouted" {
		t.Fatalf("expected rerouted outcome, got %+v", moduleResult)
	}
	if moduleResult.FallbackReasonCode != "rule_drift" {
		t.Fatalf("expected rule_drift, got %s", moduleResult.FallbackReasonCode)
	}
	if moduleResult.NextCommand != "unit_stable_verify" {
		t.Fatalf("expected next command unit_stable_verify, got %s", moduleResult.NextCommand)
	}
	statusData, err := os.ReadFile(filepath.Join(repoRoot, "docs/specs/_status.md"))
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	if !strings.Contains(string(statusData), "| `unit` | `demo` | `yes` | `no` | `stable` | `unit_stable_verify` | stable round |") {
		t.Fatalf("status row not updated:\n%s", string(statusData))
	}
}

func TestSyncImpactStableGlobalRuleAffectsEveryCurrentUnit(t *testing.T) {
	for _, tc := range []struct {
		name    string
		options Options
	}{
		{
			name:    "exact ref",
			options: Options{RuleRefs: []string{"s_g_rule_repository_baseline@1.1.0"}},
		},
		{
			name:    "rule id",
			options: Options{RuleIDs: []string{"g_rule_repository_baseline"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			setupStableGlobalRuleRepo(t, repoRoot)

			result, err := SyncImpact(repoRoot, tc.options)
			if err != nil {
				t.Fatalf("SyncImpact: %v", err)
			}
			if strings.Join(result.ScopedModules, ",") != "agent,demo" {
				t.Fatalf("expected every current unit in scope, got %+v", result.ScopedModules)
			}

			resultsByUnit := map[string]ModuleResult{}
			for _, item := range result.ModuleResults {
				resultsByUnit[item.Module] = item
			}
			agent := resultsByUnit["agent"]
			if agent.Outcome != "rerouted" || agent.FallbackReasonCode != "rule_drift" || agent.NextCommand != "unit_stable_verify" {
				t.Fatalf("expected stable agent to reroute to unit_stable_verify, got %+v", agent)
			}
			demo := resultsByUnit["demo"]
			if demo.Outcome != "invalidated" || demo.FallbackReasonCode != "rule_drift" || demo.NextCommand != "unit_check" {
				t.Fatalf("expected candidate demo to fall back to unit_check, got %+v", demo)
			}
		})
	}
}

func TestSyncImpactRejectsStableModuleBindingCandidateShared(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.StableDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/candidate"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))

	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | stable round |",
	}, "\n")+"\n")

	mainSpecRef, err := specpaths.MainSpecFileRef("stable", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: stable",
		"version: 1.0.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs:",
		"   - c_b_rule_demo@0.1.0",
		"",
	}, "\n"))

	writeSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.1.0
bound_objects:
  - unit:demo
---

# Shared

Body stays the same.
`)

	result, err := SyncImpact(repoRoot, Options{RuleRefs: []string{"c_b_rule_demo@0.1.0"}})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.ModuleResults) != 1 {
		t.Fatalf("expected one module result, got %d", len(result.ModuleResults))
	}
	moduleResult := result.ModuleResults[0]
	if moduleResult.Outcome != "rerouted" {
		t.Fatalf("expected rerouted outcome, got %+v", moduleResult)
	}
	if moduleResult.FallbackReasonCode != "binding_drift" {
		t.Fatalf("expected binding_drift, got %s", moduleResult.FallbackReasonCode)
	}
	if moduleResult.NextCommand != "unit_stable_verify" {
		t.Fatalf("expected next command unit_stable_verify, got %s", moduleResult.NextCommand)
	}
	if len(moduleResult.Diagnostics) == 0 || !strings.Contains(moduleResult.Diagnostics[0], "stable-layer object binding must use an s_ rule ref") {
		t.Fatalf("expected stable binding diagnostic, got %+v", moduleResult.Diagnostics)
	}
}

func TestSyncImpactKeepsStableLandingModule(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupStableSharedRepo(t, repoRoot)

	writeStableSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: stable
rule_version: 1.0.0
bound_objects:
  - unit:demo
---

# Shared

Body changed.
`)

	result, err := SyncImpact(repoRoot, Options{
		Modules:               []string{"demo"},
		RuleRefs:              []string{sharedRef},
		StableLandingModule:   "demo",
		StableLandingRuleRefs: []string{sharedRef},
	})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.ModuleResults) != 1 {
		t.Fatalf("expected one module result, got %d", len(result.ModuleResults))
	}
	moduleResult := result.ModuleResults[0]
	if moduleResult.Outcome != "unchanged" {
		t.Fatalf("expected unchanged outcome for stable landing unit, got %+v", moduleResult)
	}
	if moduleResult.NextCommand != "unit_fork" {
		t.Fatalf("expected next command unit_fork, got %s", moduleResult.NextCommand)
	}
}

func TestSyncImpactInvalidatesRetargetedCandidateUnit(t *testing.T) {
	repoRoot := t.TempDir()
	oldRef, stableRef := setupStableLandingRetargetRepo(t, repoRoot, true, false)

	result, err := SyncImpact(repoRoot, Options{
		RuleRefs:              []string{oldRef, stableRef},
		StableLandingModule:   "demo",
		StableLandingRuleRefs: []string{stableRef},
		RetargetedUnits:       []string{"agent"},
	})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.RetargetedUnits) != 1 || result.RetargetedUnits[0] != "agent" {
		t.Fatalf("expected retargeted unit output, got %+v", result.RetargetedUnits)
	}
	if len(result.ModuleResults) != 2 {
		t.Fatalf("expected owner and retargeted unit results, got %+v", result.ModuleResults)
	}
	resultsByUnit := map[string]ModuleResult{}
	for _, item := range result.ModuleResults {
		resultsByUnit[item.Module] = item
	}
	if owner := resultsByUnit["demo"]; owner.Outcome != "unchanged" || owner.NextCommand != "unit_fork" {
		t.Fatalf("expected stable landing owner to stay unchanged, got %+v", owner)
	}
	agent := resultsByUnit["agent"]
	if agent.Outcome != "invalidated" || agent.NextCommand != "unit_check" {
		t.Fatalf("expected retargeted agent fallback to unit_check, got %+v", agent)
	}
	if agent.FallbackReasonCode == "" {
		t.Fatalf("expected fallback reason for retargeted agent, got %+v", agent)
	}
	checkPath := filepath.Join(repoRoot, "docs/specs/_check_result/unit/agent.md")
	if _, err := os.Stat(checkPath); !os.IsNotExist(err) {
		t.Fatalf("expected retargeted agent check file to be deleted, stat err=%v", err)
	}
	statusData, err := os.ReadFile(filepath.Join(repoRoot, "docs/specs/_status.md"))
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	if !strings.Contains(string(statusData), "| `unit` | `agent` | `no` | `yes` | `candidate` | `unit_check` | current round |") {
		t.Fatalf("expected agent status fallback to unit_check:\n%s", string(statusData))
	}
}

func TestSyncImpactRejectsRetargetedStableUnit(t *testing.T) {
	repoRoot := t.TempDir()
	oldRef, stableRef := setupStableLandingRetargetRepo(t, repoRoot, true, true)

	_, err := SyncImpact(repoRoot, Options{
		RuleRefs:              []string{oldRef, stableRef},
		StableLandingModule:   "demo",
		StableLandingRuleRefs: []string{stableRef},
		RetargetedUnits:       []string{"agent"},
	})
	if err == nil || !strings.Contains(err.Error(), "retargeted unit \"agent\" must currently be at active layer candidate") {
		t.Fatalf("expected retargeted stable unit error, got %v", err)
	}
}

func TestSyncImpactRejectsRetargetedUnitWithoutStableLandingBinding(t *testing.T) {
	repoRoot := t.TempDir()
	oldRef, stableRef := setupStableLandingRetargetRepo(t, repoRoot, false, false)

	_, err := SyncImpact(repoRoot, Options{
		RuleRefs:              []string{oldRef, stableRef},
		StableLandingModule:   "demo",
		StableLandingRuleRefs: []string{stableRef},
		RetargetedUnits:       []string{"agent"},
	})
	if err == nil || !strings.Contains(err.Error(), "must currently bind at least one stable landing rule ref") {
		t.Fatalf("expected retargeted unit binding error, got %v", err)
	}
}

func TestSyncImpactRejectsRetargetedUnitWithoutCandidateLandingSourceRef(t *testing.T) {
	repoRoot := t.TempDir()
	_, stableRef := setupStableLandingRetargetRepo(t, repoRoot, true, false)

	_, err := SyncImpact(repoRoot, Options{
		RuleRefs:              []string{stableRef},
		StableLandingModule:   "demo",
		StableLandingRuleRefs: []string{stableRef},
		RetargetedUnits:       []string{"agent"},
	})
	if err == nil || !strings.Contains(err.Error(), "requires a candidate-layer rule ref with the same rule_id and rule_version in --rule-refs") {
		t.Fatalf("expected retargeted unit candidate source ref error, got %v", err)
	}
}

func TestSyncImpactRejectsRetargetedUnitWithMismatchedCandidateLandingSourceRef(t *testing.T) {
	repoRoot := t.TempDir()
	_, stableRef := setupStableLandingRetargetRepo(t, repoRoot, true, false)
	writeSharedFileAtPath(t, repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md", `---
rule_id: different_demo
rule_scope: bound
layer: candidate
rule_version: 1.0.0
promotion_owner_unit: demo
bound_objects: none
---

# Shared

Body is not the same formal rule object.
`)

	_, err := SyncImpact(repoRoot, Options{
		RuleRefs:              []string{"c_b_rule_demo@1.0.0", stableRef},
		StableLandingModule:   "demo",
		StableLandingRuleRefs: []string{stableRef},
		RetargetedUnits:       []string{"agent"},
	})
	if err == nil || !strings.Contains(err.Error(), "requires a candidate-layer rule ref with the same rule_id and rule_version in --rule-refs") {
		t.Fatalf("expected retargeted unit candidate source mismatch error, got %v", err)
	}
}

func TestSyncImpactRejectsUnknownRetargetedUnit(t *testing.T) {
	repoRoot := t.TempDir()
	oldRef, stableRef := setupStableLandingRetargetRepo(t, repoRoot, true, false)

	_, err := SyncImpact(repoRoot, Options{
		RuleRefs:              []string{oldRef, stableRef},
		StableLandingModule:   "demo",
		StableLandingRuleRefs: []string{stableRef},
		RetargetedUnits:       []string{"missing"},
	})
	if err == nil || !strings.Contains(err.Error(), "retargeted unit \"missing\" is not registered") {
		t.Fatalf("expected unknown retargeted unit error, got %v", err)
	}
}

func TestSyncImpactStableLandingModuleStillReroutesOnUnrelatedSharedDrift(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupStableSharedRepo(t, repoRoot)

	mainSpecRef, err := specpaths.MainSpecFileRef("stable", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: stable",
		"version: 1.0.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs:",
		"   - s_b_rule_demo@1.0.0",
		"   - s_b_rule_extra@1.1.0",
		"",
	}, "\n"))

	writeSharedFileAtPath(t, repoRoot, "docs/specs/rules/stable/s_b_rule_extra.md", `---
rule_id: shared_extra
rule_scope: bound
layer: stable
rule_version: 1.1.0
bound_objects:
  - unit:demo
---

# Shared

Body changed.
`)

	result, err := SyncImpact(repoRoot, Options{
		Modules:               []string{"demo"},
		RuleRefs:              []string{sharedRef, "s_b_rule_extra@1.1.0"},
		StableLandingModule:   "demo",
		StableLandingRuleRefs: []string{sharedRef},
	})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.ModuleResults) != 1 {
		t.Fatalf("expected one module result, got %d", len(result.ModuleResults))
	}
	moduleResult := result.ModuleResults[0]
	if moduleResult.Outcome != "rerouted" {
		t.Fatalf("expected rerouted outcome for unrelated shared drift, got %+v", moduleResult)
	}
	if moduleResult.FallbackReasonCode != "rule_drift" {
		t.Fatalf("expected rule_drift, got %+v", moduleResult)
	}
	if moduleResult.NextCommand != "unit_stable_verify" {
		t.Fatalf("expected next command unit_stable_verify, got %s", moduleResult.NextCommand)
	}
}

func TestSyncImpactMixedRuleRefsStillInvalidateOnNonExemptRef(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupCandidateSharedRepo(t, repoRoot)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs:",
		"   - c_b_rule_demo@0.1.0",
		"   - c_b_rule_extra@0.2.0",
		"",
	}, "\n"))

	writeSharedFileAtPath(t, repoRoot, "docs/specs/rules/candidate/c_b_rule_extra.md", `---
rule_id: shared_extra
rule_scope: bound
layer: candidate
rule_version: 0.2.0
bound_objects:
  - unit:demo
---

# Shared

Body stays the same.
`)

	snap, err := snapshot.RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	writeProcessFile(t, repoRoot, "check", renderModuleProcessSnapshotForTest(
		t,
		repoRoot,
		"check",
		"demo",
		snap.ModuleAppendixSnapshot,
		snap.RuleSnapshot,
	))

	writeSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.1.0
bound_objects:
  - unit:demo
  - unit:module_other
---

# Shared

Body stays the same.
`)
	writeSharedFileAtPath(t, repoRoot, "docs/specs/rules/candidate/c_b_rule_extra.md", `---
rule_id: shared_extra
rule_scope: bound
layer: candidate
rule_version: 0.2.0
bound_objects:
  - unit:demo
---

# Shared

Body changed.
`)

	result, err := SyncImpact(repoRoot, Options{
		RuleRefs:                     []string{sharedRef, "c_b_rule_extra@0.2.0"},
		BoundObjectsOnlyRuleFileRefs: []string{"docs/specs/rules/candidate/c_b_rule_demo.md"},
	})
	if err == nil || !strings.Contains(err.Error(), "bound_objects-only sync is no longer supported") {
		t.Fatalf("expected bound_objects-only rejection, got result=%+v err=%v", result, err)
	}
}

func TestSyncImpactDoesNotExpandScopeWithExplicitModuleSelector(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupCandidateSharedRepo(t, repoRoot)

	statusPath := filepath.Join(repoRoot, "docs/specs/_status.md")
	mustWriteFile(t, statusPath, strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_plan` | current round |",
		"| `unit` | `module_other` | `no` | `yes` | `candidate` | `unit_plan` | current round |",
	}, "\n")+"\n")

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "module_other")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: module_other",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs: none",
		"",
	}, "\n"))

	moduleDemoRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(moduleDemoRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs: none",
		"",
	}, "\n"))

	result, err := SyncImpact(repoRoot, Options{
		Modules:  []string{"module_other"},
		RuleRefs: []string{sharedRef},
	})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.ScopedModules) != 0 {
		t.Fatalf("expected explicit module selector not to widen scope, got %+v", result.ScopedModules)
	}
	if len(result.ModuleResults) != 0 {
		t.Fatalf("expected no module results, got %+v", result.ModuleResults)
	}
}

func TestSyncImpactIncludesCandidateModuleWhenSelectedBindingWasRemovedFromCurrentTruth(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupCandidateSharedRepo(t, repoRoot)
	processSnapshot, err := snapshot.LoadProcessSnapshot(repoRoot, "unit", "demo", "check")
	if err != nil {
		t.Fatalf("LoadProcessSnapshot: %v", err)
	}
	storedProcess := renderModuleProcessSnapshotForTest(
		t,
		repoRoot,
		"check",
		"demo",
		processSnapshot.ModuleAppendixSnapshot,
		processSnapshot.RuleSnapshot,
	)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs: none",
		"",
	}, "\n"))
	writeProcessFile(t, repoRoot, "check", storedProcess)

	result, err := SyncImpact(repoRoot, Options{RuleRefs: []string{sharedRef}})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.ScopedModules) != 0 {
		t.Fatalf("expected removed binding to be ignored after frontmatter migration, got %+v", result.ScopedModules)
	}
	if len(result.ModuleResults) != 0 {
		t.Fatalf("expected no module fallback after frontmatter migration, got %+v", result.ModuleResults)
	}
}

func TestSyncImpactIgnoresIncompleteRemovedBindingEvidenceForModule(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupCandidateSharedRepo(t, repoRoot)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs: none",
		"",
	}, "\n"))
	writeProcessFile(t, repoRoot, "check", strings.Join([]string{
		"rule_snapshot:",
		"  - rule_id: shared_demo",
		"    layer: candidate",
		"    file_ref: docs/specs/rules/candidate/c_b_rule_demo.md",
		"    version_ref: c_b_rule_demo@0.1.0",
		"    fingerprint: " + fingerprintForTest(t, filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md")),
	}, "\n"))

	result, err := SyncImpact(repoRoot, Options{RuleRefs: []string{sharedRef}})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.ScopedModules) != 0 {
		t.Fatalf("expected incomplete module evidence to be ignored, got %+v", result.ScopedModules)
	}
	if len(result.ModuleResults) != 0 {
		t.Fatalf("expected no module fallback from incomplete evidence, got %+v", result.ModuleResults)
	}
}

func TestSyncImpactIgnoresModuleEvidenceThatDoesNotMatchCurrentModuleIdentity(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupCandidateSharedRepo(t, repoRoot)
	processSnapshot, err := snapshot.LoadProcessSnapshot(repoRoot, "unit", "demo", "check")
	if err != nil {
		t.Fatalf("LoadProcessSnapshot: %v", err)
	}

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs: none",
		"",
	}, "\n"))

	processPath := filepath.Join(repoRoot, "docs/specs/_check_result/unit/demo.md")
	validProcess := renderModuleProcessSnapshotForTest(
		t,
		repoRoot,
		"check",
		"demo",
		processSnapshot.ModuleAppendixSnapshot,
		processSnapshot.RuleSnapshot,
	)
	rewritten := strings.Replace(validProcess, "truth_file_ref: docs/specs/units/candidate/c_unit_demo.md", "truth_file_ref: docs/specs/units/candidate/c_unit_other.md", 1)
	rewritten = strings.Replace(rewritten, "truth_fingerprint: ", "truth_fingerprint: wrong-", 1)
	mustWriteFile(t, processPath, "# check\n\n```yaml\n"+rewritten+"\n```\n")

	result, err := SyncImpact(repoRoot, Options{RuleRefs: []string{sharedRef}})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.ScopedModules) != 0 {
		t.Fatalf("expected mismatched module evidence to be rejected, got %+v", result.ScopedModules)
	}
	if len(result.ModuleResults) != 0 {
		t.Fatalf("expected no module fallback from mismatched module evidence, got %+v", result.ModuleResults)
	}
}

func TestSyncImpactIgnoresModuleEvidenceWhenCurrentTruthChangedBeyondRemovedBinding(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupCandidateSharedRepo(t, repoRoot)
	processSnapshot, err := snapshot.LoadProcessSnapshot(repoRoot, "unit", "demo", "check")
	if err != nil {
		t.Fatalf("LoadProcessSnapshot: %v", err)
	}
	storedProcess := renderModuleProcessSnapshotForTest(
		t,
		repoRoot,
		"check",
		"demo",
		processSnapshot.ModuleAppendixSnapshot,
		processSnapshot.RuleSnapshot,
	)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo Updated",
		"",
		"Body changed outside shared bindings.",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs: none",
		"",
	}, "\n"))
	writeProcessFile(t, repoRoot, "check", storedProcess)

	result, err := SyncImpact(repoRoot, Options{RuleRefs: []string{sharedRef}})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.ScopedModules) != 0 {
		t.Fatalf("expected stale module evidence to be rejected, got %+v", result.ScopedModules)
	}
	if len(result.ModuleResults) != 0 {
		t.Fatalf("expected no module fallback from stale module evidence, got %+v", result.ModuleResults)
	}
}

func TestSyncImpactRejectsAmbiguousRemovedBindingSharedID(t *testing.T) {
	repoRoot := t.TempDir()
	setupCandidateSharedRepo(t, repoRoot)

	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/stable"))
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_b_rule_demo.md"), strings.Join([]string{
		"---",
		"rule_id: shared_demo",
		"layer: stable",
		"rule_version: 1.0.0",
		"bound_objects:",
		"  - unit:demo",
		"---",
		"",
		"# Shared",
		"",
		"Stable body.",
		"",
	}, "\n"))
	writeSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.1.0
promotion_owner_unit: demo
bound_objects:
  - unit:demo
---

# Shared

Body stays the same.
`)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs:",
		"   - s_b_rule_demo@1.0.0",
		"",
	}, "\n"))
	storedProcess := renderModuleProcessSnapshotForTest(
		t,
		repoRoot,
		"check",
		"demo",
		nil,
		[]snapshot.RuleEntry{{
			RuleID:      "shared_demo",
			Layer:       "stable",
			FileRef:     "docs/specs/rules/stable/s_b_rule_demo.md",
			VersionRef:  "s_b_rule_demo@1.0.0",
			Fingerprint: fingerprintForTest(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_b_rule_demo.md")),
		}},
	)
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs: none",
		"",
	}, "\n"))
	writeProcessFile(t, repoRoot, "check", storedProcess)

	result, err := SyncImpact(repoRoot, Options{RuleIDs: []string{"shared_demo"}})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.ScopedModules) != 0 {
		t.Fatalf("expected ambiguous removed-binding evidence to be ignored after frontmatter migration, got %+v", result.ScopedModules)
	}
}

func TestSyncImpactRejectsAmbiguousCurrentBindingSharedID(t *testing.T) {
	repoRoot := t.TempDir()
	setupCandidateSharedRepo(t, repoRoot)

	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/stable"))
	writeSharedFileAtPath(t, repoRoot, "docs/specs/rules/stable/s_b_rule_demo.md", `---
rule_id: shared_demo
rule_scope: bound
layer: stable
rule_version: 1.0.0
bound_objects:
  - unit:demo
---

# Shared

Stable body.
`)
	writeSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.1.0
promotion_owner_unit: demo
bound_objects:
  - unit:demo
---

# Shared

Body stays the same.
`)

	_, err := SyncImpact(repoRoot, Options{RuleIDs: []string{"shared_demo"}})
	if err == nil || !strings.Contains(err.Error(), "multiple current shared layers exist") {
		t.Fatalf("expected ambiguous current-binding shared-id error, got %v", err)
	}
}

func TestSyncImpactRejectsUnsortedRuleRefsInCurrentTruth(t *testing.T) {
	repoRoot := t.TempDir()
	setupCandidateSharedRepo(t, repoRoot)

	writeSharedFileAtPath(t, repoRoot, "docs/specs/rules/candidate/c_b_rule_alpha.md", `---
rule_id: shared_alpha
rule_scope: bound
layer: candidate
rule_version: 0.1.0
bound_objects:
  - unit:demo
---

# Shared Alpha

Body stays the same.
`)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs:",
		"   - c_b_rule_demo@0.1.0",
		"   - c_b_rule_alpha@0.1.0",
		"",
	}, "\n"))

	_, err = SyncImpact(repoRoot, Options{RuleRefs: []string{"c_b_rule_demo@0.1.0"}})
	if err == nil || !strings.Contains(err.Error(), "rule_refs must be sorted") {
		t.Fatalf("expected unsorted rule_refs error, got %v", err)
	}
}

func TestSyncImpactIncludesRemovedBindingWhenSharedIDIsUnambiguous(t *testing.T) {
	repoRoot := t.TempDir()
	setupCandidateSharedRepo(t, repoRoot)

	processSnapshot, err := snapshot.LoadProcessSnapshot(repoRoot, "unit", "demo", "check")
	if err != nil {
		t.Fatalf("LoadProcessSnapshot: %v", err)
	}
	storedProcess := renderModuleProcessSnapshotForTest(
		t,
		repoRoot,
		"check",
		"demo",
		processSnapshot.ModuleAppendixSnapshot,
		processSnapshot.RuleSnapshot,
	)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs: none",
		"",
	}, "\n"))
	writeProcessFile(t, repoRoot, "check", storedProcess)

	result, err := SyncImpact(repoRoot, Options{RuleIDs: []string{"shared_demo"}})
	if err != nil {
		t.Fatalf("SyncImpact: %v", err)
	}
	if len(result.ScopedModules) != 0 {
		t.Fatalf("expected removed-binding shared-id evidence to be ignored after frontmatter migration, got %+v", result.ScopedModules)
	}
	if len(result.ModuleResults) != 0 {
		t.Fatalf("expected no module fallback after frontmatter migration, got %+v", result.ModuleResults)
	}
}

func TestSyncImpactRejectsEmptyRuleRefsList(t *testing.T) {
	repoRoot := t.TempDir()
	setupCandidateSharedRepo(t, repoRoot)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs:",
		"",
	}, "\n"))

	_, err = SyncImpact(repoRoot, Options{
		Modules:  []string{"demo"},
		RuleRefs: []string{"c_b_rule_demo@0.1.0"},
	})
	if err == nil || !strings.Contains(err.Error(), "must not be an empty list") {
		t.Fatalf("expected empty-list error, got %v", err)
	}
}

func TestSyncImpactRejectsDuplicateRuleRefs(t *testing.T) {
	repoRoot := t.TempDir()
	setupCandidateSharedRepo(t, repoRoot)

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs:",
		"   - c_b_rule_demo@0.1.0",
		"   - c_b_rule_demo@0.1.0",
		"",
	}, "\n"))

	_, err = SyncImpact(repoRoot, Options{
		Modules:  []string{"demo"},
		RuleRefs: []string{"c_b_rule_demo@0.1.0"},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate item") {
		t.Fatalf("expected duplicate-item error, got %v", err)
	}
}

func TestSyncImpactRejectsModulesOnlyScope(t *testing.T) {
	repoRoot := t.TempDir()
	setupStableSharedRepo(t, repoRoot)

	_, err := SyncImpact(repoRoot, Options{
		Modules: []string{"demo"},
	})
	if err == nil || !strings.Contains(err.Error(), "at least one of rule refs, rule ids, or deleted rule refs is required") {
		t.Fatalf("expected modules-only scope to be rejected, got %v", err)
	}
}

func TestSyncImpactRejectsUnknownSharedRefWithoutCurrentBindingReference(t *testing.T) {
	repoRoot := t.TempDir()
	setupCandidateSharedRepo(t, repoRoot)

	_, err := SyncImpact(repoRoot, Options{
		RuleRefs: []string{"c_b_rule_missing@9.9.9"},
	})
	if err == nil || !strings.Contains(err.Error(), "is not present under docs/specs/rules/ and is not referenced by current downstream bindings") {
		t.Fatalf("expected unknown rule ref error, got %v", err)
	}
}

func TestSyncImpactRejectsStableLandingModuleWithoutStableLandingRuleRefs(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupStableSharedRepo(t, repoRoot)

	_, err := SyncImpact(repoRoot, Options{
		RuleRefs:            []string{sharedRef},
		StableLandingModule: "demo",
	})
	if err == nil || !strings.Contains(err.Error(), "stable landing rule refs are required") {
		t.Fatalf("expected missing stable landing rule refs error, got %v", err)
	}
}

func TestSyncImpactRejectsUnknownStableLandingSharedRef(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupStableSharedRepo(t, repoRoot)

	_, err := SyncImpact(repoRoot, Options{
		RuleRefs:              []string{sharedRef},
		StableLandingModule:   "demo",
		StableLandingRuleRefs: []string{"s_b_rule_missing@9.9.9"},
	})
	if err == nil || !strings.Contains(err.Error(), "stable landing rule ref") {
		t.Fatalf("expected unknown stable landing rule ref error, got %v", err)
	}
}

func TestSyncImpactRejectsNonStableStableLandingModule(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupCandidateSharedRepo(t, repoRoot)

	_, err := SyncImpact(repoRoot, Options{
		RuleRefs:              []string{sharedRef},
		StableLandingModule:   "demo",
		StableLandingRuleRefs: []string{sharedRef},
	})
	if err == nil || !strings.Contains(err.Error(), "must currently be at active layer stable") {
		t.Fatalf("expected non-stable stable landing unit error, got %v", err)
	}
}

func TestSyncImpactRejectsStableLandingSharedRefOutsideSelectedScope(t *testing.T) {
	repoRoot := t.TempDir()
	setupStableSharedRepo(t, repoRoot)

	writeSharedFileAtPath(t, repoRoot, "docs/specs/rules/stable/s_b_rule_extra.md", `---
rule_id: shared_extra
rule_scope: bound
layer: stable
rule_version: 1.1.0
bound_objects:
  - unit:demo
---

# Shared

Body stays the same.
`)

	_, err := SyncImpact(repoRoot, Options{
		RuleRefs:              []string{"s_b_rule_demo@1.0.0"},
		StableLandingModule:   "demo",
		StableLandingRuleRefs: []string{"s_b_rule_extra@1.1.0"},
	})
	if err == nil || !strings.Contains(err.Error(), "is not selected for stable landing unit") {
		t.Fatalf("expected stable landing rule ref outside scope error, got %v", err)
	}
}

func TestReconcileBoundModulesUpdatesTouchedSharedFiles(t *testing.T) {
	repoRoot := t.TempDir()
	sharedRef := setupCandidateSharedRepo(t, repoRoot)

	writeSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.1.0
bound_objects:
  - unit:module_other
---

# Shared

Body stays the same.
`)

	result, err := ReconcileBoundModules(repoRoot, ReconcileBoundModulesOptions{
		RuleRefs: []string{sharedRef},
	})
	if err == nil || !strings.Contains(err.Error(), "reconcile-bound-objects is no longer supported") {
		t.Fatalf("expected reconcile-bound-objects rejection, got result=%+v err=%v", result, err)
	}
}

func setupCandidateSharedRepo(t *testing.T, repoRoot string) string {
	t.Helper()
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/candidate"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))

	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_plan` | current round |",
	}, "\n")+"\n")

	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: candidate",
		"version: 0.1.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs:",
		"   - c_b_rule_demo@0.1.0",
		"",
	}, "\n"))

	writeSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.1.0
bound_objects:
  - unit:demo
---

# Shared

Body stays the same.
`)

	snap, err := snapshot.RebuildCurrent(repoRoot, "demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	writeProcessFile(t, repoRoot, "check", renderModuleProcessSnapshotForTest(t, repoRoot, "check", "demo", snap.ModuleAppendixSnapshot, snap.RuleSnapshot))
	initGitRepo(t, repoRoot)
	return "c_b_rule_demo@0.1.0"
}

func setupStableSharedRepo(t *testing.T, repoRoot string) string {
	t.Helper()
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.StableDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/stable"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))

	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | stable round |",
	}, "\n")+"\n")

	mainSpecRef, err := specpaths.MainSpecFileRef("stable", "demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join([]string{
		"---",
		"id: demo",
		"layer: stable",
		"version: 1.0.0",
		"---",
		"",
		"# Demo",
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs:",
		"   - s_b_rule_demo@1.0.0",
		"",
	}, "\n"))

	writeStableSharedFile(t, repoRoot, `---
rule_id: shared_demo
rule_scope: bound
layer: stable
rule_version: 1.0.0
bound_objects:
  - unit:demo
---

# Shared

Body stays the same.
`)

	initGitRepo(t, repoRoot)
	return "s_b_rule_demo@1.0.0"
}

func setupStableGlobalRuleRepo(t *testing.T, repoRoot string) string {
	t.Helper()
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.StableDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/stable"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))

	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `agent` | `yes` | `no` | `stable` | `unit_fork` | stable round |",
		"| `unit` | `demo` | `no` | `yes` | `candidate` | `unit_plan` | current round |",
	}, "\n")+"\n")

	writeUnitSpecWithRuleRefs(t, repoRoot, "stable", "agent", nil)
	writeUnitSpecWithRuleRefs(t, repoRoot, "candidate", "demo", nil)
	writeSharedFileAtPath(t, repoRoot, "docs/specs/rules/stable/s_g_rule_repository_baseline.md", `---
rule_id: g_rule_repository_baseline
rule_scope: global
layer: stable
rule_version: 1.1.0
---

# Repository Baseline

Global baseline.
`)

	initGitRepo(t, repoRoot)
	return "s_g_rule_repository_baseline@1.1.0"
}

func setupStableLandingRetargetRepo(t *testing.T, repoRoot string, retargetAgentToStable, agentStable bool) (string, string) {
	t.Helper()
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.StableDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/candidate"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/stable"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/_check_result/unit"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))

	agentLayer := "candidate"
	agentStableCol := "no"
	agentCandidateCol := "yes"
	agentNext := "unit_plan"
	if agentStable {
		agentLayer = "stable"
		agentStableCol = "yes"
		agentCandidateCol = "no"
		agentNext = "unit_fork"
	}
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), strings.Join([]string{
		"# Spec Status",
		"",
		"## Formal Objects",
		"",
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |",
		"|---|---|---|---|---|---|---|",
		"| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | stable round |",
		"| `unit` | `agent` | `" + agentStableCol + "` | `" + agentCandidateCol + "` | `" + agentLayer + "` | `" + agentNext + "` | current round |",
	}, "\n")+"\n")

	writeUnitSpecWithRuleRefs(t, repoRoot, "stable", "demo", []string{"s_b_rule_demo@1.0.0"})
	if agentStable {
		initialRefs := []string{"c_b_rule_demo@1.0.0"}
		if retargetAgentToStable {
			initialRefs = []string{"s_b_rule_demo@1.0.0"}
		}
		writeUnitSpecWithRuleRefs(t, repoRoot, "stable", "agent", initialRefs)
	} else {
		writeUnitSpecWithRuleRefs(t, repoRoot, "candidate", "agent", []string{"c_b_rule_demo@1.0.0"})
	}

	writeSharedFileAtPath(t, repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md", `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 1.0.0
promotion_owner_unit: demo
bound_objects:
  - unit:agent
---

# Shared

Body promoted unchanged.
`)
	writeSharedFileAtPath(t, repoRoot, "docs/specs/rules/stable/s_b_rule_demo.md", `---
rule_id: shared_demo
rule_scope: bound
layer: stable
rule_version: 1.0.0
bound_objects:
  - unit:demo
  - unit:agent
---

# Shared

Body promoted unchanged.
`)

	if !agentStable {
		snap, err := snapshot.RebuildCurrent(repoRoot, "agent")
		if err != nil {
			t.Fatalf("RebuildCurrent: %v", err)
		}
		writeNamedProcessFile(t, repoRoot, "check", "agent", renderModuleProcessSnapshotForTest(t, repoRoot, "check", "agent", snap.ModuleAppendixSnapshot, snap.RuleSnapshot))
	}

	if retargetAgentToStable && !agentStable {
		writeUnitSpecWithRuleRefs(t, repoRoot, agentLayer, "agent", []string{"s_b_rule_demo@1.0.0"})
	}

	initGitRepo(t, repoRoot)
	return "c_b_rule_demo@1.0.0", "s_b_rule_demo@1.0.0"
}

func writeUnitSpecWithRuleRefs(t *testing.T, repoRoot, layer, unit string, ruleRefs []string) {
	t.Helper()
	mainSpecRef, err := specpaths.MainSpecFileRef(layer, unit)
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	lines := []string{
		"---",
		"id: " + unit,
		"layer: " + layer,
		"version: 0.1.0",
		"---",
		"",
		"# " + unit,
		"",
		"## Rule Alignment",
		"",
		"2. rule_refs:",
	}
	if len(ruleRefs) == 0 {
		lines[len(lines)-1] = "2. rule_refs: none"
	} else {
		for _, ref := range ruleRefs {
			lines = append(lines, "   - "+ref)
		}
	}
	lines = append(lines, "")
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), strings.Join(lines, "\n"))
}

func writeSharedFileAtPath(t *testing.T, repoRoot, relPath, content string) {
	t.Helper()
	mustWriteFile(t, filepath.Join(repoRoot, relPath), content)
}

func writeRepositoryMappingFile(t *testing.T, repoRoot, version string) {
	t.Helper()
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.RepositoryMappingFileRef)), strings.Join([]string{
		"---",
		"id: repository_mapping",
		"version: " + version,
		"---",
		"",
		"# Repository Mapping",
		"",
	}, "\n"))
}

func writeProcessFile(t *testing.T, repoRoot, processKind, snapshotBody string) {
	t.Helper()
	writeNamedProcessFile(t, repoRoot, processKind, "demo", snapshotBody)
}

func writeNamedProcessFile(t *testing.T, repoRoot, processKind, object, snapshotBody string) {
	t.Helper()
	objectType := "unit"
	dir := map[string]string{
		"check":  filepath.ToSlash(filepath.Join("docs/specs/_check_result", objectType)),
		"plan":   "docs/specs/_plans",
		"verify": filepath.ToSlash(filepath.Join("docs/specs/_verify_result", objectType)),
	}[processKind]
	mustMkdirAll(t, filepath.Join(repoRoot, dir))
	content := fmt.Sprintf("# %s\n\n```yaml\n%s\n```\n", processKind, snapshotBody)
	mustWriteFile(t, filepath.Join(repoRoot, dir, object+".md"), content)
}

func writeSharedFile(t *testing.T, repoRoot, content string) {
	t.Helper()
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md"), content)
}

func writeStableSharedFile(t *testing.T, repoRoot, content string) {
	t.Helper()
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_b_rule_demo.md"), content)
}

func initGitRepo(t *testing.T, repoRoot string) {
	t.Helper()
	runGit(t, repoRoot, "init")
	runGit(t, repoRoot, "config", "user.name", "SpecFlow Test")
	runGit(t, repoRoot, "config", "user.email", "specflow@example.com")
	runGit(t, repoRoot, "add", ".")
	runGit(t, repoRoot, "commit", "-m", "init")
}

func runGit(t *testing.T, repoRoot string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
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
	content = withCandidateAcceptanceFixture(path, content)
	content = testfixtures.NormalizeSpecFlowContent(path, content)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func withCandidateAcceptanceFixture(path, content string) string {
	object, ok := candidateObjectNameFromPath(path)
	if !ok {
		return content
	}
	if strings.Contains(content, "acceptance_item_set:") {
		return content
	}
	lines := append([]string{
		strings.TrimRight(content, "\n"),
		"",
	}, acceptanceSectionFixtureLines(object)...)
	return strings.Join(lines, "\n") + "\n"
}

func candidateObjectNameFromPath(path string) (string, bool) {
	normalizedPath := filepath.ToSlash(path)
	base := strings.TrimSuffix(filepath.Base(path), ".md")
	switch {
	case strings.Contains(normalizedPath, "docs/specs/units/candidate/c_unit_"):
		return strings.TrimPrefix(base, "c_unit_"), true
	default:
		return "", false
	}
}

func acceptanceSectionFixtureLines(object string) []string {
	return []string{
		"## Testability / Acceptance Criteria",
		"",
		"acceptance_item_set:",
		"  - id: " + object + ".acceptance",
		"    target: " + object + " behavior is accepted.",
		"    verification_surface: internal_flow",
		"    implementation_surface: AgentCore/internal/" + object,
		"    verification_method: Go test for " + object + " behavior.",
		"    pass_condition: " + object + " behavior passes the declared checks.",
		"    not_runnable_yet: no",
	}
}

func acceptanceItemSnapshotFixtureLines(object string) []string {
	return []string{
		"id: " + object + ".acceptance",
		"verification_surface: internal_flow",
		"not_runnable_yet: no",
	}
}

func repositoryMappingSnapshotFixtureLines(entry snapshot.RepositoryMappingEntry) []string {
	return []string{
		"file_ref: " + entry.FileRef,
		"version_ref: " + entry.VersionRef,
		"fingerprint: " + entry.Fingerprint,
	}
}

func renderModuleProcessSnapshotForTest(t *testing.T, repoRoot, processKind, module string, appendixEntries []snapshot.AppendixEntry, sharedEntries []snapshot.RuleEntry) string {
	t.Helper()
	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", module)
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	truthFingerprint := fingerprintForTest(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)))
	scalars := []string{
		"object_type: unit",
		"object_ref: " + module,
		"gate: " + map[string]string{"check": "unit_check", "verify": "unit_verify"}[processKind],
		"decision: pass",
		"allow_next: true",
		"next_command: " + map[string]string{"check": "unit_plan", "verify": "unit_promote"}[processKind],
		"blocking_summary: none",
		"coverage_summary: current candidate",
		"truth_layer_ref: candidate",
		"truth_file_ref: " + mainSpecRef,
		"truth_version_ref: c_unit_" + module + "@0.1.0",
		"truth_fingerprint: " + truthFingerprint,
		"evaluation_mode: independent",
		"reviewer_result: pass",
		"reviewer_context: minimal_context",
		"review_input_refs: " + mainSpecRef,
		"review_findings: none",
		"human_decision_refs: none",
	}
	if processKind == "verify" {
		scalars = append(scalars, "verification_scope_ref: current candidate")
	}
	appendixLines := []string{}
	for _, entry := range appendixEntries {
		appendixLines = append(appendixLines,
			"file_ref: "+entry.FileRef,
			"fingerprint: "+entry.Fingerprint,
		)
	}
	sharedLines := []string{}
	for _, entry := range sharedEntries {
		sharedLines = append(sharedLines,
			"rule_id: "+entry.RuleID,
			"layer: "+entry.Layer,
			"file_ref: "+entry.FileRef,
			"version_ref: "+entry.VersionRef,
			"fingerprint: "+entry.Fingerprint,
		)
	}
	lists := [][]string{
		append([]string{"acceptance_item_set: "}, prefixNestedList(acceptanceItemSnapshotFixtureLines(module))...),
		append([]string{"unit_appendix_snapshot: " + noneOrBlank(appendixLines)}, prefixNestedList(appendixLines)...),
		append([]string{"rule_snapshot: " + noneOrBlank(sharedLines)}, prefixNestedList(sharedLines)...),
	}
	if processKind == "verify" {
		lists = append(lists, []string{"acceptance_item_evidence_matrix: none"})
	}
	return renderSnapshotBodyForTest(scalars, lists, false)
}

func renderSnapshotBodyForTest(scalars []string, lists [][]string, bulletFormat bool) string {
	lines := []string{}
	for _, scalar := range scalars {
		if bulletFormat {
			key, value, _ := strings.Cut(scalar, ": ")
			lines = append(lines, fmt.Sprintf("- `%s`: `%s`", key, value))
			continue
		}
		lines = append(lines, scalar)
	}
	for _, list := range lists {
		header := list[0]
		items := list[1:]
		if bulletFormat {
			key, value, _ := strings.Cut(header, ": ")
			if value == "none" {
				lines = append(lines, fmt.Sprintf("- `%s`: `none`", key))
				continue
			}
			lines = append(lines, fmt.Sprintf("- `%s`:", key))
			for _, item := range items {
				trimmed := strings.TrimSpace(item)
				trimmed = strings.TrimPrefix(trimmed, "- ")
				key, value, _ := strings.Cut(trimmed, ": ")
				lines = append(lines, fmt.Sprintf("  - `%s`: `%s`", key, value))
			}
			continue
		}
		lines = append(lines, header)
		lines = append(lines, items...)
	}
	return strings.Join(lines, "\n")
}

func noneOrBlank(lines []string) string {
	if len(lines) == 0 {
		return "none"
	}
	return ""
}

func prefixNestedList(lines []string) []string {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		result = append(result, "  - "+line)
	}
	return result
}

func fingerprintForTest(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	text := strings.ReplaceAll(string(content), "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")
	text += "\n"
	sum := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", sum)
}
