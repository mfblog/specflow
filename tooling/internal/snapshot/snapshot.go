package snapshot

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Bingordinary/SpecFlow/tooling/internal/rulebinding"
	"github.com/Bingordinary/SpecFlow/tooling/internal/rulerefs"
	"github.com/Bingordinary/SpecFlow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/tooling/internal/statusfile"
)

type AppendixEntry struct {
	FileRef     string
	Fingerprint string
}

type RuleEntry struct {
	RuleID      string
	Layer       string
	FileRef     string
	VersionRef  string
	Fingerprint string
}

type ObjectSnapshotEntry struct {
	ObjectRef   string
	Layer       string
	FileRef     string
	VersionRef  string
	Fingerprint string
}

type RepositoryMappingEntry struct {
	FileRef     string
	VersionRef  string
	Fingerprint string
}

type AcceptanceItemEntry struct {
	ID                    string
	Target                string
	VerificationSurface   string
	ImplementationSurface string
	VerificationMethod    string
	PassCondition         string
	NotRunnableYet        string
	NotRunnableYetReason  string
}

type AcceptancePlanCoverageEntry struct {
	ID       string
	Coverage string
}

type AcceptanceEvidenceEntry struct {
	ID     string
	Status string
}

type Snapshot struct {
	ObjectType             string
	Object                 string
	Module                 string
	TruthLayerRef          string
	SpecFileRef            string
	SpecVersionRef         string
	SpecFingerprint        string
	ModuleAppendixSnapshot []AppendixEntry
	RepositoryMapping      RepositoryMappingEntry
	UnitSnapshot           []ObjectSnapshotEntry
	RuleSnapshot           []RuleEntry
	AcceptanceItemSet      []AcceptanceItemEntry
}

type ValidationResult struct {
	ObjectType   string
	Object       string
	ProcessKind  string
	ProcessFile  string
	Valid        bool
	Mismatches   []string
	FailureLayer string
	NextCommand  string
	Expected     Snapshot
}

type ProcessSnapshotData struct {
	ProcessKind            string
	ProcessFile            string
	PresentFields          map[string]bool
	Scalars                map[string]string
	ModuleAppendixSnapshot []AppendixEntry
	RepositoryMapping      RepositoryMappingEntry
	ModuleSnapshot         []ObjectSnapshotEntry
	RuleSnapshot           []RuleEntry
	AcceptanceItemSet      []AcceptanceItemEntry
	AcceptancePlanCoverage []AcceptancePlanCoverageEntry
	AcceptanceEvidence     []AcceptanceEvidenceEntry
}

var markdownLinkPattern = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)

var requiredUnitProcessSnapshotFields = map[string][]string{
	"check": {
		"object_type",
		"object_ref",
		"gate",
		"decision",
		"allow_next",
		"next_command",
		"blocking_summary",
		"coverage_summary",
		"truth_layer_ref",
		"truth_file_ref",
		"truth_version_ref",
		"truth_fingerprint",
		"acceptance_item_set",
		"unit_appendix_snapshot",
		"rule_snapshot",
	},
	"plan": {
		"spec_file_ref",
		"spec_version_ref",
		"spec_fingerprint",
		"unit_appendix_snapshot",
		"rule_snapshot",
		"acceptance_item_plan_coverage",
	},
	"verify": {
		"object_type",
		"object_ref",
		"gate",
		"decision",
		"allow_next",
		"next_command",
		"blocking_summary",
		"coverage_summary",
		"truth_layer_ref",
		"truth_file_ref",
		"truth_version_ref",
		"truth_fingerprint",
		"acceptance_item_set",
		"unit_appendix_snapshot",
		"verification_scope_ref",
		"rule_snapshot",
		"acceptance_item_evidence_matrix",
	},
}

var allowedAcceptanceEvidenceStatuses = map[string]bool{
	"pass":             true,
	"fail":             true,
	"partial":          true,
	"not_checked":      true,
	"not_runnable_yet": true,
}

var allowedVerificationSurfaces = map[string]bool{
	"public_api":     true,
	"internal_flow":  true,
	"error_handling": true,
	"eventing":       true,
	"storage":        true,
	"integration":    true,
	"manual_effect":  true,
}

func RebuildCurrent(repoRoot, module string) (Snapshot, error) {
	return RebuildCurrentObject(repoRoot, "unit", module)
}

func RebuildCurrentObject(repoRoot, objectType, object string) (Snapshot, error) {
	objectType = strings.TrimSpace(objectType)
	object = strings.TrimSpace(object)
	if objectType != "unit" {
		return Snapshot{}, fmt.Errorf("object type %q is not supported; only unit is supported", objectType)
	}
	status, err := statusfile.LookupObjectStatus(repoRoot, objectType, object)
	if err != nil {
		return Snapshot{}, err
	}

	mainSpecRef, err := specpaths.ObjectMainSpecFileRef(objectType, status.ActiveLayer, status.Object)
	if err != nil {
		return Snapshot{}, err
	}
	mainSpecAbs := filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef))
	mainSpecContent, err := os.ReadFile(mainSpecAbs)
	if err != nil {
		return Snapshot{}, fmt.Errorf("read %s: %w", mainSpecRef, err)
	}
	frontmatter, body, err := parseFrontmatter(string(mainSpecContent))
	if err != nil {
		return Snapshot{}, fmt.Errorf("%s: %w", mainSpecRef, err)
	}

	result := Snapshot{
		ObjectType:      objectType,
		Object:          object,
		Module:          object,
		TruthLayerRef:   status.ActiveLayer,
		SpecFileRef:     mainSpecRef,
		SpecFingerprint: hashNormalizedText(string(mainSpecContent)),
	}
	version := strings.TrimSpace(frontmatter["version"])
	if version == "" {
		return Snapshot{}, fmt.Errorf("%s: missing frontmatter.version", mainSpecRef)
	}
	result.SpecVersionRef = fmt.Sprintf("%s@%s", strings.TrimSuffix(filepath.Base(mainSpecRef), ".md"), version)

	appendixEntries, err := buildAppendixSnapshot(repoRoot, mainSpecRef, frontmatter, body)
	if err != nil {
		return Snapshot{}, err
	}
	result.ModuleAppendixSnapshot = appendixEntries

	unitRefs, _, err := parseFrontmatterNamedRefs(string(mainSpecContent), "unit_refs")
	if err != nil {
		return Snapshot{}, fmt.Errorf("%s: %w", mainSpecRef, err)
	}
	unitSnapshot, err := buildObjectDependencySnapshot(repoRoot, "unit", unitRefs)
	if err != nil {
		return Snapshot{}, err
	}
	result.UnitSnapshot = unitSnapshot

	ruleRefs, err := rulerefs.ParseObjectRuleRefs(mainSpecRef, string(mainSpecContent))
	if err != nil {
		return Snapshot{}, err
	}
	sharedEntries, err := buildRuleSnapshot(repoRoot, status.ActiveLayer, ruleRefs)
	if err != nil {
		return Snapshot{}, err
	}
	result.RuleSnapshot = sharedEntries

	acceptanceItems, err := buildAcceptanceItemSet(mainSpecRef, body)
	if err != nil {
		return Snapshot{}, err
	}
	result.AcceptanceItemSet = acceptanceItems
	return result, nil
}

func ValidateProcessFile(repoRoot, module, processKind string) (ValidationResult, error) {
	return ValidateProcessFileForObject(repoRoot, "unit", module, processKind)
}

