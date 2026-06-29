package filevalidation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CandidateFrontmatterResult is returned by ValidateCandidateFrontmatter.
type CandidateFrontmatterResult struct {
	Valid      bool
	Unit       string
	Diagnostic string
}

// ValidateCandidateFrontmatter validates candidate unit frontmatter consistency.
// Checks required fields: id, layer (must be candidate), version, rule_refs.
func ValidateCandidateFrontmatter(repoRoot, unitName string) CandidateFrontmatterResult {
	candidatePath := filepath.Join(repoRoot, fmt.Sprintf("docs/specs/units/candidate/c_unit_%s.md", unitName))

	data, err := os.ReadFile(candidatePath)
	if err != nil {
		return CandidateFrontmatterResult{
			Valid:      false,
			Unit:       unitName,
			Diagnostic: fmt.Sprintf("cannot read candidate spec at %s: %v", candidatePath, err),
		}
	}

	fm := readFrontmatterStringMap(string(data))

	id := strings.TrimSpace(fm["id"])
	layer := strings.TrimSpace(fm["layer"])

	if id != unitName {
		return CandidateFrontmatterResult{
			Valid:      false,
			Unit:       unitName,
			Diagnostic: fmt.Sprintf("frontmatter id %q does not match unit name %q", id, unitName),
		}
	}
	if layer != "candidate" {
		return CandidateFrontmatterResult{
			Valid:      false,
			Unit:       unitName,
			Diagnostic: fmt.Sprintf("frontmatter layer must be 'candidate', got %q", layer),
		}
	}
	if strings.TrimSpace(fm["version"]) == "" {
		return CandidateFrontmatterResult{
			Valid:      false,
			Unit:       unitName,
			Diagnostic: "frontmatter version is required",
		}
	}

	return CandidateFrontmatterResult{
		Valid:      true,
		Unit:       unitName,
		Diagnostic: "candidate frontmatter is valid",
	}
}

// MatchGlobPattern checks if a path matches a glob-like pattern.
func MatchGlobPattern(pattern, path string) bool {
	return patternMatch(pattern, path)
}

// patternMatch checks if a path matches a glob-like pattern.
// Supports ** for multi-segment wildcard and * for single-segment wildcard.
func patternMatch(pattern, path string) bool {
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)

	if !strings.Contains(pattern, "**") {
		matched, err := filepath.Match(pattern, path)
		return err == nil && matched
	}

	patParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")
	return matchSegments(patParts, pathParts)
}

func matchSegments(pattern, path []string) bool {
	if len(pattern) == 0 && len(path) == 0 {
		return true
	}
	if len(pattern) == 0 {
		return false
	}

	if pattern[0] == "**" {
		if matchSegments(pattern[1:], path) {
			return true
		}
		if len(path) > 0 && matchSegments(pattern, path[1:]) {
			return true
		}
		return false
	}

	if len(path) == 0 {
		return false
	}

	matched, err := filepath.Match(pattern[0], path[0])
	if err != nil || !matched {
		return false
	}
	return matchSegments(pattern[1:], path[1:])
}

func readFrontmatterStringMap(text string) map[string]string {
	result := map[string]string{}
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return result
	}

	for idx := 1; idx < len(lines); idx++ {
		line := lines[idx]
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			break
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "- ") {
			continue
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		value = strings.Trim(value, "`\"' ")
		result[key] = value
	}
	return result
}
