package reviewscope

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectDefaultSpecFlowScopeDoesNotRequireProjectStandardsRegistry(t *testing.T) {
	repoRoot := t.TempDir()
	writeLayeredFrameworkFiles(t, repoRoot)
	mustWrite(t, filepath.Join(repoRoot, "specflow/framework/candidate_intents/repair.md"), "# repair\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/framework/candidate_intents/change.md"), "# change\n")
	writeGuidanceSkillFiles(t, repoRoot)
	mustWrite(t, filepath.Join(repoRoot, "specflow/templates/docs/specs/_status.md"), "# status\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/templates/docs/specs/_check_result/README.md"), "# check\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/templates/docs/specs/_check_work/README.md"), "# check work\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/templates/docs/specs/_plans/README.md"), "# plans\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/templates/docs/specs/_plans/draft/README.md"), "# draft plans\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/templates/docs/specs/_plans/active/README.md"), "# active plans\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/templates/docs/specs/_verify_result/README.md"), "# verify\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/templates/docs/specs/_stable_verify_result/README.md"), "# stable verify\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/templates/docs/specs/_governance_review/README.md"), "# governance review\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/templates/docs/specs/repository_mapping.md"), "# repository mapping template\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/templates/docs/specs/rules/stable/s_g_rule_repository_baseline.md"), "# global rule template\n")
	mustWrite(t, filepath.Join(repoRoot, "docs/specs/_check_result/README.md"), "# project check\n")
	mustWrite(t, filepath.Join(repoRoot, "docs/specs/_plans/README.md"), "# project plans\n")
	mustWrite(t, filepath.Join(repoRoot, "docs/specs/_plans/draft/README.md"), "# project draft plans\n")
	mustWrite(t, filepath.Join(repoRoot, "docs/specs/_plans/active/README.md"), "# project active plans\n")
	mustWrite(t, filepath.Join(repoRoot, "docs/specs/_verify_result/README.md"), "# project verify\n")
	mustWrite(t, filepath.Join(repoRoot, "docs/specs/_stable_verify_result/README.md"), "# project stable verify\n")
	mustWrite(t, filepath.Join(repoRoot, "docs/specs/_governance_review/spec_flow_review.md"), "# ignored run state\n")
	mustWrite(t, filepath.Join(repoRoot, "docs/specs/units/candidate/c_unit_demo.md"), "# demo candidate\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/templates/AGENTS.md"), "host\n<!-- SPECFLOW:BEGIN -->\nmanaged\n<!-- SPECFLOW:END -->\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/templates/GEMINI.md"), "host\n<!-- SPECFLOW:BEGIN -->\nmanaged\n<!-- SPECFLOW:END -->\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/templates/CLAUDE.md"), "host\n<!-- SPECFLOW:BEGIN -->\nmanaged\n<!-- SPECFLOW:END -->\n")
	mustWrite(t, filepath.Join(repoRoot, "AGENTS.md"), "host\n<!-- SPECFLOW:BEGIN -->\nmanaged\n<!-- SPECFLOW:END -->\n")
	mustWrite(t, filepath.Join(repoRoot, "GEMINI.md"), "host\n<!-- SPECFLOW:BEGIN -->\nmanaged\n<!-- SPECFLOW:END -->\n")
	mustWrite(t, filepath.Join(repoRoot, "CLAUDE.md"), "host\n<!-- SPECFLOW:BEGIN -->\nmanaged\n<!-- SPECFLOW:END -->\n")
	mustWrite(t, filepath.Join(repoRoot, "docs/specs/_status.md"), "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n")
	mustWrite(t, filepath.Join(repoRoot, "docs/specs/repository_mapping.md"), "# Repository Mapping\n")
	mustWrite(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_g_rule_repository_baseline.md"), "# Global Rules\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/tooling/README.md"), "# tooling\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/tooling/cmd/specflowctl/main.go"), "package main\nfunc main() {}\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/tooling/internal/demo/demo.go"), "package demo\nfunc Value() string { return \"demo\" }\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/tooling/go.mod"), "module example.com/specflow\n\ngo 1.22\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/tooling/manifest.tsv"), "templates/AGENTS.md\tAGENTS.md\tframework\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/tooling/go.sum"), "example.com/mod v1.0.0 h1:demo\n")
	writeToolingScriptFiles(t, repoRoot)
	writeReaderWebFiles(t, repoRoot)

	scope, err := CollectDefaultSpecFlowScope(repoRoot)
	if err != nil {
		t.Fatalf("CollectDefaultSpecFlowScope: %v", err)
	}
	if !containsString(scope.AgentOperabilityFiles, "specflow/framework/agent_operability_standard.md") {
		t.Fatalf("expected agent operability file in scope, got %+v", scope.AgentOperabilityFiles)
	}
	for _, input := range []string{
		"specflow/framework/advance_policy.md",
		"specflow/framework/core/adoption_modes.md",
		"specflow/framework/core/context_card.md",
		"specflow/framework/core/freshness.md",
		"specflow/framework/core/independent_evaluation.md",
		"specflow/framework/governance/review_scope.md",
		"specflow/framework/spec_flow_review.md",
		"specflow/framework/spec_policy.md",
		"specflow/framework/spec_writing_guide.md",
	} {
		if !containsString(scope.AgentOperabilityFiles, input) {
			t.Fatalf("expected agent operability input %s, got %+v", input, scope.AgentOperabilityFiles)
		}
	}
	if containsString(scope.FrameworkGuidelineFiles, deletedCommandPolicyPath("specflow/framework")) {
		t.Fatalf("deleted flat command owner must stay outside default scope, got %+v", scope.FrameworkGuidelineFiles)
	}
	if !containsString(scope.FrameworkGuidelineFiles, "specflow/framework/spec_flow_review.md") {
		t.Fatalf("expected root review owner in default scope, got %+v", scope.FrameworkGuidelineFiles)
	}
	if !containsString(scope.AgentOperabilityFiles, "specflow/framework/lifecycle/unit_check.md") {
		t.Fatalf("expected lifecycle command files in agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if containsString(scope.AgentOperabilityFiles, deletedCommandFilePath("specflow/framework", "unit_check.md")) {
		t.Fatalf("deleted command files must stay outside active agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if !containsString(scope.CommandFiles, "specflow/framework/lifecycle/unit_plan.md") {
		t.Fatalf("expected lifecycle files in command scope, got %+v", scope.CommandFiles)
	}
	if !containsString(scope.GuidanceSkillFiles, "specflow/framework/skills/using-specflow-guidance/SKILL.md") {
		t.Fatalf("expected guidance skill files in scope, got %+v", scope.GuidanceSkillFiles)
	}
	if !containsString(scope.CandidateIntentFiles, "specflow/framework/candidate_intent_policy.md") {
		t.Fatalf("expected candidate intent policy in candidate intent scope, got %+v", scope.CandidateIntentFiles)
	}
	if !containsString(scope.CandidateIntentFiles, "specflow/framework/candidate_intents/repair.md") {
		t.Fatalf("expected repair intent standard in candidate intent scope, got %+v", scope.CandidateIntentFiles)
	}
	if !containsString(scope.CandidateIntentFiles, "specflow/framework/candidate_intents/change.md") {
		t.Fatalf("expected change intent standard in candidate intent scope, got %+v", scope.CandidateIntentFiles)
	}
	if !containsString(scope.AgentOperabilityFiles, "specflow/framework/candidate_intent_policy.md") {
		t.Fatalf("expected candidate intent policy in agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if !containsString(scope.AgentOperabilityFiles, "specflow/framework/candidate_intents/repair.md") {
		t.Fatalf("expected repair intent standard in agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if !containsString(scope.AgentOperabilityFiles, "specflow/framework/skills/using-specflow-guidance/SKILL.md") {
		t.Fatalf("expected guidance skill files in agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if containsString(scope.RuleGovernanceFiles, "specflow/framework/onboarding_decision_policy.md") {
		t.Fatalf("expected onboarding decision policy outside rule-governance scope, got %+v", scope.RuleGovernanceFiles)
	}
	if !containsString(scope.AgentOperabilityFiles, "specflow/framework/operations/migration.md") {
		t.Fatalf("expected migration policy in agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if !containsString(scope.AgentOperabilityFiles, "specflow/framework/operations/entry_routing.md") {
		t.Fatalf("expected layered entry routing in agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if !containsString(scope.AgentOperabilityFiles, "specflow/framework/governance/rule_system.md") {
		t.Fatalf("expected rule-governance files in agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if !containsString(scope.AgentOperabilityFiles, "specflow/framework/governance/rules/rule_sync.md") {
		t.Fatalf("expected rule flow files in agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if !containsString(scope.AgentOperabilityFiles, "specflow/framework/process_snapshot_contract.md") {
		t.Fatalf("expected process-state contract files in agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if !containsString(scope.AgentOperabilityFiles, "specflow/framework/tooling_execution_policy.md") {
		t.Fatalf("expected tooling execution policy in agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if !containsString(scope.AgentOperabilityFiles, "specflow/framework/slice_work_state_protocol.md") {
		t.Fatalf("expected slice work-state protocol in agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if !containsString(scope.ToolingContractFiles, "specflow/framework/slice_work_state_protocol.md") {
		t.Fatalf("expected slice work-state protocol in tooling contract scope, got %+v", scope.ToolingContractFiles)
	}
	if !containsString(scope.AgentOperabilityFiles, "specflow/tooling/README.md") {
		t.Fatalf("expected tooling README in agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if containsString(scope.AgentOperabilityFiles, "docs/specs/_verify_result/README.md") {
		t.Fatalf("expected project-side process files to stay outside agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if !containsString(scope.AgentOperabilityFiles, "specflow/templates/docs/specs/_stable_verify_result/README.md") {
		t.Fatalf("expected stable verify process contract in agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if !containsString(scope.AgentOperabilityFiles, "specflow/templates/docs/specs/_governance_review/README.md") {
		t.Fatalf("expected governance review process contract in agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if containsString(scope.AgentOperabilityFiles, "docs/specs/repository_mapping.md") {
		t.Fatalf("expected project repository truth to stay outside agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
	if !containsString(scope.ProjectInstanceCompatibilityFiles, "docs/specs/_status.md") {
		t.Fatalf("expected project status in compatibility scope, got %+v", scope.ProjectInstanceCompatibilityFiles)
	}
	if !containsString(scope.ProjectInstanceCompatibilityFiles, "docs/specs/repository_mapping.md") {
		t.Fatalf("expected project repository mapping in compatibility scope, got %+v", scope.ProjectInstanceCompatibilityFiles)
	}
	if !containsString(scope.ProjectInstanceCompatibilityFiles, "docs/specs/rules/stable/s_g_rule_repository_baseline.md") {
		t.Fatalf("expected project global rules in compatibility scope, got %+v", scope.ProjectInstanceCompatibilityFiles)
	}
	if !containsString(scope.TemplateProjectInstanceFiles, "specflow/templates/docs/specs/repository_mapping.md") {
		t.Fatalf("expected repository mapping template in template project-instance scope, got %+v", scope.TemplateProjectInstanceFiles)
	}
	if !containsString(scope.TemplateProjectInstanceFiles, "specflow/templates/docs/specs/rules/stable/s_g_rule_repository_baseline.md") {
		t.Fatalf("expected global rule template in template project-instance scope, got %+v", scope.TemplateProjectInstanceFiles)
	}
	if !containsString(scope.ProjectInstanceCompatibilityFiles, "docs/specs/units/candidate/c_unit_demo.md") {
		t.Fatalf("expected project truth file shape input in compatibility scope, got %+v", scope.ProjectInstanceCompatibilityFiles)
	}
	if containsString(scope.ProjectInstanceCompatibilityFiles, "docs/specs/_governance_review/spec_flow_review.md") {
		t.Fatalf("expected active governance review run state outside compatibility scope, got %+v", scope.ProjectInstanceCompatibilityFiles)
	}
	if !containsString(scope.ProjectEntryFiles, "AGENTS.md") {
		t.Fatalf("expected project entry files in scope, got %+v", scope.ProjectEntryFiles)
	}
	if len(scope.SourceRepoEntryExampleFiles) != 0 {
		t.Fatalf("installed project scope must not include source repo entry examples, got %+v", scope.SourceRepoEntryExampleFiles)
	}
	if !containsString(scope.ToolingSourceFiles, "specflow/tooling/go.mod") {
		t.Fatalf("expected tooling go.mod in tooling source scope, got %+v", scope.ToolingSourceFiles)
	}
	if !containsString(scope.ToolingSourceFiles, "specflow/tooling/manifest.tsv") {
		t.Fatalf("expected tooling manifest in tooling source scope, got %+v", scope.ToolingSourceFiles)
	}
	if !containsString(scope.ToolingSourceFiles, "specflow/tooling/go.sum") {
		t.Fatalf("expected tooling go.sum in tooling source scope when present, got %+v", scope.ToolingSourceFiles)
	}
	if !containsString(scope.ToolingScriptFiles, "specflow/tooling/scripts/tooling_fingerprint.sh") {
		t.Fatalf("expected shell fingerprint script in tooling script scope, got %+v", scope.ToolingScriptFiles)
	}
	if !containsString(scope.ToolingScriptFiles, "specflow/tooling/scripts/tooling_fingerprint.ps1") {
		t.Fatalf("expected PowerShell fingerprint script in tooling script scope, got %+v", scope.ToolingScriptFiles)
	}
	if !containsString(scope.ToolingScriptFiles, "specflow/tooling/scripts/build_release.sh") {
		t.Fatalf("expected build release script in tooling script scope, got %+v", scope.ToolingScriptFiles)
	}
	for _, relPath := range currentToolingScriptFiles() {
		if !containsString(scope.ToolingScriptFiles, relPath) {
			t.Fatalf("expected current tooling script in tooling script scope: %s, got %+v", relPath, scope.ToolingScriptFiles)
		}
	}
	if containsString(scope.ToolingSourceFiles, "specflow/tooling/scripts/tooling_fingerprint.sh") {
		t.Fatalf("expected shell fingerprint script outside tooling source scope, got %+v", scope.ToolingSourceFiles)
	}
	if containsString(scope.ToolingSourceFiles, "specflow/tooling/scripts/build_release.sh") {
		t.Fatalf("expected build release script outside tooling source scope, got %+v", scope.ToolingSourceFiles)
	}
	if !containsString(scope.ToolingRuntimeFiles, "specflow/tooling/reader/web/app.js") {
		t.Fatalf("expected reader app.js in tooling runtime scope, got %+v", scope.ToolingRuntimeFiles)
	}
	if containsString(scope.ToolingSourceFiles, "specflow/tooling/reader/web/app.js") {
		t.Fatalf("expected reader runtime file outside tooling source scope, got %+v", scope.ToolingSourceFiles)
	}
}

func TestCollectDefaultSpecFlowDesignScopeIncludesGovernanceReviewProcessContract(t *testing.T) {
	repoRoot := t.TempDir()
	writeLayeredFrameworkFiles(t, repoRoot)
	for _, relPath := range []string{
		"specflow/framework/candidate_intents/repair.md",
		"specflow/framework/candidate_intents/change.md",
		"specflow/templates/docs/specs/_status.md",
		"specflow/templates/docs/specs/_check_result/README.md",
		"specflow/templates/docs/specs/_check_work/README.md",
		"specflow/templates/docs/specs/_plans/README.md",
		"specflow/templates/docs/specs/_plans/draft/README.md",
		"specflow/templates/docs/specs/_plans/active/README.md",
		"specflow/templates/docs/specs/_verify_result/README.md",
		"specflow/templates/docs/specs/_stable_verify_result/README.md",
		"specflow/templates/docs/specs/_governance_review/README.md",
		"specflow/templates/AGENTS.md",
		"specflow/templates/GEMINI.md",
		"specflow/templates/CLAUDE.md",
		"AGENTS.md",
		"GEMINI.md",
		"CLAUDE.md",
	} {
		mustWrite(t, filepath.Join(repoRoot, relPath), "# "+filepath.Base(relPath)+"\n")
	}
	mustWrite(t, filepath.Join(repoRoot, "docs/specs/_status.md"), ""+
		"# Spec Status\n\n"+
		"## Formal Objects\n\n"+
		"| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n"+
		"|---|---|---|---|---|---|---|\n")

	scope, err := CollectDefaultSpecFlowDesignScope(repoRoot)
	if err != nil {
		t.Fatalf("CollectDefaultSpecFlowDesignScope: %v", err)
	}
	if !containsString(scope.TemplateGovernanceFiles, "specflow/templates/docs/specs/_governance_review/README.md") {
		t.Fatalf("expected governance review process contract in design scope, got %+v", scope.TemplateGovernanceFiles)
	}
	if !containsString(scope.TemplateGovernanceFiles, "specflow/templates/docs/specs/_stable_verify_result/README.md") {
		t.Fatalf("expected stable verify process contract in design scope, got %+v", scope.TemplateGovernanceFiles)
	}
	if !containsString(scope.FrameworkGuidelineFiles, "specflow/framework/operations/migration.md") {
		t.Fatalf("expected migration policy in design foundation scope, got %+v", scope.FrameworkGuidelineFiles)
	}
	if containsString(scope.FrameworkGuidelineFiles, deletedCommandPolicyPath("specflow/framework")) {
		t.Fatalf("deleted flat command owner must stay outside design foundation scope, got %+v", scope.FrameworkGuidelineFiles)
	}
	if !containsString(scope.FrameworkGuidelineFiles, "specflow/framework/lifecycle/overview.md") {
		t.Fatalf("expected lifecycle overview in design foundation scope, got %+v", scope.FrameworkGuidelineFiles)
	}
	if !containsString(scope.CandidateIntentFiles, "specflow/framework/candidate_intent_policy.md") {
		t.Fatalf("expected candidate intent policy in design candidate intent scope, got %+v", scope.CandidateIntentFiles)
	}
	if !containsString(scope.CandidateIntentFiles, "specflow/framework/candidate_intents/repair.md") {
		t.Fatalf("expected repair candidate intent in design candidate intent scope, got %+v", scope.CandidateIntentFiles)
	}
	if !containsString(scope.FrameworkGuidelineFiles, "specflow/framework/candidate_intents/change.md") {
		t.Fatalf("expected change candidate intent in design foundation scope, got %+v", scope.FrameworkGuidelineFiles)
	}
	if !containsString(scope.FrameworkGuidelineFiles, "specflow/framework/operations/output_standard.md") {
		t.Fatalf("expected output standard in design foundation scope, got %+v", scope.FrameworkGuidelineFiles)
	}
	if !containsString(scope.FrameworkGuidelineFiles, "specflow/framework/slice_work_state_protocol.md") {
		t.Fatalf("expected slice work-state protocol in design foundation scope, got %+v", scope.FrameworkGuidelineFiles)
	}
	if !containsString(scope.TemplateGovernanceFiles, "specflow/framework/slice_work_state_protocol.md") {
		t.Fatalf("expected slice work-state protocol in lifecycle contract scope, got %+v", scope.TemplateGovernanceFiles)
	}
}

func TestCollectDefaultSpecFlowDesignScopeRequiresCandidateIntentStandards(t *testing.T) {
	repoRoot := t.TempDir()
	writeLayeredFrameworkFiles(t, repoRoot)
	for _, relPath := range []string{
		"specflow/templates/docs/specs/_status.md",
		"specflow/templates/docs/specs/_check_work/README.md",
		"specflow/templates/docs/specs/_check_result/README.md",
		"specflow/templates/docs/specs/_plans/README.md",
		"specflow/templates/docs/specs/_plans/draft/README.md",
		"specflow/templates/docs/specs/_plans/active/README.md",
		"specflow/templates/docs/specs/_verify_result/README.md",
		"specflow/templates/docs/specs/_stable_verify_result/README.md",
		"specflow/templates/docs/specs/_governance_review/README.md",
		"specflow/templates/AGENTS.md",
		"specflow/templates/GEMINI.md",
		"specflow/templates/CLAUDE.md",
		"AGENTS.md",
		"GEMINI.md",
		"CLAUDE.md",
	} {
		mustWrite(t, filepath.Join(repoRoot, relPath), "# "+filepath.Base(relPath)+"\n")
	}

	_, err := CollectDefaultSpecFlowDesignScope(repoRoot)
	if err == nil || !strings.Contains(err.Error(), "default design candidate intent files are incomplete") {
		t.Fatalf("expected missing candidate intent standards error, got %v", err)
	}
}

func TestCollectDefaultSpecFlowScopeSupportsSourceRepoLayout(t *testing.T) {
	repoRoot := t.TempDir()
	writeSourceScopeRepo(t, repoRoot)

	scope, err := CollectDefaultSpecFlowScopeForLayout(repoRoot, "source")
	if err != nil {
		t.Fatalf("CollectDefaultSpecFlowScopeForLayout source: %v", err)
	}
	if scope.Layout != LayoutSourceRepo {
		t.Fatalf("expected source_repo layout, got %s", scope.Layout)
	}
	if scope.FrameworkRoot != "framework" || scope.TemplateRoot != "templates" || scope.ToolingRoot != "tooling" {
		t.Fatalf("unexpected source roots: %+v", scope)
	}
	if scope.ProjectInstanceCompatibilityMode != CompatibilityTemplateBootstrap {
		t.Fatalf("expected template compatibility mode, got %s", scope.ProjectInstanceCompatibilityMode)
	}
	if !containsString(scope.CommandFiles, "framework/lifecycle/unit_check.md") {
		t.Fatalf("expected local lifecycle file in source scope, got %+v", scope.CommandFiles)
	}
	if containsString(scope.FrameworkGuidelineFiles, deletedCommandPolicyPath("framework")) {
		t.Fatalf("deleted flat command owner must stay outside source scope, got %+v", scope.FrameworkGuidelineFiles)
	}
	if containsString(scope.CommandFiles, "specflow/framework/lifecycle/unit_check.md") {
		t.Fatalf("expected source scope to avoid installed paths, got %+v", scope.CommandFiles)
	}
	if !containsString(scope.ProjectInstanceCompatibilityFiles, "templates/docs/specs/_status.md") {
		t.Fatalf("expected template status compatibility input, got %+v", scope.ProjectInstanceCompatibilityFiles)
	}
	if containsString(scope.ProjectInstanceCompatibilityFiles, "docs/specs/_status.md") {
		t.Fatalf("source compatibility must not require project docs/specs, got %+v", scope.ProjectInstanceCompatibilityFiles)
	}
	if !containsString(scope.ToolingSourceFiles, "tooling/go.mod") {
		t.Fatalf("expected local tooling source, got %+v", scope.ToolingSourceFiles)
	}
	if len(scope.ProjectEntryFiles) != 0 {
		t.Fatalf("source scope must not require local ignored project entry files, got %+v", scope.ProjectEntryFiles)
	}
	if !containsString(scope.SourceRepoEntryExampleFiles, "example.md") {
		t.Fatalf("expected source repo entry example in source scope, got %+v", scope.SourceRepoEntryExampleFiles)
	}
	if !containsString(scope.AgentOperabilityFiles, "example.md") {
		t.Fatalf("expected source repo entry example in agent operability scope, got %+v", scope.AgentOperabilityFiles)
	}
}

func TestCollectDefaultSpecFlowDesignScopeAutoDetectsSourceRepoLayout(t *testing.T) {
	repoRoot := t.TempDir()
	writeSourceScopeRepo(t, repoRoot)

	scope, err := CollectDefaultSpecFlowDesignScope(repoRoot)
	if err != nil {
		t.Fatalf("CollectDefaultSpecFlowDesignScope source auto: %v", err)
	}
	if scope.Layout != LayoutSourceRepo {
		t.Fatalf("expected auto source layout, got %s", scope.Layout)
	}
	if !containsString(scope.FrameworkGuidelineFiles, "framework/operations/output_standard.md") {
		t.Fatalf("expected source output standard in design scope, got %+v", scope.FrameworkGuidelineFiles)
	}
	if containsString(scope.FrameworkGuidelineFiles, deletedCommandPolicyPath("framework")) {
		t.Fatalf("deleted flat command owner must stay outside source design scope, got %+v", scope.FrameworkGuidelineFiles)
	}
	if len(scope.ProjectEntryFiles) != 0 {
		t.Fatalf("source design scope must not require local ignored project entry files, got %+v", scope.ProjectEntryFiles)
	}
	if !containsString(scope.SourceRepoEntryExampleFiles, "example.md") {
		t.Fatalf("expected source repo entry example in source design scope, got %+v", scope.SourceRepoEntryExampleFiles)
	}
}

func TestCollectDefaultSpecFlowScopeAutoRejectsAmbiguousLayout(t *testing.T) {
	repoRoot := t.TempDir()
	writeSourceScopeRepo(t, repoRoot)
	for _, relDir := range []string{
		"specflow/framework/core",
		"specflow/framework/lifecycle",
		"specflow/framework/governance",
		"specflow/framework/operations",
		"specflow/templates",
		"specflow/tooling",
	} {
		if err := os.MkdirAll(filepath.Join(repoRoot, relDir), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", relDir, err)
		}
	}

	_, err := CollectDefaultSpecFlowScope(repoRoot)
	if err == nil || !strings.Contains(err.Error(), "ambiguous review layout") {
		t.Fatalf("expected ambiguous layout error, got %v", err)
	}
}

func writeLayeredFrameworkFiles(t *testing.T, repoRoot string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(repoRoot, "specflow/tooling"), 0o755); err != nil {
		t.Fatalf("mkdir tooling marker: %v", err)
	}
	for _, relPath := range []string{
		"specflow/framework/advance_policy.md",
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
		"specflow/framework/lifecycle/unit_plan.md",
		"specflow/framework/lifecycle/unit_impl.md",
		"specflow/framework/lifecycle/unit_verify.md",
		"specflow/framework/lifecycle/unit_promote.md",
		"specflow/framework/lifecycle/unit_stable_verify.md",
		"specflow/framework/lifecycle/recovery.md",
		"specflow/framework/governance/rule_system.md",
		"specflow/framework/governance/impact_sync.md",
		"specflow/framework/governance/review.md",
		"specflow/framework/governance/review_scope.md",
		"specflow/framework/governance/rules/rule_bind.md",
		"specflow/framework/governance/rules/rule_escape.md",
		"specflow/framework/governance/rules/rule_extract.md",
		"specflow/framework/governance/rules/rule_new.md",
		"specflow/framework/governance/rules/rule_sync.md",
		"specflow/framework/governance/rules/rule_topology.md",
		"specflow/framework/operations/entry_routing.md",
		"specflow/framework/operations/implementation_change.md",
		"specflow/framework/operations/output_standard.md",
		"specflow/framework/operations/migration.md",
		"specflow/framework/severity_policy.md",
		"specflow/framework/spec_authoring_baseline.md",
		"specflow/framework/spec_flow_design_review.md",
		"specflow/framework/spec_flow_review.md",
		"specflow/framework/spec_policy.md",
		"specflow/framework/spec_writing_guide.md",
		"specflow/framework/agent_operability_standard.md",
		"specflow/framework/onboarding_decision_policy.md",
		"specflow/framework/candidate_intent_policy.md",
		"specflow/framework/candidate_handoff_contract.md",
		"specflow/framework/downgrade_policy.md",
		"specflow/framework/process_snapshot_contract.md",
		"specflow/framework/slice_work_state_protocol.md",
		"specflow/framework/tooling_execution_policy.md",
		"specflow/framework/entry_index_registry.md",
	} {
		mustWrite(t, filepath.Join(repoRoot, relPath), "# "+filepath.Base(relPath)+"\n")
	}
}

func writeSourceScopeRepo(t *testing.T, repoRoot string) {
	t.Helper()
	for _, relPath := range []string{
		"framework/advance_policy.md",
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
		"framework/lifecycle/unit_plan.md",
		"framework/lifecycle/unit_impl.md",
		"framework/lifecycle/unit_verify.md",
		"framework/lifecycle/unit_promote.md",
		"framework/lifecycle/unit_stable_verify.md",
		"framework/lifecycle/recovery.md",
		"framework/governance/rule_system.md",
		"framework/governance/impact_sync.md",
		"framework/governance/review.md",
		"framework/governance/review_scope.md",
		"framework/governance/rules/rule_bind.md",
		"framework/governance/rules/rule_escape.md",
		"framework/governance/rules/rule_extract.md",
		"framework/governance/rules/rule_new.md",
		"framework/governance/rules/rule_sync.md",
		"framework/governance/rules/rule_topology.md",
		"framework/operations/entry_routing.md",
		"framework/operations/implementation_change.md",
		"framework/operations/output_standard.md",
		"framework/operations/migration.md",
		"framework/spec_flow_review.md",
		"framework/spec_flow_design_review.md",
		"framework/spec_policy.md",
		"framework/spec_writing_guide.md",
		"framework/spec_authoring_baseline.md",
		"framework/agent_operability_standard.md",
		"framework/onboarding_decision_policy.md",
		"framework/candidate_intent_policy.md",
		"framework/candidate_handoff_contract.md",
		"framework/downgrade_policy.md",
		"framework/process_snapshot_contract.md",
		"framework/slice_work_state_protocol.md",
		"framework/tooling_execution_policy.md",
		"framework/entry_index_registry.md",
		"framework/severity_policy.md",
		"framework/candidate_intents/repair.md",
		"framework/candidate_intents/change.md",
		"framework/skills/using-specflow-guidance/SKILL.md",
		"framework/skills/project-framing/SKILL.md",
		"framework/skills/scope-cutting/SKILL.md",
		"framework/skills/solution-design/SKILL.md",
		"framework/skills/design-quality-review/SKILL.md",
		"framework/skills/spec-writeback-guidance/SKILL.md",
		"templates/docs/specs/_status.md",
		"templates/docs/specs/_check_result/README.md",
		"templates/docs/specs/_check_work/README.md",
		"templates/docs/specs/_plans/README.md",
		"templates/docs/specs/_plans/draft/README.md",
		"templates/docs/specs/_plans/active/README.md",
		"templates/docs/specs/_verify_result/README.md",
		"templates/docs/specs/_stable_verify_result/README.md",
		"templates/docs/specs/_governance_review/README.md",
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
		mustWrite(t, filepath.Join(repoRoot, relPath), "# "+filepath.Base(relPath)+"\n")
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
		mustWrite(t, filepath.Join(repoRoot, relPath), "# script\n")
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func writeGuidanceSkillFiles(t *testing.T, repoRoot string) {
	t.Helper()
	for _, relPath := range []string{
		"specflow/framework/skills/using-specflow-guidance/SKILL.md",
		"specflow/framework/skills/project-framing/SKILL.md",
		"specflow/framework/skills/scope-cutting/SKILL.md",
		"specflow/framework/skills/solution-design/SKILL.md",
		"specflow/framework/skills/design-quality-review/SKILL.md",
		"specflow/framework/skills/spec-writeback-guidance/SKILL.md",
	} {
		mustWrite(t, filepath.Join(repoRoot, relPath), "# skill\n")
	}
}

func writeReaderWebFiles(t *testing.T, repoRoot string) {
	t.Helper()
	mustWrite(t, filepath.Join(repoRoot, "specflow/tooling/reader/web/index.html"), "<!doctype html>\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/tooling/reader/web/styles.css"), "body { color: #111; }\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/tooling/reader/web/app.js"), "console.log('demo');\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/tooling/reader/web/cytoscape.min.js"), "window.cytoscape = function() {};\n")
	mustWrite(t, filepath.Join(repoRoot, "specflow/tooling/reader/web/mermaid.min.js"), "window.mermaid = { initialize() {}, run() {} };\n")
}

func writeToolingScriptFiles(t *testing.T, repoRoot string) {
	t.Helper()
	for _, relPath := range currentToolingScriptFiles() {
		mustWrite(t, filepath.Join(repoRoot, relPath), "# script\n")
	}
}

func currentToolingScriptFiles() []string {
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

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func deletedCommandPolicyPath(root string) string {
	return root + "/" + "command_" + "policy.md"
}

func deletedCommandFilePath(root, name string) string {
	return root + "/" + "command" + "s/" + name
}
