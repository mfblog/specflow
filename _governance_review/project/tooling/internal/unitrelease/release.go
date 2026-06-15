package unitrelease

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/unitrefs"
)

type Options struct {
	Unit    string
	FromRef string
	ToRef   string
}

type Result struct {
	Unit                string
	FromRef             string
	ToRef               string
	CandidateUpdated    []string
	StableRerouted      []string
	MainSpecsUpdated    []string
	ProcessFilesRemoved []string
	StatusUpdated       []string
	Noop                bool
}

var stableUnitRefPattern = regexp.MustCompile(`^s_unit_([a-z0-9_]+)@([0-9]+\.[0-9]+\.[0-9]+)$`)

func ReleaseVersion(repoRoot string, options Options) (Result, error) {
	normalized := Options{
		Unit:    strings.TrimSpace(options.Unit),
		FromRef: strings.TrimSpace(options.FromRef),
		ToRef:   strings.TrimSpace(options.ToRef),
	}
	if normalized.Unit == "" || normalized.FromRef == "" || normalized.ToRef == "" {
		return Result{}, fmt.Errorf("unit, from-ref, and to-ref are required")
	}
	if normalized.FromRef == normalized.ToRef {
		return Result{}, fmt.Errorf("from-ref and to-ref must be different")
	}

	fromUnit, _, err := parseStableUnitRef(normalized.FromRef)
	if err != nil {
		return Result{}, err
	}
	toUnit, toVersion, err := parseStableUnitRef(normalized.ToRef)
	if err != nil {
		return Result{}, err
	}
	if fromUnit != normalized.Unit || toUnit != normalized.Unit {
		return Result{}, fmt.Errorf("from-ref and to-ref must both point to unit %q", normalized.Unit)
	}
	if fromUnit != toUnit {
		return Result{}, fmt.Errorf("from-ref and to-ref must refer to the same stable unit")
	}
	if err := validateToRef(repoRoot, normalized.Unit, toVersion, normalized.ToRef); err != nil {
		return Result{}, err
	}

	statuses, err := statusfile.LoadObjectStatuses(repoRoot)
	if err != nil {
		return Result{}, err
	}
	result := Result{
		Unit:    normalized.Unit,
		FromRef: normalized.FromRef,
		ToRef:   normalized.ToRef,
	}

	for _, status := range statuses {
		if status.ObjectType != "unit" {
			continue
		}
		fileRef, err := specpaths.ObjectMainSpecFileRef(status.ObjectType, status.ActiveLayer, status.Object)
		if err != nil {
			return Result{}, err
		}
		contentBytes, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(fileRef)))
		if err != nil {
			return Result{}, fmt.Errorf("read %s: %w", fileRef, err)
		}
		refs, err := unitrefs.ParseObjectUnitRefs(fileRef, string(contentBytes))
		if err != nil {
			return Result{}, err
		}
		if !containsString(refs, normalized.FromRef) {
			continue
		}

		key := status.ObjectType + ":" + status.Object
		switch status.ActiveLayer {
		case "candidate":
			nextRefs, changed, err := unitrefs.ReplaceUnitRef(refs, normalized.FromRef, normalized.ToRef)
			if err != nil {
				return Result{}, fmt.Errorf("%s: %w", fileRef, err)
			}
			if !changed {
				continue
			}
			updated, err := unitrefs.UpdateObjectUnitRefs(fileRef, string(contentBytes), nextRefs)
			if err != nil {
				return Result{}, err
			}
			if err := os.WriteFile(filepath.Join(repoRoot, filepath.FromSlash(fileRef)), []byte(updated), 0o644); err != nil {
				return Result{}, fmt.Errorf("write %s: %w", fileRef, err)
			}
			result.MainSpecsUpdated = append(result.MainSpecsUpdated, fileRef)

			removed, err := removeProcessArtifacts(repoRoot, status.ObjectType, status.Object)
			if err != nil {
				return Result{}, err
			}
			result.ProcessFilesRemoved = append(result.ProcessFilesRemoved, removed...)
			status.NextCommand = "unit_check"
			status.Notes = releaseNote(normalized.FromRef, normalized.ToRef, false)
			if _, err := statusfile.UpsertObjectStatus(repoRoot, status, false); err != nil {
				return Result{}, err
			}
			result.CandidateUpdated = append(result.CandidateUpdated, key)
			result.StatusUpdated = append(result.StatusUpdated, statusUpdateLabel(status.Object, status.NextCommand))
		case "stable":
			removed, err := removeStableVerifyArtifacts(repoRoot, status.ObjectType, status.Object)
			if err != nil {
				return Result{}, err
			}
			result.ProcessFilesRemoved = append(result.ProcessFilesRemoved, removed...)
			status.NextCommand = "unit_stable_verify"
			status.Notes = releaseNote(normalized.FromRef, normalized.ToRef, true)
			if _, err := statusfile.UpsertObjectStatus(repoRoot, status, false); err != nil {
				return Result{}, err
			}
			result.StableRerouted = append(result.StableRerouted, key)
			result.StatusUpdated = append(result.StatusUpdated, statusUpdateLabel(status.Object, status.NextCommand))
		default:
			return Result{}, fmt.Errorf("unit %q has unsupported active layer %q", status.Object, status.ActiveLayer)
		}
	}

	result.CandidateUpdated = normalizeStrings(result.CandidateUpdated)
	result.StableRerouted = normalizeStrings(result.StableRerouted)
	result.MainSpecsUpdated = normalizeStrings(result.MainSpecsUpdated)
	result.ProcessFilesRemoved = normalizeStrings(result.ProcessFilesRemoved)
	result.StatusUpdated = normalizeStrings(result.StatusUpdated)
	result.Noop = len(result.CandidateUpdated) == 0 && len(result.StableRerouted) == 0

	if diagnostics := ValidateCurrentCandidateRefs(repoRoot, normalized.FromRef); len(diagnostics) > 0 {
		return Result{}, fmt.Errorf("post-release current candidate unit ref validation failed: %s", strings.Join(diagnostics, "; "))
	}
	return result, nil
}

