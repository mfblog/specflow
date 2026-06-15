package context

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper: write a file to the temp repo.
func writeFile(t *testing.T, repoRoot, relPath, content string) {
	t.Helper()
	absPath := filepath.Join(repoRoot, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Helper: create a minimal unit spec with frontmatter.
func writeUnitSpec(t *testing.T, repoRoot, layer, unit, version string, ruleRefs, unitRefs []string) {
	t.Helper()
	dir := "candidate"
	prefix := "c_unit_"
	if layer == "stable" {
		dir = "stable"
		prefix = "s_unit_"
	}
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("id: " + unit + "\n")
	b.WriteString("layer: " + layer + "\n")
	b.WriteString("version: " + version + "\n")
	if len(unitRefs) == 0 {
		b.WriteString("unit_refs: none\n")
	} else {
		b.WriteString("unit_refs:\n")
		for _, ref := range unitRefs {
			b.WriteString("  - " + ref + "\n")
		}
	}
	if len(ruleRefs) == 0 {
		b.WriteString("rule_refs: none\n")
	} else {
		b.WriteString("rule_refs:\n")
		for _, ref := range ruleRefs {
			b.WriteString("  - " + ref + "\n")
		}
	}
	b.WriteString("---\n")
	b.WriteString("# " + unit + " " + layer + " spec\n")
	b.WriteString("Acceptance: tested\n")
	path := "docs/specs/units/" + dir + "/" + prefix + unit + ".md"
	writeFile(t, repoRoot, path, b.String())
}

// Helper: create a rule file.
func writeRuleFile(t *testing.T, repoRoot, ruleID, layer, version string) {
	t.Helper()
	dir := "stable"
	if layer == "candidate" {
		dir = "candidate"
	}
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("rule_id: " + ruleID + "\n")
	b.WriteString("rule_scope: global\n")
	b.WriteString("layer: " + layer + "\n")
	b.WriteString("rule_version: " + version + "\n")
	b.WriteString("---\n")
	b.WriteString("# " + ruleID + "\n")
	path := "docs/specs/rules/" + dir + "/" + ruleID + ".md"
	writeFile(t, repoRoot, path, b.String())
}

// Helper: create a status file.
func writeStatus(t *testing.T, repoRoot string, rows ...string) {
	t.Helper()
	var b strings.Builder
	b.WriteString("# Status\n")
	b.WriteString("| object_type | object | stable | candidate | active_layer | next_command | notes |\n")
	b.WriteString("|---|---|---|---|---|---|---|\n")
	for _, row := range rows {
		b.WriteString("| " + row + " |\n")
	}
	writeFile(t, repoRoot, "docs/specs/_status.md", b.String())
}

// Helper: create a minimal repo with auth unit.
func createMiniRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()

	// Status
	writeStatus(t, repoRoot, "unit | auth | yes | yes | candidate | unit_verify |")

	// Candidate spec with rule ref
	writeUnitSpec(t, repoRoot, "candidate", "auth", "0.2.0",
		[]string{"s_g_rule_repository_baseline@1.1.0"},
		[]string{"s_unit_user@1.0.0"},
	)

	// Stable spec
	writeUnitSpec(t, repoRoot, "stable", "auth", "0.1.0", nil, nil)

	// Referenced unit
	writeUnitSpec(t, repoRoot, "stable", "user", "1.0.0", nil, nil)

	// Global rule
	writeRuleFile(t, repoRoot, "s_g_rule_repository_baseline", "stable", "1.1.0")

	// Candidate appendix
	writeFile(t, repoRoot, "docs/specs/units/candidate/appendix/c_unit_auth_details.md",
		"---\nunit: auth\nlayer: candidate\n---\n# Appendix details\n")

	// Check result (optional)
	writeFile(t, repoRoot, "docs/specs/_check_result/unit/auth.md",
		"---\ncheck_result: pass\n---\nCheck passed.\n")

	// Repository mapping
	writeFile(t, repoRoot, "docs/specs/repository_mapping.md",
		"# Repository Mapping\n\n## 2. Object Registry\n| kind | id | registration_state | implementation_paths | spec_files | responsibility |\n|---|---|---|---|---|---|\n| unit | auth | landed | src/auth/ | docs/specs/units/stable/s_unit_auth.md | auth module |\n")

	// Framework lifecycle overview
	writeFile(t, repoRoot, "framework/lifecycle/overview.md",
		"# Lifecycle Overview\n")

	// Unit impl trigger command Context Card
	writeFile(t, repoRoot, "framework/lifecycle/unit_impl.md",
		"# Unit Implementation (Trigger)\n")

	return repoRoot
}

func TestLifecycleCollector_UnitImpl(t *testing.T) {
	repoRoot := createMiniRepo(t)
	collector, err := NewLifecycleCollector("unit_impl")
	if err != nil {
		t.Fatalf("NewLifecycleCollector(unit_impl) failed: %v", err)
	}

	pack, err := collector.Collect(repoRoot, "auth")
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	if pack.Flow != "lifecycle" {
		t.Errorf("expected flow lifecycle, got %s", pack.Flow)
	}
	if pack.Command != "unit_impl" {
		t.Errorf("expected command unit_impl, got %s", pack.Command)
	}
	if pack.Object != "auth" {
		t.Errorf("expected object auth, got %s", pack.Object)
	}

	// Trigger command has no essential files — all files are on-demand references.
	for _, f := range pack.Files {
		if f.Essential {
			t.Errorf("unit_impl trigger command should not have essential files, got essential: %s", f.Path)
		}
	}

	// Context Card must be present as a reference.
	foundContextCard := false
	for _, f := range pack.Files {
		if f.Path == "framework/lifecycle/unit_impl.md" {
			foundContextCard = true
			break
		}
	}
	if !foundContextCard {
		t.Error("unit_impl missing Context Card reference")
	}
}

func TestLifecycleCollector_UnitCheck(t *testing.T) {
	repoRoot := createMiniRepo(t)
	collector, err := NewLifecycleCollector("unit_check")
	if err != nil {
		t.Fatalf("NewLifecycleCollector(unit_check) failed: %v", err)
	}

	pack, err := collector.Collect(repoRoot, "auth")
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	essentials := map[string]bool{}
	for _, f := range pack.Files {
		if f.Essential {
			essentials[f.Path] = f.Exists
		}
	}

	// unit_check needs _status.md.
	if !essentials["docs/specs/_status.md"] {
		t.Error("unit_check missing _status.md in essential files")
	}

	// unit_check needs candidate spec.
	if !essentials["docs/specs/units/candidate/c_unit_auth.md"] {
		t.Error("unit_check missing candidate spec")
	}

	// Context Card.
	found := false
	for _, f := range pack.Files {
		if f.Path == "framework/lifecycle/unit_check.md" {
			found = true
			break
		}
	}
	if !found {
		t.Error("unit_check missing Context Card reference")
	}
}

func TestLifecycleCollector_UnitVerify(t *testing.T) {
	repoRoot := createMiniRepo(t)
	collector, err := NewLifecycleCollector("unit_verify")
	if err != nil {
		t.Fatalf("NewLifecycleCollector(unit_verify) failed: %v", err)
	}

	pack, err := collector.Collect(repoRoot, "auth")
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	essentials := map[string]bool{}
	for _, f := range pack.Files {
		if f.Essential {
			essentials[f.Path] = f.Exists
		}
	}

	// unit_verify needs _status.md (as a lifecycle command).
	if !essentials["docs/specs/_status.md"] {
		t.Error("unit_verify missing _status.md in essential files")
	}

	// unit_verify needs candidate + stable.
	if !essentials["docs/specs/units/candidate/c_unit_auth.md"] {
		t.Error("unit_verify missing candidate spec")
	}
	if !essentials["docs/specs/units/stable/s_unit_auth.md"] {
		t.Error("unit_verify missing stable spec")
	}
}

func TestLifecycleCollector_UnitPromote(t *testing.T) {
	repoRoot := createMiniRepo(t)
	collector, err := NewLifecycleCollector("unit_promote")
	if err != nil {
		t.Fatalf("NewLifecycleCollector(unit_promote) failed: %v", err)
	}

	pack, err := collector.Collect(repoRoot, "auth")
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	essentials := map[string]bool{}
	for _, f := range pack.Files {
		if f.Essential {
			essentials[f.Path] = f.Exists
		}
	}

	// unit_promote needs _status.md, candidate, stable, appendixes.
	if !essentials["docs/specs/_status.md"] {
		t.Error("unit_promote missing _status.md")
	}
	if !essentials["docs/specs/units/candidate/c_unit_auth.md"] {
		t.Error("unit_promote missing candidate spec")
	}
	if !essentials["docs/specs/units/stable/s_unit_auth.md"] {
		t.Error("unit_promote missing stable spec")
	}
	if !essentials["docs/specs/units/candidate/appendix/c_unit_auth_details.md"] {
		t.Error("unit_promote missing appendix")
	}

	// Verify result in reference.
	found := false
	for _, f := range pack.Files {
		if f.Path == "docs/specs/_verify_result/unit/auth.md" {
			found = true
			break
		}
	}
	if !found {
		t.Error("unit_promote missing verify result reference")
	}
}

func TestLifecycleCollector_UnitFork(t *testing.T) {
	repoRoot := createMiniRepo(t)
	collector, err := NewLifecycleCollector("unit_fork")
	if err != nil {
		t.Fatalf("NewLifecycleCollector(unit_fork) failed: %v", err)
	}

	pack, err := collector.Collect(repoRoot, "auth")
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	essentials := map[string]bool{}
	for _, f := range pack.Files {
		if f.Essential {
			essentials[f.Path] = f.Exists
		}
	}

	// unit_fork needs _status.md, repository_mapping.md, stable spec.
	if !essentials["docs/specs/_status.md"] {
		t.Error("unit_fork missing _status.md")
	}
	if !essentials["docs/specs/repository_mapping.md"] {
		t.Error("unit_fork missing repository_mapping.md")
	}
	if !essentials["docs/specs/units/stable/s_unit_auth.md"] {
		t.Error("unit_fork missing stable spec")
	}
}

func TestLifecycleCollector_UnitStableVerify(t *testing.T) {
	repoRoot := createMiniRepo(t)
	collector, err := NewLifecycleCollector("unit_stable_verify")
	if err != nil {
		t.Fatalf("NewLifecycleCollector(unit_stable_verify) failed: %v", err)
	}

	pack, err := collector.Collect(repoRoot, "auth")
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	essentials := map[string]bool{}
	for _, f := range pack.Files {
		if f.Essential {
			essentials[f.Path] = f.Exists
		}
	}

	// Must include _status.md, repository_mapping.md, stable spec, rules.
	if !essentials["docs/specs/_status.md"] {
		t.Error("unit_stable_verify missing _status.md")
	}
	if !essentials["docs/specs/repository_mapping.md"] {
		t.Error("unit_stable_verify missing repository_mapping.md")
	}
	if !essentials["docs/specs/units/stable/s_unit_auth.md"] {
		t.Error("unit_stable_verify missing stable spec")
	}
	if !essentials["docs/specs/rules/stable/s_g_rule_repository_baseline.md"] {
		t.Error("unit_stable_verify missing rule file")
	}
}

func TestLifecycleCollector_UnitAdvance(t *testing.T) {
	repoRoot := createMiniRepo(t)
	collector, err := NewLifecycleCollector("unit_advance")
	if err != nil {
		t.Fatalf("NewLifecycleCollector(unit_advance) failed: %v", err)
	}

	pack, err := collector.Collect(repoRoot, "auth")
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	essentials := map[string]bool{}
	for _, f := range pack.Files {
		if f.Essential {
			essentials[f.Path] = f.Exists
		}
	}

	if !essentials["docs/specs/_status.md"] {
		t.Error("unit_advance missing _status.md")
	}
	if !essentials["docs/specs/units/candidate/c_unit_auth.md"] {
		t.Error("unit_advance missing candidate spec")
	}
	if !essentials["framework/lifecycle/overview.md"] {
		t.Error("unit_advance missing overview.md")
	}
}

func TestCollector_UnknownCommand(t *testing.T) {
	_, err := NewLifecycleCollector("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown command, got nil")
	}
}

