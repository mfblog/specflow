package processcleanup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

type CleanupResult struct {
	ObjectType    string
	Object        string
	Module        string
	FromCommand   string
	Reason        string
	FailureLayer  string
	NextCommand   string
	DeletedFiles  []string
	MissingFiles  []string
	StatusUpdated bool
}

type SuccessCleanupResult struct {
	ObjectType   string
	Object       string
	Module       string
	Mode         string
	DeletedFiles []string
	MissingFiles []string
}

type cleanupRule struct {
	NextCommand string
	FileKinds   []string
}

var fallbackReasonLayers = map[string]string{
	"truth_drift":              "truth_layer",
	"binding_drift":            "truth_layer",
	"baseline_drift":           "truth_layer",
	"rule_drift":               "truth_layer",
	"truth_incomplete":         "truth_layer",
	"gate_missing":             "gate_layer",
		"evidence_incomplete":      "evidence_layer",
	"stable_verify_invalid":    "evidence_layer",
}

var layeredRules = map[string]map[string]cleanupRule{
	"unit": {
		"truth_layer":          {NextCommand: "unit_check", FileKinds: []string{"check_work", "check", "verify"}},
		"gate_layer":           {NextCommand: "unit_check", FileKinds: []string{"check_work", "check"}},
		"evidence_layer":       {NextCommand: "unit_verify", FileKinds: []string{"verify"}},
	},
}

func ValidateFallbackReason(reason, failureLayer string) error {
	reason = strings.TrimSpace(reason)
	failureLayer = strings.TrimSpace(failureLayer)
	if reason == "" {
		return fmt.Errorf("fallback reason is required")
	}
	if failureLayer == "" {
		return fmt.Errorf("failure layer is required")
	}
	expectedLayer, ok := fallbackReasonLayer(reason)
	if !ok {
		return fmt.Errorf("unsupported fallback reason %q", reason)
	}
	if failureLayer != expectedLayer {
		return fmt.Errorf("fallback reason %q requires failure layer %q, got %q", reason, expectedLayer, failureLayer)
	}
	return nil
}

func ApplyFallback(repoRoot, module, fromCommand, reason string) (CleanupResult, error) {
	layer, ok := fallbackReasonLayer(reason)
	if !ok {
		reason = strings.TrimSpace(reason)
		return CleanupResult{
			ObjectType:   "unit",
			Object:       strings.TrimSpace(module),
			Module:       strings.TrimSpace(module),
			FromCommand:  strings.TrimSpace(fromCommand),
			Reason:       reason,
			FailureLayer: "",
		}, fmt.Errorf("unsupported fallback reason %q", reason)
	}
	return ApplyObjectFallback(repoRoot, "unit", module, fromCommand, reason, layer)
}

func ApplyObjectFallback(repoRoot, objectType, object, fromCommand, reason, failureLayer string) (CleanupResult, error) {
	result := CleanupResult{
		ObjectType:   strings.TrimSpace(objectType),
		Object:       strings.TrimSpace(object),
		Module:       strings.TrimSpace(object),
		FromCommand:  strings.TrimSpace(fromCommand),
		Reason:       strings.TrimSpace(reason),
		FailureLayer: strings.TrimSpace(failureLayer),
	}

	if result.ObjectType == "" || result.Object == "" || result.FromCommand == "" || result.Reason == "" || result.FailureLayer == "" {
		return result, fmt.Errorf("object type, object, from command, reason, and failure layer are required")
	}
	if _, err := ensureFormalObject(repoRoot, result.ObjectType, result.Object); err != nil {
		return result, err
	}
	if err := ValidateFallbackReason(result.Reason, result.FailureLayer); err != nil {
		return result, err
	}

	rule, err := lookupFallbackCleanupRule(result.ObjectType, result.FailureLayer, result.FromCommand, result.Reason)
	if err != nil {
		return result, err
	}
	result.NextCommand = rule.NextCommand

	for _, relPath := range filePathsForObject(result.ObjectType, result.Object, rule.FileKinds) {
		absPath := filepath.Join(repoRoot, filepath.FromSlash(relPath))
		if _, err := os.Stat(absPath); err != nil {
			if os.IsNotExist(err) {
				result.MissingFiles = append(result.MissingFiles, relPath)
				continue
			}
			return result, fmt.Errorf("stat %s: %w", relPath, err)
		}
		if err := os.Remove(absPath); err != nil {
			return result, fmt.Errorf("delete %s: %w", relPath, err)
		}
		result.DeletedFiles = append(result.DeletedFiles, relPath)
	}

	updated, err := statusfile.UpdateObjectNextCommand(repoRoot, result.ObjectType, result.Object, result.NextCommand)
	if err != nil {
		return result, err
	}
	result.StatusUpdated = updated
	return result, nil
}

func ApplySuccessCleanup(repoRoot, module, mode string) (SuccessCleanupResult, error) {
	return ApplyObjectSuccessCleanup(repoRoot, "unit", module, mode)
}

