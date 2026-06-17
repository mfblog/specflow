package snapshot

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/rulebinding"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/rulerefs"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/unitappendix"
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
	ID           string
	Status       string
	EvidenceRefs string
}

type RetirementTargetEntry struct {
	ID                 string
	TargetRef          string
	TargetKind         string
	RetirementMethod   string
	VerificationAction string
	AcceptanceItemIDs  string
}

type RetirementEvidenceEntry struct {
	ID                 string
	Result             string
	MainlineDependency string
	EvidenceRefs       string
}

type PlannedChangeScopeEntry struct {
	ID                 string
	BasisRefs          string
	AcceptanceItemIDs  string
	ImplementationRefs string
	VerificationAction string
}

type PackageDeltaVerificationEntry struct {
	PlannedChangeScopeID string
	Result               string
	EvidenceRefs         string
}

type Snapshot struct {
	ObjectType                    string
	Object                        string
	Module                        string
	TruthLayerRef                 string
	SpecFileRef                   string
	SpecVersionRef                string
	SpecFingerprint               string
	AcceptanceBehaviorFingerprint string
	ModuleAppendixSnapshot        []AppendixEntry
	RepositoryMapping             RepositoryMappingEntry
	UnitSnapshot                  []ObjectSnapshotEntry
	RuleSnapshot                  []RuleEntry
	AcceptanceItemSet             []AcceptanceItemEntry
}

type ValidationResult struct {
	ObjectType      string
	Object          string
	ProcessKind     string
	ProcessFile     string
	Valid           bool
	Mismatches      []string
	FailureLayer    string
	NextCommand     string
	FreshnessImpact string
	EvidenceReuse   string
	Expected        Snapshot
}

type validationOptions struct {
	SkipIndependentEvaluationReceipt bool
}

type ProcessSnapshotData struct {
	ProcessKind                   string
	ProcessFile                   string
	PresentFields                 map[string]bool
	Scalars                       map[string]string
	AcceptanceBehaviorFingerprint string
	ModuleAppendixSnapshot        []AppendixEntry
	RepositoryMapping             RepositoryMappingEntry
	UnitSnapshot                []ObjectSnapshotEntry
	RuleSnapshot                  []RuleEntry
	AcceptanceItemSet             []AcceptanceItemEntry
	AcceptancePlanCoverage        []AcceptancePlanCoverageEntry
	AcceptanceEvidence            []AcceptanceEvidenceEntry
	RetirementTargets             []RetirementTargetEntry
	RetirementEvidence            []RetirementEvidenceEntry
	PlannedChangeScope            []PlannedChangeScopeEntry
	PackageDeltaVerification      []PackageDeltaVerificationEntry
}

const (
	FreshnessCurrent         = "current"
	FreshnessTextDrift       = "text_drift"
	FreshnessSemanticDrift   = "semantic_drift"
	FreshnessAcceptanceDrift = "acceptance_drift"
	FreshnessDependencyDrift = "dependency_drift"
	FreshnessSchemaDrift     = "schema_drift"
	FreshnessUnknownDrift    = "unknown_drift"

	EvidenceReuseNotNeeded     = "not_needed"
	EvidenceReusePendingReview = "pending_review"
	EvidenceReuseAccepted      = "accepted"
	EvidenceReuseRejected      = "rejected"
	EvidenceReuseNotEligible   = "not_eligible"
)

var independentEvaluationReceiptFields = []string{
	"evaluation_mode",
	"reviewer_result",
	"reviewer_context",
	"review_input_refs",
	"review_findings",
	"human_decision_refs",
}

var freshnessReceiptFields = []string{
	"freshness_impact",
	"evidence_reuse",
	"freshness_current_fingerprint",
	"freshness_review_mode",
	"freshness_reviewer_result",
	"freshness_reviewer_context",
	"freshness_review_input_refs",
	"freshness_review_findings",
}

var requiredUnitProcessSnapshotFields = map[string][]string{
	"check": withIndependentEvaluationReceipt(
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
		"acceptance_behavior_fingerprint",
		"acceptance_item_set",
		"unit_appendix_snapshot",
		"unit_snapshot",
		"rule_snapshot",
	),
	"plan": {
		"spec_file_ref",
		"spec_version_ref",
		"spec_fingerprint",
		"acceptance_behavior_fingerprint",
		"stable_candidate_diff_refs",
		"implementation_gap_refs",
		"unit_appendix_snapshot",
		"unit_snapshot",
		"rule_snapshot",
		"acceptance_item_plan_coverage",
		"retirement_targets",
		"planned_change_scope",
		"package_constraint_review",
		"package_constraint_refs",
		"package_constraint_summary",
	},
	"verify": withIndependentEvaluationReceipt(
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
		"acceptance_behavior_fingerprint",
		"acceptance_item_set",
		"unit_appendix_snapshot",
		"unit_snapshot",
		"rule_snapshot",
		"acceptance_item_evidence_matrix",
	),
	"stable_verify": withIndependentEvaluationReceipt(
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
		"acceptance_behavior_fingerprint",
		"repository_mapping_snapshot",
		"acceptance_item_set",
		"unit_appendix_snapshot",
		"unit_snapshot",
		"rule_snapshot",
		"acceptance_item_evidence_matrix",
		"implementation_surface_refs",
		"evidence_refs",
	),
}

var allowedAcceptanceEvidenceStatuses = map[string]bool{
	"pass":             true,
	"fail":             true,
	"partial":          true,
	"not_checked":      true,
	"not_runnable_yet": true,
}

var allowedRetirementTargetKinds = map[string]bool{
	"path":         true,
	"helper":       true,
	"wrapper":      true,
	"compat_layer": true,
	"dependency":   true,
	"other":        true,
}

var allowedRetirementMethods = map[string]bool{
	"remove":  true,
	"reroute": true,
	"replace": true,
	"isolate": true,
}

var allowedRetirementEvidenceResults = map[string]bool{
	"pass":        true,
	"fail":        true,
	"not_checked": true,
}

var allowedMainlineDependencyResults = map[string]bool{
	"not_required":   true,
	"still_required": true,
	"unknown":        true,
}

var allowedPackageDeltaVerificationResults = map[string]bool{
	"pass":        true,
	"fail":        true,
	"not_checked": true,
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

var stableVerifyDecisions = map[string]struct {
	AllowNext   string
	NextCommand string
}{
	"aligned":                    {AllowNext: "true", NextCommand: "unit_fork"},
	"controlled_repair_required": {AllowNext: "true", NextCommand: "unit_fork"},
	"controlled_change_required": {AllowNext: "true", NextCommand: "unit_fork"},
	"small_repair_required":      {AllowNext: "false", NextCommand: "unit_stable_verify"},
	"evidence_incomplete":        {AllowNext: "false", NextCommand: "unit_stable_verify"},
	"truth_rejudge_required":     {AllowNext: "false", NextCommand: "unit_stable_verify"},
	"truth_text_change_required": {AllowNext: "true", NextCommand: "unit_fork"},
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

	return RebuildObjectLayer(repoRoot, objectType, status.Object, status.ActiveLayer)
}

func RebuildObjectLayer(repoRoot, objectType, object, layer string) (Snapshot, error) {
	objectType = strings.TrimSpace(objectType)
	object = strings.TrimSpace(object)
	layer = strings.TrimSpace(layer)
	if objectType != "unit" {
		return Snapshot{}, fmt.Errorf("object type %q is not supported; only unit is supported", objectType)
	}
	if object == "" {
		return Snapshot{}, fmt.Errorf("object is required")
	}
	if layer != "candidate" && layer != "stable" {
		return Snapshot{}, fmt.Errorf("layer %q is not supported", layer)
	}

	mainSpecRef, err := specpaths.ObjectMainSpecFileRef(objectType, layer, object)
	if err != nil {
		return Snapshot{}, err
	}
	return rebuildObjectSnapshotFromMainSpec(repoRoot, objectType, object, layer, mainSpecRef)
}

func rebuildObjectSnapshotFromMainSpec(repoRoot, objectType, object, layer, mainSpecRef string) (Snapshot, error) {
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
		TruthLayerRef:   layer,
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
	sharedEntries, err := buildRuleSnapshot(repoRoot, layer, ruleRefs)
	if err != nil {
		return Snapshot{}, err
	}
	result.RuleSnapshot = sharedEntries

	acceptanceItems, err := buildAcceptanceItemSet(mainSpecRef, body)
	if err != nil {
		return Snapshot{}, err
	}
	result.AcceptanceItemSet = acceptanceItems
	result.AcceptanceBehaviorFingerprint = fingerprintAcceptanceBehavior(acceptanceItems)
	return result, nil
}

func CandidateIntentForObject(repoRoot, objectType, object string) (string, error) {
	frontmatter, err := candidateFrontmatterForObject(repoRoot, objectType, object)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(frontmatter["candidate_intent"]), nil
}

func CandidateSourceBasisForObject(repoRoot, objectType, object string) (string, error) {
	frontmatter, err := candidateFrontmatterForObject(repoRoot, objectType, object)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(frontmatter["source_basis"]), nil
}

func candidateFrontmatterForObject(repoRoot, objectType, object string) (map[string]string, error) {
	objectType = strings.TrimSpace(objectType)
	object = strings.TrimSpace(object)
	if objectType != "unit" {
		return nil, fmt.Errorf("object type %q is not supported; only unit is supported", objectType)
	}
	mainSpecRef, err := specpaths.ObjectMainSpecFileRef(objectType, "candidate", object)
	if err != nil {
		return nil, err
	}
	mainSpecAbs := filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef))
	mainSpecContent, err := os.ReadFile(mainSpecAbs)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", mainSpecRef, err)
	}
	frontmatter, _, err := parseFrontmatter(string(mainSpecContent))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", mainSpecRef, err)
	}
	return frontmatter, nil
}

