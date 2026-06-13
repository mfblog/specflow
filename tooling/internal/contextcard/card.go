package contextcard

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/context"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/unitappendix"
)

// UnitState classifies a unit's current lifecycle state.
type UnitState string

const (
	StateStableIdle       UnitState = "stable_idle"
	StateStableVerify     UnitState = "stable_verify"
	StateCandidateCheck   UnitState = "candidate_check"
	StateCandidatePending UnitState = "candidate_pending_impl"
	StateCandidateVerify  UnitState = "candidate_verify"
	StateCandidatePromote UnitState = "candidate_promote"
	StateUnregistered     UnitState = "unregistered"
	StateAdoptionLimited  UnitState = "adoption_limited"
)

// RuleState classifies a rule's current state.
type RuleState string

const (
	RuleStableBound        RuleState = "rule_stable_bound"
	RuleStableGlobal       RuleState = "rule_stable_global"
	RuleCandidateBound     RuleState = "rule_candidate_bound"
	RuleCandidateGlobal    RuleState = "rule_candidate_global"
	RuleCandidateNewBound  RuleState = "rule_candidate_new_bound"
	RuleCandidateNewGlobal RuleState = "rule_candidate_new_global"
	RuleUnregistered       RuleState = "rule_unregistered"
)

type AffectedUnitRow struct {
	Name        string
	Layer       string
	NextCommand string
	Impact      string
}

// --- state classification (kept from original) ---

func classifyUnitState(s statusfile.ObjectStatus) UnitState {
	if s.Object == "" {
		return StateUnregistered
	}
	stable := strings.TrimSpace(strings.ToLower(s.Stable))
	candidate := strings.TrimSpace(strings.ToLower(s.Candidate))
	active := strings.TrimSpace(strings.ToLower(s.ActiveLayer))
	next := strings.TrimSpace(strings.ToLower(s.NextCommand))
	notes := strings.TrimSpace(strings.ToLower(s.Notes))

	switch {
	case stable == "yes" && candidate == "no" && active == "stable" && next == "unit_fork":
		return StateStableIdle
	case stable == "yes" && candidate == "no" && active == "stable" && next == "unit_stable_verify":
		return StateStableVerify
	case next == "unit_init":
		return StateUnregistered
	case next == "unit_new":
		return StateCandidateCheck
	case candidate == "yes" && next == "unit_check":
		return StateCandidateCheck
	case candidate == "yes" && next == "unit_verify" && strings.Contains(notes, "pending_impl"):
		return StateCandidatePending
	case candidate == "yes" && next == "unit_verify":
		return StateCandidateVerify
	case candidate == "yes" && next == "unit_promote":
		return StateCandidatePromote
	default:
		return StateUnregistered
	}
}

func classifyRuleState(stableFile, candidateFile, scope string) RuleState {
	hasStable := stableFile != ""
	hasCandidate := candidateFile != ""

	switch {
	case !hasStable && !hasCandidate:
		return RuleUnregistered
	case hasStable && !hasCandidate && scope == "global":
		return RuleStableGlobal
	case hasStable && !hasCandidate && scope == "bound":
		return RuleStableBound
	case hasStable && hasCandidate && scope == "global":
		return RuleCandidateGlobal
	case hasStable && hasCandidate && scope == "bound":
		return RuleCandidateBound
	case !hasStable && hasCandidate && scope == "global":
		return RuleCandidateNewGlobal
	case !hasStable && hasCandidate && scope == "bound":
		return RuleCandidateNewBound
	default:
		return RuleUnregistered
	}
}

// ============================================================================
// unit card generation
// ============================================================================

// UnitCard generates a self-contained context card for a unit.
// It collects and inlines all essential files (specs, rules, appendices)
// so the agent can get everything it needs in one command output.
func UnitCard(repoRoot, unitName string) (string, error) {
	statuses, err := statusfile.LoadObjectStatuses(repoRoot)
	if err != nil {
		statuses = nil
	}

	var st *statusfile.ObjectStatus
	for i := range statuses {
		if statuses[i].Object == unitName {
			st = &statuses[i]
			break
		}
	}
	if st == nil {
		st = &statusfile.ObjectStatus{Object: unitName}
	}

	state := classifyUnitState(*st)

	adoptionMode := readAdoptionMode(repoRoot)
	implPaths := resolveImplPaths(repoRoot, unitName)

	intent, specContent := readSpecAndIntent(repoRoot, unitName)
	checkDate := readCheckDate(repoRoot, unitName)

	// Collect files to inline (essential) and reference
	essential, reference := collectUnitFiles(repoRoot, unitName, state, specContent)

	var buf strings.Builder

	// --- header ---
	writeCardHeader(&buf, unitName, "unit")

	// --- status ---
	writeUnitStatus(&buf, state, st, intent, checkDate, adoptionMode)

	// --- guidance ---
	lifecycleContent := writeUnitGuidance(&buf, repoRoot, unitName, implPaths, state, adoptionMode)

	// --- core truth (inlined files) ---
	if len(essential) > 0 {
		buf.WriteString("## Core Truth\n\n")
		for _, f := range essential {
			writeInlinedFile(&buf, f)
		}
	}

	// --- references ---
	if len(reference) > 0 {
		buf.WriteString("## References (read if needed)\n\n")
		for _, f := range reference {
			label := ""
			if !f.Exists {
				label = " (missing)"
			}
			fmt.Fprintf(&buf, "- %s%s\n", f.Path, label)
		}
		buf.WriteString("\n")
	}

	// --- writes ---
	writeUnitWrites(&buf, state, unitName, implPaths)

	// --- reads ---
	writeUnitReads(&buf, state, unitName, essential, reference)

	// --- blocked ---
	writeUnitBlocked(&buf, state, lifecycleContent)

	// --- close ---
	writeUnitClose(&buf, state, unitName, lifecycleContent)

	return buf.String(), nil
}

// ============================================================================
// rule card generation
// ============================================================================

