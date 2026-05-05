package rulesync

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/impactsync"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/rulebinding"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

type Options struct {
	Modules                      []string
	RuleRefs                     []string
	RuleIDs                      []string
	StableLandingModule          string
	StableLandingRuleRefs        []string
	BoundObjectsOnlyRuleFileRefs []string
}

type Result struct {
	ScopedModules                []string
	ScopedFlows                  []string
	ScopedRuleRefs               []string
	ScopedRuleIDs                []string
	StableLandingModule          string
	StableLandingRuleRefs        []string
	BoundObjectsOnlyRuleFileRefs []string
	ModuleResults                []ModuleResult
	FlowResults                  []ObjectResult
	BoundObjectDrifts            []BoundObjectDrift
}

type ModuleResult = impactsync.ModuleResult
type ObjectResult = impactsync.ObjectResult

type BoundObjectDrift struct {
	RuleID                string
	FileRef               string
	VersionRef            string
	DeclaredObjects       []string
	ActualObjects         []string
	BoundObjectsOnlyDelta bool
}

type moduleBinding struct {
	Status        statusfile.ModuleStatus
	RuleRefs      []string
	BindingIssues []string
}

type sharedFile struct {
	RuleID             string
	Layer              string
	FileRef            string
	VersionRef         string
	BoundObjects       []string
	PromotionOwnerUnit string
}

type ReconcileBoundModulesOptions struct {
	Modules  []string
	RuleRefs []string
	RuleIDs  []string
}

type ReconcileBoundModulesResult struct {
	ScopedModules  []string
	ScopedRuleRefs []string
	ScopedRuleIDs  []string
	TouchedFiles   []string
	UpdatedFiles   []string
	UnchangedFiles []string
}

