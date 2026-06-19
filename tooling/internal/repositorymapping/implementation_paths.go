package repositorymapping

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
)

func GetImplementationPaths(repoRoot, unit string) ([]string, error) {
	relPath := specpaths.RepositoryMappingFileRef
	data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relPath)))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", relPath, err)
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	inRegistry := false
	headerSeen := false

	for _, line := range lines {
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
		kind := cleanCell(cells[0])
		id := cleanCell(cells[1])
		if kind != "unit" || id != unit {
			continue
		}
		return parsePathList(cells[3]), nil
	}
	return nil, fmt.Errorf("unit %q not found in Object Registry of %s", unit, relPath)
}
