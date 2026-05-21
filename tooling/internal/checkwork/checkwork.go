package checkwork

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

const (
	statusInProgress        = "in_progress"
	statusBlockedOnFinding  = "blocked_on_finding"
	statusReadyForFinal     = "ready_for_final"
	statusClosedPass        = "closed_pass"
	statusClosedBlocked     = "closed_blocked"
	statusClosedFixRequired = "closed_fix_required"

	slicePending              = "pending"
	slicePassed               = "passed"
	sliceBlocked              = "blocked"
	sliceStale                = "stale"
	sliceSkippedNotApplicable = "skipped_not_applicable"

	timestampLayout = "2006-01-02T15:04:05Z"
)

var sliceColumns = []string{
	"slice_id",
	"slice_origin",
	"slice_type",
	"status",
	"review_question",
	"why_added",
	"parent_slice_id",
	"input_files",
	"input_fingerprint",
	"depends_on",
	"finding_refs",
	"result_summary",
	"exit_condition",
	"resume_next_step",
}

var markdownLinkPattern = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)

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
	StaleSlices      []string
	MissingInputs    []string
	ChangedSlices    []string
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
	Fields   map[string]string
	Baseline []sliceEntry
	Dynamic  []sliceEntry
	Findings string
	Resume   string
}

type sliceEntry struct {
	SliceID          string
	SliceOrigin      string
	SliceType        string
	Status           string
	ReviewQuestion   string
	WhyAdded         string
	ParentSliceID    string
	InputFiles       []string
	InputFingerprint string
	DependsOn        []string
	FindingRefs      string
	ResultSummary    string
	ExitCondition    string
	ResumeNextStep   string
}

