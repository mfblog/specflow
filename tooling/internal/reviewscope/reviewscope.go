package reviewscope

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	LayoutAuto             = "auto"
	LayoutInstalledProject = "installed_project"
	LayoutSourceRepo       = "source_repo"

	CompatibilityInstalledProject  = "project_instance"
	CompatibilityTemplateBootstrap = "template_bootstrap"
)

type SpecFlowScope struct {
	Profile                           string
	Layout                            string
	FrameworkRoot                     string
	TemplateRoot                      string
	ToolingRoot                       string
	ProjectInstanceCompatibilityMode  string
	FrameworkGuidelineFiles           []string
	CommandFiles                      []string
	CandidateIntentFiles              []string
	GuidanceSkillFiles                []string
	RuleGovernanceFiles               []string
	TemplateGovernanceFiles           []string
	TemplateProjectInstanceFiles      []string
	TemplateEntryFiles                []string
	ProjectEntryFiles                 []string
	SourceRepoEntryExampleFiles       []string
	AgentOperabilityFiles             []string
	ProjectInstanceCompatibilityFiles []string
	ToolingContractFiles              []string
	ToolingSourceFiles                []string
	ToolingScriptFiles                []string
	ToolingRuntimeFiles               []string
}

type layoutRoots struct {
	Layout                           string
	FrameworkRoot                    string
	TemplateRoot                     string
	ToolingRoot                      string
	ProjectInstanceCompatibilityMode string
}

func NormalizeLayout(value string) (string, error) {
	switch strings.TrimSpace(value) {
	case "", LayoutAuto:
		return LayoutAuto, nil
	case "installed", LayoutInstalledProject:
		return LayoutInstalledProject, nil
	case "source", LayoutSourceRepo:
		return LayoutSourceRepo, nil
	default:
		return "", fmt.Errorf("unsupported review layout %q", value)
	}
}

func ResolveLayout(repoRoot, requested string) (string, error) {
	normalized, err := NormalizeLayout(requested)
	if err != nil {
		return "", err
	}
	if normalized != LayoutAuto {
		if !layoutMarkersPresent(repoRoot, normalized) {
			return "", fmt.Errorf("requested review layout %s is not present under repository root", normalized)
		}
		return normalized, nil
	}

	sourcePresent := layoutMarkersPresent(repoRoot, LayoutSourceRepo)
	installedPresent := layoutMarkersPresent(repoRoot, LayoutInstalledProject)
	switch {
	case sourcePresent && installedPresent:
		return "", fmt.Errorf("ambiguous review layout: both source_repo and installed_project markers are present; pass --layout source or --layout installed")
	case sourcePresent:
		return LayoutSourceRepo, nil
	case installedPresent:
		return LayoutInstalledProject, nil
	default:
		return "", fmt.Errorf("could not detect review layout: expected source_repo markers or installed_project markers")
	}
}

func CollectDefaultSpecFlowScope(repoRoot string) (SpecFlowScope, error) {
	return CollectDefaultSpecFlowScopeForLayout(repoRoot, LayoutAuto)
}

