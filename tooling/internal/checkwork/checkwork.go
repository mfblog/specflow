package checkwork

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specflowlayout"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/unitappendix"
)

const (
	statusInProgress        = "in_progress"
	statusBlockedOnFinding  = "blocked_on_finding"
	statusReadyForFinal     = "ready_for_final"
	statusClosedPass        = "closed_pass"
	statusClosedBlocked     = "closed_blocked"
	statusClosedFixRequired = "closed_fix_required"

	itemPending       = "pending"
	itemClear         = "clear"
	itemIncomplete    = "incomplete"
	itemBlocked       = "blocked"
	itemStale         = "stale"
	itemNotApplicable = "not_applicable"

	timestampLayout = "2006-01-02T15:04:05Z"
)

var checklistColumns = []string{
	"item_id",
	"status",
	"question",
	"input_files",
	"input_fingerprint",
	"finding_refs",
	"result_summary",
}

type InitResult struct {
	File         string
	Created      bool
	Reused       bool
	DeletedFiles []DeletedWorkStateFile
}

type DeletedWorkStateFile struct {
	File   string
	Reason string
}

type ValidationResult struct {
	File        string
	Valid       bool
	Diagnostics []string
}

type RefreshResult struct {
	File             string
	StaleItems       []string
	MissingInputs    []string
	ChangedItems     []string
	LastUpdatedAtUTC string
}

type TouchResult struct {
	File             string
	LastUpdatedAtUTC string
}

type fixedWorkStateFile struct {
	LastUpdated time.Time
	Reason      string
}

type workState struct {
	Fields    map[string]string
	Checklist []checklistItem
	Findings  string
	Resume    string
}

type checklistItem struct {
	ItemID           string
	Status           string
	Question         string
	InputFiles       []string
	InputFingerprint string
	FindingRefs      string
	ResultSummary    string
}

type checklistDefinition struct {
	ID         string
	Question   string
	InputFiles func(inputContext) []string
}

type inputContext struct {
	TruthFile              string
	AppendixFiles          []string
	DependencyUnitFiles    []string
	DependencyRuleFiles    []string
	RepositoryMappingFiles []string
	GlobalRuleFiles        []string
	FrameworkFiles         []string
	AcceptanceRuleFiles    []string
	HandoffRuleFiles       []string
}

type truthInfo struct {
	Layer       string
	FileRef     string
	VersionRef  string
	Fingerprint string
	Context     inputContext
}

func WorkFilePath(repoRoot, objectType, object string) (string, error) {
	if strings.TrimSpace(objectType) != "unit" {
		return "", fmt.Errorf("object type %q is not supported; only unit is supported", objectType)
	}
	object = strings.TrimSpace(object)
	if object == "" {
		return "", fmt.Errorf("object is required")
	}
	return filepath.Join(repoRoot, filepath.FromSlash(fmt.Sprintf("docs/specs/_check_work/%s/%s.md", objectType, object))), nil
}

func Init(repoRoot, objectType, object string, now time.Time) (InitResult, error) {
	objectType = strings.TrimSpace(objectType)
	object = strings.TrimSpace(object)
	now = now.UTC()
	file, err := WorkFilePath(repoRoot, objectType, object)
	if err != nil {
		return InitResult{}, err
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return InitResult{}, err
	}

	existing, err := inspectWorkState(repoRoot, objectType, object, file, now)
	if err != nil {
		return InitResult{}, err
	}
	result := InitResult{File: file}
	if existing != nil {
		if existing.Reason == "" {
			age := now.Sub(existing.LastUpdated)
			switch {
			case age <= 2*time.Hour:
				return InitResult{File: file, Reused: true}, nil
			case age <= 7*24*time.Hour:
				return InitResult{}, fmt.Errorf("open check checklist requires manual reuse decision before check-work-init can continue: %s last_updated_at=%s age=%s", file, formatUTC(existing.LastUpdated), age.Round(time.Second))
			default:
				existing.Reason = "expired_over_7_days"
			}
		}
		if existing.Reason != "missing" {
			if err := os.Remove(file); err != nil {
				return InitResult{}, err
			}
			result.DeletedFiles = append(result.DeletedFiles, DeletedWorkStateFile{File: file, Reason: existing.Reason})
		}
	}

	truth, err := collectTruthInfo(repoRoot, objectType, object)
	if err != nil {
		return InitResult{}, err
	}
	workID := fmt.Sprintf("%s-unit_check-%s", now.Format("20060102-150405"), object)
	state := workState{
		Fields: map[string]string{
			"work_flow":         "unit_check",
			"work_id":           workID,
			"object_type":       objectType,
			"object_ref":        object,
			"status":            statusInProgress,
			"created_at":        formatUTC(now),
			"last_updated_at":   formatUTC(now),
			"truth_layer_ref":   truth.Layer,
			"truth_file_ref":    truth.FileRef,
			"truth_version_ref": truth.VersionRef,
			"truth_fingerprint": truth.Fingerprint,
			"checklist_table":   "present",
			"finding_refs":      "none",
			"blocked_reason":    "none",
			"resume_next_step":  "review checklist item goal_and_responsibility",
		},
		Findings: "none",
		Resume:   "none",
	}
	state.Checklist, err = buildChecklistItems(repoRoot, truth.Context)
	if err != nil {
		return InitResult{}, err
	}
	if err := os.WriteFile(file, []byte(renderState(state)), 0o644); err != nil {
		return InitResult{}, err
	}
	result.Created = true
	return result, nil
}