// RuleCard generates a self-contained context card for a rule.
func RuleCard(repoRoot, ruleID string) (string, error) {
	scope := "bound"
	if strings.HasPrefix(ruleID, "g_rule_") {
		scope = "global"
	}

	stablePath := filepath.Join(repoRoot, filepath.FromSlash("docs/specs/rules/stable/s_"+ruleID+".md"))
	candidatePath := filepath.Join(repoRoot, filepath.FromSlash("docs/specs/rules/candidate/c_"+ruleID+".md"))

	stableExists := false
	candidateExists := false
	version := ""
	candidateVersion := ""

	if data, err := os.ReadFile(stablePath); err == nil {
		stableExists = true
		content := string(data)
		version = extractFrontmatter(content, "rule_version")
		if version == "" {
			version = "1.0.0"
		}
	}
	if data, err := os.ReadFile(candidatePath); err == nil {
		candidateExists = true
		cv := extractFrontmatter(string(data), "rule_version")
		if cv != "" {
			candidateVersion = cv
		} else {
			candidateVersion = "0.1.0"
		}
	}
	if !stableExists && version == "" {
		version = "0.1.0"
	}

	state := classifyRuleState(
		map[bool]string{true: stablePath, false: ""}[stableExists],
		map[bool]string{true: candidatePath, false: ""}[candidateExists],
		scope,
	)

	// Collect affected units
	affected := findAffectedUnits(repoRoot, ruleID, scope)

	// Collect rule spec files to inline
	essential := collectRuleFiles(repoRoot, ruleID, state, stableExists, candidateExists, stablePath, candidatePath)

	// Collect reference files
	reference := collectRuleRefs(state)

	// Load rule_system.md content for section extraction in guidance/blocked
	ruleSystemPath := filepath.Join(repoRoot, filepath.FromSlash("framework/governance/rule_system.md"))
	var ruleSystemContent string
	if data, err := os.ReadFile(ruleSystemPath); err == nil {
		ruleSystemContent = strings.ReplaceAll(string(data), "\r\n", "\n")
	}

	var buf strings.Builder

	writeCardHeader(&buf, ruleID, "rule")
	writeRuleStatus(&buf, state, ruleID, scope, version, candidateVersion, len(affected))
	writeRuleGuidance(&buf, state, ruleID, scope, ruleSystemContent)

	if len(essential) > 0 {
		buf.WriteString("## Core Truth\n\n")
		for _, f := range essential {
			writeInlinedFile(&buf, f)
		}
	}

	if len(reference) > 0 {
		buf.WriteString("## References (read if needed)\n\n")
		for _, f := range reference {
			label := ""
			if !f.Exists {
				label = " (missing)"
			}
			fmt.Fprintf(&buf, "- %s%s\n", f.Path, label)
		}
		buf.WriteString("\n")
	}

	// affected units table
	if len(affected) > 0 {
		buf.WriteString("## IMPACTS (consumer units)\n\n")
		buf.WriteString("| Unit | Layer | Next Command |\n")
		buf.WriteString("|------|-------|-------------|\n")
		for _, u := range affected {
			fmt.Fprintf(&buf, "| %s | %s | %s |\n", u.Name, u.Layer, u.NextCommand)
		}
		buf.WriteString("\n")
	} else {
		buf.WriteString("## IMPACTS (consumer units)\n\n")
		buf.WriteString("(none)\n\n")
	}

	writeRuleBlocked(&buf, state, ruleID, scope, ruleSystemContent)
	writeRuleClose(&buf, state, ruleID, ruleSystemContent)

	return buf.String(), nil
}

// ============================================================================
// file collection — unit
// ============================================================================

func collectUnitFiles(repoRoot, unitName string, state UnitState, specContent string) (essential, reference []context.FileItem) {
	rules := unitFileInputRules(state)
	if rules == nil {
		return nil, nil
	}
	return resolveEssentialAndRef(repoRoot, unitName, rules.essential, rules.reference)
}

type cardInputRules struct {
	essential []context.InputRule
	reference []context.InputRule
}