type sliceDefinition struct {
	ID             string
	SliceType      string
	ReviewQuestion string
	InputFiles     func(inputContext) []string
	DependsOn      []string
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
				return InitResult{}, fmt.Errorf("open check-work file requires manual reuse decision before check-work-init can continue: %s last_updated_at=%s age=%s", file, formatUTC(existing.LastUpdated), age.Round(time.Second))
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
			"work_flow":            "unit_check",
			"work_id":              workID,
			"object_type":          objectType,
			"object_ref":           object,
			"status":               statusInProgress,
			"created_at":           formatUTC(now),
			"last_updated_at":      formatUTC(now),
			"active_slice":         "goal_and_responsibility",
			"truth_layer_ref":      truth.Layer,
			"truth_file_ref":       truth.FileRef,
			"truth_version_ref":    truth.VersionRef,
			"truth_fingerprint":    truth.Fingerprint,
			"baseline_slice_table": "present",
			"dynamic_slice_table":  "none",
			"finding_refs":         "none",
			"blocked_reason":       "none",
			"resume_next_step":     "review slice goal_and_responsibility",
		},
		Findings: "none",
		Resume:   "none",
	}
	state.Baseline, err = buildBaselineSlices(repoRoot, truth.Context)
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
		return RefreshResult{}, fmt.Errorf("check-work validation failed: %s", strings.Join(diagnostics, "; "))
	}

	truth, err := collectTruthInfo(repoRoot, objectType, object)
	if err != nil {
		return RefreshResult{}, err
	}
	state.Fields["truth_layer_ref"] = truth.Layer
	state.Fields["truth_file_ref"] = truth.FileRef
	state.Fields["truth_version_ref"] = truth.VersionRef
	state.Fields["truth_fingerprint"] = truth.Fingerprint

	definitions := baselineDefinitions()
	definitionsByID := map[string]sliceDefinition{}
	for _, definition := range definitions {
		definitionsByID[definition.ID] = definition
	}
	inputFilesChanged := map[string]bool{}
	for i := range state.Baseline {
		definition, ok := definitionsByID[state.Baseline[i].SliceID]
		if !ok {
			continue
		}
		currentInputFiles := union(definition.InputFiles(truth.Context))
		if !sameStringSlice(state.Baseline[i].InputFiles, currentInputFiles) {
			inputFilesChanged[state.Baseline[i].SliceID] = true
		}
		state.Baseline[i].SliceType = definition.SliceType
		state.Baseline[i].ReviewQuestion = definition.ReviewQuestion
		state.Baseline[i].WhyAdded = "baseline_catalog"
		state.Baseline[i].ParentSliceID = "none"
		state.Baseline[i].InputFiles = currentInputFiles
		state.Baseline[i].DependsOn = definition.DependsOn
	}

	result := RefreshResult{File: file, LastUpdatedAtUTC: formatUTC(now)}
	staleSet := map[string]bool{}
	allSlices := append([]sliceEntry{}, state.Baseline...)
	allSlices = append(allSlices, state.Dynamic...)
	for _, slice := range allSlices {
		if slice.Status == sliceStale {
			staleSet[slice.SliceID] = true
		}
	}
	for i := range allSlices {
		fingerprint, missing, err := computeFingerprint(repoRoot, allSlices[i].InputFiles)
		if err != nil {
			return RefreshResult{}, err
		}
		changed := len(missing) == 0 && (fingerprint != allSlices[i].InputFingerprint || inputFilesChanged[allSlices[i].SliceID])
		if changed {
			result.ChangedSlices = append(result.ChangedSlices, allSlices[i].SliceID)
		}
		for _, missingPath := range missing {
			result.MissingInputs = append(result.MissingInputs, allSlices[i].SliceID+":"+missingPath)
		}
		if len(missing) == 0 {
			allSlices[i].InputFingerprint = fingerprint
		}
		if allSlices[i].Status == slicePassed && (changed || len(missing) > 0) {
			allSlices[i].Status = sliceStale
			allSlices[i].ResultSummary = "stale: input changed"
			if len(missing) > 0 {
				allSlices[i].ResultSummary = "stale: input missing"
			}
			staleSet[allSlices[i].SliceID] = true
			result.StaleSlices = append(result.StaleSlices, allSlices[i].SliceID)
		}
	}
	for changed := true; changed; {
		changed = false
		for i := range allSlices {
			if allSlices[i].SliceType != "cross_convergence" || allSlices[i].Status != slicePassed {
				continue
			}
			for _, dependency := range allSlices[i].DependsOn {
				if staleSet[dependency] {
					allSlices[i].Status = sliceStale
					allSlices[i].ResultSummary = "stale: dependency stale"
					staleSet[allSlices[i].SliceID] = true
					result.StaleSlices = append(result.StaleSlices, allSlices[i].SliceID)
					changed = true
					break
				}
			}
		}
	}

	state.Baseline = allSlices[:len(state.Baseline)]
	state.Dynamic = allSlices[len(state.Baseline):]
	state.Fields["last_updated_at"] = formatUTC(now)
	state.Fields["dynamic_slice_table"] = "none"
	if len(state.Dynamic) > 0 {
		state.Fields["dynamic_slice_table"] = "present"
	}
	if err := os.WriteFile(file, []byte(renderState(state)), 0o644); err != nil {
		return RefreshResult{}, err
	}
	sort.Strings(result.StaleSlices)
	sort.Strings(result.ChangedSlices)
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
		return TouchResult{}, fmt.Errorf("check-work validation failed: %s", strings.Join(diagnostics, "; "))
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
	frontmatter, body, err := parseFrontmatter(string(content))
	if err != nil {
		return truthInfo{}, fmt.Errorf("%s: %w", mainSpecRef, err)
	}
	version := strings.TrimSpace(frontmatter["version"])
	if version == "" {
		return truthInfo{}, fmt.Errorf("%s: missing frontmatter.version", mainSpecRef)
	}

	appendices, err := collectAppendixFiles(repoRoot, mainSpecRef, frontmatter, body)
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
	frameworkFiles := existingFiles(repoRoot, []string{
		"specflow/framework/commands/unit_check.md",
		"specflow/framework/process_snapshot_contract.md",
		"specflow/framework/slice_work_state_protocol.md",
		"specflow/framework/candidate_handoff_contract.md",
		"specflow/framework/spec_writing_guide.md",
		"specflow/framework/candidate_intent_policy.md",
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
				"specflow/framework/spec_writing_guide.md",
				"specflow/framework/commands/unit_check.md",
			}),
			HandoffRuleFiles: existingFiles(repoRoot, []string{
				"specflow/framework/candidate_handoff_contract.md",
				"specflow/framework/process_snapshot_contract.md",
				"specflow/framework/commands/unit_plan.md",
			}),
		},
	}, nil
}