type StableVerifyIntentRequirement struct {
	Decision       string
	RequiredIntent string
	Required       bool
}

func StableVerifyCandidateIntentRequirement(repoRoot, objectType, object string) (StableVerifyIntentRequirement, error) {
	validation, err := ValidateProcessFileForObject(repoRoot, objectType, object, "stable_verify")
	if err == nil && validation.Valid {
		processData, err := LoadProcessSnapshot(repoRoot, objectType, object, "stable_verify")
		if err != nil {
			return StableVerifyIntentRequirement{}, err
		}
		return stableVerifyCandidateIntentRequirementFromData(processData), nil
	}
	processData, err := LoadProcessSnapshot(repoRoot, objectType, object, "stable_verify")
	if err != nil {
		return StableVerifyIntentRequirement{}, nil
	}
	requirement := stableVerifyCandidateIntentRequirementFromData(processData)
	if !requirement.Required {
		return requirement, nil
	}
	if !stableVerifyIntentGuardValid(repoRoot, objectType, object, processData) {
		return StableVerifyIntentRequirement{}, nil
	}
	return requirement, nil
}

func stableVerifyCandidateIntentRequirementFromData(processData ProcessSnapshotData) StableVerifyIntentRequirement {
	decision := processData.Scalars["decision"]
	switch decision {
	case "controlled_repair_required":
		return StableVerifyIntentRequirement{Decision: decision, RequiredIntent: "repair", Required: true}
	case "truth_text_change_required":
		return StableVerifyIntentRequirement{Decision: decision, RequiredIntent: "repair", Required: true}
	case "controlled_change_required":
		return StableVerifyIntentRequirement{Decision: decision, RequiredIntent: "change", Required: true}
	default:
		return StableVerifyIntentRequirement{Decision: decision}
	}
}

func stableVerifyIntentGuardValid(repoRoot, objectType, object string, processData ProcessSnapshotData) bool {
	if objectType != "unit" {
		return false
	}
	for _, field := range []string{
		"object_type",
		"object_ref",
		"gate",
		"decision",
		"allow_next",
		"next_command",
		"truth_layer_ref",
		"truth_file_ref",
		"truth_version_ref",
		"truth_fingerprint",
		"acceptance_behavior_fingerprint",
		"repository_mapping_snapshot",
		"acceptance_item_set",
		"unit_appendix_snapshot",
		"unit_snapshot",
		"rule_snapshot",
		"acceptance_item_evidence_matrix",
		"implementation_surface_refs",
		"evidence_refs",
	} {
		if !processData.PresentFields[field] {
			return false
		}
	}
	if processData.Scalars["object_type"] != objectType ||
		processData.Scalars["object_ref"] != object ||
		processData.Scalars["gate"] != "unit_stable_verify" ||
		processData.Scalars["truth_layer_ref"] != "stable" {
		return false
	}
	route, ok := stableVerifyDecisions[processData.Scalars["decision"]]
	if !ok ||
		processData.Scalars["allow_next"] != route.AllowNext ||
		processData.Scalars["next_command"] != route.NextCommand ||
		route.NextCommand != "unit_fork" {
		return false
	}

	expected, err := stableTruthForIntentGuard(repoRoot, objectType, object)
	if err != nil {
		return false
	}
	if processData.Scalars["truth_file_ref"] != expected.SpecFileRef ||
		processData.Scalars["truth_version_ref"] != expected.SpecVersionRef ||
		processData.Scalars["truth_fingerprint"] != expected.SpecFingerprint {
		return false
	}
	if processData.Scalars["acceptance_behavior_fingerprint"] != expected.AcceptanceBehaviorFingerprint {
		return false
	}

	repositoryMapping, err := BuildRepositoryMappingSnapshot(repoRoot)
	if err != nil {
		return false
	}
	if normalizeRepositoryMapping(processData.RepositoryMapping) != normalizeRepositoryMapping(repositoryMapping) {
		return false
	}
	if normalizeAcceptanceItemList(processData.AcceptanceItemSet) != normalizeAcceptanceItemList(expected.AcceptanceItemSet) {
		return false
	}
	if normalizeAppendixList(processData.ModuleAppendixSnapshot) != normalizeAppendixList(expected.ModuleAppendixSnapshot) {
		return false
	}
	if normalizeObjectSnapshotList(processData.UnitSnapshot) != normalizeObjectSnapshotList(expected.UnitSnapshot) {
		return false
	}
	return true
}

func stableTruthForIntentGuard(repoRoot, objectType, object string) (Snapshot, error) {
	mainSpecRef, err := specpaths.ObjectMainSpecFileRef(objectType, "stable", object)
	if err != nil {
		return Snapshot{}, err
	}
	mainSpecContent, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)))
	if err != nil {
		return Snapshot{}, fmt.Errorf("read %s: %w", mainSpecRef, err)
	}
	frontmatter, body, err := parseFrontmatter(string(mainSpecContent))
	if err != nil {
		return Snapshot{}, fmt.Errorf("%s: %w", mainSpecRef, err)
	}
	version := strings.TrimSpace(frontmatter["version"])
	if version == "" {
		return Snapshot{}, fmt.Errorf("%s: missing frontmatter.version", mainSpecRef)
	}
	appendixEntries, err := buildAppendixSnapshot(repoRoot, mainSpecRef, frontmatter, body)
	if err != nil {
		return Snapshot{}, err
	}
	unitRefs, _, err := parseFrontmatterNamedRefs(string(mainSpecContent), "unit_refs")
	if err != nil {
		return Snapshot{}, err
	}
	unitSnapshot, err := buildObjectDependencySnapshot(repoRoot, "unit", unitRefs)
	if err != nil {
		return Snapshot{}, err
	}
	acceptanceItems, err := buildAcceptanceItemSet(mainSpecRef, body)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{
		ObjectType:                    objectType,
		Object:                        object,
		Module:                        object,
		TruthLayerRef:                 "stable",
		SpecFileRef:                   mainSpecRef,
		SpecVersionRef:                fmt.Sprintf("%s@%s", strings.TrimSuffix(filepath.Base(mainSpecRef), ".md"), version),
		SpecFingerprint:               hashNormalizedText(string(mainSpecContent)),
		ModuleAppendixSnapshot:        appendixEntries,
		UnitSnapshot:                  unitSnapshot,
		AcceptanceItemSet:             acceptanceItems,
		AcceptanceBehaviorFingerprint: fingerprintAcceptanceBehavior(acceptanceItems),
	}, nil
}

func ValidateProcessFile(repoRoot, module, processKind string) (ValidationResult, error) {
	return ValidateProcessFileForObject(repoRoot, "unit", module, processKind)
}

func ValidateProcessFileForObject(repoRoot, objectType, object, processKind string) (ValidationResult, error) {
	return validateProcessFileForObject(repoRoot, objectType, object, processKind, validationOptions{})
}

func ValidateProcessFileForIndependentEvaluationRequest(repoRoot, objectType, object, processKind string) (ValidationResult, error) {
	return validateProcessFileForObject(repoRoot, objectType, object, processKind, validationOptions{
		SkipIndependentEvaluationReceipt: true,
	})
}