func TestRenderPack(t *testing.T) {
	repoRoot := createMiniRepo(t)
	collector, _ := NewLifecycleCollector("unit_impl")
	pack, _ := collector.Collect(repoRoot, "auth")

	var buf bytes.Buffer
	if err := pack.Render(&buf, FormatPack); err != nil {
		t.Fatalf("Render pack failed: %v", err)
	}

	output := buf.String()

	// Must contain header
	if !strings.Contains(output, "Context Pack for lifecycle/unit_impl:auth") {
		t.Error("pack missing header")
	}

	// Must contain "Core Truth" section or no essential files
	if len(pack.Files) > 0 {
		hasEssential := false
		for _, f := range pack.Files {
			if f.Essential {
				hasEssential = true
				break
			}
		}
		if hasEssential {
			if !strings.Contains(output, "## Core Truth") {
				t.Error("pack missing Core Truth section for essential files")
			}
			if !strings.Contains(output, "Acceptance: tested") {
				t.Error("pack missing inlined spec content")
			}
			if !strings.Contains(output, "```markdown") {
				t.Error("pack missing markdown code fences")
			}
		}
	}

	// Must contain "References" section
	if !strings.Contains(output, "## References") {
		t.Error("pack missing References section")
	}

	// Must contain "Inventory"
	if !strings.Contains(output, "## Inventory") {
		t.Error("pack missing Inventory section")
	}

	// Context Card reference must be present
	if !strings.Contains(output, "framework/lifecycle/unit_impl.md") {
		t.Error("pack missing Context Card reference")
	}
}