func baselineDefinitions() []sliceDefinition {
	local := []sliceDefinition{
		{
			ID:             "goal_and_responsibility",
			SliceType:      "local",
			ReviewQuestion: "Does the candidate connect user goal, unit responsibility, scope, non-goals, and owner fit.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.AppendixFiles, ctx.FrameworkFiles)
			},
		},
		{
			ID:             "dependency_truth_surface",
			SliceType:      "local",
			ReviewQuestion: "Does the candidate depend only on formal stable truth and explicit same-candidate appendix inputs.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.AppendixFiles, ctx.DependencyUnitFiles, ctx.DependencyRuleFiles, ctx.GlobalRuleFiles, ctx.RepositoryMappingFiles, ctx.FrameworkFiles)
			},
		},
		{
			ID:             "main_flow_and_state",
			SliceType:      "local",
			ReviewQuestion: "Does the candidate define the normal flow, state changes, ordering, and lifecycle semantics.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.AppendixFiles, ctx.FrameworkFiles)
			},
		},
		{
			ID:             "boundary_and_protocol",
			SliceType:      "local",
			ReviewQuestion: "Does every flow boundary name the public contract, port, adapter, store, event, trace source, or owner it needs.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.DependencyUnitFiles, ctx.DependencyRuleFiles, ctx.RepositoryMappingFiles, ctx.FrameworkFiles)
			},
		},
		{
			ID:             "data_artifact_and_output",
			SliceType:      "local",
			ReviewQuestion: "Does the candidate define produced artifacts, evidence, reports, traces, persistence records, and consumers.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.AppendixFiles, ctx.FrameworkFiles)
			},
		},
		{
			ID:             "error_edge_and_gap",
			SliceType:      "local",
			ReviewQuestion: "Does the candidate define important error states, missing dependency behavior, diagnostic gaps, and failure owners.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.AppendixFiles, ctx.DependencyUnitFiles, ctx.DependencyRuleFiles, ctx.FrameworkFiles)
			},
		},
		{
			ID:             "acceptance_and_testability",
			SliceType:      "local",
			ReviewQuestion: "Do acceptance items name test surfaces, proof methods, runnable status, and pass conditions that prove the goal.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.AppendixFiles, ctx.AcceptanceRuleFiles)
			},
		},
		{
			ID:             "implementation_handoff",
			SliceType:      "local",
			ReviewQuestion: "Can unit_plan plan implementation without inventing missing design, adapter, output, or test choices.",
			InputFiles: func(ctx inputContext) []string {
				return union([]string{ctx.TruthFile}, ctx.AppendixFiles, ctx.HandoffRuleFiles)
			},
		},
	}
	byID := map[string]sliceDefinition{}
	for _, definition := range local {
		byID[definition.ID] = definition
	}
	cross := []sliceDefinition{
		crossDefinition("goal_to_acceptance_convergence", "Do goal, responsibility, and acceptance items converge.", []string{"goal_and_responsibility", "acceptance_and_testability"}, byID),
		crossDefinition("flow_to_boundary_convergence", "Do main-flow steps converge with boundary protocols and owners.", []string{"main_flow_and_state", "boundary_and_protocol"}, byID),
		crossDefinition("dependency_truth_convergence", "Does dependency truth actually support the candidate behavior that relies on it.", []string{"dependency_truth_surface", "boundary_and_protocol"}, byID),
		crossDefinition("output_to_acceptance_convergence", "Do output artifacts converge with acceptance items and downstream handoff.", []string{"data_artifact_and_output", "acceptance_and_testability", "implementation_handoff"}, byID),
	}
	return append(local, cross...)
}