func Validate(repoRoot, objectType, object string, now time.Time) ValidationResult {
	file, err := WorkFilePath(repoRoot, objectType, object)
	result := ValidationResult{File: file}
	if err != nil {
		result.Diagnostics = append(result.Diagnostics, err.Error())
		return result
	}
	state, err := parseFile(file)
	if err != nil {
		result.Diagnostics = append(result.Diagnostics, err.Error())
		return result
	}
	result.Diagnostics = validateState(repoRoot, objectType, object, state, now.UTC(), false)
	result.Valid = len(result.Diagnostics) == 0
	return result
}

func Refresh(repoRoot, objectType, object string, now time.Time) (RefreshResult, error) {
	objectType = strings.TrimSpace(objectType)
	object = strings.TrimSpace(object)
	now = now.UTC()
	file, err := WorkFilePath(repoRoot, objectType, object)
	if err != nil {
		return RefreshResult{}, err
	}
	state, err := parseFile(file)
	if err != nil {
		return RefreshResult{}, err
	}
	if diagnostics := validateState(repoRoot, objectType, object, state, now, true); len(diagnostics) > 0 {
		return RefreshResult{}, fmt.Errorf("check checklist validation failed: %s", strings.Join(diagnostics, "; "))
	}

	truth, err := collectTruthInfo(repoRoot, objectType, object)
	if err != nil {
		return RefreshResult{}, err
	}
	state.Fields["truth_layer_ref"] = truth.Layer
	state.Fields["truth_file_ref"] = truth.FileRef
	state.Fields["truth_version_ref"] = truth.VersionRef
	state.Fields["truth_fingerprint"] = truth.Fingerprint

	definitions := checklistDefinitions()
	definitionsByID := map[string]checklistDefinition{}
	for _, definition := range definitions {
		definitionsByID[definition.ID] = definition
	}
	inputFilesChanged := map[string]bool{}
	for i := range state.Checklist {
		definition, ok := definitionsByID[state.Checklist[i].ItemID]
		if !ok {
			continue
		}
		currentInputFiles := union(definition.InputFiles(truth.Context))
		if !sameStringList(state.Checklist[i].InputFiles, currentInputFiles) {
			inputFilesChanged[state.Checklist[i].ItemID] = true
		}
		state.Checklist[i].Question = definition.Question
		state.Checklist[i].InputFiles = currentInputFiles
	}

	result := RefreshResult{File: file, LastUpdatedAtUTC: formatUTC(now)}
	for i := range state.Checklist {
		fingerprint, missing, err := computeFingerprint(repoRoot, state.Checklist[i].InputFiles)
		if err != nil {
			return RefreshResult{}, err
		}
		changed := len(missing) == 0 && (fingerprint != state.Checklist[i].InputFingerprint || inputFilesChanged[state.Checklist[i].ItemID])
		if changed {
			result.ChangedItems = append(result.ChangedItems, state.Checklist[i].ItemID)
		}
		for _, missingPath := range missing {
			result.MissingInputs = append(result.MissingInputs, state.Checklist[i].ItemID+":"+missingPath)
		}
		if len(missing) == 0 {
			state.Checklist[i].InputFingerprint = fingerprint
		}
		if state.Checklist[i].Status == itemClear && (changed || len(missing) > 0) {
			state.Checklist[i].Status = itemStale
			state.Checklist[i].ResultSummary = "stale: input changed"
			if len(missing) > 0 {
				state.Checklist[i].ResultSummary = "stale: input missing"
			}
			result.StaleItems = append(result.StaleItems, state.Checklist[i].ItemID)
		}
	}

	state.Fields["last_updated_at"] = formatUTC(now)
	if err := os.WriteFile(file, []byte(renderState(state)), 0o644); err != nil {
		return RefreshResult{}, err
	}
	sort.Strings(result.StaleItems)
	sort.Strings(result.ChangedItems)
	sort.Strings(result.MissingInputs)
	return result, nil
}