func SyncImpact(repoRoot string, options Options) (Result, error) {
	normalized := Options{
		Modules:                      normalizeStrings(options.Modules),
		RuleRefs:                     normalizeStrings(options.RuleRefs),
		RuleIDs:                      normalizeStrings(options.RuleIDs),
		StableLandingModule:          strings.TrimSpace(options.StableLandingModule),
		StableLandingRuleRefs:        normalizeStrings(options.StableLandingRuleRefs),
		BoundObjectsOnlyRuleFileRefs: normalizeStrings(options.BoundObjectsOnlyRuleFileRefs),
	}
	if len(normalized.RuleRefs) == 0 && len(normalized.RuleIDs) == 0 {
		return Result{}, fmt.Errorf("at least one of rule refs or rule ids is required")
	}

	sharedFilesByRef, err := loadSharedFiles(repoRoot)
	if err != nil {
		return Result{}, err
	}
	sharedFilesByID := buildSharedFilesByID(sharedFilesByRef)
	sharedFilesByFileRef := buildSharedFilesByFileRef(sharedFilesByRef)

	moduleBindings, actualModulesByRef, _, unresolvedRuleRefs, err := loadModuleBindings(repoRoot)
	if err != nil {
		return Result{}, err
	}
	flowBindings, unresolvedFlowRefs, err := loadObjectBindings(repoRoot, "scenario")
	if err != nil {
		return Result{}, err
	}
	allUnresolvedRefs := normalizeStrings(append(unresolvedRuleRefs, unresolvedFlowRefs...))
	referencedRuleRefs := allReferencedRuleRefs(moduleBindings, flowBindings)
	for _, sharedID := range normalized.RuleIDs {
		if len(allUnresolvedRefs) > 0 {
			return Result{}, fmt.Errorf(
				"cannot determine affected downstream objects safely for rule id %q because unresolved rule refs remain in downstream bindings: %s",
				sharedID,
				strings.Join(allUnresolvedRefs, ", "),
			)
		}
		if _, ok := sharedFilesByID[sharedID]; !ok {
			return Result{}, fmt.Errorf("rule id %q is not present under docs/specs/rules/", sharedID)
		}
	}
	for _, ref := range normalized.RuleRefs {
		if _, ok := sharedFilesByRef[ref]; ok {
			continue
		}
		if referencedRuleRefs[ref] {
			continue
		}
		return Result{}, fmt.Errorf("rule ref %q is not present under docs/specs/rules/ and is not referenced by current downstream bindings", ref)
	}
	for _, module := range normalized.Modules {
		if _, ok := moduleBindings[module]; !ok {
			return Result{}, fmt.Errorf("module %q is not registered in docs/specs/_status.md", module)
		}
	}
	if normalized.StableLandingModule != "" {
		binding, ok := moduleBindings[normalized.StableLandingModule]
		if !ok {
			return Result{}, fmt.Errorf("stable landing unit %q is not registered in docs/specs/_status.md", normalized.StableLandingModule)
		}
		if len(normalized.StableLandingRuleRefs) == 0 {
			return Result{}, fmt.Errorf("stable landing rule refs are required when stable landing unit %q is set", normalized.StableLandingModule)
		}
		if binding.Status.ActiveLayer != "stable" {
			return Result{}, fmt.Errorf("stable landing unit %q must currently be at active layer stable", normalized.StableLandingModule)
		}
		landingSelectedRefs, err := selectedRuleRefsForObject(binding.RuleRefs, normalized.RuleRefs, normalized.RuleIDs, sharedFilesByRef, sharedFilesByID)
		if err != nil {
			return Result{}, err
		}
		landingSelectedRefSet := makeStringSet(landingSelectedRefs)
		for _, ref := range normalized.StableLandingRuleRefs {
			if _, ok := sharedFilesByRef[ref]; !ok {
				return Result{}, fmt.Errorf("stable landing rule ref %q is not present under docs/specs/rules/", ref)
			}
			if !landingSelectedRefSet[ref] {
				return Result{}, fmt.Errorf("stable landing rule ref %q is not selected for stable landing unit %q in this shared sync scope", ref, normalized.StableLandingModule)
			}
		}
	} else if len(normalized.StableLandingRuleRefs) > 0 {
		return Result{}, fmt.Errorf("stable landing rule refs require stable landing unit")
	}

	boundObjectsOnlyFileRefs := map[string]bool{}
	for _, fileRef := range normalized.BoundObjectsOnlyRuleFileRefs {
		shared, ok := sharedFilesByFileRef[fileRef]
		if !ok {
			return Result{}, fmt.Errorf("bound_objects-only rule file ref %q is not present under docs/specs/rules/", fileRef)
		}
		boundObjectsOnlyFileRefs[shared.FileRef] = true
	}

	actualBoundObjectsByRef, err := collectActualBoundObjectsByRef(repoRoot, actualModulesByRef, flowBindings)
	if err != nil {
		return Result{}, err
	}
	drifts, err := collectBoundObjectDrifts(sharedFilesByRef, actualBoundObjectsByRef, boundObjectsOnlyFileRefs)
	if err != nil {
		return Result{}, err
	}

	removedBindingModules, err := candidateModulesWithRemovedSelectedBinding(repoRoot, moduleBindings, normalized.RuleRefs, normalized.RuleIDs, sharedFilesByRef, sharedFilesByID)
	if err != nil {
		return Result{}, err
	}
	removedBindingFlows, err := candidateObjectsWithRemovedSelectedBinding(repoRoot, flowBindings, normalized.RuleRefs, normalized.RuleIDs, sharedFilesByRef, sharedFilesByID)
	if err != nil {
		return Result{}, err
	}
	scopeModules, err := buildScopeModules(moduleBindings, sharedFilesByRef, sharedFilesByID, normalized, removedBindingModules)
	if err != nil {
		return Result{}, err
	}
	scopeFlows, err := buildScopeObjects(flowBindings, sharedFilesByRef, sharedFilesByID, normalized.RuleRefs, normalized.RuleIDs, removedBindingFlows)
	if err != nil {
		return Result{}, err
	}

	moduleImpactScope, err := scopedModulesForImpact(scopeModules, moduleBindings, normalized, sharedFilesByRef, sharedFilesByID, boundObjectsOnlyFileRefs, removedBindingModules)
	if err != nil {
		return Result{}, err
	}
	flowImpactScope, err := scopedObjectsForImpact("scenario", scopeFlows, flowBindings, normalized.RuleRefs, normalized.RuleIDs, sharedFilesByRef, sharedFilesByID, boundObjectsOnlyFileRefs, removedBindingFlows)
	if err != nil {
		return Result{}, err
	}

	impactResult, err := impactsync.Apply(repoRoot, impactsync.Input{
		Modules: moduleImpactScope,
		Flows:   flowImpactScope,
	})
	if err != nil {
		return Result{}, err
	}

	return Result{
		ScopedModules:                scopeModules,
		ScopedFlows:                  scopeFlows,
		ScopedRuleRefs:               normalized.RuleRefs,
		ScopedRuleIDs:                normalized.RuleIDs,
		StableLandingModule:          normalized.StableLandingModule,
		StableLandingRuleRefs:        normalized.StableLandingRuleRefs,
		BoundObjectsOnlyRuleFileRefs: normalized.BoundObjectsOnlyRuleFileRefs,
		ModuleResults:                impactResult.ModuleResults,
		FlowResults:                  impactResult.FlowResults,
		BoundObjectDrifts:            drifts,
	}, nil
}

