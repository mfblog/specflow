package repositorymapping

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

const unitDefaultRule = "unit_default"

type Result struct {
	Diagnostics []string
}

func (r Result) Valid() bool {
	return len(r.Diagnostics) == 0
}

type mappingData struct {
	Units       map[string]mappingUnit
	SharedPaths []mappingPath
	Diagnostics []string
}

type mappingUnit struct {
	ID      string
	Rule    string
	RuleSet bool
}

type mappingPath struct {
	Path string
	Line int
}

var numberedCodeSpanPattern = regexp.MustCompile("^\\s*\\d+\\.\\s+`([^`]+)`\\s*$")

func Validate(repoRoot string) (Result, error) {
	mapping, err := loadMapping(repoRoot)
	if err != nil {
		return Result{}, err
	}
	result := Result{Diagnostics: append([]string(nil), mapping.Diagnostics...)}

	statuses, err := statusfile.LoadObjectStatuses(repoRoot)
	if err != nil {
		return Result{}, err
	}

	for _, status := range statuses {
		if status.ObjectType != "unit" && status.ObjectType != "scenario" {
			continue
		}
		if status.ObjectType == "unit" {
			unit, ok := mapping.Units[status.Object]
			if !ok {
				result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("unit %q is present in _status.md but missing from repository_mapping.md", status.Object))
			} else {
				if !unit.RuleSet {
					result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("unit %q must declare truth_surface_rule: %s", status.Object, unitDefaultRule))
				} else if unit.Rule != unitDefaultRule {
					result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("unit %q uses unsupported truth_surface_rule %q", status.Object, unit.Rule))
				}
			}
		}
		validateStatusTruthFiles(&result, repoRoot, status)
	}

	for _, sharedPath := range mapping.SharedPaths {
		if !exists(repoRoot, sharedPath.Path) {
			result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("shared contract path does not exist: %s", sharedPath.Path))
		}
	}

	return result, nil
}

func validateStatusTruthFiles(result *Result, repoRoot string, status statusfile.ObjectStatus) {
	if status.ActiveLayer != "stable" && status.ActiveLayer != "candidate" {
		result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("%s %q has unsupported Active Layer %q", status.ObjectType, status.Object, status.ActiveLayer))
		return
	}

	if status.Stable == "yes" {
		validateLayerFile(result, repoRoot, status.ObjectType, status.Object, "stable")
	}
	if status.Candidate == "yes" {
		validateLayerFile(result, repoRoot, status.ObjectType, status.Object, "candidate")
	}
	validateLayerFile(result, repoRoot, status.ObjectType, status.Object, status.ActiveLayer)
}

func validateLayerFile(result *Result, repoRoot, objectType, object, layer string) {
	path, err := specpaths.ObjectMainSpecFileRef(objectType, layer, object)
	if err != nil {
		result.Diagnostics = append(result.Diagnostics, err.Error())
		return
	}
	if !exists(repoRoot, path) {
		result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("%s %q %s truth file does not exist: %s", objectType, object, layer, path))
	}
}

func loadMapping(repoRoot string) (mappingData, error) {
	relPath := specpaths.RepositoryMappingFileRef
	data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relPath)))
	if err != nil {
		return mappingData{}, fmt.Errorf("read %s: %w", relPath, err)
	}

	result := mappingData{Units: map[string]mappingUnit{}}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	section := ""
	currentUnit := ""
	for idx, line := range lines {
		lineNo := idx + 1
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "### 2.1 "):
			section = "unit_map"
			currentUnit = ""
			continue
		case strings.HasPrefix(trimmed, "### 2.2 "):
			section = ""
			currentUnit = ""
			continue
		case strings.HasPrefix(trimmed, "### 4.5 "):
			section = "shared_paths"
			currentUnit = ""
			continue
		case strings.HasPrefix(trimmed, "### 4.6 "):
			section = "unit_paths"
			currentUnit = ""
			continue
		case strings.HasPrefix(trimmed, "### 4.7 ") || strings.HasPrefix(trimmed, "## 5."):
			section = ""
			currentUnit = ""
			continue
		}

		switch section {
		case "unit_map":
			if id, ok := parseNumberedCodeSpan(trimmed); ok {
				unit := result.Units[id]
				unit.ID = id
				result.Units[id] = unit
			}
		case "shared_paths":
			if strings.HasPrefix(trimmed, "- `") {
				if path := firstCodeSpan(trimmed); strings.HasPrefix(path, "docs/specs/shared_contracts/") {
					result.SharedPaths = append(result.SharedPaths, mappingPath{Path: path, Line: lineNo})
				}
			}
		case "unit_paths":
			if id, ok := parseNumberedCodeSpan(trimmed); ok {
				if _, known := result.Units[id]; known {
					currentUnit = id
				} else {
					currentUnit = ""
				}
				continue
			}
			if currentUnit == "" {
				continue
			}
			if strings.Contains(trimmed, "`truth_surface`") && !strings.Contains(trimmed, "truth_surface_rule") {
				result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("%s:%d uses deprecated truth_surface; use truth_surface_rule: %s", relPath, lineNo, unitDefaultRule))
			}
			if strings.Contains(trimmed, "docs/specs/units/candidate/") || strings.Contains(trimmed, "docs/specs/units/stable/") {
				result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("%s:%d lists a concrete unit truth path; use truth_surface_rule: %s", relPath, lineNo, unitDefaultRule))
			}
			if strings.Contains(trimmed, "truth_surface_rule") {
				unit := result.Units[currentUnit]
				unit.Rule = parseRuleValue(trimmed)
				unit.RuleSet = unit.Rule != ""
				result.Units[currentUnit] = unit
			}
		}
	}
	return result, nil
}

func exists(repoRoot, relPath string) bool {
	_, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(relPath)))
	return err == nil
}

func parseNumberedCodeSpan(line string) (string, bool) {
	matches := numberedCodeSpanPattern.FindStringSubmatch(line)
	if len(matches) != 2 {
		return "", false
	}
	return strings.TrimSpace(matches[1]), true
}

func parseRuleValue(line string) string {
	spans := codeSpans(line)
	if len(spans) >= 2 {
		return spans[1]
	}
	if idx := strings.Index(line, ":"); idx >= 0 {
		return strings.Trim(strings.TrimSpace(line[idx+1:]), "` ")
	}
	return ""
}

func firstCodeSpan(line string) string {
	spans := codeSpans(line)
	if len(spans) == 0 {
		return ""
	}
	return spans[0]
}

func codeSpans(line string) []string {
	spans := []string{}
	rest := line
	for {
		start := strings.Index(rest, "`")
		if start < 0 {
			return spans
		}
		rest = rest[start+1:]
		end := strings.Index(rest, "`")
		if end < 0 {
			return spans
		}
		spans = append(spans, strings.TrimSpace(rest[:end]))
		rest = rest[end+1:]
	}
}