func validateProcessFileForObject(repoRoot, objectType, object, processKind string, options validationOptions) (ValidationResult, error) {
	expected, err := RebuildCurrentObject(repoRoot, objectType, object)
	if err != nil {
		return ValidationResult{}, err
	}
	requiredFields, ok := requiredFieldsForObjectProcess(objectType, processKind)
	if !ok {
		return ValidationResult{}, fmt.Errorf("process kind %q is not supported for object type %q", processKind, objectType)
	}
	if options.SkipIndependentEvaluationReceipt {
		requiredFields = withoutIndependentEvaluationReceipt(requiredFields)
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
		ObjectType:      objectType,
		Object:          object,
		ProcessKind:     processKind,
		ProcessFile:     processFile,
		Expected:        expected,
		Valid:           true,
		FreshnessImpact: FreshnessCurrent,
		EvidenceReuse:   EvidenceReuseNotNeeded,
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
	if !options.SkipIndependentEvaluationReceipt && processKind != "plan" {
		validateIndependentEvaluationReceipt(&result, actual)
	}
	validateCandidateAppendixCoverage(repoRoot, &result, expected)

	if processKind == "plan" {
		compareScalar(&result, "spec_file_ref", actual.scalars["spec_file_ref"], expected.SpecFileRef)
		compareScalar(&result, "spec_version_ref", actual.scalars["spec_version_ref"], expected.SpecVersionRef)
		comparePrimaryFingerprint(&result, "spec_fingerprint", actual.scalars["spec_fingerprint"], expected.SpecFingerprint)
		compareAcceptanceBehaviorFingerprint(&result, actual, expected, false)

		if _, ok := actual.scalars["unit_appendix_snapshot"]; ok || actual.appendixPresent {
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
		validateAcceptancePlanCoverage(&result, actual, expected.AcceptanceItemSet)
		validatePlanReferenceFields(repoRoot, &result, actual, expected)
		validateRetirementTargets(repoRoot, &result, actual, expected)
		validatePackageConstraintReview(&result, actual, expected)
		validatePlannedChangeScope(repoRoot, &result, actual, expected)
		finalizeValidationResult(&result, actual)
		return result, nil
	}

	if processKind == "stable_verify" {
		validateStableVerifyProcess(repoRoot, &result, actual, expected)
		finalizeValidationResult(&result, actual)
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
	comparePrimaryFingerprint(&result, "truth_fingerprint", actual.scalars["truth_fingerprint"], expected.SpecFingerprint)
	compareAcceptanceBehaviorFingerprint(&result, actual, expected, true)
	compareAcceptanceItemSet(&result, actual, expected.AcceptanceItemSet)
	if processKind == "verify" {
		validateAcceptanceEvidenceMatrix(&result, actual, expected.AcceptanceItemSet, true)
		if actual.presentFields["active_plan_file_ref"] {
			validateActivePlanBinding(repoRoot, &result, actual, expected)
		}
		planParsed, planErr := parseActivePlan(repoRoot, expected.Object)
		var targetEntries []RetirementTargetEntry
		var plannedEntries []PlannedChangeScopeEntry
		if planErr == nil {
			targetEntries, _ = retirementTargetEntriesFromParsed(planParsed)
			plannedEntries, _ = plannedChangeScopeEntriesFromParsed(planParsed)
		}
		validateRetirementEvidence(repoRoot, &result, actual, expected, targetEntries)
		validatePackageDeltaVerification(repoRoot, &result, actual, expected, plannedEntries)
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

	finalizeValidationResult(&result, actual)
	return result, nil
}

func validateCandidateAppendixCoverage(repoRoot string, result *ValidationResult, expected Snapshot) {
	if expected.ObjectType != "unit" || expected.TruthLayerRef != "candidate" {
		return
	}
	mismatches, err := unitappendix.CandidateCoverageMismatches(repoRoot, expected.ObjectType, expected.Object)
	if err != nil {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, err.Error())
		return
	}
	if len(mismatches) == 0 {
		return
	}
	result.Valid = false
	result.Mismatches = append(result.Mismatches, mismatches...)
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

func withIndependentEvaluationReceipt(fields ...string) []string {
	result := append([]string{}, fields...)
	result = append(result, independentEvaluationReceiptFields...)
	return result
}

func withoutIndependentEvaluationReceipt(fields []string) []string {
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		if isIndependentEvaluationReceiptField(field) {
			continue
		}
		result = append(result, field)
	}
	return result
}

func isIndependentEvaluationReceiptField(field string) bool {
	for _, receiptField := range independentEvaluationReceiptFields {
		if field == receiptField {
			return true
		}
	}
	return false
}

// expectedProcessRouting returns (expectedGate, expectedNextCommand) for a
// given object type and process kind.
//
// For check results: gate is "unit_check" (the gate that produced this result)
// and next_command is "unit_verify" (the advancement target — matching
// process_snapshot_contract.md Section 2's advancement-target convention).
//
// For verify results: gate is "unit_verify" and next_command is "unit_promote".
// The verify result records the promotion target.
//
// See process_snapshot_contract.md Section 2 for the common fields documentation.
func expectedProcessRouting(objectType, processKind string) (string, string, error) {
	switch objectType {
	case "unit":
		switch processKind {
		case "check":
			return "unit_check", "unit_verify", nil
		case "verify":
			return "unit_verify", "unit_promote", nil
		}
	}
	return "", "", fmt.Errorf("process kind %q is not supported for object type %q", processKind, objectType)
}

func validateIndependentEvaluationReceipt(result *ValidationResult, actual processSnapshot) {
	if actual.presentFields["evaluation_mode"] {
		compareScalar(result, "evaluation_mode", actual.scalars["evaluation_mode"], "independent")
	}
	if actual.presentFields["reviewer_result"] {
		compareScalar(result, "reviewer_result", actual.scalars["reviewer_result"], "pass")
	}
	if actual.presentFields["reviewer_context"] {
		compareScalar(result, "reviewer_context", actual.scalars["reviewer_context"], "minimal_context")
	}
	if actual.presentFields["review_findings"] {
		compareScalar(result, "review_findings", actual.scalars["review_findings"], "none")
	}
	if actual.presentFields["review_input_refs"] {
		validateEvaluationInputRefs(result, "review_input_refs", actual.scalars["review_input_refs"], packForProcessKind(result.ProcessKind))
	}
	if actual.presentFields["human_decision_refs"] {
		value := strings.TrimSpace(actual.scalars["human_decision_refs"])
		if value != "" && value != "none" && isChatOnlyHumanDecisionRef(value) {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, "human_decision_refs must be none or explicit durable refs, not chat-only")
		}
	}
}

func validateEvaluationInputRefs(result *ValidationResult, field, value, expectedPack string) {
	refs := splitProcessRefList(value)
	if len(refs) == 0 || strings.TrimSpace(value) == "none" {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, field+" must include reviewer pack, request file, and durable input refs")
		return
	}
	for _, ref := range refs {
		if ref == "none" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, field+" must not contain none")
			return
		}
	}
	if expectedPack == "" {
		return
	}
	expectedRequest := evaluationRequestFileRef(result.ObjectType, result.Object, expectedPack)
	if !containsRef(refs, expectedPack) {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, fmt.Sprintf("%s must contain reviewer pack: %s", field, expectedPack))
	}
	if !containsRef(refs, expectedRequest) {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, fmt.Sprintf("%s must contain request file ref: %s", field, expectedRequest))
	}
	hasDurableInput := false
	for _, ref := range refs {
		if ref != expectedPack && ref != expectedRequest {
			hasDurableInput = true
			break
		}
	}
	if !hasDurableInput {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, field+" must contain at least one durable input ref")
	}
}

func splitProcessRefList(value string) []string {
	parts := strings.Split(value, ";")
	refs := []string{}
	for _, part := range parts {
		ref := strings.TrimSpace(filepath.ToSlash(part))
		if ref == "" {
			continue
		}
		refs = append(refs, ref)
	}
	return refs
}

func containsRef(refs []string, want string) bool {
	for _, ref := range refs {
		if ref == want {
			return true
		}
	}
	return false
}

func evaluationRequestFileRef(objectType, object, pack string) string {
	return filepath.ToSlash(filepath.Join("docs/specs/_independent_evaluation/requests", objectType, object, pack+".md"))
}

func packForProcessKind(processKind string) string {
	switch processKind {
	case "check":
		return "unit_check_pass"
	case "verify":
		return "unit_verify_ready_to_promote"
	case "stable_verify":
		return "unit_stable_verify_advancing"
	default:
		return ""
	}
}

func isChatOnlyHumanDecisionRef(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return normalized == "chat" ||
		normalized == "conversation" ||
		normalized == "thread" ||
		strings.Contains(normalized, "chat-only") ||
		strings.Contains(normalized, "conversation-only")
}

func validateStableVerifyProcess(repoRoot string, result *ValidationResult, actual processSnapshot, expected Snapshot) {
	decision := actual.scalars["decision"]
	route, ok := stableVerifyDecisions[decision]
	if !ok {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, fmt.Sprintf("decision invalid for stable_verify: %s", decision))
	}

	if expected.TruthLayerRef != "stable" {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, fmt.Sprintf("stable_verify requires stable truth, got %s", expected.TruthLayerRef))
	}

	compareScalar(result, "object_type", actual.scalars["object_type"], expected.ObjectType)
	compareScalar(result, "object_ref", actual.scalars["object_ref"], expected.Object)
	compareScalar(result, "gate", actual.scalars["gate"], "unit_stable_verify")
	if ok {
		compareScalar(result, "allow_next", actual.scalars["allow_next"], route.AllowNext)
		compareScalar(result, "next_command", actual.scalars["next_command"], route.NextCommand)
	}
	compareScalar(result, "truth_layer_ref", actual.scalars["truth_layer_ref"], "stable")
	compareScalar(result, "truth_file_ref", actual.scalars["truth_file_ref"], expected.SpecFileRef)
	compareScalar(result, "truth_version_ref", actual.scalars["truth_version_ref"], expected.SpecVersionRef)
	comparePrimaryFingerprint(result, "truth_fingerprint", actual.scalars["truth_fingerprint"], expected.SpecFingerprint)
	compareAcceptanceBehaviorFingerprint(result, actual, expected, true)

	repositoryMapping, err := BuildRepositoryMappingSnapshot(repoRoot)
	if err != nil {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, err.Error())
	} else {
		actualMapping := normalizeRepositoryMapping(actual.repositoryMapping)
		expectedMapping := normalizeRepositoryMapping(repositoryMapping)
		if actualMapping != expectedMapping {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("repository_mapping_snapshot mismatch: actual=%s expected=%s", actualMapping, expectedMapping))
		}
	}

	compareAcceptanceItemSet(result, actual, expected.AcceptanceItemSet)
	validateAcceptanceEvidenceMatrix(result, actual, expected.AcceptanceItemSet, false)
	validateStableVerifyEvidenceRefs(result, actual, expected.AcceptanceItemSet)
	if decision == "aligned" {
		validateStableAlignedEvidencePasses(result, actual, expected.AcceptanceItemSet)
	}

	actualAppendix := normalizeAppendixList(actual.appendixEntries)
	expectedAppendix := normalizeAppendixList(expected.ModuleAppendixSnapshot)
	if actualAppendix != expectedAppendix {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, fmt.Sprintf("unit_appendix_snapshot mismatch: actual=%s expected=%s", actualAppendix, expectedAppendix))
	}

	actualUnits := normalizeObjectSnapshotList(actual.moduleEntries)
	expectedUnits := normalizeObjectSnapshotList(expected.UnitSnapshot)
	if actualUnits != expectedUnits {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, fmt.Sprintf("unit_snapshot mismatch: actual=%s expected=%s", actualUnits, expectedUnits))
	}

	actualShared := normalizeSharedList(actual.sharedEntries)
	expectedShared := normalizeSharedList(expected.RuleSnapshot)
	if actualShared != expectedShared {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, fmt.Sprintf("rule_snapshot mismatch: actual=%s expected=%s", actualShared, expectedShared))
	}
}

