package manifest

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const RelativePath = "specflow/tooling/manifest.tsv"

type Item struct {
	SourceRelative      string
	DestinationRelative string
	Mode                string
}

func Load(repoRoot string) ([]Item, error) {
	path := filepath.Join(repoRoot, filepath.FromSlash(RelativePath))
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", RelativePath, err)
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
		return nil, fmt.Errorf("scan %s: %w", RelativePath, err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no manifest entries found in %s", RelativePath)
	}
	return items, nil
}