func ValidateProcessFileForObject(repoRoot, objectType, object, processKind string) (ValidationResult, error) {
	expected, err := RebuildCurrentObject(repoRoot, objectType, object)
	if err != nil {
		return ValidationResult{}, err
	}
	requiredFields, ok := requiredFieldsForObjectProcess(objectType, processKind)
	if !ok {
		return ValidationResult{}, fmt.Errorf("process kind %q is not supported for object type %q", processKind, objectType)
	}

	processFile, err := ProcessFilePath(objectType, object, processKind)
	if err != nil {
		return ValidationResult{}, err
	}
	processAbs := filepath.Join(repoRoot, filepath.FromSlash(processFile))
	content, err := os.ReadFile(processAbs)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("read %s: %w", processFile, err)
	}

	actual, err := parseProcessSnapshot(string(content))
	if err != nil {
		return ValidationResult{}, fmt.Errorf("%s: %w", processFile, err)
	}

	result := ValidationResult{
		ObjectType:  objectType,
		Object:      object,
		ProcessKind: processKind,
		ProcessFile: processFile,
		Expected:    expected,
		Valid:       true,
	}

	for _, field := range requiredFields {
		if !actual.presentFields[field] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("missing required field: %s", field))
			continue
		}
		if actualValue, ok := actual.scalars[field]; ok && strings.TrimSpace(actualValue) == "" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("%s must not be empty", field))
		}
	}
	for _, field := range actual.invalidFields {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, field)
	}

	if processKind == "plan" {
		compareScalar(&result, "spec_file_ref", actual.scalars["spec_file_ref"], expected.SpecFileRef)
		compareScalar(&result, "spec_version_ref", actual.scalars["spec_version_ref"], expected.SpecVersionRef)
		compareScalar(&result, "spec_fingerprint", actual.scalars["spec_fingerprint"], expected.SpecFingerprint)

		if _, ok := actual.scalars["unit_appendix_snapshot"]; ok || actual.appendixPresent {
			actualAppendix := normalizeAppendixList(actual.appendixEntries)
			expectedAppendix := normalizeAppendixList(expected.ModuleAppendixSnapshot)
			if actualAppendix != expectedAppendix {
				result.Valid = false
				result.Mismatches = append(result.Mismatches, fmt.Sprintf("unit_appendix_snapshot mismatch: actual=%s expected=%s", actualAppendix, expectedAppendix))
			}
		}
		if _, ok := actual.scalars["rule_snapshot"]; ok || actual.sharedPresent {
			actualShared := normalizeSharedList(actual.sharedEntries)
			expectedShared := normalizeSharedList(expected.RuleSnapshot)
			if actualShared != expectedShared {
				result.Valid = false
				result.Mismatches = append(result.Mismatches, fmt.Sprintf("rule_snapshot mismatch: actual=%s expected=%s", actualShared, expectedShared))
			}
		}
		validateAcceptancePlanCoverage(&result, actual, expected.AcceptanceItemSet)
		finalizeValidationResult(&result)
		return result, nil
	}

	expectedGate, expectedNextCommand, err := expectedProcessRouting(objectType, processKind)
	if err != nil {
		return ValidationResult{}, err
	}
	compareScalar(&result, "object_type", actual.scalars["object_type"], objectType)
	compareScalar(&result, "object_ref", actual.scalars["object_ref"], expected.Object)
	compareScalar(&result, "gate", actual.scalars["gate"], expectedGate)
	compareScalar(&result, "decision", actual.scalars["decision"], "pass")
	compareScalar(&result, "allow_next", actual.scalars["allow_next"], "true")
	compareScalar(&result, "next_command", actual.scalars["next_command"], expectedNextCommand)
	compareScalar(&result, "truth_layer_ref", actual.scalars["truth_layer_ref"], expected.TruthLayerRef)
	compareScalar(&result, "truth_file_ref", actual.scalars["truth_file_ref"], expected.SpecFileRef)
	compareScalar(&result, "truth_version_ref", actual.scalars["truth_version_ref"], expected.SpecVersionRef)
	compareScalar(&result, "truth_fingerprint", actual.scalars["truth_fingerprint"], expected.SpecFingerprint)
	compareAcceptanceItemSet(&result, actual, expected.AcceptanceItemSet)
	if processKind == "verify" {
		validateAcceptanceEvidenceMatrix(&result, actual, expected.AcceptanceItemSet)
	}

	if actual.scalars["unit_appendix_snapshot"] != "" || actual.appendixPresent {
		actualAppendix := normalizeAppendixList(actual.appendixEntries)
		expectedAppendix := normalizeAppendixList(expected.ModuleAppendixSnapshot)
		if actualAppendix != expectedAppendix {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("unit_appendix_snapshot mismatch: actual=%s expected=%s", actualAppendix, expectedAppendix))
		}
	}
	if actual.modulePresent || actual.scalars["unit_snapshot"] != "" {
		actualUnits := normalizeObjectSnapshotList(actual.moduleEntries)
		expectedUnits := normalizeObjectSnapshotList(expected.UnitSnapshot)
		if actualUnits != expectedUnits {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("unit_snapshot mismatch: actual=%s expected=%s", actualUnits, expectedUnits))
		}
	}
	if _, ok := actual.scalars["rule_snapshot"]; ok || actual.sharedPresent {
		actualShared := normalizeSharedList(actual.sharedEntries)
		expectedShared := normalizeSharedList(expected.RuleSnapshot)
		if actualShared != expectedShared {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("rule_snapshot mismatch: actual=%s expected=%s", actualShared, expectedShared))
		}
	}

	finalizeValidationResult(&result)
	return result, nil
}

func requiredFieldsForObjectProcess(objectType, processKind string) ([]string, bool) {
	switch objectType {
	case "unit":
		fields, ok := requiredUnitProcessSnapshotFields[processKind]
		return fields, ok
	default:
		return nil, false
	}
}

func expectedProcessRouting(objectType, processKind string) (string, string, error) {
	switch objectType {
	case "unit":
		switch processKind {
		case "check":
			return "unit_check", "unit_plan", nil
		case "plan":
			return "unit_plan", "unit_impl", nil
		case "verify":
			return "unit_verify", "unit_promote", nil
		}
	}
	return "", "", fmt.Errorf("process kind %q is not supported for object type %q", processKind, objectType)
}

func LoadProcessSnapshot(repoRoot, objectType, object, processKind string) (ProcessSnapshotData, error) {
	processFile, err := ProcessFilePath(objectType, object, processKind)
	if err != nil {
		return ProcessSnapshotData{}, err
	}
	processAbs := filepath.Join(repoRoot, filepath.FromSlash(processFile))
	content, err := os.ReadFile(processAbs)
	if err != nil {
		return ProcessSnapshotData{}, fmt.Errorf("read %s: %w", processFile, err)
	}

	parsed, err := parseProcessSnapshot(string(content))
	if err != nil {
		return ProcessSnapshotData{}, fmt.Errorf("%s: %w", processFile, err)
	}

	scalars := make(map[string]string, len(parsed.scalars))
	for key, value := range parsed.scalars {
		scalars[key] = value
	}
	return ProcessSnapshotData{
		ProcessKind:            processKind,
		ProcessFile:            processFile,
		PresentFields:          copyStringBoolMap(parsed.presentFields),
		Scalars:                scalars,
		ModuleAppendixSnapshot: append([]AppendixEntry(nil), parsed.appendixEntries...),
		ModuleSnapshot:         append([]ObjectSnapshotEntry(nil), parsed.moduleEntries...),
		RuleSnapshot:           append([]RuleEntry(nil), parsed.sharedEntries...),
		RepositoryMapping:      parsed.repositoryMapping,
		AcceptanceItemSet:      append([]AcceptanceItemEntry(nil), parsed.acceptanceItemEntries...),
		AcceptancePlanCoverage: append([]AcceptancePlanCoverageEntry(nil), parsed.acceptancePlanEntries...),
		AcceptanceEvidence:     append([]AcceptanceEvidenceEntry(nil), parsed.acceptanceEvidenceEntries...),
	}, nil
}

