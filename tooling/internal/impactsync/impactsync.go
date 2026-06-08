package impactsync

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/processcleanup"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

type ModuleBinding struct {
	Module        string
	ActiveLayer   string
	NextCommand   string
	BindingIssues []string
}

type ScopedModule struct {
	Binding                               ModuleBinding
	InvalidatingRuleRefs                  []string
	ExplicitFallbackScope                 bool
	AllowedSharedSnapshotMismatchFileRefs []string
}

type Input struct {
	Modules []ScopedModule
}

type Result struct {
	ModuleResults []ModuleResult
}

type ModuleResult struct {
	Module             string
	ActiveLayer        string
	Outcome            string
	FallbackReasonCode string
	FailureLayer       string
	NextCommand        string
	DeletedFiles       []string
	MissingFiles       []string
	StatusUpdated      bool
	Diagnostics        []string
}

func Apply(repoRoot string, input Input) (Result, error) {
	moduleResults := make([]ModuleResult, 0, len(input.Modules))
	for _, scoped := range input.Modules {
		result, err := reconcileModule(repoRoot, scoped)
		if err != nil {
			return Result{}, err
		}
		moduleResults = append(moduleResults, result)
	}

	return Result{
		ModuleResults: moduleResults,
	}, nil
}

func reconcileModule(repoRoot string, scoped ScopedModule) (ModuleResult, error) {
	binding := scoped.Binding
	result := ModuleResult{
		Module:      binding.Module,
		ActiveLayer: binding.ActiveLayer,
		Outcome:     "unchanged",
		NextCommand: binding.NextCommand,
		Diagnostics: append([]string{}, binding.BindingIssues...),
	}

	bindingIssue := len(binding.BindingIssues) > 0

	switch binding.ActiveLayer {
	case "candidate":
		return reconcileCandidate(repoRoot, binding, result, scoped.InvalidatingRuleRefs, scoped.ExplicitFallbackScope, bindingIssue, scoped.AllowedSharedSnapshotMismatchFileRefs)
	case "stable":
		return reconcileStable(repoRoot, binding, result, scoped.InvalidatingRuleRefs, scoped.ExplicitFallbackScope, bindingIssue)
	default:
		return ModuleResult{}, fmt.Errorf("unsupported active layer %q for module %s", binding.ActiveLayer, binding.Module)
	}
}