func ReconcileBoundModules(repoRoot string, options ReconcileBoundModulesOptions) (ReconcileBoundModulesResult, error) {
	normalized := ReconcileBoundModulesOptions{
		Modules:  normalizeStrings(options.Modules),
		RuleRefs: normalizeStrings(options.RuleRefs),
		RuleIDs:  normalizeStrings(options.RuleIDs),
	}
	if len(normalized.Modules) == 0 && len(normalized.RuleRefs) == 0 && len(normalized.RuleIDs) == 0 {
		return ReconcileBoundModulesResult{}, fmt.Errorf("at least one of modules, rule refs, or rule ids is required")
	}

	sharedFilesByRef, err := loadSharedFiles(repoRoot)
	if err != nil {
		return ReconcileBoundModulesResult{}, err
	}
	sharedFilesByID := buildSharedFilesByID(sharedFilesByRef)
	moduleBindings, actualModulesByRef, _, _, err := loadModuleBindings(repoRoot)
	if err != nil {
		return ReconcileBoundModulesResult{}, err
	}
	flowBindings, _, err := loadObjectBindings(repoRoot, "scenario")
	if err != nil {
		return ReconcileBoundModulesResult{}, err
	}
	for _, module := range normalized.Modules {
		if _, ok := moduleBindings[module]; !ok {
			return ReconcileBoundModulesResult{}, fmt.Errorf("module %q is not registered in docs/specs/_status.md", module)
		}
	}
	for _, ref := range normalized.RuleRefs {
		if _, ok := sharedFilesByRef[ref]; !ok {
			return ReconcileBoundModulesResult{}, fmt.Errorf("rule ref %q is not present under docs/specs/rules/", ref)
		}
	}
	for _, sharedID := range normalized.RuleIDs {
		if _, ok := sharedFilesByID[sharedID]; !ok {
			return ReconcileBoundModulesResult{}, fmt.Errorf("rule id %q is not present under docs/specs/rules/", sharedID)
		}
	}

	touchedFiles := buildScopeSharedFiles(moduleBindings, sharedFilesByRef, sharedFilesByID, normalized)
	result := ReconcileBoundModulesResult{
		ScopedModules:  normalized.Modules,
		ScopedRuleRefs: normalized.RuleRefs,
		ScopedRuleIDs:  normalized.RuleIDs,
		TouchedFiles:   touchedFiles,
	}
	actualBoundObjectsByRef, err := collectActualBoundObjectsByRef(repoRoot, actualModulesByRef, flowBindings)
	if err != nil {
		return ReconcileBoundModulesResult{}, err
	}
	sharedFilesByFileRef := buildSharedFilesByFileRef(sharedFilesByRef)
	for _, fileRef := range touchedFiles {
		shared := sharedFilesByFileRef[fileRef]
		actualObjects := normalizeStrings(actualBoundObjectsByRef[shared.VersionRef])
		if sameStringSlice(shared.BoundObjects, actualObjects) {
			result.UnchangedFiles = append(result.UnchangedFiles, shared.FileRef)
			continue
		}
		if err := rewriteSharedBoundObjects(repoRoot, shared.FileRef, actualObjects); err != nil {
			return ReconcileBoundModulesResult{}, err
		}
		result.UpdatedFiles = append(result.UpdatedFiles, shared.FileRef)
	}
	result.UpdatedFiles = normalizeStrings(result.UpdatedFiles)
	result.UnchangedFiles = normalizeStrings(result.UnchangedFiles)
	return result, nil
}