func unitFileInputRules(state UnitState) *cardInputRules {
	base := func() *cardInputRules {
		return &cardInputRules{
			essential: []context.InputRule{
				{PathTemplate: "docs/specs/_status.md", Essential: true},
				{PathTemplate: "docs/specs/repository_mapping.md", Essential: true},
			},
			reference: nil,
		}
	}

	switch state {
	case StateStableIdle:
		r := base()
		r.reference = append(r.reference,
			context.InputRule{PathTemplate: "docs/specs/repository_mapping.md", Essential: false},
			context.InputRule{PathTemplate: "docs/specs/units/stable/s_unit_{object}.md", Essential: false},
			context.InputRule{Resolve: resolveStableAppendices, Essential: false},
			context.InputRule{Resolve: resolveRuleRefs, Essential: false},
			context.InputRule{Resolve: resolveUnitRefs, Essential: false},
			context.InputRule{PathTemplate: "framework/lifecycle/overview.md", Essential: false},
			context.InputRule{PathTemplate: "framework/core/object_model.md", Essential: false},
		)
		return r
	case StateStableVerify:
		r := base()
		r.reference = append(r.reference,
			context.InputRule{PathTemplate: "docs/specs/repository_mapping.md", Essential: false},
			context.InputRule{PathTemplate: "docs/specs/units/stable/s_unit_{object}.md", Essential: false},
			context.InputRule{Resolve: resolveStableAppendices, Essential: false},
			context.InputRule{Resolve: resolveRuleRefs, Essential: false},
			context.InputRule{Resolve: resolveUnitRefs, Essential: false},
			context.InputRule{PathTemplate: "docs/specs/_stable_verify_result/unit/{object}.md", Essential: false, Optional: true},
			context.InputRule{PathTemplate: "framework/lifecycle/unit_stable_verify.md", Essential: false},
			context.InputRule{PathTemplate: "framework/process_snapshot_contract.md", Essential: false},
			context.InputRule{PathTemplate: "framework/core/independent_evaluation.md", Essential: false},
		)
		return r
	case StateCandidateCheck:
		r := base()
		r.reference = append(r.reference,
			context.InputRule{PathTemplate: "docs/specs/units/candidate/c_unit_{object}.md", Essential: false},
			context.InputRule{PathTemplate: "docs/specs/units/stable/s_unit_{object}.md", Essential: false, Optional: true},
			context.InputRule{Resolve: resolveCandidateAppendices, Essential: false},
			context.InputRule{Resolve: resolveRuleRefs, Essential: false},
			context.InputRule{Resolve: resolveUnitRefs, Essential: false},
			context.InputRule{PathTemplate: "docs/specs/_check_result/unit/{object}.md", Essential: false, Optional: true},
			context.InputRule{PathTemplate: "framework/lifecycle/unit_check.md", Essential: false},
			context.InputRule{PathTemplate: "framework/process_snapshot_contract.md", Essential: false},
			context.InputRule{PathTemplate: "framework/spec_writing_guide.md", Essential: false},
		)
		return r
	case StateCandidatePending:
		r := base()
		r.reference = append(r.reference,
			context.InputRule{PathTemplate: "docs/specs/repository_mapping.md", Essential: false},
			context.InputRule{PathTemplate: "docs/specs/units/candidate/c_unit_{object}.md", Essential: false},
			context.InputRule{PathTemplate: "docs/specs/units/stable/s_unit_{object}.md", Essential: false, Optional: true},
			context.InputRule{Resolve: resolveCandidateAppendices, Essential: false},
			context.InputRule{Resolve: resolveRuleRefs, Essential: false},
			context.InputRule{Resolve: resolveUnitRefs, Essential: false},
			context.InputRule{PathTemplate: "docs/specs/_check_result/unit/{object}.md", Essential: false, Optional: true},
			context.InputRule{PathTemplate: "framework/lifecycle/unit_impl.md", Essential: false},
			context.InputRule{PathTemplate: "framework/spec_writing_guide.md", Essential: false},
		)
		return r
	case StateCandidateVerify:
		r := base()
		r.reference = append(r.reference,
			context.InputRule{PathTemplate: "docs/specs/units/candidate/c_unit_{object}.md", Essential: false},
			context.InputRule{PathTemplate: "docs/specs/units/stable/s_unit_{object}.md", Essential: false, Optional: true},
			context.InputRule{Resolve: resolveCandidateAppendices, Essential: false},
			context.InputRule{Resolve: resolveRuleRefs, Essential: false},
			context.InputRule{Resolve: resolveUnitRefs, Essential: false},
			context.InputRule{PathTemplate: "docs/specs/repository_mapping.md", Essential: false},
			context.InputRule{PathTemplate: "docs/specs/_check_result/unit/{object}.md", Essential: false, Optional: true},
			context.InputRule{PathTemplate: "framework/lifecycle/unit_verify.md", Essential: false},
			context.InputRule{PathTemplate: "framework/process_snapshot_contract.md", Essential: false},
			context.InputRule{PathTemplate: "framework/core/independent_evaluation.md", Essential: false},
		)
		return r
	case StateCandidatePromote:
		r := base()
		r.reference = append(r.reference,
			context.InputRule{PathTemplate: "docs/specs/units/candidate/c_unit_{object}.md", Essential: false},
			context.InputRule{PathTemplate: "docs/specs/units/stable/s_unit_{object}.md", Essential: false},
			context.InputRule{Resolve: resolveCandidateAppendices, Essential: false},
			context.InputRule{PathTemplate: "docs/specs/_verify_result/unit/{object}.md", Essential: false, Optional: true},
			context.InputRule{PathTemplate: "framework/lifecycle/unit_promote.md", Essential: false},
			context.InputRule{PathTemplate: "framework/spec_writing_guide.md", Essential: false},
			context.InputRule{PathTemplate: "framework/candidate_intent.md", Essential: false},
			context.InputRule{PathTemplate: "framework/process_snapshot_contract.md", Essential: false},
		)
		return r
	case StateUnregistered:
		r := base()
		r.reference = append(r.reference,
			context.InputRule{PathTemplate: "docs/specs/repository_mapping.md", Essential: false},
			context.InputRule{PathTemplate: "framework/operations/entry_routing.md", Essential: false},
			context.InputRule{PathTemplate: "framework/lifecycle/unit_init_new_fork.md", Essential: false},
			context.InputRule{PathTemplate: "framework/spec_writing_guide.md", Essential: false},
			context.InputRule{PathTemplate: "framework/candidate_intent.md", Essential: false},
		)
		return r
	default:
		return nil
	}
}

// ============================================================================
// file collection — rule
// ============================================================================

func collectRuleFiles(repoRoot, ruleID string, state RuleState, stableExists, candidateExists bool, stablePath, candidatePath string) []context.FileItem {
	var rules []context.InputRule
	switch {
	case stableExists:
		rules = append(rules, context.InputRule{PathTemplate: filepath.ToSlash(stablePath)[len(filepath.ToSlash(repoRoot)):], Essential: true})
		fallthrough
	case candidateExists:
		rules = append(rules, context.InputRule{PathTemplate: filepath.ToSlash(candidatePath)[len(filepath.ToSlash(repoRoot)):], Essential: true})
	}
	// Also collect referent framework files
	rules = append(rules,
		context.InputRule{PathTemplate: "framework/governance/rule_system.md", Essential: true},
	)
	items := context.ResolveInputRules(repoRoot, ruleID, rules)
	// Only return the ones that exist
	var result []context.FileItem
	for _, item := range items {
		if item.Exists {
			result = append(result, item)
		}
	}
	return result
}

func collectRuleRefs(state RuleState) []context.FileItem {
	refs := []context.InputRule{
		{PathTemplate: "framework/governance/rule_system.md", Essential: false},
	}
	switch state {
	case RuleStableBound, RuleCandidateBound, RuleCandidateNewBound, RuleCandidateGlobal, RuleCandidateNewGlobal, RuleStableGlobal:
		refs = append(refs,
			context.InputRule{PathTemplate: "framework/governance/rules/rule_new.md", Essential: false},
			context.InputRule{PathTemplate: "framework/governance/rules/rule_sync.md", Essential: false},
		)
	case RuleUnregistered:
		refs = append(refs,
			context.InputRule{PathTemplate: "framework/governance/rule_system.md", Essential: false},
		)
	}
	items := context.ResolveInputRules("", "", refs)
	var result []context.FileItem
	for _, item := range items {
		if item.Exists {
			result = append(result, item)
		}
	}
	return result
}

// ============================================================================
// resolve helpers (via context.InputRule.Resolve callback)
// ============================================================================

func resolveCandidateAppendices(repoRoot, object string) ([]context.FileItem, error) {
	return resolveAppendices(repoRoot, object, "candidate")
}

