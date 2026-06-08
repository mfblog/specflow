package context

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/unitappendix"
)

// --- lifecycle collector ---

type lifecycleCollector struct {
	command string
	rules   inputRules
}

type inputRules struct {
	essential []InputRule
	reference []InputRule
}

func (c *lifecycleCollector) Flow() string       { return "lifecycle" }
func (c *lifecycleCollector) Command() string     { return c.command }
func (c *lifecycleCollector) Collect(repoRoot, object string) (*Pack, error) {
	if object == "" {
		return nil, fmt.Errorf("object is required for lifecycle command %q", c.command)
	}
	return CollectByRules("lifecycle", c.command, repoRoot, object, c.rules.essential, c.rules.reference)
}

// NewLifecycleCollector creates a collector for the given lifecycle command.
func NewLifecycleCollector(command string) (Collector, error) {
	rules, ok := lifecycleInputs[command]
	if !ok {
		return nil, fmt.Errorf("unknown lifecycle command %q", command)
	}
	return &lifecycleCollector{command: command, rules: rules}, nil
}

// --- lifecycle input rule definitions ---
//
// Each entry is derived from the matching Context Card in framework/lifecycle/*.md.
// Essential = core truth files that the agent must read (inlined).
// Reference = optional or supplementary files (path links only).

var lifecycleInputs = map[string]inputRules{
	// unit_init — 将已有能力直接记录为首个稳定 truth（无候选层）
	// Context Card: unit_init_new_fork.md
	"unit_init": {
		essential: []InputRule{
			{PathTemplate: "docs/specs/_status.md", Essential: true},
			{PathTemplate: "docs/specs/repository_mapping.md", Essential: true},
		},
		reference: []InputRule{
			{PathTemplate: "framework/lifecycle/unit_init_new_fork.md", Essential: false},
		},
	},

	// unit_new — 为一个全新的 unit 创建首个候选 truth
	// Context Card: unit_init_new_fork.md
	"unit_new": {
		essential: []InputRule{
			{PathTemplate: "docs/specs/_status.md", Essential: true},
			{PathTemplate: "docs/specs/repository_mapping.md", Essential: true},
		},
		reference: []InputRule{
			{PathTemplate: "framework/lifecycle/unit_init_new_fork.md", Essential: false},
		},
	},

	// unit_fork — 从现有稳定 truth 分支出一个候选轮次
	// Context Card: unit_init_new_fork.md
	"unit_fork": {
		essential: []InputRule{
			{PathTemplate: "docs/specs/_status.md", Essential: true},
			{PathTemplate: "docs/specs/repository_mapping.md", Essential: true},
			{PathTemplate: "docs/specs/units/stable/s_unit_{object}.md", Essential: true},
		},
		reference: []InputRule{
			{PathTemplate: "framework/lifecycle/unit_init_new_fork.md", Essential: false},
		},
	},

	// unit_check — 候选 truth 质量检查
	// Context Card: unit_check.md
	"unit_check": {
		essential: []InputRule{
			{PathTemplate: "docs/specs/_status.md", Essential: true},
			{PathTemplate: "docs/specs/units/candidate/c_unit_{object}.md", Essential: true},
			{PathTemplate: "docs/specs/units/stable/s_unit_{object}.md", Essential: true},
			{Resolve: resolveCandidateAppendices, Essential: true},
			{Resolve: resolveRuleRefs, Essential: true},
			{Resolve: resolveUnitRefs, Essential: true},
		},
		reference: []InputRule{
			{PathTemplate: "framework/lifecycle/unit_check.md", Essential: false},
		},
	},

	// unit_impl — 实现阶段（unit_check pass 后自动推进）
	// Context Card: unit_impl.md
	"unit_impl": {
		essential: []InputRule{
			{PathTemplate: "docs/specs/units/candidate/c_unit_{object}.md", Essential: true},
			{PathTemplate: "docs/specs/units/stable/s_unit_{object}.md", Essential: true},
			{Resolve: resolveCandidateAppendices, Essential: true},
			{Resolve: resolveRuleRefs, Essential: true},
			{Resolve: resolveUnitRefs, Essential: true},
		},
		reference: []InputRule{
			{PathTemplate: "docs/specs/_check_result/unit/{object}.md", Essential: false, Optional: true},
			{PathTemplate: "framework/lifecycle/unit_impl.md", Essential: false},
		},
	},

	// unit_verify — 验证实现是否满足候选 truth
	// Context Card: unit_verify.md
	"unit_verify": {
		essential: []InputRule{
			{PathTemplate: "docs/specs/_status.md", Essential: true},
			{PathTemplate: "docs/specs/units/candidate/c_unit_{object}.md", Essential: true},
			{PathTemplate: "docs/specs/units/stable/s_unit_{object}.md", Essential: true},
			{Resolve: resolveCandidateAppendices, Essential: true},
			{Resolve: resolveRuleRefs, Essential: true},
			{Resolve: resolveUnitRefs, Essential: true},
		},
		reference: []InputRule{
			{PathTemplate: "docs/specs/_check_result/unit/{object}.md", Essential: false, Optional: true},
			{PathTemplate: "framework/lifecycle/unit_verify.md", Essential: false},
		},
	},

	// unit_promote — 候选 truth 晋升为稳定 truth
	// Context Card: unit_promote.md
	"unit_promote": {
		essential: []InputRule{
			{PathTemplate: "docs/specs/_status.md", Essential: true},
			{PathTemplate: "docs/specs/units/candidate/c_unit_{object}.md", Essential: true},
			{PathTemplate: "docs/specs/units/stable/s_unit_{object}.md", Essential: true},
			{Resolve: resolveCandidateAppendices, Essential: true},
		},
		reference: []InputRule{
			{PathTemplate: "docs/specs/_verify_result/unit/{object}.md", Essential: false},
			{PathTemplate: "framework/lifecycle/unit_promote.md", Essential: false},
		},
	},

	// unit_stable_verify — 实现与稳定 truth 的一致性检查
	// Context Card: unit_stable_verify.md
	"unit_stable_verify": {
		essential: []InputRule{
			{PathTemplate: "docs/specs/_status.md", Essential: true},
			{PathTemplate: "docs/specs/repository_mapping.md", Essential: true},
			{PathTemplate: "docs/specs/units/stable/s_unit_{object}.md", Essential: true},
			{Resolve: resolveStableAppendices, Essential: true},
			{Resolve: resolveRuleRefs, Essential: true},
			{Resolve: resolveUnitRefs, Essential: true},
		},
		reference: []InputRule{
			{PathTemplate: "docs/specs/_stable_verify_result/unit/{object}.md", Essential: false, Optional: true},
			{PathTemplate: "framework/lifecycle/unit_stable_verify.md", Essential: false},
		},
	},

	// unit_advance — 自动推进（组合 check + verify + relation preflight）
	// Context Card: advance_policy.md + dynamic Context Card
	"unit_advance": {
		essential: []InputRule{
			{PathTemplate: "docs/specs/_status.md", Essential: true},
			{PathTemplate: "docs/specs/units/candidate/c_unit_{object}.md", Essential: true},
			{PathTemplate: "framework/lifecycle/overview.md", Essential: true},
		},
		reference: []InputRule{
			{PathTemplate: "framework/advance_policy.md", Essential: false},
		},
	},
}