func TestRenderRefs(t *testing.T) {
	repoRoot := createMiniRepo(t)
	collector, _ := NewLifecycleCollector("unit_impl")
	pack, _ := collector.Collect(repoRoot, "auth")

	var buf bytes.Buffer
	if err := pack.Render(&buf, FormatRefs); err != nil {
		t.Fatalf("Render refs failed: %v", err)
	}

	output := buf.String()

	// Trigger command has no essential files — only reference section
	if strings.Contains(output, "essential:") {
		t.Error("unit_impl trigger command should not have essential files in refs output")
	}
	if !strings.Contains(output, "reference:") {
		t.Error("refs missing reference section")
	}

	// Context Card must be listed
	if !strings.Contains(output, "framework/lifecycle/unit_impl.md") {
		t.Error("refs missing Context Card reference")
	}
}

func TestCollector_MissingObject(t *testing.T) {
	repoRoot := createMiniRepo(t)
	collector, _ := NewLifecycleCollector("unit_impl")

	// Object "nonexistent" — should not error, just produce minimal pack.
	pack, err := collector.Collect(repoRoot, "nonexistent")
	if err != nil {
		t.Fatalf("Collect with nonexistent object should not error: %v", err)
	}

	// Trigger command files (Context Card) should still resolve regardless of object name.
	foundContextCard := false
	for _, f := range pack.Files {
		if f.Path == "framework/lifecycle/unit_impl.md" && f.Exists {
			foundContextCard = true
		}
	}
	if !foundContextCard {
		t.Error("trigger command missing Context Card reference despite object name")
	}
}

