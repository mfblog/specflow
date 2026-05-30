package evaluationrequest

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
)

const (
	PackUnitCheckPass             = "unit_check_pass"
	PackUnitPlanPlanReady         = "unit_plan_plan_ready"
	PackUnitVerifyReadyToPromote  = "unit_verify_ready_to_promote"
	PackUnitStableVerifyAdvancing = "unit_stable_verify_advancing"
	PackFreshnessTextDriftReuse   = "freshness_text_drift_reuse"
)

type Options struct {
	RepoRoot    string
	ObjectType  string
	Object      string
	Pack        string
	ProcessKind string
	Now         time.Time
}

type Result struct {
	RequestFile        string
	Pack               string
	ProcessKind        string
	ProcessFile        string
	TriggerInstruction string
	ReviewInputRefs    []string
	ReviewFileRefs     []string
	ReviewEvidenceRefs []string
	Validation         snapshot.ValidationResult
}

type packConfig struct {
	Pack                string
	ProcessKind         string
	RequiresProcess     bool
	LifecycleRef        string
	ReviewTitle         string
	ReviewGoal          string
	AllowedInputs       []string
	ForbiddenInputs     []string
	EvaluationQuestions []string
}

type reviewRefs struct {
	FileRefs     []string
	EvidenceRefs []string
}

func Create(options Options) (Result, error) {
	normalized, config, err := normalizeOptions(options)
	if err != nil {
		return Result{}, err
	}

	var validation snapshot.ValidationResult
	if config.Pack == PackFreshnessTextDriftReuse {
		validation, err = snapshot.ValidateProcessFileForObject(normalized.RepoRoot, normalized.ObjectType, normalized.Object, normalized.ProcessKind)
		if err != nil {
			return Result{}, err
		}
		if validation.FreshnessImpact != snapshot.FreshnessTextDrift || validation.EvidenceReuse != snapshot.EvidenceReusePendingReview {
			return Result{}, fmt.Errorf("freshness review request requires freshness_impact=%s and evidence_reuse=%s; actual freshness_impact=%s evidence_reuse=%s", snapshot.FreshnessTextDrift, snapshot.EvidenceReusePendingReview, validation.FreshnessImpact, validation.EvidenceReuse)
		}
	} else {
		validation, err = snapshot.ValidateProcessFileForIndependentEvaluationRequest(normalized.RepoRoot, normalized.ObjectType, normalized.Object, normalized.ProcessKind)
		if err != nil {
			return Result{}, err
		}
		if !validation.Valid {
			return Result{}, fmt.Errorf("process artifact is not ready for independent evaluation request: %s", strings.Join(validation.Mismatches, "; "))
		}
	}

	processFile, err := snapshot.ProcessFilePath(normalized.ObjectType, normalized.Object, normalized.ProcessKind)
	if err != nil {
		return Result{}, err
	}
	processData, err := snapshot.LoadProcessSnapshot(normalized.RepoRoot, normalized.ObjectType, normalized.Object, normalized.ProcessKind)
	if err != nil {
		return Result{}, err
	}
	if config.Pack == PackUnitStableVerifyAdvancing && !stableVerifyDecisionAdvances(processData.Scalars["decision"]) {
		return Result{}, fmt.Errorf("pack %q requires stable_verify decision aligned, controlled_repair_required, or controlled_change_required; actual decision=%s", config.Pack, strings.TrimSpace(processData.Scalars["decision"]))
	}
	requestFile := requestFilePath(normalized.ObjectType, normalized.Object, config.Pack)
	refs := collectReviewRefs(normalized, config, processFile, validation, processData)
	inputRefs := refs.Combined()
	trigger := fmt.Sprintf("Read and execute this independent evaluation request: %s", requestFile)

	body := renderRequest(normalized, config, processFile, requestFile, refs, trigger, validation)
	absRequestFile := filepath.Join(normalized.RepoRoot, filepath.FromSlash(requestFile))
	if err := os.MkdirAll(filepath.Dir(absRequestFile), 0o755); err != nil {
		return Result{}, fmt.Errorf("mkdir %s: %w", filepath.Dir(requestFile), err)
	}
	if err := os.WriteFile(absRequestFile, []byte(body), 0o644); err != nil {
		return Result{}, fmt.Errorf("write %s: %w", requestFile, err)
	}

	return Result{
		RequestFile:        requestFile,
		Pack:               config.Pack,
		ProcessKind:        normalized.ProcessKind,
		ProcessFile:        processFile,
		TriggerInstruction: trigger,
		ReviewInputRefs:    inputRefs,
		ReviewFileRefs:     refs.FileRefs,
		ReviewEvidenceRefs: refs.EvidenceRefs,
		Validation:         validation,
	}, nil
}

