package unitappendix

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
)

type Entry struct {
	ObjectType string
	Object     string
	Layer      string
	Name       string
	FileRef    string
	Content    string
}

func Scan(repoRoot, objectType, object, layer string) ([]Entry, error) {
	if objectType != "unit" {
		return nil, fmt.Errorf("unsupported object type %q", objectType)
	}
	object = strings.TrimSpace(object)
	if object == "" {
		return nil, fmt.Errorf("object is required")
	}
	dir, prefix, err := unitAppendixPathParts(layer, object)
	if err != nil {
		return nil, err
	}
	pattern := filepath.Join(repoRoot, filepath.FromSlash(dir), prefix+"*.md")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	entries := make([]Entry, 0, len(matches))
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			continue
		}
		relPath, err := filepath.Rel(repoRoot, match)
		if err != nil {
			return nil, err
		}
		fileRef := filepath.ToSlash(relPath)
		name := appendixName(fileRef, prefix)
		if name == "" {
			continue
		}
		contentBytes, err := os.ReadFile(match)
		if err != nil {
			return nil, fmt.Errorf("read appendix %s: %w", fileRef, err)
		}
		content := string(contentBytes)
		frontmatter := parseFrontmatter(content)
		if err := validateAppendixFrontmatter(fileRef, frontmatter, object, layer); err != nil {
			return nil, err
		}
		entries = append(entries, Entry{
			ObjectType: objectType,
			Object:     object,
			Layer:      layer,
			Name:       name,
			FileRef:    fileRef,
			Content:    content,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].FileRef < entries[j].FileRef
	})
	return entries, nil
}

func ValidateCandidateCoverage(repoRoot, objectType, object string) error {
	mismatches, err := CandidateCoverageMismatches(repoRoot, objectType, object)
	if err != nil {
		return err
	}
	if len(mismatches) > 0 {
		return errors.New(strings.Join(mismatches, "; "))
	}
	return nil
}

func CandidateCoverageMismatches(repoRoot, objectType, object string) ([]string, error) {
	return CandidateCoverageMismatchesWithExclusions(repoRoot, objectType, object, nil)
}

func CandidateCoverageMismatchesWithExclusions(repoRoot, objectType, object string, excludedStableRefs []string) ([]string, error) {
	stableEntries, err := Scan(repoRoot, objectType, object, "stable")
	if err != nil {
		return nil, err
	}
	candidateEntries, err := Scan(repoRoot, objectType, object, "candidate")
	if err != nil {
		return nil, err
	}
	candidateNames := map[string]bool{}
	for _, entry := range candidateEntries {
		candidateNames[entry.Name] = true
	}
	excluded := map[string]bool{}
	for _, ref := range excludedStableRefs {
		excluded[ref] = true
	}
	missing := []string{}
	for _, entry := range stableEntries {
		if candidateNames[entry.Name] {
			continue
		}
		if excluded[entry.FileRef] {
			continue
		}
		missing = append(missing, fmt.Sprintf("%s -> %s", entry.FileRef, CandidateFileRef(objectType, object, entry.Name)))
	}
	sort.Strings(missing)
	if len(missing) == 0 {
		return nil, nil
	}
	return []string{"unit_appendix_snapshot mismatch: missing candidate appendix for stable appendix: " + strings.Join(missing, ", ")}, nil
}



func CandidateFileRef(objectType, object, name string) string {
	if objectType != "unit" {
		return ""
	}
	return fmt.Sprintf("%s/c_unit_%s_%s.md", specpaths.CandidateAppendixDir, object, name)
}

func StableFileRef(objectType, object, name string) string {
	if objectType != "unit" {
		return ""
	}
	return fmt.Sprintf("%s/s_unit_%s_%s.md", specpaths.StableAppendixDir, object, name)
}

func CandidateCounterpartForStable(objectType, object, stableFileRef string) (string, bool) {
	if objectType != "unit" {
		return "", false
	}
	prefix := fmt.Sprintf("%s/s_unit_%s_", specpaths.StableAppendixDir, object)
	if !strings.HasPrefix(stableFileRef, prefix) || !strings.HasSuffix(stableFileRef, ".md") {
		return "", false
	}
	name := strings.TrimSuffix(strings.TrimPrefix(stableFileRef, prefix), ".md")
	if name == "" {
		return "", false
	}
	return CandidateFileRef(objectType, object, name), true
}

func StableCounterpartForCandidatePath(candidateFileRef string) (string, bool) {
	const candidatePrefix = "docs/specs/units/candidate/appendix/c_unit_"
	if !strings.HasPrefix(candidateFileRef, candidatePrefix) || !strings.HasSuffix(candidateFileRef, ".md") {
		return "", false
	}
	return "docs/specs/units/stable/appendix/s_unit_" + strings.TrimPrefix(candidateFileRef, candidatePrefix), true
}

func unitAppendixPathParts(layer, object string) (string, string, error) {
	switch layer {
	case "candidate":
		return specpaths.CandidateAppendixDir, fmt.Sprintf("c_unit_%s_", object), nil
	case "stable":
		return specpaths.StableAppendixDir, fmt.Sprintf("s_unit_%s_", object), nil
	default:
		return "", "", fmt.Errorf("unsupported layer %q", layer)
	}
}

func appendixName(fileRef, prefix string) string {
	base := strings.TrimSuffix(filepath.Base(fileRef), ".md")
	if !strings.HasPrefix(base, prefix) {
		return ""
	}
	return strings.TrimPrefix(base, prefix)
}

func validateAppendixFrontmatter(fileRef string, frontmatter map[string]string, object, layer string) error {
	unitValue, ok := frontmatter["unit"]
	if !ok || strings.TrimSpace(unitValue) == "" {
		return fmt.Errorf("%s: missing frontmatter.unit", fileRef)
	}
	if strings.TrimSpace(unitValue) != object {
		return fmt.Errorf("%s: frontmatter.unit mismatch: actual=%s expected=%s", fileRef, strings.TrimSpace(unitValue), object)
	}
	layerValue, ok := frontmatter["layer"]
	if !ok || strings.TrimSpace(layerValue) == "" {
		return fmt.Errorf("%s: missing frontmatter.layer", fileRef)
	}
	if strings.TrimSpace(layerValue) != layer {
		return fmt.Errorf("%s: frontmatter.layer mismatch: actual=%s expected=%s", fileRef, strings.TrimSpace(layerValue), layer)
	}
	return nil
}

func parseFrontmatter(content string) map[string]string {
	result := map[string]string{}
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return result
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return result
	}
	for _, line := range lines[1:end] {
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		result[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return result
}