func Render(snapshot Snapshot) string {
	objectType := snapshot.ObjectType
	if objectType == "" {
		objectType = "unit"
	}
	object := snapshot.Object
	if object == "" {
		object = snapshot.Module
	}
	lines := []string{
		fmt.Sprintf("object_type: %s", objectType),
		fmt.Sprintf("object_ref: %s", object),
		fmt.Sprintf("truth_layer_ref: %s", snapshot.TruthLayerRef),
		fmt.Sprintf("truth_file_ref: %s", snapshot.SpecFileRef),
		fmt.Sprintf("truth_version_ref: %s", snapshot.SpecVersionRef),
		fmt.Sprintf("truth_fingerprint: %s", snapshot.SpecFingerprint),
		"acceptance_item_set:",
	}
	lines = append(lines, renderAcceptanceItemLines(snapshot.AcceptanceItemSet)...)
	lines = append(lines, "unit_appendix_snapshot:")
	lines = append(lines, renderAppendixLines(snapshot.ModuleAppendixSnapshot)...)
	lines = append(lines, "unit_snapshot:")
	lines = append(lines, renderObjectSnapshotLines("unit", snapshot.UnitSnapshot)...)
	lines = append(lines, "rule_snapshot:")
	lines = append(lines, renderSharedLines(snapshot.RuleSnapshot)...)
	return strings.Join(lines, "\n")
}

func buildAcceptanceItemSet(mainSpecRef, body string) ([]AcceptanceItemEntry, error) {
	parsed, err := parseProcessSnapshot(body)
	if err != nil {
		return nil, err
	}
	if !parsed.presentFields["acceptance_item_set"] {
		if isStableMainSpecRef(mainSpecRef) {
			return nil, nil
		}
		if hasAcceptanceSection(body) {
			return nil, fmt.Errorf("%s: acceptance section must define acceptance_item_set", mainSpecRef)
		}
		return nil, fmt.Errorf("%s: main Spec must define Testability / Acceptance Criteria with acceptance_item_set", mainSpecRef)
	}
	entries, err := acceptanceMainItemEntriesFromParsed(parsed)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", mainSpecRef, err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("%s: acceptance_item_set must contain at least one item", mainSpecRef)
	}
	return entries, nil
}

func BuildAcceptanceItemSetFromBody(mainSpecRef, body string) ([]AcceptanceItemEntry, error) {
	return buildAcceptanceItemSet(mainSpecRef, body)
}

func BuildRepositoryMappingSnapshot(repoRoot string) (RepositoryMappingEntry, error) {
	fileRef := specpaths.RepositoryMappingFileRef
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(fileRef)))
	if err != nil {
		return RepositoryMappingEntry{}, fmt.Errorf("read %s: %w", fileRef, err)
	}
	frontmatter, _, err := parseFrontmatter(string(content))
	if err != nil {
		return RepositoryMappingEntry{}, fmt.Errorf("%s: %w", fileRef, err)
	}
	version := strings.TrimSpace(frontmatter["version"])
	if version == "" {
		return RepositoryMappingEntry{}, fmt.Errorf("%s: missing frontmatter.version", fileRef)
	}
	return RepositoryMappingEntry{
		FileRef:     fileRef,
		VersionRef:  fmt.Sprintf("repository_mapping@%s", version),
		Fingerprint: hashNormalizedText(string(content)),
	}, nil
}

func hasAcceptanceSection(body string) bool {
	for _, line := range strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		lower := strings.ToLower(trimmed)
		if strings.Contains(lower, "acceptance criteria") || strings.Contains(trimmed, "验收标准") {
			return true
		}
	}
	return false
}

func isStableMainSpecRef(mainSpecRef string) bool {
	return strings.HasPrefix(mainSpecRef, "docs/specs/units/stable/")
}

func buildAppendixSnapshot(repoRoot, mainSpecRef string, frontmatter map[string]string, body string) ([]AppendixEntry, error) {
	mainDir := filepath.Dir(filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)))
	currentLayer := mainSpecLayer(mainSpecRef)
	currentObject, ownerKey, err := mainSpecObject(mainSpecRef)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	entries := []AppendixEntry{}
	addAppendix := func(relPath string) error {
		relPath = filepath.ToSlash(filepath.Clean(relPath))
		if relPath == "." || filepath.IsAbs(relPath) || filepath.Ext(relPath) != ".md" {
			return nil
		}
		if relPath == mainSpecRef {
			return nil
		}
		if seen[relPath] {
			return nil
		}
		seen[relPath] = true

		absPath := filepath.Join(repoRoot, filepath.FromSlash(relPath))
		content, err := os.ReadFile(absPath)
		if err != nil {
			return fmt.Errorf("read appendix %s: %w", relPath, err)
		}
		frontmatter, _, err := parseFrontmatter(string(content))
		if err != nil {
			return fmt.Errorf("%s: %w", relPath, err)
		}
		if layer := strings.TrimSpace(frontmatter["layer"]); layer != "" && layer != currentLayer {
			return fmt.Errorf("%s: appendix layer %q does not match main spec layer %q", relPath, layer, currentLayer)
		}
		if owner := strings.TrimSpace(frontmatter[ownerKey]); owner != "" && owner != currentObject {
			return fmt.Errorf("%s: appendix %s %q does not match main spec %s %q", relPath, ownerKey, owner, ownerKey, currentObject)
		}
		entries = append(entries, AppendixEntry{
			FileRef:     relPath,
			Fingerprint: hashNormalizedText(string(content)),
		})
		return nil
	}
	if evidenceRef := strings.TrimSpace(frontmatter["evidence_appendix_ref"]); evidenceRef != "" && evidenceRef != "none" {
		relPath, err := resolveAppendixRef(repoRoot, mainDir, evidenceRef)
		if err != nil {
			return nil, err
		}
		if !strings.Contains(relPath, "/appendix/") {
			return nil, fmt.Errorf("%s: evidence appendix ref %s is not under an appendix directory", mainSpecRef, relPath)
		}
		if err := addAppendix(relPath); err != nil {
			return nil, err
		}
	}
	for _, destination := range markdownLinkPattern.FindAllStringSubmatch(body, -1) {
		if len(destination) != 2 {
			continue
		}
		linkDestination := strings.TrimSpace(destination[1])
		if linkDestination == "" || strings.HasPrefix(linkDestination, "/") || strings.Contains(linkDestination, "://") {
			continue
		}
		absPath := filepath.Clean(filepath.Join(mainDir, filepath.FromSlash(linkDestination)))
		relWithinLayerRoot, err := filepath.Rel(mainDir, absPath)
		if err != nil {
			return nil, err
		}
		relWithinLayerRoot = filepath.ToSlash(relWithinLayerRoot)
		if strings.HasPrefix(relWithinLayerRoot, "../") || relWithinLayerRoot == ".." || filepath.Ext(relWithinLayerRoot) != ".md" {
			continue
		}
		relPath, err := filepath.Rel(repoRoot, absPath)
		if err != nil {
			return nil, err
		}
		relPath = filepath.ToSlash(relPath)
		if relPath == mainSpecRef {
			continue
		}
		if filepath.Dir(relWithinLayerRoot) == "." {
			return nil, fmt.Errorf("%s: module-local supporting file %s remains in the layer root; this is directory drift", mainSpecRef, relPath)
		}
		if err := addAppendix(relPath); err != nil {
			return nil, err
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].FileRef < entries[j].FileRef
	})
	return entries, nil
}