func ApplyObjectSuccessCleanup(repoRoot, objectType, object, mode string) (SuccessCleanupResult, error) {
	result := SuccessCleanupResult{
		ObjectType: strings.TrimSpace(objectType),
		Object:     strings.TrimSpace(object),
		Module:     strings.TrimSpace(object),
		Mode:       strings.TrimSpace(mode),
	}
	if result.ObjectType == "" || result.Object == "" || result.Mode == "" {
		return result, fmt.Errorf("object type, object, and mode are required")
	}
	if _, err := ensureFormalObject(repoRoot, result.ObjectType, result.Object); err != nil {
		return result, err
	}
	if err := ensureSuccessCleanupPrerequisites(repoRoot, result.ObjectType, result.Object, result.Mode); err != nil {
		return result, err
	}

	paths, err := successCleanupPaths(repoRoot, result.ObjectType, result.Object, result.Mode)
	if err != nil {
		return result, err
	}
	for _, relPath := range paths {
		absPath := filepath.Join(repoRoot, filepath.FromSlash(relPath))
		if _, err := os.Stat(absPath); err != nil {
			if os.IsNotExist(err) {
				result.MissingFiles = append(result.MissingFiles, relPath)
				continue
			}
			return result, fmt.Errorf("stat %s: %w", relPath, err)
		}
		if err := os.Remove(absPath); err != nil {
			return result, fmt.Errorf("delete %s: %w", relPath, err)
		}
		result.DeletedFiles = append(result.DeletedFiles, relPath)
	}
	return result, nil
}

func ensureSuccessCleanupPrerequisites(repoRoot, objectType, object, mode string) error {
	if mode != "unit_promote" {
		return nil
	}
	if objectType != "unit" {
		return fmt.Errorf("mode %q requires object type unit", mode)
	}
	summaryRef := snapshot.StablePromotionSummaryFilePath(objectType, object)
	if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(summaryRef))); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("stable promotion summary is required before unit_promote cleanup: %s", summaryRef)
		}
		return fmt.Errorf("stat %s: %w", summaryRef, err)
	}
	return nil
}

func lookupFallbackCleanupRule(objectType, failureLayer, fromCommand, reason string) (cleanupRule, error) {
	if objectType == "unit" && failureLayer == "evidence_layer" && isStableVerifyEvidenceFallback(fromCommand, reason) {
		return cleanupRule{NextCommand: "unit_stable_verify", FileKinds: []string{"stable_verify"}}, nil
	}
	return lookupLayeredRule(objectType, failureLayer)
}

func isStableVerifyEvidenceFallback(fromCommand, reason string) bool {
	return fromCommand == "unit_stable_verify" || reason == "stable_verify_invalid"
}

func lookupLayeredRule(objectType, failureLayer string) (cleanupRule, error) {
	objectRules, ok := layeredRules[objectType]
	if !ok {
		return cleanupRule{}, fmt.Errorf("unsupported object type %q", objectType)
	}
	rule, ok := objectRules[failureLayer]
	if !ok {
		return cleanupRule{}, fmt.Errorf("no deterministic fallback cleanup is defined for %q + %q", objectType, failureLayer)
	}
	return rule, nil
}

func filePathsForModule(module string, fileKinds []string) []string {
	return filePathsForObject("unit", module, fileKinds)
}

func filePathsForObject(objectType, object string, fileKinds []string) []string {
	paths := make([]string, 0, len(fileKinds))
	for _, fileKind := range fileKinds {
		filePaths, err := snapshot.ProcessArtifactPaths(objectType, object, fileKind)
		if err != nil {
			continue
		}
		paths = append(paths, filePaths...)
	}
	return sortAndDedupeStrings(paths)
}

func fallbackReasonLayer(reason string) (string, bool) {
	layer, ok := fallbackReasonLayers[strings.TrimSpace(reason)]
	return layer, ok
}

func successCleanupPaths(repoRoot, objectType, object, mode string) ([]string, error) {
	paths := []string{}
	switch mode {
	case "unit_fork":
		if objectType != "unit" {
			return nil, fmt.Errorf("mode %q requires object type unit", mode)
		}
		paths = append(paths, filePathsForObject(objectType, object, []string{"check_work", "check", "plan", "verify", "stable_verify"})...)
	case "unit_promote":
		if objectType != "unit" {
			return nil, fmt.Errorf("mode %q requires object type unit", mode)
		}
		candidateMainRef, err := specpaths.ObjectMainSpecFileRef(objectType, "candidate", object)
		if err != nil {
			return nil, err
		}
		paths = append(paths, candidateMainRef)
		paths = append(paths, filePathsForObject(objectType, object, []string{"check_work", "check", "plan", "verify", "stable_verify"})...)
		appendixPaths, err := candidateAppendixPaths(repoRoot, objectType, object)
		if err != nil {
			return nil, err
		}
		paths = append(paths, appendixPaths...)
	default:
		return nil, fmt.Errorf("unsupported success cleanup mode %q", mode)
	}
	return sortAndDedupeStrings(paths), nil
}

func candidateAppendixPaths(repoRoot, objectType, object string) ([]string, error) {
	glob, err := specpaths.ObjectCandidateAppendixGlob(objectType, object)
	if err != nil {
		return nil, err
	}
	pattern := filepath.Join(repoRoot, filepath.FromSlash(glob))
	matches, err := filepath.Glob(pattern)
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
	return result, nil
}

func ensureFormalModule(repoRoot, module string) (bool, error) {
	return ensureFormalObject(repoRoot, "unit", module)
}

func ensureFormalObject(repoRoot, objectType, object string) (bool, error) {
	if objectType != "unit" {
		return false, fmt.Errorf("object type %q is not supported; only unit is supported", objectType)
	}
	statuses, err := statusfile.LoadObjectStatuses(repoRoot)
	if err != nil {
		return false, err
	}
	for _, candidate := range statuses {
		if candidate.ObjectType == objectType && candidate.Object == object {
			return true, nil
		}
	}
	return false, fmt.Errorf("%s %q is not registered in docs/specs/_status.md", objectType, object)
}

func sortAndDedupeStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
