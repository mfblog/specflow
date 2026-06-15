package evaluationrequest

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specflowlayout"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
)

const (
	PackUnitCheckPass             = "unit_check_pass"
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
	ReviewStandardRefs  []reviewStandardRef
	AllowedInputs       []string
	ForbiddenInputs     []string
	EvaluationQuestions []string
}

type reviewStandardRef struct {
	Ref       string
	Authority string
}

type reviewRefs struct {
	SubjectRefs  []string
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
		ReviewFileRefs:     refs.SubjectRefs,
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
	configs, err := configsByPack(normalized.RepoRoot)
	if err != nil {
		return Options{}, packConfig{}, fmt.Errorf("load pack configs: %w", err)
	}
	config, ok := configs[normalized.Pack]
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

func configsByPack(repoRoot string) (map[string]packConfig, error) {
	layout, err := specflowlayout.Resolve(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve layout for independent_evaluation.md: %w", err)
	}
	path := filepath.Join(repoRoot, filepath.FromSlash(layout.FrameworkRoot), filepath.FromSlash("core/independent_evaluation.md"))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read independent_evaluation.md: %w", err)
	}
	content := strings.ReplaceAll(string(data), "\r\n", "\n")

	// Find the Reviewer Packs section
	idx := strings.Index(content, "\n## Reviewer Packs\n")
	if idx < 0 {
		return nil, fmt.Errorf("independent_evaluation.md: missing ## Reviewer Packs section")
	}
	// Stop at next ## section after Reviewer Packs
	body := content[idx+len("\n## Reviewer Packs\n"):]
	if end := strings.Index(body, "\n## "); end > 0 {
		body = body[:end]
	}

	packs := parsePackSections(body)
	result := make(map[string]packConfig, len(packs))
	for packName := range packs {
		meta := packMetaByID[packName]
		if meta.Pack == "" {
			continue
		}
		meta.AllowedInputs = packs[packName].AllowedInputs
		meta.ForbiddenInputs = packs[packName].ForbiddenInputs
		meta.EvaluationQuestions = packs[packName].EvaluationQuestions
		meta.ReviewStandardRefs = packs[packName].ReviewStandardRefs
		meta.LifecycleRef = findLifecycleRef(meta.ReviewStandardRefs)
		result[packName] = meta.toPackConfig()
	}
	return result, nil
}

type packMeta struct {
	Pack            string
	ProcessKind     string
	RequiresProcess bool
	ReviewTitle     string
	ReviewGoal      string
	LifecycleRef    string
	AllowedInputs   []string
	ForbiddenInputs []string
	EvaluationQuestions []string
	ReviewStandardRefs  []reviewStandardRef
}

func (m packMeta) toPackConfig() packConfig {
	return packConfig{
		Pack:                m.Pack,
		ProcessKind:         m.ProcessKind,
		RequiresProcess:     m.RequiresProcess,
		LifecycleRef:        m.LifecycleRef,
		ReviewTitle:         m.ReviewTitle,
		ReviewGoal:          m.ReviewGoal,
		ReviewStandardRefs:  m.ReviewStandardRefs,
		AllowedInputs:       m.AllowedInputs,
		ForbiddenInputs:     m.ForbiddenInputs,
		EvaluationQuestions: m.EvaluationQuestions,
	}
}

var packMetaByID = map[string]packMeta{
	PackUnitCheckPass: {
		Pack: PackUnitCheckPass, ProcessKind: "check",
		ReviewTitle: "Unit Check Pass Review",
		ReviewGoal:  "Decide whether candidate unit truth is clear enough for downstream work.",
	},
	PackUnitVerifyReadyToPromote: {
		Pack: PackUnitVerifyReadyToPromote, ProcessKind: "verify",
		ReviewTitle: "Unit Verify Ready-To-Promote Review",
		ReviewGoal:  "Decide whether candidate verification is ready for promotion.",
	},
	PackUnitStableVerifyAdvancing: {
		Pack: PackUnitStableVerifyAdvancing, ProcessKind: "stable_verify",
		ReviewTitle: "Stable Verify Advancing Review",
		ReviewGoal:  "Decide whether the stable verify result supports the stored advancing decision.",
	},
	PackFreshnessTextDriftReuse: {
		Pack: PackFreshnessTextDriftReuse, RequiresProcess: true,
		ReviewTitle: "Freshness Text-Drift Reuse Review",
		ReviewGoal:  "Decide whether text-only drift can reuse existing process evidence.",
	},
}