func collectBoundObjectDrifts(sharedFilesByRef map[string]sharedFile, actualBoundObjectsByRef map[string][]string, boundObjectsOnlyFileRefs map[string]bool) ([]BoundObjectDrift, error) {
	refs := make([]string, 0, len(sharedFilesByRef))
	for ref := range sharedFilesByRef {
		refs = append(refs, ref)
	}
	sort.Strings(refs)

	drifts := []BoundObjectDrift{}
	for _, ref := range refs {
		shared := sharedFilesByRef[ref]
		actual := normalizeStrings(actualBoundObjectsByRef[ref])
		declared := normalizeStrings(shared.BoundObjects)
		if sameStringSlice(actual, declared) {
			continue
		}
		drifts = append(drifts, BoundObjectDrift{
			RuleID:                shared.RuleID,
			FileRef:               shared.FileRef,
			VersionRef:            shared.VersionRef,
			DeclaredObjects:       declared,
			ActualObjects:         actual,
			BoundObjectsOnlyDelta: boundObjectsOnlyFileRefs[shared.FileRef],
		})
	}
	return drifts, nil
}

func collectActualBoundObjectsByRef(repoRoot string, actualUnitsByRef map[string][]string, flowBindings map[string]objectBinding) (map[string][]string, error) {
	result := map[string][]string{}
	appendTypedRefs := func(versionRef, objectType string, objects []string) {
		for _, object := range objects {
			result[versionRef] = append(result[versionRef], typedBoundObjectRef(objectType, object))
		}
	}

	for versionRef, units := range actualUnitsByRef {
		appendTypedRefs(versionRef, "unit", units)
	}
	for object, binding := range flowBindings {
		for _, ref := range binding.RuleRefs {
			resolved, err := rulebinding.ResolveRef(repoRoot, binding.Status.ActiveLayer, ref)
			if err != nil {
				continue
			}
			result[resolved.VersionRef] = append(result[resolved.VersionRef], typedBoundObjectRef("scenario", object))
		}
	}
	for versionRef := range result {
		result[versionRef] = normalizeStrings(result[versionRef])
	}
	return result, nil
}

func loadSharedFiles(repoRoot string) (map[string]sharedFile, error) {
	result := map[string]sharedFile{}
	for _, root := range []struct {
		layer string
		dir   string
	}{
		{layer: "candidate", dir: "docs/specs/rules/candidate"},
		{layer: "stable", dir: "docs/specs/rules/stable"},
	} {
		pattern := filepath.Join(repoRoot, filepath.FromSlash(root.dir), "*.md")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			relPath, err := filepath.Rel(repoRoot, match)
			if err != nil {
				return nil, err
			}
			relPath = filepath.ToSlash(relPath)
			content, err := os.ReadFile(match)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", relPath, err)
			}
			shared, err := parseSharedFile(relPath, string(content))
			if err != nil {
				return nil, err
			}
			if err := rulebinding.ValidatePromotionOwnerUnit(repoRoot, shared.FileRef, shared.Layer, shared.PromotionOwnerUnit); err != nil {
				return nil, err
			}
			if shared.Layer != root.layer {
				return nil, fmt.Errorf("%s: frontmatter.layer=%s does not match path layer %s", relPath, shared.Layer, root.layer)
			}
			if _, exists := result[shared.VersionRef]; exists {
				return nil, fmt.Errorf("duplicate shared version ref %s", shared.VersionRef)
			}
			result[shared.VersionRef] = shared
		}
	}
	return result, nil
}

func buildSharedFilesByID(sharedFilesByRef map[string]sharedFile) map[string][]sharedFile {
	result := map[string][]sharedFile{}
	for _, shared := range sharedFilesByRef {
		result[shared.RuleID] = append(result[shared.RuleID], shared)
	}
	for sharedID := range result {
		sort.Slice(result[sharedID], func(i, j int) bool {
			return result[sharedID][i].VersionRef < result[sharedID][j].VersionRef
		})
	}
	return result
}

func buildSharedFilesByFileRef(sharedFilesByRef map[string]sharedFile) map[string]sharedFile {
	result := make(map[string]sharedFile, len(sharedFilesByRef))
	for _, shared := range sharedFilesByRef {
		result[shared.FileRef] = shared
	}
	return result
}

