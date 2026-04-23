package sharedbinding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveRefRequiresPromotionOwnerUnitWhenCandidateSharedHasStableSibling(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/shared_contracts/candidate"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/shared_contracts/stable"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))

	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | note |\n")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/shared_contracts/stable/s_shared_demo.md"), `---
shared_contract_id: shared_demo
layer: stable
shared_version: 0.1.0
bound_objects:
  - unit:demo
system_constraints_stable_ref: s_system_constraints@1.1.0
---

# Stable
`)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/shared_contracts/candidate/c_shared_demo.md"), `---
shared_contract_id: shared_demo
layer: candidate
shared_version: 0.2.0
bound_objects:
  - unit:demo
system_constraints_stable_ref: s_system_constraints@1.1.0
---

# Candidate
`)

	_, err := ResolveRef(repoRoot, "candidate", "c_shared_demo@0.2.0")
	if err == nil || !strings.Contains(err.Error(), "missing promotion_owner_unit") {
		t.Fatalf("expected missing promotion_owner_unit error, got %v", err)
	}
}

func TestResolveRefAcceptsPromotionOwnerUnitWhenCandidateSharedHasStableSibling(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/shared_contracts/candidate"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/shared_contracts/stable"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))

	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), "# Spec Status\n\n## Formal Objects\n\n| Object Type | Object | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|---|\n| `unit` | `demo` | `yes` | `no` | `stable` | `unit_fork` | note |\n")
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/shared_contracts/stable/s_shared_demo.md"), `---
shared_contract_id: shared_demo
layer: stable
shared_version: 0.1.0
bound_objects:
  - unit:demo
system_constraints_stable_ref: s_system_constraints@1.1.0
---

# Stable
`)
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/shared_contracts/candidate/c_shared_demo.md"), `---
shared_contract_id: shared_demo
layer: candidate
shared_version: 0.2.0
promotion_owner_unit: demo
bound_objects:
  - unit:demo
system_constraints_stable_ref: s_system_constraints@1.1.0
---

# Candidate
`)

	resolved, err := ResolveRef(repoRoot, "candidate", "c_shared_demo@0.2.0")
	if err != nil {
		t.Fatalf("ResolveRef: %v", err)
	}
	if resolved.SharedContractID != "shared_demo" {
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
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}
