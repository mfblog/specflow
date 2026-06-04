package rulesync

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
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

type objectBinding struct {
	Status        statusfile.ObjectStatus
	RuleRefs      []string
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

		refs, err := readObjectRuleRefs(repoRoot, status)
		if err != nil {
			return nil, nil, err
		}

		bindingIssues := []string{}
		for _, ref := range refs {
			if _, err := rulebinding.ResolveRef(repoRoot, status.ActiveLayer, ref); err != nil {
				bindingIssues = append(bindingIssues, err.Error())
				unresolvedRefs = append(unresolvedRefs, ref)
			}
		}

		bindings[status.Object] = objectBinding{
			Status:        status,
			RuleRefs:      refs,
			BindingIssues: normalizeStrings(bindingIssues),
		}
	}

	return bindings, normalizeStrings(unresolvedRefs), nil
}

func readObjectRuleRefs(repoRoot string, status statusfile.ObjectStatus) ([]string, error) {
	mainSpecRef, err := specpaths.ObjectMainSpecFileRef(status.ObjectType, status.ActiveLayer, status.Object)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", mainSpecRef, err)
	}
	refs, err := rulerefs.ParseObjectRuleRefs(mainSpecRef, string(content))
	if err != nil {
		return nil, err
	}
	return normalizeStrings(refs), nil
}

func buildScopeObjects(bindings map[string]objectBinding, sharedFilesByRef map[string]sharedFile, sharedFilesByID map[string][]sharedFile, scopedRefs, scopedIDs []string, removedBindingScope map[string]bool) ([]string, error) {
	if len(scopedRefs) == 0 && len(scopedIDs) == 0 {
		return nil, nil
	}

	scope := map[string]bool{}
	for object, binding := range bindings {
		selectedRefs, err := selectedRuleRefsForObject(binding.RuleRefs, scopedRefs, scopedIDs, sharedFilesByRef, sharedFilesByID)
		if err != nil {
			return nil, err
		}
		if len(selectedRefs) > 0 {
			scope[object] = true
		}
	}
	for object := range removedBindingScope {
		scope[object] = true
	}
	return sortedKeys(scope), nil
}