func parsePackSections(body string) map[string]packMeta {
	sections := splitPackSections(body)
	result := make(map[string]packMeta, len(sections))
	for name, content := range sections {
		meta := packMetaByID[name]
		meta.AllowedInputs = parseListSection(content, "Allowed Inputs")
		meta.ForbiddenInputs = parseListSection(content, "Forbidden Inputs")
		meta.EvaluationQuestions = parseListSection(content, "Evaluation Questions")
		meta.ReviewStandardRefs = parseStandardRefs(content)
		meta.LifecycleRef = findLifecycleRef(meta.ReviewStandardRefs)
		result[name] = meta
	}
	return result
}

func splitPackSections(body string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(body, "\n### `")
	for _, part := range parts[1:] {
		if idx := strings.Index(part, "`\n"); idx > 0 {
			name := part[:idx]
			content := part[idx+2:]
			result[name] = content
		}
	}
	return result
}

func parseListSection(content, heading string) []string {
	idx := strings.Index(content, heading+":")
	if idx < 0 {
		return nil
	}
	start := idx + len(heading) + 1
	// Skip blank lines after heading
	for start < len(content) && content[start] == '\n' {
		start++
	}
	// Collect lines until next section heading (ends with ":") or end of content
	end := start
	lines := strings.Split(content[start:], "\n")
	var items []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Check if this is a new section heading (ends with ":" but is not a ** sub-heading)
		if strings.HasSuffix(line, ":") && !strings.HasPrefix(line, "**") {
			break
		}
		// Strip numbered prefix (only if the number is at the start: "1. ", "10. ", etc.)
		if dot := strings.Index(line, ". "); dot > 0 && dot <= 3 {
			text := strings.TrimSpace(line[dot+2:])
			if text != "" && !strings.HasPrefix(text, "```") {
				items = append(items, text)
			}
		} else if strings.HasPrefix(line, "**") {
			// Sub-section label, skip
			continue
		} else {
			// Could be a continuation line
			if len(items) > 0 {
				items[len(items)-1] += " " + line
			}
		}
	}
	_ = end
	return items
}

func parseStandardRefs(content string) []reviewStandardRef {
	idx := strings.Index(content, "Review Standard Refs:")
	if idx < 0 {
		return nil
	}
	section := content[idx+len("Review Standard Refs:"):]
	end := strings.Index(section, "\n\n")
	if end > 0 {
		section = section[:end]
	}
	var refs []reviewStandardRef
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		i := strings.Index(line, ". ")
		if i < 0 {
			continue
		}
		rest := strings.TrimSpace(line[i+2:])
		// Split at " - " to get ref and authority
		if j := strings.Index(rest, " - "); j > 0 {
			ref := strings.Trim(strings.TrimSpace(rest[:j]), "`")
			authority := strings.TrimSuffix(strings.TrimSpace(rest[j+3:]), ".")
			refs = append(refs, reviewStandardRef{Ref: ref, Authority: authority})
		}
	}
	return refs
}

func findLifecycleRef(refs []reviewStandardRef) string {
	for _, r := range refs {
		if strings.HasPrefix(r.Ref, "framework/lifecycle/") {
			return r.Ref
		}
	}
	return ""
}

func requestFilePath(objectType, object, pack string) string {
	return filepath.ToSlash(filepath.Join("docs/specs/_independent_evaluation/requests", objectType, object, pack+".md"))
}

