package sharedsync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/sharedbinding"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

type objectBinding struct {
	Status        statusfile.ObjectStatus
	SharedRefs    []string
	BindingIssues []string
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

func buildScopeObjects(bindings map[string]objectBinding, sharedFilesByRef map[string]sharedFile, scopedRefs, scopedIDs []string, removedBindingScope map[string]bool) []string {
	if len(scopedRefs) == 0 && len(scopedIDs) == 0 {
		return nil
	}

	scope := map[string]bool{}
	for object, binding := range bindings {
		if len(selectedSharedRefsForObject(binding.SharedRefs, scopedRefs, scopedIDs, sharedFilesByRef)) > 0 {
			scope[object] = true
		}
	}
	for object := range removedBindingScope {
		scope[object] = true
	}
	return sortedKeys(scope)
}

func candidateObjectsWithRemovedSelectedBinding(repoRoot string, bindings map[string]objectBinding, scopedRefs, scopedIDs []string) (map[string]bool, error) {
	result := map[string]bool{}
	for object, binding := range bindings {
		if binding.Status.ActiveLayer != "candidate" {
			continue
		}
		if len(selectedSharedRefsForObject(binding.SharedRefs, scopedRefs, scopedIDs, nil)) > 0 {
			continue
		}
		matched, err := processSnapshotContainsSelectedShared(
			repoRoot,
			binding.Status.ObjectType,
			binding.Status.Object,
			binding.Status.ActiveLayer,
			[]string{"check", "verify"},
			scopedRefs,
			scopedIDs,
		)
		if err != nil {
			return nil, err
		}
		if matched {
			result[object] = true
		}
	}
	return result, nil
}

func processSnapshotContainsSelectedShared(repoRoot, objectType, object, activeLayer string, processKinds, scopedRefs, scopedIDs []string) (bool, error) {
	refSet := makeStringSet(scopedRefs)
	idSet := makeStringSet(scopedIDs)
	for _, processKind := range processKinds {
		processPath, err := snapshot.ProcessFilePath(object, processKind)
		if err != nil {
			return false, err
		}
		processAbs := filepath.Join(repoRoot, filepath.FromSlash(processPath))
		if _, err := os.Stat(processAbs); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return false, fmt.Errorf("stat %s: %w", processPath, err)
		}
		processSnapshot, err := snapshot.LoadProcessSnapshot(repoRoot, object, processKind)
		if err != nil {
			return false, err
		}
		if objectType != "module" {
			rawContent, err := os.ReadFile(processAbs)
			if err != nil {
				return false, fmt.Errorf("read %s: %w", processPath, err)
			}
			if !isValidRemovedBindingEvidence(string(rawContent), processSnapshot, objectType, object, activeLayer, processKind) {
				continue
			}
		}
		for _, entry := range processSnapshot.SharedContractSnapshot {
			if refSet[entry.VersionRef] || idSet[entry.SharedContractID] {
				return true, nil
			}
		}
	}
	return false, nil
}

func isValidRemovedBindingEvidence(content string, processSnapshot snapshot.ProcessSnapshotData, objectType, object, activeLayer, processKind string) bool {
	requiredScalars := []string{
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
		"system_constraints_stable_file_ref",
		"system_constraints_stable_version_ref",
		"system_constraints_stable_fingerprint",
	}
	if processKind == "verify" {
		requiredScalars = append(requiredScalars, "verification_scope_ref")
	}
	for _, field := range requiredScalars {
		if strings.TrimSpace(processSnapshot.Scalars[field]) == "" {
			return false
		}
		if !hasSnapshotField(content, field) {
			return false
		}
	}

	requiredListFields := []string{"shared_contract_snapshot", "module_snapshot"}
	if objectType == "project" {
		requiredListFields = append(requiredListFields, "flow_snapshot")
	}
	for _, field := range requiredListFields {
		if !hasSnapshotField(content, field) {
			return false
		}
	}
	if !allSharedSnapshotEntriesComplete(processSnapshot.SharedContractSnapshot) {
		return false
	}

	expectedGate, expectedNextCommand, ok := expectedObjectProcessRouting(objectType, processKind)
	if !ok {
		return false
	}
	if processSnapshot.Scalars["object_type"] != objectType {
		return false
	}
	if processSnapshot.Scalars["object_ref"] != object {
		return false
	}
	if processSnapshot.Scalars["truth_layer_ref"] != activeLayer {
		return false
	}
	expectedTruthFileRef, err := specpaths.ObjectMainSpecFileRef(objectType, activeLayer, object)
	if err != nil {
		return false
	}
	if processSnapshot.Scalars["truth_file_ref"] != expectedTruthFileRef {
		return false
	}
	if processSnapshot.Scalars["gate"] != expectedGate {
		return false
	}
	if processSnapshot.Scalars["decision"] != "pass" {
		return false
	}
	if processSnapshot.Scalars["allow_next"] != "true" {
		return false
	}
	if processSnapshot.Scalars["next_command"] != expectedNextCommand {
		return false
	}
	if processSnapshot.Scalars["system_constraints_stable_file_ref"] != specpaths.SystemConstraintsStableFileRef {
		return false
	}
	return true
}

func hasSnapshotField(content, field string) bool {
	return strings.Contains(strings.ReplaceAll(content, "\r\n", "\n"), "\n"+field+":") ||
		strings.HasPrefix(strings.ReplaceAll(content, "\r\n", "\n"), field+":")
}

func allSharedSnapshotEntriesComplete(entries []snapshot.SharedContractEntry) bool {
	if len(entries) == 0 {
		return true
	}
	for _, entry := range entries {
		if strings.TrimSpace(entry.SharedContractID) == "" ||
			strings.TrimSpace(entry.Layer) == "" ||
			strings.TrimSpace(entry.FileRef) == "" ||
			strings.TrimSpace(entry.VersionRef) == "" ||
			strings.TrimSpace(entry.Fingerprint) == "" {
			return false
		}
	}
	return true
}

func expectedObjectProcessRouting(objectType, processKind string) (string, string, bool) {
	switch objectType {
	case "flow":
		switch processKind {
		case "check":
			return "flow_check", "flow_verify", true
		case "verify":
			return "flow_verify", "flow_promote", true
		}
	case "project":
		switch processKind {
		case "check":
			return "project_check", "project_verify", true
		case "verify":
			return "project_verify", "project_promote", true
		}
	}
	return "", "", false
}
