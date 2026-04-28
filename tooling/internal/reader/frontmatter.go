package reader

import (
	"strings"
)

type frontmatter struct {
	Scalars      map[string]string
	BoundObjects []string
}

func parseFrontmatter(text string) frontmatter {
	result := frontmatter{Scalars: map[string]string{}}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return result
	}

	currentList := ""
	for idx := 1; idx < len(lines); idx++ {
		line := lines[idx]
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			break
		}
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") && currentList == "bound_objects" {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			if value != "" {
				result.BoundObjects = append(result.BoundObjects, value)
			}
			continue
		}
		currentList = ""
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if value == "" {
			currentList = key
			continue
		}
		result.Scalars[key] = trimQuote(value)
	}
	return result
}

func trimQuote(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "`")
	value = strings.Trim(value, "\"")
	value = strings.Trim(value, "'")
	return value
}
