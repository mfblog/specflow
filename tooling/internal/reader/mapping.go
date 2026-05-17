package reader

import (
	"os"
	"path/filepath"
	"strings"
)

const registryHeader = "kind|id|registration_state|implementation_paths|spec_files|responsibility"

type repositoryMapping struct {
	Units       map[string]mappingUnit
	Rules       map[string]mappingShared
	GlobalRules map[string]mappingShared
	Registry    map[string]mappingRegistryEntry
	InvalidRows []mappingInvalidRow
	Diagnostics []Diagnostic
}

type mappingRegistryEntry struct {
	Kind                string
	ID                  string
	RegistrationState   string
	ImplementationPaths []SourceRef
	SpecFiles           []SourceRef
	Responsibility      string
	Source              SourceRef
	Invalid             bool
	InvalidReason       string
}

type mappingInvalidRow struct {
	ID      string
	Message string
	Source  SourceRef
}

type mappingUnit struct {
	ID                  string
	Responsibility      string
	TruthSurfaceRule    string
	ImplementationPaths []SourceRef
	Source              SourceRef
}

type mappingShared struct {
	ID             string
	Responsibility string
	TruthPaths     []SourceRef
	Source         SourceRef
}

func loadRepositoryMapping(repoRoot string) repositoryMapping {
	relPath := "docs/specs/repository_mapping.md"
	path := filepath.Join(repoRoot, filepath.FromSlash(relPath))
	data, err := os.ReadFile(path)
	result := repositoryMapping{
		Units:       map[string]mappingUnit{},
		Rules:       map[string]mappingShared{},
		GlobalRules: map[string]mappingShared{},
		Registry:    map[string]mappingRegistryEntry{},
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
	inRegistrySection := false
	headerSeen := false
	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "## 2. Object Registry"):
			inRegistrySection = true
			headerSeen = false
			continue
		case inRegistrySection && strings.HasPrefix(trimmed, "## ") && !strings.HasPrefix(trimmed, "## 2. Object Registry"):
			inRegistrySection = false
			continue
		}
		if !inRegistrySection || !strings.HasPrefix(trimmed, "|") {
			continue
		}
		cells := splitMarkdownTableRow(trimmed)
		if len(cells) == 0 || isMarkdownSeparatorRow(cells) {
			continue
		}
		if !headerSeen {
			if normalizeHeader(cells) != registryHeader {
				result.InvalidRows = append(result.InvalidRows, mappingInvalidRow{
					ID:      "header",
					Message: "Object Registry header must be: | kind | id | registration_state | implementation_paths | spec_files | responsibility |",
					Source:  SourceRef{Path: relPath, Line: idx + 1, Label: "Object Registry"},
				})
				result.Diagnostics = append(result.Diagnostics, Diagnostic{
					Severity: "error",
					Message:  "invalid Object Registry header",
					Source:   &SourceRef{Path: relPath, Line: idx + 1, Label: "Object Registry"},
				})
				return result
			}
			headerSeen = true
			continue
		}
		entry := parseRegistryEntry(cells, SourceRef{Path: relPath, Line: idx + 1, Label: "Object Registry"})
		key := entry.Kind + ":" + entry.ID
		if entry.Invalid {
			if key == ":" {
				key = "invalid:" + entry.Source.Path + ":" + stringLine(entry.Source.Line)
			}
			result.InvalidRows = append(result.InvalidRows, mappingInvalidRow{ID: key, Message: entry.InvalidReason, Source: entry.Source})
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: "error",
				Message:  entry.InvalidReason,
				Source:   &entry.Source,
			})
			continue
		}
		result.Registry[key] = entry
		applyRegistryEntry(&result, entry)
	}
	if !headerSeen {
		ref := SourceRef{Path: relPath, Label: "Object Registry"}
		result.InvalidRows = append(result.InvalidRows, mappingInvalidRow{
			ID:      "missing_object_registry",
			Message: "repository_mapping.md must contain ## 2. Object Registry with the fixed registry table",
			Source:  ref,
		})
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Severity: "error",
			Message:  "repository_mapping.md is missing Object Registry table",
			Source:   &ref,
		})
	}
	return result
}