func Touch(repoRoot, objectType, object string, now time.Time) (TouchResult, error) {
	now = now.UTC()
	file, err := WorkFilePath(repoRoot, objectType, object)
	if err != nil {
		return TouchResult{}, err
	}
	state, err := parseFile(file)
	if err != nil {
		return TouchResult{}, err
	}
	if diagnostics := validateState(repoRoot, objectType, object, state, now, false); len(diagnostics) > 0 {
		return TouchResult{}, fmt.Errorf("check checklist validation failed: %s", strings.Join(diagnostics, "; "))
	}
	state.Fields["last_updated_at"] = formatUTC(now)
	if err := os.WriteFile(file, []byte(renderState(state)), 0o644); err != nil {
		return TouchResult{}, err
	}
	return TouchResult{File: file, LastUpdatedAtUTC: formatUTC(now)}, nil
}

func inspectWorkState(repoRoot, objectType, object, file string, now time.Time) (*fixedWorkStateFile, error) {
	state, err := parseFile(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &fixedWorkStateFile{Reason: "missing"}, nil
		}
		return &fixedWorkStateFile{Reason: "invalid_work_state"}, nil
	}
	if isClosedStatus(state.Fields["status"]) {
		return &fixedWorkStateFile{Reason: "closed_work_state"}, nil
	}
	if diagnostics := validateState(repoRoot, objectType, object, state, now, true); len(diagnostics) > 0 {
		return &fixedWorkStateFile{Reason: "invalid_work_state"}, nil
	}
	lastUpdated, err := parseTimestamp(state.Fields["last_updated_at"])
	if err != nil {
		return &fixedWorkStateFile{Reason: "invalid_work_state"}, nil
	}
	return &fixedWorkStateFile{LastUpdated: lastUpdated}, nil
}