func reconcileCandidate(repoRoot string, binding ModuleBinding, result ModuleResult, invalidatingRuleRefs []string, explicitFallbackScope, bindingIssue bool, allowedSharedSnapshotMismatchFileRefs []string) (ModuleResult, error) {
	fallbackReason := ""
	if bindingIssue {
		fallbackReason = "binding_drift"
	}

	allowedSharedSnapshotMismatchFileRefSet := make(map[string]bool, len(allowedSharedSnapshotMismatchFileRefs))
	for _, fileRef := range allowedSharedSnapshotMismatchFileRefs {
		if strings.TrimSpace(fileRef) != "" {
			allowedSharedSnapshotMismatchFileRefSet[strings.TrimSpace(fileRef)] = true
		}
	}

	expectedSnapshot, err := snapshot.RebuildCurrent(repoRoot, binding.Module)
	if err != nil {
		if fallbackReason != "" {
			return applyCandidateFallback(repoRoot, result, fallbackReason, "truth_layer")
		}
		return ModuleResult{}, err
	}

	processFound := false
	sharedMismatch := false
	nonSharedMismatch := false
	freshnessReviewRequired := false
	failureLayer := ""
	for _, processKind := range []string{"check", "plan", "verify"} {
		processPath, err := snapshot.ProcessFilePath("unit", binding.Module, processKind)
		if err != nil {
			return ModuleResult{}, err
		}
		processAbs := filepath.Join(repoRoot, filepath.FromSlash(processPath))
		if _, err := os.Stat(processAbs); err != nil {
			if os.IsNotExist(err) {
				result.MissingFiles = append(result.MissingFiles, processPath)
				continue
			}
			return ModuleResult{}, fmt.Errorf("stat %s: %w", processPath, err)
		}
		processFound = true

		validation, err := snapshot.ValidateProcessFile(repoRoot, binding.Module, processKind)
		if err != nil {
			return ModuleResult{}, err
		}
		if validation.Valid {
			continue
		}
		if validation.FailureLayer == "freshness_layer" {
			freshnessReviewRequired = true
			result.Diagnostics = append(result.Diagnostics, prefixItems(validation.Mismatches, processKind)...)
			continue
		}
		if failureLayer == "" {
			failureLayer = validation.FailureLayer
		}

		processSnapshot, err := snapshot.LoadProcessSnapshot(repoRoot, "unit", binding.Module, processKind)
		if err != nil {
			return ModuleResult{}, err
		}
		if hasNonSharedMismatch(validation.Mismatches) {
			nonSharedMismatch = true
			result.Diagnostics = append(result.Diagnostics, prefixItems(validation.Mismatches, processKind)...)
			continue
		}

		equivalent, err := sharedSnapshotsEquivalentAllowingFileRefs(processSnapshot.RuleSnapshot, expectedSnapshot.RuleSnapshot, allowedSharedSnapshotMismatchFileRefSet)
		if err != nil {
			return ModuleResult{}, err
		}
		if equivalent {
			result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("%s snapshot differs only on caller-allowed rule file mismatch", processKind))
			continue
		}
		sharedMismatch = true
		result.Diagnostics = append(result.Diagnostics, prefixItems(validation.Mismatches, processKind)...)
	}

	switch {
	case fallbackReason != "":
		failureLayer = "truth_layer"
	case nonSharedMismatch:
		switch failureLayer {
			case "evidence_layer":
			fallbackReason = "evidence_incomplete"
		case "gate_layer":
			fallbackReason = "gate_missing"
		default:
			fallbackReason = "truth_drift"
			failureLayer = "truth_layer"
		}
	case sharedMismatch:
		fallbackReason = "rule_drift"
		failureLayer = "truth_layer"
	case !processFound && len(invalidatingRuleRefs) > 0:
		fallbackReason = "rule_drift"
		failureLayer = "truth_layer"
	case !processFound && explicitFallbackScope && binding.NextCommand != "unit_check":
		fallbackReason = "binding_drift"
		failureLayer = "truth_layer"
	case freshnessReviewRequired:
		result.Outcome = "freshness_review_required"
		result.FailureLayer = "freshness_layer"
	}

	if fallbackReason == "" {
		return result, nil
	}
	return applyCandidateFallback(repoRoot, result, fallbackReason, failureLayer)
}

func reconcileStable(repoRoot string, binding ModuleBinding, result ModuleResult, invalidatingRuleRefs []string, explicitFallbackScope, bindingIssue bool) (ModuleResult, error) {
	fallbackReason := ""
	switch {
	case bindingIssue:
		fallbackReason = "binding_drift"
	case len(invalidatingRuleRefs) > 0:
		fallbackReason = "rule_drift"
	case explicitFallbackScope && len(invalidatingRuleRefs) == 0:
		fallbackReason = "binding_drift"
	}

	if fallbackReason == "" {
		return result, nil
	}
	if err := processcleanup.ValidateFallbackReason(fallbackReason, "truth_layer"); err != nil {
		return ModuleResult{}, err
	}
	result.FallbackReasonCode = fallbackReason
	result.Outcome = "rerouted"
	result.NextCommand = "unit_stable_verify"
	processPaths, err := snapshot.ProcessArtifactPaths("unit", result.Module, "stable_verify")
	if err != nil {
		return ModuleResult{}, err
	}
	for _, processPath := range processPaths {
		processAbs := filepath.Join(repoRoot, filepath.FromSlash(processPath))
		if _, err := os.Stat(processAbs); err != nil {
			if os.IsNotExist(err) {
				if !contains(result.MissingFiles, processPath) {
					result.MissingFiles = append(result.MissingFiles, processPath)
				}
				continue
			}
			return ModuleResult{}, fmt.Errorf("stat %s: %w", processPath, err)
		}
		if err := os.Remove(processAbs); err != nil {
			return ModuleResult{}, fmt.Errorf("delete %s: %w", processPath, err)
		}
		result.DeletedFiles = append(result.DeletedFiles, processPath)
	}
	updated, err := statusfile.UpdateNextCommand(repoRoot, binding.Module, result.NextCommand)
	if err != nil {
		return ModuleResult{}, err
	}
	result.StatusUpdated = updated
	return result, nil
}