func validateStableAlignedEvidencePasses(result *ValidationResult, actual processSnapshot, expected []AcceptanceItemEntry) {
	entries, err := acceptanceEvidenceEntriesFromParsed(actual)
	if err != nil {
		return
	}
	expectedByID := acceptanceItemsByID(expected)
	for _, entry := range entries {
		expectedItem, ok := expectedByID[entry.ID]
		if !ok {
			continue
		}
		if expectedItem.NotRunnableYet == "yes" {
			if entry.Status != "not_runnable_yet" {
				result.Valid = false
				result.Mismatches = append(result.Mismatches, fmt.Sprintf("stable_verify aligned evidence for %s must be not_runnable_yet", entry.ID))
			}
			continue
		}
		if entry.Status != "pass" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("stable_verify aligned evidence for %s must be pass", entry.ID))
		}
	}
}

func validateStableVerifyEvidenceRefs(result *ValidationResult, actual processSnapshot, expected []AcceptanceItemEntry) {
	entries, err := acceptanceEvidenceEntriesFromParsed(actual)
	if err != nil {
		return
	}
	expectedByID := acceptanceItemsByID(expected)
	for _, entry := range entries {
		expectedItem, ok := expectedByID[entry.ID]
		if !ok || expectedItem.NotRunnableYet == "yes" {
			continue
		}
		evidenceRefs := strings.TrimSpace(entry.EvidenceRefs)
		if evidenceRefs == "" || evidenceRefs == "none" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("acceptance_item_evidence_matrix evidence_refs for %s must be durable refs", entry.ID))
		}
	}
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
		ProcessKind:                   processKind,
		ProcessFile:                   processFile,
		PresentFields:                 copyStringBoolMap(parsed.presentFields),
		Scalars:                       scalars,
		AcceptanceBehaviorFingerprint: scalars["acceptance_behavior_fingerprint"],
		ModuleAppendixSnapshot:        append([]AppendixEntry(nil), parsed.appendixEntries...),
		UnitSnapshot:                append([]ObjectSnapshotEntry(nil), parsed.moduleEntries...),
		RuleSnapshot:                  append([]RuleEntry(nil), parsed.sharedEntries...),
		RepositoryMapping:             parsed.repositoryMapping,
		AcceptanceItemSet:             append([]AcceptanceItemEntry(nil), parsed.acceptanceItemEntries...),
		AcceptancePlanCoverage:        append([]AcceptancePlanCoverageEntry(nil), parsed.acceptancePlanEntries...),
		AcceptanceEvidence:            append([]AcceptanceEvidenceEntry(nil), parsed.acceptanceEvidenceEntries...),
		RetirementTargets:             append([]RetirementTargetEntry(nil), parsed.retirementTargetEntries...),
		RetirementEvidence:            append([]RetirementEvidenceEntry(nil), parsed.retirementEvidenceEntries...),
		PlannedChangeScope:            append([]PlannedChangeScopeEntry(nil), parsed.plannedChangeEntries...),
		PackageDeltaVerification:      append([]PackageDeltaVerificationEntry(nil), parsed.packageDeltaEntries...),
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
		fmt.Sprintf("acceptance_behavior_fingerprint: %s", snapshot.AcceptanceBehaviorFingerprint),
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

// RenderWithAppendixCoverage returns the rendered snapshot text with appendix
// coverage information appended as YAML comments for candidate-layer objects.
// For non-candidate objects, it behaves identically to Render().
func RenderWithAppendixCoverage(snapshot Snapshot, repoRoot string) string {
	text := Render(snapshot)
	if snapshot.TruthLayerRef != "candidate" {
		return text
	}
	mismatches, err := unitappendix.CandidateCoverageMismatches(
		repoRoot, snapshot.ObjectType, snapshot.Object,
	)
	if err != nil {
		return text + "\n# appendix coverage: error (appendix coverage check failed — see stderr for details)\n"
	}
	if len(mismatches) == 0 {
		return text + "\n# appendix coverage: valid (every stable appendix has a candidate counterpart)\n"
	}
	var buf strings.Builder
	buf.WriteString(text)
	buf.WriteString("\n# appendix coverage: invalid (missing candidate appendix for stable appendix)\n")
	for _, m := range mismatches {
		buf.WriteString("#   " + m + "\n")
	}
	return buf.String()
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
	currentObject, _, err := mainSpecObject(mainSpecRef)
	if err != nil {
		return nil, err
	}
	scanned, err := unitappendix.Scan(repoRoot, "unit", currentObject, currentLayer)
	if err != nil {
		return nil, err
	}
	entries := []AppendixEntry{}
	entryByRef := map[string]bool{}
	for _, appendix := range scanned {
		entryByRef[appendix.FileRef] = true
		entries = append(entries, AppendixEntry{
			FileRef:     appendix.FileRef,
			Fingerprint: hashNormalizedText(appendix.Content),
		})
	}
	if evidenceRef := strings.TrimSpace(frontmatter["evidence_appendix_ref"]); evidenceRef != "" && evidenceRef != "none" {
		relPath, err := resolveAppendixRef(repoRoot, mainDir, evidenceRef)
		if err != nil {
			return nil, err
		}
		if !strings.Contains(relPath, "/appendix/") {
			return nil, fmt.Errorf("%s: evidence appendix ref %s is not under an appendix directory", mainSpecRef, relPath)
		}
		if currentLayer != "candidate" {
			return nil, fmt.Errorf("%s: evidence appendix ref %s must point to a current candidate appendix", mainSpecRef, relPath)
		}
		if !entryByRef[relPath] {
			return nil, fmt.Errorf("%s: evidence appendix ref %s is not a current candidate appendix for unit %s", mainSpecRef, relPath, currentObject)
		}
	}
	_ = body

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

// ComputeFileFingerprint applies the Section 7 normalization rules to the
// given file content and returns its SHA-256 hex fingerprint.
// This is the same algorithm used for truth_fingerprint, spec_fingerprint,
// and all snapshot item fingerprints.
func ComputeFileFingerprint(content string) string {
	return hashNormalizedText(content)
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
	retirementTargetEntries   []RetirementTargetEntry
	retirementTargetPresent   bool
	retirementEvidenceEntries []RetirementEvidenceEntry
	retirementEvidencePresent bool
	plannedChangeEntries      []PlannedChangeScopeEntry
	plannedChangePresent      bool
	packageDeltaEntries       []PackageDeltaVerificationEntry
	packageDeltaPresent       bool
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
				case "retirement_targets":
					result.retirementTargetPresent = true
					currentList = key
				case "retirement_evidence_matrix":
					result.retirementEvidencePresent = true
					currentList = key
				case "planned_change_scope":
					result.plannedChangePresent = true
					currentList = key
				case "package_delta_verification":
					result.packageDeltaPresent = true
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
			case "retirement_targets":
				if currentIndex < 0 || (listItemStart && key == "id") {
					result.retirementTargetEntries = append(result.retirementTargetEntries, RetirementTargetEntry{})
					currentIndex = len(result.retirementTargetEntries) - 1
				}
				assignRetirementTargetField(&result.retirementTargetEntries[currentIndex], key, value)
			case "retirement_evidence_matrix":
				if currentIndex < 0 || (listItemStart && key == "id") {
					result.retirementEvidenceEntries = append(result.retirementEvidenceEntries, RetirementEvidenceEntry{})
					currentIndex = len(result.retirementEvidenceEntries) - 1
				}
				assignRetirementEvidenceField(&result.retirementEvidenceEntries[currentIndex], key, value)
			case "planned_change_scope":
				if currentIndex < 0 || (listItemStart && key == "id") {
					result.plannedChangeEntries = append(result.plannedChangeEntries, PlannedChangeScopeEntry{})
					currentIndex = len(result.plannedChangeEntries) - 1
				}
				assignPlannedChangeScopeField(&result.plannedChangeEntries[currentIndex], key, value)
			case "package_delta_verification":
				if currentIndex < 0 || (listItemStart && key == "planned_change_scope_id") {
					result.packageDeltaEntries = append(result.packageDeltaEntries, PackageDeltaVerificationEntry{})
					currentIndex = len(result.packageDeltaEntries) - 1
				}
				assignPackageDeltaVerificationField(&result.packageDeltaEntries[currentIndex], key, value)
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
	if raw, ok := result.scalars["retirement_targets"]; ok && raw == "none" {
		result.retirementTargetPresent = true
		result.retirementTargetEntries = nil
	}
	if raw, ok := result.scalars["retirement_evidence_matrix"]; ok && raw == "none" {
		result.retirementEvidencePresent = true
		result.retirementEvidenceEntries = nil
	}
	if raw, ok := result.scalars["planned_change_scope"]; ok && raw == "none" {
		result.plannedChangePresent = true
		result.plannedChangeEntries = nil
	}
	if raw, ok := result.scalars["package_delta_verification"]; ok && raw == "none" {
		result.packageDeltaPresent = true
		result.packageDeltaEntries = nil
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
	case "evidence_refs":
		entry.EvidenceRefs = value
	}
}

func assignRetirementTargetField(entry *RetirementTargetEntry, key, value string) {
	switch key {
	case "id":
		entry.ID = value
	case "target_ref":
		entry.TargetRef = value
	case "target_kind":
		entry.TargetKind = value
	case "retirement_method":
		entry.RetirementMethod = value
	case "verification_action":
		entry.VerificationAction = value
	case "acceptance_item_ids":
		entry.AcceptanceItemIDs = value
	}
}

func assignRetirementEvidenceField(entry *RetirementEvidenceEntry, key, value string) {
	switch key {
	case "id":
		entry.ID = value
	case "result":
		entry.Result = value
	case "mainline_dependency":
		entry.MainlineDependency = value
	case "evidence_refs":
		entry.EvidenceRefs = value
	}
}

func assignPlannedChangeScopeField(entry *PlannedChangeScopeEntry, key, value string) {
	switch key {
	case "id":
		entry.ID = value
	case "basis_refs":
		entry.BasisRefs = value
	case "acceptance_item_ids":
		entry.AcceptanceItemIDs = value
	case "implementation_refs":
		entry.ImplementationRefs = value
	case "verification_action":
		entry.VerificationAction = value
	}
}

func assignPackageDeltaVerificationField(entry *PackageDeltaVerificationEntry, key, value string) {
	switch key {
	case "planned_change_scope_id":
		entry.PlannedChangeScopeID = value
	case "result":
		entry.Result = value
	case "evidence_refs":
		entry.EvidenceRefs = value
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

func comparePrimaryFingerprint(result *ValidationResult, field, actual, expected string) {
	if actual == "" {
		result.Valid = false
		result.Mismatches = append(result.Mismatches,
			fmt.Sprintf("%s: empty fingerprint in process snapshot — evidence does not record the truth version it was based on; recreate or migrate the snapshot", field))
		return
	}
	if actual != expected {
		result.Valid = false
		result.Mismatches = append(result.Mismatches,
			fmt.Sprintf("%s mismatch: actual=%s expected=%s [see process_snapshot_contract.md Sections 6-7 for fingerprint algorithm]", field, actual, expected))
	}
}

func compareAcceptanceBehaviorFingerprint(result *ValidationResult, actual processSnapshot, expected Snapshot, required bool) {
	if !actual.presentFields["acceptance_behavior_fingerprint"] {
		if required {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, "acceptance_behavior_fingerprint is required but absent [see process_snapshot_contract.md Section 2]")
		}
		return
	}
	actualValue := strings.TrimSpace(actual.scalars["acceptance_behavior_fingerprint"])
	if actualValue == "" {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, "acceptance_behavior_fingerprint must not be empty")
		return
	}
	if actualValue != expected.AcceptanceBehaviorFingerprint {
		result.Valid = false
		result.Mismatches = append(result.Mismatches,
			fmt.Sprintf("acceptance_behavior_fingerprint mismatch: actual=%s expected=%s [see process_snapshot_contract.md Sections 6-6a for acceptance behavior fingerprint serialization]", actualValue, expected.AcceptanceBehaviorFingerprint))
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

func validatePlanReferenceFields(repoRoot string, result *ValidationResult, actual processSnapshot, expected Snapshot) {
	diffRefs := strings.TrimSpace(actual.scalars["stable_candidate_diff_refs"])
	if diffRefs == "" {
		return
	}
	if !validateNoneOrRefList(result, "stable_candidate_diff_refs", diffRefs) {
		return
	}
	if diffRefs != "none" {
		validateRepositoryRelativeExistingRefs(repoRoot, result, "stable_candidate_diff_refs", splitProcessRefList(diffRefs))
	}

	stableRef, stablePresent, err := stableMainSpecRefIfPresent(repoRoot, expected.ObjectType, expected.Object)
	if err != nil {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, err.Error())
		return
	}
	if stablePresent {
		if diffRefs == "none" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, "stable_candidate_diff_refs must cite stable and candidate main specs when stable truth exists")
		} else {
			refs := splitProcessRefList(diffRefs)
			if !containsRef(refs, stableRef) {
				result.Valid = false
				result.Mismatches = append(result.Mismatches, fmt.Sprintf("stable_candidate_diff_refs must contain stable spec ref: %s", stableRef))
			}
			if !containsRef(refs, expected.SpecFileRef) {
				result.Valid = false
				result.Mismatches = append(result.Mismatches, fmt.Sprintf("stable_candidate_diff_refs must contain candidate spec ref: %s", expected.SpecFileRef))
			}
		}
	}

	gapRefs := strings.TrimSpace(actual.scalars["implementation_gap_refs"])
	if gapRefs == "" {
		return
	}
	if !validateNoneOrRefList(result, "implementation_gap_refs", gapRefs) {
		return
	}
	if gapRefs == "none" && acceptanceItemsDeclareImplementationSurface(expected.AcceptanceItemSet) {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, "implementation_gap_refs must cite durable refs when acceptance items declare implementation surfaces")
	}
	if gapRefs != "none" {
		validateRepositoryRelativeExistingRefs(repoRoot, result, "implementation_gap_refs", splitProcessRefList(gapRefs))
	}
}

