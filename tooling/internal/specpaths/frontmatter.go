package specpaths

import "strings"

// ReadFrontmatterStringMap extracts a flat string map from YAML frontmatter.
// Supports both inline list syntax (unit_refs: [a, b]) and block-style YAML
// lists. Block-style list items for the same key are accumulated into an
// inline-formatted string [a, b] for downstream parsing.
func ReadFrontmatterStringMap(text string) map[string]string {
	result := map[string]string{}
	normalized := NormalizeText(text)
	lines := strings.Split(normalized, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return result
	}

	currentKey := ""
	for idx := 1; idx < len(lines); idx++ {
		line := lines[idx]
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			break
		}
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "- ") && currentKey != "" {
			// Accumulate block-style YAML list items into inline list format [a, b]
			item := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			if item == "" {
				continue
			}
			existing := result[currentKey]
			if existing == "" {
				result[currentKey] = "[" + item + "]"
			} else if strings.HasPrefix(existing, "[") {
				result[currentKey] = existing[:len(existing)-1] + ", " + item + "]"
			}
			continue
		}

		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if value == "" {
			currentKey = key
			continue
		}
		value = strings.Trim(value, "`\"' ")
		result[key] = value
	}
	return result
}

// ParseRefList parses a YAML list or comma-separated string into ref tokens.
// Accepts both "[a, b, c]" and "a, b, c" formats.
func ParseRefList(value string) []string {
	value = strings.TrimSpace(value)

	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")

	parts := strings.Split(value, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, "\"'")
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