func normalizeOptions(options Options) (Options, packConfig, error) {
	normalized := Options{
		RepoRoot:    strings.TrimSpace(options.RepoRoot),
		ObjectType:  strings.TrimSpace(options.ObjectType),
		Object:      strings.TrimSpace(options.Object),
		Pack:        strings.TrimSpace(options.Pack),
		ProcessKind: strings.TrimSpace(options.ProcessKind),
		Now:         options.Now.UTC(),
	}
	if normalized.RepoRoot == "" {
		normalized.RepoRoot = "."
	}
	if normalized.ObjectType == "" || normalized.Object == "" || normalized.Pack == "" {
		return Options{}, packConfig{}, fmt.Errorf("object-type, object, and pack are required")
	}
	if normalized.ObjectType != "unit" {
		return Options{}, packConfig{}, fmt.Errorf("object type %q is not supported; only unit is supported", normalized.ObjectType)
	}
	if invalidPathPart(normalized.Object) {
		return Options{}, packConfig{}, fmt.Errorf("object %q is not a valid path segment", normalized.Object)
	}
	config, ok := configsByPack()[normalized.Pack]
	if !ok {
		return Options{}, packConfig{}, fmt.Errorf("unsupported independent evaluation pack %q", normalized.Pack)
	}
	if normalized.ProcessKind == "" {
		if config.RequiresProcess {
			return Options{}, packConfig{}, fmt.Errorf("process is required for pack %q", normalized.Pack)
		}
		normalized.ProcessKind = config.ProcessKind
	} else if !config.RequiresProcess && normalized.ProcessKind != config.ProcessKind {
		return Options{}, packConfig{}, fmt.Errorf("pack %q requires process %q, got %q", normalized.Pack, config.ProcessKind, normalized.ProcessKind)
	}
	if normalized.Now.IsZero() {
		normalized.Now = time.Now().UTC()
	}
	return normalized, config, nil
}