func BuildAppendixSnapshot(repoRoot, mainSpecRef, body string) ([]AppendixEntry, error) {
	return buildAppendixSnapshot(repoRoot, mainSpecRef, nil, body)
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

func buildRuleSnapshot(repoRoot, moduleLayer string, refs []string) ([]RuleEntry, error) {
	entries, err := buildStableGlobalRuleEntries(repoRoot)
	if err != nil {
		return nil, err
	}
	if len(refs) == 0 {
		return sortRuleEntries(entries), nil
	}
	for _, ref := range refs {
		resolved, err := rulebinding.ResolveRef(repoRoot, moduleLayer, ref)
		if err != nil {
			return nil, err
		}
		entries = append(entries, RuleEntry{
			RuleID:      resolved.RuleID,
			Layer:       resolved.Layer,
			FileRef:     resolved.FileRef,
			VersionRef:  resolved.VersionRef,
			Fingerprint: hashNormalizedText(resolved.Content),
		})
	}
	return sortRuleEntries(entries), nil
}

func buildObjectDependencySnapshot(repoRoot, expectedObjectType string, refs []string) ([]ObjectSnapshotEntry, error) {
	entries := make([]ObjectSnapshotEntry, 0, len(refs))
	for _, ref := range refs {
		entry, err := resolveObjectVersionRef(repoRoot, expectedObjectType, ref)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return sortObjectSnapshotEntries(entries), nil
}

func resolveObjectVersionRef(repoRoot, expectedObjectType, ref string) (ObjectSnapshotEntry, error) {
	ref = strings.Trim(strings.TrimSpace(ref), "`")
	prefix, version, ok := strings.Cut(ref, "@")
	if !ok || strings.TrimSpace(version) == "" {
		return ObjectSnapshotEntry{}, fmt.Errorf("invalid %s ref %q", expectedObjectType, ref)
	}
	layer := ""
	object := ""
	switch expectedObjectType {
	case "unit":
		switch {
		case strings.HasPrefix(prefix, "c_unit_"):
			layer = "candidate"
			object = strings.TrimPrefix(prefix, "c_unit_")
		case strings.HasPrefix(prefix, "s_unit_"):
			layer = "stable"
			object = strings.TrimPrefix(prefix, "s_unit_")
		}
	}
	if layer == "" || object == "" {
		return ObjectSnapshotEntry{}, fmt.Errorf("invalid %s ref %q", expectedObjectType, ref)
	}
	if expectedObjectType == "unit" && layer != "stable" {
		return ObjectSnapshotEntry{}, fmt.Errorf("unit_refs must reference stable units; got %q", ref)
	}
	fileRef, err := specpaths.ObjectMainSpecFileRef(expectedObjectType, layer, object)
	if err != nil {
		return ObjectSnapshotEntry{}, err
	}
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(fileRef)))
	if err != nil {
		return ObjectSnapshotEntry{}, fmt.Errorf("read %s: %w", fileRef, err)
	}
	frontmatter, _, err := parseFrontmatter(string(content))
	if err != nil {
		return ObjectSnapshotEntry{}, fmt.Errorf("%s: %w", fileRef, err)
	}
	currentVersion := strings.TrimSpace(frontmatter["version"])
	if currentVersion == "" {
		return ObjectSnapshotEntry{}, fmt.Errorf("%s: missing frontmatter.version", fileRef)
	}
	expectedVersionRef := fmt.Sprintf("%s@%s", strings.TrimSuffix(filepath.Base(fileRef), ".md"), currentVersion)
	if ref != expectedVersionRef {
		return ObjectSnapshotEntry{}, fmt.Errorf("%s ref %q does not match current version %q", expectedObjectType, ref, expectedVersionRef)
	}
	return ObjectSnapshotEntry{
		ObjectRef:   object,
		Layer:       layer,
		FileRef:     fileRef,
		VersionRef:  expectedVersionRef,
		Fingerprint: hashNormalizedText(string(content)),
	}, nil
}

func buildStableGlobalRuleEntries(repoRoot string) ([]RuleEntry, error) {
	matches, err := filepath.Glob(filepath.Join(repoRoot, filepath.FromSlash("docs/specs/rules/stable/s_g_rule_*.md")))
	if err != nil {
		return nil, err
	}
	entries := []RuleEntry{}
	for _, absPath := range matches {
		rel, err := filepath.Rel(repoRoot, absPath)
		if err != nil {
			return nil, err
		}
		fileRef := filepath.ToSlash(rel)
		content, err := os.ReadFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", fileRef, err)
		}
		hasBoundObjects, err := rulerefs.HasRuleBoundObjects(fileRef, string(content))
		if err != nil {
			return nil, err
		}
		if hasBoundObjects {
			return nil, fmt.Errorf("%s: bound_objects is forbidden; derive consumers from current-layer rule_refs", fileRef)
		}
		frontmatter, _, err := parseFrontmatter(string(content))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fileRef, err)
		}
		ruleID := strings.TrimSpace(frontmatter["rule_id"])
		ruleScope := strings.TrimSpace(frontmatter["rule_scope"])
		layer := strings.TrimSpace(frontmatter["layer"])
		version := strings.TrimSpace(frontmatter["rule_version"])
		if ruleID == "" || ruleScope != "global" || layer != "stable" || version == "" {
			return nil, fmt.Errorf("%s: stable global rule must record rule_id, rule_scope=global, layer=stable, and rule_version", fileRef)
		}
		prefix := strings.TrimSuffix(filepath.Base(fileRef), ".md")
		entries = append(entries, RuleEntry{
			RuleID:      ruleID,
			Layer:       layer,
			FileRef:     fileRef,
			VersionRef:  fmt.Sprintf("%s@%s", prefix, version),
			Fingerprint: hashNormalizedText(string(content)),
		})
	}
	return entries, nil
}

func sortRuleEntries(entries []RuleEntry) []RuleEntry {
	items := append([]RuleEntry(nil), entries...)
	sort.Slice(items, func(i, j int) bool {
		if items[i].RuleID != items[j].RuleID {
			return items[i].RuleID < items[j].RuleID
		}
		if items[i].Layer != items[j].Layer {
			return items[i].Layer < items[j].Layer
		}
		return items[i].FileRef < items[j].FileRef
	})
	return items
}

func sortObjectSnapshotEntries(entries []ObjectSnapshotEntry) []ObjectSnapshotEntry {
	items := append([]ObjectSnapshotEntry(nil), entries...)
	sort.Slice(items, func(i, j int) bool {
		if items[i].ObjectRef != items[j].ObjectRef {
			return items[i].ObjectRef < items[j].ObjectRef
		}
		if items[i].Layer != items[j].Layer {
			return items[i].Layer < items[j].Layer
		}
		return items[i].FileRef < items[j].FileRef
	})
	return items
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
	body := strings.Join(lines[endIdx+1:], "\n")
	return result, body, nil
}

func parseFrontmatterNamedRefs(content, fieldName string) ([]string, bool, error) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, false, nil
	}
	endIdx := -1
	for idx := 1; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "---" {
			endIdx = idx
			break
		}
	}
	if endIdx == -1 {
		return nil, false, nil
	}

	for idx := 1; idx < endIdx; idx++ {
		trimmed := strings.TrimSpace(lines[idx])
		right, matched, err := parseNamedFieldLine(trimmed, fieldName)
		if err != nil {
			return nil, false, err
		}
		if !matched {
			continue
		}
		if right == "`none`" || right == "none" {
			return nil, true, nil
		}
		if right != "" {
			return nil, false, fmt.Errorf("%s must use literal none or a YAML list", fieldName)
		}
		refs := []string{}
		seen := map[string]bool{}
		for next := idx + 1; next < endIdx; next++ {
			nextTrimmed := strings.TrimSpace(lines[next])
			if nextTrimmed == "" || strings.HasPrefix(nextTrimmed, "#") {
				continue
			}
			if !strings.HasPrefix(nextTrimmed, "- ") {
				break
			}
			ref := strings.TrimSpace(strings.TrimPrefix(nextTrimmed, "- "))
			ref = strings.Trim(strings.Trim(ref, "`"), "\"'")
			if ref == "" {
				return nil, false, fmt.Errorf("%s contains an empty item", fieldName)
			}
			if seen[ref] {
				return nil, false, fmt.Errorf("%s contains duplicate item %q", fieldName, ref)
			}
			seen[ref] = true
			refs = append(refs, ref)
		}
		if len(refs) == 0 {
			return nil, false, fmt.Errorf("%s must not be an empty list", fieldName)
		}
		if err := validateOrderedRefs(fieldName, refs); err != nil {
			return nil, false, err
		}
		return refs, true, nil
	}
	return nil, false, nil
}

