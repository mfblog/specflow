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

var layeredRules = map[string]map[string]cleanupRule{
	"unit": {
		"truth_layer":          {NextCommand: "unit_check", FileKinds: []string{"check", "plan", "verify"}},
		"gate_layer":           {NextCommand: "unit_check", FileKinds: []string{"check"}},
		"plan_layer":           {NextCommand: "unit_plan", FileKinds: []string{"plan", "verify"}},
		"implementation_layer": {NextCommand: "unit_impl", FileKinds: []string{"verify"}},
		"evidence_layer":       {NextCommand: "unit_verify", FileKinds: []string{"verify"}},
	},
	"scenario": {
		"truth_layer":                {NextCommand: "scenario_check", FileKinds: []string{"check", "verify"}},
		"gate_layer":                 {NextCommand: "scenario_check", FileKinds: []string{"check"}},
		"evidence_layer":             {NextCommand: "scenario_verify", FileKinds: []string{"verify"}},
		"dependency_readiness_layer": {NextCommand: "scenario_promote", FileKinds: nil},
	},
}

func ApplyFallback(repoRoot, module, fromCommand, reason string) (CleanupResult, error) {
	return ApplyObjectFallback(repoRoot, "unit", module, fromCommand, reason, inferFailureLayer("unit", fromCommand, reason))
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

	rule, err := lookupLayeredRule(result.ObjectType, result.FailureLayer)
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

func inferFailureLayer(objectType, fromCommand, reason string) string {
	switch reason {
	case "truth_drift", "binding_drift", "baseline_drift", "rule_drift", "truth_incomplete", "shared_truth_conflict", "governance_drift":
		return "truth_layer"
	case "implementation_deviation":
		return "implementation_layer"
	case "evidence_incomplete":
		return "evidence_layer"
	case "stable_dependency_not_ready":
		return "dependency_readiness_layer"
	case "gate_missing":
		if strings.HasSuffix(fromCommand, "_impl") {
			return "plan_layer"
		}
		return "gate_layer"
	default:
		if objectType == "unit" && strings.Contains(fromCommand, "_plan") {
			return "plan_layer"
		}
		return "truth_layer"
	}
}

func successCleanupPaths(repoRoot, objectType, object, mode string) ([]string, error) {
	paths := []string{}
	switch mode {
	case "unit_fork":
		if objectType != "unit" {
			return nil, fmt.Errorf("mode %q requires object type unit", mode)
		}
		paths = append(paths, filePathsForObject(objectType, object, []string{"check", "plan", "verify"})...)
	case "unit_promote":
		if objectType != "unit" {
			return nil, fmt.Errorf("mode %q requires object type unit", mode)
		}
		candidateMainRef, err := specpaths.ObjectMainSpecFileRef(objectType, "candidate", object)
		if err != nil {
			return nil, err
		}
		paths = append(paths, candidateMainRef)
		paths = append(paths, filePathsForObject(objectType, object, []string{"check", "plan", "verify"})...)
		appendixPaths, err := candidateAppendixPaths(repoRoot, objectType, object)
		if err != nil {
			return nil, err
		}
		paths = append(paths, appendixPaths...)
	case "scenario_fork":
		if objectType != "scenario" {
			return nil, fmt.Errorf("mode %q requires object type scenario", mode)
		}
		paths = append(paths, filePathsForObject(objectType, object, []string{"check", "verify"})...)
	case "scenario_promote":
		if objectType != "scenario" {
			return nil, fmt.Errorf("mode %q requires object type scenario", mode)
		}
		candidateMainRef, err := specpaths.ObjectMainSpecFileRef(objectType, "candidate", object)
		if err != nil {
			return nil, err
		}
		paths = append(paths, candidateMainRef)
		paths = append(paths, filePathsForObject(objectType, object, []string{"check", "verify"})...)
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