func configsByPack() map[string]packConfig {
	return map[string]packConfig{
		PackUnitCheckPass: {
			Pack:         PackUnitCheckPass,
			ProcessKind:  "check",
			LifecycleRef: "framework/lifecycle/unit_check.md",
			ReviewTitle:  "Unit Check Pass Review",
			ReviewGoal:   "Decide whether candidate unit truth is ready for planning.",
			AllowedInputs: []string{
				"user goal or exact `unit_check:{unit}` target.",
				"candidate unit truth and explicitly referenced appendices, stable truth, and rules.",
				"`_check_result/unit/{unit}.md`.",
				"`framework/lifecycle/unit_check.md` check questions.",
			},
			ForbiddenInputs: []string{
				"implementation plan drafts.",
				"implementation files unless repository mapping is part of the boundary question.",
				"executor rationale not present in durable truth or `_check_result`.",
			},
			EvaluationQuestions: []string{
				"Is the unit goal, responsibility, boundary, dependency truth, and rule binding explicit enough for planning?",
				"Are acceptance items testable without inventing behavior?",
				"Does the check result match the candidate truth and evidence refs?",
			},
		},
		PackUnitPlanPlanReady: {
			Pack:         PackUnitPlanPlanReady,
			ProcessKind:  "plan",
			LifecycleRef: "framework/lifecycle/unit_plan.md",
			ReviewTitle:  "Unit Plan Ready Review",
			ReviewGoal:   "Decide whether the active plan is ready for implementation handoff.",
			AllowedInputs: []string{
				"user goal or exact `unit_plan:{unit}` target.",
				"candidate unit truth and valid `_check_result/unit/{unit}.md`.",
				"active plan under review.",
				"plan acceptance coverage criteria from `framework/lifecycle/unit_plan.md`.",
			},
			ForbiddenInputs: []string{
				"implementation work not authorized by the active plan.",
				"executor-only design rationale that is absent from the active plan.",
				"unrelated repository architecture exploration.",
			},
			EvaluationQuestions: []string{
				"Does the plan cover every accepted acceptance item?",
				"Does the plan stay inside checked truth and named implementation surfaces?",
				"Can unit_impl execute without inventing behavior or ownership?",
			},
		},
		PackUnitVerifyReadyToPromote: {
			Pack:         PackUnitVerifyReadyToPromote,
			ProcessKind:  "verify",
			LifecycleRef: "framework/lifecycle/unit_verify.md",
			ReviewTitle:  "Unit Verify Ready-To-Promote Review",
			ReviewGoal:   "Decide whether candidate verification is ready for promotion.",
			AllowedInputs: []string{
				"user goal or exact `unit_verify:{unit}` target.",
				"candidate unit truth, valid check result, and active plan.",
				"verify result under review.",
				"evidence refs needed to inspect acceptance coverage.",
			},
			ForbiddenInputs: []string{
				"unrecorded executor claims that tests passed.",
				"implementation changes not represented by plan or evidence refs.",
				"promotion judgment not grounded in verify evidence.",
			},
			EvaluationQuestions: []string{
				"Does the verify result cover every executable acceptance item?",
				"Are evidence refs inspectable and aligned with the active plan?",
				"Is the candidate ready for promotion without hiding unresolved gaps?",
			},
		},
		PackUnitStableVerifyAdvancing: {
			Pack:         PackUnitStableVerifyAdvancing,
			ProcessKind:  "stable_verify",
			LifecycleRef: "framework/lifecycle/unit_stable_verify.md",
			ReviewTitle:  "Stable Verify Advancing Review",
			ReviewGoal:   "Decide whether the stable verify result supports the stored advancing decision.",
			AllowedInputs: []string{
				"exact `unit_stable_verify:{unit}` target.",
				"stable unit truth, referenced appendices, rules, and repository mapping snapshot.",
				"stable verify result under review.",
				"implementation surface refs and evidence refs needed to inspect stable alignment.",
				"decision criteria from `framework/lifecycle/unit_stable_verify.md`.",
			},
			ForbiddenInputs: []string{
				"candidate truth unless the stable verify result explicitly cites it as historical context.",
				"proposed repairs or changes not captured in the stable verify result.",
				"executor preference for aligned, controlled repair, or controlled change outcomes.",
			},
			EvaluationQuestions: []string{
				"Does current implementation align with stable truth, or does the stored decision correctly identify the controlled next step?",
				"Does the evidence matrix cover every current stable acceptance item?",
				"Are implementation surface refs and evidence refs sufficient for the stored decision?",
			},
		},
		PackFreshnessTextDriftReuse: {
			Pack:            PackFreshnessTextDriftReuse,
			RequiresProcess: true,
			ReviewTitle:     "Freshness Text-Drift Reuse Review",
			ReviewGoal:      "Decide whether text-only drift can reuse existing process evidence.",
			AllowedInputs: []string{
				"current truth or spec file.",
				"prior process evidence being reused.",
				"deterministic freshness classification showing `text_drift`.",
				"acceptance behavior fingerprint comparison and current fingerprint reported by tooling.",
			},
			ForbiddenInputs: []string{
				"reuse claims when deterministic validation reports `semantic_drift`, `acceptance_drift`, `dependency_drift`, `schema_drift`, or `unknown_drift`.",
				"executor assertions that the text change is harmless without current file refs.",
				"unrelated changes outside the file and process evidence under review.",
			},
			EvaluationQuestions: []string{
				"Is the change only wording, formatting, or clarification that preserves the acceptance behavior already reviewed?",
				"Does the prior evidence still answer the same gate question?",
				"Is recreating evidence unnecessary for semantic safety?",
			},
		},
	}
}

func requestFilePath(objectType, object, pack string) string {
	return filepath.ToSlash(filepath.Join("docs/specs/_independent_evaluation/requests", objectType, object, pack+".md"))
}

