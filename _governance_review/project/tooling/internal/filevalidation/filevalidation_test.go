package filevalidation

import (
	"testing"
)

func TestValidateWriteAllowedByDefault(t *testing.T) {
	c := Constraints{}
	r := ValidateWrite("unit_impl", "src/foo.go", c)
	if !r.Allowed {
		t.Errorf("expected allowed by default, got denied: %s", r.Reason)
	}
}

func TestValidateWriteForbiddenPattern(t *testing.T) {
	c := Constraints{
		ForbiddenWrites: []WriteRule{
			{Pattern: "docs/specs/stable/*"},
		},
		AllowedWrites: []WriteRule{
			{Pattern: "src/*", Phases: []string{"unit_impl"}},
		},
	}
	r := ValidateWrite("unit_impl", "docs/specs/stable/s_unit_x.md", c)
	if r.Allowed {
		t.Errorf("expected denied for forbidden pattern, got allowed")
	}
	if r.Reason == "" {
		t.Errorf("expected non-empty reason")
	}
}

func TestValidateWriteAllowedMatch(t *testing.T) {
	c := Constraints{
		AllowedWrites: []WriteRule{
			{Pattern: "src/*", Phases: []string{"unit_impl"}},
		},
	}
	r := ValidateWrite("unit_impl", "src/foo.go", c)
	if !r.Allowed {
		t.Errorf("expected allowed for matching pattern, got denied: %s", r.Reason)
	}
}

func TestValidateWriteNoAllowedMatch(t *testing.T) {
	c := Constraints{
		AllowedWrites: []WriteRule{
			{Pattern: "src/*", Phases: []string{"unit_impl"}},
		},
	}
	r := ValidateWrite("unit_impl", "other/foo.go", c)
	if r.Allowed {
		t.Errorf("expected denied for non-matching pattern, got allowed")
	}
}

func TestValidateWritePhaseMismatch(t *testing.T) {
	c := Constraints{
		AllowedWrites: []WriteRule{
			{Pattern: "src/*", Phases: []string{"unit_impl"}},
		},
	}
	r := ValidateWrite("unit_check", "src/foo.go", c)
	if r.Allowed {
		t.Errorf("expected denied for phase mismatch, got allowed")
	}
}

func TestValidateWritePhaseMatchAnyWhenEmpty(t *testing.T) {
	c := Constraints{
		AllowedWrites: []WriteRule{
			{Pattern: "src/*"}, // no phases = all phases
		},
	}
	r := ValidateWrite("unit_verify", "src/foo.go", c)
	if !r.Allowed {
		t.Errorf("expected allowed for empty phases (all), got denied: %s", r.Reason)
	}
}

func TestValidateWriteForbiddenTakesPrecedence(t *testing.T) {
	c := Constraints{
		ForbiddenWrites: []WriteRule{
			{Pattern: "src/secret/*"},
		},
		AllowedWrites: []WriteRule{
			{Pattern: "src/*"},
		},
	}
	r := ValidateWrite("unit_impl", "src/secret/key.go", c)
	if r.Allowed {
		t.Errorf("expected forbidden to take precedence, got allowed")
	}
}

func TestValidateWriteGlobDoubleStarPattern(t *testing.T) {
	c := Constraints{
		AllowedWrites: []WriteRule{
			{Pattern: "src/my_feature/**"},
		},
	}
	r := ValidateWrite("unit_impl", "src/my_feature/internal/handler.go", c)
	if !r.Allowed {
		t.Errorf("expected allowed for ** pattern, got denied: %s", r.Reason)
	}
}

func TestValidateWriteDoubleStarMatchPrefix(t *testing.T) {
	c := Constraints{
		AllowedWrites: []WriteRule{
			{Pattern: "src/**/*.go"},
		},
	}
	r := ValidateWrite("unit_impl", "src/my_feature/handler.go", c)
	if !r.Allowed {
		t.Errorf("expected allowed for src/**/*.go pattern, got denied: %s", r.Reason)
	}
}

func TestValidateWritePathCleaned(t *testing.T) {
	c := Constraints{
		AllowedWrites: []WriteRule{
			{Pattern: "src/*"},
		},
	}
	r := ValidateWrite("unit_impl", "./src/foo.go", c)
	if !r.Allowed {
		t.Errorf("expected allowed for path with ./ prefix, got denied: %s", r.Reason)
	}
}

func TestParseConstraintsEmpty(t *testing.T) {
	c, err := ParseConstraints("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.AllowedWrites) != 0 {
		t.Errorf("expected 0 allowed, got %d", len(c.AllowedWrites))
	}
	if len(c.ForbiddenWrites) != 0 {
		t.Errorf("expected 0 forbidden, got %d", len(c.ForbiddenWrites))
	}
}

func TestParseConstraintsBasic(t *testing.T) {
	input := `allowed_writes:
  - pattern: "src/my_feature/**"
    phases: [unit_impl, unit_verify]
  - pattern: "tests/my_feature/**"
    phases: [unit_impl, unit_verify]
forbidden_writes:
  - pattern: "docs/specs/units/stable/**"
  - pattern: "docs/specs/_status.md"`

	c, err := ParseConstraints(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.AllowedWrites) != 2 {
		t.Errorf("expected 2 allowed, got %d", len(c.AllowedWrites))
	}
	if len(c.ForbiddenWrites) != 2 {
		t.Errorf("expected 2 forbidden, got %d", len(c.ForbiddenWrites))
	}
	if c.AllowedWrites[0].Pattern != "src/my_feature/**" {
		t.Errorf("expected src/my_feature/**, got %q", c.AllowedWrites[0].Pattern)
	}
	if len(c.AllowedWrites[0].Phases) != 2 {
		t.Errorf("expected 2 phases, got %d", len(c.AllowedWrites[0].Phases))
	}
}

func TestValidateWriteIntegration(t *testing.T) {
	input := `allowed_writes:
  - pattern: "src/my_feature/**"
    phases: [unit_impl, unit_verify]
  - pattern: "tests/my_feature/**"
    phases: [unit_verify]
forbidden_writes:
  - pattern: "docs/specs/units/stable/**"
  - pattern: "docs/specs/_status.md"`

	c, err := ParseConstraints(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	tests := []struct {
		path    string
		phase   string
		allowed bool
	}{
		{path: "src/my_feature/handler.go", phase: "unit_impl", allowed: true},
		{path: "src/my_feature/handler.go", phase: "unit_verify", allowed: true},
		{path: "src/my_feature/handler.go", phase: "unit_check", allowed: false},
		{path: "tests/my_feature/main_test.go", phase: "unit_verify", allowed: true},
		{path: "tests/my_feature/main_test.go", phase: "unit_impl", allowed: false},
		{path: "docs/specs/units/stable/s_unit_x.md", phase: "unit_verify", allowed: false},
		{path: "docs/specs/_status.md", phase: "unit_verify", allowed: false},
		{path: "other/file.go", phase: "unit_impl", allowed: false},
	}

	for _, tt := range tests {
		r := ValidateWrite(tt.phase, tt.path, c)
		if r.Allowed != tt.allowed {
			t.Errorf("ValidateWrite(%q, %q): expected allowed=%v, got %v (reason: %s)", tt.path, tt.phase, tt.allowed, r.Allowed, r.Reason)
		}
	}
}