func validatePackageConstraintReview(result *ValidationResult, actual processSnapshot, expected Snapshot) {
	compareScalar(result, "package_constraint_review", actual.scalars["package_constraint_review"], "pass")
	rawRefs := strings.TrimSpace(actual.scalars["package_constraint_refs"])
	if rawRefs == "" {
			result.Valid = false
                     result.Mismatches = append(result.Mismatches,
  "package_constraint_refs must not be empty")
		return
	}
	if rawRefs == "none" {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, "package_constraint_refs must cite package refs")
		return
	}
	if !validateNoneOrRefList(result, "package_constraint_refs", rawRefs) {
		return
	}
	validatePackageBasisRefs(result, "package_constraint_refs", splitProcessRefList(rawRefs), expected)
}

func validatePlannedChangeScope(repoRoot string, result *ValidationResult, actual processSnapshot, expected Snapshot) {
	entries, err := plannedChangeScopeEntriesFromParsed(actual)
	if err != nil {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, "planned_change_scope invalid: "+err.Error())
		return
	}
	if len(entries) == 0 {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, "planned_change_scope must contain at least one pcs.<slug> item")
		return
	}
	expectedAcceptanceIDs := acceptanceItemIDSet(expected.AcceptanceItemSet)
	seen := map[string]bool{}
	for _, entry := range entries {
		if seen[entry.ID] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("planned_change_scope duplicate id: %s", entry.ID))
			continue
		}
		seen[entry.ID] = true
		if !strings.HasPrefix(entry.ID, "pcs.") || strings.TrimSpace(strings.TrimPrefix(entry.ID, "pcs.")) == "" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("planned_change_scope id must use pcs.<slug>: %s", entry.ID))
		}
		if strings.TrimSpace(entry.VerificationAction) == "" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("planned_change_scope verification_action for %s must not be empty", entry.ID))
		}
		basisRefs := strings.TrimSpace(entry.BasisRefs)
		if basisRefs == "" || basisRefs == "none" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("planned_change_scope basis_refs for %s must cite package refs", entry.ID))
		} else if validateNoneOrRefList(result, "planned_change_scope basis_refs for "+entry.ID, basisRefs) {
			validatePackageBasisRefs(result, "planned_change_scope basis_refs for "+entry.ID, splitProcessRefList(basisRefs), expected)
		}
		acceptanceIDs, err := splitCommaSeparatedField(entry.AcceptanceItemIDs)
		if err != nil {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("planned_change_scope acceptance_item_ids invalid for %s: %v", entry.ID, err))
		} else {
			seenAcceptanceIDs := map[string]bool{}
			for _, id := range acceptanceIDs {
				if seenAcceptanceIDs[id] {
					result.Valid = false
					result.Mismatches = append(result.Mismatches, fmt.Sprintf("planned_change_scope duplicate acceptance_item_ids for %s: %s", entry.ID, id))
					continue
				}
				seenAcceptanceIDs[id] = true
				if !expectedAcceptanceIDs[id] {
					result.Valid = false
					result.Mismatches = append(result.Mismatches, fmt.Sprintf("planned_change_scope unknown acceptance_item_ids for %s: %s", entry.ID, id))
				}
			}
		}
		implementationRefs := strings.TrimSpace(entry.ImplementationRefs)
		if implementationRefs == "" || implementationRefs == "none" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("planned_change_scope implementation_refs for %s must cite durable refs", entry.ID))
		} else if validateNoneOrRefList(result, "planned_change_scope implementation_refs for "+entry.ID, implementationRefs) {
			validateRepositoryRelativeExistingRefs(repoRoot, result, "planned_change_scope implementation_refs for "+entry.ID, splitProcessRefList(implementationRefs))
		}
	}
}

func validatePackageDeltaVerification(repoRoot string, result *ValidationResult, actual processSnapshot, expected Snapshot, plannedEntries []PlannedChangeScopeEntry) {
	actualEntries, err := packageDeltaVerificationEntriesFromParsed(actual)
	if err != nil {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, "package_delta_verification invalid: "+err.Error())
		return
	}
	expectedIDs := map[string]bool{}
	for _, entry := range plannedEntries {
		expectedIDs[entry.ID] = true
	}
	actualIDs := map[string]bool{}
	for _, entry := range actualEntries {
		if actualIDs[entry.PlannedChangeScopeID] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("package_delta_verification duplicate planned_change_scope_id: %s", entry.PlannedChangeScopeID))
			continue
		}
		actualIDs[entry.PlannedChangeScopeID] = true
		if !expectedIDs[entry.PlannedChangeScopeID] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("package_delta_verification unknown planned_change_scope_id: %s", entry.PlannedChangeScopeID))
			continue
		}
		if !allowedPackageDeltaVerificationResults[entry.Result] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("package_delta_verification invalid result for %s: %s", entry.PlannedChangeScopeID, entry.Result))
		}
		if entry.Result != "pass" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("package_delta_verification result for %s must be pass", entry.PlannedChangeScopeID))
		}
		if strings.TrimSpace(entry.EvidenceRefs) == "" || strings.TrimSpace(entry.EvidenceRefs) == "none" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("package_delta_verification evidence_refs for %s must be durable refs", entry.PlannedChangeScopeID))
		}
	}
	for id := range expectedIDs {
		if !actualIDs[id] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("package_delta_verification missing planned_change_scope_id: %s", id))
		}
	}
}

