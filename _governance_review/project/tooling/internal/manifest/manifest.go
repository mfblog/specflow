package manifest

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specflowlayout"
)

type Item struct {
	SourceRelative      string
	DestinationRelative string
	Mode                string
}

func Load(repoRoot string) ([]Item, error) {
	layout, err := specflowlayout.Resolve(repoRoot)
	if err != nil {
		return nil, err
	}
	relativePath := specflowlayout.Relative(layout.ToolingRoot, "manifest.tsv")
	path := filepath.Join(repoRoot, filepath.FromSlash(relativePath))
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", relativePath, err)
	}
	defer file.Close()

	items := []Item{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid manifest row %q", line)
		}
		items = append(items, Item{
			SourceRelative:      parts[0],
			DestinationRelative: parts[1],
			Mode:                parts[2],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", relativePath, err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no manifest entries found in %s", relativePath)
	}
	return items, nil
}