func loadModuleBindings(repoRoot string) (map[string]moduleBinding, map[string][]string, map[string][]string, []string, error) {
	statuses, err := statusfile.LoadModuleStatuses(repoRoot)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	bindings := make(map[string]moduleBinding, len(statuses))
	actualByRef := map[string][]string{}
	actualByID := map[string][]string{}
	unresolvedRefs := []string{}
	for _, status := range statuses {
		refs, err := readModuleRuleRefs(repoRoot, status)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		bindingIssues := []string{}
		bindings[status.Module] = moduleBinding{
			Status:   status,
			RuleRefs: refs,
		}
		for _, ref := range refs {
			resolved, err := rulebinding.ResolveRef(repoRoot, status.ActiveLayer, ref)
			if err != nil {
				bindingIssues = append(bindingIssues, err.Error())
				unresolvedRefs = append(unresolvedRefs, ref)
				continue
			}
			actualByRef[resolved.VersionRef] = append(actualByRef[resolved.VersionRef], status.Module)
			actualByID[resolved.RuleID] = append(actualByID[resolved.RuleID], status.Module)
		}
		bindings[status.Module] = moduleBinding{
			Status:        status,
			RuleRefs:      refs,
			BindingIssues: normalizeStrings(bindingIssues),
		}
	}
	for ref := range actualByRef {
		actualByRef[ref] = normalizeStrings(actualByRef[ref])
	}
	for sharedID := range actualByID {
		actualByID[sharedID] = normalizeStrings(actualByID[sharedID])
	}
	return bindings, actualByRef, actualByID, normalizeStrings(unresolvedRefs), nil
}

func buildScopeModules(moduleBindings map[string]moduleBinding, sharedFilesByRef map[string]sharedFile, sharedFilesByID map[string][]sharedFile, options Options, removedBindingScope map[string]bool) ([]string, error) {
	affected := map[string]bool{}
	for module, binding := range moduleBindings {
		selectedRefs, err := selectedRuleRefsForObject(binding.RuleRefs, options.RuleRefs, options.RuleIDs, sharedFilesByRef, sharedFilesByID)
		if err != nil {
			return nil, err
		}
		if len(selectedRefs) > 0 {
			affected[module] = true
		}
	}
	for module := range removedBindingScope {
		affected[module] = true
	}
	if len(options.Modules) == 0 {
		return sortedKeys(affected), nil
	}
	narrowed := map[string]bool{}
	for _, module := range options.Modules {
		if affected[module] {
			narrowed[module] = true
		}
	}
	return sortedKeys(narrowed), nil
}

func scopedModulesForImpact(scopeModules []string, moduleBindings map[string]moduleBinding, options Options, sharedFilesByRef map[string]sharedFile, sharedFilesByID map[string][]sharedFile, boundModulesOnlyFileRefs map[string]bool, removedBindingScope map[string]bool) ([]impactsync.ScopedModule, error) {
	result := make([]impactsync.ScopedModule, 0, len(scopeModules))
	stableLandingSharedRefSet := makeStringSet(options.StableLandingRuleRefs)
	for _, module := range scopeModules {
		binding := moduleBindings[module]
		selectedRuleRefs, err := selectedRuleRefsForObject(binding.RuleRefs, options.RuleRefs, options.RuleIDs, sharedFilesByRef, sharedFilesByID)
		if err != nil {
			return nil, err
		}
		invalidatingRuleRefs := filterInvalidatingRuleRefs(selectedRuleRefs, sharedFilesByRef, boundModulesOnlyFileRefs)
		if binding.Status.ActiveLayer == "stable" && options.StableLandingModule == module {
			invalidatingRuleRefs = subtractStringSet(invalidatingRuleRefs, stableLandingSharedRefSet)
		}
		result = append(result, impactsync.ScopedModule{
			Binding: impactsync.ModuleBinding{
				Module:        binding.Status.Module,
				ActiveLayer:   binding.Status.ActiveLayer,
				NextCommand:   binding.Status.NextCommand,
				BindingIssues: append([]string{}, binding.BindingIssues...),
			},
			InvalidatingRuleRefs:                  invalidatingRuleRefs,
			ExplicitFallbackScope:                 removedBindingScope[module],
			AllowedSharedSnapshotMismatchFileRefs: allowedSharedSnapshotMismatchFileRefs(selectedRuleRefs, sharedFilesByRef, boundModulesOnlyFileRefs),
		})
	}
	return result, nil
}