func candidateObjectsWithRemovedSelectedBinding(repoRoot string, bindings map[string]objectBinding, scopedRefs, scopedIDs []string, sharedFilesByRef map[string]sharedFile, sharedFilesByID map[string][]sharedFile) (map[string]bool, error) {
	result := map[string]bool{}
	for object, binding := range bindings {
		if binding.Status.ActiveLayer != "candidate" {
			continue
		}
		selectedRefs, err := selectedRuleRefsForObject(binding.RuleRefs, scopedRefs, scopedIDs, sharedFilesByRef, sharedFilesByID)
		if err != nil {
			return nil, err
		}
		if len(selectedRefs) > 0 {
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
			sharedFilesByID,
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

func processSnapshotContainsSelectedShared(repoRoot, objectType, object, activeLayer string, processKinds, scopedRefs, scopedIDs []string, sharedFilesByID map[string][]sharedFile) (bool, error) {
	refSet := makeStringSet(scopedRefs)
	idSet := makeStringSet(scopedIDs)
	for _, processKind := range processKinds {
		processPath, err := snapshot.ProcessFilePath(objectType, object, processKind)
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
		processSnapshot, err := snapshot.LoadProcessSnapshot(repoRoot, objectType, object, processKind)
		if err != nil {
			return false, err
		}
		if objectType != "unit" {
			return false, fmt.Errorf("unsupported object type %q", objectType)
		}
		validEvidence, err := isValidModuleRemovedBindingEvidence(repoRoot, object, activeLayer, processKind, processSnapshot, scopedRefs, scopedIDs, sharedFilesByID)
		if err != nil {
			return false, err
		}
		if !validEvidence {
			continue
		}
		for _, entry := range processSnapshot.RuleSnapshot {
			matched, err := matchesSelectedSharedEntry(entry, refSet, idSet, sharedFilesByID)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
	}
	return false, nil
}

func isValidModuleRemovedBindingEvidence(repoRoot, module, activeLayer, processKind string, processSnapshot snapshot.ProcessSnapshotData, scopedRefs, scopedIDs []string, sharedFilesByID map[string][]sharedFile) (bool, error) {
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
	}
	if processKind == "verify" {
	}
	for _, field := range requiredScalars {
		if !processSnapshot.PresentFields[field] {
			return false, nil
		}
		if strings.TrimSpace(processSnapshot.Scalars[field]) == "" {
			return false, nil
		}
	}
	requiredListFields := []string{"unit_appendix_snapshot", "unit_snapshot", "rule_snapshot", "acceptance_item_set"}
	if processKind == "verify" {
		requiredListFields = append(requiredListFields, "acceptance_item_evidence_matrix")
	}
	for _, field := range requiredListFields {
		if !processSnapshot.PresentFields[field] {
			return false, nil
		}
	}
	if !allSharedSnapshotEntriesComplete(processSnapshot.RuleSnapshot) {
		return false, nil
	}

	expectedGate, expectedNextCommand, ok := expectedModuleProcessRouting(processKind)
	if !ok {
		return false, nil
	}
	currentSnapshot, err := snapshot.RebuildCurrent(repoRoot, module)
	if err != nil {
		return false, err
	}
	currentTruthContent, err := readCurrentObjectTruthContent(repoRoot, "unit", module, activeLayer)
	if err != nil {
		return false, err
	}
	if processSnapshot.Scalars["object_type"] != "unit" {
		return false, nil
	}
	if processSnapshot.Scalars["object_ref"] != module {
		return false, nil
	}
	if processSnapshot.Scalars["gate"] != expectedGate {
		return false, nil
	}
	if processSnapshot.Scalars["decision"] != "pass" {
		return false, nil
	}
	if processSnapshot.Scalars["allow_next"] != "true" {
		return false, nil
	}
	if processSnapshot.Scalars["next_command"] != expectedNextCommand {
		return false, nil
	}
	if processSnapshot.Scalars["truth_layer_ref"] != activeLayer {
		return false, nil
	}
	truthMatches, err := matchesRemovedBindingTruth(processSnapshot, currentSnapshot.SpecFileRef, currentTruthContent, processSnapshot.RuleSnapshot)
	if err != nil {
		return false, err
	}
	if !truthMatches {
		return false, nil
	}
	if !equalAppendixEntries(processSnapshot.ModuleAppendixSnapshot, currentSnapshot.ModuleAppendixSnapshot) {
		return false, nil
	}
	if !equalObjectSnapshotEntries(processSnapshot.UnitSnapshot, currentSnapshot.UnitSnapshot) {
		return false, nil
	}
	if !equalAcceptanceItemEntries(processSnapshot.AcceptanceItemSet, currentSnapshot.AcceptanceItemSet) {
		return false, nil
	}
	if processKind == "verify" && !acceptanceEvidenceMatrixCovers(processSnapshot.AcceptanceEvidence, currentSnapshot.AcceptanceItemSet) {
		return false, nil
	}

	return sharedSnapshotMatchesRemovedBindingEvidence(
		processSnapshot.RuleSnapshot,
		currentSnapshot.RuleSnapshot,
		scopedRefs,
		scopedIDs,
		sharedFilesByID,
	)
}

func allSharedSnapshotEntriesComplete(entries []snapshot.RuleEntry) bool {
	if len(entries) == 0 {
		return true
	}
	for _, entry := range entries {
		if strings.TrimSpace(entry.RuleID) == "" ||
			strings.TrimSpace(entry.Layer) == "" ||
			strings.TrimSpace(entry.FileRef) == "" ||
			strings.TrimSpace(entry.VersionRef) == "" ||
			strings.TrimSpace(entry.Fingerprint) == "" {
			return false
		}
	}
	return true
}

func equalObjectSnapshotEntries(actual, expected []snapshot.ObjectSnapshotEntry) bool {
	actual = normalizeObjectSnapshotEntries(actual)
	expected = normalizeObjectSnapshotEntries(expected)
	if len(actual) != len(expected) {
		return false
	}
	for idx := range actual {
		if actual[idx] != expected[idx] {
			return false
		}
	}
	return true
}

func equalAppendixEntries(actual, expected []snapshot.AppendixEntry) bool {
	actual = normalizeAppendixEntries(actual)
	expected = normalizeAppendixEntries(expected)
	if len(actual) != len(expected) {
		return false
	}
	for idx := range actual {
		if actual[idx] != expected[idx] {
			return false
		}
	}
	return true
}

func equalSharedSnapshotEntries(actual, expected []snapshot.RuleEntry) bool {
	actual = normalizeSharedSnapshotEntries(actual)
	expected = normalizeSharedSnapshotEntries(expected)
	if len(actual) != len(expected) {
		return false
	}
	for idx := range actual {
		if actual[idx] != expected[idx] {
			return false
		}
	}
	return true
}

func equalAcceptanceItemEntries(actual, expected []snapshot.AcceptanceItemEntry) bool {
	actual = normalizeAcceptanceItemEntries(actual)
	expected = normalizeAcceptanceItemEntries(expected)
	if len(actual) != len(expected) {
		return false
	}
	for idx := range actual {
		if actual[idx] != expected[idx] {
			return false
		}
	}
	return true
}

func acceptanceEvidenceMatrixCovers(actual []snapshot.AcceptanceEvidenceEntry, expected []snapshot.AcceptanceItemEntry) bool {
	actual = normalizeAcceptanceEvidenceEntries(actual)
	expected = normalizeAcceptanceItemEntries(expected)
	if len(actual) != len(expected) {
		return false
	}
	expectedByID := map[string]snapshot.AcceptanceItemEntry{}
	for _, item := range expected {
		expectedByID[item.ID] = item
	}
	seen := map[string]bool{}
	for _, entry := range actual {
		if seen[entry.ID] {
			return false
		}
		seen[entry.ID] = true
		item, ok := expectedByID[entry.ID]
		if !ok {
			return false
		}
		if !allowedAcceptanceEvidenceStatus(entry.Status) {
			return false
		}
		if item.NotRunnableYet == "yes" && entry.Status != "not_runnable_yet" {
			return false
		}
		if item.NotRunnableYet == "no" && entry.Status == "not_runnable_yet" {
			return false
		}
	}
	return true
}

func normalizeAppendixEntries(entries []snapshot.AppendixEntry) []snapshot.AppendixEntry {
	if len(entries) == 0 {
		return nil
	}
	items := append([]snapshot.AppendixEntry(nil), entries...)
	sort.Slice(items, func(i, j int) bool {
		return items[i].FileRef < items[j].FileRef
	})
	return items
}

func normalizeAcceptanceItemEntries(entries []snapshot.AcceptanceItemEntry) []snapshot.AcceptanceItemEntry {
	if len(entries) == 0 {
		return nil
	}
	items := append([]snapshot.AcceptanceItemEntry(nil), entries...)
	sort.Slice(items, func(i, j int) bool {
		if items[i].ID != items[j].ID {
			return items[i].ID < items[j].ID
		}
		return items[i].VerificationSurface < items[j].VerificationSurface
	})
	for idx := range items {
		items[idx] = snapshot.AcceptanceItemEntry{
			ID:                  items[idx].ID,
			VerificationSurface: items[idx].VerificationSurface,
			NotRunnableYet:      items[idx].NotRunnableYet,
		}
	}
	return items
}

func normalizeAcceptanceEvidenceEntries(entries []snapshot.AcceptanceEvidenceEntry) []snapshot.AcceptanceEvidenceEntry {
	if len(entries) == 0 {
		return nil
	}
	items := append([]snapshot.AcceptanceEvidenceEntry(nil), entries...)
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

func allowedAcceptanceEvidenceStatus(status string) bool {
	switch status {
	case "pass", "fail", "partial", "not_checked", "not_runnable_yet":
		return true
	default:
		return false
	}
}

func normalizeObjectSnapshotEntries(entries []snapshot.ObjectSnapshotEntry) []snapshot.ObjectSnapshotEntry {
	if len(entries) == 0 {
		return nil
	}
	items := append([]snapshot.ObjectSnapshotEntry(nil), entries...)
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

func normalizeSharedSnapshotEntries(entries []snapshot.RuleEntry) []snapshot.RuleEntry {
	if len(entries) == 0 {
		return nil
	}
	items := append([]snapshot.RuleEntry(nil), entries...)
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

type currentObjectSnapshot struct {
	TruthFileRef      string
	TruthVersionRef   string
	TruthFingerprint  string
	AppendixSnapshot  []snapshot.AppendixEntry
	UnitSnapshot      []snapshot.ObjectSnapshotEntry
	RuleSnapshot      []snapshot.RuleEntry
	AcceptanceItemSet []snapshot.AcceptanceItemEntry
}

func rebuildCurrentObjectSnapshot(repoRoot, objectType, object, activeLayer string) (currentObjectSnapshot, error) {
	mainSpecRef, err := specpaths.ObjectMainSpecFileRef(objectType, activeLayer, object)
	if err != nil {
		return currentObjectSnapshot{}, err
	}
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)))
	if err != nil {
		return currentObjectSnapshot{}, fmt.Errorf("read %s: %w", mainSpecRef, err)
	}
	frontmatter, body, err := parseFrontmatter(string(content))
	if err != nil {
		return currentObjectSnapshot{}, fmt.Errorf("%s: %w", mainSpecRef, err)
	}
	version := strings.TrimSpace(frontmatter["version"])
	if version == "" {
		return currentObjectSnapshot{}, fmt.Errorf("%s: missing frontmatter.version", mainSpecRef)
	}
	result := currentObjectSnapshot{
		TruthFileRef:     mainSpecRef,
		TruthVersionRef:  fmt.Sprintf("%s@%s", strings.TrimSuffix(filepath.Base(mainSpecRef), ".md"), version),
		TruthFingerprint: hashNormalizedText(string(content)),
	}
	result.AcceptanceItemSet, err = snapshot.BuildAcceptanceItemSetFromBody(mainSpecRef, body)
	if err != nil {
		return currentObjectSnapshot{}, err
	}
	result.AppendixSnapshot, err = snapshot.BuildAppendixSnapshot(repoRoot, mainSpecRef, body)
	if err != nil {
		return currentObjectSnapshot{}, err
	}

	moduleRefs, hasField, err := parseNamedRefList(string(content), "unit_refs")
	if err != nil {
		return currentObjectSnapshot{}, err
	}
	if hasField {
		result.UnitSnapshot, err = buildObjectDependencySnapshot(repoRoot, "unit", moduleRefs)
		if err != nil {
			return currentObjectSnapshot{}, err
		}
	}
	sharedRefs, err := rulerefs.ParseObjectRuleRefs(mainSpecRef, string(content))
	if err != nil {
		return currentObjectSnapshot{}, err
	}
	result.RuleSnapshot, err = buildRuleSnapshot(repoRoot, activeLayer, sharedRefs)
	if err != nil {
		return currentObjectSnapshot{}, err
	}

	return result, nil
}

func buildRuleSnapshot(repoRoot, activeLayer string, refs []string) ([]snapshot.RuleEntry, error) {
	entries, err := buildStableGlobalRuleEntries(repoRoot)
	if err != nil {
		return nil, err
	}
	for _, ref := range refs {
		resolved, err := rulebinding.ResolveRef(repoRoot, activeLayer, ref)
		if err != nil {
			return nil, err
		}
		entries = append(entries, snapshot.RuleEntry{
			RuleID:      resolved.RuleID,
			Layer:       resolved.Layer,
			FileRef:     resolved.FileRef,
			VersionRef:  resolved.VersionRef,
			Fingerprint: hashNormalizedText(resolved.Content),
		})
	}
	return normalizeSharedSnapshotEntries(entries), nil
}

func buildStableGlobalRuleEntries(repoRoot string) ([]snapshot.RuleEntry, error) {
	matches, err := filepath.Glob(filepath.Join(repoRoot, filepath.FromSlash("docs/specs/rules/stable/s_g_rule_*.md")))
	if err != nil {
		return nil, err
	}
	entries := []snapshot.RuleEntry{}
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
		entries = append(entries, snapshot.RuleEntry{
			RuleID:      ruleID,
			Layer:       layer,
			FileRef:     fileRef,
			VersionRef:  fmt.Sprintf("%s@%s", prefix, version),
			Fingerprint: hashNormalizedText(string(content)),
		})
	}
	return entries, nil
}

func buildObjectDependencySnapshot(repoRoot, objectType string, refs []string) ([]snapshot.ObjectSnapshotEntry, error) {
	entries := make([]snapshot.ObjectSnapshotEntry, 0, len(refs))
	for _, ref := range refs {
		entry, err := resolveObjectVersionRef(repoRoot, objectType, ref)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return normalizeObjectSnapshotEntries(entries), nil
}

func resolveObjectVersionRef(repoRoot, expectedObjectType, ref string) (snapshot.ObjectSnapshotEntry, error) {
	prefix, _, found := strings.Cut(strings.TrimSpace(ref), "@")
	if !found {
		return snapshot.ObjectSnapshotEntry{}, fmt.Errorf("invalid %s ref %q", expectedObjectType, ref)
	}
	objectType, layer, object, err := parseObjectVersionRefPrefix(prefix)
	if err != nil {
		return snapshot.ObjectSnapshotEntry{}, err
	}
	if objectType != expectedObjectType {
		return snapshot.ObjectSnapshotEntry{}, fmt.Errorf("%s ref %q resolves to object type %q", expectedObjectType, ref, objectType)
	}
	if expectedObjectType == "unit" && layer != "stable" {
		return snapshot.ObjectSnapshotEntry{}, fmt.Errorf("unit_refs must reference stable units; got %q", ref)
	}
	fileRef, err := specpaths.ObjectMainSpecFileRef(objectType, layer, object)
	if err != nil {
		return snapshot.ObjectSnapshotEntry{}, err
	}
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(fileRef)))
	if err != nil {
		return snapshot.ObjectSnapshotEntry{}, fmt.Errorf("read %s: %w", fileRef, err)
	}
	frontmatter, _, err := parseFrontmatter(string(content))
	if err != nil {
		return snapshot.ObjectSnapshotEntry{}, fmt.Errorf("%s: %w", fileRef, err)
	}
	version := strings.TrimSpace(frontmatter["version"])
	if version == "" {
		return snapshot.ObjectSnapshotEntry{}, fmt.Errorf("%s: missing frontmatter.version", fileRef)
	}
	expectedVersionRef := fmt.Sprintf("%s@%s", prefix, version)
	if expectedVersionRef != ref {
		return snapshot.ObjectSnapshotEntry{}, fmt.Errorf("%s ref %q does not match current version %q", expectedObjectType, ref, expectedVersionRef)
	}
	return snapshot.ObjectSnapshotEntry{
		ObjectRef:   object,
		Layer:       layer,
		FileRef:     fileRef,
		VersionRef:  expectedVersionRef,
		Fingerprint: hashNormalizedText(string(content)),
	}, nil
}

func parseObjectVersionRefPrefix(prefix string) (string, string, string, error) {
	switch {
	case strings.HasPrefix(prefix, "c_unit_"):
		return "unit", "candidate", strings.TrimPrefix(prefix, "c_unit_"), nil
	case strings.HasPrefix(prefix, "s_unit_"):
		return "unit", "stable", strings.TrimPrefix(prefix, "s_unit_"), nil
	default:
		return "", "", "", fmt.Errorf("unsupported object version ref prefix %q", prefix)
	}
}

func parseNamedRefList(body, fieldName string) ([]string, bool, error) {
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		right, matched, err := parseObjectNamedFieldLine(trimmed, fieldName)
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
			if nextTrimmed == "---" {
				break
			}
			if strings.HasPrefix(nextTrimmed, "## ") || regexp.MustCompile(`^\d+\.`).MatchString(nextTrimmed) {
				break
			}
			if strings.Contains(nextTrimmed, ":") && !strings.HasPrefix(nextTrimmed, "- ") {
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
		return refs, true, nil
	}
	return nil, false, nil
}

func parseObjectNamedFieldLine(trimmed, fieldName string) (string, bool, error) {
	parts := strings.SplitN(trimmed, ":", 2)
	if len(parts) != 2 {
		return "", false, nil
	}
	left := normalizeObjectFieldKey(strings.TrimSpace(parts[0]))
	if left != fieldName {
		return "", false, nil
	}
	return strings.TrimSpace(parts[1]), true, nil
}

func normalizeObjectFieldKey(value string) string {
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

func hashNormalizedText(content string) string {
	text := strings.ReplaceAll(content, "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")
	text += "\n"
	sum := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", sum)
}

func sharedSnapshotMatchesRemovedBindingEvidence(stored, current []snapshot.RuleEntry, scopedRefs []string, scopedIDs []string, sharedFilesByID map[string][]sharedFile) (bool, error) {
	stored = normalizeSharedSnapshotEntries(stored)
	current = normalizeSharedSnapshotEntries(current)
	if equalSharedSnapshotEntries(stored, current) {
		return true, nil
	}
	refSet := makeStringSet(scopedRefs)
	idSet := makeStringSet(scopedIDs)

	currentSet := map[string]bool{}
	for _, entry := range current {
		currentSet[sharedSnapshotEntryKey(entry)] = true
	}
	storedSet := map[string]bool{}
	for _, entry := range stored {
		storedSet[sharedSnapshotEntryKey(entry)] = true
	}

	for _, entry := range current {
		if !storedSet[sharedSnapshotEntryKey(entry)] {
			return false, nil
		}
	}
	for _, entry := range stored {
		if currentSet[sharedSnapshotEntryKey(entry)] {
			continue
		}
		matched, err := matchesSelectedSharedEntry(entry, refSet, idSet, sharedFilesByID)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}
	return true, nil
}

func matchesSelectedSharedEntry(entry snapshot.RuleEntry, scopedRefSet, scopedIDSet map[string]bool, sharedFilesByID map[string][]sharedFile) (bool, error) {
	if scopedRefSet[entry.VersionRef] {
		return true, nil
	}
	return matchesSelectedSharedIDEntry(entry, scopedIDSet, sharedFilesByID)
}

func matchesSelectedSharedIDEntry(entry snapshot.RuleEntry, scopedIDSet map[string]bool, sharedFilesByID map[string][]sharedFile) (bool, error) {
	if !scopedIDSet[entry.RuleID] {
		return false, nil
	}
	candidates := sharedFilesByID[entry.RuleID]
	if len(candidates) > 1 {
		return false, fmt.Errorf("shared_id %q resolves to multiple current rule files; removed-binding scope is ambiguous", entry.RuleID)
	}
	if len(candidates) != 1 {
		return false, nil
	}
	current := candidates[0]
	if current.FileRef == entry.FileRef && current.Layer == entry.Layer {
		return true, nil
	}
	return false, nil
}

func sharedSnapshotEntryKey(entry snapshot.RuleEntry) string {
	return strings.Join([]string{
		entry.RuleID,
		entry.Layer,
		entry.FileRef,
		entry.VersionRef,
		entry.Fingerprint,
	}, "\x00")
}

func hasOnlySharedSnapshotMismatch(mismatches []string) bool {
	if len(mismatches) == 0 {
		return false
	}
	for _, mismatch := range mismatches {
		if !strings.HasPrefix(mismatch, "rule_snapshot mismatch") {
			return false
		}
	}
	return true
}

func readCurrentObjectTruthContent(repoRoot, objectType, object, activeLayer string) (string, error) {
	mainSpecRef, err := specpaths.ObjectMainSpecFileRef(objectType, activeLayer, object)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)))
	if err != nil {
		return "", fmt.Errorf("read %s: %w", mainSpecRef, err)
	}
	return string(content), nil
}