func collectTruthInfo(repoRoot, objectType, object string) (truthInfo, error) {
	if objectType != "unit" {
		return truthInfo{}, fmt.Errorf("object type %q is not supported; only unit is supported", objectType)
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, objectType, object)
	if err != nil {
		return truthInfo{}, err
	}
	if status.Candidate != "yes" {
		return truthInfo{}, fmt.Errorf("unit %q has no candidate truth", object)
	}
	mainSpecRef, err := specpaths.ObjectMainSpecFileRef(objectType, "candidate", object)
	if err != nil {
		return truthInfo{}, err
	}
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)))
	if err != nil {
		return truthInfo{}, fmt.Errorf("read %s: %w", mainSpecRef, err)
	}
	frontmatter, _, err := parseFrontmatter(string(content))
	if err != nil {
		return truthInfo{}, fmt.Errorf("%s: %w", mainSpecRef, err)
	}
	version := strings.TrimSpace(frontmatter["version"])
	if version == "" {
		return truthInfo{}, fmt.Errorf("%s: missing frontmatter.version", mainSpecRef)
	}

	appendices, err := collectAppendixFiles(repoRoot, objectType, object, mainSpecRef, frontmatter)
	if err != nil {
		return truthInfo{}, err
	}
	unitFiles := resolveUnitRefFiles(parseNamedRefs(string(content), "unit_refs"))
	ruleFiles := resolveRuleRefFiles(parseNamedRefs(string(content), "rule_refs"))
	globalRules, err := globExisting(repoRoot, "docs/specs/rules/stable/s_g_rule_*.md")
	if err != nil {
		return truthInfo{}, err
	}
	repositoryMapping := existingFiles(repoRoot, []string{specpaths.RepositoryMappingFileRef})
	layout, err := specflowlayout.Resolve(repoRoot)
	if err != nil {
		return truthInfo{}, err
	}
	frameworkPath := func(relPath string) string {
		return specflowlayout.Relative(layout.FrameworkRoot, relPath)
	}
	frameworkFiles := existingFiles(repoRoot, []string{
		frameworkPath("core/object_model.md"),
		frameworkPath("core/repository_mapping.md"),
		frameworkPath("lifecycle/unit_check.md"),
		frameworkPath("process_snapshot_contract.md"),
		frameworkPath("candidate_handoff_contract.md"),
		frameworkPath("candidate_intent_policy.md"),
		frameworkPath("operations/implementation_change.md"),
	})
	return truthInfo{
		Layer:       "candidate",
		FileRef:     mainSpecRef,
		VersionRef:  fmt.Sprintf("%s@%s", strings.TrimSuffix(filepath.Base(mainSpecRef), ".md"), version),
		Fingerprint: hashNormalizedText(string(content)),
		Context: inputContext{
			TruthFile:              mainSpecRef,
			AppendixFiles:          appendices,
			DependencyUnitFiles:    unitFiles,
			DependencyRuleFiles:    ruleFiles,
			RepositoryMappingFiles: repositoryMapping,
			GlobalRuleFiles:        globalRules,
			FrameworkFiles:         frameworkFiles,
			AcceptanceRuleFiles: existingFiles(repoRoot, []string{
				frameworkPath("core/object_model.md"),
				frameworkPath("lifecycle/unit_check.md"),
			}),
			HandoffRuleFiles: existingFiles(repoRoot, []string{
				frameworkPath("candidate_handoff_contract.md"),
				frameworkPath("process_snapshot_contract.md"),
				frameworkPath("lifecycle/unit_plan.md"),
			}),
		},
	}, nil
}

func checklistDefinitions() []checklistDefinition {
	return []checklistDefinition{
		{
			ID:       "goal_and_responsibility",
			Question: "Is the unit goal, responsibility, scope, non-goals, and owner fit clear enough to plan from.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.AppendixFiles, ctx.FrameworkFiles)
			},
		},
		{
			ID:       "dependency_truth_surface",
			Question: "Are unit_refs, rule_refs, stable global rules, appendices, and repository mapping explicit and consistent.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.AppendixFiles, ctx.DependencyUnitFiles, ctx.DependencyRuleFiles, ctx.GlobalRuleFiles, ctx.RepositoryMappingFiles, ctx.FrameworkFiles)
			},
		},
		{
			ID:       "main_flow_and_state",
			Question: "Are the normal flow, state changes, ordering, lifecycle effects, and important transitions clear.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.AppendixFiles, ctx.FrameworkFiles)
			},
		},
		{
			ID:       "boundary_and_protocol",
			Question: "Are public contracts, ports, adapters, stores, events, trace sources, and owners explicit.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.DependencyUnitFiles, ctx.DependencyRuleFiles, ctx.RepositoryMappingFiles, ctx.FrameworkFiles)
			},
		},
		{
			ID:       "data_artifact_and_output",
			Question: "Are produced artifacts, evidence, reports, traces, persistence records, outputs, and consumers clear.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.AppendixFiles, ctx.FrameworkFiles)
			},
		},
		{
			ID:       "error_edge_and_gap",
			Question: "Are important error states, edge cases, missing-dependency behavior, diagnostic gaps, and failure owners clear.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.AppendixFiles, ctx.DependencyUnitFiles, ctx.DependencyRuleFiles, ctx.FrameworkFiles)
			},
		},
		{
			ID:       "acceptance_and_testability",
			Question: "Do acceptance items name test surfaces, proof methods, runnable status, and pass conditions.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.AppendixFiles, ctx.AcceptanceRuleFiles)
			},
		},
		{
			ID:       "implementation_handoff",
			Question: "Can unit_plan create an implementation handoff without inventing missing design, adapter, output, or test choices.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.AppendixFiles, ctx.HandoffRuleFiles)
			},
		},
	}
}

