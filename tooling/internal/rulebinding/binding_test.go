package rulebinding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/testfixtures"
)

func TestResolveRefRequiresPromotionOwnerUnitWhenCandidateSharedHasStableSibling(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/stable"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/candidate"))

	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_b_rule_demo.md"), `---
rule_id: shared_demo
rule_scope: bound
layer: stable
rule_version: 0.1.0
---

# Stable
`)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md"), `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.2.0
---

# Candidate
`)

	_, err := ResolveRef(repoRoot, "candidate", "c_b_rule_demo@0.2.0")
	if err == nil || !strings.Contains(err.Error(), "missing promotion_owner_unit") {
		t.Fatalf("expected missing promotion_owner_unit error, got %v", err)
	}
}

func TestResolveRefAcceptsPromotionOwnerUnitWhenCandidateSharedHasStableSibling(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/stable"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/rules/candidate"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/units/stable"))

	// Create spec file for promotion_owner_unit (file existence is state)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_demo.md"), "---\nid: demo\nlayer: stable\nversion: 1.0.0\n---\n# Demo\n")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/stable/s_b_rule_demo.md"), `---
rule_id: shared_demo
rule_scope: bound
layer: stable
rule_version: 0.1.0
---

# Stable
`)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/rules/candidate/c_b_rule_demo.md"), `---
rule_id: shared_demo
rule_scope: bound
layer: candidate
rule_version: 0.2.0
promotion_owner_unit: demo
---

# Candidate
`)

	resolved, err := ResolveRef(repoRoot, "candidate", "c_b_rule_demo@0.2.0")
	if err != nil {
		t.Fatalf("ResolveRef: %v", err)
	}
	if resolved.RuleID != "shared_demo" {
		t.Fatalf("unexpected resolved ref: %+v", resolved)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	content = testfixtures.NormalizeSpecFlowContent(path, content)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}