func matchesRemovedBindingTruth(processSnapshot snapshot.ProcessSnapshotData, currentTruthFileRef, currentTruthContent string, storedShared []snapshot.RuleEntry) (bool, error) {
	if processSnapshot.Scalars["truth_file_ref"] != currentTruthFileRef {
		return false, nil
	}
	reconstructedContents, err := reconstructTruthWithStoredSharedSnapshot(currentTruthContent, storedShared)
	if err != nil {
		return false, err
	}
	for _, reconstructedContent := range reconstructedContents {
		frontmatter, _, err := parseFrontmatter(reconstructedContent)
		if err != nil {
			return false, err
		}
		version := strings.TrimSpace(frontmatter["version"])
		if version == "" {
			return false, fmt.Errorf("%s: missing frontmatter.version", currentTruthFileRef)
		}
		reconstructedVersionRef := fmt.Sprintf("%s@%s", strings.TrimSuffix(filepath.Base(currentTruthFileRef), ".md"), version)
		if processSnapshot.Scalars["truth_version_ref"] != reconstructedVersionRef {
			continue
		}
		if processSnapshot.Scalars["truth_fingerprint"] == hashNormalizedText(reconstructedContent) {
			return true, nil
		}
	}
	return false, nil
}

func reconstructTruthWithStoredSharedSnapshot(currentTruthContent string, storedShared []snapshot.RuleEntry) ([]string, error) {
	refs := make([]string, 0, len(storedShared))
	for _, entry := range storedShared {
		refs = append(refs, strings.TrimSpace(entry.VersionRef))
	}
	sort.Strings(refs)
	rewritten, err := rulerefs.UpdateObjectRuleRefs("stored-truth", currentTruthContent, refs)
	if err != nil {
		return nil, err
	}
	return []string{rewritten}, nil
}

