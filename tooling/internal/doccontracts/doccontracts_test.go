package doccontracts

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/reviewscope"
)

func TestActiveDocsDoNotDeprecateUnitCheck(t *testing.T) {
	repoRoot := findRepoRoot(t)
	files := activeDocFiles(t, repoRoot)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)unit_check.{0,80}deprecated`),
		regexp.MustCompile(`(?is)deprecated.{0,80}unit_check`),
		regexp.MustCompile(`(?is)unit_check.{0,80}merged.{0,40}unit_plan`),
		regexp.MustCompile(`(?is)unit_check.{0,80}removed`),
		regexp.MustCompile(`(?is)unit_check.{0,80}废弃`),
		regexp.MustCompile(`(?is)废弃.{0,80}unit_check`),
		regexp.MustCompile(`(?is)unit_check.{0,80}合并.{0,40}unit_plan`),
		regexp.MustCompile(`(?is)unit_check.{0,80}并入.{0,40}unit_plan`),
	}

	for _, relPath := range files {
		content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relPath)))
		if err != nil {
			t.Fatalf("read %s: %v", relPath, err)
		}
		for _, pattern := range patterns {
			if pattern.Match(content) {
				t.Fatalf("%s matches forbidden unit_check deprecation/merge contract pattern %q", relPath, pattern.String())
			}
		}
	}
}

func TestLifecycleContextCardsUseFixedSections(t *testing.T) {
	repoRoot := findRepoRoot(t)
	files := []string{
		"framework/lifecycle/unit_init_new_fork.md",
		"framework/lifecycle/unit_check.md",
		"framework/lifecycle/unit_plan.md",
		"framework/lifecycle/unit_impl.md",
		"framework/lifecycle/unit_verify.md",
		"framework/lifecycle/unit_promote.md",
		"framework/lifecycle/unit_stable_verify.md",
	}
	sections := []string{
		"## Required Context",
		"## Allowed Writes",
		"## Forbidden Writes",
		"## On-Demand Expansions",
		"## Independent Evaluation",
		"## Close Requirements",
	}

	for _, relPath := range files {
		content := readDocContractFile(t, repoRoot, relPath)
		lastIndex := -1
		for _, section := range sections {
			index := strings.Index(content, section)
			if index < 0 {
				t.Fatalf("%s missing required Context Card section %q", relPath, section)
			}
			if index <= lastIndex {
				t.Fatalf("%s Context Card section %q is out of order", relPath, section)
			}
			lastIndex = index
		}
	}
}

func TestLifecycleCardsUseCanonicalUnitTruthPaths(t *testing.T) {
	repoRoot := findRepoRoot(t)
	for _, relPath := range []string{
		"framework/lifecycle/unit_init_new_fork.md",
		"framework/lifecycle/unit_check.md",
		"framework/lifecycle/unit_plan.md",
		"framework/lifecycle/unit_impl.md",
		"framework/lifecycle/unit_verify.md",
		"framework/lifecycle/unit_promote.md",
		"framework/lifecycle/unit_stable_verify.md",
	} {
		content := readDocContractFile(t, repoRoot, relPath)
		for _, forbidden := range []string{
			"docs/specs/units/candidate/{unit}.md",
			"docs/specs/units/stable/{unit}.md",
		} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("%s must not reference non-canonical unit truth path %q", relPath, forbidden)
			}
		}
	}

	required := map[string][]string{
		"framework/lifecycle/unit_init_new_fork.md": {
			"docs/specs/units/candidate/c_unit_{unit}.md",
			"docs/specs/units/stable/s_unit_{unit}.md",
		},
		"framework/lifecycle/unit_check.md": {
			"docs/specs/units/candidate/c_unit_{unit}.md",
		},
		"framework/lifecycle/unit_plan.md": {
			"docs/specs/units/candidate/c_unit_{unit}.md",
		},
		"framework/lifecycle/unit_impl.md": {
			"docs/specs/units/candidate/c_unit_{unit}.md",
		},
		"framework/lifecycle/unit_verify.md": {
			"docs/specs/units/candidate/c_unit_{unit}.md",
		},
		"framework/lifecycle/unit_promote.md": {
			"docs/specs/units/candidate/c_unit_{unit}.md",
			"docs/specs/units/stable/s_unit_{unit}.md",
		},
		"framework/lifecycle/unit_stable_verify.md": {
			"docs/specs/units/stable/s_unit_{unit}.md",
		},
	}
	for relPath, phrases := range required {
		content := readDocContractFile(t, repoRoot, relPath)
		for _, phrase := range phrases {
			if !strings.Contains(content, phrase) {
				t.Fatalf("%s missing canonical unit truth path %q", relPath, phrase)
			}
		}
	}
}

func TestLifecycleRuleExpansionsRouteToRuleGovernance(t *testing.T) {
	repoRoot := findRepoRoot(t)
	files := []string{
		"framework/lifecycle/unit_init_new_fork.md",
		"framework/lifecycle/unit_check.md",
		"framework/lifecycle/unit_plan.md",
		"framework/lifecycle/unit_impl.md",
		"framework/lifecycle/unit_verify.md",
		"framework/lifecycle/unit_promote.md",
		"framework/lifecycle/unit_stable_verify.md",
	}

	for _, relPath := range files {
		content := readDocContractFile(t, repoRoot, relPath)
		onDemand := sectionBetween(t, content, "## On-Demand Expansions", "## Independent Evaluation", relPath)
		if strings.Contains(onDemand, "framework/governance/review.md") {
			t.Fatalf("%s must route rule changes to rule governance, not governance review", relPath)
		}
		for _, phrase := range []string{
			"framework/governance/rule_system.md",
			"framework/governance/rules/rule_escape.md",
		} {
			if !strings.Contains(onDemand, phrase) {
				t.Fatalf("%s On-Demand Expansions missing rule-governance route %q", relPath, phrase)
			}
		}
	}
}

func TestLifecycleOnDemandExpansionsUseExplicitOwnerRefs(t *testing.T) {
	repoRoot := findRepoRoot(t)
	files := []string{
		"framework/lifecycle/unit_check.md",
		"framework/lifecycle/unit_plan.md",
		"framework/lifecycle/unit_impl.md",
		"framework/lifecycle/unit_verify.md",
		"framework/lifecycle/unit_promote.md",
		"framework/lifecycle/unit_stable_verify.md",
	}

	for _, relPath := range files {
		content := readDocContractFile(t, repoRoot, relPath)
		onDemand := sectionBetween(t, content, "## On-Demand Expansions", "## Independent Evaluation", relPath)
		for _, forbidden := range []string{
			"recovery guidance",
			"migration guidance",
			"lifecycle fork guidance",
		} {
			if strings.Contains(onDemand, forbidden) {
				t.Fatalf("%s On-Demand Expansions must name the owner file instead of %q", relPath, forbidden)
			}
		}
		for _, phrase := range []string{
			"framework/lifecycle/recovery.md",
			"framework/operations/migration.md",
		} {
			if !strings.Contains(onDemand, phrase) {
				t.Fatalf("%s On-Demand Expansions missing explicit owner ref %q", relPath, phrase)
			}
		}
	}

	stableVerify := readDocContractFile(t, repoRoot, "framework/lifecycle/unit_stable_verify.md")
	onDemand := sectionBetween(t, stableVerify, "## On-Demand Expansions", "## Independent Evaluation", "framework/lifecycle/unit_stable_verify.md")
	if !strings.Contains(onDemand, "framework/lifecycle/unit_init_new_fork.md") {
		t.Fatalf("unit_stable_verify On-Demand Expansions missing explicit fork owner ref")
	}
}

func TestLifecycleCardsRequirePreWriteCommandPreflight(t *testing.T) {
	repoRoot := findRepoRoot(t)
	checks := map[string]string{
		"framework/lifecycle/unit_plan.md":    "<tooling-root>/bin/specflowctl-<os>-<arch> command preflight --repo-root <repo-root> --command unit_plan --object-type unit --object <unit>",
		"framework/lifecycle/unit_impl.md":    "<tooling-root>/bin/specflowctl-<os>-<arch> command preflight --repo-root <repo-root> --command unit_impl --object-type unit --object <unit>",
		"framework/lifecycle/unit_verify.md":  "<tooling-root>/bin/specflowctl-<os>-<arch> command preflight --repo-root <repo-root> --command unit_verify --object-type unit --object <unit>",
		"framework/lifecycle/unit_promote.md": "<tooling-root>/bin/specflowctl-<os>-<arch> command preflight --repo-root <repo-root> --command unit_promote --object-type unit --object <unit>",
	}

	for relPath, phrase := range checks {
		content := readDocContractFile(t, repoRoot, relPath)
		if !strings.Contains(content, phrase) {
			t.Fatalf("%s missing pre-write command preflight rule %q", relPath, phrase)
		}
		if !strings.Contains(strings.ToLower(content), "before") {
			t.Fatalf("%s preflight rule must be written as a before-write gate", relPath)
		}
	}
}

func TestLifecycleCardsUseImplementedCommandCloseCLI(t *testing.T) {
	repoRoot := findRepoRoot(t)
	checks := map[string]string{
		"framework/lifecycle/unit_check.md":         "--command unit_check --object-type unit --object <unit> --outcome pass --apply",
		"framework/lifecycle/unit_plan.md":          "--command unit_plan --object-type unit --object <unit> --outcome plan_ready --apply",
		"framework/lifecycle/unit_impl.md":          "--command unit_impl --object-type unit --object <unit> --outcome ready_for_verify --apply",
		"framework/lifecycle/unit_verify.md":        "--command unit_verify --object-type unit --object <unit> --outcome ready_to_promote --apply",
		"framework/lifecycle/unit_promote.md":       "--command unit_promote --object-type unit --object <unit> --outcome promoted --apply",
		"framework/lifecycle/unit_stable_verify.md": "--command unit_stable_verify --object-type unit --object <unit> --outcome <outcome> --apply",
	}
	oldForm := regexp.MustCompile(`command close unit_[a-z_]+:\{unit\}`)

	for relPath, phrase := range checks {
		content := readDocContractFile(t, repoRoot, relPath)
		if oldForm.MatchString(content) {
			t.Fatalf("%s contains unsupported command close shorthand", relPath)
		}
		if !strings.Contains(content, phrase) {
			t.Fatalf("%s missing implemented command close form %q", relPath, phrase)
		}
	}
}

func TestLifecycleToolingCommandsUseToolingRoot(t *testing.T) {
	repoRoot := findRepoRoot(t)
	files := []string{
		"framework/lifecycle/unit_check.md",
		"framework/lifecycle/unit_plan.md",
		"framework/lifecycle/unit_impl.md",
		"framework/lifecycle/unit_verify.md",
		"framework/lifecycle/unit_promote.md",
		"framework/lifecycle/unit_stable_verify.md",
		"framework/process_snapshot_contract.md",
	}

	contextCard := readDocContractFile(t, repoRoot, "framework/core/context_card.md")
	for _, phrase := range []string{
		"`<tooling-root>/...`",
		"`specflow/tooling/...`",
		"`tooling/...`",
	} {
		if !strings.Contains(contextCard, phrase) {
			t.Fatalf("context_card.md missing tooling-root resolution phrase %q", phrase)
		}
	}

	for _, relPath := range files {
		content := readDocContractFile(t, repoRoot, relPath)
		if strings.Contains(content, "specflow/tooling/bin/specflowctl-<os>-<arch>") {
			t.Fatalf("%s must use <tooling-root>/bin for lifecycle tooling commands", relPath)
		}
		if !strings.Contains(content, "<tooling-root>/bin/specflowctl-<os>-<arch>") {
			t.Fatalf("%s missing <tooling-root>/bin tooling command", relPath)
		}
	}
}

func TestProcessSnapshotFallbackLayersMatchRecovery(t *testing.T) {
	repoRoot := findRepoRoot(t)
	processContract := readDocContractFile(t, repoRoot, "framework/process_snapshot_contract.md")
	recovery := readDocContractFile(t, repoRoot, "framework/lifecycle/recovery.md")

	for _, phrase := range []string{
		"`implementation_layer`",
		"`implementation_layer` -> `unit_impl`",
	} {
		if !strings.Contains(processContract, phrase) {
			t.Fatalf("process_snapshot_contract.md missing fallback phrase %q", phrase)
		}
	}
	for _, phrase := range []string{
		"`implementation_layer`",
		"`unit_impl`",
	} {
		if !strings.Contains(recovery, phrase) {
			t.Fatalf("recovery.md missing fallback phrase %q", phrase)
		}
	}
}

func TestImpactSyncDefinesStableUnitReleaseHandoff(t *testing.T) {
	repoRoot := findRepoRoot(t)
	impactSync := readDocContractFile(t, repoRoot, "framework/governance/impact_sync.md")
	for _, phrase := range []string{
		"Current candidate consumers may be mechanically retargeted",
		"Current stable consumers must not have stable truth rewritten by release-version tooling",
		"Remove stale `unit_stable_verify` evidence",
		"routes through `unit_fork:{unit}` and the owning unit lifecycle",
	} {
		if !strings.Contains(impactSync, phrase) {
			t.Fatalf("impact_sync.md missing stable unit release handoff phrase %q", phrase)
		}
	}

	toolingReadme := readDocContractFile(t, repoRoot, "tooling/README.md")
	unitRelease := sectionBetween(t, toolingReadme, "28. `unit release-version`", "29. `relation candidates`", "tooling/README.md")
	if strings.Contains(unitRelease, "stable current-layer units are rewritten directly") {
		t.Fatalf("tooling README must not authorize direct stable truth rewrite for unit release-version")
	}
	for _, phrase := range []string{
		"candidate current-layer units are rewritten directly",
		"stable current-layer units are not rewritten",
		"stale stable-verify evidence is removed",
		"routed to `unit_stable_verify`",
	} {
		if !strings.Contains(unitRelease, phrase) {
			t.Fatalf("tooling README unit release-version section missing phrase %q", phrase)
		}
	}
}

func TestUnitPromoteRequiresStablePromotionSummaryBeforeCleanup(t *testing.T) {
	repoRoot := findRepoRoot(t)
	unitPromote := readDocContractFile(t, repoRoot, "framework/lifecycle/unit_promote.md")
	for _, phrase := range []string{
		"`docs/specs/_verify_result/stable/unit/{unit}.md` as the stable promotion summary",
		"written before candidate verify evidence is cleaned up",
		"stable promotion summary writeback",
	} {
		if !strings.Contains(unitPromote, phrase) {
			t.Fatalf("unit_promote.md missing stable promotion summary cleanup guard phrase %q", phrase)
		}
	}

	toolingReadme := readDocContractFile(t, repoRoot, "tooling/README.md")
	if !strings.Contains(toolingReadme, "`unit_promote` cleanup requires the stable promotion summary") {
		t.Fatalf("tooling README must document stable promotion summary prerequisite for unit_promote cleanup")
	}
}

func TestInstalledEntriesDefineFrameworkRootRelativeRefs(t *testing.T) {
	repoRoot := findRepoRoot(t)
	for _, relPath := range []string{"templates/AGENTS.md", "templates/CLAUDE.md", "templates/GEMINI.md"} {
		block := managedSpecFlowBlock(t, readDocContractFile(t, repoRoot, relPath), relPath)
		for _, phrase := range []string{
			"Framework-root relative paths",
			"`framework/...`",
			"`specflow/framework/...`",
		} {
			if !strings.Contains(block, phrase) {
				t.Fatalf("%s managed block missing framework-root path rule %q", relPath, phrase)
			}
		}
	}
}

func TestEntryManagedBlocksDefineActionGuide(t *testing.T) {
	repoRoot := findRepoRoot(t)
	for _, relPath := range []string{"templates/AGENTS.md", "templates/CLAUDE.md", "templates/GEMINI.md"} {
		block := managedSpecFlowBlock(t, readDocContractFile(t, repoRoot, relPath), relPath)
		for _, phrase := range []string{
			"### 1. What specFlow Is",
			"### 2. Spec Document Types",
			"### 3. Spec Document Layers",
			"### 4. State Files",
			"### 5. Command Format",
			"### 6. Command Index",
			"### 7. Development Loop",
			"### 8. Natural-Language Requests",
			"### 9. First Read",
			"### 10. Pre-Action Rules",
			"### 11. No Custom Flow",
			"### 12. Hard Stops",
			"### 13. Required Output",
			"### 14. Rule Locations",
			"This repository uses specFlow to manage development work.",
			"specFlow maintains project documents",
			"Spec documents have two types:",
			"Spec documents have two layers:",
			"specFlow has two important state files:",
			"specFlow commands use this format:",
			"Exact commands have priority over natural-language routing.",
			"Before any lifecycle action, implementation proposal, reconciliation plan, test-repair plan, or repo-tracked file edit",
			"When the user request exactly matches one of these commands, read the linked owner file first and follow that file.",
			"Before reading any lifecycle command owner file except `unit_advance:{unit}`",
			"`specflow/framework/advance_policy.md`",
			"specflow/framework/lifecycle/overview.md",
			"`specflow/framework/governance/review.md`",
			"`specflow/framework/operations/migration.md`",
			"specflow/framework/operations/entry_routing.md",
			"specflow/framework/operations/implementation_change.md",
			"`unit_check:{unit}`",
			"`unit_plan:{unit}`",
			"`unit_verify:{unit}`",
			"`unit_promote:{unit}`",
			"`unit_stable_verify:{unit}`",
			"`unit_advance:{unit}`",
			"`spec_flow_review`",
			"`spec_flow_review:full`",
			"`spec_flow_design_review`",
			"`spec_flow_migrate`",
			"unit_new / unit_fork -> unit_check -> unit_plan -> unit_impl -> unit_verify -> unit_promote",
			"Lifecycle state may advance only through legal command closure.",
			"If the user request does not exactly match a specFlow command",
			"formal truth creation or change, no formal truth",
			"behavior, protocol, boundary, acceptance, rule, ownership, lifecycle",
			"lifecycle state, Next Command, stable/candidate state, unit phase",
			"skipping `_status.md` or owner checks",
			"field meaning, schema fields, output fields, fixture fields",
			"contract-like log fields, or downstream compatibility",
			"repository mapping, guidance",
			"reconciliation, audit, alignment, or gap-review",
			"specflow/framework/operations/entry_routing.md",
			"limited to implementation-side code, tests, configs, prompts, fixtures, integration scripts",
			"After the first owner routes the request, continue only through the routed owner.",
			"Before editing any implementation file, prove that the active owner allows the implementation edit.",
			"Testing, debugging, review, and exploration may inspect or verify. They do not authorize mutation by themselves.",
			"Do not guess from directory shape, code shape, or chat.",
			"Do not create a custom reconciliation, audit, alignment, or gap-review flow",
			"still route the request through the legal owner first",
			"A unit is one governed engineering responsibility.",
			"Stable is the accepted current project truth.",
			"Candidate is proposed next project truth.",
			"A rule is shared truth",
			"Do not close an advancing gate from self-assessment.",
			"The user-facing answer must not require the user to understand internal object-family names",
		} {
			if !strings.Contains(block, phrase) {
				t.Fatalf("%s managed block missing action-guide phrase %q", relPath, phrase)
			}
		}
		for _, forbidden := range []string{
			"### 1. First Required Step",
			"### 2. Classification Table",
			"| Classification | Next legal action | Forbidden action |",
			"`governing object`",
			"`next legal owner`",
			"`allowed next action`",
			"report these fields first",
			"This addendum is inserted at the start",
			"Treat yourself",
			"fresh executor",
			"no prior `specFlow` memory",
			"Read `specflow/framework/operations/implementation_change.md` before editing implementation files for a formal unit.",
			"Before editing any repo-tracked implementation file",
			"### 1. Before Editing Files",
			"Classify the request as implementation_only, truth_writeback_required, or boundary_unclear before editing.",
			"stop ordinary implementation and reroute through",
			"If the request may touch repo-tracked code, tests, configs, prompts, fixtures, integration scripts, or other implementation-side files and no exact lifecycle Context Card is already active",
		} {
			if strings.Contains(block, forbidden) {
				t.Fatalf("%s managed block must not use obsolete weak-entry phrase %q", relPath, forbidden)
			}
		}
	}
}

func TestFrameworkDocsUseFrameworkRootRelativeRefs(t *testing.T) {
	repoRoot := findRepoRoot(t)
	allowed := map[string]bool{
		"framework/core/context_card.md":       true,
		"framework/governance/review.md":       true,
		"framework/governance/review_scope.md": true,
		"framework/spec_flow_review.md":        true,
		"framework/spec_flow_design_review.md": true,
	}

	root := filepath.Join(repoRoot, "framework")
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".md") {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		relPath := filepath.ToSlash(rel)
		if allowed[relPath] {
			return nil
		}
		content := readDocContractFile(t, repoRoot, relPath)
		if strings.Contains(content, "specflow/framework/") {
			t.Fatalf("%s must use framework-root relative refs (`framework/...`) instead of installed-project refs", relPath)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk framework docs: %v", err)
	}
}

func TestSpecFlowReviewScopeUsesMergedEntryContextCard(t *testing.T) {
	repoRoot := findRepoRoot(t)
	content := readDocContractFile(t, repoRoot, "framework/spec_flow_review.md")
	if strings.Contains(content, "unit_new.md") {
		t.Fatalf("spec_flow_review.md must not reference removed scope file unit_new.md")
	}
	for _, phrase := range []string{
		"`lifecycle/unit_init_new_fork.md`",
		"for `unit_init`, `unit_new`, and `unit_fork`",
	} {
		if !strings.Contains(content, phrase) {
			t.Fatalf("spec_flow_review.md missing merged entry Context Card phrase %q", phrase)
		}
	}
}

func TestLayoutSensitiveContractsUseResolvedRoots(t *testing.T) {
	repoRoot := findRepoRoot(t)

	migration := readDocContractFile(t, repoRoot, "framework/operations/migration.md")
	migrationReads := sectionBetween(t, migration, "## 2. Required Reads", "## 3. Target Shape Rule", "framework/operations/migration.md")
	for _, phrase := range []string{
		"`installed_project`, `<template-root>` is `specflow/templates/` and `<tooling-root>` is `specflow/tooling/`",
		"`source_repo`, `<template-root>` is `templates/` and `<tooling-root>` is `tooling/`",
		"`<template-root>/**`",
		"`<tooling-root>/README.md`",
	} {
		if !strings.Contains(migrationReads, phrase) {
			t.Fatalf("migration required reads missing layout-root phrase %q", phrase)
		}
	}
	for _, forbidden := range []string{
		"`specflow/templates/**`",
		"`specflow/tooling/README.md`",
	} {
		if strings.Contains(migrationReads, forbidden) {
			t.Fatalf("migration required reads must use resolved roots instead of %q", forbidden)
		}
	}

	toolingPolicy := readDocContractFile(t, repoRoot, "framework/tooling_execution_policy.md")
	toolingIdentity := sectionBetween(t, toolingPolicy, "## 2. What Counts As Governance Tooling", "## 3. Tooling Necessity Contract", "framework/tooling_execution_policy.md")
	toolingReviewSet := sectionBetween(t, toolingPolicy, "The required tooling-contract document set is:", "Default `spec_flow_review` must not issue", "framework/tooling_execution_policy.md")
	toolingFreshness := sectionBetween(t, toolingPolicy, "## 7. Compiled Tooling Freshness", "## 8. Non-Goals", "framework/tooling_execution_policy.md")
	for _, phrase := range []string{
		"`installed_project`, `<tooling-root>` is `specflow/tooling/`",
		"`source_repo`, `<tooling-root>` is `tooling/`",
		"`<tooling-root>/README.md`",
		"`<tooling-root>/cmd/**/*.go`",
		"`<tooling-root>/internal/**/*.go`",
		"`<tooling-root>/manifest.tsv`",
		"`<tooling-root>/bin/`",
		"`<tooling-root>/scripts/`",
	} {
		if !strings.Contains(toolingPolicy, phrase) {
			t.Fatalf("tooling execution policy missing resolved tooling-root phrase %q", phrase)
		}
	}
	for _, section := range []struct {
		name    string
		content string
	}{
		{name: "tooling identity", content: toolingIdentity},
		{name: "tooling review set", content: toolingReviewSet},
		{name: "tooling freshness", content: toolingFreshness},
	} {
		for _, forbidden := range []string{
			"`specflow/tooling/README.md`",
			"`specflow/tooling/cmd/**/*.go`",
			"`specflow/tooling/internal/**/*.go`",
			"`specflow/tooling/manifest.tsv`",
			"`specflow/tooling/bin/`",
			"`specflow/tooling/scripts/`",
		} {
			if strings.Contains(section.content, forbidden) {
				t.Fatalf("%s must use <tooling-root> instead of %q", section.name, forbidden)
			}
		}
	}

	toolingReadme := readDocContractFile(t, repoRoot, "tooling/README.md")
	toolingInputSet := sectionBetween(t, toolingReadme, "## Tooling Input Set", "## Unified Status Table", "tooling/README.md")
	toolingReadmeFreshness := sectionBetween(t, toolingReadme, "## Freshness Rule", "The minimal stale-binary recovery", "tooling/README.md")
	for _, phrase := range []string{
		"`<tooling-root>/cmd/**/*.go`",
		"`<tooling-root>/internal/**/*.go`",
		"`<tooling-root>/manifest.tsv`",
		"`<tooling-root>/reader/web/**`",
		"`<tooling-root>/bin/`",
		"tooling-root-relative keys",
	} {
		if !strings.Contains(toolingReadme, phrase) {
			t.Fatalf("tooling/README.md missing resolved tooling-root phrase %q", phrase)
		}
	}
	for _, forbidden := range []string{
		"`specflow/tooling/cmd/**/*.go`",
		"`specflow/tooling/internal/**/*.go`",
		"`specflow/tooling/manifest.tsv`",
		"`specflow/tooling/reader/web/**`",
	} {
		if strings.Contains(toolingInputSet, forbidden) {
			t.Fatalf("tooling input set must use <tooling-root> instead of %q", forbidden)
		}
	}
	if strings.Contains(toolingReadmeFreshness, "`specflow/tooling/bin/`") {
		t.Fatalf("tooling freshness contract must use <tooling-root> for binary paths")
	}
}

func TestUnitAdvanceRoutesThroughAdvancePolicy(t *testing.T) {
	repoRoot := findRepoRoot(t)
	for _, relPath := range []string{"templates/AGENTS.md", "templates/CLAUDE.md", "templates/GEMINI.md"} {
		block := managedSpecFlowBlock(t, readDocContractFile(t, repoRoot, relPath), relPath)
		if !strings.Contains(block, "specflow/framework/advance_policy.md") {
			t.Fatalf("%s managed block must route unit_advance through advance_policy.md", relPath)
		}
		if strings.Contains(block, "If the request exactly matches `unit_advance:{unit}`, read `specflow/framework/lifecycle/overview.md`.") {
			t.Fatalf("%s managed block must not route unit_advance only to lifecycle overview", relPath)
		}
	}

	routing := readDocContractFile(t, repoRoot, "framework/operations/entry_routing.md")
	if !strings.Contains(routing, "framework/advance_policy.md") {
		t.Fatalf("entry_routing.md must route unit_advance through advance_policy.md")
	}
	if strings.Contains(routing, "If the request is `unit_advance:{unit}`, read `framework/lifecycle/overview.md`.") {
		t.Fatalf("entry_routing.md must not route unit_advance only to lifecycle overview")
	}

	overview := readDocContractFile(t, repoRoot, "framework/lifecycle/overview.md")
	for _, phrase := range []string{
		"framework/advance_policy.md",
		"not the execution owner for `unit_advance:{unit}`",
	} {
		if !strings.Contains(overview, phrase) {
			t.Fatalf("lifecycle overview missing unit_advance boundary phrase %q", phrase)
		}
	}
}

func TestNaturalLanguageUnitRoutingSelectsLifecycleContextCard(t *testing.T) {
	repoRoot := findRepoRoot(t)
	routing := readDocContractFile(t, repoRoot, "framework/operations/entry_routing.md")
	for _, phrase := range []string{
		"For a natural-language unit lifecycle request, select one existing lifecycle command and its Context Card before any lifecycle write:",
		"its recorded `Next Command` is the only legal lifecycle command",
		"Select `unit_init:{unit}` only when an existing accepted capability already satisfies every direct first-stable onboarding condition.",
		"Select `unit_new:{unit}` when the request creates new candidate truth",
		"selects `unit_fork:{unit}` only when the recorded `Next Command` is `unit_fork`",
		"Do not invent a command alias, enter a generic unit lifecycle without an active Context Card, or ask the user to choose an internal command name.",
		"a natural-language unit request cannot be resolved to one legal existing lifecycle command and active Context Card from current durable truth",
	} {
		if !strings.Contains(routing, phrase) {
			t.Fatalf("entry_routing.md missing natural-language lifecycle selection contract %q", phrase)
		}
	}

	overview := readDocContractFile(t, repoRoot, "framework/lifecycle/overview.md")
	for _, phrase := range []string{
		"A lifecycle Context Card may be selected in either of two ways:",
		"the request exactly states one of the command forms below",
		"`framework/operations/entry_routing.md` resolves a natural-language request to one of the existing command forms below from current durable truth",
		"Only the existing exact command forms may select a lifecycle Context Card:",
	} {
		if !strings.Contains(overview, phrase) {
			t.Fatalf("lifecycle overview missing natural-language Context Card boundary phrase %q", phrase)
		}
	}
}

func TestImplementationSideRoutingUsesDirectImplementationGate(t *testing.T) {
	repoRoot := findRepoRoot(t)
	routing := readDocContractFile(t, repoRoot, "framework/operations/entry_routing.md")
	for _, phrase := range []string{
		"no exact lifecycle Context Card is already active",
		"`framework/onboarding_decision_policy.md`",
		"repo-tracked code, tests, configs, prompts, fixtures, integration scripts, or other implementation-side files",
		"That operation owns the implementation-only, truth-writeback-required, and boundary-unclear classification.",
		"Implementation permission must be proven before proposing or editing implementation-side files.",
	} {
		if !strings.Contains(routing, phrase) {
			t.Fatalf("entry_routing.md missing implementation-side routing contract %q", phrase)
		}
	}
}

func TestCompatibilityRoutingUsesGovernanceFrontDoor(t *testing.T) {
	repoRoot := findRepoRoot(t)
	content := readDocContractFile(t, repoRoot, "framework/operations/entry_routing.md")
	for _, phrase := range []string{
		"`spec_flow_review` -> `framework/governance/review.md`",
		"`spec_flow_design_review` -> `framework/governance/review.md`",
		"`framework/governance/review.md` decides the default path for each review entry.",
		"For `spec_flow_review`, the default is `scoped_review`; it delegates to `framework/spec_flow_review.md` only for exact `spec_flow_review:full`.",
		"For `spec_flow_design_review`, there is no scoped mode; `framework/governance/review.md` delegates to `framework/spec_flow_design_review.md` for the default full-scope design-baseline review.",
	} {
		if !strings.Contains(content, phrase) {
			t.Fatalf("entry_routing.md missing scoped governance front-door route %q", phrase)
		}
	}
	for _, forbidden := range []string{
		"`spec_flow_review` -> `framework/spec_flow_review.md`",
		"`spec_flow_design_review` -> `framework/spec_flow_design_review.md`",
		"`spec_flow_review` -> `specflow/framework/spec_flow_review.md`",
		"`spec_flow_design_review` -> `specflow/framework/spec_flow_design_review.md`",
		"`framework/governance/review.md` decides the default scoped review path and delegates to deep-audit owners only when explicit deep-audit intent is present.",
		"only when explicit mechanism deep-audit intent is present",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("entry_routing.md must not route plain governance review directly to %q", forbidden)
		}
	}
}

func TestEntryManagedBlocksRouteEntryCommandsToExistingOwner(t *testing.T) {
	repoRoot := findRepoRoot(t)
	ownerPath := "specflow/framework/lifecycle/unit_init_new_fork.md"
	if _, err := os.Stat(filepath.Join(repoRoot, "framework/lifecycle/unit_init_new_fork.md")); err != nil {
		t.Fatalf("expected source owner file to exist: %v", err)
	}
	for _, relPath := range []string{"templates/AGENTS.md", "templates/CLAUDE.md", "templates/GEMINI.md"} {
		block := managedSpecFlowBlock(t, readDocContractFile(t, repoRoot, relPath), relPath)
		for _, row := range []string{
			"| `unit_init:{unit}` | `" + ownerPath + "` |",
			"| `unit_new:{unit}` | `" + ownerPath + "` |",
			"| `unit_fork:{unit}` | `" + ownerPath + "` |",
		} {
			if !strings.Contains(block, row) {
				t.Fatalf("%s managed block missing entry-command owner row %q", relPath, row)
			}
		}
		for _, removedOwner := range []string{
			"specflow/framework/lifecycle/unit_init.md",
			"specflow/framework/lifecycle/unit_new.md",
			"specflow/framework/lifecycle/unit_fork.md",
		} {
			if strings.Contains(block, removedOwner) {
				t.Fatalf("%s managed block routes to removed owner %s", relPath, removedOwner)
			}
		}
	}
}

func TestEntryManagedBlocksStayLightweight(t *testing.T) {
	repoRoot := findRepoRoot(t)
	for _, relPath := range []string{"templates/AGENTS.md", "templates/CLAUDE.md", "templates/GEMINI.md"} {
		block := managedSpecFlowBlock(t, readDocContractFile(t, repoRoot, relPath), relPath)
		for _, forbidden := range []string{
			"recovery",
			"rule_topology",
			"rule topology",
		} {
			if strings.Contains(strings.ToLower(block), forbidden) {
				t.Fatalf("%s managed block should not default-reference %q", relPath, forbidden)
			}
		}
	}
}

func TestActiveDocsDoNotForceFlatPolicyForExactCommands(t *testing.T) {
	repoRoot := findRepoRoot(t)
	for _, relPath := range []string{"templates/AGENTS.md", "templates/CLAUDE.md", "templates/GEMINI.md"} {
		block := managedSpecFlowBlock(t, readDocContractFile(t, repoRoot, relPath), relPath)
		for _, forbidden := range []string{
			deletedFrameworkPolicyPath("command_", "policy.md"),
			"complete flat policy",
			"full flat policy",
			"read all policy",
		} {
			if strings.Contains(strings.ToLower(block), forbidden) {
				t.Fatalf("%s managed block should not require exact commands to read %q", relPath, forbidden)
			}
		}
	}
}

func TestLifecycleAuthorityDoesNotRequireFullScopeRun(t *testing.T) {
	repoRoot := findRepoRoot(t)
	files := []string{
		"framework/core/lifecycle_authority.md",
		"framework/lifecycle/overview.md",
		"framework/lifecycle/unit_check.md",
		"framework/lifecycle/unit_plan.md",
		"framework/lifecycle/unit_impl.md",
		"framework/lifecycle/unit_verify.md",
		"framework/lifecycle/unit_stable_verify.md",
		"framework/process_snapshot_contract.md",
	}
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)lifecycle progression.{0,80}only.{0,80}new.{0,40}full-scope`),
		regexp.MustCompile(`(?is)only one new.{0,40}full-scope`),
		regexp.MustCompile(`(?is)valid only from a new.{0,40}full-scope`),
		regexp.MustCompile(`(?is)new independent full-scope run`),
		regexp.MustCompile(`(?is)non-authoritative follow-up`),
		regexp.MustCompile(`(?is)authoritative validation`),
		regexp.MustCompile(`(?is)authoritative process validation`),
	}

	for _, relPath := range files {
		content := readDocContractFile(t, repoRoot, relPath)
		for _, pattern := range patterns {
			if pattern.MatchString(content) {
				t.Fatalf("%s reintroduces run-based lifecycle authority pattern %q", relPath, pattern.String())
			}
		}
	}
}

func TestLifecycleOwnersCarryCommandAuthority(t *testing.T) {
	repoRoot := findRepoRoot(t)
	authority := readDocContractFile(t, repoRoot, "framework/core/lifecycle_authority.md")
	for _, phrase := range []string{
		"current valid evidence",
		"required independent evaluation receipt",
		"deterministic validation",
		"successful `command close`",
	} {
		if !strings.Contains(authority, phrase) {
			t.Fatalf("lifecycle_authority.md missing lifecycle authority phrase %q", phrase)
		}
	}

	overview := readDocContractFile(t, repoRoot, "framework/lifecycle/overview.md")
	for _, phrase := range []string{
		"Only the existing exact command forms may select a lifecycle Context Card:",
		"Do not invent scenario commands, command aliases, or object-type shortcuts.",
		"Command close is the only lifecycle advancement authority for `_status.md`.",
	} {
		if !strings.Contains(overview, phrase) {
			t.Fatalf("lifecycle/overview.md missing command authority phrase %q", phrase)
		}
	}
}

func TestStableVerifyFallbackReturnsToStableVerify(t *testing.T) {
	repoRoot := findRepoRoot(t)
	status := readDocContractFile(t, repoRoot, "framework/core/status.md")
	for _, phrase := range []string{
		"Candidate verify evidence fallback returns to `unit_verify`.",
		"Stable verify evidence fallback returns to `unit_stable_verify`.",
	} {
		if !strings.Contains(status, phrase) {
			t.Fatalf("status.md missing stable verify fallback phrase %q", phrase)
		}
	}
	if strings.Contains(status, "Evidence fallback returns to `unit_verify`.") {
		t.Fatalf("status.md must not collapse stable verify evidence fallback into unit_verify")
	}
}

func TestUnitForkReadsControlledStableVerifyIntent(t *testing.T) {
	repoRoot := findRepoRoot(t)
	content := readDocContractFile(t, repoRoot, "framework/lifecycle/unit_init_new_fork.md")
	required := sectionBetween(t, content, "## Required Context", "## Allowed Writes", "framework/lifecycle/unit_init_new_fork.md")
	for _, phrase := range []string{
		"current valid `docs/specs/_stable_verify_result/unit/{unit}.md`",
		"`Next Command` is `unit_fork`",
		"`decision: controlled_repair_required`",
		"`candidate_intent=repair`",
		"`decision: controlled_change_required`",
		"`candidate_intent=change`",
		"`decision: aligned`",
		"does not force a candidate intent",
	} {
		if !strings.Contains(required, phrase) {
			t.Fatalf("unit_init_new_fork.md missing controlled stable verify handoff phrase %q", phrase)
		}
	}
}

func TestActiveDocsAndDefaultScopeDoNotReferenceRemovedEntryPaths(t *testing.T) {
	repoRoot := findRepoRoot(t)
	removed := removedEntryPathStrings()
	for _, relPath := range activeDocFiles(t, repoRoot) {
		content := readDocContractFile(t, repoRoot, relPath)
		for _, phrase := range removed {
			if strings.Contains(content, phrase) {
				t.Fatalf("%s still references removed entry path %q", relPath, phrase)
			}
		}
	}

	scope, err := reviewscope.CollectDefaultSpecFlowScopeForLayout(repoRoot, reviewscope.LayoutSourceRepo)
	if err != nil {
		t.Fatalf("CollectDefaultSpecFlowScopeForLayout source: %v", err)
	}
	allScopeFiles := append([]string{}, scope.FrameworkGuidelineFiles...)
	allScopeFiles = append(allScopeFiles, scope.CommandFiles...)
	allScopeFiles = append(allScopeFiles, scope.RuleGovernanceFiles...)
	allScopeFiles = append(allScopeFiles, scope.AgentOperabilityFiles...)
	allScopeFiles = append(allScopeFiles, scope.ToolingContractFiles...)
	for _, path := range allScopeFiles {
		for _, phrase := range removed {
			if strings.Contains(path, phrase) {
				t.Fatalf("default review scope still includes removed entry path %q in %q", phrase, path)
			}
		}
	}
}

func TestFreshnessCoreDefinesFixedLevels(t *testing.T) {
	repoRoot := findRepoRoot(t)
	content := readDocContractFile(t, repoRoot, "framework/core/freshness.md")
	for _, level := range []string{
		"`current`",
		"`text_drift`",
		"`semantic_drift`",
		"`acceptance_drift`",
		"`dependency_drift`",
		"`schema_drift`",
		"`unknown_drift`",
	} {
		if !strings.Contains(content, level) {
			t.Fatalf("freshness.md missing level %s", level)
		}
	}
	for _, phrase := range []string{
		"Only `text_drift` can reuse existing process evidence.",
		"deterministic tooling classifies",
		"independent reviewer receipt",
	} {
		if !strings.Contains(content, phrase) {
			t.Fatalf("freshness.md missing contract phrase %q", phrase)
		}
	}
}

func TestLifecycleCardsUseFreshnessOnlyOnDemand(t *testing.T) {
	repoRoot := findRepoRoot(t)
	files := []string{
		"framework/lifecycle/unit_check.md",
		"framework/lifecycle/unit_plan.md",
		"framework/lifecycle/unit_impl.md",
		"framework/lifecycle/unit_verify.md",
		"framework/lifecycle/unit_stable_verify.md",
	}
	for _, relPath := range files {
		content := readDocContractFile(t, repoRoot, relPath)
		required := sectionBetween(t, content, "## Required Context", "## Allowed Writes", relPath)
		if strings.Contains(required, "framework/core/freshness.md") {
			t.Fatalf("%s must not make freshness a default required context", relPath)
		}
		onDemand := sectionBetween(t, content, "## On-Demand Expansions", "## Independent Evaluation", relPath)
		if !strings.Contains(onDemand, "framework/core/freshness.md") {
			t.Fatalf("%s must reference freshness as on-demand expansion", relPath)
		}
	}
}

func TestActiveDocsDoNotForceFallbackForEveryFingerprintMismatch(t *testing.T) {
	repoRoot := findRepoRoot(t)
	files := activeDocFiles(t, repoRoot)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)every fingerprint mismatch.{0,80}(fallback|fall back|unit_check|reroute)`),
		regexp.MustCompile(`(?is)any fingerprint mismatch.{0,80}(fallback|fall back|unit_check|reroute)`),
		regexp.MustCompile(`(?is)fingerprint mismatch.{0,80}automatically.{0,80}(fallback|fall back|unit_check|reroute)`),
		regexp.MustCompile(`(?is)fingerprint drift.{0,80}always.{0,80}(fallback|fall back|unit_check|reroute)`),
	}

	for _, relPath := range files {
		content := readDocContractFile(t, repoRoot, relPath)
		for _, pattern := range patterns {
			if pattern.MatchString(content) {
				t.Fatalf("%s reintroduces unconditional fingerprint fallback pattern %q", relPath, pattern.String())
			}
		}
	}
}

func TestAdoptionModesContractDefinesIncrementalStarts(t *testing.T) {
	repoRoot := findRepoRoot(t)
	files := []string{
		"README.md",
		"README.zh-CN.md",
		"framework/core/adoption_modes.md",
	}
	modes := []string{
		"`reader-only`",
		"`implementation-only`",
		"`single-unit-trial`",
		"`unit-check-only`",
	}

	for _, relPath := range files {
		content := readDocContractFile(t, repoRoot, relPath)
		for _, mode := range modes {
			if !strings.Contains(content, mode) {
				t.Fatalf("%s missing adoption mode %s", relPath, mode)
			}
		}
	}

	contract := readDocContractFile(t, repoRoot, "framework/core/adoption_modes.md")
	for _, phrase := range []string{
		"not lifecycle states",
		"not process schema",
		"not harness commands",
		"not a mode-selection flag for `specflowctl init`",
		"do not introduce a new process file, lifecycle state, lifecycle command, harness command, or process schema field",
	} {
		if !strings.Contains(contract, phrase) {
			t.Fatalf("adoption_modes.md missing no-new-mechanism phrase %q", phrase)
		}
	}
}

func TestReadmeQuickStartDoesNotForceFullLifecycle(t *testing.T) {
	repoRoot := findRepoRoot(t)
	checks := map[string][]string{
		"README.md": {
			"After this step, choose an [Adoption Mode](#adoption-modes).",
			"`init` prepares the shared skeleton; it does not require you to run the whole lifecycle immediately.",
			"Installing specFlow does not commit a project to promotion, stable verification, governance review, or full lifecycle use.",
		},
		"README.zh-CN.md": {
			"完成这一步后，先选择一个[增量采用模式](#增量采用模式)。",
			"`init` 只是准备共享骨架，不要求你立刻使用完整生命周期。",
			"安装 specFlow 不等于必须立刻承诺 promotion、stable verification、governance review 或完整生命周期。",
		},
	}

	for relPath, phrases := range checks {
		content := readDocContractFile(t, repoRoot, relPath)
		for _, phrase := range phrases {
			if !strings.Contains(content, phrase) {
				t.Fatalf("%s missing incremental adoption quick-start phrase %q", relPath, phrase)
			}
		}
	}
}

func TestAdoptionModesAreNotUnimplementedCliModes(t *testing.T) {
	repoRoot := findRepoRoot(t)
	files := activeDocFiles(t, repoRoot)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)specflowctl\s+init\s+--mode`),
		regexp.MustCompile(`(?is)<specflow-binary>\s+init\s+--mode`),
		regexp.MustCompile(`(?is)init\s+--mode\s+(reader|implementation|single|unit)`),
	}

	for _, relPath := range files {
		content := readDocContractFile(t, repoRoot, relPath)
		for _, pattern := range patterns {
			if pattern.MatchString(content) {
				t.Fatalf("%s claims an unimplemented adoption-mode CLI pattern %q", relPath, pattern.String())
			}
		}
	}
}

func TestActiveDocsDoNotMakeLightweightModesEnterHeavyGatesByDefault(t *testing.T) {
	repoRoot := findRepoRoot(t)
	files := []string{
		"README.md",
		"README.zh-CN.md",
		"framework/core/adoption_modes.md",
		"framework/lifecycle/overview.md",
		"framework/operations/implementation_change.md",
	}
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)(reader-only|implementation-only|single-unit-trial|unit-check-only).{0,160}(must|requires|required).{0,80}(promotion|promote|stable verification|governance review)`),
		regexp.MustCompile(`(?is)(reader-only|implementation-only|single-unit-trial|unit-check-only).{0,160}(default|by default).{0,80}(promotion|promote|stable verification|governance review)`),
		regexp.MustCompile(`(?is)(reader-only|implementation-only|single-unit-trial|unit-check-only).{0,160}(完整生命周期|默认进入)`),
	}

	for _, relPath := range files {
		content := readDocContractFile(t, repoRoot, relPath)
		for _, line := range strings.Split(content, "\n") {
			if !strings.Contains(line, "reader-only") &&
				!strings.Contains(line, "implementation-only") &&
				!strings.Contains(line, "single-unit-trial") &&
				!strings.Contains(line, "unit-check-only") {
				continue
			}
			checkedLine := strings.ReplaceAll(line, "do not require", "do-not-need")
			checkedLine = strings.ReplaceAll(checkedLine, "does not require", "does-not-need")
			checkedLine = strings.ReplaceAll(checkedLine, "fix-required", "fix-needed")
			for _, pattern := range patterns {
				if pattern.MatchString(checkedLine) {
					t.Fatalf("%s makes a lightweight adoption mode enter a heavy gate by default: %q", relPath, line)
				}
			}
		}
	}

	overview := readDocContractFile(t, repoRoot, "framework/lifecycle/overview.md")
	for _, phrase := range []string{
		"framework/core/adoption_modes.md",
		"they do not require every project or task to run the full lifecycle by default",
	} {
		if !strings.Contains(overview, phrase) {
			t.Fatalf("lifecycle overview missing adoption-mode boundary phrase %q", phrase)
		}
	}

	implementation := readDocContractFile(t, repoRoot, "framework/operations/implementation_change.md")
	for _, phrase := range []string{
		"the `implementation-only` adoption mode",
		"not a shortcut around truth",
		"the user asks for an implementation-side proposal or asks to modify repo-tracked code, tests, or other implementation-side files",
		"Classify the request before proposing or editing implementation-side files:",
		"Before any implementation-side proposal or edit",
		"`implementation_only` must not authorize an implementation proposal or implementation-side edit.",
		"stop at the smallest legal truth step",
	} {
		if !strings.Contains(implementation, phrase) {
			t.Fatalf("implementation_change.md missing implementation-only boundary phrase %q", phrase)
		}
	}

	specWriteback := readDocContractFile(t, repoRoot, "framework/skills/spec-writeback-guidance/SKILL.md")
	if !strings.Contains(specWriteback, "Read `framework/operations/implementation_change.md` before any implementation-side proposal or edit.") {
		t.Fatalf("spec-writeback-guidance must not narrow implementation permission to edits only")
	}
}

func TestIndependentEvaluationDefinesFixedReviewerPacks(t *testing.T) {
	repoRoot := findRepoRoot(t)
	content := readDocContractFile(t, repoRoot, "framework/core/independent_evaluation.md")
	packs := []string{
		"`unit_check_pass`",
		"`unit_plan_plan_ready`",
		"`unit_verify_ready_to_promote`",
		"`unit_stable_verify_advancing`",
		"`freshness_text_drift_reuse`",
	}
	for _, pack := range packs {
		if !strings.Contains(content, pack) {
			t.Fatalf("independent_evaluation.md missing reviewer pack %s", pack)
		}
	}
	for _, phrase := range []string{
		"## Minimal Context",
		"## Reviewer Packs",
		"Review Standard Refs:",
		"Allowed Inputs:",
		"Forbidden Inputs:",
		"Evaluation Questions:",
		"Legal Output:",
		"## Handoff Requests",
		"specflowctl-<os>-<arch> evaluation request",
		"## Anti-Patterns",
		"specFlow does not create harness commands, reviewer sessions, tokens, or task scheduling.",
		"The reviewer must not inherit the executor's full working context as authority",
	} {
		if !strings.Contains(content, phrase) {
			t.Fatalf("independent_evaluation.md missing independent review contract phrase %q", phrase)
		}
	}
}

func TestLifecycleCardsReferenceFixedReviewerPacks(t *testing.T) {
	repoRoot := findRepoRoot(t)
	checks := map[string]string{
		"framework/lifecycle/unit_check.md":         "`unit_check_pass`",
		"framework/lifecycle/unit_plan.md":          "`unit_plan_plan_ready`",
		"framework/lifecycle/unit_verify.md":        "`unit_verify_ready_to_promote`",
		"framework/lifecycle/unit_stable_verify.md": "`unit_stable_verify_advancing`",
	}

	for relPath, pack := range checks {
		content := readDocContractFile(t, repoRoot, relPath)
		section := sectionBetween(t, content, "## Independent Evaluation", "## Close Requirements", relPath)
		if !strings.Contains(section, pack) {
			t.Fatalf("%s Independent Evaluation section missing reviewer pack %s", relPath, pack)
		}
		if !strings.Contains(section, "evaluation request") {
			t.Fatalf("%s Independent Evaluation section missing handoff request command", relPath)
		}
		if strings.Contains(section, "minimal review pack:") {
			t.Fatalf("%s should reference the fixed reviewer pack instead of an inline free-form pack", relPath)
		}
	}

	impl := readDocContractFile(t, repoRoot, "framework/lifecycle/unit_impl.md")
	implSection := sectionBetween(t, impl, "## Independent Evaluation", "## Close Requirements", "framework/lifecycle/unit_impl.md")
	for _, phrase := range []string{
		"`unit_impl` does not require an independent reviewer receipt",
		"it does not approve that code",
	} {
		if !strings.Contains(implSection, phrase) {
			t.Fatalf("unit_impl Independent Evaluation section missing phrase %q", phrase)
		}
	}
}

func TestIndependentEvaluationDoesNotClaimToolingProvesIsolationOrSemantics(t *testing.T) {
	repoRoot := findRepoRoot(t)
	files := activeDocFiles(t, repoRoot)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)tooling.{0,80}(proves|guarantees|ensures).{0,80}(reviewer|review).{0,80}(isolation|isolated|separate session)`),
		regexp.MustCompile(`(?is)tooling.{0,80}(proves|guarantees|ensures).{0,80}(semantic|meaning|business).{0,80}(correct|quality|decision)`),
		regexp.MustCompile(`(?is)snapshot validate-process.{0,80}(proves|guarantees|ensures).{0,80}(semantic|meaning|business)`),
		regexp.MustCompile(`(?is)receipt.{0,80}(proves|guarantees|ensures).{0,80}(reviewer|review).{0,80}(isolation|isolated|separate session)`),
	}

	for _, relPath := range files {
		content := readDocContractFile(t, repoRoot, relPath)
		for _, pattern := range patterns {
			if pattern.MatchString(content) {
				t.Fatalf("%s overclaims independent evaluation tooling boundary with pattern %q", relPath, pattern.String())
			}
		}
	}

	contract := readDocContractFile(t, repoRoot, "framework/process_snapshot_contract.md")
	for _, phrase := range []string{
		"Tooling validates the receipt fields mechanically",
		"it does not prove reviewer session isolation",
		"does not judge whether the reviewer made a good semantic decision",
	} {
		if !strings.Contains(contract, phrase) {
			t.Fatalf("process_snapshot_contract.md missing tooling boundary phrase %q", phrase)
		}
	}
}

func TestIndependentEvaluationDoesNotInventReviewerCli(t *testing.T) {
	repoRoot := findRepoRoot(t)
	files := activeDocFiles(t, repoRoot)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)specflowctl\s+reviewer\b`),
		regexp.MustCompile(`(?is)specflowctl\s+(independent-review|reviewer-review|review-session)\b`),
		regexp.MustCompile(`(?is)reviewer\s+(begin|submit|token)\b`),
		regexp.MustCompile(`(?is)harness\s+(begin|submit|token)\b`),
	}

	for _, relPath := range files {
		content := readDocContractFile(t, repoRoot, relPath)
		for _, pattern := range patterns {
			if pattern.MatchString(content) {
				t.Fatalf("%s invents an unsupported reviewer/harness CLI pattern %q", relPath, pattern.String())
			}
		}
	}
}

func TestGovernanceReviewDefaultsToScopedReview(t *testing.T) {
	repoRoot := findRepoRoot(t)
	contract := readDocContractFile(t, repoRoot, "framework/governance/review_scope.md")
	for _, phrase := range []string{
		"`scoped_review` is the default for ordinary framework changes",
		"`deep_audit` is explicit",
		"It must not use `_governance_review/` run-state by default.",
		"It must not require baseline slice table, dynamic slice table, score-state table, or full-scope run-state startup.",
		"Every real finding in a scoped review must be written as a short repairable story before trace details.",
		"Do not present a raw field dump as the user-facing finding.",
		"Every real finding in a scoped review must still include:",
		"`severity: P0|P1|P2|P3`",
		"`blocking: yes|no`",
		"Severity uses `framework/severity_policy.md`.",
		"Severity describes harm level; it does not replace explicit blocking status.",
		"Use it only for exact `spec_flow_review:full`.",
		"Use `deep_audit` only when the user explicitly requests exact `spec_flow_review:full` for mechanism correctness.",
		"scoped_pass | scoped_blocked | needs_deep_audit",
		"It must not claim full governance-baseline pass or full design-baseline pass.",
		"`scoped_blocked` means at least one in-scope finding has `blocking: yes`.",
	} {
		if !strings.Contains(contract, phrase) {
			t.Fatalf("review_scope.md missing scoped review contract phrase %q", phrase)
		}
	}

	frontDoor := readDocContractFile(t, repoRoot, "framework/governance/review.md")
	for _, phrase := range []string{
		"Ordinary `spec_flow_review` and governance review use `scoped_review` by default.",
		"Read `framework/governance/review_scope.md` first.",
		"`scoped_review` does not use `_governance_review/` run-state",
		"Plain exact `spec_flow_review` routes through this file first and remains scoped.",
		"Plain exact `spec_flow_design_review` routes through this file first, then directly delegates to `framework/spec_flow_design_review.md`.",
		"The only full-scope mechanism review entry is exact `spec_flow_review:full`.",
		"When the entry is exact `spec_flow_review:full`, `spec_flow_review` delegates to `framework/spec_flow_review.md`.",
		"`framework/spec_flow_review.md` is the mechanism deep-audit owner",
		"`framework/spec_flow_design_review.md` is the ordinary owner for every `spec_flow_design_review`.",
	} {
		if !strings.Contains(frontDoor, phrase) {
			t.Fatalf("governance/review.md missing scoped front-door phrase %q", phrase)
		}
	}

	for _, relPath := range []string{"README.md", "README.zh-CN.md"} {
		content := readDocContractFile(t, repoRoot, relPath)
		if !strings.Contains(content, "scoped review") {
			t.Fatalf("%s must describe scoped review as the routine governance review default", relPath)
		}
		if !strings.Contains(content, "spec_flow_review:full") {
			t.Fatalf("%s must name exact spec_flow_review:full as the full-scope mechanism review entry", relPath)
		}
		for _, forbidden := range []string{
			"ask for `full-scope`, `baseline`, `deep audit`, `resumable review`, or run-state-backed review",
			"明确说 `full-scope`、`baseline`、`deep audit`、`resumable review` 或 run-state-backed review",
		} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("%s still advertises natural-language full-scope mechanism triggers %q", relPath, forbidden)
			}
		}
	}
}

func TestSpecFlowReviewFindingsRequireNarrativeStory(t *testing.T) {
	repoRoot := findRepoRoot(t)
	review := readDocContractFile(t, repoRoot, "framework/spec_flow_review.md")
	findingContract := sectionBetween(t, review, "### 8.2 Narrative Finding Contract", "## 9. Non-Goals", "framework/spec_flow_review.md")
	for _, phrase := range []string{
		"one self-contained repairable story",
		"The first paragraph must use plain language and must be 4 to 6 sentences.",
		"who is executing the flow",
		"what the executor is trying to complete",
		"what the governing rule should make clear",
		"where the actual rule, handoff, or state path loses direction",
		"how the executor can take the wrong next step",
		"what the smallest correct repair point is",
		"Do not present a raw field dump as the user-facing finding.",
		"they do not satisfy the user-facing finding requirement by themselves.",
		"The first use of an internal term must explain the term in place.",
		"`Context Card`",
		"Every real finding must still contain these information items:",
		"required for every real finding and must be one of `P0`, `P1`, `P2`, or `P3`",
		"recommended fix",
		"evidence",
		"Recommended user-facing shape:",
		"Finding F-006: Natural-language unit requests can send the executor into the wrong lifecycle path.",
		"Status:",
		"Trace:",
	} {
		if !strings.Contains(findingContract, phrase) {
			t.Fatalf("spec_flow_review.md narrative finding contract missing %q", phrase)
		}
	}
	if strings.Contains(findingContract, "The minimum required fields are:") {
		t.Fatalf("spec_flow_review.md must not present findings as a raw minimum-field list")
	}

	scope := readDocContractFile(t, repoRoot, "framework/governance/review_scope.md")
	scoped := sectionBetween(t, scope, "## `scoped_review`", "## `deep_audit`", "framework/governance/review_scope.md")
	for _, phrase := range []string{
		"Every real finding in a scoped review must be written as a short repairable story before trace details.",
		"execution path, expected rule behavior, actual gap, possible wrong next step, and smallest correct repair",
		"Do not present a raw field dump as the user-facing finding.",
		"Every real finding in a scoped review must still include:",
		"`severity: P0|P1|P2|P3`",
		"`blocking: yes|no`",
		"`evidence`",
		"`recommended fix`",
		"Severity and blocking may appear in a status line after the story.",
		"Evidence and trace details should appear after the reader can already understand the problem.",
	} {
		if !strings.Contains(scoped, phrase) {
			t.Fatalf("review_scope.md narrative finding contract missing %q", phrase)
		}
	}
}

func TestGovernanceReviewDeepAuditRequiresExplicitIntent(t *testing.T) {
	repoRoot := findRepoRoot(t)
	content := readDocContractFile(t, repoRoot, "framework/spec_flow_review.md")
	for _, phrase := range []string{
		"This file owns explicit `deep_audit` review",
		"Ordinary or plain exact `spec_flow_review` entry routes through `framework/governance/review.md` first and stays `scoped_review`.",
		"The only full-scope mechanism review entry is exact `spec_flow_review:full`.",
		"Deep audit must use exact `spec_flow_review:full`.",
		"Plain exact entry must not automatically start full-scope run-state review.",
		"the state carrier for exact `spec_flow_review:full` is `docs/specs/_governance_review/spec_flow_review.md`",
		"ordinary scoped `spec_flow_review` does not use that carrier",
		"exact `spec_flow_review:full` must use the run-state file procedure in this section",
		"ordinary scoped `spec_flow_review` must use `framework/governance/review_scope.md` and must not use full-scope run state",
		"This section applies only to explicit `deep_audit`.",
		"This output contract applies to explicit `deep_audit`.",
		"replace the default `scoped_review` front door in `framework/governance/review.md`",
		"required for every real finding and must be one of `P0`, `P1`, `P2`, or `P3`",
		"`P0` and `P1` are normally blocking",
		"`P2` and `P3` are normally non-blocking",
	} {
		if !strings.Contains(content, phrase) {
			t.Fatalf("framework/spec_flow_review.md missing explicit deep-audit phrase %q", phrase)
		}
	}
	for _, forbidden := range []string{
		"resumable slice review",
		"narrowed reviews do not use",
		"a narrowed review may use a run-state file",
		"For narrowed review:",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("framework/spec_flow_review.md still contains obsolete narrowed/resumable phrase %q", forbidden)
		}
	}

	files := activeDocFiles(t, repoRoot)
	patterns := []*regexp.Regexp{
		regexp.MustCompile("(?is)Plain input `spec_flow_review` means the default governance-baseline review"),
		regexp.MustCompile("(?is)Plain exact entry.{0,120}automatically starts? full-scope"),
		regexp.MustCompile("(?is)plain `spec_flow_review`.{0,120}full-scope run-state"),
		regexp.MustCompile("(?is)spec_flow_review.{0,120}asks? for `full-scope`, `baseline`, `deep audit`"),
	}
	for _, relPath := range files {
		content := readDocContractFile(t, repoRoot, relPath)
		content = strings.ReplaceAll(content, "must not automatically start", "must-not-automatically-start")
		for _, pattern := range patterns {
			if pattern.MatchString(content) {
				t.Fatalf("%s reintroduces implicit full-scope governance review with pattern %q", relPath, pattern.String())
			}
		}
	}
}

func TestDesignReviewAlwaysUsesFullScopeBaseline(t *testing.T) {
	repoRoot := findRepoRoot(t)
	content := readDocContractFile(t, repoRoot, "framework/spec_flow_design_review.md")
	for _, phrase := range []string{
		"This file owns the only `spec_flow_design_review` mode: the default full-scope design-baseline review.",
		"Ordinary or plain exact `spec_flow_design_review` entry routes through `framework/governance/review.md` first, then enters this file.",
		"It must not be narrowed into `scoped_review`.",
		"Plain exact `spec_flow_design_review` starts the full-scope design-baseline review.",
		"This section applies to every `spec_flow_design_review`.",
		"Every `spec_flow_design_review` uses a run-state process file.",
		"This output contract applies to every `spec_flow_design_review`.",
		"whether full-scope run state was created, reused, or deleted and recreated",
		"the run-state file path",
		"the baseline slice table and slice statuses",
		"the dynamic risk slice table and slice statuses, or explicit `none`",
		"the score-state table",
		"the stale slice result",
		"entry_control_chain_check",
		"`startup_entry_control`",
		"`first_owner_selection`",
		"`owner_only_continuation`",
		"`pre_action_permission_gate`",
		"`route_specificity_before_implementation_gate`",
		"`diagnostic_work_not_mutation`",
		"chat-claimed lifecycle state",
		"skipped status or owner checks",
		"contract-like fields",
		"downstream compatibility",
		"`exact_command_precedence`",
		"`drift_stop_and_reroute`",
		"`no_ad_hoc_flow_substitution`",
		"`hard_stop_clarity`",
		"`owner_reachability`",
		"`entry_robustness_probe`",
		"tool-neutral",
		"independent executor",
		"`mixed_intent_prompt`",
		"`disguised_truth_change_prompt`",
		"`chat_claimed_state_prompt`",
		"`skip_owner_or_status_prompt`",
		"`custom_flow_substitution_prompt`",
		"`exact_command_with_noise_prompt`",
		"`clean_implementation_only_control_prompt`",
		"`prompt_family`",
		"`expected_control`",
		"`observed_first_owner`",
		"`diagnostic_allowed`",
		"`mutation_allowed`",
		"`probe_source`",
		"`executor_independence`",
		"`independent_agent_session`",
		"`reviewer_role_play`",
		"`recorded_replay_harness`",
		"`local_multi_executor_tool`",
		"`manual_black_box_exercise`",
		"`confirmed_independent_no_project_specific_context`",
		"`failure_class`",
		"`wrong_first_owner`",
		"`mutation_leak`",
		"`diagnostic_overblock`",
		"`chat_truth_trusted`",
		"`custom_flow_accepted`",
		"`exact_command_displaced`",
		"`implementation_gate_overmatch`",
		"the `entry_control_chain_check result`:",
		"report evidence for `startup_entry_control`, `first_owner_selection`, `owner_only_continuation`, `pre_action_permission_gate`, `route_specificity_before_implementation_gate`, `diagnostic_work_not_mutation`",
		"report probe evidence using `prompt_family`, `expected_control`, `observed_first_owner`, `diagnostic_allowed`, `mutation_allowed`, `result`, `failure_class`, `probe_source`, and `executor_independence`",
		"must be `passed`, `blocked`, or `incomplete`",
		"does not create another review flow, score question, score group, baseline slice, run-state field, or CLI",
		"must not depend on product, integration, vendor, or domain examples",
		"every code edit must change spec documents",
		"code-only or implementation-only work has a smaller legal path",
		"create a scoped or narrowed `spec_flow_design_review` mode",
		"required for every real finding and must be one of `P0`, `P1`, `P2`, or `P3`",
		"`P0` and `P1` are normally blocking",
		"`P2` and `P3` are normally non-blocking",
	} {
		if !strings.Contains(content, phrase) {
			t.Fatalf("framework/spec_flow_design_review.md missing full-scope design phrase %q", phrase)
		}
	}
	defaultScope := sectionBetween(t, content, "That default scope includes:", "The default scope excludes:", "framework/spec_flow_design_review.md")
	designFoundation := sectionBetween(t, content, "1. `design_foundation`", "2. `lifecycle_and_gate_design`", "framework/spec_flow_design_review.md")
	for _, section := range []struct {
		name    string
		content string
	}{
		{name: "default scope", content: defaultScope},
		{name: "design_foundation block", content: designFoundation},
	} {
		for _, phrase := range []string{
			"`governance/review.md`",
			"`governance/review_scope.md`",
		} {
			if !strings.Contains(section.content, phrase) {
				t.Fatalf("framework/spec_flow_design_review.md %s missing governance review input %q", section.name, phrase)
			}
		}
	}
	for _, forbidden := range []string{
		"stays `scoped_review`",
		"Plain exact entry without explicit deep-audit intent",
		"This section applies only to explicit `deep_audit`.",
		"This output contract applies to explicit `deep_audit`.",
		"narrowed reviews do not use",
		"replace the default `scoped_review` front door",
		"or not used",
		"when full-scope run state is used",
		"when run state is used",
		"`pre_solution_classification_report`",
		"`classification_next_action_table`",
		"`forbidden_action_visibility`",
		"`entry_applicability`",
		"`pre_mutation_gate`",
		"`authority_resolution`",
		"`drift_reclassification`",
		"spawn_agent",
		"multi_agent",
		"multi_agent_v1",
		"subagent",
		"provider",
		"adapter",
		"OpenAI",
		"Aliyun",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("framework/spec_flow_design_review.md contains forbidden design-review phrase %q", forbidden)
		}
	}
}

func TestScopedReviewDoesNotRequireRunStateOrSliceTables(t *testing.T) {
	repoRoot := findRepoRoot(t)
	content := readDocContractFile(t, repoRoot, "framework/governance/review_scope.md")
	scoped := sectionBetween(t, content, "## `scoped_review`", "## `deep_audit`", "framework/governance/review_scope.md")
	for _, forbidden := range []string{
		"must use `_governance_review/`",
		"requires `_governance_review/`",
		"must require baseline slice table",
		"must require dynamic slice table",
		"must require score-state table",
		"review run-init",
		"review run-refresh",
	} {
		if strings.Contains(scoped, forbidden) {
			t.Fatalf("scoped review contract must not require deep-audit machinery phrase %q", forbidden)
		}
	}
	for _, phrase := range []string{
		"input refs",
		"owner refs",
		"boundary refs",
		"checks performed",
		"findings",
		"conclusion",
	} {
		if !strings.Contains(scoped, phrase) {
			t.Fatalf("scoped review output contract missing %q", phrase)
		}
	}
}

func TestGovernanceReviewRunStatePolicy(t *testing.T) {
	repoRoot := findRepoRoot(t)
	content := readDocContractFile(t, repoRoot, "templates/docs/specs/_governance_review/README.md")
	for _, phrase := range []string{
		"process files for explicit full-scope mechanism reviews and every `spec_flow_design_review`",
		"Explicit deep-audit `spec_flow_review` uses:",
		"Every `spec_flow_design_review` uses:",
		"Scoped `spec_flow_review` and ordinary scoped governance reviews do not use full-scope run state by default.",
		"Full-scope review mechanical writes are maintained by `specflowctl review run-* --flow <review_flow> --layout auto|installed|source`",
		"When an explicit full-scope mechanism review or any `spec_flow_design_review` resumes a run-state file",
		"A full-scope review result must not claim a passing conclusion until every required baseline slice and dynamic slice is closed by the owning review policy.",
		"Starting a new full-scope review deletes the previous fixed file before the new run state is written.",
		"Each run-state file records `review_layout`.",
		"`source_repo` layout reviews template bootstrap compatibility under `templates/docs/specs/`",
	} {
		if !strings.Contains(content, phrase) {
			t.Fatalf("_governance_review README missing deep-audit run-state phrase %q", phrase)
		}
	}
	for _, forbidden := range []string{
		"resumable governance reviews",
		"When a deep-audit review resumes a run-state file",
		"A deep-audit review result must not claim a passing conclusion",
		"Starting a new deep-audit review deletes the previous fixed file",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("_governance_review README still contains deep-audit-only run-state phrase %q", forbidden)
		}
	}
}

func TestGovernanceReviewDoesNotInventNewReviewCli(t *testing.T) {
	repoRoot := findRepoRoot(t)
	files := activeDocFiles(t, repoRoot)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)specflowctl\s+review\s+scoped\b`),
		regexp.MustCompile(`(?is)specflowctl\s+review\s+deep-audit\b`),
		regexp.MustCompile(`(?is)specflowctl\s+review-scope\b`),
		regexp.MustCompile(`(?is)specflowctl\s+deep-audit\b`),
	}

	for _, relPath := range files {
		content := readDocContractFile(t, repoRoot, relPath)
		for _, pattern := range patterns {
			if pattern.MatchString(content) {
				t.Fatalf("%s invents an unsupported governance review CLI pattern %q", relPath, pattern.String())
			}
		}
	}
}

func TestSourceRepoEntryExampleRoutesGovernanceAndDesignReview(t *testing.T) {
	repoRoot := findRepoRoot(t)
	content := readDocContractFile(t, repoRoot, "example.md")
	for _, phrase := range []string{
		"# Source Repository Agent Instructions Example",
		"Personal preferences should stay in local `AGENTS.md`, `CLAUDE.md`, or `GEMINI.md` files",
		"# Governance Review Shortcut",
		"This is the specFlow source repository, so use local `framework/...` paths.",
		"For `spec_flow_review` or ordinary governance review requests:",
		"Read `framework/governance/review.md`.",
		"Read `framework/governance/review_scope.md`.",
		"Default to `scoped_review`.",
		"Use `framework/spec_flow_review.md`, `_governance_review/` run-state files, baseline slice tables, or dynamic slice tables only for exact `spec_flow_review:full`.",
		"For `spec_flow_design_review`:",
		"Read `framework/spec_flow_design_review.md`.",
		"Run the default full-scope design-baseline review. Do not narrow it to `scoped_review`.",
	} {
		if !strings.Contains(content, phrase) {
			t.Fatalf("example.md missing source-repo entry phrase %q", phrase)
		}
	}
	for _, forbidden := range []string{
		"specflow/framework/lifecycle/overview.md",
		"specflow/framework/operations/entry_routing.md",
		"specflow/framework/governance/review.md",
		"specflow/framework/governance/review_scope.md",
		"<!-- SPECFLOW:BEGIN -->",
		"## First Read",
		"## Authority Boundary",
		"## Active Surface",
		"spec_design_review",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("example.md should use source-repo local entry paths and no managed block; found %q", forbidden)
		}
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if fileExists(filepath.Join(dir, "framework", "lifecycle", "unit_check.md")) &&
			fileExists(filepath.Join(dir, "tooling", "go.mod")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate repo root from %s", dir)
		}
		dir = parent
	}
}

func TestRecoveryAndImpactSyncUseCanonicalFallbackReasons(t *testing.T) {
	repoRoot := findRepoRoot(t)
	files := []string{
		"framework/lifecycle/recovery.md",
		"framework/governance/impact_sync.md",
	}
	forbidden := []string{
		"`truth_changed`",
		"`candidate_truth_repaired`",
		"`rule_binding_changed`",
		"`repository_boundary_changed`",
		"`check_invalid`",
		"`check_missing`",
		"`check_scope_unclear`",
		"`plan_invalid`",
		"`plan_missing`",
		"`implementation_plan_drift`",
		"`implementation_incomplete`",
		"`implementation_blocked`",
		"`implementation_drift`",
		"`verify_invalid`",
		"`verify_missing`",
		"`evidence_drift`",
		"`stable_alignment_drift`",
	}
	for _, relPath := range files {
		content := readDocContractFile(t, repoRoot, relPath)
		for _, reason := range forbidden {
			if strings.Contains(content, reason) {
				t.Fatalf("%s must not use non-canonical fallback reason %s", relPath, reason)
			}
		}
	}

	recovery := readDocContractFile(t, repoRoot, "framework/lifecycle/recovery.md")
	for _, reason := range []string{
		"`truth_drift`",
		"`binding_drift`",
		"`baseline_drift`",
		"`rule_drift`",
		"`truth_incomplete`",
		"`plan_drift`",
		"`gate_missing`",
		"`implementation_deviation`",
		"`evidence_incomplete`",
		"`stable_verify_invalid`",
	} {
		if !strings.Contains(recovery, reason) {
			t.Fatalf("framework/lifecycle/recovery.md missing canonical fallback reason %s", reason)
		}
	}
}

func TestProcessSnapshotContractStoresOnlyAdvancingCheckAndVerifyEvidence(t *testing.T) {
	repoRoot := findRepoRoot(t)
	content := readDocContractFile(t, repoRoot, "framework/process_snapshot_contract.md")
	if strings.Contains(content, "decision: pass|blocked|fix_required") {
		t.Fatalf("process_snapshot_contract.md must not authorize non-pass check or verify evidence")
	}
	for _, phrase := range []string{
		"decision: pass",
		"allow_next: true",
		"`_check_result` and candidate `_verify_result` are consumable evidence only for advancing pass gates.",
		"Non-advancing command outcomes such as `blocked` or `fix_required` must not be stored as these process snapshots.",
	} {
		if !strings.Contains(content, phrase) {
			t.Fatalf("process_snapshot_contract.md missing pass-only process evidence phrase %q", phrase)
		}
	}
}

func activeDocFiles(t *testing.T, repoRoot string) []string {
	t.Helper()
	files := []string{"README.md", "README.zh-CN.md", "tooling/README.md"}
	for _, relDir := range []string{"framework", "templates"} {
		root := filepath.Join(repoRoot, relDir)
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".md") && entry.Name() != "SKILL.md" {
				return nil
			}
			rel, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(rel))
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", relDir, err)
		}
	}
	return files
}

func removedEntryPathStrings() []string {
	return []string{
		deletedFrameworkPolicyPath("command_", "policy.md"),
		deletedFrameworkPolicyPath("recovery_", "policy.md"),
		deletedFrameworkPolicyPath("impact_sync_", "policy.md"),
		deletedFrameworkPolicyPath("spec_flow_", "migrate.md"),
		deletedFrameworkRulePath("new"),
		deletedFrameworkRulePath("extract"),
		deletedFrameworkRulePath("bind"),
		deletedFrameworkRulePath("topology"),
		deletedFrameworkRulePath("sync"),
		deletedFrameworkRulePath("escape"),
		"framework/" + "command" + "s",
	}
}

func deletedFrameworkPolicyPath(prefix, suffix string) string {
	return "framework/" + prefix + suffix
}

func deletedFrameworkRulePath(name string) string {
	return "framework/" + "rule_" + name + ".md"
}

func readDocContractFile(t *testing.T, repoRoot, relPath string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relPath)))
	if err != nil {
		t.Fatalf("read %s: %v", relPath, err)
	}
	return string(content)
}

func managedSpecFlowBlock(t *testing.T, content, relPath string) string {
	t.Helper()
	start := strings.Index(content, "<!-- SPECFLOW:BEGIN -->")
	end := strings.Index(content, "<!-- SPECFLOW:END -->")
	if start < 0 || end < 0 || end <= start {
		t.Fatalf("%s missing managed specFlow block", relPath)
	}
	return content[start:end]
}

func sectionBetween(t *testing.T, content, start, end, relPath string) string {
	t.Helper()
	startIndex := strings.Index(content, start)
	if startIndex < 0 {
		t.Fatalf("%s missing section %q", relPath, start)
	}
	endIndex := strings.Index(content[startIndex+len(start):], end)
	if endIndex < 0 {
		t.Fatalf("%s missing section %q after %q", relPath, end, start)
	}
	endIndex += startIndex + len(start)
	return content[startIndex:endIndex]
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
