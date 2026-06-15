package contextcard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// frameworkRoot returns the abs path to the framework/ directory from the tooling module.
func frameworkRoot(t *testing.T) string {
	t.Helper()
	// Walk up from tooling/ to repo root, then into framework/
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "framework", "lifecycle", "overview.md")); err == nil {
			return filepath.Join(wd, "framework")
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("cannot find repo root containing framework/lifecycle/overview.md")
		}
		wd = parent
	}
}

func loadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}

// TestLifecycleSections_NotAllowed verifies every lifecycle file has a
// well-formed "Not Allowed" section that parseListItems can extract.
func TestLifecycleSections_NotAllowed(t *testing.T) {
	root := frameworkRoot(t)
	pattern := filepath.Join(root, "lifecycle", "unit_*.md")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatalf("no lifecycle files found at %s", pattern)
	}

	for _, path := range matches {
		name := filepath.Base(path)
		content := loadFile(t, path)
		items := parseListItems(stripFrontmatter(content), "Not Allowed")
		if len(items) == 0 {
			t.Errorf("%s: 'Not Allowed' section is empty or missing — BLOCKED extraction will fail for this lifecycle file", name)
		}

		// Verify the section itself exists (not just empty items)
		section := extractMarkdownSection(stripFrontmatter(content), "Not Allowed")
		if section == "" {
			t.Errorf("%s: 'Not Allowed' section heading not found — check heading text matches exactly 'Not Allowed'", name)
		}
	}
}

// TestLifecycleSections_HowToEnd verifies every lifecycle file has a
// well-formed "How to End" section that parseCloseCommand can extract from.
func TestLifecycleSections_HowToEnd(t *testing.T) {
	root := frameworkRoot(t)
	pattern := filepath.Join(root, "lifecycle", "unit_*.md")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range matches {
		name := filepath.Base(path)
		content := loadFile(t, path)

		section := extractMarkdownSection(stripFrontmatter(content), "How to End")
		if section == "" {
			t.Errorf("%s: 'How to End' section heading not found — check heading text matches exactly 'How to End'", name)
			continue
		}

		cmd := parseCloseCommand(stripFrontmatter(content))
		if cmd == "" {
			t.Errorf("%s: 'How to End' section has no extractable close command — add a line containing 'specflowctl command close', 'command close --command', or 'Terminal outcome:'", name)
		}
	}
}

// TestLifecycleMapping verifies that every UnitState (except stable_idle and
// unregistered) maps to a lifecycle file and that file actually exists.
func TestLifecycleMapping(t *testing.T) {
	root := frameworkRoot(t)
	repoRoot := filepath.Dir(root) // stateToLifecycleFile values start with framework/

	for state, file := range stateToLifecycleFile {
		absPath := filepath.Join(repoRoot, filepath.FromSlash(file))
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			t.Errorf("state %s maps to %s but file does not exist at %s", state, file, absPath)
		}
	}
}

// TestRuleSystemSections verifies rule_system.md has the sections that
// writeRuleGuidance and writeRuleBlocked extract from.
func TestRuleSystemSections(t *testing.T) {
	root := frameworkRoot(t)
	content := loadFile(t, filepath.Join(root, "governance", "rule_system.md"))

	t.Run("Rule Scopes section", func(t *testing.T) {
		section := extractMarkdownSection(content, "Rule Scopes")
		if section == "" {
			t.Error("rule_system.md: 'Rule Scopes' section not found — rule card GUIDANCE scope extraction will fail")
		}
		lower := strings.ToLower(section)
		if !strings.Contains(lower, "global") || !strings.Contains(lower, "bound") {
			t.Error("rule_system.md: 'Rule Scopes' section should describe global and bound rule scopes")
		}
	})

	t.Run("Governance Flows section", func(t *testing.T) {
		section := extractMarkdownSection(content, "Governance Flows")
		if section == "" {
			t.Error("rule_system.md: 'Governance Flows' section not found — rule card CLOSE extraction will fail")
		}
	})
}
