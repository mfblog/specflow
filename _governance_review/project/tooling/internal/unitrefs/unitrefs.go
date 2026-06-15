package unitrefs

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var unitRefPattern = regexp.MustCompile(`^s_unit_[a-z0-9_]+@[0-9]+\.[0-9]+\.[0-9]+$`)

type document struct {
	lines    []string
	endIndex int
}

func ParseObjectUnitRefs(fileRef string, content string) ([]string, error) {
	doc, err := parseDocument(content)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fileRef, err)
	}
	refs, found, err := parseUnitRefsFromFrontmatter(doc.lines[1:doc.endIndex])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fileRef, err)
	}
	if !found {
		return nil, nil
	}
	return refs, nil
}

func UpdateObjectUnitRefs(fileRef string, content string, refs []string) (string, error) {
	doc, err := parseDocument(content)
	if err != nil {
		return "", fmt.Errorf("%s: %w", fileRef, err)
	}
	normalized, err := NormalizeUnitRefs(refs)
	if err != nil {
		return "", fmt.Errorf("%s: %w", fileRef, err)
	}

	fm := make([]string, 0, doc.endIndex+4)
	fm = append(fm, "---")
	wroteRefs := false
	bodyStart := doc.endIndex + 1
	for i := 1; i < doc.endIndex; i++ {
		line := doc.lines[i]
		key, ok := frontmatterKey(line)
		if !ok {
			fm = append(fm, line)
			continue
		}
		if key == "unit_refs" {
			if !wroteRefs {
				fm = appendUnitRefs(fm, normalized)
				wroteRefs = true
			}
			i = skipFrontmatterBlock(doc.lines, i+1, doc.endIndex) - 1
			continue
		}
		fm = append(fm, line)
	}
	if !wroteRefs {
		fm = appendUnitRefs(fm, normalized)
	}
	fm = append(fm, "---")

	out := strings.Join(fm, "\n")
	body := strings.TrimSuffix(strings.Join(doc.lines[bodyStart:], "\n"), "\n")
	if strings.TrimSpace(body) != "" {
		out += "\n" + body
	}
	if strings.HasSuffix(content, "\n") {
		out += "\n"
	}
	return out, nil
}

func FrontmatterScalar(fileRef string, content string, key string) (string, error) {
	doc, err := parseDocument(content)
	if err != nil {
		return "", fmt.Errorf("%s: %w", fileRef, err)
	}
	for i := 1; i < doc.endIndex; i++ {
		lineKey, ok := frontmatterKey(doc.lines[i])
		if !ok || lineKey != key {
			continue
		}
		parts := strings.SplitN(doc.lines[i], ":", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("%s: invalid frontmatter field %q", fileRef, key)
		}
		return strings.TrimSpace(parts[1]), nil
	}
	return "", fmt.Errorf("%s: frontmatter field %q is required", fileRef, key)
}

func NormalizeUnitRefs(refs []string) ([]string, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	seen := map[string]bool{}
	normalized := make([]string, 0, len(refs))
	for _, ref := range refs {
		ref = strings.TrimSpace(strings.Trim(ref, "`"))
		if ref == "" {
			return nil, fmt.Errorf("unit_refs contains an empty item")
		}
		if !unitRefPattern.MatchString(ref) {
			return nil, fmt.Errorf("invalid unit ref %q", ref)
		}
		if seen[ref] {
			return nil, fmt.Errorf("unit_refs contains duplicate item %q", ref)
		}
		seen[ref] = true
		normalized = append(normalized, ref)
	}
	sort.Strings(normalized)
	return normalized, nil
}

func ReplaceUnitRef(refs []string, fromRef, toRef string) ([]string, bool, error) {
	changed := false
	next := make([]string, 0, len(refs))
	for _, ref := range refs {
		if ref == fromRef {
			next = append(next, toRef)
			changed = true
			continue
		}
		next = append(next, ref)
	}
	if !changed {
		return refs, false, nil
	}
	normalized, err := NormalizeUnitRefs(next)
	if err != nil {
		return nil, true, err
	}
	return normalized, true, nil
}

func parseDocument(content string) (document, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return document{}, fmt.Errorf("frontmatter block is required")
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return document{lines: lines, endIndex: i}, nil
		}
	}
	return document{}, fmt.Errorf("frontmatter closing marker is required")
}

func parseUnitRefsFromFrontmatter(lines []string) ([]string, bool, error) {
	for idx, line := range lines {
		key, right, matched := frontmatterField(line)
		if !matched || key != "unit_refs" {
			continue
		}
		if right == "none" || right == "`none`" {
			return nil, true, nil
		}
		if right != "" {
			return nil, false, fmt.Errorf("unit_refs must use literal none or a YAML list")
		}
		refs := []string{}
		seen := map[string]bool{}
		for next := idx + 1; next < len(lines); next++ {
			trimmed := strings.TrimSpace(lines[next])
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			if !strings.HasPrefix(trimmed, "- ") {
				break
			}
			ref := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			ref = strings.Trim(strings.Trim(ref, "`"), "\"'")
			if ref == "" {
				return nil, false, fmt.Errorf("unit_refs contains an empty item")
			}
			if seen[ref] {
				return nil, false, fmt.Errorf("unit_refs contains duplicate item %q", ref)
			}
			seen[ref] = true
			refs = append(refs, ref)
		}
		if len(refs) == 0 {
			return nil, false, fmt.Errorf("unit_refs must not be an empty list")
		}
		normalized, err := NormalizeUnitRefs(refs)
		if err != nil {
			return nil, false, err
		}
		for i := range refs {
			if refs[i] != normalized[i] {
				return nil, false, fmt.Errorf("unit_refs must be sorted by exact ref string in ascending lexical order")
			}
		}
		return refs, true, nil
	}
	return nil, false, nil
}

func appendUnitRefs(lines []string, refs []string) []string {
	if len(refs) == 0 {
		return append(lines, "unit_refs: none")
	}
	lines = append(lines, "unit_refs:")
	for _, ref := range refs {
		lines = append(lines, "  - "+ref)
	}
	return lines
}

func skipFrontmatterBlock(lines []string, start, end int) int {
	idx := start
	for idx < end {
		trimmed := strings.TrimSpace(lines[idx])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			idx++
			continue
		}
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(lines[idx], " ") || strings.HasPrefix(lines[idx], "\t") {
			idx++
			continue
		}
		break
	}
	return idx
}

func frontmatterKey(line string) (string, bool) {
	key, _, matched := frontmatterField(line)
	return key, matched
}

func frontmatterField(line string) (string, string, bool) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	key := strings.TrimSpace(parts[0])
	if key == "" || strings.Contains(key, " ") {
		return "", "", false
	}
	return key, strings.TrimSpace(parts[1]), true
}