func resolveStableAppendices(repoRoot, object string) ([]context.FileItem, error) {
	return resolveAppendices(repoRoot, object, "stable")
}

func resolveAppendices(repoRoot, object, layer string) ([]context.FileItem, error) {
	entries, err := unitappendix.Scan(repoRoot, "unit", object, layer)
	if err != nil {
		return nil, nil
	}
	var items []context.FileItem
	for _, entry := range entries {
		items = append(items, context.FileItem{
			Path:      entry.FileRef,
			Essential: true,
			Exists:    true,
			Content:   entry.Content,
			LineCount: strings.Count(entry.Content, "\n"),
		})
	}
	return items, nil
}

func resolveRuleRefs(repoRoot, object string) ([]context.FileItem, error) {
	content := readActiveSpecContent(repoRoot, object)
	if content == "" {
		return nil, nil
	}
	refs := parseNamedRefs(content, "rule_refs")
	var items []context.FileItem
	seen := map[string]bool{}

	// Global rules always apply
	globalPattern := filepath.Join(repoRoot, filepath.FromSlash("docs/specs/rules/stable/s_g_rule_*.md"))
	globals, _ := filepath.Glob(globalPattern)
	for _, match := range globals {
		rel, _ := filepath.Rel(repoRoot, match)
		relSlash := filepath.ToSlash(rel)
		if seen[relSlash] {
			continue
		}
		seen[relSlash] = true
		item := context.ResolveFileItem(repoRoot, relSlash, "", true, false)
		if item.Exists {
			items = append(items, item)
		}
	}

	for _, ref := range refs {
		path := resolveRuleRefPath(ref)
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		item := context.ResolveFileItem(repoRoot, path, "", true, false)
		if item.Exists {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	return items, nil
}

func resolveUnitRefs(repoRoot, object string) ([]context.FileItem, error) {
	content := readActiveSpecContent(repoRoot, object)
	if content == "" {
		return nil, nil
	}
	refs := parseNamedRefs(content, "unit_refs")
	var items []context.FileItem
	seen := map[string]bool{}
	for _, ref := range refs {
		path := resolveUnitRefPath(ref)
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		item := context.ResolveFileItem(repoRoot, path, "", true, false)
		if item.Exists {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	return items, nil
}

func resolveEssentialAndRef(repoRoot, object string, essential, reference []context.InputRule) ([]context.FileItem, []context.FileItem) {
	ess := context.ResolveInputRules(repoRoot, object, essential)
	ref := context.ResolveInputRules(repoRoot, object, reference)
	// Remove duplicates: reference should not repeat files already in essential
	essPaths := map[string]bool{}
	for _, e := range ess {
		essPaths[e.Path] = true
	}
	var filtered []context.FileItem
	for _, r := range ref {
		if !essPaths[r.Path] {
			filtered = append(filtered, r)
		}
	}
	return ess, filtered
}

// ============================================================================
// section writers — unit
// ============================================================================

func writeCardHeader(buf *strings.Builder, name, kind string) {
	fmt.Fprintf(buf, "# Context Card: %s/%s\n\n", kind, name)
}

func writeUnitStatus(buf *strings.Builder, state UnitState, st *statusfile.ObjectStatus, intent, checkDate, adoptionMode string) {
	buf.WriteString("## STATUS\n\n")

	if adoptionMode != "" {
		buf.WriteString("- Adoption mode: **")
		buf.WriteString(adoptionMode)
		buf.WriteString("**\n")
	}

	switch state {
	case StateStableIdle:
		buf.WriteString("- Stage: stable (idle) | Next: unit_fork\n")
		buf.WriteString("- Layer: stable\n")
		buf.WriteString("- Stable: yes | Candidate: no\n")
	case StateStableVerify:
		buf.WriteString("- Stage: unit_stable_verify | Next: unit_fork (after verification)\n")
		buf.WriteString("- Layer: stable\n")
	case StateCandidateCheck:
		buf.WriteString("- Stage: unit_check | Next: unit_verify (on pass)\n")
		buf.WriteString("- Layer: candidate")
		if intent != "" {
			fmt.Fprintf(buf, " | Intent: %s", intent)
		}
		buf.WriteString("\n")
		buf.WriteString("- Stable: ")
		buf.WriteString(st.Stable)
		buf.WriteString(" | Candidate: ")
		buf.WriteString(st.Candidate)
		buf.WriteString("\n")
	case StateCandidatePending:
		buf.WriteString("- Stage: pending_impl | Next: unit_verify\n")
		buf.WriteString("- Layer: candidate")
		if intent != "" {
			fmt.Fprintf(buf, " | Intent: %s", intent)
		}
		buf.WriteString("\n")
		buf.WriteString("- Notes: pending_impl\n")
		if checkDate != "" {
			fmt.Fprintf(buf, "- Check passed: %s\n", checkDate)
		}
	case StateCandidateVerify:
		buf.WriteString("- Stage: unit_verify | Next: unit_promote\n")
		buf.WriteString("- Layer: candidate")
		if intent != "" {
			fmt.Fprintf(buf, " | Intent: %s", intent)
		}
		buf.WriteString("\n")
	case StateCandidatePromote:
		buf.WriteString("- Stage: unit_promote | Next: unit_fork (on success)\n")
		buf.WriteString("- Layer: candidate → stable\n")
	case StateUnregistered:
		buf.WriteString("- This unit is not yet registered in `docs/specs/_status.md`\n")
	}
	buf.WriteString("\n")
}

var stateToLifecycleFile = map[UnitState]string{
	StateStableIdle:       "framework/lifecycle/unit_init_new_fork.md",
	StateStableVerify:     "framework/lifecycle/unit_stable_verify.md",
	StateCandidateCheck:   "framework/lifecycle/unit_check.md",
	StateCandidatePending: "framework/lifecycle/unit_impl.md",
	StateCandidateVerify:  "framework/lifecycle/unit_verify.md",
	StateCandidatePromote: "framework/lifecycle/unit_promote.md",
	StateUnregistered:     "framework/lifecycle/unit_init_new_fork.md",
}

func writeUnitGuidance(buf *strings.Builder, repoRoot, unitName, implPaths string, state UnitState, adoptionMode string) string {
	if adoptionMode != "" {
		writeAdoptionPreamble(buf, adoptionMode)
	}

	buf.WriteString("## GUIDANCE\n\n")

	file, ok := stateToLifecycleFile[state]
	if !ok {
		return ""
	}

	absPath := filepath.Join(repoRoot, filepath.FromSlash(file))
	data, err := os.ReadFile(absPath)
	if err != nil {
		buf.WriteString("(lifecycle file not found: ")
		buf.WriteString(file)
		buf.WriteString(")\n\n")
		return ""
	}
	rawText := strings.ReplaceAll(string(data), "\r\n", "\n")
	rawText = strings.ReplaceAll(rawText, "{unit}", unitName)

	displayText := stripFrontmatter(rawText)
	displayText = strings.TrimSpace(displayText)
	displayText = demoteHeadings(displayText)

	buf.WriteString("The complete execution procedure from `")
	buf.WriteString(file)
	buf.WriteString("`:\n\n")
	buf.WriteString(displayText)
	buf.WriteString("\n\n")

	return rawText
}

// demoteHeadings adds two '#' to every ATX heading line so the inlined
// content nests cleanly under the card's own ## sections (e.g. # Title → ### Title).
func demoteHeadings(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "#") {
			lines[i] = "##" + line
		}
	}
	return strings.Join(lines, "\n")
}

// stripFrontmatter removes YAML frontmatter delimited by --- markers.
func stripFrontmatter(s string) string {
	if !strings.HasPrefix(s, "---\n") {
		return s
	}
	if end := strings.Index(s[4:], "\n---\n"); end >= 0 {
		return s[end+5:]
	}
	return s
}

// extractMarkdownSection finds a section by ATX heading title and returns its
// body content (everything between that heading and the next heading at any level).
func extractMarkdownSection(md, sectionName string) string {
	lines := strings.Split(md, "\n")
	inSection := false
	var body []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			if inSection {
				break
			}
			rest := trimmed
			for strings.HasPrefix(rest, "#") {
				rest = rest[1:]
			}
			if strings.TrimSpace(rest) == sectionName {
				inSection = true
				continue
			}
		}
		if inSection {
			body = append(body, line)
		}
	}
	return strings.TrimSpace(strings.Join(body, "\n"))
}

