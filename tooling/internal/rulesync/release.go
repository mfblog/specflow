package rulesync

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/rulebinding"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/rulerefs"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

type ConsumerOptions struct {
	RuleID  string
	RuleRef string
}

type Consumer struct {
	ObjectType  string
	Object      string
	ActiveLayer string
	FileRef     string
	RuleRefs    []string
}

type ConsumerResult struct {
	RuleID    string
	RuleRef   string
	Consumers []Consumer
}

type ReleaseVersionOptions struct {
	RuleID  string
	FromRef string
	ToRef   string
}

type ReleaseVersionResult struct {
	RuleID              string
	FromRef             string
	ToRef               string
	CandidateUpdated    []string
	StableForked        []string
	ProcessFilesRemoved []string
	Sync                Result
}

func Consumers(repoRoot string, options ConsumerOptions) (ConsumerResult, error) {
	ruleID := strings.TrimSpace(options.RuleID)
	ruleRef := strings.TrimSpace(options.RuleRef)
	if (ruleID == "") == (ruleRef == "") {
		return ConsumerResult{}, fmt.Errorf("exactly one of rule id or rule ref is required")
	}

	sharedFilesByRef, err := loadSharedFiles(repoRoot)
	if err != nil {
		return ConsumerResult{}, err
	}
	statuses, err := statusfile.LoadObjectStatuses(repoRoot)
	if err != nil {
		return ConsumerResult{}, err
	}

	result := ConsumerResult{RuleID: ruleID, RuleRef: ruleRef}
	for _, status := range statuses {
		fileRef, err := specpaths.ObjectMainSpecFileRef(status.ObjectType, status.ActiveLayer, status.Object)
		if err != nil {
			return ConsumerResult{}, err
		}
		refs, err := readObjectRuleRefs(repoRoot, status)
		if err != nil {
			return ConsumerResult{}, err
		}
		matchedRefs := []string{}
		for _, ref := range refs {
			if ruleRef != "" {
				if ref == ruleRef {
					matchedRefs = append(matchedRefs, ref)
				}
				continue
			}
			shared, ok := sharedFilesByRef[ref]
			if ok && shared.RuleID == ruleID {
				matchedRefs = append(matchedRefs, ref)
			}
		}
		if len(matchedRefs) == 0 {
			continue
		}
		result.Consumers = append(result.Consumers, Consumer{
			ObjectType:  status.ObjectType,
			Object:      status.Object,
			ActiveLayer: status.ActiveLayer,
			FileRef:     fileRef,
			RuleRefs:    normalizeStrings(matchedRefs),
		})
	}
	return result, nil
}

