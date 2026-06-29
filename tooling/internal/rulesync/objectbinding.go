package rulesync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// readObjectRuleRefs reads a unit spec file and returns its rule_refs.
// In the simplified model, the layer is determined directly from the file path.
func readObjectRuleRefs(repoRoot, object, layer string) ([]string, error) {
	prefix := "c"
	if layer == "stable" {
		prefix = "s"
	}
	fileRef := fmt.Sprintf("docs/specs/units/%s/%s_unit_%s.md", layer, prefix, object)
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(fileRef)))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", fileRef, err)
	}

	// Simple frontmatter parsing to find rule_refs
	lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
	inFrontmatter := false
	foundRuleRefs := false
	var refs []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			break
		}
		if !inFrontmatter {
			continue
		}
		if foundRuleRefs {
			if !strings.HasPrefix(trimmed, "- ") {
				break
			}
			ref := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			ref = strings.Trim(ref, "`")
			if ref != "" {
				refs = append(refs, ref)
			}
			continue
		}
		if strings.HasPrefix(trimmed, "rule_refs:") {
			right := strings.TrimSpace(trimmed[len("rule_refs:"):])
			if right == "none" || right == "`none`" {
				return nil, nil
			}
			if right != "" {
				continue
			}
			foundRuleRefs = true
		}
	}

	return normalizeStrings(refs), nil
}