func hashNormalizedText(content string) string {
	text := strings.ReplaceAll(content, "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")
	text += "\n"
	sum := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", sum)
}

func parseNamedRefs(body, fieldName string) ([]string, bool, error) {
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		right, matched, err := parseNamedFieldLine(trimmed, fieldName)
		if err != nil {
			return nil, false, err
		}
		if !matched {
			continue
		}
		if right == "`none`" || right == "none" {
			return nil, true, nil
		}
		if right != "" {
			return nil, false, fmt.Errorf("%s must use literal none or a markdown list", fieldName)
		}
		refs := []string{}
		seen := map[string]bool{}
		for next := idx + 1; next < len(lines); next++ {
			nextTrimmed := strings.TrimSpace(lines[next])
			if nextTrimmed == "" {
				continue
			}
			if strings.HasPrefix(nextTrimmed, "## ") || regexp.MustCompile(`^\d+\.`).MatchString(nextTrimmed) {
				break
			}
			if _, matched, _ := parseNamedFieldLine(nextTrimmed, "rule_refs"); matched && fieldName != "rule_refs" {
				break
			}
			if _, matched, _ := parseNamedFieldLine(nextTrimmed, "unit_refs"); matched && fieldName != "unit_refs" {
				break
			}
			if !strings.HasPrefix(nextTrimmed, "- ") {
				return nil, false, fmt.Errorf("%s must be a markdown list", fieldName)
			}
			ref := strings.TrimSpace(strings.TrimPrefix(nextTrimmed, "- "))
			ref = strings.Trim(ref, "`")
			if ref == "" {
				return nil, false, fmt.Errorf("%s contains an empty item", fieldName)
			}
			if seen[ref] {
				return nil, false, fmt.Errorf("%s contains duplicate item %q", fieldName, ref)
			}
			seen[ref] = true
			refs = append(refs, ref)
		}
		if len(refs) == 0 {
			return nil, false, fmt.Errorf("%s must not be an empty list", fieldName)
		}
		if err := validateOrderedRefs(fieldName, refs); err != nil {
			return nil, false, err
		}
		return refs, true, nil
	}
	return nil, false, nil
}

func validateOrderedRuleRefs(refs []string) error {
	return validateOrderedRefs("rule_refs", refs)
}

func validateOrderedRefs(fieldName string, refs []string) error {
	if len(refs) < 2 {
		return nil
	}
	expected := append([]string(nil), refs...)
	sort.Strings(expected)
	for idx := range refs {
		if refs[idx] != expected[idx] {
			return fmt.Errorf("%s must be sorted by exact ref string in ascending lexical order", fieldName)
		}
	}
	return nil
}

func parseNamedFieldLine(trimmed, fieldName string) (string, bool, error) {
	parts := strings.SplitN(trimmed, ":", 2)
	if len(parts) != 2 {
		return "", false, nil
	}
	left := normalizeFieldKey(strings.TrimSpace(parts[0]))
	if left != fieldName {
		return "", false, nil
	}
	return strings.TrimSpace(parts[1]), true, nil
}

func normalizeFieldKey(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "`", "")
	if idx := strings.Index(value, ". "); idx > 0 {
		allDigits := true
		for _, ch := range value[:idx] {
			if ch < '0' || ch > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			value = value[idx+2:]
		}
	}
	return strings.TrimSpace(value)
}

func mainSpecLayer(mainSpecRef string) string {
	base := strings.TrimSuffix(filepath.Base(mainSpecRef), ".md")
	if strings.HasPrefix(base, "c_") {
		return "candidate"
	}
	return "stable"
}

func mainSpecObject(mainSpecRef string) (string, string, error) {
	base := strings.TrimSuffix(filepath.Base(mainSpecRef), ".md")
	switch {
	case strings.HasPrefix(base, "c_unit_"):
		return strings.TrimPrefix(base, "c_unit_"), "unit", nil
	case strings.HasPrefix(base, "s_unit_"):
		return strings.TrimPrefix(base, "s_unit_"), "unit", nil
	default:
		return "", "", fmt.Errorf("unsupported main spec file ref %q", mainSpecRef)
	}
}

type processSnapshot struct {
	presentFields             map[string]bool
	scalars                   map[string]string
	appendixEntries           []AppendixEntry
	appendixPresent           bool
	repositoryMapping         RepositoryMappingEntry
	repositoryMappingPresent  bool
	moduleEntries             []ObjectSnapshotEntry
	modulePresent             bool
	sharedEntries             []RuleEntry
	sharedPresent             bool
	acceptanceItemEntries     []AcceptanceItemEntry
	acceptanceItemPresent     bool
	acceptancePlanEntries     []AcceptancePlanCoverageEntry
	acceptancePlanPresent     bool
	acceptanceEvidenceEntries []AcceptanceEvidenceEntry
	acceptanceEvidencePresent bool
	invalidFields             []string
}