func scopedObjectsForImpact(objectType string, scopeObjects []string, bindings map[string]objectBinding, scopedRefs, scopedIDs []string, sharedFilesByRef map[string]sharedFile, sharedFilesByID map[string][]sharedFile, boundModulesOnlyFileRefs map[string]bool, removedBindingScope map[string]bool) ([]impactsync.ScopedObject, error) {
	result := make([]impactsync.ScopedObject, 0, len(scopeObjects))
	for _, object := range scopeObjects {
		binding := bindings[object]
		selectedRuleRefs, err := selectedRuleRefsForObject(binding.RuleRefs, scopedRefs, scopedIDs, sharedFilesByRef, sharedFilesByID)
		if err != nil {
			return nil, err
		}
		result = append(result, impactsync.ScopedObject{
			Binding: impactsync.ObjectBinding{
				ObjectType:    objectType,
				Object:        binding.Status.Object,
				ActiveLayer:   binding.Status.ActiveLayer,
				NextCommand:   binding.Status.NextCommand,
				BindingIssues: append([]string{}, binding.BindingIssues...),
			},
			ExplicitFallbackScope: removedBindingScope[object],
			InvalidatingRuleRefs:  filterInvalidatingRuleRefs(selectedRuleRefs, sharedFilesByRef, boundModulesOnlyFileRefs),
		})
	}
	return result, nil
}

func candidateModulesWithRemovedSelectedBinding(repoRoot string, moduleBindings map[string]moduleBinding, scopedRefs, scopedIDs []string, sharedFilesByRef map[string]sharedFile, sharedFilesByID map[string][]sharedFile) (map[string]bool, error) {
	result := map[string]bool{}
	for module, binding := range moduleBindings {
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
			"unit",
			binding.Status.Module,
			binding.Status.ActiveLayer,
			[]string{"check", "plan", "verify"},
			scopedRefs,
			scopedIDs,
			sharedFilesByID,
		)
		if err != nil {
			return nil, err
		}
		if matched {
			result[module] = true
		}
	}
	return result, nil
}

func allReferencedRuleRefs(moduleBindings map[string]moduleBinding, flowBindings map[string]objectBinding) map[string]bool {
	result := map[string]bool{}
	for _, binding := range moduleBindings {
		for _, ref := range binding.RuleRefs {
			result[ref] = true
		}
	}
	for _, binding := range flowBindings {
		for _, ref := range binding.RuleRefs {
			result[ref] = true
		}
	}
	return result
}

func sortedKeys(scope map[string]bool) []string {
	result := make([]string, 0, len(scope))
	for item := range scope {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

func buildScopeSharedFiles(moduleBindings map[string]moduleBinding, sharedFilesByRef map[string]sharedFile, sharedFilesByID map[string][]sharedFile, options ReconcileBoundModulesOptions) []string {
	scope := map[string]bool{}
	for _, module := range options.Modules {
		for _, ref := range moduleBindings[module].RuleRefs {
			if shared, ok := sharedFilesByRef[ref]; ok {
				scope[shared.FileRef] = true
			}
		}
	}
	for _, ref := range options.RuleRefs {
		if shared, ok := sharedFilesByRef[ref]; ok {
			scope[shared.FileRef] = true
		}
	}
	for _, sharedID := range options.RuleIDs {
		for _, shared := range sharedFilesByID[sharedID] {
			scope[shared.FileRef] = true
		}
	}
	result := make([]string, 0, len(scope))
	for fileRef := range scope {
		result = append(result, fileRef)
	}
	sort.Strings(result)
	return result
}

func selectedRuleRefsForObject(objectRefs, scopedRefs, scopedIDs []string, sharedFilesByRef map[string]sharedFile, sharedFilesByID map[string][]sharedFile) ([]string, error) {
	refSet := map[string]bool{}
	for _, ref := range scopedRefs {
		refSet[ref] = true
	}
	idSet := map[string]bool{}
	for _, sharedID := range scopedIDs {
		idSet[sharedID] = true
	}

	result := []string{}
	for _, ref := range objectRefs {
		if refSet[ref] {
			result = append(result, ref)
			continue
		}
		shared, ok := sharedFilesByRef[ref]
		if ok && idSet[shared.RuleID] {
			if len(sharedFilesByID[shared.RuleID]) > 1 {
				return nil, fmt.Errorf("cannot determine affected downstream objects safely for rule id %q because multiple current shared layers exist", shared.RuleID)
			}
			result = append(result, ref)
		}
	}
	return normalizeStrings(result), nil
}

func filterInvalidatingRuleRefs(selectedRefs []string, sharedFilesByRef map[string]sharedFile, boundObjectsOnlyFileRefs map[string]bool) []string {
	result := make([]string, 0, len(selectedRefs))
	for _, ref := range selectedRefs {
		shared, ok := sharedFilesByRef[ref]
		if !ok || !boundObjectsOnlyFileRefs[shared.FileRef] {
			result = append(result, ref)
		}
	}
	return normalizeStrings(result)
}

func allowedSharedSnapshotMismatchFileRefs(selectedRefs []string, sharedFilesByRef map[string]sharedFile, boundObjectsOnlyFileRefs map[string]bool) []string {
	result := []string{}
	for _, ref := range selectedRefs {
		shared, ok := sharedFilesByRef[ref]
		if ok && boundObjectsOnlyFileRefs[shared.FileRef] {
			result = append(result, shared.FileRef)
		}
	}
	return normalizeStrings(result)
}

func subtractStringSet(values []string, excluded map[string]bool) []string {
	if len(values) == 0 || len(excluded) == 0 {
		return normalizeStrings(values)
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if !excluded[value] {
			result = append(result, value)
		}
	}
	return normalizeStrings(result)
}

func makeStringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result[value] = true
		}
	}
	return result
}