func CollectDefaultSpecFlowScopeForLayout(repoRoot, requestedLayout string) (SpecFlowScope, error) {
	roots, err := resolveRoots(repoRoot, requestedLayout)
	if err != nil {
		return SpecFlowScope{}, err
	}
	scope := SpecFlowScope{
		Profile:                          "default_governance_baseline",
		Layout:                           roots.Layout,
		FrameworkRoot:                    roots.FrameworkRoot,
		TemplateRoot:                     roots.TemplateRoot,
		ToolingRoot:                      roots.ToolingRoot,
		ProjectInstanceCompatibilityMode: roots.ProjectInstanceCompatibilityMode,
	}

	frameworkFiles, err := layeredFrameworkFiles(repoRoot, roots)
	if err != nil {
		return scope, err
	}
	commandFiles, err := globRelative(repoRoot, joinPath(roots.FrameworkRoot, "lifecycle/*.md"))
	if err != nil {
		return scope, err
	}
	candidateIntentStandardFiles, err := globRelative(repoRoot, joinPath(roots.FrameworkRoot, "candidate_intents/*.md"))
	if err != nil {
		return scope, err
	}
	candidateIntentFiles := sortAndDedupe(append([]string{scope.FrameworkPath("candidate_intent_policy.md")}, candidateIntentStandardFiles...))
	guidanceSkillFiles, err := globRelative(repoRoot, joinPath(roots.FrameworkRoot, "skills/*/SKILL.md"))
	if err != nil {
		return scope, err
	}
	ruleFlowFiles, err := globRelative(repoRoot, joinPath(roots.FrameworkRoot, "governance/rules/*.md"))
	if err != nil {
		return scope, err
	}
	if len(frameworkFiles) == 0 || len(commandFiles) == 0 || len(candidateIntentStandardFiles) == 0 || len(guidanceSkillFiles) == 0 || len(ruleFlowFiles) == 0 {
		return scope, fmt.Errorf("default governance files are incomplete")
	}

	minimumGuidanceSkillFiles := []string{
		scope.FrameworkPath("skills/using-specflow-guidance/SKILL.md"),
		scope.FrameworkPath("skills/project-framing/SKILL.md"),
		scope.FrameworkPath("skills/scope-cutting/SKILL.md"),
		scope.FrameworkPath("skills/solution-design/SKILL.md"),
		scope.FrameworkPath("skills/design-quality-review/SKILL.md"),
		scope.FrameworkPath("skills/spec-writeback-guidance/SKILL.md"),
	}
	ruleFiles := []string{
		scope.FrameworkPath("governance/rule_system.md"),
		scope.FrameworkPath("governance/impact_sync.md"),
	}
	ruleFiles = sortAndDedupe(append(ruleFiles, ruleFlowFiles...))
	templateProcessStateFiles := templateProcessStateFiles(scope)
	templateGovernanceFiles := append([]string{}, templateProcessStateFiles...)
	templateProjectInstanceFiles := []string{
		scope.TemplatePath("docs/specs/repository_mapping.md"),
		scope.TemplatePath("docs/specs/rules/stable/s_g_rule_repository_baseline.md"),
	}
	templateEntryFiles := templateEntryFiles(scope)
	projectEntryFiles := projectEntryFiles(roots.Layout)
	sourceRepoEntryExampleFiles := sourceRepoEntryExampleFiles(roots.Layout)
	toolingContractFiles := []string{
		scope.FrameworkPath("tooling_execution_policy.md"),
		scope.FrameworkPath("slice_work_state_protocol.md"),
		scope.ToolingPath("README.md"),
	}
	processStateContractFiles := []string{
		scope.FrameworkPath("candidate_handoff_contract.md"),
		scope.FrameworkPath("downgrade_policy.md"),
		scope.FrameworkPath("process_snapshot_contract.md"),
		scope.FrameworkPath("slice_work_state_protocol.md"),
		scope.FrameworkPath("lifecycle/recovery.md"),
	}
	agentOperabilityFiles := collectAgentOperabilityFiles(scope, projectEntryFiles, sourceRepoEntryExampleFiles, templateEntryFiles, templateProcessStateFiles, commandFiles, candidateIntentFiles, guidanceSkillFiles, ruleFiles, processStateContractFiles, toolingContractFiles)
	projectInstanceCompatibilityFiles, err := collectProjectInstanceCompatibilityFiles(repoRoot, scope)
	if err != nil {
		return scope, err
	}

	toolingCmdFiles, err := walkRelativeFiles(repoRoot, scope.ToolingPath("cmd"), ".go")
	if err != nil {
		return scope, err
	}
	toolingInternalFiles, err := walkRelativeFiles(repoRoot, scope.ToolingPath("internal"), ".go")
	if err != nil {
		return scope, err
	}
	toolingSourceFiles := append([]string{}, toolingCmdFiles...)
	toolingSourceFiles = append(toolingSourceFiles, toolingInternalFiles...)
	toolingSourceFiles = append(toolingSourceFiles, scope.ToolingPath("go.mod"))
	toolingSourceFiles = append(toolingSourceFiles, scope.ToolingPath("manifest.tsv"))
	if fileExists(repoRoot, scope.ToolingPath("go.sum")) {
		toolingSourceFiles = append(toolingSourceFiles, scope.ToolingPath("go.sum"))
	}
	if len(toolingSourceFiles) == 0 {
		return scope, fmt.Errorf("default tooling source files are incomplete")
	}
	toolingScriptFiles, err := walkRelativeFiles(repoRoot, scope.ToolingPath("scripts"), "")
	if err != nil {
		return scope, err
	}
	if len(toolingScriptFiles) == 0 {
		return scope, fmt.Errorf("default tooling script files are incomplete")
	}
	toolingRuntimeFiles, err := walkRelativeFiles(repoRoot, scope.ToolingPath("reader/web"), "")
	if err != nil {
		return scope, err
	}
	if len(toolingRuntimeFiles) == 0 {
		return scope, fmt.Errorf("default tooling runtime files are incomplete")
	}

	required := append([]string{}, ruleFiles...)
	required = append(required, candidateIntentFiles...)
	required = append(required, minimumGuidanceSkillFiles...)
	required = append(required, templateGovernanceFiles...)
	required = append(required, templateProjectInstanceFiles...)
	required = append(required, templateEntryFiles...)
	required = append(required, projectEntryFiles...)
	required = append(required, sourceRepoEntryExampleFiles...)
	required = append(required, agentOperabilityFiles...)
	required = append(required, toolingContractFiles...)
	required = append(required, toolingSourceFiles...)
	required = append(required, toolingScriptFiles...)
	required = append(required, toolingRuntimeFiles...)
	if err := ensureRelativeFiles(repoRoot, required); err != nil {
		return scope, err
	}

	scope.FrameworkGuidelineFiles = frameworkFiles
	scope.CommandFiles = commandFiles
	scope.CandidateIntentFiles = sortAndDedupe(candidateIntentFiles)
	scope.GuidanceSkillFiles = sortAndDedupe(guidanceSkillFiles)
	scope.RuleGovernanceFiles = ruleFiles
	scope.TemplateGovernanceFiles = templateGovernanceFiles
	scope.TemplateProjectInstanceFiles = templateProjectInstanceFiles
	scope.TemplateEntryFiles = templateEntryFiles
	scope.ProjectEntryFiles = projectEntryFiles
	scope.SourceRepoEntryExampleFiles = sourceRepoEntryExampleFiles
	scope.AgentOperabilityFiles = agentOperabilityFiles
	scope.ProjectInstanceCompatibilityFiles = projectInstanceCompatibilityFiles
	scope.ToolingContractFiles = toolingContractFiles
	scope.ToolingSourceFiles = sortAndDedupe(toolingSourceFiles)
	scope.ToolingScriptFiles = toolingScriptFiles
	scope.ToolingRuntimeFiles = sortAndDedupe(toolingRuntimeFiles)
	return scope, nil
}