func parseRegistryEntry(cells []string, source SourceRef) mappingRegistryEntry {
	entry := mappingRegistryEntry{Source: source}
	if len(cells) != 6 {
		entry.Invalid = true
		entry.InvalidReason = "invalid Object Registry row: expected 6 columns"
		return entry
	}
	entry.Kind = cleanRegistryCell(cells[0])
	entry.ID = cleanRegistryCell(cells[1])
	entry.RegistrationState = cleanRegistryCell(cells[2])
	entry.ImplementationPaths = parseRegistryPathList(cells[3], source)
	entry.SpecFiles = parseRegistryPathList(cells[4], source)
	entry.Responsibility = cleanRegistryCell(cells[5])

	if entry.Kind != "unit" && entry.Kind != "rule" {
		entry.Invalid = true
		entry.InvalidReason = "invalid Object Registry row: kind must be unit or rule"
		return entry
	}
	if entry.ID == "" {
		entry.Invalid = true
		entry.InvalidReason = "invalid Object Registry row: id is required"
		return entry
	}
	if entry.RegistrationState != "planned" && entry.RegistrationState != "landed" {
		entry.Invalid = true
		entry.InvalidReason = "invalid Object Registry row: registration_state must be planned or landed"
		return entry
	}
	if entry.RegistrationState == "planned" && len(entry.ImplementationPaths) != 0 {
		entry.Invalid = true
		entry.InvalidReason = "invalid Object Registry row: planned objects must use implementation_paths=none"
		return entry
	}
	if entry.RegistrationState == "landed" && len(entry.ImplementationPaths) == 0 {
		entry.Invalid = true
		entry.InvalidReason = "invalid Object Registry row: landed objects must declare implementation_paths"
		return entry
	}
	return entry
}

func applyRegistryEntry(result *repositoryMapping, entry mappingRegistryEntry) {
	switch entry.Kind {
	case "unit":
		result.Units[entry.ID] = mappingUnit{
			ID:                  entry.ID,
			Responsibility:      entry.Responsibility,
			TruthSurfaceRule:    "object_registry",
			ImplementationPaths: entry.ImplementationPaths,
			Source:              entry.Source,
		}
	case "rule":
		shared := mappingShared{
			ID:             entry.ID,
			Responsibility: entry.Responsibility,
			TruthPaths:     entry.SpecFiles,
			Source:         entry.Source,
		}
		if inferredRuleScope(entry.ID, "") == "global" {
			result.GlobalRules[entry.ID] = shared
		} else {
			result.Rules[entry.ID] = shared
		}
	}
}

func splitMarkdownTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	raw := strings.Split(line, "|")
	cells := make([]string, 0, len(raw))
	for _, cell := range raw {
		cells = append(cells, strings.TrimSpace(cell))
	}
	return cells
}

func normalizeHeader(cells []string) string {
	normalized := make([]string, 0, len(cells))
	for _, cell := range cells {
		normalized = append(normalized, cleanRegistryCell(cell))
	}
	return strings.Join(normalized, "|")
}

func isMarkdownSeparatorRow(cells []string) bool {
	if len(cells) == 0 {
		return false
	}
	for _, cell := range cells {
		cell = strings.TrimSpace(cell)
		if cell == "" {
			return false
		}
		for _, char := range cell {
			if char != '-' && char != ':' {
				return false
			}
		}
	}
	return true
}

func parseRegistryPathList(cell string, source SourceRef) []SourceRef {
	cell = cleanRegistryCell(cell)
	if cell == "" || cell == "none" {
		return nil
	}
	parts := strings.Split(cell, ";")
	refs := []SourceRef{}
	for _, part := range parts {
		path := cleanRegistryCell(part)
		if path == "" || path == "none" {
			continue
		}
		refs = appendSourceUnique(refs, SourceRef{Path: filepath.ToSlash(path), Line: source.Line, Label: source.Label})
	}
	return refs
}

func cleanRegistryCell(cell string) string {
	cell = strings.TrimSpace(cell)
	cell = strings.Trim(cell, "` ")
	return strings.TrimSpace(cell)
}

func stringLine(line int) string {
	if line == 0 {
		return "0"
	}
	digits := []byte{}
	for line > 0 {
		digits = append([]byte{byte('0' + line%10)}, digits...)
		line = line / 10
	}
	return string(digits)
}
