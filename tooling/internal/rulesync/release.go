package rulesync

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/commandclose"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/rulebinding"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/rulerefs"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/unitappendix"
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
	AppendixRetargeted  []string
	AppendixRemoved     []string
	ProcessFilesRemoved []string
	Sync                Result
}

type candidateReleasePlan struct {
	key     string
	status  statusfile.ObjectStatus
	fileRef string
	content string
}

type stableReleasePlan struct {
	key                string
	status             statusfile.ObjectStatus
	candidateFileRef   string
	content            string
	appendices         []preparedAppendix
	appendixRetargeted []string
	appendixRemoved    []string
}

type preparedAppendix struct {
	fileRef string
	content string
}

var markdownLinkTargetPattern = regexp.MustCompile(`\]\(([^)\n]+)\)`)

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
		if status.ObjectType != "unit" {
			continue
		}
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
	candidatePlans := []candidateReleasePlan{}
	stablePlans := []stableReleasePlan{}

	for _, status := range statuses {
		if status.ObjectType != "unit" {
			continue
		}
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
			candidatePlans = append(candidatePlans, candidateReleasePlan{
				key:     key,
				status:  status,
				fileRef: fileRef,
				content: updated,
			})
			continue
		}
		if status.ActiveLayer != "stable" {
			return ReleaseVersionResult{}, fmt.Errorf("%s %q has unsupported active layer %q", status.ObjectType, status.Object, status.ActiveLayer)
		}
		if status.NextCommand != "unit_fork" {
			return ReleaseVersionResult{}, fmt.Errorf("rule release-version stable auto-fork for %s %q requires current Next Command unit_fork, got %s", status.ObjectType, status.Object, status.NextCommand)
		}
		requirement, err := snapshot.StableVerifyCandidateIntentRequirement(repoRoot, status.ObjectType, status.Object)
		if err != nil {
			return ReleaseVersionResult{}, err
		}
		if requirement.Required && requirement.RequiredIntent != "change" {
			return ReleaseVersionResult{}, fmt.Errorf("rule release-version stable auto-fork for %s %q conflicts with stable verify decision %s: required candidate_intent=%s, release-version writes candidate_intent=change", status.ObjectType, status.Object, requirement.Decision, requirement.RequiredIntent)
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
			"candidate_intent":      "change",
			"layer":                 "candidate",
			"version":               nextVersion,
			"evidence_appendix_ref": "none",
		}
		updated, err := rulerefs.RewriteObjectFrontmatter(candidateRef, string(contentBytes), frontmatterUpdates, nextRefs)
		if err != nil {
			return ReleaseVersionResult{}, err
		}
		updated, appendices, appendixRetargeted, appendixRemoved, err := prepareStableAppendicesForFork(repoRoot, status.ObjectType, status.Object, fileRef, candidateRef, string(contentBytes), updated)
		if err != nil {
			return ReleaseVersionResult{}, err
		}
		stablePlans = append(stablePlans, stableReleasePlan{
			key:                key,
			status:             status,
			candidateFileRef:   candidateRef,
			content:            updated,
			appendices:         appendices,
			appendixRetargeted: appendixRetargeted,
			appendixRemoved:    appendixRemoved,
		})
	}

	if len(candidatePlans)+len(stablePlans) == 0 {
		return ReleaseVersionResult{}, fmt.Errorf("no current-layer unit binds from-ref %q", normalized.FromRef)
	}

	for _, plan := range candidatePlans {
		if err := os.WriteFile(filepath.Join(repoRoot, filepath.FromSlash(plan.fileRef)), []byte(plan.content), 0o644); err != nil {
			return ReleaseVersionResult{}, fmt.Errorf("write %s: %w", plan.fileRef, err)
		}
		if removed, err := removeProcessArtifacts(repoRoot, plan.status.ObjectType, plan.status.Object); err != nil {
			return ReleaseVersionResult{}, err
		} else {
			result.ProcessFilesRemoved = append(result.ProcessFilesRemoved, removed...)
		}
		status := plan.status
		status.NextCommand = checkCommandForObject(status.ObjectType)
		status.Notes = releaseNote(normalized.FromRef, normalized.ToRef, false)
		if _, err := statusfile.UpsertObjectStatus(repoRoot, status, false); err != nil {
			return ReleaseVersionResult{}, err
		}
		result.CandidateUpdated = append(result.CandidateUpdated, plan.key)
	}

	for _, plan := range stablePlans {
		candidateAbs := filepath.Join(repoRoot, filepath.FromSlash(plan.candidateFileRef))
		if err := os.MkdirAll(filepath.Dir(candidateAbs), 0o755); err != nil {
			return ReleaseVersionResult{}, fmt.Errorf("create candidate dir for %s: %w", plan.candidateFileRef, err)
		}
		if err := os.WriteFile(candidateAbs, []byte(plan.content), 0o644); err != nil {
			return ReleaseVersionResult{}, fmt.Errorf("write %s: %w", plan.candidateFileRef, err)
		}
		if err := removeAppendixRefs(repoRoot, plan.appendixRemoved); err != nil {
			return ReleaseVersionResult{}, err
		}
		if err := writePreparedAppendices(repoRoot, plan.appendices); err != nil {
			return ReleaseVersionResult{}, err
		}
		closeResult, err := commandclose.Close(commandclose.Options{
			RepoRoot:   repoRoot,
			Command:    "unit_fork",
			ObjectType: plan.status.ObjectType,
			Object:     plan.status.Object,
			Outcome:    "candidate_created",
			Notes:      releaseNote(normalized.FromRef, normalized.ToRef, true),
			Apply:      true,
		})
		if err != nil {
			return ReleaseVersionResult{}, err
		}
		result.ProcessFilesRemoved = append(result.ProcessFilesRemoved, closeResult.SuccessCleanup.DeletedFiles...)
		result.AppendixRetargeted = append(result.AppendixRetargeted, plan.appendixRetargeted...)
		result.AppendixRemoved = append(result.AppendixRemoved, plan.appendixRemoved...)
		result.StableForked = append(result.StableForked, plan.key)
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
	result.AppendixRetargeted = normalizeStrings(result.AppendixRetargeted)
	result.AppendixRemoved = normalizeStrings(result.AppendixRemoved)
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
		if status.ObjectType != "unit" {
			continue
		}
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
	if objectType != "unit" {
		return nil, fmt.Errorf("unsupported object type %q", objectType)
	}
	kinds := []string{"check_work", "check", "plan", "verify", "stable_verify"}
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

func prepareStableAppendicesForFork(repoRoot, objectType, object, stableMainRef, candidateMainRef, stableContent, candidateContent string) (string, []preparedAppendix, []string, []string, error) {
	_ = stableContent
	stableEntries, err := unitappendix.Scan(repoRoot, objectType, object, "stable")
	if err != nil {
		return "", nil, nil, nil, err
	}
	stableAppendices := make([]string, 0, len(stableEntries))
	stableContentByRef := stableEntriesByRef(stableEntries)
	for _, entry := range stableEntries {
		stableAppendices = append(stableAppendices, entry.FileRef)
	}
	refMap := map[string]string{
		path.Clean(stableMainRef): path.Clean(candidateMainRef),
	}
	prepared := []preparedAppendix{}
	for _, stableAppendixRef := range stableAppendices {
		candidateAppendixRef, err := candidateAppendixRefForStable(objectType, object, stableAppendixRef)
		if err != nil {
			return "", nil, nil, nil, err
		}
		refMap[stableAppendixRef] = candidateAppendixRef
	}
	for _, stableAppendixRef := range stableAppendices {
		candidateAppendixRef := refMap[stableAppendixRef]
		appendixContent := stableContentByRef[stableAppendixRef]
		appendixContent = rewriteCandidateAppendixFrontmatter(appendixContent, objectType, object)
		appendixContent = rewriteMarkdownDocRefs(appendixContent, stableAppendixRef, candidateAppendixRef, refMap)
		appendixContent = rewriteKnownDocRefLiterals(appendixContent, refMap)
		prepared = append(prepared, preparedAppendix{fileRef: candidateAppendixRef, content: appendixContent})
	}

	removed, err := candidateAppendices(repoRoot, objectType, object)
	if err != nil {
		return "", nil, nil, nil, err
	}
	if len(stableAppendices) == 0 {
		return candidateContent, nil, nil, normalizeStrings(removed), nil
	}

	nextCandidateContent := rewriteMarkdownDocRefs(candidateContent, stableMainRef, candidateMainRef, refMap)
	nextCandidateContent = rewriteKnownDocRefLiterals(nextCandidateContent, refMap)
	retargeted := []string{}
	for _, appendix := range prepared {
		retargeted = append(retargeted, appendix.fileRef)
	}

	return nextCandidateContent, prepared, normalizeStrings(retargeted), normalizeStrings(removed), nil
}

func stableEntriesByRef(entries []unitappendix.Entry) map[string]string {
	result := map[string]string{}
	for _, entry := range entries {
		result[entry.FileRef] = entry.Content
	}
	return result
}

func candidateAppendixRefForStable(objectType, object, stableAppendixRef string) (string, error) {
	stablePrefix, candidatePrefix, err := appendixPrefixes(objectType, object)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(stableAppendixRef, stablePrefix) || !strings.HasSuffix(stableAppendixRef, ".md") {
		return "", fmt.Errorf("%s is not a same-%s stable appendix for %q", stableAppendixRef, objectType, object)
	}
	return candidatePrefix + strings.TrimPrefix(stableAppendixRef, stablePrefix), nil
}

func appendixPrefixes(objectType, object string) (string, string, error) {
	switch objectType {
	case "unit":
		return fmt.Sprintf("docs/specs/units/stable/appendix/s_unit_%s_", object),
			fmt.Sprintf("docs/specs/units/candidate/appendix/c_unit_%s_", object), nil
	default:
		return "", "", fmt.Errorf("unsupported object type %q", objectType)
	}
}

func rewriteMarkdownDocRefs(content, resolveFromRef, outputFromRef string, refMap map[string]string) string {
	return markdownLinkTargetPattern.ReplaceAllStringFunc(content, func(token string) string {
		match := markdownLinkTargetPattern.FindStringSubmatch(token)
		if len(match) != 2 {
			return token
		}
		target, suffix, ok := splitMarkdownTarget(match[1])
		if !ok {
			return token
		}
		resolvedRef, anchor, ok := resolveMarkdownDocRef(resolveFromRef, target)
		if !ok {
			return token
		}
		nextRef, ok := refMap[resolvedRef]
		if !ok {
			return token
		}
		nextTarget := relativeDocRef(outputFromRef, nextRef) + anchor + suffix
		return strings.Replace(token, "("+match[1]+")", "("+nextTarget+")", 1)
	})
}

func rewriteKnownDocRefLiterals(content string, refMap map[string]string) string {
	type refPair struct {
		from string
		to   string
	}
	pairs := make([]refPair, 0, len(refMap))
	for from, to := range refMap {
		pairs = append(pairs, refPair{from: from, to: to})
	}
	sort.Slice(pairs, func(left, right int) bool {
		if len(pairs[left].from) == len(pairs[right].from) {
			return pairs[left].from < pairs[right].from
		}
		return len(pairs[left].from) > len(pairs[right].from)
	})
	for _, pair := range pairs {
		content = strings.ReplaceAll(content, pair.from, pair.to)
		content = replaceDelimitedDocRefToken(content, path.Base(pair.from), path.Base(pair.to))
	}
	return content
}

func replaceDelimitedDocRefToken(content, from, to string) string {
	if from == "" || from == to {
		return content
	}
	pattern := regexp.MustCompile(`(^|[^A-Za-z0-9_-])` + regexp.QuoteMeta(from) + `($|[^A-Za-z0-9_-])`)
	return pattern.ReplaceAllStringFunc(content, func(match string) string {
		indexes := pattern.FindStringSubmatchIndex(match)
		if len(indexes) != 6 {
			return match
		}
		return match[indexes[2]:indexes[3]] + to + match[indexes[4]:indexes[5]]
	})
}

func splitMarkdownTarget(raw string) (string, string, bool) {
	target := strings.TrimSpace(raw)
	if target == "" {
		return "", "", false
	}
	if strings.HasPrefix(target, "<") {
		end := strings.Index(target, ">")
		if end <= 0 {
			return "", "", false
		}
		return target[1:end], target[end+1:], true
	}
	if idx := strings.IndexAny(target, " \t\r\n"); idx >= 0 {
		return target[:idx], target[idx:], true
	}
	return target, "", true
}

func resolveMarkdownDocRef(fromRef, target string) (string, string, bool) {
	target = strings.TrimSpace(strings.ReplaceAll(target, "\\", "/"))
	if target == "" ||
		strings.HasPrefix(target, "#") ||
		strings.Contains(target, "://") ||
		strings.HasPrefix(target, "mailto:") {
		return "", "", false
	}

	docTarget, anchor, hasAnchor := strings.Cut(target, "#")
	if docTarget == "" {
		return "", "", false
	}
	if hasAnchor {
		anchor = "#" + anchor
	} else {
		anchor = ""
	}

	var ref string
	if strings.HasPrefix(docTarget, "docs/") {
		ref = path.Clean(docTarget)
	} else {
		ref = path.Clean(path.Join(path.Dir(fromRef), docTarget))
	}
	if !strings.HasPrefix(ref, "docs/specs/") {
		return "", "", false
	}
	return ref, anchor, true
}

func relativeDocRef(fromRef, targetRef string) string {
	rel, err := filepath.Rel(filepath.FromSlash(path.Dir(fromRef)), filepath.FromSlash(targetRef))
	if err != nil {
		return targetRef
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, ".") {
		return rel
	}
	return "./" + rel
}

func rewriteCandidateAppendixFrontmatter(content, objectType, object string) string {
	ownerKey := objectType
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return fmt.Sprintf("---\n%s: %s\nlayer: candidate\n---\n\n%s", ownerKey, object, content)
	}

	end := -1
	for idx := 1; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "---" {
			end = idx
			break
		}
	}
	if end == -1 {
		return fmt.Sprintf("---\n%s: %s\nlayer: candidate\n---\n\n%s", ownerKey, object, content)
	}

	nextFrontmatter := []string{"---"}
	ownerWritten := false
	layerWritten := false
	for _, line := range lines[1:end] {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, ownerKey+":"):
			nextFrontmatter = append(nextFrontmatter, fmt.Sprintf("%s: %s", ownerKey, object))
			ownerWritten = true
		case strings.HasPrefix(trimmed, "layer:"):
			nextFrontmatter = append(nextFrontmatter, "layer: candidate")
			layerWritten = true
		case strings.HasPrefix(trimmed, "spec_version_ref:"):
			continue
		default:
			nextFrontmatter = append(nextFrontmatter, line)
		}
	}
	if !ownerWritten {
		nextFrontmatter = append(nextFrontmatter, fmt.Sprintf("%s: %s", ownerKey, object))
	}
	if !layerWritten {
		nextFrontmatter = append(nextFrontmatter, "layer: candidate")
	}
	nextFrontmatter = append(nextFrontmatter, "---")

	body := strings.Join(lines[end+1:], "\n")
	if body == "" {
		return strings.Join(nextFrontmatter, "\n") + "\n"
	}
	return strings.Join(nextFrontmatter, "\n") + "\n" + body
}