func parseProcessSnapshot(content string) (processSnapshot, error) {
	result := processSnapshot{
		presentFields: map[string]bool{},
		scalars:       map[string]string{},
	}
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	currentList := ""
	currentIndex := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := leadingSpaceCount(line)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			continue
		}

		if indent == 0 {
			currentList = ""
			currentIndex = -1
			key, value, ok := parseSnapshotFieldLine(trimmed)
			if !ok {
				continue
			}
			result.presentFields[key] = true
			if value == "" {
				switch key {
				case "unit_appendix_snapshot":
					result.appendixPresent = true
					currentList = key
				case "repository_mapping_snapshot":
					result.repositoryMappingPresent = true
					currentList = key
				case "unit_snapshot":
					result.modulePresent = true
					currentList = key
				case "rule_snapshot":
					result.sharedPresent = true
					currentList = key
				case "acceptance_item_set":
					result.acceptanceItemPresent = true
					currentList = key
				case "acceptance_item_plan_coverage":
					result.acceptancePlanPresent = true
					currentList = key
				case "acceptance_item_evidence_matrix":
					result.acceptanceEvidencePresent = true
					currentList = key
				}
				continue
			}
			result.scalars[key] = value
			continue
		}

		if indent >= 2 {
			key, value, ok := parseSnapshotFieldLine(trimmed)
			if !ok {
				continue
			}
			listItemStart := strings.HasPrefix(trimmed, "- ")
			switch currentList {
			case "unit_appendix_snapshot":
				if !allowedAppendixSnapshotField(key) {
					result.invalidFields = append(result.invalidFields, fmt.Sprintf("unsupported field: %s.%s", currentList, key))
					continue
				}
				if currentIndex < 0 || (listItemStart && key == "file_ref") {
					result.appendixEntries = append(result.appendixEntries, AppendixEntry{})
					currentIndex = len(result.appendixEntries) - 1
				}
				assignAppendixField(&result.appendixEntries[currentIndex], key, value)
			case "repository_mapping_snapshot":
				assignRepositoryMappingField(&result.repositoryMapping, key, value)
			case "unit_snapshot":
				if currentIndex < 0 || (listItemStart && key == "unit") {
					result.moduleEntries = append(result.moduleEntries, ObjectSnapshotEntry{})
					currentIndex = len(result.moduleEntries) - 1
				}
				assignObjectSnapshotField(&result.moduleEntries[currentIndex], key, value)
			case "rule_snapshot":
				if currentIndex < 0 || (listItemStart && key == "rule_id") {
					result.sharedEntries = append(result.sharedEntries, RuleEntry{})
					currentIndex = len(result.sharedEntries) - 1
				}
				assignSharedField(&result.sharedEntries[currentIndex], key, value)
			case "acceptance_item_set":
				if currentIndex < 0 || (listItemStart && key == "id") {
					result.acceptanceItemEntries = append(result.acceptanceItemEntries, AcceptanceItemEntry{})
					currentIndex = len(result.acceptanceItemEntries) - 1
				}
				assignAcceptanceItemField(&result.acceptanceItemEntries[currentIndex], key, value)
			case "acceptance_item_plan_coverage":
				if currentIndex < 0 || (listItemStart && key == "id") {
					result.acceptancePlanEntries = append(result.acceptancePlanEntries, AcceptancePlanCoverageEntry{})
					currentIndex = len(result.acceptancePlanEntries) - 1
				}
				assignAcceptancePlanCoverageField(&result.acceptancePlanEntries[currentIndex], key, value)
			case "acceptance_item_evidence_matrix":
				if currentIndex < 0 || (listItemStart && key == "id") {
					result.acceptanceEvidenceEntries = append(result.acceptanceEvidenceEntries, AcceptanceEvidenceEntry{})
					currentIndex = len(result.acceptanceEvidenceEntries) - 1
				}
				assignAcceptanceEvidenceField(&result.acceptanceEvidenceEntries[currentIndex], key, value)
			}
		}
	}

	if raw, ok := result.scalars["unit_appendix_snapshot"]; ok && raw == "none" {
		result.appendixPresent = true
		result.appendixEntries = nil
	}
	if raw, ok := result.scalars["repository_mapping_snapshot"]; ok && raw == "none" {
		result.repositoryMappingPresent = true
		result.repositoryMapping = RepositoryMappingEntry{}
	}
	if raw, ok := result.scalars["unit_snapshot"]; ok && raw == "none" {
		result.modulePresent = true
		result.moduleEntries = nil
	}
	if raw, ok := result.scalars["rule_snapshot"]; ok && raw == "none" {
		result.sharedPresent = true
		result.sharedEntries = nil
	}
	if raw, ok := result.scalars["acceptance_item_set"]; ok && raw == "none" {
		result.acceptanceItemPresent = true
		result.acceptanceItemEntries = nil
	}
	if raw, ok := result.scalars["acceptance_item_plan_coverage"]; ok && raw == "none" {
		result.acceptancePlanPresent = true
		result.acceptancePlanEntries = nil
	}
	if raw, ok := result.scalars["acceptance_item_evidence_matrix"]; ok && raw == "none" {
		result.acceptanceEvidencePresent = true
		result.acceptanceEvidenceEntries = nil
	}
	return result, nil
}

func parseSnapshotFieldLine(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "- ")
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	key := normalizeFieldKey(strings.TrimSpace(parts[0]))
	if key == "" {
		return "", "", false
	}
	value := strings.Trim(strings.TrimSpace(parts[1]), "`")
	return key, value, true
}

func leadingSpaceCount(line string) int {
	count := 0
	for count < len(line) && line[count] == ' ' {
		count++
	}
	return count
}

func splitKeyValue(line string) (string, string) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", ""
	}
	key := strings.TrimSpace(parts[0])
	value := strings.Trim(strings.TrimSpace(parts[1]), "`")
	return key, value
}

func allowedAppendixSnapshotField(key string) bool {
	switch key {
	case "file_ref", "fingerprint":
		return true
	default:
		return false
	}
}

func assignAppendixField(entry *AppendixEntry, key, value string) {
	switch key {
	case "file_ref":
		entry.FileRef = value
	case "fingerprint":
		entry.Fingerprint = value
	}
}

func assignSharedField(entry *RuleEntry, key, value string) {
	switch key {
	case "rule_id":
		entry.RuleID = value
	case "layer":
		entry.Layer = value
	case "file_ref":
		entry.FileRef = value
	case "version_ref":
		entry.VersionRef = value
	case "fingerprint":
		entry.Fingerprint = value
	}
}

func assignRepositoryMappingField(entry *RepositoryMappingEntry, key, value string) {
	switch key {
	case "file_ref":
		entry.FileRef = value
	case "version_ref":
		entry.VersionRef = value
	case "fingerprint":
		entry.Fingerprint = value
	}
}

func assignObjectSnapshotField(entry *ObjectSnapshotEntry, key, value string) {
	switch key {
	case "unit":
		entry.ObjectRef = value
	case "layer":
		entry.Layer = value
	case "file_ref":
		entry.FileRef = value
	case "version_ref":
		entry.VersionRef = value
	case "fingerprint":
		entry.Fingerprint = value
	}
}

func assignAcceptanceItemField(entry *AcceptanceItemEntry, key, value string) {
	switch key {
	case "id":
		entry.ID = value
	case "target":
		entry.Target = value
	case "verification_surface":
		entry.VerificationSurface = value
	case "implementation_surface":
		entry.ImplementationSurface = value
	case "verification_method":
		entry.VerificationMethod = value
	case "pass_condition":
		entry.PassCondition = value
	case "not_runnable_yet":
		entry.NotRunnableYet = value
	case "not_runnable_yet_reason":
		entry.NotRunnableYetReason = value
	}
}

func assignAcceptancePlanCoverageField(entry *AcceptancePlanCoverageEntry, key, value string) {
	switch key {
	case "id":
		entry.ID = value
	case "coverage":
		entry.Coverage = value
	}
}

func assignAcceptanceEvidenceField(entry *AcceptanceEvidenceEntry, key, value string) {
	switch key {
	case "id":
		entry.ID = value
	case "status":
		entry.Status = value
	}
}

func compareScalar(result *ValidationResult, field, actual, expected string) {
	if actual == "" {
		return
	}
	if actual != expected {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, fmt.Sprintf("%s mismatch: actual=%s expected=%s", field, actual, expected))
	}
}

func compareAcceptanceItemSet(result *ValidationResult, actual processSnapshot, expected []AcceptanceItemEntry) {
	actualEntries, err := acceptanceItemEntriesFromParsed(actual)
	if err != nil {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, "acceptance_item_set invalid: "+err.Error())
		return
	}
	actualValue := normalizeAcceptanceItemList(actualEntries)
	expectedValue := normalizeAcceptanceItemList(expected)
	if actualValue != expectedValue {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, fmt.Sprintf("acceptance_item_set mismatch: actual=%s expected=%s", actualValue, expectedValue))
	}
}

func validateAcceptancePlanCoverage(result *ValidationResult, actual processSnapshot, expected []AcceptanceItemEntry) {
	actualEntries, err := acceptancePlanCoverageEntriesFromParsed(actual)
	if err != nil {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, "acceptance_item_plan_coverage invalid: "+err.Error())
		return
	}
	expectedIDs := acceptanceItemIDSet(expected)
	actualIDs := map[string]bool{}
	for _, entry := range actualEntries {
		if actualIDs[entry.ID] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("acceptance_item_plan_coverage duplicate id: %s", entry.ID))
			continue
		}
		actualIDs[entry.ID] = true
		if !expectedIDs[entry.ID] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("acceptance_item_plan_coverage unknown id: %s", entry.ID))
		}
	}
	for _, item := range normalizeAcceptanceItemEntries(expected) {
		if !actualIDs[item.ID] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("acceptance_item_plan_coverage missing id: %s", item.ID))
		}
	}
}