func applyCandidateFallback(repoRoot string, result ModuleResult, fallbackReason string, failureLayer string) (ModuleResult, error) {
	if err := processcleanup.ValidateFallbackReason(fallbackReason, failureLayer); err != nil {
		return ModuleResult{}, err
	}
	result.FallbackReasonCode = fallbackReason
	result.FailureLayer = failureLayer
	result.Outcome = "invalidated"
	processKinds := []string{"check_work", "check", "plan", "verify"}
	switch failureLayer {
	case "truth_layer":
		result.NextCommand = "unit_check"
	case "gate_layer":
		result.NextCommand = "unit_check"
		processKinds = []string{"check_work", "check"}
	case "evidence_layer":
		result.NextCommand = "unit_verify"
		processKinds = []string{"verify"}
	default:
		return ModuleResult{}, fmt.Errorf("unsupported failure layer %q", failureLayer)
	}
	for _, processKind := range processKinds {
		processPaths, err := snapshot.ProcessArtifactPaths("unit", result.Module, processKind)
		if err != nil {
			return ModuleResult{}, err
		}
		for _, processPath := range processPaths {
			processAbs := filepath.Join(repoRoot, filepath.FromSlash(processPath))
			if _, err := os.Stat(processAbs); err != nil {
				if os.IsNotExist(err) {
					if !contains(result.MissingFiles, processPath) {
						result.MissingFiles = append(result.MissingFiles, processPath)
					}
					continue
				}
				return ModuleResult{}, fmt.Errorf("stat %s: %w", processPath, err)
			}
			if err := os.Remove(processAbs); err != nil {
				return ModuleResult{}, fmt.Errorf("delete %s: %w", processPath, err)
			}
			result.DeletedFiles = append(result.DeletedFiles, processPath)
		}
	}
	updated, err := statusfile.UpdateNextCommand(repoRoot, result.Module, result.NextCommand)
	if err != nil {
		return ModuleResult{}, err
	}
	result.StatusUpdated = updated
	result.DeletedFiles = normalizeStrings(result.DeletedFiles)
	result.MissingFiles = normalizeStrings(result.MissingFiles)
	return result, nil
}

func hasNonSharedMismatch(mismatches []string) bool {
	for _, mismatch := range mismatches {
		if !strings.HasPrefix(mismatch, "rule_snapshot mismatch") {
			return true
		}
	}
	return false
}

func sharedSnapshotsEquivalentAllowingFileRefs(actual, expected []snapshot.RuleEntry, allowedFileRefs map[string]bool) (bool, error) {
	actual = normalizeSharedEntries(actual)
	expected = normalizeSharedEntries(expected)
	if len(actual) != len(expected) {
		return false, nil
	}
	for idx := range actual {
		if actual[idx].RuleID != expected[idx].RuleID ||
			actual[idx].Layer != expected[idx].Layer ||
			actual[idx].FileRef != expected[idx].FileRef ||
			actual[idx].VersionRef != expected[idx].VersionRef {
			return false, nil
		}
		if actual[idx].Fingerprint == expected[idx].Fingerprint {
			continue
		}
		if !allowedFileRefs[expected[idx].FileRef] {
			return false, nil
		}
	}
	return true, nil
}

func normalizeStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func normalizeSharedEntries(values []snapshot.RuleEntry) []snapshot.RuleEntry {
	result := append([]snapshot.RuleEntry(nil), values...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].RuleID != result[j].RuleID {
			return result[i].RuleID < result[j].RuleID
		}
		if result[i].Layer != result[j].Layer {
			return result[i].Layer < result[j].Layer
		}
		return result[i].FileRef < result[j].FileRef
	})
	return result
}

func prefixItems(items []string, prefix string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, fmt.Sprintf("%s: %s", prefix, item))
	}
	return result
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
