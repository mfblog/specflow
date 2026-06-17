package repositorymapping

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
)

// UpdateSpecFilesForFork updates the repository_mapping.md Object Registry
// entry for the given unit, replacing its stable-layer spec file path with
// the candidate-layer path. This is called during rule release-version
// auto-fork to reflect that the unit's active spec has moved to candidate.
//
// It returns an error if the unit's row is not found, or if the spec_files
// column does not contain the expected stable path.
func UpdateSpecFilesForFork(repoRoot, unit string) error {
	relPath := specpaths.RepositoryMappingFileRef
	absPath := filepath.Join(repoRoot, filepath.FromSlash(relPath))
	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", relPath, err)
	}

	stableRef, err := specpaths.ObjectMainSpecFileRef("unit", "stable", unit)
	if err != nil {
		return fmt.Errorf("stable spec path for %q: %w", unit, err)
	}
	candidateRef, err := specpaths.ObjectMainSpecFileRef("unit", "candidate", unit)
	if err != nil {
		return fmt.Errorf("candidate spec path for %q: %w", unit, err)
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	inRegistry := false
	headerSeen := false
	updated := false

	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "## 2. Object Registry"):
			inRegistry = true
			headerSeen = false
			continue
		case inRegistry && strings.HasPrefix(trimmed, "## ") && !strings.HasPrefix(trimmed, "## 2. Object Registry"):
			inRegistry = false
			continue
		}
		if !inRegistry || !strings.HasPrefix(trimmed, "|") {
			continue
		}
		cells := splitMarkdownTableRow(trimmed)
		if len(cells) == 0 || isMarkdownSeparatorRow(cells) {
			continue
		}
		if !headerSeen {
			headerSeen = true
			continue
		}
		// We are in a data row. Check if it belongs to the target unit.
		kind := cleanCell(cells[0])
		id := cleanCell(cells[1])
		if kind != "unit" || id != unit {
			continue
		}
		// Found the unit's row. Update spec_files column (index 4).
		specFiles := cleanCell(cells[4])
		if !strings.Contains(specFiles, stableRef) {
			return fmt.Errorf("%s row for unit %q: spec_files %q does not contain expected stable path %q",
				relPath, unit, specFiles, stableRef)
		}
		newSpecFiles := strings.Replace(specFiles, stableRef, candidateRef, 1)
		if newSpecFiles == specFiles {
			return fmt.Errorf("%s row for unit %q: replacing %q with %q produced no change",
				relPath, unit, stableRef, candidateRef)
		}
		lines[idx] = reconstructTableRow(cells, newSpecFiles)
		updated = true
		break
	}

	if !updated {
		return fmt.Errorf("%s: no Object Registry row found for unit %q", relPath, unit)
	}

	output := strings.Join(lines, "\n")
	if err := os.WriteFile(absPath, []byte(output), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", relPath, err)
	}
	return nil
}

// reconstructTableRow rebuilds a markdown table row, replacing the spec_files
// column (index 4) with the given value. Other cells are preserved as-is
// from the original split.
func reconstructTableRow(cells []string, newSpecFiles string) string {
	updated := make([]string, len(cells))
	copy(updated, cells)
	if len(updated) > 4 {
		updated[4] = newSpecFiles
	}
	return "| " + strings.Join(updated, " | ") + " |"
}
