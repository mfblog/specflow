package reader

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type repositoryMapping struct {
	Units           map[string]mappingUnit
	SharedContracts map[string]mappingShared
	Diagnostics     []Diagnostic
}

type mappingUnit struct {
	ID                  string
	Responsibility      string
	TruthPaths          []SourceRef
	ImplementationPaths []SourceRef
}

type mappingShared struct {
	ID             string
	Responsibility string
	TruthPaths     []SourceRef
}

var numberedCodeSpan = regexp.MustCompile("^\\s*\\d+\\.\\s+`([^`]+)`\\s*$")

func loadRepositoryMapping(repoRoot string) repositoryMapping {
	relPath := "docs/specs/repository_mapping.md"
	path := filepath.Join(repoRoot, filepath.FromSlash(relPath))
	data, err := os.ReadFile(path)
	result := repositoryMapping{
		Units:           map[string]mappingUnit{},
		SharedContracts: map[string]mappingShared{},
	}
	if err != nil {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Severity: "error",
			Message:  "cannot read repository mapping: " + err.Error(),
			Source:   &SourceRef{Path: relPath},
		})
		return result
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	section := ""
	currentID := ""
	currentPathKind := ""
	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "### 2.1 "):
			section = "unit_map"
			currentID = ""
			continue
		case strings.HasPrefix(trimmed, "### 2.2 "):
			section = ""
			currentID = ""
			continue
		case strings.HasPrefix(trimmed, "### 2.3 "):
			section = "shared_map"
			currentID = ""
			continue
		case strings.HasPrefix(trimmed, "## 3.") || strings.HasPrefix(trimmed, "### 4.1 "):
			section = ""
			currentID = ""
			continue
		case strings.HasPrefix(trimmed, "### 4.5 "):
			section = "shared_paths"
			currentID = ""
			currentPathKind = ""
			continue
		case strings.HasPrefix(trimmed, "### 4.6 "):
			section = "unit_paths"
			currentID = ""
			currentPathKind = ""
			continue
		case strings.HasPrefix(trimmed, "### 4.7 ") || strings.HasPrefix(trimmed, "## 5."):
			section = ""
			currentID = ""
			currentPathKind = ""
			continue
		}

		switch section {
		case "unit_map":
			if id, ok := parseNumberedCodeSpan(trimmed); ok {
				currentID = id
				unit := result.Units[id]
				unit.ID = id
				result.Units[id] = unit
				continue
			}
			if currentID != "" && strings.HasPrefix(trimmed, "- ") {
				unit := result.Units[currentID]
				unit.Responsibility = strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
				result.Units[currentID] = unit
			}
		case "shared_map":
			if id, ok := parseNumberedCodeSpan(trimmed); ok {
				currentID = normalizedSharedID(id)
				shared := result.SharedContracts[currentID]
				shared.ID = currentID
				result.SharedContracts[currentID] = shared
				continue
			}
			if currentID != "" && strings.HasPrefix(trimmed, "- ") {
				shared := result.SharedContracts[currentID]
				shared.Responsibility = strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
				result.SharedContracts[currentID] = shared
			}
		case "shared_paths":
			if id, ok := parseNumberedCodeSpan(trimmed); ok {
				currentID = normalizedSharedID(id)
				shared := result.SharedContracts[currentID]
				shared.ID = currentID
				result.SharedContracts[currentID] = shared
				continue
			}
			if currentID != "" && strings.HasPrefix(trimmed, "- `") {
				if pathRef, ok := parseListCodePath(trimmed, relPath, idx+1); ok {
					shared := result.SharedContracts[currentID]
					shared.TruthPaths = append(shared.TruthPaths, pathRef)
					result.SharedContracts[currentID] = shared
				}
			}
		case "unit_paths":
			if id, ok := parseNumberedCodeSpan(trimmed); ok {
				currentID = id
				currentPathKind = ""
				unit := result.Units[id]
				unit.ID = id
				result.Units[id] = unit
				continue
			}
			if currentID == "" {
				continue
			}
			switch {
			case strings.Contains(trimmed, "`truth_surface`"):
				currentPathKind = "truth"
				continue
			case strings.Contains(trimmed, "`implementation_surface`"):
				currentPathKind = "implementation"
				continue
			case strings.HasPrefix(trimmed, "- `"):
				pathRef, ok := parseListCodePath(trimmed, relPath, idx+1)
				if !ok {
					continue
				}
				unit := result.Units[currentID]
				if currentPathKind == "truth" {
					unit.TruthPaths = append(unit.TruthPaths, pathRef)
				}
				if currentPathKind == "implementation" {
					unit.ImplementationPaths = append(unit.ImplementationPaths, pathRef)
				}
				result.Units[currentID] = unit
			}
		}
	}
	return result
}

func parseNumberedCodeSpan(line string) (string, bool) {
	matches := numberedCodeSpan.FindStringSubmatch(line)
	if len(matches) != 2 {
		return "", false
	}
	return strings.TrimSpace(matches[1]), true
}

func parseListCodePath(line, sourcePath string, sourceLine int) (SourceRef, bool) {
	value := extractFirstCodeSpan(line)
	if value == "" {
		return SourceRef{}, false
	}
	return SourceRef{Path: filepath.ToSlash(value), Line: sourceLine, Label: sourcePath}, true
}

func extractFirstCodeSpan(line string) string {
	start := strings.Index(line, "`")
	if start < 0 {
		return ""
	}
	end := strings.Index(line[start+1:], "`")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(line[start+1 : start+1+end])
}

func normalizedSharedID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" || strings.HasPrefix(id, "shared_") {
		return id
	}
	return "shared_" + id
}