func validateAcceptanceEvidenceMatrix(result *ValidationResult, actual processSnapshot, expected []AcceptanceItemEntry) {
	actualEntries, err := acceptanceEvidenceEntriesFromParsed(actual)
	if err != nil {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, "acceptance_item_evidence_matrix invalid: "+err.Error())
		return
	}
	expectedByID := acceptanceItemsByID(expected)
	actualIDs := map[string]bool{}
	for _, entry := range actualEntries {
		if actualIDs[entry.ID] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("acceptance_item_evidence_matrix duplicate id: %s", entry.ID))
			continue
		}
		actualIDs[entry.ID] = true
		expectedItem, ok := expectedByID[entry.ID]
		if !ok {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("acceptance_item_evidence_matrix unknown id: %s", entry.ID))
			continue
		}
		if !allowedAcceptanceEvidenceStatuses[entry.Status] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("acceptance_item_evidence_matrix invalid status for %s: %s", entry.ID, entry.Status))
			continue
		}
		if expectedItem.NotRunnableYet == "yes" && entry.Status != "not_runnable_yet" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("acceptance_item_evidence_matrix status for %s must be not_runnable_yet", entry.ID))
		}
		if expectedItem.NotRunnableYet == "no" && entry.Status == "not_runnable_yet" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("acceptance_item_evidence_matrix status for %s cannot be not_runnable_yet", entry.ID))
		}
	}
	for _, item := range normalizeAcceptanceItemEntries(expected) {
		if !actualIDs[item.ID] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("acceptance_item_evidence_matrix missing id: %s", item.ID))
		}
	}
}

func acceptanceItemEntriesFromParsed(parsed processSnapshot) ([]AcceptanceItemEntry, error) {
	if raw, ok := parsed.scalars["acceptance_item_set"]; ok && raw != "none" {
		return nil, fmt.Errorf("must be literal none or a list")
	}
	items := append([]AcceptanceItemEntry(nil), parsed.acceptanceItemEntries...)
	if len(items) == 0 {
		return nil, nil
	}
	if err := validateAcceptanceItemEntries(items); err != nil {
		return nil, err
	}
	return snapshotAcceptanceItemEntries(items), nil
}

func acceptanceMainItemEntriesFromParsed(parsed processSnapshot) ([]AcceptanceItemEntry, error) {
	if raw, ok := parsed.scalars["acceptance_item_set"]; ok && raw != "none" {
		return nil, fmt.Errorf("must be literal none or a list")
	}
	items := append([]AcceptanceItemEntry(nil), parsed.acceptanceItemEntries...)
	if len(items) == 0 {
		return nil, nil
	}
	if err := validateAcceptanceItemEntries(items); err != nil {
		return nil, err
	}
	for _, item := range items {
		if strings.TrimSpace(item.Target) == "" ||
			strings.TrimSpace(item.ImplementationSurface) == "" ||
			strings.TrimSpace(item.VerificationMethod) == "" ||
			strings.TrimSpace(item.PassCondition) == "" {
			return nil, fmt.Errorf("acceptance item %s must include target, implementation_surface, verification_method, and pass_condition", item.ID)
		}
		if item.NotRunnableYet == "yes" && strings.TrimSpace(item.NotRunnableYetReason) == "" {
			return nil, fmt.Errorf("not_runnable_yet acceptance item %s must include not_runnable_yet_reason", item.ID)
		}
	}
	return normalizeAcceptanceItemEntries(items), nil
}

func acceptancePlanCoverageEntriesFromParsed(parsed processSnapshot) ([]AcceptancePlanCoverageEntry, error) {
	if raw, ok := parsed.scalars["acceptance_item_plan_coverage"]; ok && raw != "none" {
		return nil, fmt.Errorf("must be literal none or a list")
	}
	items := append([]AcceptancePlanCoverageEntry(nil), parsed.acceptancePlanEntries...)
	for _, item := range items {
		if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Coverage) == "" {
			return nil, fmt.Errorf("each item must include id and coverage")
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func acceptanceEvidenceEntriesFromParsed(parsed processSnapshot) ([]AcceptanceEvidenceEntry, error) {
	if raw, ok := parsed.scalars["acceptance_item_evidence_matrix"]; ok && raw != "none" {
		return nil, fmt.Errorf("must be literal none or a list")
	}
	items := append([]AcceptanceEvidenceEntry(nil), parsed.acceptanceEvidenceEntries...)
	for _, item := range items {
		if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Status) == "" {
			return nil, fmt.Errorf("each item must include id and status")
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func acceptanceItemIDSet(entries []AcceptanceItemEntry) map[string]bool {
	result := map[string]bool{}
	for _, item := range entries {
		result[item.ID] = true
	}
	return result
}

func acceptanceItemsByID(entries []AcceptanceItemEntry) map[string]AcceptanceItemEntry {
	result := map[string]AcceptanceItemEntry{}
	for _, item := range entries {
		result[item.ID] = item
	}
	return result
}

func normalizeAcceptanceItemEntries(entries []AcceptanceItemEntry) []AcceptanceItemEntry {
	if len(entries) == 0 {
		return nil
	}
	items := append([]AcceptanceItemEntry(nil), entries...)
	sort.Slice(items, func(i, j int) bool {
		if items[i].ID != items[j].ID {
			return items[i].ID < items[j].ID
		}
		return items[i].VerificationSurface < items[j].VerificationSurface
	})
	return items
}

func snapshotAcceptanceItemEntries(entries []AcceptanceItemEntry) []AcceptanceItemEntry {
	items := normalizeAcceptanceItemEntries(entries)
	for idx := range items {
		items[idx] = AcceptanceItemEntry{
			ID:                  items[idx].ID,
			VerificationSurface: items[idx].VerificationSurface,
			NotRunnableYet:      items[idx].NotRunnableYet,
		}
	}
	return items
}

func validateAcceptanceItemEntries(entries []AcceptanceItemEntry) error {
	seen := map[string]bool{}
	for _, item := range entries {
		if strings.TrimSpace(item.ID) == "" ||
			strings.TrimSpace(item.VerificationSurface) == "" ||
			strings.TrimSpace(item.NotRunnableYet) == "" {
			return fmt.Errorf("each item must include id, verification_surface, and not_runnable_yet")
		}
		if item.NotRunnableYet != "yes" && item.NotRunnableYet != "no" {
			return fmt.Errorf("not_runnable_yet for %s must be yes or no", item.ID)
		}
		if !allowedVerificationSurfaces[item.VerificationSurface] {
			return fmt.Errorf("verification_surface for %s must be one of public_api, internal_flow, error_handling, eventing, storage, integration, manual_effect", item.ID)
		}
		if seen[item.ID] {
			return fmt.Errorf("duplicate id %s", item.ID)
		}
		seen[item.ID] = true
	}
	return nil
}

func normalizeAppendixList(entries []AppendixEntry) string {
	if len(entries) == 0 {
		return "none"
	}
	items := make([]AppendixEntry, len(entries))
	copy(items, entries)
	sort.Slice(items, func(i, j int) bool {
		return items[i].FileRef < items[j].FileRef
	})
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("%s|%s", item.FileRef, item.Fingerprint))
	}
	return strings.Join(parts, ";")
}

func normalizeAcceptanceItemList(entries []AcceptanceItemEntry) string {
	if len(entries) == 0 {
		return "none"
	}
	items := normalizeAcceptanceItemEntries(entries)
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("%s|%s|%s", item.ID, item.VerificationSurface, item.NotRunnableYet))
	}
	return strings.Join(parts, ";")
}