func CollectDefaultSpecFlowDesignScope(repoRoot string) (SpecFlowScope, error) {
	return CollectDefaultSpecFlowDesignScopeForLayout(repoRoot, LayoutAuto)
}

func CollectDefaultSpecFlowDesignScopeForLayout(repoRoot, requestedLayout string) (SpecFlowScope, error) {
	roots, err := resolveRoots(repoRoot, requestedLayout)
	if err != nil {
		return SpecFlowScope{}, err
	}
	scope := SpecFlowScope{
		Profile:                          "default_design_baseline",
		Layout:                           roots.Layout,
		FrameworkRoot:                    roots.FrameworkRoot,
		TemplateRoot:                     roots.TemplateRoot,
		ToolingRoot:                      roots.ToolingRoot,
		ProjectInstanceCompatibilityMode: roots.ProjectInstanceCompatibilityMode,
	}

	commandFiles, err := globRelative(repoRoot, joinPath(roots.FrameworkRoot, "lifecycle/*.md"))
	if err != nil {
		return scope, err
	}
	if len(commandFiles) == 0 {
		return scope, fmt.Errorf("default design lifecycle files are incomplete")
	}
	candidateIntentStandardFiles, err := globRelative(repoRoot, joinPath(roots.FrameworkRoot, "candidate_intents/*.md"))
	if err != nil {
		return scope, err
	}
	if len(candidateIntentStandardFiles) == 0 {
		return scope, fmt.Errorf("default design candidate intent files are incomplete")
	}
	candidateIntentFiles := sortAndDedupe(append([]string{scope.FrameworkPath("candidate_intent_policy.md")}, candidateIntentStandardFiles...))

	designFoundationFiles := []string{
		scope.FrameworkPath("spec_flow_design_review.md"),
		scope.FrameworkPath("governance/review.md"),
		scope.FrameworkPath("governance/review_scope.md"),
		scope.FrameworkPath("governance/rule_system.md"),
		scope.FrameworkPath("agent_operability_standard.md"),
		scope.FrameworkPath("spec_policy.md"),
		scope.FrameworkPath("advance_policy.md"),
		scope.FrameworkPath("core/object_model.md"),
		scope.FrameworkPath("core/status.md"),
		scope.FrameworkPath("core/repository_mapping.md"),
		scope.FrameworkPath("core/lifecycle_authority.md"),
		scope.FrameworkPath("lifecycle/overview.md"),
		scope.FrameworkPath("operations/entry_routing.md"),
		scope.FrameworkPath("operations/migration.md"),
		scope.FrameworkPath("onboarding_decision_policy.md"),
		scope.FrameworkPath("operations/implementation_change.md"),
		scope.FrameworkPath("spec_writing_guide.md"),
		scope.FrameworkPath("entry_index_registry.md"),
		scope.FrameworkPath("operations/output_standard.md"),
		scope.FrameworkPath("slice_work_state_protocol.md"),
	}
	designFoundationFiles = append(designFoundationFiles, candidateIntentFiles...)
	lifecycleContractFiles := []string{
		scope.FrameworkPath("candidate_handoff_contract.md"),
		scope.FrameworkPath("downgrade_policy.md"),
		scope.FrameworkPath("process_snapshot_contract.md"),
		scope.FrameworkPath("slice_work_state_protocol.md"),
		scope.FrameworkPath("lifecycle/recovery.md"),
		scope.FrameworkPath("operations/output_standard.md"),
	}
	templateProcessStateFiles := templateProcessStateFiles(scope)
	templateEntryFiles := templateEntryFiles(scope)
	projectEntryFiles := projectEntryFiles(roots.Layout)
	sourceRepoEntryExampleFiles := sourceRepoEntryExampleFiles(roots.Layout)

	required := append([]string{}, designFoundationFiles...)
	required = append(required, lifecycleContractFiles...)
	required = append(required, templateProcessStateFiles...)
	required = append(required, templateEntryFiles...)
	required = append(required, projectEntryFiles...)
	required = append(required, sourceRepoEntryExampleFiles...)
	if err := ensureRelativeFiles(repoRoot, required); err != nil {
		return scope, err
	}

	scope.FrameworkGuidelineFiles = sortAndDedupe(designFoundationFiles)
	scope.CommandFiles = commandFiles
	scope.CandidateIntentFiles = sortAndDedupe(candidateIntentFiles)
	scope.TemplateGovernanceFiles = sortAndDedupe(append(lifecycleContractFiles, templateProcessStateFiles...))
	scope.TemplateEntryFiles = templateEntryFiles
	scope.ProjectEntryFiles = projectEntryFiles
	scope.SourceRepoEntryExampleFiles = sourceRepoEntryExampleFiles
	return scope, nil
}