func removeCandidateAppendices(repoRoot, objectType, object string) ([]string, error) {
	refs, err := candidateAppendices(repoRoot, objectType, object)
	if err != nil {
		return nil, err
	}
	if err := removeAppendixRefs(repoRoot, refs); err != nil {
		return nil, err
	}
	return normalizeStrings(refs), nil
}

func candidateAppendices(repoRoot, objectType, object string) ([]string, error) {
	glob, err := specpaths.ObjectCandidateAppendixGlob(objectType, object)
	if err != nil {
		return nil, err
	}
	matches, err := filepath.Glob(filepath.Join(repoRoot, filepath.FromSlash(glob)))
	if err != nil {
		return nil, err
	}
	removed := []string{}
	for _, match := range matches {
		relPath, err := filepath.Rel(repoRoot, match)
		if err != nil {
			return nil, err
		}
		fileRef := filepath.ToSlash(relPath)
		removed = append(removed, fileRef)
	}
	return normalizeStrings(removed), nil
}

func removeAppendixRefs(repoRoot string, refs []string) error {
	for _, ref := range refs {
		abs := filepath.Join(repoRoot, filepath.FromSlash(ref))
		if err := os.Remove(abs); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("remove %s: %w", ref, err)
		}
	}
	return nil
}

func writePreparedAppendices(repoRoot string, appendices []preparedAppendix) error {
	for _, appendix := range appendices {
		candidateAppendixAbs := filepath.Join(repoRoot, filepath.FromSlash(appendix.fileRef))
		if err := os.MkdirAll(filepath.Dir(candidateAppendixAbs), 0o755); err != nil {
			return fmt.Errorf("create candidate appendix dir for %s: %w", appendix.fileRef, err)
		}
		if err := os.WriteFile(candidateAppendixAbs, []byte(appendix.content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", appendix.fileRef, err)
		}
	}
	return nil
}

func checkCommandForObject(objectType string) string {
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