func collectReviewRefs(options Options, config packConfig, processFile string, validation snapshot.ValidationResult, processData snapshot.ProcessSnapshotData) reviewRefs {
	expected := validation.Expected
	fileRefs := []string{
		"framework/core/independent_evaluation.md",
		processFile,
	}
	if config.LifecycleRef != "" {
		fileRefs = append(fileRefs, config.LifecycleRef)
	}
	fileRefs = appendExistingSpecRefs(fileRefs, options.RepoRoot, options.ObjectType, options.Object, expected.TruthLayerRef)
	fileRefs = appendSnapshotRefs(fileRefs, expected)
	evidenceRefs := []string{}

	switch config.Pack {
	case PackUnitPlanPlanReady:
		fileRefs = append(fileRefs, snapshot.CheckResultFilePath(options.ObjectType, options.Object))
	case PackUnitVerifyReadyToPromote:
		fileRefs = append(fileRefs, snapshot.CheckResultFilePath(options.ObjectType, options.Object), snapshot.ActivePlanFilePath(options.Object))
	case PackUnitStableVerifyAdvancing:
		fileRefs = append(fileRefs, "docs/specs/repository_mapping.md")
		evidenceRefs = appendScalarRefs(evidenceRefs, processData, "implementation_surface_refs", "evidence_refs")
	case PackFreshnessTextDriftReuse:
		fileRefs = append(fileRefs, "framework/core/freshness.md")
		switch options.ProcessKind {
		case "plan":
			fileRefs = append(fileRefs, snapshot.ActivePlanFilePath(options.Object))
		case "check":
			fileRefs = append(fileRefs, snapshot.CheckResultFilePath(options.ObjectType, options.Object))
		case "verify":
			fileRefs = append(fileRefs, snapshot.VerifyResultFilePath(options.ObjectType, options.Object))
		case "stable_verify":
			fileRefs = append(fileRefs, snapshot.StableVerifyResultFilePath(options.ObjectType, options.Object))
		}
	}
	return reviewRefs{
		FileRefs:     sortedUnique(fileRefs),
		EvidenceRefs: sortedUnique(evidenceRefs),
	}
}

func (refs reviewRefs) Combined() []string {
	return sortedUnique(append(append([]string{}, refs.FileRefs...), refs.EvidenceRefs...))
}

func appendExistingSpecRefs(refs []string, repoRoot, objectType, object, activeLayer string) []string {
	layers := []string{activeLayer}
	if activeLayer == "candidate" {
		layers = append(layers, "stable")
	}
	for _, layer := range layers {
		ref, err := specpaths.ObjectMainSpecFileRef(objectType, layer, object)
		if err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(ref))); err == nil {
			refs = append(refs, ref)
		}
	}
	return refs
}

func appendSnapshotRefs(refs []string, expected snapshot.Snapshot) []string {
	refs = append(refs, expected.SpecFileRef)
	for _, entry := range expected.ModuleAppendixSnapshot {
		refs = append(refs, entry.FileRef)
	}
	for _, entry := range expected.UnitSnapshot {
		refs = append(refs, entry.FileRef)
	}
	for _, entry := range expected.RuleSnapshot {
		refs = append(refs, entry.FileRef)
	}
	return refs
}

func appendScalarRefs(refs []string, processData snapshot.ProcessSnapshotData, keys ...string) []string {
	for _, key := range keys {
		value := strings.TrimSpace(processData.Scalars[key])
		if value == "" || value == "none" {
			continue
		}
		for _, part := range strings.Split(value, ";") {
			refs = append(refs, strings.TrimSpace(part))
		}
	}
	return refs
}

func stableVerifyDecisionAdvances(decision string) bool {
	switch strings.TrimSpace(decision) {
	case "aligned", "controlled_repair_required", "controlled_change_required":
		return true
	default:
		return false
	}
}

func renderRequest(options Options, config packConfig, processFile, requestFile string, refs reviewRefs, trigger string, validation snapshot.ValidationResult) string {
	inputRefs := refs.Combined()
	var b strings.Builder
	b.WriteString("# Independent Evaluation Request\n\n")
	b.WriteString("## Request\n\n")
	writeField(&b, "object_type", options.ObjectType)
	writeField(&b, "object_ref", options.Object)
	writeField(&b, "reviewer_pack", config.Pack)
	writeField(&b, "review_title", config.ReviewTitle)
	writeField(&b, "process_kind", options.ProcessKind)
	writeField(&b, "process_file", processFile)
	writeField(&b, "request_file", requestFile)
	writeField(&b, "created_at", options.Now.UTC().Format("2006-01-02T15:04:05Z"))
	b.WriteString("\n## Reviewer Role\n\n")
	b.WriteString("You are the independent reviewer for this request. Do not modify repository files. Read only the files listed in Review File Refs and the reviewer pack details in this request.\n\n")
	b.WriteString("Use Review Evidence Refs only to judge whether the recorded evidence is sufficient and traceable; do not treat every evidence ref as a readable file.\n\n")
	b.WriteString("## Review Goal\n\n")
	b.WriteString(config.ReviewGoal)
	b.WriteString("\n\n## Allowed Inputs\n\n")
	writeBullets(&b, config.AllowedInputs)
	b.WriteString("\n## Forbidden Inputs\n\n")
	writeBullets(&b, config.ForbiddenInputs)
	b.WriteString("\n## Review File Refs\n\n")
	writeRefList(&b, refs.FileRefs)
	b.WriteString("\n## Review Evidence Refs\n\n")
	writeRefList(&b, refs.EvidenceRefs)
	b.WriteString("\n## Evaluation Questions\n\n")
	writeBullets(&b, config.EvaluationQuestions)
	if config.Pack == PackFreshnessTextDriftReuse {
		b.WriteString("\n## Mechanical Validation\n\n")
		writeField(&b, "freshness_impact", validation.FreshnessImpact)
		writeField(&b, "evidence_reuse", validation.EvidenceReuse)
		writeField(&b, "freshness_current_fingerprint", validation.Expected.SpecFingerprint)
	}
	b.WriteString("\n## Reviewer Output\n\n")
	b.WriteString("Return exactly one reviewer result:\n\n")
	b.WriteString("```text\npass | blocked | needs_human_decision\n```\n\n")
	b.WriteString("If the result is `blocked` or `needs_human_decision`, include concrete blocking findings. If the result is `pass`, include no findings.\n\n")
	b.WriteString("## Executor Receipt After Pass\n\n")
	b.WriteString("Only the executor writes this receipt into process evidence after receiving reviewer result `pass`.\n\n")
	if config.Pack == PackFreshnessTextDriftReuse {
		renderFreshnessReceipt(&b, config.Pack, requestFile, inputRefs, validation.Expected.SpecFingerprint)
	} else {
		renderIndependentEvaluationReceipt(&b, config.Pack, requestFile, inputRefs)
	}
	b.WriteString("## Trigger Instruction\n\n")
	b.WriteString(trigger)
	b.WriteString("\n")
	return b.String()
}

