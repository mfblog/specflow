package filevalidation

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Constraints defines file write permissions for a unit at a given lifecycle phase.
type Constraints struct {
	AllowedWrites   []WriteRule
	ForbiddenWrites []WriteRule
}

// WriteRule defines a glob pattern permission for specific lifecycle phases.
type WriteRule struct {
	Pattern string
	Phases  []string // empty = applies to all phases
}

// Result is returned by ValidateWrite.
type Result struct {
	Allowed bool
	Reason  string
	Phase   string
	Path    string
}

// ValidateWrite checks whether a file write at a given lifecycle phase is allowed.
func ValidateWrite(phase string, path string, constraints Constraints) Result {
	// Normalize path: clean and use forward slashes
	normalizedPath := filepath.ToSlash(filepath.Clean(path))

	// Check forbidden rules first (forbidden takes precedence)
	for _, rule := range constraints.ForbiddenWrites {
		if !phaseMatches(phase, rule.Phases) {
			continue
		}
		if patternMatch(rule.Pattern, normalizedPath) {
			return Result{
				Allowed: false,
				Reason:  fmt.Sprintf("path %q matches forbidden pattern %q", path, rule.Pattern),
				Phase:   phase,
				Path:    path,
			}
		}
	}

	// Check allowed rules: if no allowed rules are defined, allow all
	if len(constraints.AllowedWrites) == 0 {
		return Result{
			Allowed: true,
			Reason:  "no allowed_writes constraints defined; write permitted by default",
			Phase:   phase,
			Path:    path,
		}
	}

	// If there are allowed rules, at least one must match
	for _, rule := range constraints.AllowedWrites {
		if !phaseMatches(phase, rule.Phases) {
			continue
		}
		if patternMatch(rule.Pattern, normalizedPath) {
			return Result{
				Allowed: true,
				Reason:  fmt.Sprintf("path %q matches allowed pattern %q in phase %s", path, rule.Pattern, phase),
				Phase:   phase,
				Path:    path,
			}
		}
	}

	return Result{
		Allowed: false,
		Reason:  fmt.Sprintf("path %q does not match any allowed_writes pattern for phase %q", path, phase),
		Phase:   phase,
		Path:    path,
	}
}

// patternMatch checks if a path matches a glob-like pattern.
// Supports ** for multi-segment wildcard and * for single-segment wildcard.
func patternMatch(pattern, path string) bool {
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)

	// Patterns without ** use filepath.Match
	if !strings.Contains(pattern, "**") {
		matched, err := filepath.Match(pattern, path)
		return err == nil && matched
	}

	// With **, use segment-based recursive matching
	patParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")
	return matchSegments(patParts, pathParts)
}

// matchSegments recursively matches pattern segments against path segments.
// ** matches zero or more segments; * matches one segment.
func matchSegments(pattern, path []string) bool {
	// Both exhausted → match
	if len(pattern) == 0 && len(path) == 0 {
		return true
	}
	// Pattern exhausted but path remains → no match
	if len(pattern) == 0 {
		return false
	}

	// Handle **
	if pattern[0] == "**" {
		// Try matching ** as zero segments (skip **)
		if matchSegments(pattern[1:], path) {
			return true
		}
		// Try matching ** as one or more segments
		if len(path) > 0 && matchSegments(pattern, path[1:]) {
			return true
		}
		return false
	}

	// Path exhausted but pattern still has non-** segments → no match
	if len(path) == 0 {
		return false
	}

	// Match current segment
	matched, err := filepath.Match(pattern[0], path[0])
	if err != nil || !matched {
		return false
	}
	return matchSegments(pattern[1:], path[1:])
}

// phaseMatches checks if a phase is in the allowed phases list.
// An empty phases list means "all phases".
func phaseMatches(phase string, phases []string) bool {
	if len(phases) == 0 {
		return true
	}
	for _, p := range phases {
		if strings.EqualFold(p, phase) {
			return true
		}
	}
	return false
}

// ParseConstraints parses a constraints string from _status.md.
func ParseConstraints(constraintsStr string) (Constraints, error) {
	var c Constraints

	lines := strings.Split(constraintsStr, "\n")
	var currentSection string
	var lastRuleIndex int

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		switch {
		case trimmed == "allowed_writes:":
			currentSection = "allowed"
			continue
		case trimmed == "forbidden_writes:":
			currentSection = "forbidden"
			continue
		}

		if currentSection == "" {
			continue
		}

		// Parse list items: "- pattern: src/**"
		if strings.HasPrefix(trimmed, "- pattern:") {
			pattern := strings.TrimSpace(strings.TrimPrefix(trimmed, "- pattern:"))
			pattern = strings.Trim(pattern, "\"'")
			rule := WriteRule{Pattern: pattern}

			switch currentSection {
			case "allowed":
				c.AllowedWrites = append(c.AllowedWrites, rule)
				lastRuleIndex = len(c.AllowedWrites) - 1
			case "forbidden":
				c.ForbiddenWrites = append(c.ForbiddenWrites, rule)
				lastRuleIndex = len(c.ForbiddenWrites) - 1
			}
			continue
		}

		// Parse phases sub-field: "  phases: [unit_impl, unit_verify]"
		if strings.HasPrefix(trimmed, "phases:") {
			phasesStr := strings.TrimSpace(strings.TrimPrefix(trimmed, "phases:"))
			phasesStr = strings.Trim(phasesStr, "[]")
			rawPhases := strings.Split(phasesStr, ",")
			var cleanPhases []string
			for _, p := range rawPhases {
				p = strings.TrimSpace(p)
				p = strings.Trim(p, "\"' ")
				if p != "" {
					cleanPhases = append(cleanPhases, p)
				}
			}

			// Add phases to the last added rule
			switch currentSection {
			case "allowed":
				if len(c.AllowedWrites) > 0 {
					c.AllowedWrites[len(c.AllowedWrites)-1].Phases = cleanPhases
				}
			case "forbidden":
				if len(c.ForbiddenWrites) > 0 {
					c.ForbiddenWrites[len(c.ForbiddenWrites)-1].Phases = cleanPhases
				}
			}
			_ = lastRuleIndex
		}
	}

	return c, nil
}