func TestPackInventory(t *testing.T) {
	repoRoot := createMiniRepo(t)
	collector, _ := NewLifecycleCollector("unit_impl")
	pack, _ := collector.Collect(repoRoot, "auth")

	inv := pack.Inventory()
	if !strings.Contains(inv, "Total:") {
		t.Error("inventory missing Total")
	}
	if !strings.Contains(inv, "inlined") {
		t.Error("inventory missing inlined count")
	}
	if !strings.Contains(inv, "referenced") {
		t.Error("inventory missing referenced count")
	}
}

func TestNewLifecycleCollector_Map(t *testing.T) {
	commands, collectors := RegisterLifecycleCommands()

	expected := []string{"unit_advance", "unit_check", "unit_fork", "unit_impl", "unit_init", "unit_new", "unit_promote", "unit_stable_verify"}
	for _, cmd := range expected {
		found := false
		for _, registered := range commands {
			if registered == cmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected command %s not registered", cmd)
		}
		if _, ok := collectors[cmd]; !ok {
			t.Errorf("collector for %s not found", cmd)
		}
	}
}

func TestFileItemContent(t *testing.T) {
	repoRoot := createMiniRepo(t)

	// Use unit_check (has essential files) to verify inlined content.
	collector, _ := NewLifecycleCollector("unit_check")
	pack, _ := collector.Collect(repoRoot, "auth")

	for _, f := range pack.Files {
		if f.Path == "docs/specs/units/candidate/c_unit_auth.md" && f.Exists {
			if !strings.Contains(f.Content, "candidate") {
				t.Error("inlined candidate spec content doesn't contain expected text")
			}
			if f.LineCount == 0 {
				t.Error("inlined file has 0 line count")
			}
		}
	}
}