func buildChecklistItems(repoRoot string, ctx inputContext) ([]checklistItem, error) {
	result := []checklistItem{}
	for _, definition := range checklistDefinitions() {
		inputFiles := union(definition.InputFiles(ctx))
		fingerprint, missing, err := computeFingerprint(repoRoot, inputFiles)
		if err != nil {
			return nil, err
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf("checklist item %s has missing input files: %s", definition.ID, strings.Join(missing, ", "))
		}
		result = append(result, checklistItem{
			ItemID:           definition.ID,
			Status:           itemPending,
			Question:         definition.Question,
			InputFiles:       inputFiles,
			InputFingerprint: fingerprint,
			FindingRefs:      "none",
			ResultSummary:    "pending",
		})
	}
	return result, nil
}

func validateState(repoRoot, objectType, object string, state workState, now time.Time, requireOpen bool) []string {
	diagnostics := []string{}
	requiredFields := []string{
		"work_flow",
		"work_id",
		"object_type",
		"object_ref",
		"status",
		"created_at",
		"last_updated_at",
		"truth_layer_ref",
		"truth_file_ref",
		"truth_version_ref",
		"truth_fingerprint",
		"checklist_table",
		"finding_refs",
		"blocked_reason",
		"resume_next_step",
	}
	for _, field := range requiredFields {
		if strings.TrimSpace(state.Fields[field]) == "" {
			diagnostics = append(diagnostics, "missing work field: "+field)
		}
	}
	if state.Fields["work_flow"] != "unit_check" {
		diagnostics = append(diagnostics, "work_flow must be unit_check")
	}
	if state.Fields["object_type"] != objectType {
		diagnostics = append(diagnostics, "object_type must be "+objectType)
	}
	if state.Fields["object_ref"] != object {
		diagnostics = append(diagnostics, "object_ref must be "+object)
	}
	if state.Fields["truth_layer_ref"] != "candidate" {
		diagnostics = append(diagnostics, "truth_layer_ref must be candidate")
	}
	status := strings.TrimSpace(state.Fields["status"])
	if !isWorkStatus(status) {
		diagnostics = append(diagnostics, "invalid work status: "+status)
	} else if requireOpen && !isOpenStatus(status) {
		diagnostics = append(diagnostics, "closed check checklist files cannot be reused")
	}
	for _, field := range []string{"created_at", "last_updated_at"} {
		if _, err := parseTimestamp(state.Fields[field]); err != nil {
			diagnostics = append(diagnostics, field+" must use UTC format YYYY-MM-DDTHH:MM:SSZ")
		}
	}
	if lastUpdated, err := parseTimestamp(state.Fields["last_updated_at"]); err == nil && lastUpdated.After(now) {
		diagnostics = append(diagnostics, "last_updated_at must not be later than current UTC time")
	}
	expected := checklistIDSet()
	seen := map[string]bool{}
	for _, item := range state.Checklist {
		if seen[item.ItemID] {
			diagnostics = append(diagnostics, "duplicate checklist item: "+item.ItemID)
		}
		seen[item.ItemID] = true
		if !expected[item.ItemID] {
			diagnostics = append(diagnostics, "unexpected checklist item: "+item.ItemID)
		}
		diagnostics = append(diagnostics, validateChecklistItem(repoRoot, item)...)
	}
	for id := range expected {
		if !seen[id] {
			diagnostics = append(diagnostics, "missing checklist item: "+id)
		}
	}
	return diagnostics
}

func validateChecklistItem(repoRoot string, item checklistItem) []string {
	diagnostics := []string{}
	for _, field := range []struct {
		name  string
		value string
	}{
		{"item_id", item.ItemID},
		{"status", item.Status},
		{"question", item.Question},
		{"input_fingerprint", item.InputFingerprint},
		{"finding_refs", item.FindingRefs},
		{"result_summary", item.ResultSummary},
	} {
		if strings.TrimSpace(field.value) == "" {
			diagnostics = append(diagnostics, field.name+" is required for checklist item: "+item.ItemID)
		}
	}
	if !isChecklistStatus(item.Status) {
		diagnostics = append(diagnostics, "invalid checklist status for "+item.ItemID+": "+item.Status)
	}
	if len(item.InputFiles) == 0 {
		diagnostics = append(diagnostics, "input_files is required for checklist item: "+item.ItemID)
	}
	for _, relPath := range item.InputFiles {
		if strings.TrimSpace(relPath) == "" || filepath.IsAbs(relPath) || strings.Contains(relPath, "\\") {
			diagnostics = append(diagnostics, "input_files must use repository-relative slash paths: "+item.ItemID)
			continue
		}
		if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(relPath))); err != nil && !os.IsNotExist(err) {
			diagnostics = append(diagnostics, "cannot inspect input file for "+item.ItemID+": "+relPath)
		}
	}
	return diagnostics
}