func ReleaseVersion(repoRoot string, options ReleaseVersionOptions) (ReleaseVersionResult, error) {
	normalized := ReleaseVersionOptions{
		RuleID:  strings.TrimSpace(options.RuleID),
		FromRef: strings.TrimSpace(options.FromRef),
		ToRef:   strings.TrimSpace(options.ToRef),
	}
	if normalized.RuleID == "" || normalized.FromRef == "" || normalized.ToRef == "" {
		return ReleaseVersionResult{}, fmt.Errorf("rule id, from-ref, and to-ref are required")
	}
	if normalized.FromRef == normalized.ToRef {
		return ReleaseVersionResult{}, fmt.Errorf("from-ref and to-ref must be different")
	}

	sharedFilesByRef, err := loadSharedFiles(repoRoot)
	if err != nil {
		return ReleaseVersionResult{}, err
	}
	toShared, ok := sharedFilesByRef[normalized.ToRef]
	if !ok {
		return ReleaseVersionResult{}, fmt.Errorf("to-ref %q is not present under docs/specs/rules/", normalized.ToRef)
	}
	if toShared.RuleID != normalized.RuleID {
		return ReleaseVersionResult{}, fmt.Errorf("to-ref %q belongs to rule_id %q, not %q", normalized.ToRef, toShared.RuleID, normalized.RuleID)
	}
	if toShared.Layer != "stable" {
		return ReleaseVersionResult{}, fmt.Errorf("to-ref %q must point to a stable-layer rule file", normalized.ToRef)
	}
	fromPrefix, err := ruleRefPrefix(normalized.FromRef)
	if err != nil {
		return ReleaseVersionResult{}, err
	}
	toPrefix, err := ruleRefPrefix(normalized.ToRef)
	if err != nil {
		return ReleaseVersionResult{}, err
	}
	if !strings.HasPrefix(fromPrefix, "s_") {
		return ReleaseVersionResult{}, fmt.Errorf("from-ref %q must point to a stable-layer rule ref", normalized.FromRef)
	}
	if fromPrefix != toPrefix {
		return ReleaseVersionResult{}, fmt.Errorf("from-ref %q and to-ref %q must refer to the same rule file prefix", normalized.FromRef, normalized.ToRef)
	}

	statuses, err := statusfile.LoadObjectStatuses(repoRoot)
	if err != nil {
		return ReleaseVersionResult{}, err
	}
	result := ReleaseVersionResult{
		RuleID:  normalized.RuleID,
		FromRef: normalized.FromRef,
		ToRef:   normalized.ToRef,
	}
	changedObjects := map[string]statusfile.ObjectStatus{}

	for _, status := range statuses {
		fileRef, err := specpaths.ObjectMainSpecFileRef(status.ObjectType, status.ActiveLayer, status.Object)
		if err != nil {
			return ReleaseVersionResult{}, err
		}
		contentBytes, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(fileRef)))
		if err != nil {
			return ReleaseVersionResult{}, fmt.Errorf("read %s: %w", fileRef, err)
		}
		refs, err := rulerefs.ParseObjectRuleRefs(fileRef, string(contentBytes))
		if err != nil {
			return ReleaseVersionResult{}, err
		}
		nextRefs, changed := replaceRuleRef(refs, normalized.FromRef, normalized.ToRef)
		if !changed {
			continue
		}
		key := status.ObjectType + ":" + status.Object
		if status.ActiveLayer == "candidate" {
			updated, err := rulerefs.UpdateObjectRuleRefs(fileRef, string(contentBytes), nextRefs)
			if err != nil {
				return ReleaseVersionResult{}, err
			}
			if err := os.WriteFile(filepath.Join(repoRoot, filepath.FromSlash(fileRef)), []byte(updated), 0o644); err != nil {
				return ReleaseVersionResult{}, fmt.Errorf("write %s: %w", fileRef, err)
			}
			if removed, err := removeProcessArtifacts(repoRoot, status.ObjectType, status.Object); err != nil {
				return ReleaseVersionResult{}, err
			} else {
				result.ProcessFilesRemoved = append(result.ProcessFilesRemoved, removed...)
			}
			status.NextCommand = checkCommandForObject(status.ObjectType)
			status.Notes = releaseNote(normalized.FromRef, normalized.ToRef, false)
			if _, err := statusfile.UpsertObjectStatus(repoRoot, status, false); err != nil {
				return ReleaseVersionResult{}, err
			}
			result.CandidateUpdated = append(result.CandidateUpdated, key)
			changedObjects[key] = status
			continue
		}
		if status.ActiveLayer != "stable" {
			return ReleaseVersionResult{}, fmt.Errorf("%s %q has unsupported active layer %q", status.ObjectType, status.Object, status.ActiveLayer)
		}

		candidateRef, err := specpaths.ObjectMainSpecFileRef(status.ObjectType, "candidate", status.Object)
		if err != nil {
			return ReleaseVersionResult{}, err
		}
		version, err := rulerefs.FrontmatterScalar(fileRef, string(contentBytes), "version")
		if err != nil {
			return ReleaseVersionResult{}, err
		}
		nextVersion, err := bumpPatchVersion(version)
		if err != nil {
			return ReleaseVersionResult{}, fmt.Errorf("%s: %w", fileRef, err)
		}
		frontmatterUpdates := map[string]string{
			"layer":                 "candidate",
			"version":               nextVersion,
			"source_basis":          "new_design",
			"evidence_appendix_ref": "none",
		}
		if status.ObjectType == "unit" {
			frontmatterUpdates["candidate_intent"] = "change"
		}
		updated, err := rulerefs.RewriteObjectFrontmatter(candidateRef, string(contentBytes), frontmatterUpdates, nextRefs)
		if err != nil {
			return ReleaseVersionResult{}, err
		}
		candidateAbs := filepath.Join(repoRoot, filepath.FromSlash(candidateRef))
		if err := os.MkdirAll(filepath.Dir(candidateAbs), 0o755); err != nil {
			return ReleaseVersionResult{}, fmt.Errorf("create candidate dir for %s: %w", candidateRef, err)
		}
		if err := os.WriteFile(candidateAbs, []byte(updated), 0o644); err != nil {
			return ReleaseVersionResult{}, fmt.Errorf("write %s: %w", candidateRef, err)
		}
		if removed, err := removeProcessArtifacts(repoRoot, status.ObjectType, status.Object); err != nil {
			return ReleaseVersionResult{}, err
		} else {
			result.ProcessFilesRemoved = append(result.ProcessFilesRemoved, removed...)
		}
		status.Candidate = "yes"
		status.ActiveLayer = "candidate"
		status.NextCommand = checkCommandForObject(status.ObjectType)
		status.Notes = releaseNote(normalized.FromRef, normalized.ToRef, true)
		if _, err := statusfile.UpsertObjectStatus(repoRoot, status, false); err != nil {
			return ReleaseVersionResult{}, err
		}
		result.StableForked = append(result.StableForked, key)
		changedObjects[key] = status
	}

	if len(changedObjects) == 0 {
		return ReleaseVersionResult{}, fmt.Errorf("no current-layer unit/scenario binds from-ref %q", normalized.FromRef)
	}

	syncResult, err := SyncImpact(repoRoot, Options{RuleRefs: []string{normalized.ToRef}})
	if err != nil {
		return ReleaseVersionResult{}, err
	}
	result.Sync = syncResult

	if diagnostics := ValidateCurrentBindings(repoRoot, normalized.FromRef); len(diagnostics) > 0 {
		return ReleaseVersionResult{}, fmt.Errorf("post-release current binding validation failed: %s", strings.Join(diagnostics, "; "))
	}
	result.CandidateUpdated = normalizeStrings(result.CandidateUpdated)
	result.StableForked = normalizeStrings(result.StableForked)
	result.ProcessFilesRemoved = normalizeStrings(result.ProcessFilesRemoved)
	return result, nil
}