func writeBullets(b *strings.Builder, items []string) {
	if len(items) == 0 {
		b.WriteString("- none\n")
		return
	}
	for _, item := range items {
		b.WriteString("- ")
		b.WriteString(item)
		b.WriteString("\n")
	}
}

func writeRefList(b *strings.Builder, refs []string) {
	if len(refs) == 0 {
		b.WriteString("- none\n")
		return
	}
	for _, ref := range refs {
		b.WriteString("- ")
		b.WriteString(ref)
		b.WriteString("\n")
	}
}

func renderIndependentEvaluationReceipt(b *strings.Builder, pack, requestFile string, inputRefs []string) {
	b.WriteString("```yaml\n")
	b.WriteString("evaluation_mode: independent\n")
	b.WriteString("reviewer_result: pass\n")
	b.WriteString("reviewer_context: minimal_context\n")
	b.WriteString("review_input_refs: ")
	writeReceiptRefs(b, pack, requestFile, inputRefs)
	b.WriteString("\n")
	b.WriteString("review_findings: none\n")
	b.WriteString("human_decision_refs: none\n")
	b.WriteString("```\n\n")
}

func renderFreshnessReceipt(b *strings.Builder, pack, requestFile string, inputRefs []string, currentFingerprint string) {
	b.WriteString("```yaml\n")
	b.WriteString("freshness_impact: text_drift\n")
	b.WriteString("evidence_reuse: accepted\n")
	b.WriteString("freshness_current_fingerprint: ")
	b.WriteString(currentFingerprint)
	b.WriteString("\n")
	b.WriteString("freshness_review_mode: independent\n")
	b.WriteString("freshness_reviewer_result: pass\n")
	b.WriteString("freshness_reviewer_context: minimal_context\n")
	b.WriteString("freshness_review_input_refs: ")
	writeReceiptRefs(b, pack, requestFile, inputRefs)
	b.WriteString("\n")
	b.WriteString("freshness_review_findings: none\n")
	b.WriteString("```\n\n")
}

func writeReceiptRefs(b *strings.Builder, pack, requestFile string, inputRefs []string) {
	b.WriteString(pack)
	b.WriteString(";")
	b.WriteString(requestFile)
	for _, ref := range inputRefs {
		b.WriteString(";")
		b.WriteString(ref)
	}
}

func writeField(b *strings.Builder, key, value string) {
	b.WriteString("- `")
	b.WriteString(key)
	b.WriteString("`: `")
	b.WriteString(value)
	b.WriteString("`\n")
}

func invalidPathPart(value string) bool {
	return value == "" || strings.Contains(value, "/") || strings.Contains(value, "\\") || value == "." || value == ".." || strings.Contains(value, "..")
}

func sortedUnique(items []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, item := range items {
		item = strings.TrimSpace(filepath.ToSlash(item))
		if item == "" || item == "none" || seen[item] {
			continue
		}
		seen[item] = true
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}