func TestLifecycleCollector_UnitNew(t *testing.T) {
	repoRoot := createMiniRepo(t)
	collector, err := NewLifecycleCollector("unit_new")
	if err != nil {
		t.Fatalf("NewLifecycleCollector(unit_new) failed: %v", err)
	}

	pack, err := collector.Collect(repoRoot, "auth")
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	essentials := map[string]bool{}
	for _, f := range pack.Files {
		if f.Essential {
			essentials[f.Path] = f.Exists
		}
	}

	if !essentials["docs/specs/_status.md"] {
		t.Error("unit_new missing _status.md")
	}
	if !essentials["docs/specs/repository_mapping.md"] {
		t.Error("unit_new missing repository_mapping.md")
	}
}

func TestLifecycleCollector_UnitInit(t *testing.T) {
	repoRoot := createMiniRepo(t)
	collector, err := NewLifecycleCollector("unit_init")
	if err != nil {
		t.Fatalf("NewLifecycleCollector(unit_init) failed: %v", err)
	}

	pack, err := collector.Collect(repoRoot, "auth")
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	essentials := map[string]bool{}
	for _, f := range pack.Files {
		if f.Essential {
			essentials[f.Path] = f.Exists
		}
	}

	if !essentials["docs/specs/_status.md"] {
		t.Error("unit_init missing _status.md")
	}
	if !essentials["docs/specs/repository_mapping.md"] {
		t.Error("unit_init missing repository_mapping.md")
	}
}

func TestResolveRuleRefs_NoDuplicateGlobals(t *testing.T) {
	repoRoot := createMiniRepo(t)

	items, err := resolveRuleRefs(repoRoot, "auth")
	if err != nil {
		t.Fatalf("resolveRuleRefs failed: %v", err)
	}

	count := 0
	for _, item := range items {
		if item.Path == "docs/specs/rules/stable/s_g_rule_repository_baseline.md" {
			count++
		}
	}
	if count > 1 {
		t.Errorf("global rule appears %d times (expected 1)", count)
	}
}