func (scope SpecFlowScope) FrameworkPath(relPath string) string {
	return joinPath(scope.FrameworkRoot, relPath)
}

func (scope SpecFlowScope) TemplatePath(relPath string) string {
	return joinPath(scope.TemplateRoot, relPath)
}

func (scope SpecFlowScope) ToolingPath(relPath string) string {
	return joinPath(scope.ToolingRoot, relPath)
}

func collectAgentOperabilityFiles(scope SpecFlowScope, projectEntryFiles, sourceRepoEntryExampleFiles, templateEntryFiles, templateProcessStateFiles, commandFiles, candidateIntentFiles, guidanceSkillFiles, sharedGovernanceFiles, processStateContractFiles, toolingContractFiles []string) []string {
	files := []string{
		scope.FrameworkPath("advance_policy.md"),
		scope.FrameworkPath("agent_operability_standard.md"),
		scope.FrameworkPath("core/adoption_modes.md"),
		scope.FrameworkPath("core/context_card.md"),
		scope.FrameworkPath("core/freshness.md"),
		scope.FrameworkPath("core/independent_evaluation.md"),
		scope.FrameworkPath("core/object_model.md"),
		scope.FrameworkPath("core/status.md"),
		scope.FrameworkPath("core/repository_mapping.md"),
		scope.FrameworkPath("core/lifecycle_authority.md"),
		scope.FrameworkPath("lifecycle/overview.md"),
		scope.FrameworkPath("governance/review.md"),
		scope.FrameworkPath("governance/review_scope.md"),
		scope.FrameworkPath("operations/migration.md"),
		scope.FrameworkPath("onboarding_decision_policy.md"),
		scope.FrameworkPath("operations/entry_routing.md"),
		scope.FrameworkPath("operations/implementation_change.md"),
		scope.FrameworkPath("operations/output_standard.md"),
		scope.FrameworkPath("severity_policy.md"),
		scope.FrameworkPath("slice_work_state_protocol.md"),
		scope.FrameworkPath("spec_flow_design_review.md"),
		scope.FrameworkPath("spec_flow_review.md"),
		scope.FrameworkPath("spec_policy.md"),
		scope.FrameworkPath("spec_writing_guide.md"),
		scope.FrameworkPath("spec_authoring_baseline.md"),
	}
	files = append(files, projectEntryFiles...)
	files = append(files, sourceRepoEntryExampleFiles...)
	files = append(files, templateEntryFiles...)
	files = append(files, commandFiles...)
	files = append(files, candidateIntentFiles...)
	files = append(files, guidanceSkillFiles...)
	files = append(files, sharedGovernanceFiles...)
	files = append(files, processStateContractFiles...)
	files = append(files, toolingContractFiles...)
	files = append(files, templateProcessStateFiles...)
	return sortAndDedupe(files)
}

