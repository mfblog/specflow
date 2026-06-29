package promote

import (
	"io"
	"os"
	"strings"
)

// parseFrontmatter extracts YAML frontmatter fields from a markdown file.
// It handles the --- delimited frontmatter block at the start of the file.
func parseFrontmatter(content string) map[string]string {
	result := map[string]string{}

	if !strings.HasPrefix(strings.TrimSpace(content), "---") {
		return result
	}

	lines := strings.Split(content, "\n")
	startIdx := -1
	endIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if startIdx == -1 {
				startIdx = i
			} else {
				endIdx = i
				break
			}
		}
	}

	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx+1 {
		return result
	}

	for _, line := range lines[startIdx+1 : endIdx] {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")
		result[key] = value
	}

	return result
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	if err := os.MkdirAll(dirName(dst), 0755); err != nil {
		return err
	}

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	return err
}

func dirName(path string) string {
	idx := strings.LastIndex(path, string(os.PathSeparator))
	if idx == -1 {
		return "."
	}
	return path[:idx]
}