func ValidateCurrentBindings(repoRoot string, forbiddenRef string) []string {
	diagnostics := []string{}
	if _, err := loadSharedFiles(repoRoot); err != nil {
		diagnostics = append(diagnostics, err.Error())
	}
	statuses, err := statusfile.LoadObjectStatuses(repoRoot)
	if err != nil {
		return append(diagnostics, err.Error())
	}
	for _, status := range statuses {
		fileRef, err := specpaths.ObjectMainSpecFileRef(status.ObjectType, status.ActiveLayer, status.Object)
		if err != nil {
			diagnostics = append(diagnostics, err.Error())
			continue
		}
		content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(fileRef)))
		if err != nil {
			diagnostics = append(diagnostics, fmt.Sprintf("read %s: %v", fileRef, err))
			continue
		}
		refs, err := rulerefs.ParseObjectRuleRefs(fileRef, string(content))
		if err != nil {
			diagnostics = append(diagnostics, err.Error())
			continue
		}
		for _, ref := range refs {
			if forbiddenRef != "" && ref == forbiddenRef {
				diagnostics = append(diagnostics, fmt.Sprintf("%s: forbidden rule ref %s remains", fileRef, forbiddenRef))
			}
			if _, err := rulebinding.ResolveRef(repoRoot, status.ActiveLayer, ref); err != nil {
				diagnostics = append(diagnostics, err.Error())
			}
		}
	}
	return normalizeStrings(diagnostics)
}

func replaceRuleRef(refs []string, fromRef, toRef string) ([]string, bool) {
	changed := false
	next := make([]string, 0, len(refs))
	for _, ref := range refs {
		if ref == fromRef {
			next = append(next, toRef)
			changed = true
			continue
		}
		next = append(next, ref)
	}
	if !changed {
		return refs, false
	}
	normalized, err := rulerefs.NormalizeRuleRefs(next)
	if err != nil {
		return next, true
	}
	return normalized, true
}

func ruleRefPrefix(ref string) (string, error) {
	normalized, err := rulerefs.NormalizeRuleRefs([]string{ref})
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(normalized[0], "@", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid rule ref %q", ref)
	}
	return parts[0], nil
}

func removeProcessArtifacts(repoRoot, objectType, object string) ([]string, error) {
	kinds := []string{"check", "verify"}
	if objectType == "unit" {
		kinds = []string{"check", "plan", "verify"}
	}
	removed := []string{}
	for _, kind := range kinds {
		paths, err := snapshot.ProcessArtifactPaths(objectType, object, kind)
		if err != nil {
			return nil, err
		}
		for _, fileRef := range paths {
			abs := filepath.Join(repoRoot, filepath.FromSlash(fileRef))
			if err := os.Remove(abs); err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, fmt.Errorf("remove %s: %w", fileRef, err)
			}
			removed = append(removed, fileRef)
		}
	}
	return removed, nil
}

func checkCommandForObject(objectType string) string {
	if objectType == "scenario" {
		return "scenario_check"
	}
	return "unit_check"
}

func releaseNote(fromRef, toRef string, forked bool) string {
	if forked {
		return fmt.Sprintf("Auto-forked by rule release-version from %s to %s; rerun check.", fromRef, toRef)
	}
	return fmt.Sprintf("Retargeted by rule release-version from %s to %s; rerun check.", fromRef, toRef)
}

func bumpPatchVersion(version string) (string, error) {
	parts := strings.Split(strings.TrimSpace(version), ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("version %q must use MAJOR.MINOR.PATCH", version)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", fmt.Errorf("version %q has invalid PATCH: %w", version, err)
	}
	return fmt.Sprintf("%s.%s.%d", parts[0], parts[1], patch+1), nil
}