func layeredFrameworkFiles(repoRoot string, roots layoutRoots) ([]string, error) {
	result, err := globRelative(repoRoot, joinPath(roots.FrameworkRoot, "*.md"))
	if err != nil {
		return nil, err
	}
	for _, relDir := range []string{
		joinPath(roots.FrameworkRoot, "core"),
		joinPath(roots.FrameworkRoot, "lifecycle"),
		joinPath(roots.FrameworkRoot, "governance"),
		joinPath(roots.FrameworkRoot, "operations"),
	} {
		files, err := walkRelativeFiles(repoRoot, relDir, ".md")
		if err != nil {
			return nil, err
		}
		result = append(result, files...)
	}
	return sortAndDedupe(result), nil
}

func collectProjectInstanceCompatibilityFiles(repoRoot string, scope SpecFlowScope) ([]string, error) {
	if scope.ProjectInstanceCompatibilityMode == CompatibilityTemplateBootstrap {
		return walkRelativeFiles(repoRoot, scope.TemplatePath("docs/specs"), ".md")
	}

	required := []string{
		"docs/specs/_status.md",
		"docs/specs/repository_mapping.md",
		"docs/specs/rules/stable/s_g_rule_repository_baseline.md",
	}
	if err := ensureRelativeFiles(repoRoot, required); err != nil {
		return nil, err
	}

	root := filepath.Join(repoRoot, filepath.FromSlash("docs/specs"))
	if _, err := os.Stat(root); err != nil {
		return nil, fmt.Errorf("required scope directory missing: docs/specs")
	}

	result := append([]string{}, required...)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		relPath := filepath.ToSlash(rel)
		if strings.HasPrefix(relPath, "docs/specs/_governance_review/") {
			return nil
		}
		result = append(result, relPath)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return sortAndDedupe(result), nil
}

func resolveRoots(repoRoot, requestedLayout string) (layoutRoots, error) {
	layout, err := ResolveLayout(repoRoot, requestedLayout)
	if err != nil {
		return layoutRoots{}, err
	}
	switch layout {
	case LayoutInstalledProject:
		return layoutRoots{
			Layout:                           LayoutInstalledProject,
			FrameworkRoot:                    "specflow/framework",
			TemplateRoot:                     "specflow/templates",
			ToolingRoot:                      "specflow/tooling",
			ProjectInstanceCompatibilityMode: CompatibilityInstalledProject,
		}, nil
	case LayoutSourceRepo:
		return layoutRoots{
			Layout:                           LayoutSourceRepo,
			FrameworkRoot:                    "framework",
			TemplateRoot:                     "templates",
			ToolingRoot:                      "tooling",
			ProjectInstanceCompatibilityMode: CompatibilityTemplateBootstrap,
		}, nil
	default:
		return layoutRoots{}, fmt.Errorf("unsupported resolved review layout %q", layout)
	}
}