// --- resolve functions ---

// resolveRuleRefs reads the candidate spec's frontmatter rule_refs
// and resolves each ref to an actual file path.
func resolveRuleRefs(repoRoot, object string) ([]FileItem, error) {
	// Try candidate spec first, fall back to stable spec.
	paths := []string{
		filepath.Join(repoRoot, filepath.FromSlash("docs/specs/units/candidate/c_unit_"+object+".md")),
		filepath.Join(repoRoot, filepath.FromSlash("docs/specs/units/stable/s_unit_"+object+".md")),
	}
	var content string
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			content = string(data)
			break
		}
	}
	if content == "" {
		return nil, nil
	}

	// Also collect global rule files (always apply).
	var items []FileItem
	seen := map[string]bool{}
	globalPattern := filepath.Join(repoRoot, filepath.FromSlash("docs/specs/rules/stable/s_g_rule_*.md"))
	globals, _ := filepath.Glob(globalPattern)
	for _, match := range globals {
		rel, err := filepath.Rel(repoRoot, match)
		if err != nil {
			continue
		}
		relSlash := filepath.ToSlash(rel)
		if seen[relSlash] {
			continue
		}
		seen[relSlash] = true
		item := ResolveFileItem(repoRoot, relSlash, "", true, false)
		if item.Exists {
			items = append(items, item)
		}
	}

	// Parse rule_refs from frontmatter.
	refs := parseNamedRefs(content, "rule_refs")
	for _, ref := range refs {
		path := resolveRuleRefToPath(ref)
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		item := ResolveFileItem(repoRoot, path, "", true, false)
		if item.Exists {
			items = append(items, item)
		}
	}
	return items, nil
}

// resolveUnitRefs reads the candidate spec's frontmatter unit_refs
// and resolves each ref to a stable unit spec file path.
func resolveUnitRefs(repoRoot, object string) ([]FileItem, error) {
	paths := []string{
		filepath.Join(repoRoot, filepath.FromSlash("docs/specs/units/candidate/c_unit_"+object+".md")),
		filepath.Join(repoRoot, filepath.FromSlash("docs/specs/units/stable/s_unit_"+object+".md")),
	}
	var content string
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			content = string(data)
			break
		}
	}
	if content == "" {
		return nil, nil
	}

	refs := parseNamedRefs(content, "unit_refs")
	var items []FileItem
	for _, ref := range refs {
		path := resolveUnitRefToPath(ref)
		if path == "" {
			continue
		}
		item := ResolveFileItem(repoRoot, path, "", true, false)
		if item.Exists {
			items = append(items, item)
		}
	}
	return items, nil
}