type sharedRefRenderStyle struct {
	wrapWithBackticks bool
}

func rewriteRuleRefsInBody(body string, refs []string, style sharedRefRenderStyle) (string, error) {
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		_, matched, err := parseObjectNamedFieldLine(trimmed, "rule_refs")
		if err != nil {
			return "", err
		}
		if !matched {
			continue
		}
		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			return "", fmt.Errorf("rule_refs line missing colon")
		}
		left := line[:colonIdx]
		end := idx + 1
		for end < len(lines) {
			nextTrimmed := strings.TrimSpace(lines[end])
			if nextTrimmed == "" {
				break
			}
			if strings.HasPrefix(nextTrimmed, "## ") || regexp.MustCompile(`^\d+\.`).MatchString(nextTrimmed) {
				break
			}
			if !strings.HasPrefix(nextTrimmed, "- ") {
				return "", fmt.Errorf("rule_refs must be a markdown list of rule refs")
			}
			end++
		}

		replacement := []string{}
		if len(refs) == 0 {
			replacement = append(replacement, left+": none")
		} else {
			replacement = append(replacement, left+":")
			indent := detectRuleRefIndent(line, lines[idx+1:end])
			for _, ref := range refs {
				renderedRef := ref
				if style.wrapWithBackticks {
					renderedRef = "`" + ref + "`"
				}
				replacement = append(replacement, indent+"- "+renderedRef)
			}
		}
		updated := append([]string{}, lines[:idx]...)
		updated = append(updated, replacement...)
		updated = append(updated, lines[end:]...)
		return strings.Join(updated, "\n"), nil
	}
	return "", fmt.Errorf("rule_refs field not found")
}

func detectRuleRefIndent(fieldLine string, existingListLines []string) string {
	for _, line := range existingListLines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			if idx := strings.Index(line, "- "); idx >= 0 {
				return line[:idx]
			}
		}
	}
	leading := fieldLine[:len(fieldLine)-len(strings.TrimLeft(fieldLine, " \t"))]
	if regexp.MustCompile(`^\s*\d+\.`).MatchString(fieldLine) {
		return leading + "   "
	}
	return leading + "  "
}

func expectedModuleProcessRouting(processKind string) (string, string, bool) {
	switch processKind {
	case "check":
		return "unit_check", "unit_plan", true
	case "plan":
		return "unit_plan", "unit_impl", true
	case "verify":
		return "unit_verify", "unit_promote", true
	default:
		return "", "", false
	}
}