func layoutMarkersPresent(repoRoot, layout string) bool {
	var requiredDirs []string
	switch layout {
	case LayoutSourceRepo:
		requiredDirs = []string{
			"framework/core",
			"framework/lifecycle",
			"framework/governance",
			"framework/operations",
			"templates",
			"tooling",
		}
	case LayoutInstalledProject:
		requiredDirs = []string{
			"specflow/framework/core",
			"specflow/framework/lifecycle",
			"specflow/framework/governance",
			"specflow/framework/operations",
			"specflow/templates",
			"specflow/tooling",
		}
	default:
		return false
	}
	for _, relDir := range requiredDirs {
		info, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(relDir)))
		if err != nil || !info.IsDir() {
			return false
		}
	}
	return true
}

func templateProcessStateFiles(scope SpecFlowScope) []string {
	return []string{
		scope.TemplatePath("docs/specs/_status.md"),
		scope.TemplatePath("docs/specs/_check_work/README.md"),
		scope.TemplatePath("docs/specs/_check_result/README.md"),
		scope.TemplatePath("docs/specs/_plans/README.md"),
		scope.TemplatePath("docs/specs/_plans/draft/README.md"),
		scope.TemplatePath("docs/specs/_plans/active/README.md"),
		scope.TemplatePath("docs/specs/_verify_result/README.md"),
		scope.TemplatePath("docs/specs/_stable_verify_result/README.md"),
		scope.TemplatePath("docs/specs/_governance_review/README.md"),
		scope.TemplatePath("docs/specs/_independent_evaluation/README.md"),
	}
}

func templateEntryFiles(scope SpecFlowScope) []string {
	return []string{
		scope.TemplatePath("AGENTS.md"),
		scope.TemplatePath("GEMINI.md"),
		scope.TemplatePath("CLAUDE.md"),
	}
}

func projectEntryFiles(layout string) []string {
	if layout != LayoutInstalledProject {
		return nil
	}
	return []string{
		"AGENTS.md",
		"GEMINI.md",
		"CLAUDE.md",
	}
}

func sourceRepoEntryExampleFiles(layout string) []string {
	if layout != LayoutSourceRepo {
		return nil
	}
	return []string{"example.md"}
}

func globRelative(repoRoot, pattern string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(repoRoot, filepath.FromSlash(pattern)))
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		rel, err := filepath.Rel(repoRoot, match)
		if err != nil {
			return nil, err
		}
		result = append(result, filepath.ToSlash(rel))
	}
	sort.Strings(result)
	return result, nil
}

func walkRelativeFiles(repoRoot, relDir, suffix string) ([]string, error) {
	root := filepath.Join(repoRoot, filepath.FromSlash(relDir))
	if _, err := os.Stat(root); err != nil {
		return nil, fmt.Errorf("required scope directory missing: %s", relDir)
	}

	result := []string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), suffix) {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		result = append(result, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(result)
	return result, nil
}

func ensureRelativeFiles(repoRoot string, relPaths []string) error {
	for _, relPath := range relPaths {
		if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(relPath))); err != nil {
			return fmt.Errorf("required scope file missing: %s", relPath)
		}
	}
	return nil
}

func fileExists(repoRoot, relPath string) bool {
	_, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(relPath)))
	return err == nil
}

func joinPath(root, relPath string) string {
	root = strings.Trim(strings.ReplaceAll(root, "\\", "/"), "/")
	relPath = strings.Trim(strings.ReplaceAll(relPath, "\\", "/"), "/")
	if root == "" {
		return relPath
	}
	if relPath == "" {
		return root
	}
	return root + "/" + relPath
}

func sortAndDedupe(items []string) []string {
	set := map[string]bool{}
	for _, item := range items {
		set[item] = true
	}
	result := make([]string, 0, len(set))
	for item := range set {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}