// parseListItems extracts bullet-list items from a named markdown section.
func parseListItems(md, sectionName string) []string {
	section := extractMarkdownSection(md, sectionName)
	if section == "" {
		return nil
	}
	var items []string
	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			items = append(items, strings.TrimPrefix(trimmed, "- "))
		}
	}
	return items
}

// parseCloseCommand extracts the tooling invocation or terminal outcome line
// from the "How to End" section of a lifecycle file.
func parseCloseCommand(md string) string {
	section := extractMarkdownSection(md, "How to End")
	if section == "" {
		return ""
	}
	var lastToolingLine, lastTerminalLine string
	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") {
			continue
		}
		if strings.Contains(trimmed, "specflowctl command close") {
			lastToolingLine = trimmed
		}
		if strings.Contains(trimmed, "command close --command") {
			lastToolingLine = trimmed
		}
		if strings.Contains(trimmed, "Terminal outcome:") {
			lastTerminalLine = trimmed
		}
	}
	if lastToolingLine != "" {
		return lastToolingLine
	}
	if lastTerminalLine != "" {
		return lastTerminalLine
	}
	return ""
}

func writeInlinedFile(buf *strings.Builder, f context.FileItem) {
	if !f.Exists {
		fmt.Fprintf(buf, "### %s (missing)\n\n", f.Path)
		return
	}
	fmt.Fprintf(buf, "### %s (%d lines)\n\n", f.Path, f.LineCount)
	buf.WriteString("```markdown\n")
	buf.WriteString(f.Content)
	if !strings.HasSuffix(f.Content, "\n") {
		buf.WriteByte('\n')
	}
	buf.WriteString("```\n\n")
}

func writeAdoptionPreamble(buf *strings.Builder, mode string) {
	buf.WriteString("This repository is using **")
	buf.WriteString(mode)
	buf.WriteString("** adoption mode. Your scope is limited.\n")
	buf.WriteString("See `framework/core/adoption_modes.md` for the complete mode boundary table.\n\n")
}

func writeUnitWrites(buf *strings.Builder, state UnitState, unitName, implPaths string) {
	buf.WriteString("## WRITES (owned by this unit)\n\n")

	switch state {
	case StateStableIdle:
		fmt.Fprintf(buf, "- %s/**\n", implPaths)
		buf.WriteString("- tests/**\n")
	case StateStableVerify:
		fmt.Fprintf(buf, "- docs/specs/_stable_verify_result/unit/%s.md\n", unitName)
		fmt.Fprintf(buf, "- %s/**\n", implPaths)
	case StateCandidateCheck:
		fmt.Fprintf(buf, "- docs/specs/_check_result/unit/%s.md\n", unitName)
		fmt.Fprintf(buf, "- docs/specs/_check_work/unit/%s.md\n", unitName)
		fmt.Fprintf(buf, "- docs/specs/units/candidate/c_unit_%s.md\n", unitName)
	case StateCandidatePending:
		fmt.Fprintf(buf, "- docs/specs/units/candidate/c_unit_%s.md\n", unitName)
		buf.WriteString("- docs/specs/units/candidate/appendix/c_unit_")
		buf.WriteString(unitName)
		buf.WriteString("_*.md\n")
		fmt.Fprintf(buf, "- %s/**\n", implPaths)
		buf.WriteString("- tests/**\n")
	case StateCandidateVerify:
		fmt.Fprintf(buf, "- docs/specs/_verify_result/unit/%s.md\n", unitName)
		fmt.Fprintf(buf, "- docs/specs/_check_work/unit/%s.md\n", unitName)
		fmt.Fprintf(buf, "- %s/**\n", implPaths)
		buf.WriteString("- tests/**\n")
	case StateCandidatePromote:
		fmt.Fprintf(buf, "- docs/specs/units/stable/s_unit_%s.md\n", unitName)
		buf.WriteString("- docs/specs/units/stable/appendix/s_unit_")
		buf.WriteString(unitName)
		buf.WriteString("_*.md\n")
		fmt.Fprintf(buf, "- docs/specs/_verify_result/stable/unit/%s.md\n", unitName)
	case StateUnregistered:
		buf.WriteString("- To be determined after registration\n")
	}
	buf.WriteString("\n")
}