func parseFile(file string) (workState, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return workState{}, err
	}
	text := string(raw)
	state := workState{Fields: map[string]string{}}
	sections := splitSections(text)
	runSection, ok := sections["Run State"]
	if !ok {
		return state, fmt.Errorf("missing section: Run State")
	}
	fields, err := parseKeyValueTable(runSection)
	if err != nil {
		return state, err
	}
	state.Fields = fields
	checklistSection, ok := sections["Checklist"]
	if !ok {
		return state, fmt.Errorf("missing section: Checklist")
	}
	state.Checklist, err = parseChecklistTable(checklistSection)
	if err != nil {
		return state, fmt.Errorf("checklist table: %w", err)
	}
	state.Findings = strings.TrimSpace(sections["Findings"])
	if state.Findings == "" {
		state.Findings = "none"
	}
	state.Resume = strings.TrimSpace(sections["Resume"])
	if state.Resume == "" {
		state.Resume = "none"
	}
	return state, nil
}

func splitSections(text string) map[string]string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	sections := map[string]string{}
	current := ""
	var body []string
	flush := func() {
		if current != "" {
			sections[current] = strings.TrimSpace(strings.Join(body, "\n"))
		}
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			flush()
			current = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			body = []string{}
			continue
		}
		if current != "" {
			body = append(body, line)
		}
	}
	flush()
	return sections
}

func parseKeyValueTable(section string) (map[string]string, error) {
	rows := parseMarkdownRows(section)
	result := map[string]string{}
	for _, row := range rows {
		if len(row) != 2 || row[0] == "field" {
			continue
		}
		result[row[0]] = row[1]
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("run field table is empty")
	}
	return result, nil
}

func parseChecklistTable(section string) ([]checklistItem, error) {
	rows := parseMarkdownRows(section)
	if len(rows) == 0 {
		return nil, fmt.Errorf("checklist table is empty")
	}
	header := rows[0]
	if len(header) != len(checklistColumns) {
		return nil, fmt.Errorf("checklist table header has %d columns, want %d", len(header), len(checklistColumns))
	}
	for i, column := range checklistColumns {
		if header[i] != column {
			return nil, fmt.Errorf("checklist table column %d is %q, want %q", i+1, header[i], column)
		}
	}
	result := []checklistItem{}
	for _, row := range rows[1:] {
		if len(row) != len(checklistColumns) {
			return nil, fmt.Errorf("checklist row has %d columns, want %d", len(row), len(checklistColumns))
		}
		result = append(result, checklistItem{
			ItemID:           row[0],
			Status:           row[1],
			Question:         row[2],
			InputFiles:       parseList(row[3]),
			InputFingerprint: row[4],
			FindingRefs:      row[5],
			ResultSummary:    row[6],
		})
	}
	return result, nil
}