func normalizeSharedList(entries []RuleEntry) string {
	if len(entries) == 0 {
		return "none"
	}
	items := make([]RuleEntry, len(entries))
	copy(items, entries)
	sort.Slice(items, func(i, j int) bool {
		if items[i].RuleID != items[j].RuleID {
			return items[i].RuleID < items[j].RuleID
		}
		if items[i].Layer != items[j].Layer {
			return items[i].Layer < items[j].Layer
		}
		return items[i].FileRef < items[j].FileRef
	})
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("%s|%s|%s|%s|%s", item.RuleID, item.Layer, item.FileRef, item.VersionRef, item.Fingerprint))
	}
	return strings.Join(parts, ";")
}

func normalizeRepositoryMapping(entry RepositoryMappingEntry) string {
	if entry.FileRef == "" && entry.VersionRef == "" && entry.Fingerprint == "" {
		return "none"
	}
	return fmt.Sprintf("%s|%s|%s", entry.FileRef, entry.VersionRef, entry.Fingerprint)
}

func normalizeObjectSnapshotList(entries []ObjectSnapshotEntry) string {
	if len(entries) == 0 {
		return "none"
	}
	items := sortObjectSnapshotEntries(entries)
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("%s|%s|%s|%s|%s", item.ObjectRef, item.Layer, item.FileRef, item.VersionRef, item.Fingerprint))
	}
	return strings.Join(parts, ";")
}

func finalizeValidationResult(result *ValidationResult) {
	if result.Valid {
		result.FailureLayer = "none"
		result.NextCommand = ""
		return
	}
	result.FailureLayer = classifyFailureLayer(result.ObjectType, result.ProcessKind, result.Mismatches)
	result.NextCommand = nextCommandForFailureLayer(result.ObjectType, result.ProcessKind, result.FailureLayer)
}

func classifyFailureLayer(objectType, processKind string, mismatches []string) string {
	for _, mismatch := range mismatches {
		switch {
		case strings.Contains(mismatch, "truth_"),
			strings.Contains(mismatch, "spec_file_ref mismatch"),
			strings.Contains(mismatch, "spec_version_ref mismatch"),
			strings.Contains(mismatch, "spec_fingerprint mismatch"),
			strings.Contains(mismatch, "acceptance_item_set mismatch"),
			strings.Contains(mismatch, "unit_appendix_snapshot mismatch"),
			strings.Contains(mismatch, "unit_snapshot mismatch"),
			strings.Contains(mismatch, "rule_snapshot mismatch"):
			return "truth_layer"
		}
	}
	switch processKind {
	case "plan":
		return "plan_layer"
	case "verify":
		return "evidence_layer"
	case "check":
		return "gate_layer"
	default:
		return "truth_layer"
	}
}

func nextCommandForFailureLayer(objectType, processKind, failureLayer string) string {
	switch objectType {
	case "unit":
		switch failureLayer {
		case "truth_layer", "gate_layer":
			return "unit_check"
		case "plan_layer":
			return "unit_plan"
		case "implementation_layer":
			return "unit_impl"
		case "evidence_layer":
			return "unit_verify"
		}
	}
	return ""
}

func copyStringBoolMap(source map[string]bool) map[string]bool {
	result := make(map[string]bool, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func renderAppendixLines(entries []AppendixEntry) []string {
	if len(entries) == 0 {
		return []string{"  none"}
	}
	lines := []string{}
	for _, entry := range entries {
		lines = append(lines,
			fmt.Sprintf("  - file_ref: %s", entry.FileRef),
			fmt.Sprintf("    fingerprint: %s", entry.Fingerprint),
		)
	}
	return lines
}

func renderSharedLines(entries []RuleEntry) []string {
	if len(entries) == 0 {
		return []string{"  none"}
	}
	lines := []string{}
	for _, entry := range entries {
		lines = append(lines,
			fmt.Sprintf("  - rule_id: %s", entry.RuleID),
			fmt.Sprintf("    layer: %s", entry.Layer),
			fmt.Sprintf("    file_ref: %s", entry.FileRef),
			fmt.Sprintf("    version_ref: %s", entry.VersionRef),
			fmt.Sprintf("    fingerprint: %s", entry.Fingerprint),
		)
	}
	return lines
}

func renderRepositoryMappingLines(entry RepositoryMappingEntry) []string {
	if entry.FileRef == "" && entry.VersionRef == "" && entry.Fingerprint == "" {
		return []string{"  none"}
	}
	return []string{
		fmt.Sprintf("  file_ref: %s", entry.FileRef),
		fmt.Sprintf("  version_ref: %s", entry.VersionRef),
		fmt.Sprintf("  fingerprint: %s", entry.Fingerprint),
	}
}

func renderObjectSnapshotLines(objectField string, entries []ObjectSnapshotEntry) []string {
	if len(entries) == 0 {
		return []string{"  none"}
	}
	lines := []string{}
	for _, entry := range sortObjectSnapshotEntries(entries) {
		lines = append(lines,
			fmt.Sprintf("  - %s: %s", objectField, entry.ObjectRef),
			fmt.Sprintf("    layer: %s", entry.Layer),
			fmt.Sprintf("    file_ref: %s", entry.FileRef),
			fmt.Sprintf("    version_ref: %s", entry.VersionRef),
			fmt.Sprintf("    fingerprint: %s", entry.Fingerprint),
		)
	}
	return lines
}

func renderAcceptanceItemLines(entries []AcceptanceItemEntry) []string {
	if len(entries) == 0 {
		return []string{"  none"}
	}
	lines := []string{}
	for _, entry := range normalizeAcceptanceItemEntries(entries) {
		lines = append(lines,
			fmt.Sprintf("  - id: %s", entry.ID),
			fmt.Sprintf("    verification_surface: %s", entry.VerificationSurface),
			fmt.Sprintf("    not_runnable_yet: %s", entry.NotRunnableYet),
		)
	}
	return lines
}

func ActivePlanFilePath(module string) string {
	return fmt.Sprintf("docs/specs/_plans/active/%s.md", module)
}

func DraftPlanFilePath(module string) string {
	return fmt.Sprintf("docs/specs/_plans/draft/%s.md", module)
}

func CheckResultFilePath(objectType, object string) string {
	return fmt.Sprintf("docs/specs/_check_result/%s/%s.md", objectType, object)
}

func VerifyResultFilePath(objectType, object string) string {
	return fmt.Sprintf("docs/specs/_verify_result/%s/%s.md", objectType, object)
}

func ProcessArtifactPaths(objectType, object, processKind string) ([]string, error) {
	if objectType != "unit" {
		return nil, fmt.Errorf("object type %q is not supported; only unit is supported", objectType)
	}
	switch processKind {
	case "check":
		return []string{CheckResultFilePath(objectType, object)}, nil
	case "plan":
		if objectType != "unit" {
			return nil, fmt.Errorf("process kind %q is not supported for object type %q", processKind, objectType)
		}
		return []string{DraftPlanFilePath(object), ActivePlanFilePath(object)}, nil
	case "verify":
		return []string{VerifyResultFilePath(objectType, object)}, nil
	default:
		return nil, fmt.Errorf("unsupported process kind %q", processKind)
	}
}

func ProcessFilePath(objectType, object, processKind string) (string, error) {
	if objectType != "unit" {
		return "", fmt.Errorf("object type %q is not supported; only unit is supported", objectType)
	}
	switch processKind {
	case "check":
		return CheckResultFilePath(objectType, object), nil
	case "plan":
		if objectType != "unit" {
			return "", fmt.Errorf("process kind %q is not supported for object type %q", processKind, objectType)
		}
		return ActivePlanFilePath(object), nil
	case "verify":
		return VerifyResultFilePath(objectType, object), nil
	default:
		return "", fmt.Errorf("unsupported process kind %q", processKind)
	}
}