func writeUnitReads(buf *strings.Builder, state UnitState, unitName string, essential, reference []context.FileItem) {
	buf.WriteString("## READS (read-only context)\n\n")

	seen := map[string]bool{}

	for _, f := range essential {
		if f.Exists {
			seen[f.Path] = true
			fmt.Fprintf(buf, "- %s\n", f.Path)
		}
	}
	for _, f := range reference {
		if seen[f.Path] {
			continue
		}
		seen[f.Path] = true
		label := ""
		if !f.Exists {
			label = " (missing)"
		}
		fmt.Fprintf(buf, "- %s%s\n", f.Path, label)
	}

	switch state {
	case StateStableIdle:
		if !seen["Implementation files"] {
			buf.WriteString("- Implementation files (for implementation-only path)\n")
			buf.WriteString("- Test files (for implementation-only path)\n")
		}
	case StateStableVerify:
		buf.WriteString("- Implementation files\n")
		buf.WriteString("- Test files\n")
	case StateCandidateVerify:
		buf.WriteString("- Implementation files\n")
		buf.WriteString("- Test files\n")
	}
	buf.WriteString("\n")
}

func writeUnitBlocked(buf *strings.Builder, state UnitState, lifecycleContent string) {
	buf.WriteString("## BLOCKED\n\n")

	parseFromLifecycle := state != StateUnregistered && lifecycleContent != ""
	hasParsedItems := false
	if parseFromLifecycle {
		items := parseListItems(lifecycleContent, "Not Allowed")
		if len(items) > 0 {
			hasParsedItems = true
			for _, item := range items {
				fmt.Fprintf(buf, "- %s\n", item)
			}
		}
	}

	if !hasParsedItems {
		switch state {
		case StateStableIdle:
			buf.WriteString("- Modifying stable-layer truth (must go through fork)\n")
			buf.WriteString("- Modifying `_status.md` (use command close)\n")
			buf.WriteString("- Any rule files\n")
			buf.WriteString("- Other units' specs or status\n")
		case StateStableVerify:
			buf.WriteString("- Modifying stable-layer or candidate-layer truth\n")
			buf.WriteString("- Modifying lifecycle state\n")
			buf.WriteString("- Modifying rule truth\n")
			buf.WriteString("- Modifying implementation files (except small_repair_required)\n")
		case StateCandidateCheck:
			buf.WriteString("- `_status.md` (use command close)\n")
			buf.WriteString("- Implementation files\n")
			buf.WriteString("- Stable-layer truth\n")
			buf.WriteString("- Any rule files\n")
			buf.WriteString("- Other units' specs or status\n")
		case StateCandidatePending:
			buf.WriteString("- `_status.md` (use command close)\n")
			buf.WriteString("- Any rule files (use rule governance)\n")
			buf.WriteString("- Any other unit's spec or status\n")
			buf.WriteString("- Stable-layer truth (must fork first)\n")
			buf.WriteString("- `unit_promote` (must verify first)\n")
		case StateCandidateVerify:
			buf.WriteString("- `_status.md` (use command close)\n")
			buf.WriteString("- Candidate spec (must go back to unit_check to modify)\n")
			buf.WriteString("- Stable-layer truth\n")
			buf.WriteString("- Any rule files\n")
			buf.WriteString("- Other units' specs or status\n")
		case StateCandidatePromote:
			buf.WriteString("- Modifying implementation files\n")
			buf.WriteString("- Manually modifying lifecycle state\n")
		case StateUnregistered:
			buf.WriteString("- Modifying implementation files (not yet registered)\n")
			buf.WriteString("- Advancing lifecycle state\n")
			buf.WriteString("- Modifying any spec files\n")
		}
	}
	buf.WriteString("\n")
}

func writeUnitClose(buf *strings.Builder, state UnitState, unitName string, lifecycleContent string) {
	buf.WriteString("## CLOSE\n\n")

	switch state {
	case StateStableIdle:
		buf.WriteString("stable_idle does not require command close. Enter a candidate round via unit_fork, or check alignment via unit_stable_verify.\n\n")
		return
	case StateUnregistered:
		buf.WriteString("Unit is unregistered, cannot execute command close. Register first by running unit_init or unit_new.\n\n")
		return
	}

	if lifecycleContent == "" {
		fmt.Fprintf(buf, "(lifecycle file not available for state %s; run `specflowctl context card` from a project with the framework installed)\n\n", state)
		return
	}

	closeCmd := parseCloseCommand(lifecycleContent)
	if closeCmd != "" {
		closeCmd = strings.ReplaceAll(closeCmd, "<unit>", unitName)
		closeCmd = strings.ReplaceAll(closeCmd, "{unit}", unitName)
		buf.WriteString(closeCmd)
		buf.WriteString("\n\n")
		return
	}

	if state == StateCandidatePending {
		fmt.Fprintf(buf, "pending_impl uses the unit_impl trigger command; no command close. On completion, run unit_verify:%s.\n\n", unitName)
		return
	}

	fmt.Fprintf(buf, "(close information not found in lifecycle file for state %s — check How to End in GUIDANCE above)\n\n", state)
}

// ============================================================================
// section writers — rule
// ============================================================================

func writeRuleStatus(buf *strings.Builder, state RuleState, ruleID, scope, version, candidateVersion string, consumerCount int) {
	buf.WriteString("## STATUS\n\n")
	switch state {
	case RuleStableGlobal:
		fmt.Fprintf(buf, "- Scope: global | Layer: stable\n- Rule version: %s\n- Consumers: all current-layer units\n", version)
	case RuleStableBound:
		fmt.Fprintf(buf, "- Scope: bound | Layer: stable\n- Rule version: %s\n- Consumers: %d unit(s)\n", version, consumerCount)
	case RuleCandidateGlobal:
		fmt.Fprintf(buf, "- Scope: global | Layer: candidate\n- Stable baseline: s_%s@%s\n- Candidate version: %s\n- Consumers: all current-layer units\n", ruleID, version, candidateVersion)
	case RuleCandidateBound:
		fmt.Fprintf(buf, "- Scope: bound | Layer: candidate\n- Stable baseline: s_%s@%s\n- Candidate version: %s\n- Consumers: %d unit(s)\n", ruleID, version, candidateVersion, consumerCount)
	case RuleCandidateNewGlobal:
		fmt.Fprintf(buf, "- Scope: global | Layer: candidate (new)\n- Candidate version: %s\n- No stable baseline\n", candidateVersion)
	case RuleCandidateNewBound:
		fmt.Fprintf(buf, "- Scope: bound | Layer: candidate (new)\n- Candidate version: %s\n- No stable baseline\n", candidateVersion)
	case RuleUnregistered:
		buf.WriteString("- Rule file does not exist\n")
	}
	buf.WriteString("\n")
}