func validatePackageBasisRefs(result *ValidationResult, field string, refs []string, expected Snapshot) {
	allowed := packageBasisRefSet(expected)
	for _, ref := range refs {
		if !allowed[ref] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("%s unknown package ref: %s", field, ref))
		}
	}
}

func packageBasisRefSet(expected Snapshot) map[string]bool {
	refs := map[string]bool{}
	if expected.SpecFileRef != "" {
		refs[expected.SpecFileRef] = true
	}
	for _, entry := range expected.ModuleAppendixSnapshot {
		if entry.FileRef != "" {
			refs[entry.FileRef] = true
		}
	}
	for _, entry := range expected.UnitSnapshot {
		if entry.FileRef != "" {
			refs[entry.FileRef] = true
		}
	}
	for _, entry := range expected.RuleSnapshot {
		if entry.FileRef != "" {
			refs[entry.FileRef] = true
		}
	}
	return refs
}

func validateNoneOrRefList(result *ValidationResult, field, value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "none" {
		return true
	}
	refs := splitProcessRefList(trimmed)
	if len(refs) == 0 {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, field+" must be none or durable refs")
		return false
	}
	for _, ref := range refs {
		if ref == "none" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, field+" must not mix none with refs")
			return false
		}
	}
	return true
}

func acceptanceItemsDeclareImplementationSurface(entries []AcceptanceItemEntry) bool {
	for _, entry := range entries {
		surface := strings.TrimSpace(entry.ImplementationSurface)
		if surface != "" && !strings.EqualFold(surface, "none") {
			return true
		}
	}
	return false
}

func validateRepositoryRelativeExistingRefs(repoRoot string, result *ValidationResult, field string, refs []string) {
	for _, ref := range refs {
		if !isRepositoryRelativePathRef(ref) {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("%s ref must be a repository-relative path: %s", field, ref))
			continue
		}
		absPath := filepath.Join(repoRoot, filepath.FromSlash(ref))
		rel, err := filepath.Rel(repoRoot, absPath)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("%s ref escapes repository root: %s", field, ref))
			continue
		}
		if _, err := os.Stat(absPath); err != nil {
			result.Valid = false
			if os.IsNotExist(err) {
				result.Mismatches = append(result.Mismatches, fmt.Sprintf("%s ref does not exist: %s", field, ref))
			} else {
				result.Mismatches = append(result.Mismatches, fmt.Sprintf("stat %s ref %s: %v", field, ref, err))
			}
		}
	}
}

func isRepositoryRelativePathRef(ref string) bool {
	ref = strings.TrimSpace(ref)
	if ref == "" ||
		ref == "." ||
		ref == ".." ||
		strings.Contains(ref, "\\") ||
		strings.Contains(ref, ":") ||
		strings.Contains(ref, "://") ||
		strings.Contains(ref, "\t") ||
		strings.Contains(ref, "\n") ||
		filepath.IsAbs(ref) ||
		filepath.VolumeName(ref) != "" {
		return false
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(ref)))
	return clean == ref && !strings.HasPrefix(clean, "../")
}

func stableMainSpecRefIfPresent(repoRoot, objectType, object string) (string, bool, error) {
	ref, err := specpaths.ObjectMainSpecFileRef(objectType, "stable", object)
	if err != nil {
		return "", false, err
	}
	_, err = os.Stat(filepath.Join(repoRoot, filepath.FromSlash(ref)))
	if err == nil {
		return ref, true, nil
	}
	if os.IsNotExist(err) {
		return ref, false, nil
	}
	return ref, false, fmt.Errorf("stat %s: %w", ref, err)
}

func validateAcceptanceEvidenceMatrix(result *ValidationResult, actual processSnapshot, expected []AcceptanceItemEntry, requirePromotionReadyEvidence bool) {
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
		if expectedItem.NotRunnableYet == "no" {
			if entry.Status == "not_runnable_yet" {
				result.Valid = false
				result.Mismatches = append(result.Mismatches, fmt.Sprintf("acceptance_item_evidence_matrix status for %s cannot be not_runnable_yet", entry.ID))
			}
			if requirePromotionReadyEvidence && entry.Status != "pass" {
				result.Valid = false
				result.Mismatches = append(result.Mismatches, fmt.Sprintf("acceptance_item_evidence_matrix status for %s must be pass", entry.ID))
			}
			if requirePromotionReadyEvidence && (strings.TrimSpace(entry.EvidenceRefs) == "" || strings.TrimSpace(entry.EvidenceRefs) == "none") {
				result.Valid = false
				result.Mismatches = append(result.Mismatches, fmt.Sprintf("acceptance_item_evidence_matrix evidence_refs for %s must be durable refs", entry.ID))
			}
		}
	}
	for _, item := range normalizeAcceptanceItemEntries(expected) {
		if !actualIDs[item.ID] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("acceptance_item_evidence_matrix missing id: %s", item.ID))
		}
	}
}

func validateRetirementTargets(repoRoot string, result *ValidationResult, actual processSnapshot, expected Snapshot) {
	actualEntries, err := retirementTargetEntriesFromParsed(actual)
	if err != nil {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, "retirement_targets invalid: "+err.Error())
		return
	}
	intent, sourceBasis, err := candidateIntentAndSourceBasis(repoRoot, expected.ObjectType, expected.Object)
	if err != nil {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, "retirement_targets candidate metadata unavailable: "+err.Error())
		return
	}
	if intent == "change" && sourceBasis == "replacement" && len(actualEntries) == 0 {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, "retirement_targets must not be none for change replacement candidates")
	}
	expectedIDs := acceptanceItemIDSet(expected.AcceptanceItemSet)
	actualIDs := map[string]bool{}
	for _, entry := range actualEntries {
		if actualIDs[entry.ID] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("retirement_targets duplicate id: %s", entry.ID))
			continue
		}
		actualIDs[entry.ID] = true
		if !strings.HasPrefix(entry.ID, "rt.") || strings.TrimSpace(strings.TrimPrefix(entry.ID, "rt.")) == "" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("retirement_targets id must use rt.<slug>: %s", entry.ID))
		}
		if !allowedRetirementTargetKinds[entry.TargetKind] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("retirement_targets target_kind for %s must be one of path, helper, wrapper, compat_layer, dependency, other", entry.ID))
		}
		if !allowedRetirementMethods[entry.RetirementMethod] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("retirement_targets retirement_method for %s must be one of remove, reroute, replace, isolate", entry.ID))
		}
		acceptanceIDs, err := splitCommaSeparatedField(entry.AcceptanceItemIDs)
		if err != nil {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("retirement_targets acceptance_item_ids invalid for %s: %v", entry.ID, err))
			continue
		}
		seenAcceptanceIDs := map[string]bool{}
		for _, id := range acceptanceIDs {
			if seenAcceptanceIDs[id] {
				result.Valid = false
				result.Mismatches = append(result.Mismatches, fmt.Sprintf("retirement_targets duplicate acceptance_item_ids for %s: %s", entry.ID, id))
				continue
			}
			seenAcceptanceIDs[id] = true
			if !expectedIDs[id] {
				result.Valid = false
				result.Mismatches = append(result.Mismatches, fmt.Sprintf("retirement_targets unknown acceptance_item_ids for %s: %s", entry.ID, id))
			}
		}
	}
}

func candidateIntentAndSourceBasis(repoRoot, objectType, object string) (string, string, error) {
	frontmatter, err := candidateFrontmatterForObject(repoRoot, objectType, object)
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(frontmatter["candidate_intent"]), strings.TrimSpace(frontmatter["source_basis"]), nil
}

func validateActivePlanBinding(repoRoot string, result *ValidationResult, actual processSnapshot, expected Snapshot) {
	expectedPlanRef := ActivePlanFilePath(expected.Object)
	compareScalar(result, "active_plan_file_ref", actual.scalars["active_plan_file_ref"], expectedPlanRef)
	expectedFingerprint, err := fileFingerprint(repoRoot, expectedPlanRef)
	if err != nil {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, "active_plan_fingerprint unavailable: "+err.Error())
		return
	}
	compareScalar(result, "active_plan_fingerprint", actual.scalars["active_plan_fingerprint"], expectedFingerprint)
}