func parseMarkdownRows(section string) [][]string {
	rows := [][]string{}
	for _, rawLine := range strings.Split(section, "\n") {
		line := strings.TrimSpace(rawLine)
		if !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		row := make([]string, 0, len(cells))
		separator := true
		for _, cell := range cells {
			trimmed := strings.TrimSpace(cell)
			row = append(row, trimmed)
			if strings.Trim(trimmed, "-: ") != "" {
				separator = false
			}
		}
		if separator {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

func renderState(state workState) string {
	var b strings.Builder
	b.WriteString("# Unit Check Checklist\n\n")
	b.WriteString("## Run State\n\n")
	b.WriteString("| field | value |\n")
	b.WriteString("|---|---|\n")
	for _, field := range []string{"work_flow", "work_id", "object_type", "object_ref", "status", "created_at", "last_updated_at", "truth_layer_ref", "truth_file_ref", "truth_version_ref", "truth_fingerprint", "checklist_table", "finding_refs", "blocked_reason", "resume_next_step"} {
		b.WriteString(fmt.Sprintf("| %s | %s |\n", field, cleanCell(state.Fields[field])))
	}
	b.WriteString("\n## Checklist\n\n")
	renderChecklistTable(&b, state.Checklist)
	b.WriteString("\n## Findings\n\n")
	b.WriteString(defaultText(state.Findings))
	b.WriteString("\n\n## Resume\n\n")
	b.WriteString(defaultText(state.Resume))
	b.WriteString("\n")
	return b.String()
}

func renderChecklistTable(b *strings.Builder, items []checklistItem) {
	b.WriteString("| " + strings.Join(checklistColumns, " | ") + " |\n")
	b.WriteString("|" + strings.Repeat("---|", len(checklistColumns)) + "\n")
	for _, item := range items {
		values := []string{
			item.ItemID,
			item.Status,
			item.Question,
			joinList(item.InputFiles),
			item.InputFingerprint,
			item.FindingRefs,
			item.ResultSummary,
		}
		for i := range values {
			values[i] = cleanCell(values[i])
		}
		b.WriteString("| " + strings.Join(values, " | ") + " |\n")
	}
}

func computeFingerprint(repoRoot string, inputFiles []string) (string, []string, error) {
	files := union(inputFiles)
	payload := strings.Builder{}
	missing := []string{}
	for _, relPath := range files {
		content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relPath)))
		if err != nil {
			if !os.IsNotExist(err) {
				return "", nil, err
			}
			missing = append(missing, relPath)
			continue
		}
		sum := sha256.Sum256([]byte(normalizeText(string(content))))
		payload.WriteString("file_ref: ")
		payload.WriteString(relPath)
		payload.WriteString("\nfile_sha256: ")
		payload.WriteString(hex.EncodeToString(sum[:]))
		payload.WriteString("\n\n")
	}
	if len(missing) > 0 {
		return "", missing, nil
	}
	sum := sha256.Sum256([]byte(payload.String()))
	return hex.EncodeToString(sum[:]), missing, nil
}

func collectAppendixFiles(repoRoot, objectType, object, mainSpecRef string, frontmatter map[string]string) ([]string, error) {
	mainDir := filepath.Dir(filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)))
	entries, err := unitappendix.Scan(repoRoot, objectType, object, "candidate")
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(entries))
	entrySet := map[string]bool{}
	for _, entry := range entries {
		result = append(result, entry.FileRef)
		entrySet[entry.FileRef] = true
	}
	if evidenceRef := strings.TrimSpace(frontmatter["evidence_appendix_ref"]); evidenceRef != "" && evidenceRef != "none" {
		relPath, err := resolveAppendixRef(repoRoot, mainDir, evidenceRef)
		if err != nil {
			return nil, err
		}
		if !entrySet[relPath] {
			return nil, fmt.Errorf("%s: evidence appendix ref %s is not a current candidate appendix for unit %s", mainSpecRef, relPath, object)
		}
	}
	return union(result), nil
}

func resolveAppendixRef(repoRoot, mainDir, ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" || strings.HasPrefix(ref, "/") || strings.Contains(ref, "://") {
		return "", fmt.Errorf("invalid appendix ref %q", ref)
	}
	var absPath string
	if strings.HasPrefix(filepath.ToSlash(ref), "docs/") {
		absPath = filepath.Join(repoRoot, filepath.FromSlash(ref))
	} else {
		absPath = filepath.Join(mainDir, filepath.FromSlash(ref))
	}
	relPath, err := filepath.Rel(repoRoot, filepath.Clean(absPath))
	if err != nil {
		return "", err
	}
	relPath = filepath.ToSlash(relPath)
	if strings.HasPrefix(relPath, "../") || relPath == ".." {
		return "", fmt.Errorf("appendix ref %q resolves outside repository", ref)
	}
	return relPath, nil
}

func parseFrontmatter(content string) (map[string]string, string, error) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, "", fmt.Errorf("missing frontmatter start marker")
	}
	endIdx := -1
	for idx := 1; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "---" {
			endIdx = idx
			break
		}
	}
	if endIdx == -1 {
		return nil, "", fmt.Errorf("missing frontmatter end marker")
	}
	result := map[string]string{}
	for _, line := range lines[1:endIdx] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return result, strings.Join(lines[endIdx+1:], "\n"), nil
}

func parseNamedRefs(content, fieldName string) []string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	result := []string{}
	seen := map[string]bool{}
	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		right, matched := parseNamedFieldLine(trimmed, fieldName)
		if !matched {
			continue
		}
		if right == "`none`" || right == "none" {
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
	sort.Strings(result)
	return result
}