func collectReviewRefs(options Options, config packConfig, processFile string, validation snapshot.ValidationResult, processData snapshot.ProcessSnapshotData) reviewRefs {
	expected := validation.Expected
	subjectRefs := []string{processFile}
	evidenceRefs := []string{}

	if config.LifecycleRef != "" {
		subjectRefs = append(subjectRefs, config.LifecycleRef)
	}
	for _, sr := range config.ReviewStandardRefs {
		if sr.Ref != config.LifecycleRef {
			subjectRefs = append(subjectRefs, sr.Ref)
		}
	}
	subjectRefs = appendExistingSpecRefs(subjectRefs, options.RepoRoot, options.ObjectType, options.Object, expected.TruthLayerRef)
	subjectRefs = appendSnapshotRefs(subjectRefs, expected)

	switch config.Pack {
	case PackUnitVerifyReadyToPromote:
		checkPath := filepath.Join(options.RepoRoot, snapshot.CheckResultFilePath(options.ObjectType, options.Object))
		if _, err := os.Stat(checkPath); err == nil {
			subjectRefs = append(subjectRefs, snapshot.CheckResultFilePath(options.ObjectType, options.Object))
		}
		planPath := filepath.Join(options.RepoRoot, snapshot.ActivePlanFilePath(options.Object))
		if _, err := os.Stat(planPath); err == nil {
			subjectRefs = append(subjectRefs, snapshot.ActivePlanFilePath(options.Object))
		}
		evidenceRefs = appendScalarRefs(evidenceRefs, processData, "evidence_refs")
		for _, entry := range processData.AcceptanceEvidence {
			evidenceRefs = appendSplitRefs(evidenceRefs, entry.EvidenceRefs)
		}
		for _, entry := range processData.RetirementEvidence {
			evidenceRefs = appendSplitRefs(evidenceRefs, entry.EvidenceRefs)
		}
		for _, entry := range processData.PackageDeltaVerification {
			evidenceRefs = appendSplitRefs(evidenceRefs, entry.EvidenceRefs)
		}
	case PackUnitStableVerifyAdvancing:
		subjectRefs = append(subjectRefs, "docs/specs/repository_mapping.md")
		evidenceRefs = appendScalarRefs(evidenceRefs, processData, "implementation_surface_refs", "evidence_refs")
		for _, entry := range processData.AcceptanceEvidence {
			evidenceRefs = appendSplitRefs(evidenceRefs, entry.EvidenceRefs)
		}
	case PackFreshnessTextDriftReuse:
		subjectRefs = append(subjectRefs, "framework/core/freshness.md")
		switch options.ProcessKind {
		case "plan":
			subjectRefs = append(subjectRefs, snapshot.ActivePlanFilePath(options.Object))
		case "check":
			subjectRefs = append(subjectRefs, snapshot.CheckResultFilePath(options.ObjectType, options.Object))
		case "verify":
			subjectRefs = append(subjectRefs, snapshot.VerifyResultFilePath(options.ObjectType, options.Object))
		case "stable_verify":
			subjectRefs = append(subjectRefs, snapshot.StableVerifyResultFilePath(options.ObjectType, options.Object))
		}
	}
	return reviewRefs{
		SubjectRefs:  sortedUnique(subjectRefs),
		EvidenceRefs: sortedUnique(evidenceRefs),
	}
}

func (refs reviewRefs) Combined() []string {
	var all []string
	all = append(all, refs.SubjectRefs...)
	all = append(all, refs.EvidenceRefs...)
	return sortedUnique(all)
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
		refs = appendSplitRefs(refs, value)
	}
	return refs
}

func appendSplitRefs(refs []string, value string) []string {
	for _, part := range strings.Split(value, ";") {
		ref := strings.TrimSpace(part)
		if ref == "" || ref == "none" {
			continue
		}
		refs = append(refs, ref)
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
	b.WriteString("You are the independent reviewer for this request. Do not modify repository files.\n")
	b.WriteString("Review Subject lists all files you may need to examine (paths only). ")
	b.WriteString("Evaluation Questions are the authoritative review criteria.\n\n")
	b.WriteString("Use Evaluation Questions as the authoritative review criteria.\n\n")
	b.WriteString("## Review Goal\n\n")
	b.WriteString(config.ReviewGoal)
	b.WriteString("\n\n")
	b.WriteString("\n## Allowed Inputs\n\n")
	writeBullets(&b, replacePlaceholder(config.AllowedInputs, options.Object))
	b.WriteString("\n## Forbidden Inputs\n\n")
	writeBullets(&b, replacePlaceholder(config.ForbiddenInputs, options.Object))
	b.WriteString("\n## Review Subject (artifact under review)\n\n")
	writeRefList(&b, refs.SubjectRefs)
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

func writeStandardRefs(b *strings.Builder, refs []reviewStandardRef) {
	if len(refs) == 0 {
		b.WriteString("- none\n")
		return
	}
	for _, ref := range refs {
		b.WriteString("- `")
		b.WriteString(ref.Ref)
		b.WriteString("`: ")
		b.WriteString(ref.Authority)
		b.WriteString("\n")
	}
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


func replacePlaceholder(items []string, object string) []string {
	var result []string
	for _, item := range items {
		result = append(result, strings.ReplaceAll(item, "{unit}", object))
	}
	return result
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