func validateRetirementEvidence(repoRoot string, result *ValidationResult, actual processSnapshot, expected Snapshot, targetEntries []RetirementTargetEntry) {
	actualEntries, err := retirementEvidenceEntriesFromParsed(actual)
	if err != nil {
		result.Valid = false
		result.Mismatches = append(result.Mismatches, "retirement_evidence_matrix invalid: "+err.Error())
		return
	}
	if len(targetEntries) == 0 {
		if len(actualEntries) != 0 {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, "retirement_evidence_matrix must be none when retirement_targets is none")
		}
		return
	}

	expectedIDs := map[string]bool{}
	for _, entry := range targetEntries {
		expectedIDs[entry.ID] = true
	}
	actualIDs := map[string]bool{}
	for _, entry := range actualEntries {
		if actualIDs[entry.ID] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("retirement_evidence_matrix duplicate id: %s", entry.ID))
			continue
		}
		actualIDs[entry.ID] = true
		if !expectedIDs[entry.ID] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("retirement_evidence_matrix unknown id: %s", entry.ID))
			continue
		}
		if !allowedRetirementEvidenceResults[entry.Result] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("retirement_evidence_matrix invalid result for %s: %s", entry.ID, entry.Result))
		}
		if !allowedMainlineDependencyResults[entry.MainlineDependency] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("retirement_evidence_matrix invalid mainline_dependency for %s: %s", entry.ID, entry.MainlineDependency))
		}
		if strings.TrimSpace(entry.EvidenceRefs) == "" || strings.TrimSpace(entry.EvidenceRefs) == "none" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("retirement_evidence_matrix evidence_refs for %s must be durable refs", entry.ID))
		}
		if entry.Result != "pass" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("retirement_evidence_matrix result for %s must be pass", entry.ID))
		}
		if entry.MainlineDependency != "not_required" {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("retirement_evidence_matrix mainline_dependency for %s must be not_required", entry.ID))
		}
	}
	for _, entry := range targetEntries {
		if !actualIDs[entry.ID] {
			result.Valid = false
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("retirement_evidence_matrix missing id: %s", entry.ID))
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
		if strings.TrimSpace(item.ID) == "" ||
			strings.TrimSpace(item.Status) == "" ||
			strings.TrimSpace(item.EvidenceRefs) == "" {
			return nil, fmt.Errorf("each item must include id, status, and evidence_refs")
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func retirementTargetEntriesFromParsed(parsed processSnapshot) ([]RetirementTargetEntry, error) {
	if raw, ok := parsed.scalars["retirement_targets"]; ok {
		if raw != "none" {
			return nil, fmt.Errorf("must be literal none or a list")
		}
		return nil, nil
	}
	if !parsed.retirementTargetPresent {
		return nil, nil
	}
	items := append([]RetirementTargetEntry(nil), parsed.retirementTargetEntries...)
	if len(items) == 0 {
		return nil, fmt.Errorf("must be literal none or a non-empty list")
	}
	for _, item := range items {
		if strings.TrimSpace(item.ID) == "" ||
			strings.TrimSpace(item.TargetRef) == "" ||
			strings.TrimSpace(item.TargetKind) == "" ||
			strings.TrimSpace(item.RetirementMethod) == "" ||
			strings.TrimSpace(item.VerificationAction) == "" ||
			strings.TrimSpace(item.AcceptanceItemIDs) == "" {
			return nil, fmt.Errorf("each item must include id, target_ref, target_kind, retirement_method, verification_action, and acceptance_item_ids")
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func retirementEvidenceEntriesFromParsed(parsed processSnapshot) ([]RetirementEvidenceEntry, error) {
	if raw, ok := parsed.scalars["retirement_evidence_matrix"]; ok {
		if raw != "none" {
			return nil, fmt.Errorf("must be literal none or a list")
		}
		return nil, nil
	}
	if !parsed.retirementEvidencePresent {
		return nil, nil
	}
	items := append([]RetirementEvidenceEntry(nil), parsed.retirementEvidenceEntries...)
	if len(items) == 0 {
		return nil, fmt.Errorf("must be literal none or a non-empty list")
	}
	for _, item := range items {
		if strings.TrimSpace(item.ID) == "" ||
			strings.TrimSpace(item.Result) == "" ||
			strings.TrimSpace(item.MainlineDependency) == "" ||
			strings.TrimSpace(item.EvidenceRefs) == "" {
			return nil, fmt.Errorf("each item must include id, result, mainline_dependency, and evidence_refs")
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func plannedChangeScopeEntriesFromParsed(parsed processSnapshot) ([]PlannedChangeScopeEntry, error) {
	if raw, ok := parsed.scalars["planned_change_scope"]; ok {
		if raw != "none" {
			return nil, fmt.Errorf("must be literal none or a list")
		}
		return nil, nil
	}
	if !parsed.plannedChangePresent {
		return nil, nil
	}
	items := append([]PlannedChangeScopeEntry(nil), parsed.plannedChangeEntries...)
	if len(items) == 0 {
		return nil, fmt.Errorf("must be a non-empty list")
	}
	for _, item := range items {
		if strings.TrimSpace(item.ID) == "" ||
			strings.TrimSpace(item.BasisRefs) == "" ||
			strings.TrimSpace(item.AcceptanceItemIDs) == "" ||
			strings.TrimSpace(item.ImplementationRefs) == "" ||
			strings.TrimSpace(item.VerificationAction) == "" {
			return nil, fmt.Errorf("each item must include id, basis_refs, acceptance_item_ids, implementation_refs, and verification_action")
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func packageDeltaVerificationEntriesFromParsed(parsed processSnapshot) ([]PackageDeltaVerificationEntry, error) {
	if raw, ok := parsed.scalars["package_delta_verification"]; ok {
		if raw != "none" {
			return nil, fmt.Errorf("must be literal none or a list")
		}
		return nil, nil
	}
	if !parsed.packageDeltaPresent {
		return nil, nil
	}
	items := append([]PackageDeltaVerificationEntry(nil), parsed.packageDeltaEntries...)
	if len(items) == 0 {
		return nil, fmt.Errorf("must be a non-empty list")
	}
	for _, item := range items {
		if strings.TrimSpace(item.PlannedChangeScopeID) == "" ||
			strings.TrimSpace(item.Result) == "" ||
			strings.TrimSpace(item.EvidenceRefs) == "" {
			return nil, fmt.Errorf("each item must include planned_change_scope_id, result, and evidence_refs")
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].PlannedChangeScopeID < items[j].PlannedChangeScopeID
	})
	return items, nil
}


func parseActivePlan(repoRoot, object string) (processSnapshot, error) {
	planRef := ActivePlanFilePath(object)
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(planRef)))
	if err != nil {
		return processSnapshot{}, fmt.Errorf("read %s: %w", planRef, err)
	}
	parsed, err := parseProcessSnapshot(string(content))
	if err != nil {
		return processSnapshot{}, fmt.Errorf("%s: %w", planRef, err)
	}
	return parsed, nil
}

func activePlanRetirementTargets(repoRoot, object string) ([]RetirementTargetEntry, error) {
	planRef := ActivePlanFilePath(object)
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(planRef)))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", planRef, err)
	}
	parsed, err := parseProcessSnapshot(string(content))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", planRef, err)
	}
	if !parsed.presentFields["retirement_targets"] {
		return nil, fmt.Errorf("missing required field: retirement_targets")
	}
	return retirementTargetEntriesFromParsed(parsed)
}

func fileFingerprint(repoRoot, fileRef string) (string, error) {
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(fileRef)))
	if err != nil {
		return "", fmt.Errorf("read %s: %w", fileRef, err)
	}
	return hashNormalizedText(string(content)), nil
}

func splitCommaSeparatedField(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "none" {
		return nil, fmt.Errorf("must be comma-separated current acceptance item ids")
	}
	rawParts := strings.Split(value, ",")
	parts := make([]string, 0, len(rawParts))
	for _, raw := range rawParts {
		part := strings.TrimSpace(raw)
		if part == "" {
			return nil, fmt.Errorf("must not contain empty ids")
		}
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("must be comma-separated current acceptance item ids")
	}
	return parts, nil
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
		if items[i].VerificationSurface != items[j].VerificationSurface {
			return items[i].VerificationSurface < items[j].VerificationSurface
		}
		return items[i].NotRunnableYet < items[j].NotRunnableYet
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

func fingerprintAcceptanceBehavior(entries []AcceptanceItemEntry) string {
	items := normalizeAcceptanceItemEntries(entries)
	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, strings.Join([]string{
			"id=" + item.ID,
			"target=" + item.Target,
			"verification_surface=" + item.VerificationSurface,
			"implementation_surface=" + item.ImplementationSurface,
			"verification_method=" + item.VerificationMethod,
			"pass_condition=" + item.PassCondition,
			"not_runnable_yet=" + item.NotRunnableYet,
			"not_runnable_yet_reason=" + item.NotRunnableYetReason,
		}, "\x1f"))
	}
	return hashNormalizedText(strings.Join(parts, "\n"))
}

// ComputeAcceptanceBehaviorFingerprint computes the acceptance_behavior_fingerprint
// hash for the given acceptance items. See process_snapshot_contract.md Sections 6-6a
// for the fingerprint contract and the serialization format.
func ComputeAcceptanceBehaviorFingerprint(entries []AcceptanceItemEntry) string {
	return fingerprintAcceptanceBehavior(entries)
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

func finalizeValidationResult(result *ValidationResult, actual processSnapshot) {
	result.FreshnessImpact = classifyFreshnessImpact(result, actual)
	result.EvidenceReuse = evidenceReuseForImpact(result.FreshnessImpact)

	if result.FreshnessImpact == FreshnessTextDrift {
		primaryField := primaryFingerprintField(result.ProcessKind)
		if freshnessReceiptValid(result, actual, primaryField) && hasOnlyPrimaryFingerprintMismatch(result.Mismatches, primaryField) {
			result.Mismatches = removePrimaryFingerprintMismatches(result.Mismatches, primaryField)
			result.Valid = true
			result.EvidenceReuse = EvidenceReuseAccepted
		} else {
			result.Valid = false
			if result.EvidenceReuse != EvidenceReuseRejected {
				result.EvidenceReuse = EvidenceReusePendingReview
			}
		}
	}

	if result.Valid {
		result.FailureLayer = "none"
		result.NextCommand = ""
		return
	}
	if result.FreshnessImpact == FreshnessTextDrift {
		result.FailureLayer = "freshness_layer"
		result.NextCommand = ""
		return
	}
	result.FailureLayer = classifyFailureLayer(result.ObjectType, result.ProcessKind, result.Mismatches)
	result.NextCommand = nextCommandForFailureLayer(result.ObjectType, result.ProcessKind, result.FailureLayer)
}

func classifyFreshnessImpact(result *ValidationResult, actual processSnapshot) string {
	primaryField := primaryFingerprintField(result.ProcessKind)
	hasPrimaryDrift := hasPrimaryFingerprintMismatch(result.Mismatches, primaryField)
	if !hasPrimaryDrift {
		switch {
		case len(result.Mismatches) == 0:
			return FreshnessCurrent
		case hasDependencyMismatch(result.Mismatches):
			return FreshnessDependencyDrift
		case hasAcceptanceSetMismatch(result.Mismatches):
			return FreshnessAcceptanceDrift
		default:
			return FreshnessSchemaDrift
		}
	}

	switch {
	case hasDependencyMismatch(result.Mismatches):
		return FreshnessDependencyDrift
	case hasAcceptanceSetMismatch(result.Mismatches):
		return FreshnessAcceptanceDrift
	case hasPrimaryRefMismatch(result.Mismatches):
		return FreshnessSemanticDrift
	case hasSchemaMismatch(result.Mismatches, primaryField):
		return FreshnessSchemaDrift
	}

	actualBehavior := strings.TrimSpace(actual.scalars["acceptance_behavior_fingerprint"])
	if actualBehavior == "" {
		return FreshnessUnknownDrift
	}
	if actualBehavior != result.Expected.AcceptanceBehaviorFingerprint {
		return FreshnessSemanticDrift
	}
	return FreshnessTextDrift
}

func evidenceReuseForImpact(impact string) string {
	switch impact {
	case FreshnessCurrent:
		return EvidenceReuseNotNeeded
	case FreshnessTextDrift:
		return EvidenceReusePendingReview
	default:
		return EvidenceReuseNotEligible
	}
}

func primaryFingerprintField(processKind string) string {
	switch processKind {
	case "plan":
		return "spec_fingerprint"
	case "check", "verify", "stable_verify":
		return "truth_fingerprint"
	default:
		return ""
	}
}

func freshnessReceiptValid(result *ValidationResult, actual processSnapshot, primaryField string) bool {
	valid := true
	for _, field := range freshnessReceiptFields {
		if !actual.presentFields[field] {
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("missing required freshness field: %s", field))
			valid = false
		}
	}
	if !valid {
		return false
	}

	expectedCurrentFingerprint := result.Expected.SpecFingerprint
	if primaryField == "" {
		expectedCurrentFingerprint = ""
	}
	checks := []struct {
		Field    string
		Expected string
	}{
		{"freshness_impact", FreshnessTextDrift},
		{"evidence_reuse", EvidenceReuseAccepted},
		{"freshness_current_fingerprint", expectedCurrentFingerprint},
		{"freshness_review_mode", "independent"},
		{"freshness_reviewer_result", "pass"},
		{"freshness_reviewer_context", "minimal_context"},
		{"freshness_review_findings", "none"},
	}
	for _, check := range checks {
		actualValue := strings.TrimSpace(actual.scalars[check.Field])
		if actualValue != check.Expected {
			result.Mismatches = append(result.Mismatches, fmt.Sprintf("%s mismatch: actual=%s expected=%s", check.Field, actualValue, check.Expected))
			valid = false
			if check.Field == "freshness_reviewer_result" || check.Field == "freshness_review_findings" {
				result.EvidenceReuse = EvidenceReuseRejected
			}
		}
	}
	before := len(result.Mismatches)
	validateEvaluationInputRefs(result, "freshness_review_input_refs", actual.scalars["freshness_review_input_refs"], "freshness_text_drift_reuse")
	if len(result.Mismatches) != before {
		valid = false
	}
	return valid
}

func hasPrimaryFingerprintMismatch(mismatches []string, primaryField string) bool {
	if primaryField == "" {
		return false
	}
	prefix := primaryField + " mismatch:"
	for _, mismatch := range mismatches {
		if strings.HasPrefix(mismatch, prefix) {
			return true
		}
	}
	return false
}

func hasOnlyPrimaryFingerprintMismatch(mismatches []string, primaryField string) bool {
	if len(mismatches) == 0 || primaryField == "" {
		return false
	}
	for _, mismatch := range mismatches {
		if !strings.HasPrefix(mismatch, primaryField+" mismatch:") {
			return false
		}
	}
	return true
}

func removePrimaryFingerprintMismatches(mismatches []string, primaryField string) []string {
	result := make([]string, 0, len(mismatches))
	for _, mismatch := range mismatches {
		if strings.HasPrefix(mismatch, primaryField+" mismatch:") {
			continue
		}
		result = append(result, mismatch)
	}
	return result
}

func hasDependencyMismatch(mismatches []string) bool {
	for _, mismatch := range mismatches {
		if strings.Contains(mismatch, "unit_appendix_snapshot mismatch") ||
			strings.Contains(mismatch, "repository_mapping_snapshot mismatch") ||
			strings.Contains(mismatch, "unit_snapshot mismatch") ||
			strings.Contains(mismatch, "rule_snapshot mismatch") {
			return true
		}
	}
	return false
}

func hasAcceptanceSetMismatch(mismatches []string) bool {
	for _, mismatch := range mismatches {
		if strings.Contains(mismatch, "acceptance_item_set mismatch") {
			return true
		}
	}
	return false
}

func hasPrimaryRefMismatch(mismatches []string) bool {
	for _, mismatch := range mismatches {
		if strings.Contains(mismatch, "truth_file_ref mismatch") ||
			strings.Contains(mismatch, "truth_version_ref mismatch") ||
			strings.Contains(mismatch, "spec_file_ref mismatch") ||
			strings.Contains(mismatch, "spec_version_ref mismatch") {
			return true
		}
	}
	return false
}

func hasSchemaMismatch(mismatches []string, primaryField string) bool {
	for _, mismatch := range mismatches {
		if primaryField != "" && strings.HasPrefix(mismatch, primaryField+" mismatch:") {
			continue
		}
		if strings.Contains(mismatch, "acceptance_behavior_fingerprint mismatch") ||
			strings.Contains(mismatch, "acceptance_behavior_fingerprint must not be empty") {
			continue
		}
		if strings.Contains(mismatch, "acceptance_item_set mismatch") ||
			strings.Contains(mismatch, "unit_appendix_snapshot mismatch") ||
			strings.Contains(mismatch, "repository_mapping_snapshot mismatch") ||
			strings.Contains(mismatch, "unit_snapshot mismatch") ||
			strings.Contains(mismatch, "rule_snapshot mismatch") {
			continue
		}
		return true
	}
	return false
}

func classifyFailureLayer(objectType, processKind string, mismatches []string) string {
	for _, mismatch := range mismatches {
		switch {
		case strings.Contains(mismatch, "truth_"),
			strings.Contains(mismatch, "spec_file_ref mismatch"),
			strings.Contains(mismatch, "spec_version_ref mismatch"),
			strings.Contains(mismatch, "spec_fingerprint mismatch"),
			strings.Contains(mismatch, "acceptance_behavior_fingerprint mismatch"),
			strings.Contains(mismatch, "acceptance_item_set mismatch"),
			strings.Contains(mismatch, "unit_appendix_snapshot mismatch"),
			strings.Contains(mismatch, "repository_mapping_snapshot mismatch"),
			strings.Contains(mismatch, "unit_snapshot mismatch"),
			strings.Contains(mismatch, "rule_snapshot mismatch"):
			return "truth_layer"
		}
	}
	switch processKind {
	case "verify":
		return "evidence_layer"
	case "stable_verify":
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
		if processKind == "stable_verify" {
			return "unit_stable_verify"
		}
		switch failureLayer {
		case "freshness_layer":
			return ""
		case "truth_layer", "gate_layer":
			return "unit_check"
		case "evidence_layer":
			if processKind == "stable_verify" {
				return "unit_stable_verify"
			}
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

func CheckWorkFilePath(objectType, object string) string {
	return fmt.Sprintf("docs/specs/_check_work/%s/%s.md", objectType, object)
}

func VerifyResultFilePath(objectType, object string) string {
	return fmt.Sprintf("docs/specs/_verify_result/%s/%s.md", objectType, object)
}

func StablePromotionSummaryFilePath(objectType, object string) string {
	return fmt.Sprintf("docs/specs/_verify_result/stable/%s/%s.md", objectType, object)
}

func StableVerifyResultFilePath(objectType, object string) string {
	return fmt.Sprintf("docs/specs/_stable_verify_result/%s/%s.md", objectType, object)
}

func ProcessArtifactPaths(objectType, object, processKind string) ([]string, error) {
	if objectType != "unit" {
		return nil, fmt.Errorf("object type %q is not supported; only unit is supported", objectType)
	}
	switch processKind {
	case "check_work":
		return []string{CheckWorkFilePath(objectType, object)}, nil
	case "check":
		return []string{CheckResultFilePath(objectType, object)}, nil
	case "plan":
		if objectType != "unit" {
			return nil, fmt.Errorf("process kind %q is not supported for object type %q", processKind, objectType)
		}
		return []string{DraftPlanFilePath(object), ActivePlanFilePath(object)}, nil
	case "verify":
		return []string{VerifyResultFilePath(objectType, object)}, nil
	case "stable_verify":
		return []string{StableVerifyResultFilePath(objectType, object)}, nil
	default:
		return nil, fmt.Errorf("unsupported process kind %q", processKind)
	}
}

func ProcessFilePath(objectType, object, processKind string) (string, error) {
	if objectType != "unit" {
		return "", fmt.Errorf("object type %q is not supported; only unit is supported", objectType)
	}
	switch processKind {
	case "check_work":
		return CheckWorkFilePath(objectType, object), nil
	case "check":
		return CheckResultFilePath(objectType, object), nil
	case "plan":
		if objectType != "unit" {
			return "", fmt.Errorf("process kind %q is not supported for object type %q", processKind, objectType)
		}
		return ActivePlanFilePath(object), nil
	case "verify":
		return VerifyResultFilePath(objectType, object), nil
	case "stable_verify":
		return StableVerifyResultFilePath(objectType, object), nil
	default:
		return "", fmt.Errorf("unsupported process kind %q", processKind)
	}
}
