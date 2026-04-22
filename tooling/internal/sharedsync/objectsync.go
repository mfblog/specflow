package sharedsync

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/sharedbinding"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

type ObjectResult struct {
	Object             string
	ActiveLayer        string
	Outcome            string
	FallbackReasonCode string
	NextCommand        string
	DeletedFiles       []string
	MissingFiles       []string
	StatusUpdated      bool
	Diagnostics        []string
}

type objectBinding struct {
	Status        statusfile.ObjectStatus
	SharedRefs    []string
	BindingIssues []string
}

type objectFamilyConfig struct {
	ObjectType            string
	CandidateNextCommand  string
	StableNextCommand     string
	CandidateProcessKinds []string
}

func loadObjectBindings(repoRoot, objectType string) (map[string]objectBinding, []string, error) {
	statuses, err := statusfile.LoadObjectStatuses(repoRoot)
	if err != nil {
		return nil, nil, err
	}

	bindings := map[string]objectBinding{}
	unresolvedRefs := []string{}
	for _, status := range statuses {
		if status.ObjectType != objectType {
			continue
		}

		refs, err := readObjectSharedRefs(repoRoot, status)
		if err != nil {
			return nil, nil, err
		}

		bindingIssues := []string{}
		for _, ref := range refs {
			if _, err := sharedbinding.ResolveRef(repoRoot, status.ActiveLayer, ref); err != nil {
				bindingIssues = append(bindingIssues, err.Error())
				unresolvedRefs = append(unresolvedRefs, ref)
			}
		}

		bindings[status.Object] = objectBinding{
			Status:        status,
			SharedRefs:    refs,
			BindingIssues: normalizeStrings(bindingIssues),
		}
	}

	return bindings, normalizeStrings(unresolvedRefs), nil
}

func readObjectSharedRefs(repoRoot string, status statusfile.ObjectStatus) ([]string, error) {
	mainSpecRef, err := specpaths.ObjectMainSpecFileRef(status.ObjectType, status.ActiveLayer, status.Object)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", mainSpecRef, err)
	}
	_, body, err := parseFrontmatter(string(content))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", mainSpecRef, err)
	}
	refs, _, err := parseSharedContractRefs(body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", mainSpecRef, err)
	}
	return normalizeStrings(refs), nil
}

func buildScopeObjects(bindings map[string]objectBinding, sharedFilesByRef map[string]sharedFile, options Options) []string {
	if len(options.SharedRefs) == 0 && len(options.SharedIDs) == 0 {
		return nil
	}

	scope := map[string]bool{}
	for object, binding := range bindings {
		if len(selectedSharedRefsForObject(binding.SharedRefs, options.SharedRefs, options.SharedIDs, sharedFilesByRef)) > 0 {
			scope[object] = true
		}
	}

	result := make([]string, 0, len(scope))
	for object := range scope {
		result = append(result, object)
	}
	return normalizeStrings(result)
}

func reconcileObject(repoRoot string, binding objectBinding, relevantSelectedRefs []string, sharedFilesByRef map[string]sharedFile, boundModulesOnlyFileRefs map[string]bool, config objectFamilyConfig) (ObjectResult, error) {
	result := ObjectResult{
		Object:      binding.Status.Object,
		ActiveLayer: binding.Status.ActiveLayer,
		Outcome:     "unchanged",
		NextCommand: binding.Status.NextCommand,
		Diagnostics: append([]string{}, binding.BindingIssues...),
	}

	fallbackReason := ""
	switch {
	case len(binding.BindingIssues) > 0:
		fallbackReason = "binding_drift"
	case len(relevantSelectedRefs) > 0 && hasNonBoundModulesSelectedChange(relevantSelectedRefs, sharedFilesByRef, boundModulesOnlyFileRefs):
		fallbackReason = "shared_contract_drift"
	default:
		return result, nil
	}

	if binding.Status.ActiveLayer == "candidate" {
		return applyObjectCandidateFallback(repoRoot, result, config, fallbackReason)
	}
	return applyObjectStableReroute(repoRoot, result, config, fallbackReason)
}

func applyObjectCandidateFallback(repoRoot string, result ObjectResult, config objectFamilyConfig, fallbackReason string) (ObjectResult, error) {
	result.FallbackReasonCode = fallbackReason
	result.Outcome = "invalidated"
	result.NextCommand = config.CandidateNextCommand

	for _, processPath := range objectProcessPaths(result.Object, config.CandidateProcessKinds) {
		processAbs := filepath.Join(repoRoot, filepath.FromSlash(processPath))
		if _, err := os.Stat(processAbs); err != nil {
			if os.IsNotExist(err) {
				result.MissingFiles = append(result.MissingFiles, processPath)
				continue
			}
			return ObjectResult{}, fmt.Errorf("stat %s: %w", processPath, err)
		}
		if err := os.Remove(processAbs); err != nil {
			return ObjectResult{}, fmt.Errorf("delete %s: %w", processPath, err)
		}
		result.DeletedFiles = append(result.DeletedFiles, processPath)
	}

	updated, err := statusfile.UpdateObjectNextCommand(repoRoot, config.ObjectType, result.Object, result.NextCommand)
	if err != nil {
		return ObjectResult{}, err
	}
	result.StatusUpdated = updated
	result.DeletedFiles = normalizeStrings(result.DeletedFiles)
	result.MissingFiles = normalizeStrings(result.MissingFiles)
	return result, nil
}

func applyObjectStableReroute(repoRoot string, result ObjectResult, config objectFamilyConfig, fallbackReason string) (ObjectResult, error) {
	result.FallbackReasonCode = fallbackReason
	result.Outcome = "rerouted"
	result.NextCommand = config.StableNextCommand

	updated, err := statusfile.UpdateObjectNextCommand(repoRoot, config.ObjectType, result.Object, result.NextCommand)
	if err != nil {
		return ObjectResult{}, err
	}
	result.StatusUpdated = updated
	return result, nil
}

func objectProcessPaths(object string, processKinds []string) []string {
	paths := make([]string, 0, len(processKinds))
	for _, processKind := range processKinds {
		switch processKind {
		case "check":
			paths = append(paths, fmt.Sprintf("docs/specs/_check_result/%s.md", object))
		case "plan":
			paths = append(paths, fmt.Sprintf("docs/specs/_plans/%s.md", object))
		case "verify":
			paths = append(paths, fmt.Sprintf("docs/specs/_verify_result/%s.md", object))
		}
	}
	return normalizeStrings(paths)
}