func parseSharedFile(relPath, content string) (sharedFile, error) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return sharedFile{}, fmt.Errorf("%s: missing frontmatter start marker", relPath)
	}
	endIdx := -1
	for idx := 1; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "---" {
			endIdx = idx
			break
		}
	}
	if endIdx == -1 {
		return sharedFile{}, fmt.Errorf("%s: missing frontmatter end marker", relPath)
	}

	shared := sharedFile{FileRef: relPath}
	inBoundObjects := false
	for _, line := range lines[1:endIdx] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "bound_objects:") {
			inBoundObjects = true
			if strings.HasSuffix(trimmed, "none") {
				inBoundObjects = false
			}
			continue
		}
		if inBoundObjects {
			if strings.HasPrefix(trimmed, "- ") {
				ref := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
				ref = strings.Trim(ref, "`")
				boundRef, err := normalizeTypedBoundObjectRef(ref)
				if err != nil {
					return sharedFile{}, fmt.Errorf("%s: %w", relPath, err)
				}
				shared.BoundObjects = append(shared.BoundObjects, boundRef)
				continue
			}
			inBoundObjects = false
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "rule_id":
			shared.RuleID = value
		case "layer":
			shared.Layer = value
		case "rule_version":
			shared.VersionRef = fmt.Sprintf("%s@%s", strings.TrimSuffix(filepath.Base(relPath), ".md"), value)
		case "promotion_owner_unit":
			shared.PromotionOwnerUnit = value
		}
	}
	if shared.RuleID == "" || shared.Layer == "" || shared.VersionRef == "" {
		return sharedFile{}, fmt.Errorf("%s: missing rule_id, layer, or rule_version", relPath)
	}
	shared.BoundObjects = normalizeStrings(shared.BoundObjects)
	return shared, nil
}

func rewriteSharedBoundObjects(repoRoot, fileRef string, boundObjects []string) error {
	path := filepath.Join(repoRoot, filepath.FromSlash(fileRef))
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", fileRef, err)
	}
	content := strings.ReplaceAll(string(contentBytes), "\r\n", "\n")
	hadTrailingNewline := strings.HasSuffix(content, "\n")
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return fmt.Errorf("%s: missing frontmatter start marker", fileRef)
	}
	endIdx := -1
	for idx := 1; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "---" {
			endIdx = idx
			break
		}
	}
	if endIdx == -1 {
		return fmt.Errorf("%s: missing frontmatter end marker", fileRef)
	}

	boundLines := []string{"bound_objects: none"}
	if len(boundObjects) > 0 {
		boundLines = []string{"bound_objects:"}
		for _, boundObject := range normalizeStrings(boundObjects) {
			boundLines = append(boundLines, fmt.Sprintf("  - %s", boundObject))
		}
	}

	frontmatter := []string{"---"}
	inserted := false
	skipping := false
	for _, line := range lines[1:endIdx] {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "bound_objects:") {
			frontmatter = append(frontmatter, boundLines...)
			inserted = true
			skipping = true
			continue
		}
		if skipping {
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") || trimmed == "" {
				continue
			}
			skipping = false
		}
		frontmatter = append(frontmatter, line)
	}
	if !inserted {
		frontmatter = append(frontmatter, boundLines...)
	}
	frontmatter = append(frontmatter, "---")
	rewritten := strings.Join(append(frontmatter, lines[endIdx+1:]...), "\n")
	if hadTrailingNewline {
		rewritten += "\n"
	}
	if err := os.WriteFile(path, []byte(rewritten), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", fileRef, err)
	}
	return nil
}

