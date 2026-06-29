package repositorymapping

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
)

const registryHeader = "kind|id|registration_state|implementation_paths|spec_files|responsibility"

type Result struct {
	Diagnostics []string
}

func (r Result) Valid() bool {
	return len(r.Diagnostics) == 0
}

type registryEntry struct {
	Kind                string
	ID                  string
	RegistrationState   string
	ImplementationPaths []string
	SpecFiles           []string
	Line                int
}

type mappingData struct {
	Entries     map[string]registryEntry
	Diagnostics []string
}

func Validate(repoRoot string) (Result, error) {
	mapping, err := loadMapping(repoRoot)
	if err != nil {
		return Result{}, err
	}
	result := Result{Diagnostics: append([]string(nil), mapping.Diagnostics...)}

	for _, entry := range mapping.Entries {
		switch entry.RegistrationState {
		case "planned":
			if len(entry.ImplementationPaths) > 0 {
				result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("planned object %s %q must use implementation_paths=none", entry.Kind, entry.ID))
			}
		case "landed":
			if len(entry.ImplementationPaths) == 0 {
				result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("landed object %s %q must declare implementation_paths", entry.Kind, entry.ID))
			}
			for _, implementationPath := range entry.ImplementationPaths {
				if !implementationPathExists(repoRoot, implementationPath) {
					result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("registered implementation path does not exist: %s", implementationPath))
				}
			}
		}
		for _, specFile := range entry.SpecFiles {
			if !exists(repoRoot, specFile) {
				result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("registered spec file does not exist: %s", specFile))
			}
		}
		// Validate that landed units have at least one spec file on the filesystem
		if entry.RegistrationState == "landed" && entry.Kind == "unit" {
			stablePath, _ := specpaths.ObjectMainSpecFileRef("unit", "stable", entry.ID)
			candidatePath, _ := specpaths.ObjectMainSpecFileRef("unit", "candidate", entry.ID)
			if !exists(repoRoot, stablePath) && !exists(repoRoot, candidatePath) {
				result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("landed %s %q has no spec file (checked stable: %s, candidate: %s)", entry.Kind, entry.ID, stablePath, candidatePath))
			}
		}
	}

	return result, nil
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

	result := mappingData{Entries: map[string]registryEntry{}}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	inRegistrySection := false
	headerSeen := false
	for idx, line := range lines {
		lineNo := idx + 1
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
				result.Diagnostics = append(result.Diagnostics, "Object Registry header must be: | kind | id | registration_state | implementation_paths | spec_files | responsibility |")
				return result, nil
			}
			headerSeen = true
			continue
		}
		entry, diagnostics := parseRegistryEntry(cells, lineNo)
		result.Diagnostics = append(result.Diagnostics, diagnostics...)
		if len(diagnostics) == 0 {
			result.Entries[entry.Kind+":"+entry.ID] = entry
		}
	}
	if !headerSeen {
		result.Diagnostics = append(result.Diagnostics, "repository_mapping.md must contain ## 2. Object Registry with the fixed registry table")
	}
	return result, nil
}

func parseRegistryEntry(cells []string, line int) (registryEntry, []string) {
	entry := registryEntry{Line: line}
	if len(cells) != 6 {
		return entry, []string{fmt.Sprintf("Object Registry row %d must have 6 columns", line)}
	}
	entry.Kind = cleanCell(cells[0])
	entry.ID = cleanCell(cells[1])
	entry.RegistrationState = cleanCell(cells[2])
	entry.ImplementationPaths = parsePathList(cells[3])
	entry.SpecFiles = parsePathList(cells[4])

	diagnostics := []string{}
	if entry.Kind != "unit" && entry.Kind != "rule" {
		diagnostics = append(diagnostics, fmt.Sprintf("Object Registry row %d has invalid kind %q", line, entry.Kind))
	}
	if entry.ID == "" {
		diagnostics = append(diagnostics, fmt.Sprintf("Object Registry row %d has empty id", line))
	}
	if entry.RegistrationState != "planned" && entry.RegistrationState != "landed" {
		diagnostics = append(diagnostics, fmt.Sprintf("Object Registry row %d has invalid registration_state %q", line, entry.RegistrationState))
	}
	if entry.RegistrationState == "planned" && len(entry.ImplementationPaths) > 0 {
		diagnostics = append(diagnostics, fmt.Sprintf("Object Registry row %d planned object must use implementation_paths=none", line))
	}
	if entry.RegistrationState == "landed" && len(entry.ImplementationPaths) == 0 {
		diagnostics = append(diagnostics, fmt.Sprintf("Object Registry row %d landed object must declare implementation_paths", line))
	}
	return entry, diagnostics
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
		normalized = append(normalized, cleanCell(cell))
	}
	return strings.Join(normalized, "|")
}

func isMarkdownSeparatorRow(cells []string) bool {
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
	return len(cells) > 0
}

func parsePathList(cell string) []string {
	cell = cleanCell(cell)
	if cell == "" || cell == "none" {
		return nil
	}
	parts := strings.Split(cell, ";")
	paths := []string{}
	for _, part := range parts {
		path := cleanCell(part)
		if path == "" || path == "none" {
			continue
		}
		paths = append(paths, filepath.ToSlash(path))
	}
	return paths
}

func cleanCell(cell string) string {
	cell = strings.TrimSpace(cell)
	cell = strings.Trim(cell, "` ")
	return strings.TrimSpace(cell)
}

func exists(repoRoot, relPath string) bool {
	_, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(relPath)))
	return err == nil
}

func implementationPathExists(repoRoot, relPath string) bool {
	relPath = strings.TrimSpace(filepath.ToSlash(relPath))
	if relPath == "" || relPath == "none" {
		return false
	}
	if strings.HasSuffix(relPath, "/**") {
		base := strings.TrimSuffix(relPath, "/**")
		info, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(base)))
		return err == nil && info.IsDir()
	}
	if strings.ContainsAny(relPath, "*?[") {
		matches, err := filepath.Glob(filepath.Join(repoRoot, filepath.FromSlash(relPath)))
		return err == nil && len(matches) > 0
	}
	return exists(repoRoot, relPath)
}