func crossDefinition(id, question string, dependsOn []string, byID map[string]sliceDefinition) sliceDefinition {
	return sliceDefinition{
		ID:             id,
		SliceType:      "cross_convergence",
		ReviewQuestion: question,
		DependsOn:      dependsOn,
		InputFiles: func(ctx inputContext) []string {
			sets := make([][]string, 0, len(dependsOn))
			for _, dependency := range dependsOn {
				if definition, ok := byID[dependency]; ok {
					sets = append(sets, definition.InputFiles(ctx))
				}
			}
			return union(sets...)
		},
	}
}

func buildBaselineSlices(repoRoot string, ctx inputContext) ([]sliceEntry, error) {
	result := []sliceEntry{}
	for _, definition := range baselineDefinitions() {
		inputFiles := union(definition.InputFiles(ctx))
		fingerprint, missing, err := computeFingerprint(repoRoot, inputFiles)
		if err != nil {
			return nil, err
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf("baseline slice %s has missing input files: %s", definition.ID, strings.Join(missing, ", "))
		}
		result = append(result, sliceEntry{
			SliceID:          definition.ID,
			SliceOrigin:      "baseline",
			SliceType:        definition.SliceType,
			Status:           slicePending,
			ReviewQuestion:   definition.ReviewQuestion,
			WhyAdded:         "baseline_catalog",
			ParentSliceID:    "none",
			InputFiles:       inputFiles,
			InputFingerprint: fingerprint,
			DependsOn:        definition.DependsOn,
			FindingRefs:      "none",
			ResultSummary:    "pending",
			ExitCondition:    "agent records passed blocked stale or skipped_not_applicable according to unit_check evidence",
			ResumeNextStep:   "review slice " + definition.ID,
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
		"active_slice",
		"truth_layer_ref",
		"truth_file_ref",
		"truth_version_ref",
		"truth_fingerprint",
		"baseline_slice_table",
		"dynamic_slice_table",
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
		diagnostics = append(diagnostics, "closed check-work files cannot be reused")
	}
	for _, field := range []string{"created_at", "last_updated_at"} {
		if _, err := parseTimestamp(state.Fields[field]); err != nil {
			diagnostics = append(diagnostics, field+" must use UTC format YYYY-MM-DDTHH:MM:SSZ")
		}
	}
	if lastUpdated, err := parseTimestamp(state.Fields["last_updated_at"]); err == nil && lastUpdated.After(now) {
		diagnostics = append(diagnostics, "last_updated_at must not be later than current UTC time")
	}
	expected := baselineIDSet()
	seen := map[string]bool{}
	allIDs := map[string]bool{}
	for _, slice := range state.Baseline {
		if seen[slice.SliceID] {
			diagnostics = append(diagnostics, "duplicate baseline slice: "+slice.SliceID)
		}
		seen[slice.SliceID] = true
		allIDs[slice.SliceID] = true
		if !expected[slice.SliceID] {
			diagnostics = append(diagnostics, "unexpected baseline slice: "+slice.SliceID)
		}
		if slice.SliceOrigin != "baseline" {
			diagnostics = append(diagnostics, "baseline slice must use slice_origin baseline: "+slice.SliceID)
		}
		if slice.ParentSliceID != "none" {
			diagnostics = append(diagnostics, "baseline slice parent_slice_id must be none: "+slice.SliceID)
		}
		diagnostics = append(diagnostics, validateSlice(repoRoot, slice)...)
	}
	for id := range expected {
		if !seen[id] {
			diagnostics = append(diagnostics, "missing baseline slice: "+id)
		}
	}
	for _, slice := range state.Dynamic {
		if allIDs[slice.SliceID] {
			diagnostics = append(diagnostics, "duplicate slice_id: "+slice.SliceID)
		}
		allIDs[slice.SliceID] = true
	}
	for _, slice := range state.Dynamic {
		if slice.SliceOrigin != "dynamic" {
			diagnostics = append(diagnostics, "dynamic slice must use slice_origin dynamic: "+slice.SliceID)
		}
		if slice.ParentSliceID == "" || slice.ParentSliceID == "none" || slice.ParentSliceID == slice.SliceID || !allIDs[slice.ParentSliceID] {
			diagnostics = append(diagnostics, "dynamic slice parent_slice_id must reference an existing slice: "+slice.SliceID)
		}
		for _, field := range []struct {
			name  string
			value string
		}{
			{"review_question", slice.ReviewQuestion},
			{"why_added", slice.WhyAdded},
			{"exit_condition", slice.ExitCondition},
		} {
			if strings.TrimSpace(field.value) == "" || strings.TrimSpace(field.value) == "none" {
				diagnostics = append(diagnostics, "dynamic slice missing "+field.name+": "+slice.SliceID)
			}
		}
		diagnostics = append(diagnostics, validateSlice(repoRoot, slice)...)
	}
	return diagnostics
}

func validateSlice(repoRoot string, slice sliceEntry) []string {
	diagnostics := []string{}
	for _, field := range []struct {
		name  string
		value string
	}{
		{"slice_id", slice.SliceID},
		{"slice_origin", slice.SliceOrigin},
		{"slice_type", slice.SliceType},
		{"status", slice.Status},
		{"review_question", slice.ReviewQuestion},
		{"why_added", slice.WhyAdded},
		{"parent_slice_id", slice.ParentSliceID},
		{"input_fingerprint", slice.InputFingerprint},
		{"finding_refs", slice.FindingRefs},
		{"result_summary", slice.ResultSummary},
		{"exit_condition", slice.ExitCondition},
		{"resume_next_step", slice.ResumeNextStep},
	} {
		if strings.TrimSpace(field.value) == "" {
			diagnostics = append(diagnostics, field.name+" is required for slice: "+slice.SliceID)
		}
	}
	if slice.SliceType != "local" && slice.SliceType != "cross_convergence" {
		diagnostics = append(diagnostics, "invalid slice_type for "+slice.SliceID+": "+slice.SliceType)
	}
	if !isSliceStatus(slice.Status) {
		diagnostics = append(diagnostics, "invalid slice status for "+slice.SliceID+": "+slice.Status)
	}
	if len(slice.InputFiles) == 0 {
		diagnostics = append(diagnostics, "input_files is required for slice: "+slice.SliceID)
	}
	for _, relPath := range slice.InputFiles {
		if strings.TrimSpace(relPath) == "" || filepath.IsAbs(relPath) || strings.Contains(relPath, "\\") {
			diagnostics = append(diagnostics, "input_files must use repository-relative slash paths: "+slice.SliceID)
			continue
		}
		if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(relPath))); err != nil && !os.IsNotExist(err) {
			diagnostics = append(diagnostics, "cannot inspect input file for "+slice.SliceID+": "+relPath)
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
	baselineSection, ok := sections["Baseline Slices"]
	if !ok {
		return state, fmt.Errorf("missing section: Baseline Slices")
	}
	state.Baseline, err = parseSliceTable(baselineSection)
	if err != nil {
		return state, fmt.Errorf("baseline slice table: %w", err)
	}
	dynamicSection, ok := sections["Dynamic Slices"]
	if !ok {
		return state, fmt.Errorf("missing section: Dynamic Slices")
	}
	if strings.TrimSpace(dynamicSection) != "none" {
		state.Dynamic, err = parseSliceTable(dynamicSection)
		if err != nil {
			return state, fmt.Errorf("dynamic slice table: %w", err)
		}
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

func parseSliceTable(section string) ([]sliceEntry, error) {
	rows := parseMarkdownRows(section)
	if len(rows) == 0 {
		return nil, fmt.Errorf("slice table is empty")
	}
	header := rows[0]
	if len(header) != len(sliceColumns) {
		return nil, fmt.Errorf("slice table header has %d columns, want %d", len(header), len(sliceColumns))
	}
	for i, column := range sliceColumns {
		if header[i] != column {
			return nil, fmt.Errorf("slice table column %d is %q, want %q", i+1, header[i], column)
		}
	}
	result := []sliceEntry{}
	for _, row := range rows[1:] {
		if len(row) != len(sliceColumns) {
			return nil, fmt.Errorf("slice row has %d columns, want %d", len(row), len(sliceColumns))
		}
		result = append(result, sliceEntry{
			SliceID:          row[0],
			SliceOrigin:      row[1],
			SliceType:        row[2],
			Status:           row[3],
			ReviewQuestion:   row[4],
			WhyAdded:         row[5],
			ParentSliceID:    row[6],
			InputFiles:       parseList(row[7]),
			InputFingerprint: row[8],
			DependsOn:        parseList(row[9]),
			FindingRefs:      row[10],
			ResultSummary:    row[11],
			ExitCondition:    row[12],
			ResumeNextStep:   row[13],
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
	b.WriteString("# Unit Check Work State\n\n")
	b.WriteString("## Run State\n\n")
	b.WriteString("| field | value |\n")
	b.WriteString("|---|---|\n")
	for _, field := range []string{"work_flow", "work_id", "object_type", "object_ref", "status", "created_at", "last_updated_at", "active_slice", "truth_layer_ref", "truth_file_ref", "truth_version_ref", "truth_fingerprint", "baseline_slice_table", "dynamic_slice_table", "finding_refs", "blocked_reason", "resume_next_step"} {
		b.WriteString(fmt.Sprintf("| %s | %s |\n", field, cleanCell(state.Fields[field])))
	}
	b.WriteString("\n## Baseline Slices\n\n")
	renderSliceTable(&b, state.Baseline)
	b.WriteString("\n## Dynamic Slices\n\n")
	if len(state.Dynamic) == 0 {
		b.WriteString("none\n")
	} else {
		renderSliceTable(&b, state.Dynamic)
	}
	b.WriteString("\n## Findings\n\n")
	b.WriteString(defaultText(state.Findings))
	b.WriteString("\n\n## Resume\n\n")
	b.WriteString(defaultText(state.Resume))
	b.WriteString("\n")
	return b.String()
}

func renderSliceTable(b *strings.Builder, slices []sliceEntry) {
	b.WriteString("| " + strings.Join(sliceColumns, " | ") + " |\n")
	b.WriteString("|" + strings.Repeat("---|", len(sliceColumns)) + "\n")
	for _, slice := range slices {
		values := []string{
			slice.SliceID,
			slice.SliceOrigin,
			slice.SliceType,
			slice.Status,
			slice.ReviewQuestion,
			slice.WhyAdded,
			slice.ParentSliceID,
			joinList(slice.InputFiles),
			slice.InputFingerprint,
			joinList(slice.DependsOn),
			slice.FindingRefs,
			slice.ResultSummary,
			slice.ExitCondition,
			slice.ResumeNextStep,
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

func collectAppendixFiles(repoRoot, mainSpecRef string, frontmatter map[string]string, body string) ([]string, error) {
	mainDir := filepath.Dir(filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)))
	result := []string{}
	if evidenceRef := strings.TrimSpace(frontmatter["evidence_appendix_ref"]); evidenceRef != "" && evidenceRef != "none" {
		relPath, err := resolveAppendixRef(repoRoot, mainDir, evidenceRef)
		if err != nil {
			return nil, err
		}
		result = append(result, relPath)
	}
	for _, destination := range markdownLinkPattern.FindAllStringSubmatch(body, -1) {
		if len(destination) != 2 {
			continue
		}
		linkDestination := strings.TrimSpace(destination[1])
		if linkDestination == "" || strings.HasPrefix(linkDestination, "/") || strings.Contains(linkDestination, "://") || filepath.Ext(linkDestination) != ".md" {
			continue
		}
		absPath := filepath.Clean(filepath.Join(mainDir, filepath.FromSlash(linkDestination)))
		relPath, err := filepath.Rel(repoRoot, absPath)
		if err != nil {
			return nil, err
		}
		relPath = filepath.ToSlash(relPath)
		if strings.HasPrefix(relPath, "../") || relPath == ".." || relPath == mainSpecRef {
			continue
		}
		if strings.Contains(relPath, "/appendix/") {
			result = append(result, relPath)
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

func baselineIDSet() map[string]bool {
	result := map[string]bool{}
	for _, definition := range baselineDefinitions() {
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

func isSliceStatus(status string) bool {
	switch status {
	case slicePending, slicePassed, sliceBlocked, sliceStale, sliceSkippedNotApplicable:
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

func sameStringSlice(left, right []string) bool {
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