func parseNamedFieldLine(trimmed, fieldName string) (string, bool) {
	parts := strings.SplitN(trimmed, ":", 2)
	if len(parts) != 2 {
		return "", false
	}
	left := strings.ReplaceAll(strings.TrimSpace(parts[0]), "`", "")
	if idx := strings.Index(left, ". "); idx > 0 {
		allDigits := true
		for _, ch := range left[:idx] {
			if ch < '0' || ch > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			left = left[idx+2:]
		}
	}
	if strings.TrimSpace(left) != fieldName {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}

func resolveUnitRefFiles(refs []string) []string {
	result := []string{}
	for _, ref := range refs {
		prefix, _, ok := strings.Cut(strings.Trim(ref, "`"), "@")
		if !ok {
			continue
		}
		switch {
		case strings.HasPrefix(prefix, "s_unit_"):
			result = append(result, "docs/specs/units/stable/"+prefix+".md")
		case strings.HasPrefix(prefix, "c_unit_"):
			result = append(result, "docs/specs/units/candidate/"+prefix+".md")
		}
	}
	return union(result)
}

func resolveRuleRefFiles(refs []string) []string {
	result := []string{}
	for _, ref := range refs {
		prefix, _, ok := strings.Cut(strings.Trim(ref, "`"), "@")
		if !ok {
			continue
		}
		switch {
		case strings.HasPrefix(prefix, "s_"):
			result = append(result, "docs/specs/rules/stable/"+prefix+".md")
		case strings.HasPrefix(prefix, "c_"):
			result = append(result, "docs/specs/rules/candidate/"+prefix+".md")
		}
	}
	return union(result)
}

func globExisting(repoRoot, pattern string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(repoRoot, filepath.FromSlash(pattern)))
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, match := range matches {
		rel, err := filepath.Rel(repoRoot, match)
		if err != nil {
			return nil, err
		}
		result = append(result, filepath.ToSlash(rel))
	}
	return union(result), nil
}

func existingFiles(repoRoot string, relPaths []string) []string {
	result := []string{}
	for _, relPath := range relPaths {
		if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(relPath))); err == nil {
			result = append(result, relPath)
		}
	}
	return union(result)
}

func checklistIDSet() map[string]bool {
	result := map[string]bool{}
	for _, definition := range checklistDefinitions() {
		result[definition.ID] = true
	}
	return result
}

func isWorkStatus(status string) bool {
	switch status {
	case statusInProgress, statusBlockedOnFinding, statusReadyForFinal, statusClosedPass, statusClosedBlocked, statusClosedFixRequired:
		return true
	default:
		return false
	}
}

func isOpenStatus(status string) bool {
	switch status {
	case statusInProgress, statusBlockedOnFinding, statusReadyForFinal:
		return true
	default:
		return false
	}
}

func isClosedStatus(status string) bool {
	switch status {
	case statusClosedPass, statusClosedBlocked, statusClosedFixRequired:
		return true
	default:
		return false
	}
}

func isChecklistStatus(status string) bool {
	switch status {
	case itemPending, itemClear, itemIncomplete, itemBlocked, itemStale, itemNotApplicable:
		return true
	default:
		return false
	}
}

func parseTimestamp(value string) (time.Time, error) {
	return time.Parse(timestampLayout, value)
}

func formatUTC(value time.Time) string {
	return value.UTC().Format(timestampLayout)
}

func hashNormalizedText(content string) string {
	sum := sha256.Sum256([]byte(normalizeText(content)))
	return fmt.Sprintf("%x", sum)
}

func normalizeText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return text
}

func union(sets ...[]string) []string {
	seen := map[string]bool{}
	for _, set := range sets {
		for _, item := range set {
			item = strings.TrimSpace(filepath.ToSlash(item))
			if item == "" || item == "none" {
				continue
			}
			seen[item] = true
		}
	}
	result := make([]string, 0, len(seen))
	for item := range seen {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

func sameStringList(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func parseList(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" || value == "none" {
		return nil
	}
	items := strings.Split(value, ";")
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" && item != "none" {
			result = append(result, item)
		}
	}
	sort.Strings(result)
	return result
}

func joinList(values []string) string {
	values = union(values)
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, ";")
}

func cleanCell(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "|", "/")
	if value == "" {
		return "none"
	}
	return value
}

func defaultText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "none"
	}
	return value
}
