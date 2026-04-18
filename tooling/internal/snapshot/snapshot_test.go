package snapshot

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
)

func TestRebuildCurrentCollectsAppendixAndSharedSnapshot(t *testing.T) {
	repoRoot := t.TempDir()
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs"))
	mustMkdirAll(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateAppendixDir)))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/shared_contracts/candidate"))
	mustMkdirAll(t, filepath.Join(repoRoot, "docs/specs/system/stable"))

	status := "# Spec Status\n\n## Formal Modules\n\n| Module | Stable | Candidate | Active Layer | Next Command | Notes |\n|---|---|---|---|---|---|\n| `module_demo` | `no` | `yes` | `candidate` | `cand_check` | note |\n"
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/_status.md"), status)

	mainSpec := `---
id: module_demo
layer: candidate
version: 0.1.0
---

# Demo

See [appendix](./appendix/c_module_demo_prompt.md).

## Global Constraint Alignment

1. ` + "`system_constraints_stable_ref`: `s_system_constraints@1.1.0`" + `
2. ` + "`shared_contract_refs`:" + `
   - ` + "`c_shared_demo@0.2.0`" + `
`
	mainSpecRef, err := specpaths.MainSpecFileRef("candidate", "module_demo")
	if err != nil {
		t.Fatalf("MainSpecFileRef: %v", err)
	}
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(mainSpecRef)), mainSpec)

	appendix := `---
module: module_demo
layer: candidate
spec_version_ref: c_module_demo@0.1.0
---

# Appendix
`
	mustWriteFile(t, filepath.Join(repoRoot, filepath.FromSlash(specpaths.CandidateAppendixDir), "c_module_demo_prompt.md"), appendix)

	shared := `---
shared_contract_id: shared_demo
layer: candidate
shared_version: 0.2.0
bound_modules:
  - module_demo
system_constraints_stable_ref: s_system_constraints@1.1.0
---

# Shared
`
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/shared_contracts/candidate/c_shared_demo.md"), shared)

	system := `---
version: 1.1.0
---

# System
`
	mustWriteFile(t, filepath.Join(repoRoot, "docs/specs/system/stable/s_system_constraints.md"), system)

	result, err := RebuildCurrent(repoRoot, "module_demo")
	if err != nil {
		t.Fatalf("RebuildCurrent: %v", err)
	}
	if result.SpecFileRef != mainSpecRef {
		t.Fatalf("unexpected spec file ref: %s", result.SpecFileRef)
	}
	if result.SpecVersionRef != "c_module_demo@0.1.0" {
		t.Fatalf("unexpected spec version ref: %s", result.SpecVersionRef)
	}
	if len(result.ModuleAppendixSnapshot) != 1 {
		t.Fatalf("expected one appendix snapshot entry, got %d", len(result.ModuleAppendixSnapshot))
	}
	if result.ModuleAppendixSnapshot[0].AppendixRef != "c_module_demo@0.1.0" {
		t.Fatalf("unexpected appendix ref: %s", result.ModuleAppendixSnapshot[0].AppendixRef)
	}
	if result.SystemConstraintsStableVersionRef != "s_system_constraints@1.1.0" {
		t.Fatalf("unexpected system constraints version ref: %s", result.SystemConstraintsStableVersionRef)
	}
	if len(result.SharedContractSnapshot) != 1 {
		t.Fatalf("expected one shared snapshot entry, got %d", len(result.SharedContractSnapshot))
	}
	if result.SharedContractSnapshot[0].VersionRef != "c_shared_demo@0.2.0" {
		t.Fatalf("unexpected shared version ref: %s", result.SharedContractSnapshot[0].VersionRef)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