// resolveCandidateAppendices scans candidate appendix files for the unit.
func resolveCandidateAppendices(repoRoot, object string) ([]FileItem, error) {
	entries, err := unitappendix.Scan(repoRoot, "unit", object, "candidate")
	if err != nil {
		return nil, nil // silently skip if appendix dir doesn't exist
	}
	var items []FileItem
	for _, entry := range entries {
		items = append(items, FileItem{
			Path:      entry.FileRef,
			Essential: true,
			Exists:    true,
			Content:   entry.Content,
			LineCount: strings.Count(entry.Content, "\n"),
		})
	}
	return items, nil
}

// resolveStableAppendices scans stable appendix files for the unit.
func resolveStableAppendices(repoRoot, object string) ([]FileItem, error) {
	entries, err := unitappendix.Scan(repoRoot, "unit", object, "stable")
	if err != nil {
		return nil, nil
	}
	var items []FileItem
	for _, entry := range entries {
		items = append(items, FileItem{
			Path:      entry.FileRef,
			Essential: true,
			Exists:    true,
			Content:   entry.Content,
			LineCount: strings.Count(entry.Content, "\n"),
		})
	}
	return items, nil
}

// --- internal helpers ---

// parseNamedRefs parses a named YAML list or scalar from markdown frontmatter.
// It handles both "field: value" and "field:\n  - item1\n  - item2" formats.
func parseNamedRefs(content, fieldName string) []string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	result := []string{}
	seen := map[string]bool{}
	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		right, matched := namedField(trimmed, fieldName)
		if !matched {
			continue
		}
		if right == "`none`" || right == "none" {
			continue
		}
		if right != "" {
			item := strings.Trim(strings.TrimSpace(right), "`\"'")
			if item != "" && !seen[item] {
				result = append(result, item)
				seen[item] = true
			}
			continue
		}
		for next := idx + 1; next < len(lines); next++ {
			nextTrimmed := strings.TrimSpace(lines[next])
			if nextTrimmed == "" {
				continue
			}
			if !strings.HasPrefix(nextTrimmed, "- ") {
				break
			}
			item := strings.TrimSpace(strings.TrimPrefix(nextTrimmed, "- "))
			item = strings.Trim(item, "`\"'")
			if item != "" && !seen[item] {
				result = append(result, item)
				seen[item] = true
			}
		}
	}
	sort.Strings(result)
	return result
}

func namedField(trimmed, fieldName string) (string, bool) {
	parts := strings.SplitN(trimmed, ":", 2)
	if len(parts) != 2 {
		return "", false
	}
	left := strings.ReplaceAll(strings.TrimSpace(parts[0]), "`", "")
	if idx := strings.Index(left, ". "); idx > 0 {
		allDigits := true
		for _, ch := range left[:idx] {
			if ch < '0' || ch > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			left = left[idx+2:]
		}
	}
	if strings.TrimSpace(left) != fieldName {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}

// resolveRuleRefToPath converts a rule ref like "s_b_rule_demo@1.0.0"
// to a file path like "docs/specs/rules/stable/s_b_rule_demo.md".
func resolveRuleRefToPath(ref string) string {
	ref = strings.Trim(ref, "`")
	prefix, _, ok := strings.Cut(ref, "@")
	if !ok {
		return ""
	}
	switch {
	case strings.HasPrefix(prefix, "s_"):
		return "docs/specs/rules/stable/" + prefix + ".md"
	case strings.HasPrefix(prefix, "c_"):
		return "docs/specs/rules/candidate/" + prefix + ".md"
	}
	return ""
}

// resolveUnitRefToPath converts a unit ref like "s_unit_user@1.2.0"
// to a file path like "docs/specs/units/stable/s_unit_user.md".
func resolveUnitRefToPath(ref string) string {
	ref = strings.Trim(ref, "`")
	prefix, _, ok := strings.Cut(ref, "@")
	if !ok {
		return ""
	}
	switch {
	case strings.HasPrefix(prefix, "s_unit_"):
		return "docs/specs/units/stable/" + prefix + ".md"
	case strings.HasPrefix(prefix, "c_unit_"):
		return "docs/specs/units/candidate/" + prefix + ".md"
	}
	return ""
}

// Ensure lifecycleCollector implements Collector.
var _ Collector = (*lifecycleCollector)(nil)

// RegisterLifecycleCommands registers all lifecycle collectors.
func RegisterLifecycleCommands() ([]string, map[string]Collector) {
	commands := make([]string, 0, len(lifecycleInputs))
	collectors := make(map[string]Collector, len(lifecycleInputs))
	for cmd := range lifecycleInputs {
		commands = append(commands, cmd)
		collectors[cmd] = &lifecycleCollector{command: cmd, rules: lifecycleInputs[cmd]}
	}
	sort.Strings(commands)
	return commands, collectors
}