func writeRuleGuidance(buf *strings.Builder, state RuleState, ruleID, scope string, ruleSystemContent string) {
	buf.WriteString("## GUIDANCE\n\n")

	scopeSection := extractMarkdownSection(ruleSystemContent, "Rule Scopes")

	switch state {
	case RuleStableGlobal:
		buf.WriteString("Stable global rule — applies to every current-layer unit.\n")
		if scopeSection != "" {
			buf.WriteString("From `rule_system.md`:\n\n")
			buf.WriteString(scopeSection)
			buf.WriteString("\n\n")
		}
		buf.WriteString("To modify: create a candidate version through the rule governance process.\n")
		buf.WriteString("After changes, run `rule_sync` to coordinate affected units.\n")
		buf.WriteString("See governance flows in Core Truth above.\n\n")
	case RuleStableBound:
		buf.WriteString("Stable bound rule — applies only to units that list it in `rule_refs`.\n")
		if scopeSection != "" {
			buf.WriteString("From `rule_system.md`:\n\n")
			buf.WriteString(scopeSection)
			buf.WriteString("\n\n")
		}
		buf.WriteString("To modify: create a candidate version through the rule governance process.\n")
		buf.WriteString("After changes, run `rule_sync` to coordinate consumers.\n")
		buf.WriteString("See governance flows in Core Truth above.\n\n")
	case RuleCandidateGlobal, RuleCandidateNewGlobal:
		buf.WriteString("Active candidate round. The candidate version proposes changes to the stable rule.\n")
		buf.WriteString("After promotion, all current-layer units are affected.\n\n")
	case RuleCandidateBound, RuleCandidateNewBound:
		buf.WriteString("Active candidate round. The candidate version proposes changes to the stable rule.\n")
		buf.WriteString("After finalizing, run `rule_sync` to coordinate consumers.\n\n")
	case RuleUnregistered:
		buf.WriteString("This rule is not registered.\n")
		buf.WriteString("To create it, follow the governance flows in Core Truth above (`rule_system.md` → Governance Flows).\n")
		buf.WriteString("Start with `rule_new` at `framework/governance/rules/rule_new.md`.\n\n")
	}
}

func writeRuleBlocked(buf *strings.Builder, state RuleState, ruleID, scope, ruleSystemContent string) {
	buf.WriteString("## BLOCKED\n\n")

	scopeSection := extractMarkdownSection(ruleSystemContent, "Rule Scopes")
	if scopeSection != "" && strings.Contains(scopeSection, "Rule files must not") && state != RuleUnregistered {
		buf.WriteString("- Rule files must not store consumer lists (from `rule_system.md`)\n")
	}

	switch state {
	case RuleStableGlobal, RuleStableBound:
		buf.WriteString("- Editing rule files outside the rule governance process\n")
		buf.WriteString("- Editing consuming units' specs (their lifecycle owns them)\n")
		buf.WriteString("- Editing `_status.md`\n")
	case RuleCandidateGlobal, RuleCandidateBound, RuleCandidateNewGlobal, RuleCandidateNewBound:
		buf.WriteString("- Editing consuming units' specs\n")
		buf.WriteString("- Promoting without `rule_sync` coordination\n")
		buf.WriteString("- Editing `_status.md` directly\n")
	case RuleUnregistered:
		buf.WriteString("- Using the rule before it is registered\n")
		buf.WriteString("- Creating rule files outside the governance process\n")
	}
	buf.WriteString("\n")
}

func writeRuleClose(buf *strings.Builder, state RuleState, ruleID, ruleSystemContent string) {
	buf.WriteString("## CLOSE\n\n")

	governanceFlows := extractMarkdownSection(ruleSystemContent, "Governance Flows")
	hasGovFlows := governanceFlows != ""

	switch state {
	case RuleStableGlobal, RuleStableBound:
		if hasGovFlows {
			buf.WriteString("Rule operations close through their own governance flows")
			buf.WriteString(" (see Governance Flows in Core Truth).\n")
			buf.WriteString("Not via unit command close.\n\n")
		} else {
			buf.WriteString("Rule governance flows close through their own procedures. Not via unit command close.\n\n")
		}
	case RuleCandidateGlobal, RuleCandidateBound, RuleCandidateNewGlobal, RuleCandidateNewBound:
		buf.WriteString("Rule operations close through their own governance flows")
		if hasGovFlows {
			buf.WriteString(" (see Governance Flows in Core Truth).\n")
		} else {
			buf.WriteString(".\n")
		}
		buf.WriteString("After finalizing, run `rule_sync` to coordinate consumers.\n\n")
	case RuleUnregistered:
		buf.WriteString("Rule is unregistered. Register first through `rule_new` before any close operations.\n\n")
	}
}

// ============================================================================
// helpers
// ============================================================================