func ValidateCurrentCandidateRefs(repoRoot string, forbiddenRef string) []string {
	diagnostics := []string{}
	statuses, err := statusfile.LoadObjectStatuses(repoRoot)
	if err != nil {
		return append(diagnostics, err.Error())
	}
	for _, status := range statuses {
		if status.ObjectType != "unit" {
			continue
		}
		if status.ActiveLayer != "candidate" {
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
		refs, err := unitrefs.ParseObjectUnitRefs(fileRef, string(content))
		if err != nil {
			diagnostics = append(diagnostics, err.Error())
			continue
		}
		for _, ref := range refs {
			if forbiddenRef != "" && ref == forbiddenRef {
				diagnostics = append(diagnostics, fmt.Sprintf("%s: forbidden unit ref %s remains", fileRef, forbiddenRef))
			}
		}
	}
	return normalizeStrings(diagnostics)
}

func validateToRef(repoRoot, unit, version, toRef string) error {
	fileRef, err := specpaths.ObjectMainSpecFileRef("unit", "stable", unit)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(fileRef)))
	if err != nil {
		return fmt.Errorf("read %s: %w", fileRef, err)
	}
	currentVersion, err := unitrefs.FrontmatterScalar(fileRef, string(content), "version")
	if err != nil {
		return err
	}
	if currentVersion != version {
		return fmt.Errorf("to-ref %q does not match current stable version %q", toRef, "s_unit_"+unit+"@"+currentVersion)
	}
	return nil
}

func parseStableUnitRef(ref string) (string, string, error) {
	matches := stableUnitRefPattern.FindStringSubmatch(strings.TrimSpace(ref))
	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid stable unit ref %q", ref)
	}
	return matches[1], matches[2], nil
}

func removeProcessArtifacts(repoRoot, objectType, object string) ([]string, error) {
	kinds := []string{"check_work", "check", "plan", "verify"}
	return removeProcessArtifactsByKind(repoRoot, objectType, object, kinds)
}

func removeStableVerifyArtifacts(repoRoot, objectType, object string) ([]string, error) {
	return removeProcessArtifactsByKind(repoRoot, objectType, object, []string{"stable_verify"})
}

func removeProcessArtifactsByKind(repoRoot, objectType, object string, kinds []string) ([]string, error) {
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

func releaseNote(fromRef, toRef string, stable bool) string {
	if stable {
		return fmt.Sprintf("Rerouted by unit release-version from %s to %s; stable truth still references the prior ref until verified or forked.", fromRef, toRef)
	}
	return fmt.Sprintf("Retargeted by unit release-version from %s to %s; rerun check.", fromRef, toRef)
}

func statusUpdateLabel(object, nextCommand string) string {
	return fmt.Sprintf("unit:%s -> %s", object, nextCommand)
}

func normalizeStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := map[string]bool{}
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		normalized = append(normalized, item)
	}
	sort.Strings(normalized)
	return normalized
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