func normalizeTypedBoundObjectRef(raw string) (string, error) {
	raw = strings.TrimSpace(strings.Trim(raw, "`"))
	if raw == "" {
		return "", fmt.Errorf("bound_objects contains an empty item")
	}
	objectType, object, ok := strings.Cut(raw, ":")
	if !ok {
		return "", fmt.Errorf("bound_objects item %q must use typed ref syntax <object_type>:<object>", raw)
	}
	objectType = strings.TrimSpace(objectType)
	object = strings.TrimSpace(object)
	switch objectType {
	case "unit", "scenario":
	default:
		return "", fmt.Errorf("bound_objects item %q uses unsupported object type %q", raw, objectType)
	}
	if object == "" {
		return "", fmt.Errorf("bound_objects item %q has an empty object id", raw)
	}
	return typedBoundObjectRef(objectType, object), nil
}

func typedBoundObjectRef(objectType, object string) string {
	return fmt.Sprintf("%s:%s", strings.TrimSpace(objectType), strings.TrimSpace(object))
}

func readModuleRuleRefs(repoRoot string, status statusfile.ModuleStatus) ([]string, error) {
	return readObjectRuleRefs(repoRoot, statusfile.ObjectStatus{
		ObjectType:  "unit",
		Object:      status.Module,
		ActiveLayer: status.ActiveLayer,
	})
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

	values := map[string]string{}
	for _, line := range lines[1:endIdx] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		values[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return values, strings.Join(lines[endIdx+1:], "\n"), nil
}

func parseRuleRefs(body string) ([]string, bool, error) {
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		right, matched, err := parseNamedFieldLine(trimmed, "rule_refs")
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
			return nil, false, fmt.Errorf("rule_refs must use literal none or a markdown list")
		}
		refs := []string{}
		seen := map[string]bool{}
		for next := idx + 1; next < len(lines); next++ {
			nextTrimmed := strings.TrimSpace(lines[next])
			if nextTrimmed == "" {
				continue
			}
			if strings.HasPrefix(nextTrimmed, "## ") || isNumberedListLine(nextTrimmed) {
				break
			}
			if !strings.HasPrefix(nextTrimmed, "- ") {
				return nil, false, fmt.Errorf("rule_refs must be a markdown list of rule refs")
			}
			ref := strings.TrimSpace(strings.TrimPrefix(nextTrimmed, "- "))
			ref = strings.Trim(ref, "`")
			if ref == "" {
				return nil, false, fmt.Errorf("rule_refs contains an empty item")
			}
			if seen[ref] {
				return nil, false, fmt.Errorf("rule_refs contains duplicate item %q", ref)
			}
			seen[ref] = true
			refs = append(refs, ref)
		}
		if len(refs) == 0 {
			return nil, false, fmt.Errorf("rule_refs must not be an empty list")
		}
		if err := validateOrderedRuleRefs(refs); err != nil {
			return nil, false, err
		}
		return refs, true, nil
	}
	return nil, false, nil
}

func validateOrderedRuleRefs(refs []string) error {
	if len(refs) < 2 {
		return nil
	}
	expected := append([]string(nil), refs...)
	sort.Strings(expected)
	for idx := range refs {
		if refs[idx] != expected[idx] {
			return fmt.Errorf("rule_refs must be sorted by exact rule ref string in ascending lexical order")
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

func isNumberedListLine(line string) bool {
	if line == "" {
		return false
	}
	digits := 0
	for digits < len(line) && line[digits] >= '0' && line[digits] <= '9' {
		digits++
	}
	return digits > 0 && digits < len(line) && line[digits] == '.'
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

func sameStringSlice(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for idx := range left {
		if left[idx] != right[idx] {
			return false
		}
	}
	return true
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
