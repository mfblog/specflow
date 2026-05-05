package rulebinding

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

type ResolvedRef struct {
	VersionRef  string
	FileRef     string
	Layer       string
	RuleID      string
	RuleScope   string
	RuleVersion string
	Content     string
}

func ResolveRef(repoRoot, moduleLayer, ref string) (ResolvedRef, error) {
	versionRef := strings.TrimSpace(ref)
	prefix, expectedVersion, err := splitVersionRef(versionRef)
	if err != nil {
		return ResolvedRef{}, err
	}

	layer, err := layerFromPrefix(prefix)
	if err != nil {
		return ResolvedRef{}, err
	}
	if moduleLayer == "stable" && layer != "stable" {
		return ResolvedRef{}, fmt.Errorf("stable-layer object binding must use an s_ rule ref, got %q", versionRef)
	}

	fileRef := ruleFileRef(prefix, layer)
	contentBytes, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(fileRef)))
	if err != nil {
		return ResolvedRef{}, fmt.Errorf("read rule %s: %w", fileRef, err)
	}
	content := string(contentBytes)

	frontmatter, err := parseFrontmatter(content)
	if err != nil {
		return ResolvedRef{}, fmt.Errorf("%s: %w", fileRef, err)
	}

	ruleID := strings.TrimSpace(frontmatter["rule_id"])
	ruleScope := strings.TrimSpace(frontmatter["rule_scope"])
	actualLayer := strings.TrimSpace(frontmatter["layer"])
	actualVersion := strings.TrimSpace(frontmatter["rule_version"])
	if ruleID == "" || ruleScope == "" || actualLayer == "" || actualVersion == "" {
		return ResolvedRef{}, fmt.Errorf("%s: missing rule_id/rule_scope/layer/rule_version", fileRef)
	}
	if ruleScope != "global" && ruleScope != "bound" {
		return ResolvedRef{}, fmt.Errorf("%s: rule_scope must be global or bound", fileRef)
	}
	if actualLayer != layer {
		return ResolvedRef{}, fmt.Errorf("%s: frontmatter.layer=%s does not match bound layer %s", fileRef, actualLayer, layer)
	}
	if actualVersion != expectedVersion {
		return ResolvedRef{}, fmt.Errorf("%s: bound version %q does not match frontmatter rule_version %q", fileRef, expectedVersion, actualVersion)
	}
	if err := ValidatePromotionOwnerUnit(repoRoot, fileRef, actualLayer, strings.TrimSpace(frontmatter["promotion_owner_unit"])); err != nil {
		return ResolvedRef{}, err
	}

	return ResolvedRef{
		VersionRef:  versionRef,
		FileRef:     fileRef,
		Layer:       actualLayer,
		RuleID:      ruleID,
		RuleScope:   ruleScope,
		RuleVersion: actualVersion,
		Content:     content,
	}, nil
}

func ValidatePromotionOwnerUnit(repoRoot, fileRef, layer, promotionOwnerUnit string) error {
	owner := strings.TrimSpace(promotionOwnerUnit)
	if layer != "candidate" {
		if owner != "" {
			return fmt.Errorf("%s: promotion_owner_unit is allowed only on candidate-layer rule files with a stable sibling", fileRef)
		}
		return nil
	}

	stableSiblingRef := "docs/specs/rules/stable/s_" + strings.TrimPrefix(filepath.Base(fileRef), "c_")
	_, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(stableSiblingRef)))
	hasStableSibling := err == nil
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat %s: %w", stableSiblingRef, err)
	}
	if !hasStableSibling {
		if owner != "" {
			return fmt.Errorf("%s: promotion_owner_unit must not be recorded when no stable-layer sibling exists", fileRef)
		}
		return nil
	}
	if owner == "" {
		return fmt.Errorf("%s: missing promotion_owner_unit for candidate-layer rule file with stable sibling %s", fileRef, stableSiblingRef)
	}
	if _, err := statusfile.LookupModuleStatus(repoRoot, owner); err != nil {
		return fmt.Errorf("%s: promotion_owner_unit %q is not a registered formal unit", fileRef, owner)
	}
	return nil
}

func splitVersionRef(ref string) (string, string, error) {
	parts := strings.SplitN(ref, "@", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid rule ref %q", ref)
	}
	prefix := strings.TrimSpace(parts[0])
	version := strings.TrimSpace(parts[1])
	if prefix == "" || version == "" {
		return "", "", fmt.Errorf("invalid rule ref %q", ref)
	}
	return prefix, version, nil
}

func layerFromPrefix(prefix string) (string, error) {
	switch {
	case strings.HasPrefix(prefix, "c_"):
		return "candidate", nil
	case strings.HasPrefix(prefix, "s_"):
		return "stable", nil
	default:
		return "", fmt.Errorf("invalid rule ref prefix %q", prefix)
	}
}

func ruleFileRef(prefix, layer string) string {
	return fmt.Sprintf("docs/specs/rules/%s/%s.md", layer, prefix)
}

func parseFrontmatter(content string) (map[string]string, error) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, fmt.Errorf("missing frontmatter start marker")
	}
	endIdx := -1
	for idx := 1; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "---" {
			endIdx = idx
			break
		}
	}
	if endIdx == -1 {
		return nil, fmt.Errorf("missing frontmatter end marker")
	}

	values := map[string]string{}
	for _, line := range lines[1:endIdx] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		values[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return values, nil
}