func readAdoptionMode(repoRoot string) string {
	adoptionPath := filepath.Join(repoRoot, filepath.FromSlash("docs/specs/_adoption_mode.txt"))
	data, err := os.ReadFile(adoptionPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func readSpecAndIntent(repoRoot, unitName string) (string, string) {
	candidatePaths := []string{
		filepath.Join(repoRoot, filepath.FromSlash("docs/specs/units/candidate/c_unit_"+unitName+".md")),
		filepath.Join(repoRoot, filepath.FromSlash("docs/specs/units/stable/s_unit_"+unitName+".md")),
	}
	for _, p := range candidatePaths {
		data, err := os.ReadFile(p)
		if err == nil {
			content := string(data)
			return extractFrontmatter(content, "candidate_intent"), content
		}
	}
	return "", ""
}

func readActiveSpecContent(repoRoot, unitName string) string {
	candidatePaths := []string{
		filepath.Join(repoRoot, filepath.FromSlash("docs/specs/units/candidate/c_unit_"+unitName+".md")),
		filepath.Join(repoRoot, filepath.FromSlash("docs/specs/units/stable/s_unit_"+unitName+".md")),
	}
	for _, p := range candidatePaths {
		data, err := os.ReadFile(p)
		if err == nil {
			return string(data)
		}
	}
	return ""
}

func readCheckDate(repoRoot, unitName string) string {
	checkResultPath := filepath.Join(repoRoot, "docs/specs/_check_result/unit", unitName+".md")
	data, err := os.ReadFile(checkResultPath)
	if err != nil {
		return ""
	}
	return extractFrontmatter(string(data), "created_at")
}

func findAffectedUnits(repoRoot, ruleID, scope string) []AffectedUnitRow {
	statuses, err := statusfile.LoadObjectStatuses(repoRoot)
	if err != nil {
		return nil
	}
	var affected []AffectedUnitRow
	for _, st := range statuses {
		if scope == "global" {
			affected = append(affected, AffectedUnitRow{
				Name:        st.Object,
				Layer:       st.ActiveLayer,
				NextCommand: st.NextCommand,
			})
			continue
		}
		// Bound rule: check if unit references this rule
		unitSpecPath := filepath.Join(repoRoot, filepath.FromSlash("docs/specs/units/candidate/c_unit_"+st.Object+".md"))
		data, err := os.ReadFile(unitSpecPath)
		if err != nil {
			unitSpecPath = filepath.Join(repoRoot, filepath.FromSlash("docs/specs/units/stable/s_unit_"+st.Object+".md"))
			data, err = os.ReadFile(unitSpecPath)
		}
		if err != nil {
			continue
		}
		refs := parseNamedRefs(string(data), "rule_refs")
		if slices.Contains(refs, "s_"+ruleID) || slices.Contains(refs, "c_"+ruleID) ||
			slices.Contains(refs, ruleID) {
			affected = append(affected, AffectedUnitRow{
				Name:        st.Object,
				Layer:       st.ActiveLayer,
				NextCommand: st.NextCommand,
			})
		}
	}
	return affected
}

// ============================================================================
// frontmatter helpers
// ============================================================================

func extractFrontmatter(content, field string) string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	inFM := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			inFM = !inFM
			continue
		}
		if !inFM {
			continue
		}
		if strings.HasPrefix(trimmed, field+":") {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, field+":"))
			val = strings.Trim(val, "\"`'")
			return val
		}
	}
	return ""
}

func parseNamedRefs(content, field string) []string {
	if content == "" {
		return nil
	}
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	var result []string
	seen := map[string]bool{}
	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, field+":") {
			continue
		}
		right := strings.TrimSpace(strings.TrimPrefix(trimmed, field+":"))
		if right == "none" || right == "`none`" {
			continue
		}
		if right != "" {
			item := strings.Trim(strings.TrimSpace(right), "`\"'")
			if item != "" && !seen[item] {
				result = append(result, item)
				seen[item] = true
			}
			continue
		}
		for next := idx + 1; next < len(lines); next++ {
			nextTrimmed := strings.TrimSpace(lines[next])
			if nextTrimmed == "" {
				continue
			}
			if !strings.HasPrefix(nextTrimmed, "- ") {
				break
			}
			item := strings.TrimSpace(strings.TrimPrefix(nextTrimmed, "- "))
			item = strings.Trim(item, "`\"'")
			if item != "" && !seen[item] {
				result = append(result, item)
				seen[item] = true
			}
		}
	}
	return result
}

func resolveRuleRefPath(ref string) string {
	ref = strings.Trim(ref, "`")
	prefix, _, _ := strings.Cut(ref, "@")
	switch {
	case strings.HasPrefix(prefix, "s_"):
		return "docs/specs/rules/stable/" + prefix + ".md"
	case strings.HasPrefix(prefix, "c_"):
		return "docs/specs/rules/candidate/" + prefix + ".md"
	}
	return ""
}

func resolveUnitRefPath(ref string) string {
	ref = strings.Trim(ref, "`")
	prefix, _, _ := strings.Cut(ref, "@")
	switch {
	case strings.HasPrefix(prefix, "s_unit_"):
		return "docs/specs/units/stable/" + prefix + ".md"
	case strings.HasPrefix(prefix, "c_unit_"):
		return "docs/specs/units/candidate/" + prefix + ".md"
	}
	return ""
}

func resolveImplPaths(repoRoot, unitName string) string {
	mappingPath := filepath.Join(repoRoot, filepath.FromSlash("docs/specs/repository_mapping.md"))
	data, err := os.ReadFile(mappingPath)
	if err != nil {
		return "src"
	}
	content := string(data)
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")

	inRegistry := false
	tableStarted := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## 2.") && strings.Contains(trimmed, "Object Registry") {
			inRegistry = true
			continue
		}
		if !inRegistry {
			continue
		}
		if strings.HasPrefix(trimmed, "## ") && !strings.Contains(trimmed, "Object Registry") {
			break
		}
		if !tableStarted {
			if strings.Contains(trimmed, "kind") && strings.Contains(trimmed, "implementation_paths") {
				tableStarted = true
			}
			continue
		}
		if !strings.HasPrefix(trimmed, "|") {
			continue
		}
		cells := parseMarkdownRow(trimmed)
		if len(cells) >= 5 {
			kind := strings.Trim(strings.TrimSpace(cells[0]), "`")
			id := strings.Trim(strings.TrimSpace(cells[1]), "`")
			if kind == "unit" && id == unitName {
				implPaths := strings.Trim(strings.TrimSpace(cells[3]), "`")
				implPaths = strings.TrimSpace(implPaths)
				if implPaths != "" && strings.ToLower(implPaths) != "none" {
					return implPaths
				}
				return "src"
			}
		}
	}
	return "src"
}

func parseMarkdownRow(line string) []string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "|") {
		return nil
	}
	line = strings.TrimPrefix(line, "|")
	if strings.HasSuffix(line, "|") {
		line = strings.TrimSuffix(line, "|")
	}
	cells := strings.Split(line, "|")
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	return cells
}
